package abc

// Go-only semantic tests for bmcFault.c structural fault transforms.

import (
	"errors"
	"reflect"
	"testing"
)

func faultTransformSource(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	first, err := g.AppendAnd(a, b.Not())
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.AppendAnd(first.Not(), a)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = g.AppendCO(first)
	_, _ = g.AppendCO(second)
	return g
}

func evaluateFaultModel(t *testing.T, source *GiaManager, primary []bool, controls [][]bool, mode int, enabled []bool) []bool {
	t.Helper()
	values := make([]bool, source.ObjectCount())
	for i, id := range source.cis {
		values[id] = primary[i]
	}
	read := func(l GiaLit) bool {
		value := values[l.ID()]
		if l.Complemented() {
			value = !value
		}
		return value
	}
	row := 0
	for i, object := range source.objects {
		if !object.IsAnd() {
			continue
		}
		id := GiaID(i)
		first, _ := object.Fanin0(id)
		second, _ := object.Fanin1(id)
		switch mode {
		case 0:
			value := read(first) && read(second)
			if enabled[2*row] && controls[row][0] {
				value = false
			}
			if enabled[2*row+1] && controls[row][1] {
				value = true
			}
			values[id] = value
		case 1:
			value := read(first) && read(second)
			if enabled[row] && controls[row][0] {
				value = !value
			}
			values[id] = value
		case 2:
			index := 0
			if values[first.ID()] {
				index++
			}
			if values[second.ID()] {
				index += 2
			}
			nominal := 3
			if first.Complemented() {
				nominal--
			}
			if second.Complemented() {
				nominal -= 2
			}
			value := enabled[4*row+index] && controls[row][index]
			if index == nominal {
				value = !value
			}
			values[id] = value
		}
		row++
	}
	outputs := make([]bool, source.OutputCount())
	for i, id := range source.cos {
		driver, _ := source.objects[id].Fanin0(id)
		outputs[i] = read(driver)
	}
	return outputs
}

func checkFaultTransform(t *testing.T, mode int, enabled []bool, build func(*GiaManager, []bool, GiaFaultTransformOptions) (*GiaFaultTransform, error)) {
	t.Helper()
	source := faultTransformSource(t)
	result, err := build(source, enabled, GiaFaultTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	perGate := len(result.Controls[0])
	if result.Graph.InputCount() != source.InputCount()+source.AndCount()*perGate || len(result.SourceGates) != source.AndCount() {
		t.Fatal(result.Graph.InputCount(), result.SourceGates, result.Controls)
	}
	for mask := 0; mask < 1<<result.Graph.InputCount(); mask++ {
		inputs := make([]bool, result.Graph.InputCount())
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		controls := make([][]bool, source.AndCount())
		position := source.PrimaryInputCount()
		for row := range controls {
			controls[row] = append([]bool(nil), inputs[position:position+perGate]...)
			position += perGate
		}
		want := evaluateFaultModel(t, source, inputs[:source.PrimaryInputCount()], controls, mode, enabled)
		got, err := result.Graph.Evaluate(inputs)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("mode=%d mask=%b got=%v want=%v err=%v", mode, mask, got, want, err)
		}
	}
}

func TestGiaFaultStructuralTransforms(t *testing.T) {
	checkFaultTransform(t, 0, []bool{true, false, false, true}, UnfoldGiaStuckAt)
	checkFaultTransform(t, 1, []bool{true, true}, UnfoldGiaFlip)
	checkFaultTransform(t, 2, []bool{true, false, true, true, false, true, false, true}, UnfoldGiaFourOutputFunction)
}

func TestGiaFaultInputCofactor(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	r, _ := source.AppendCI()
	first, _ := source.AppendAnd(a, b)
	next, _ := source.AppendAnd(first, r.Not())
	_, _ = source.AppendCO(first)
	_, _ = source.AppendCO(next)
	_ = source.SetRegisterCount(1)
	result, err := CofactorGiaFaultInputs(source, []bool{true, false}, GiaFaultTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.InputCount() != source.InputCount() || result.RegisterCount() != 1 {
		t.Fatal(result.InputCount(), result.RegisterCount())
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := result.Evaluate(inputs)
		want, sourceErr := source.Evaluate([]bool{true, false, inputs[2]})
		if err != nil || sourceErr != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(mask, got, want, err, sourceErr)
		}
	}
}

func temporalFaultSource(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	a, _ := g.AppendCI()
	r, _ := g.AppendCI()
	gate, _ := g.AppendAnd(a, r)
	_, _ = g.AppendCO(gate)
	_, _ = g.AppendCO(a)
	_ = g.SetRegisterCount(1)
	return g
}

func TestGiaTemporalFaultUnfold(t *testing.T) {
	source := temporalFaultSource(t)
	for _, tc := range []struct {
		muxes, registers bool
	}{
		{false, false}, {true, false}, {false, true}, {true, true},
	} {
		result, err := UnfoldGiaTemporalFault(source, tc.muxes, tc.registers, GiaFaultTransformOptions{})
		if err != nil {
			t.Fatal(tc, err)
		}
		if result.Graph.RegisterCount() != 0 || result.Graph.InputCount() != 4 || result.Graph.OutputCount() != 2 || len(result.Controls) != 1 {
			t.Fatal(tc, result.Graph.InputCount(), result.Graph.OutputCount(), result.Controls)
		}
		for mask := 0; mask < 16; mask++ {
			ro0 := mask&1 != 0
			pi0 := mask&2 != 0
			pi1 := mask&4 != 0
			control := mask&8 != 0
			got, err := result.Graph.Evaluate([]bool{ro0, pi0, pi1, control})
			po := pi1 && pi0
			ri := pi1
			if tc.muxes && control {
				if tc.registers {
					ri = pi0
				} else {
					po = pi0 && ro0
				}
			}
			want := []bool{po, ri}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal(tc, mask, got, want, err)
			}
		}
	}
}

func TestGiaFaultSingleFrameSequentialPartition(t *testing.T) {
	source := temporalFaultSource(t)
	result, err := UnfoldGiaFlip(source, []bool{true}, GiaFaultTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph.PrimaryInputCount() != 2 || result.Graph.RegisterCount() != 1 || result.Graph.InputCount() != 3 {
		t.Fatal(result.Graph.PrimaryInputCount(), result.Graph.RegisterCount(), result.Graph.InputCount())
	}
	for mask := 0; mask < 8; mask++ {
		a := mask&1 != 0
		control := mask&2 != 0
		r := mask&4 != 0
		outputs, next, err := result.Graph.Step([]bool{a, control}, []bool{r})
		want := a && r
		if control {
			want = !want
		}
		if err != nil || !reflect.DeepEqual(outputs, []bool{want}) || !reflect.DeepEqual(next, []bool{a}) {
			t.Fatal(mask, outputs, next, want, err)
		}
	}
}

func TestGiaFaultTransformFailures(t *testing.T) {
	source := faultTransformSource(t)
	for _, call := range []func() error{
		func() error { _, err := UnfoldGiaStuckAt(source, nil, GiaFaultTransformOptions{}); return err },
		func() error { _, err := UnfoldGiaFlip(source, nil, GiaFaultTransformOptions{}); return err },
		func() error {
			_, err := UnfoldGiaFourOutputFunction(source, nil, GiaFaultTransformOptions{})
			return err
		},
		func() error {
			_, err := CofactorGiaFaultInputs(source, []bool{true, false, true}, GiaFaultTransformOptions{})
			return err
		},
	} {
		if err := call(); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
	}
	for _, options := range []GiaFaultTransformOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {Stop: func() bool { return true }}} {
		if result, err := UnfoldGiaFlip(source, []bool{true, true}, options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(result, err)
		}
	}
	extended := NewGiaManager()
	a, _ := extended.AppendCI()
	b, _ := extended.AppendCI()
	x, _ := extended.AppendXorReal(a, b)
	_, _ = extended.AppendCO(x)
	if result, err := UnfoldGiaFlip(extended, []bool{true}, GiaFaultTransformOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzGiaFaultFlipSemantics(f *testing.F) {
	for _, seed := range []byte{0, 1, 3, 7, 15, 255} {
		f.Add(seed, seed^0x55)
	}
	f.Fuzz(func(t *testing.T, flags, assignment byte) {
		enabled := []bool{flags&1 != 0, flags&2 != 0}
		source := faultTransformSource(t)
		result, err := UnfoldGiaFlip(source, enabled, GiaFaultTransformOptions{})
		if err != nil {
			t.Fatal(err)
		}
		inputs := make([]bool, 4)
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		controls := [][]bool{{inputs[2]}, {inputs[3]}}
		want := evaluateFaultModel(t, source, inputs[:2], controls, 1, enabled)
		got, err := result.Graph.Evaluate(inputs)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(flags, assignment, got, want, err)
		}
	})
}
