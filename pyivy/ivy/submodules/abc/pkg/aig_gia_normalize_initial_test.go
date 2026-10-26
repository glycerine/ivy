package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaInitialNormalizationTrace(t *testing.T) {
	m := NewGiaManager()
	m.Name = "resets"
	m.Spec = "resets.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	mux, _ := m.AppendMuxReal(c, x, b)
	for _, l := range []GiaLit{mux, c, a.Not(), x, b.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(3)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for encoding := 0; encoding < 27; encoding++ {
		initial := make([]LatchInit, 3)
		value := encoding
		unknown := make([]int, 0)
		for i := range initial {
			initial[i] = LatchInit(value % 3)
			value /= 3
			if initial[i] == LatchInitUnknown {
				unknown = append(unknown, i)
			}
		}
		r, e := m.NormalizeInitialState(initial, 1, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(r.UnknownRegisters, unknown) || r.Graph.constraints != 1 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source", r)
		}
		for unknownBits := 0; unknownBits < 1<<len(unknown); unknownBits++ {
			for sequence := 0; sequence < 16; sequence++ {
				source := make([]bool, 3)
				for i, init := range initial {
					source[i] = init == LatchInitOne
				}
				for i, index := range unknown {
					source[index] = unknownBits&(1<<i) != 0
				}
				target := make([]bool, r.Graph.registers)
				for frame := 0; frame < 4; frame++ {
					p := sequence&(1<<frame) != 0
					input := make([]bool, r.Graph.PrimaryInputCount())
					input[0] = p
					for i := range unknown {
						input[1+i] = unknownBits&(1<<i) != 0
						if frame > 0 {
							input[1+i] = !input[1+i]
						}
					}
					input[len(input)-1] = frame%2 == 0
					want, next, e := m.Step([]bool{p}, source)
					if e != nil {
						t.Fatal(e)
					}
					got, nextTarget, e := r.Graph.Step(input, target)
					if e != nil || !slices.Equal(got, want) {
						t.Fatal(encoding, unknownBits, sequence, frame, got, want, e)
					}
					for i, init := range initial {
						if nextTarget[i] != (next[i] != (init == LatchInitOne)) {
							t.Fatal("state encoding", encoding, frame, i)
						}
					}
					if len(unknown) > 0 && !nextTarget[len(nextTarget)-1] {
						t.Fatal("selector did not become one")
					}
					source, target = next, nextTarget
				}
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.NormalizeInitialState([]LatchInit{0, 1, 2}, 0, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 15} {
		calls := 0
		r, e := m.NormalizeInitialState([]LatchInit{0, 1, 2}, 0, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, initial := range [][]LatchInit{nil, {0, 1, 3}} {
		if _, e := m.NormalizeInitialState(initial, 0, GiaCofactorOptions{}); e == nil {
			t.Fatal(initial)
		}
	}
}
func FuzzGiaInitialNormalization(f *testing.F) {
	f.Add(uint8(5), uint8(42))
	f.Fuzz(func(t *testing.T, encoding, sequence uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(b.Not())
		m.AppendCO(x)
		m.AppendCO(a.Not())
		m.SetRegisterCount(2)
		initial := []LatchInit{LatchInit(encoding % 3), LatchInit((encoding / 3) % 3)}
		r, e := m.NormalizeInitialState(initial, 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		source := []bool{initial[0] == 1, initial[1] == 1}
		for i, index := range r.UnknownRegisters {
			source[index] = encoding&(1<<(i+4)) != 0
		}
		target := make([]bool, r.Graph.registers)
		for frame := 0; frame < 8; frame++ {
			pv := sequence&(1<<frame) != 0
			input := make([]bool, r.Graph.PrimaryInputCount())
			input[0] = pv
			for i, index := range r.UnknownRegisters {
				input[1+i] = source[index]
				if frame > 0 {
					input[1+i] = frame%2 == 0
				}
			}
			got, next, e := r.Graph.Step(input, target)
			if e != nil || got[0] != !source[1] {
				t.Fatal(got, source, e)
			}
			source = []bool{pv != source[0], !source[0]}
			target = next
		}
	})
}

func TestGiaInitialNormalizationNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	for _, l := range []GiaLit{x, c, a.Not(), x, b.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(3)
	raw, e := os.ReadFile("test_vectors/oracle/gia_normalize_initial.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Inputs, Registers int
		Drivers                    []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 3 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		initial := [][]LatchInit{{0, 0, 0}, {0, 1, 2}, {2, 2, 2}}[mode]
		r, e := m.NormalizeInitialState(initial, 1, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || r.Graph.InputCount() != row.Inputs || r.Graph.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row, r.Graph.ObjectCount())
		}
	}
}
func TestGiaInitialNormalizationUnusedUnknown(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	m.AppendCI()
	m.AppendCO(p)
	m.AppendCO(GiaConst0)
	m.SetRegisterCount(1)
	r, e := m.NormalizeInitialState([]LatchInit{LatchInitUnknown}, 0, GiaCofactorOptions{})
	if e != nil || r.Graph.AndCount() != 0 || r.Graph.PrimaryInputCount() != 2 || r.ResetRegister != 1 || r.Graph.registers != 2 || !slices.Equal(r.Initial, []LatchInit{0, 0}) {
		t.Fatal(r, e)
	}
}

func TestGiaInitialNormalizationMuxControlOnlyFanout(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	control, _ := m.AppendCI()
	mux, _ := m.AppendMuxReal(control, a, b)
	m.AppendCO(mux)
	m.AppendCO(GiaConst0)
	m.SetRegisterCount(1)
	r, e := m.NormalizeInitialState([]LatchInit{LatchInitUnknown}, 0, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for bits := 0; bits < 8; bits++ {
		av, bv, initial := bits&1 != 0, bits&2 != 0, bits&4 != 0
		out, next, e := r.Graph.Step([]bool{av, bv, initial}, []bool{false, false})
		want := bv
		if initial {
			want = av
		}
		if e != nil || out[0] != want {
			t.Fatal(bits, out, want, e)
		}
		later, _, e := r.Graph.Step([]bool{av, bv, !initial}, next)
		if e != nil || later[0] != bv {
			t.Fatal("later PI leaked into state", later, e)
		}
	}
}
