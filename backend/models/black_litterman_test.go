package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestBlackLittermanConfig_JSONMapSerialization(t *testing.T) {
	config := BlackLittermanConfig{
		PortfolioID:    1,
		RiskAversion:   decimal.NewFromFloat(2.5),
		PriorType:      PriorTypeEqualWeight,
		PriorWeights:   JSONMap{"0": 0.25, "1": 0.25, "2": 0.25, "3": 0.25},
		ImpliedReturns: JSONMap{"0": 0.08, "1": 0.06, "2": 0.07, "3": 0.05},
		OmegaMethod:    OmegaMethodIdzorek,
		OmegaMatrix:    JSONMap{"0": JSONMap{"0": 0.01, "1": 0}, "1": JSONMap{"0": 0, "1": 0.01}},
		IsActive:       true,
		LastCalculated: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded BlackLittermanConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.PriorWeights == nil {
		t.Error("PriorWeights should not be nil after round-trip")
	}
	pw, ok := decoded.PriorWeights["0"].(float64)
	if !ok || pw != 0.25 {
		t.Errorf("Expected PriorWeights[0]=0.25, got %v (%T)", decoded.PriorWeights["0"], decoded.PriorWeights["0"])
	}

	if decoded.OmegaMatrix == nil {
		t.Error("OmegaMatrix should not be nil after round-trip")
	}
	row0, ok := decoded.OmegaMatrix["0"].(map[string]any)
	if !ok {
		t.Fatalf("OmegaMatrix[0] should be map, got %T", decoded.OmegaMatrix["0"])
	}
	cell00, ok := row0["0"].(float64)
	if !ok || cell00 != 0.01 {
		t.Errorf("Expected OmegaMatrix[0][0]=0.01, got %v", cell00)
	}

	if decoded.PortfolioID != 1 {
		t.Errorf("PortfolioID mismatch: expected 1, got %d", decoded.PortfolioID)
	}
	if decoded.PriorType != PriorTypeEqualWeight {
		t.Errorf("PriorType mismatch: expected %s, got %s", PriorTypeEqualWeight, decoded.PriorType)
	}
}

func TestBlackLittermanConfig_NilJSONMap(t *testing.T) {
	config := BlackLittermanConfig{
		PortfolioID:  1,
		PriorType:    PriorTypeEqualWeight,
		PriorWeights: nil,
		OmegaMatrix:  nil,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded BlackLittermanConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.PriorWeights != nil {
		t.Errorf("Expected nil PriorWeights, got %v", decoded.PriorWeights)
	}
	if decoded.OmegaMatrix != nil {
		t.Errorf("Expected nil OmegaMatrix, got %v", decoded.OmegaMatrix)
	}
}

func TestBLPosteriorReturn_JSONMapSerialization(t *testing.T) {
	posterior := BLPosteriorReturn{
		ConfigID:         1,
		CalculationDate:  time.Now(),
		PosteriorReturns: JSONMap{"AAPL": 0.09, "MSFT": 0.07},
		PosteriorWeights: JSONMap{"AAPL": 0.4, "MSFT": 0.35, "GOOG": 0.25},
		PosteriorCov:     JSONMap{"AAPL": JSONMap{"AAPL": 0.04}, "MSFT": JSONMap{"MSFT": 0.03}},
		NumViews:         2,
		ViewImpact:       decimal.NewFromFloat(0.05),
		CreatedAt:        time.Now(),
	}

	data, err := json.Marshal(posterior)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded BLPosteriorReturn
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.PosteriorReturns == nil {
		t.Error("PosteriorReturns should not be nil after round-trip")
	}
	if decoded.NumViews != 2 {
		t.Errorf("NumViews mismatch: expected 2, got %d", decoded.NumViews)
	}
	if !decoded.ViewImpact.Equal(decimal.NewFromFloat(0.05)) {
		t.Errorf("ViewImpact mismatch: expected 0.05, got %s", decoded.ViewImpact.String())
	}
}

func TestPriorType_Validation(t *testing.T) {
	tests := []struct {
		name     string
		p        PriorType
		expected bool
	}{
		{"valid equal_weight", PriorTypeEqualWeight, true},
		{"valid min_variance", PriorTypeMinVariance, true},
		{"valid market_cap", PriorTypeMarketCap, true},
		{"invalid empty", PriorType(""), false},
		{"invalid random", PriorType("random_type"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsValid(); got != tt.expected {
				t.Errorf("PriorType(%q).IsValid() = %v, want %v", tt.p, got, tt.expected)
			}
		})
	}
}

func TestOmegaMethod_Validation(t *testing.T) {
	tests := []struct {
		name     string
		o        OmegaMethod
		expected bool
	}{
		{"valid Idzorek", OmegaMethodIdzorek, true},
		{"valid HeLitterman", OmegaMethodHeLitterman, true},
		{"invalid empty", OmegaMethod(""), false},
		{"invalid random", OmegaMethod("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.IsValid(); got != tt.expected {
				t.Errorf("OmegaMethod(%q).IsValid() = %v, want %v", tt.o, got, tt.expected)
			}
		})
	}
}

func TestBlackLittermanConfig_TableName(t *testing.T) {
	config := BlackLittermanConfig{}
	if config.TableName() != "black_litterman_configs" {
		t.Errorf("Expected TableName()=black_litterman_configs, got %s", config.TableName())
	}
}

func TestBLPosteriorReturn_TableName(t *testing.T) {
	posterior := BLPosteriorReturn{}
	if posterior.TableName() != "bl_posterior_returns" {
		t.Errorf("Expected TableName()=bl_posterior_returns, got %s", posterior.TableName())
	}
}
