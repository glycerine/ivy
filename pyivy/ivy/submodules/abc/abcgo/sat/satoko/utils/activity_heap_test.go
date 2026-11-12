package utils

import (
	"github.com/glycerine/abcgo/sat/satoko/utils/vec"
	"math"
	"testing"
)

func TestActivityArithmetic(t *testing.T) {
	for _, a := range []float64{1, 1.5, 2, 3.5, 100, 1e30} {
		for _, b := range []float64{1, 2, 7} {
			x, y := ActivityFromFloat(a), ActivityFromFloat(b)
			for _, pair := range [][2]float64{{ActivityFloat(ActivityAdd(x, y)), a + b}, {ActivityFloat(ActivityMultiply(x, y)), a * b}} {
				if math.Abs(pair[0]-pair[1])/pair[1] > 1e-12 {
					t.Fatal(a, b, pair)
				}
			}
		}
	}
	if ActivityRescale(SDBL_CONST1, 200) != SDBL_CONST1 {
		t.Fatal("activity lower bound")
	}
}
func TestHeapUpdates(t *testing.T) {
	weights := vec.NewActivities(0)
	for _, w := range []uint64{2, 8, 3, 4} {
		weights.Push(w)
	}
	h := NewHeap(weights)
	for i := uint32(0); i < 4; i++ {
		h.Insert(i)
	}
	if h.Pop() != 1 {
		t.Fatal("max weight")
	}
	weights.Set(0, 9)
	h.Update(0)
	if h.Pop() != 0 {
		t.Fatal("updated weight")
	}
	weights.Set(3, 1)
	h.Update(3)
	if h.Pop() != 2 || h.Pop() != 3 {
		t.Fatal("decrease")
	}
	h.Build([]uint32{3, 0})
	if h.Pop() != 0 {
		t.Fatal("build")
	}
}
