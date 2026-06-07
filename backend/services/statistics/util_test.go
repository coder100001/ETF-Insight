package statistics

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMean(t *testing.T) {
	assert.Equal(t, 3.0, Mean([]float64{1, 2, 3, 4, 5}))
	assert.Equal(t, 0.0, Mean([]float64{}))
}

func TestSampleVariance(t *testing.T) {
	// Known data: [2, 4, 4, 4, 5, 5, 7, 9], sample variance = 4.5714
	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	v := SampleVariance(data)
	assert.InDelta(t, 4.5714, v, 0.001)
}

func TestSampleVariance_PopulationDiff(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	sample := SampleVariance(data)  // /4
	pop := PopulationVariance(data) // /5
	assert.True(t, sample > pop)
	assert.InDelta(t, 2.5, sample, 0.001)
	assert.InDelta(t, 2.0, pop, 0.001)
}

func TestSampleVariance_SingleElement(t *testing.T) {
	assert.Equal(t, 0.0, SampleVariance([]float64{5}))
}

func TestPopulationVariance_Empty(t *testing.T) {
	assert.Equal(t, 0.0, PopulationVariance([]float64{}))
}

func TestSampleStdDev(t *testing.T) {
	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	std := SampleStdDev(data)
	assert.InDelta(t, math.Sqrt(4.5714), std, 0.001)
}

func TestNormalCDF(t *testing.T) {
	tests := []struct{ x, expected float64 }{
		{0, 0.5},
		{1.645, 0.95},
		{-1.645, 0.05},
		{2.326, 0.99},
		{-2.326, 0.01},
		{1.0, 0.8413},
	}
	for _, tt := range tests {
		result := NormalCDF(tt.x)
		assert.InDelta(t, tt.expected, result, 0.0001,
			"NormalCDF(%f) = %f, want %f", tt.x, result, tt.expected)
	}
}

func TestNormalCDF_Extremes(t *testing.T) {
	assert.Equal(t, 0.0, NormalCDF(-10))
	assert.Equal(t, 1.0, NormalCDF(10))
}

func TestNormalPDF(t *testing.T) {
	// phi(0) = 1/sqrt(2*pi) ~ 0.3989
	assert.InDelta(t, 0.3989, NormalPDF(0), 0.0001)
	// phi(1) ~ 0.2420
	assert.InDelta(t, 0.2420, NormalPDF(1), 0.0001)
}

func TestNormalQuantile(t *testing.T) {
	tests := []struct{ p, expected float64 }{
		{0.5, 0},
		{0.95, 1.645},
		{0.05, -1.645},
		{0.975, 1.96},
	}
	for _, tt := range tests {
		result := NormalQuantile(tt.p)
		assert.InDelta(t, tt.expected, result, 0.01,
			"NormalQuantile(%f) = %f, want %f", tt.p, result, tt.expected)
	}
}

func TestNormalQuantile_Invalid(t *testing.T) {
	assert.Equal(t, 0.0, NormalQuantile(0))
	assert.Equal(t, 0.0, NormalQuantile(1))
	assert.Equal(t, 0.0, NormalQuantile(-0.1))
}

func TestNormalQuantile_InverseOfCDF(t *testing.T) {
	probs := []float64{0.01, 0.05, 0.10, 0.25, 0.5, 0.75, 0.90, 0.95, 0.99}
	for _, p := range probs {
		q := NormalQuantile(p)
		cdf := NormalCDF(q)
		assert.InDelta(t, p, cdf, 0.001,
			"CDF(Quantile(%f)) = CDF(%f) = %f, want %f", p, q, cdf, p)
	}
}
