package abc

// Standalone tests for the Go port of ABC src/base/io.

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func ioTestNetwork(t *testing.T) *IONetwork {
	t.Helper()
	n := &IONetwork{
		Name:    "demo",
		Inputs:  []string{"a", "b"},
		Outputs: []string{"y", "z"},
		Nodes: []IONode{
			{Name: "x", Fanins: []string{"a", "b"}, Cubes: []IOCube{{Pattern: "10", Value: true}, {Pattern: "01", Value: true}}},
			{Name: "y", Fanins: []string{"x", "a"}, Cubes: []IOCube{{Pattern: "11", Value: true}}},
			{Name: "z", Fanins: []string{"a", "b"}, Cubes: []IOCube{{Pattern: "1-", Value: true}, {Pattern: "-1", Value: true}}},
		},
	}
	if err := n.Validate(IOLimits{}); err != nil {
		t.Fatal(err)
	}
	return n
}

func ioTruth(t *testing.T, n *IONetwork) []uint64 {
	t.Helper()
	truth := make([]uint64, len(n.Outputs))
	for mask := 0; mask < 1<<len(n.Inputs); mask++ {
		inputs := make(map[string]bool, len(n.Inputs))
		for i, name := range n.Inputs {
			inputs[name] = mask&(1<<i) != 0
		}
		got, err := n.Evaluate(inputs, map[string]bool{}, IOLimits{})
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range got {
			if v {
				truth[i] |= 1 << mask
			}
		}
	}
	return truth
}

func TestIOValidationEvaluationAndLifecycle(t *testing.T) {
	n := ioTestNetwork(t)
	if got := ioTruth(t, n); !reflect.DeepEqual(got, []uint64{0x2, 0xe}) {
		t.Fatalf("truth = %#v", got)
	}
	bad := cloneIONetwork(n)
	bad.Nodes[0].Fanins[0] = "x"
	if err := bad.Validate(IOLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("self-cycle error = %v", err)
	}
	bad = cloneIONetwork(n)
	bad.Nodes[0].Cubes[0].Pattern = "2-"
	if err := bad.Validate(IOLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatalf("cube error = %v", err)
	}
	before := cloneIONetwork(n)
	if err := n.Validate(IOLimits{MaxEdges: 1}); !errors.Is(err, ErrLimit) || !reflect.DeepEqual(n, before) {
		t.Fatalf("limit = %v mutation=%v", err, !reflect.DeepEqual(n, before))
	}
	n.Close()
	if err := n.Validate(IOLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed error = %v", err)
	}
}

func TestIOBLIFRoundTripAndLatch(t *testing.T) {
	source := []byte(".model seq\n.inputs a b\n.outputs y\n.latch x q 0\n.names a b x\n11 1\n.names q a y\n1- 1\n-1 1\n.end\n")
	n, err := ParseIONetwork(source, IOFormatBLIF, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Latches) != 1 || n.Latches[0].Init != LatchInitZero {
		t.Fatalf("latches = %#v", n.Latches)
	}
	got, err := n.Evaluate(map[string]bool{"a": false, "b": true}, map[string]bool{"q": true}, IOLimits{})
	if err != nil || !reflect.DeepEqual(got, []bool{true}) {
		t.Fatalf("evaluate = %v, %v", got, err)
	}
	next, err := n.NextState(map[string]bool{"a": true, "b": true}, map[string]bool{"q": false}, IOLimits{})
	if err != nil || !next["q"] {
		t.Fatalf("next = %v, %v", next, err)
	}
	encoded, err := FormatIONetwork(n, IOFormatBLIF, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	round, err := ParseIONetwork(encoded, IOFormatBLIF, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruthWithState(t, n, false), ioTruthWithState(t, round, false)) {
		t.Fatalf("round trip = %v", err)
	}
}

func ioTruthWithState(t *testing.T, n *IONetwork, q bool) []uint64 {
	t.Helper()
	truth := make([]uint64, len(n.Outputs))
	for mask := 0; mask < 1<<len(n.Inputs); mask++ {
		in := make(map[string]bool)
		for i, name := range n.Inputs {
			in[name] = mask&(1<<i) != 0
		}
		state := make(map[string]bool)
		for _, l := range n.Latches {
			state[l.Output] = q
		}
		out, err := n.Evaluate(in, state, IOLimits{})
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range out {
			if v {
				truth[i] |= 1 << mask
			}
		}
	}
	return truth
}

func TestIOBenchRoundTrip(t *testing.T) {
	data := []byte("INPUT(a)\nINPUT(b)\nOUTPUT(y)\nx = AND(a, b)\ny = NOT(x)\n")
	n, err := ParseIONetwork(data, IOFormatBench, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got := ioTruth(t, n); !reflect.DeepEqual(got, []uint64{0x7}) {
		t.Fatalf("truth = %#v", got)
	}
	encoded, err := FormatIONetwork(n, IOFormatBench, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	round, err := ParseIONetwork(encoded, IOFormatBench, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruth(t, n), ioTruth(t, round)) {
		t.Fatalf("round = %s, %v", encoded, err)
	}
}

func TestIOEquationAndVerilog(t *testing.T) {
	eqn := []byte("INORDER = a b; OUTORDER = y; x = a ^ b; y = x & !a;")
	n, err := ParseIONetwork(eqn, IOFormatEquation, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruth(t, n), []uint64{0x4}) {
		t.Fatalf("equation = %v", err)
	}
	verilog, err := FormatIONetwork(n, IOFormatVerilog, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	round, err := ParseIONetwork(verilog, IOFormatVerilog, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruth(t, n), ioTruth(t, round)) {
		t.Fatalf("verilog round = %q %v", verilog, err)
	}
	encoded, err := FormatIONetwork(n, IOFormatEquation, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	round, err = ParseIONetwork(encoded, IOFormatEquation, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruth(t, n), ioTruth(t, round)) {
		t.Fatalf("equation round = %q %v", encoded, err)
	}
}

func TestIOPLARoundTrip(t *testing.T) {
	data := []byte(".i 2\n.o 2\n.ilb a b\n.ob x y\n.type f\n10 10\n01 10\n1- 01\n-1 01\n.e\n")
	n, err := ParseIONetwork(data, IOFormatPLA, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruth(t, n), []uint64{0x6, 0xe}) {
		t.Fatalf("PLA = %#v %v", ioTruth(t, n), err)
	}
	encoded, err := FormatIONetwork(n, IOFormatPLA, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	round, err := ParseIONetwork(encoded, IOFormatPLA, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruth(t, n), ioTruth(t, round)) {
		t.Fatalf("round = %q %v", encoded, err)
	}
}

func TestIOJSONRoundTrip(t *testing.T) {
	n := ioTestNetwork(t)
	data, err := FormatIONetwork(n, IOFormatJSON, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	round, err := ParseIONetwork(data, IOFormatJSON, IOLimits{})
	if err != nil || !reflect.DeepEqual(ioTruth(t, n), ioTruth(t, round)) {
		t.Fatalf("JSON = %v", err)
	}
	good, _ := FormatIONetwork(n, IOFormatJSON, IOLimits{})
	good = append(good, []byte("{}")...)
	if _, err := ParseIONetwork(good, IOFormatJSON, IOLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatalf("trailing JSON = %v", err)
	}
}

func TestIOProjectionWritersAndLimits(t *testing.T) {
	n := ioTestNetwork(t)
	for _, format := range []IOFormat{IOFormatDOT, IOFormatGML, IOFormatList, IOFormatSMV, IOFormatCNF} {
		a, err := FormatIONetwork(n, format, IOLimits{})
		if err != nil || len(a) == 0 {
			t.Fatalf("format %d = %q %v", format, a, err)
		}
		b, err := FormatIONetwork(n, format, IOLimits{})
		if err != nil || !bytes.Equal(a, b) {
			t.Fatalf("format %d nondeterministic", format)
		}
		if got, err := FormatIONetwork(n, format, IOLimits{MaxOutputBytes: len(a) - 1}); got != nil || !errors.Is(err, ErrLimit) {
			t.Fatalf("format %d limit = %q %v", format, got, err)
		}
	}
	if got, err := ParseIONetwork([]byte("x"), IOFormatEDIF, IOLimits{}); got != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("unsupported = %v %v", got, err)
	}
	for _, format := range []IOFormat{IOFormatBAF, IOFormatBook, IOFormatBBLIF, IOFormatBLIFMV} {
		if got, err := FormatIONetwork(n, format, IOLimits{}); got != nil || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("unsupported format %d = %q %v", format, got, err)
		}
	}
	if _, err := ParseIONetwork([]byte("INPUT(a)\nOUTPUT(a)\n"), IOFormatBench, IOLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel = %v", err)
	}
}

func TestIODetectAndInventory(t *testing.T) {
	rows := map[string]IOFormat{"x.blif": IOFormatBLIF, "x.BENCH": IOFormatBench, "x.eqn": IOFormatEquation, "x.pla": IOFormatPLA, "x.sv": IOFormatVerilog, "x.json": IOFormatJSON, "x.unknown": IOFormatUnknown}
	for name, want := range rows {
		if got := IODetectFormat(name); got != want {
			t.Fatalf("%s = %d", name, got)
		}
	}
	if len(IOBaseBuildSources) != 31 || IOBaseBuildSources[0] != "io.c" || IOBaseBuildSources[30] != "ioWriteSmv.c" {
		t.Fatalf("inventory = %#v", IOBaseBuildSources)
	}
}

func TestIOIndependentModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0x10bace))
	for trial := 0; trial < 5000; trial++ {
		vars := 1 + rng.Intn(4)
		inputs := make([]string, vars)
		for i := range inputs {
			inputs[i] = fmt.Sprintf("x%d", i)
		}
		n := &IONetwork{Name: "m", Inputs: inputs, Outputs: []string{"y"}}
		truth := uint64(rng.Intn(1 << (1 << vars)))
		cubes := make([]IOCube, 0)
		for mask := 0; mask < 1<<vars; mask++ {
			if truth&(1<<mask) != 0 {
				cubes = append(cubes, IOCube{Pattern: binaryPattern(mask, vars), Value: true})
			}
		}
		n.Nodes = []IONode{{Name: "y", Fanins: slicesCloneStrings(inputs), Cubes: cubes}}
		if err := n.Validate(IOLimits{}); err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 1<<vars; mask++ {
			assignment := make(map[string]bool, vars)
			for i, name := range inputs {
				assignment[name] = mask&(1<<i) != 0
			}
			got, err := n.Evaluate(assignment, map[string]bool{}, IOLimits{})
			if err != nil || got[0] != (truth&(1<<mask) != 0) {
				t.Fatalf("trial %d mask %d = %v %v", trial, mask, got, err)
			}
		}
		data, err := FormatIONetwork(n, IOFormatBLIF, IOLimits{})
		if err != nil {
			t.Fatal(err)
		}
		round, err := ParseIONetwork(data, IOFormatBLIF, IOLimits{})
		if err != nil || !reflect.DeepEqual(ioTruth(t, n), ioTruth(t, round)) {
			t.Fatalf("trial %d round = %v", trial, err)
		}
	}
}

func slicesCloneStrings(s []string) []string { return append([]string(nil), s...) }

func TestIOMalformedCorpus(t *testing.T) {
	cases := []struct {
		format IOFormat
		data   string
	}{
		{IOFormatBench, "INPUT(a)\nOUTPUT(y)\ny = AND(a, missing)\n"},
		{IOFormatBLIF, ".model m\n.inputs a\n.outputs y\n.names a y\n2 1\n.end\n"},
		{IOFormatBLIF, ".model m\n.inputs a\n.outputs y\n.names a y\n1 1\n"},
		{IOFormatPLA, ".i 1\n.o 1\n2 1\n.e\n"},
		{IOFormatEquation, "INORDER = a; OUTORDER = y; y = (a;"},
		{IOFormatVerilog, "module m(a,y); input a; output y; assign y = a;"},
	}
	for _, tc := range cases {
		if got, err := ParseIONetwork([]byte(tc.data), tc.format, IOLimits{}); got != nil || err == nil {
			t.Fatalf("accepted %d %q: %#v %v", tc.format, tc.data, got, err)
		}
	}
}

func FuzzIOFormatRoundTrips(f *testing.F) {
	f.Add([]byte{2, 0x6, 0xe})
	f.Add([]byte{3, 0x96, 0x80})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		vars := 1 + int(data[0]%4)
		inputs := make([]string, vars)
		for i := range inputs {
			inputs[i] = fmt.Sprintf("x%d", i)
		}
		cubes := make([]IOCube, 0)
		for mask := 0; mask < 1<<vars; mask++ {
			if data[1+mask%(len(data)-1)]&(1<<uint(mask&7)) != 0 {
				cubes = append(cubes, IOCube{Pattern: binaryPattern(mask, vars), Value: true})
			}
		}
		n := &IONetwork{Name: "f", Inputs: inputs, Outputs: []string{"y"}, Nodes: []IONode{{Name: "y", Fanins: slicesCloneStrings(inputs), Cubes: cubes}}}
		for _, format := range []IOFormat{IOFormatBLIF, IOFormatPLA, IOFormatJSON, IOFormatVerilog, IOFormatEquation} {
			encoded, err := FormatIONetwork(n, format, IOLimits{MaxOperations: 1 << 20, MaxOutputBytes: 1 << 20})
			if err != nil {
				t.Fatalf("format %d: %v", format, err)
			}
			round, err := ParseIONetwork(encoded, format, IOLimits{MaxOperations: 1 << 20, MaxInputBytes: 1 << 20})
			if err != nil {
				t.Fatalf("parse %d %q: %v", format, encoded, err)
			}
			if !reflect.DeepEqual(ioTruth(t, n), ioTruth(t, round)) {
				t.Fatalf("format %d truth mismatch", format)
			}
		}
	})
}

func TestIOTextBoundsPreflight(t *testing.T) {
	n := ioTestNetwork(t)
	for limit := 1; limit < 64; limit++ {
		got, err := FormatIONetwork(n, IOFormatBLIF, IOLimits{MaxOutputBytes: limit})
		if got != nil || !errors.Is(err, ErrLimit) {
			t.Fatalf("limit %d = %q %v", limit, got, err)
		}
	}
	large := strings.Repeat("x", 100)
	if got, err := ParseIONetwork([]byte(large), IOFormatBench, IOLimits{MaxInputBytes: 10}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("input limit = %#v %v", got, err)
	}
}

func TestIOReviewSourceSemantics(t *testing.T) {
	offset, err := ParseIONetwork([]byte(".model m\n.inputs a\n.outputs y\n.names a y\n0 0\n.end\n"), IOFormatBLIF, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []bool{false, true} {
		got, err := offset.Evaluate(map[string]bool{"a": value}, nil, IOLimits{})
		if err != nil || len(got) != 1 || got[0] != value {
			t.Fatalf("offset a=%v = %v, %v", value, got, err)
		}
	}
	constant, err := ParseIONetwork([]byte(".model m\n.outputs q\n.latch 1 q 1\n.end\n"), IOFormatBLIF, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	next, err := constant.NextState(nil, map[string]bool{"q": false}, IOLimits{})
	if err != nil || !next["q"] {
		t.Fatalf("constant latch = %v, %v", next, err)
	}
	if err := (&IONetwork{Name: "m", Inputs: []string{"a"}, Outputs: []string{"y"}, Nodes: []IONode{{Name: "y", Fanins: []string{"a", "a"}, Cubes: []IOCube{{Pattern: "11", Value: true}}}}}).Validate(IOLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatalf("duplicate fanin = %v", err)
	}
	for _, source := range []string{
		".i 1\n.i 1\n.o 1\n0 1\n.e\n",
		".i 1\n.o 1\n0 1\n.e\n0 1\n",
	} {
		if _, err := ParseIONetwork([]byte(source), IOFormatPLA, IOLimits{}); !errors.Is(err, ErrFormat) {
			t.Fatalf("malformed PLA %q = %v", source, err)
		}
	}
}

func TestIOCumulativeBudgetsAndExpressionDepth(t *testing.T) {
	n := &IONetwork{Name: "m", Inputs: []string{"a"}, Outputs: []string{"y"}, Nodes: []IONode{{Name: "y", Fanins: []string{"a"}, Cubes: []IOCube{{Pattern: "1", Value: true}}}}}
	if err := n.Validate(IOLimits{MaxOperations: 5}); err != nil {
		t.Fatalf("validation threshold = %v", err)
	}
	if got, err := n.Evaluate(map[string]bool{"a": true}, nil, IOLimits{MaxOperations: 5}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("cumulative evaluation = %v, %v", got, err)
	}
	deep := "INORDER = a; OUTORDER = y; y = " + strings.Repeat("!", 20) + "a;"
	if got, err := ParseIONetwork([]byte(deep), IOFormatEquation, IOLimits{MaxDepth: 8, MaxOperations: 1 << 20}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("deep expression = %#v, %v", got, err)
	}
	chain := &IONetwork{Name: "m", Inputs: []string{"a"}, Outputs: []string{"y"}, Nodes: []IONode{
		{Name: "x", Fanins: []string{"a"}, Cubes: []IOCube{{Pattern: "1", Value: true}}},
		{Name: "y", Fanins: []string{"x"}, Cubes: []IOCube{{Pattern: "1", Value: true}}},
	}}
	if err := chain.Validate(IOLimits{MaxDepth: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("logic depth = %v", err)
	}
}

func TestIOVerilogStructureAndSequentialRejection(t *testing.T) {
	bad := []string{
		"module m(a,y); input a; output z; assign z = a; endmodule;",
		"module m(a,y); input a; output y; endmodule;",
		"module m(a,y); input a; output y; wire y; assign y = a; endmodule;",
		"module m(a,y); input a; output y; assign y = a; assign y = a; endmodule;",
	}
	for _, source := range bad {
		if got, err := ParseIONetwork([]byte(source), IOFormatVerilog, IOLimits{}); got != nil || !errors.Is(err, ErrFormat) {
			t.Fatalf("malformed Verilog %q = %#v, %v", source, got, err)
		}
	}
	sequential := &IONetwork{Name: "m", Outputs: []string{"q"}, Latches: []IOLatch{{Input: "1", Output: "q", Init: LatchInitZero}}}
	for _, format := range []IOFormat{IOFormatBench, IOFormatEquation, IOFormatPLA, IOFormatVerilog, IOFormatSMV, IOFormatCNF} {
		if got, err := FormatIONetwork(sequential, format, IOLimits{}); got != nil || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("sequential format %d = %q, %v", format, got, err)
		}
	}
}

func TestIOCNFConstantFaninSemantics(t *testing.T) {
	n := &IONetwork{Name: "m", Inputs: []string{"a"}, Outputs: []string{"y"}, Nodes: []IONode{{Name: "y", Fanins: []string{"1", "a"}, Cubes: []IOCube{{Pattern: "11", Value: true}}}}}
	data, err := FormatIONetwork(n, IOFormatCNF, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 || lines[0] != "p cnf 2 2" {
		t.Fatalf("CNF = %q", data)
	}
	for a := 0; a < 2; a++ {
		for y := 0; y < 2; y++ {
			satisfied := true
			for _, line := range lines[1:] {
				clause := false
				for _, field := range strings.Fields(line) {
					lit, err := strconv.Atoi(field)
					if err != nil {
						t.Fatal(err)
					}
					if lit == 0 {
						continue
					}
					value := a != 0
					if lit < 0 {
						lit = -lit
						value = !value
					}
					if lit == 2 {
						value = y != 0
						if field[0] == '-' {
							value = !value
						}
					}
					clause = clause || value
				}
				satisfied = satisfied && clause
			}
			if satisfied != (a == y) {
				t.Fatalf("a=%d y=%d satisfied=%v CNF=%q", a, y, satisfied, data)
			}
		}
	}
}
