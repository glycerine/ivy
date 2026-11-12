package mf

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type cnfGate struct {
	a, b, c int32
	kind    int
}
type cnfNetwork struct {
	gates           []cnfGate
	inputs, outputs []int32
}

func (n cnfNetwork) ObjectCount() int        { return len(n.gates) }
func (n cnfNetwork) InputLiterals() []int32  { return n.inputs }
func (n cnfNetwork) OutputLiterals() []int32 { return n.outputs }
func (n cnfNetwork) Gate(id int32) (int32, int32, int32, int) {
	g := n.gates[id]
	return g.a, g.b, g.c, g.kind
}
func cnfAndMapping() *Mapping {
	n := cnfNetwork{gates: []cnfGate{{}, {kind: 1}, {kind: 1}, {a: 2, b: 4, kind: 2}, {a: 6, kind: 6}}, inputs: []int32{2, 4}, outputs: []int32{6}}
	return &Mapping{Network: n, Options: DefaultOptions(), Cuts: [][]Cut{nil, nil, nil, {{Leaves: []int32{1, 2}, Function: 4}}, nil}, Truths: [][]uint64{{0, 0, 0, 0}, {0xaaaaaaaaaaaaaaaa, 0xaaaaaaaaaaaaaaaa, 0xaaaaaaaaaaaaaaaa, 0xaaaaaaaaaaaaaaaa}, {0x8888888888888888, 0x8888888888888888, 0x8888888888888888, 0x8888888888888888}}, ClauseSizes: []int{1, 2, 3}, Refs: []int32{0, 1, 1, 1, 0}}
}
func TestCNFSourceANDClauseOrder(t *testing.T) {
	m := cnfAndMapping()
	cnf, err := DeriveCNF(m, false, true, false)
	if err != nil {
		t.Fatal(err)
	}
	// Original ./abc: &r -s and.aig; &write_cnf -K8 and.cnf yields
	// 2; 2 -3; -2 3; 3 -5 -6; -3 6; -3 5; -4 (DIMACS uses var+1).
	want := [][]int32{{2}, {2, 5}, {3, 4}, {4, 9, 11}, {5, 10}, {5, 8}, {7}}
	if cnf.Variables != 6 || !reflect.DeepEqual(cnf.Clauses, want) {
		t.Fatalf("vars%d clauses%v", cnf.Variables, cnf.Clauses)
	}
	if !reflect.DeepEqual(cnf.VarNums[:5], []int32{3, 4, 5, 2, 1}) {
		t.Fatal("source variable ordering", cnf.VarNums)
	}
}
func cnfSatisfied(clauses [][]int32, assignment uint64) bool {
	for _, clause := range clauses {
		satisfied := false
		for _, lit := range clause {
			value := assignment>>uint(lit>>1)&1 != 0
			if value != (lit&1 != 0) {
				satisfied = true
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}
func TestCNFAllPhasesAndObjectIDs(t *testing.T) {
	for _, phase := range []int32{0, 1} {
		for _, objectIDs := range []bool{false, true} {
			m := cnfAndMapping()
			m.Cuts[3][0].Function |= phase
			cnf, err := DeriveCNF(m, objectIDs, false, true)
			if err != nil {
				t.Fatal(err)
			}
			for assignment := uint64(0); assignment < 1<<cnf.Variables; assignment++ {
				value := func(id int) bool { return assignment>>uint(cnf.VarNums[id])&1 != 0 }
				want := !value(0) && value(3) == ((value(1) && value(2)) != (phase != 0)) && value(4) == value(3)
				if cnfSatisfied(cnf.Clauses, assignment) != want {
					t.Fatalf("phase%d ids%v assignment%x", phase, objectIDs, assignment)
				}
			}
			offset := cnf.Mapping[cnf.VarNums[3]]
			truth := uint32(cnf.Mapping[offset+3])
			want := uint32(0x88888888)
			if phase != 0 {
				want = ^want
			}
			if truth != want {
				t.Fatalf("mapping phase truth%x want%x", truth, want)
			}
		}
	}
}

func TestCNFConstantsSyntheticOutputsAndEmptyOr(t *testing.T) {
	n := cnfNetwork{gates: []cnfGate{{}}, outputs: []int32{0, 1}}
	m := &Mapping{Network: n, Options: DefaultOptions()}
	cnf, err := DeriveCNF(m, false, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if cnf.Variables != 4 || !reflect.DeepEqual(cnf.Clauses, [][]int32{{2, 4}, {2, 7}, {3, 6}, {4, 6}, {5, 7}, {7}}) {
		t.Fatal("constant output clauses", cnf)
	}
	n.outputs = nil
	m.Network = n
	cnf, err = DeriveCNF(m, false, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(cnf.Clauses) != 2 || len(cnf.Clauses[0]) != 0 {
		t.Fatal("no-output OR must be empty/unsatisfiable")
	}
}
func TestCNFSmallLUTAndConstantCut(t *testing.T) {
	m := cnfAndMapping()
	m.Options.LUTSize = 3
	for i := range m.Truths {
		m.Truths[i] = m.Truths[i][:1]
	}
	if _, err := DeriveCNF(m, false, false, false); err != nil {
		t.Fatal(err)
	}
	m.Cuts[3][0] = Cut{Function: 1}
	cnf, err := DeriveCNF(m, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cnf.Clauses[2], []int32{4}) {
		t.Fatal("constant-one cut", cnf.Clauses)
	}
	if _, err := DeriveCNF(m, false, false, true); err == nil {
		t.Fatal("source constant-cut mapping limitation lost")
	}
}

func TestCNFWideRandomFunctionsExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(84239))
	for _, leaves := range []int{7, 8} {
		for sample := 0; sample < 12; sample++ {
			words := 1 << (leaves - 6)
			table := make([]uint64, words)
			for i := range table {
				table[i] = rng.Uint64()
			}
			table[0] &^= 1
			n := cnfNetwork{gates: []cnfGate{{}}}
			cutLeaves := make([]int32, leaves)
			for i := 0; i < leaves; i++ {
				n.gates = append(n.gates, cnfGate{kind: 1})
				n.inputs = append(n.inputs, int32(2*(i+1)))
				cutLeaves[i] = int32(i + 1)
			}
			root := int32(len(n.gates))
			n.gates = append(n.gates, cnfGate{a: 2, b: 4, kind: 2})
			n.gates = append(n.gates, cnfGate{a: root << 1, kind: 6})
			n.outputs = []int32{root << 1}
			for _, phase := range []int32{0, 1} {
				cuts := make([][]Cut, len(n.gates))
				cuts[root] = []Cut{{Leaves: cutLeaves, Function: 4 | phase}}
				refs := make([]int32, len(n.gates))
				refs[root] = 1
				padded := make([]uint64, 4)
				copy(padded, table)
				m := &Mapping{Network: n, Options: DefaultOptions(), Cuts: cuts, Truths: [][]uint64{nil, nil, table}, ClauseSizes: []int{1, 2, int(util.Abc_Tt8CnfSize(padded, int32(leaves)))}, Refs: refs}
				cnf, err := DeriveCNF(m, false, false, true)
				if err != nil {
					t.Fatal(err)
				}
				for pattern := 0; pattern < 1<<leaves; pattern++ {
					for output := 0; output < 2; output++ {
						assignment := uint64(0)
						for i, id := range cutLeaves {
							if pattern>>i&1 != 0 {
								assignment |= 1 << uint(cnf.VarNums[id])
							}
						}
						if output != 0 {
							assignment |= 1 << uint(cnf.VarNums[root])
							assignment |= 1 << uint(cnf.VarNums[root+1])
						}
						want := int(table[pattern>>6]>>uint(pattern&63)&1) ^ int(phase)
						if cnfSatisfied(cnf.Clauses, assignment) != (output == want) {
							t.Fatalf("leaves%d sample%d phase%d pattern%x output%d", leaves, sample, phase, pattern, output)
						}
					}
				}
				offset := cnf.Mapping[cnf.VarNums[root]]
				start := int(offset) + 1 + leaves
				for i, word := range table {
					want := word
					if phase != 0 {
						want = ^want
					}
					got := uint64(uint32(cnf.Mapping[start+2*i])) | uint64(uint32(cnf.Mapping[start+2*i+1]))<<32
					if got != want {
						t.Fatalf("wide flat truth word%d got%x want%x", i, got, want)
					}
				}
			}
		}
	}
}

func TestCNFWideOriginalCClauseOrder(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for original-header oracle")
	}
	_, file, _, _ := runtime.Caller(0)
	include := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../src"))
	dir := t.TempDir()
	source := filepath.Join(dir, "cnf_oracle.c")
	binary := filepath.Join(dir, "cnf_oracle")
	code := `#include "misc/util/abc_global.h"
#include "misc/util/utilTruth.h"
int main(void){int n,k,count,cover[512];unsigned long long t[4];while(scanf("%d%llx%llx%llx%llx",&n,&t[0],&t[1],&t[2],&t[3])==5){count=Abc_Tt8Cnf((word*)t,n,cover);printf("%d",count);for(k=0;k<count;k++)printf(" %d",cover[k]);printf("\n");}return 0;}`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-DLIN64", "-I"+include, source, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("original C oracle compile: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(61061))
	for _, leaves := range []int{7, 8} {
		for sample := 0; sample < 4; sample++ {
			table := make([]uint64, 4)
			for i := 0; i < 1<<(leaves-6); i++ {
				table[i] = rng.Uint64()
			}
			table[0] &^= 1
			request := fmt.Sprintf("%d %x %x %x %x\n", leaves, table[0], table[1], table[2], table[3])
			command := exec.Command(binary)
			command.Stdin = strings.NewReader(request)
			out, err := command.Output()
			if err != nil {
				t.Fatal(err)
			}
			reader := strings.NewReader(string(out))
			var count int
			if _, err := fmt.Fscan(reader, &count); err != nil {
				t.Fatal(err)
			}
			cubes := make([]int32, count)
			for i := range cubes {
				if _, err := fmt.Fscan(reader, &cubes[i]); err != nil {
					t.Fatal(err)
				}
			}
			for _, phase := range []int32{0, 1} {
				n := cnfNetwork{gates: make([]cnfGate, leaves+3)}
				ids := make([]int32, leaves)
				for i := range ids {
					ids[i] = int32(i + 1)
					n.gates[i+1].kind = 1
					n.inputs = append(n.inputs, ids[i]<<1)
				}
				root := int32(leaves + 1)
				n.gates[root] = cnfGate{a: 2, b: 4, kind: 2}
				n.gates[root+1] = cnfGate{a: root << 1, kind: 6}
				n.outputs = []int32{root << 1}
				m := &Mapping{Network: n, Options: DefaultOptions(), Cuts: make([][]Cut, len(n.gates)), Truths: [][]uint64{nil, nil, table}, ClauseSizes: []int{1, 2, count}, Refs: make([]int32, len(n.gates))}
				m.Cuts[root] = []Cut{{Leaves: ids, Function: 4 | phase}}
				m.Refs[root] = 1
				cnf, err := DeriveCNF(m, false, false, false)
				if err != nil {
					t.Fatal(err)
				}
				expected := make([][]int32, count)
				for i, cube := range cubes {
					expected[i] = []int32{encodedLiteral(cnf.VarNums[root], (cubeLiteral(cube, leaves) == 2) != (phase != 0))}
					for k, id := range ids {
						lit := cubeLiteral(cube, k)
						if lit != 0 {
							expected[i] = append(expected[i], encodedLiteral(cnf.VarNums[id], lit == 2))
						}
					}
				}
				if !reflect.DeepEqual(cnf.Clauses[2:len(cnf.Clauses)-1], expected) {
					t.Fatalf("original C clause ordering differs leaves%d sample%d phase%d", leaves, sample, phase)
				}
			}
		}
	}
}
