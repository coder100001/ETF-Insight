package services

import (
	"testing"
	"time"

	"etf-insight/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPluginTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.PluginRegistry{},
		&models.PluginConfiguration{},
		&models.PluginExecutionLog{},
		&models.ModelBenchmarkMatrix{},
		&models.StrategyExperiment{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func cleanupPluginTestDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func TestPluginService_RegisterPlugin(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-alpha-generator",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test alpha generator plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	err := service.RegisterPlugin(plugin)
	if err != nil {
		t.Errorf("RegisterPlugin failed: %v", err)
	}

	if plugin.ID == 0 {
		t.Error("Plugin ID should not be zero after creation")
	}
}

func TestPluginService_RegisterPlugin_Duplicate(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	duplicate := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Duplicate plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	err := service.RegisterPlugin(duplicate)
	if err == nil {
		t.Error("Expected error for duplicate plugin")
	}

	if err != ErrPluginAlreadyExists {
		t.Errorf("Expected ErrPluginAlreadyExists, got %v", err)
	}
}

func TestPluginService_GetPlugin(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	retrieved, err := service.GetPlugin(plugin.ID)
	if err != nil {
		t.Errorf("GetPlugin failed: %v", err)
	}

	if retrieved.PluginName != plugin.PluginName {
		t.Errorf("Expected PluginName %s, got %s", plugin.PluginName, retrieved.PluginName)
	}
}

func TestPluginService_GetPluginByName(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	retrieved, err := service.GetPluginByName("test-plugin")
	if err != nil {
		t.Errorf("GetPluginByName failed: %v", err)
	}

	if retrieved.ID != plugin.ID {
		t.Errorf("Expected ID %d, got %d", plugin.ID, retrieved.ID)
	}
}

func TestPluginService_ListPlugins(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugins := []*models.PluginRegistry{
		{
			PluginName:   "alpha-gen-1",
			PluginType:   models.PluginTypeAlphaGenerator,
			Version:      "1.0.0",
			Description:  "Alpha generator 1",
			Author:       "Test Author",
			Status:       models.PluginStatusActive,
			RegisteredAt: time.Now(),
		},
		{
			PluginName:   "risk-model-1",
			PluginType:   models.PluginTypeRiskModel,
			Version:      "1.0.0",
			Description:  "Risk model 1",
			Author:       "Test Author",
			Status:       models.PluginStatusActive,
			RegisteredAt: time.Now(),
		},
	}

	for _, p := range plugins {
		_ = service.RegisterPlugin(p)
	}

	all, err := service.ListPlugins("")
	if err != nil {
		t.Errorf("ListPlugins failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(all))
	}

	alphaPlugins, err := service.ListPlugins(models.PluginTypeAlphaGenerator)
	if err != nil {
		t.Errorf("ListPlugins with type filter failed: %v", err)
	}

	if len(alphaPlugins) != 1 {
		t.Errorf("Expected 1 alpha plugin, got %d", len(alphaPlugins))
	}
}

func TestPluginService_DeactivatePlugin(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	err := service.DeactivatePlugin(plugin.ID)
	if err != nil {
		t.Errorf("DeactivatePlugin failed: %v", err)
	}

	retrieved, _ := service.GetPlugin(plugin.ID)
	if retrieved.Status != models.PluginStatusDisabled {
		t.Errorf("Expected status %s, got %s", models.PluginStatusDisabled, retrieved.Status)
	}
}

func TestPluginService_CreateConfiguration(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	config := &models.PluginConfiguration{
		PluginID:   plugin.ID,
		ConfigName: "default-config",
		Parameters: models.JSONMap{"param1": "value1"},
		IsActive:   true,
		IsDefault:  true,
	}

	err := service.CreateConfiguration(config)
	if err != nil {
		t.Errorf("CreateConfiguration failed: %v", err)
	}

	if config.ID == 0 {
		t.Error("Config ID should not be zero after creation")
	}
}

func TestPluginService_ListConfigurations(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	configs := []*models.PluginConfiguration{
		{
			PluginID:   plugin.ID,
			ConfigName: "config-1",
			Parameters: models.JSONMap{"param1": "value1"},
			IsActive:   true,
		},
		{
			PluginID:   plugin.ID,
			ConfigName: "config-2",
			Parameters: models.JSONMap{"param2": "value2"},
			IsActive:   true,
		},
	}

	for _, c := range configs {
		_ = service.CreateConfiguration(c)
	}

	list, err := service.ListConfigurations(plugin.ID)
	if err != nil {
		t.Errorf("ListConfigurations failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 configurations, got %d", len(list))
	}
}

func TestPluginService_ExecutePlugin(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	input := map[string]any{
		"test_param": "test_value",
	}

	log, err := service.ExecutePlugin(plugin.ID, input)
	if err != nil {
		t.Errorf("ExecutePlugin failed: %v", err)
	}

	if log.ID == 0 {
		t.Error("Log ID should not be zero after creation")
	}

	if log.Status != models.ExecutionStatusSuccess {
		t.Errorf("Expected status %s, got %s", models.ExecutionStatusSuccess, log.Status)
	}

	t.Logf("Execution log: PluginID=%d, Duration=%dms, Status=%s", log.PluginID, log.Duration, log.Status)
}

func TestPluginService_GetExecutionLogs(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	for i := range 3 {
		_, _ = service.ExecutePlugin(plugin.ID, map[string]any{"test": i})
	}

	logs, err := service.GetExecutionLogs(plugin.ID, 10)
	if err != nil {
		t.Errorf("GetExecutionLogs failed: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

func TestPluginService_CreateBenchmarkMatrix(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	matrix := &models.ModelBenchmarkMatrix{
		ComparisonName:     "test-comparison",
		AlphaPluginID:      plugin.ID,
		OptimizerPluginID:  plugin.ID,
		RiskPluginID:       plugin.ID,
		BacktestWindow:     252,
		RebalanceFrequency: "monthly",
		ComparisonDate:     time.Now(),
	}

	err := service.CreateBenchmarkMatrix(matrix)
	if err != nil {
		t.Errorf("CreateBenchmarkMatrix failed: %v", err)
	}

	if matrix.ID == 0 {
		t.Error("Matrix ID should not be zero after creation")
	}
}

func TestPluginService_ListBenchmarkMatrices(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	plugin := &models.PluginRegistry{
		PluginName:   "test-plugin",
		PluginType:   models.PluginTypeAlphaGenerator,
		Version:      "1.0.0",
		Description:  "Test plugin",
		Author:       "Test Author",
		Status:       models.PluginStatusActive,
		RegisteredAt: time.Now(),
	}

	_ = service.RegisterPlugin(plugin)

	for i := range 2 {
		matrix := &models.ModelBenchmarkMatrix{
			ComparisonName:     "test-comparison-" + string(rune('A'+i)),
			AlphaPluginID:      plugin.ID,
			OptimizerPluginID:  plugin.ID,
			RiskPluginID:       plugin.ID,
			BacktestWindow:     252,
			RebalanceFrequency: "monthly",
			ComparisonDate:     time.Now(),
		}
		_ = service.CreateBenchmarkMatrix(matrix)
	}

	matrices, err := service.ListBenchmarkMatrices()
	if err != nil {
		t.Errorf("ListBenchmarkMatrices failed: %v", err)
	}

	if len(matrices) != 2 {
		t.Errorf("Expected 2 matrices, got %d", len(matrices))
	}
}

func TestPluginService_CreateExperiment(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	experiment := &models.StrategyExperiment{
		ExperimentName: "test-experiment",
		Description:    "Test experiment description",
		Status:         models.ExperimentStatusRunning,
	}

	err := service.CreateExperiment(experiment)
	if err != nil {
		t.Errorf("CreateExperiment failed: %v", err)
	}

	if experiment.ID == 0 {
		t.Error("Experiment ID should not be zero after creation")
	}
}

func TestPluginService_ListExperiments(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	experiments := []*models.StrategyExperiment{
		{
			ExperimentName: "experiment-1",
			Description:    "Running experiment",
			Status:         models.ExperimentStatusRunning,
		},
		{
			ExperimentName: "experiment-2",
			Description:    "Completed experiment",
			Status:         models.ExperimentStatusCompleted,
		},
	}

	for _, e := range experiments {
		_ = service.CreateExperiment(e)
	}

	all, err := service.ListExperiments("")
	if err != nil {
		t.Errorf("ListExperiments failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("Expected 2 experiments, got %d", len(all))
	}

	running, err := service.ListExperiments(models.ExperimentStatusRunning)
	if err != nil {
		t.Errorf("ListExperiments with status filter failed: %v", err)
	}

	if len(running) != 1 {
		t.Errorf("Expected 1 running experiment, got %d", len(running))
	}
}

func TestPluginService_CompleteExperiment(t *testing.T) {
	db := setupPluginTestDB(t)
	defer cleanupPluginTestDB(db)

	service := NewPluginService(db)

	experiment := &models.StrategyExperiment{
		ExperimentName: "test-experiment",
		Description:    "Test experiment",
		Status:         models.ExperimentStatusRunning,
	}

	_ = service.CreateExperiment(experiment)

	results := `{"total_return": 0.15, "sharpe_ratio": 1.5}`
	err := service.CompleteExperiment(experiment.ID, results)
	if err != nil {
		t.Errorf("CompleteExperiment failed: %v", err)
	}

	retrieved, _ := service.GetExperiment(experiment.ID)
	if retrieved.Status != models.ExperimentStatusCompleted {
		t.Errorf("Expected status %s, got %s", models.ExperimentStatusCompleted, retrieved.Status)
	}
}
