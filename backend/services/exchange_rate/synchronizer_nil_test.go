package exchange_rate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExchangeRateService_UpdateRates_NilDB_ShouldNotPanic(t *testing.T) {
	service := NewExchangeRateService(nil)

	err := service.UpdateRates()

	assert.NoError(t, err, "UpdateRates with nil service should not panic, got error: %v", err)
	t.Logf("✅ UpdateRates completed without panic: %v", err)
}

func TestExchangeRateService_MultipleUpdateRatesCalls(t *testing.T) {
	service := NewExchangeRateService(nil)

	for i := 0; i < 3; i++ {
		err := service.UpdateRates()
		if err != nil {
			t.Logf("Call %d: Expected error (acceptable): %v", i+1, err)
		} else {
			t.Logf("Call %d: No error (gracefully handled)", i+1)
		}
	}

	t.Log("✅ Multiple calls completed without panic")
}
