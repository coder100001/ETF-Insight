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
			NewSinaProvider(),      // 新浪财经
			NewEastMoneyProvider(), // 东方财富
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

	// 尝试所有数据源
	for _, provider := range s.providers {
		if !provider.IsAvailable() {
			continue
		}

		data, err := provider.GetPrice(etf.Symbol)
		if err == nil && data != nil {
			priceData = data
			utils.Info("获取价格成功",
				"symbol", etf.Symbol,
				"provider", provider.GetName(),
				"price", data.CurrentPrice)
			break
		}
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

	return nil
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

	resp, err := p.client.Get(url)
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

	priceChange := result.Data.CurrentPrice - result.Data.PreviousClose

	return &PriceData{
		Symbol:         symbol,
		CurrentPrice:   result.Data.CurrentPrice,
		PreviousClose:  result.Data.PreviousClose,
		PriceChange:    priceChange,
		PriceChangePct: result.Data.PriceChangePct,
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
