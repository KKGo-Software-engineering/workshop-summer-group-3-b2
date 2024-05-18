package summary

import (
	"testing"
)

func TestSummary(t *testing.T) {
	testCases := []struct {
		name string
		data []Data
		want Summary
	}{
		{
			name: "EmptyData",
			data: []Data{},
			want: Summary{Total: 0, Average: 0, Count: 0},
		},
		{
			name: "SingleData",
			data: []Data{
				{SumAmount: 10, CountExpenses: 1},
			},
			want: Summary{Total: 10, Average: 10, Count: 1},
		},
		{
			name: "MultipleData",
			data: []Data{
				{SumAmount: 20, CountExpenses: 2},
				{SumAmount: 30, CountExpenses: 3},
			},
			want: Summary{Total: 50, Average: 25, Count: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := summary(tc.data)
			if got != tc.want {
				t.Errorf("summary(%v) = %v; want %v", tc.data, got, tc.want)
			}
		})
	}
}
