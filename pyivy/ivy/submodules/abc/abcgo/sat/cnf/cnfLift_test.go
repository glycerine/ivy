package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
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

func TestLiftAndFlipOnlySelectedLiterals(t *testing.T) {
	data := &CNF{Clauses: [][]int32{{0, 3}, {4, 7}}}
	flips := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(flips, 1)
	vec.Vec_IntPush(flips, 2)
	Cnf_DataLiftAndFlipLits(data, 3, flips)
	want := [][]int32{{0, 8}, {11, 7}}
	if !reflect.DeepEqual(data.Clauses, want) || !reflect.DeepEqual(data.pClauses, want) {
		t.Fatalf("clauses = %v / %v, want %v", data.Clauses, data.pClauses, want)
	}
}

func TestDataLiftPreservesLocalCountAndReverses(t *testing.T) {
	data := &CNF{Variables: 2, nVars: 2, VarNums: []int32{-1, 0, 1}, Clauses: [][]int32{{0, 3}}}
	Cnf_DataLift(data, 4)
	if data.Variables != 2 || data.nVars != 2 {
		t.Fatalf("lift changed local variable count to %d/%d", data.Variables, data.nVars)
	}
	Cnf_DataLift(data, -4)
	if !reflect.DeepEqual(data.VarNums, []int32{-1, 0, 1}) || !reflect.DeepEqual(data.Clauses, [][]int32{{0, 3}}) {
		t.Fatalf("lift round trip changed mapping or clauses: %v %v", data.VarNums, data.Clauses)
	}
}
