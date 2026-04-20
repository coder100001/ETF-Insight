package backtest

import (
	"time"

	"github.com/shopspring/decimal"
)

// Bar K线数据
type Bar struct {
	Symbol   string          `json:"symbol"`
	Date     time.Time       `json:"date"`
	Open     decimal.Decimal `json:"open"`
	High     decimal.Decimal `json:"high"`
	Low      decimal.Decimal `json:"low"`
	Close    decimal.Decimal `json:"close"`
	Volume   int64           `json:"volume"`
	Dividend decimal.Decimal `json:"dividend"` // 每股股息
	Split    decimal.Decimal `json:"split"`    // 拆股比例
}

// Signal 交易信号
type Signal struct {
	Type     SignalType      `json:"type"`
	Symbol   string          `json:"symbol"`
	Quantity decimal.Decimal `json:"quantity"`
	Price    decimal.Decimal `json:"price"` // 限价单价格，0表示市价单
	Reason   string          `json:"reason"`
}

// SignalType 信号类型
type SignalType string

const (
	SignalBuy  SignalType = "BUY"
	SignalSell SignalType = "SELL"
)

// Strategy 策略接口
type Strategy interface {
	// GenerateSignals 生成交易信号
	GenerateSignals(engine *BacktestEngine, bar *Bar) []*Signal

	// GetName 获取策略名称
	GetName() string

	// GetDescription 获取策略描述
	GetDescription() string
}

// FactorStrategy 因子策略接口
type FactorStrategy interface {
	Strategy

	// CalculateFactors 计算因子值
	CalculateFactors(engine *BacktestEngine, bar *Bar) map[string]decimal.Decimal

	// GetFactorNames 获取因子名称列表
	GetFactorNames() []string
}

// DataProvider 数据提供者接口
type DataProvider interface {
	// GetData 获取历史数据
	GetData(startDate, endDate time.Time) ([]*Bar, error)

	// GetSymbols 获取可用标的列表
	GetSymbols() []string
}

// SlippageModel 滑点模型接口
type SlippageModel interface {
	// Calculate 计算滑点
	// price: 当前价格
	// quantity: 交易数量
	// isBuy: 是否为买入
	Calculate(price, quantity decimal.Decimal, isBuy bool) decimal.Decimal
}

// CommissionModel 手续费模型接口
type CommissionModel interface {
	// Calculate 计算手续费
	// price: 成交价格
	// quantity: 交易数量
	Calculate(price, quantity decimal.Decimal) decimal.Decimal
}

// DividendModel 股息模型接口
type DividendModel interface {
	// Calculate 计算股息收入
	// dividendPerShare: 每股股息
	// quantity: 持仓数量
	Calculate(dividendPerShare, quantity decimal.Decimal) decimal.Decimal
}

// DefaultSlippageModel 默认滑点模型
type DefaultSlippageModel struct {
	SlippageRate decimal.Decimal // 滑点率
}

func (m *DefaultSlippageModel) Calculate(price, quantity decimal.Decimal, isBuy bool) decimal.Decimal {
	return price.Mul(m.SlippageRate)
}

// DefaultCommissionModel 默认手续费模型
type DefaultCommissionModel struct {
	CommissionRate decimal.Decimal // 手续费率
	MinCommission  decimal.Decimal // 最低手续费
}

func (m *DefaultCommissionModel) Calculate(price, quantity decimal.Decimal) decimal.Decimal {
	commission := price.Mul(quantity).Mul(m.CommissionRate)
	if m.MinCommission.GreaterThan(decimal.Zero) && commission.LessThan(m.MinCommission) {
		return m.MinCommission
	}
	return commission
}

// DefaultDividendModel 默认股息模型
type DefaultDividendModel struct {
	TaxRate decimal.Decimal // 股息税率
}

func (m *DefaultDividendModel) Calculate(dividendPerShare, quantity decimal.Decimal) decimal.Decimal {
	grossDividend := dividendPerShare.Mul(quantity)
	if m.TaxRate.GreaterThan(decimal.Zero) {
		return grossDividend.Mul(decimal.NewFromInt(1).Sub(m.TaxRate))
	}
	return grossDividend
}

// VolumeSlippageModel 基于成交量的滑点模型
type VolumeSlippageModel struct {
	BaseSlippageRate decimal.Decimal // 基础滑点率
	VolumeImpact     decimal.Decimal // 成交量冲击系数
	AvgVolume        int64           // 平均成交量
}

func (m *VolumeSlippageModel) Calculate(price, quantity decimal.Decimal, isBuy bool) decimal.Decimal {
	// 基础滑点
	baseSlippage := price.Mul(m.BaseSlippageRate)

	// 成交量冲击滑点
	// 假设交易量占平均成交量的比例越大，滑点越大
	volumeRatio := quantity.Div(decimal.NewFromInt(m.AvgVolume))
	volumeImpact := price.Mul(m.VolumeImpact).Mul(volumeRatio)

	return baseSlippage.Add(volumeImpact)
}

// TieredCommissionModel 阶梯式手续费模型
type TieredCommissionModel struct {
	Tiers []CommissionTier
}

// CommissionTier 手续费阶梯
type CommissionTier struct {
	MinValue decimal.Decimal // 最小交易金额
	MaxValue decimal.Decimal // 最大交易金额
	Rate     decimal.Decimal // 手续费率
}

func (m *TieredCommissionModel) Calculate(price, quantity decimal.Decimal) decimal.Decimal {
	tradeValue := price.Mul(quantity)

	for _, tier := range m.Tiers {
		if tradeValue.GreaterThanOrEqual(tier.MinValue) &&
			(tier.MaxValue.IsZero() || tradeValue.LessThan(tier.MaxValue)) {
			return tradeValue.Mul(tier.Rate)
		}
	}

	// 默认返回交易金额的0.1%
	return tradeValue.Mul(decimal.NewFromFloat(0.001))
}

// DividendReinvestmentModel 股息再投资模型
type DividendReinvestmentModel struct {
	ReinvestEnabled bool            // 是否启用再投资
	TaxRate         decimal.Decimal // 股息税率
	ReinvestDelay   int             // 再投资延迟天数
}

func (m *DividendReinvestmentModel) Calculate(dividendPerShare, quantity decimal.Decimal) decimal.Decimal {
	grossDividend := dividendPerShare.Mul(quantity)
	if m.TaxRate.GreaterThan(decimal.Zero) {
		return grossDividend.Mul(decimal.NewFromInt(1).Sub(m.TaxRate))
	}
	return grossDividend
}

// IsReinvestEnabled 是否启用再投资
func (m *DividendReinvestmentModel) IsReinvestEnabled() bool {
	return m.ReinvestEnabled
}

// GetReinvestDelay 获取再投资延迟天数
func (m *DividendReinvestmentModel) GetReinvestDelay() int {
	return m.ReinvestDelay
}
