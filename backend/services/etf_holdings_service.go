package services

import "fmt"

// ETFHoldingsService provides ETF holdings data and analysis methods.
type ETFHoldingsService struct{}

// NewETFHoldingsService creates a new ETFHoldingsService instance.
func NewETFHoldingsService() *ETFHoldingsService {
	return &ETFHoldingsService{}
}

// GetHoldings returns the static holdings for a given ETF symbol.
func (s *ETFHoldingsService) GetHoldings(symbol string) ([]StaticHolding, error) {
	if holdings, ok := staticHoldings[symbol]; ok {
		return holdings, nil
	}
	return nil, fmt.Errorf("no holdings data for %s", symbol)
}

// CalculateOverlap calculates portfolio overlap between two ETFs using the
// minimum weight method. Returns overlap percentage (0-100).
func (s *ETFHoldingsService) CalculateOverlap(etf1, etf2 string) (float64, error) {
	h1, err := s.GetHoldings(etf1)
	if err != nil {
		return 0, err
	}
	h2, err := s.GetHoldings(etf2)
	if err != nil {
		return 0, err
	}

	weights1 := make(map[string]float64)
	for _, h := range h1 {
		weights1[h.Symbol] = h.Weight
	}

	overlap := 0.0
	for _, h := range h2 {
		if w, ok := weights1[h.Symbol]; ok {
			if w < h.Weight {
				overlap += w
			} else {
				overlap += h.Weight
			}
		}
	}
	return overlap, nil
}

// GetSectorAllocation returns aggregated sector weights for an ETF.
func (s *ETFHoldingsService) GetSectorAllocation(symbol string) (map[string]float64, error) {
	holdings, err := s.GetHoldings(symbol)
	if err != nil {
		return nil, err
	}

	sectors := make(map[string]float64)
	for _, h := range holdings {
		sectors[h.Sector] += h.Weight
	}
	return sectors, nil
}

// GetConcentrationMetrics returns the top-N holdings weight and Herfindahl-Hirschman Index (HHI).
func (s *ETFHoldingsService) GetConcentrationMetrics(symbol string, topN int) (topWeight float64, hhi float64, err error) {
	holdings, err := s.GetHoldings(symbol)
	if err != nil {
		return 0, 0, err
	}

	for i, h := range holdings {
		if i < topN {
			topWeight += h.Weight
		}
		hhi += (h.Weight / 100) * (h.Weight / 100)
	}
	return topWeight, hhi, nil
}
