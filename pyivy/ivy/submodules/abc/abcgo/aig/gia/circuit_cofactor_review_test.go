package gia

import (
	"fmt"
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
					defer Ccf_ManStop(p)
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
					if p.pSat.NumVars() != g.nObjs || p.pSat.NumVars() <= old {
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

func TestCircuitCofactoringHonorsLatchResetStates(t *testing.T) {
	for _, reset := range []int32{0, 1, -1} {
		t.Run(fmt.Sprintf("reset%d", reset), func(t *testing.T) {
			src := Gia_ManStart(8)
			state := Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, state)
			Gia_ManAppendCo(src, state)
			Gia_ManSetRegNum(src, 1)
			if err := src.SetAigerLatchInitialStates([]int32{reset}); err != nil {
				t.Fatal(err)
			}
			text := hashReviewStdout(t, func() { Gia_ManCofTest(src, 3, 0, 0, 0) })
			if reset == 0 {
				if !strings.Contains(text, "Property holds.") {
					t.Fatalf("zero reset should be safe: %q", text)
				}
			} else if !strings.Contains(text, "Property failed in frame 0.") || strings.Contains(text, "Property holds.") {
				t.Fatalf("reset%d includes a bad initial state: %q", reset, text)
			}
		})
	}
}

func TestCircuitCofactoringModelSubstitution(t *testing.T) {
	checks := 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				func() {
					g := windowReviewGraph(signs, typed, layout)
					Gia_ManHashStart(g)
					source := Gia_ManStart(8)
					for i := 0; i < 2; i++ {
						lit := Gia_ManAppendCi(source)
						Gia_ManAppendCo(source, lit)
					}
					Gia_ManSetRegNum(source, 2)
					p := &Ccf_Man_t{pGia: source, pFrames: g, pSat: bsat.NewSolver(), vCopies: Vec_IntAlloc(0)}
					defer Ccf_ManStop(p)
					Gia_ManCofExtendSolver(p)
					var roots []int32
					for _, id := range Vec_IntArray(g.vCos) {
						roots = append(roots, Gia_ObjFaninLit0p(g, g.pObjs[id]))
					}
					roots = append(roots, 0, 1, Gia_ObjToLit(g, Gia_ManCi(g, 0))^1, Gia_ObjToLit(g, Gia_ManCi(g, 3)))
					for model := 0; model < 16; model++ {
						for _, root := range roots {
							assumptions := make([]int32, 4)
							for i, id := range Vec_IntArray(g.vCis) {
								assumptions[i] = id<<1 ^ int32(1-(model>>i&1))
							}
							if p.pSat.Solve(assumptions...) != bsat.Satisfiable {
								t.Fatal("fixed source model is not SAT")
							}
							blocked := Gia_ManCofOneDerive(p, root)
							for state := 0; state < 4; state++ {
								original := []bool{state&1 != 0, state&2 != 0, model&4 != 0, model&8 != 0}
								values, _, err := g.evaluateObjects(original)
								if err != nil {
									t.Fatal(err)
								}
								expected := values[root>>1] != (root&1 != 0)
								// Changing the fixed frame PIs must not affect the emitted state cofactor.
								actual := []bool{original[0], original[1], !original[2], !original[3]}
								values, outputs, err := g.evaluateObjects(actual)
								if err != nil {
									t.Fatal(err)
								}
								if outputs[len(outputs)-1] != expected || (values[blocked>>1] != (blocked&1 != 0)) == expected {
									t.Fatalf("phase%d typed%d layout%d model%d state%d: wrong cofactor or blocking polarity", signs, typed, layout, model, state)
								}
								checks++
							}
						}
					}
				}()
			}
		}
	}
	t.Logf("%d raw model-cofactor/blocking checks, including typed gates, layout changes, constants and CI roots", checks)
}

func TestCircuitCofactoringVerdictsMatchExhaustiveResets(t *testing.T) {
	cases := 0
	for property := 0; property < 16; property++ {
		for transition := 0; transition < 16; transition++ {
			for _, reset := range []int32{0, 1, -1} {
				unsafe := false
				for initial := 0; initial < 2; initial++ {
					if reset != -1 && initial != int(reset) {
						continue
					}
					for history := 0; history < 8; history++ {
						state := initial
						for frame := 0; frame < 3; frame++ {
							pi := history >> frame & 1
							if property>>(pi|state<<1)&1 != 0 {
								unsafe = true
							}
							state = transition >> (pi | state<<1) & 1
						}
					}
				}
				src := ccofReviewMachine(property, transition)
				if err := src.SetAigerLatchInitialStates([]int32{reset}); err != nil {
					t.Fatal(err)
				}
				var out *Gia_Man_t
				text := hashReviewStdout(t, func() { out = Gia_ManCofTest(src, 3, 0, 0, 0) })
				failed, holds := strings.Contains(text, "Property failed in frame"), strings.Contains(text, "Property holds.")
				if failed != unsafe || holds == unsafe {
					t.Fatalf("property%04b transition%04b reset%d: unsafe%v, verdict%q", property, transition, reset, unsafe, text)
				}
				if out == nil || out.nObjs == 0 || Gia_ManRegNum(out) != 0 {
					t.Fatal("cofactoring did not return a live combinational image")
				}
				// Stop released its internal unroller without stopping this returned graph.
				if _, err := out.Evaluate(make([]bool, Gia_ManCiNum(out))); err != nil {
					t.Fatal(err)
				}
				cases++
			}
		}
	}
	t.Logf("%d complete one-bit checks with zero/one/unknown resets match exhaustive three-frame reachability", cases)
}

func TestCircuitCofactoringInitialIntersectionIgnoresImageBlocks(t *testing.T) {
	cases := 0
	for table := 0; table < 16; table++ {
		for _, r0 := range []int32{0, 1, -1} {
			for _, r1 := range []int32{0, 1, -1} {
				func() {
					g := Gia_ManStart(32)
					a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
					Gia_ManHashStart(g)
					root := ccofReviewTable(g, a, b, table)
					Gia_ManAppendCo(g, root)
					src := ccofReviewStateSource(2)
					if err := src.SetAigerLatchInitialStates([]int32{r0, r1}); err != nil {
						t.Fatal(err)
					}
					p := &Ccf_Man_t{pGia: src, pFrames: g, pSat: bsat.NewSolver()}
					defer Ccf_ManStop(p)
					Gia_ManCofExtendSolver(p)
					// The real enumerator can exclude every image state. Its clauses must
					// never constrain the separate question about admissible initial states.
					p.pSat.AddClause(0)
					expected := false
					for state := 0; state < 4; state++ {
						if r0 != -1 && int(r0) != state&1 || r1 != -1 && int(r1) != state>>1&1 {
							continue
						}
						expected = expected || table>>state&1 != 0
					}
					status := giaCofInitialFailure(p)
					if (status == bsat.Satisfiable) != expected || status == bsat.Unknown {
						t.Fatalf("table%04b resets%v: status%v expected intersection%v", table, []int32{r0, r1}, status, expected)
					}
					cases++
				}()
			}
		}
	}
	t.Logf("%d exact two-register reset/image intersections, independent of contradictory enumeration clauses", cases)
}

func ccofReviewPigeonhole() *Gia_Man_t {
	g := Gia_ManStart(256)
	var vars [5][4]int32
	for pigeon := range vars {
		for hole := range vars[pigeon] {
			vars[pigeon][hole] = Gia_ManAppendCi(g)
		}
	}
	Gia_ManHashStart(g)
	root := int32(1)
	for pigeon := range vars {
		var clause int32
		for _, lit := range vars[pigeon] {
			clause = Gia_ManHashOr(g, clause, lit)
		}
		root = Gia_ManHashAnd(g, root, clause)
	}
	for hole := 0; hole < 4; hole++ {
		for a := 0; a < 5; a++ {
			for b := a + 1; b < 5; b++ {
				clause := Gia_ManHashOr(g, vars[a][hole]^1, vars[b][hole]^1)
				root = Gia_ManHashAnd(g, root, clause)
			}
		}
	}
	Gia_ManAppendCo(g, root)
	Gia_ManHashStop(g)
	return g
}
func TestCircuitCofactoringLimitsAndLifecycle(t *testing.T) {
	src := ccofReviewMachine(8, 12)
	p := Ccf_ManStart(src, 3, 0, 5, 0)
	if p.pSat.Deadline().IsZero() {
		t.Fatal("runtime limit did not reach solver")
	}
	p.pFrames = Gia_ManUnrollAdd(p.pUnr, 1).(*Gia_Man_t)
	Gia_ManCofExtendSolver(p)
	p.pSat.SetRuntimeDeadline(func() int64 { return 10 }, 1)
	if status := Gia_ManCofGetReachable(p, Gia_ManUnrollLastLit(p.pUnr)); status != -1 || Gia_ManCoNum(p.pFrames) != 0 {
		t.Fatalf("expired enumeration status%d outputs%d", status, Gia_ManCoNum(p.pFrames))
	}
	p.pSat.SetRuntimeDeadline(nil, 0)
	if status := Gia_ManCofGetReachable(p, Gia_ManUnrollLastLit(p.pUnr)); status != 0 {
		t.Fatalf("resumed image status%d", status)
	}
	if status := Gia_ManCofGetReachable(p, Gia_ManUnrollLastLit(p.pUnr)); status != 1 {
		t.Fatalf("repeated completed image status%d", status)
	}
	savedFrames, unroller := p.pFrames, p.pUnr
	Ccf_ManStop(p)
	Ccf_ManStop(p)
	Ccf_ManStop(nil)
	if savedFrames.nObjs != 0 || Gia_ManUnrollLastLit(unroller) != -1 || p.pSat != nil || p.pGia != nil {
		t.Fatal("stop did not release owned frame/unroller state")
	}
	if _, err := src.Evaluate([]bool{false, false}); err != nil {
		t.Fatal("stop destroyed source graph", err)
	}
	hard := Ccf_ManStart(ccofReviewPigeonhole(), 1, 1, 0, 0)
	defer Ccf_ManStop(hard)
	hard.pFrames = Gia_ManUnrollAdd(hard.pUnr, 1).(*Gia_Man_t)
	Gia_ManCofExtendSolver(hard)
	if status := Gia_ManCofGetReachable(hard, Gia_ManUnrollLastLit(hard.pUnr)); status != -1 {
		t.Fatalf("small conflict budget returned%d, expected unknown", status)
	}
	hard.pSat.SetLimits(0, 0)
	if status := Gia_ManCofGetReachable(hard, Gia_ManUnrollLastLit(hard.pUnr)); status != 1 {
		t.Fatalf("unlimited UNSAT property returned%d", status)
	}
}
func TestCircuitCofactoringResetQueryHonorsResources(t *testing.T) {
	src := ccofReviewStateSource(2)
	if err := src.SetAigerLatchInitialStates([]int32{-1, -1}); err != nil {
		t.Fatal(err)
	}
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	p := &Ccf_Man_t{pGia: src, pFrames: g, pSat: bsat.NewSolver()}
	defer Ccf_ManStop(p)
	p.pSat.SetRuntimeDeadline(func() int64 { return 10 }, 1)
	if status := giaCofInitialFailure(p); status != bsat.Unknown {
		t.Fatalf("expired reset intersection returned%v, expected unknown", status)
	}
	p.pSat.SetRuntimeDeadline(nil, 0)
	if status := giaCofInitialFailure(p); status != bsat.Satisfiable {
		t.Fatalf("resumed reset intersection returned%v", status)
	}
	hard := ccofReviewPigeonhole()
	source := ccofReviewStateSource(int(Gia_ManCiNum(hard)))
	resets := make([]int32, Gia_ManCiNum(hard))
	for i := range resets {
		resets[i] = -1
	}
	if err := source.SetAigerLatchInitialStates(resets); err != nil {
		t.Fatal(err)
	}
	if p.initialSolver != nil {
		p.initialSolver.Close()
		p.initialSolver = nil
	}
	p.pGia, p.pFrames, p.nConfMax = source, hard, 1
	if status := giaCofInitialFailure(p); status != bsat.Unknown {
		t.Fatalf("reset query ignored conflict limit: %v", status)
	}
	p.nConfMax = 0
	if status := giaCofInitialFailure(p); status != bsat.Unsatisfiable {
		t.Fatalf("unlimited impossible reset intersection returned%v", status)
	}
}

func TestCircuitCofactoringReportsAttemptedFrames(t *testing.T) {
	for _, test := range []struct {
		src               *Gia_Man_t
		frames, conflicts int32
		expected          string
	}{
		{ccofReviewMachine(0, 0), 3, 0, "Backward reachability converged after 1 iterations."},
		{ccofReviewMachine(12, 12), 3, 0, "Backward reachability converged after 2 iterations."},
		{ccofReviewPigeonhole(), 1, 1, "Conflict limit or timeout is reached after 1 frames."},
	} {
		text := hashReviewStdout(t, func() { Gia_ManCofTest(test.src, test.frames, test.conflicts, 0, 0) })
		if !strings.Contains(text, test.expected) || strings.Contains(text, "after -") || strings.Contains(text, "Runtime limit (0") {
			t.Errorf("expected %q, got%q", test.expected, text)
		}
	}
}
func BenchmarkCircuitCofactoringImage(b *testing.B) {
	src := ccofReviewMachine(8, 12)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := Ccf_ManStart(src, 3, 0, 0, 0)
		p.pFrames = Gia_ManUnrollAdd(p.pUnr, 1).(*Gia_Man_t)
		Gia_ManCofExtendSolver(p)
		Gia_ManCofGetReachable(p, Gia_ManUnrollLastLit(p.pUnr))
		Ccf_ManStop(p)
	}
}

func TestCircuitCofactoringResetCheckerExtendsImages(t *testing.T) {
	src := ccofReviewStateSource(2)
	if err := src.SetAigerLatchInitialStates([]int32{0, -1}); err != nil {
		t.Fatal(err)
	}
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashStart(g)
	Gia_ManAppendCo(g, Gia_ManHashAnd(g, a, b))
	p := &Ccf_Man_t{pGia: src, pFrames: g, pSat: bsat.NewSolver()}
	defer Ccf_ManStop(p)
	if status := giaCofInitialFailure(p); status != bsat.Unsatisfiable {
		t.Fatalf("first image intersection: %v", status)
	}
	Gia_ManAppendCo(g, Gia_ManHashAnd(g, a^1, b))
	for i := 0; i < 2; i++ {
		if status := giaCofInitialFailure(p); status != bsat.Satisfiable {
			t.Fatalf("extended image intersection: %v", status)
		}
	}
}
func TestCircuitCofactoringIncompleteResetQueryIsUndecided(t *testing.T) {
	hard := ccofReviewPigeonhole()
	src := Gia_ManStart(hard.nObjs + 32)
	control := Gia_ManAppendCi(src)
	hard.pObjs[0].Value = 0
	for _, id := range Vec_IntArray(hard.vCis) {
		hard.pObjs[id].Value = uint32(Gia_ManAppendCi(src))
	}
	Gia_ManHashStart(src)
	giaComposeNodes(src, hard, true)
	property := Gia_ManHashOr(src, control^1, Gia_ObjFanin0Copy(Gia_ManCo(hard, 0)))
	Gia_ManAppendCo(src, property)
	for _, id := range Vec_IntArray(src.vCis) {
		Gia_ManAppendCo(src, id<<1)
	}
	Gia_ManHashStop(src)
	Gia_ManSetRegNum(src, Gia_ManCiNum(src))
	resets := make([]int32, Gia_ManRegNum(src))
	for i := range resets {
		resets[i] = -1
	}
	resets[0] = 1
	if err := src.SetAigerLatchInitialStates(resets); err != nil {
		t.Fatal(err)
	}
	var out *Gia_Man_t
	text := hashReviewStdout(t, func() { out = Gia_ManCofTest(src, 1, 4, 0, 0) })
	if Gia_ManCoNum(out) == 0 {
		t.Fatalf("did not reach reset intersection after image enumeration: %q", text)
	}
	if !strings.Contains(text, "Property is undecided.") || !strings.Contains(text, "Conflict limit or timeout") || strings.Contains(text, "Property holds.") || strings.Contains(text, "Property failed") {
		t.Fatalf("incomplete initial-state intersection must remain undecided: %q", text)
	}
}
func BenchmarkCircuitCofactoringResetQueries(b *testing.B) {
	for _, reuse := range []bool{false, true} {
		name := "fresh"
		if reuse {
			name = "reused"
		}
		b.Run(name, func(b *testing.B) {
			src := ccofReviewStateSource(2)
			if err := src.SetAigerLatchInitialStates([]int32{-1, -1}); err != nil {
				b.Fatal(err)
			}
			g := Gia_ManStart(8)
			a, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, c))
			p := &Ccf_Man_t{pGia: src, pFrames: g, pSat: bsat.NewSolver()}
			defer Ccf_ManStop(p)
			giaCofInitialFailure(p)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if !reuse {
					p.initialSolver.Close()
					p.initialSolver = nil
				}
				giaCofInitialFailure(p)
			}
		})
	}
}
