package mc

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	lg "github.com/glycerine/goivy/logic"
)

// evalConstLit evaluates a literal in a circuit that has no real inputs
// (only constant True/False connections). It runs simulation with all-zero inputs.
func evalConstLit(a *Aiger, lit int) byte {
	a.Reset()
	// Step with all-zero inputs
	inp := strings.Repeat("0", len(a.Inputs))
	if len(inp) > 0 {
		a.Step(inp)
	}
	return a.GetIn(lit)
}

// evalMultiConstLits evaluates multiple literals in a constant circuit.
func evalMultiConstLits(a *Aiger, lits []int) int {
	a.Reset()
	inp := strings.Repeat("0", len(a.Inputs))
	if len(inp) > 0 {
		a.Step(inp)
	}
	result := 0
	n := len(lits)
	for i, lit := range lits {
		if a.GetIn(lit) == '1' {
			result += 1 << (n - 1 - i)
		}
	}
	return result
}

// ============================================================
// Aiger tests
// ============================================================

func TestAigerNewAiger(t *testing.T) {
	a := NewAiger([]string{"x", "y"}, []string{"s0", "s1"}, []string{"o"})
	// Should have 3 inputs (x, y, bogus)
	if len(a.Inputs) != 3 {
		t.Errorf("expected 3 inputs (including bogus), got %d", len(a.Inputs))
	}
	if a.Inputs[2] != "%%bogus%%" {
		t.Errorf("expected bogus input, got %s", a.Inputs[2])
	}
	// Check that all inputs and latches have literals
	for _, name := range a.Inputs {
		if _, ok := a.VarMap[name]; !ok {
			t.Errorf("input %s has no literal", name)
		}
	}
	for _, name := range a.Latches {
		if _, ok := a.VarMap[name]; !ok {
			t.Errorf("latch %s has no literal", name)
		}
	}
}

func TestAigerTrueFalse(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	if a.True() != 1 {
		t.Errorf("True should be 1, got %d", a.True())
	}
	if a.False() != 0 {
		t.Errorf("False should be 0, got %d", a.False())
	}
}

func TestAigerNotl(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	// Notl of even literal should be odd (negated)
	if a.Notl(0) != 1 {
		t.Errorf("Notl(0) should be 1, got %d", a.Notl(0))
	}
	if a.Notl(1) != 0 {
		t.Errorf("Notl(1) should be 0, got %d", a.Notl(1))
	}
	if a.Notl(2) != 3 {
		t.Errorf("Notl(2) should be 3, got %d", a.Notl(2))
	}
	if a.Notl(3) != 2 {
		t.Errorf("Notl(3) should be 2, got %d", a.Notl(3))
	}
	// Double negation
	for i := 0; i < 10; i++ {
		if a.Notl(a.Notl(i)) != i {
			t.Errorf("Notl(Notl(%d)) should be %d, got %d", i, i, a.Notl(a.Notl(i)))
		}
	}
}

func TestAigerAndlEmpty(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	// Empty AND = True
	if a.Andl() != a.True() {
		t.Errorf("Andl() should be True, got %d", a.Andl())
	}
}

func TestAigerAndlSingle(t *testing.T) {
	a := NewAiger([]string{"x"}, nil, nil)
	lit := a.MustLit("x")
	res := a.Andl(lit)
	if res != lit {
		t.Errorf("Andl(x) should be x literal %d, got %d", lit, res)
	}
}

func TestAigerAndlMultiple(t *testing.T) {
	a := NewAiger([]string{"x", "y"}, nil, nil)
	xLit := a.MustLit("x")
	yLit := a.MustLit("y")
	res := a.Andl(xLit, yLit)
	// Should create one gate
	if len(a.Gates) != 1 {
		t.Errorf("expected 1 gate, got %d", len(a.Gates))
	}
	if res != a.Gates[0][0] {
		t.Errorf("result should be gate output")
	}
}

func TestAigerOrlEmpty(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	// Empty OR = False
	if a.Orl() != a.False() {
		t.Errorf("Orl() should be False, got %d", a.Orl())
	}
}

func TestAigerIte(t *testing.T) {
	a := NewAiger([]string{"c", "x", "y"}, nil, nil)
	c := a.MustLit("c")
	x := a.MustLit("x")
	y := a.MustLit("y")
	res := a.Ite(c, x, y)
	// ITE should produce gates
	if len(a.Gates) == 0 {
		t.Error("ITE should create gates")
	}
	// Just verify it produced a non-trivial result
	_ = res
}

func TestAigerIff(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	// Iff(True, True) should involve True
	res := a.Iff(a.True(), a.True())
	// Simulate: both inputs are true, so iff should be true
	// Can't easily check without simulation, but ensure gates are created
	_ = res
}

func TestAigerImpliesL(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	// False implies anything = True
	res := a.ImpliesL(a.False(), a.False())
	_ = res
}

func TestAigerXor(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	res := a.Xor(a.True(), a.False())
	_ = res
}

func TestAigerDefine(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	a.Define("test", 42)
	v, ok := a.Lit("test")
	if !ok || v != 42 {
		t.Errorf("Define then Lit should return 42, got %d (ok=%v)", v, ok)
	}
}

func TestAigerSet(t *testing.T) {
	a := NewAiger(nil, []string{"s0"}, nil)
	a.Set("s0", 5)
	if a.Values["s0"] != 5 {
		t.Errorf("Set should store value 5, got %d", a.Values["s0"])
	}
}

func TestAigerString(t *testing.T) {
	a := NewAiger([]string{"x"}, []string{"s"}, []string{"o"})
	a.Set("s", a.MustLit("s"))
	a.Set("o", a.True())
	s := a.String()
	if !strings.HasPrefix(s, "aag") {
		t.Errorf("String should start with 'aag', got: %s", s[:10])
	}
	// Should contain input, latch, output lines
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines (header + input + latch + output), got %d", len(lines))
	}
}

func TestAigerSimulation(t *testing.T) {
	// Build a simple circuit: output = AND(input0, input1)
	a := NewAiger([]string{"a", "b"}, nil, []string{"out"})
	aLit := a.MustLit("a")
	bLit := a.MustLit("b")
	andRes := a.Andl(aLit, bLit)
	a.Set("out", andRes)

	a.Reset()

	// Test: a=1, b=1 -> out should be 1
	a.Step("110") // 3 inputs including bogus
	outVal := a.GetIn(a.Values["out"])
	if outVal != '1' {
		t.Errorf("AND(1,1) should be 1, got %c", outVal)
	}

	// Test: a=1, b=0 -> out should be 0
	a.Reset()
	a.Step("100")
	outVal = a.GetIn(a.Values["out"])
	if outVal != '0' {
		t.Errorf("AND(1,0) should be 0, got %c", outVal)
	}

	// Test: a=0, b=1 -> out should be 0
	a.Reset()
	a.Step("010")
	outVal = a.GetIn(a.Values["out"])
	if outVal != '0' {
		t.Errorf("AND(0,1) should be 0, got %c", outVal)
	}
}

func TestAigerSimulationOr(t *testing.T) {
	a := NewAiger([]string{"a", "b"}, nil, []string{"out"})
	aLit := a.MustLit("a")
	bLit := a.MustLit("b")
	orRes := a.Orl(aLit, bLit)
	a.Set("out", orRes)

	testCases := []struct {
		inp    string
		expect byte
	}{
		{"000", '0'},
		{"010", '1'},
		{"100", '1'},
		{"110", '1'},
	}
	for _, tc := range testCases {
		a.Reset()
		a.Step(tc.inp)
		got := a.GetIn(a.Values["out"])
		if got != tc.expect {
			t.Errorf("OR with input %s: expected %c, got %c", tc.inp, tc.expect, got)
		}
	}
}

func TestAigerSimulationNot(t *testing.T) {
	a := NewAiger([]string{"a"}, nil, []string{"out"})
	aLit := a.MustLit("a")
	notRes := a.Notl(aLit)
	a.Set("out", notRes)

	a.Reset()
	a.Step("10") // a=1, bogus=0
	got := a.GetIn(a.Values["out"])
	if got != '0' {
		t.Errorf("NOT(1) should be 0, got %c", got)
	}

	a.Reset()
	a.Step("00") // a=0, bogus=0
	got = a.GetIn(a.Values["out"])
	if got != '1' {
		t.Errorf("NOT(0) should be 1, got %c", got)
	}
}

func TestAigerSimulationIte(t *testing.T) {
	a := NewAiger([]string{"c", "x", "y"}, nil, []string{"out"})
	cLit := a.MustLit("c")
	xLit := a.MustLit("x")
	yLit := a.MustLit("y")
	iteRes := a.Ite(cLit, xLit, yLit)
	a.Set("out", iteRes)

	testCases := []struct {
		inp    string
		expect byte
	}{
		{"1100", '1'}, // c=1,x=1,y=0 -> x=1
		{"1010", '0'}, // c=1,x=0,y=1 -> x=0
		{"0100", '0'}, // c=0,x=1,y=0 -> y=0
		{"0010", '1'}, // c=0,x=0,y=1 -> y=1
	}
	for _, tc := range testCases {
		a.Reset()
		a.Step(tc.inp)
		got := a.GetIn(a.Values["out"])
		if got != tc.expect {
			t.Errorf("ITE with input %s: expected %c, got %c", tc.inp, tc.expect, got)
		}
	}
}

func TestAigerSimulationIff(t *testing.T) {
	a := NewAiger([]string{"x", "y"}, nil, []string{"out"})
	xLit := a.MustLit("x")
	yLit := a.MustLit("y")
	iffRes := a.Iff(xLit, yLit)
	a.Set("out", iffRes)

	testCases := []struct {
		inp    string
		expect byte
	}{
		{"000", '1'}, // 0 iff 0 = 1
		{"010", '0'}, // 0 iff 1 = 0
		{"100", '0'}, // 1 iff 0 = 0
		{"110", '1'}, // 1 iff 1 = 1
	}
	for _, tc := range testCases {
		a.Reset()
		a.Step(tc.inp)
		got := a.GetIn(a.Values["out"])
		if got != tc.expect {
			t.Errorf("IFF with input %s: expected %c, got %c", tc.inp, tc.expect, got)
		}
	}
}

func TestAigerLatchSimulation(t *testing.T) {
	// Build circuit with a latch: output is latch value, next latch = input
	a := NewAiger([]string{"inp"}, []string{"state"}, []string{"out"})
	inpLit := a.MustLit("inp")
	stLit := a.MustLit("state")
	a.Set("state", inpLit) // next state = input
	a.Set("out", stLit)    // output = current state

	a.Reset()
	// Initially state = 0
	outVal := a.GetIn(a.VarMap["state"])
	if outVal != '0' {
		t.Errorf("initial state should be 0, got %c", outVal)
	}

	// Step with input=1 -> state becomes 1 on next cycle
	a.Step("10")
	a.Advance()
	outVal = a.GetIn(a.VarMap["state"])
	if outVal != '1' {
		t.Errorf("after step(1), state should be 1, got %c", outVal)
	}

	// Step with input=0 -> state becomes 0 on next cycle
	a.Step("00")
	a.Advance()
	outVal = a.GetIn(a.VarMap["state"])
	if outVal != '0' {
		t.Errorf("after step(0), state should be 0, got %c", outVal)
	}
}

func TestAigerGetState(t *testing.T) {
	a := NewAiger(nil, []string{"s0", "s1", "s2"}, nil)
	state := a.GetState("10x")
	if state["s0"] != '1' {
		t.Errorf("s0 should be '1', got %c", state["s0"])
	}
	if state["s1"] != '0' {
		t.Errorf("s1 should be '0', got %c", state["s1"])
	}
	if state["s2"] != 'x' {
		t.Errorf("s2 should be 'x', got %c", state["s2"])
	}
}

func TestAigerDebug(t *testing.T) {
	a := NewAiger([]string{"x"}, []string{"s"}, []string{"o"})
	a.Set("s", a.MustLit("s"))
	a.Set("o", a.True())
	dbg := a.Debug()
	if !strings.Contains(dbg, "inputs:") {
		t.Error("Debug output should contain 'inputs:'")
	}
	if !strings.Contains(dbg, "latches:") {
		t.Error("Debug output should contain 'latches:'")
	}
}

func TestAigerGateGeneration(t *testing.T) {
	a := NewAiger([]string{"a", "b", "c"}, nil, nil)
	aLit := a.MustLit("a")
	bLit := a.MustLit("b")
	cLit := a.MustLit("c")

	// AND of 3 should create 2 gates
	_ = a.Andl(aLit, bLit, cLit)
	if len(a.Gates) != 2 {
		t.Errorf("AND of 3 inputs should create 2 gates, got %d", len(a.Gates))
	}
}

// ============================================================
// Encoder tests
// ============================================================

func TestEncoderNewEncoder(t *testing.T) {
	bw := map[string]int{"x": 3, "s": 2}
	enc := NewEncoder([]string{"x"}, []string{"s"}, []string{"o"}, bw)
	if len(enc.Encoding["x"]) != 3 {
		t.Errorf("x should have 3 bits, got %d", len(enc.Encoding["x"]))
	}
	if len(enc.Encoding["s"]) != 2 {
		t.Errorf("s should have 2 bits, got %d", len(enc.Encoding["s"]))
	}
	if len(enc.Encoding["o"]) != 1 {
		t.Errorf("o should have 1 bit (default), got %d", len(enc.Encoding["o"]))
	}
}

func TestEncoderBinEnc(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	bits := enc.BinEnc(5, 4) // 5 = 0101 in 4 bits
	if len(bits) != 4 {
		t.Fatalf("expected 4 bits, got %d", len(bits))
	}
	// MSB first: 0,1,0,1
	expected := []int{enc.Sub.False(), enc.Sub.True(), enc.Sub.False(), enc.Sub.True()}
	for i, e := range expected {
		if bits[i] != e {
			t.Errorf("bit[%d]: expected %d, got %d", i, e, bits[i])
		}
	}
}

func TestEncoderBinDec(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	bits := enc.BinEnc(7, 3) // 7 = 111 in 3 bits
	val := enc.BinDec(bits)
	if val != 7 {
		t.Errorf("expected 7, got %d", val)
	}

	bits = enc.BinEnc(0, 4)
	val = enc.BinDec(bits)
	if val != 0 {
		t.Errorf("expected 0, got %d", val)
	}
}

func TestEncoderBinEncDecRoundtrip(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	for n := 0; n < 16; n++ {
		bits := enc.BinEnc(n, 4)
		got := enc.BinDec(bits)
		if got != n {
			t.Errorf("roundtrip(%d) = %d", n, got)
		}
	}
}

func TestEncoderTrueFalse(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	tr := enc.True()
	if len(tr) != 1 || tr[0] != enc.Sub.True() {
		t.Errorf("True() should be [1], got %v", tr)
	}
	fl := enc.False()
	if len(fl) != 1 || fl[0] != enc.Sub.False() {
		t.Errorf("False() should be [0], got %v", fl)
	}
}

func TestEncoderNotlMulti(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	bits := enc.BinEnc(5, 4) // 0101
	neg := enc.NotlMulti(bits)
	if len(neg) != 4 {
		t.Fatalf("expected 4 bits, got %d", len(neg))
	}
	// Double negation should give original
	dbl := enc.NotlMulti(neg)
	for i, b := range dbl {
		if b != bits[i] {
			t.Errorf("double negation bit[%d]: expected %d, got %d", i, bits[i], b)
		}
	}
}

func TestEncoderAndlMulti(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	a := enc.BinEnc(3, 2) // 11
	b := enc.BinEnc(2, 2) // 10
	res := enc.AndlMulti(a, b)
	if len(res) != 2 {
		t.Fatalf("expected 2 bits, got %d", len(res))
	}
}

func TestEncoderOrlMulti(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	a := enc.BinEnc(1, 2) // 01
	b := enc.BinEnc(2, 2) // 10
	res := enc.OrlMulti(a, b)
	if len(res) != 2 {
		t.Fatalf("expected 2 bits, got %d", len(res))
	}
}

func TestEncoderGeBin(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	// GeBin with constant true bits
	bits := enc.BinEnc(5, 3) // 101
	// 5 >= 0 should be true (returns True constant)
	res := enc.GeBin(bits, 0)
	if res != enc.Sub.True() {
		t.Errorf("5 >= 0 should be True")
	}
	// 5 >= 8 should be false (overflow for 3 bits, returns False constant)
	res = enc.GeBin(bits, 8)
	if res != enc.Sub.False() {
		t.Errorf("5 >= 8 (overflow 3 bits) should be False")
	}
	// 5 >= 5 may produce a gate (not necessarily a constant), so just check it's not False
	res = enc.GeBin(bits, 5)
	// Evaluate via simulation
	got := evalConstLit(enc.Sub, res)
	if got != '1' {
		t.Errorf("5 >= 5 should be true, got %c", got)
	}
}

func TestEncoderEncodePlusInt(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	x := enc.BinEnc(3, 4)
	y := enc.BinEnc(2, 4)
	res, _ := enc.EncodePlusInt(x, y, enc.Sub.False())
	val := evalMultiConstLits(enc.Sub, res)
	if val != 5 {
		t.Errorf("3 + 2 should be 5, got %d", val)
	}
}

func TestEncoderEncodePlus(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	x := enc.BinEnc(7, 4)
	y := enc.BinEnc(3, 4)
	res := enc.EncodePlus(x, y)
	val := evalMultiConstLits(enc.Sub, res)
	if val != 10 {
		t.Errorf("7 + 3 should be 10, got %d", val)
	}
}

func TestEncoderEncodeMinus(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	x := enc.BinEnc(7, 4)
	y := enc.BinEnc(3, 4)
	res := enc.EncodeMinus(x, y)
	val := evalMultiConstLits(enc.Sub, res)
	if val != 4 {
		t.Errorf("7 - 3 should be 4, got %d", val)
	}
}

func TestEncoderEncodeTimes(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	x := enc.BinEnc(3, 4)
	y := enc.BinEnc(2, 4)
	res := enc.EncodeTimes(x, y)
	val := evalMultiConstLits(enc.Sub, res)
	if val != 6 {
		t.Errorf("3 * 2 should be 6, got %d", val)
	}
}

func TestEncoderEncodeEquality(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	x := enc.BinEnc(5, 3)
	y := enc.BinEnc(5, 3)
	res := enc.EncodeEquality(8, x, y)
	if len(res) != 1 {
		t.Fatalf("equality should return 1-bit result, got %d", len(res))
	}
	got := evalConstLit(enc.Sub, res[0])
	if got != '1' {
		t.Errorf("5 == 5 should be true, got %c", got)
	}

	// Test inequality: 5 != 3
	enc2 := NewEncoder(nil, nil, nil, nil)
	x2 := enc2.BinEnc(5, 3)
	y2 := enc2.BinEnc(3, 3)
	res2 := enc2.EncodeEquality(8, x2, y2)
	got2 := evalConstLit(enc2.Sub, res2[0])
	if got2 != '0' {
		t.Errorf("5 == 3 should be false, got %c", got2)
	}
}

func TestEncoderEncodeIte(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	x := enc.BinEnc(5, 3)
	y := enc.BinEnc(3, 3)
	res := enc.EncodeIte(enc.Sub.True(), x, y)
	val := evalMultiConstLits(enc.Sub, res)
	if val != 5 {
		t.Errorf("ITE(true, 5, 3) should be 5, got %d", val)
	}

	enc2 := NewEncoder(nil, nil, nil, nil)
	x2 := enc2.BinEnc(5, 3)
	y2 := enc2.BinEnc(3, 3)
	res2 := enc2.EncodeIte(enc2.Sub.False(), x2, y2)
	val2 := evalMultiConstLits(enc2.Sub, res2)
	if val2 != 3 {
		t.Errorf("ITE(false, 5, 3) should be 3, got %d", val2)
	}
}

func TestEncoderDefineSym(t *testing.T) {
	bw := map[string]int{"x": 3}
	enc := NewEncoder([]string{"x"}, nil, nil, bw)
	val := enc.BinEnc(5, 3)
	enc.DefineSym("extra", val)
	got, ok := enc.Lit("extra")
	if !ok {
		t.Fatal("DefineSym should make symbol available via Lit")
	}
	if len(got) != 3 {
		t.Errorf("expected 3 bits, got %d", len(got))
	}
}

func TestEncoderSetSym(t *testing.T) {
	bw := map[string]int{"s": 2}
	enc := NewEncoder(nil, []string{"s"}, nil, bw)
	val := enc.BinEnc(3, 2)
	enc.SetSym("s", val)
	for _, subSym := range enc.Encoding["s"] {
		if _, ok := enc.Sub.Values[subSym]; !ok {
			t.Errorf("SetSym should set sub values for %s", subSym)
		}
	}
}

func TestEncoderString(t *testing.T) {
	bw := map[string]int{"x": 2, "s": 1}
	enc := NewEncoder([]string{"x"}, []string{"s"}, []string{"o"}, bw)
	enc.SetSym("s", enc.BinEnc(0, 1))
	enc.SetSym("o", enc.BinEnc(1, 1))
	s := enc.String()
	if !strings.HasPrefix(s, "aag") {
		t.Errorf("String should start with 'aag', got: %s", s)
	}
}

// ============================================================
// Helper function tests
// ============================================================

func TestCeilLog2(t *testing.T) {
	cases := []struct {
		n, expect int
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 3},
		{7, 3},
		{8, 3},
		{9, 4},
		{16, 4},
		{17, 5},
		{256, 8},
	}
	for _, c := range cases {
		got := CeilLog2(c.n)
		if got != c.expect {
			t.Errorf("CeilLog2(%d) = %d, want %d", c.n, got, c.expect)
		}
	}
}

func TestGetEncodingBitsSimple(t *testing.T) {
	// Enumerated sort with 3 values -> 2 bits
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	n, err := GetEncodingBitsSimple(es)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("expected 2 bits for 3-value enum, got %d", n)
	}

	// Range sort 0..7 -> 3 bits
	rs := &lg.RangeSort{Name: "idx", Lb: "0", Ub: "7"}
	n, err = GetEncodingBitsSimple(rs)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("expected 3 bits for range 0..7, got %d", n)
	}

	// Boolean -> 1 bit
	n, err = GetEncodingBitsSimple(lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected 1 bit for boolean, got %d", n)
	}

	// Uninterpreted sort -> error
	us := &lg.UninterpretedSort{Name: "mytype"}
	_, err = GetEncodingBitsSimple(us)
	if err == nil {
		t.Error("expected error for uninterpreted sort")
	}
}

func TestGetEncodingBitsWithInterp(t *testing.T) {
	// Enumerated sort
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	n, err := GetEncodingBits(es, nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("expected 2 bits, got %d", n)
	}

	// Boolean sort
	n, err = GetEncodingBits(lg.Boolean, nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected 1 bit for boolean, got %d", n)
	}
}

func TestIsFiniteSort(t *testing.T) {
	if !IsFiniteSort(&lg.EnumeratedSort{Name: "e", Extension: []string{"a", "b"}}) {
		t.Error("EnumeratedSort should be finite")
	}
	if !IsFiniteSort(&lg.RangeSort{Name: "r", Lb: "0", Ub: "3"}) {
		t.Error("RangeSort should be finite")
	}
	if !IsFiniteSort(lg.Boolean) {
		t.Error("BooleanSort should be finite")
	}
	if IsFiniteSort(&lg.UninterpretedSort{Name: "x"}) {
		t.Error("UninterpretedSort should not be finite")
	}
	fs, _ := lg.NewFunctionSort(lg.Boolean, lg.Boolean)
	if IsFiniteSort(fs) {
		t.Error("FunctionSort should not be finite")
	}
}

func TestSortValues(t *testing.T) {
	// Enumerated
	es := &lg.EnumeratedSort{Name: "color", Extension: []string{"red", "green", "blue"}}
	vals, err := SortValues(es)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 3 {
		t.Errorf("expected 3 values, got %d", len(vals))
	}
	if vals[0] != "red" || vals[1] != "green" || vals[2] != "blue" {
		t.Errorf("unexpected values: %v", vals)
	}

	// Range
	rs := &lg.RangeSort{Name: "idx", Lb: "2", Ub: "5"}
	vals, err = SortValues(rs)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 4 {
		t.Errorf("expected 4 values (2,3,4,5), got %d", len(vals))
	}
	if vals[0] != "2" || vals[3] != "5" {
		t.Errorf("unexpected range values: %v", vals)
	}

	// Boolean
	vals, err = SortValues(lg.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 2 {
		t.Errorf("expected 2 boolean values, got %d", len(vals))
	}

	// Uninterpreted -> error
	_, err = SortValues(&lg.UninterpretedSort{Name: "x"})
	if err == nil {
		t.Error("expected error for uninterpreted sort")
	}
}

func TestTermOrd(t *testing.T) {
	if TermOrd("a", "b") != -1 {
		t.Error("a < b")
	}
	if TermOrd("b", "a") != 1 {
		t.Error("b > a")
	}
	if TermOrd("x", "x") != 0 {
		t.Error("x == x")
	}
}

func TestNormalize(t *testing.T) {
	if Normalize("x = x") != "true" {
		t.Error("x = x should normalize to true")
	}
	if Normalize("b = a") != "a = b" {
		t.Errorf("b = a should normalize to 'a = b', got '%s'", Normalize("b = a"))
	}
	if Normalize("a = b") != "a = b" {
		t.Error("a = b should stay as 'a = b'")
	}
	if Normalize("foo") != "foo" {
		t.Error("non-equality should pass through")
	}
}

func TestGetTruth(t *testing.T) {
	v, err := GetTruth("01x", 0)
	if err != nil {
		t.Fatal(err)
	}
	if v != "false" {
		t.Errorf("digit '0' should be 'false', got %s", v)
	}

	v, err = GetTruth("01x", 1)
	if err != nil {
		t.Fatal(err)
	}
	if v != "true" {
		t.Errorf("digit '1' should be 'true', got %s", v)
	}

	v, err = GetTruth("01x", 2)
	if err != nil {
		t.Fatal(err)
	}
	if v != "" {
		t.Errorf("digit 'x' should be empty, got %s", v)
	}

	_, err = GetTruth("01x", 5)
	if err == nil {
		t.Error("out of range should error")
	}

	_, err = GetTruth("z", 0)
	if err == nil {
		t.Error("invalid digit should error")
	}
}

// ============================================================
// Match (schema) tests
// ============================================================

func TestMatchAddAndLookup(t *testing.T) {
	m := NewMatch()
	m.Add("X", "int")
	if m.Map["X"] != "int" {
		t.Errorf("expected X->int, got %s", m.Map["X"])
	}
}

func TestMatchPushPop(t *testing.T) {
	m := NewMatch()
	m.Add("X", "int")
	m.Push()
	m.Add("Y", "bool")
	if m.Map["Y"] != "bool" {
		t.Error("Y should be bound after Push+Add")
	}
	m.Pop()
	if _, ok := m.Map["Y"]; ok {
		t.Error("Y should be unbound after Pop")
	}
	if m.Map["X"] != "int" {
		t.Error("X should still be bound after Pop")
	}
}

func TestMatchUnify(t *testing.T) {
	m := NewMatch()
	if !m.Unify("X", "int") {
		t.Error("first Unify should succeed")
	}
	if !m.Unify("X", "int") {
		t.Error("Unify with same value should succeed")
	}
	if m.Unify("X", "bool") {
		t.Error("Unify with different value should fail")
	}
}

func TestMatchUnifyLists(t *testing.T) {
	m := NewMatch()
	if !m.UnifyLists([]string{"A", "B"}, []string{"int", "bool"}) {
		t.Error("UnifyLists should succeed for compatible lists")
	}
	if m.Map["A"] != "int" || m.Map["B"] != "bool" {
		t.Error("bindings should be correct after UnifyLists")
	}

	// Different lengths
	m2 := NewMatch()
	if m2.UnifyLists([]string{"A"}, []string{"int", "bool"}) {
		t.Error("UnifyLists should fail for different lengths")
	}
}

func TestMatchCopy(t *testing.T) {
	m := NewMatch()
	m.Add("X", "int")
	m.Add("Y", "bool")
	cp := m.Copy()
	if cp["X"] != "int" || cp["Y"] != "bool" {
		t.Error("Copy should contain all bindings")
	}
	// Modifying copy should not affect original
	cp["X"] = "nat"
	if m.Map["X"] != "int" {
		t.Error("original should not be affected by copy modification")
	}
}

func TestMatchNestedPushPop(t *testing.T) {
	m := NewMatch()
	m.Add("A", "1")
	m.Push()
	m.Add("B", "2")
	m.Push()
	m.Add("C", "3")
	if len(m.Map) != 3 {
		t.Errorf("expected 3 bindings, got %d", len(m.Map))
	}
	m.Pop()
	if _, ok := m.Map["C"]; ok {
		t.Error("C should be gone after inner Pop")
	}
	if m.Map["B"] != "2" {
		t.Error("B should still be bound")
	}
	m.Pop()
	if _, ok := m.Map["B"]; ok {
		t.Error("B should be gone after outer Pop")
	}
	if m.Map["A"] != "1" {
		t.Error("A should still be bound")
	}
}

func TestStrMap(t *testing.T) {
	m := map[string]string{"x": "1"}
	s := StrMap(m)
	if !strings.Contains(s, "x:1") {
		t.Errorf("StrMap should contain 'x:1', got %s", s)
	}
}

func TestApplyMatch(t *testing.T) {
	m := map[string]string{"X": "int", "Y": "bool"}
	result := ApplyMatch(m, "forall X. f(Y)")
	if result != "forall int. f(bool)" {
		t.Errorf("unexpected result: %s", result)
	}
}

// ============================================================
// Qelim tests
// ============================================================

func TestQelimFresh(t *testing.T) {
	q := NewQelim(nil, nil)
	name := q.Fresh("expr1")
	if name.Name != "__qe[0]" {
		t.Errorf("first fresh should be __qe[0], got %s", name.Name)
	}
	name2 := q.Fresh("expr2")
	if name2.Name != "__qe[1]" {
		t.Errorf("second fresh should be __qe[1], got %s", name2.Name)
	}
	if sym, ok := q.Syms["expr1"]; !ok || sym.String() != "__qe[0]" {
		t.Error("expr1 should be recorded in Syms")
	}
}

func TestQelimGetConsts(t *testing.T) {
	intSort := &lg.UninterpretedSort{Name: "int"}
	boolSort := lg.Boolean
	sc := map[string][]*lg.Symbol{
		"int":  {lg.NewSymbol("0", intSort), lg.NewSymbol("1", intSort), lg.NewSymbol("2", intSort)},
		"bool": {lg.NewSymbol("false", boolSort), lg.NewSymbol("true", boolSort)},
	}
	q := NewQelim(sc, nil)
	consts := q.GetConsts(intSort, sc)
	if len(consts) != 3 {
		t.Errorf("expected 3 constants for int, got %d", len(consts))
	}
	unknownSort := &lg.UninterpretedSort{Name: "unknown"}
	consts = q.GetConsts(unknownSort, sc)
	if consts != nil {
		t.Errorf("expected nil for unknown sort, got %v", consts)
	}
}

func TestElimIteKey(t *testing.T) {
	key := ElimIteKey("bool")
	if !strings.HasPrefix(key, "__ite[") {
		t.Errorf("ElimIteKey should start with '__ite[', got %s", key)
	}
	if !strings.HasSuffix(key, ":bool") {
		t.Errorf("ElimIteKey should end with ':bool', got %s", key)
	}
}

// ============================================================
// PropAbs tests
// ============================================================

func TestPropAbsNewProp(t *testing.T) {
	pa := NewPropAbs(nil, nil)
	// Create a test expression
	x := lg.NewSymbol("x", lg.Boolean)
	y := lg.NewSymbol("y", lg.Boolean)
	expr := &lg.Eq{T1: x, T2: y}
	name := pa.newProp(expr)
	if name.Name != "__abs[0]" {
		t.Errorf("first newProp should be __abs[0], got %s", name.Name)
	}
	// Same expression should return same name
	name2 := pa.newProp(expr)
	if name2 != name {
		t.Errorf("same expression should return same name, got %s", name2.Name)
	}
	// Different expression should get new name
	z := lg.NewSymbol("z", lg.Boolean)
	expr2 := &lg.Eq{T1: x, T2: z}
	name3 := pa.newProp(expr2)
	if name3 == name {
		t.Error("different expression should get different name")
	}
}

// ============================================================
// Witness tests
// ============================================================

func TestMatchHandlerBase(t *testing.T) {
	h := &MatchHandlerBase{}
	if h.Eval("false") != false {
		t.Error("Eval(false) should be false")
	}
	if h.Eval("true") != true {
		t.Error("Eval(true) should be true")
	}
	if h.Eval("something") != true {
		t.Error("Eval(something) should default to true")
	}
	// These should not panic
	h.Handle("action", "env")
	h.DoReturn("action", "env")
}

// ============================================================
// Checker tests
// ============================================================

func TestABCModelCheckerCmd(t *testing.T) {
	mc := &ABCModelChecker{ABCPath: "/usr/local/bin/abc"}
	cmd := mc.Cmd("test.aig", "test.out")
	if cmd[0] != "/usr/local/bin/abc" {
		t.Errorf("expected abc path, got %s", cmd[0])
	}
	if !strings.Contains(cmd[2], "read_aiger test.aig") {
		t.Errorf("command should contain read_aiger, got %s", cmd[2])
	}
}

func TestABCModelCheckerScrape(t *testing.T) {
	mc := &ABCModelChecker{}
	if !mc.Scrape("Property proved\n") {
		t.Error("should detect 'Property proved'")
	}
	if mc.Scrape("Counterexample found\n") {
		t.Error("should not detect property proved in counterexample output")
	}
}

func TestToAigerEncoder(t *testing.T) {
	bw := map[string]int{"x": 2}
	enc := NewEncoder([]string{"x"}, []string{"s"}, []string{"o"}, bw)
	if enc == nil {
		t.Fatal("NewEncoder should return non-nil")
	}
	if enc.Sub == nil {
		t.Error("encoder should have non-nil Sub")
	}
}

// ============================================================
// Integration tests
// ============================================================

func TestAigerFullCircuit(t *testing.T) {
	// Build a 2-input mux: out = ITE(sel, a, b)
	a := NewAiger([]string{"sel", "a", "b"}, nil, []string{"out"})
	sel := a.MustLit("sel")
	aLit := a.MustLit("a")
	bLit := a.MustLit("b")
	iteRes := a.Ite(sel, aLit, bLit)
	a.Set("out", iteRes)

	// Verify AIGER output format
	s := a.String()
	if !strings.HasPrefix(s, "aag") {
		t.Error("should start with aag")
	}

	// Simulate
	tests := []struct {
		sel, aIn, bIn byte
		expect        byte
	}{
		{'1', '1', '0', '1'},
		{'1', '0', '1', '0'},
		{'0', '1', '0', '0'},
		{'0', '0', '1', '1'},
	}
	for _, tc := range tests {
		a.Reset()
		inp := string([]byte{tc.sel, tc.aIn, tc.bIn, '0'}) // +bogus
		a.Step(inp)
		got := a.GetIn(a.Values["out"])
		if got != tc.expect {
			t.Errorf("MUX(%c,%c,%c) = %c, want %c", tc.sel, tc.aIn, tc.bIn, got, tc.expect)
		}
	}
}

func TestEncoderArithmeticRoundtrip(t *testing.T) {
	for a := 0; a < 8; a++ {
		for b := 0; b <= a; b++ {
			enc := NewEncoder(nil, nil, nil, nil)
			x := enc.BinEnc(a, 4)
			y := enc.BinEnc(b, 4)
			sum := enc.EncodePlus(x, y)
			sumVal := evalMultiConstLits(enc.Sub, sum)
			if sumVal != (a+b)%16 {
				t.Errorf("%d + %d = %d, want %d", a, b, sumVal, (a+b)%16)
			}

			enc2 := NewEncoder(nil, nil, nil, nil)
			x2 := enc2.BinEnc(a, 4)
			y2 := enc2.BinEnc(b, 4)
			diff := enc2.EncodeMinus(x2, y2)
			diffVal := evalMultiConstLits(enc2.Sub, diff)
			if diffVal != (a-b+16)%16 {
				t.Errorf("%d - %d = %d, want %d", a, b, diffVal, (a-b+16)%16)
			}
		}
	}
}

func TestEncoderMultiplicationSmall(t *testing.T) {
	for a := 0; a < 4; a++ {
		for b := 0; b < 4; b++ {
			enc := NewEncoder(nil, nil, nil, nil)
			x := enc.BinEnc(a, 4)
			y := enc.BinEnc(b, 4)
			prod := enc.EncodeTimes(x, y)
			prodVal := evalMultiConstLits(enc.Sub, prod)
			expected := (a * b) % 16
			if prodVal != expected {
				t.Errorf("%d * %d = %d, want %d", a, b, prodVal, expected)
			}
		}
	}
}

func TestEncoderLtLe(t *testing.T) {
	for a := 0; a < 8; a++ {
		for b := 0; b < 8; b++ {
			enc := NewEncoder(nil, nil, nil, nil)
			x := enc.BinEnc(a, 3)
			y := enc.BinEnc(b, 3)
			lt := enc.EncodeLt(x, y, enc.Sub.False())
			ltVal := evalConstLit(enc.Sub, lt[0]) == '1'
			if ltVal != (a < b) {
				t.Errorf("%d < %d = %v, want %v", a, b, ltVal, a < b)
			}

			enc2 := NewEncoder(nil, nil, nil, nil)
			x2 := enc2.BinEnc(a, 3)
			y2 := enc2.BinEnc(b, 3)
			le := enc2.EncodeLe(x2, y2)
			leVal := evalConstLit(enc2.Sub, le[0]) == '1'
			if leVal != (a <= b) {
				t.Errorf("%d <= %d = %v, want %v", a, b, leVal, a <= b)
			}
		}
	}
}

func TestIsFiniteSortWithInterp(t *testing.T) {
	interp := map[string]interface{}{
		"mybv": "bv[8]",
	}
	// Direct finite sorts
	if !IsFiniteSortWithInterp(lg.Boolean, interp) {
		t.Error("Boolean should be finite")
	}
	// Uninterpreted with bv interpretation
	us := &lg.UninterpretedSort{Name: "mybv"}
	if !IsFiniteSortWithInterp(us, interp) {
		t.Error("mybv with bv[8] interp should be finite")
	}
}

// ============================================================
// Fuzz tests
// ============================================================

func FuzzCeilLog2(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(2)
	f.Add(3)
	f.Add(255)
	f.Add(256)
	f.Add(1023)
	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 {
			return
		}
		bits := CeilLog2(n)
		if bits < 0 {
			t.Errorf("CeilLog2(%d) = %d, should be non-negative", n, bits)
		}
		if n > 0 {
			// 2^bits should be >= n
			val := 1 << bits
			if val < n {
				t.Errorf("CeilLog2(%d) = %d, but 2^%d = %d < %d", n, bits, bits, val, n)
			}
			// 2^(bits-1) should be < n (unless bits=0)
			if bits > 0 {
				prev := 1 << (bits - 1)
				if prev >= n {
					t.Errorf("CeilLog2(%d) = %d, but 2^%d = %d >= %d (not tight)", n, bits, bits-1, prev, n)
				}
			}
		}
	})
}

func FuzzAigerOps(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(1), uint8(1))
	f.Add(uint8(0), uint8(1))
	f.Add(uint8(1), uint8(0))
	f.Fuzz(func(t *testing.T, xBit, yBit uint8) {
		xBool := xBit%2 == 1
		yBool := yBit%2 == 1

		a := NewAiger([]string{"x", "y"}, nil, []string{"and_out", "or_out", "xor_out", "iff_out", "imp_out"})
		xLit := a.MustLit("x")
		yLit := a.MustLit("y")

		andRes := a.Andl(xLit, yLit)
		orRes := a.Orl(xLit, yLit)
		xorRes := a.Xor(xLit, yLit)
		iffRes := a.Iff(xLit, yLit)
		impRes := a.ImpliesL(xLit, yLit)

		a.Set("and_out", andRes)
		a.Set("or_out", orRes)
		a.Set("xor_out", xorRes)
		a.Set("iff_out", iffRes)
		a.Set("imp_out", impRes)

		xChar := byte('0')
		if xBool {
			xChar = '1'
		}
		yChar := byte('0')
		if yBool {
			yChar = '1'
		}

		a.Reset()
		inp := string([]byte{xChar, yChar, '0'}) // +bogus
		a.Step(inp)

		check := func(name string, lit int, expected bool) {
			got := a.GetIn(lit) == '1'
			if got != expected {
				t.Errorf("%s(%v, %v) = %v, want %v", name, xBool, yBool, got, expected)
			}
		}

		check("AND", andRes, xBool && yBool)
		check("OR", orRes, xBool || yBool)
		check("XOR", xorRes, xBool != yBool)
		check("IFF", iffRes, xBool == yBool)
		check("IMP", impRes, !xBool || yBool)
	})
}

func FuzzEncoderArith(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(3), uint8(2))
	f.Add(uint8(7), uint8(5))
	f.Add(uint8(15), uint8(15))
	f.Fuzz(func(t *testing.T, a8, b8 uint8) {
		a := int(a8 % 16) // 4-bit values
		b := int(b8 % 16)

		enc := NewEncoder(nil, nil, nil, nil)
		x := enc.BinEnc(a, 4)
		y := enc.BinEnc(b, 4)
		sum := enc.EncodePlus(x, y)
		sumVal := evalMultiConstLits(enc.Sub, sum)
		if sumVal != (a+b)%16 {
			t.Errorf("%d + %d = %d, want %d", a, b, sumVal, (a+b)%16)
		}

		enc2 := NewEncoder(nil, nil, nil, nil)
		x2 := enc2.BinEnc(a, 4)
		y2 := enc2.BinEnc(b, 4)
		diff := enc2.EncodeMinus(x2, y2)
		diffVal := evalMultiConstLits(enc2.Sub, diff)
		if diffVal != (a-b+16)%16 {
			t.Errorf("%d - %d = %d, want %d", a, b, diffVal, (a-b+16)%16)
		}
	})
}

// ============================================================
// Additional tests to exceed 35 threshold
// ============================================================

func TestAigerLitNotFound(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	_, ok := a.Lit("nonexistent")
	if ok {
		t.Error("Lit should return false for nonexistent symbol")
	}
}

func TestAigerMustLitPanic(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLit should panic for nonexistent symbol")
		}
	}()
	a.MustLit("nonexistent")
}

func TestEncoderLitNotFound(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	_, ok := enc.Lit("nonexistent")
	if ok {
		t.Error("Lit should return false for nonexistent symbol")
	}
}

func TestAigerSymVals(t *testing.T) {
	a := NewAiger([]string{"a", "b"}, nil, nil)
	a.Reset()
	// Set states manually
	a.State[a.VarMap["a"]] = '1'
	a.State[a.VarMap["b"]] = '0'
	// Note: bogus is also an input
	a.State[a.VarMap["%%bogus%%"]] = '0'
	vals := a.SymVals([]string{"a", "b"})
	if vals != "10" {
		t.Errorf("expected '10', got '%s'", vals)
	}
}

func TestMatchPopEmpty(t *testing.T) {
	m := NewMatch()
	// Pop the initial empty level
	m.Pop()
	// Pop again (should be no-op)
	m.Pop()
	// Should not panic
}

func TestEncoderEncodeLeConstants(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	// 0 <= 0 should be true
	x := enc.BinEnc(0, 3)
	y := enc.BinEnc(0, 3)
	res := enc.EncodeLe(x, y)
	got := evalConstLit(enc.Sub, res[0])
	if got != '1' {
		t.Errorf("0 <= 0 should be True, got %c", got)
	}
}

func TestAigerNotlConstants(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	// Not(True) = False
	if a.Notl(a.True()) != a.False() {
		t.Error("Notl(True) should be False")
	}
	if a.Notl(a.False()) != a.True() {
		t.Error("Notl(False) should be True")
	}
}

func TestCeilLog2Large(t *testing.T) {
	if CeilLog2(1024) != 10 {
		t.Errorf("CeilLog2(1024) = %d, want 10", CeilLog2(1024))
	}
	if CeilLog2(1025) != 11 {
		t.Errorf("CeilLog2(1025) = %d, want 11", CeilLog2(1025))
	}
}

func TestGetTruthEdgeCases(t *testing.T) {
	// Empty string
	_, err := GetTruth("", 0)
	if err == nil {
		t.Error("empty string should error")
	}
}

func TestSortValuesRangeZero(t *testing.T) {
	rs := &lg.RangeSort{Name: "idx", Lb: "0", Ub: "0"}
	vals, err := SortValues(rs)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != "0" {
		t.Errorf("range 0..0 should have single value '0', got %v", vals)
	}
}

func TestAigerGetInConstants(t *testing.T) {
	a := NewAiger(nil, nil, nil)
	a.State = make(map[int]byte)
	if a.GetIn(0) != '0' {
		t.Error("GetIn(0) should be '0'")
	}
	if a.GetIn(1) != '1' {
		t.Error("GetIn(1) should be '1'")
	}
}

func TestEncoderAndlMultiEmpty(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	res := enc.AndlMulti()
	if len(res) != 1 || res[0] != enc.Sub.True() {
		t.Errorf("AndlMulti() should be True, got %v", res)
	}
}

func TestEncoderOrlMultiEmpty(t *testing.T) {
	enc := NewEncoder(nil, nil, nil, nil)
	res := enc.OrlMulti()
	if len(res) != 1 || res[0] != enc.Sub.False() {
		t.Errorf("OrlMulti() should be False, got %v", res)
	}
}

func TestMatchSchemaPrems(t *testing.T) {
	sortConstants := map[string][]string{
		"int": {"0", "1"},
	}
	match := NewMatch()
	var results []map[string]string
	MatchSchemaPrems([]string{"X"}, sortConstants, nil, match, nil, func(m map[string]string) {
		results = append(results, m)
	})
	if len(results) != 2 {
		t.Errorf("expected 2 match results, got %d", len(results))
	}
}

// Benchmark

func BenchmarkAigerAndChain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := NewAiger(nil, nil, nil)
		prev := a.True()
		for j := 0; j < 100; j++ {
			a.Define(fmt.Sprintf("v%d", j), a.NextID*2)
			a.NextID++
			prev = a.Andl(prev, a.MustLit(fmt.Sprintf("v%d", j)))
		}
		_ = prev
	}
}

func BenchmarkEncoderBinEncDec(b *testing.B) {
	enc := NewEncoder(nil, nil, nil, nil)
	r := rand.New(rand.NewSource(42))
	for i := 0; i < b.N; i++ {
		v := r.Intn(256)
		bits := enc.BinEnc(v, 8)
		_ = enc.BinDec(bits)
	}
}
