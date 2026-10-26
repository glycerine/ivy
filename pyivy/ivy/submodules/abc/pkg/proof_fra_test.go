package abc

// Standalone Go tests for the src/proof/fra port. Expected results are derived
// from direct Boolean and explicit-state models, never from native ABC code.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func fraTestCombinational(t *testing.T) (*AigManager, AigLit, AigLit, AigLit, AigLit) {
	t.Helper()
	g := NewAigManager()
	a, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	f, err := g.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	either, err := g.Or(a, b)
	if err != nil {
		t.Fatal(err)
	}
	redundant, err := g.And(f, either)
	if err != nil {
		t.Fatal(err)
	}
	gEquivalent, err := g.CreateBuffer(redundant)
	if err != nil {
		t.Fatal(err)
	}
	notA, err := g.CreateBuffer(a.Not())
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range []AigLit{f, gEquivalent, notA, a} {
		if _, err = g.CreateCO(output); err != nil {
			t.Fatal(err)
		}
	}
	if err = g.Check(); err != nil {
		t.Fatal(err)
	}
	return g, a, b, f, gEquivalent
}

func fraPatterns(width int) []FRAPattern {
	patterns := make([]FRAPattern, 1<<uint(width))
	for mask := range patterns {
		patterns[mask].Inputs = make([]bool, width)
		for bit := 0; bit < width; bit++ {
			patterns[mask].Inputs[bit] = mask>>uint(bit)&1 != 0
		}
	}
	return patterns
}

func TestFRASourceInventory(t *testing.T) {
	units := FRASourceInventory()
	if len(units) != 20 {
		t.Fatalf("inventory has %d units", len(units))
	}
	want := []string{"fra.h", "fra_.c", "fraBmc.c", "fraCec.c", "fraClass.c", "fraClau.c", "fraClaus.c", "fraCnf.c", "fraCore.c", "fraHot.c", "fraImp.c", "fraInd.c", "fraIndVer.c", "fraLcr.c", "fraMan.c", "fraPart.c", "fraSat.c", "fraSec.c", "fraSim.c", "module.make"}
	for i := range want {
		if units[i].File != want[i] || units[i].Disposition == "" {
			t.Fatalf("inventory[%d]=%+v", i, units[i])
		}
	}
	copyUnits := FRASourceInventory()
	copyUnits[0].File = "changed"
	if FRASourceInventory()[0].File != "fra.h" {
		t.Fatal("inventory is not owned")
	}
}

func TestFRASignaturesClassesAndSAT(t *testing.T) {
	g, a, b, f, equivalent := fraTestCombinational(t)
	m, err := NewFRAManager(g, FRAOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	patterns := fraPatterns(2)
	sigs, err := m.Signatures(patterns)
	if err != nil {
		t.Fatal(err)
	}
	for id := range sigs {
		for mask, p := range patterns {
			values := make([]bool, g.ObjectCount())
			values[0] = true
			values[a.ID()] = p.Inputs[0]
			values[b.ID()] = p.Inputs[1]
			for n, obj := range g.objects {
				read := func(l AigLit) bool { return values[l.ID()] != l.Complemented() }
				switch obj.Type {
				case AigAND:
					values[n] = read(obj.Fanin0) && read(obj.Fanin1)
				case AigBuffer, AigCO:
					values[n] = read(obj.Fanin0)
				}
			}
			got := sigs[id].Bits[0]>>uint(mask)&1 != 0
			if got != values[id] {
				t.Fatalf("signature node %d mask %d", id, mask)
			}
		}
	}
	classes, err := m.Classes(patterns)
	if err != nil {
		t.Fatal(err)
	}
	foundEquivalent, foundComplement := false, false
	for _, class := range classes {
		for _, member := range class.Members {
			if class.Representative == f.ID() && member.Node == equivalent.ID() && !member.Complement {
				foundEquivalent = true
			}
			if class.Representative == a.ID() && member.Complement {
				foundComplement = true
			}
		}
	}
	if !foundEquivalent || !foundComplement {
		t.Fatalf("classes=%+v", classes)
	}
	verified, err := m.VerifiedClasses(patterns)
	if err != nil {
		t.Fatal(err)
	}
	if len(verified) == 0 {
		t.Fatal("no SAT-verified classes")
	}
	proof, err := m.VerifyEquivalent(f, equivalent)
	if err != nil || proof.Status != FraigRelationProved {
		t.Fatalf("proof=%+v err=%v", proof, err)
	}
	disproof, err := m.VerifyEquivalent(a, b)
	if err != nil || disproof.Status != FraigRelationDisproved {
		t.Fatalf("disproof=%+v err=%v", disproof, err)
	}
	if len(disproof.Witness.Inputs) != 2 || disproof.Witness.Inputs[0] == disproof.Witness.Inputs[1] {
		t.Fatalf("bad witness %+v", disproof.Witness)
	}
	owned := disproof.Witness.Clone()
	disproof.Witness.Inputs[0] = !disproof.Witness.Inputs[0]
	if slices.Equal(owned.Inputs, disproof.Witness.Inputs) {
		t.Fatal("witness clone aliases")
	}
}

func TestFRASweepAtomicAndPhaseCorrect(t *testing.T) {
	g, _, _, f, equivalent := fraTestCombinational(t)
	originalObjects, originalAnds := g.ObjectCount(), g.AndCount()
	m, err := NewFRAManager(g, FRAOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r, err := m.Sweep(fraPatterns(2))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Mapping) != g.ObjectCount() || r.Mapping[f.ID()] != r.Mapping[equivalent.ID()] {
		t.Fatalf("mapping did not merge equivalent nodes: %v %v", r.Mapping[f.ID()], r.Mapping[equivalent.ID()])
	}
	for mask := 0; mask < 4; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0}
		before, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		after, e := r.Graph.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(before, after) {
			t.Fatalf("mask %d: %v != %v", mask, before, after)
		}
	}
	if g.ObjectCount() != originalObjects || g.AndCount() != originalAnds {
		t.Fatal("source graph mutated")
	}
	if err = r.Graph.Check(); err != nil {
		t.Fatal(err)
	}
	if m.Stats().VerifiedMerges == 0 {
		t.Fatal("no verified merge recorded")
	}
}

func fraTestSequential(t *testing.T) (*AigManager, AigLit, AigLit, AigLit) {
	t.Helper()
	g := NewAigManager()
	pi, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	latch, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	copyLatch, err := g.CreateBuffer(latch)
	if err != nil {
		t.Fatal(err)
	}
	next, err := g.Xor(pi, latch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCO(latch); err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = g.Check(); err != nil {
		t.Fatal(err)
	}
	return g, pi, latch, copyLatch
}

func TestFRASequentialFramesWitnessAndReplay(t *testing.T) {
	g, pi, latch, copyLatch := fraTestSequential(t)
	m, err := NewFRAManager(g, FRAOptions{MaxFrames: 8, MaxDepth: 8})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	proved, err := m.VerifySequentialEquivalent(latch, copyLatch, 4)
	if err != nil || proved.Status != FraigRelationProved {
		t.Fatalf("proved=%+v err=%v", proved, err)
	}
	if !proved.Bounded || proved.BoundedFrames != 4 {
		t.Fatalf("bounded proof mislabeled: %+v", proved)
	}
	disproved, err := m.VerifySequentialEquivalent(pi, latch, 3)
	if err != nil || disproved.Status != FraigRelationDisproved {
		t.Fatalf("disproved=%+v err=%v", disproved, err)
	}
	if !disproved.Bounded || disproved.BoundedFrames != len(disproved.Witness.Frames) {
		t.Fatalf("witness bound mismatch: %+v", disproved)
	}
	if err = m.ReplaySequential(pi, latch, disproved.Witness); err != nil {
		t.Fatalf("replay: %v witness=%+v", err, disproved.Witness)
	}
	bad := disproved.Witness.Clone()
	for i := range bad.Frames {
		for j := range bad.Frames[i] {
			bad.Frames[i][j] = false
		}
	}
	if err = m.ReplaySequential(pi, latch, bad); err == nil {
		t.Fatal("accepted non-distinguishing trace")
	}
	bad = disproved.Witness.Clone()
	bad.States[0][0] = true
	if err = m.ReplaySequential(pi, latch, bad); !errors.Is(err, ErrDimension) {
		t.Fatalf("accepted nonzero initial state: %v", err)
	}
	bad = disproved.Witness.Clone()
	bad.States = nil
	if err = m.ReplaySequential(pi, latch, bad); !errors.Is(err, ErrDimension) {
		t.Fatalf("accepted missing state replay: %v", err)
	}
	if err = m.ReplaySequential(InvalidAigLit, latch, disproved.Witness); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid replay literal: %v", err)
	}
}

func TestFRASequentialPatternsAreCumulative(t *testing.T) {
	g, _, latch, copyLatch := fraTestSequential(t)
	m, err := NewFRAManager(g, FRAOptions{MaxFrames: 4, MaxDepth: 4, MaxPatterns: 3})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err = m.VerifySequentialEquivalent(latch, copyLatch, 2); !errors.Is(err, ErrLimit) {
		t.Fatalf("unbounded transition patterns: %v", err)
	}
	if m.Stats().Patterns != 3 {
		t.Fatalf("pattern accounting: %+v", m.Stats())
	}
}

func TestFRASequentialExactMultiFrameReplay(t *testing.T) {
	g, _, latch, _ := fraTestSequential(t)
	m, err := NewFRAManager(g, FRAOptions{MaxFrames: 4, MaxDepth: 4})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	result, err := m.VerifySequentialEquivalent(latch, AigConst0, 2)
	if err != nil || result.Status != FraigRelationDisproved || len(result.Witness.Frames) != 2 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if err = m.ReplaySequential(latch, AigConst0, result.Witness); err != nil {
		t.Fatal(err)
	}
	bad := result.Witness.Clone()
	bad.States[1][0] = !bad.States[1][0]
	if err = m.ReplaySequential(latch, AigConst0, bad); !errors.Is(err, ErrDimension) {
		t.Fatalf("accepted altered intermediate state: %v", err)
	}
	bad = result.Witness.Clone()
	bad.Frames = bad.Frames[1:]
	bad.States = bad.States[1:]
	if err = m.ReplaySequential(latch, AigConst0, bad); !errors.Is(err, ErrDimension) {
		t.Fatalf("accepted nonzero-state suffix: %v", err)
	}
}

func TestFRALimitsCancellationMalformedAndLifecycle(t *testing.T) {
	g, a, _, _, _ := fraTestCombinational(t)
	cases := []FRAOptions{{MaxNodes: 1}, {MaxClauses: 1}, {MaxDepth: 1}, {MaxBytes: 1}}
	for _, o := range cases {
		if m, err := NewFRAManager(g, o); m != nil || !errors.Is(err, ErrLimit) {
			t.Fatalf("options %+v: m=%v err=%v", o, m, err)
		}
	}
	if m, err := NewFRAManager(g, FRAOptions{Stop: func() bool { return true }}); m != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel m=%v err=%v", m, err)
	}
	corrupt := g.Clone()
	corrupt.objects[a.ID()].Type = AigAND
	if m, err := NewFRAManager(corrupt, FRAOptions{}); m != nil || !errors.Is(err, ErrInvariant) {
		t.Fatalf("malformed m=%v err=%v", m, err)
	}
	m, err := NewFRAManager(g, FRAOptions{MaxPatterns: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Signatures(fraPatterns(2)); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Signatures([]FRAPattern{{Inputs: []bool{false, false}}}); !errors.Is(err, ErrLimit) {
		t.Fatalf("cumulative patterns: %v", err)
	}
	clone := m.Graph()
	m.Close()
	m.Close()
	if m.Graph() != nil {
		t.Fatal("closed graph exposed")
	}
	if _, err = m.VerifyEquivalent(a, a); !errors.Is(err, ErrFRAClosed) {
		t.Fatalf("closed: %v", err)
	}
	if clone == nil || clone.Check() != nil {
		t.Fatal("owned graph invalid after close")
	}
	if _, err = NewFRAManager(nil, FRAOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("nil graph: %v", err)
	}
	cancel := false
	cancellable, err := NewFRAManager(g, FRAOptions{Stop: func() bool { return cancel }})
	if err != nil {
		t.Fatal(err)
	}
	cancel = true
	beforeObjects := g.ObjectCount()
	if result, sweepErr := cancellable.Sweep(fraPatterns(2)); !errors.Is(sweepErr, ErrCancelled) || result.Graph != nil {
		t.Fatalf("cancelled sweep result=%+v err=%v", result, sweepErr)
	}
	if g.ObjectCount() != beforeObjects || g.Check() != nil {
		t.Fatal("cancelled sweep mutated source")
	}
}

func TestFRADeterministic5000IndependentCases(t *testing.T) {
	g, a, b, f, equivalent := fraTestCombinational(t)
	m, err := NewFRAManager(g, FRAOptions{MaxPatterns: 6000, MaxWork: 1 << 24, MaxBytes: 1 << 24})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	rng := rand.New(rand.NewSource(0x465241))
	patterns := make([]FRAPattern, 5000)
	for i := range patterns {
		av, bv := rng.Intn(2) != 0, rng.Intn(2) != 0
		patterns[i] = FRAPattern{Inputs: []bool{av, bv}}
		values, err := m.values(patterns[i].Inputs)
		if err != nil {
			t.Fatal(err)
		}
		if values[f.ID()] != (av && bv) || values[equivalent.ID()] != (av && bv) || values[a.ID()] != av || values[b.ID()] != bv {
			t.Fatalf("case %d", i)
		}
	}
	sigs, err := m.Signatures(patterns)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(sigs[f.ID()].Bits, sigs[equivalent.ID()].Bits) {
		t.Fatal("5000-case signatures disagree")
	}
}

func TestFRAFuzzCampaign5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0x66757a7a))
	for trial := 0; trial < 5000; trial++ {
		g := NewAigManager()
		x, _ := g.CreateCI()
		y, _ := g.CreateCI()
		left, right := x, y
		if rng.Intn(2) != 0 {
			left = left.Not()
		}
		if rng.Intn(2) != 0 {
			right = right.Not()
		}
		node, err := g.And(left, right)
		if err != nil {
			t.Fatal(err)
		}
		copyNode, err := g.CreateBuffer(node)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = g.CreateCO(node)
		_, _ = g.CreateCO(copyNode)
		m, err := NewFRAManager(g, FRAOptions{MaxPatterns: 4, MaxWork: 10000, MaxBytes: 1 << 16})
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		proof, err := m.VerifyEquivalent(node, copyNode)
		if err != nil || proof.Status != FraigRelationProved {
			t.Fatalf("trial %d proof=%+v err=%v", trial, proof, err)
		}
		for mask := 0; mask < 4; mask++ {
			av, bv := mask&1 != 0, mask&2 != 0
			want := (av != left.Complemented()) && (bv != right.Complemented())
			out, err := g.Evaluate([]bool{av, bv})
			if err != nil || out[0] != want || out[1] != want {
				t.Fatalf("trial %d mask %d out=%v want=%v err=%v", trial, mask, out, want, err)
			}
		}
		m.Close()
	}
}

func FuzzFRAEquivalentBuffer(f *testing.F) {
	for _, seed := range []byte{0, 1, 2, 3, 0xff} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, controls byte) {
		g := NewAigManager()
		x, err := g.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
		y, err := g.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
		left, right := x.NotIf(controls&1 != 0), y.NotIf(controls&2 != 0)
		node, err := g.And(left, right)
		if err != nil {
			t.Fatal(err)
		}
		if controls&4 != 0 {
			node = node.Not()
		}
		copyNode, err := g.CreateBuffer(node)
		if err != nil {
			t.Fatal(err)
		}
		m, err := NewFRAManager(g, FRAOptions{MaxWork: 10000, MaxBytes: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		proof, err := m.VerifyEquivalent(node, copyNode)
		if err != nil || proof.Status != FraigRelationProved {
			t.Fatalf("controls=%x proof=%+v err=%v", controls, proof, err)
		}
	})
}
