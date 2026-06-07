package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDividendYield_Fallback(t *testing.T) {
	assert.Equal(t, 0.035, fallbackYields["SCHD"])
	assert.Equal(t, 0.095, fallbackYields["JEPQ"])
	assert.Equal(t, 0.075, fallbackYields["JEPI"])
	assert.Equal(t, 0.006, fallbackYields["QQQ"])
	assert.Equal(t, 0.015, fallbackYields["VTI"])
	assert.Equal(t, 0.013, fallbackYields["SPY"])
	assert.Equal(t, 0.028, fallbackYields["VYM"])
	assert.Equal(t, 0.030, fallbackYields["BND"])
}

func TestGetDividendYield_Default(t *testing.T) {
	_, ok := fallbackYields["UNKNOWN"]
	assert.False(t, ok)
}
