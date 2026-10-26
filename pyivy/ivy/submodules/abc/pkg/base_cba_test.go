package abc

// Standalone Go tests for the CBA port.

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
)

func buildCBATestNetwork(t *testing.T, limits CBALimits) *CBANetwork {
	t.Helper()
	n, err := NewCBANetwork("top", limits)
	if err != nil {
		t.Fatal(err)
	}
	a, err := n.AddInput("a", 4, false)
	if err != nil {
		t.Fatal(err)
	}
	b, err := n.AddInput("b", 4, false)
	if err != nil {
		t.Fatal(err)
	}
	s, err := n.AddInput("s", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := n.AddNode("sum", CBABoxAdd, 4, false, []int{a, b})
	if err != nil {
		t.Fatal(err)
	}
	xor, err := n.AddNode("xorv", CBABoxXor, 4, false, []int{a, b})
	if err != nil {
		t.Fatal(err)
	}
	mux, err := n.AddNode("choice", CBABoxMux, 4, false, []int{s, sum, xor})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = n.AddOutput("y", 4, mux); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestCBAInventoryAndTypes(t *testing.T) {
	want := []string{"cbaBlast.c", "cbaCba.c", "cbaCom.c", "cbaNtk.c", "cbaReadBlif.c", "cbaReadVer.c", "cbaWriteBlif.c", "cbaWriteVer.c"}
	if len(CBABuildInventory) != len(want) {
		t.Fatalf("inventory length %d", len(CBABuildInventory))
	}
	for i := range want {
		if CBABuildInventory[i] != want[i] {
			t.Fatalf("inventory[%d]=%q", i, CBABuildInventory[i])
		}
	}
	if int(CBAObjectTypeCount) != 90 || len(CBATypeNames) != 90 {
		t.Fatalf("type count %d/%d", CBAObjectTypeCount, len(CBATypeNames))
	}
	if CBABoxSlice != 88 || CBABoxConcat != 89 {
		t.Fatalf("terminal type values %d %d", CBABoxSlice, CBABoxConcat)
	}
	for i, name := range CBATypeNames {
		if name == "" {
			t.Fatalf("empty type name %d", i)
		}
	}
	for _, tc := range []struct {
		r CBARange
		w int
	}{{CBARange{7, 0}, 8}, {CBARange{0, 7}, 8}, {CBARange{-2, -5}, 4}} {
		w, err := tc.r.Width()
		if err != nil || w != tc.w {
			t.Fatalf("range %+v: %d %v", tc.r, w, err)
		}
	}
	if _, err := (CBARange{Left: math.MaxInt, Right: math.MinInt}).Width(); !errors.Is(err, ErrLimit) {
		t.Fatalf("overflowing range: %v", err)
	}
}

func TestCBAParserDirectOutputsConstantsAndOwnership(t *testing.T) {
	verilog := "module direct(a, b, y); input a; input b; output y; assign y = (a & b); endmodule"
	d, err := CBAParseVerilog(verilog, CBALimits{})
	if err != nil {
		t.Fatalf("direct Verilog output: %v", err)
	}
	n, _ := d.Root()
	for m := 0; m < 4; m++ {
		got, e := n.Evaluate(map[string]uint64{"a": uint64(m & 1), "b": uint64(m >> 1)})
		want := uint64(0)
		if m&1 != 0 && m&2 != 0 {
			want = 1
		}
		if e != nil || got["y"] != want {
			t.Fatalf("direct output m=%d got=%v err=%v", m, got, e)
		}
	}
	for _, text := range []string{
		".model one\n.outputs y\n.names y\n1\n.end\n",
		".model zero\n.outputs y\n.names y\n.end\n",
		".model and\n.inputs a b\n.outputs y\n.names a b y\n11 1\n.end\n",
	} {
		parsed, e := CBAParseBLIF(text, CBALimits{})
		if e != nil {
			t.Fatalf("direct BLIF output: %v\n%s", e, text)
		}
		root, _ := parsed.Root()
		if e = root.Validate(); e != nil {
			t.Fatal(e)
		}
	}
	if _, err = CBAParseBLIF(".model bad\n.outputs y\n.names y\n2\n.end\n", CBALimits{}); err == nil {
		t.Fatal("accepted malformed zero-input cube")
	}
	owned, _ := NewCBANetwork("owned", CBALimits{})
	_, _ = owned.AddInput("a", 1, false)
	d1, _ := NewCBADesign("d1", CBALimits{})
	d2, _ := NewCBADesign("d2", CBALimits{})
	if _, err = d1.AddNetwork(owned); err != nil {
		t.Fatal(err)
	}
	if _, err = d2.AddNetwork(owned); !errors.Is(err, ErrInvariant) {
		t.Fatalf("shared owned network: %v", err)
	}
}

func TestCBAValidationRejectsMutatedShapes(t *testing.T) {
	n := buildCBATestNetwork(t, CBALimits{})
	n.objects[n.outputs[0]].Inputs = nil
	if err := n.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("malformed output: %v", err)
	}
	n = buildCBATestNetwork(t, CBALimits{})
	n.objects[4].Width = 1
	if err := n.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("malformed node width: %v", err)
	}
}

func TestCBANetworkExhaustiveEvaluation(t *testing.T) {
	n := buildCBATestNetwork(t, CBALimits{})
	if err := n.Validate(); err != nil {
		t.Fatal(err)
	}
	for a := uint64(0); a < 16; a++ {
		for b := uint64(0); b < 16; b++ {
			for s := uint64(0); s < 2; s++ {
				got, err := n.Evaluate(map[string]uint64{"a": a, "b": b, "s": s})
				if err != nil {
					t.Fatal(err)
				}
				want := (a ^ b) & 15
				if s != 0 {
					want = (a + b) & 15
				}
				if got["y"] != want {
					t.Fatalf("a=%d b=%d s=%d got=%d want=%d", a, b, s, got["y"], want)
				}
			}
		}
	}
	stats, err := n.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.Objects != 7 || stats.Inputs != 3 || stats.Outputs != 1 || stats.Edges != 8 || stats.Bits != 25 {
		t.Fatalf("stats %+v", stats)
	}
	objects, err := n.Objects()
	if err != nil {
		t.Fatal(err)
	}
	objects[0].Name = "changed"
	if obj, _ := n.Object(1); obj.Name != "a" {
		t.Fatal("object snapshot aliases manager")
	}
}

func TestCBAPrimitiveIndependentModel(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 { state ^= state << 7; state ^= state >> 9; state ^= state << 8; return state }
	types := []CBAObjectType{CBABoxAnd, CBABoxNand, CBABoxOr, CBABoxNor, CBABoxXor, CBABoxXnor, CBABoxAdd, CBABoxSub, CBABoxMul, CBABoxDiv, CBABoxMod, CBABoxLess, CBABoxEqual, CBABoxNotEqual, CBABoxShiftLeft, CBABoxShiftRight, CBABoxRotateLeft, CBABoxRotateRight}
	for i := 0; i < 5000; i++ {
		w := 1 + int(next()%16)
		mask := cbaMask(w)
		a, b := next()&mask, next()&mask
		typ := types[next()%uint64(len(types))]
		got, err := CBAEvaluatePrimitive(typ, w, false, []uint64{a, b}, w, 0)
		if err != nil {
			t.Fatalf("%d %v", i, err)
		}
		var want uint64
		switch typ {
		case CBABoxAnd:
			want = a & b
		case CBABoxNand:
			want = ^(a & b)
		case CBABoxOr:
			want = a | b
		case CBABoxNor:
			want = ^(a | b)
		case CBABoxXor:
			want = a ^ b
		case CBABoxXnor:
			want = ^(a ^ b)
		case CBABoxAdd:
			want = a + b
		case CBABoxSub:
			want = a - b
		case CBABoxMul:
			want = a * b
		case CBABoxDiv:
			if b != 0 {
				want = a / b
			}
		case CBABoxMod:
			if b != 0 {
				want = a % b
			}
		case CBABoxLess:
			want = cbaBoolWord(a < b)
		case CBABoxEqual:
			want = cbaBoolWord(a == b)
		case CBABoxNotEqual:
			want = cbaBoolWord(a != b)
		case CBABoxShiftLeft:
			if b < uint64(w) {
				want = a << uint(b)
			}
		case CBABoxShiftRight:
			if b < uint64(w) {
				want = a >> uint(b)
			}
		case CBABoxRotateLeft:
			want = cbaRotateLeft(a, int(b%uint64(w)), w)
		case CBABoxRotateRight:
			s := int(b % uint64(w))
			want = cbaRotateLeft(a, (w-s)%w, w)
		}
		if got != want&mask {
			t.Fatalf("iter=%d type=%v width=%d a=%x b=%x got=%x want=%x", i, typ, w, a, b, got, want&mask)
		}
	}
}

func TestCBAVerilogRoundTrip(t *testing.T) {
	n := buildCBATestNetwork(t, CBALimits{})
	d, err := NewCBADesign("d", CBALimits{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = d.AddNetwork(n); err != nil {
		t.Fatal(err)
	}
	text, err := CBAWriteVerilog(d, CBALimits{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "module top") || !strings.Contains(text, "assign choice = (s ? sum : xorv);") {
		t.Fatalf("unexpected text:\n%s", text)
	}
	parsed, err := CBAParseVerilog(text, CBALimits{})
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, text)
	}
	p, err := parsed.Root()
	if err != nil {
		t.Fatal(err)
	}
	for a := uint64(0); a < 16; a++ {
		for b := uint64(0); b < 16; b++ {
			for s := uint64(0); s < 2; s++ {
				in := map[string]uint64{"a": a, "b": b, "s": s}
				x, e1 := n.Evaluate(in)
				y, e2 := p.Evaluate(in)
				if e1 != nil || e2 != nil || x["y"] != y["y"] {
					t.Fatalf("roundtrip %v %v %v %v", x, y, e1, e2)
				}
			}
		}
	}
}

func TestCBAVerilogNegatedBinaryRoundTrip(t *testing.T) {
	n, _ := NewCBANetwork("neg", CBALimits{})
	a, _ := n.AddInput("a", 1, false)
	b, _ := n.AddInput("b", 1, false)
	for i, typ := range []CBAObjectType{CBABoxNand, CBABoxNor, CBABoxXnor} {
		id, err := n.AddNode(fmt.Sprintf("n%d", i), typ, 1, false, []int{a, b})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = n.AddOutput(fmt.Sprintf("y%d", i), 1, id); err != nil {
			t.Fatal(err)
		}
	}
	d, _ := NewCBADesign("neg", CBALimits{})
	_, _ = d.AddNetwork(n)
	text, err := CBAWriteVerilog(d, CBALimits{})
	if err != nil {
		t.Fatal(err)
	}
	pd, err := CBAParseVerilog(text, CBALimits{})
	if err != nil {
		t.Fatalf("%v\n%s", err, text)
	}
	p, _ := pd.Root()
	for m := 0; m < 4; m++ {
		in := map[string]uint64{"a": uint64(m & 1), "b": uint64(m >> 1)}
		x, _ := n.Evaluate(in)
		y, e := p.Evaluate(in)
		if e != nil || fmt.Sprint(x) != fmt.Sprint(y) {
			t.Fatalf("m=%d %v %v %v", m, x, y, e)
		}
	}
}

func buildCBABoolean(t *testing.T) *CBADesign {
	t.Helper()
	n, err := NewCBANetwork("logic", CBALimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddInput("a", 1, false)
	b, _ := n.AddInput("b", 1, false)
	x, _ := n.AddNode("x", CBABoxXor, 1, false, []int{a, b})
	m, _ := n.AddNode("m", CBABoxAnd, 1, false, []int{a, b})
	y, _ := n.AddNode("z", CBABoxOr, 1, false, []int{x, m})
	if _, err = n.AddOutput("y", 1, y); err != nil {
		t.Fatal(err)
	}
	d, _ := NewCBADesign("logic", CBALimits{})
	if _, err = d.AddNetwork(n); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestCBABLIFRoundTrip(t *testing.T) {
	d := buildCBABoolean(t)
	text, err := CBAWriteBLIF(d, CBALimits{})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := CBAParseBLIF(text, CBALimits{})
	if err != nil {
		t.Fatalf("parse %v\n%s", err, text)
	}
	a, _ := d.Root()
	b, _ := parsed.Root()
	for v := 0; v < 4; v++ {
		in := map[string]uint64{"a": uint64(v & 1), "b": uint64(v >> 1)}
		x, e1 := a.Evaluate(in)
		y, e2 := b.Evaluate(in)
		if e1 != nil || e2 != nil || x["y"] != y["y"] {
			t.Fatalf("v=%d x=%v y=%v e=%v/%v\n%s", v, x, y, e1, e2, text)
		}
	}
	if _, err = CBAParseBLIF(".model x\n.inputs a\n.outputs y\n.names a y\n2 1\n.end\n", CBALimits{}); err == nil {
		t.Fatal("accepted malformed cube")
	}
}

func TestCBALimitsCancellationAndAtomicity(t *testing.T) {
	if _, err := NewCBANetwork("bad-name", CBALimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("name err %v", err)
	}
	n, err := NewCBANetwork("n", CBALimits{MaxObjects: 2, MaxEdges: 2, MaxBits: 4, MaxBytes: 1000})
	if err != nil {
		t.Fatal(err)
	}
	a, err := n.AddInput("a", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = n.AddNode("b", CBABoxBuf, 1, false, []int{a}); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit err %v", err)
	}
	if len(n.objects) != 2 || len(n.names) != 1 {
		t.Fatal("failed add mutated network")
	}
	if _, err = n.AddNode("bad", CBABoxAnd, 1, false, []int{a, 99}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("bad edge %v", err)
	}
	calls := 0
	enabled := false
	n2 := buildCBATestNetwork(t, CBALimits{Stop: func() bool {
		if !enabled {
			return false
		}
		calls++
		return calls > 3
	}})
	enabled = true
	if _, err = n2.Evaluate(map[string]uint64{"a": 1, "b": 2, "s": 1}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel err %v", err)
	}
	if _, err := CBAParseVerilog("module x(a); input a; endmodule", CBALimits{MaxTextBytes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatalf("text limit %v", err)
	}
	if _, err := CBAParseBLIF(".model x\n.inputs a\n.outputs y\n.end\n", CBALimits{}); err == nil {
		t.Fatal("accepted undriven output")
	}
	n3 := buildCBATestNetwork(t, CBALimits{})
	n3.objects[4].Inputs[0] = 999
	if err = n3.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("corrupt graph %v", err)
	}
	n4 := buildCBATestNetwork(t, CBALimits{})
	n4.Close()
	if _, err = n4.Evaluate(nil); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed err %v", err)
	}
	n5 := buildCBATestNetwork(t, CBALimits{})
	n5.names["a"] = 2
	if err = n5.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("corrupt names %v", err)
	}
	if tiny, e := NewCBANetwork("x", CBALimits{MaxBytes: 100}); e != nil {
		t.Fatal(e)
	} else if _, e = tiny.AddInput("a", 1, false); !errors.Is(e, ErrLimit) {
		t.Fatalf("retained bytes %v", e)
	}
}

func TestCBAParserRejections(t *testing.T) {
	cases := []string{
		"module x(a); input a; output y; assign y = missing; endmodule",
		"module x(a); input a; input [1:0] a; endmodule",
		"module x(a); input a; output y; assign y = a && a; endmodule",
		"module x(a); input a; output y; assign y = a;",
		"garbage module x(a); input a; endmodule",
		"module x(a); input a; endmodule garbage",
	}
	for _, text := range cases {
		if d, err := CBAParseVerilog(text, CBALimits{}); err == nil || d != nil {
			t.Fatalf("accepted %q: %v", text, err)
		}
	}
	if d, err := CBAParseVerilog("module x(a); input a; output y; assign y = a; endmodule", CBALimits{MaxOperations: 1}); !errors.Is(err, ErrLimit) || d != nil {
		t.Fatalf("op limit %v %v", d, err)
	}
	if d, err := CBAParseBLIF(".model x\n.inputs a\n.outputs y\n.names a y\n1 1\n.end\n", CBALimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || d != nil {
		t.Fatalf("cancel %v %v", d, err)
	}
}

func TestCBASliceConcatSigned(t *testing.T) {
	n, err := NewCBANetwork("bits", CBALimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddInput("a", 8, false)
	hi, _ := n.AddSlice("hi", a, 7, 4)
	lo, _ := n.AddSlice("lo", a, 3, 0)
	cat, _ := n.AddNode("cat", CBABoxConcat, 8, false, []int{hi, lo})
	if _, err = n.AddOutput("y", 8, cat); err != nil {
		t.Fatal(err)
	}
	for v := uint64(0); v < 256; v++ {
		out, e := n.Evaluate(map[string]uint64{"a": v})
		if e != nil || out["y"] != v {
			t.Fatalf("slice concat %d %v %v", v, out, e)
		}
	}
	got, err := CBAEvaluatePrimitive(CBABoxSignedLess, 1, false, []uint64{15, 1}, 4, 0)
	if err != nil || got != 1 {
		t.Fatalf("signed less %d %v", got, err)
	}
	if got, err = CBAEvaluatePrimitive(CBABoxMin, 4, true, []uint64{3}, 4, 0); err != nil || got != 13 {
		t.Fatalf("minus %d %v", got, err)
	}
	if got, err = CBAEvaluatePrimitive(CBABoxSqrt, 4, false, []uint64{15}, 4, 0); err != nil || got != 3 {
		t.Fatalf("sqrt %d %v", got, err)
	}
	if got, err = CBAEvaluatePrimitive(CBABoxDiv, 4, true, []uint64{14, 2}, 4, 0); err != nil || got != 15 {
		t.Fatalf("signed div %d %v", got, err)
	}
}

func FuzzCBAStructuralRoundTrip(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x55, 0xaa, 0xffff, 0x123456789abcdef0} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		state := seed | 1
		next := func() uint64 { state = state*6364136223846793005 + 1442695040888963407; return state }
		n, err := NewCBANetwork("f", CBALimits{MaxObjects: 64, MaxEdges: 128, MaxBits: 256, MaxOperations: 4096, MaxTextBytes: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := n.AddInput("a", 4, false)
		b, _ := n.AddInput("b", 4, false)
		ids := []int{a, b}
		ops := []CBAObjectType{CBABoxAnd, CBABoxOr, CBABoxXor, CBABoxAdd, CBABoxSub}
		count := 1 + int(next()%12)
		for i := 0; i < count; i++ {
			x := ids[next()%uint64(len(ids))]
			y := ids[next()%uint64(len(ids))]
			id, e := n.AddNode(fmt.Sprintf("n%d", i), ops[next()%uint64(len(ops))], 4, false, []int{x, y})
			if e != nil {
				t.Fatal(e)
			}
			ids = append(ids, id)
		}
		if _, err = n.AddOutput("y", 4, ids[len(ids)-1]); err != nil {
			t.Fatal(err)
		}
		d, _ := NewCBADesign("f", n.limits)
		if _, err = d.AddNetwork(n); err != nil {
			t.Fatal(err)
		}
		text, err := CBAWriteVerilog(d, n.limits)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := CBAParseVerilog(text, n.limits)
		if err != nil {
			t.Fatalf("parse %v\n%s", err, text)
		}
		p, _ := parsed.Root()
		for k := 0; k < 16; k++ {
			in := map[string]uint64{"a": next() & 15, "b": next() & 15}
			x, e1 := n.Evaluate(in)
			y, e2 := p.Evaluate(in)
			if e1 != nil || e2 != nil || x["y"] != y["y"] {
				t.Fatalf("mismatch seed=%x in=%v x=%v y=%v e=%v/%v", seed, in, x, y, e1, e2)
			}
		}
	})
}
