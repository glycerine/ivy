package abc

// Tests for the Go FRAIG orchestration derived from src/proof/fraig.

import (
	"math/rand"
	"slices"
	"testing"
)

func newFraigTestOrchestrator(t testing.TB, functional bool) *FraigOrchestrator {
	t.Helper()
	o, err := NewFraigOrchestrator(FraigOrchestrationOptions{
		Manager:     FraigOptions{MaxNodes: 4096, MaxInputs: 16, MaxOutputs: 64, MaxTruthInputs: 6, FunctionalReduce: functional},
		Parameters:  FraigParameters{RandomPatterns: 128, DynamicPatterns: 128, BacktrackLimit: 99, Seconds: 20, FunctionalReduce: functional, Feedback: true, DistanceOne: true, TryProve: true},
		MaxPatterns: 256, MaxChoices: 256, MaxClauses: 256, MaxVisits: 1 << 16,
	})
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func fraigTruthForEdge(t testing.TB, o *FraigOrchestrator, edge FraigEdge) uint64 {
	t.Helper()
	inputs := len(o.manager.inputs)
	var truth uint64
	values := make([]bool, inputs)
	for assignment := 0; assignment < 1<<inputs; assignment++ {
		for i := range values {
			values[i] = assignment&(1<<i) != 0
		}
		value, err := o.manager.Evaluate(edge, values)
		if err != nil {
			t.Fatal(err)
		}
		if value {
			truth |= uint64(1) << assignment
		}
	}
	return truth
}

func TestFraigParametersAndOwnedAPI(t *testing.T) {
	defaults := DefaultFraigParameters()
	if defaults.RandomPatterns != 2048 || defaults.DynamicPatterns != 2048 || defaults.BacktrackLimit != 99 || defaults.Seconds != 20 || !defaults.FunctionalReduce || !defaults.Feedback || !defaults.DistanceOne || defaults.Sparse || defaults.Choices || !defaults.TryProve {
		t.Fatalf("defaults=%+v", defaults)
	}
	full := FullFraigParameters()
	if full.BacktrackLimit != -1 || !full.Sparse || full.TryProve {
		t.Fatalf("full=%+v", full)
	}
	o := newFraigTestOrchestrator(t, false)
	a, err := o.NewInput("a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := o.Input(2)
	if err != nil || len(o.manager.inputs) != 3 || b != o.manager.inputs[2] {
		t.Fatalf("input=%v count=%d err=%v", b, len(o.manager.inputs), err)
	}
	if got := o.InputNames(); !slices.Equal(got, []string{"a", "", ""}) {
		t.Fatalf("names=%v", got)
	}
	and, _ := o.And(a, b)
	or, _ := o.Or(a, b)
	xor, _ := o.Xor(a, b)
	mux, _ := o.Mux(o.manager.inputs[1], a, b)
	if fraigTruthForEdge(t, o, and) != 0xa0 || fraigTruthForEdge(t, o, or) != 0xfa || fraigTruthForEdge(t, o, xor) != 0x5a || fraigTruthForEdge(t, o, mux) != 0xb8 {
		t.Fatalf("truth and=%x or=%x xor=%x mux=%x", fraigTruthForEdge(t, o, and), fraigTruthForEdge(t, o, or), fraigTruthForEdge(t, o, xor), fraigTruthForEdge(t, o, mux))
	}
	if err = o.AddOutput(xor, "diff"); err != nil {
		t.Fatal(err)
	}
	names := o.OutputNames()
	names[0] = "changed"
	if o.OutputNames()[0] != "diff" {
		t.Fatal("output names were not owned")
	}
	if err = o.Validate(); err != nil {
		t.Fatal(err)
	}
	o.Close()
	if _, err = o.And(a, b); err != ErrInvariant {
		t.Fatalf("closed err=%v", err)
	}
}

func TestFraigSATRelationsClausesAndMiter(t *testing.T) {
	o := newFraigTestOrchestrator(t, false)
	a, _ := o.NewInput("a")
	b, _ := o.NewInput("b")
	beforeTrivial := o.Stats().SATQueries
	if trivial, trivialErr := o.Equivalent(a, a, FraigSATLimits{}); trivialErr != nil || trivial.Status != FraigRelationProved || o.Stats().SATQueries != beforeTrivial {
		t.Fatalf("trivial=%+v err=%v stats=%+v", trivial, trivialErr, o.Stats())
	}
	ab, _ := o.And(a, b)
	ba, _ := o.And(b, a)
	result, err := o.Equivalent(ab, ba, FraigSATLimits{})
	if err != nil || result.Status != FraigRelationProved {
		t.Fatalf("equivalent=%+v err=%v", result, err)
	}
	result, err = o.Equivalent(a, b, FraigSATLimits{})
	if err != nil || result.Status != FraigRelationDisproved || len(result.Model) != 2 {
		t.Fatalf("different=%+v err=%v", result, err)
	}
	av, _ := o.manager.Evaluate(a, result.Model)
	bv, _ := o.manager.Evaluate(b, result.Model)
	if av == bv {
		t.Fatalf("model does not replay: %v", result.Model)
	}
	result, err = o.Implies(ab, a, FraigSATLimits{})
	if err != nil || result.Status != FraigRelationProved {
		t.Fatalf("implication=%+v err=%v", result, err)
	}
	result, err = o.Implies(a, ab, FraigSATLimits{})
	if err != nil || result.Status != FraigRelationDisproved {
		t.Fatalf("counter implication=%+v err=%v", result, err)
	}
	if err = o.AddClause([]FraigEdge{a}); err != nil {
		t.Fatal(err)
	}
	result, err = o.Implies(o.manager.ConstantOne(), a, FraigSATLimits{})
	if err != nil || result.Status != FraigRelationProved {
		t.Fatalf("clause implication=%+v err=%v", result, err)
	}
	zero, _ := o.Xor(a, a)
	if err = o.AddOutput(zero, "zero"); err != nil {
		t.Fatal(err)
	}
	result, err = o.CheckMiter(FraigSATLimits{})
	if err != nil || result.Status != FraigRelationProved {
		t.Fatalf("zero miter=%+v err=%v", result, err)
	}
	if err = o.AddOutput(ab, "ab"); err != nil {
		t.Fatal(err)
	}
	result, err = o.CheckMiter(FraigSATLimits{})
	if err != nil || result.Status != FraigRelationDisproved {
		t.Fatalf("bug miter=%+v err=%v", result, err)
	}
	if got := o.Stats(); got.SATQueries == 0 || got.SATProved == 0 || got.SATDisproved == 0 {
		t.Fatalf("stats=%+v", got)
	}
}

func TestFraigChoicesDFSRecognizersAndSupergate(t *testing.T) {
	o := newFraigTestOrchestrator(t, false)
	a, _ := o.NewInput("a")
	b, _ := o.NewInput("b")
	c, _ := o.NewInput("c")
	ab, _ := o.And(a, b)
	left, _ := o.And(ab, c)
	bc, _ := o.And(b, c)
	right, _ := o.And(a, bc)
	if err := o.AddChoice(left.Not(), right.Not(), FraigSATLimits{}); err != nil {
		t.Fatal(err)
	}
	classes := o.ChoiceClasses()
	if len(classes) != 1 || classes[0].Representative != left.Not() || !slices.Equal(classes[0].Alternatives, []FraigEdge{right.Not()}) {
		t.Fatalf("classes=%v", classes)
	}
	post, err := o.DFS([]FraigEdge{left}, true, false)
	if err != nil || len(post) < 4 || post[len(post)-1] != left.ID() {
		t.Fatalf("post=%v err=%v", post, err)
	}
	if in, inErr := o.InTFI(a.ID(), left.ID()); inErr != nil || !in {
		t.Fatalf("tfi=%v err=%v", in, inErr)
	}
	if in, inErr := o.InTFI(left.ID(), a.ID()); inErr != nil || in {
		t.Fatalf("reverse tfi=%v err=%v", in, inErr)
	}
	mux, _ := o.Mux(a, b, c)
	control, whenTrue, whenFalse, ok := o.RecognizeMux(mux)
	if !ok {
		t.Fatal("mux not recognized")
	}
	rebuilt, _ := o.Mux(control, whenTrue, whenFalse)
	if fraigTruthForEdge(t, o, rebuilt) != fraigTruthForEdge(t, o, mux) {
		t.Fatal("recognized mux differs")
	}
	xor, _ := o.Xor(a, b)
	xa, xb, ok := o.RecognizeXor(xor)
	if !ok {
		t.Fatal("xor not recognized")
	}
	rebuiltXor, _ := o.Xor(xa, xb)
	if fraigTruthForEdge(t, o, rebuiltXor) != fraigTruthForEdge(t, o, xor) {
		t.Fatal("recognized xor differs")
	}
	super, err := o.CollectSupergate(left, false)
	if err != nil || len(super) != 2 {
		t.Fatalf("super=%v err=%v", super, err)
	}
	product := o.manager.ConstantOne()
	for _, factor := range super {
		product, err = o.And(product, factor)
		if err != nil {
			t.Fatal(err)
		}
	}
	if fraigTruthForEdge(t, o, product) != fraigTruthForEdge(t, o, left) {
		t.Fatal("supergate product differs")
	}
	if err = o.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFraigAssociativeChoicesAndUnknown(t *testing.T) {
	o := newFraigTestOrchestrator(t, false)
	a, _ := o.NewInput("a")
	b, _ := o.NewInput("b")
	c, _ := o.NewInput("c")
	ab, _ := o.And(a, b)
	root, _ := o.And(ab, c)
	added, err := o.AddAssociativeChoices(4, FraigSATLimits{})
	if err != nil || added == 0 || len(o.ChoiceClasses()) == 0 {
		t.Fatalf("added=%d classes=%v err=%v", added, o.ChoiceClasses(), err)
	}
	for _, class := range o.ChoiceClasses() {
		for _, alternative := range class.Alternatives {
			if fraigTruthForEdge(t, o, class.Representative) != fraigTruthForEdge(t, o, alternative) {
				t.Fatal("associative choice changed function")
			}
		}
	}
	before := o.Stats()
	if err = o.AddChoice(root, a, FraigSATLimits{}); err != ErrInvariant {
		t.Fatalf("unequal choice err=%v", err)
	}
	if after := o.Stats(); after.Choices != before.Choices || after.ChoiceClasses != before.ChoiceClasses {
		t.Fatalf("rejected choice mutated classes: before=%+v after=%+v", before, after)
	}
	result, err := o.Equivalent(root, a, FraigSATLimits{Stop: func() bool { return true }})
	if err != nil || result.Status != FraigRelationUnknown {
		t.Fatalf("cancelled relation=%+v err=%v", result, err)
	}
	if err = o.AddClause(nil); err != ErrDimension {
		t.Fatalf("empty clause err=%v", err)
	}
}

func TestFraigChoicePhaseAndChaining(t *testing.T) {
	o := newFraigTestOrchestrator(t, false)
	a, _ := o.NewInput("a")
	b, _ := o.NewInput("b")
	c, _ := o.NewInput("c")
	ab, _ := o.And(a, b)
	left, _ := o.And(ab, c)
	bc, _ := o.And(b, c)
	right, _ := o.And(a, bc)
	ac, _ := o.And(a, c)
	third, _ := o.And(ac, b)
	if err := o.AddChoice(left.Not(), right.Not(), FraigSATLimits{}); err != nil {
		t.Fatal(err)
	}
	before := o.ChoiceClasses()
	if err := o.AddChoice(left, third, FraigSATLimits{}); err != ErrInvariant {
		t.Fatalf("opposite representative phase err=%v", err)
	}
	if !slices.EqualFunc(before, o.ChoiceClasses(), func(a, b FraigChoiceClass) bool {
		return a.Representative == b.Representative && slices.Equal(a.Alternatives, b.Alternatives)
	}) {
		t.Fatal("rejected phase change mutated choices")
	}
	if err := o.AddChoice(right.Not(), third.Not(), FraigSATLimits{}); err != nil {
		t.Fatalf("chained choice: %v", err)
	}
	classes := o.ChoiceClasses()
	if len(classes) != 1 || classes[0].Representative != left.Not() || !slices.Equal(classes[0].Alternatives, []FraigEdge{right.Not(), third.Not()}) {
		t.Fatalf("classes=%v", classes)
	}
}

func TestFraigMuxNormalizationAndSupergateSourceSemantics(t *testing.T) {
	o := newFraigTestOrchestrator(t, false)
	a, _ := o.NewInput("a")
	b, _ := o.NewInput("b")
	c, _ := o.NewInput("c")
	mux, _ := o.Mux(a.Not(), b, c)
	control, whenTrue, whenFalse, ok := o.RecognizeMux(mux)
	if !ok || control != a || whenTrue != c || whenFalse != b {
		t.Fatalf("mux control=%v true=%v false=%v ok=%v", control, whenTrue, whenFalse, ok)
	}
	if stopped, err := o.CollectSupergate(mux.Not(), true); err != nil || !slices.Equal(stopped, []FraigEdge{mux.Not()}) {
		t.Fatalf("stopped=%v err=%v", stopped, err)
	}
	g := newFraigTestOrchestrator(t, false)
	ga, _ := g.NewInput("a")
	gb, _ := g.NewInput("b")
	gc, _ := g.NewInput("c")
	gd, _ := g.NewInput("d")
	ab, _ := g.And(ga, gb)
	ac, _ := g.And(ga, gc)
	root, _ := g.And(ab, ac)
	_, _ = g.And(root, gd)
	if err := g.AddOutput(root, "shared"); err != nil {
		t.Fatal(err)
	}
	leaves, err := g.CollectSupergate(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(leaves, []FraigEdge{ga, gb, gc}) {
		t.Fatalf("leaves=%v", leaves)
	}
}

func TestFraigVisitBudgetsAndCombinedStop(t *testing.T) {
	globalStop := false
	o, err := NewFraigOrchestrator(FraigOrchestrationOptions{
		Manager:    FraigOptions{MaxNodes: 16, MaxInputs: 4},
		Parameters: FraigParameters{RandomPatterns: 128, DynamicPatterns: 128},
		MaxVisits:  1,
		Stop:       func() bool { return globalStop },
	})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := o.NewInput("a")
	b, _ := o.NewInput("b")
	if result, solveErr := o.Equivalent(a, b, FraigSATLimits{}); solveErr != ErrLimit || result.Status != FraigRelationUnknown {
		t.Fatalf("solver budget result=%+v err=%v", result, solveErr)
	}
	if added, feedbackErr := o.AddFeedbackPattern([]bool{true, false}); feedbackErr != ErrLimit || added || len(o.FeedbackPatterns()) != 0 {
		t.Fatalf("feedback added=%v patterns=%v err=%v", added, o.FeedbackPatterns(), feedbackErr)
	}
	localStop := func() bool { return false }
	stop := o.satLimits(FraigSATLimits{Stop: localStop}).Stop
	globalStop = true
	if stop == nil || !stop() {
		t.Fatal("caller stop masked the manager stop")
	}
}

func TestFraigFeedbackOwnershipLimitsAndCancellation(t *testing.T) {
	o := newFraigTestOrchestrator(t, false)
	a, _ := o.NewInput("a")
	b, _ := o.NewInput("b")
	ab, _ := o.And(a, b)
	patterns := [][]bool{{false, false}, {true, false}, {false, true}, {true, true}}
	for _, pattern := range patterns {
		added, err := o.AddFeedbackPattern(pattern)
		if err != nil || !added {
			t.Fatalf("add=%v err=%v", added, err)
		}
	}
	if ones, err := o.FeedbackOnes(ab.ID()); err != nil || ones != 1 {
		t.Fatalf("ones=%d err=%v", ones, err)
	}
	copyPatterns := o.FeedbackPatterns()
	copyPatterns[0][0] = true
	if o.FeedbackPatterns()[0][0] {
		t.Fatal("patterns alias")
	}
	if added, err := o.AddFeedbackPattern(patterns[0]); err != nil || added {
		t.Fatalf("duplicate added=%v err=%v", added, err)
	}
	limited, err := NewFraigOrchestrator(FraigOrchestrationOptions{Manager: FraigOptions{MaxNodes: 8, MaxInputs: 2}, Parameters: FraigParameters{RandomPatterns: 128, DynamicPatterns: 128}, MaxPatterns: 1, MaxChoices: 1, MaxClauses: 1, MaxVisits: 8})
	if err != nil {
		t.Fatal(err)
	}
	la, _ := limited.NewInput("a")
	if _, err = limited.NewInput("b"); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.AddFeedbackPattern([]bool{false, false}); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.AddFeedbackPattern([]bool{true, false}); err != ErrLimit {
		t.Fatalf("pattern limit=%v", err)
	}
	if err = limited.AddClause([]FraigEdge{la}); err != nil {
		t.Fatal(err)
	}
	if err = limited.AddClause([]FraigEdge{la.Not()}); err != ErrLimit {
		t.Fatalf("clause limit=%v", err)
	}
	calls := 0
	cancelled, err := NewFraigOrchestrator(FraigOrchestrationOptions{Manager: FraigOptions{MaxNodes: 32, MaxInputs: 2}, Parameters: FraigParameters{RandomPatterns: 128, DynamicPatterns: 128}, Stop: func() bool { calls++; return calls >= 4 }})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = cancelled.NewInput("a")
	_, _ = cancelled.NewInput("b")
	before := cancelled.FeedbackPatterns()
	if _, err = cancelled.AddFeedbackPattern([]bool{true, true}); err != ErrCancelled {
		t.Fatalf("cancel err=%v calls=%d", err, calls)
	}
	if !slices.EqualFunc(before, cancelled.FeedbackPatterns(), slices.Equal[[]bool, bool]) {
		t.Fatal("cancelled feedback mutated state")
	}
}

func TestFraigCompositeFailuresAreAtomic(t *testing.T) {
	limited, err := NewFraigOrchestrator(FraigOrchestrationOptions{
		Manager:     FraigOptions{MaxNodes: 5, MaxInputs: 5, MaxTruthInputs: 3},
		Parameters:  FraigParameters{RandomPatterns: 128, DynamicPatterns: 128},
		MaxPatterns: 8, MaxChoices: 8, MaxClauses: 8, MaxVisits: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := limited.NewInput("a")
	b, _ := limited.NewInput("b")
	c, _ := limited.NewInput("c")
	before := limited.manager.Stats()
	if _, err = limited.Mux(a, b, c); err != ErrLimit {
		t.Fatalf("mux err=%v", err)
	}
	if after := limited.manager.Stats(); after != before {
		t.Fatalf("mux mutated manager: before=%+v after=%+v", before, after)
	}
	if err = limited.Validate(); err != nil {
		t.Fatal(err)
	}
	inputs, names := limited.manager.Inputs(), limited.InputNames()
	if _, err = limited.Input(4); err != ErrLimit {
		t.Fatalf("input err=%v", err)
	}
	if !slices.Equal(inputs, limited.manager.Inputs()) || !slices.Equal(names, limited.InputNames()) {
		t.Fatal("failed sparse input mutated state")
	}
}

func TestFraigIndependentModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0x46524149))
	for trial := 0; trial < 5000; trial++ {
		o := newFraigTestOrchestrator(t, false)
		edges := make([]FraigEdge, 0, 16)
		truths := make([]uint64, 0, 16)
		for variable := 0; variable < 3; variable++ {
			edge, err := o.NewInput("")
			if err != nil {
				t.Fatal(err)
			}
			edges = append(edges, edge)
			var truth uint64
			for assignment := 0; assignment < 8; assignment++ {
				if assignment&(1<<variable) != 0 {
					truth |= uint64(1) << assignment
				}
			}
			truths = append(truths, truth)
		}
		for step := 0; step < 8; step++ {
			i, j := rng.Intn(len(edges)), rng.Intn(len(edges))
			a, b := edges[i], edges[j]
			ta, tb := truths[i], truths[j]
			if rng.Intn(2) != 0 {
				a, ta = a.Not(), (^ta)&0xff
			}
			if rng.Intn(2) != 0 {
				b, tb = b.Not(), (^tb)&0xff
			}
			var edge FraigEdge
			var truth uint64
			var err error
			switch rng.Intn(3) {
			case 0:
				edge, err, truth = func() (FraigEdge, error, uint64) { e, x := o.And(a, b); return e, x, ta & tb }()
			case 1:
				edge, err, truth = func() (FraigEdge, error, uint64) { e, x := o.Or(a, b); return e, x, ta | tb }()
			default:
				edge, err, truth = func() (FraigEdge, error, uint64) { e, x := o.Xor(a, b); return e, x, ta ^ tb }()
			}
			if err != nil {
				t.Fatal(err)
			}
			edges, truths = append(edges, edge), append(truths, truth&0xff)
		}
		last := len(edges) - 1
		if got := fraigTruthForEdge(t, o, edges[last]); got != truths[last] {
			t.Fatalf("trial=%d got=%02x want=%02x", trial, got, truths[last])
		}
	}
}

func FuzzFraigOrchestration(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{255, 128, 64, 32, 16})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 48 {
			data = data[:48]
		}
		o := newFraigTestOrchestrator(t, false)
		edges := make([]FraigEdge, 0, 64)
		for i := 0; i < 4; i++ {
			edge, err := o.NewInput("")
			if err != nil {
				t.Fatal(err)
			}
			edges = append(edges, edge)
		}
		for pos := 0; pos+2 < len(data); pos += 3 {
			a := edges[int(data[pos])%len(edges)]
			b := edges[int(data[pos+1])%len(edges)]
			if data[pos+2]&4 != 0 {
				a = a.Not()
			}
			if data[pos+2]&8 != 0 {
				b = b.Not()
			}
			var edge FraigEdge
			var err error
			switch data[pos+2] % 3 {
			case 0:
				edge, err = o.And(a, b)
			case 1:
				edge, err = o.Or(a, b)
			default:
				edge, err = o.Xor(a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			edges = append(edges, edge)
		}
		if len(edges) < 2 {
			return
		}
		a, b := edges[len(edges)-1], edges[len(edges)/2]
		result, err := o.Equivalent(a, b, FraigSATLimits{Conflicts: 10000, Decisions: 10000, Propagations: 100000})
		if err != nil {
			t.Fatal(err)
		}
		equal := fraigTruthForEdge(t, o, a) == fraigTruthForEdge(t, o, b)
		if result.Status == FraigRelationProved != equal {
			t.Fatalf("status=%v equal=%v", result.Status, equal)
		}
		if result.Status == FraigRelationDisproved {
			av, _ := o.manager.Evaluate(a, result.Model)
			bv, _ := o.manager.Evaluate(b, result.Model)
			if av == bv {
				t.Fatal("counterexample does not replay")
			}
		}
		if err = o.Validate(); err != nil {
			t.Fatal(err)
		}
	})
}
