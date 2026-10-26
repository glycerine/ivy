package abc

import (
	"errors"
	"slices"
	"testing"
)

func buildGiaMuxEnable(t testing.TB, flags byte) (*GiaManager, GiaLit, GiaLit) {
	t.Helper()
	g := NewGiaManager()
	control := giaTestCI(t, g)
	data := giaTestCI(t, g)
	observe := giaTestCI(t, g)
	state := giaTestCI(t, g)
	giaTestHash(t, g)
	control = control.NotIf(flags&1 != 0)
	data = data.NotIf(flags&2 != 0)
	feedback := state.NotIf(flags&4 != 0)
	var next GiaLit
	var err error
	if flags&8 != 0 {
		next, err = g.HashMuxReal(control, feedback, data)
	} else {
		next, err = g.HashMuxReal(control, data, feedback)
	}
	if err != nil {
		t.Fatal(err)
	}
	next = next.NotIf(flags&16 != 0)
	po, err := g.HashXorReal(observe, state)
	if err != nil {
		t.Fatal(err)
	}
	g.HashStop()
	giaTestCO(t, g, po)
	giaTestCO(t, g, next)
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	g.Name, g.Spec = "enable", "enable.aig"
	giaTestCheck(t, g)
	return g, control, data.NotIf(flags&16 != 0)
}

func TestGiaRemoveMuxEnablesExhaustive(t *testing.T) {
	for flags := byte(0); flags < 32; flags++ {
		source, control, data := buildGiaMuxEnable(t, flags)
		before := slices.Clone(source.objects)
		result, err := source.RemoveMuxEnables(GiaCofactorOptions{})
		if err != nil {
			t.Fatal(flags, err)
		}
		if result.Graph == nil || !result.Recognized[0] || result.Data[0] != data || result.Graph.Name != source.Name || result.Graph.Spec != source.Spec || result.Graph.ConstraintCount() != 1 {
			t.Fatal(flags, result)
		}
		if !slices.Equal(before, source.objects) {
			t.Fatal("source mutation", flags)
		}
		for assignment := 0; assignment < 16; assignment++ {
			primary := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			state := []bool{assignment&8 != 0}
			oldOut, _, err := source.Step(primary, state)
			if err != nil {
				t.Fatal(err)
			}
			newOut, next, err := result.Graph.Step(primary, state)
			if err != nil || !slices.Equal(oldOut, newOut) {
				t.Fatal(flags, assignment, oldOut, newOut, err)
			}
			read := func(l GiaLit) bool {
				values := []bool{false, primary[0], primary[1], primary[2], state[0]}
				return values[l.ID()] != l.Complemented()
			}
			if len(next) != 1 || next[0] != read(data) {
				t.Fatal("data branch", flags, assignment, next, data)
			}
			update := read(control)
			if flags&8 != 0 {
				update = !update
			}
			if read(result.Controls[0]) != update {
				t.Fatal("normalized control", flags, assignment, result.Controls[0], update)
			}
		}
	}
}

func TestGiaRemoveMuxEnablesNonmatchAndBounds(t *testing.T) {
	g := NewGiaManager()
	input := giaTestCI(t, g)
	state := giaTestCI(t, g)
	giaTestHash(t, g)
	next := giaTestAnd(t, g, input, state)
	g.HashStop()
	giaTestCO(t, g, next)
	if err := g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	result, err := g.RemoveMuxEnables(GiaCofactorOptions{})
	if err != nil || result.Recognized[0] || result.Controls[0] != InvalidGiaLit || result.Data[0] != InvalidGiaLit {
		t.Fatal(result, err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		_, want, _ := g.Step([]bool{assignment&1 != 0}, []bool{assignment&2 != 0})
		_, got, stepErr := result.Graph.Step([]bool{assignment&1 != 0}, []bool{assignment&2 != 0})
		if stepErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, stepErr)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxObjects: 1}, {MaxMappingEntries: 1}} {
		if r, e := g.RemoveMuxEnables(options); e == nil || r.Graph != nil {
			t.Fatal(options, r, e)
		}
	}
	if r, e := (*GiaManager)(nil).RemoveMuxEnables(GiaCofactorOptions{}); !errors.Is(e, ErrDimension) || r.Graph != nil {
		t.Fatal(r, e)
	}
	calls := 0
	if r, e := g.RemoveMuxEnables(GiaCofactorOptions{Stop: func() bool { calls++; return calls == 2 }}); !errors.Is(e, ErrCancelled) || r.Graph != nil {
		t.Fatal(calls, r, e)
	}
}

func TestGiaRemoveStructuralEnablesClasses(t *testing.T) {
	g := NewGiaManager()
	control := giaTestCI(t, g)
	data0 := giaTestCI(t, g)
	data1 := giaTestCI(t, g)
	state0 := giaTestCI(t, g)
	state1 := giaTestCI(t, g)
	giaTestHash(t, g)
	next0, err := g.HashMux(control, data0, state0)
	if err != nil {
		t.Fatal(err)
	}
	next1, err := g.HashMux(control, data1, state1)
	if err != nil {
		t.Fatal(err)
	}
	g.HashStop()
	giaTestCO(t, g, next0)
	giaTestCO(t, g, next1)
	if err = g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	result, err := g.RemoveStructuralEnables(GiaCofactorOptions{})
	if err != nil || !slices.Equal(result.Recognized, []bool{true, true}) || !slices.Equal(result.Classes, []int{1, 1}) || result.Controls[0] != result.Controls[1] {
		t.Fatal(result, err)
	}
	for assignment := 0; assignment < 32; assignment++ {
		primary := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		state := []bool{assignment&8 != 0, assignment&16 != 0}
		_, next, stepErr := result.Graph.Step(primary, state)
		if stepErr != nil || !slices.Equal(next, primary[1:]) {
			t.Fatal(assignment, next, stepErr)
		}
	}
	newer, err := g.RemoveMuxEnables(GiaCofactorOptions{})
	if err != nil || !slices.Equal(newer.Recognized, result.Recognized) {
		t.Fatal(newer, err)
	}
}

func TestGiaSupergateAndSignalFrequencyReport(t *testing.T) {
	g := NewGiaManager()
	a := giaTestCI(t, g)
	b := giaTestCI(t, g)
	c := giaTestCI(t, g)
	giaTestHash(t, g)
	left := giaTestAnd(t, g, a, b)
	root := giaTestAnd(t, g, left, c.Not())
	g.HashStop()
	immediate, err := g.CollectSupergate(root, false, 0)
	if err != nil || !slices.Equal(immediate, []GiaID{c.ID(), left.ID()}) {
		t.Fatal(immediate, err)
	}
	flat, err := g.CollectSupergate(root, true, 0)
	if err != nil || !slices.Equal(flat, []GiaID{c.ID(), a.ID(), b.ID()}) {
		t.Fatal(flat, err)
	}
	flat, err = g.CollectSupergate(left.Not(), true, 0)
	if err != nil || !slices.Equal(flat, []GiaID{left.ID()}) {
		t.Fatal(flat, err)
	}
	if _, err = g.CollectSupergate(root.Not(), false, 0); err == nil {
		t.Fatal("complemented immediate root")
	}
	if _, err = g.CollectSupergate(root, true, g.ObjectCount()-1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	giaTestCO(t, g, root)
	frequencies := make([]uint64, g.ObjectCount())
	frequencies[a.ID()] = 2
	frequencies[b.ID()] = 11
	frequencies[left.ID()] = 12
	report, err := g.SignalFrequencyReport(frequencies, "Signals", 0)
	if err != nil || report.RepeatedSignals != 3 || report.TotalUse != 25 || len(report.Frequent) != 2 || report.Frequent[0].ID != b.ID() || report.Frequent[1].ID != left.ID() {
		t.Fatal(report, err)
	}
	want := "Signals (total = 3  driven = 25)\n" +
		"  1 :   Obj =      2   Refs =      1   Freq =     11\n" +
		"  2 :   Obj =      4   Refs =      1   Freq =     12\n"
	if got := report.Format(); got != want {
		t.Fatalf("%q != %q", got, want)
	}
	report.Frequent[0].Frequency = 99
	again, err := g.SignalFrequencyReport(frequencies, "Signals", 0)
	if err != nil || again.Frequent[0].Frequency != 11 {
		t.Fatal(again, err)
	}
	if _, err = g.SignalFrequencyReport(frequencies[:len(frequencies)-1], "bad", 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzGiaRemoveMuxEnables(f *testing.F) {
	for _, flags := range []byte{0, 1, 2, 4, 8, 16, 31} {
		f.Add(flags)
	}
	f.Fuzz(func(t *testing.T, flags byte) {
		flags &= 31
		source, _, data := buildGiaMuxEnable(t, flags)
		result, err := source.RemoveMuxEnables(GiaCofactorOptions{})
		if err != nil || !result.Recognized[0] {
			t.Fatal(flags, result, err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			primary := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			state := []bool{assignment&8 != 0}
			oldOut, _, oldErr := source.Step(primary, state)
			newOut, next, newErr := result.Graph.Step(primary, state)
			values := []bool{false, primary[0], primary[1], primary[2], state[0]}
			want := values[data.ID()] != data.Complemented()
			if oldErr != nil || newErr != nil || !slices.Equal(oldOut, newOut) || len(next) != 1 || next[0] != want {
				t.Fatal(flags, assignment, oldOut, newOut, next, want, oldErr, newErr)
			}
		}
	})
}
