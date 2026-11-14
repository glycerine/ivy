package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func quantificationReviewGraph(signs int32, registers, single bool) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a^(signs&1), b)
	buffer := Gia_ManAppendBuf(g, x^((signs>>1)&1))
	mux := Gia_ManAppendMuxReal(g, c^((signs>>2)&1), buffer, d)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, mux, a^1))
	if !single {
		Gia_ManAppendCo(g, Gia_ManAppendAnd(g, c, x))
		Gia_ManAppendCo(g, mux^1)
	}
	if registers {
		Gia_ManSetRegNum(g, 1)
		g.nConstrs = 1
	}
	g.SetIdentity("quantification", "quantification.aig")
	return g
}

func TestQuantificationMatchesOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "quantification.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C quantification: %v: %s", err, want)
	}
	var records strings.Builder
	for regs := 0; regs < 2; regs++ {
		for signs := int32(0); signs < 8; signs++ {
			for variable := int32(0); variable < int32(4-regs); variable++ {
				for op := 0; op < 8; op++ {
					if regs != 0 && op >= 3 {
						continue
					}
					g := quantificationReviewGraph(signs, regs != 0, op == 4 || op == 5)
					var out *Gia_Man_t
					switch op {
					case 0:
						out = Gia_ManDupCofactorVar(g, variable, 0)
					case 1:
						out = Gia_ManDupCofactorVar(g, variable, 1)
					case 2:
						out = Gia_ManDupExist2(g, variable)
					case 3:
						out = Gia_ManDupUniv(g, variable)
					case 4:
						out = Gia_ManDupExist(g, variable)
					case 5:
						negative := quantificationReviewGraph(signs^7, false, true)
						g.SetIdentity("positive", "positive.aig")
						negative.SetIdentity("negative", "negative.aig")
						out = Gia_ManDupMux(variable, g, negative)
					case 6:
						out = Gia_ManDupCofactorObj(g, Gia_ObjFaninId0p(g, Gia_ManCo(g, 0)), variable&1)
					default:
						out = Gia_ManDupCofactorObj(g, Gia_ManCi(g, variable).id, signs&1)
					}
					positiveName := 0
					if name, spec := out.Identity(); name == "positive" && spec == "positive.aig" {
						positiveName = 1
					}
					fmt.Fprintf(&records, "%d %d %d %d %d %d %d %d %d", regs, signs, variable, op,
						Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, out.nConstrs, positiveName)
					for mask := 0; mask < 16; mask++ {
						values, err := out.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0})
						if err != nil {
							t.Fatal(err)
						}
						packed := 0
						for i, value := range values {
							if value {
								packed |= 1 << uint(i)
							}
						}
						fmt.Fprintf(&records, " %d", packed)
					}
					records.WriteByte('\n')
				}
			}
		}
	}
	gotRows := strings.Split(strings.TrimSpace(records.String()), "\n")
	wantRows := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(gotRows) != len(wantRows) {
		t.Fatalf("Go emitted %d cases, C emitted %d", len(gotRows), len(wantRows))
	}
	for i := range gotRows {
		if gotRows[i] != wantRows[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, gotRows[i], wantRows[i])
		}
	}
}

func checkQuantificationOutputs(t *testing.T, g *Gia_Man_t, inputs, want []bool) {
	t.Helper()
	got, err := g.Evaluate(inputs)
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("inputs %v: got %v (%v), want %v", inputs, got, err, want)
	}
}

func TestCofactorMuxAcceptsSameGraphForBothBranches(t *testing.T) {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	for variable := int32(0); variable < 2; variable++ {
		out := Gia_ManDupMux(variable, g, g)
		for mask := 0; mask < 4; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0}
			got, err := out.Evaluate(inputs)
			if err != nil || len(got) != 1 || got[0] != (inputs[0] && inputs[1]) {
				t.Fatalf("variable %d inputs %v: output %v (%v), want unchanged conjunction", variable, inputs, got, err)
			}
		}
	}
}

func TestQuantificationTypedGatesAndSequentialInterfaces(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		for _, registers := range []bool{false, true} {
			g := quantificationReviewGraph(signs, registers, false)
			for variable := int32(0); variable < Gia_ManPiNum(g); variable++ {
				zero := Gia_ManDupCofactorVar(g, variable, 0)
				one := Gia_ManDupCofactorVar(g, variable, 1)
				exists := Gia_ManDupExist2(g, variable)
				outs := []*Gia_Man_t{zero, one, exists}
				var universal *Gia_Man_t
				if !registers {
					universal = Gia_ManDupUniv(g, variable)
					outs = append(outs, universal)
				}
				for _, out := range outs {
					if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 3 || out.nRegs != g.nRegs || out.nConstrs != g.nConstrs {
						t.Fatal("quantification changed interface, registers or constraints")
					}
				}
				for mask := 0; mask < 16; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					fixed := slices.Clone(inputs)
					fixed[variable] = false
					negative, err := g.Evaluate(fixed)
					if err != nil {
						t.Fatal(err)
					}
					fixed[variable] = true
					positive, err := g.Evaluate(fixed)
					if err != nil {
						t.Fatal(err)
					}
					checkQuantificationOutputs(t, zero, inputs, negative)
					checkQuantificationOutputs(t, one, inputs, positive)
					want := make([]bool, len(positive))
					for i := range want {
						want[i] = negative[i] || positive[i]
					}
					checkQuantificationOutputs(t, exists, inputs, want)
					if universal != nil {
						// Source convention: quantify PO zero; take the positive
						// cofactor for all auxiliary outputs.
						want = slices.Clone(positive)
						want[0] = negative[0] && positive[0]
						checkQuantificationOutputs(t, universal, inputs, want)
					}
				}
			}
		}
	}
}

func TestInternalCofactorsTypedGateGuards(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		g := quantificationReviewGraph(signs, false, false)
		for id := int32(1); id < g.nObjs; id++ {
			if Gia_ObjIsCand(g.pObjs[id]) == 0 {
				continue
			}
			for value := int32(0); value < 2; value++ {
				out := Gia_ManDupCofactorObj(g, id, value)
				for mask := 0; mask < 16; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					values, want, err := g.evaluateObjects(inputs)
					if err != nil {
						t.Fatal(err)
					}
					guard := values[id] == (value != 0)
					for i := range want {
						want[i] = want[i] && guard
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			}
		}
	}
}

func TestCofactorMuxBranchesAndSingleOutputExistence(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		positive := quantificationReviewGraph(signs, false, true)
		negative := quantificationReviewGraph(signs^7, false, true)
		positive.SetIdentity("positive", "positive.aig")
		negative.SetIdentity("negative", "negative.aig")
		for variable := int32(0); variable < 4; variable++ {
			out := Gia_ManDupMux(variable, positive, negative)
			same := Gia_ManDupMux(variable, positive, positive)
			exists := Gia_ManDupExist(positive, variable)
			if name, spec := out.Identity(); name != "positive" || spec != "positive.aig" {
				t.Errorf("MUX identity %q/%q, want the positive cofactor's source identity", name, spec)
			}
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				want, err := positive.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, same, inputs, want)
				if !inputs[variable] {
					want, err = negative.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
				}
				checkQuantificationOutputs(t, out, inputs, want)
				fixed := slices.Clone(inputs)
				fixed[variable] = false
				lo, err := positive.Evaluate(fixed)
				if err != nil {
					t.Fatal(err)
				}
				fixed[variable] = true
				hi, err := positive.Evaluate(fixed)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, exists, inputs, []bool{lo[0] || hi[0]})
			}
		}
	}
}
