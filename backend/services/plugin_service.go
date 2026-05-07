package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"etf-insight/models"

	"gorm.io/gorm"
)

var (
	ErrPluginNotFound      = errors.New("plugin not found")
	ErrPluginAlreadyExists = errors.New("plugin already exists")
	ErrInvalidPluginConfig = errors.New("invalid plugin configuration")
	ErrPluginNotActive     = errors.New("plugin is not active")
)

type PluginService struct {
	db *gorm.DB
}

func NewPluginService(db *gorm.DB) *PluginService {
	return &PluginService{db: db}
}

func (s *PluginService) RegisterPlugin(plugin *models.PluginRegistry) error {
	var existing models.PluginRegistry
	err := s.db.Where("plugin_name = ? AND version = ?", plugin.PluginName, plugin.Version).First(&existing).Error
	if err == nil {
		return ErrPluginAlreadyExists
	}

	if err := s.validatePlugin(plugin); err != nil {
		return err
	}

	plugin.Status = models.PluginStatusActive
	plugin.CreatedAt = time.Now()
	plugin.UpdatedAt = time.Now()

	return s.db.Create(plugin).Error
}

func (s *PluginService) GetPlugin(id uint) (*models.PluginRegistry, error) {
	var plugin models.PluginRegistry
	err := s.db.First(&plugin, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPluginNotFound
		}
		return nil, err
	}
	return &plugin, nil
}

func (s *PluginService) GetPluginByName(name string) (*models.PluginRegistry, error) {
	var plugin models.PluginRegistry
	err := s.db.Where("plugin_name = ?", name).First(&plugin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPluginNotFound
		}
		return nil, err
	}
	return &plugin, nil
}

func (s *PluginService) ListPlugins(pluginType models.PluginType) ([]models.PluginRegistry, error) {
	var plugins []models.PluginRegistry
	query := s.db.Model(&models.PluginRegistry{})

	if pluginType != "" {
		query = query.Where("plugin_type = ?", pluginType)
	}

	err := query.Order("created_at DESC").Find(&plugins).Error
	return plugins, err
}

func (s *PluginService) UpdatePlugin(plugin *models.PluginRegistry) error {
	plugin.UpdatedAt = time.Now()
	return s.db.Save(plugin).Error
}

func (s *PluginService) DeactivatePlugin(id uint) error {
	return s.db.Model(&models.PluginRegistry{}).
		Where("id = ?", id).
		Update("status", models.PluginStatusDisabled).Error
}

func (s *PluginService) ActivatePlugin(id uint) error {
	return s.db.Model(&models.PluginRegistry{}).
		Where("id = ?", id).
		Update("status", models.PluginStatusActive).Error
}

func (s *PluginService) validatePlugin(plugin *models.PluginRegistry) error {
	if plugin.PluginName == "" {
		return errors.New("plugin name is required")
	}

	if plugin.Version == "" {
		return errors.New("plugin version is required")
	}

	if !plugin.PluginType.IsValid() {
		return errors.New("invalid plugin type")
	}

	if plugin.InputSchema != "" {
		var schema map[string]any
		if err := json.Unmarshal([]byte(plugin.InputSchema), &schema); err != nil {
			return errors.New("invalid input schema JSON")
		}
	}

	if plugin.OutputSchema != "" {
		var schema map[string]any
		if err := json.Unmarshal([]byte(plugin.OutputSchema), &schema); err != nil {
			return errors.New("invalid output schema JSON")
		}
	}

	return nil
}

func (s *PluginService) CreateConfiguration(config *models.PluginConfiguration) error {
	plugin, err := s.GetPlugin(config.PluginID)
	if err != nil {
		return err
	}

	if plugin.Status != models.PluginStatusActive {
		return ErrPluginNotActive
	}

	if err := s.validateConfiguration(config, plugin); err != nil {
		return err
	}

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	return s.db.Create(config).Error
}

func (s *PluginService) GetConfiguration(id uint) (*models.PluginConfiguration, error) {
	var config models.PluginConfiguration
	err := s.db.First(&config, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("configuration not found")
		}
		return nil, err
	}
	return &config, nil
}

func (s *PluginService) ListConfigurations(pluginID uint) ([]models.PluginConfiguration, error) {
	var configs []models.PluginConfiguration
	err := s.db.Where("plugin_id = ?", pluginID).
		Order("created_at DESC").
		Find(&configs).Error
	return configs, err
}

func (s *PluginService) UpdateConfiguration(config *models.PluginConfiguration) error {
	config.UpdatedAt = time.Now()
	return s.db.Save(config).Error
}

func (s *PluginService) DeleteConfiguration(id uint) error {
	return s.db.Delete(&models.PluginConfiguration{}, id).Error
}

func (s *PluginService) validateConfiguration(config *models.PluginConfiguration, plugin *models.PluginRegistry) error {
	if config.Parameters != "" {
		var params map[string]any
		if err := json.Unmarshal([]byte(config.Parameters), &params); err != nil {
			return errors.New("invalid parameters JSON")
		}
	}

	return nil
}

func (s *PluginService) ExecutePlugin(pluginID uint, input any) (*models.PluginExecutionLog, error) {
	plugin, err := s.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	if plugin.Status != models.PluginStatusActive {
		return nil, ErrPluginNotActive
	}

	startTime := time.Now()
	inputJSON, _ := json.Marshal(input)

	executionID := fmt.Sprintf("exec-%d-%d", pluginID, startTime.UnixNano())

	log := &models.PluginExecutionLog{
		PluginID:    pluginID,
		ExecutionID: executionID,
		InputData:   string(inputJSON),
		Status:      models.ExecutionStatusSuccess,
		StartTime:   startTime,
		CreatedAt:   startTime,
	}

	output, err := s.runPlugin(plugin, input)
	endTime := time.Now()
	duration := int(endTime.Sub(startTime).Milliseconds())

	log.EndTime = endTime
	log.Duration = duration

	if err != nil {
		log.Status = models.ExecutionStatusFailed
		log.ErrorMessage = err.Error()
	} else {
		outputJSON, _ := json.Marshal(output)
		log.OutputData = string(outputJSON)
	}

	if err := s.db.Create(log).Error; err != nil {
		return nil, err
	}

	return log, nil
}

func (s *PluginService) runPlugin(plugin *models.PluginRegistry, input any) (any, error) {
	switch plugin.PluginType {
	case models.PluginTypeAlphaGenerator:
		return s.runAlphaGeneratorPlugin(plugin, input)
	case models.PluginTypePortfolioOptimizer:
		return s.runPortfolioOptimizerPlugin(plugin, input)
	case models.PluginTypeRiskModel:
		return s.runRiskModelPlugin(plugin, input)
	default:
		return nil, errors.New("unsupported plugin type")
	}
}

func (s *PluginService) runAlphaGeneratorPlugin(plugin *models.PluginRegistry, input any) (any, error) {
	return map[string]any{
		"plugin_name": plugin.PluginName,
		"plugin_type": plugin.PluginType,
		"result":      "Alpha generator plugin executed successfully",
		"timestamp":   time.Now(),
	}, nil
}

func (s *PluginService) runPortfolioOptimizerPlugin(plugin *models.PluginRegistry, input any) (any, error) {
	return map[string]any{
		"plugin_name": plugin.PluginName,
		"plugin_type": plugin.PluginType,
		"result":      "Portfolio optimizer plugin executed successfully",
		"timestamp":   time.Now(),
	}, nil
}

func (s *PluginService) runRiskModelPlugin(plugin *models.PluginRegistry, input any) (any, error) {
	return map[string]any{
		"plugin_name": plugin.PluginName,
		"plugin_type": plugin.PluginType,
		"result":      "Risk model plugin executed successfully",
		"timestamp":   time.Now(),
	}, nil
}

func (s *PluginService) GetExecutionLogs(pluginID uint, limit int) ([]models.PluginExecutionLog, error) {
	var logs []models.PluginExecutionLog
	query := s.db.Where("plugin_id = ?", pluginID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("start_time DESC").Find(&logs).Error
	return logs, err
}

func (s *PluginService) GetExecutionLog(id uint) (*models.PluginExecutionLog, error) {
	var log models.PluginExecutionLog
	err := s.db.First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("execution log not found")
		}
		return nil, err
	}
	return &log, nil
}

func (s *PluginService) CreateBenchmarkMatrix(matrix *models.ModelBenchmarkMatrix) error {
	matrix.CreatedAt = time.Now()
	return s.db.Create(matrix).Error
}

func (s *PluginService) GetBenchmarkMatrix(id uint) (*models.ModelBenchmarkMatrix, error) {
	var matrix models.ModelBenchmarkMatrix
	err := s.db.First(&matrix, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("benchmark matrix not found")
		}
		return nil, err
	}
	return &matrix, nil
}

func (s *PluginService) ListBenchmarkMatrices() ([]models.ModelBenchmarkMatrix, error) {
	var matrices []models.ModelBenchmarkMatrix
	err := s.db.Order("created_at DESC").Find(&matrices).Error
	return matrices, err
}

func (s *PluginService) UpdateBenchmarkMatrix(matrix *models.ModelBenchmarkMatrix) error {
	return s.db.Save(matrix).Error
}

func (s *PluginService) DeleteBenchmarkMatrix(id uint) error {
	return s.db.Delete(&models.ModelBenchmarkMatrix{}, id).Error
}

func (s *PluginService) CreateExperiment(experiment *models.StrategyExperiment) error {
	experiment.CreatedAt = time.Now()
	experiment.UpdatedAt = time.Now()
	return s.db.Create(experiment).Error
}

func (s *PluginService) GetExperiment(id uint) (*models.StrategyExperiment, error) {
	var experiment models.StrategyExperiment
	err := s.db.First(&experiment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("experiment not found")
		}
		return nil, err
	}
	return &experiment, nil
}

func (s *PluginService) ListExperiments(status models.ExperimentStatus) ([]models.StrategyExperiment, error) {
	var experiments []models.StrategyExperiment
	query := s.db.Model(&models.StrategyExperiment{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&experiments).Error
	return experiments, err
}

func (s *PluginService) UpdateExperiment(experiment *models.StrategyExperiment) error {
	experiment.UpdatedAt = time.Now()
	return s.db.Save(experiment).Error
}

func (s *PluginService) DeleteExperiment(id uint) error {
	return s.db.Delete(&models.StrategyExperiment{}, id).Error
}

func (s *PluginService) StartExperiment(id uint) error {
	return s.db.Model(&models.StrategyExperiment{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     models.ExperimentStatusRunning,
			"start_date": time.Now(),
		}).Error
}

func (s *PluginService) CompleteExperiment(id uint, results string) error {
	return s.db.Model(&models.StrategyExperiment{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":            models.ExperimentStatusCompleted,
			"experiment_result": results,
			"end_date":          time.Now(),
		}).Error
}

func (s *PluginService) FailExperiment(id uint, errorMsg string) error {
	return s.db.Model(&models.StrategyExperiment{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":   models.ExperimentStatusFailed,
			"end_date": time.Now(),
		}).Error
}
