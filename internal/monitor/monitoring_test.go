package monitor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAvg тестирует функцию AvgFloat,
// которая получает список float32 и отдает среднее значение.
func TestAvg(t *testing.T) {
	tests := []struct {
		input  []float32
		result float32
	}{
		{
			input:  []float32{3.0, 3.0, 3.0},
			result: 3.0,
		},
		{
			input:  []float32{1.0, 2.0, 3.0},
			result: 2.0,
		},
		{
			input:  []float32{33.69, 77.53, 123.12},
			result: 78.113335,
		},
		{
			input:  []float32{0.0, 0.0, 3.0},
			result: 1.0,
		},
		{
			input:  []float32{-3.0, 0.0, 3.0},
			result: 0,
		},
		{
			input:  []float32{8.5, 10.5, 55.4, 65.3, 15.3, 0.0, 20.0},
			result: 25.000002, // mantisa ?
		},
		{
			input:  []float32{-4.0, 1.0, 1.0},
			result: -0.6666667,
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test-%d", i), func(t *testing.T) {
			require.Equal(t, tt.result, AvgFloat(tt.input...))
		})
	}
}
