package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// TreasuryRateService fetches real-time 10-year Treasury yield from FRED API
type TreasuryRateService struct {
	apiKey    string
	cache     float64
	mu        sync.RWMutex
	lastFetch time.Time
	client    *http.Client
}

// NewTreasuryRateService creates a new TreasuryRateService
func NewTreasuryRateService(apiKey string) *TreasuryRateService {
	return &TreasuryRateService{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type fredResponse struct {
	Observations []struct {
		Date  string `json:"date"`
		Value string `json:"value"`
	} `json:"observations"`
}

// Get10YearRate returns the current 10-year Treasury yield as a decimal (e.g., 0.0435 for 4.35%).
// Results are cached for 1 hour.
func (s *TreasuryRateService) Get10YearRate() (float64, error) {
	s.mu.RLock()
	if time.Since(s.lastFetch) < 1*time.Hour && s.cache > 0 {
		defer s.mu.RUnlock()
		return s.cache, nil
	}
	s.mu.RUnlock()

	if s.apiKey == "" {
		return 0, fmt.Errorf("FRED API key not configured")
	}

	url := fmt.Sprintf(
		"https://api.stlouisfed.org/fred/series/observations?series_id=DGS10&api_key=%s&file_type=json&sort_order=desc&limit=1",
		s.apiKey,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		if s.cache > 0 {
			return s.cache, nil
		}
		return 0, fmt.Errorf("failed to fetch treasury rate: %w", err)
	}
	defer resp.Body.Close()

	var result fredResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode FRED response: %w", err)
	}

	if len(result.Observations) == 0 {
		return 0, fmt.Errorf("no observations from FRED")
	}

	var rate float64
	if _, err := fmt.Sscanf(result.Observations[0].Value, "%f", &rate); err != nil {
		return 0, fmt.Errorf("failed to parse rate: %w", err)
	}

	s.mu.Lock()
	s.cache = rate / 100 // Convert percentage to decimal
	s.lastFetch = time.Now()
	s.mu.Unlock()

	return s.cache, nil
}

// GetCachedRate returns the cached rate without making an API call.
// Returns (rate, true) if cache is populated, (0, false) otherwise.
func (s *TreasuryRateService) GetCachedRate() (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache > 0 {
		return s.cache, true
	}
	return 0, false
}
