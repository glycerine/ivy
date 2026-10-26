package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func halfSeqSource(regs int, extended bool) *GiaManager {
	m := NewGiaManager()
	ins := make([]GiaLit, regs+2)
	for i := range ins {
		ins[i], _ = m.AppendCI()
	}
	var x, y GiaLit
	if extended {
		x, _ = m.AppendXorReal(ins[0], ins[1])
		y, _ = m.AppendMuxReal(ins[0], x, ins[regs+1])
	} else {
		x, _ = m.AppendAnd(ins[0], ins[1])
		y, _ = m.AppendAnd(x.Not(), ins[regs+1])
	}
	m.AppendCO(y)
	m.AppendCO(x.Not())
	for i := 0; i < regs; i++ {
		m.AppendCO(ins[(i+1)%(regs+2)].NotIf(i&1 != 0))
	}
	m.SetRegisterCount(regs)
	m.SetConstraintCount(1)
	m.Name = "halves"
	m.Spec = "halves.aig"
	return m
}

func TestGiaHalfSeqNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_half_seq.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Registers, Second, Objects, TargetRegisters int
		CoIds                                       []GiaID
		Drivers                                     []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 8 {
		t.Fatal(len(rows))
	}
	for _, r := range rows {
		m := halfSeqSource(r.Registers, false)
		part, e := m.SplitRegisterHalf(r.Second != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := part.Graph
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != r.Objects || g.registers != r.TargetRegisters || !slices.Equal(g.cos, r.CoIds) || !slices.Equal(drivers, r.Drivers) {
			t.Fatal(r, g.ObjectCount(), g.cos, drivers)
		}
	}
}

func checkHalfSeq(t *testing.T, regs int, second, extended bool) {
	t.Helper()
	m := halfSeqSource(regs, extended)
	before := slices.Clone(m.objects)
	part, e := m.SplitRegisterHalf(second, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	g := part.Graph
	begin, end := 0, regs/2
	if second {
		begin, end = regs/2, regs
	}
	expectedInputs := []int{0, 1}
	for i := 0; i < regs; i++ {
		if i < begin || i >= end {
			expectedInputs = append(expectedInputs, 2+i)
		}
	}
	var expectedRegisters []int
	for i := begin; i < end; i++ {
		expectedInputs = append(expectedInputs, 2+i)
		expectedRegisters = append(expectedRegisters, i)
	}
	if !slices.Equal(part.Inputs, expectedInputs) || !slices.Equal(part.Registers, expectedRegisters) || g.registers != end-begin || g.Name != m.Name || g.Spec != m.Spec || g.constraints != 0 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/projections", part)
	}
	for bits := 0; bits < 1<<m.InputCount(); bits++ {
		source := make([]bool, m.InputCount())
		for i := range source {
			source[i] = bits&(1<<i) != 0
		}
		target := make([]bool, len(part.Inputs))
		for i, j := range part.Inputs {
			target[i] = source[j]
		}
		outputs, e := m.Evaluate(source)
		if e != nil {
			t.Fatal(e)
		}
		want := append([]bool(nil), outputs[:2]...)
		want = append(want, outputs[2+begin:2+end]...)
		got, e := g.Evaluate(target)
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(regs, second, bits, got, want, e)
		}
	}
}
func TestGiaHalfSeqSemantics(t *testing.T) {
	for regs := 0; regs < 6; regs++ {
		for _, second := range []bool{false, true} {
			checkHalfSeq(t, regs, second, true)
		}
	}
	m := halfSeqSource(3, true)
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		r, e := m.SplitRegisterHalf(false, o)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.SplitRegisterHalf(false, o)
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 11} {
		calls := 0
		r, e := m.SplitRegisterHalf(false, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if r, e := (*GiaManager)(nil).SplitRegisterHalf(false, GiaCofactorOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzGiaHalfSeq(f *testing.F) {
	f.Add(uint8(3))
	f.Fuzz(func(t *testing.T, flags uint8) { checkHalfSeq(t, int(flags%6), flags&8 != 0, flags&16 != 0) })
}
