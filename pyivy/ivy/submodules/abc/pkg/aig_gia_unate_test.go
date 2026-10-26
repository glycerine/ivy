package abc

// Go-only path model, exhaustive properties and fuzz tests for
// src/aig/gia/giaUnate.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func giaUnateFixture(t testing.TB) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	negativePositive, _ := m.AppendAnd(a.Not(), b)
	left, _ := m.AppendAnd(a, b.Not())
	right, _ := m.AppendAnd(a.Not(), b)
	xor, _ := m.AppendAnd(left.Not(), right.Not())
	positiveBranch, _ := m.AppendAnd(a, b)
	cancellation, _ := m.AppendAnd(positiveBranch, right)
	buffer, _ := m.AppendBuffer(c.Not())
	for _, driver := range []GiaLit{negativePositive, negativePositive.Not(), xor, cancellation, buffer} {
		if _, err := m.AppendCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	return m
}

func TestGiaStructuralUnatenessClassesSelectionsAndStats(t *testing.T) {
	m := giaUnateFixture(t)
	before := slices.Clone(m.objects)
	report, err := m.AnalyzeStructuralUnateness(nil, nil, GiaUnateOptions{})
	if err != nil || !slices.Equal(before, m.objects) || !slices.Equal(report.Inputs, []int{0, 1, 2}) || !slices.Equal(report.Outputs, []int{0, 1, 2, 3, 4}) {
		t.Fatal(report, err)
	}
	want := [][]GiaUnateClass{
		{GiaUnateNegative, GiaUnatePositive, GiaUnateIndependent},
		{GiaUnatePositive, GiaUnateNegative, GiaUnateIndependent},
		{GiaUnateBinate, GiaUnateBinate, GiaUnateIndependent},
		{GiaUnateBinate, GiaUnatePositive, GiaUnateIndependent},
		{GiaUnateIndependent, GiaUnateIndependent, GiaUnateNegative},
	}
	if len(report.Classes) != len(want) {
		t.Fatal(report.Classes)
	}
	for row := range want {
		if !slices.Equal(report.Classes[row], want[row]) {
			t.Fatal(row, report.Classes[row], want[row])
		}
	}
	if report.SupportPairs != 9 || report.UnatePairs != 6 || report.BinatePairs != 3 {
		t.Fatal(report.SupportPairs, report.UnatePairs, report.BinatePairs)
	}
	for output := range want {
		for input := range want[output] {
			class, pairErr := m.StructuralUnate(input, output, GiaUnateOptions{})
			if pairErr != nil || class != want[output][input] {
				t.Fatal(input, output, class, pairErr)
			}
		}
	}

	selected, err := m.AnalyzeStructuralUnateness([]int{2, 0, 0}, []int{4, 0, 4}, GiaUnateOptions{})
	if err != nil || !slices.Equal(selected.Inputs, []int{2, 0, 0}) || !slices.Equal(selected.Outputs, []int{4, 0, 4}) {
		t.Fatal(selected, err)
	}
	wantSelected := [][]GiaUnateClass{
		{GiaUnateNegative, GiaUnateIndependent, GiaUnateIndependent},
		{GiaUnateIndependent, GiaUnateNegative, GiaUnateNegative},
		{GiaUnateNegative, GiaUnateIndependent, GiaUnateIndependent},
	}
	for row := range wantSelected {
		if !slices.Equal(selected.Classes[row], wantSelected[row]) {
			t.Fatal(row, selected.Classes[row], wantSelected[row])
		}
	}
	empty, err := m.AnalyzeStructuralUnateness([]int{}, []int{}, GiaUnateOptions{})
	if err != nil || len(empty.Inputs) != 0 || len(empty.Outputs) != 0 || len(empty.Classes) != 0 || empty.SupportPairs != 0 {
		t.Fatal(empty, err)
	}
}

func TestGiaStructuralUnatenessErrorsCancellationAndDeep(t *testing.T) {
	m := giaUnateFixture(t)
	for _, options := range []GiaUnateOptions{{MaxObjects: -1}, {MaxEntries: -1}, {MaxObjects: 1}, {MaxEntries: 1}} {
		if report, err := m.AnalyzeStructuralUnateness(nil, nil, options); err == nil || report.Classes != nil {
			t.Fatal(options, report, err)
		}
	}
	for _, selection := range []struct{ inputs, outputs []int }{
		{[]int{-1}, []int{0}}, {[]int{3}, []int{0}}, {[]int{0}, []int{-1}}, {[]int{0}, []int{5}},
	} {
		if report, err := m.AnalyzeStructuralUnateness(selection.inputs, selection.outputs, GiaUnateOptions{}); !errors.Is(err, ErrDimension) || report.Classes != nil {
			t.Fatal(selection, report, err)
		}
	}
	if report, err := (*GiaManager)(nil).AnalyzeStructuralUnateness(nil, nil, GiaUnateOptions{}); !errors.Is(err, ErrDimension) || report.Classes != nil {
		t.Fatal(report, err)
	}
	for threshold := 1; threshold <= 8; threshold++ {
		calls := 0
		report, err := m.AnalyzeStructuralUnateness(nil, nil, GiaUnateOptions{Stop: func() bool { calls++; return calls == threshold }})
		if !errors.Is(err, ErrCancelled) || report.Classes != nil {
			t.Fatal(threshold, calls, report, err)
		}
	}
	for _, native := range []func(*GiaManager, GiaLit, GiaLit) error{
		func(g *GiaManager, a, b GiaLit) error { _, err := g.AppendXorReal(a, b); return err },
		func(g *GiaManager, a, b GiaLit) error { _, err := g.AppendMuxReal(a, b, GiaConst0); return err },
	} {
		g := NewGiaManager()
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		if err := native(g, a, b); err != nil {
			t.Fatal(err)
		}
		if report, err := g.AnalyzeStructuralUnateness(nil, nil, GiaUnateOptions{}); !errors.Is(err, ErrUnsupported) || report.Classes != nil {
			t.Fatal(report, err)
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
	report, err := deep.AnalyzeStructuralUnateness(nil, nil, GiaUnateOptions{})
	if err != nil || len(report.Classes) != 1 || len(report.Classes[0]) != 2 || report.Classes[0][0] != GiaUnateBinate || report.Classes[0][1] != GiaUnatePositive {
		t.Fatal(report, err)
	}
}

func giaUnatePathModel(m *GiaManager, input, output int) GiaUnateClass {
	target := m.cis[input]
	co := m.cos[output]
	driver, _ := m.objects[co].Fanin0(co)
	type item struct {
		id    GiaID
		phase bool
	}
	stack := []item{{driver.ID(), driver.Complemented()}}
	seen := make([][2]bool, len(m.objects))
	positive, negative := false, false
	for len(stack) != 0 {
		next := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		index := 0
		if next.phase {
			index = 1
		}
		if seen[next.id][index] {
			continue
		}
		seen[next.id][index] = true
		if next.id == target {
			if next.phase {
				negative = true
			} else {
				positive = true
			}
			continue
		}
		object := m.objects[next.id]
		if !object.IsAnd() {
			continue
		}
		first, _ := object.Fanin0(next.id)
		second, _ := object.Fanin1(next.id)
		stack = append(stack, item{second.ID(), next.phase != second.Complemented()}, item{first.ID(), next.phase != first.Complemented()})
	}
	switch {
	case positive && negative:
		return GiaUnateBinate
	case negative:
		return GiaUnateNegative
	case positive:
		return GiaUnatePositive
	default:
		return GiaUnateIndependent
	}
}

func checkGiaUnateFunctionalImplication(t *testing.T, m *GiaManager, report GiaUnateReport) {
	t.Helper()
	if m.InputCount() > 8 {
		return
	}
	for row, output := range report.Outputs {
		for column, input := range report.Inputs {
			class := report.Classes[row][column]
			if class == GiaUnateBinate {
				continue
			}
			for assignment := 0; assignment < 1<<m.InputCount(); assignment++ {
				if assignment&(1<<input) != 0 {
					continue
				}
				low := make([]bool, m.InputCount())
				for bit := range low {
					low[bit] = assignment&(1<<bit) != 0
				}
				high := slices.Clone(low)
				high[input] = true
				lowOutputs, _ := m.Evaluate(low)
				highOutputs, _ := m.Evaluate(high)
				valid := class == GiaUnatePositive && (!lowOutputs[output] || highOutputs[output]) || class == GiaUnateNegative && (!highOutputs[output] || lowOutputs[output]) || class == GiaUnateIndependent && lowOutputs[output] == highOutputs[output]
				if !valid {
					t.Fatal(input, output, class, assignment, lowOutputs[output], highOutputs[output])
				}
			}
		}
	}
}

func FuzzGiaStructuralUnateness(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 17, 1001} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		m := NewGiaManager()
		inputCount := 2 + random.Intn(5)
		literals := make([]GiaLit, inputCount)
		for i := range literals {
			literals[i], _ = m.AppendCI()
		}
		for gate := 0; gate < 32; gate++ {
			if gate%11 == 10 {
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
			result, err := m.AppendAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			literals = append(literals, result)
		}
		outputCount := 1 + random.Intn(5)
		for output := 0; output < outputCount; output++ {
			driver := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			if _, err := m.AppendCO(driver); err != nil {
				t.Fatal(err)
			}
		}
		report, err := m.AnalyzeStructuralUnateness(nil, nil, GiaUnateOptions{})
		if err != nil {
			t.Fatal(seed, err)
		}
		for output := 0; output < outputCount; output++ {
			for input := 0; input < inputCount; input++ {
				want := giaUnatePathModel(m, input, output)
				if report.Classes[output][input] != want {
					t.Fatal(seed, input, output, report.Classes[output][input], want)
				}
			}
		}
		checkGiaUnateFunctionalImplication(t, m, report)
	})
}
