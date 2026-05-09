package unified

import (
	"context"
	"testing"
	"time"

	"etf-insight/services/datasource"
	erdatasource "etf-insight/services/exchange_rate/datasource"

	"github.com/shopspring/decimal"
)

type mockETFProvider struct {
	name      string
	available bool
	rateLimit int
}

func (m *mockETFProvider) GetName() string                      { return m.name }
func (m *mockETFProvider) GetProviderType() ProviderType        { return ProviderTypeETF }
func (m *mockETFProvider) IsAvailable(ctx context.Context) bool { return m.available }
func (m *mockETFProvider) GetRateLimit() int                    { return m.rateLimit }
func (m *mockETFProvider) GetHealth(ctx context.Context) *ProviderHealth {
	return &ProviderHealth{Name: m.name, Type: ProviderTypeETF, TypeStr: ProviderTypeETF.String(), Available: m.available}
}

type mockFullETFProvider struct {
	name      string
	available bool
	rateLimit int
}

func (m *mockFullETFProvider) GetName() string { return m.name }
func (m *mockFullETFProvider) GetQuote(ctx context.Context, symbol string) (*datasource.QuoteData, error) {
	return &datasource.QuoteData{Symbol: symbol, DataSource: m.name}, nil
}
func (m *mockFullETFProvider) GetQuotes(ctx context.Context, symbols []string) ([]*datasource.QuoteData, error) {
	return nil, nil
}
func (m *mockFullETFProvider) GetETFHoldings(ctx context.Context, symbol string, date time.Time) ([]*datasource.ETFHoldingData, error) {
	return nil, nil
}
func (m *mockFullETFProvider) IsAvailable(ctx context.Context) bool { return m.available }
func (m *mockFullETFProvider) GetRateLimit() int                    { return m.rateLimit }

type mockFullFXProvider struct {
	name      string
	available bool
	rateLimit int
}

func (m *mockFullFXProvider) GetName() string         { return m.name }
func (m *mockFullFXProvider) GetBaseCurrency() string { return "USD" }
func (m *mockFullFXProvider) GetRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	return decimal.NewFromFloat(1.0), nil
}
func (m *mockFullFXProvider) GetRates(ctx context.Context, baseCurrency string) (*erdatasource.BatchRateResult, error) {
	return nil, nil
}
func (m *mockFullFXProvider) IsAvailable(ctx context.Context) bool    { return m.available }
func (m *mockFullFXProvider) GetRateLimit() int                       { return m.rateLimit }
func (m *mockFullFXProvider) GetResponseTime() time.Duration          { return 0 }
func (m *mockFullFXProvider) GetSuccessRate() float64                 { return 1.0 }
func (m *mockFullFXProvider) GetAPIKey() string                       { return "test-key" }
func (m *mockFullFXProvider) GetSupportedCurrencies() []string        { return []string{"USD", "EUR"} }
func (m *mockFullFXProvider) ValidateAPIKey(ctx context.Context) bool { return true }

type mockFullAShareProvider struct {
	name      string
	available bool
	rateLimit int
}

func (m *mockFullAShareProvider) GetName() string { return m.name }
func (m *mockFullAShareProvider) FetchETFPrice(ctx context.Context, symbol string) (float64, error) {
	return 100.0, nil
}
func (m *mockFullAShareProvider) IsAvailable(ctx context.Context) bool { return m.available }
func (m *mockFullAShareProvider) GetRateLimit() int                    { return m.rateLimit }

type mockFullPythonService struct {
	name      string
	available bool
	rateLimit int
}

func (m *mockFullPythonService) GetName() string { return m.name }
func (m *mockFullPythonService) FetchData(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	return []byte(`{"data": "test"}`), nil
}
func (m *mockFullPythonService) IsAvailable(ctx context.Context) bool { return m.available }
func (m *mockFullPythonService) GetRateLimit() int                    { return m.rateLimit }

func TestUnifiedRegistry_RegisterAndGet(t *testing.T) {
	reg := GetUnifiedRegistry()
	p := &mockETFProvider{name: "test-etf", available: true, rateLimit: 10}
	reg.Register("test-etf", p)

	got, ok := reg.Get("test-etf")
	if !ok {
		t.Fatal("expected provider to be found")
	}
	if got.GetName() != "test-etf" {
		t.Errorf("expected name test-etf, got %s", got.GetName())
	}
	if got.GetProviderType() != ProviderTypeETF {
		t.Errorf("expected ProviderTypeETF, got %v", got.GetProviderType())
	}

	reg.Unregister("test-etf")
	if _, ok := reg.Get("test-etf"); ok {
		t.Error("expected provider to be unregistered")
	}
}

func TestUnifiedRegistry_List(t *testing.T) {
	reg := GetUnifiedRegistry()
	providers := []*mockETFProvider{
		{name: "b-provider", available: true, rateLimit: 5},
		{name: "a-provider", available: true, rateLimit: 10},
	}
	for _, p := range providers {
		reg.Register(p.name, p)
	}

	list := reg.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(list))
	}
	if list[0].GetName() != "a-provider" {
		t.Errorf("expected first to be a-provider, got %s", list[0].GetName())
	}
	if list[1].GetName() != "b-provider" {
		t.Errorf("expected second to be b-provider, got %s", list[1].GetName())
	}
	for _, p := range providers {
		reg.Unregister(p.name)
	}
}

type mockETFType struct {
	name      string
	available bool
	rateLimit int
}

func (m *mockETFType) GetName() string                      { return m.name }
func (m *mockETFType) GetProviderType() ProviderType        { return ProviderTypeETF }
func (m *mockETFType) IsAvailable(ctx context.Context) bool { return m.available }
func (m *mockETFType) GetRateLimit() int                    { return m.rateLimit }
func (m *mockETFType) GetHealth(ctx context.Context) *ProviderHealth {
	return &ProviderHealth{Name: m.name, Type: ProviderTypeETF, TypeStr: ProviderTypeETF.String(), Available: m.available}
}

type mockFXType struct {
	name      string
	available bool
	rateLimit int
}

func (m *mockFXType) GetName() string                      { return m.name }
func (m *mockFXType) GetProviderType() ProviderType        { return ProviderTypeFX }
func (m *mockFXType) IsAvailable(ctx context.Context) bool { return m.available }
func (m *mockFXType) GetRateLimit() int                    { return m.rateLimit }
func (m *mockFXType) GetHealth(ctx context.Context) *ProviderHealth {
	return &ProviderHealth{Name: m.name, Type: ProviderTypeFX, TypeStr: ProviderTypeFX.String(), Available: m.available}
}

func TestUnifiedRegistry_GetByType(t *testing.T) {
	reg := GetUnifiedRegistry()
	reg.Register("etf1", &mockETFType{name: "etf1"})
	reg.Register("etf2", &mockETFType{name: "etf2"})
	reg.Register("fx1", &mockFXType{name: "fx1"})

	etfs := reg.GetByType(ProviderTypeETF)
	if len(etfs) != 2 {
		t.Errorf("expected 2 ETF providers, got %d", len(etfs))
	}

	reg.Unregister("etf1")
	reg.Unregister("etf2")
	reg.Unregister("fx1")
}

func TestUnifiedRegistry_HealthCheck(t *testing.T) {
	reg := GetUnifiedRegistry()
	reg.Register("good", &mockETFProvider{name: "good", available: true})
	reg.Register("bad", &mockETFProvider{name: "bad", available: false})

	checks := reg.HealthCheck(context.Background())
	if len(checks) != 2 {
		t.Fatalf("expected 2 health checks, got %d", len(checks))
	}

	goodFound, badFound := false, false
	for _, c := range checks {
		switch c.Name {
		case "good":
			goodFound = true
			if !c.Available {
				t.Error("expected good provider to be available")
			}
		case "bad":
			badFound = true
			if c.Available {
				t.Error("expected bad provider to be unavailable")
			}
		}
	}
	if !goodFound || !badFound {
		t.Error("expected both providers in health check results")
	}

	reg.Unregister("good")
	reg.Unregister("bad")
}

func TestProviderType_String(t *testing.T) {
	tests := []struct {
		pt   ProviderType
		want string
	}{
		{ProviderTypeETF, "etf"},
		{ProviderTypeFX, "fx"},
		{ProviderTypeAShare, "ashare"},
		{ProviderTypePython, "python"},
		{ProviderType(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.pt.String(); got != tt.want {
			t.Errorf("ProviderType(%d).String() = %q, want %q", tt.pt, got, tt.want)
		}
	}
}

func TestETFAdapter(t *testing.T) {
	inner := &mockFullETFProvider{name: "finnhub", available: true, rateLimit: 10}
	adapter := NewETFAdapter(inner)

	if adapter.GetName() != "finnhub" {
		t.Errorf("expected finnhub, got %s", adapter.GetName())
	}
	if adapter.GetProviderType() != ProviderTypeETF {
		t.Errorf("expected ProviderTypeETF")
	}
	if !adapter.IsAvailable(context.Background()) {
		t.Error("expected available")
	}
	if adapter.GetRateLimit() != 10 {
		t.Errorf("expected rate limit 10, got %d", adapter.GetRateLimit())
	}
}

func TestFXAdapter(t *testing.T) {
	inner := &mockFullFXProvider{name: "openexchangerates", available: true, rateLimit: 30}
	adapter := NewFXAdapter(inner)

	if adapter.GetName() != "openexchangerates" {
		t.Errorf("expected openexchangerates, got %s", adapter.GetName())
	}
	if adapter.GetProviderType() != ProviderTypeFX {
		t.Errorf("expected ProviderTypeFX")
	}
	if !adapter.IsAvailable(context.Background()) {
		t.Error("expected available")
	}
}

func TestAShareAdapter(t *testing.T) {
	inner := &mockFullAShareProvider{name: "akshare", available: true, rateLimit: 5}
	adapter := NewAShareAdapter(inner)

	if adapter.GetName() != "akshare" {
		t.Errorf("expected akshare, got %s", adapter.GetName())
	}
	if adapter.GetProviderType() != ProviderTypeAShare {
		t.Errorf("expected ProviderTypeAShare")
	}
}

func TestPythonProxyAdapter(t *testing.T) {
	inner := &mockFullPythonService{name: "yfinance", available: true, rateLimit: 20}
	adapter := NewPythonProxyAdapter(inner)

	if adapter.GetName() != "yfinance" {
		t.Errorf("expected yfinance, got %s", adapter.GetName())
	}
	if adapter.GetProviderType() != ProviderTypePython {
		t.Errorf("expected ProviderTypePython")
	}
	if !adapter.IsAvailable(context.Background()) {
		t.Error("expected available")
	}
}

func TestUnifiedRegistry_Count(t *testing.T) {
	reg := GetUnifiedRegistry()
	before := reg.Count()

	reg.Register("count-test", &mockETFProvider{name: "count-test"})
	if reg.Count() != before+1 {
		t.Errorf("expected count %d, got %d", before+1, reg.Count())
	}

	reg.Unregister("count-test")
	if reg.Count() != before {
		t.Errorf("expected count %d, got %d", before, reg.Count())
	}
}
