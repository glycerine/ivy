package bmc

import "testing"

func TestTernaryNotTruthTable(t *testing.T) {
	want := map[int32]int32{absZero: absOne, absOne: absZero, absUnknown: absUnknown}
	for input, output := range want {
		if got := ternaryNot(input); got != output {
			t.Errorf("ternaryNot(%d) = %d, want %d", input, got, output)
		}
	}
}

func TestTernaryAndTruthTable(t *testing.T) {
	states := []int32{absZero, absOne, absUnknown}
	want := [3][3]int32{
		{absZero, absZero, absZero},
		{absZero, absOne, absUnknown},
		{absZero, absUnknown, absUnknown},
	}
	for i, left := range states {
		for j, right := range states {
			if got := ternaryAnd(left, right); got != want[i][j] {
				t.Errorf("ternaryAnd(%d, %d) = %d, want %d", left, right, got, want[i][j])
			}
		}
	}
}

func TestTernaryHelpersRejectInvalidState(t *testing.T) {
	for _, value := range []int32{0, 4} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("ternaryNot(%d) did not reject invalid state", value)
				}
			}()
			ternaryNot(value)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("ternaryAnd(%d, %d) did not reject invalid state", value, absZero)
				}
			}()
			ternaryAnd(value, absZero)
		}()
	}
}
