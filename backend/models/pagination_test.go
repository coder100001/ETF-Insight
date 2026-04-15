package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationQuery_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int
	}{
		{
			name:     "Page 1",
			page:     1,
			pageSize: 20,
			expected: 0,
		},
		{
			name:     "Page 2",
			page:     2,
			pageSize: 20,
			expected: 20,
		},
		{
			name:     "Page 5",
			page:     5,
			pageSize: 10,
			expected: 40,
		},
		{
			name:     "Zero page",
			page:     0,
			pageSize: 20,
			expected: 0,
		},
		{
			name:     "Negative page",
			page:     -1,
			pageSize: 20,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PaginationQuery{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}
			assert.Equal(t, tt.expected, p.GetOffset())
		})
	}
}

func TestPaginationQuery_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		expected int
	}{
		{
			name:     "Default size",
			pageSize: 0,
			expected: 20,
		},
		{
			name:     "Normal size",
			pageSize: 10,
			expected: 10,
		},
		{
			name:     "Maximum size",
			pageSize: 100,
			expected: 100,
		},
		{
			name:     "Exceed maximum",
			pageSize: 150,
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PaginationQuery{
				PageSize: tt.pageSize,
			}
			assert.Equal(t, tt.expected, p.GetLimit())
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	page := 1
	pageSize := 10
	total := int64(25)

	response := NewPaginatedResponse(data, page, pageSize, total)

	assert.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, data, response.Data)
	assert.Equal(t, page, response.Pagination.Page)
	assert.Equal(t, pageSize, response.Pagination.PageSize)
	assert.Equal(t, total, response.Pagination.Total)
	assert.Equal(t, 3, response.Pagination.TotalPages)
	assert.True(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)
}

func TestNewPaginatedResponse_LastPage(t *testing.T) {
	data := []string{"item1", "item2"}
	page := 3
	pageSize := 10
	total := int64(22)

	response := NewPaginatedResponse(data, page, pageSize, total)

	assert.Equal(t, 3, response.Pagination.TotalPages)
	assert.False(t, response.Pagination.HasNext)
	assert.True(t, response.Pagination.HasPrev)
}

func TestNewPaginatedResponse_SinglePage(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	page := 1
	pageSize := 10
	total := int64(3)

	response := NewPaginatedResponse(data, page, pageSize, total)

	assert.Equal(t, 1, response.Pagination.TotalPages)
	assert.False(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)
}

func TestNewPaginatedResponse_EmptyData(t *testing.T) {
	data := []string{}
	page := 1
	pageSize := 10
	total := int64(0)

	response := NewPaginatedResponse(data, page, pageSize, total)

	assert.Equal(t, 0, response.Pagination.TotalPages)
	assert.False(t, response.Pagination.HasNext)
	assert.False(t, response.Pagination.HasPrev)
}
