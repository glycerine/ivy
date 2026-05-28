package goivy

import (
	"strings"
	"testing"
)

// TestDivergence10UndoInReportInterpOverVar verifies that reportInterpOverVar
// reverses alpha-conversion via varUniq.Undo() before including variable names
// and formulas in error messages. This matches Python ivy_fragment.py lines
// 444 and 446 which call var_uniq.undo(v) and var_uniq.undo(fmla).
func TestDivergence10UndoInReportInterpOverVar(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)

	// Set up varUniq that renames X -> X_a.
	// Uniquifying the same forall-X formula twice: first X->X, second X->X_a.
	c.varUniq = NewVariableUniqifier(nil)
	origX, err := NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}
	fmla1 := &ForAll{Variables: []*LogicVariable{origX}, Body: origX}
	c.varUniq.Uniquify(fmla1)         // X -> X (first pass, no rename needed)
	res2 := c.varUniq.Uniquify(fmla1) // X -> X_a (name collision with first pass)

	// Extract the uniquified variable X_a from the result.
	fa2, ok := res2.(*ForAll)
	if !ok {
		t.Fatalf("expected *lg.ForAll, got %T", res2)
	}
	xA := fa2.Variables[0] // This is X_a
	if xA.Name == "X" {
		t.Fatalf("expected uniquified name != X, got %s", xA.Name)
	}

	// Register X_a as a universal variable at line 99.
	vid := makeVarID(xA)
	c.universallyQuantifiedVars[vid] = xA
	c.universalVarLoc[vid] = Location{Filename: "origin.ivy", Line: 99}

	// Create a strat map entry for X_a.
	node := NewUFNode()
	vk := varKey(xA)
	c.stratMap[vk] = node
	c.stratInfo[vk] = stratEntry{v: xA}

	// Build a formula containing the uniquified variable.
	boolSort, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	p := NewConst("p", boolSort)
	appFmla, err := NewApply(p, xA)
	if err != nil {
		t.Fatal(err)
	}

	// Call reportInterpOverVar and catch the panic.
	var errMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				if fe, ok := r.(*FragmentError); ok {
					errMsg = fe.Message
				} else {
					t.Fatalf("unexpected panic type: %T: %v", r, r)
				}
			}
		}()
		c.reportInterpOverVar(appFmla, Location{Filename: "use.ivy", Line: 42}, node)
	}()

	if errMsg == "" {
		t.Fatal("expected FragmentError panic, got none")
	}

	// The error message should contain the ORIGINAL name "X", not the
	// uniquified name like "X_a".
	if strings.Contains(errMsg, xA.Name) && xA.Name != "X" {
		t.Errorf("error message contains uniquified name %q instead of original name:\n%s", xA.Name, errMsg)
	}
	if !strings.Contains(errMsg, "The quantified variable is") {
		t.Errorf("error message missing 'The quantified variable is' text:\n%s", errMsg)
	}
	if !strings.Contains(errMsg, "use.ivy: line 42:") || !strings.Contains(errMsg, "origin.ivy: line 99:") {
		t.Errorf("error message should include full source locations:\n%s", errMsg)
	}
}

// TestDivergence11SkolemMapInReportArc verifies that reportArc checks
// skolemMap for terms and adds "skolem function defined by:" context,
// matching Python ivy_fragment.py lines 418-420.
func TestDivergence11SkolemMapInReportArc(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)

	// Create an existential variable E.
	eVar, err := NewVariable("E", S)
	if err != nil {
		t.Fatal(err)
	}

	// Create a skolem entry with a known formula and AST node with lineno.
	acfg := NewAstConfig()
	astNode := acfg.NewAtom("witness")
	astNode.SetLineno(Location{Filename: "test.ivy", Line: 10})

	skolemFmla := &ForAll{
		Variables: []*LogicVariable{eVar},
		Body:      eVar,
	}
	c.skolemMap[makeVarID(eVar)] = skolemEntry{
		fmla: skolemFmla,
		ast:  astNode,
	}

	// Build formula Apply(f, E) to make an arc where term = E.
	fSort, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	f := NewConst("f", fSort)
	appFmla, err := NewApply(f, eVar)
	if err != nil {
		t.Fatal(err)
	}

	// Set up a from-node with a sort in stratMap so getNodeSort works.
	fromNode := NewUFNode()
	fromKey := varKey(eVar)
	c.stratMap[fromKey] = fromNode
	c.stratInfo[fromKey] = stratEntry{v: eVar}

	a := arc{
		from:   fromNode,
		to:     NewUFNode(),
		fmla:   appFmla,
		loc:    Location{Line: 5},
		argIdx: 0, // term = E (first arg after function in NodeArgs)
		hasIdx: true,
	}

	result := c.reportArc(a)

	if !strings.Contains(result, "skolem function defined by:") {
		t.Errorf("reportArc should include 'skolem function defined by:' when term is in skolemMap, got:\n%s", result)
	}
	if !strings.Contains(result, "test.ivy: line 10:") {
		t.Errorf("reportArc should include the skolem origin lineno 'test.ivy: line 10:', got:\n%s", result)
	}
}

// TestDivergence11NoSkolemEntry verifies that reportArc does NOT include
// skolem info when the term is not in skolemMap.
func TestDivergence11NoSkolemEntry(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	sig := NewSig()
	sig.AddSort(S)
	c := newChecker(sig, nil)

	// Create a variable that is NOT in skolemMap.
	xVar, err := NewVariable("X", S)
	if err != nil {
		t.Fatal(err)
	}

	fSort, err := NewFunctionSort(S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	f := NewConst("f", fSort)
	appFmla, err := NewApply(f, xVar)
	if err != nil {
		t.Fatal(err)
	}

	fromNode := NewUFNode()
	fromKey := varKey(xVar)
	c.stratMap[fromKey] = fromNode
	c.stratInfo[fromKey] = stratEntry{v: xVar}

	a := arc{
		from:   fromNode,
		to:     NewUFNode(),
		fmla:   appFmla,
		loc:    Location{Line: 5},
		argIdx: 0,
		hasIdx: true,
	}

	result := c.reportArc(a)

	if strings.Contains(result, "skolem function defined by:") {
		t.Errorf("reportArc should NOT include skolem info when term is not in skolemMap, got:\n%s", result)
	}
}
