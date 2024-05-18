package summary

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSummary(t *testing.T) {
	testCases := []struct {
		name string
		data []RawData
		want Summary
	}{
		{
			name: "empty data",
			data: []RawData{},
			want: Summary{Total: 0, Average: 0, Count: 0},
		},
		{
			name: "single data",
			data: []RawData{
				{SumAmount: 10, CountExpenses: 1},
			},
			want: Summary{Total: 10, Average: 10, Count: 1},
		},
		{
			name: "multiple data",
			data: []RawData{
				{SumAmount: 20, CountExpenses: 2},
				{SumAmount: 30, CountExpenses: 3},
			},
			want: Summary{Total: 50, Average: 25, Count: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := summary(tc.data)
			assert.Equal(t, tc.want, got)
		})
	}
}
