package gia

import (
	"strings"
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

func TestCircuitCofactoringPreservesTypedProperty(t *testing.T) {
	for _, kind := range []string{"XOR", "MUX"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(16)
			input, state := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			var root int32
			if kind == "XOR" {
				root = Gia_ManAppendXorReal(src, input, state)
			} else {
				src.EnableMuxes()
				root = Gia_ManAppendMuxReal(src, input, 1, state)
			}
			Gia_ManAppendCo(src, root)
			Gia_ManAppendCo(src, state)
			Gia_ManSetRegNum(src, 1)
			checkQuantificationOutputs(t, src, []bool{true, false}, []bool{true, false})
			var out *Gia_Man_t
			text := hashReviewStdout(t, func() { out = Gia_ManCofTest(src, 3, 0, 0, 0) })
			if !strings.Contains(text, "Property failed in frame 0.") || strings.Contains(text, "Property holds.") {
				t.Fatalf("%s is reachable from the zero initial state, got verdict %q", kind, text)
			}
			if out == nil || Gia_ManCoNum(out) == 0 {
				t.Fatal("cofactoring returned no image")
			}
		})
	}
}

func TestCircuitCofactoringEncodingMatchesAllGateValues(t *testing.T) {
	checks := 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				func() {
					g := windowReviewGraph(signs, typed, layout)
					p := &Ccf_Man_t{pFrames: g, pSat: bsat.NewSolver()}
					defer p.pSat.Close()
					Gia_ManCofExtendSolver(p)
					clauses := p.pSat.Statistics().Clauses
					Gia_ManCofExtendSolver(p)
					if p.pSat.Statistics().Clauses != clauses || p.pSat.NumVars() != g.nObjs {
						t.Fatal("repeated extension duplicated constraints or variables")
					}
					for mask := 0; mask < 16; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
						values, _, err := g.evaluateObjects(inputs)
						if err != nil {
							t.Fatal(err)
						}
						assumptions := make([]int32, 0, 5)
						for i, id := range Vec_IntArray(g.vCis) {
							assumptions = append(assumptions, id<<1^giaBool(!inputs[i]))
						}
						if p.pSat.Solve(assumptions...) != bsat.Satisfiable {
							t.Fatal("fixed input assignment became UNSAT")
						}
						for _, obj := range g.pObjs {
							if Gia_ObjIsCo(obj) != 0 {
								continue
							}
							wrong := obj.id<<1 ^ giaBool(values[obj.id])
							trial := append(assumptions, wrong)
							if p.pSat.Solve(trial...) != bsat.Unsatisfiable {
								t.Fatalf("signs%d typed%d layout%d assignment%d node%d admits wrong value", signs, typed, layout, mask, obj.id)
							}
							checks++
						}
					}
					old := g.nObjs
					g.EnableMuxes()
					fresh := Gia_ManAppendXorReal(g, Gia_ObjToLit(g, Gia_ManCi(g, 0)), Gia_ObjToLit(g, Gia_ManCi(g, 3))^1)
					Gia_ManCofExtendSolver(p)
					if p.encoded != g.nObjs || p.encoded <= old {
						t.Fatal("new gates were not incrementally encoded")
					}
					for mask := 0; mask < 16; mask++ {
						assumptions := make([]int32, 0, 5)
						for i, id := range Vec_IntArray(g.vCis) {
							assumptions = append(assumptions, id<<1^int32(1-(mask>>i&1)))
						}
						expected := (mask&1 != 0) != (mask&8 == 0)
						assumptions = append(assumptions, fresh^giaBool(expected))
						if p.pSat.Solve(assumptions...) != bsat.Unsatisfiable {
							t.Fatal("incremental XOR admits wrong value")
						}
					}
				}()
			}
		}
	}
	t.Logf("%d fixed-input gate contradictions plus 1024 incremental XOR checks", checks)
}

func ccofReviewTable(g *Gia_Man_t, input, state int32, table int) int32 {
	var root int32
	for assignment := 0; assignment < 4; assignment++ {
		if table>>assignment&1 == 0 {
			continue
		}
		term := Gia_ManHashAnd(g, input^int32(1-(assignment&1)), state^int32(1-(assignment>>1&1)))
		root = Gia_ManHashOr(g, root, term)
	}
	return root
}
func ccofReviewMachine(property, transition int) *Gia_Man_t {
	g := Gia_ManStart(32)
	input, state := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashStart(g)
	bad, next := ccofReviewTable(g, input, state, property), ccofReviewTable(g, input, state, transition)
	Gia_ManAppendCo(g, bad)
	Gia_ManAppendCo(g, next)
	Gia_ManSetRegNum(g, 1)
	Gia_ManHashStop(g)
	return g
}
func TestCircuitCofactoringImagesMatchExhaustiveMachines(t *testing.T) {
	steps, rows := 0, 0
	for property := 0; property < 16; property++ {
		for transition := 0; transition < 16; transition++ {
			func() {
				src := ccofReviewMachine(property, transition)
				p := Ccf_ManStart(src, 3, 0, 0, 0)
				defer Ccf_ManStop(p)
				seen := 0
				for frame := 1; frame <= 3; frame++ {
					frameBad := 0
					for initial := 0; initial < 2; initial++ {
						for history := 0; history < 1<<frame; history++ {
							state := initial
							bad := false
							for step := 0; step < frame; step++ {
								pi := history >> step & 1
								bad = property>>(pi|state<<1)&1 != 0
								state = transition >> (pi | state<<1) & 1
							}
							if bad {
								frameBad |= 1 << initial
							}
						}
					}
					expectedStatus := int32(0)
					if frameBad&^seen == 0 {
						expectedStatus = 1
					}
					seen |= frameBad
					p.pFrames = Gia_ManUnrollAdd(p.pUnr, int32(frame)).(*Gia_Man_t)
					Gia_ManCofExtendSolver(p)
					status := Gia_ManCofGetReachable(p, Gia_ManUnrollLastLit(p.pUnr))
					if status != expectedStatus {
						t.Fatalf("property%04b transition%04b frame%d: status%d want%d, image%02b", property, transition, frame, status, expectedStatus, seen)
					}
					for assignment := 0; assignment < 1<<(frame+1); assignment++ {
						inputs := make([]bool, frame+1)
						for i := range inputs {
							inputs[i] = assignment>>i&1 != 0
						}
						outputs, err := p.pFrames.Evaluate(inputs)
						if err != nil {
							t.Fatal(err)
						}
						got := false
						for _, out := range outputs {
							got = got || out
						}
						if expected := seen>>(assignment&1)&1 != 0; got != expected {
							t.Fatalf("property%04b transition%04b frame%d input%d: image%v want%v", property, transition, frame, assignment, got, expected)
						}
						rows++
					}
					steps++
				}
			}()
		}
	}
	t.Logf("256 exhaustive one-bit machines, %d frame images and %d assignments", steps, rows)
}
