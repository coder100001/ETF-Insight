package handlers

import (
	"net/http"
	"time"

	"etf-insight/models"
	"etf-insight/services/ashare"

	"github.com/gin-gonic/gin"
)

// AShareDataHandler A股数据处理器
type AShareDataHandler struct {
	etfService *ashare.ETFDataService
}

// NewAShareDataHandler 创建A股数据处理器
func NewAShareDataHandler() *AShareDataHandler {
	return &AShareDataHandler{
		etfService: ashare.NewETFDataService(models.DB),
	}
}

// EnableAKShareRequest 启用AKShare请求
type EnableAKShareRequest struct {
	BaseURL string `json:"base_url"` // Python服务地址，如 http://localhost:8000
}

// EnableAKShareResponse 启用AKShare响应
type EnableAKShareResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// EnableAKShare 启用AKShare数据源
// @Summary 启用AKShare数据源
// @Description 配置并启用AKShare数据源
// @Tags a-share-data
// @Accept json
// @Produce json
// @Param request body EnableAKShareRequest true "配置参数"
// @Success 200 {object} EnableAKShareResponse
// @Router /api/a-share/enable-akshare [post]
func (h *AShareDataHandler) EnableAKShare(c *gin.Context) {
	var req EnableAKShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, EnableAKShareResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	if req.BaseURL == "" {
		req.BaseURL = "http://localhost:8000"
	}

	h.etfService.EnableAKShare(req.BaseURL)

	c.JSON(http.StatusOK, EnableAKShareResponse{
		Success: true,
		Message: "AKShare数据源已启用: " + req.BaseURL,
	})
}

// SyncETFListResponse 同步ETF列表响应
type SyncETFListResponse struct {
	Success bool   `json:"success"`
	Count   int    `json:"count"`
	Error   string `json:"error,omitempty"`
}

// SyncETFList 同步ETF列表
// @Summary 同步ETF列表
// @Description 从AKShare同步A股ETF列表到数据库
// @Tags a-share-data
// @Produce json
// @Success 200 {object} SyncETFListResponse
// @Router /api/a-share/sync-etf-list [post]
func (h *AShareDataHandler) SyncETFList(c *gin.Context) {
	if err := h.etfService.SyncETFList(); err != nil {
		c.JSON(http.StatusInternalServerError, SyncETFListResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// 查询同步后的数量
	var count int64
	models.DB.Model(&models.AShareDividendETF{}).Count(&count)

	c.JSON(http.StatusOK, SyncETFListResponse{
		Success: true,
		Count:   int(count),
	})
}

// SyncETFPricesResponse 同步ETF价格响应
type SyncETFPricesResponse struct {
	Success bool   `json:"success"`
	Count   int    `json:"count"`
	Error   string `json:"error,omitempty"`
}

// SyncETFPrices 同步ETF价格
// @Summary 同步ETF价格
// @Description 从AKShare同步所有ETF的实时价格
// @Tags a-share-data
// @Produce json
// @Success 200 {object} SyncETFPricesResponse
// @Router /api/a-share/sync-prices [post]
func (h *AShareDataHandler) SyncETFPrices(c *gin.Context) {
	var etfs []models.AShareDividendETF
	models.DB.Where("status = ?", 1).Pluck("symbol", &etfs)

	symbols := make([]string, len(etfs))
	for i, etf := range etfs {
		symbols[i] = etf.Symbol
	}

	if err := h.etfService.SyncETFPrices(symbols); err != nil {
		c.JSON(http.StatusInternalServerError, SyncETFPricesResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SyncETFPricesResponse{
		Success: true,
		Count:   len(symbols),
	})
}

// GetETFPriceResponse 获取ETF价格响应
type GetETFPriceResponse struct {
	Success bool                      `json:"success"`
	Data    *models.AShareDividendETF `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// GetETFPrice 获取单个ETF价格
// @Summary 获取ETF实时价格
// @Description 获取指定ETF的实时价格信息
// @Tags a-share-data
// @Produce json
// @Param symbol path string true "ETF代码"
// @Success 200 {object} GetETFPriceResponse
// @Router /api/a-share/price/{symbol} [get]
func (h *AShareDataHandler) GetETFPrice(c *gin.Context) {
	symbol := c.Param("symbol")

	etf, err := h.etfService.GetETFPrice(symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, GetETFPriceResponse{
			Success: false,
			Error:   "ETF未找到: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetETFPriceResponse{
		Success: true,
		Data:    etf,
	})
}

// GetAllETFPricesResponse 获取所有ETF价格响应
type GetAllETFPricesResponse struct {
	Success bool                       `json:"success"`
	Data    []models.AShareDividendETF `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

// GetPortfolioETFPricesResponse 核心ETF价格响应
type GetPortfolioETFPricesResponse struct {
	Success bool                       `json:"success"`
	Data    []models.AShareDividendETF `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

// GetPortfolioETFPrices 获取投资组合核心ETF价格
// @Summary 获取核心ETF价格
// @Description 获取A股投资组合内8个核心ETF的实时价格（白名单过滤，只返回组合内ETF）
// @Tags a-share-data
// @Produce json
// @Success 200 {object} GetPortfolioETFPricesResponse
// @Router /api/a-share/prices [get]
func (h *AShareDataHandler) GetPortfolioETFPrices(c *gin.Context) {
	etfs, err := h.etfService.GetCoreETFPrices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetPortfolioETFPricesResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetPortfolioETFPricesResponse{
		Success: true,
		Data:    etfs,
	})
}

// HistoricalDataRequest 历史数据请求
type HistoricalDataRequest struct {
	StartDate string `json:"start_date"` // 开始日期，格式：2006-01-02
	EndDate   string `json:"end_date"`   // 结束日期，格式：2006-01-02
}

// HistoricalDataResponse 历史数据响应
type HistoricalDataResponse struct {
	Success bool                       `json:"success"`
	Data    []ashare.ETFHistoricalData `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

// GetHistoricalData 获取历史数据
// @Summary 获取ETF历史数据
// @Description 获取指定ETF的历史价格数据
// @Tags a-share-data
// @Accept json
// @Produce json
// @Param symbol path string true "ETF代码"
// @Param request body HistoricalDataRequest true "日期范围"
// @Success 200 {object} HistoricalDataResponse
// @Router /api/a-share/historical/{symbol} [post]
func (h *AShareDataHandler) GetHistoricalData(c *gin.Context) {
	symbol := c.Param("symbol")

	var req HistoricalDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, HistoricalDataResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, HistoricalDataResponse{
			Success: false,
			Error:   "开始日期格式错误",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, HistoricalDataResponse{
			Success: false,
			Error:   "结束日期格式错误",
		})
		return
	}

	data, err := h.etfService.GetHistoricalPrices(symbol, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HistoricalDataResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, HistoricalDataResponse{
		Success: true,
		Data:    data,
	})
}

// SearchETFsResponse 搜索ETF响应
type SearchETFsResponse struct {
	Success bool                       `json:"success"`
	Data    []models.AShareDividendETF `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

// SearchETFs 搜索ETF
// @Summary 搜索ETF
// @Description 根据关键词搜索ETF
// @Tags a-share-data
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} SearchETFsResponse
// @Router /api/a-share/search [get]
func (h *AShareDataHandler) SearchETFs(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, SearchETFsResponse{
			Success: false,
			Error:   "搜索关键词不能为空",
		})
		return
	}

	etfs, err := h.etfService.SearchETFs(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SearchETFsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SearchETFsResponse{
		Success: true,
		Data:    etfs,
	})
}

// GetETFsByFrequencyResponse 按频率获取ETF响应
type GetETFsByFrequencyResponse struct {
	Success bool                       `json:"success"`
	Data    []models.AShareDividendETF `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
}

// GetETFsByFrequency 按分红频率获取ETF
// @Summary 按分红频率获取ETF
// @Description 根据分红频率筛选ETF
// @Tags a-share-data
// @Produce json
// @Param frequency path string true "分红频率: 月分/季分/年分"
// @Success 200 {object} GetETFsByFrequencyResponse
// @Router /api/a-share/by-frequency/{frequency} [get]
func (h *AShareDataHandler) GetETFsByFrequency(c *gin.Context) {
	frequencyStr := c.Param("frequency")

	var frequency models.DividendFrequency
	switch frequencyStr {
	case "月分":
		frequency = models.FrequencyMonthly
	case "季分":
		frequency = models.FrequencyQuarterly
	case "年分":
		frequency = models.FrequencyYearly
	default:
		c.JSON(http.StatusBadRequest, GetETFsByFrequencyResponse{
			Success: false,
			Error:   "无效的分红频率",
		})
		return
	}

	etfs, err := h.etfService.GetETFByFrequency(frequency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetETFsByFrequencyResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetETFsByFrequencyResponse{
		Success: true,
		Data:    etfs,
	})
}

// CalculateDividendYieldResponse 计算股息率响应
type CalculateDividendYieldResponse struct {
	Success       bool    `json:"success"`
	Symbol        string  `json:"symbol"`
	DividendYield float64 `json:"dividend_yield"`
	Error         string  `json:"error,omitempty"`
}

// CalculateDividendYield 计算股息率
// @Summary 计算ETF股息率
// @Description 计算指定ETF的股息率
// @Tags a-share-data
// @Produce json
// @Param symbol path string true "ETF代码"
// @Success 200 {object} CalculateDividendYieldResponse
// @Router /api/a-share/dividend-yield/{symbol} [get]
func (h *AShareDataHandler) CalculateDividendYield(c *gin.Context) {
	symbol := c.Param("symbol")

	yield, err := h.etfService.CalculateDividendYield(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CalculateDividendYieldResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CalculateDividendYieldResponse{
		Success:       true,
		Symbol:        symbol,
		DividendYield: yield.InexactFloat64(),
	})
}

// RefreshAllDataResponse 刷新所有数据响应
type RefreshAllDataResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// RefreshAllData 刷新所有数据
// @Summary 刷新所有A股数据
// @Description 同步ETF列表和价格
// @Tags a-share-data
// @Produce json
// @Success 200 {object} RefreshAllDataResponse
// @Router /api/a-share/refresh-all [post]
func (h *AShareDataHandler) RefreshAllData(c *gin.Context) {
	if err := h.etfService.RefreshAllData(); err != nil {
		c.JSON(http.StatusInternalServerError, RefreshAllDataResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RefreshAllDataResponse{
		Success: true,
		Message: "所有数据已刷新",
	})
}

// DataSourceStatusResponse 数据源状态响应
type DataSourceStatusResponse struct {
	Success        bool   `json:"success"`
	AKShareEnabled bool   `json:"akshare_enabled"`
	TuShareEnabled bool   `json:"tushare_enabled"`
	ETFCount       int    `json:"etf_count"`
	Error          string `json:"error,omitempty"`
}

// GetDataSourceStatus 获取数据源状态
// @Summary 获取数据源状态
// @Description 获取A股数据源的配置状态
// @Tags a-share-data
// @Produce json
// @Success 200 {object} DataSourceStatusResponse
// @Router /api/a-share/data-source-status [get]
func (h *AShareDataHandler) GetDataSourceStatus(c *gin.Context) {
	var count int64
	models.DB.Model(&models.AShareDividendETF{}).Count(&count)

	c.JSON(http.StatusOK, DataSourceStatusResponse{
		Success:        true,
		AKShareEnabled: true, // 简化处理
		TuShareEnabled: false,
		ETFCount:       int(count),
	})
}
