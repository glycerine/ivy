package core

import "testing"

func TestInitLinearMatrixIdentityAndMemoryAccounting(t *testing.T) {
	dd := NewBDDManager(65)
	before := Cudd_ReadMemoryInUse(dd)
	if !Cudd_InitLinearMatrix(dd) {
		t.Fatal("Cudd_InitLinearMatrix failed")
	}
	words, size := Cudd_LinearMatrixState(dd)
	if size != 65 {
		t.Fatalf("matrix size = %d, want 65", size)
	}
	if len(words) != 130 {
		t.Fatalf("matrix words = %d, want 130", len(words))
	}
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			got := words[row*2+col/64]&(uint64(1)<<uint(col%64)) != 0
			if got != (row == col) {
				t.Fatalf("matrix[%d,%d] = %t, want %t", row, col, got, row == col)
			}
		}
	}
	if got, want := Cudd_ReadMemoryInUse(dd)-before, uint64(130*8); got != want {
		t.Fatalf("memory accounting delta = %d, want %d", got, want)
	}
}

func TestInitLinearMatrixFailureSetsMemoryError(t *testing.T) {
	dd := NewBDDManager(1)
	Cudd_Quit(dd)
	if Cudd_InitLinearMatrix(dd) {
		t.Fatal("initializing a closed manager should fail")
	}
	if got := Cudd_ReadErrorCode(dd); got != CUDD_MEMORY_OUT {
		t.Fatalf("error code = %v, want CUDD_MEMORY_OUT", got)
	}
}
