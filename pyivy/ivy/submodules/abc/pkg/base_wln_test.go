package abc

// Stand-alone Go tests for src/base/wln.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func newWlnTestNetwork(t *testing.T, name string) *WlnNetwork {
	t.Helper()
	network, err := NewWlnNetwork(name, WlnLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return network
}

func addWlnTestObject(t *testing.T, network *WlnNetwork, object WlnObject) int {
	t.Helper()
	id, err := network.AddObject(object)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestWlnOperationLayoutRangesAndConstants(t *testing.T) {
	if WlnCI != 3 || WlnCO != 4 || WlnTable != 77 || WlnDFFRSE != 89 || WlnConst != 97 || WlnOperationLast != 99 {
		t.Fatal("operation layout")
	}
	if WlnDFFRSE.String() != "DFFRSE" || WlnConcat.String() != "{}" {
		t.Fatal("operation names")
	}
	for _, test := range []struct {
		wordRange WlnRange
		width     int
	}{{WlnRange{7, 0}, 8}, {WlnRange{0, 7}, 8}, {WlnRange{-2, -5}, 4}} {
		if got := test.wordRange.Width(); got != test.width {
			t.Fatal(test, got)
		}
	}
	constant, err := WlnConstantFromBits([]uint64{0xa5}, 8)
	if err != nil || constant != "8'b10100101" {
		t.Fatal(constant, err)
	}
	constant, err = WlnConstantFromString("10xZ", 4)
	if err != nil || constant != "4'b10xZ" {
		t.Fatal(constant, err)
	}
	if _, err = WlnConstantFromString("102", 3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestWlnNetworkConstructionOwnershipFanoutAndDFS(t *testing.T) {
	n := newWlnTestNetwork(t, "owned")
	a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{3, 0}, Name: "a"})
	b := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{3, 0}, Name: "b"})
	fans := []int{a, b}
	sum := addWlnTestObject(t, n, WlnObject{Type: WlnAdd, Range: WlnRange{3, 0}, Fanins: fans, Name: "sum"})
	fans[0] = 0
	out := addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{3, 0}, Fanins: []int{sum}, Name: "out"})
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	objects := n.Objects()
	objects[sum].Fanins[0] = 0
	if n.Objects()[sum].Fanins[0] != a {
		t.Fatal("borrowed object storage")
	}
	refs, err := n.References()
	if err != nil || refs[a] != 1 || refs[b] != 1 || refs[sum] != 1 {
		t.Fatal(refs, err)
	}
	fanouts, err := n.Fanouts()
	if err != nil || !slices.Equal(fanouts[sum], []int{out}) {
		t.Fatal(fanouts, err)
	}
	acyclic, err := n.IsAcyclic()
	if err != nil || !acyclic {
		t.Fatal(acyclic, err)
	}
	dup, mapping, err := n.DuplicateDFS()
	if err != nil || dup.Check() != nil || mapping[out] == 0 {
		t.Fatal(mapping, err)
	}
	if text, _ := dup.Verilog(); !strings.Contains(text, "sum = a + b") {
		t.Fatal(text)
	}
	if err = n.SetFanin(sum, 0, sum); err != nil {
		t.Fatal(err)
	}
	acyclic, err = n.IsAcyclic()
	if err != nil || acyclic {
		t.Fatal(acyclic, err)
	}
}

func TestWlnNDRRoundTripOrderingAndMalformedAtomicity(t *testing.T) {
	n := newWlnTestNetwork(t, "ndr")
	n.Spec, n.SMTLib = "source.smt2", true
	a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Signed: true, Range: WlnRange{7, 0}, Name: "a"})
	c := addWlnTestObject(t, n, WlnObject{Type: WlnConst, Range: WlnRange{7, 0}, Constant: "8'b00000101", Name: "five"})
	x := addWlnTestObject(t, n, WlnObject{Type: WlnBitXor, Range: WlnRange{7, 0}, Fanins: []int{a, c}, Name: "x", InstanceID: 7})
	addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{7, 0}, Fanins: []int{x}, Name: "z"})
	n.UseInstanceDelays = true
	design, err := n.ToNDR()
	if err != nil {
		t.Fatal(err)
	}
	copyNet, err := WlnNetworkFromNDR(design, WlnLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if copyNet.Name != "ndr" || copyNet.Spec != "source.smt2" || !copyNet.SMTLib || !copyNet.UseInstanceDelays || !slices.Equal(copyNet.CIs(), []int{1}) || copyNet.Objects()[3].Constant != "" || copyNet.Objects()[2].Constant != "8'b00000101" {
		t.Fatal(copyNet.Objects())
	}
	first, _ := n.Verilog()
	second, _ := copyNet.Verilog()
	if first != second {
		t.Fatalf("round trip differs\n%s\n%s", first, second)
	}
	bad := &WlnNDRDesign{Name: "bad", Objects: []WlnNDRObject{{Type: WlnBitBuf, Range: WlnRange{}, Fanins: []int{2}, Name: "x"}}}
	if got, badErr := WlnNetworkFromNDR(bad, WlnLimits{}); got != nil || !errors.Is(badErr, ErrDimension) {
		t.Fatal(got, badErr)
	}
}

func TestWlnNDRTablePayloadRoundTrip(t *testing.T) {
	n := newWlnTestNetwork(t, "table")
	index := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{1, 0}, Name: "index"})
	table := addWlnTestObject(t, n, WlnObject{Type: WlnTable, Range: WlnRange{3, 0}, Fanins: []int{index}, Table: []uint64{0x3210}, Parameter: 7, Name: "value"})
	addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{3, 0}, Fanins: []int{table}, Name: "out"})
	design, err := n.ToNDR()
	if err != nil {
		t.Fatal(err)
	}
	round, err := WlnNetworkFromNDR(design, WlnLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(round.Objects()[table].Table, []uint64{0x3210}) {
		t.Fatal(round.Objects()[table].Table)
	}
	want, err := n.Verilog()
	if err != nil {
		t.Fatal(err)
	}
	got, err := round.Verilog()
	if err != nil || got != want {
		t.Fatalf("roundtrip table: %v\n%s\n%s", err, want, got)
	}
}

func TestWlnSimulationAndIndependentCampaign(t *testing.T) {
	n := newWlnTestNetwork(t, "sim")
	a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "a"})
	b := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "b"})
	x := addWlnTestObject(t, n, WlnObject{Type: WlnBitXor, Range: WlnRange{7, 0}, Fanins: []int{a, b}, Name: "x"})
	s := addWlnTestObject(t, n, WlnObject{Type: WlnAdd, Range: WlnRange{7, 0}, Fanins: []int{x, a}, Name: "s"})
	m := addWlnTestObject(t, n, WlnObject{Type: WlnBitMux, Range: WlnRange{7, 0}, Fanins: []int{a, x, s}, Name: "m"})
	addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{7, 0}, Fanins: []int{m}, Name: "z"})
	state := uint64(0x243f6a8885a308d3)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		left, right := uint64(byte(state)), uint64(byte(state>>8))
		outputs, next, err := n.Evaluate([]uint64{left, right}, nil)
		if err != nil || len(next) != 0 {
			t.Fatal(execution, err)
		}
		xor := left ^ right
		sum := byte(xor + left)
		want := uint64(sum)
		if left&1 != 0 {
			want = xor
		}
		if outputs[0] != want {
			t.Fatalf("execution %d got=%x want=%x", execution, outputs[0], want)
		}
	}
}

func TestWlnSignedShiftComparisonAndRotateSemantics(t *testing.T) {
	n := newWlnTestNetwork(t, "widths")
	value := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "value"})
	amount := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "amount"})
	signedShift := addWlnTestObject(t, n, WlnObject{Type: WlnArithShiftRight, Range: WlnRange{7, 0}, Fanins: []int{value, amount}, Name: "signed_shift"})
	logicalShift := addWlnTestObject(t, n, WlnObject{Type: WlnShiftRight, Range: WlnRange{7, 0}, Fanins: []int{value, amount}, Name: "logical_shift"})
	minusOne := addWlnTestObject(t, n, WlnObject{Type: WlnConst, Range: WlnRange{7, 0}, Constant: "8'b11111111", Name: "minus_one"})
	signedLess := addWlnTestObject(t, n, WlnObject{Type: WlnSignedLess, Range: WlnRange{}, Fanins: []int{value, minusOne}, Name: "signed_less"})
	rotate := addWlnTestObject(t, n, WlnObject{Type: WlnRotateLeft, Range: WlnRange{7, 0}, Fanins: []int{value}, Parameter: 3, Name: "rotate"})
	for index, id := range []int{signedShift, logicalShift, signedLess, rotate} {
		addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: n.objects[id].Range, Fanins: []int{id}, Name: "out" + string(rune('0'+index))})
	}
	outputs, _, err := n.Evaluate([]uint64{0x80, 1}, nil)
	if err != nil || !slices.Equal(outputs, []uint64{0xc0, 0x40, 1, 0x04}) {
		t.Fatal(outputs, err)
	}
	outputs, _, err = n.Evaluate([]uint64{0x80, 64}, nil)
	if err != nil || outputs[0] != 0xff || outputs[1] != 0 {
		t.Fatal(outputs, err)
	}

	bad := newWlnTestNetwork(t, "bad_rotate")
	in := addWlnTestObject(t, bad, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "in"})
	rot := addWlnTestObject(t, bad, WlnObject{Type: WlnRotateRight, Range: WlnRange{7, 0}, Fanins: []int{in}, Parameter: 8, Name: "rot"})
	addWlnTestObject(t, bad, WlnObject{Type: WlnCO, Range: WlnRange{7, 0}, Fanins: []int{rot}, Name: "out"})
	if _, _, err = bad.Evaluate([]uint64{1}, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = bad.Verilog(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestWlnNetworkAndRetimeIndependentModelCampaigns(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		n := newWlnTestNetwork(t, "model")
		a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "a"})
		b := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "b"})
		nodes := []int{a, b}
		count := 1 + int(state%6)
		for i := 0; i < count; i++ {
			state = state*6364136223846793005 + 1
			left := nodes[len(nodes)-1]
			state = state*6364136223846793005 + 1
			right := nodes[int(state%uint64(len(nodes)))]
			id := addWlnTestObject(t, n, WlnObject{Type: WlnBitXor, Range: WlnRange{7, 0}, Fanins: []int{left, right}, Name: "v" + string(rune('a'+i))})
			nodes = append(nodes, id)
		}
		root := nodes[len(nodes)-1]
		addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{7, 0}, Fanins: []int{root}, Name: "z"})
		design, err := n.ToNDR()
		if err != nil {
			t.Fatal(execution, err)
		}
		round, err := WlnNetworkFromNDR(design, WlnLimits{})
		if err != nil || round.ObjectCount() != n.ObjectCount() {
			t.Fatal(execution, err)
		}
		for id, object := range n.Objects() {
			other := round.Objects()[id]
			if object.Type != other.Type || object.Range != other.Range || !slices.Equal(object.Fanins, other.Fanins) {
				t.Fatal(execution, id, object, other)
			}
		}
		delays := make([]int, n.ObjectCount())
		distance := make([]int, n.ObjectCount())
		want := 0
		for id, object := range n.Objects() {
			if id == 0 {
				continue
			}
			state = state*6364136223846793005 + 1
			delays[id] = int(state % 6)
			for _, fanin := range object.Fanins {
				if distance[id] < distance[fanin] {
					distance[id] = distance[fanin]
				}
			}
			if object.Type != WlnCI {
				distance[id] += delays[id]
			}
			if !isWlnTerminal(object.Type) && distance[id] > want {
				want = distance[id]
			}
		}
		result, err := n.Retime(delays, 1)
		if err != nil || result.BestDelay != want || len(result.Moves) != 0 {
			t.Fatalf("execution %d got=%+v want=%d err=%v", execution, result, want, err)
		}
	}
}

func TestWlnSequentialSimulationRetimeAndDelayPrecedence(t *testing.T) {
	n := newWlnTestNetwork(t, "seq")
	a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{}, Name: "a"})
	b := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{}, Name: "b"})
	ffFaninsA := []int{a, 0, 0, 0, 0, 0, 0, 0}
	ffFaninsB := []int{b, 0, 0, 0, 0, 0, 0, 0}
	fa := addWlnTestObject(t, n, WlnObject{Type: WlnDFFRSE, Range: WlnRange{}, Fanins: ffFaninsA, Name: "fa"})
	fb := addWlnTestObject(t, n, WlnObject{Type: WlnDFFRSE, Range: WlnRange{}, Fanins: ffFaninsB, Name: "fb"})
	and := addWlnTestObject(t, n, WlnObject{Type: WlnBitAnd, Range: WlnRange{}, Fanins: []int{fa, fb}, Name: "and", NameID: 5})
	tail := addWlnTestObject(t, n, WlnObject{Type: WlnBitInv, Range: WlnRange{}, Fanins: []int{and}, Name: "tail"})
	addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{}, Fanins: []int{tail}, Name: "z"})
	out, next, err := n.Evaluate([]uint64{1, 0}, []uint64{1, 1})
	if err != nil || out[0] != 0 || !slices.Equal(next, []uint64{1, 0}) {
		t.Fatal(out, next, err)
	}
	delays, err := n.DefaultDelays()
	if err != nil || delays[fa] != 1 || delays[tail] != 1 {
		t.Fatal(delays, err)
	}
	delays[and] = 10
	result, err := n.Retime(delays, 4)
	if err != nil || len(result.Moves) == 0 || !result.Moves[0].Forward || !slices.Equal(result.Moves[0].Objects, []int{and}) {
		t.Fatal(result, err)
	}
	recorded, err := n.RetimeWithOptions(delays, WlnRetimeOptions{MaxMoves: 1})
	if err != nil || recorded.MoveCount != 1 || len(recorded.Instructions) != 5 || recorded.Instructions[1] != 0 || recorded.Instructions[3] != -5 || recorded.Instructions[4] != 0 {
		t.Fatal(recorded, err)
	}
	if got, limitErr := n.RetimeWithOptions(delays, WlnRetimeOptions{MaxMoves: 1, MaxEdgeLinks: 2}); got.Moves != nil || !errors.Is(limitErr, ErrLimit) {
		t.Fatal(got, limitErr)
	}
	if got, cancelErr := n.RetimeWithOptions(delays, WlnRetimeOptions{Stop: func() bool { return true }}); got.Moves != nil || !errors.Is(cancelErr, ErrCancelled) {
		t.Fatal(got, cancelErr)
	}
	n.UseInstanceDelays = true
	n.objects[and].InstanceID = 19
	user, err := n.DefaultDelays()
	if err != nil || user[and] != 19 || user[fa] != 0 {
		t.Fatal(user, err)
	}
}

func TestWlnRetimeCountsFlopDelayAndIgnoresCIDelay(t *testing.T) {
	n := newWlnTestNetwork(t, "delay")
	a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{}, Name: "a"})
	ff := addWlnTestObject(t, n, WlnObject{Type: WlnDFFRSE, Range: WlnRange{}, Fanins: []int{a, 0, 0, 0, 0, 0, 0, 0}, Name: "ff"})
	addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{}, Fanins: []int{ff}, Name: "out"})
	delays := make([]int, n.ObjectCount())
	delays[a], delays[ff] = 20, 7
	result, err := n.Retime(delays, 1)
	if err != nil || result.InitialDelay != 7 || result.BestDelay != 7 {
		t.Fatal(result, err)
	}
}

func TestWlnVerilogGoldenOrderingAndSpecialForms(t *testing.T) {
	n := newWlnTestNetwork(t, "gold")
	a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Signed: true, Range: WlnRange{3, 0}, Name: "a"})
	b := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{3, 0}, Name: "b"})
	mux := addWlnTestObject(t, n, WlnObject{Type: WlnBitMux, Range: WlnRange{3, 0}, Fanins: []int{a, a, b}, Name: "mux"})
	cat := addWlnTestObject(t, n, WlnObject{Type: WlnConcat, Range: WlnRange{7, 0}, Fanins: []int{mux, b}, Name: "cat"})
	addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{7, 0}, Fanins: []int{cat}, Name: "z"})
	got, err := n.Verilog()
	if err != nil {
		t.Fatal(err)
	}
	want := "module gold (a, b, z);\n  input wire signed [3:0] a;\n  input wire [3:0] b;\n  wire [3:0] mux = a ? a : b;\n  wire [7:0] cat = { mux, b };\n  output wire [7:0] z = cat;\nendmodule\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	again, _ := n.Verilog()
	if got != again {
		t.Fatal("nondeterministic output")
	}
}

func TestWlnBoundsCancellationAndAtomicity(t *testing.T) {
	if _, err := NewWlnNetwork("bad", WlnLimits{MaxObjects: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	n, err := NewWlnNetwork("limit", WlnLimits{MaxObjects: 2})
	if err != nil {
		t.Fatal(err)
	}
	addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{}})
	before := n.Objects()
	if _, err = n.AddObject(WlnObject{Type: WlnCI, Range: WlnRange{}}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if !slices.EqualFunc(before, n.Objects(), func(a, b WlnObject) bool { return a.Type == b.Type && slices.Equal(a.Fanins, b.Fanins) }) {
		t.Fatal("limit mutated network")
	}
	n.limits.Stop = func() bool { return true }
	if _, err = n.References(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func FuzzWlnNetworkSimulationAndNDR(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{0xff, 0, 17, 99})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 || len(data) > 128 {
			return
		}
		n := newWlnTestNetwork(t, "fuzz")
		a := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "a"})
		b := addWlnTestObject(t, n, WlnObject{Type: WlnCI, Range: WlnRange{7, 0}, Name: "b"})
		nodes := []int{a, b}
		for i, value := range data {
			left := nodes[int(value)%len(nodes)]
			right := nodes[int(value>>3)%len(nodes)]
			operations := []WlnOperation{WlnBitAnd, WlnBitOr, WlnBitXor, WlnAdd}
			id := addWlnTestObject(t, n, WlnObject{Type: operations[int(value>>6)], Range: WlnRange{7, 0}, Fanins: []int{left, right}, Name: "n" + string(rune('a'+i%26))})
			nodes = append(nodes, id)
		}
		root := nodes[len(nodes)-1]
		addWlnTestObject(t, n, WlnObject{Type: WlnCO, Range: WlnRange{7, 0}, Fanins: []int{root}, Name: "z"})
		inputs := []uint64{uint64(data[0]), uint64(data[len(data)-1])}
		first, _, err := n.Evaluate(inputs, nil)
		if err != nil {
			t.Fatal(err)
		}
		design, err := n.ToNDR()
		if err != nil {
			t.Fatal(err)
		}
		copyNet, err := WlnNetworkFromNDR(design, WlnLimits{})
		if err != nil {
			t.Fatal(err)
		}
		second, _, err := copyNet.Evaluate(inputs, nil)
		if err != nil || !slices.Equal(first, second) {
			t.Fatal(first, second, err)
		}
		if _, err = copyNet.Verilog(); err != nil {
			t.Fatal(err)
		}
	})
}
