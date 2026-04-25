package docs

// getSchemas 返回所有 Schema 定义
func getSchemas() map[string]interface{} {
	return map[string]interface{}{
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

func getErrorResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean", "example": false},
			"error":   map[string]interface{}{"type": "string", "example": "Invalid request parameters"},
			"code":    map[string]interface{}{"type": "string", "example": "INVALID_REQUEST"},
		},
	}
}

func getPaginatedETFListResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean", "example": true},
			"data": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "object"},
			},
			"pagination": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page":       map[string]interface{}{"type": "integer"},
					"pageSize":   map[string]interface{}{"type": "integer"},
					"total":      map[string]interface{}{"type": "integer"},
					"totalPages": map[string]interface{}{"type": "integer"},
				},
			},
		},
	}
}

func getETFDetailResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol":         map[string]interface{}{"type": "string"},
					"name":           map[string]interface{}{"type": "string"},
					"current_price":  map[string]interface{}{"type": "number"},
					"previous_close": map[string]interface{}{"type": "number"},
					"change":         map[string]interface{}{"type": "number"},
					"change_percent": map[string]interface{}{"type": "number"},
					"volume":         map[string]interface{}{"type": "integer"},
					"market_cap":     map[string]interface{}{"type": "integer"},
					"dividend_yield": map[string]interface{}{"type": "number"},
					"pe_ratio":       map[string]interface{}{"type": "number"},
					"beta":           map[string]interface{}{"type": "number"},
				},
			},
		},
	}
}

func getETFHistoryResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"$ref": "#/components/schemas/ETFData"},
			},
		},
	}
}

func getETFCompareRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []string{"symbols"},
		"properties": map[string]interface{}{
			"symbols": map[string]interface{}{
				"type":    "array",
				"items":   map[string]interface{}{"type": "string"},
				"example": []string{"SPY", "VTI", "VEA"},
			},
			"metrics": map[string]interface{}{
				"type":    "array",
				"items":   map[string]interface{}{"type": "string"},
				"example": []string{"dividend_yield", "expense_ratio", "volatility"},
			},
		},
	}
}

func getETFCompareResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"comparison": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"symbol":         map[string]interface{}{"type": "string"},
								"name":           map[string]interface{}{"type": "string"},
								"dividend_yield": map[string]interface{}{"type": "number"},
								"expense_ratio":  map[string]interface{}{"type": "number"},
								"volatility":     map[string]interface{}{"type": "number"},
							},
						},
					},
					"metrics": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
	}
}

func getPortfolioRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []string{"allocation"},
		"properties": map[string]interface{}{
			"allocation": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": map[string]interface{}{"type": "number"},
				"example":              map[string]interface{}{"SPY": 0.4, "VTI": 0.3, "BND": 0.3},
			},
			"total_investment": map[string]interface{}{
				"type":    "number",
				"default": 10000,
				"example": 10000,
			},
			"tax_rate": map[string]interface{}{
				"type":    "number",
				"default": 0.1,
				"example": 0.1,
			},
		},
	}
}

func getPortfolioAnalysisResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"total_value":                        map[string]interface{}{"type": "number"},
					"total_return":                       map[string]interface{}{"type": "number"},
					"total_return_pct":                   map[string]interface{}{"type": "number"},
					"annual_dividend_before_tax":         map[string]interface{}{"type": "number"},
					"annual_dividend_after_tax":          map[string]interface{}{"type": "number"},
					"dividend_yield":                     map[string]interface{}{"type": "number"},
					"tax_rate":                           map[string]interface{}{"type": "number"},
					"after_tax_return":                   map[string]interface{}{"type": "number"},
					"dividend_tax":                       map[string]interface{}{"type": "number"},
					"total_return_with_dividend":         map[string]interface{}{"type": "number"},
					"total_return_with_dividend_percent": map[string]interface{}{"type": "number"},
					"total_investment":                   map[string]interface{}{"type": "number"},
					"holdings":                           map[string]interface{}{"type": "array"},
				},
			},
		},
	}
}

func getPortfolioOptimizeRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []string{"symbols"},
		"properties": map[string]interface{}{
			"symbols": map[string]interface{}{
				"type":    "array",
				"items":   map[string]interface{}{"type": "string"},
				"example": []string{"SPY", "VTI", "BND"},
			},
			"optimization_type": map[string]interface{}{
				"type":    "string",
				"enum":    []string{"max_sharpe", "min_volatility", "equal_weight"},
				"default": "max_sharpe",
				"example": "max_sharpe",
			},
			"risk_free_rate": map[string]interface{}{
				"type":    "number",
				"default": 0.04,
				"example": 0.04,
			},
			"constraints": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"min_weight": map[string]interface{}{"type": "number", "example": 0.05},
					"max_weight": map[string]interface{}{"type": "number", "example": 0.5},
				},
			},
		},
	}
}

func getPortfolioOptimizeResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"optimization_type": map[string]interface{}{"type": "string"},
					"weights": map[string]interface{}{
						"type":        "object",
						"description": "资产权重配置",
						"example":     map[string]interface{}{"SPY": 0.4, "VTI": 0.3, "BND": 0.3},
					},
					"expected_return": map[string]interface{}{
						"type":        "number",
						"description": "预期收益率",
					},
					"volatility": map[string]interface{}{
						"type":        "number",
						"description": "波动率",
					},
					"sharpe_ratio": map[string]interface{}{"type": "number"},
				},
			},
		},
	}
}

func getEfficientFrontierRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []string{"symbols"},
		"properties": map[string]interface{}{
			"symbols": map[string]interface{}{
				"type":    "array",
				"items":   map[string]interface{}{"type": "string"},
				"example": []string{"SPY", "VTI", "BND"},
			},
			"risk_free_rate": map[string]interface{}{
				"type":    "number",
				"default": 0.04,
				"example": 0.04,
			},
			"points": map[string]interface{}{
				"type":    "integer",
				"default": 20,
				"example": 20,
			},
		},
	}
}

func getEfficientFrontierResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"risk_free_rate": map[string]interface{}{"type": "number"},
					"frontier": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"return":       map[string]interface{}{"type": "number", "description": "预期收益率"},
								"volatility":   map[string]interface{}{"type": "number", "description": "波动率"},
								"sharpe_ratio": map[string]interface{}{"type": "number"},
								"weights": map[string]interface{}{
									"type":        "object",
									"description": "资产权重",
								},
							},
						},
					},
					"max_sharpe_portfolio":     map[string]interface{}{"$ref": "#/components/schemas/PortfolioOptimizeResponse"},
					"min_volatility_portfolio": map[string]interface{}{"$ref": "#/components/schemas/PortfolioOptimizeResponse"},
				},
			},
		},
	}
}

func getAShareETFListResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean", "example": true},
			"data": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"$ref": "#/components/schemas/AShareDividendETF"},
			},
		},
	}
}

func getASharePriceResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"$ref": "#/components/schemas/AShareETFPrice"},
			},
		},
	}
}

func getASharePriceRefreshResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"updated_count": map[string]interface{}{"type": "integer"},
					"updated_at":    map[string]interface{}{"type": "string"},
				},
			},
		},
	}
}

func getASharePortfolioRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []string{"allocation", "total_investment"},
		"properties": map[string]interface{}{
			"allocation": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": map[string]interface{}{"type": "number"},
				"example":              map[string]interface{}{"515080": 0.4, "515180": 0.3, "515300": 0.3},
			},
			"total_investment": map[string]interface{}{
				"type":    "number",
				"example": 100000,
			},
		},
	}
}

func getAShareDividendCalculationResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"portfolio_id":             map[string]interface{}{"type": "integer"},
			"total_investment":         map[string]interface{}{"type": "number", "example": 100000},
			"expected_annual_dividend": map[string]interface{}{"type": "number", "example": 4500},
			"average_dividend_yield":   map[string]interface{}{"type": "number", "example": 4.5},
			"monthly_dividend":         map[string]interface{}{"type": "number", "example": 375},
			"quarterly_dividend":       map[string]interface{}{"type": "number", "example": 1125},
			"holdings": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"$ref": "#/components/schemas/AShareHoldingDetailResponse"},
			},
		},
	}
}

func getAShareHoldingDetailResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"symbol":                map[string]interface{}{"type": "string", "example": "515080"},
			"name":                  map[string]interface{}{"type": "string", "example": "中证红利ETF"},
			"investment":            map[string]interface{}{"type": "number", "example": 40000},
			"weight":                map[string]interface{}{"type": "number", "example": 0.4},
			"dividend_yield":        map[string]interface{}{"type": "number", "example": 4.95},
			"dividend_frequency":    map[string]interface{}{"type": "string", "example": "quarterly"},
			"expected_dividend":     map[string]interface{}{"type": "number", "example": 1800},
			"dividend_contribution": map[string]interface{}{"type": "number", "example": 0.4},
		},
	}
}

func getExchangeRateListResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"$ref": "#/components/schemas/ExchangeRate"},
			},
		},
	}
}

func getExchangeRateResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{"type": "boolean"},
			"data":    map[string]interface{}{"$ref": "#/components/schemas/ExchangeRate"},
		},
	}
}

func getHealthResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status":    map[string]interface{}{"type": "string", "example": "healthy"},
			"version":   map[string]interface{}{"type": "string", "example": "2.6.0"},
			"timestamp": map[string]interface{}{"type": "string", "example": "2026-04-14T10:30:00Z"},
		},
	}
}

func getReadyResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ready":     map[string]interface{}{"type": "boolean", "example": true},
			"database":  map[string]interface{}{"type": "string", "example": "connected"},
			"timestamp": map[string]interface{}{"type": "string"},
		},
	}
}

func getLiveResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status":    map[string]interface{}{"type": "string", "example": "alive"},
			"timestamp": map[string]interface{}{"type": "string"},
		},
	}
}

func getETFDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"symbol":      map[string]interface{}{"type": "string"},
			"date":        map[string]interface{}{"type": "string", "example": "2026-04-14"},
			"open_price":  map[string]interface{}{"type": "number"},
			"high_price":  map[string]interface{}{"type": "number"},
			"low_price":   map[string]interface{}{"type": "number"},
			"close_price": map[string]interface{}{"type": "number"},
			"volume":      map[string]interface{}{"type": "integer"},
			"dividend":    map[string]interface{}{"type": "number"},
			"split_ratio": map[string]interface{}{"type": "number"},
		},
	}
}

func getAShareDividendETFSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":                 map[string]interface{}{"type": "integer"},
			"symbol":             map[string]interface{}{"type": "string", "example": "515080"},
			"name":               map[string]interface{}{"type": "string", "example": "中证红利ETF"},
			"dividend_yield_min": map[string]interface{}{"type": "number", "example": 4.8},
			"dividend_yield_max": map[string]interface{}{"type": "number", "example": 5.1},
			"dividend_frequency": map[string]interface{}{"type": "string", "example": "quarterly"},
			"benchmark":          map[string]interface{}{"type": "string", "example": "中证红利指数"},
			"exchange":           map[string]interface{}{"type": "string", "example": "SSE"},
			"management_fee":     map[string]interface{}{"type": "number", "example": 0.005},
			"description":        map[string]interface{}{"type": "string"},
			"status":             map[string]interface{}{"type": "integer"},
			"created_at":         map[string]interface{}{"type": "string"},
			"updated_at":         map[string]interface{}{"type": "string"},
		},
	}
}

func getAShareETFPriceSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":             map[string]interface{}{"type": "integer"},
			"etf_id":         map[string]interface{}{"type": "integer"},
			"symbol":         map[string]interface{}{"type": "string", "description": "ETF代码"},
			"current_price":  map[string]interface{}{"type": "number", "description": "当前价格"},
			"previous_close": map[string]interface{}{"type": "number", "description": "昨收价"},
			"change":         map[string]interface{}{"type": "number", "description": "涨跌额"},
			"change_percent": map[string]interface{}{"type": "number", "description": "涨跌幅百分比"},
			"volume":         map[string]interface{}{"type": "integer", "description": "成交量"},
			"created_at":     map[string]interface{}{"type": "string"},
			"updated_at":     map[string]interface{}{"type": "string"},
		},
	}
}

func getExchangeRateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":            map[string]interface{}{"type": "integer"},
			"currency_code": map[string]interface{}{"type": "string", "example": "USD"},
			"rate":          map[string]interface{}{"type": "number", "example": 7.25},
			"created_at":    map[string]interface{}{"type": "string"},
			"updated_at":    map[string]interface{}{"type": "string"},
		},
	}
}
