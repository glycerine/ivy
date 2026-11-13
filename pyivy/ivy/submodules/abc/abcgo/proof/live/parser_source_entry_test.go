package live

import "testing"

// The compiled original-ABC parser fixture prints "8:>signal" and "8:alpha".
func TestSourceParserTreatsImplicationMarkerAsNameAndIgnoresTail(t *testing.T) {
	resetGlobalVar()
	defer resetGlobalVar()
	node := parseFormulaCreateAST(">signal trailing")
	if node == nil || node.Type != BOOL || node.Name != ">signal" {
		t.Fatalf("source AST %+v", node)
	}
	node = readLtlFormula("alpha beta")
	if node == nil || node.Type != BOOL || node.Name != "alpha" {
		t.Fatalf("source trailing-token AST %+v", node)
	}
	if native, err := ParseFormula(">signal trailing"); err != nil || native.Type != IMPLY {
		t.Fatalf("native parser should recognize implication: %+v %v", native, err)
	}
	if _, err := ParseFormula("alpha beta"); err == nil {
		t.Fatal("native parser should reject trailing input")
	}
}

func TestSourceLTLWrapperUsesSourceTokenRules(t *testing.T) {
	n := liveFixture(false)
	n.OutputNames[0] = ">signal"
	n.LTLProperties = []string{":G :F >signal trailing"}
	result := LivenessToSafetyTransformationWithLTL(IGNORE_SAFETY_KEEP_LIVENESS_MODE, n, n.Graph, nil, nil, nil, nil)
	if trueOutputs(result) != 1 {
		t.Fatalf("source accepted expression should produce one output, got %d", trueOutputs(result))
	}
}

func TestSourceParserCursorAndReset(t *testing.T) {
	resetGlobalVar()
	defer resetGlobalVar()
	text := "alpha beta"
	first := readLtlFormula(text)
	if first == nil || first.Name != "alpha" || startOfSuffixString != 5 {
		t.Fatalf("first=%+v cursor=%d", first, startOfSuffixString)
	}
	second := readLtlFormula(text)
	if second == nil || second.Name != "beta" || startOfSuffixString != 10 {
		t.Fatalf("second=%+v cursor=%d", second, startOfSuffixString)
	}
	if readLtlFormula(text) != nil {
		t.Fatal("source cursor past last token should return NULL")
	}
	resetGlobalVar()
	if node := parseFormulaCreateAST(text); node == nil || node.Name != "alpha" || startOfSuffixString != 0 {
		t.Fatalf("AST wrapper must reset cursor: %+v %d", node, startOfSuffixString)
	}
}

func TestSourceLTLWrapperRetainsExpressionFromExistingCursor(t *testing.T) {
	resetGlobalVar()
	defer resetGlobalVar()
	if node := readLtlFormula("skip"); node == nil || startOfSuffixString != 4 {
		t.Fatal("could not prime source parser cursor")
	}
	n := liveFixture(false)
	n.OutputNames[0] = "signal"
	n.LTLProperties = []string{"skip:G :F signal trailing"}
	result := LivenessToSafetyTransformationWithLTL(IGNORE_SAFETY_KEEP_LIVENESS_MODE, n, n.Graph, nil, nil, nil, nil)
	if trueOutputs(result) != 1 || startOfSuffixString != 0 {
		t.Fatalf("source AST cursor was not retained/reset: outputs=%d cursor=%d", trueOutputs(result), startOfSuffixString)
	}
}
