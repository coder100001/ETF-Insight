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

	return s.db.Create(view).Error
}

func (s *AlphaViewService) GetAlphaView(id uint) (*models.AlphaView, error) {
	var view models.AlphaView
	err := s.db.First(&view, id).Error
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (s *AlphaViewService) GetActiveAlphaViews(assetSymbol string) ([]models.AlphaView, error) {
	var views []models.AlphaView
	query := s.db.Where("status = ?", models.ViewStatusActive)

	if assetSymbol != "" {
		query = query.Where("asset_symbol = ?", assetSymbol)
	}

	err := query.Order("created_at DESC").Find(&views).Error
	return views, err
}

func (s *AlphaViewService) UpdateAlphaView(view *models.AlphaView) error {
	return s.db.Save(view).Error
}

func (s *AlphaViewService) DeactivateView(id uint) error {
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
	return s.db.Create(performance).Error
}

func (s *AlphaViewService) GetViewPerformance(viewID uint) (*models.AlphaViewPerformance, error) {
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

	return s.db.Create(config).Error
}

func (s *BlackLittermanService) GetConfig(id uint) (*models.BlackLittermanConfig, error) {
	var config models.BlackLittermanConfig
	err := s.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *BlackLittermanService) UpdateConfig(config *models.BlackLittermanConfig) error {
	return s.db.Save(config).Error
}

func (s *BlackLittermanService) CalculatePosteriorReturns(configID uint, views []models.AlphaView) (*models.BLPosteriorReturn, error) {
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

	P, Q, Omega, err := s.buildViewMatrices(views, len(marketWeights))
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

func (s *BlackLittermanService) GetPosteriorReturns(configID uint) (*models.BLPosteriorReturn, error) {
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

func (s *BlackLittermanService) parseMarketWeights(weightsJSON string) ([]decimal.Decimal, error) {
	var weights []float64
	if err := json.Unmarshal([]byte(weightsJSON), &weights); err != nil {
		return nil, err
	}

	result := make([]decimal.Decimal, len(weights))
	for i, w := range weights {
		result[i] = decimal.NewFromFloat(w)
	}
	return result, nil
}

func (s *BlackLittermanService) parseCovarianceMatrix(covJSON string) ([][]decimal.Decimal, error) {
	var cov [][]float64
	if err := json.Unmarshal([]byte(covJSON), &cov); err != nil {
		return nil, err
	}

	result := make([][]decimal.Decimal, len(cov))
	for i, row := range cov {
		result[i] = make([]decimal.Decimal, len(row))
		for j, val := range row {
			result[i][j] = decimal.NewFromFloat(val)
		}
	}
	return result, nil
}

func (s *BlackLittermanService) calculateEquilibriumReturns(weights []decimal.Decimal, cov [][]decimal.Decimal, riskAversion decimal.Decimal) []decimal.Decimal {
	n := len(weights)
	returns := make([]decimal.Decimal, n)

	for i := 0; i < n; i++ {
		sum := decimal.Zero
		for j := 0; j < n; j++ {
			sum = sum.Add(cov[i][j].Mul(weights[j]))
		}
		returns[i] = riskAversion.Mul(sum)
	}

	return returns
}

func (s *BlackLittermanService) buildViewMatrices(views []models.AlphaView, nAssets int) ([][]decimal.Decimal, []decimal.Decimal, [][]decimal.Decimal, error) {
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
		P[i][0] = decimal.NewFromInt(1)
		Q[i] = view.ViewReturn

		omega := decimal.NewFromInt(100).Sub(view.Confidence).Div(decimal.NewFromInt(100))
		Omega[i][i] = omega.Mul(omega).Mul(decimal.NewFromFloat(0.01))
	}

	return P, Q, Omega, nil
}

func (s *BlackLittermanService) blFormula(pi []decimal.Decimal, cov [][]decimal.Decimal, P [][]decimal.Decimal, Q []decimal.Decimal, Omega [][]decimal.Decimal, tau, riskAversion decimal.Decimal) (string, string) {
	n := len(pi)

	tauCov := make([][]decimal.Decimal, n)
	for i := 0; i < n; i++ {
		tauCov[i] = make([]decimal.Decimal, n)
		for j := 0; j < n; j++ {
			tauCov[i][j] = cov[i][j].Mul(tau)
		}
	}

	posteriorReturns := make([]decimal.Decimal, n)
	for i := 0; i < n; i++ {
		posteriorReturns[i] = pi[i].Add(Q[0].Mul(decimal.NewFromFloat(0.1)))
	}

	posteriorCov := make([][]decimal.Decimal, n)
	for i := 0; i < n; i++ {
		posteriorCov[i] = make([]decimal.Decimal, n)
		for j := 0; j < n; j++ {
			posteriorCov[i][j] = cov[i][j].Add(tauCov[i][j])
		}
	}

	returnsJSON, _ := json.Marshal(posteriorReturns)
	covJSON, _ := json.Marshal(posteriorCov)

	return string(returnsJSON), string(covJSON)
}

func calculateMatrixInverse(matrix [][]decimal.Decimal) ([][]decimal.Decimal, error) {
	n := len(matrix)
	inverse := make([][]decimal.Decimal, n)
	for i := range inverse {
		inverse[i] = make([]decimal.Decimal, n)
		inverse[i][i] = decimal.NewFromInt(1)
	}

	for i := 0; i < n; i++ {
		pivot := matrix[i][i]
		if pivot.IsZero() {
			return nil, errors.New("matrix is singular")
		}

		for j := 0; j < n; j++ {
			matrix[i][j] = matrix[i][j].Div(pivot)
			inverse[i][j] = inverse[i][j].Div(pivot)
		}

		for k := 0; k < n; k++ {
			if k != i {
				factor := matrix[k][i]
				for j := 0; j < n; j++ {
					matrix[k][j] = matrix[k][j].Sub(factor.Mul(matrix[i][j]))
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
	for i := 0; i < m; i++ {
		result[i] = make([]decimal.Decimal, n)
		for j := 0; j < n; j++ {
			sum := decimal.Zero
			for k := 0; k < p; k++ {
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
	for i := 0; i < n; i++ {
		result[i] = make([]decimal.Decimal, m)
		for j := 0; j < m; j++ {
			result[i][j] = matrix[j][i]
		}
	}
	return result
}

func decimalSqrt(d decimal.Decimal) decimal.Decimal {
	return decimal.NewFromFloat(math.Sqrt(d.InexactFloat64()))
}
