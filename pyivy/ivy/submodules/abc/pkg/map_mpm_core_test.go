package abc

// Go-only tests for the MPM subject graph and mapping core.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func mpmTestEvaluate(g *MPMSubjectGraph, assignment uint64) []bool {
	values := make([]bool, g.ObjectCount())
	for id, node := range g.nodes {
		read := func(l MPMSubjectLit) bool { return values[l.ID()] != l.Complemented() }
		switch node.kind {
		case MPMInput:
			values[id] = assignment&(uint64(1)<<uint(node.terminal)) != 0
		case MPMOutput, MPMBuffer:
			values[id] = read(node.fanins[0])
		case MPMAnd:
			values[id] = read(node.fanins[0]) && read(node.fanins[1])
		case MPMXor:
			values[id] = read(node.fanins[0]) != read(node.fanins[1])
		case MPMMux:
			values[id] = read(node.fanins[1])
			if !read(node.fanins[2]) {
				values[id] = read(node.fanins[0])
			}
		case MPMMajority:
			a, b, c := read(node.fanins[0]), read(node.fanins[1]), read(node.fanins[2])
			values[id] = a && b || a && c || b && c
		}
	}
	return values
}

func mpmTestFixture(t testing.TB, size int) (*MPMSubjectGraph, MPMParameters) {
	t.Helper()
	g, err := NewMPMSubjectGraph(64)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := g.AppendInput()
	b, _ := g.AppendInput()
	c, _ := g.AppendInput()
	x, _ := g.AppendAnd(a.Not(), b)
	y, _ := g.AppendXor(x, c.Not())
	z, _ := g.AppendMux(b, y, a)
	w, _ := g.AppendMajority(z, x, c)
	buf, _ := g.AppendBuffer(w.Not())
	if _, err = g.AppendOutput(buf); err != nil {
		t.Fatal(err)
	}
	library, err := NewSimpleMPMLUTLibrary(size)
	if err != nil {
		t.Fatal(err)
	}
	p := DefaultMPMParameters()
	p.Library, p.CutLimit, p.MinimizeSupport = library, 8, true
	return g, p
}

func mpmCheckMappingSemantics(t testing.TB, g *MPMSubjectGraph, result *MPMMappingResult) {
	t.Helper()
	for assignment := uint64(0); assignment < uint64(1)<<uint(g.InputCount()); assignment++ {
		values := mpmTestEvaluate(g, assignment)
		for _, record := range result.Records {
			var local uint64
			for i, leaf := range record.Leaves {
				if values[leaf] {
					local |= uint64(1) << uint(i)
				}
			}
			cut := MPMCut{Leaves: record.Leaves, Truth: record.Truth, Complement: record.Complement}
			if got := cut.Value(local); got != values[record.Root] {
				t.Fatalf("assignment %d root %d: mapped=%v graph=%v leaves=%v", assignment, record.Root, got, values[record.Root], record.Leaves)
			}
		}
	}
}

func TestMPMSubjectGraphAndMapping(t *testing.T) {
	g, p := mpmTestFixture(t, 3)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	if g.AndCount() != 1 || g.XorCount() != 1 || g.MuxCount() != 1 || g.TypeCount(MPMMajority) != 1 {
		t.Fatalf("type counts: and=%d xor=%d mux=%d maj=%d", g.AndCount(), g.XorCount(), g.MuxCount(), g.TypeCount(MPMMajority))
	}
	if got, err := g.SupportSize(g.outputs[0], 0, nil); err != nil || got != 3 {
		t.Fatalf("support=(%d,%v)", got, err)
	}
	if got, err := g.SmallSupportNodeCount(3, nil); err != nil || got != 4 {
		t.Fatalf("small support=(%d,%v)", got, err)
	}
	result, err := g.MapLUTs(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = result.Validate(g, p); err != nil {
		t.Fatal(err)
	}
	if result.Delay <= 0 || result.Area <= 0 || result.Edges <= 0 || result.Statistics.Rounds != 5 {
		t.Fatalf("bad summary: %+v", result)
	}
	mpmCheckMappingSemantics(t, g, result)
	if len(result.Records) == 0 {
		t.Fatal("no mapped records")
	}
	original := result.Cuts[result.Records[0].Root][0].Leaves[0]
	result.Records[0].Leaves[0] = 999
	if result.Cuts[result.Records[0].Root][0].Leaves[0] != original {
		t.Fatal("record aliases retained cuts")
	}
}

func TestMPMGiaConversionAndPackedMapping(t *testing.T) {
	gia := NewGiaManager()
	a, _ := gia.AppendCI()
	b, _ := gia.AppendCI()
	c, _ := gia.AppendCI()
	x, _ := gia.AppendAnd(a.Not(), b)
	y, _ := gia.AppendXorReal(x, c)
	z, _ := gia.AppendMuxReal(a, y, b.Not())
	buf, _ := gia.AppendBuffer(z)
	gia.AppendCO(buf.Not())
	g, err := MPMSubjectGraphFromGia(gia, 0)
	if err != nil {
		t.Fatal(err)
	}
	if g.ObjectCount() != gia.ObjectCount() || g.NodeCount() != gia.AndCount() {
		t.Fatalf("conversion counts: MPM %d/%d GIA %d/%d", g.ObjectCount(), g.NodeCount(), gia.ObjectCount(), gia.AndCount())
	}
	library, _ := NewSimpleMPMLUTLibrary(3)
	p := DefaultMPMParameters()
	p.Library = library
	result, err := g.MapLUTs(p)
	if err != nil {
		t.Fatal(err)
	}
	mpmCheckMappingSemantics(t, g, result)
	packed, err := result.GiaMapping(0)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := packed.Verify(gia, GiaLUTAnalysisOptions{}); err != nil || !ok {
		t.Fatalf("packed verify=(%v,%v)", ok, err)
	}
}

func TestMPMTruthMinimizationChoiceAndOrdering(t *testing.T) {
	g, _ := NewMPMSubjectGraph(32)
	a, _ := g.AppendInput()
	b, _ := g.AppendInput()
	c, _ := g.AppendInput()
	x, _ := g.AppendXor(a, b)
	y, _ := g.AppendAnd(x, c)
	z, _ := g.AppendAnd(x, c)
	if err := g.SetChoice(z.ID(), y.ID(), false); err != nil {
		t.Fatal(err)
	}
	g.AppendOutput(z)
	library, _ := NewSimpleMPMLUTLibrary(3)
	p := DefaultMPMParameters()
	p.Library, p.MinimizeSupport = library, true
	first, err := g.MapLUTs(p)
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.MapLUTs(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Records) != len(second.Records) {
		t.Fatal("nondeterministic record count")
	}
	for i := range first.Records {
		if first.Records[i].Root != second.Records[i].Root || !slices.Equal(first.Records[i].Leaves, second.Records[i].Leaves) || !slices.Equal(first.Records[i].Truth, second.Records[i].Truth) {
			t.Fatal("nondeterministic mapping")
		}
	}
	mpmCheckMappingSemantics(t, g, first)
}

func TestMPMWideTruthAndVariablePinDelays(t *testing.T) {
	g, _ := NewMPMSubjectGraph(64)
	inputs := make([]MPMSubjectLit, 8)
	for i := range inputs {
		inputs[i], _ = g.AppendInput()
	}
	root := inputs[0]
	for _, input := range inputs[1:] {
		root, _ = g.AppendAnd(root, input)
	}
	g.AppendOutput(root)
	library, _ := NewSimpleMPMLUTLibrary(8)
	library.VariablePinDelays = true
	library.Areas[8] = 40
	for i := 0; i < 8; i++ {
		library.Delays[8][i] = 1
	}
	p := DefaultMPMParameters()
	p.Library, p.OneRound, p.CutLimit = library, true, 16
	result, err := g.MapLUTs(p)
	if err != nil {
		t.Fatal(err)
	}
	mpmCheckMappingSemantics(t, g, result)
	last := result.Cuts[root.ID()][0]
	if len(last.Leaves) != 8 || len(last.Truth) != 4 || last.Arrival != 1 {
		t.Fatalf("wide cut: leaves=%d words=%d arrival=%d", len(last.Leaves), len(last.Truth), last.Arrival)
	}

	pins, _ := NewMPMSubjectGraph(8)
	a, _ := pins.AppendInput()
	b, _ := pins.AppendInput()
	and, _ := pins.AppendAnd(a, b)
	pins.AppendOutput(and)
	pinLibrary, _ := NewSimpleMPMLUTLibrary(2)
	pinLibrary.VariablePinDelays = true
	pinLibrary.Delays[2][0], pinLibrary.Delays[2][1] = 5, 2
	p.Library = pinLibrary
	pinResult, err := pins.MapLUTs(p)
	if err != nil {
		t.Fatal(err)
	}
	if pinResult.Delay != 5 {
		t.Fatalf("pin delay=%d", pinResult.Delay)
	}
}

func TestMPMFailuresAreAtomic(t *testing.T) {
	g, p := mpmTestFixture(t, 2)
	before := mpmTestEvaluate(g, 5)
	p.Stop = func() bool { return true }
	if result, err := g.MapLUTs(p); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel=(%v,%v)", result, err)
	}
	p.Stop, p.MaxCombinations = nil, 1
	if result, err := g.MapLUTs(p); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("limit=(%v,%v)", result, err)
	}
	if after := mpmTestEvaluate(g, 5); !slices.Equal(before, after) {
		t.Fatal("failed map mutated source")
	}
	if got, err := g.SupportSize(g.outputs[0], 1, nil); got != 0 || !errors.Is(err, ErrLimit) {
		t.Fatalf("support limit=(%d,%v)", got, err)
	}
	if _, err := g.SupportSize(g.outputs[0], 0, func() bool { return true }); !errors.Is(err, ErrCancelled) {
		t.Fatalf("support cancel=%v", err)
	}
	bad := DefaultMPMParameters()
	if result, err := g.MapLUTs(bad); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("missing library=(%v,%v)", result, err)
	}
}

func mpmRandomGraph(seed int64) (*MPMSubjectGraph, error) {
	rng := rand.New(rand.NewSource(seed))
	g, _ := NewMPMSubjectGraph(96)
	signals := []MPMSubjectLit{}
	for i := 0; i < 2+int(seed%4+4)%4; i++ {
		l, _ := g.AppendInput()
		signals = append(signals, l)
	}
	for i := 0; i < 3+int(seed%9+9)%9; i++ {
		pick := func() MPMSubjectLit {
			l := signals[rng.Intn(len(signals))]
			if rng.Intn(2) == 1 {
				l = l.Not()
			}
			return l
		}
		var l MPMSubjectLit
		var err error
		switch rng.Intn(4) {
		case 0:
			l, err = g.AppendAnd(pick(), pick())
		case 1:
			l, err = g.AppendXor(pick(), pick())
		case 2:
			l, err = g.AppendMux(pick(), pick(), pick())
		default:
			l, err = g.AppendMajority(pick(), pick(), pick())
		}
		if err != nil {
			return nil, err
		}
		signals = append(signals, l)
	}
	_, err := g.AppendOutput(signals[len(signals)-1])
	return g, err
}

func TestMPMIndependentRandomModel5000(t *testing.T) {
	library, _ := NewSimpleMPMLUTLibrary(4)
	for seed := int64(0); seed < 5000; seed++ {
		g, err := mpmRandomGraph(seed)
		if err != nil {
			t.Fatalf("seed %d graph: %v", seed, err)
		}
		p := DefaultMPMParameters()
		p.Library = library
		p.CutLimit = 5
		p.OneRound = true
		p.MaxCombinations = 1 << 18
		result, err := g.MapLUTs(p)
		if err != nil {
			t.Fatalf("seed %d map: %v", seed, err)
		}
		if err = result.Validate(g, p); err != nil {
			t.Fatalf("seed %d validate: %v", seed, err)
		}
		mpmCheckMappingSemantics(t, g, result)
	}
}

func FuzzMPMMappingModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 7, 17, 63, 255, 1024} {
		f.Add(seed)
	}
	library, _ := NewSimpleMPMLUTLibrary(4)
	f.Fuzz(func(t *testing.T, seed uint64) {
		g, err := mpmRandomGraph(int64(seed))
		if err != nil {
			t.Skip()
		}
		p := DefaultMPMParameters()
		p.Library = library
		p.CutLimit = 5
		p.OneRound = true
		p.MaxCombinations = 1 << 18
		result, err := g.MapLUTs(p)
		if err != nil {
			t.Fatalf("map: %v", err)
		}
		if err = result.Validate(g, p); err != nil {
			t.Fatal(err)
		}
		mpmCheckMappingSemantics(t, g, result)
	})
}
