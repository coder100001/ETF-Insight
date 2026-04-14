package docs

// SwaggerSpec 返回 Swagger JSON 规范
func SwaggerSpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "ETF-Insight API",
			"description": "开源专业的 ETF 量化分析平台 API 文档。提供ETF数据查询、投资组合分析、量化指标计算等功能。",
			"version":     "2.4.0",
			"contact": map[string]interface{}{
				"name":  "ETF-Insight Team",
				"url":   "https://github.com/coder100001/ETF-Insight",
				"email": "support@etf-insight.com",
			},
			"license": map[string]interface{}{
				"name": "MIT",
				"url":  "https://opensource.org/licenses/MIT",
			},
		},
		"servers": []map[string]interface{}{
			{
				"url":         "http://localhost:8080",
				"description": "本地开发服务器",
			},
		},
		"paths": getPaths(),
		"components": map[string]interface{}{
			"schemas":         getSchemas(),
			"securitySchemes": getSecuritySchemes(),
		},
	}
}

// getPaths 返回所有 API 路径定义
func getPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/etf/list":                     getETFListPath(),
		"/api/etf/detail/{symbol}":          getETFDetailPath(),
		"/api/etf/history/{symbol}":         getETFHistoryPath(),
		"/api/etf/compare":                  getETFComparePath(),
		"/api/portfolio/analysis":           getPortfolioAnalysisPath(),
		"/api/portfolio/optimize":           getPortfolioOptimizePath(),
		"/api/portfolio/efficient-frontier": getEfficientFrontierPath(),
		"/api/a-share/etfs":                 getAShareETFsPath(),
		"/api/a-share/prices":               getASharePricesPath(),
		"/api/a-share/prices/refresh":       getASharePricesRefreshPath(),
		"/api/a-share/portfolio/dividend":   getAShareDividendPath(),
		"/api/exchange-rates":               getExchangeRatesPath(),
		"/api/exchange-rates/{currency}":    getExchangeRatePath(),
		"/health":                           getHealthPath(),
		"/ready":                            getReadyPath(),
		"/live":                             getLivePath(),
	}
}

// getETFListPath 获取 ETF 列表
func getETFListPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"ETF"},
			"summary":     "获取ETF列表",
			"description": "获取ETF列表，支持分页",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"parameters": []map[string]interface{}{
				{
					"name":        "page",
					"in":          "query",
					"description": "页码",
					"schema":      map[string]interface{}{"type": "integer", "default": 1},
				},
				{
					"name":        "pageSize",
					"in":          "query",
					"description": "每页数量",
					"schema":      map[string]interface{}{"type": "integer", "default": 10, "maximum": 100},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/PaginatedETFListResponse"},
						},
					},
				},
				"401": getUnauthorizedResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getETFDetailPath 获取 ETF 详情
func getETFDetailPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"ETF"},
			"summary":     "获取ETF详情",
			"description": "获取指定ETF的详细信息",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"parameters": []map[string]interface{}{
				{
					"name":        "symbol",
					"in":          "path",
					"required":    true,
					"description": "ETF代码",
					"schema":      map[string]interface{}{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ETFDetailResponse"},
						},
					},
				},
				"400": getBadRequestResponse(),
				"404": getNotFoundResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getETFHistoryPath 获取 ETF 历史数据
func getETFHistoryPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"ETF"},
			"summary":     "获取ETF历史数据",
			"description": "获取ETF历史价格数据",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"parameters": []map[string]interface{}{
				{
					"name":        "symbol",
					"in":          "path",
					"required":    true,
					"description": "ETF代码",
					"schema":      map[string]interface{}{"type": "string"},
				},
				{
					"name":        "start",
					"in":          "query",
					"description": "开始日期 (YYYY-MM-DD)",
					"schema":      map[string]interface{}{"type": "string"},
				},
				{
					"name":        "end",
					"in":          "query",
					"description": "结束日期 (YYYY-MM-DD)",
					"schema":      map[string]interface{}{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ETFHistoryResponse"},
						},
					},
				},
				"400": getBadRequestResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getETFComparePath ETF 对比分析
func getETFComparePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{"ETF"},
			"summary":     "ETF对比分析",
			"description": "对比多个ETF的关键指标",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{"$ref": "#/components/schemas/ETFCompareRequest"},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ETFCompareResponse"},
						},
					},
				},
				"400": getBadRequestResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getPortfolioAnalysisPath 投资组合分析
func getPortfolioAnalysisPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{"Portfolio"},
			"summary":     "投资组合分析",
			"description": "分析投资组合的收益、风险、分红等指标",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{"$ref": "#/components/schemas/PortfolioRequest"},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/PortfolioAnalysisResponse"},
						},
					},
				},
				"400": getBadRequestResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getPortfolioOptimizePath 投资组合优化
func getPortfolioOptimizePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{"Portfolio"},
			"summary":     "投资组合优化",
			"description": "使用马科维茨模型优化投资组合权重",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{"$ref": "#/components/schemas/PortfolioOptimizeRequest"},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/PortfolioOptimizeResponse"},
						},
					},
				},
				"400": getBadRequestResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getEfficientFrontierPath 有效前沿计算
func getEfficientFrontierPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{"Portfolio"},
			"summary":     "有效前沿计算",
			"description": "生成投资组合的有效前沿数据",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{"$ref": "#/components/schemas/EfficientFrontierRequest"},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/EfficientFrontierResponse"},
						},
					},
				},
				"400": getBadRequestResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getAShareETFsPath A股ETF列表
func getAShareETFsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"A-Share"},
			"summary":     "获取A股ETF列表",
			"description": "获取A股红利ETF列表",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/AShareETFListResponse"},
						},
					},
				},
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getASharePricesPath A股ETF价格
func getASharePricesPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"A-Share"},
			"summary":     "获取A股ETF价格",
			"description": "获取A股ETF实时价格",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ASharePriceResponse"},
						},
					},
				},
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getASharePricesRefreshPath 刷新A股ETF价格
func getASharePricesRefreshPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{"A-Share"},
			"summary":     "刷新A股ETF价格",
			"description": "手动刷新A股ETF价格数据",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ASharePriceRefreshResponse"},
						},
					},
				},
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getAShareDividendPath A股分红计算
func getAShareDividendPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{"A-Share"},
			"summary":     "计算分红收益",
			"description": "计算A股ETF组合的分红收益",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{"$ref": "#/components/schemas/ASharePortfolioRequest"},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/AShareDividendCalculationResponse"},
						},
					},
				},
				"400": getBadRequestResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getExchangeRatesPath 汇率列表
func getExchangeRatesPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"Exchange Rate"},
			"summary":     "获取汇率列表",
			"description": "获取所有汇率数据",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ExchangeRateListResponse"},
						},
					},
				},
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getExchangeRatePath 单币种汇率
func getExchangeRatePath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"Exchange Rate"},
			"summary":     "获取单币种汇率",
			"description": "获取指定货币的汇率",
			"security":    []map[string]interface{}{{"BearerAuth": []interface{}{}}},
			"parameters": []map[string]interface{}{
				{
					"name":        "currency",
					"in":          "path",
					"required":    true,
					"description": "货币代码 (如: USD, EUR)",
					"schema":      map[string]interface{}{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "成功",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ExchangeRateResponse"},
						},
					},
				},
				"404": getNotFoundResponse(),
				"500": getInternalErrorResponse(),
			},
		},
	}
}

// getHealthPath 健康检查
func getHealthPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"Health"},
			"summary":     "健康检查",
			"description": "检查服务健康状态",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "服务正常",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/HealthResponse"},
						},
					},
				},
			},
		},
	}
}

// getReadyPath 就绪检查
func getReadyPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"Health"},
			"summary":     "就绪检查",
			"description": "检查服务是否就绪",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "服务就绪",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ReadyResponse"},
						},
					},
				},
				"503": map[string]interface{}{
					"description": "服务未就绪",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/ErrorResponse"},
						},
					},
				},
			},
		},
	}
}

// getLivePath 存活检查
func getLivePath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"Health"},
			"summary":     "存活检查",
			"description": "检查服务是否存活",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "服务存活",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/LiveResponse"},
						},
					},
				},
			},
		},
	}
}

// Helper functions for responses
func getBadRequestResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "无效的请求",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
}

func getUnauthorizedResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "未授权",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
}

func getNotFoundResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "资源未找到",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
}

func getInternalErrorResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "服务器错误",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{"$ref": "#/components/schemas/ErrorResponse"},
			},
		},
	}
}
