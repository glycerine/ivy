package abc

// Go-only tests for the remaining src/map/amap source units.

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const amapTestLibrary = `
# constants and primitive cells
GATE ZERO 0.25 Y=0;
GATE ONE 0.30 Y=1;
GATE BUF 0.50 Y=A;
PIN A NONINV 1 2 0.1 0.2 0.3 0.4
GATE INV 0.45 Y=!A;
PIN A INV 1 2 0.4 0.2 0.3 0.1
GATE AND_SLOW 2.00 Y=A*B;
PIN A NONINV 1 2 1 0 1 0
PIN B NONINV 1 2 1 0 1 0
GATE AND_FAST 1.00 Y=A&B;
PIN A NONINV 1 2 1 0 1 0
PIN B NONINV 1 2 1 0 1 0
GATE XOR2 1.50 Y=A^B;
PIN A UNKNOWN 1 2 1 0 1 0
PIN B UNKNOWN 1 2 1 0 1 0
GATE MUX2 3.00 Y=(A&C)|(B&!C);
PIN A UNKNOWN 1 2 1 0 1 0
PIN B UNKNOWN 1 2 1 0 1 0
PIN C UNKNOWN 1 2 1 0 1 0
GATE DUAL 4.00 Y=A;
PIN A NONINV 1 2 1 0 1 0
GATE DUAL 4.00 Z=!A;
PIN A INV 1 2 1 0 1 0
.end
`

func mustAMapLibrary(t testing.TB) *AMapCellLibrary {
	t.Helper()
	lib, err := ParseAMapLibrary("test.genlib", []byte(amapTestLibrary), AMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

func TestAMapLibraryParsePrepareTruthAndOwnership(t *testing.T) {
	lib := mustAMapLibrary(t)
	if len(lib.Gates) != 10 || lib.MaxPins() != 3 || !lib.HasXOR || !lib.HasMUX {
		t.Fatalf("library gates=%d max=%d xor=%v mux=%v", len(lib.Gates), lib.MaxPins(), lib.HasXOR, lib.HasMUX)
	}
	if lib.Gates[8].Twin != 9 || lib.Gates[9].Twin != 8 {
		t.Fatalf("twins = %d %d", lib.Gates[8].Twin, lib.Gates[9].Twin)
	}
	if lib.Gates[4].Truth[0]&15 != 8 || lib.Gates[6].Truth[0]&15 != 6 || lib.Gates[7].Truth[0]&255 != 0xac {
		t.Fatalf("truth AND=%x XOR=%x MUX=%x", lib.Gates[4].Truth[0], lib.Gates[6].Truth[0], lib.Gates[7].Truth[0])
	}
	if lib.Sorted[0] != 0 || lib.Sorted[1] != 1 || lib.Sorted[2] != 3 || lib.Constant0 != 0 || lib.Constant1 != 1 || lib.Buffer != 2 || lib.Inverter != 3 {
		t.Fatalf("prepared sorted=%v specials=%d/%d/%d/%d", lib.Sorted, lib.Constant0, lib.Constant1, lib.Buffer, lib.Inverter)
	}
	and, ok := lib.FindTruth(2, []uint64{8})
	if !ok || and != 5 { // minimum-area equivalent AND
		t.Fatalf("AND = %d %v", and, ok)
	}
	clone := lib.Clone()
	clone.Gates[0].Name = "changed"
	clone.Gates[4].Pins[0].Name = "changed"
	clone.Gates[4].Truth[0] = 0
	if lib.Gates[0].Name != "ZERO" || lib.Gates[4].Pins[0].Name != "A" || lib.Gates[4].Truth[0]&15 != 8 {
		t.Fatal("clone aliases source")
	}
}

func TestAMapGenlibRoundTripWildcardCommentsAndFile(t *testing.T) {
	source := []byte("# wildcard pin\nGATE NAND2 1 Y=!(A*B);\\\n\n PIN * UNKNOWN 1 2 3 4 5 6\n.end\n")
	lib, err := ParseAMapLibrary("wild.genlib", source, AMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{lib.Gates[0].Pins[0].Name, lib.Gates[0].Pins[1].Name}; !slices.Equal(got, []string{"A", "B"}) {
		t.Fatalf("wildcard pins %v", got)
	}
	encoded, err := lib.WriteGenlib()
	if err != nil {
		t.Fatal(err)
	}
	reparsed, err := ParseAMapLibrary("round.genlib", encoded, AMapLimits{})
	if err != nil || !slices.Equal(reparsed.Gates[0].Truth, lib.Gates[0].Truth) {
		t.Fatalf("roundtrip err=%v encoded=%q", err, encoded)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "one.genlib")
	if err = os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	fromFile, err := ReadAMapLibrary(path, AMapLimits{})
	if err != nil || len(fromFile.Gates) != 1 {
		t.Fatalf("file library=%v err=%v", fromFile, err)
	}
}

func TestAMapRuleUniquenessPermutationGraphMappingAndOutput(t *testing.T) {
	lib := mustAMapLibrary(t)
	andRule, err := lib.CreateRuleNode(AMapRuleAnd, 1, 0, 0, AMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	andRuleAgain, err := lib.CreateRuleNode(AMapRuleAnd, 0, 1, 0, AMapLimits{})
	if err != nil || andRuleAgain != andRule || len(lib.RuleNodes) != 2 {
		t.Fatalf("unique rule %d %d nodes=%v err=%v", andRule, andRuleAgain, lib.RuleNodes, err)
	}
	perm, err := DeriveAMapPermutation("(C&A)^B", []string{"A", "B", "C"}, AMapLimits{})
	if err != nil || !slices.Equal(perm, []int{2, 0, 1}) {
		t.Fatalf("permutation=%v err=%v", perm, err)
	}
	g, err := NewAMapGraph(AMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := g.CreatePI()
	b, _ := g.CreatePI()
	n, err := g.CreateNode(AMapGraphAND, a, b, 0)
	if err != nil {
		t.Fatal(err)
	}
	n2, err := g.CreateNode(AMapGraphAND, b, a, 0)
	if err != nil || n2 != n {
		t.Fatalf("graph uniqueness %d %d err=%v", n, n2, err)
	}
	if _, err = g.CreatePO(n); err != nil {
		t.Fatal(err)
	}
	matches, err := MapAMapGraph(g, lib, DefaultAMapParams())
	if err != nil || len(matches) != 1 || lib.Gates[matches[0].Gate].Name != "AND_FAST" || matches[0].Delay != 1 {
		t.Fatalf("matches=%+v err=%v", matches, err)
	}
	outputs, err := ProduceAMapOutput(g, lib, matches)
	if err != nil || len(outputs) != 4 || outputs[2].Name != "AND_FAST" || outputs[2].Type != 0 || outputs[3].Type != 1 {
		t.Fatalf("outputs=%+v err=%v", outputs, err)
	}
}

func TestAMapConcreteAIGGraphConversion(t *testing.T) {
	aig := NewAigManager()
	a, err := aig.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := aig.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	node, err := aig.And(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = aig.CreateCO(node); err != nil {
		t.Fatal(err)
	}
	if _, err = aig.And(a, b.Not()); err != nil { // dangling and absent from native DFS transfer
		t.Fatal(err)
	}
	graph, err := CreateAMapGraphFromAIG(aig, AMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Nodes) != 5 || len(graph.Inputs) != 2 || len(graph.Outputs) != 1 || graph.Nodes[3].Type != AMapGraphAND || graph.Nodes[3].Fanins[1]&1 == 0 {
		t.Fatalf("graph=%+v inputs=%v outputs=%v", graph.Nodes, graph.Inputs, graph.Outputs)
	}
}

func TestAMapMatchOrderingSourceTies(t *testing.T) {
	a := AMapMatch{Area: 2, Delay: 3, AverageFanout: 4}
	b := AMapMatch{Area: 2.1, Delay: 2, AverageFanout: 5}
	if CompareAMapMatches(a, b, true, .001) != -1 || CompareAMapMatches(a, b, false, .001) != 1 {
		t.Fatal("area/delay primary order differs")
	}
	// The source's epsilon fanout comparison retains the existing match on a
	// full tie, providing a deterministic first-candidate tie break.
	if CompareAMapMatches(a, a, true, .001) != -1 || CompareAMapMatches(a, a, false, .001) != -1 {
		t.Fatal("tie is not source-faithful")
	}
}

func TestAMapLibertyScalarCombinationalConversion(t *testing.T) {
	source := []byte(`library(x) {
cell(AND2) { area : 1.25; pin(A) { direction : input; } pin(B) { direction : input; }
pin(Y) { direction : output; function : "A & B"; } }
cell(SKIP) { area : 9; dont_use : true; pin(Y) { direction : output; function : "1"; } }
}`)
	lib, err := ParseAMapLiberty("x.lib", source, AMapLimits{})
	if err != nil || len(lib.Gates) != 3 || lib.Gates[0].Name != "_const0_" || lib.Gates[1].Name != "_const1_" || lib.Gates[2].Name != "AND2" || lib.Gates[2].Area != 1.25 || lib.Gates[2].Truth[0]&15 != 8 {
		t.Fatalf("lib=%+v err=%v", lib, err)
	}
	if lib.Gates[2].Pins[0].InputLoad != 1 || lib.Gates[2].Pins[0].MaximumLoad != 999 || lib.Gates[2].Pins[0].BlockMaximum != 1 {
		t.Fatal("Liberty default pin timing", lib.Gates[2].Pins[0])
	}
}

func TestAMapReviewedParserSupportAndGraphDefects(t *testing.T) {
	source := []byte("GATE IMPLICIT 2 Y=A B;\nPIN A UNKNOWN 0 0 0 0 0 0\nPIN B UNKNOWN 0 0 0 0 0 0\nLATCH ignored tokens\nGATE BRACKET 3 Y=A[0]&B;\nPIN A[0] UNKNOWN 0 0 0 0 0 0\nPIN B UNKNOWN 0 0 0 0 0 0\nGATE UNUSED 4 Y=A;\nPIN A UNKNOWN 0 0 0 0 0 0\nPIN B UNKNOWN 0 0 0 0 0 0\n.end\n")
	lib, err := ParseAMapLibrary("review.genlib", source, AMapLimits{})
	if err != nil || len(lib.Gates) != 3 {
		t.Fatal("review genlib", err, lib)
	}
	if lib.Gates[0].Truth[0]&15 != 8 || lib.Gates[1].Truth[0]&15 != 8 {
		t.Fatal("implicit AND or punctuation pin", lib.Gates[0].Truth, lib.Gates[1].Truth)
	}
	if lib.Gates[2].Truth != nil || slices.Contains(lib.Selected, 2) {
		t.Fatal("gate without full support was selected", lib.Gates[2], lib.Selected)
	}

	graph, _ := NewAMapGraph(AMapLimits{})
	a, _ := graph.CreatePI()
	b, _ := graph.CreatePI()
	before := graph.Nodes[amapLitVar(a)].References + graph.Nodes[amapLitVar(b)].References
	if _, err = graph.CreateNode(AMapGraphAND, a, 999<<1, 0); !errors.Is(err, ErrDimension) {
		t.Fatal("invalid fanin", err)
	}
	after := graph.Nodes[amapLitVar(a)].References + graph.Nodes[amapLitVar(b)].References
	if after != before {
		t.Fatal("failed graph mutation was not atomic", before, after)
	}
	xor, _ := graph.CreateNode(AMapGraphXOR, amapLitNot(a), b, 0)
	if !graph.Nodes[amapLitVar(xor)].Phase {
		t.Fatal("XOR phase")
	}
	mux, _ := graph.CreateNode(AMapGraphMUX, a, amapLitNot(b), a)
	if graph.Nodes[amapLitVar(mux)].Phase {
		t.Fatal("MUX phase")
	}
}

func TestAMapReviewedOutputPermutationAndInverter(t *testing.T) {
	lib := mustAMapLibrary(t)
	graph, _ := NewAMapGraph(AMapLimits{})
	x, _ := graph.CreatePI()
	y, _ := graph.CreatePI()
	node, _ := graph.CreateNode(AMapGraphAND, amapLitNot(x), y, 0)
	_, _ = graph.CreatePO(node)
	andGate, _ := lib.FindTruth(2, []uint64{8})
	outputs, err := ProduceAMapOutput(graph, lib, []AMapMatch{{Node: amapLitVar(node), Gate: andGate, PinPermutation: []int{1, 0}}})
	if err != nil || len(outputs) != 5 || outputs[2].Name != "INV" || outputs[3].Name != "AND_FAST" || !slices.Equal(outputs[3].Fanins, []int{2, 1}) || !slices.Equal(outputs[4].Fanins, []int{3}) {
		t.Fatalf("permuted/inverted output = %+v err=%v", outputs, err)
	}
}

func TestAMapReviewedLibertySkipRules(t *testing.T) {
	source := []byte(`library(x) {
cell(NOAREA) { pin(A) { direction : input; } pin(Y) { direction : output; function : "A"; } }
cell(CONST) { area : 2; pin(Y) { direction : output; function : "1"; } }
cell(FALSE_DONT_USE) { area : 3; dont_use : false; pin(A) { direction : input; } pin(Y) { direction : output; function : "A"; } }
cell(SEQ) { area : 4; ff(IQ,IQN) { next_state : "D"; } pin(D) { direction : input; } pin(Q) { direction : output; function : "D"; } }
cell(KEEP) { area : 5; pin(A) { direction : input; } pin(Y) { direction : output; function : "A"; } }
}`)
	lib, err := ParseAMapLiberty("skip.lib", source, AMapLimits{})
	if err != nil || len(lib.Gates) != 3 || lib.Gates[2].Name != "KEEP" {
		t.Fatalf("Liberty source skip rules lib=%+v err=%v", lib, err)
	}
}

func TestAMapRemainingLimitsCancellationAndAtomicErrors(t *testing.T) {
	bad := []string{
		"", "PIN A UNKNOWN 0 0 0 0 0 0\n.end", "GATE X bad Y=A;\nPIN A UNKNOWN 0 0 0 0 0 0\n.end",
		"GATE X 1 Y=A;\nPIN A SIDEWAYS 0 0 0 0 0 0\n.end", "GATE X 1 Y=A;\n.end",
		"GATE X 1 Y=A;\nPIN A UNKNOWN 0 0 0 0 0 0\nPIN A UNKNOWN 0 0 0 0 0 0\n.end",
	}
	for _, source := range bad {
		if lib, err := ParseAMapLibrary("bad", []byte(source), AMapLimits{}); lib != nil || err == nil {
			t.Fatalf("bad source %q lib=%v err=%v", source, lib, err)
		}
	}
	if lib, err := ParseAMapLibrary("limit", []byte(amapTestLibrary), AMapLimits{MaxBytes: 5}); lib != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit lib=%v err=%v", lib, err)
	}
	if lib, err := ParseAMapLibrary("limit", []byte(amapTestLibrary), AMapLimits{MaxGates: 1}); lib != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("gate limit lib=%v err=%v", lib, err)
	}
	calls := 0
	if lib, err := ParseAMapLibrary("cancel", []byte(amapTestLibrary), AMapLimits{Stop: func() bool { calls++; return calls == 3 }}); lib != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel lib=%v err=%v calls=%d", lib, err, calls)
	}
}

func independentAMapFormula(a, b, c bool, selector int) bool {
	switch selector {
	case 0:
		return a && b
	case 1:
		return a != b
	case 2:
		return (a && c) || (b && !c)
	case 3:
		return !(a || b)
	default:
		return (!a && b) || c
	}
}

func amapModelFormula(selector int) (string, []string) {
	switch selector {
	case 0:
		return "A*B", []string{"A", "B"}
	case 1:
		return "A^B", []string{"A", "B"}
	case 2:
		return "(A&C)|(B&!C)", []string{"A", "B", "C"}
	case 3:
		return "!(A+B)", []string{"A", "B"}
	default:
		return "(A'*B)+C", []string{"A", "B", "C"}
	}
}

func checkAMapTruthModel(t testing.TB, seed uint64) {
	t.Helper()
	selector := int(seed % 5)
	formula, names := amapModelFormula(selector)
	var source strings.Builder
	fmt.Fprintf(&source, "GATE G 1 Y=%s;\n", formula)
	for _, name := range names {
		fmt.Fprintf(&source, "PIN %s UNKNOWN 0 0 0 0 0 0\n", name)
	}
	source.WriteString(".end\n")
	lib, err := ParseAMapLibrary("model", []byte(source.String()), AMapLimits{})
	if err != nil {
		t.Fatalf("seed=%d: %v", seed, err)
	}
	truth := lib.Gates[0].Truth[0]
	for assignment := 0; assignment < 1<<len(names); assignment++ {
		a := assignment&1 != 0
		b := assignment&2 != 0
		c := assignment&4 != 0
		got := truth&(uint64(1)<<uint(assignment)) != 0
		want := independentAMapFormula(a, b, c, selector)
		if got != want {
			t.Fatalf("seed=%d selector=%d assignment=%d got=%v want=%v truth=%x", seed, selector, assignment, got, want, truth)
		}
	}
}

func TestAMapRemainingIndependentTruthTimingMappingModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		checkAMapTruthModel(t, state)
		a := AMapMatch{Area: float64(state & 255), Delay: float64((state >> 8) & 255), AverageFanout: float64((state >> 16) & 255)}
		b := AMapMatch{Area: a.Area + 1, Delay: a.Delay, AverageFanout: a.AverageFanout}
		if CompareAMapMatches(a, b, true, 0.001) != -1 {
			t.Fatalf("execution %d area ordering", execution)
		}
		graph, err := NewAMapGraph(AMapLimits{})
		if err != nil {
			t.Fatal(err)
		}
		inputs := make([]int, 3)
		for index := range inputs {
			inputs[index], _ = graph.CreatePI()
			if state&(1<<uint(index)) != 0 {
				inputs[index] = amapLitNot(inputs[index])
			}
		}
		kind := AMapGraphAND
		wantPhase := state&1 != 0 && state&2 != 0
		if state>>24&1 != 0 {
			kind = AMapGraphXOR
			wantPhase = state&1 != 0 != (state&2 != 0)
		}
		if state>>25&1 != 0 {
			kind = AMapGraphMUX
			control := state&4 != 0
			wantPhase = state&1 != 0 && !control || state&2 != 0 && control
		}
		literal, err := graph.CreateNode(kind, inputs[0], inputs[1], inputs[2])
		if err != nil || graph.Nodes[amapLitVar(literal)].Phase != wantPhase {
			t.Fatalf("execution %d phase kind=%d got=%v want=%v err=%v", execution, kind, graph.Nodes[amapLitVar(literal)].Phase, wantPhase, err)
		}
	}
}

func FuzzAMapRemainingTruthModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, math.MaxUint64} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) { checkAMapTruthModel(t, seed) })
}

func FuzzAMapGenlibRejectOrRoundTrip(f *testing.F) {
	for _, seed := range []string{amapTestLibrary, "", "GATE G 1 Y=A;\nPIN A UNKNOWN 0 0 0 0 0 0\n.end"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 4096 {
			t.Skip()
		}
		lib, err := ParseAMapLibrary("fuzz", []byte(text), AMapLimits{MaxBytes: 4096, MaxGates: 64, MaxPins: 15, MaxFormulaNodes: 256, MaxRuleNodes: 256})
		if err != nil {
			if lib != nil {
				t.Fatal("failed parse returned partial library")
			}
			return
		}
		encoded, err := lib.WriteGenlib()
		if err != nil {
			t.Fatal(err)
		}
		round, err := ParseAMapLibrary("round", encoded, AMapLimits{MaxBytes: 1 << 20})
		if err != nil || len(round.Gates) != len(lib.Gates) {
			t.Fatalf("roundtrip gates=%d err=%v text=%q", len(lib.Gates), err, text)
		}
		for i := range lib.Gates {
			if !slices.Equal(lib.Gates[i].Truth, round.Gates[i].Truth) {
				t.Fatalf("gate %d truth mismatch", i)
			}
		}
	})
}
