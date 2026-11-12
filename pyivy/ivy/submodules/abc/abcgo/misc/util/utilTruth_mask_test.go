package util

import (
	"strconv"
	"testing"
)

func TestSmallTruthVariableMasks(t *testing.T) {
	for variable := int32(0); variable < 6; variable++ {
		got := ttMask(variable)
		var expected word
		for assignment := int32(0); assignment < 64; assignment++ {
			if assignment&(1<<variable) != 0 {
				expected |= word(1) << assignment
			}
		}
		if got != expected {
			t.Fatalf("variable%d mask %016x want%016x", variable, got, expected)
		}
		// Returning one variable's mask must not change another lookup.
		if ttMask(variable) != got {
			t.Fatal("variable mask changed")
		}
	}
}
func TestSmallTruthVariableMaskBounds(t *testing.T) {
	for _, variable := range []int32{-2147483648, -1, 6, 7, 2147483647} {
		t.Run(strconv.FormatInt(int64(variable), 10), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("out-of-range variable %d accepted", variable)
				}
			}()
			_ = ttMask(variable)
		})
	}
}

var truthMaskBenchmarkSink word

func BenchmarkSmallTruthVariableMask(b *testing.B) {
	var result word
	for i := 0; i < b.N; i++ {
		result ^= ttMask(int32(i % 6))
	}
	truthMaskBenchmarkSink = result
}
