package statistics

import "math"

func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// SampleVariance computes sample variance (divides by N-1).
func SampleVariance(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return 0
	}
	mean := Mean(values)
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(n-1)
}

func SampleStdDev(values []float64) float64 {
	return math.Sqrt(SampleVariance(values))
}

// PopulationVariance computes population variance (divides by N).
func PopulationVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := Mean(values)
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(values))
}

// NormalCDF computes the standard normal CDF using the Hastings approximation
// (error < 1.5e-7).
// Reference: Hastings, C. (1955). Approximations for Digital Computers.
func NormalCDF(x float64) float64 {
	if x < -8 {
		return 0
	}
	if x > 8 {
		return 1
	}
	const (
		b1 = 0.319381530
		b2 = -0.356563782
		b3 = 1.781477937
		b4 = -1.821255978
		b5 = 1.330274429
		p  = 0.2316419
	)
	sign := 1.0
	if x < 0 {
		sign = -1.0
		x = -x
	}
	t := 1.0 / (1.0 + p*x)
	poly := ((((b5*t+b4)*t)+b3)*t+b2)*t + b1
	// φ(x) = (1/√(2π)) × exp(-x²/2)
	phiX := (1.0 / math.Sqrt(2*math.Pi)) * math.Exp(-x*x/2)
	// Φ(x) = 1 - φ(x) × poly × t   (for x ≥ 0)
	return 0.5 + sign*(0.5-phiX*poly*t)
}

// NormalPDF computes the standard normal probability density function.
func NormalPDF(x float64) float64 {
	return (1.0 / math.Sqrt(2*math.Pi)) * math.Exp(-x*x/2)
}

// NormalQuantile computes the standard normal quantile (inverse CDF) using the
// Beasley-Springer-Moro algorithm.
func NormalQuantile(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	if p < 0.5 {
		return -rationalApprox(math.Sqrt(-2 * math.Log(p)))
	}
	return rationalApprox(math.Sqrt(-2 * math.Log(1-p)))
}

func rationalApprox(t float64) float64 {
	const (
		c0 = 2.515517
		c1 = 0.802853
		c2 = 0.010328
		d1 = 1.432788
		d2 = 0.189269
		d3 = 0.001308
	)
	return t - (c0+c1*t+c2*t*t)/(1+d1*t+d2*t*t+d3*t*t*t)
}
