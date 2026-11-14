package bmc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test every normalized two-input function that depends on both inputs. The
// selected fanins and LUT bits are read directly from the SAT model; this is
// independent of Exa3_ManEvalV2's truth-table implementation.
func TestExa32IndependentTruthTables(t *testing.T) {
	for _, truth := range []uint64{2, 4, 6, 8, 14} {
		p := Exa3_ManAllocV2(&Bmc_EsPar_t{NVars: 2, NNodes: 1, NLutSize: 2}, []uint64{truth})
		if Exa3_ManAddCnfStartV2(p, 0) == 0 {
			t.Fatalf("truth=%x: startup", truth)
		}
		mint := int32(1)
		for step := 0; step < 20 && mint >= 0; step++ {
			if Exa3_ManAddCnfV2(p, mint) == 0 || p.pSat.Solve() != bsat.Satisfiable {
				t.Fatalf("truth=%x: unsatisfiable", truth)
			}
			mint = Exa3_ManEvalV2(p)
		}
		if mint != -1 {
			t.Fatalf("truth=%x: did not converge", truth)
		}
		for bits := 0; bits < 4; bits++ {
			lut := 0
			for k := 0; k < 2; k++ {
				selected := -1
				for id := 0; id < 2; id++ {
					v := p.VarMarks[2][k][id]
					if v != 0 {
						value, ok := p.pSat.Value(v)
						if !ok {
							t.Fatal("missing model")
						}
						if value {
							if selected != -1 {
								t.Fatal("multiple fanins")
							}
							selected = id
						}
					}
				}
				if selected < 0 {
					t.Fatal("missing fanin")
				}
				if bits&(1<<selected) != 0 {
					lut |= 1 << k
				}
			}
			value := false
			if lut > 0 {
				var ok bool
				value, ok = p.pSat.Value(int32(lut))
				if !ok {
					t.Fatal("missing LUT bit")
				}
			}
			if value != (truth&(1<<bits) != 0) {
				t.Fatalf("truth=%x bits=%d", truth, bits)
			}
		}
		Exa3_ManFreeV2(p)
	}
}

func TestExa3Entrypoints(t *testing.T) {
	for _, tc := range []struct {
		name, truth string
		run         func(*Bmc_EsPar_t)
		mode        int32
		want        string
	}{
		{"base", "8", Exa3_ManExactSynthesis, 0, "Realization"},
		{"incremental", "8", Exa3_ManExactSynthesis, 1, "Realization"},
		{"base-complement", "7", Exa3_ManExactSynthesis, 0, "Realization"},
		{"v2", "8", Exa3_ManExactSynthesisV2, 0, "Realization"},
		{"v2-complement", "7", Exa3_ManExactSynthesisV2, 0, "Realization"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := captureMeshOutput(t, func() { tc.run(&Bmc_EsPar_t{NVars: 2, NNodes: 1, NLutSize: 2, PTtStr: tc.truth, FUseIncr: tc.mode}) })
			if !strings.Contains(out, tc.want) || strings.Contains(out, "no solution") {
				t.Fatalf("unexpected output %q", out)
			}
		})
	}
}

func TestOtherExactSynthesisEntrypoints(t *testing.T) {
	t.Run("two-input-v2", func(t *testing.T) {
		output := captureMeshOutput(t, func() { Exa_ManExactSynthesis2(&Bmc_EsPar_t{NVars: 2, NNodes: 1, PTtStr: "8"}) })
		if !strings.Contains(output, "Realization") {
			t.Fatalf("missing solution: %q", output)
		}
	})
	t.Run("zyx-majority", func(t *testing.T) {
		output := captureMeshOutput(t, func() { Zyx_ManExactSynthesis(&Bmc_EsPar_t{NVars: 3, NNodes: 1, NLutSize: 3, FMajority: 1}) })
		if !strings.Contains(output, "MAJ3") || strings.Contains(output, "no solution") {
			t.Fatalf("missing majority solution: %q", output)
		}
	})
}

func TestExa32InitialPolarityPreservesSolution(t *testing.T) {
	p := Exa3_ManAllocV2(&Bmc_EsPar_t{NVars: 2, NNodes: 1, NLutSize: 2}, []uint64{8})
	Exa3_ManInitPolarity(p)
	if Exa3_ManAddCnfStartV2(p, 0) == 0 {
		t.Fatal("startup unsatisfiable")
	}
	mint := int32(1)
	for step := 0; step < 16 && mint >= 0; step++ {
		if Exa3_ManAddCnfV2(p, mint) == 0 || p.pSat.Solve() != bsat.Satisfiable {
			t.Fatal("AND should remain satisfiable")
		}
		mint = Exa3_ManEvalV2(p)
	}
	if mint != -1 {
		t.Fatalf("did not converge; minterm %d", mint)
	}
}

func TestExactSynthesisEntrypointWritesBlif(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	output := captureMeshOutput(t, func() { Exa_ManExactSynthesis(&Bmc_EsPar_t{NVars: 2, NNodes: 1, PTtStr: "8"}) })
	if !strings.Contains(output, "Realization") {
		t.Fatalf("no realization: %q", output)
	}
	files, err := filepath.Glob("*.blif")
	if err != nil || len(files) != 1 {
		t.Fatalf("BLIF files %v: %v", files, err)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	// One AND gate has precisely the 11 on-set cube, independent of pin order.
	if !strings.Contains(string(data), ".names b a F\n11 1\n") && !strings.Contains(string(data), ".names a b F\n11 1\n") {
		t.Fatalf("unexpected BLIF: %s", data)
	}
}

func TestZyxBlockingRemovesTopology(t *testing.T) {
	p, err := newZyxSynthesis(&Bmc_EsPar_t{NVars: 2, NNodes: 1, NLutSize: 2, PTtStr: "8"})
	if err != nil {
		t.Fatal(err)
	}
	net, status := p.solve()
	if status != bsat.Satisfiable {
		t.Fatalf("expected solution, got %v", status)
	}
	if p.blockTopology(net) {
		_, status = p.solve()
		if status != bsat.Unsatisfiable {
			t.Fatal("the unique two-input topology was not blocked")
		}
	}
}

func TestMajorityConnect2Topology(t *testing.T) {
	nodes := make([][3]int32, 16)
	Maj_ManConnect2(nodes, 7, 16, 0)
	expected := map[int][3]int32{10: {0, 0, 9}, 12: {0, 0, 11}, 13: {0, 10, 9}, 14: {0, 12, 11}, 15: {0, 14, 13}}
	for i, node := range nodes {
		if node != expected[i] {
			t.Fatalf("node %d = %v, want %v", i, node, expected[i])
		}
		for _, fanin := range node {
			if fanin != 0 && fanin >= int32(i) {
				t.Fatalf("non-topological edge %d -> %d", fanin, i)
			}
		}
	}
}
