package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestFaultAddOne(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	for _, fixed := range []int32{0, 1} {
		for _, positive := range []int32{0, 1} {
			enc, err := gia.Mf_ManGenerateCnfResult(p, 6, true, false, false)
			if err != nil {
				t.Fatal(err)
			}
			base := &cnf.Cnf_Dat_t{Variables: enc.Variables, Clauses: enc.Clauses, VarNums: enc.VarNums}
			s := bsat.NewSolver()
			for i := 0; i < enc.Variables; i++ {
				s.NewVariable()
			}
			for _, cl := range enc.Clauses {
				if !s.AddClause(cl...) {
					t.Fatal("base encoding")
				}
			}
			values := vec.Vec_IntAllocArrayCopy([]int32{fixed}, 1)
			result := Gia_ManFaultAddOne(p, base, s, values, 1, positive, p)
			if fixed == 0 && positive == 1 {
				if result != 0 && s.Solve() != bsat.Unsatisfiable {
					t.Fatal("false cofactor must be unsatisfiable")
				}
				continue
			}
			if result == 0 {
				t.Fatalf("unexpected inconsistency")
			}
			variable := base.VarNums[b>>1]
			for _, v := range []bool{false, true} {
				lit := bsat.Negative(variable)
				if v {
					lit = bsat.Positive(variable)
				}
				want := (fixed != 0 && v) == (positive != 0)
				if got := s.Solve(lit) == bsat.Satisfiable; got != want {
					t.Fatalf("fixed=%d positive=%d b=%v got=%v enc=%v", fixed, positive, v, got, enc.Clauses)
				}
			}
		}
	}
}

func TestDumpUntestableFaults(t *testing.T) {
	p := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	data := &cnf.Cnf_Dat_t{VarNums: make([]int32, gia.Gia_ManObjNum(p))}
	data.VarNums[a>>1] = 0
	data.VarNums[b>>1] = 1
	s := bsat.NewSolver()
	s.NewVariable()
	s.NewVariable()
	s.AddClause(bsat.Positive(0), bsat.Positive(1))
	name := filepath.Join(t.TempDir(), "faults")
	Gia_ManDumpUntests(p, data, s, 1, name, 0)
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSpace(string(content)), "\n")
	for i := range rows {
		rows[i] = strings.TrimSpace(rows[i])
	}
	sort.Strings(rows)
	if !reflect.DeepEqual(rows, []string{"0", "0 1", "1"}) {
		t.Fatalf("fault rows %q", content)
	}
}

func TestDumpNewFaults(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	prior := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	next := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	if Gia_ManFaultDumpNewFaults(p, 1, prior, next, &Bmc_ParFf_t{}) != 1 {
		t.Fatal("failed enumeration")
	}
	data, err := os.ReadFile("newfaults.txt")
	if err != nil || string(data) != "0 \n" {
		t.Fatalf("faults %q %v", data, err)
	}
}

func TestDumpUntestableFaultCountMatchesRows(t *testing.T) {
	p := gia.Gia_ManStart(4)
	a := gia.Gia_ManAppendCi(p)
	data := &cnf.Cnf_Dat_t{VarNums: make([]int32, gia.Gia_ManObjNum(p))}
	data.VarNums[a>>1] = 0
	for _, value := range []bool{false, true} {
		s := bsat.NewSolver()
		s.NewVariable()
		lit := bsat.Negative(0)
		if value {
			lit = bsat.Positive(0)
		}
		if !s.AddClause(lit) {
			t.Fatal("unit failed")
		}
		name := filepath.Join(t.TempDir(), "faults")
		got := Gia_ManDumpUntests(p, data, s, 0, name, 0)
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		want := int32(0)
		if value {
			want = 1
		}
		if got != want || strings.Count(string(content), "\n") != int(want) {
			t.Fatalf("value=%v count=%d file=%q", value, got, content)
		}
	}
}
