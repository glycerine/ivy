package acec

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// Expected values come from test_vectors/acec_polynomial_order.c linked with ABC.
func TestSourcePolynomialAdderOrder(t *testing.T) {
	p := reviewAdderChain(-1)
	want := []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 0, 0, 0}
	got := vec.Vec_IntArray(Gia_PolynReorder(p, 0, 0))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Go order %v, C order %v", got, want)
	}
}
