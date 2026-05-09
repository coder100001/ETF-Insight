package mathutil

import (
	"math"
	"testing"
)

func TestMatrixInverse(t *testing.T) {
	mat := [][]float64{
		{4, 7},
		{2, 6},
	}
	inv, err := MatrixInverse(mat)
	if err != nil {
		t.Fatalf("MatrixInverse failed: %v", err)
	}
	result := MatrixMultiply(mat, inv)
	expected := [][]float64{
		{1, 0},
		{0, 1},
	}
	for i := range result {
		for j := range result[i] {
			if math.Abs(result[i][j]-expected[i][j]) > 1e-10 {
				t.Errorf("MatrixInverse: [%d][%d] = %f, expected %f", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

func TestMatrixInverse_Singular(t *testing.T) {
	mat := [][]float64{
		{1, 2},
		{2, 4},
	}
	_, err := MatrixInverse(mat)
	if err == nil {
		t.Fatal("Expected error for singular matrix, got nil")
	}
}

func TestMatrixTranspose(t *testing.T) {
	mat := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}
	result := MatrixTranspose(mat)
	if len(result) != 3 || len(result[0]) != 2 {
		t.Fatalf("Transpose dimensions: got %dx%d, expected 3x2", len(result), len(result[0]))
	}
	if result[0][0] != 1 || result[0][1] != 4 {
		t.Errorf("Transpose[0] = %v, expected [1, 4]", result[0])
	}
	if result[2][0] != 3 || result[2][1] != 6 {
		t.Errorf("Transpose[2] = %v, expected [3, 6]", result[2])
	}
}

func TestMatrixMultiply(t *testing.T) {
	a := [][]float64{
		{1, 2},
		{3, 4},
	}
	b := [][]float64{
		{5, 6},
		{7, 8},
	}
	result := MatrixMultiply(a, b)
	expected := [][]float64{
		{19, 22},
		{43, 50},
	}
	for i := range result {
		for j := range result[i] {
			if result[i][j] != expected[i][j] {
				t.Errorf("Multiply[%d][%d] = %f, expected %f", i, j, result[i][j], expected[i][j])
			}
		}
	}
}

func TestMatrixVectorMultiply(t *testing.T) {
	mat := [][]float64{
		{1, 2},
		{3, 4},
	}
	vec := []float64{5, 6}
	result := MatrixVectorMultiply(mat, vec)
	expected := []float64{17, 39}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("MatrixVectorMultiply[%d] = %f, expected %f", i, result[i], expected[i])
		}
	}
}

func TestMatrixInverse_EmptyMatrix(t *testing.T) {
	_, err := MatrixInverse([][]float64{})
	if err == nil {
		t.Fatal("Expected error for empty matrix, got nil")
	}
}

func TestMatrixInverse_NonSquare(t *testing.T) {
	_, err := MatrixInverse([][]float64{{1, 2, 3}, {4, 5, 6}})
	if err == nil {
		t.Fatal("Expected error for non-square matrix, got nil")
	}
}