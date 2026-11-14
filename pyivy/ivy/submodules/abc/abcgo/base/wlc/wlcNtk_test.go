package wlc

import (
	"reflect"
	"testing"
)

func TestVecWrdSelectSortCost2(t *testing.T) {
	values := []uint64{12, 22, 31, 42}
	costs := []uint64{3, 1, 2, 1}
	Vec_WrdSelectSortCost2(values, len(values), costs)
	if want := []uint64{22, 42, 31, 12}; !reflect.DeepEqual(values, want) {
		t.Fatalf("sorted values = %v, want %v", values, want)
	}
	if want := []uint64{1, 1, 2, 3}; !reflect.DeepEqual(costs, want) {
		t.Fatalf("sorted costs = %v, want %v", costs, want)
	}
}
