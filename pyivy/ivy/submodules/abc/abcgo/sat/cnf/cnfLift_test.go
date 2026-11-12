package cnf

import (
	"reflect"
	"testing"
)

func TestLiftVariableOffsetsPreservesLocalVariableCount(t *testing.T) {
	data := &CNF{
		Variables: 2,
		nVars:     2,
		VarNums:   []int32{0, 1},
		pVarNums:  []int32{0, 1},
		Clauses:   [][]int32{{0, 3}},
		pClauses:  [][]int32{{0, 3}},
	}
	data.LiftVariableOffsets(4)
	if data.Variables != 2 || data.nVars != 2 {
		t.Fatalf("local variable count changed: Variables=%d nVars=%d", data.Variables, data.nVars)
	}
	if !reflect.DeepEqual(data.VarNums, []int32{4, 5}) || !reflect.DeepEqual(data.pVarNums, []int32{4, 5}) {
		t.Fatalf("variable maps = %v / %v, want [4 5]", data.VarNums, data.pVarNums)
	}
	wantClauses := [][]int32{{8, 11}}
	if !reflect.DeepEqual(data.Clauses, wantClauses) || !reflect.DeepEqual(data.pClauses, wantClauses) {
		t.Fatalf("clauses = %v / %v, want %v", data.Clauses, data.pClauses, wantClauses)
	}
}
