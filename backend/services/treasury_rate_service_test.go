package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTreasuryRateService_NoAPIKey(t *testing.T) {
	svc := NewTreasuryRateService("")
	_, err := svc.Get10YearRate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not configured")
}

func TestTreasuryRateService_CachedRate(t *testing.T) {
	svc := NewTreasuryRateService("")
	svc.cache = 0.0435
	svc.lastFetch = time.Now()

	rate, ok := svc.GetCachedRate()
	assert.True(t, ok)
	assert.Equal(t, 0.0435, rate)
}

func TestTreasuryRateService_NoCache(t *testing.T) {
	svc := NewTreasuryRateService("")

	rate, ok := svc.GetCachedRate()
	assert.False(t, ok)
	assert.Equal(t, 0.0, rate)
}

func TestTreasuryRateService_StaleCacheFallback(t *testing.T) {
	svc := NewTreasuryRateService("")
	svc.cache = 0.0435
	svc.lastFetch = time.Now().Add(-2 * time.Hour) // stale

	// Without API key, stale cache is not returned (cache > 0 check passes but API key is empty)
	// The method checks API key before returning stale cache on request failure
	_, err := svc.Get10YearRate()
	assert.Error(t, err)
}

func TestFredResponseParsing(t *testing.T) {
	jsonStr := `{"observations":[{"date":"2026-06-05","value":"4.35"}]}`
	var result fredResponse
	err := json.Unmarshal([]byte(jsonStr), &result)
	assert.NoError(t, err)
	assert.Len(t, result.Observations, 1)
	assert.Equal(t, "4.35", result.Observations[0].Value)
	assert.Equal(t, "2026-06-05", result.Observations[0].Date)
}

func TestFredResponseParsing_MultipleObservations(t *testing.T) {
	jsonStr := `{"observations":[
		{"date":"2026-06-05","value":"4.35"},
		{"date":"2026-06-04","value":"4.30"}
	]}`
	var result fredResponse
	err := json.Unmarshal([]byte(jsonStr), &result)
	assert.NoError(t, err)
	assert.Len(t, result.Observations, 2)
}

func TestFredResponseParsing_EmptyObservations(t *testing.T) {
	jsonStr := `{"observations":[]}`
	var result fredResponse
	err := json.Unmarshal([]byte(jsonStr), &result)
	assert.NoError(t, err)
	assert.Len(t, result.Observations, 0)
}
