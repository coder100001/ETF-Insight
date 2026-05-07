package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidViewType    = errors.New("invalid view type")
	ErrInvalidConfidence  = errors.New("confidence must be between 0 and 100")
	ErrViewExpired        = errors.New("view has expired")
	ErrInvalidPriorType   = errors.New("invalid prior type")
	ErrInvalidOmegaMethod = errors.New("invalid omega method")
)

type AlphaViewService struct {
	db            *gorm.DB
	factorService *FactorDataService
}

func NewAlphaViewService(db *gorm.DB, factorService *FactorDataService) *AlphaViewService {
	return &AlphaViewService{
		db:            db,
		factorService: factorService,
	}
}

func (s *AlphaViewService) CreateAlphaView(view *models.AlphaView) error {
	if err := s.validateView(view); err != nil {
		return err
	}
	if s.db == nil {
		return ErrDatabaseNotInitialized
	}

	return s.db.Create(view).Error
}

func (s *AlphaViewService) GetAlphaView(id uint) (*models.AlphaView, error) {
	if s.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	var view models.AlphaView
	err := s.db.First(&view, id).Error
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (s *AlphaViewService) GetActiveAlphaViews(assetSymbol string) ([]models.AlphaView, error) {
	if s.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	var views []models.AlphaView
	query := s.db.Where("status = ?", models.ViewStatusActive)

	if assetSymbol != "" {
		query = query.Where("asset_symbol = ?", assetSymbol)
	}

	err := query.Order("created_at DESC").Find(&views).Error
	return views, err
}

func (s *AlphaViewService) UpdateAlphaView(view *models.AlphaView) error {
	if s.db == nil {
		return ErrDatabaseNotInitialized
	}
	return s.db.Save(view).Error
}

func (s *AlphaViewService) DeactivateView(id uint) error {
	if s.db == nil {
		return ErrDatabaseNotInitialized
	}
	return s.db.Model(&models.AlphaView{}).
		Where("id = ?", id).
		Update("status", models.ViewStatusExpired).Error
}

func (s *AlphaViewService) GenerateViewFromFactorTiming(factorName, assetSymbol string) (*models.AlphaView, error) {
	signal, err := s.factorService.GetLatestTimingSignal(factorName)
	if err != nil {
		return nil, fmt.Errorf("failed to get timing signal: %w", err)
	}

	viewType := models.ViewTypeAbsolute
	view := &models.AlphaView{
		AssetSymbol:  assetSymbol,
		ViewType:     viewType,
		ViewReturn:   signal.ExpectedReturn,
		Confidence:   signal.Confidence,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		ViewMethod:   models.ViewMethodFactorTiming,
		SourceFactor: factorName,
		Status:       models.ViewStatusActive,
		CreatedAt:    time.Now(),
	}

	if err := s.CreateAlphaView(view); err != nil {
		return nil, err
	}

	return view, nil
}

func (s *AlphaViewService) RecordViewPerformance(performance *models.AlphaViewPerformance) error {
	if s.db == nil {
		return ErrDatabaseNotInitialized
	}
	return s.db.Create(performance).Error
}

func (s *AlphaViewService) GetViewPerformance(viewID uint) (*models.AlphaViewPerformance, error) {
	if s.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	var performance models.AlphaViewPerformance
	err := s.db.Where("view_id = ?", viewID).First(&performance).Error
	if err != nil {
		return nil, err
	}
	return &performance, nil
}

func (s *AlphaViewService) validateView(view *models.AlphaView) error {
	if !view.ViewType.IsValid() {
		return ErrInvalidViewType
	}

	if view.Confidence.LessThan(decimal.Zero) || view.Confidence.GreaterThan(decimal.NewFromInt(100)) {
		return ErrInvalidConfidence
	}

	if view.ValidUntil.Before(time.Now()) {
		return ErrViewExpired
	}

	return nil
}

type BlackLittermanService struct {
	db           *gorm.DB
	alphaService *AlphaViewService
}

func NewBlackLittermanService(db *gorm.DB, alphaService *AlphaViewService) *BlackLittermanService {
	return &BlackLittermanService{
		db:           db,
		alphaService: alphaService,
	}
}

func (s *BlackLittermanService) CreateConfig(config *models.BlackLittermanConfig) error {
	if err := s.validateConfig(config); err != nil {
		return err
	}
	if s.db == nil {
		return ErrDatabaseNotInitialized
	}

	// Atomic upsert - handles race condition safely
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "portfolio_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"risk_aversion", "prior_type", "prior_weights", "implied_returns", "omega_method", "omega_matrix", "is_active", "last_calculated", "updated_at"}),
	}).Create(config).Error
}

func (s *BlackLittermanService) GetConfig(id uint) (*models.BlackLittermanConfig, error) {
	if s.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	var config models.BlackLittermanConfig
	err := s.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *BlackLittermanService) UpdateConfig(config *models.BlackLittermanConfig) error {
	if s.db == nil {
		return ErrDatabaseNotInitialized
	}
	return s.db.Save(config).Error
}

func (s *BlackLittermanService) CalculatePosteriorReturns(configID uint, views []models.AlphaView) (*models.BLPosteriorReturn, error) {
	if s.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	config, err := s.GetConfig(configID)
	if err != nil {
		return nil, err
	}

	marketWeights, err := s.parseMarketWeights(config.PriorWeights)
	if err != nil {
		return nil, err
	}

	covMatrix, err := s.parseCovarianceMatrix(config.OmegaMatrix)
	if err != nil {
		return nil, err
	}

	pi := s.calculateEquilibriumReturns(marketWeights, covMatrix, config.RiskAversion)

	var portfolio models.Portfolio
	err = s.db.Preload("Positions").First(&portfolio, config.PortfolioID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load portfolio %d: %w", config.PortfolioID, err)
	}

	assetSymbols := make([]string, len(portfolio.Positions))
	for i, pos := range portfolio.Positions {
		assetSymbols[i] = pos.Symbol
	}

	P, Q, Omega, err := s.buildViewMatrices(views, len(marketWeights), assetSymbols)
	if err != nil {
		return nil, err
	}

	tau := decimal.NewFromFloat(0.05)
	posteriorReturns, posteriorCov := s.blFormula(pi, covMatrix, P, Q, Omega, tau, config.RiskAversion)

	posterior := &models.BLPosteriorReturn{
		ConfigID:         configID,
		CalculationDate:  time.Now(),
		PosteriorReturns: posteriorReturns,
		PosteriorCov:     posteriorCov,
		NumViews:         len(views),
		CreatedAt:        time.Now(),
	}

	if err := s.db.Create(posterior).Error; err != nil {
		return nil, err
	}

	return posterior, nil
}

func (s *BlackLittermanService) CalculatePosteriorReturnsByIDs(configID uint, viewIDs []uint) (*models.BLPosteriorReturn, error) {
	if s.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	var views []models.AlphaView
	if err := s.db.Where("id IN ?", viewIDs).Find(&views).Error; err != nil {
		return nil, fmt.Errorf("failed to load alpha views: %w", err)
	}

	if len(views) == 0 {
		return nil, fmt.Errorf("no alpha views found for the given IDs")
	}

	return s.CalculatePosteriorReturns(configID, views)
}

func (s *BlackLittermanService) GetPosteriorReturns(configID uint) (*models.BLPosteriorReturn, error) {
	if s.db == nil {
		return nil, ErrDatabaseNotInitialized
	}

	var posterior models.BLPosteriorReturn
	err := s.db.Where("config_id = ?", configID).
		Order("calculation_date DESC").
		First(&posterior).Error
	if err != nil {
		return nil, err
	}
	return &posterior, nil
}

func (s *BlackLittermanService) validateConfig(config *models.BlackLittermanConfig) error {
	if !config.PriorType.IsValid() {
		return ErrInvalidPriorType
	}

	if !config.OmegaMethod.IsValid() {
		return ErrInvalidOmegaMethod
	}

	if config.RiskAversion.LessThanOrEqual(decimal.Zero) {
		return errors.New("risk aversion must be positive")
	}

	return nil
}

func (s *BlackLittermanService) parseMarketWeights(weightsJSON models.JSONMap) ([]decimal.Decimal, error) {
	if weightsJSON == nil {
		return nil, errors.New("weights JSON is nil")
	}

	weightsBytes, err := json.Marshal(weightsJSON)
	if err != nil {
		return nil, err
	}

	// Try array format first: [0.25, 0.25, 0.25, 0.25]
	var weights []float64
	if err := json.Unmarshal(weightsBytes, &weights); err == nil {
		result := make([]decimal.Decimal, len(weights))
		for i, w := range weights {
			result[i] = decimal.NewFromFloat(w)
		}
		return result, nil
	}

	// Fall back to map format: {"0": 0.25, "1": 0.25, ...}
	var weightsMap map[string]float64
	if err := json.Unmarshal(weightsBytes, &weightsMap); err != nil {
		return nil, fmt.Errorf("failed to parse weights: %w", err)
	}

	// Convert map to ordered slice (keys are "0", "1", "2", ...)
	result := make([]decimal.Decimal, len(weightsMap))
	for i := 0; i < len(weightsMap); i++ {
		key := fmt.Sprintf("%d", i)
		val, ok := weightsMap[key]
		if !ok {
			return nil, fmt.Errorf("missing weight at index %d", i)
		}
		result[i] = decimal.NewFromFloat(val)
	}
	return result, nil
}

func (s *BlackLittermanService) parseCovarianceMatrix(covJSON models.JSONMap) ([][]decimal.Decimal, error) {
	if covJSON == nil {
		return nil, errors.New("covariance JSON is nil")
	}

	covBytes, err := json.Marshal(covJSON)
	if err != nil {
		return nil, err
	}

	// Try array format first: [[0.01, 0], [0, 0.01]]
	var cov [][]float64
	if err := json.Unmarshal(covBytes, &cov); err == nil {
		result := make([][]decimal.Decimal, len(cov))
		for i, row := range cov {
			result[i] = make([]decimal.Decimal, len(row))
			for j, val := range row {
				result[i][j] = decimal.NewFromFloat(val)
			}
		}
		return result, nil
	}

	// Fall back to map format: {"0": {"0": 0.01, "1": 0}, "1": {"0": 0, "1": 0.01}}
	var covMap map[string]map[string]float64
	if err := json.Unmarshal(covBytes, &covMap); err != nil {
		return nil, fmt.Errorf("failed to parse covariance matrix: %w", err)
	}

	// Convert map to ordered 2D slice
	n := len(covMap)
	result := make([][]decimal.Decimal, n)
	for i := range n {
		key := fmt.Sprintf("%d", i)
		row, ok := covMap[key]
		if !ok {
			return nil, fmt.Errorf("missing row %d in covariance matrix", i)
		}
		if len(row) != n {
			return nil, fmt.Errorf("row %d has %d elements, expected %d", i, len(row), n)
		}
		result[i] = make([]decimal.Decimal, n)
		for j := range n {
			colKey := fmt.Sprintf("%d", j)
			val, ok := row[colKey]
			if !ok {
				return nil, fmt.Errorf("missing column %d in row %d of covariance matrix", j, i)
			}
			result[i][j] = decimal.NewFromFloat(val)
		}
	}
	return result, nil
}

func (s *BlackLittermanService) calculateEquilibriumReturns(weights []decimal.Decimal, cov [][]decimal.Decimal, riskAversion decimal.Decimal) []decimal.Decimal {
	n := len(weights)
	returns := make([]decimal.Decimal, n)

	for i := range n {
		sum := decimal.Zero
		for j := range n {
			sum = sum.Add(cov[i][j].Mul(weights[j]))
		}
		returns[i] = riskAversion.Mul(sum)
	}

	return returns
}

func (s *BlackLittermanService) buildViewMatrices(views []models.AlphaView, nAssets int, assetSymbols []string) ([][]decimal.Decimal, []decimal.Decimal, [][]decimal.Decimal, error) {
	viewCount := len(views)

	P := make([][]decimal.Decimal, viewCount)
	for i := range P {
		P[i] = make([]decimal.Decimal, nAssets)
	}

	Q := make([]decimal.Decimal, viewCount)

	Omega := make([][]decimal.Decimal, viewCount)
	for i := range Omega {
		Omega[i] = make([]decimal.Decimal, viewCount)
	}

	for i, view := range views {
		assetIndex := -1
		for j, symbol := range assetSymbols {
			if symbol == view.AssetSymbol {
				assetIndex = j
				break
			}
		}
		if assetIndex == -1 {
			return nil, nil, nil, errors.New("asset symbol not found in portfolio: " + view.AssetSymbol)
		}
		P[i][assetIndex] = decimal.NewFromInt(1)
		Q[i] = view.ViewReturn

		omega := decimal.NewFromInt(100).Sub(view.Confidence).Div(decimal.NewFromInt(100))
		Omega[i][i] = omega.Mul(omega).Mul(decimal.NewFromFloat(0.01))
	}

	return P, Q, Omega, nil
}

func (s *BlackLittermanService) blFormula(pi []decimal.Decimal, cov [][]decimal.Decimal, P [][]decimal.Decimal, Q []decimal.Decimal, Omega [][]decimal.Decimal, tau, riskAversion decimal.Decimal) (models.JSONMap, models.JSONMap) {
	n := len(pi)
	if n == 0 {
		return models.JSONMap{}, models.JSONMap{}
	}

	tauCov := scaleMatrix(cov, tau)
	tauCovInv, err := calculateMatrixInverse(tauCov)
	if err != nil {
		return jsonToMap(pi), jsonToMap(cov)
	}

	k := len(Q)
	if k == 0 || P == nil || len(P) == 0 || Omega == nil || len(Omega) == 0 {
		return jsonToMap(pi), jsonToMap(cov)
	}

	omegaInv, err := calculateMatrixInverse(Omega)
	if err != nil {
		return jsonToMap(pi), jsonToMap(cov)
	}

	pT := calculateMatrixTranspose(P)
	pTOmegaInv, err := calculateMatrixMultiply(pT, omegaInv)
	if err != nil {
		return jsonToMap(pi), jsonToMap(cov)
	}

	pTOmegaInvP, err := calculateMatrixMultiply(pTOmegaInv, P)
	if err != nil {
		return jsonToMap(pi), jsonToMap(cov)
	}

	sum := addMatrices(tauCovInv, pTOmegaInvP)
	sumInv, err := calculateMatrixInverse(sum)
	if err != nil {
		return jsonToMap(pi), jsonToMap(cov)
	}

	tauCovInvPi := matrixVectorMultiply(tauCovInv, pi)
	pTOmegaInvQ := matrixVectorMultiply(pTOmegaInv, Q)
	rhs := addVectors(tauCovInvPi, pTOmegaInvQ)

	posteriorReturns := matrixVectorMultiply(sumInv, rhs)

	posteriorCov := addMatrices(cov, sumInv)

	return jsonToMap(posteriorReturns), jsonToMap(posteriorCov)
}

func jsonToMap(v any) models.JSONMap {
	switch val := v.(type) {
	case []decimal.Decimal:
		// Convert vector to map: [0.1, 0.2] -> {"0": 0.1, "1": 0.2}
		result := make(models.JSONMap, len(val))
		for i, d := range val {
			result[fmt.Sprintf("%d", i)] = d.InexactFloat64()
		}
		return result
	case [][]decimal.Decimal:
		// Convert matrix to map: [[0.1, 0.2], [0.3, 0.4]] -> {"0": {"0": 0.1, "1": 0.2}, ...}
		result := make(models.JSONMap, len(val))
		for i, row := range val {
			rowMap := make(map[string]any, len(row))
			for j, d := range row {
				rowMap[fmt.Sprintf("%d", j)] = d.InexactFloat64()
			}
			result[fmt.Sprintf("%d", i)] = rowMap
		}
		return result
	default:
		// Fallback: marshal and unmarshal
		bytes, err := json.Marshal(v)
		if err != nil {
			return models.JSONMap{}
		}
		var result models.JSONMap
		if err := json.Unmarshal(bytes, &result); err != nil {
			return models.JSONMap{}
		}
		return result
	}
}

func scaleMatrix(matrix [][]decimal.Decimal, scalar decimal.Decimal) [][]decimal.Decimal {
	n := len(matrix)
	result := make([][]decimal.Decimal, n)
	for i := range n {
		result[i] = make([]decimal.Decimal, len(matrix[i]))
		for j := 0; j < len(matrix[i]); j++ {
			result[i][j] = matrix[i][j].Mul(scalar)
		}
	}
	return result
}

func addMatrices(A, B [][]decimal.Decimal) [][]decimal.Decimal {
	n := len(A)
	result := make([][]decimal.Decimal, n)
	for i := range n {
		result[i] = make([]decimal.Decimal, len(A[i]))
		for j := 0; j < len(A[i]); j++ {
			result[i][j] = A[i][j].Add(B[i][j])
		}
	}
	return result
}

func addVectors(a, b []decimal.Decimal) []decimal.Decimal {
	n := len(a)
	result := make([]decimal.Decimal, n)
	for i := range n {
		result[i] = a[i].Add(b[i])
	}
	return result
}

func matrixVectorMultiply(matrix [][]decimal.Decimal, vector []decimal.Decimal) []decimal.Decimal {
	m := len(matrix)
	result := make([]decimal.Decimal, m)
	for i := range m {
		sum := decimal.Zero
		for j := range vector {
			sum = sum.Add(matrix[i][j].Mul(vector[j]))
		}
		result[i] = sum
	}
	return result
}

func calculateMatrixInverse(matrix [][]decimal.Decimal) ([][]decimal.Decimal, error) {
	n := len(matrix)

	workMatrix := make([][]decimal.Decimal, n)
	for i := range n {
		workMatrix[i] = make([]decimal.Decimal, len(matrix[i]))
		copy(workMatrix[i], matrix[i])
	}

	inverse := make([][]decimal.Decimal, n)
	for i := range inverse {
		inverse[i] = make([]decimal.Decimal, n)
		inverse[i][i] = decimal.NewFromInt(1)
	}

	for i := range n {
		pivot := workMatrix[i][i]
		if pivot.IsZero() {
			return nil, errors.New("matrix is singular")
		}

		for j := range n {
			workMatrix[i][j] = workMatrix[i][j].Div(pivot)
			inverse[i][j] = inverse[i][j].Div(pivot)
		}

		for k := range n {
			if k != i {
				factor := workMatrix[k][i]
				for j := range n {
					workMatrix[k][j] = workMatrix[k][j].Sub(factor.Mul(workMatrix[i][j]))
					inverse[k][j] = inverse[k][j].Sub(factor.Mul(inverse[i][j]))
				}
			}
		}
	}

	return inverse, nil
}

func calculateMatrixMultiply(A, B [][]decimal.Decimal) ([][]decimal.Decimal, error) {
	if len(A[0]) != len(B) {
		return nil, errors.New("matrix dimensions do not match")
	}

	m := len(A)
	n := len(B[0])
	p := len(B)

	result := make([][]decimal.Decimal, m)
	for i := range m {
		result[i] = make([]decimal.Decimal, n)
		for j := range n {
			sum := decimal.Zero
			for k := range p {
				sum = sum.Add(A[i][k].Mul(B[k][j]))
			}
			result[i][j] = sum
		}
	}

	return result, nil
}

func calculateMatrixTranspose(matrix [][]decimal.Decimal) [][]decimal.Decimal {
	if len(matrix) == 0 {
		return matrix
	}

	m := len(matrix)
	n := len(matrix[0])
	result := make([][]decimal.Decimal, n)
	for i := range n {
		result[i] = make([]decimal.Decimal, m)
		for j := range m {
			result[i][j] = matrix[j][i]
		}
	}
	return result
}

func decimalSqrt(d decimal.Decimal) decimal.Decimal {
	return decimal.NewFromFloat(math.Sqrt(d.InexactFloat64()))
}
