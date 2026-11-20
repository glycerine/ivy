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

func TestPackedClauseMemoryAndOwnership(t *testing.T) {
	const count = 512
	offsets := make([]int32, count+1)
	lits := make([]int32, 2*count)
	src := make([][]int32, count)
	for i := range src {
		offsets[i+1] = int32(2*i + 2)
		lits[2*i], lits[2*i+1] = int32(i*2), int32(i*2+1)
		src[i] = lits[2*i : 2*i+2]
	}
	var dup [][]int32
	var data Cnf_Dat_t
	for _, tc := range []struct {
		name string
		run  func()
	}{
		{"clone", func() { dup = cloneClauses(src) }},
		{"import", func() { Cnf_DataSetPackedClauses(&data, count, offsets, lits) }},
	} {
		if allocs := testing.AllocsPerRun(10, tc.run); allocs > 2 {
			t.Errorf("%s allocated %.0f objects", tc.name, allocs)
		}
	}
	for _, clauses := range [][][]int32{dup, data.Clauses} {
		clauses[0][0] = 99
		clauses[0] = append(clauses[0], 77)
		if clauses[1][0] != 2 || lits[0] != 0 {
			t.Fatal("clause storage aliases input or adjacent clause")
		}
	}
}

func BenchmarkCopyPackedClauses(b *testing.B) {
	const count = 8192
	offsets := make([]int32, count+1)
	lits := make([]int32, 3*count)
	for i := 1; i < len(offsets); i++ {
		offsets[i] = int32(3 * i)
	}
	b.ReportAllocs()
	for b.Loop() {
		data := new(Cnf_Dat_t)
		Cnf_DataSetPackedClauses(data, 16, offsets, lits)
	}
}

func loaderBenchmarkCNF() *Cnf_Dat_t {
	const variables, count = 2048, 8192
	offsets := make([]int32, count+1)
	lits := make([]int32, 3*count)
	for i := 0; i < count; i++ {
		v, group := i%variables, i/variables
		offsets[i+1] = int32(3 * (i + 1))
		lits[3*i] = int32(2 * v)
		lits[3*i+1] = int32(2*((v+17)%variables) + (group & 1))
		lits[3*i+2] = int32(2*((v+71)%variables) + (group >> 1))
	}
	data := new(Cnf_Dat_t)
	Cnf_DataSetPackedClauses(data, variables, offsets, lits)
	return data
}

func BenchmarkLoadTernaryClauses(b *testing.B) {
	data := loaderBenchmarkCNF()
	b.ReportAllocs()
	for b.Loop() {
		s := Cnf_DataWriteIntoSolver(data, 1, 0)
		if s == nil || s.NumVars() != 2048 {
			b.Fatal("loading SAT clauses failed")
		}
		s.Close()
	}
}

func BenchmarkDuplicateTernaryClauses(b *testing.B) {
	data := loaderBenchmarkCNF()
	b.ReportAllocs()
	for b.Loop() {
		d := Cnf_DataDup(data)
		if d.nLiterals != 24576 || d.nClauses != 8192 {
			b.Fatal("invalid duplicate dimensions")
		}
		Cnf_DataFree(d)
	}
}
