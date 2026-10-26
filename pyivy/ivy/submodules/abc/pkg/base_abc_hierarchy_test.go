package abc

// Stand-alone tests for the hierarchy and BLIF-MV behavior ported from src/base/abc.

import (
	"errors"
	"math/rand"
	"testing"
)

func hierarchyTestDesign(t *testing.T, options ABCHierarchyOptions) *ABCHierarchyDesign {
	t.Helper()
	child, err := NewABCHierarchyModule("cell", []string{"a", "b"}, options)
	if err != nil {
		t.Fatal(err)
	}
	x, err := child.AddGate("xor", ABCSOPXOR(), []ABCHierarchyEdge{{Node: child.Inputs[0]}, {Node: child.Inputs[1]}})
	if err != nil {
		t.Fatal(err)
	}
	if err := child.SetOutputs([]string{"y"}, []ABCHierarchyEdge{x}); err != nil {
		t.Fatal(err)
	}
	top, err := NewABCHierarchyModule("top", []string{"p", "q", "r"}, options)
	if err != nil {
		t.Fatal(err)
	}
	box, err := top.AddInstance("u0", "cell", []ABCHierarchyEdge{{Node: top.Inputs[0]}, {Node: top.Inputs[1], Complemented: true}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	or, err := top.AddGate("or", ABCSOP{Variables: 2, Cubes: []string{"1-", "-1"}, Phase: '1'}, []ABCHierarchyEdge{box[0], {Node: top.Inputs[2]}})
	if err != nil {
		t.Fatal(err)
	}
	if err := top.SetOutputs([]string{"out", "trace"}, []ABCHierarchyEdge{or, box[0].Not()}); err != nil {
		t.Fatal(err)
	}
	design, err := NewABCHierarchyDesign("tiny", options)
	if err != nil {
		t.Fatal(err)
	}
	if err := design.AddModule(child, false); err != nil {
		t.Fatal(err)
	}
	if err := design.AddModule(top, true); err != nil {
		t.Fatal(err)
	}
	return design
}

func TestABCMVRelationRoundTripEvaluateAndCost(t *testing.T) {
	text := "# deterministic multi-valued relation\nd 0\n0 - 1\n1 2 2\n"
	relation, err := ParseABCMVRelation(text, []int{2, 3}, 3, ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	formatted, err := relation.Format(ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	again, err := ParseABCMVRelation(formatted, []int{2, 3}, 3, ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if formatted2, err := again.Format(ABCHierarchyOptions{}); err != nil || formatted2 != formatted {
		t.Fatalf("non-deterministic round trip: %q, %v", formatted2, err)
	}
	checks := []struct {
		in   []int
		want []bool
	}{{[]int{0, 1}, []bool{false, true, false}}, {[]int{1, 2}, []bool{false, false, true}}, {[]int{1, 0}, []bool{true, false, false}}}
	for _, check := range checks {
		got, err := relation.Evaluate(check.in)
		if err != nil || !equalHierarchyBools(got, check.want) {
			t.Fatalf("Evaluate(%v) = %v, %v; want %v", check.in, got, err, check.want)
		}
	}
	if cost, err := relation.Cost(); err != nil || cost != 5 {
		t.Fatalf("Cost = %d, %v; want 5", cost, err)
	}
	copyDefault, err := ParseABCMVRelation("d =0\n", []int{3}, 3, ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := copyDefault.Evaluate([]int{2})
	if err != nil || !equalHierarchyBools(got, []bool{false, false, true}) {
		t.Fatalf("copied default = %v, %v", got, err)
	}
}

func TestABCMVRelationAbsentDefaultAndDefaultWidth(t *testing.T) {
	relation, err := ParseABCMVRelation("0 1\n", []int{2}, 2, ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if text, err := relation.Format(ABCHierarchyOptions{}); err != nil || text != "0 1\n" {
		t.Fatalf("absent default round trip = %q, %v", text, err)
	}
	got, err := relation.Evaluate([]int{1})
	if err != nil || !equalHierarchyBools(got, []bool{false, false}) {
		t.Fatalf("uncovered relation = %v, %v; want no asserted value", got, err)
	}
	if _, err := ParseABCMVRelation("d =0\n", []int{2}, 3, ABCHierarchyOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("accepted unequal copied-default domains: %v", err)
	}
}

func TestABCMVRelationRejectsMalformedAndCancels(t *testing.T) {
	bad := []string{"d 0\n0\n", "d =2\n", "d 0\n2 1\n", "d 0\n0 3\n", "d 0\x00\n"}
	for _, text := range bad {
		if _, err := ParseABCMVRelation(text, []int{2}, 3, ABCHierarchyOptions{}); err == nil {
			t.Fatalf("accepted malformed relation %q", text)
		}
	}
	_, err := ParseABCMVRelation("d 0\n", []int{2}, 2, ABCHierarchyOptions{Stop: func() bool { return true }})
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation = %v", err)
	}
	if _, err := ParseABCMVRelation("d 0\n", []int{2}, 2, ABCHierarchyOptions{MaxTextBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("text limit = %v", err)
	}
	relation, _ := ParseABCMVRelation("d 0\n", []int{2}, 2, ABCHierarchyOptions{})
	if _, err := relation.Format(ABCHierarchyOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("format cancellation = %v", err)
	}
}

func TestABCHierarchyFlattenTraceAndGia(t *testing.T) {
	design := hierarchyTestDesign(t, ABCHierarchyOptions{})
	serialized, err := design.Format(ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	serializedAgain, err := design.Format(ABCHierarchyOptions{})
	if err != nil || serializedAgain != serialized {
		t.Fatalf("non-deterministic hierarchy serialization: %v", err)
	}
	if _, err := design.Format(ABCHierarchyOptions{MaxTextBytes: 16}); !errors.Is(err, ErrLimit) {
		t.Fatalf("serialization limit = %v", err)
	}
	if _, err := design.Format(ABCHierarchyOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("serialization cancellation = %v", err)
	}
	flat, trace, err := design.Flatten()
	if err != nil {
		t.Fatal(err)
	}
	if len(flat.Instances) != 0 || trace.Instances != 1 || trace.ModuleUses["top"] != 1 || trace.ModuleUses["cell"] != 1 {
		t.Fatalf("bad flatten trace: instances=%d, trace=%+v", len(flat.Instances), trace)
	}
	if len(trace.FlatNames) != 2 || trace.FlatNames[0] != "u0/xor" || trace.FlatNames[1] != "or" {
		t.Fatalf("flat names lost deterministic traversal: %v", trace.FlatNames)
	}
	gia, err := design.ToGia()
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		cell := inputs[0] != !inputs[1]
		want := []bool{cell || inputs[2], !cell}
		got, err := flat.Evaluate(inputs)
		if err != nil || !equalHierarchyBools(got, want) {
			t.Fatalf("flat Evaluate(%v) = %v, %v; want %v", inputs, got, err, want)
		}
		giaGot, err := evaluateHierarchyGia(gia, inputs)
		if err != nil || !equalHierarchyBools(giaGot, want) {
			t.Fatalf("GIA Evaluate(%v) = %v, %v; want %v", inputs, giaGot, err, want)
		}
	}
	modules, instances, nodes, err := design.Statistics()
	if err != nil || modules != 2 || instances != 1 || nodes != 10 {
		t.Fatalf("Statistics = %d,%d,%d,%v", modules, instances, nodes, err)
	}
}

func TestABCHierarchySpecialXORMarkers(t *testing.T) {
	for _, phase := range []byte{'x', 'n'} {
		module, err := NewABCHierarchyModule("parity", []string{"a", "b", "c"}, ABCHierarchyOptions{})
		if err != nil {
			t.Fatal(err)
		}
		gate, err := module.AddGate("xor3", ABCSOP{Variables: 3, Cubes: []string{"---"}, Phase: phase}, []ABCHierarchyEdge{{Node: module.Inputs[0]}, {Node: module.Inputs[1]}, {Node: module.Inputs[2]}})
		if err != nil {
			t.Fatal(err)
		}
		if err := module.SetOutputs([]string{"y"}, []ABCHierarchyEdge{gate}); err != nil {
			t.Fatal(err)
		}
		design, _ := NewABCHierarchyDesign("xor-marker", ABCHierarchyOptions{})
		if err := design.AddModule(module, true); err != nil {
			t.Fatal(err)
		}
		gia, err := design.ToGia()
		if err != nil {
			t.Fatal(err)
		}
		for bits := 0; bits < 8; bits++ {
			inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			want := (inputs[0] != inputs[1]) != inputs[2]
			if phase == 'n' {
				want = !want
			}
			got, err := module.Evaluate(inputs)
			if err != nil || len(got) != 1 || got[0] != want {
				t.Fatalf("phase %c module(%v) = %v, %v; want %v", phase, inputs, got, err, want)
			}
			giaGot, err := evaluateHierarchyGia(gia, inputs)
			if err != nil || len(giaGot) != 1 || giaGot[0] != want {
				t.Fatalf("phase %c GIA(%v) = %v, %v; want %v", phase, inputs, giaGot, err, want)
			}
		}
	}
}

func TestABCHierarchyValidationLimitsCancellationAndRollback(t *testing.T) {
	design := hierarchyTestDesign(t, ABCHierarchyOptions{})
	extra, err := NewABCHierarchyModule("extra", nil, ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := extra.SetOutputs(nil, nil); err != nil {
		t.Fatal(err)
	}
	before := len(design.Modules)
	if err := design.AddModule(extra, true); !errors.Is(err, ErrFormat) || len(design.Modules) != before {
		t.Fatalf("failed root addition was not atomic: err=%v modules=%d", err, len(design.Modules))
	}

	cyclic, err := NewABCHierarchyDesign("cycle", ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := NewABCHierarchyModule("a", nil, ABCHierarchyOptions{})
	b, _ := NewABCHierarchyModule("b", nil, ABCHierarchyOptions{})
	aOut, _ := a.AddInstance("ib", "b", nil, 1)
	bOut, _ := b.AddInstance("ia", "a", nil, 1)
	_ = a.SetOutputs([]string{"o"}, aOut)
	_ = b.SetOutputs([]string{"o"}, bOut)
	_ = cyclic.AddModule(a, true)
	_ = cyclic.AddModule(b, false)
	if err := cyclic.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("recursive hierarchy validation = %v", err)
	}

	cancelled := hierarchyTestDesign(t, ABCHierarchyOptions{Stop: func() bool { return true }})
	if _, _, err := cancelled.Flatten(); !errors.Is(err, ErrCancelled) {
		t.Fatalf("flatten cancellation = %v", err)
	}
	limited := hierarchyTestDesign(t, ABCHierarchyOptions{MaxOperations: 1})
	if _, _, err := limited.Flatten(); !errors.Is(err, ErrLimit) {
		t.Fatalf("flatten operation limit = %v", err)
	}
}

func TestABCHierarchyRejectsCorruptLookupAndAtomicModuleLimit(t *testing.T) {
	design := hierarchyTestDesign(t, ABCHierarchyOptions{})
	design.byName[design.Root] = 0
	if err := design.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("corrupt module lookup = %v", err)
	}

	options := ABCHierarchyOptions{MaxNodes: 3}
	limited, err := NewABCHierarchyDesign("limited", options)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := NewABCHierarchyModule("root", []string{"a", "b"}, options)
	if err := root.SetOutputs(nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := limited.AddModule(root, true); err != nil {
		t.Fatal(err)
	}
	extra, _ := NewABCHierarchyModule("extra", nil, options)
	_ = extra.SetOutputs(nil, nil)
	before := len(limited.Modules)
	if err := limited.AddModule(extra, false); !errors.Is(err, ErrLimit) || len(limited.Modules) != before {
		t.Fatalf("limited module addition = %v, modules %d -> %d", err, before, len(limited.Modules))
	}
}

func TestABCHierarchyRejectsReorderedInstancePortsAndUnusedRecursion(t *testing.T) {
	module, _ := NewABCHierarchyModule("ports", []string{"a"}, ABCHierarchyOptions{})
	outputs, err := module.AddInstance("box", "model", []ABCHierarchyEdge{{Node: module.Inputs[0]}}, 2)
	if err != nil {
		t.Fatal(err)
	}
	module.Instances[0].Outputs[0], module.Instances[0].Outputs[1] = module.Instances[0].Outputs[1], module.Instances[0].Outputs[0]
	module.Nodes[outputs[0].Node].Port = 1
	module.Nodes[outputs[1].Node].Port = 0
	module.Instances[0].Inputs[0] = outputs[0]
	if err := module.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("reordered self-dependent instance ports = %v", err)
	}

	design, _ := NewABCHierarchyDesign("unused-cycle", ABCHierarchyOptions{})
	root, _ := NewABCHierarchyModule("root", nil, ABCHierarchyOptions{})
	_ = root.SetOutputs(nil, nil)
	a, _ := NewABCHierarchyModule("a", nil, ABCHierarchyOptions{})
	b, _ := NewABCHierarchyModule("b", nil, ABCHierarchyOptions{})
	aOut, _ := a.AddInstance("b0", "b", nil, 1)
	bOut, _ := b.AddInstance("a0", "a", nil, 1)
	_ = a.SetOutputs([]string{"o"}, aOut)
	_ = b.SetOutputs([]string{"o"}, bOut)
	_ = design.AddModule(root, true)
	_ = design.AddModule(a, false)
	_ = design.AddModule(b, false)
	if err := design.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("unused recursive modules = %v", err)
	}
}

func TestABCHierarchyBoundedCEC(t *testing.T) {
	left := hierarchyTestDesign(t, ABCHierarchyOptions{})
	right := hierarchyTestDesign(t, ABCHierarchyOptions{})
	result, err := left.CompareTruth(right, 3)
	if err != nil || !result.Equivalent || result.Counterexample != nil {
		t.Fatalf("equivalent comparison = %+v, %v", result, err)
	}
	top, _ := right.module(right.Root)
	top.Outputs[0].Complemented = !top.Outputs[0].Complemented
	result, err = left.CompareTruth(right, 3)
	if err != nil || result.Equivalent || len(result.Counterexample) != 3 {
		t.Fatalf("inequivalent comparison = %+v, %v", result, err)
	}
	if _, err := left.CompareTruth(right, 2); !errors.Is(err, ErrLimit) {
		t.Fatalf("CEC input limit = %v", err)
	}
}

func TestABCHierarchyIndependentRandomModel(t *testing.T) {
	random := rand.New(rand.NewSource(0xabc9))
	for iteration := 0; iteration < 5000; iteration++ {
		kind := random.Intn(3)
		invertA := random.Intn(2) != 0
		invertB := random.Intn(2) != 0
		outputInvert := random.Intn(2) != 0
		design, flat := randomHierarchyModel(t, kind, invertA, invertB, outputInvert)
		if err := design.Validate(); err != nil {
			t.Fatalf("iteration %d: %v", iteration, err)
		}
		for bits := 0; bits < 4; bits++ {
			a := bits&1 != 0
			b := bits&2 != 0
			left, right := a != invertA, b != invertB
			want := left && right
			if kind == 1 {
				want = left || right
			} else if kind == 2 {
				want = left != right
			}
			want = want != outputInvert
			got, err := flat.Evaluate([]bool{a, b})
			if err != nil || len(got) != 1 || got[0] != want {
				t.Fatalf("iteration %d bits %d: got %v, %v; want %v", iteration, bits, got, err, want)
			}
		}
	}
}

func FuzzABCHierarchyFlatten(f *testing.F) {
	f.Add(byte(0), byte(0))
	f.Add(byte(2), byte(7))
	f.Fuzz(func(t *testing.T, operation, flags byte) {
		kind := int(operation % 3)
		design, flat := randomHierarchyModel(t, kind, flags&1 != 0, flags&2 != 0, flags&4 != 0)
		if err := design.Validate(); err != nil {
			t.Fatal(err)
		}
		for bits := 0; bits < 4; bits++ {
			a, b := bits&1 != 0, bits&2 != 0
			left, right := a != (flags&1 != 0), b != (flags&2 != 0)
			want := left && right
			if kind == 1 {
				want = left || right
			} else if kind == 2 {
				want = left != right
			}
			want = want != (flags&4 != 0)
			got, err := flat.Evaluate([]bool{a, b})
			if err != nil || len(got) != 1 || got[0] != want {
				t.Fatalf("got %v, %v; want %v", got, err, want)
			}
		}
	})
}

func randomHierarchyModel(t *testing.T, kind int, invertA, invertB, outputInvert bool) (*ABCHierarchyDesign, *ABCHierarchyModule) {
	t.Helper()
	child, err := NewABCHierarchyModule("child", []string{"x", "y"}, ABCHierarchyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var sop ABCSOP
	switch kind {
	case 0:
		sop, _ = ABCSOPAnd(2, nil, false)
	case 1:
		sop = ABCSOP{Variables: 2, Cubes: []string{"1-", "-1"}, Phase: '1'}
	default:
		sop = ABCSOPXOR()
	}
	gate, err := child.AddGate("g", sop, []ABCHierarchyEdge{{Node: child.Inputs[0]}, {Node: child.Inputs[1]}})
	if err != nil {
		t.Fatal(err)
	}
	_ = child.SetOutputs([]string{"z"}, []ABCHierarchyEdge{gate})
	top, _ := NewABCHierarchyModule("root", []string{"a", "b"}, ABCHierarchyOptions{})
	outs, err := top.AddInstance("i", "child", []ABCHierarchyEdge{{Node: top.Inputs[0], Complemented: invertA}, {Node: top.Inputs[1], Complemented: invertB}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	outs[0].Complemented = outputInvert
	_ = top.SetOutputs([]string{"o"}, outs)
	design, _ := NewABCHierarchyDesign("random", ABCHierarchyOptions{})
	_ = design.AddModule(child, false)
	_ = design.AddModule(top, true)
	flat, _, err := design.Flatten()
	if err != nil {
		t.Fatal(err)
	}
	return design, flat
}

func equalHierarchyBools(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func evaluateHierarchyGia(manager *GiaManager, inputs []bool) ([]bool, error) {
	if manager == nil || len(inputs) != manager.InputCount() {
		return nil, ErrDimension
	}
	values := make([]bool, manager.ObjectCount())
	for index, id := range manager.Inputs() {
		values[id] = inputs[index]
	}
	read := func(lit GiaLit) bool { return values[lit.ID()] != lit.Complemented() }
	for rawID := 1; rawID < manager.ObjectCount(); rawID++ {
		id := GiaID(rawID)
		object, err := manager.Object(id)
		if err != nil {
			return nil, err
		}
		if object.IsAnd() {
			a, err := object.Fanin0(id)
			if err != nil {
				return nil, err
			}
			b, err := object.Fanin1(id)
			if err != nil {
				return nil, err
			}
			values[id] = read(a) && read(b)
		}
	}
	outputs := make([]bool, manager.OutputCount())
	for index, id := range manager.Outputs() {
		object, err := manager.Object(id)
		if err != nil {
			return nil, err
		}
		driver, err := object.Fanin0(id)
		if err != nil {
			return nil, err
		}
		outputs[index] = read(driver)
	}
	return outputs, nil
}
