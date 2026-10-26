package abc

// Tests for the ports of src/aig/aig/aigCuts.c and aigFact.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func aigCutTestGraph(t *testing.T) (*AigManager, []AigLit, AigLit) {
	t.Helper()
	g := NewAigManager()
	inputs := make([]AigLit, 4)
	for i := range inputs {
		var err error
		inputs[i], err = g.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	ab, err := g.And(inputs[0], inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	ac, err := g.And(inputs[0], inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	root, err := g.And(ab, ac)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	return g, inputs, root
}

func aigCutKey(leaves []AigID) string {
	b := make([]byte, 0, 2*len(leaves))
	for _, id := range leaves {
		b = append(b, byte(id), byte(id>>8))
	}
	return string(b)
}

func aigReferenceCuts(g *AigManager, id AigID, maximum int, memo map[AigID][][]AigID) [][]AigID {
	if cuts, exists := memo[id]; exists {
		return cuts
	}
	result := [][]AigID{{id}}
	object := g.objects[id]
	if object.Type == AigAND {
		for _, left := range aigReferenceCuts(g, object.Fanin0.ID(), maximum, memo) {
			for _, right := range aigReferenceCuts(g, object.Fanin1.ID(), maximum, memo) {
				if merged, ok := aigCutMergeLeaves(left, right, maximum); ok {
					result = append(result, merged)
				}
			}
		}
	}
	unique := make([][]AigID, 0, len(result))
	seen := make(map[string]bool)
	for _, cut := range result {
		key := aigCutKey(cut)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, cut)
		}
	}
	keep := make([][]AigID, 0, len(unique))
	for i, cut := range unique {
		if len(cut) == 1 && cut[0] == id {
			keep = append(keep, cut)
			continue
		}
		dominated := false
		for j, other := range unique {
			if i != j && len(other) >= 2 && len(other) <= len(cut) && aigCutContained(other, cut) {
				dominated = true
				break
			}
		}
		if !dominated {
			keep = append(keep, cut)
		}
	}
	memo[id] = keep
	return keep
}

func aigCutScalarValue(g *AigManager, root AigLit, leaves []AigID, assignment uint64) bool {
	boundary := make(map[AigID]bool, len(leaves))
	for i, id := range leaves {
		boundary[id] = assignment&(uint64(1)<<i) != 0
	}
	memo := make(map[AigID]bool)
	var evaluate func(AigLit) bool
	evaluate = func(literal AigLit) bool {
		value, found := boundary[literal.ID()]
		if !found {
			value, found = memo[literal.ID()]
		}
		if !found {
			object := g.objects[literal.ID()]
			switch object.Type {
			case AigConstant1:
				value = true
			case AigAND:
				value = evaluate(object.Fanin0) && evaluate(object.Fanin1)
			default:
				panic("cut does not bound a CI")
			}
			memo[literal.ID()] = value
		}
		return value != literal.Complemented()
	}
	return evaluate(root)
}

func TestAigCutEnumerationTruthAndOwnership(t *testing.T) {
	g, _, root := aigCutTestGraph(t)
	before := g.ObjectCount()
	manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 32, MaximumLeaves: 4, ComputeTruth: true})
	if err != nil {
		t.Fatal(err)
	}
	if g.ObjectCount() != before {
		t.Fatal("cut computation mutated its source")
	}
	cuts, err := manager.Cuts(root.ID())
	if err != nil {
		t.Fatal(err)
	}
	want := aigReferenceCuts(g, root.ID(), 4, make(map[AigID][][]AigID))
	wantSet := make(map[string]bool)
	for _, cut := range want {
		wantSet[aigCutKey(cut)] = true
	}
	if len(cuts) != len(wantSet) {
		t.Fatalf("cuts=%v reference=%v", cuts, want)
	}
	for _, cut := range cuts {
		if !wantSet[aigCutKey(cut.Leaves)] || !slices.IsSorted(cut.Leaves) || cut.Signature != aigCutSignature(cut.Leaves) {
			t.Fatalf("invalid cut %+v", cut)
		}
		reference, truthErr := g.CutTruth(root, cut.Leaves, AigCutTruthOptions{})
		if truthErr != nil || cut.Truth == nil || !cut.Truth.Equal(reference.Table) {
			t.Fatalf("truth mismatch %+v: %v", cut, truthErr)
		}
		for assignment := uint64(0); assignment < uint64(1)<<len(cut.Leaves); assignment++ {
			got, getErr := cut.Truth.Get(assignment)
			if getErr != nil || got != aigCutScalarValue(g, root, cut.Leaves, assignment) {
				t.Fatalf("scalar mismatch cut=%v assignment=%d got=%v err=%v", cut.Leaves, assignment, got, getErr)
			}
		}
	}
	first := cuts[0].Leaves[0]
	cuts[0].Leaves[0] = 0
	again, err := manager.Cuts(root.ID())
	if err != nil || again[0].Leaves[0] != first {
		t.Fatal("returned cut aliases manager storage", err)
	}
	total, largest, err := manager.Count()
	if err != nil || total < len(cuts) || largest < 0 {
		t.Fatal(total, largest, err)
	}
	if cost, costErr := manager.ReciprocalCost(cuts[len(cuts)-1]); costErr != nil || cost < 0 {
		t.Fatal(cost, costErr)
	}
	text, err := manager.Format(root.ID())
	if err != nil || len(text) == 0 || text[0] != '{' {
		t.Fatal(text, err)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Cuts(root.ID()); !errors.Is(err, ErrAigCutClosed) {
		t.Fatal(err)
	}
}

func TestAigCutTruthExhaustiveSixInputs(t *testing.T) {
	g := NewAigManager()
	level := make([]AigLit, 6)
	for i := range level {
		level[i], _ = g.CreateCI()
	}
	for len(level) > 1 {
		next := make([]AigLit, 0, (len(level)+1)/2)
		for i := 0; i < len(level); i += 2 {
			if i+1 == len(level) {
				next = append(next, level[i])
				continue
			}
			literal, err := g.And(level[i], level[i+1])
			if err != nil {
				t.Fatal(err)
			}
			next = append(next, literal)
		}
		level = next
	}
	root := level[0]
	_, _ = g.CreateCO(root)
	manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 256, MaximumLeaves: 6, ComputeTruth: true})
	if err != nil {
		t.Fatal(err)
	}
	cuts, err := manager.Cuts(root.ID())
	if err != nil {
		t.Fatal(err)
	}
	foundSix := false
	for _, cut := range cuts {
		if len(cut.Leaves) == 6 {
			foundSix = true
		}
		for assignment := uint64(0); assignment < uint64(1)<<len(cut.Leaves); assignment++ {
			got, getErr := cut.Truth.Get(assignment)
			if getErr != nil || got != aigCutScalarValue(g, root, cut.Leaves, assignment) {
				t.Fatalf("cut=%v assignment=%d got=%v err=%v", cut.Leaves, assignment, got, getErr)
			}
		}
	}
	if !foundSix {
		t.Fatal("missing six-input cut")
	}
}

func TestAigCutPriorityDominanceAndSupportMinimize(t *testing.T) {
	g, inputs, root := aigCutTestGraph(t)
	manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 2, MaximumLeaves: 3})
	if err != nil {
		t.Fatal(err)
	}
	cuts, err := manager.Cuts(root.ID())
	if err != nil || len(cuts) > 2 || len(cuts) == 0 || !slices.Equal(cuts[0].Leaves, []AigID{root.ID()}) {
		t.Fatal(cuts, err)
	}
	x, _ := NewVariableTruthTable(3, 0, 64)
	y, _ := NewVariableTruthTable(3, 1, 64)
	xy, _ := x.And(y, false, false)
	reduced, err := manager.MinimizeSupport(AigKCut{Node: root.ID(), Leaves: []AigID{inputs[0].ID(), inputs[1].ID(), inputs[3].ID()}, Truth: xy})
	if err != nil || !slices.Equal(reduced.Leaves, []AigID{inputs[0].ID(), inputs[1].ID()}) || reduced.Truth.Variables() != 2 {
		t.Fatal(reduced, err)
	}
}

func aigEvalLiteral(g *AigManager, literal AigLit, inputs []bool) (bool, error) {
	values := make([]bool, len(g.objects))
	values[0] = true
	for i, id := range g.cis {
		if i >= len(inputs) {
			return false, ErrDimension
		}
		values[id] = inputs[i]
	}
	read := func(l AigLit) bool { return values[l.ID()] != l.Complemented() }
	for i, object := range g.objects {
		switch object.Type {
		case AigAND:
			values[i] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[i] = read(object.Fanin0)
		}
	}
	return read(literal), nil
}

func TestAigAlgebraicFactor(t *testing.T) {
	g := NewAigManager()
	a, _ := g.CreateCI()
	b, _ := g.CreateCI()
	c, _ := g.CreateCI()
	ab, _ := g.And(a, b)
	ac, _ := g.And(a, c)
	_, _ = g.CreateCO(ab)
	_, _ = g.CreateCO(ac)
	before := g.ObjectCount()
	implications, err := g.AigImplications(ab, AigFactorOptions{})
	if err != nil || !slices.Equal(implications, []AigLit{a, b}) {
		t.Fatal(implications, err)
	}
	if overlap, overlapErr := g.AigConeOverlaps(ac, implications, AigFactorOptions{}); overlapErr != nil || !overlap {
		t.Fatal(overlap, overlapErr)
	}
	derived, deriveErr := g.DeriveAigCone(ac, implications, AigFactorOptions{})
	if deriveErr != nil || !derived.Found {
		t.Fatal(derived, deriveErr)
	}
	for assignment := 0; assignment < 8; assignment++ {
		in := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, _ := aigEvalLiteral(derived.Graph, derived.Literal, in)
		if got != in[2] {
			t.Fatalf("derived assignment %d: got %v want c", assignment, got)
		}
	}
	result, err := g.FactorAlgebraic(0, 1, true, AigFactorOptions{})
	if err != nil || !result.Found || result.Graph == nil || result.Literal == InvalidAigLit || g.ObjectCount() != before {
		t.Fatal(result.Found, result.Literal, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		in := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, evalErr := aigEvalLiteral(result.Graph, result.Literal, in)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		if in[0] && in[1] && got != in[2] {
			t.Fatalf("assignment %d: factor=%v want c", assignment, got)
		}
	}
	missing, err := g.FactorAlgebraic(0, 1, false, AigFactorOptions{})
	if err != nil || missing.Found {
		t.Fatal(missing, err)
	}
}

func TestAigSupportMinimization(t *testing.T) {
	g := NewAigManager()
	x, _ := g.CreateCI()
	y, _ := g.CreateCI()
	z, _ := g.CreateCI()
	nxny, _ := g.And(x.Not(), y.Not())
	condition := nxny.Not()
	xz, _ := g.And(x, z)
	yz, _ := g.And(y, z)
	function, _ := g.Or(xz, yz)
	before := g.ObjectCount()
	result, err := g.MinimizeSupportUnderCondition(condition, function, AigFactorOptions{})
	if err != nil || result.Status != AigSupportMinimized || result.Literal == InvalidAigLit || g.ObjectCount() != before {
		t.Fatal(result.Status, result.Literal, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		cond, _ := aigEvalLiteral(g, condition, inputs)
		if !cond {
			continue
		}
		want, _ := aigEvalLiteral(g, function, inputs)
		got, evalErr := aigEvalLiteral(result.Graph, result.Literal, inputs)
		if evalErr != nil || got != want {
			t.Fatalf("assignment=%d got=%v want=%v err=%v", assignment, got, want, evalErr)
		}
	}
	notOR, err := g.MinimizeSupportUnderCondition(x, function, AigFactorOptions{})
	if err != nil || notOR.Status != AigSupportNotOR {
		t.Fatal(notOR.Status, err)
	}
}

func TestAigCutFactorErrorsLimitsAndAtomicity(t *testing.T) {
	g, _, root := aigCutTestGraph(t)
	before := g.ObjectCount()
	bad := []AigCutOptions{{MaximumCuts: 1, MaximumLeaves: 4}, {MaximumCuts: 2, MaximumLeaves: 0}, {MaximumCuts: 2, MaximumLeaves: 17}, {MaximumCuts: 2, MaximumLeaves: 4, MaximumWork: -1}}
	for _, options := range bad {
		if manager, err := ComputeAigCuts(g, options); manager != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(options, manager, err)
		}
	}
	if manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 4, MaximumLeaves: 4, MaximumWork: 1}); manager != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(manager, err)
	}
	if manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 4, MaximumLeaves: 16, MaximumWords: 1}); err != nil || manager == nil {
		t.Fatal("word bound should not apply without truth", err)
	}
	if manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 4, MaximumLeaves: 4, ComputeTruth: true, MaximumWords: 1}); manager != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(manager, err)
	}
	if manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 4, MaximumLeaves: 4, Stop: func() bool { return true }}); manager != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(manager, err)
	}
	if _, err := g.AigImplications(root, AigFactorOptions{MaximumWork: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := g.FactorAlgebraic(0, 0, true, AigFactorOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if g.ObjectCount() != before {
		t.Fatal("failed operation mutated graph")
	}
}

func FuzzAigPriorityCuts(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 7, 19, 0xdeadbeef} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		g := NewAigManager()
		literals := make([]AigLit, 5)
		for i := range literals {
			literals[i], _ = g.CreateCI()
		}
		for i := 0; i < 8; i++ {
			left := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			right := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			literal, err := g.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			literals = append(literals, literal)
		}
		root := literals[len(literals)-1]
		_, _ = g.CreateCO(root)
		manager, err := ComputeAigCuts(g, AigCutOptions{MaximumCuts: 24, MaximumLeaves: 4, ComputeTruth: true})
		if err != nil {
			t.Fatal(err)
		}
		for id, object := range g.objects {
			if object.Type != AigAND {
				continue
			}
			cuts, cutErr := manager.Cuts(AigID(id))
			if cutErr != nil {
				t.Fatal(cutErr)
			}
			for i, cut := range cuts {
				if len(cut.Leaves) == 0 || len(cut.Leaves) > 4 || !slices.IsSorted(cut.Leaves) {
					t.Fatalf("bad cut %+v", cut)
				}
				truth, truthErr := g.CutTruth(AigLit(id)<<1, cut.Leaves, AigCutTruthOptions{})
				if truthErr != nil || !cut.Truth.Equal(truth.Table) {
					t.Fatalf("truth %+v %v", cut, truthErr)
				}
				for j, other := range cuts {
					if i != j && len(cut.Leaves) >= 2 && len(other.Leaves) >= 2 && len(other.Leaves) <= len(cut.Leaves) && aigCutContained(other.Leaves, cut.Leaves) {
						t.Fatalf("dominated cuts %+v %+v", other, cut)
					}
				}
			}
		}
	})
}

func FuzzAigAlgebraicFactor(f *testing.F) {
	for _, seed := range []uint64{0, 1, 3, 17, 99, 0xabcdef} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		g := NewAigManager()
		literals := make([]AigLit, 5)
		for i := range literals {
			literals[i], _ = g.CreateCI()
		}
		internals := make([]AigLit, 0, 10)
		for len(internals) < 10 {
			left := literals[random.Intn(len(literals))].NotIf(random.Intn(3) == 0)
			right := literals[random.Intn(len(literals))].NotIf(random.Intn(3) == 0)
			literal, err := g.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			if literal.ID() == 0 || g.objects[literal.ID()].Type != AigAND {
				continue
			}
			literals = append(literals, literal)
			internals = append(internals, literal)
		}
		driverA := internals[random.Intn(len(internals))].NotIf(random.Intn(2) != 0)
		driverB := internals[random.Intn(len(internals))].NotIf(random.Intn(2) != 0)
		_, _ = g.CreateCO(driverA)
		_, _ = g.CreateCO(driverB)
		value := random.Intn(2) != 0
		result, err := g.FactorAlgebraic(0, 1, value, AigFactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Found {
			return
		}
		for assignment := 0; assignment < 1<<len(g.cis); assignment++ {
			inputs := make([]bool, len(g.cis))
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			condition, _ := aigEvalLiteral(g, driverA, inputs)
			if condition != value {
				continue
			}
			want, _ := aigEvalLiteral(g, driverB, inputs)
			got, evalErr := aigEvalLiteral(result.Graph, result.Literal, inputs)
			if evalErr != nil || got != want {
				t.Fatalf("seed=%d value=%v assignment=%d got=%v want=%v err=%v", seed, value, assignment, got, want, evalErr)
			}
		}
	})
}
