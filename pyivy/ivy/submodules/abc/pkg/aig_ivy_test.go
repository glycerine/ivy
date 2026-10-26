package abc

// Standalone tests for the remaining src/aig/ivy behavior.

import (
	"context"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

func ivyExtendedFixture(t testing.TB) (*IvyManager, []IvyEdge, IvyEdge) {
	t.Helper()
	m, err := NewIvyManager(IvyOptions{MaxObjects: 10000, MaxWork: 100000})
	if err != nil {
		t.Fatal(err)
	}
	pis := make([]IvyEdge, 3)
	for i := range pis {
		pis[i], err = m.CreatePI()
		if err != nil {
			t.Fatal(err)
		}
	}
	a, err := m.And(pis[0], pis[1])
	if err != nil {
		t.Fatal(err)
	}
	root, err := m.And(a, pis[2])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.CreatePO(root); err != nil {
		t.Fatal(err)
	}
	if err = m.Check(); err != nil {
		t.Fatal(err)
	}
	return m, pis, root
}

func ivyTruthBit(words []uint64, assignment int) bool {
	return words[assignment/64]&(1<<uint(assignment%64)) != 0
}

func TestIvyAttributesAndInventory(t *testing.T) {
	if len(IvyExtendedSources) != 22 {
		t.Fatalf("build inventory %d", len(IvyExtendedSources))
	}
	if IvyExtendedSources[0] != "ivyBalance.c" || IvyExtendedSources[len(IvyExtendedSources)-1] != "ivyUtil.c" {
		t.Fatalf("inventory endpoints %v", IvyExtendedSources)
	}
	if len(IvyExtendedCoverage) != 19 {
		t.Fatalf("coverage inventory %d", len(IvyExtendedCoverage))
	}
	if IvyExtendedCoverage[8] != "ivyFraig.c:validated:deterministic simulation classes and SAT-certified full-support reduction/miter proof" {
		t.Fatalf("Ivy FRAIG coverage regressed to a non-validated disposition: %q", IvyExtendedCoverage[8])
	}
	a, err := NewIvyIntAttributes(2, 8)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Write(5, 17); err != nil {
		t.Fatal(err)
	}
	if got, _ := a.Read(5); got != 17 {
		t.Fatalf("int attr %d", got)
	}
	if got, _ := a.Read(7); got != 0 {
		t.Fatalf("default attr %d", got)
	}
	snap, err := a.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	snap[5] = 99
	if got, _ := a.Read(5); got != 17 {
		t.Fatal("snapshot aliases storage")
	}
	if err = a.Write(8, 1); err != ErrDimension {
		t.Fatalf("limit %v", err)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Read(0); err != ErrInvariant {
		t.Fatalf("closed read %v", err)
	}
	b, err := NewIvyByteAttributes(1, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	v := []byte{1, 2, 3}
	if err = b.Write(2, v); err != nil {
		t.Fatal(err)
	}
	v[0] = 9
	got, err := b.Read(2)
	if err != nil || !slices.Equal(got, []byte{1, 2, 3}) {
		t.Fatalf("byte attr %v %v", got, err)
	}
	got[1] = 8
	again, _ := b.Read(2)
	if again[1] != 2 {
		t.Fatal("byte read aliases storage")
	}
	if err = b.Write(1, []byte{1}); err != ErrDimension {
		t.Fatalf("width %v", err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestIvyCutsTruthMFFCMapAndProfiles(t *testing.T) {
	m, pis, root := ivyExtendedFixture(t)
	o := IvyExtendedOptions{MaxObjects: 1000, MaxEdges: 2000, MaxCuts: 1000, MaxLeaves: 6, MaxTruth: 64, MaxBytes: 1 << 20, MaxWork: 100000, MaxDepth: 100}
	truth, err := m.CutTruth(root, []uint32{pis[0].ID(), pis[1].ID(), pis[2].ID()}, o)
	if err != nil {
		t.Fatal(err)
	}
	for a := 0; a < 8; a++ {
		want := a == 7
		if ivyTruthBit(truth, a) != want {
			t.Fatalf("truth %d got %v", a, ivyTruthBit(truth, a))
		}
	}
	cuts, err := m.EnumerateCuts(3, o)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range cuts[root.ID()] {
		if slices.Equal(c.Leaves, []uint32{1, 2, 3}) {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing PI cut %#v", cuts[root.ID()])
	}
	mffc, err := m.MFFC(root.ID(), o)
	if err != nil {
		t.Fatal(err)
	}
	if len(mffc.Nodes) != 2 || !slices.Equal(mffc.Boundary, []uint32{1, 2, 3}) {
		t.Fatalf("mffc %#v", mffc)
	}
	mapping, err := m.FastMap(3, o)
	if err != nil {
		t.Fatal(err)
	}
	if mapping.Delay != 1 || mapping.Area != 2 {
		t.Fatalf("map %#v", mapping)
	}
	stats, err := m.ExtendedStats(o)
	if err != nil {
		t.Fatal(err)
	}
	if stats.ANDs != 2 || stats.Inputs != 3 || stats.Outputs != 1 || stats.Edges != 5 {
		t.Fatalf("stats %#v", stats)
	}
	hp, err := m.HashProfile(17, o)
	if err != nil {
		t.Fatal(err)
	}
	if hp.Entries != 2 || len(hp.Buckets) != 17 {
		t.Fatalf("hash %#v", hp)
	}
	if sum := func() int {
		n := 0
		for _, v := range hp.Buckets {
			n += v
		}
		return n
	}(); sum != 2 {
		t.Fatalf("buckets %d", sum)
	}
	dot, err := m.FormatDOT([]uint32{root.ID()}, o)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dot, "digraph ivy") || !strings.Contains(dot, "style=bold") {
		t.Fatalf("dot output %q", dot)
	}
}

func TestIvyStepBufferAssertChoices(t *testing.T) {
	m, err := NewIvyManager(IvyOptions{MaxObjects: 1000, MaxWork: 10000})
	if err != nil {
		t.Fatal(err)
	}
	p, _ := m.CreatePI()
	latch, err := m.Latch(p, IvyInit0)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := m.CreateBuffer(latch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.CreatePO(buf); err != nil {
		t.Fatal(err)
	}
	if _, err = m.CreateAssert(p.Not()); err != nil {
		t.Fatal(err)
	}
	out, next, err := m.IvyStep([]bool{true}, []bool{false}, IvyExtendedOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0] || !slices.Equal(next, []bool{true}) {
		t.Fatalf("step %v %v", out, next)
	}
	c, err := NewIvyChoices(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Add(m, buf.ID(), []IvyEdge{latch, latch, buf}, IvyExtendedOptions{}); err != nil {
		t.Fatal(err)
	}
	snap := c.Snapshot()
	if len(snap) != 1 || len(snap[0].Equivalents) != 2 {
		t.Fatalf("choices %#v", snap)
	}
	snap[0].Equivalents[0] = 0
	if c.Snapshot()[0].Equivalents[0] == 0 {
		t.Fatal("choice snapshot alias")
	}
}

func TestIvyVerifiedChoicesAndEquivalence(t *testing.T) {
	m, err := NewIvyManager(IvyOptions{MaxObjects: 100, MaxWork: 10000})
	if err != nil {
		t.Fatal(err)
	}
	p, _ := m.CreatePI()
	q, _ := m.CreatePI()
	buf, err := m.CreateBuffer(p)
	if err != nil {
		t.Fatal(err)
	}
	and, err := m.And(p, q)
	if err != nil {
		t.Fatal(err)
	}
	o := IvyExtendedOptions{MaxLeaves: 4, MaxTruth: 16, MaxWork: 10000, MaxBytes: 1 << 20}
	equal, err := m.Equivalent(p, buf, o)
	if err != nil || !equal {
		t.Fatalf("equivalence %v %v", equal, err)
	}
	equal, err = m.Equivalent(p, and, o)
	if err != nil || equal {
		t.Fatalf("inequivalence %v %v", equal, err)
	}
	c, err := NewIvyChoices(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.AddVerified(m, p, []IvyEdge{buf}, o); err != nil {
		t.Fatal(err)
	}
	if err = c.AddVerified(m, and, []IvyEdge{p}, o); err != ErrInvariant {
		t.Fatalf("invalid class %v", err)
	}
}

func TestIvyBalanceCanonicalAndDOT(t *testing.T) {
	m, err := NewIvyManager(IvyOptions{MaxObjects: 10000, MaxWork: 100000})
	if err != nil {
		t.Fatal(err)
	}
	pis := make([]IvyEdge, 6)
	for i := range pis {
		pis[i], err = m.CreatePI()
		if err != nil {
			t.Fatal(err)
		}
	}
	root := pis[0]
	for i := 1; i < len(pis); i++ {
		root, err = m.And(root, pis[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = m.CreatePO(root); err != nil {
		t.Fatal(err)
	}
	balanced, err := m.BalanceCombinational(IvyExtendedOptions{MaxDepth: 100, MaxWork: 100000})
	if err != nil {
		t.Fatal(err)
	}
	if err = balanced.Check(); err != nil {
		t.Fatal(err)
	}
	for a := 0; a < 64; a++ {
		in := make([]bool, 6)
		for i := range in {
			in[i] = a&(1<<uint(i)) != 0
		}
		x, e := m.Evaluate(in, nil)
		if e != nil {
			t.Fatal(e)
		}
		y, e := balanced.Evaluate(in, nil)
		if e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(x, y) {
			t.Fatalf("balance assignment %d %v %v", a, x, y)
		}
	}
	orig, _ := m.Object(root.ID())
	brot := balanced.objects[balanced.outputs[0]].Fanin0.ID()
	bo, _ := balanced.Object(brot)
	if bo.Level >= orig.Level {
		t.Fatalf("not balanced old %d new %d", orig.Level, bo.Level)
	}
	res, err := m.Resynthesize(IvyExtendedOptions{MaxDepth: 100, MaxWork: 100000})
	if err != nil {
		t.Fatal(err)
	}
	if err = res.Check(); err != nil {
		t.Fatal(err)
	}
	canon, err := IvyCanonicalTruth(0x8, 2, IvyExtendedOptions{MaxWork: 10000})
	if err != nil {
		t.Fatal(err)
	}
	if canon.Truth != 1 {
		t.Fatalf("canonical %#v", canon)
	}
	all, err := IvyCanonicalTruth(^uint64(0), 6, IvyExtendedOptions{MaxWork: 10000000})
	if err != nil {
		t.Fatal(err)
	}
	if all.Truth != 0 || !all.OutputComplement {
		t.Fatalf("constant canonical %#v", all)
	}
}

func TestIvyExtendedLimitsCancellationAndErrors(t *testing.T) {
	m, pis, root := ivyExtendedFixture(t)
	if _, err := m.CutTruth(root, []uint32{pis[0].ID(), pis[1].ID(), pis[2].ID()}, IvyExtendedOptions{MaxTruth: 4}); err != ErrLimit {
		t.Fatalf("truth limit %v", err)
	}
	if _, err := m.EnumerateCuts(3, IvyExtendedOptions{MaxCuts: 2}); err != ErrLimit {
		t.Fatalf("cut limit %v", err)
	}
	if _, err := m.EnumerateCuts(3, IvyExtendedOptions{MaxBytes: 32}); err != ErrLimit {
		t.Fatalf("cut bytes %v", err)
	}
	if _, err := m.FormatDOT(nil, IvyExtendedOptions{MaxBytes: 8}); err != ErrLimit {
		t.Fatalf("text limit %v", err)
	}
	if _, err := m.ExtendedStats(IvyExtendedOptions{Stop: func() bool { return true }}); err != ErrCancelled {
		t.Fatalf("cancel %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !ivyContextStop(ctx, nil)() {
		t.Fatal("context stop")
	}
	if _, err := IvyCanonicalTruth(0, 7, IvyExtendedOptions{}); err != ErrDimension {
		t.Fatalf("vars %v", err)
	}
	if _, err := m.CutTruth(root, []uint32{pis[0].ID()}, IvyExtendedOptions{}); err != ErrDimension {
		t.Fatalf("incomplete cut %v", err)
	}
	if _, err := m.BalanceCombinational(IvyExtendedOptions{MaxObjects: 2}); err != ErrLimit {
		t.Fatalf("balance object limit %v", err)
	}
	bad := m.snapshot()
	m.objects[root.ID()].Fanin0 = IvyEdge(uint32(len(m.objects)+1) << 1)
	if _, err := m.ExtendedStats(IvyExtendedOptions{}); err == nil {
		t.Fatal("malformed accepted")
	}
	m.restore(bad)
	if err := m.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestIvyExtendedCumulativeBudgetsAndMalformedUnary(t *testing.T) {
	m, pis, root := ivyExtendedFixture(t)
	minimum := func(name string, call func(int) error) int {
		t.Helper()
		for limit := 1; limit <= 5000; limit++ {
			err := call(limit)
			if err == nil {
				return limit
			}
			if err != ErrLimit {
				t.Fatalf("%s limit %d: %v", name, limit, err)
			}
		}
		t.Fatalf("%s never completed", name)
		return 0
	}
	cutWork := minimum("cuts", func(limit int) error {
		_, err := m.EnumerateCuts(3, IvyExtendedOptions{MaxWork: limit})
		return err
	})
	mapWork := minimum("map", func(limit int) error {
		_, err := m.FastMap(3, IvyExtendedOptions{MaxWork: limit})
		return err
	})
	if mapWork <= cutWork {
		t.Fatalf("mapping did not include cut work: cuts=%d map=%d", cutWork, mapWork)
	}
	truthWork := minimum("truth", func(limit int) error {
		_, err := m.CutTruth(root, []uint32{pis[0].ID(), pis[1].ID(), pis[2].ID()}, IvyExtendedOptions{MaxWork: limit})
		return err
	})
	equivalentWork := minimum("equivalent", func(limit int) error {
		_, err := m.Equivalent(root, root, IvyExtendedOptions{MaxLeaves: 3, MaxTruth: 8, MaxWork: limit})
		return err
	})
	if equivalentWork <= truthWork {
		t.Fatalf("equivalence did not include both truth computations: truth=%d equivalent=%d", truthWork, equivalentWork)
	}
	if _, _, err := m.IvyStep(make([]bool, len(pis)), nil, IvyExtendedOptions{MaxBytes: 1}); err != ErrLimit {
		t.Fatalf("step byte limit %v", err)
	}
	if _, err := NewIvyIntAttributes(0, ivyAttributeMaxBytes/8+1); err != ErrLimit {
		t.Fatalf("int attribute allocation bound %v", err)
	}
	if _, err := NewIvyByteAttributes(0, 0, ivyAttributeMaxBytes/24+1); err != ErrLimit {
		t.Fatalf("byte attribute table bound %v", err)
	}
	var nilManager *IvyManager
	if _, err := nilManager.CreateBuffer(IvyConst1); err == nil {
		t.Fatal("nil buffer manager accepted")
	}
	if _, err := nilManager.CreateAssert(IvyConst1); err == nil {
		t.Fatal("nil assertion manager accepted")
	}

	corrupt, err := NewIvyManager(IvyOptions{MaxObjects: 16})
	if err != nil {
		t.Fatal(err)
	}
	pi, _ := corrupt.CreatePI()
	buffer, _ := corrupt.CreateBuffer(pi)
	corrupt.objects[pi.ID()].Refs--
	corrupt.objects[buffer.ID()].Fanin0 = buffer
	corrupt.objects[buffer.ID()].Refs++
	if err = corrupt.Check(); err != nil {
		t.Fatalf("fixture should expose core unary-check gap: %v", err)
	}
	if _, err = corrupt.ExtendedStats(IvyExtendedOptions{}); err != ErrInvariant {
		t.Fatalf("malformed unary edge accepted: %v", err)
	}
}

func TestIvyExtendedIndependentModel5000(t *testing.T) {
	r := rand.New(rand.NewSource(0x1a71))
	const variables = 4
	mask := uint64((1 << 16) - 1)
	for trial := 0; trial < 5000; trial++ {
		m, err := NewIvyManager(IvyOptions{MaxObjects: 200, MaxWork: 10000})
		if err != nil {
			t.Fatal(err)
		}
		edges := make([]IvyEdge, variables)
		truths := make([]uint64, variables)
		for i := 0; i < variables; i++ {
			edges[i], err = m.CreatePI()
			if err != nil {
				t.Fatal(err)
			}
			var v uint64
			for a := 0; a < 16; a++ {
				if a&(1<<uint(i)) != 0 {
					v |= 1 << uint(a)
				}
			}
			truths[i] = v
		}
		steps := 1 + r.Intn(10)
		for s := 0; s < steps; s++ {
			i, j := r.Intn(len(edges)), r.Intn(len(edges))
			a, b := edges[i], edges[j]
			ta, tb := truths[i], truths[j]
			if r.Intn(2) == 0 {
				a = a.Not()
				ta = (^ta) & mask
			}
			if r.Intn(2) == 0 {
				b = b.Not()
				tb = (^tb) & mask
			}
			var e IvyEdge
			var tv uint64
			switch r.Intn(3) {
			case 0:
				e, err = m.And(a, b)
				tv = ta & tb
			case 1:
				e, err = m.Or(a, b)
				tv = ta | tb
			default:
				e, err = m.Xor(a, b)
				tv = ta ^ tb
			}
			if err != nil {
				t.Fatal(err)
			}
			edges = append(edges, e)
			truths = append(truths, tv&mask)
		}
		pick := r.Intn(len(edges))
		if _, err = m.CreatePO(edges[pick]); err != nil {
			t.Fatal(err)
		}
		for a := 0; a < 16; a++ {
			in := make([]bool, variables)
			for i := range in {
				in[i] = a&(1<<uint(i)) != 0
			}
			out, err := m.Evaluate(in, nil)
			if err != nil {
				t.Fatal(err)
			}
			want := truths[pick]&(1<<uint(a)) != 0
			if out[0] != want {
				t.Fatalf("trial %d assignment %d got %v want %v", trial, a, out[0], want)
			}
		}
	}
}

func FuzzIvyExtendedContracts(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 7, 15, 31, 63, 127, 255, 1024, 0xdeadbeef} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		r := rand.New(rand.NewSource(int64(seed)))
		m, err := NewIvyManager(IvyOptions{MaxObjects: 256, MaxWork: 100000})
		if err != nil {
			t.Fatal(err)
		}
		pis := make([]IvyEdge, 3)
		for i := range pis {
			pis[i], _ = m.CreatePI()
		}
		edges := slices.Clone(pis)
		for i := 0; i < 1+r.Intn(12); i++ {
			a, b := edges[r.Intn(len(edges))], edges[r.Intn(len(edges))]
			if seed&(1<<uint(i%63)) != 0 {
				a = a.Not()
			}
			var e IvyEdge
			switch r.Intn(4) {
			case 0:
				e, err = m.And(a, b)
			case 1:
				e, err = m.Or(a, b)
			case 2:
				e, err = m.Xor(a, b)
			default:
				e, err = m.Mux(pis[r.Intn(3)], a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			edges = append(edges, e)
		}
		root := edges[r.Intn(len(edges))]
		if _, err = m.CreatePO(root); err != nil {
			t.Fatal(err)
		}
		o := IvyExtendedOptions{MaxObjects: 1000, MaxEdges: 2000, MaxCuts: 10000, MaxLeaves: 3, MaxTruth: 8, MaxBytes: 1 << 20, MaxWork: 100000, MaxDepth: 100}
		truth, err := m.CutTruth(root, []uint32{1, 2, 3}, o)
		if err != nil {
			t.Fatal(err)
		}
		for a := 0; a < 8; a++ {
			in := []bool{a&1 != 0, a&2 != 0, a&4 != 0}
			out, er := m.Evaluate(in, nil)
			if er != nil {
				t.Fatal(er)
			}
			if out[0] != ivyTruthBit(truth, a) {
				t.Fatalf("assignment %d", a)
			}
		}
		balanced, err := m.BalanceCombinational(o)
		if err != nil {
			t.Fatal(err)
		}
		for a := 0; a < 8; a++ {
			in := []bool{a&1 != 0, a&2 != 0, a&4 != 0}
			x, _ := m.Evaluate(in, nil)
			y, _ := balanced.Evaluate(in, nil)
			if !slices.Equal(x, y) {
				t.Fatalf("balance %d", a)
			}
		}
		if _, err = m.EnumerateCuts(3, o); err != nil {
			t.Fatal(err)
		}
		if _, err = m.FastMap(3, o); err != nil {
			t.Fatal(err)
		}
	})
}
