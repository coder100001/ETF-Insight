package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/models"
	"etf-insight/services/etf"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// UniversalETFHandler 通用ETF处理器
type UniversalETFHandler struct {
	service *etf.UniversalETFService
}

// NewUniversalETFHandler 创建通用ETF处理器
func NewUniversalETFHandler() *UniversalETFHandler {
	return &UniversalETFHandler{
		service: etf.NewUniversalETFService(models.DB),
	}
}

// InitializeETFsResponse 初始化ETF响应
type InitializeETFsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// InitializeDefaultETFs 初始化默认ETF数据
// @Summary 初始化默认ETF数据
// @Description 加载预定义的跨资产类别ETF数据
// @Tags universal-etf
// @Produce json
// @Success 200 {object} InitializeETFsResponse
// @Router /api/universal-etf/initialize [post]
func (h *UniversalETFHandler) InitializeDefaultETFs(c *gin.Context) {
	if err := h.service.InitializeDefaultETFs(); err != nil {
		c.JSON(http.StatusInternalServerError, InitializeETFsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, InitializeETFsResponse{
		Success: true,
		Message: "ETF数据初始化成功",
	})
}

// GetETFBySymbolResponse 获取ETF响应
type GetETFBySymbolResponse struct {
	Success bool                 `json:"success"`
	Data    *models.UniversalETF `json:"data,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// GetETFBySymbol 根据代码获取ETF
// @Summary 获取ETF详情
// @Description 根据代码获取ETF详细信息
// @Tags universal-etf
// @Produce json
// @Param symbol path string true "ETF代码"
// @Success 200 {object} GetETFBySymbolResponse
// @Router /api/universal-etf/{symbol} [get]
func (h *UniversalETFHandler) GetETFBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")

	etf, err := h.service.GetETFBySymbol(symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, GetETFBySymbolResponse{
			Success: false,
			Error:   "ETF未找到",
		})
		return
	}

	c.JSON(http.StatusOK, GetETFBySymbolResponse{
		Success: true,
		Data:    etf,
	})
}

// GetAllETFsResponse 获取所有ETF响应
type GetAllETFsResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// GetAllETFs 获取所有ETF
// @Summary 获取所有ETF
// @Description 获取所有可用的ETF列表
// @Tags universal-etf
// @Produce json
// @Success 200 {object} GetAllETFsResponse
// @Router /api/universal-etf [get]
func (h *UniversalETFHandler) GetAllETFs(c *gin.Context) {
	etfs, err := h.service.GetAllETFs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetAllETFsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetAllETFsResponse{
		Success: true,
		Data:    etfs,
		Count:   len(etfs),
	})
}

// GetETFsByAssetClassResponse 按资产类别获取ETF响应
type GetETFsByAssetClassResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// GetETFsByAssetClass 根据资产类别获取ETF
// @Summary 按资产类别获取ETF
// @Description 根据资产类别筛选ETF
// @Tags universal-etf
// @Produce json
// @Param asset_class path string true "资产类别: equity/bond/commodity/reit/currency/multi_asset/alternative"
// @Success 200 {object} GetETFsByAssetClassResponse
// @Router /api/universal-etf/asset-class/{asset_class} [get]
func (h *UniversalETFHandler) GetETFsByAssetClass(c *gin.Context) {
	assetClassStr := c.Param("asset_class")
	assetClass := models.AssetClass(assetClassStr)

	etfs, err := h.service.GetETFsByAssetClass(assetClass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetETFsByAssetClassResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetETFsByAssetClassResponse{
		Success: true,
		Data:    etfs,
		Count:   len(etfs),
	})
}

// GetETFsByRegionResponse 按地区获取ETF响应
type GetETFsByRegionResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// GetETFsByRegion 根据地区获取ETF
// @Summary 按地区获取ETF
// @Description 根据地区筛选ETF
// @Tags universal-etf
// @Produce json
// @Param region path string true "地区: global/us/china/europe/japan/emerging/asia_pacific/latin_america"
// @Success 200 {object} GetETFsByRegionResponse
// @Router /api/universal-etf/region/{region} [get]
func (h *UniversalETFHandler) GetETFsByRegion(c *gin.Context) {
	regionStr := c.Param("region")
	region := models.Region(regionStr)

	etfs, err := h.service.GetETFsByRegion(region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetETFsByRegionResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetETFsByRegionResponse{
		Success: true,
		Data:    etfs,
		Count:   len(etfs),
	})
}

// GetETFsByTypeResponse 按类型获取ETF响应
type GetETFsByTypeResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// GetETFsByType 根据类型获取ETF
// @Summary 按类型获取ETF
// @Description 根据ETF类型筛选
// @Tags universal-etf
// @Produce json
// @Param etf_type path string true "ETF类型: index/sector/factor/thematic/active/leveraged/inverse"
// @Success 200 {object} GetETFsByTypeResponse
// @Router /api/universal-etf/type/{etf_type} [get]
func (h *UniversalETFHandler) GetETFsByType(c *gin.Context) {
	etfTypeStr := c.Param("etf_type")
	etfType := models.ETFType(etfTypeStr)

	etfs, err := h.service.GetETFsByType(etfType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetETFsByTypeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetETFsByTypeResponse{
		Success: true,
		Data:    etfs,
		Count:   len(etfs),
	})
}

// SearchUniversalETFsResponse 搜索ETF响应
type SearchUniversalETFsResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// SearchETFs 搜索ETF
// @Summary 搜索ETF
// @Description 根据关键词搜索ETF
// @Tags universal-etf
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} SearchUniversalETFsResponse
// @Router /api/universal-etf/search [get]
func (h *UniversalETFHandler) SearchETFs(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, SearchUniversalETFsResponse{
			Success: false,
			Error:   "搜索关键词不能为空",
		})
		return
	}

	etfs, err := h.service.SearchETFs(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SearchUniversalETFsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SearchUniversalETFsResponse{
		Success: true,
		Data:    etfs,
		Count:   len(etfs),
	})
}

// FilterETFsRequest ETF筛选请求
type FilterETFsRequest struct {
	AssetClass      string          `json:"asset_class"`
	Region          string          `json:"region"`
	ETFType         string          `json:"etf_type"`
	Sector          string          `json:"sector"`
	Provider        string          `json:"provider"`
	Currency        string          `json:"currency"`
	MinExpenseRatio decimal.Decimal `json:"min_expense_ratio"`
	MaxExpenseRatio decimal.Decimal `json:"max_expense_ratio"`
}

// FilterETFsResponse ETF筛选响应
type FilterETFsResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// FilterETFs 多条件筛选ETF
// @Summary 多条件筛选ETF
// @Description 根据多个条件筛选ETF
// @Tags universal-etf
// @Accept json
// @Produce json
// @Param request body FilterETFsRequest true "筛选条件"
// @Success 200 {object} FilterETFsResponse
// @Router /api/universal-etf/filter [post]
func (h *UniversalETFHandler) FilterETFs(c *gin.Context) {
	var req FilterETFsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, FilterETFsResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	filter := etf.ETFFilter{
		AssetClass:      models.AssetClass(req.AssetClass),
		Region:          models.Region(req.Region),
		ETFType:         models.ETFType(req.ETFType),
		Sector:          req.Sector,
		Provider:        req.Provider,
		Currency:        req.Currency,
		MinExpenseRatio: req.MinExpenseRatio,
		MaxExpenseRatio: req.MaxExpenseRatio,
	}

	etfs, err := h.service.GetETFsByFilter(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, FilterETFsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, FilterETFsResponse{
		Success: true,
		Data:    etfs,
		Count:   len(etfs),
	})
}

// GetAssetClassDistributionResponse 资产类别分布响应
type GetAssetClassDistributionResponse struct {
	Success      bool           `json:"success"`
	Distribution map[string]int `json:"distribution,omitempty"`
	Error        string         `json:"error,omitempty"`
}

// GetAssetClassDistribution 获取资产类别分布
// @Summary 资产类别分布
// @Description 获取ETF资产类别的分布统计
// @Tags universal-etf
// @Produce json
// @Success 200 {object} GetAssetClassDistributionResponse
// @Router /api/universal-etf/distribution/asset-class [get]
func (h *UniversalETFHandler) GetAssetClassDistribution(c *gin.Context) {
	distribution, err := h.service.GetAssetClassDistribution()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetAssetClassDistributionResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetAssetClassDistributionResponse{
		Success:      true,
		Distribution: distribution,
	})
}

// GetRegionDistributionResponse 地区分布响应
type GetRegionDistributionResponse struct {
	Success      bool           `json:"success"`
	Distribution map[string]int `json:"distribution,omitempty"`
	Error        string         `json:"error,omitempty"`
}

// GetRegionDistribution 获取地区分布
// @Summary 地区分布
// @Description 获取ETF地区的分布统计
// @Tags universal-etf
// @Produce json
// @Success 200 {object} GetRegionDistributionResponse
// @Router /api/universal-etf/distribution/region [get]
func (h *UniversalETFHandler) GetRegionDistribution(c *gin.Context) {
	distribution, err := h.service.GetRegionDistribution()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetRegionDistributionResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetRegionDistributionResponse{
		Success:      true,
		Distribution: distribution,
	})
}

// CompareETFsRequest ETF对比请求
type CompareETFsRequest struct {
	Symbols []string `json:"symbols" binding:"required"`
}

// CompareETFsResponse ETF对比响应
type CompareETFsResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// CompareETFs 对比ETF
// @Summary 对比ETF
// @Description 对比多个ETF的详细信息
// @Tags universal-etf
// @Accept json
// @Produce json
// @Param request body CompareETFsRequest true "ETF代码列表"
// @Success 200 {object} CompareETFsResponse
// @Router /api/universal-etf/compare [post]
func (h *UniversalETFHandler) CompareETFs(c *gin.Context) {
	var req CompareETFsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CompareETFsResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	if len(req.Symbols) < 2 {
		c.JSON(http.StatusBadRequest, CompareETFsResponse{
			Success: false,
			Error:   "至少需要选择2个ETF进行对比",
		})
		return
	}

	etfs, err := h.service.CompareETFs(req.Symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CompareETFsResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CompareETFsResponse{
		Success: true,
		Data:    etfs,
		Count:   len(etfs),
	})
}

// GetPortfolioAllocationResponse 组合配置响应
type GetPortfolioAllocationResponse struct {
	Success    bool               `json:"success"`
	Strategy   string             `json:"strategy"`
	Allocation map[string]float64 `json:"allocation,omitempty"`
	Error      string             `json:"error,omitempty"`
}

// GetPortfolioAllocation 获取组合配置建议
// @Summary 组合配置建议
// @Description 获取基于不同策略的ETF配置建议
// @Tags universal-etf
// @Produce json
// @Param strategy query string true "策略: conservative/balanced/aggressive/dividend"
// @Success 200 {object} GetPortfolioAllocationResponse
// @Router /api/universal-etf/portfolio-allocation [get]
func (h *UniversalETFHandler) GetPortfolioAllocation(c *gin.Context) {
	strategy := c.Query("strategy")
	if strategy == "" {
		strategy = "balanced"
	}

	allocation, err := h.service.GetPortfolioAllocation(strategy)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetPortfolioAllocationResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetPortfolioAllocationResponse{
		Success:    true,
		Strategy:   strategy,
		Allocation: allocation,
	})
}

// GetCategoriesResponse 获取分类列表响应
type GetCategoriesResponse struct {
	Success      bool     `json:"success"`
	AssetClasses []string `json:"asset_classes,omitempty"`
	Regions      []string `json:"regions,omitempty"`
	ETFTypes     []string `json:"etf_types,omitempty"`
	Sectors      []string `json:"sectors,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// GetCategories 获取分类列表
// @Summary 获取分类列表
// @Description 获取所有可用的ETF分类选项
// @Tags universal-etf
// @Produce json
// @Success 200 {object} GetCategoriesResponse
// @Router /api/universal-etf/categories [get]
func (h *UniversalETFHandler) GetCategories(c *gin.Context) {
	// 资产类别
	assetClasses := []string{
		string(models.AssetClassEquity),
		string(models.AssetClassBond),
		string(models.AssetClassCommodity),
		string(models.AssetClassREIT),
		string(models.AssetClassCurrency),
		string(models.AssetClassMultiAsset),
		string(models.AssetClassAlternative),
	}

	// 地区
	regions := []string{
		string(models.RegionGlobal),
		string(models.RegionUS),
		string(models.RegionChina),
		string(models.RegionEurope),
		string(models.RegionJapan),
		string(models.RegionEmerging),
		string(models.RegionAsiaPacific),
		string(models.RegionLatinAmerica),
	}

	// ETF类型
	etfTypes := []string{
		string(models.ETFTypeIndex),
		string(models.ETFTypeSector),
		string(models.ETFTypeFactor),
		string(models.ETFTypeThematic),
		string(models.ETFTypeActive),
		string(models.ETFTypeLeveraged),
		string(models.ETFTypeInverse),
	}

	// 行业板块
	sectors := []string{
		"Technology",
		"Healthcare",
		"Financials",
		"Consumer",
		"Energy",
		"Industrials",
		"Materials",
		"Utilities",
		"Real Estate",
		"Communication",
	}

	c.JSON(http.StatusOK, GetCategoriesResponse{
		Success:      true,
		AssetClasses: assetClasses,
		Regions:      regions,
		ETFTypes:     etfTypes,
		Sectors:      sectors,
	})
}

// GetTopPerformersResponse 获取表现最佳ETF响应
type GetTopPerformersResponse struct {
	Success bool                  `json:"success"`
	Data    []models.UniversalETF `json:"data,omitempty"`
	Period  string                `json:"period"`
	Count   int                   `json:"count"`
	Error   string                `json:"error,omitempty"`
}

// GetTopPerformers 获取表现最佳ETF
// @Summary 表现最佳ETF
// @Description 获取指定时间段内表现最佳的ETF
// @Tags universal-etf
// @Produce json
// @Param period query string true "时间段: 1m/3m/6m/1y/ytd"
// @Param limit query int false "返回数量，默认10"
// @Success 200 {object} GetTopPerformersResponse
// @Router /api/universal-etf/top-performers [get]
func (h *UniversalETFHandler) GetTopPerformers(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		period = "1y"
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10
	}

	// 获取所有ETF
	etfs, err := h.service.GetAllETFs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetTopPerformersResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// 根据时间段排序
	var sortedETFs []models.UniversalETF
	switch period {
	case "1m":
		sortedETFs = sortETFsByReturn(etfs, "1m")
	case "3m":
		sortedETFs = sortETFsByReturn(etfs, "3m")
	case "6m":
		sortedETFs = sortETFsByReturn(etfs, "6m")
	case "1y":
		sortedETFs = sortETFsByReturn(etfs, "1y")
	case "ytd":
		sortedETFs = sortETFsByReturn(etfs, "ytd")
	default:
		sortedETFs = etfs
	}

	// 限制返回数量
	if len(sortedETFs) > limit {
		sortedETFs = sortedETFs[:limit]
	}

	c.JSON(http.StatusOK, GetTopPerformersResponse{
		Success: true,
		Data:    sortedETFs,
		Period:  period,
		Count:   len(sortedETFs),
	})
}

// sortETFsByReturn 根据收益率排序ETF
func sortETFsByReturn(etfs []models.UniversalETF, period string) []models.UniversalETF {
	// 这里简化处理，实际应该根据数据库查询排序
	// 返回原始列表（实际实现需要按收益率排序）
	return etfs
}
