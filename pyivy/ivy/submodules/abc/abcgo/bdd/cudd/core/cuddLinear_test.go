package core

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

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

func TestLinearReadAndPrintMatchPacked64BitLayout(t *testing.T) {
	dd := NewBDDManager(2)
	if !Cudd_InitLinearMatrix(dd) {
		t.Fatal("Cudd_InitLinearMatrix failed")
	}
	if Cudd_ReadLinear(dd, 0, 0) != 1 || Cudd_ReadLinear(dd, 0, 1) != 0 || Cudd_ReadLinear(dd, 1, 1) != 1 {
		t.Fatal("linear matrix identity bits were read incorrectly")
	}
	var out bytes.Buffer
	Cudd_SetOutput(dd, &out, nil)
	if got := Cudd_PrintLinear(dd); got != 1 {
		t.Fatalf("Cudd_PrintLinear returned %d, want 1", got)
	}
	if got, want := out.String(), "1"+strings.Repeat("0", 63)+"\n01"+strings.Repeat("0", 62)+"\n"; got != want {
		t.Fatalf("printed matrix = %q, want %q", got, want)
	}
}

type linearFailWriter struct{}

func (linearFailWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestLinearPrintPropagatesWriterFailure(t *testing.T) {
	dd := NewBDDManager(1)
	if !Cudd_InitLinearMatrix(dd) {
		t.Fatal("Cudd_InitLinearMatrix failed")
	}
	Cudd_SetOutput(dd, linearFailWriter{}, nil)
	if got := Cudd_PrintLinear(dd); got != 0 {
		t.Fatalf("Cudd_PrintLinear returned %d after writer failure, want 0", got)
	}
}

func TestResizeLinearMatrixAddsIdentityRow(t *testing.T) {
	dd := NewBDDManager(2)
	if !Cudd_InitLinearMatrix(dd) || Cudd_bddIthVar(dd, 2) == nil {
		t.Fatal("failed to initialize matrix or add variable")
	}
	if !Cudd_ResizeLinearMatrix(dd) {
		t.Fatal("Cudd_ResizeLinearMatrix failed")
	}
	if got := Cudd_ReadLinear(dd, 2, 2); got != 1 {
		t.Fatalf("new diagonal entry = %d, want 1", got)
	}
	if got := Cudd_ReadLinear(dd, 0, 0); got != 1 {
		t.Fatalf("preserved diagonal entry = %d, want 1", got)
	}
}

func TestXorLinearRows(t *testing.T) {
	dd := NewBDDManager(70)
	if !Cudd_InitLinearMatrix(dd) {
		t.Fatal("Cudd_InitLinearMatrix failed")
	}
	Cudd_XorLinearRows(dd, 2, 67)
	for _, bit := range []struct{ row, col, want int }{
		{2, 2, 1}, {2, 67, 1}, {67, 67, 1}, {67, 2, 0},
	} {
		if got := Cudd_ReadLinear(dd, bit.row, bit.col); got != bit.want {
			t.Errorf("Cudd_ReadLinear(%d, %d) = %d, want %d", bit.row, bit.col, got, bit.want)
		}
	}
}
