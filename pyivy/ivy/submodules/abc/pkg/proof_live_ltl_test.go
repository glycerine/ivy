package abc

// Stand-alone tests for ABC src/proof/live/ltl_parser.c.

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

func TestLTLPrefixGrammarRoundTrip(t *testing.T) {
	formula := "& :G :F + request !stall :U ready done"
	node, err := ParseLTLPrefix(formula, LTLParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if node.Operator != LTLAnd || !node.Left.IsGFWellFormed() || node.IsGFWellFormed() {
		t.Fatalf("unexpected source grammar tree: %#v", node)
	}
	encoded, err := node.Prefix()
	if err != nil {
		t.Fatal(err)
	}
	again, err := ParseLTLPrefix(encoded, LTLParseOptions{})
	if err != nil {
		t.Fatalf("round-trip %q: %v", encoded, err)
	}
	second, _ := again.Prefix()
	if second != encoded {
		t.Fatalf("noncanonical round trip: %q != %q", second, encoded)
	}
	if infix, err := node.Infix(); err != nil || infix != "(GF(request+~stall)&(readyUdone))" {
		t.Fatalf("infix = %q, %v", infix, err)
	}
	for _, atom := range []string{"Glob", "Finally", "Xray", "Until"} {
		n, e := ParseLTLPrefix(atom, LTLParseOptions{})
		if e != nil || n.Operator != LTLAtom || n.Name != atom {
			t.Fatalf("operator-leading variable %q: %#v %v", atom, n, e)
		}
	}
}

func TestLTLPrefixRejectsAndLimits(t *testing.T) {
	tests := []string{"", ":", ":Q p", "& p", "p q", "\x00", ":U p"}
	for _, input := range tests {
		if node, err := ParseLTLPrefix(input, LTLParseOptions{}); node != nil || err == nil {
			t.Fatalf("accepted %q: %#v", input, node)
		}
	}
	if n, err := ParseLTLPrefix("& p q", LTLParseOptions{MaxBytes: 4}); n != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %#v %v", n, err)
	}
	if n, err := ParseLTLPrefix("& p q", LTLParseOptions{MaxNodes: 2}); n != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit: %#v %v", n, err)
	}
	if n, err := ParseLTLPrefix("!!!p", LTLParseOptions{MaxDepth: 3}); n != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("depth limit: %#v %v", n, err)
	}
	if n, err := ParseLTLPrefix("p", LTLParseOptions{Stop: func() bool { return true }}); n != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation: %#v %v", n, err)
	}
	if n, err := ParseLTLPrefix("p", LTLParseOptions{MaxNodes: -1}); n != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("bad limits: %#v %v", n, err)
	}
	invalid := &LTLNode{Operator: LTLNot}
	if err := invalid.Validate(LTLParseOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("invalid tree: %v", err)
	}
	cycle := &LTLNode{Operator: LTLNot}
	cycle.Left = cycle
	if err := cycle.Validate(LTLParseOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("cyclic tree: %v", err)
	}
	if cycle.IsNonTemporal() || cycle.IsGFWellFormed() {
		t.Fatal("cyclic tree accepted by classifiers")
	}
}

func TestLTLFiniteTraceSemantics(t *testing.T) {
	trace := []LTLValuation{
		{"p": true, "q": false},
		{"p": true, "q": false},
		{"p": false, "q": true},
	}
	for _, test := range []struct {
		formula string
		frame   int
		want    bool
	}{
		{":G p", 0, false}, {":F q", 0, true}, {":X q", 1, true},
		{":X q", 2, false}, {":U p q", 0, true}, {":U q p", 2, false},
		{"& p !q", 0, true}, {"+ false q", 2, true},
	} {
		node, err := ParseLTLPrefix(test.formula, LTLParseOptions{})
		if err != nil {
			t.Fatal(err)
		}
		got, err := EvaluateLTL(node, trace, test.frame, LTLParseOptions{})
		if err != nil || got != test.want {
			t.Fatalf("%q frame %d = %v, %v; want %v", test.formula, test.frame, got, err, test.want)
		}
	}
	if _, err := EvaluateLTL(&LTLNode{Operator: LTLAtom, Name: "p"}, trace, 3, LTLParseOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad frame: %v", err)
	}
}

func TestLTLDatabaseSignalsAndCombinationalAIG(t *testing.T) {
	input := []string{"p", ":G :F q"}
	database, err := NewLTLDatabase(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = "changed"
	copyOf := database.Formulas()
	copyOf[1] = "changed"
	if got := database.Formulas(); got[0] != "p" || got[1] != ":G :F q" {
		t.Fatalf("database aliases caller: %v", got)
	}
	node, _ := ParseLTLPrefix("+ & a b !c", LTLParseOptions{})
	if err := CheckLTLSignalNames(node, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	if err := CheckLTLSignalNames(node, []string{"a", "b"}); err == nil {
		t.Fatal("missing signal accepted")
	}
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	c, _ := graph.CreateCI()
	root, err := BuildLTLCombinational(graph, node, []LTLSignal{{"a", a}, {"b", b}, {"c", c}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		values := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		outputs, err := graph.Evaluate(values)
		want := (values[0] && values[1]) || !values[2]
		if err != nil || len(outputs) != 1 || outputs[0] != want {
			t.Fatalf("assignment %03b: %v %v want %v", bits, outputs, err, want)
		}
	}
	if _, err := BuildLTLCombinational(graph, node, []LTLSignal{{"a", a}}); err == nil {
		t.Fatal("missing binding accepted")
	}
	temporal, _ := ParseLTLPrefix(":F a", LTLParseOptions{})
	if _, err := BuildLTLCombinational(graph, temporal, []LTLSignal{{"a", a}}); !errors.Is(err, ErrDimension) {
		t.Fatalf("temporal builder: %v", err)
	}
}

func ltlOracle(node *LTLNode, trace []LTLValuation, at int) bool {
	switch node.Operator {
	case LTLAtom:
		return node.Name == "true" || node.Name != "false" && trace[at][node.Name]
	case LTLNot:
		return !ltlOracle(node.Left, trace, at)
	case LTLAnd:
		return ltlOracle(node.Left, trace, at) && ltlOracle(node.Right, trace, at)
	case LTLOr:
		return ltlOracle(node.Left, trace, at) || ltlOracle(node.Right, trace, at)
	case LTLImplies:
		return !ltlOracle(node.Left, trace, at) || ltlOracle(node.Right, trace, at)
	case LTLNext:
		return at+1 < len(trace) && ltlOracle(node.Left, trace, at+1)
	case LTLGlobally:
		for i := at; i < len(trace); i++ {
			if !ltlOracle(node.Left, trace, i) {
				return false
			}
		}
		return true
	case LTLEventually:
		for i := at; i < len(trace); i++ {
			if ltlOracle(node.Left, trace, i) {
				return true
			}
		}
		return false
	case LTLUntil:
		for i := at; i < len(trace); i++ {
			if ltlOracle(node.Right, trace, i) {
				return true
			}
			if !ltlOracle(node.Left, trace, i) {
				return false
			}
		}
	}
	return false
}

func randomLTLNode(random *rand.Rand, depth int) *LTLNode {
	if depth == 0 || random.Intn(4) == 0 {
		return &LTLNode{Operator: LTLAtom, Name: []string{"a", "b", "c", "true", "false"}[random.Intn(5)]}
	}
	operator := LTLOperator(random.Intn(int(LTLAtom)))
	node := &LTLNode{Operator: operator, Left: randomLTLNode(random, depth-1)}
	if operator == LTLAnd || operator == LTLOr || operator == LTLImplies || operator == LTLUntil {
		node.Right = randomLTLNode(random, depth-1)
	}
	return node
}

func TestLTLIndependentModel5000(t *testing.T) {
	random := rand.New(rand.NewSource(0x4c544c))
	for iteration := 0; iteration < 5000; iteration++ {
		node := randomLTLNode(random, 4)
		trace := make([]LTLValuation, 1+random.Intn(6))
		for frame := range trace {
			trace[frame] = LTLValuation{"a": random.Intn(2) != 0, "b": random.Intn(2) != 0, "c": random.Intn(2) != 0}
		}
		at := random.Intn(len(trace))
		got, err := EvaluateLTL(node, trace, at, LTLParseOptions{MaxNodes: 100})
		if err != nil {
			t.Fatalf("iteration %d: %v", iteration, err)
		}
		if want := ltlOracle(node, trace, at); got != want {
			t.Fatalf("iteration %d got %v want %v", iteration, got, want)
		}
	}
}

func FuzzLTLPrefix(f *testing.F) {
	for _, seed := range []string{"p", "& p q", ":G :F p", ":U p q", "+ & a b !c", "", ":", "p q"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 2048 {
			t.Skip()
		}
		node, err := ParseLTLPrefix(input, LTLParseOptions{MaxBytes: 2048, MaxNodes: 256, MaxDepth: 128})
		if err != nil {
			return
		}
		encoded, err := node.Prefix()
		if err != nil {
			t.Fatal(err)
		}
		again, err := ParseLTLPrefix(encoded, LTLParseOptions{MaxBytes: 4096, MaxNodes: 256, MaxDepth: 128})
		if err != nil {
			t.Fatalf("encoded %q from %q: %v", encoded, input, err)
		}
		second, err := again.Prefix()
		if err != nil || second != encoded {
			t.Fatalf("noncanonical %q -> %q, %v", encoded, second, err)
		}
		if strings.ContainsRune(encoded, 0) {
			t.Fatal("NUL in canonical encoding")
		}
	})
}
