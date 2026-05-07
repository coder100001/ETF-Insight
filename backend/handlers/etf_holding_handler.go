package handlers

import (
	"fmt"
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/services/etf"
	"etf-insight/services/event"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ETFHoldingHandler ETF持仓处理器
type ETFHoldingHandler struct {
	db              *gorm.DB
	holdingsService *etf.CachedHoldingsService
	cacheService    *etf.OverlapCacheService
}

// NewETFHoldingHandler 创建ETF持仓处理器（带缓存）
func NewETFHoldingHandler(db *gorm.DB) *ETFHoldingHandler {
	return &ETFHoldingHandler{
		db:              db,
		holdingsService: etf.NewCachedHoldingsService(db),
		cacheService:    etf.NewOverlapCacheService(db),
	}
}

// GetETFHoldings 获取ETF底层持仓
// GET /api/etf/:symbol/holdings
func (h *ETFHoldingHandler) GetETFHoldings(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ETF代码不能为空",
			"code":    "INVALID_SYMBOL",
		})
		return
	}

	// 解析日期参数（可选）
	var date time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效，请使用 YYYY-MM-DD 格式",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	// 获取持仓数据
	holdings, err := h.holdingsService.GetETFHoldings(c.Request.Context(), symbol, date)
	if err != nil {
		utils.Error("获取ETF持仓失败", err, "symbol", symbol)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到该ETF的持仓数据",
			"code":    "HOLDINGS_NOT_FOUND",
		})
		return
	}

	// 转换为响应格式
	var responseHoldings []gin.H
	for _, h := range holdings {
		responseHoldings = append(responseHoldings, gin.H{
			"symbol":       h.Symbol,
			"name":         h.Name,
			"weight":       h.Weight,
			"shares":       h.Shares,
			"market_value": h.MarketValue,
			"date":         h.Date.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"symbol":   symbol,
			"date":     getHoldingsDate(holdings),
			"holdings": responseHoldings,
			"total":    len(holdings),
		},
	})
}

// GetETFOverlap 计算两只ETF的持仓重叠度
// GET /api/etf/overlap?sym1=XXX&sym2=YYY&date=YYYY-MM-DD
func (h *ETFHoldingHandler) GetETFOverlap(c *gin.Context) {
	sym1 := c.Query("sym1")
	sym2 := c.Query("sym2")

	if sym1 == "" || sym2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请提供两个ETF代码（sym1和sym2）",
			"code":    "MISSING_SYMBOLS",
		})
		return
	}

	if sym1 == sym2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "不能计算同一只ETF的重叠度",
			"code":    "SAME_SYMBOL",
		})
		return
	}

	// 解析日期参数（可选）
	var date time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效，请使用 YYYY-MM-DD 格式",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	// 计算重叠度（使用缓存）
	result, err := h.holdingsService.CalculateOverlapWithCache(c.Request.Context(), sym1, sym2, date)
	if err != nil {
		utils.Error("计算ETF重叠度失败", err, "sym1", sym1, "sym2", sym2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "计算重叠度失败: " + err.Error(),
			"code":    "OVERLAP_CALCULATION_ERROR",
		})
		return
	}

	// 转换明细为响应格式
	var details []gin.H
	for _, d := range result.Details {
		details = append(details, gin.H{
			"symbol":     d.Symbol,
			"name":       d.Name,
			"weight_a":   d.WeightA,
			"weight_b":   d.WeightB,
			"min_weight": d.MinWeight,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"etf_a":           result.ETFA,
			"etf_b":           result.ETFB,
			"overlap_score":   result.OverlapScore,
			"common_holdings": result.CommonHoldings,
			"total_weight_a":  result.TotalWeightA,
			"total_weight_b":  result.TotalWeightB,
			"details":         details,
			"calculated_at":   result.CalculatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// GetETFHoldingsComparison 对比多只ETF的持仓
// POST /api/etf/holdings/comparison
func (h *ETFHoldingHandler) GetETFHoldingsComparison(c *gin.Context) {
	var req struct {
		Symbols []string `json:"symbols" binding:"required,min=2,max=5"`
		Date    string   `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
			"code":    "INVALID_REQUEST",
		})
		return
	}

	// 解析日期
	var date time.Time
	if req.Date != "" {
		var err error
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	// 获取每只ETF的持仓
	comparison := make(map[string]any)
	for _, symbol := range req.Symbols {
		holdings, err := h.holdingsService.GetETFHoldings(c.Request.Context(), symbol, date)
		if err != nil {
			utils.Warn("获取ETF持仓失败", "symbol", symbol, "error", err)
			continue
		}

		var holdingList []gin.H
		for _, h := range holdings {
			holdingList = append(holdingList, gin.H{
				"symbol": h.Symbol,
				"name":   h.Name,
				"weight": h.Weight,
			})
		}

		comparison[symbol] = gin.H{
			"holdings":      holdingList,
			"total":         len(holdings),
			"holdings_date": getHoldingsDate(holdings),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    comparison,
	})
}

// GetTopHoldings 获取ETF前N大持仓
// GET /api/etf/:symbol/top-holdings?n=10
func (h *ETFHoldingHandler) GetTopHoldings(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ETF代码不能为空",
			"code":    "INVALID_SYMBOL",
		})
		return
	}

	// 解析n参数
	n := 10
	if nStr := c.Query("n"); nStr != "" {
		if parsedN, err := parseInt(nStr); err == nil && parsedN > 0 && parsedN <= 50 {
			n = parsedN
		}
	}

	// 解析日期参数（可选）
	var date time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	holdings, err := h.holdingsService.GetTopHoldings(c.Request.Context(), symbol, n, date)
	if err != nil {
		utils.Error("获取前十大持仓失败", err, "symbol", symbol)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到持仓数据",
			"code":    "HOLDINGS_NOT_FOUND",
		})
		return
	}

	var responseHoldings []gin.H
	for _, h := range holdings {
		responseHoldings = append(responseHoldings, gin.H{
			"symbol":       h.Symbol,
			"name":         h.Name,
			"weight":       h.Weight,
			"shares":       h.Shares,
			"market_value": h.MarketValue,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"symbol":   symbol,
			"top_n":    n,
			"holdings": responseHoldings,
		},
	})
}

// GetSectorAllocation 获取ETF行业分布
// GET /api/etf/:symbol/sector-allocation
func (h *ETFHoldingHandler) GetSectorAllocation(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ETF代码不能为空",
			"code":    "INVALID_SYMBOL",
		})
		return
	}

	var date time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "日期格式无效",
				"code":    "INVALID_DATE_FORMAT",
			})
			return
		}
	}

	sectorAllocation, err := h.holdingsService.GetSectorAllocation(c.Request.Context(), symbol, date)
	if err != nil {
		utils.Error("获取行业分布失败", err, "symbol", symbol)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取行业分布失败",
			"code":    "SECTOR_ALLOCATION_ERROR",
		})
		return
	}

	// 转换为有序数组
	var sectors []gin.H
	for sector, weight := range sectorAllocation {
		sectors = append(sectors, gin.H{
			"sector": sector,
			"weight": weight,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"symbol":  symbol,
			"sectors": sectors,
		},
	})
}

// SaveETFHoldings 保存ETF持仓数据（管理员接口）
// POST /api/etf/:symbol/holdings
func (h *ETFHoldingHandler) SaveETFHoldings(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ETF代码不能为空",
			"code":    "INVALID_SYMBOL",
		})
		return
	}

	var req struct {
		Holdings []struct {
			Symbol      string          `json:"symbol" binding:"required"`
			Name        string          `json:"name" binding:"required"`
			Weight      decimal.Decimal `json:"weight" binding:"required"`
			Shares      int64           `json:"shares"`
			MarketValue decimal.Decimal `json:"market_value"`
		} `json:"holdings" binding:"required,min=1"`
		Date string `json:"date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
			"code":    "INVALID_REQUEST",
		})
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "日期格式无效",
			"code":    "INVALID_DATE_FORMAT",
		})
		return
	}

	// 转换为模型
	var holdings []models.ETFHolding
	for _, h := range req.Holdings {
		holdings = append(holdings, models.ETFHolding{
			Symbol:      h.Symbol,
			Name:        h.Name,
			Weight:      h.Weight,
			Shares:      h.Shares,
			MarketValue: h.MarketValue,
			Date:        date,
			DataSource:  "manual",
		})
	}

	if err := h.holdingsService.SaveHoldings(c.Request.Context(), symbol, holdings); err != nil {
		utils.Error("保存持仓数据失败", err, "symbol", symbol)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存持仓数据失败",
			"code":    "SAVE_HOLDINGS_ERROR",
		})
		return
	}

	// 触发缓存失效事件
	if event.GlobalEventBus != nil {
		go event.GlobalEventBus.PublishETFHoldingsUpdated(c.Request.Context(), symbol, "api")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"symbol":  symbol,
			"date":    req.Date,
			"count":   len(holdings),
			"message": "持仓数据保存成功",
		},
	})
}

// getHoldingsDate 获取持仓日期
func getHoldingsDate(holdings []models.ETFHolding) string {
	if len(holdings) > 0 {
		return holdings[0].Date.Format("2006-01-02")
	}
	return ""
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// GetCacheStats 获取缓存统计信息
// GET /api/cache/overlap/stats
func (h *ETFHoldingHandler) GetCacheStats(c *gin.Context) {
	stats, err := h.cacheService.GetCacheStats(c.Request.Context())
	if err != nil {
		utils.Error("获取缓存统计失败", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取缓存统计失败",
			"code":    "CACHE_STATS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// InvalidateCache 使指定ETF的缓存失效
// POST /api/cache/overlap/invalidate
func (h *ETFHoldingHandler) InvalidateCache(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效: " + err.Error(),
			"code":    "INVALID_REQUEST",
		})
		return
	}

	if err := h.cacheService.InvalidateOverlapCache(c.Request.Context(), req.Symbol); err != nil {
		utils.Error("使缓存失效失败", err, "symbol", req.Symbol)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "使缓存失效失败: " + err.Error(),
			"code":    "CACHE_INVALIDATE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"symbol":  req.Symbol,
			"message": "缓存已失效，下次请求将重新计算",
		},
	})
}

// CleanExpiredCache 清理过期缓存
// POST /api/cache/overlap/clean
func (h *ETFHoldingHandler) CleanExpiredCache(c *gin.Context) {
	deletedCount, err := h.cacheService.CleanExpiredCache(c.Request.Context())
	if err != nil {
		utils.Error("清理过期缓存失败", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "清理过期缓存失败: " + err.Error(),
			"code":    "CACHE_CLEAN_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"deleted_count": deletedCount,
			"message":       "过期缓存已清理",
		},
	})
}
