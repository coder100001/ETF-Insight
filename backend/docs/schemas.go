package docs

// getSchemas 返回所有 Schema 定义
func getSchemas() map[string]any {
	return map[string]any{
		"ErrorResponse":                     getErrorResponseSchema(),
		"PaginatedETFListResponse":          getPaginatedETFListResponseSchema(),
		"ETFDetailResponse":                 getETFDetailResponseSchema(),
		"ETFHistoryResponse":                getETFHistoryResponseSchema(),
		"ETFCompareRequest":                 getETFCompareRequestSchema(),
		"ETFCompareResponse":                getETFCompareResponseSchema(),
		"PortfolioRequest":                  getPortfolioRequestSchema(),
		"PortfolioAnalysisResponse":         getPortfolioAnalysisResponseSchema(),
		"PortfolioOptimizeRequest":          getPortfolioOptimizeRequestSchema(),
		"PortfolioOptimizeResponse":         getPortfolioOptimizeResponseSchema(),
		"EfficientFrontierRequest":          getEfficientFrontierRequestSchema(),
		"EfficientFrontierResponse":         getEfficientFrontierResponseSchema(),
		"AShareETFListResponse":             getAShareETFListResponseSchema(),
		"ASharePriceResponse":               getASharePriceResponseSchema(),
		"ASharePriceRefreshResponse":        getASharePriceRefreshResponseSchema(),
		"ASharePortfolioRequest":            getASharePortfolioRequestSchema(),
		"AShareDividendCalculationResponse": getAShareDividendCalculationResponseSchema(),
		"AShareHoldingDetailResponse":       getAShareHoldingDetailResponseSchema(),
		"ExchangeRateListResponse":          getExchangeRateListResponseSchema(),
		"ExchangeRateResponse":              getExchangeRateResponseSchema(),
		"HealthResponse":                    getHealthResponseSchema(),
		"ReadyResponse":                     getReadyResponseSchema(),
		"LiveResponse":                      getLiveResponseSchema(),
		"ETFData":                           getETFDataSchema(),
		"AShareDividendETF":                 getAShareDividendETFSchema(),
		"AShareETFPrice":                    getAShareETFPriceSchema(),
		"ExchangeRate":                      getExchangeRateSchema(),
	}
}

func getErrorResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean", "example": false},
			"error":   map[string]any{"type": "string", "example": "Invalid request parameters"},
			"code":    map[string]any{"type": "string", "example": "INVALID_REQUEST"},
		},
	}
}

func getPaginatedETFListResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean", "example": true},
			"data": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "object"},
			},
			"pagination": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"page":       map[string]any{"type": "integer"},
					"pageSize":   map[string]any{"type": "integer"},
					"total":      map[string]any{"type": "integer"},
					"totalPages": map[string]any{"type": "integer"},
				},
			},
		},
	}
}

func getETFDetailResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"symbol":         map[string]any{"type": "string"},
					"name":           map[string]any{"type": "string"},
					"current_price":  map[string]any{"type": "number"},
					"previous_close": map[string]any{"type": "number"},
					"change":         map[string]any{"type": "number"},
					"change_percent": map[string]any{"type": "number"},
					"volume":         map[string]any{"type": "integer"},
					"market_cap":     map[string]any{"type": "integer"},
					"dividend_yield": map[string]any{"type": "number"},
					"pe_ratio":       map[string]any{"type": "number"},
					"beta":           map[string]any{"type": "number"},
				},
			},
		},
	}
}

func getETFHistoryResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/components/schemas/ETFData"},
			},
		},
	}
}

func getETFCompareRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"symbols"},
		"properties": map[string]any{
			"symbols": map[string]any{
				"type":    "array",
				"items":   map[string]any{"type": "string"},
				"example": []string{"SPY", "VTI", "VEA"},
			},
			"metrics": map[string]any{
				"type":    "array",
				"items":   map[string]any{"type": "string"},
				"example": []string{"dividend_yield", "expense_ratio", "volatility"},
			},
		},
	}
}

func getETFCompareResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"comparison": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"symbol":         map[string]any{"type": "string"},
								"name":           map[string]any{"type": "string"},
								"dividend_yield": map[string]any{"type": "number"},
								"expense_ratio":  map[string]any{"type": "number"},
								"volatility":     map[string]any{"type": "number"},
							},
						},
					},
					"metrics": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
			},
		},
	}
}

func getPortfolioRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"allocation"},
		"properties": map[string]any{
			"allocation": map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "number"},
				"example":              map[string]any{"SPY": 0.4, "VTI": 0.3, "BND": 0.3},
			},
			"total_investment": map[string]any{
				"type":    "number",
				"default": 10000,
				"example": 10000,
			},
			"tax_rate": map[string]any{
				"type":    "number",
				"default": 0.1,
				"example": 0.1,
			},
		},
	}
}

func getPortfolioAnalysisResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"total_value":                        map[string]any{"type": "number"},
					"total_return":                       map[string]any{"type": "number"},
					"total_return_pct":                   map[string]any{"type": "number"},
					"annual_dividend_before_tax":         map[string]any{"type": "number"},
					"annual_dividend_after_tax":          map[string]any{"type": "number"},
					"dividend_yield":                     map[string]any{"type": "number"},
					"tax_rate":                           map[string]any{"type": "number"},
					"after_tax_return":                   map[string]any{"type": "number"},
					"dividend_tax":                       map[string]any{"type": "number"},
					"total_return_with_dividend":         map[string]any{"type": "number"},
					"total_return_with_dividend_percent": map[string]any{"type": "number"},
					"total_investment":                   map[string]any{"type": "number"},
					"holdings":                           map[string]any{"type": "array"},
				},
			},
		},
	}
}

func getPortfolioOptimizeRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"symbols"},
		"properties": map[string]any{
			"symbols": map[string]any{
				"type":    "array",
				"items":   map[string]any{"type": "string"},
				"example": []string{"SPY", "VTI", "BND"},
			},
			"optimization_type": map[string]any{
				"type":    "string",
				"enum":    []string{"max_sharpe", "min_volatility", "equal_weight"},
				"default": "max_sharpe",
				"example": "max_sharpe",
			},
			"risk_free_rate": map[string]any{
				"type":    "number",
				"default": 0.04,
				"example": 0.04,
			},
			"constraints": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"min_weight": map[string]any{"type": "number", "example": 0.05},
					"max_weight": map[string]any{"type": "number", "example": 0.5},
				},
			},
		},
	}
}

func getPortfolioOptimizeResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"optimization_type": map[string]any{"type": "string"},
					"weights": map[string]any{
						"type":        "object",
						"description": "资产权重配置",
						"example":     map[string]any{"SPY": 0.4, "VTI": 0.3, "BND": 0.3},
					},
					"expected_return": map[string]any{
						"type":        "number",
						"description": "预期收益率",
					},
					"volatility": map[string]any{
						"type":        "number",
						"description": "波动率",
					},
					"sharpe_ratio": map[string]any{"type": "number"},
				},
			},
		},
	}
}

func getEfficientFrontierRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"symbols"},
		"properties": map[string]any{
			"symbols": map[string]any{
				"type":    "array",
				"items":   map[string]any{"type": "string"},
				"example": []string{"SPY", "VTI", "BND"},
			},
			"risk_free_rate": map[string]any{
				"type":    "number",
				"default": 0.04,
				"example": 0.04,
			},
			"points": map[string]any{
				"type":    "integer",
				"default": 20,
				"example": 20,
			},
		},
	}
}

func getEfficientFrontierResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"risk_free_rate": map[string]any{"type": "number"},
					"frontier": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"return":       map[string]any{"type": "number", "description": "预期收益率"},
								"volatility":   map[string]any{"type": "number", "description": "波动率"},
								"sharpe_ratio": map[string]any{"type": "number"},
								"weights": map[string]any{
									"type":        "object",
									"description": "资产权重",
								},
							},
						},
					},
					"max_sharpe_portfolio":     map[string]any{"$ref": "#/components/schemas/PortfolioOptimizeResponse"},
					"min_volatility_portfolio": map[string]any{"$ref": "#/components/schemas/PortfolioOptimizeResponse"},
				},
			},
		},
	}
}

func getAShareETFListResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean", "example": true},
			"data": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/components/schemas/AShareDividendETF"},
			},
		},
	}
}

func getASharePriceResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/components/schemas/AShareETFPrice"},
			},
		},
	}
}

func getASharePriceRefreshResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"updated_count": map[string]any{"type": "integer"},
					"updated_at":    map[string]any{"type": "string"},
				},
			},
		},
	}
}

func getASharePortfolioRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"allocation", "total_investment"},
		"properties": map[string]any{
			"allocation": map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "number"},
				"example":              map[string]any{"515080": 0.4, "515180": 0.3, "515300": 0.3},
			},
			"total_investment": map[string]any{
				"type":    "number",
				"example": 100000,
			},
		},
	}
}

func getAShareDividendCalculationResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"portfolio_id":             map[string]any{"type": "integer"},
			"total_investment":         map[string]any{"type": "number", "example": 100000},
			"expected_annual_dividend": map[string]any{"type": "number", "example": 4500},
			"average_dividend_yield":   map[string]any{"type": "number", "example": 4.5},
			"monthly_dividend":         map[string]any{"type": "number", "example": 375},
			"quarterly_dividend":       map[string]any{"type": "number", "example": 1125},
			"holdings": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/components/schemas/AShareHoldingDetailResponse"},
			},
		},
	}
}

func getAShareHoldingDetailResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol":                map[string]any{"type": "string", "example": "515080"},
			"name":                  map[string]any{"type": "string", "example": "中证红利ETF"},
			"investment":            map[string]any{"type": "number", "example": 40000},
			"weight":                map[string]any{"type": "number", "example": 0.4},
			"dividend_yield":        map[string]any{"type": "number", "example": 4.95},
			"dividend_frequency":    map[string]any{"type": "string", "example": "quarterly"},
			"expected_dividend":     map[string]any{"type": "number", "example": 1800},
			"dividend_contribution": map[string]any{"type": "number", "example": 0.4},
		},
	}
}

func getExchangeRateListResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/components/schemas/ExchangeRate"},
			},
		},
	}
}

func getExchangeRateResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"success": map[string]any{"type": "boolean"},
			"data":    map[string]any{"$ref": "#/components/schemas/ExchangeRate"},
		},
	}
}

func getHealthResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status":    map[string]any{"type": "string", "example": "healthy"},
			"version":   map[string]any{"type": "string", "example": "2.6.0"},
			"timestamp": map[string]any{"type": "string", "example": "2026-04-14T10:30:00Z"},
		},
	}
}

func getReadyResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ready":     map[string]any{"type": "boolean", "example": true},
			"database":  map[string]any{"type": "string", "example": "connected"},
			"timestamp": map[string]any{"type": "string"},
		},
	}
}

func getLiveResponseSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status":    map[string]any{"type": "string", "example": "alive"},
			"timestamp": map[string]any{"type": "string"},
		},
	}
}

func getETFDataSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol":      map[string]any{"type": "string"},
			"date":        map[string]any{"type": "string", "example": "2026-04-14"},
			"open_price":  map[string]any{"type": "number"},
			"high_price":  map[string]any{"type": "number"},
			"low_price":   map[string]any{"type": "number"},
			"close_price": map[string]any{"type": "number"},
			"volume":      map[string]any{"type": "integer"},
			"dividend":    map[string]any{"type": "number"},
			"split_ratio": map[string]any{"type": "number"},
		},
	}
}

func getAShareDividendETFSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":                 map[string]any{"type": "integer"},
			"symbol":             map[string]any{"type": "string", "example": "515080"},
			"name":               map[string]any{"type": "string", "example": "中证红利ETF"},
			"dividend_yield_min": map[string]any{"type": "number", "example": 4.8},
			"dividend_yield_max": map[string]any{"type": "number", "example": 5.1},
			"dividend_frequency": map[string]any{"type": "string", "example": "quarterly"},
			"benchmark":          map[string]any{"type": "string", "example": "中证红利指数"},
			"exchange":           map[string]any{"type": "string", "example": "SSE"},
			"management_fee":     map[string]any{"type": "number", "example": 0.005},
			"description":        map[string]any{"type": "string"},
			"status":             map[string]any{"type": "integer"},
			"created_at":         map[string]any{"type": "string"},
			"updated_at":         map[string]any{"type": "string"},
		},
	}
}

func getAShareETFPriceSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":             map[string]any{"type": "integer"},
			"etf_id":         map[string]any{"type": "integer"},
			"symbol":         map[string]any{"type": "string", "description": "ETF代码"},
			"current_price":  map[string]any{"type": "number", "description": "当前价格"},
			"previous_close": map[string]any{"type": "number", "description": "昨收价"},
			"change":         map[string]any{"type": "number", "description": "涨跌额"},
			"change_percent": map[string]any{"type": "number", "description": "涨跌幅百分比"},
			"volume":         map[string]any{"type": "integer", "description": "成交量"},
			"created_at":     map[string]any{"type": "string"},
			"updated_at":     map[string]any{"type": "string"},
		},
	}
}

func getExchangeRateSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":            map[string]any{"type": "integer"},
			"currency_code": map[string]any{"type": "string", "example": "USD"},
			"rate":          map[string]any{"type": "number", "example": 7.25},
			"created_at":    map[string]any{"type": "string"},
			"updated_at":    map[string]any{"type": "string"},
		},
	}
}
