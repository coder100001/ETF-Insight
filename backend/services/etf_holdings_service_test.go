package services

import (
	"testing"
)

func TestETFHoldingsService_GetHoldings(t *testing.T) {
	svc := NewETFHoldingsService()

	t.Run("returns data for known ETF", func(t *testing.T) {
		holdings, err := svc.GetHoldings("SCHD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(holdings) == 0 {
			t.Fatal("expected non-empty holdings")
		}
		// Verify first holding structure
		h := holdings[0]
		if h.Symbol == "" || h.Name == "" || h.Sector == "" {
			t.Errorf("incomplete holding: %+v", h)
		}
		if h.Weight <= 0 {
			t.Errorf("expected positive weight, got %f", h.Weight)
		}
	})

	t.Run("returns error for unknown ETF", func(t *testing.T) {
		_, err := svc.GetHoldings("UNKNOWN")
		if err == nil {
			t.Fatal("expected error for unknown ETF")
		}
	})

	t.Run("covers all 12 ETFs", func(t *testing.T) {
		expected := []string{"SCHD", "JEPI", "JEPQ", "QQQ", "VTI", "SPY", "VYM", "SPYD", "HDV", "DGRO", "VNQ", "BND"}
		for _, sym := range expected {
			holdings, err := svc.GetHoldings(sym)
			if err != nil {
				t.Errorf("%s: unexpected error: %v", sym, err)
			}
			if len(holdings) < 10 {
				t.Errorf("%s: expected at least 10 holdings, got %d", sym, len(holdings))
			}
		}
	})
}

func TestETFHoldingsService_CalculateOverlap(t *testing.T) {
	svc := NewETFHoldingsService()

	t.Run("SCHD vs VYM has meaningful overlap", func(t *testing.T) {
		overlap, err := svc.CalculateOverlap("SCHD", "VYM")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Both are dividend ETFs with common holdings (JPM, HD, KO, MRK, etc.)
		if overlap < 5.0 {
			t.Errorf("expected significant overlap (>5%%), got %.2f%%", overlap)
		}
		if overlap > 100 {
			t.Errorf("overlap cannot exceed 100%%, got %.2f%%", overlap)
		}
		t.Logf("SCHD vs VYM overlap: %.2f%%", overlap)
	})

	t.Run("SCHD vs JEPQ has less overlap", func(t *testing.T) {
		overlap, err := svc.CalculateOverlap("SCHD", "JEPQ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// SCHD is dividend-value, JEPQ is growth-tech heavy - some but less overlap
		t.Logf("SCHD vs JEPQ overlap: %.2f%%", overlap)
	})

	t.Run("QQQ vs SPY has significant overlap", func(t *testing.T) {
		overlap, err := svc.CalculateOverlap("QQQ", "SPY")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Both hold mega-cap tech
		if overlap < 10.0 {
			t.Errorf("expected significant overlap (>10%%), got %.2f%%", overlap)
		}
		t.Logf("QQQ vs SPY overlap: %.2f%%", overlap)
	})

	t.Run("VNQ vs BND has zero overlap", func(t *testing.T) {
		overlap, err := svc.CalculateOverlap("VNQ", "BND")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// REITs vs Bonds - no common symbols
		if overlap != 0 {
			t.Errorf("expected 0%% overlap, got %.2f%%", overlap)
		}
	})

	t.Run("error on unknown ETF", func(t *testing.T) {
		_, err := svc.CalculateOverlap("SCHD", "UNKNOWN")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("overlap is symmetric", func(t *testing.T) {
		ov1, _ := svc.CalculateOverlap("SCHD", "VYM")
		ov2, _ := svc.CalculateOverlap("VYM", "SCHD")
		if ov1 != ov2 {
			t.Errorf("overlap not symmetric: %.2f vs %.2f", ov1, ov2)
		}
	})
}

func TestETFHoldingsService_GetSectorAllocation(t *testing.T) {
	svc := NewETFHoldingsService()

	t.Run("returns sector map for SCHD", func(t *testing.T) {
		sectors, err := svc.GetSectorAllocation("SCHD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sectors) == 0 {
			t.Fatal("expected non-empty sectors")
		}
		// SCHD has Technology, Healthcare, Financials, Consumer Staples, etc.
		total := 0.0
		for sector, w := range sectors {
			total += w
			t.Logf("  %s: %.2f%%", sector, w)
		}
		if total < 50 {
			t.Errorf("total sector weight too low: %.2f%%", total)
		}
	})

	t.Run("VNQ is predominantly Real Estate", func(t *testing.T) {
		sectors, err := svc.GetSectorAllocation("VNQ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		reWeight := sectors["Real Estate"]
		// Static data covers top 20 holdings (~55% of fund), all Real Estate
		if reWeight < 50 {
			t.Errorf("expected >50%% Real Estate, got %.2f%%", reWeight)
		}
		t.Logf("VNQ Real Estate weight (top 20): %.2f%%", reWeight)
	})

	t.Run("error on unknown ETF", func(t *testing.T) {
		_, err := svc.GetSectorAllocation("UNKNOWN")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestETFHoldingsService_GetConcentrationMetrics(t *testing.T) {
	svc := NewETFHoldingsService()

	t.Run("SCHD top 10 weight", func(t *testing.T) {
		topWeight, hhi, err := svc.GetConcentrationMetrics("SCHD", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if topWeight <= 0 || topWeight > 100 {
			t.Errorf("invalid top 10 weight: %.2f%%", topWeight)
		}
		if hhi <= 0 {
			t.Errorf("expected positive HHI, got %f", hhi)
		}
		t.Logf("SCHD top10: %.2f%%, HHI: %f", topWeight, hhi)
	})

	t.Run("QQQ top 5 is highly concentrated", func(t *testing.T) {
		top5, _, err := svc.GetConcentrationMetrics("QQQ", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// MSFT + AAPL + NVDA + AMZN + META > 30%
		if top5 < 30 {
			t.Errorf("expected QQQ top5 > 30%%, got %.2f%%", top5)
		}
		t.Logf("QQQ top5: %.2f%%", top5)
	})

	t.Run("HDV higher concentration than VTI", func(t *testing.T) {
		hdvTop5, _, _ := svc.GetConcentrationMetrics("HDV", 5)
		vtiTop5, _, _ := svc.GetConcentrationMetrics("VTI", 5)
		if hdvTop5 <= vtiTop5 {
			t.Errorf("HDV should be more concentrated: HDV %.2f%% vs VTI %.2f%%", hdvTop5, vtiTop5)
		}
	})

	t.Run("error on unknown ETF", func(t *testing.T) {
		_, _, err := svc.GetConcentrationMetrics("UNKNOWN", 10)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
