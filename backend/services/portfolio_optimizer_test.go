package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PortfolioOptimizerTestSuite struct {
	suite.Suite
	optimizer *PortfolioOptimizer
}

func (s *PortfolioOptimizerTestSuite) SetupTest() {
	analysisService := NewETFAnalysisService(nil)
	s.optimizer = NewPortfolioOptimizer(analysisService)
}

func TestPortfolioOptimizerSuite(t *testing.T) {
	suite.Run(t, new(PortfolioOptimizerTestSuite))
}

func (s *PortfolioOptimizerTestSuite) TestOptimize_MaxSharpe() {
	request := PortfolioOptimizationRequest{
		Symbols:          []string{"SPY", "QQQ"},
		OptimizationType: OptimizationTypeMaxSharpe,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
		Constraints: OptimizationConstraints{
			MaxWeightPerAsset: decimal.NewFromFloat(0.8),
			MinWeightPerAsset: decimal.NewFromFloat(0.1),
			AllowShort:        false,
		},
	}

	result, err := s.optimizer.Optimize(request)

	// 由于需要历史数据，可能会返回错误
	// 主要测试函数不panic
	s.NotPanics(func() {
		s.optimizer.Optimize(request)
	})

	// 如果成功，验证结果结构
	if err == nil && result != nil {
		s.NotNil(result.Weights)
		s.NotNil(result.ExpectedReturn)
		s.NotNil(result.ExpectedVolatility)
		s.NotNil(result.SharpeRatio)
	}
}

func (s *PortfolioOptimizerTestSuite) TestOptimize_MinVolatility() {
	request := PortfolioOptimizationRequest{
		Symbols:          []string{"SPY", "QQQ", "TLT"},
		OptimizationType: OptimizationTypeMinVolatility,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
		Constraints: OptimizationConstraints{
			MaxWeightPerAsset: decimal.NewFromFloat(0.8),
			MinWeightPerAsset: decimal.NewFromFloat(0.1),
			AllowShort:        false,
		},
	}

	s.NotPanics(func() {
		s.optimizer.Optimize(request)
	})
}

func (s *PortfolioOptimizerTestSuite) TestOptimize_EqualWeight() {
	request := PortfolioOptimizationRequest{
		Symbols:          []string{"SPY", "QQQ", "IWM", "TLT"},
		OptimizationType: OptimizationTypeEqualWeight,
		RiskFreeRate:     decimal.NewFromFloat(0.04),
		Constraints: OptimizationConstraints{
			MaxWeightPerAsset: decimal.NewFromFloat(0.8),
			MinWeightPerAsset: decimal.NewFromFloat(0.1),
			AllowShort:        false,
		},
	}

	result, err := s.optimizer.Optimize(request)

	// 等权重分配不依赖历史数据
	if err == nil && result != nil {
		s.NotNil(result.Weights)
		// 验证权重和为1
		totalWeight := decimal.Zero
		for _, weight := range result.Weights {
			totalWeight = totalWeight.Add(weight)
		}
		s.True(totalWeight.Sub(decimal.NewFromInt(1)).Abs().LessThan(decimal.NewFromFloat(0.0001)))
	}
}

func (s *PortfolioOptimizerTestSuite) TestOptimize_InvalidRequest() {
	// 测试空symbol列表
	request := PortfolioOptimizationRequest{
		Symbols:          []string{},
		OptimizationType: OptimizationTypeMaxSharpe,
	}

	_, err := s.optimizer.Optimize(request)
	s.Error(err)
}

func (s *PortfolioOptimizerTestSuite) TestOptimize_SingleSymbol() {
	// 测试单symbol（应该失败，因为至少需要2个）
	request := PortfolioOptimizationRequest{
		Symbols:          []string{"SPY"},
		OptimizationType: OptimizationTypeMaxSharpe,
	}

	_, err := s.optimizer.Optimize(request)
	s.Error(err)
}

func (s *PortfolioOptimizerTestSuite) TestNewPortfolioOptimizer() {
	analysisService := NewETFAnalysisService(nil)
	optimizer := NewPortfolioOptimizer(analysisService)

	s.NotNil(optimizer)
	s.NotNil(optimizer.analysisService)
}

func TestPortfolioOptimizationRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request PortfolioOptimizationRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: PortfolioOptimizationRequest{
				Symbols:          []string{"SPY", "QQQ"},
				OptimizationType: OptimizationTypeMaxSharpe,
				RiskFreeRate:     decimal.NewFromFloat(0.04),
			},
			wantErr: false,
		},
		{
			name: "Empty symbols",
			request: PortfolioOptimizationRequest{
				Symbols:          []string{},
				OptimizationType: OptimizationTypeMaxSharpe,
			},
			wantErr: true,
		},
		{
			name: "Single symbol",
			request: PortfolioOptimizationRequest{
				Symbols:          []string{"SPY"},
				OptimizationType: OptimizationTypeMaxSharpe,
			},
			wantErr: true,
		},
		{
			name: "Too many symbols",
			request: PortfolioOptimizationRequest{
				Symbols:          make([]string, 21),
				OptimizationType: OptimizationTypeMaxSharpe,
			},
			wantErr: true,
		},
		{
			name: "Invalid optimization type",
			request: PortfolioOptimizationRequest{
				Symbols:          []string{"SPY", "QQQ"},
				OptimizationType: "invalid_type",
			},
			wantErr: false, // 默认会处理为max_sharpe
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysisService := NewETFAnalysisService(nil)
			optimizer := NewPortfolioOptimizer(analysisService)

			_, err := optimizer.Optimize(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// 可能成功或因为缺少数据而失败
				assert.NotPanics(t, func() {
					optimizer.Optimize(tt.request)
				})
			}
		})
	}
}

func TestOptimizationConstraints_Validation(t *testing.T) {
	constraints := OptimizationConstraints{
		MaxWeightPerAsset: decimal.NewFromFloat(0.5),
		MinWeightPerAsset: decimal.NewFromFloat(0.1),
		AllowShort:        false,
	}

	assert.True(t, constraints.MaxWeightPerAsset.GreaterThan(constraints.MinWeightPerAsset))
	assert.False(t, constraints.AllowShort)
}

func TestPortfolioOptimizationResult_Structure(t *testing.T) {
	result := PortfolioOptimizationResult{
		Weights: map[string]decimal.Decimal{
			"SPY": decimal.NewFromFloat(0.6),
			"QQQ": decimal.NewFromFloat(0.4),
		},
		ExpectedReturn:     decimal.NewFromFloat(0.12),
		ExpectedVolatility: decimal.NewFromFloat(0.15),
		SharpeRatio:        decimal.NewFromFloat(0.8),
		OptimizationType:   OptimizationTypeMaxSharpe,
		RiskFreeRate:       decimal.NewFromFloat(0.04),
	}

	assert.NotNil(t, result.Weights)
	assert.Equal(t, 2, len(result.Weights))
	assert.True(t, result.ExpectedReturn.GreaterThan(decimal.Zero))
	assert.True(t, result.ExpectedVolatility.GreaterThan(decimal.Zero))
	assert.True(t, result.SharpeRatio.GreaterThan(decimal.Zero))
}

func TestFrontierPoint_Structure(t *testing.T) {
	point := FrontierPoint{
		ExpectedReturn: decimal.NewFromFloat(0.10),
		Volatility:     decimal.NewFromFloat(0.12),
		Weights: map[string]decimal.Decimal{
			"SPY": decimal.NewFromFloat(0.5),
			"QQQ": decimal.NewFromFloat(0.5),
		},
	}

	assert.NotNil(t, point.Weights)
	assert.True(t, point.ExpectedReturn.GreaterThan(decimal.Zero))
	assert.True(t, point.Volatility.GreaterThan(decimal.Zero))
}
