package abc

// Go-only exact models, invariants and fuzz tests for src/aig/gia/giaIso3.c.

import (
	"errors"
	"math/rand"
	"slices"
	"sort"
	"testing"
)

func giaIsoFixture(t testing.TB, swapInputs bool) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	if swapInputs {
		a, b = b, a
	}
	plain, _ := m.AppendAnd(a, b)
	oneComplement, _ := m.AppendAnd(a.Not(), c)
	twoComplements, _ := m.AppendAnd(b.Not(), c.Not())
	root, _ := m.AppendAnd(plain, oneComplement.Not())
	buffer, _ := m.AppendBuffer(twoComplements.Not())
	for _, driver := range []GiaLit{root, plain.Not(), buffer} {
		if _, err := m.AppendCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	return m
}

func giaIso3InitialModel(m *GiaManager) []uint32 {
	result := make([]uint32, len(m.objects))
	for i, object := range m.objects {
		switch {
		case object.IsAnd():
			first, _ := object.Fanin0(GiaID(i))
			second, _ := object.Fanin1(GiaID(i))
			phase := 0
			if first.Complemented() {
				phase++
			}
			if second.Complemented() {
				phase++
			}
			result[i] = [3]uint32{0x04892ed6, 0xc2523d7d, 0xdc36cd2e}[phase]
		case object.IsCI():
			result[i] = 0xf2db64f8
		case object.IsCO():
			result[i] = 0xde3126bb
		default:
			result[i] = 0xdebbdff0
		}
	}
	return result
}

func giaIso3RefineModel(m *GiaManager, signatures []uint32) []uint32 {
	result := slices.Clone(signatures)
	for i, object := range m.objects {
		if !object.IsAnd() && !object.IsCO() {
			continue
		}
		id := GiaID(i)
		add := func(fanin GiaLit) {
			complement := uint32(0x8ba63e50)
			if fanin.Complemented() {
				complement = 0x14d87f02
			}
			result[id] += signatures[fanin.ID()] + complement + 0x855ee0cf
			result[fanin.ID()] += signatures[id] + complement + 0x946e1b5f
		}
		first, _ := object.Fanin0(id)
		add(first)
		if object.IsAnd() {
			second, _ := object.Fanin1(id)
			add(second)
		}
	}
	return result
}

func TestGiaIso3InitializationRefinementAndConvergence(t *testing.T) {
	m := giaIsoFixture(t, false)
	before := slices.Clone(m.objects)
	initial, err := m.Iso3InitialSignatures(GiaIsoOptions{})
	wantInitial := giaIso3InitialModel(m)
	if err != nil || !slices.Equal(initial, wantInitial) || !slices.Equal(before, m.objects) {
		t.Fatal(initial, wantInitial, err)
	}
	next, err := m.Iso3Refine(initial, GiaIsoOptions{})
	wantNext := giaIso3RefineModel(m, initial)
	if err != nil || !slices.Equal(next, wantNext) || !slices.Equal(before, m.objects) {
		t.Fatal(next, wantNext, err)
	}
	result, err := m.AnalyzeIso3(GiaIsoOptions{})
	if err != nil || !result.Converged || len(result.UniqueCounts) < 2 || len(result.UniqueCounts) > 500 || len(result.Signatures) != m.ObjectCount() {
		t.Fatal(result, err)
	}
	for i := 1; i < len(result.UniqueCounts); i++ {
		if result.UniqueCounts[i] < result.UniqueCounts[i-1] {
			t.Fatal("signature classes merged", result.UniqueCounts)
		}
	}
	if result.UniqueCounts[len(result.UniqueCounts)-1] != result.UniqueCounts[len(result.UniqueCounts)-2] {
		t.Fatal("wrong stopping condition", result.UniqueCounts)
	}

	other := giaIsoFixture(t, true)
	otherResult, err := other.AnalyzeIso3(GiaIsoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	firstSorted := slices.Clone(result.Signatures)
	secondSorted := slices.Clone(otherResult.Signatures)
	sort.Slice(firstSorted, func(i, j int) bool { return firstSorted[i] < firstSorted[j] })
	sort.Slice(secondSorted, func(i, j int) bool { return secondSorted[i] < secondSorted[j] })
	if !slices.Equal(firstSorted, secondSorted) || !slices.Equal(result.UniqueCounts, otherResult.UniqueCounts) {
		t.Fatal("isomorphic relabeling changed signatures", result.UniqueCounts, otherResult.UniqueCounts)
	}
}

type giaIsoRandomModel struct{ z, w uint32 }

func (r *giaIsoRandomModel) next() uint32 {
	if r.z == 0 && r.w == 0 {
		r.z, r.w = 3716960521, 2174103536
	}
	r.z = 36969*(r.z&65535) + r.z>>16
	r.w = 18000*(r.w&65535) + r.w>>16
	return r.z<<16 + r.w
}

func giaIso4Model(m *GiaManager) GiaIso4Result {
	reverse := make([]uint32, len(m.objects))
	maximum := uint32(0)
	for i := len(m.objects) - 1; i >= 0; i-- {
		object := m.objects[i]
		id := GiaID(i)
		switch {
		case object.IsBuffer():
			first, _ := object.Fanin0(id)
			reverse[first.ID()] = max(reverse[first.ID()], reverse[i])
		case object.IsAnd():
			first, _ := object.Fanin0(id)
			second, _ := object.Fanin1(id)
			reverse[first.ID()] = max(reverse[first.ID()], reverse[i]+1)
			reverse[second.ID()] = max(reverse[second.ID()], reverse[i]+1)
		case object.IsCO():
			first, _ := object.Fanin0(id)
			reverse[first.ID()] = max(reverse[first.ID()], uint32(1))
		default:
			maximum = max(maximum, reverse[i])
		}
	}
	result := GiaIso4Result{Levels: make([][]GiaID, int(maximum)+1), Values: make([]uint32, len(m.objects))}
	for i, object := range m.objects {
		if i == 0 || !object.IsCO() && reverse[i] == 0 {
			continue
		}
		result.Levels[reverse[i]] = append(result.Levels[reverse[i]], GiaID(i))
	}
	var random giaIsoRandomModel
	_ = random.next()
	for level, objects := range result.Levels {
		complements := [2]uint32{random.next(), random.next()}
		for _, id := range objects {
			object := m.objects[id]
			if level == 0 {
				result.Values[id] = random.next()
				first, _ := object.Fanin0(id)
				phase := 0
				if first.Complemented() {
					phase = 1
				}
				result.Values[first.ID()] += result.Values[id] + complements[phase]
				continue
			}
			if !object.IsAnd() {
				continue
			}
			for _, fanin := range func() []GiaLit { a, _ := object.Fanin0(id); b, _ := object.Fanin1(id); return []GiaLit{a, b} }() {
				phase := 0
				if fanin.Complemented() {
					phase = 1
				}
				result.Values[fanin.ID()] += result.Values[id] + complements[phase]
			}
		}
	}
	return result
}

func TestGiaIso4PropagationAndComparison(t *testing.T) {
	m := giaIsoFixture(t, false)
	before := slices.Clone(m.objects)
	actual, err := m.AnalyzeIso4(GiaIsoOptions{})
	want := giaIso4Model(m)
	if err != nil || !slices.Equal(actual.Values, want.Values) || len(actual.Levels) != len(want.Levels) || !slices.Equal(before, m.objects) {
		t.Fatal(actual, want, err)
	}
	for level := range want.Levels {
		if !slices.Equal(actual.Levels[level], want.Levels[level]) {
			t.Fatal(level, actual.Levels[level], want.Levels[level])
		}
	}
	comparison, err := CompareIso4(m, giaIsoFixture(t, true), GiaIsoOptions{})
	if err != nil || len(comparison.Levels) != len(actual.Levels) {
		t.Fatal(comparison, err)
	}
	for level, record := range comparison.Levels {
		if record.Level != level || record.OnlyFirst != 0 || record.OnlySecond != 0 || record.Common != len(actual.Levels[level]) {
			t.Fatal(level, record)
		}
	}
}

func TestGiaIsoErrorsLimitsCancellationAndDeep(t *testing.T) {
	m := giaIsoFixture(t, false)
	for _, options := range []GiaIsoOptions{{MaxObjects: -1}, {MaxEntries: -1}, {MaxIterations: -1}, {MaxObjects: 1}, {MaxEntries: 1}} {
		if result, err := m.AnalyzeIso3(options); err == nil || result.Signatures != nil {
			t.Fatal(options, result, err)
		}
	}
	initial, _ := m.Iso3InitialSignatures(GiaIsoOptions{})
	if result, err := m.Iso3Refine(initial[:len(initial)-1], GiaIsoOptions{}); !errors.Is(err, ErrDimension) || result != nil {
		t.Fatal(result, err)
	}
	if result, err := (*GiaManager)(nil).AnalyzeIso3(GiaIsoOptions{}); !errors.Is(err, ErrDimension) || result.Signatures != nil {
		t.Fatal(result, err)
	}
	for threshold := 1; threshold <= 8; threshold++ {
		calls := 0
		result, err := m.AnalyzeIso3(GiaIsoOptions{Stop: func() bool { calls++; return calls == threshold }})
		if !errors.Is(err, ErrCancelled) || result.Signatures != nil {
			t.Fatal(threshold, calls, result, err)
		}
	}
	for _, native := range []func(*GiaManager, GiaLit, GiaLit) error{
		func(g *GiaManager, a, b GiaLit) error { _, err := g.AppendXorReal(a, b); return err },
		func(g *GiaManager, a, b GiaLit) error { _, err := g.AppendMuxReal(a, b, GiaConst0); return err },
	} {
		g := NewGiaManager()
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		native(g, a, b)
		if result, err := g.AnalyzeIso4(GiaIsoOptions{}); !errors.Is(err, ErrUnsupported) || result.Values != nil {
			t.Fatal(result, err)
		}
	}

	deep := NewGiaManager()
	a, _ := deep.AppendCI()
	b, _ := deep.AppendCI()
	chain, _ := deep.AppendAnd(a, b)
	for i := 0; i < 10000; i++ {
		chain, _ = deep.AppendAnd(chain, a.NotIf(i&1 != 0))
	}
	deep.AppendCO(chain)
	initial, err := deep.Iso3InitialSignatures(GiaIsoOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if refined, refineErr := deep.Iso3Refine(initial, GiaIsoOptions{}); refineErr != nil || len(refined) != deep.ObjectCount() {
		t.Fatal(len(refined), refineErr)
	}
	if propagated, propagateErr := deep.AnalyzeIso4(GiaIsoOptions{}); propagateErr != nil || len(propagated.Values) != deep.ObjectCount() {
		t.Fatal(len(propagated.Values), propagateErr)
	}
}

func FuzzGiaIsoSignatures(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 17, 1001} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		m := NewGiaManager()
		literals := make([]GiaLit, 2+random.Intn(5))
		for i := range literals {
			literals[i], _ = m.AppendCI()
		}
		for gate := 0; gate < 32; gate++ {
			if gate%13 == 12 {
				buffer, err := m.AppendBuffer(literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0))
				if err != nil {
					t.Fatal(err)
				}
				literals = append(literals, buffer)
				continue
			}
			left := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			right := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			for left.ID() == right.ID() {
				right = literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			}
			gate, err := m.AppendAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			literals = append(literals, gate)
		}
		for output := 0; output < 1+random.Intn(5); output++ {
			m.AppendCO(literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0))
		}
		initial, err := m.Iso3InitialSignatures(GiaIsoOptions{})
		if err != nil || !slices.Equal(initial, giaIso3InitialModel(m)) {
			t.Fatal(seed, initial, err)
		}
		refined, err := m.Iso3Refine(initial, GiaIsoOptions{})
		if err != nil || !slices.Equal(refined, giaIso3RefineModel(m, initial)) {
			t.Fatal(seed, refined, err)
		}
		propagated, err := m.AnalyzeIso4(GiaIsoOptions{})
		want := giaIso4Model(m)
		if err != nil || !slices.Equal(propagated.Values, want.Values) || len(propagated.Levels) != len(want.Levels) {
			t.Fatal(seed, propagated, want, err)
		}
		for level := range want.Levels {
			if !slices.Equal(propagated.Levels[level], want.Levels[level]) {
				t.Fatal(seed, level, propagated.Levels[level], want.Levels[level])
			}
		}
	})
}
