package cnf

import (
	"reflect"
	"testing"
)

func TestPackedClauseBridgeAndLift(t *testing.T) {
	data := new(Cnf_Dat_t)
	Cnf_DataSetPackedClauses(data, 3, []int32{0, 2, 5}, []int32{0, 3, 4, 1, 5})
	if Cnf_DataVarNum(data) != 3 || Cnf_DataClauseNum(data) != 2 {
		t.Fatalf("counts = (%d vars, %d clauses)", Cnf_DataVarNum(data), Cnf_DataClauseNum(data))
	}
	if got, want := Cnf_DataClauses(data), [][]int32{{0, 3}, {4, 1, 5}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("clauses = %v, want %v", got, want)
	}
	Cnf_DataLift(data, 3)
	if got, want := Cnf_DataClause(data, 0), []int32{6, 9}; !reflect.DeepEqual(got, want) {
		t.Fatalf("lifted first clause = %v, want %v", got, want)
	}
	if got, want := Cnf_DataClause(data, 1), []int32{10, 7, 11}; !reflect.DeepEqual(got, want) {
		t.Fatalf("lifted second clause = %v, want %v", got, want)
	}
}
