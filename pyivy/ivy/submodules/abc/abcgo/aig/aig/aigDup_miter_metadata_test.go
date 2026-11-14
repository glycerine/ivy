package aig

import "testing"

func TestCreateMiterDoesNotInheritOutputRoles(t *testing.T) {
	for _, role := range []string{"constraint", "assertion"} {
		t.Run(role, func(t *testing.T) {
			a, b := NewGraph(), NewGraph()
			a.AddOutput(a.NewInput())
			b.AddOutput(b.NewInput())
			if role == "constraint" {
				Aig_ManSetConstrNum(a, 1)
			} else {
				a.Assertions = 1
			}
			result := Aig_ManCreateMiter(a, b, 0)
			if result.nConstrs != 0 || result.Assertions != 0 {
				t.Fatalf("new XOR miter inherited operand output roles: constraints=%d assertions=%d", result.nConstrs, result.Assertions)
			}
			if err := result.Check(); err != nil {
				t.Fatal(err)
			}
			for _, value := range []bool{false, true} {
				outputs, err := result.Evaluate([]bool{value})
				if err != nil || len(outputs) != 1 || outputs[0] {
					t.Fatalf("input=%v outputs=%v err=%v", value, outputs, err)
				}
			}
		})
	}
}

func TestOutputSelectionDropsOriginalAssertionRoles(t *testing.T) {
	transforms := []struct {
		name                  string
		run                   func(*Graph) *Graph
		wantOutputs, wantRegs int
	}{
		{"without", Aig_ManDupWithoutPos, 0, 0},
		{"flops", Aig_ManDupFlopsOnly, 1, 1},
		{"one/comb", func(g *Graph) *Graph { return Aig_ManDupOneOutput(g, 0, 0) }, 1, 0},
		{"one/seq", func(g *Graph) *Graph { return Aig_ManDupOneOutput(g, 0, 1) }, 2, 1},
		{"or/comb", func(g *Graph) *Graph { return Aig_ManDupOrpos(g, 0) }, 1, 0},
		{"or/seq", func(g *Graph) *Graph { return Aig_ManDupOrpos(g, 1) }, 2, 1},
		{"unsolved/comb", func(g *Graph) *Graph { return Aig_ManDupUnsolvedOutputs(g, 0) }, 2, 0},
		{"unsolved/seq", func(g *Graph) *Graph { return Aig_ManDupUnsolvedOutputs(g, 1) }, 3, 1},
	}
	for _, tc := range transforms {
		t.Run(tc.name, func(t *testing.T) {
			src := NewGraph()
			x, state := src.NewInput(), src.NewInput()
			src.AddOutput(x)
			src.AddOutput(state)
			src.AddOutput(state.Not())
			Aig_ManSetRegNum(src, 1)
			src.Assertions = 1
			dst := tc.run(src)
			if dst.Assertions != 0 {
				t.Errorf("inherited %d assertions on transformed outputs", dst.Assertions)
			}
			if dst.OutputCount() != tc.wantOutputs || int(dst.RegisterCount()) != tc.wantRegs {
				t.Fatalf("outputs=%d regs=%d", dst.OutputCount(), dst.RegisterCount())
			}
			if src.Assertions != 1 {
				t.Fatal("source assertion metadata changed")
			}
			if err := dst.Check(); err != nil {
				t.Fatal(err)
			}
			for bits := 0; bits < 4; bits++ {
				v, s := bits&1 != 0, bits&2 != 0
				got, err := dst.Evaluate([]bool{v, s})
				if err != nil {
					t.Fatal(err)
				}
				var want []bool
				switch tc.name {
				case "without":
				case "flops":
					want = []bool{!s}
				case "one/comb", "one/seq":
					want = []bool{v}
				case "or/comb", "or/seq":
					want = []bool{v || s}
				default:
					want = []bool{v, s}
				}
				if tc.wantRegs != 0 && tc.name != "flops" {
					want = append(want, !s)
				}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("bits=%d output=%d got=%v want=%v", bits, i, got, want)
					}
				}
			}
		})
	}
}
