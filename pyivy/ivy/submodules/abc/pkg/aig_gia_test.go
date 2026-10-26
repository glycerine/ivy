package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"sync"
	"testing"
	"unsafe"
)

func giaTestCI(t testing.TB, m *GiaManager) GiaLit {
	t.Helper()
	l, err := m.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	return l
}
func giaTestCO(t testing.TB, m *GiaManager, l GiaLit) GiaLit {
	t.Helper()
	co, err := m.AppendCO(l)
	if err != nil {
		t.Fatal(err)
	}
	return co
}
func giaTestAnd(t testing.TB, m *GiaManager, a, b GiaLit) GiaLit {
	t.Helper()
	l, err := m.HashAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	return l
}
func giaTestCheck(t testing.TB, m *GiaManager) {
	t.Helper()
	if err := m.Check(); err != nil {
		t.Fatal(err)
	}
}
func giaTestHash(t testing.TB, m *GiaManager) {
	t.Helper()
	if err := m.HashStart(); err != nil {
		t.Fatal(err)
	}
}

func TestGiaLiteralAndCompactLayout(t *testing.T) {
	// Sizeof checks layout only; no pointer conversion or unsafe memory access.
	if unsafe.Sizeof(GiaObject{}) != 12 {
		t.Fatal("compact object size")
	}
	for _, id := range []GiaID{0, 1, 31, MaxGiaID} {
		for _, c := range []bool{false, true} {
			l, err := MakeGiaLit(id, c)
			if err != nil || !l.Valid() || l.ID() != id || l.Complemented() != c || l.Not().Not() != l || l.NotIf(c) != l.Regular() {
				t.Fatal(l, err)
			}
		}
	}
	if _, err := MakeGiaLit(MaxGiaID+1, false); err == nil || InvalidGiaLit.Valid() {
		t.Fatal("GIA literal bounds")
	}
	m := NewGiaManager()
	a, b := giaTestCI(t, m), giaTestCI(t, m)
	node, err := m.AppendAnd(b.Not(), a)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := m.Object(node.ID())
	if o.Words() != [3]uint32{2, 1 | giaComplement, 0} || !o.IsAnd() || o.IsBuffer() || o.IsXor() {
		t.Fatal(o.Words())
	}
	x, err := m.AppendXorReal(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	o, _ = m.Object(x.ID())
	if !o.IsXor() || o.Diff0() != 2 || o.Diff1() != 3 || o.Complement0() || !o.Complement1() {
		t.Fatal(o.Words())
	}
	buf, err := m.AppendBuffer(a.Not())
	if err != nil {
		t.Fatal(err)
	}
	o, _ = m.Object(buf.ID())
	if !o.IsBuffer() || o.Diff0() != 4 || !o.Complement0() || !o.Complement1() {
		t.Fatal(o.Words())
	}
	co := giaTestCO(t, m, node.Not())
	o, _ = m.Object(co.ID())
	index, err := o.TerminalIndex()
	if err != nil || index != 0 || !o.IsCO() || o.Diff0() != 3 || !o.Complement0() {
		t.Fatal(o.Words(), err)
	}
	ci, _ := m.Object(a.ID())
	if !ci.IsCI() || !ci.IsTerminal() || ci.IsConstant0() {
		t.Fatal(ci.Words())
	}
	if _, err := ci.Fanin0(a.ID()); err == nil {
		t.Fatal("CI fanin")
	}
	if _, err := o.Fanin1(co.ID()); err == nil {
		t.Fatal("CO second fanin")
	}
	if _, err := m.objects[node.ID()].TerminalIndex(); err == nil {
		t.Fatal("AND ordinal")
	}
	giaTestCheck(t, m)
	// The 29-bit all-ones distance is a type sentinel, not a usable edge.
	if _, err := giaDifference(MaxGiaID, GiaConst0); err == nil {
		t.Fatal("sentinel distance")
	}
}

func TestGiaExhaustiveOperations(t *testing.T) {
	m := NewGiaManager()
	giaTestHash(t, m)
	a, b, c := giaTestCI(t, m), giaTestCI(t, m), giaTestCI(t, m)
	literals := []GiaLit{GiaConst0, GiaConst1, a, a.Not(), b, b.Not(), c, c.Not()}
	truth := []uint8{0, 255, 0xaa, 0x55, 0xcc, 0x33, 0xf0, 0x0f}
	var expected []uint8
	for ai, a := range literals {
		for bi, b := range literals {
			v := giaTestAnd(t, m, a, b)
			giaTestCO(t, m, v)
			expected = append(expected, truth[ai]&truth[bi])
			v, err := m.HashXor(a, b)
			if err != nil {
				t.Fatal(err)
			}
			giaTestCO(t, m, v)
			expected = append(expected, truth[ai]^truth[bi])
			real, err := m.HashXorReal(a, b)
			if err != nil {
				t.Fatal(err)
			}
			giaTestCO(t, m, real)
			expected = append(expected, truth[ai]^truth[bi])
			for ci, c := range literals {
				v, err := m.HashMux(a, b, c)
				if err != nil {
					t.Fatal(err)
				}
				giaTestCO(t, m, v)
				expected = append(expected, (truth[ai]&truth[bi])|(^truth[ai]&truth[ci]))
				v, err = m.HashMuxReal(a, b, c)
				if err != nil {
					t.Fatal(err)
				}
				giaTestCO(t, m, v)
				expected = append(expected, (truth[ai]&truth[bi])|(^truth[ai]&truth[ci]))
				v, err = m.HashMajority(a, b, c)
				if err != nil {
					t.Fatal(err)
				}
				giaTestCO(t, m, v)
				expected = append(expected, (truth[ai]&truth[bi])|(truth[ai]&truth[ci])|(truth[bi]&truth[ci]))
			}
		}
	}
	giaTestCheck(t, m)
	for assignment := 0; assignment < 8; assignment++ {
		out, err := m.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		for i, table := range expected {
			if out[i] != (table&(1<<assignment) != 0) {
				t.Fatalf("operation %d assignment %d", i, assignment)
			}
		}
	}
	if m.XorCount() == 0 || m.MuxCount() == 0 {
		t.Fatal("native XOR/mux coverage")
	}
	if err := m.SetPhases(); err != nil {
		t.Fatal(err)
	}
	zeros, _ := m.Evaluate([]bool{false, false, false})
	for i, id := range m.cos {
		if m.objects[id].Phase() != zeros[i] {
			t.Fatal("phase")
		}
	}
	giaTestCheck(t, m)
}

func TestGiaHashLifecycleAndRawDuplicates(t *testing.T) {
	m := NewGiaManager()
	a, b := giaTestCI(t, m), giaTestCI(t, m)
	if _, err := m.HashAnd(a, b); err == nil {
		t.Fatal("missing hash table")
	}
	original, err := m.AppendAnd(b, a)
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := m.AppendAnd(a, b)
	if err != nil || duplicate == original {
		t.Fatal("raw duplicate", err)
	}
	giaTestHash(t, m)
	if got := giaTestAnd(t, m, a, b); got != original {
		t.Fatal("first occurrence hash", got, original)
	}
	if err := m.HashStart(); err == nil {
		t.Fatal("already hashing")
	}
	before := m.ObjectCount()
	if _, yes, err := m.HashAndTry(a.Not(), b); err != nil || yes {
		t.Fatal(yes, err)
	}
	if m.ObjectCount() != before {
		t.Fatal("lookup allocated")
	}
	if got, yes, err := m.HashAndTry(a, a.Not()); err != nil || !yes || got != GiaConst0 {
		t.Fatal(got, yes, err)
	}
	if got, yes, err := m.HashAndTry(a, b); err != nil || !yes || got != original {
		t.Fatal(got, yes, err)
	}
	clone := m.Clone()
	m.HashStop()
	if !clone.Hashing() || m.Hashing() {
		t.Fatal("clone hash ownership")
	}
	if _, err := m.AppendAnd(a.Not(), b); err != nil {
		t.Fatal(err)
	}
	if clone.ObjectCount() != before {
		t.Fatal("clone arena aliases")
	}
	giaTestHash(t, m)
	giaTestCheck(t, m)
	giaTestCheck(t, clone)
	ids := m.Inputs()
	ids[0] = InvalidGiaID
	if m.Inputs()[0] != a.ID() {
		t.Fatal("CI slice alias")
	}
}

func TestGiaCleanupMappingAndBuffers(t *testing.T) {
	m := NewGiaManager()
	m.Name = "cleanup"
	m.Spec = "source"
	a, b, c := giaTestCI(t, m), giaTestCI(t, m), giaTestCI(t, m)
	giaTestHash(t, m)
	live := giaTestAnd(t, m, a, b)
	dead := giaTestAnd(t, m, a.Not(), b)
	rooted := giaTestAnd(t, m, b, c)
	buf, err := m.AppendBuffer(rooted)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := m.HashMuxReal(a, b, c)
	if err != nil {
		t.Fatal(err)
	}
	xor, err := m.HashXorReal(a, c)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, m, live)
	giaTestCO(t, m, mux)
	giaTestCO(t, m, xor)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := m.Clone()
	cleaned, mapping, err := m.Cleanup()
	if err != nil {
		t.Fatal(err)
	}
	if mapping[dead.ID()] != InvalidGiaLit || mapping[buf.ID()] == InvalidGiaLit || mapping[rooted.ID()] == InvalidGiaLit {
		t.Fatal("cleanup buffer roots")
	}
	if cleaned.ObjectCount() != m.ObjectCount()-1 || cleaned.BufferCount() != 1 || cleaned.MuxCount() != 1 || cleaned.XorCount() != 1 || cleaned.RegisterCount() != 1 || cleaned.ConstraintCount() != 1 || cleaned.Name != m.Name || cleaned.Spec != m.Spec {
		t.Fatal("cleanup metadata")
	}
	for i := range m.objects {
		if m.objects[i] != before.objects[i] {
			t.Fatal("cleanup changed source scratch")
		}
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		a, _ := m.Evaluate(inputs)
		b, _ := cleaned.Evaluate(inputs)
		if !slices.Equal(a, b) {
			t.Fatal("cleanup changed outputs")
		}
		oldValues, _ := m.evaluateValues(inputs)
		newValues, _ := cleaned.evaluateValues(inputs)
		for id, l := range mapping {
			if l != InvalidGiaLit && oldValues[id] != (newValues[l.ID()] != l.Complemented()) {
				t.Fatal("invalid cleanup mapping", id, l)
			}
		}
	}
	again, _, err := cleaned.Cleanup()
	if err != nil || again.ObjectCount() != cleaned.ObjectCount() {
		t.Fatal("cleanup idempotence", err)
	}
	giaTestCheck(t, m)
	giaTestCheck(t, cleaned)
}

func TestGiaAigConversionAndSequentialOrdering(t *testing.T) {
	g := NewGiaManager()
	g.Name = "seq"
	g.Spec = "source"
	pi, ro := giaTestCI(t, g), giaTestCI(t, g)
	giaTestHash(t, g)
	xor, err := g.HashXorReal(pi, ro)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(pi.Not())
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, g, xor)
	giaTestCO(t, g, GiaConst1)
	ri := giaTestCO(t, g, buffer)
	g.SetRegisterCount(1)
	g.SetConstraintCount(1)
	if id, err := g.RegisterInputForOutput(ro.ID()); err != nil || id != ri.ID() {
		t.Fatal(id, err)
	}
	if id, err := g.RegisterOutputForInput(ri.ID()); err != nil || id != ro.ID() {
		t.Fatal(id, err)
	}
	if _, err := g.RegisterInputForOutput(pi.ID()); err == nil {
		t.Fatal("PI treated as register")
	}
	aig, toAig, err := g.ToAig()
	if err != nil {
		t.Fatal(err)
	}
	if toAig[0] != AigConst0 || aig.RegisterCount() != 1 || aig.ConstraintCount() != 1 {
		t.Fatal("constant/partition conversion")
	}
	round, toGia, err := GiaFromAig(aig)
	if err != nil {
		t.Fatal(err)
	}
	if toGia[0] != GiaConst1 || round.RegisterCount() != 1 || round.ConstraintCount() != 1 || round.Name != g.Name || round.Spec != g.Spec {
		t.Fatal("round trip metadata")
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		before, _ := g.Evaluate(inputs)
		middle, _ := aig.Evaluate(inputs)
		after, _ := round.Evaluate(inputs)
		if !slices.Equal(before, middle) || !slices.Equal(before, after) {
			t.Fatal(before, middle, after)
		}
		primary, state := inputs[:1], inputs[1:]
		out, next, err := round.Step(primary, state)
		if err != nil || !slices.Equal(out, before[:2]) || !slices.Equal(next, before[2:]) {
			t.Fatal(out, next, err)
		}
	}
	// Tombstone mapping and structural merge are explicit rather than assuming
	// equal object IDs between the different graph managers.
	dangling := aigTestAnd(t, aig, AigLit(aig.cis[0])<<1, (AigLit(aig.cis[1]) << 1).Not())
	aig.Cleanup()
	_, mapDead, err := GiaFromAig(aig)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aig.Object(dangling.ID()); err != nil && mapDead[dangling.ID()] != InvalidGiaLit {
		t.Fatal("mapped deleted object")
	}
	duplicate := NewGiaManager()
	a, b := giaTestCI(t, duplicate), giaTestCI(t, duplicate)
	x, _ := duplicate.AppendAnd(a, b)
	y, _ := duplicate.AppendAnd(a, b)
	giaTestCO(t, duplicate, x)
	giaTestCO(t, duplicate, y)
	_, mapped, err := duplicate.ToAig()
	if err != nil || mapped[x.ID()] != mapped[y.ID()] {
		t.Fatal("duplicate conversion", err)
	}
}

func TestGiaErrorsAndOutputPatching(t *testing.T) {
	m := NewGiaManager()
	a, b := giaTestCI(t, m), giaTestCI(t, m)
	co := giaTestCO(t, m, a)
	giaTestHash(t, m)
	for _, l := range []GiaLit{InvalidGiaLit, co, GiaLit(m.ObjectCount()+50) << 1} {
		count := m.ObjectCount()
		if _, err := m.HashAnd(l, a); err == nil {
			t.Fatal(l)
		}
		if _, err := m.HashXorReal(l, a); err == nil {
			t.Fatal(l)
		}
		if _, err := m.HashMuxReal(a, l, b); err == nil {
			t.Fatal(l)
		}
		if _, err := m.AppendCO(l); err == nil {
			t.Fatal(l)
		}
		if m.ObjectCount() != count {
			t.Fatal("failed validation mutated graph")
		}
	}
	if _, err := m.AppendAnd(a, a); err == nil {
		t.Fatal("raw same-variable fanins")
	}
	if _, err := m.AppendMuxReal(a, a, b); err == nil {
		t.Fatal("raw mux repeated ID")
	}
	if _, err := m.AppendCI(); err != nil {
		t.Fatal(err)
	}
	later := GiaLit(m.cis[2]) << 1
	if err := m.PatchCODriver(0, later); err == nil {
		t.Fatal("late CO driver")
	}
	m.objects[co.ID()].word0 |= 1 << 30
	if err := m.PatchCODriver(0, b.Not()); err != nil {
		t.Fatal(err)
	}
	if !m.objects[co.ID()].Mark0() {
		t.Fatal("patch destroyed user mark")
	}
	m.FlipCODriver(0)
	out, _ := m.Evaluate([]bool{false, true, false})
	if !out[0] {
		t.Fatal("flipped driver")
	}
	if m.SetRegisterCount(-1) == nil || m.SetConstraintCount(2) == nil {
		t.Fatal("partition bounds")
	}
	if _, err := new(GiaManager).AppendCI(); err == nil {
		t.Fatal("zero manager")
	}
	if new(GiaManager).Check() == nil {
		t.Fatal("zero manager check")
	}
	if _, err := m.Evaluate(nil); err == nil {
		t.Fatal("CI dimensions")
	}
	if _, _, err := m.Step(nil, nil); err == nil {
		t.Fatal("sequential dimensions")
	}
	giaTestCheck(t, m)
	for _, corrupt := range []func(*GiaManager){
		func(g *GiaManager) { g.objects[0].word0 = 0 },
		func(g *GiaManager) { g.cis[0] = InvalidGiaID },
		func(g *GiaManager) { g.cos[0] = InvalidGiaID },
		func(g *GiaManager) { g.objects[co.ID()].word0 = giaTerm },
		func(g *GiaManager) { g.muxCount++ },
	} {
		g := m.Clone()
		corrupt(g)
		if g.Check() == nil {
			t.Fatal("missed corruption")
		}
	}
}

func TestAigGiaConstraintMetadata(t *testing.T) {
	m, err := ReadAiger([]byte("aig 1 1 0 0 0 1 1\n2\n3\n"), AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if m.Graph().ConstraintCount() != 1 {
		t.Fatal("AIGER constraint propagation")
	}
	g, _, err := GiaFromAig(m.Graph())
	if err != nil || g.ConstraintCount() != 1 {
		t.Fatal(g, err)
	}
	aig, _, err := g.ToAig()
	if err != nil || aig.ConstraintCount() != 1 {
		t.Fatal(err)
	}
	rebuilt, err := NewAigerModel(aig, 0, 1, 1, nil)
	if err != nil || rebuilt.Validate() != nil {
		t.Fatal(err)
	}
	clone := aig.Clone()
	if clone.ConstraintCount() != 1 {
		t.Fatal("AIG clone constraints")
	}
	if clone.SetConstraintCount(-1) == nil || clone.SetConstraintCount(3) == nil {
		t.Fatal("AIG constraint bounds")
	}
	aig.SetConstraintCount(0)
	if rebuilt.Validate() == nil {
		t.Fatal("model/graph metadata mismatch")
	}
}

func TestGiaIndependentManagers(t *testing.T) {
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Go(func() {
			m := NewGiaManager()
			giaTestHash(t, m)
			a, b := giaTestCI(t, m), giaTestCI(t, m)
			giaTestCO(t, m, giaTestAnd(t, m, a, b))
			for j := 0; j < 20; j++ {
				g, _, err := m.Cleanup()
				if err != nil {
					t.Error(err)
					return
				}
				giaTestCheck(t, g)
				g.Evaluate([]bool{true, false})
			}
		})
	}
	group.Wait()
}

func TestGiaNativeFixture(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Operations []struct {
			Op, A, B, C int
			AC, BC, CC  bool
			Literal     GiaLit
		}
		Words           [][3]uint32
		Controls        []GiaLit
		Truths          []uint8
		RawDangling     [][2]GiaLit `json:"raw_dangling"`
		CleanupMap      []GiaLit    `json:"cleanup_map"`
		CleanupWords    [][3]uint32 `json:"cleanup_words"`
		CleanupControls []GiaLit    `json:"cleanup_controls"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Operations) != 128 || len(fixture.Truths) != 128 {
		t.Fatal("incomplete GIA fixture")
	}
	m := NewGiaManager()
	giaTestHash(t, m)
	roots := []GiaLit{GiaConst0, GiaConst1}
	for i := 0; i < 3; i++ {
		roots = append(roots, giaTestCI(t, m))
	}
	for i, op := range fixture.Operations {
		a, b, c := roots[op.A].NotIf(op.AC), roots[op.B].NotIf(op.BC), roots[op.C].NotIf(op.CC)
		var l GiaLit
		switch op.Op {
		case 0:
			l, err = m.HashAnd(a, b)
		case 1:
			l, err = m.HashOr(a, b)
		case 2:
			l, err = m.HashXor(a, b)
		case 3:
			l, err = m.HashMux(a, b, c)
		case 4:
			l, err = m.HashMajority(a, b, c)
		case 5:
			l, err = m.HashXorReal(a, b)
		case 6:
			l, err = m.HashMuxReal(a, b, c)
		case 7:
			l, err = m.AppendBuffer(a)
		default:
			t.Fatal("unknown native operation")
		}
		if err != nil {
			t.Fatal(err)
		}
		if l != op.Literal {
			t.Fatalf("native literal mismatch at operation %d: Go %d C %d", i, l, op.Literal)
		}
		roots = append(roots, l)
		giaTestCO(t, m, l)
	}
	for _, pair := range fixture.RawDangling {
		if _, err := m.AppendAnd(pair[0], pair[1]); err != nil {
			t.Fatal(err)
		}
	}
	giaTestCheck(t, m)
	if len(fixture.Words) != m.ObjectCount() || len(fixture.Controls) != m.ObjectCount() {
		t.Fatal("native object count")
	}
	for i, obj := range m.objects {
		if obj.Words() != fixture.Words[i] {
			t.Fatal("native compact words", i, obj.Words(), fixture.Words[i])
		}
		control, yes := m.MuxControl(GiaID(i))
		if !yes {
			control = 0
		}
		if control != fixture.Controls[i] {
			t.Fatal("native mux control", i)
		}
	}
	cleaned, mapping, err := m.Cleanup()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(mapping, fixture.CleanupMap) || len(cleaned.objects) != len(fixture.CleanupWords) {
		t.Fatal("native cleanup mapping")
	}
	if cleaned.ObjectCount() != m.ObjectCount()-2 {
		t.Fatal("native dead cone not removed")
	}
	for i, obj := range cleaned.objects {
		if obj.Words() != fixture.CleanupWords[i] {
			t.Fatal("native cleaned words", i)
		}
		control, yes := cleaned.MuxControl(GiaID(i))
		if !yes {
			control = 0
		}
		if control != fixture.CleanupControls[i] {
			t.Fatal("native cleaned mux control", i)
		}
	}
	for assignment := 0; assignment < 8; assignment++ {
		outputs, err := m.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		for i, table := range fixture.Truths {
			if outputs[i] != (table&(1<<assignment) != 0) {
				t.Fatal("native GIA truth", i, assignment)
			}
		}
	}
}

// A bounded expression model supplies independent truth masks for GIA,
// conversion, cleanup and every retained mapping, including native XOR/mux.
func FuzzGiaOperations(f *testing.F) {
	for _, seed := range [][]byte{{}, {5, 2, 3, 0, 6, 4, 5, 2}, {7, 2, 0, 0, 7, 5, 0, 0}, {3, 1, 0, 2, 2, 5, 4, 0}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			return
		}
		m := NewGiaManager()
		giaTestHash(t, m)
		roots := []GiaLit{GiaConst0, GiaConst1}
		tables := []uint8{0, 255}
		for _, table := range []uint8{0xaa, 0xcc, 0xf0} {
			roots = append(roots, giaTestCI(t, m))
			tables = append(tables, table)
		}
		for i := 0; i+3 < len(data); i += 4 {
			ai, bi, ci := int(data[i+1])%len(roots), int(data[i+2])%len(roots), int(data[i+3])%len(roots)
			a, b, c := roots[ai], roots[bi], roots[ci]
			av, bv, cv := tables[ai], tables[bi], tables[ci]
			if data[i]&8 != 0 {
				a = a.Not()
				av = ^av
			}
			if data[i]&16 != 0 {
				b = b.Not()
				bv = ^bv
			}
			if data[i]&32 != 0 {
				c = c.Not()
				cv = ^cv
			}
			var l GiaLit
			var table uint8
			var err error
			switch data[i] % 8 {
			case 0:
				l, err = m.HashAnd(a, b)
				table = av & bv
			case 1:
				l, err = m.HashOr(a, b)
				table = av | bv
			case 2:
				l, err = m.HashXor(a, b)
				table = av ^ bv
			case 3:
				l, err = m.HashMux(a, b, c)
				table = (av & bv) | (^av & cv)
			case 4:
				l, err = m.HashMajority(a, b, c)
				table = (av & bv) | (av & cv) | (bv & cv)
			case 5:
				l, err = m.HashXorReal(a, b)
				table = av ^ bv
			case 6:
				l, err = m.HashMuxReal(a, b, c)
				table = (av & bv) | (^av & cv)
			case 7:
				l, err = m.AppendBuffer(a)
				table = av
			}
			if err != nil {
				t.Fatal(err)
			}
			roots = append(roots, l)
			tables = append(tables, table)
			giaTestCheck(t, m)
		}
		first := max(0, len(roots)-6)
		for _, l := range roots[first:] {
			giaTestCO(t, m, l)
		}
		source := m.Clone()
		cleaned, mapping, err := m.Cleanup()
		if err != nil {
			t.Fatal(err)
		}
		aig, toAig, err := m.ToAig()
		if err != nil {
			t.Fatal(err)
		}
		round, _, err := GiaFromAig(aig)
		if err != nil {
			t.Fatal(err)
		}
		for i := range source.objects {
			if source.objects[i] != m.objects[i] {
				t.Fatal("transformation mutated source")
			}
		}
		for assignment := 0; assignment < 8; assignment++ {
			input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			outputs, err := m.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			after, _ := cleaned.Evaluate(input)
			middle, _ := aig.Evaluate(input)
			last, _ := round.Evaluate(input)
			if !slices.Equal(outputs, after) || !slices.Equal(outputs, middle) || !slices.Equal(outputs, last) {
				t.Fatal("GIA transformation semantics")
			}
			for i, table := range tables[first:] {
				if outputs[i] != (table&(1<<assignment) != 0) {
					t.Fatal("independent GIA truth")
				}
			}
			oldValues, _ := m.evaluateValues(input)
			newValues, _ := cleaned.evaluateValues(input)
			for id, l := range mapping {
				if l != InvalidGiaLit && oldValues[id] != (newValues[l.ID()] != l.Complemented()) {
					t.Fatal("GIA mapping", id)
				}
			}
			// Evaluate mapped AIG regular signals by adding COs to an independent clone.
			copyAig := aig.Clone()
			expected := make([]bool, 0)
			for id, l := range toAig {
				if !m.objects[id].IsCO() {
					if _, err := copyAig.CreateCO(l); err != nil {
						t.Fatal(err)
					}
					expected = append(expected, oldValues[id])
				}
			}
			mappedOutputs, _ := copyAig.Evaluate(input)
			if !slices.Equal(mappedOutputs[len(outputs):], expected) {
				t.Fatal("GIA/AIG literal mapping")
			}
		}
	})
}

func TestGiaInvalidSentinels(t *testing.T) {
	if GiaLit(math.MaxUint32).Valid() {
		t.Fatal("invalid sentinel")
	}
	var o GiaObject
	if _, err := o.Fanin0(0); err == nil {
		t.Fatal("zero distance")
	}
}
