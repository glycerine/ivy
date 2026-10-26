package abc

// Stand-alone tests for the Go port of src/base/bac.

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func makeBACWordDesign(t *testing.T, limits BACLimits) *BACDesign {
	t.Helper()
	d, err := NewBACDesign("word", limits)
	if err != nil {
		t.Fatal(err)
	}
	b, err := d.AddNetwork("word")
	if err != nil {
		t.Fatal(err)
	}
	a, err := b.AddInput("a", 4)
	if err != nil {
		t.Fatal(err)
	}
	c, err := b.AddInput("b", 4)
	if err != nil {
		t.Fatal(err)
	}
	s, err := b.AddInput("s", 1)
	if err != nil {
		t.Fatal(err)
	}
	and, err := b.AddNode(BACAnd, "n_and", 4, a, c)
	if err != nil {
		t.Fatal(err)
	}
	add, err := b.AddNode(BACAdd, "n_add", 4, a, c)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := b.AddNode(BACMux, "n_mux", 4, s, and, add)
	if err != nil {
		t.Fatal(err)
	}
	lt, err := b.AddNode(BACLess, "n_lt", 1, a, c)
	if err != nil {
		t.Fatal(err)
	}
	if err = b.AddOutput("y", mux); err != nil {
		t.Fatal(err)
	}
	if err = b.AddOutput("lt", lt); err != nil {
		t.Fatal(err)
	}
	if err = b.Commit(); err != nil {
		t.Fatal(err)
	}
	if err = d.Validate(); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestBACWordEvaluationAndOwnership(t *testing.T) {
	d := makeBACWordDesign(t, BACLimits{})
	for a := uint64(0); a < 16; a++ {
		for b := uint64(0); b < 16; b++ {
			for s := uint64(0); s < 2; s++ {
				got, err := d.Evaluate("word", map[string]uint64{"a": a, "b": b, "s": s})
				if err != nil {
					t.Fatal(err)
				}
				want := a & b
				if s != 0 {
					want = (a + b) & 15
				}
				if got["y"] != want || got["lt"] != bacBoolWord(a < b) {
					t.Fatalf("a=%d b=%d s=%d got=%v want y=%d", a, b, s, got, want)
				}
			}
		}
	}
	n, ok := d.Module("word")
	if !ok {
		t.Fatal("missing module")
	}
	objects := n.Objects()
	objects[0].Name = "changed"
	objects[3].Inputs[0] = 99
	n2, _ := d.Module("word")
	if n2.Objects()[0].Name != "a" || n2.Objects()[3].Inputs[0] == 99 {
		t.Fatal("snapshot aliases design")
	}
	clone, err := d.Clone()
	if err != nil {
		t.Fatal(err)
	}
	clone.Close()
	if _, err = d.Evaluate("word", map[string]uint64{"a": 1, "b": 2, "s": 0}); err != nil {
		t.Fatal("closing clone changed source")
	}
}

func bacBoolWord(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}

func TestBACPrimitiveSemantics(t *testing.T) {
	types := []BACType{BACAnd, BACNand, BACOr, BACNor, BACXor, BACXnor, BACAdd, BACSubtract, BACMultiply, BACMinimum, BACDivide, BACModulo, BACShiftLeft, BACShiftRight, BACRotateLeft, BACRotateRight, BACEqual, BACNotEqual, BACLessEqual, BACGreaterEqual}
	for _, kind := range types {
		d, _ := NewBACDesign("p", BACLimits{})
		b, _ := d.AddNetwork("p")
		a, _ := b.AddInput("a", 4)
		c, _ := b.AddInput("b", 4)
		width := 4
		if kind >= BACLess && kind <= BACNotEqual {
			width = 1
		}
		id, err := b.AddNode(kind, "n", width, a, c)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		_ = b.AddOutput("y", id)
		_ = b.Commit()
		for x := uint64(0); x < 16; x++ {
			for y := uint64(0); y < 16; y++ {
				if (kind == BACDivide || kind == BACModulo) && y == 0 {
					continue
				}
				got, e := d.Evaluate("p", map[string]uint64{"a": x, "b": y})
				if e != nil {
					t.Fatalf("%s: %v", kind, e)
				}
				var want uint64
				switch kind {
				case BACAnd:
					want = x & y
				case BACNand:
					want = ^(x & y) & 15
				case BACOr:
					want = x | y
				case BACNor:
					want = ^(x | y) & 15
				case BACXor:
					want = x ^ y
				case BACXnor:
					want = ^(x ^ y) & 15
				case BACAdd:
					want = (x + y) & 15
				case BACSubtract:
					want = (x - y) & 15
				case BACMultiply:
					want = (x * y) & 15
				case BACMinimum:
					if x < y {
						want = x
					} else {
						want = y
					}
				case BACDivide:
					want = x / y
				case BACModulo:
					want = x % y
				case BACShiftLeft:
					want = (x << y) & 15
				case BACShiftRight:
					want = x >> y
				case BACRotateLeft:
					s := y % 4
					want = ((x << s) | (x >> ((4 - s) % 4))) & 15
				case BACRotateRight:
					s := y % 4
					want = ((x >> s) | (x << ((4 - s) % 4))) & 15
				case BACEqual:
					want = bacBoolWord(x == y)
				case BACNotEqual:
					want = bacBoolWord(x != y)
				case BACLessEqual:
					want = bacBoolWord(x <= y)
				case BACGreaterEqual:
					want = bacBoolWord(x >= y)
				}
				if got["y"] != want {
					t.Fatalf("%s %d %d got %d want %d", kind, x, y, got["y"], want)
				}
			}
		}
	}
}

func TestBACBinaryRoundTripAndMalformed(t *testing.T) {
	d := makeBACWordDesign(t, BACLimits{})
	data, err := d.MarshalBAC()
	if err != nil {
		t.Fatal(err)
	}
	d2, err := UnmarshalBAC(data, BACLimits{})
	if err != nil {
		t.Fatal(err)
	}
	data2, err := d2.MarshalBAC()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(data2) {
		t.Fatal("binary round trip is not deterministic")
	}
	for _, in := range []map[string]uint64{{"a": 3, "b": 5, "s": 0}, {"a": 15, "b": 1, "s": 1}} {
		a, e := d.Evaluate("word", in)
		if e != nil {
			t.Fatal(e)
		}
		b, e := d2.Evaluate("word", in)
		if e != nil || a["y"] != b["y"] || a["lt"] != b["lt"] {
			t.Fatalf("round trip mismatch %v %v %v", a, b, e)
		}
	}
	for i := 0; i < len(data); i++ {
		if _, e := UnmarshalBAC(data[:i], BACLimits{}); e == nil {
			t.Fatalf("accepted truncation %d", i)
		}
	}
	bad := append([]byte(nil), data...)
	bad[0] = 'X'
	if _, e := UnmarshalBAC(bad, BACLimits{}); !errors.Is(e, ErrDimension) {
		t.Fatalf("bad magic: %v", e)
	}
	if _, e := UnmarshalBAC(data, BACLimits{MaxBytes: len(data) - 1}); !errors.Is(e, ErrLimit) {
		t.Fatalf("byte limit: %v", e)
	}
}

func TestBACVerilogRoundTrip(t *testing.T) {
	d := makeBACWordDesign(t, BACLimits{})
	text, err := d.Verilog("word")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "assign n_mux = (s ? n_add : n_and);") {
		t.Fatalf("unexpected text:\n%s", text)
	}
	d2, err := ParseBACVerilog(text, BACLimits{})
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, text)
	}
	text2, err := d2.Verilog("word")
	if err != nil {
		t.Fatal(err)
	}
	if text != text2 {
		t.Fatalf("unstable round trip:\n%s\n---\n%s", text, text2)
	}
	for a := uint64(0); a < 16; a++ {
		for b := uint64(0); b < 16; b++ {
			for s := uint64(0); s < 2; s++ {
				in := map[string]uint64{"a": a, "b": b, "s": s}
				x, _ := d.Evaluate("word", in)
				y, e := d2.Evaluate("word", in)
				if e != nil || x["y"] != y["y"] || x["lt"] != y["lt"] {
					t.Fatalf("mismatch %v %v %v", x, y, e)
				}
			}
		}
	}
	for _, bad := range []string{"", "module x();\n assign y = q;\nendmodule\n", "module x(a);\n input [99:0] a;\nendmodule\n", strings.Repeat("x", 100)} {
		if _, e := ParseBACVerilog(bad, BACLimits{MaxTextBytes: 64}); e == nil {
			t.Fatalf("accepted malformed %q", bad)
		}
	}
}

func makeBACScalar(t *testing.T) *BACDesign {
	t.Helper()
	d, _ := NewBACDesign("logic", BACLimits{})
	b, _ := d.AddNetwork("logic")
	a, _ := b.AddInput("a", 1)
	c, _ := b.AddInput("b", 1)
	x, _ := b.AddNode(BACXor, "x", 1, a, c)
	n, _ := b.AddNode(BACInvert, "n", 1, x)
	_ = b.AddOutput("y", n)
	_ = b.Commit()
	return d
}

func TestBACBLIFRoundTrip(t *testing.T) {
	d := makeBACScalar(t)
	text, e := d.BLIF("logic")
	if e != nil {
		t.Fatal(e)
	}
	d2, e := ParseBACBLIF(text, BACLimits{})
	if e != nil {
		t.Fatalf("%v\n%s", e, text)
	}
	text2, e := d2.BLIF("logic")
	if e != nil {
		t.Fatal(e)
	}
	if text != text2 {
		t.Fatalf("unstable BLIF:\n%s\n---\n%s", text, text2)
	}
	for a := uint64(0); a < 2; a++ {
		for b := uint64(0); b < 2; b++ {
			in := map[string]uint64{"a": a, "b": b}
			x, _ := d.Evaluate("logic", in)
			y, e := d2.Evaluate("logic", in)
			if e != nil || x["y"] != y["y"] {
				t.Fatalf("mismatch %v %v %v", x, y, e)
			}
		}
	}
	bad := strings.Replace(text, "10 1\n01 1", "garbage", 1)
	if _, e = ParseBACBLIF(bad, BACLimits{}); e == nil {
		t.Fatal("accepted malformed cube")
	}
	if _, e = d.BLIF("missing"); !errors.Is(e, ErrDimension) {
		t.Fatalf("missing module: %v", e)
	}
	if _, e = makeBACWordDesign(t, BACLimits{}).BLIF("word"); !errors.Is(e, ErrDimension) {
		t.Fatalf("word BLIF: %v", e)
	}
}

func TestBACLimitsCancellationAndAtomicity(t *testing.T) {
	d, _ := NewBACDesign("x", BACLimits{MaxObjects: 3, MaxEdges: 3})
	b, _ := d.AddNetwork("x")
	a, _ := b.AddInput("a", 1)
	c, _ := b.AddInput("b", 1)
	before := len(b.network.objects)
	if _, e := b.AddNode(BACMux, "bad", 1, a, c); e == nil {
		t.Fatal("accepted bad mux")
	}
	if len(b.network.objects) != before {
		t.Fatal("failed mutation was not atomic")
	}
	n, e := b.AddNode(BACAnd, "n", 1, a, c)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = b.AddNode(BACInvert, "too_many", 1, n); !errors.Is(e, ErrLimit) {
		t.Fatalf("object limit: %v", e)
	}
	_ = b.AddOutput("y", n)
	if e = b.Commit(); e != nil {
		t.Fatal(e)
	}
	if _, e = NewBACDesign("stop", BACLimits{Stop: func() bool { return true }}); !errors.Is(e, ErrCancelled) {
		t.Fatalf("constructor cancellation: %v", e)
	}
	armed := false
	d2 := makeBACWordDesign(t, BACLimits{Stop: func() bool { return armed }})
	armed = true
	if _, e = d2.Evaluate("word", map[string]uint64{"a": 1, "b": 2, "s": 0}); !errors.Is(e, ErrCancelled) {
		t.Fatalf("evaluation cancellation: %v", e)
	}
	d3 := makeBACWordDesign(t, BACLimits{})
	d3.limits.MaxOperations = 2
	if _, e = d3.Evaluate("word", map[string]uint64{"a": 1, "b": 2, "s": 0}); !errors.Is(e, ErrLimit) {
		t.Fatalf("operation limit: %v", e)
	}
}

func TestBACReviewAggregateLimitsAndCommitCancellation(t *testing.T) {
	d := makeBACWordDesign(t, BACLimits{})
	raw, err := d.MarshalBAC()
	if err != nil {
		t.Fatal(err)
	}
	d.limits.MaxBytes = len(raw) - 1
	if _, err = d.MarshalBAC(); !errors.Is(err, ErrLimit) {
		t.Fatalf("binary preflight: %v", err)
	}
	d.limits.MaxBytes = 1 << 26
	d.Spec = strings.Repeat("s", d.limits.MaxTextBytes)
	if err = d.Validate(); !errors.Is(err, ErrLimit) {
		t.Fatalf("public metadata limit: %v", err)
	}
	d.Spec = ""
	d.objects++
	if err = d.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("aggregate counter corruption: %v", err)
	}

	armed := false
	atomic, err := NewBACDesign("atomic", BACLimits{Stop: func() bool { return armed }})
	if err != nil {
		t.Fatal(err)
	}
	b, err := atomic.AddNetwork("atomic")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := b.AddInput("a", 1)
	n, _ := b.AddNode(BACBuffer, "n", 1, a)
	_ = b.AddOutput("y", n)
	armed = true
	if err = b.Commit(); !errors.Is(err, ErrCancelled) {
		t.Fatalf("commit cancellation: %v", err)
	}
	if len(atomic.modules) != 0 || atomic.objects != 0 || atomic.edges != 0 {
		t.Fatal("cancelled commit mutated design")
	}

	var nilDesign *BACDesign
	if err = nilDesign.Validate(); !errors.Is(err, ErrBACClosed) {
		t.Fatalf("nil validate: %v", err)
	}
	if _, err = nilDesign.MarshalBAC(); !errors.Is(err, ErrBACClosed) {
		t.Fatalf("nil marshal: %v", err)
	}
	if nilDesign.ModuleNames() != nil {
		t.Fatal("nil module-name snapshot")
	}
	if _, ok := nilDesign.Module("x"); ok {
		t.Fatal("nil module lookup")
	}

	nameLimited, err := NewBACDesign("n", BACLimits{MaxTextBytes: 8})
	if err != nil {
		t.Fatal(err)
	}
	nb, err := nameLimited.AddNetwork("m")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = nb.AddInput("toolong", 1); !errors.Is(err, ErrLimit) {
		t.Fatalf("builder name budget: %v", err)
	}
}

func TestBACReviewTextStructureAndBudgets(t *testing.T) {
	d := makeBACWordDesign(t, BACLimits{})
	text, err := d.Verilog("word")
	if err != nil {
		t.Fatal(err)
	}
	badVerilog := []string{
		strings.TrimSuffix(text, "endmodule\n"),
		strings.Replace(text, "module word(a, b, s, y, lt);", "module word(a, b, y, s, lt);", 1),
		strings.Replace(text, "  wire [3:0] n_and;", "  wire [3:0] n_and;\n  wire [3:0] n_and;", 1),
		strings.Replace(text, "  wire [3:0] n_and;", "", 1),
		text + "module extra();\nendmodule\n",
	}
	for i, bad := range badVerilog {
		if _, err = ParseBACVerilog(bad, BACLimits{}); err == nil {
			t.Fatalf("accepted malformed Verilog case %d", i)
		}
	}
	d.limits.MaxTextBytes = len(text) - 1
	if _, err = d.Verilog("word"); !errors.Is(err, ErrLimit) {
		t.Fatalf("incremental Verilog limit: %v", err)
	}

	scalar := makeBACScalar(t)
	blif, err := scalar.BLIF("logic")
	if err != nil {
		t.Fatal(err)
	}
	for i, bad := range []string{strings.TrimSuffix(blif, ".end\n"), blif + ".model late\n.end\n"} {
		if _, err = ParseBACBLIF(bad, BACLimits{}); err == nil {
			t.Fatalf("accepted malformed BLIF case %d", i)
		}
	}
	if _, err = ParseBACBLIF(blif, BACLimits{MaxOperations: 5}); !errors.Is(err, ErrLimit) {
		t.Fatalf("BLIF cube work budget: %v", err)
	}
	scalar.limits.MaxTextBytes = len(blif) - 1
	if _, err = scalar.BLIF("logic"); !errors.Is(err, ErrLimit) {
		t.Fatalf("incremental BLIF limit: %v", err)
	}

	collision, err := NewBACDesign("collision", BACLimits{})
	if err != nil {
		t.Fatal(err)
	}
	cb, _ := collision.AddNetwork("collision")
	ci, _ := cb.AddInput("a", 1)
	cn, _ := cb.AddNode(BACInvert, "y", 1, ci)
	_ = cb.AddOutput("y", cn)
	_ = cb.Commit()
	if _, err = collision.Verilog("collision"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("Verilog output/net collision: %v", err)
	}
	if _, err = collision.BLIF("collision"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("BLIF output/net collision: %v", err)
	}
}

func TestBACHierarchyAndUnsupportedValues(t *testing.T) {
	d, _ := NewBACDesign("hier", BACLimits{MaxDepth: 8})
	a, _ := d.AddNetwork("A")
	ai, _ := a.AddInput("i", 1)
	ab, err := a.AddInstance("B", "u_b", 1, ai)
	if err != nil {
		t.Fatal(err)
	}
	_ = a.AddOutput("o", ab)
	if err = a.Commit(); err != nil {
		t.Fatal(err)
	}
	b, _ := d.AddNetwork("B")
	bi, _ := b.AddInput("i", 1)
	bo, _ := b.AddNode(BACBuffer, "n", 1, bi)
	_ = b.AddOutput("o", bo)
	if err = b.Commit(); err != nil {
		t.Fatal(err)
	}
	d.Root = "A"
	if err = d.Validate(); err != nil {
		t.Fatalf("valid hierarchy: %v", err)
	}

	cycle, _ := NewBACDesign("cycle", BACLimits{MaxDepth: 8})
	ca, _ := cycle.AddNetwork("A")
	cai, _ := ca.AddInput("i", 1)
	cau, _ := ca.AddInstance("B", "u", 1, cai)
	_ = ca.AddOutput("o", cau)
	_ = ca.Commit()
	cb, _ := cycle.AddNetwork("B")
	cbi, _ := cb.AddInput("i", 1)
	cbu, _ := cb.AddInstance("A", "u", 1, cbi)
	_ = cb.AddOutput("o", cbu)
	_ = cb.Commit()
	cycle.Root = "A"
	if err = cycle.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("recursive hierarchy: %v", err)
	}

	x, _ := NewBACDesign("x", BACLimits{})
	xb, _ := x.AddNetwork("x")
	if _, err = xb.addObject(BACConstX, "unknown", 1, nil, 0, ""); !errors.Is(err, ErrDimension) {
		t.Fatalf("four-state constant: %v", err)
	}
}

func TestBACVerilogNegatedBinary(t *testing.T) {
	d, _ := NewBACDesign("neg", BACLimits{})
	b, _ := d.AddNetwork("neg")
	a, _ := b.AddInput("a", 1)
	c, _ := b.AddInput("b", 1)
	n, _ := b.AddNode(BACNand, "n", 1, a, c)
	_ = b.AddOutput("y", n)
	_ = b.Commit()
	text, e := d.Verilog("neg")
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(text, "~(a & b)") {
		t.Fatalf("invalid NAND rendering: %s", text)
	}
	copy, e := ParseBACVerilog(text, BACLimits{})
	if e != nil {
		t.Fatal(e)
	}
	for av := uint64(0); av < 2; av++ {
		for bv := uint64(0); bv < 2; bv++ {
			got, e := copy.Evaluate("neg", map[string]uint64{"a": av, "b": bv})
			if e != nil || got["y"] != (^(av&bv)&1) {
				t.Fatalf("a=%d b=%d got=%v err=%v", av, bv, got, e)
			}
		}
	}
}

func TestBACModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 { state = state*6364136223846793005 + 1442695040888963407; return state }
	d := makeBACWordDesign(t, BACLimits{})
	for i := 0; i < 5000; i++ {
		a, b, s := next()&15, next()&15, next()&1
		got, e := d.Evaluate("word", map[string]uint64{"a": a, "b": b, "s": s})
		if e != nil {
			t.Fatal(e)
		}
		want := a & b
		if s != 0 {
			want = (a + b) & 15
		}
		if got["y"] != want || got["lt"] != bacBoolWord(a < b) {
			t.Fatalf("case %d got %v", i, got)
		}
	}
}

func TestBACInventories(t *testing.T) {
	if len(BACSourceInventory) != 16 {
		t.Fatalf("source inventory %d", len(BACSourceInventory))
	}
	want := []string{"bacBlast.c", "bacBac.c", "bacCom.c", "bacLib.c", "bacNtk.c", "bacPrsBuild.c", "bacPrsTrans.c", "bacPtr.c", "bacPtrAbc.c", "bacReadBlif.c", "bacReadSmt.c", "bacReadVer.c", "bacWriteBlif.c", "bacWriteSmt.c", "bacWriteVer.c"}
	if len(BACBuildInventory) != len(want) {
		t.Fatalf("build inventory %d", len(BACBuildInventory))
	}
	for i := range want {
		if BACBuildInventory[i] != want[i] {
			t.Fatalf("build[%d]=%s", i, BACBuildInventory[i])
		}
	}
}

func FuzzBACBinaryAndText(f *testing.F) {
	f.Add([]byte{4, 3, 5, 1, 2})
	f.Add([]byte{1, 0, 0, 0})
	f.Add([]byte{8, 255, 17, 0, 3, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		width := 1
		if len(data) > 0 {
			width = 1 + int(data[0]&7)
		}
		d, e := NewBACDesign("f", BACLimits{MaxObjects: 128, MaxEdges: 256, MaxBytes: 1 << 16, MaxTextBytes: 1 << 16, MaxOperations: 1 << 16})
		if e != nil {
			t.Fatal(e)
		}
		b, _ := d.AddNetwork("f")
		a, _ := b.AddInput("a", width)
		c, _ := b.AddInput("b", width)
		last := a
		ops := []BACType{BACAnd, BACOr, BACXor, BACAdd, BACSubtract}
		for i := 1; i < len(data) && i < 12; i++ {
			kind := ops[int(data[i])%len(ops)]
			id, e := b.AddNode(kind, "n"+strconv.Itoa(i), width, last, c)
			if e != nil {
				t.Fatal(e)
			}
			last = id
		}
		if e = b.AddOutput("y", last); e != nil {
			t.Fatal(e)
		}
		if e = b.Commit(); e != nil {
			t.Fatal(e)
		}
		raw, e := d.MarshalBAC()
		if e != nil {
			t.Fatal(e)
		}
		copy, e := UnmarshalBAC(raw, BACLimits{MaxBytes: 1 << 16, MaxObjects: 128, MaxEdges: 256, MaxOperations: 1 << 16})
		if e != nil {
			t.Fatal(e)
		}
		x := uint64(0)
		y := uint64(0)
		if len(data) > 2 {
			x = uint64(data[1])
			y = uint64(data[2])
		}
		in := map[string]uint64{"a": x, "b": y}
		p, e := d.Evaluate("f", in)
		if e != nil {
			t.Fatal(e)
		}
		q, e := copy.Evaluate("f", in)
		if e != nil || p["y"] != q["y"] {
			t.Fatalf("binary mismatch %v %v %v", p, q, e)
		}
		text, e := d.Verilog("f")
		if e != nil {
			t.Fatal(e)
		}
		parsed, e := ParseBACVerilog(text, BACLimits{MaxTextBytes: 1 << 16, MaxObjects: 128, MaxEdges: 256, MaxOperations: 1 << 16})
		if e != nil {
			t.Fatal(e)
		}
		q, e = parsed.Evaluate("f", in)
		if e != nil || p["y"] != q["y"] {
			t.Fatalf("text mismatch %v %v %v", p, q, e)
		}
		if len(data) > 0 {
			_, _ = UnmarshalBAC(data, BACLimits{MaxBytes: 128, MaxObjects: 64, MaxEdges: 128, MaxOperations: 1024})
			_, _ = ParseBACBLIF(string(data), BACLimits{MaxTextBytes: 128, MaxObjects: 64, MaxEdges: 128, MaxOperations: 1024})
		}
	})
}
