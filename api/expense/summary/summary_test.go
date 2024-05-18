package summary

import (
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/assert"
	"testing"
)

func TestSummary(t *testing.T) {
	testCases := []struct {
		name string
		data []Data
		want Summary
	}{
		{
			name: "empty data",
			data: []Data{},
			want: Summary{TotalAmount: 0, AveragePerDay: 0, CountTransaction: 0},
		},
		{
			name: "single data",
			data: []Data{
				{SumAmount: 10, CountExpenses: 1},
			},
			want: Summary{TotalAmount: 10, AveragePerDay: 10, CountTransaction: 1},
		},
		{
			name: "multiple data",
			data: []Data{
				{SumAmount: 20, CountExpenses: 2},
				{SumAmount: 30, CountExpenses: 3},
			},
			want: Summary{TotalAmount: 50, AveragePerDay: 25, CountTransaction: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := summary(tc.data)
			assert.Equal(t, tc.want, got)
		})
	}
}
