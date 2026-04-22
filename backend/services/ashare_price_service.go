package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

// ASharePriceService A股ETF价格获取服务
type ASharePriceService struct {
	client    *http.Client
	baseURL   string
	providers []PriceProvider
}

// PriceProvider 价格数据提供者接口
type PriceProvider interface {
	GetName() string
	GetPrice(symbol string) (*PriceData, error)
	IsAvailable() bool
}

// PriceData 价格数据
type PriceData struct {
	Symbol         string
	CurrentPrice   float64
	PreviousClose  float64
	PriceChange    float64
	PriceChangePct float64
	Volume         int64
	Turnover       float64
	NAV            float64
	PremiumRate    float64
	UpdateTime     time.Time
}

// NewASharePriceService 创建A股价格服务
func NewASharePriceService() *ASharePriceService {
	return &ASharePriceService{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "",
		providers: []PriceProvider{
			NewSinaProvider(),      // 新浪财经（主数据源）
			NewEastMoneyProvider(), // 东方财富（备份1）
			NewTencentProvider(),   // 腾讯财经（备份2）
			NewNetEaseProvider(),   // 网易财经（备份3）
		},
	}
}

// UpdateAllETFPrices 更新所有A股ETF价格
func (s *ASharePriceService) UpdateAllETFPrices() error {
	var etfs []models.AShareDividendETF
	if err := models.DB.Where("status = ?", 1).Find(&etfs).Error; err != nil {
		return fmt.Errorf("获取ETF列表失败: %w", err)
	}

	utils.Info("开始更新A股ETF价格", "count", len(etfs))

	for _, etf := range etfs {
		if err := s.UpdateETFPrice(&etf); err != nil {
			utils.Error("更新ETF价格失败", err, "symbol", etf.Symbol)
			continue
		}
		time.Sleep(100 * time.Millisecond) // 避免请求过快
	}

	utils.Info("A股ETF价格更新完成")
	return nil
}

// UpdateETFPrice 更新单个ETF价格
func (s *ASharePriceService) UpdateETFPrice(etf *models.AShareDividendETF) error {
	var priceData *PriceData
	var lastErr error
	var successProvider string

	// 尝试所有数据源（故障转移）
	for _, provider := range s.providers {
		if !provider.IsAvailable() {
			utils.Info("数据源不可用，跳过", "provider", provider.GetName())
			continue
		}

		data, err := provider.GetPrice(etf.Symbol)
		if err == nil && data != nil && data.CurrentPrice > 0 {
			priceData = data
			successProvider = provider.GetName()
			utils.Info("获取价格成功",
				"symbol", etf.Symbol,
				"provider", provider.GetName(),
				"price", data.CurrentPrice)
			break
		}
		utils.Warn("数据源获取失败，尝试下一个",
			"symbol", etf.Symbol,
			"provider", provider.GetName(),
			"error", err)
		lastErr = err
	}

	if priceData == nil {
		return fmt.Errorf("所有数据源均不可用: %w", lastErr)
	}

	// 更新ETF数据
	etf.CurrentPrice = decimal.NewFromFloat(priceData.CurrentPrice)
	etf.PreviousClose = decimal.NewFromFloat(priceData.PreviousClose)
	etf.PriceChange = decimal.NewFromFloat(priceData.PriceChange)
	etf.PriceChangePct = decimal.NewFromFloat(priceData.PriceChangePct)
	etf.Volume = priceData.Volume
	etf.Turnover = decimal.NewFromFloat(priceData.Turnover)
	etf.NAV = decimal.NewFromFloat(priceData.NAV)
	etf.PremiumRate = decimal.NewFromFloat(priceData.PremiumRate)
	etf.PriceUpdatedAt = priceData.UpdateTime

	if err := models.DB.Save(etf).Error; err != nil {
		return fmt.Errorf("保存价格数据失败: %w", err)
	}

	utils.Info("价格更新成功",
		"symbol", etf.Symbol,
		"provider", successProvider,
		"price", priceData.CurrentPrice)

	return nil
}

// GetProviderStatus 获取所有数据源状态
func (s *ASharePriceService) GetProviderStatus() []ProviderStatus {
	statuses := make([]ProviderStatus, 0, len(s.providers))

	// 使用测试代码检查数据源可用性
	testSymbol := "515080" // 中证红利ETF

	for _, provider := range s.providers {
		status := ProviderStatus{
			Name:  provider.GetName(),
			Order: len(statuses) + 1,
		}

		// 尝试获取测试数据
		data, err := provider.GetPrice(testSymbol)
		if err == nil && data != nil && data.CurrentPrice > 0 {
			status.Available = true
			status.LastPrice = data.CurrentPrice
			status.LastCheck = time.Now()
		} else {
			status.Available = false
			status.Error = fmt.Sprintf("获取测试数据失败: %v", err)
			status.LastCheck = time.Now()
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// ProviderStatus 数据源状态
type ProviderStatus struct {
	Name      string    `json:"name"`
	Available bool      `json:"available"`
	Order     int       `json:"order"`
	LastPrice float64   `json:"last_price,omitempty"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
}

// ==================== 新浪财经数据源 ====================

type SinaProvider struct {
	client *http.Client
}

func NewSinaProvider() *SinaProvider {
	return &SinaProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *SinaProvider) GetName() string {
	return "sina"
}

func (p *SinaProvider) IsAvailable() bool {
	return true
}

// GetPrice 从新浪财经获取ETF价格
func (p *SinaProvider) GetPrice(symbol string) (*PriceData, error) {
	// 转换代码格式：515080 -> sh515080 或 sz159545
	code := p.convertSymbol(symbol)
	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", code)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return p.parseSinaData(string(body), symbol)
}

func (p *SinaProvider) convertSymbol(symbol string) string {
	// 上交所代码以5、6开头，深交所代码以0、1、3开头
	if strings.HasPrefix(symbol, "5") || strings.HasPrefix(symbol, "6") {
		return "sh" + symbol
	}
	return "sz" + symbol
}

func (p *SinaProvider) parseSinaData(data, symbol string) (*PriceData, error) {
	// 新浪返回格式：var hq_str_sh515080="名称,今日开盘价,昨日收盘价,当前价,今日最高价,今日最低价,竞买价,竞卖价,成交量,成交额,买一量,买一价,买二量,买二价...";
	prefix := fmt.Sprintf("var hq_str_%s=\"", p.convertSymbol(symbol))
	if !strings.Contains(data, prefix) {
		return nil, fmt.Errorf("无效的数据格式")
	}

	data = strings.TrimPrefix(data, prefix)
	data = strings.TrimSuffix(data, "\";")
	fields := strings.Split(data, ",")

	if len(fields) < 10 {
		return nil, fmt.Errorf("数据字段不足")
	}

	// 解析数据
	currentPrice := parseFloat(fields[2])  // 当前价
	previousClose := parseFloat(fields[1]) // 昨日收盘价
	volume := parseInt(fields[8])          // 成交量
	turnover := parseFloat(fields[9])      // 成交额

	priceChange := currentPrice - previousClose
	priceChangePct := 0.0
	if previousClose > 0 {
		priceChangePct = (priceChange / previousClose) * 100
	}

	return &PriceData{
		Symbol:         symbol,
		CurrentPrice:   currentPrice,
		PreviousClose:  previousClose,
		PriceChange:    priceChange,
		PriceChangePct: priceChangePct,
		Volume:         volume,
		Turnover:       turnover,
		NAV:            0, // 新浪接口不直接提供净值
		PremiumRate:    0,
		UpdateTime:     time.Now(),
	}, nil
}

// ==================== 东方财富数据源 ====================

type EastMoneyProvider struct {
	client *http.Client
}

func NewEastMoneyProvider() *EastMoneyProvider {
	return &EastMoneyProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *EastMoneyProvider) GetName() string {
	return "eastmoney"
}

func (p *EastMoneyProvider) IsAvailable() bool {
	return true
}

// GetPrice 从东方财富获取ETF价格
func (p *EastMoneyProvider) GetPrice(symbol string) (*PriceData, error) {
	// 东方财富API
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=f43,f44,f45,f46,f47,f48,f57,f58,f60,f170",
		p.convertSymbol(symbol))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			CurrentPrice   float64 `json:"f43"`  // 最新价
			PriceChangePct float64 `json:"f170"` // 涨跌幅
			Volume         int64   `json:"f47"`  // 成交量
			Turnover       float64 `json:"f48"`  // 成交额
			PreviousClose  float64 `json:"f60"`  // 昨收
			Open           float64 `json:"f46"`  // 开盘价
			High           float64 `json:"f44"`  // 最高价
			Low            float64 `json:"f45"`  // 最低价
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Data.CurrentPrice == 0 {
		return nil, fmt.Errorf("获取价格失败")
	}

	// 东方财富 API 返回的价格和涨跌幅都是放大 100 倍的整数，需要除以 100
	currentPrice := result.Data.CurrentPrice / 100
	previousClose := result.Data.PreviousClose / 100
	priceChangePct := result.Data.PriceChangePct / 100
	priceChange := currentPrice - previousClose

	return &PriceData{
		Symbol:         symbol,
		CurrentPrice:   currentPrice,
		PreviousClose:  previousClose,
		PriceChange:    priceChange,
		PriceChangePct: priceChangePct,
		Volume:         result.Data.Volume,
		Turnover:       result.Data.Turnover,
		NAV:            0,
		PremiumRate:    0,
		UpdateTime:     time.Now(),
	}, nil
}

func (p *EastMoneyProvider) convertSymbol(symbol string) string {
	// 东方财富格式：0.515080 或 1.159545
	if strings.HasPrefix(symbol, "5") || strings.HasPrefix(symbol, "6") {
		return "1." + symbol // 上交所
	}
	return "0." + symbol // 深交所
}

// ==================== 工具函数 ====================

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseInt(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}

// ==================== 腾讯财经数据源 ====================

type TencentProvider struct {
	client *http.Client
}

func NewTencentProvider() *TencentProvider {
	return &TencentProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *TencentProvider) GetName() string {
	return "tencent"
}

func (p *TencentProvider) IsAvailable() bool {
	return true
}

// GetPrice 从腾讯财经获取ETF价格
func (p *TencentProvider) GetPrice(symbol string) (*PriceData, error) {
	// 腾讯财经API格式
	code := p.convertSymbol(symbol)
	url := fmt.Sprintf("https://web.sqt.gtimg.cn/q=%s", code)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://gu.qq.com/")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return p.parseTencentData(string(body), symbol)
}

func (p *TencentProvider) convertSymbol(symbol string) string {
	// 腾讯格式：sh515080 或 sz159545
	if strings.HasPrefix(symbol, "5") || strings.HasPrefix(symbol, "6") {
		return "sh" + symbol
	}
	return "sz" + symbol
}

func (p *TencentProvider) parseTencentData(data, symbol string) (*PriceData, error) {
	// 腾讯返回格式：v_sh515080="1~名称~代码~当前价~昨收~今开~成交量~成交额~..."
	prefix := fmt.Sprintf("v_%s=\"", p.convertSymbol(symbol))
	if !strings.Contains(data, prefix) {
		return nil, fmt.Errorf("无效的数据格式")
	}

	data = strings.TrimPrefix(data, prefix)
	data = strings.TrimSuffix(data, "\";")
	fields := strings.Split(data, "~")

	if len(fields) < 10 {
		return nil, fmt.Errorf("数据字段不足")
	}

	currentPrice := parseFloat(fields[3])  // 当前价
	previousClose := parseFloat(fields[4]) // 昨收
	volume := parseInt(fields[6])          // 成交量
	turnover := parseFloat(fields[7])      // 成交额

	priceChange := currentPrice - previousClose
	priceChangePct := 0.0
	if previousClose > 0 {
		priceChangePct = (priceChange / previousClose) * 100
	}

	return &PriceData{
		Symbol:         symbol,
		CurrentPrice:   currentPrice,
		PreviousClose:  previousClose,
		PriceChange:    priceChange,
		PriceChangePct: priceChangePct,
		Volume:         volume,
		Turnover:       turnover,
		NAV:            0,
		PremiumRate:    0,
		UpdateTime:     time.Now(),
	}, nil
}

// ==================== 网易财经数据源（东方财富备用接口）====================

type NetEaseProvider struct {
	client *http.Client
}

func NewNetEaseProvider() *NetEaseProvider {
	return &NetEaseProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *NetEaseProvider) GetName() string {
	return "eastmoney-quote"
}

func (p *NetEaseProvider) IsAvailable() bool {
	return true
}

// GetPrice 从东方财富行情接口获取ETF价格（备用接口）
func (p *NetEaseProvider) GetPrice(symbol string) (*PriceData, error) {
	// 东方财富行情中心备用接口
	code := p.convertSymbol(symbol)
	url := fmt.Sprintf("https://qt.gtimg.cn/q=%s", code)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://gu.qq.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return p.parseData(string(body), symbol)
}

func (p *NetEaseProvider) convertSymbol(symbol string) string {
	if strings.HasPrefix(symbol, "5") || strings.HasPrefix(symbol, "6") {
		return "sh" + symbol
	}
	return "sz" + symbol
}

func (p *NetEaseProvider) parseData(data, symbol string) (*PriceData, error) {
	// 腾讯 qt.gtimg.cn 格式与 web.sqt.gtimg.cn 相同
	prefix := fmt.Sprintf("v_%s=\"", p.convertSymbol(symbol))
	if !strings.Contains(data, prefix) {
		return nil, fmt.Errorf("无效的数据格式")
	}

	data = strings.TrimPrefix(data, prefix)
	data = strings.TrimSuffix(data, "\";")
	fields := strings.Split(data, "~")

	if len(fields) < 10 {
		return nil, fmt.Errorf("数据字段不足")
	}

	currentPrice := parseFloat(fields[3])
	previousClose := parseFloat(fields[4])
	volume := parseInt(fields[6])
	turnover := parseFloat(fields[7])

	priceChange := currentPrice - previousClose
	priceChangePct := 0.0
	if previousClose > 0 {
		priceChangePct = (priceChange / previousClose) * 100
	}

	return &PriceData{
		Symbol:         symbol,
		CurrentPrice:   currentPrice,
		PreviousClose:  previousClose,
		PriceChange:    priceChange,
		PriceChangePct: priceChangePct,
		Volume:         volume,
		Turnover:       turnover,
		UpdateTime:     time.Now(),
	}, nil
}
