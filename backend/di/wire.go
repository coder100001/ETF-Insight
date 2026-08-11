//go:build wireinject
// +build wireinject

package di

import (
	"etf-insight/config"
	"etf-insight/router"
	"etf-insight/services"
	"etf-insight/services/datasource"
	"etf-insight/tasks"

	"github.com/google/wire"
)

// ServiceSet 核心服务依赖集合
var ServiceSet = wire.NewSet(
	services.NewETFAnalysisService,
	services.NewPortfolioOptimizer,
	ProvideExchangeRateConfig,
	ProvideExchangeRateService,
	ProvideDataSourceProvider,
)

// TaskSet 定时任务依赖集合
var TaskSet = wire.NewSet(
	ProvideScheduleConfig,
	tasks.NewScheduler,
	tasks.NewExchangeRateTask,
)

// RouterSet 路由依赖集合
var RouterSet = wire.NewSet(
	router.NewRouter,
)

// Container 包含所有已构造的依赖项
type Container struct {
	Config           *config.Config
	Router           *router.Router
	Scheduler        *tasks.Scheduler
	ExchangeRateTask *tasks.ExchangeRateTask
	Provider         datasource.DataSourceProvider
	AnalysisService  *services.ETFAnalysisService
	Optimizer        *services.PortfolioOptimizer
}

// InitializeContainer 使用 Wire 自动注入所有依赖
// cfg 由 Bootstrap 阶段加载并传入
func InitializeContainer(cfg *config.Config) (*Container, error) {
	wire.Build(
		ServiceSet,
		TaskSet,
		RouterSet,
		wire.Struct(new(Container), "*"),
	)
	return nil, nil
}
