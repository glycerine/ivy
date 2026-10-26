package abc

// Standalone contract, ownership, concurrency and fuzz tests for AbcApi.cpp.

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func TestABCCallMiniSat22ExactStubContract(t *testing.T) {
	if ABCMiniSat22Unavailable != -1 {
		t.Fatalf("native result constant is %d", ABCMiniSat22Unavailable)
	}
	if got := ABCCallMiniSat22(nil); got != ABCMiniSat22Unavailable {
		t.Fatalf("nil call returned %d", got)
	}
	x, _ := MakeSatLit(0, false)
	notX := x.Not()
	formula, err := NewCNFFormula(1, [][]SatLit{{x}, {notX}}, CNFTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	before := formula.Clauses()
	if got := ABCCallMiniSat22(formula); got != ABCMiniSat22Unavailable {
		t.Fatalf("formula call returned %d", got)
	}
	if formula.VariableCount() != 1 || !slices.EqualFunc(formula.Clauses(), before, slices.Equal[[]SatLit]) {
		t.Fatalf("stub mutated its ignored CNF: variables=%d clauses=%v", formula.VariableCount(), formula.Clauses())
	}
}

func TestABCCallMiniSat22Concurrent(t *testing.T) {
	formula, err := NewCNFFormula(2, nil, CNFTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 32; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for call := 0; call < 128; call++ {
				if ABCCallMiniSat22(formula) != ABCMiniSat22Unavailable {
					failures.Add(1)
				}
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent calls returned %d unexpected results", failures.Load())
	}
}

func FuzzABCCallMiniSat22(f *testing.F) {
	f.Add(uint8(0), []byte(nil))
	f.Add(uint8(3), []byte{0, 1, 2, 3, 4})
	f.Fuzz(func(t *testing.T, variableByte uint8, data []byte) {
		variables := int(variableByte % 17)
		clauses := make([][]SatLit, 0, min(len(data), 32))
		if variables > 0 {
			for index, value := range data {
				if index == 32 {
					break
				}
				literal, err := MakeSatLit(SatVar(int(value)%variables), value&1 != 0)
				if err != nil {
					t.Fatal(err)
				}
				clauses = append(clauses, []SatLit{literal})
			}
		}
		formula, err := NewCNFFormula(variables, clauses, CNFTransformOptions{})
		if err != nil {
			t.Fatal(err)
		}
		before := formula.Clauses()
		if got := ABCCallMiniSat22(formula); got != ABCMiniSat22Unavailable {
			t.Fatalf("returned %d", got)
		}
		if formula.VariableCount() != variables || !slices.EqualFunc(formula.Clauses(), before, slices.Equal[[]SatLit]) {
			t.Fatal("stub mutated its ignored CNF")
		}
	})
}
