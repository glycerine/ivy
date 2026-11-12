package cas

import (
	"bytes"
	"fmt"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func truth(dd *DdManager, table uint64, n int) *DdNode {
	var build func(int, uint64) *DdNode
	build = func(index int, table uint64) *DdNode {
		if index == n {
			if table&1 != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		width := uint(1) << uint(n-index-1)
		mask := (uint64(1) << width) - 1
		return cudd.Cudd_bddIte(dd, cudd.Cudd_bddIthVar(dd, index), build(index+1, table>>width), build(index+1, table&mask))
	}
	return build(0, table)
}
func TestCascadeAllThreeVariableFunctions(t *testing.T) {
	for table := uint64(0); table < 256; table++ {
		dd := cudd.NewBDDManager(3)
		f := cudd.Cudd_BddToAdd(dd, truth(dd, table, 3))
		luts, codes, err := Decompose(dd, f, 3, 3)
		if err != nil {
			t.Fatal(table, err)
		}
		if !VerifyCascade(dd, f, luts, codes) {
			t.Fatalf("table %x", table)
		}
		var out bytes.Buffer
		WriteLUTSintoBLIFfile(&out, dd, luts, len(luts), codes, []string{"a", "b", "c"}, 3, "")
		for a := 0; a < 8; a++ {
			env := map[string]bool{"a": a&4 != 0, "b": a&2 != 0, "c": a&1 != 0}
			if got := evalBLIF(t, out.String(), env); got != (table>>uint(a)&1 != 0) {
				t.Fatalf("BLIF table=%x assignment=%d got=%v", table, a, got)
			}
		}
	}
}

// Evaluate emitted .names blocks independently of diagram operations.
func evalBLIF(t *testing.T, text string, env map[string]bool) bool {
	type block struct {
		signals []string
		rows    []string
	}
	var blocks []block
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ".names") {
			blocks = append(blocks, block{signals: strings.Fields(line)[1:]})
		} else if !strings.HasPrefix(line, ".") {
			blocks[len(blocks)-1].rows = append(blocks[len(blocks)-1].rows, line)
		}
	}
	for pass := 0; pass <= len(blocks); pass++ {
		for _, b := range blocks {
			n := len(b.signals) - 1
			ready := true
			for _, s := range b.signals[:n] {
				if _, ok := env[s]; !ok {
					ready = false
				}
			}
			if !ready {
				continue
			}
			value := false
			for _, row := range b.rows {
				fields := strings.Fields(row)
				if n == 0 {
					value = fields[0] == "1"
					continue
				}
				if len(fields[0]) != n {
					t.Fatal("invalid BLIF cube", row)
				}
				match := true
				for j, c := range fields[0] {
					if c != '-' && (c == '1') != env[b.signals[j]] {
						match = false
					}
				}
				if match {
					value = fields[len(fields)-1] == "1"
				}
			}
			env[b.signals[n]] = value
		}
	}
	result, ok := env["F"]
	if !ok {
		t.Fatal("unresolved BLIF output")
	}
	return result
}
func TestCascadeParityMultipleStages(t *testing.T) {
	for _, order := range [][]int{{0, 1, 2, 3, 4, 5}, {5, 3, 1, 4, 2, 0}} {
		dd := cudd.NewBDDManager(6)
		f := dd.LogicZero
		for i := 0; i < 6; i++ {
			f = cudd.Cudd_bddXor(dd, f, cudd.Cudd_bddIthVar(dd, i))
		}
		cudd.Cudd_Ref(f)
		cudd.Cudd_ShuffleHeap(dd, order)
		a := cudd.Cudd_BddToAdd(dd, f)
		luts, codes, err := Decompose(dd, a, 6, 3)
		if err != nil {
			t.Fatal(err)
		}
		if len(luts) != 3 || !VerifyCascade(dd, a, luts, codes) {
			t.Fatalf("stages=%d", len(luts))
		}
		names := []string{"a", "b", "c", "d", "e", "f"}
		var out bytes.Buffer
		WriteLUTSintoBLIFfile(&out, dd, luts, len(luts), codes, names, 6, "")
		for assignment := 0; assignment < 64; assignment++ {
			env := map[string]bool{}
			ones := 0
			for i, s := range names {
				env[s] = assignment&(1<<i) != 0
				if env[s] {
					ones++
				}
			}
			if evalBLIF(t, out.String(), env) != (ones%2 == 1) {
				t.Fatalf("order=%v assignment=%d", order, assignment)
			}
		}
	}
}
func TestOutputEncodingSortAndUnusedCodes(t *testing.T) {
	dd := cudd.NewBDDManager(5)
	outputs := []*DdNode{dd.Vars[0], cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1]), dd.One}
	f := GetSingleOutputFunction(dd, outputs, 3, dd.Vars[3:], 2, false)
	if f == nil {
		t.Fatal("encoding failed")
	}
	for code := 0; code < 4; code++ {
		for a := 0; a < 4; a++ {
			assignment := []bool{a&1 != 0, a&2 != 0, false, code&2 != 0, code&1 != 0}
			got, err := dd.EvaluateADD(f, assignment)
			want := false
			switch code {
			case 0:
				want = assignment[0] != assignment[1]
			case 1:
				want = assignment[0]
			case 2:
				want = true
			}
			if err != nil || (got != 0) != want {
				t.Fatal(code, a, got, err)
			}
		}
	}
}
func TestRemapNewManager(t *testing.T) {
	dd := cudd.NewBDDManager(6)
	outputs := []*DdNode{cudd.Cudd_bddXor(dd, dd.Vars[4], dd.Vars[5]), dd.Vars[3]}
	var dest *DdManager
	f := GetSingleOutputFunctionRemappedNewDD(dd, outputs, 2, &dest)
	if f == nil || len(dest.Vars) != 3 {
		t.Fatal("remap", f, dest)
	}
	for code := 0; code < 2; code++ {
		for a := 0; a < 4; a++ {
			v := []bool{code != 0, a&2 != 0, a&1 != 0}
			got, err := dest.EvaluateADD(f, v)
			want := v[1] != v[2]
			if code == 1 {
				want = v[1]
			}
			if err != nil || (got != 0) != want {
				t.Fatal(code, a, got, err)
			}
		}
	}
}
func TestCascadeInvalidAndTooSmall(t *testing.T) {
	dd := cudd.NewBDDManager(6)
	if _, _, err := Decompose(dd, dd.One, 6, 1); err == nil {
		t.Fatal("accepted LUT1")
	}
	if _, _, err := Decompose(dd, cudd.Cudd_addConst(dd, 2), 6, 3); err == nil {
		t.Fatal("accepted nonboolean ADD")
	}
	f := truth(dd, 0x123456789abcdef, 6)
	_, _, err := Decompose(dd, cudd.Cudd_BddToAdd(dd, f), 6, 2)
	if err == nil {
		t.Fatal("accepted too-small LUT")
	}
}
func ExampleGrayCode() {
	fmt.Println(GrayCode(6)) // Output: 5
}

// The checksum is generated by test_vectors/encoding_oracle.c linked to the
// original casCore.c/CUDD; no native compiler is needed by this test.
func TestOriginalCOutputEncodingFixture(t *testing.T) {
	var hash uint64 = 14695981039346656037
	for a := 0; a < 256; a++ {
		for b := 0; b < 16; b++ {
			dd := cudd.NewBDDManager(5)
			outputs := []*DdNode{truth(dd, uint64(a), 3), truth(dd, uint64(b*17), 3), dd.One}
			f := GetSingleOutputFunction(dd, outputs, 3, dd.Vars[3:], 2, false)
			for v := 0; v < 32; v++ {
				values := make([]bool, 5)
				for i := range values {
					values[i] = v&(1<<uint(4-i)) != 0
				}
				value, err := dd.EvaluateADD(f, values)
				if err != nil {
					t.Fatal(err)
				}
				hash ^= uint64(value)
				hash *= 1099511628211
			}
		}
	}
	if hash != 0x3702fd9a65c8d325 {
		t.Fatalf("original C output-encoding checksum: %016x", hash)
	}
}
func TestForeignManagerAndZeroColumnStages(t *testing.T) {
	dd := cudd.NewBDDManager(6)
	other := cudd.NewBDDManager(6)
	if _, _, err := Decompose(dd, cudd.Cudd_addIthVar(other, 0), 6, 3); err == nil {
		t.Fatal("accepted foreign diagram")
	}
	for _, f := range []*DdNode{dd.One, dd.Zero, cudd.Cudd_addIthVar(dd, 5)} {
		n := len(dd.Vars)
		luts, codes, err := Decompose(dd, f, 6, 3)
		if err != nil || !VerifyCascade(dd, f, luts, codes) {
			t.Fatal("zero-width intermediate stage", n, err)
		}
	}
}

func TestRandomMultistageTruthTables(t *testing.T) {
	seed := uint64(0x9e3779b97f4a7c15)
	accepted := 0
	for i := 0; i < 128; i++ {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		dd := cudd.NewBDDManager(6)
		f := cudd.Cudd_BddToAdd(dd, truth(dd, seed, 6))
		luts, codes, err := Decompose(dd, f, 6, 4)
		if err != nil {
			continue
		}
		accepted++
		if !VerifyCascade(dd, f, luts, codes) {
			t.Fatalf("table %016x", seed)
		}
		names := []string{"a", "b", "c", "d", "e", "f"}
		var out bytes.Buffer
		WriteLUTSintoBLIFfile(&out, dd, luts, len(luts), codes, names, 6, "")
		for a := 0; a < 64; a++ {
			env := map[string]bool{}
			for k, n := range names {
				env[n] = a&(1<<uint(5-k)) != 0
			}
			if evalBLIF(t, out.String(), env) != (seed>>uint(a)&1 != 0) {
				t.Fatal("random BLIF", seed, a)
			}
		}
	}
	if accepted < 10 {
		t.Fatalf("insufficient feasible cascade coverage: %d", accepted)
	}
}

func TestCascadeExperimentWritesEquivalentNetworks(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	outputs := []*DdNode{cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2]), dd.Vars[1], dd.One}
	base := filepath.Join(t.TempDir(), "cascade")
	if Abc_CascadeExperiment(base, dd, outputs, 3, 3, 4, true, false) != 1 {
		t.Fatal("experiment failed")
	}
	encoded, err := os.ReadFile(base + "_ENC.blif")
	if err != nil {
		t.Fatal(err)
	}
	decomposed, err := os.ReadFile(base + "_LUT.blif")
	if err != nil {
		t.Fatal(err)
	}
	for code := 0; code < 4; code++ {
		for a := 0; a < 8; a++ {
			env := map[string]bool{"pi000": a&4 != 0, "pi001": a&2 != 0, "pi002": a&1 != 0, "OutEnc_00": code&2 != 0, "OutEnc_01": code&1 != 0}
			other := make(map[string]bool)
			for k, v := range env {
				other[k] = v
			}
			want := false
			switch code {
			case 0:
				want = env["pi000"] != env["pi002"]
			case 1:
				want = env["pi001"]
			case 2:
				want = true
			}
			if evalBLIF(t, string(encoded), env) != want || evalBLIF(t, string(decomposed), other) != want {
				t.Fatal("experiment BLIF mismatch", code, a)
			}
		}
	}
}
