package abc

// Go-only validation for src/aig/miniaig headers and module.make.

import (
	"encoding/binary"
	"errors"
	"slices"
	"strings"
	"testing"
)

func evaluateSingleAIG(t testing.TB, aig *MiniAIG, inputs []bool) bool {
	t.Helper()
	outputs, err := aig.Evaluate(inputs)
	if err != nil || len(outputs) != 1 {
		t.Fatalf("evaluate: %v %v", outputs, err)
	}
	return outputs[0]
}

func TestAbcOperatorNameInventory(t *testing.T) {
	expected := map[AbcOperator]string{
		AbcOperPI: "pi", AbcOperPO: "po", AbcOperCI: "ci", AbcOperCO: "co", AbcOperFON: "fon", AbcOperBox: "box",
		AbcOperBitBuf: "buf", AbcOperBitInv: "~", AbcOperBitAnd: "&", AbcOperBitNand: "~&", AbcOperBitOr: "|", AbcOperBitNor: "~|", AbcOperBitXor: "^", AbcOperBitNxor: "~^", AbcOperBitMux: "mux", AbcOperBitMaj: "maj",
		AbcOperRedAnd: "&", AbcOperRedNand: "~&", AbcOperRedOr: "|", AbcOperRedNor: "~|", AbcOperRedXor: "^", AbcOperRedNxor: "~^",
		AbcOperLogicNot: "!", AbcOperLogicImpl: "=>", AbcOperLogicAnd: "&&", AbcOperLogicOr: "||", AbcOperLogicXor: "^^",
		AbcOperAriAdd: "+", AbcOperAriSub: "-", AbcOperAriMul: "*", AbcOperAriSMul: "*", AbcOperAriDiv: "/", AbcOperAriRem: "%", AbcOperAriMod: "mod", AbcOperAriPow: "**", AbcOperAriMin: "-", AbcOperAriSqrt: "sqrt", AbcOperAriSquare: "squar",
		AbcOperCompEqual: "==", AbcOperCompNotEqual: "!=", AbcOperCompLess: "<", AbcOperCompMore: ">", AbcOperCompLessEqual: "<=", AbcOperCompMoreEqual: ">=",
		AbcOperShiftL: "<<", AbcOperShiftR: ">>", AbcOperShiftLA: "<<<", AbcOperShiftRA: ">>>", AbcOperShiftRotL: "rotL", AbcOperShiftRotR: "rotR",
		AbcOperDFFRSE: "DFFRSE", AbcOperSlice: "[:]", AbcOperConcat: "{}", AbcOperZeroPad: "zPad", AbcOperSignExt: "sExt", AbcOperConst: "const", AbcOperTable: "table", AbcOperLUT: "lut",
	}
	for operator := AbcOperNone; operator <= AbcOperLast; operator++ {
		got, okay := AbcOperatorName(operator)
		want, exists := expected[operator]
		if okay != exists || got != want {
			t.Fatalf("operator %d got=%q/%v want=%q/%v", operator, got, okay, want, exists)
		}
	}
}

func TestMiniAIGLiteralsOperationsAndTruthExhaustive(t *testing.T) {
	if AbcOperLast != 99 || NDRUnknown != 11 || MiniAIGNull != 0x7fffffff || MiniLUTNull2 != 0x7ffffffe {
		t.Fatal("source numeric domains changed")
	}
	if AIGVarLiteral(7, true) != 15 || AIGLiteral(15).Variable() != 7 || !AIGLiteral(15).Complemented() || AIGLiteral(15).Regular() != 14 || AIGLiteral(14).NotIf(true) != 15 {
		t.Fatal("literal algebra")
	}
	for truth4 := uint32(0); truth4 < 16; truth4++ {
		aig, err := NewMiniAIG(MiniNetworkLimits{MaxNodes: 256})
		if err != nil {
			t.Fatal(err)
		}
		x, _ := aig.CreatePI()
		y, _ := aig.CreatePI()
		truth := truth4 * 0x11111111
		result, err := aig.Truth([]AIGLiteral{x, y}, truth)
		if err != nil {
			t.Fatalf("truth %x: %v", truth4, err)
		}
		if _, err = aig.CreatePO(result); err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			got := evaluateSingleAIG(t, aig, []bool{assignment&1 != 0, assignment&2 != 0})
			want := truth4>>assignment&1 != 0
			if got != want {
				t.Fatalf("truth %x assignment %d: %v want %v", truth4, assignment, got, want)
			}
		}
	}
	for truth8 := uint32(0); truth8 < 256; truth8++ {
		aig, _ := NewMiniAIG(MiniNetworkLimits{MaxNodes: 1024})
		vars := make([]AIGLiteral, 3)
		for index := range vars {
			vars[index], _ = aig.CreatePI()
		}
		result, err := aig.Truth(vars, truth8*0x01010101)
		if err != nil {
			t.Fatalf("truth %02x: %v", truth8, err)
		}
		_, _ = aig.CreatePO(result)
		for assignment := 0; assignment < 8; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			if got, want := evaluateSingleAIG(t, aig, inputs), truth8&(1<<uint(assignment)) != 0; got != want {
				t.Fatalf("truth %02x assignment %d got=%v want=%v", truth8, assignment, got, want)
			}
		}
	}

	aig, _ := NewMiniAIG(MiniNetworkLimits{})
	x, _ := aig.CreatePI()
	y, _ := aig.CreatePI()
	c, _ := aig.CreatePI()
	and, _ := aig.AND(x, y)
	or, _ := aig.OR(x, y)
	xor, _ := aig.XOR(x, y)
	mux, _ := aig.MUX(c, x, y)
	for _, lit := range []AIGLiteral{and, or, xor, mux} {
		_, _ = aig.CreatePO(lit)
	}
	for assignment := 0; assignment < 8; assignment++ {
		xv, yv, cv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		got, err := aig.Evaluate([]bool{xv, yv, cv})
		want := []bool{xv && yv, xv || yv, xv != yv, map[bool]bool{true: xv, false: yv}[cv]}
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("%d: %v want %v: %v", assignment, got, want, err)
		}
	}
	pis, pos, ands := aig.Counts()
	if pis != 3 || pos != 4 || ands != 8 {
		t.Fatal(pis, pos, ands)
	}
	if aig.Stats() != "PI = 3. PO = 4. Node = 8." {
		t.Fatal(aig.Stats())
	}
	if text, err := aig.Verilog("m"); err != nil || !strings.Contains(text, "module m") || !strings.Contains(text, "endmodule") {
		t.Fatal(text, err)
	}
}

func TestMiniAIGLUTConversionsAndBinaryRoundTrips(t *testing.T) {
	aig, _ := NewMiniAIG(MiniNetworkLimits{})
	x, _ := aig.CreatePI()
	y, _ := aig.CreatePI()
	value, _ := aig.XOR(x.Not(), y)
	_, _ = aig.CreatePO(value.Not())
	_ = aig.SetRegisterCount(1)
	lut, err := MiniAIGToLUT(aig, MiniNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	back, err := MiniLUTToAIG(lut, MiniNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		a, _ := aig.Evaluate(inputs)
		l, _ := lut.Evaluate(inputs)
		b, _ := back.Evaluate(inputs)
		if !slices.Equal(a, l) || !slices.Equal(a, b) {
			t.Fatal(assignment, a, l, b)
		}
	}
	if lut.RegisterCount() != 1 || back.RegisterCount() != 1 {
		t.Fatal("register count")
	}
	aigBytes, _ := aig.MarshalBinary()
	aigCopy, err := ParseMiniAIG(aigBytes, MiniNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	lutBytes, _ := lut.MarshalBinary()
	lutCopy, err := ParseMiniLUT(lutBytes, MiniNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(aigBytes, func() []byte { b, _ := aigCopy.MarshalBinary(); return b }()) || !slices.Equal(lutBytes, func() []byte { b, _ := lutCopy.MarshalBinary(); return b }()) {
		t.Fatal("round trip")
	}
	if _, err = ParseMiniAIG(aigBytes[:len(aigBytes)-1], MiniNetworkLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = ParseMiniLUT(lutBytes[:len(lutBytes)-1], MiniNetworkLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestMiniBinaryLayoutsAndMalformedCanonicalRecords(t *testing.T) {
	aig, _ := NewMiniAIG(MiniNetworkLimits{})
	input, _ := aig.CreatePI()
	_, _ = aig.CreatePO(input.Not())
	_ = aig.SetRegisterCount(1)
	data, err := aig.MarshalBinary()
	want := []byte{
		6, 0, 0, 0, 1, 0, 0, 0,
		0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0x7f,
		0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0x7f,
		3, 0, 0, 0, 0xff, 0xff, 0xff, 0x7f,
	}
	if err != nil || !slices.Equal(data, want) {
		t.Fatalf("AIG bytes=%x want=%x err=%v", data, want, err)
	}
	badAIG := slices.Clone(data)
	binary.LittleEndian.PutUint32(badAIG[8:], 0)
	if _, err = ParseMiniAIG(badAIG, MiniNetworkLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	lut, _ := NewMiniLUT(2, MiniNetworkLimits{})
	lutInput, _ := lut.CreatePI()
	lutNode, _ := lut.CreateNode([]LUTNodeID{lutInput}, []uint32{0xaaaaaaaa})
	_, _ = lut.CreatePO(lutNode)
	lutData, err := lut.MarshalBinary()
	if err != nil || len(lutData) != 12+4*5*3 || binary.LittleEndian.Uint32(lutData) != 5 || binary.LittleEndian.Uint32(lutData[8:]) != 2 {
		t.Fatal(len(lutData), err)
	}
	// Fanins precede the complete truth plane. The PO marker is its second
	// fanin and each field is a fixed little-endian 32-bit word.
	if binary.LittleEndian.Uint32(lutData[12+4*(2*4+1):]) != uint32(MiniLUTNull2) || binary.LittleEndian.Uint32(lutData[12+4*2*5+4*3:]) != 0xaaaaaaaa {
		t.Fatalf("LUT layout %x", lutData)
	}
	badLUT := slices.Clone(lutData)
	binary.LittleEndian.PutUint32(badLUT[12:], 0)
	if _, err = ParseMiniLUT(badLUT, MiniNetworkLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	design, _ := NewNDRDesign(9, MiniNetworkLimits{})
	module, _ := design.AddModule(10)
	_ = design.AddObject(module, NDRObjectSpec{Operator: AbcOperConst, Outputs: []int{11}, Function: []byte("abcd")})
	ndrData, err := design.MarshalBinary()
	if err != nil || len(ndrData)%5 != 0 {
		t.Fatal(err, len(ndrData))
	}
	records := len(ndrData) / 5
	if binary.LittleEndian.Uint32(ndrData) != uint32(records) || NDRRecordType(ndrData[4*records]) != NDRDesign {
		t.Fatalf("NDR layout %x", ndrData)
	}
	objects, _ := design.Objects(module)
	function, _ := design.ObjectEntries(objects[0], NDRFunction)
	if len(function) != 2 || uint32(function[0]) != 0x64636261 || function[1] != 0 {
		t.Fatal(function)
	}
}

func TestMiniLUTToAIGCompactTruthExhaustive(t *testing.T) {
	for variables := 1; variables <= 4; variables++ {
		truthLimit := 1 << uint(1<<uint(variables))
		if variables == 4 {
			truthLimit = 1024
		}
		for truth := 0; truth < truthLimit; truth++ {
			lut, err := NewMiniLUT(5, MiniNetworkLimits{MaxNodes: 32})
			if err != nil {
				t.Fatal(err)
			}
			fanins := make([]LUTNodeID, variables)
			for index := range fanins {
				fanins[index], _ = lut.CreatePI()
			}
			words := make([]uint32, 1)
			words[0] = uint32(truth)
			node, err := lut.CreateNode(fanins, words)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = lut.CreatePO(node)
			aig, err := MiniLUTToAIG(lut, MiniNetworkLimits{MaxNodes: 512})
			if err != nil {
				t.Fatalf("vars=%d truth=%x: %v", variables, truth, err)
			}
			for assignment := 0; assignment < 1<<uint(variables); assignment++ {
				inputs := make([]bool, variables)
				for index := range inputs {
					inputs[index] = assignment&(1<<uint(index)) != 0
				}
				got := evaluateSingleAIG(t, aig, inputs)
				want := truth&(1<<uint(assignment)) != 0
				if got != want {
					t.Fatalf("vars=%d truth=%x assignment=%d got=%v want=%v", variables, truth, assignment, got, want)
				}
			}
		}
	}
	zero, _ := NewMiniLUT(2, MiniNetworkLimits{})
	truth := []uint32{1}
	id, err := zero.CreateNode(nil, truth)
	if err != nil || !zero.IsPI(id) || zero.IsNode(id) {
		t.Fatal(id, err)
	}
}

func TestMiniLUTTruthEvaluationAndOwnership(t *testing.T) {
	lut, err := NewMiniLUT(3, MiniNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := lut.CreatePI()
	b, _ := lut.CreatePI()
	c, _ := lut.CreatePI()
	truth := []uint32{0x96969696}
	node, err := lut.CreateNode([]LUTNodeID{a, b, c}, truth)
	if err != nil {
		t.Fatal(err)
	}
	truth[0] = 0
	_, _ = lut.CreatePO(node)
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, err := lut.Evaluate(inputs)
		want := inputs[0] != inputs[1] != inputs[2]
		if err != nil || len(got) != 1 || got[0] != want {
			t.Fatal(assignment, got, want, err)
		}
	}
	fanins, _ := lut.Fanins(node)
	fanins[0] = 0
	owned, _ := lut.Fanins(node)
	if owned[0] != a {
		t.Fatal(owned)
	}
	if lut.Stats() != "PI = 3. PO = 1. LUT = 1." {
		t.Fatal(lut.Stats())
	}
}

func TestNDRDesignLayoutIterationAndBinaryRoundTrip(t *testing.T) {
	design, err := NewNDRDesign(1, MiniNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	module, err := design.AddModule(2)
	if err != nil {
		t.Fatal(err)
	}
	if err = design.AddObject(module, NDRObjectSpec{Operator: AbcOperCI, RangeLeft: 3, Outputs: []int{10}}); err != nil {
		t.Fatal(err)
	}
	if err = design.AddObject(module, NDRObjectSpec{Operator: AbcOperAriAdd, RangeLeft: 3, Inputs: []int{10, 11}, Outputs: []int{12}, Function: []byte("fn")}); err != nil {
		t.Fatal(err)
	}
	if err = design.AddObject(module, NDRObjectSpec{Operator: AbcOperCO, RangeLeft: 3, Inputs: []int{12}}); err != nil {
		t.Fatal(err)
	}
	modules, _ := design.Modules()
	objects, _ := design.Objects(module)
	if !slices.Equal(modules, []int{module}) || len(objects) != 3 {
		t.Fatal(modules, objects)
	}
	cis, cos, nodes, err := design.Counts(module)
	if err != nil || cis != 1 || cos != 1 || nodes != 1 {
		t.Fatal(cis, cos, nodes, err)
	}
	inputs, _ := design.ObjectEntries(objects[1], NDRInput)
	if !slices.Equal(inputs, []int{10, 11}) {
		t.Fatal(inputs)
	}
	data, _ := design.MarshalBinary()
	copyDesign, err := ParseNDRDesign(data, MiniNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	data2, _ := copyDesign.MarshalBinary()
	if !slices.Equal(data, data2) {
		t.Fatal("NDR round trip")
	}
	if NDRRangeText([]int{7}, false) != "[7]" || NDRRangeText([]int{7, 0}, true) != "[7:0]" {
		t.Fatal("range")
	}
	if name, ok := AbcOperatorName(AbcOperAriAdd); !ok || name != "+" {
		t.Fatal(name, ok)
	}
	names := []string{"", "design", "module", "", "", "", "", "", "", "", "a", "b", "sum"}
	verilog, err := design.Verilog(names)
	if err != nil || !strings.Contains(verilog, "module module") || !strings.Contains(verilog, "assign sum = a + b;") {
		t.Fatal(verilog, err)
	}
}

func TestNDRArraysModulesAndLUTFunctionContracts(t *testing.T) {
	design, _ := NewNDRDesign(1, MiniNetworkLimits{})
	first, _ := design.AddModule(2)
	second, _ := design.AddModule(3)
	before := design.RecordCount()
	if err := design.AddObject(first, NDRObjectSpec{Operator: AbcOperCI, Outputs: []int{1}}); !errors.Is(err, ErrDimension) || design.RecordCount() != before {
		t.Fatal(err, design.RecordCount(), before)
	}
	if err := design.AddObject(second, NDRObjectSpec{Operator: AbcOperLUT, Inputs: []int{1}, Outputs: []int{2}, Function: []byte{1, 2, 3, 4}}); !errors.Is(err, ErrDimension) || design.RecordCount() != before {
		t.Fatal(err, design.RecordCount(), before)
	}
	function := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	if err := design.AddObject(second, NDRObjectSpec{Operator: AbcOperLUT, Inputs: []int{1}, Outputs: []int{2}, Function: function}); err != nil {
		t.Fatal(err)
	}
	if err := design.AddObject(second, NDRObjectSpec{Operator: AbcOperConst, Outputs: []int{3}, Function: []byte{'a', 0, 'x'}}); err != nil {
		t.Fatal(err)
	}
	objects, _ := design.Objects(second)
	words, _ := design.ObjectEntries(objects[0], NDRFunction)
	if !slices.Equal(words, []int{0x04030201, 0x08070605}) {
		t.Fatal(words)
	}
	words, _ = design.ObjectEntries(objects[1], NDRFunction)
	if !slices.Equal(words, []int{'a'}) {
		t.Fatal(words)
	}
	target := design.RecordCount()
	design.push(NDRTarget, 77)
	moduleRecord := second - 256
	design.body[moduleRecord]++
	design.body[0]++
	targets, err := design.Targets(second)
	if err != nil || !slices.Equal(targets, []int{target}) {
		t.Fatal(targets, err)
	}
}

func TestNDRVerilogSpecialOperators(t *testing.T) {
	design, _ := NewNDRDesign(1, MiniNetworkLimits{})
	module, _ := design.AddModule(2)
	objects := []NDRObjectSpec{
		{Operator: AbcOperBitBuf, Inputs: []int{3}, Outputs: []int{4}},
		{Operator: AbcOperSlice, RangeLeft: 3, RangeRight: 1, Inputs: []int{4}, Outputs: []int{5}},
		{Operator: AbcOperConcat, Inputs: []int{3, 4}, Outputs: []int{6}},
		{Operator: AbcOperAriAdd, Inputs: []int{3, 4, 5}, Outputs: []int{7}},
		{Operator: AbcOperDFF, InstanceName: 8, Inputs: []int{3, 4}, Outputs: []int{9}},
		{Operator: AbcOperRAMR, InstanceName: 10, Inputs: []int{3, 4}, Outputs: []int{11}},
	}
	for _, object := range objects {
		if err := design.AddObject(module, object); err != nil {
			t.Fatal(err)
		}
	}
	names := []string{"", "design", "top", "a", "b", "slice", "joined", "sum", "reg0", "q", "read0", "data"}
	text, err := design.Verilog(names)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"assign b = a;", "assign slice = b[3:1];", "assign joined = {a, b};",
		"assign sum = a + b + slice;", "ABC_DFF reg0 (.q(q), .d(a), .init(b));",
		"ABC_READ read0 (.data(data), .mem_in(a), .addr(b));",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("missing %q in %q", fragment, text)
		}
	}
}

func TestMiniNetworksLimitsCancellationAtomicityAndLifecycle(t *testing.T) {
	aig, _ := NewMiniAIG(MiniNetworkLimits{MaxNodes: 2})
	x, _ := aig.CreatePI()
	before := aig.NodeCount()
	if _, err := aig.AND(x, x); !errors.Is(err, ErrLimit) || aig.NodeCount() != before {
		t.Fatal(err, aig.NodeCount())
	}
	cancelled := false
	lut, _ := NewMiniLUT(2, MiniNetworkLimits{Stop: func() bool { return cancelled }})
	cancelled = true
	if _, err := lut.CreatePI(); !errors.Is(err, ErrCancelled) || lut.NodeCount() != 2 {
		t.Fatal(err, lut.NodeCount())
	}
	design, _ := NewNDRDesign(1, MiniNetworkLimits{MaxRecords: 4})
	beforeRecords := design.RecordCount()
	if _, err := design.AddModule(2); err != nil {
		t.Fatal(err)
	}
	if err := design.AddObject(258, NDRObjectSpec{Operator: AbcOperCI, Inputs: []int{1}}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if design.RecordCount() != beforeRecords+2 {
		t.Fatal("non-atomic NDR failure")
	}
	aig.Close()
	if err := aig.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	lut.Close()
	if err := lut.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	design.Close()
	if _, err := design.Modules(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func runMiniAIGModel(data []byte) error {
	aig, err := NewMiniAIG(MiniNetworkLimits{MaxNodes: 96})
	if err != nil {
		return err
	}
	truths := []uint64{0}
	for variable := 0; variable < 4; variable++ {
		if _, err = aig.CreatePI(); err != nil {
			return err
		}
		var truth uint64
		for assignment := 0; assignment < 16; assignment++ {
			if assignment>>variable&1 != 0 {
				truth |= 1 << assignment
			}
		}
		truths = append(truths, truth)
	}
	literalTruth := func(l AIGLiteral) uint64 {
		value := truths[l.Variable()]
		if l.Complemented() {
			return (^value) & 0xffff
		}
		return value
	}
	for i := 0; i+2 < len(data) && aig.NodeCount() < 90; i += 3 {
		oldNodeCount := aig.NodeCount()
		litCount := 2 * aig.NodeCount()
		x := AIGLiteral(int(data[i]) % litCount)
		y := AIGLiteral(int(data[i+1]) % litCount)
		var lit AIGLiteral
		switch data[i+2] % 3 {
		case 0:
			lit, err = aig.AND(x, y)
		case 1:
			lit, err = aig.OR(x, y)
		case 2:
			lit, err = aig.XOR(x, y)
		}
		if err != nil {
			return err
		}
		_ = lit
		for nodeID := oldNodeCount; nodeID < aig.NodeCount(); nodeID++ {
			node := aig.nodes[nodeID]
			value := literalTruth(AIGLiteral(node.fanin0)) & literalTruth(AIGLiteral(node.fanin1))
			truths = append(truths, value)
		}
	}
	target := AIGLiteral(2 * (aig.NodeCount() - 1))
	_, err = aig.CreatePO(target)
	if err != nil {
		return err
	}
	for assignment := 0; assignment < 16; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		outputs, e := aig.Evaluate(inputs)
		if e != nil || len(outputs) != 1 || outputs[0] != (literalTruth(target)>>assignment&1 != 0) {
			return ErrInvariant
		}
	}
	encoded, err := aig.MarshalBinary()
	if err != nil {
		return err
	}
	parsed, err := ParseMiniAIG(encoded, MiniNetworkLimits{MaxNodes: 96})
	if err != nil {
		return err
	}
	reencoded, err := parsed.MarshalBinary()
	if err != nil || !slices.Equal(encoded, reencoded) {
		return ErrInvariant
	}
	lut, err := MiniAIGToLUT(aig, MiniNetworkLimits{MaxNodes: 256})
	if err != nil {
		return err
	}
	lutData, err := lut.MarshalBinary()
	if err != nil {
		return err
	}
	lutParsed, err := ParseMiniLUT(lutData, MiniNetworkLimits{MaxNodes: 256})
	if err != nil {
		return err
	}
	back, err := MiniLUTToAIG(lutParsed, MiniNetworkLimits{MaxNodes: 1024})
	if err != nil {
		return err
	}
	for assignment := 0; assignment < 16; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		got, gotErr := back.Evaluate(inputs)
		want := literalTruth(target)>>assignment&1 != 0
		if gotErr != nil || len(got) != 1 || got[0] != want {
			return ErrInvariant
		}
	}
	return nil
}

func TestMiniAIGIndependentModelCampaign(t *testing.T) {
	state := uint64(0xd1b54a32d192ed03)
	data := make([]byte, 96)
	for execution := 0; execution < 5000; execution++ {
		for i := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[i] = byte(state)
		}
		if err := runMiniAIGModel(data); err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
	}
}

func FuzzMiniAIGModel(f *testing.F) {
	f.Add([]byte{1, 2, 0, 3, 4, 1, 5, 6, 2})
	f.Add([]byte{255, 0, 2, 7, 9, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			return
		}
		if err := runMiniAIGModel(data); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzMiniBinaryParsers(f *testing.F) {
	f.Add(uint8(0), []byte{2, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0x7f})
	f.Add(uint8(1), []byte{
		2, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0,
		0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0x7f,
		0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0x7f,
		0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff,
	})
	f.Add(uint8(2), []byte{2, 0, 0, 0, 9, 0, 0, 0, 1, 7})
	f.Fuzz(func(t *testing.T, kind uint8, data []byte) {
		if len(data) > 4096 {
			return
		}
		limits := MiniNetworkLimits{MaxNodes: 512, MaxRecords: 1024}
		switch kind % 3 {
		case 0:
			value, err := ParseMiniAIG(data, limits)
			if err != nil {
				return
			}
			encoded, err := value.MarshalBinary()
			if err != nil || !slices.Equal(encoded, data) {
				t.Fatal(err)
			}
		case 1:
			value, err := ParseMiniLUT(data, limits)
			if err != nil {
				return
			}
			encoded, err := value.MarshalBinary()
			if err != nil || !slices.Equal(encoded, data) {
				t.Fatal(err)
			}
		case 2:
			value, err := ParseNDRDesign(data, limits)
			if err != nil {
				return
			}
			encoded, err := value.MarshalBinary()
			if err != nil || !slices.Equal(encoded, data) {
				t.Fatal(err)
			}
		}
	})
}
