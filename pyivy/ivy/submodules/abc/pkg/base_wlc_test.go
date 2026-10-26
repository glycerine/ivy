package abc

// Standalone Go tests for the WLC port. Models below do not invoke native code.

import (
	"errors"
	"math/bits"
	"strings"
	"testing"
)

func wlcTestLimits() WLCLimits {
	return WLCLimits{MaxObjects: 256, MaxEdges: 1024, MaxBits: 4096, MaxBytes: 1 << 20, MaxOperations: 1 << 20, MaxTextBytes: 1 << 20, MaxDepth: 128, MaxFanins: 64}
}

func TestWLCInventoryTypesAndDefaults(t *testing.T) {
	if WLCObjectTypeCount != 60 || len(wlcTypeNames) != 60 {
		t.Fatalf("type inventory %d/%d", WLCObjectTypeCount, len(wlcTypeNames))
	}
	if len(WLCBuildUnits) != 19 || WLCBuildUnits[0] != "wlcAbs.c" || WLCBuildUnits[18] != "wlcWriteVer.c" {
		t.Fatalf("build inventory %#v", WLCBuildUnits)
	}
	if len(WLCSourceInventory) != 22 {
		t.Fatalf("source inventory %d", len(WLCSourceInventory))
	}
	for i := WLCObjectType(1); i < WLCObjectTypeCount; i++ {
		if i.String() == "" || i.String() == "out_of_bound" {
			t.Fatalf("missing type name %d", i)
		}
	}
	p := DefaultWLCParameters()
	if p.IterationLimit != 1000 || !p.XorOutput || !p.CheckClauses || !p.MFFC || !p.LoadTrace || !p.Hybrid {
		t.Fatalf("defaults %#v", p)
	}
	if _, err := (WLCRange{End: 63, Begin: 0}).Width(); err != nil {
		t.Fatal(err)
	}
	if _, err := (WLCRange{End: int(^uint(0) >> 1), Begin: -1}).Width(); !errors.Is(err, ErrLimit) {
		t.Fatalf("range overflow %v", err)
	}
}

func makeWLCBinary(t *testing.T, typ WLCObjectType, width int, signed bool) *WLCNetwork {
	t.Helper()
	n, err := NewWLCNetwork("bin", wlcTestLimits())
	if err != nil {
		t.Fatal(err)
	}
	a, err := n.AddInput("a", WLCRange{End: width - 1}, signed)
	if err != nil {
		t.Fatal(err)
	}
	b, err := n.AddInput("b", WLCRange{End: width - 1}, signed)
	if err != nil {
		t.Fatal(err)
	}
	y, err := n.AddNode("y", typ, WLCRange{End: width - 1}, signed, []int{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if err := n.MarkOutput(y); err != nil {
		t.Fatal(err)
	}
	return n
}

func binaryExpected(typ WLCObjectType, a, b uint64, width int, signed bool) uint64 {
	m := wlcMask(width)
	switch typ {
	case WLCObjectBitAnd:
		return a & b & m
	case WLCObjectBitOr:
		return (a | b) & m
	case WLCObjectBitXor:
		return (a ^ b) & m
	case WLCObjectBitNand:
		return ^(a & b) & m
	case WLCObjectBitNor:
		return ^(a | b) & m
	case WLCObjectBitXnor:
		return ^(a ^ b) & m
	case WLCObjectAdd:
		return (a + b) & m
	case WLCObjectSub:
		return (a - b) & m
	case WLCObjectMultiply:
		return a * b & m
	case WLCObjectDivide:
		if b == 0 {
			return m
		}
		if signed {
			return uint64(wlcSigned(a, width)/wlcSigned(b, width)) & m
		}
		return a / b & m
	case WLCObjectRemainder:
		if b == 0 {
			return a & m
		}
		if signed {
			return uint64(wlcSigned(a, width)%wlcSigned(b, width)) & m
		}
		return a % b & m
	}
	return 0
}

func TestWLCExhaustiveBinaryOperators(t *testing.T) {
	types := []WLCObjectType{WLCObjectBitAnd, WLCObjectBitOr, WLCObjectBitXor, WLCObjectBitNand, WLCObjectBitNor, WLCObjectBitXnor, WLCObjectAdd, WLCObjectSub, WLCObjectMultiply, WLCObjectDivide, WLCObjectRemainder}
	for _, typ := range types {
		for _, signed := range []bool{false, true} {
			n := makeWLCBinary(t, typ, 4, signed)
			for a := uint64(0); a < 16; a++ {
				for b := uint64(0); b < 16; b++ {
					got, err := n.Evaluate([]uint64{a, b}, nil)
					if err != nil {
						t.Fatalf("%s %x %x: %v", typ, a, b, err)
					}
					want := binaryExpected(typ, a, b, 4, signed)
					if got.Outputs[0] != want {
						t.Fatalf("%s signed=%v %x %x got %x want %x", typ, signed, a, b, got.Outputs[0], want)
					}
				}
			}
		}
	}
}

func TestWLCWordOperations(t *testing.T) {
	n, err := NewWLCNetwork("words", wlcTestLimits())
	if err != nil {
		t.Fatal(err)
	}
	a, _ := n.AddInput("a", WLCRange{End: 7}, true)
	b, _ := n.AddInput("b", WLCRange{End: 3}, false)
	sh, _ := n.AddNode("sh", WLCObjectShiftRight, WLCRange{End: 7}, false, []int{a, b})
	ash, _ := n.AddNode("ash", WLCObjectArithmeticShiftRight, WLCRange{End: 7}, true, []int{a, b})
	rol, _ := n.AddNode("rol", WLCObjectRotateLeft, WLCRange{End: 7}, false, []int{a, b})
	cat, _ := n.AddNode("cat", WLCObjectBitConcat, WLCRange{End: 11}, false, []int{a, b})
	sl, _ := n.AddSlice("sl", a, 5, 2, false)
	red, _ := n.AddNode("red", WLCObjectReduceXor, WLCRange{}, false, []int{a})
	sq, _ := n.AddNode("sq", WLCObjectSqrt, WLCRange{End: 7}, false, []int{a})
	dec, _ := n.AddNode("dec", WLCObjectDecoder, WLCRange{End: 15}, false, []int{b})
	for _, id := range []int{sh, ash, rol, cat, sl, red, sq, dec} {
		if err := n.MarkOutput(id); err != nil {
			t.Fatal(err)
		}
	}
	r, err := n.Evaluate([]uint64{0x92, 3}, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []uint64{0x12, 0xf2, 0x94, 0x923, 4, uint64(bits.OnesCount8(0x92) & 1), 12, 8}
	for i := range want {
		if r.Outputs[i] != want[i] {
			t.Fatalf("output %d got %x want %x", i, r.Outputs[i], want[i])
		}
	}
}

func TestWLCDeclaredRangeSliceAndHugeMuxSelect(t *testing.T) {
	n, _ := NewWLCNetwork("ranges", wlcTestLimits())
	a, _ := n.AddInput("a", WLCRange{End: 11, Begin: 4}, false)
	s, _ := n.AddInput("s", WLCRange{}, false)
	b, _ := n.AddConstant("b", WLCRange{End: 7}, 0x5a, false)
	c, _ := n.AddConstant("c", WLCRange{End: 7}, 0xa5, false)
	sl, err := n.AddSlice("sl", a, 9, 6, false)
	if err != nil {
		t.Fatal(err)
	}
	m, err := n.AddNode("m", WLCObjectMux, WLCRange{End: 7}, false, []int{s, b, c})
	if err != nil {
		t.Fatal(err)
	}
	_ = n.MarkOutput(sl)
	_ = n.MarkOutput(m)
	r, err := n.Evaluate([]uint64{0xd6, 1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outputs[0] != 5 || r.Outputs[1] != 0xa5 {
		t.Fatalf("outputs %x", r.Outputs)
	}
	if _, err := n.AddSlice("bad", a, 12, 8, false); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad range %v", err)
	}
	wideSel, _ := n.AddInput("wide_sel", WLCRange{End: 2}, false)
	if _, err := n.AddNode("bad_mux", WLCObjectMux, WLCRange{End: 7}, false, []int{wideSel, b, c}); !errors.Is(err, ErrDimension) {
		t.Fatalf("source-shaped mux arity %v", err)
	}
}

func TestWLCSourceShapeAndSequentialInvariants(t *testing.T) {
	n, _ := NewWLCNetwork("shape", wlcTestLimits())
	a, _ := n.AddInput("a", WLCRange{End: 3}, false)
	b, _ := n.AddInput("b", WLCRange{End: 3}, false)
	if _, err := n.AddNode("orphan", WLCObjectFO, WLCRange{End: 3}, false, nil); !errors.Is(err, ErrDimension) {
		t.Fatalf("orphan flop output %v", err)
	}
	if _, err := n.AddNode("bad_cat", WLCObjectBitConcat, WLCRange{End: 6}, false, []int{a, b}); !errors.Is(err, ErrDimension) {
		t.Fatalf("concat width %v", err)
	}
	if _, err := n.AddNode("bad_rot", WLCObjectRotateLeft, WLCRange{End: 7}, false, []int{a, b}); !errors.Is(err, ErrDimension) {
		t.Fatalf("rotate width %v", err)
	}
	ff, err := n.AddFlopOutput("ff", WLCRange{End: 7}, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := n.SetFlopNext(ff, a); !errors.Is(err, ErrDimension) {
		t.Fatalf("flop width %v", err)
	}
	next, _ := n.AddNode("next", WLCObjectBuf, WLCRange{End: 7}, false, []int{ff})
	if err := n.SetFlopNext(ff, next); err != nil {
		t.Fatal(err)
	}
	if err := n.Validate(); err != nil {
		t.Fatal(err)
	}
	n.outputs = append(n.outputs, next)
	if err := n.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("unmarked output list %v", err)
	}
}

func TestWLCSnapshotOwnershipAndLifecycle(t *testing.T) {
	var nilNetwork *WLCNetwork
	if nilNetwork.Inputs() != nil || nilNetwork.Outputs() != nil || nilNetwork.Flops() != nil {
		t.Fatal("nil snapshots")
	}
	n, _ := NewWLCNetwork("snap", wlcTestLimits())
	a, _ := n.AddInput("a", WLCRange{}, false)
	ins := n.Inputs()
	ins[0] = 99
	if n.Inputs()[0] != a {
		t.Fatal("borrowed inputs")
	}
	n.Close()
	if n.Inputs() != nil || n.Outputs() != nil || n.Flops() != nil {
		t.Fatal("closed snapshots")
	}
}

func TestWLCMuxLogicComparisonsAndLUT(t *testing.T) {
	n, _ := NewWLCNetwork("misc", wlcTestLimits())
	s, _ := n.AddInput("s", WLCRange{}, false)
	a, _ := n.AddInput("a", WLCRange{End: 3}, true)
	b, _ := n.AddInput("b", WLCRange{End: 3}, true)
	m, _ := n.AddNode("m", WLCObjectMux, WLCRange{End: 3}, true, []int{s, a, b})
	lt, _ := n.AddNode("lt", WLCObjectLess, WLCRange{}, true, []int{a, b})
	imp, _ := n.AddNode("imp", WLCObjectLogicImplies, WLCRange{}, false, []int{a, b})
	lut, _ := n.AddLUT("lut", s, 1, 2)
	for _, id := range []int{m, lt, imp, lut} {
		if err := n.MarkOutput(id); err != nil {
			t.Fatal(err)
		}
	}
	r, err := n.Evaluate([]uint64{1, 15, 2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := r.Outputs, []uint64{2, 1, 1, 1}; !wlcEqualWords(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func wlcEqualWords(a, b []uint64) bool {
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

func TestWLCSequentialEvaluation(t *testing.T) {
	n, _ := NewWLCNetwork("seq", wlcTestLimits())
	one, _ := n.AddConstant("one", WLCRange{}, 1, false)
	ff, err := n.AddFlopOutput("state", WLCRange{End: 3}, false, 7)
	if err != nil {
		t.Fatal(err)
	}
	add, _ := n.AddNode("next", WLCObjectAdd, WLCRange{End: 3}, false, []int{ff, one})
	if err := n.SetFlopNext(ff, add); err != nil {
		t.Fatal(err)
	}
	if err := n.MarkOutput(ff); err != nil {
		t.Fatal(err)
	}
	state, _ := n.InitialState()
	r, err := n.Evaluate(nil, state)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outputs[0] != 7 || r.NextState[0] != 8 {
		t.Fatalf("step %#v", r)
	}
	if _, err := WLCWriteVerilog(n); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("sequential write %v", err)
	}
}

func TestWLCAnalysisCloneReportAndLifecycle(t *testing.T) {
	n, _ := NewWLCNetwork("analysis", wlcTestLimits())
	a, _ := n.AddInput("a", WLCRange{End: 3}, false)
	b, _ := n.AddInput("b", WLCRange{End: 3}, false)
	sum, _ := n.AddNode("sum", WLCObjectAdd, WLCRange{End: 3}, false, []int{a, b})
	mul, _ := n.AddNode("mul", WLCObjectMultiply, WLCRange{End: 3}, false, []int{sum, b})
	buf, _ := n.AddNode("buf", WLCObjectBuf, WLCRange{End: 3}, false, []int{mul})
	_ = n.MarkOutput(buf)
	levels, err := n.Levels()
	if err != nil || levels[buf] != 3 {
		t.Fatalf("levels %v %v", levels, err)
	}
	rev, _ := n.ReverseLevels()
	if rev[a] <= rev[sum] || rev[sum] <= rev[mul] {
		t.Fatalf("reverse %v", rev)
	}
	refs, _ := n.References()
	if refs[b] != 2 || refs[buf] != 1 {
		t.Fatalf("refs %v", refs)
	}
	supp, _ := n.Support(buf)
	if len(supp) != 2 || supp[0] != a || supp[1] != b {
		t.Fatalf("support %v", supp)
	}
	leaves, nodes, err := n.ArithmeticWindow(mul)
	if err != nil || len(nodes) != 2 || len(leaves) != 2 {
		t.Fatalf("window leaves=%v nodes=%v err=%v", leaves, nodes, err)
	}
	s, _ := n.Stats()
	if s.Objects != 5 || s.Edges != 5 || s.MaxLevel != 3 || s.TypeCounts[WLCObjectAdd] != 1 {
		t.Fatalf("stats %#v", s)
	}
	report, err := n.Report()
	if err != nil || !strings.Contains(report, "objects=5") || !strings.Contains(report, "+=1") {
		t.Fatalf("report %q %v", report, err)
	}
	c, err := n.Clone()
	if err != nil {
		t.Fatal(err)
	}
	n.Close()
	if _, err := n.Evaluate(nil, nil); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed eval %v", err)
	}
	r, err := c.Evaluate([]uint64{2, 3}, nil)
	if err != nil || r.Outputs[0] != 15 {
		t.Fatalf("clone %v %v", r, err)
	}
}

func TestWLCVerilogRoundTrip(t *testing.T) {
	n, _ := NewWLCNetwork("round", wlcTestLimits())
	a, _ := n.AddInput("a", WLCRange{End: 3}, false)
	b, _ := n.AddInput("b", WLCRange{End: 3}, false)
	x, _ := n.AddNode("x", WLCObjectBitXor, WLCRange{End: 3}, false, []int{a, b})
	y, _ := n.AddNode("y", WLCObjectAdd, WLCRange{End: 3}, false, []int{x, b})
	_ = n.MarkOutput(y)
	text, err := WLCWriteVerilog(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "module round") {
		t.Fatalf("text %q", text)
	}
	p, err := WLCParseVerilog(text, wlcTestLimits())
	if err != nil {
		t.Fatal(err)
	}
	for aVal := uint64(0); aVal < 16; aVal++ {
		for bVal := uint64(0); bVal < 16; bVal++ {
			r1, e1 := n.Evaluate([]uint64{aVal, bVal}, nil)
			r2, e2 := p.Evaluate([]uint64{aVal, bVal}, nil)
			if e1 != nil || e2 != nil || !wlcEqualWords(r1.Outputs, r2.Outputs) {
				t.Fatalf("%x %x: %v/%v %v/%v", aVal, bVal, r1.Outputs, r2.Outputs, e1, e2)
			}
		}
	}
}

func TestWLCLimitsCancellationAndAtomicity(t *testing.T) {
	limits := wlcTestLimits()
	limits.MaxObjects = 2
	n, _ := NewWLCNetwork("limit", limits)
	a, _ := n.AddInput("a", WLCRange{}, false)
	b, _ := n.AddInput("b", WLCRange{}, false)
	if _, err := n.AddNode("x", WLCObjectBitAnd, WLCRange{}, false, []int{a, b}); !errors.Is(err, ErrLimit) {
		t.Fatalf("object limit %v", err)
	}
	objects, _ := n.Objects()
	if len(objects) != 2 {
		t.Fatalf("failed mutation retained %d", len(objects))
	}
	stop := false
	limits = wlcTestLimits()
	limits.Stop = func() bool { return stop }
	n, _ = NewWLCNetwork("cancel", limits)
	a, _ = n.AddInput("a", WLCRange{}, false)
	b, _ = n.AddInput("b", WLCRange{}, false)
	stop = true
	if _, err := n.AddNode("x", WLCObjectBitAnd, WLCRange{}, false, []int{a, b}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	stop = false
	objects, _ = n.Objects()
	if len(objects) != 2 {
		t.Fatal("cancelled mutation retained")
	}
	x, _ := n.AddNode("x", WLCObjectPower, WLCRange{End: 63}, false, []int{a, b})
	_ = n.MarkOutput(x)
	n.limits.MaxOperations = 3
	if _, err := n.Evaluate([]uint64{2, 63}, nil); !errors.Is(err, ErrLimit) {
		t.Fatalf("work limit %v", err)
	}
}

func TestWLCMalformedGraphsAndUnsupportedValues(t *testing.T) {
	n, _ := NewWLCNetwork("malformed", wlcTestLimits())
	a, _ := n.AddInput("a", WLCRange{}, false)
	if _, err := n.AddNode("bad", WLCObjectAdd, WLCRange{}, false, []int{a}); !errors.Is(err, ErrDimension) {
		t.Fatalf("shape %v", err)
	}
	x, _ := n.AddConstant("x", WLCRange{}, 0, true)
	_ = n.MarkOutput(x)
	if _, err := n.Evaluate([]uint64{0}, nil); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("x value %v", err)
	}
	n.objects[x].ID = 99
	if err := n.Validate(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("corrupt id %v", err)
	}
	if _, err := n.Object(-1); !errors.Is(err, ErrInvariant) {
		t.Fatalf("bad object %v", err)
	}
}

type wlcModelRNG struct{ state uint64 }

func (r *wlcModelRNG) next() uint64 {
	r.state = r.state*6364136223846793005 + 1442695040888963407
	return r.state
}

func TestWLCIndependentModel5000(t *testing.T) {
	rng := wlcModelRNG{state: 0x123456789abcdef}
	for trial := 0; trial < 5000; trial++ {
		w := int(rng.next()%8) + 1
		mask := wlcMask(w)
		aVal, bVal := rng.next()&mask, rng.next()&mask
		typ := []WLCObjectType{WLCObjectAdd, WLCObjectSub, WLCObjectMultiply, WLCObjectBitAnd, WLCObjectBitOr, WLCObjectBitXor}[rng.next()%6]
		n := makeWLCBinary(t, typ, w, false)
		got, err := n.Evaluate([]uint64{aVal, bVal}, nil)
		if err != nil {
			t.Fatal(err)
		}
		want := binaryExpected(typ, aVal, bVal, w, false)
		if got.Outputs[0] != want {
			t.Fatalf("trial %d %s got %x want %x", trial, typ, got.Outputs[0], want)
		}
	}
}

func FuzzWLCWordSemantics(f *testing.F) {
	for _, seed := range []struct {
		a, b  uint64
		width uint8
		op    uint8
	}{{0, 0, 1, 0}, {15, 3, 4, 1}, {0xff, 0x81, 8, 2}, {7, 0, 4, 3}} {
		f.Add(seed.a, seed.b, seed.width, seed.op)
	}
	f.Fuzz(func(t *testing.T, a, b uint64, width, op uint8) {
		w := int(width%16) + 1
		types := []WLCObjectType{WLCObjectAdd, WLCObjectSub, WLCObjectMultiply, WLCObjectBitAnd, WLCObjectBitOr, WLCObjectBitXor, WLCObjectDivide, WLCObjectRemainder}
		typ := types[int(op)%len(types)]
		n := makeWLCBinary(t, typ, w, false)
		m := wlcMask(w)
		a, b = a&m, b&m
		r, err := n.Evaluate([]uint64{a, b}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if want := binaryExpected(typ, a, b, w, false); r.Outputs[0] != want {
			t.Fatalf("%s %x %x got %x want %x", typ, a, b, r.Outputs[0], want)
		}
	})
}
