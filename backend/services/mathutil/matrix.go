package mathutil

import "math"

func MatrixInverse(mat [][]float64) ([][]float64, error) {
	n := len(mat)
	if n == 0 {
		return nil, ErrEmptyMatrix
	}
	augmented := make([][]float64, n)
	for i := range mat {
		if len(mat[i]) != n {
			return nil, ErrNotSquareMatrix
		}
		augmented[i] = make([]float64, 2*n)
		copy(augmented[i], mat[i])
		augmented[i][n+i] = 1
	}
	for i := 0; i < n; i++ {
		pivot := augmented[i][i]
		if math.Abs(pivot) < 1e-10 {
			return nil, ErrSingularMatrix
		}
		for j := 0; j < 2*n; j++ {
			augmented[i][j] /= pivot
		}
		for k := 0; k < n; k++ {
			if k == i {
				continue
			}
			factor := augmented[k][i]
			for j := 0; j < 2*n; j++ {
				augmented[k][j] -= factor * augmented[i][j]
			}
		}
	}
	inv := make([][]float64, n)
	for i := range inv {
		inv[i] = augmented[i][n:]
	}
	return inv, nil
}

func MatrixMultiply(a, b [][]float64) [][]float64 {
	m, n := len(a), len(a[0])
	p := len(b[0])
	result := make([][]float64, m)
	for i := range result {
		result[i] = make([]float64, p)
		for j := 0; j < p; j++ {
			for k := 0; k < n; k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}

func MatrixTranspose(mat [][]float64) [][]float64 {
	if len(mat) == 0 {
		return mat
	}
	m, n := len(mat), len(mat[0])
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			result[i][j] = mat[j][i]
		}
	}
	return result
}

func MatrixVectorMultiply(mat [][]float64, vec []float64) []float64 {
	m := len(mat)
	result := make([]float64, m)
	for i := 0; i < m; i++ {
		for j := 0; j < len(vec); j++ {
			result[i] += mat[i][j] * vec[j]
		}
	}
	return result
}
