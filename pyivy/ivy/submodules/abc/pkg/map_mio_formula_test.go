package abc

// Stand-alone semantic tests for the MIO formula/SOP core.

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type mioTestExpr struct {
	op          byte
	variable    int
	left, right *mioTestExpr
}

func (expression *mioTestExpr) eval(assignment uint64) bool {
	switch expression.op {
	case 'v':
		return assignment>>uint(expression.variable)&1 != 0
	case '0':
		return false
	case '1':
		return true
	case '!':
		return !expression.left.eval(assignment)
	case '&':
		return expression.left.eval(assignment) && expression.right.eval(assignment)
	case '|':
		return expression.left.eval(assignment) || expression.right.eval(assignment)
	case '^':
		return expression.left.eval(assignment) != expression.right.eval(assignment)
	}
	return false
}

func mioTestCheckFormula(t testing.TB, formula string, names []string, expected *mioTestExpr) {
	t.Helper()
	expression, err := ParseMioFormula(formula, names, MioFormulaLimits{})
	if err != nil {
		t.Fatalf("ParseMioFormula(%q): %v", formula, err)
	}
	truth, err := expression.Truth(MioFormulaLimits{})
	if err != nil {
		t.Fatalf("Truth(%q): %v", formula, err)
	}
	cover, err := expression.SOP(MioFormulaLimits{})
	if err != nil {
		t.Fatalf("SOP(%q): %v", formula, err)
	}
	if _, err = cover.Format(); err != nil {
		t.Fatalf("Format(%q): %v", formula, err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(len(names)); assignment++ {
		want := expected.eval(assignment)
		got, getErr := truth.Get(assignment)
		if getErr != nil || got != want {
			t.Fatalf("truth %q assignment %d = %v, %v; want %v", formula, assignment, got, getErr, want)
		}
		got, getErr = cover.Evaluate(assignment)
		if getErr != nil || got != want {
			t.Fatalf("SOP %q assignment %d = %v, %v; want %v; cover %+v", formula, assignment, got, getErr, want, cover)
		}
	}
}

func TestMioFormulaSyntaxTruthAndSOP(t *testing.T) {
	variables := func(index int) *mioTestExpr { return &mioTestExpr{op: 'v', variable: index} }
	not := func(value *mioTestExpr) *mioTestExpr { return &mioTestExpr{op: '!', left: value} }
	binary := func(op byte, left, right *mioTestExpr) *mioTestExpr {
		return &mioTestExpr{op: op, left: left, right: right}
	}
	a, b, c, d, e := variables(0), variables(1), variables(2), variables(3), variables(4)
	rows := []struct {
		formula string
		names   []string
		want    *mioTestExpr
	}{
		{"0", nil, &mioTestExpr{op: '0'}},
		{"1", nil, &mioTestExpr{op: '1'}},
		{"a", []string{"a"}, a},
		{"!a'", []string{"a"}, a},
		{"a b", []string{"a", "b"}, binary('&', a, b)},
		{"a(!b)", []string{"a", "b"}, binary('&', a, not(b))},
		{"a+b*c", []string{"a", "b", "c"}, binary('|', a, binary('&', b, c))},
		{"a^b+c", []string{"a", "b", "c"}, binary('|', binary('^', a, b), c)},
		{"a|b^c&d", []string{"a", "b", "c", "d"}, binary('|', a, binary('^', b, binary('&', c, d)))},
		{"((a+b')^(c&!d))*e", []string{"a", "b", "c", "d", "e"}, binary('&', binary('^', binary('|', a, not(b)), binary('&', c, not(d))), e)},
	}
	for _, row := range rows {
		t.Run(strings.NewReplacer("!", "n", "'", "p", "|", "o", "&", "a", "^", "x", "+", "O", "*", "A", " ", "_").Replace(row.formula), func(t *testing.T) {
			mioTestCheckFormula(t, row.formula, row.names, row.want)
		})
	}
}

func TestMioExpressionNativeShapeAndText(t *testing.T) {
	expression, err := ParseMioFormula("a^b", []string{"a", "b"}, MioFormulaLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if expression.NodeCount() != 3 || expression.Root() != MioExprLit(2*(2+2)+1) {
		t.Fatalf("XOR expression shape = %d nodes, root %d", expression.NodeCount(), expression.Root())
	}
	text, err := expression.Text()
	if err != nil || !strings.Contains(text, " 2 = !a &  b\n") || !strings.HasSuffix(text, " F = !4\n") {
		t.Fatalf("expression text %q, %v", text, err)
	}
	truth, err := expression.Truth6()
	if err != nil || truth != 0x6666666666666666 {
		t.Fatalf("XOR truth = %016x, %v", truth, err)
	}
	if expression.VariableNames()[0] = "changed"; expression.VariableNames()[0] != "a" {
		t.Fatal("variable-name snapshot aliases expression")
	}
	nodes := expression.Nodes()
	nodes[0].First = MioExprConstant(true)
	if expression.Nodes()[0] == nodes[0] {
		t.Fatal("node snapshot aliases expression")
	}
}

func TestMioSOPPrimitivesAndFormatting(t *testing.T) {
	zero, err := MioSOPLiteral(1, false)
	if err != nil {
		t.Fatal(err)
	}
	one, _ := MioSOPLiteral(1, true)
	if zero|one != 3<<2 || !(zero | one).Empty() || !MioSOPCube(0).Contains(zero) {
		t.Fatalf("cube operations: zero=%x one=%x", zero, one)
	}
	cover := MioSOP{Variables: 2, Cubes: []MioSOPCube{zero}, OnValue: true}
	text, err := cover.Format()
	if err != nil || text != "-0 1\n" {
		t.Fatalf("cover format %q, %v", text, err)
	}
	for _, row := range []struct {
		cover MioSOP
		text  string
	}{
		{MioSOP{Variables: 2, OnValue: true}, " 0\n"},
		{MioSOP{Variables: 2, OnValue: false}, " 1\n"},
		{MioSOP{Variables: 2, Cubes: []MioSOPCube{0}, OnValue: true}, " 1\n"},
		{MioSOP{Variables: 2, Cubes: []MioSOPCube{0}, OnValue: false}, " 0\n"},
	} {
		got, formatErr := row.cover.Format()
		if formatErr != nil || got != row.text {
			t.Fatalf("Format(%+v) = %q, %v; want %q", row.cover, got, formatErr, row.text)
		}
	}
}

func TestMioFormulaNamesGateAndLibrary(t *testing.T) {
	names, err := MioCollectFormulaNames("!slow + fast*slow + aux'", 4)
	if err != nil || fmt.Sprint(names) != "[slow fast aux]" {
		t.Fatalf("names = %v, %v", names, err)
	}
	gate, err := CompileMioFormulaGate(MioFormulaGate{
		Name: "muxlike", Formula: "slow+fast*aux",
		Pins: []MioFormulaPin{{Name: "*", MaxDelay: 2.5}},
	}, MioFormulaLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(gate.Expression.VariableNames()) != "[slow fast aux]" || len(gate.Pins) != 3 || gate.MaxDelay != 2.5 {
		t.Fatalf("compiled generic gate = %+v, names %v", gate, gate.Expression.VariableNames())
	}
	if _, err := CompileMioFormulaGate(MioFormulaGate{
		Name: "ambiguous", Formula: "1a", Pins: []MioFormulaPin{{Name: "*"}},
	}, MioFormulaLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("ambiguous leading constant token: %v", err)
	}
	explicit, err := CompileMioFormulaGate(MioFormulaGate{
		Name: "ordered", Formula: "a+b",
		Pins: []MioFormulaPin{{Name: "b", MaxDelay: 1}, {Name: "a", MaxDelay: 3}},
	}, MioFormulaLimits{})
	if err != nil || fmt.Sprint(explicit.Expression.VariableNames()) != "[b a]" || explicit.MaxDelay != 3 {
		t.Fatalf("explicit gate = %+v, %v", explicit, err)
	}
	library, err := CompileMioFormulaLibrary([]MioFormulaGate{
		{Name: "zero", Formula: "0"},
		{Name: "one", Formula: "1"},
		explicit,
	}, MioFormulaLimits{})
	if err != nil || library.ConstantZero != 0 || library.ConstantOne != 1 || len(library.Gates) != 3 {
		t.Fatalf("library = %+v, %v", library, err)
	}
	if explicit.Expression == library.Gates[2].Expression {
		t.Fatal("library retained caller's derived expression")
	}
}

func TestMioFormulaErrorsLimitsCancellationAndAtomicity(t *testing.T) {
	invalid := []string{"", "(", ")", "a+", "+a", "a**a", "a(b", "a)'", "unknown", "a\x00b"}
	for _, formula := range invalid {
		if expression, err := ParseMioFormula(formula, []string{"a", "b"}, MioFormulaLimits{}); expression != nil || err == nil {
			t.Fatalf("invalid formula %q returned %+v, %v", formula, expression, err)
		}
	}
	if expression, err := ParseMioFormula("a&a", []string{"a", "a"}, MioFormulaLimits{}); expression != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("duplicate variables = %+v, %v", expression, err)
	}
	for _, limits := range []MioFormulaLimits{
		{MaxVariables: 1}, {MaxNodes: 1}, {MaxTokens: 1}, {MaxDepth: 1},
	} {
		if expression, err := ParseMioFormula("((a^b)+a)", []string{"a", "b"}, limits); expression != nil || !errors.Is(err, ErrLimit) {
			t.Fatalf("limits %+v returned %+v, %v", limits, expression, err)
		}
	}
	if expression, err := ParseMioFormula("a", []string{"a"}, MioFormulaLimits{Stop: func() bool { return true }}); expression != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled parse = %+v, %v", expression, err)
	}
	expression, err := ParseMioFormula("a+b", []string{"a", "b"}, MioFormulaLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if table, truthErr := expression.Truth(MioFormulaLimits{Stop: func() bool { return true }}); table != nil || !errors.Is(truthErr, ErrCancelled) {
		t.Fatalf("cancelled truth = %+v, %v", table, truthErr)
	}
	if cover, sopErr := expression.SOP(MioFormulaLimits{MaxCubes: 1}); cover.Cubes != nil || !errors.Is(sopErr, ErrLimit) {
		t.Fatalf("limited SOP = %+v, %v", cover, sopErr)
	}
	original := MioFormulaGate{Name: "bad", Formula: "a+", Pins: []MioFormulaPin{{Name: "a"}}}
	if compiled, compileErr := CompileMioFormulaGate(original, MioFormulaLimits{}); compiled.Name != "" || compileErr == nil || original.Expression != nil {
		t.Fatalf("atomic gate compile = %+v, %v; original %+v", compiled, compileErr, original)
	}
}

func mioTestGenerated(data []byte) (string, *mioTestExpr) {
	leaves := []*mioTestExpr{
		{op: 'v', variable: 0}, {op: 'v', variable: 1}, {op: 'v', variable: 2}, {op: 'v', variable: 3},
		{op: '0'}, {op: '1'},
	}
	texts := []string{"a", "b", "c", "d", "0", "1"}
	if len(data) == 0 {
		return texts[0], leaves[0]
	}
	current := int(data[0]) % len(leaves)
	expression := leaves[current]
	text := texts[current]
	for index := 1; index < len(data) && index < 25; index++ {
		value := data[index]
		if value&7 == 0 {
			expression = &mioTestExpr{op: '!', left: expression}
			if value&8 != 0 {
				text = "(" + text + ")'"
			} else {
				text = "!(" + text + ")"
			}
			continue
		}
		rightIndex := int(value>>3) % len(leaves)
		right := leaves[rightIndex]
		rightText := texts[rightIndex]
		operations := []byte{'&', '|', '^'}
		op := operations[int(value)%len(operations)]
		expression = &mioTestExpr{op: op, left: expression, right: right}
		if op == '&' && value&4 != 0 {
			text = "(" + text + ")(" + rightText + ")"
		} else {
			text = "(" + text + string(op) + rightText + ")"
		}
	}
	return text, expression
}

func runMioFormulaModel(data []byte) error {
	formula, expected := mioTestGenerated(data)
	expression, err := ParseMioFormula(formula, []string{"a", "b", "c", "d"}, MioFormulaLimits{MaxNodes: 256, MaxTokens: 512, MaxDepth: 64, MaxCubes: 4096})
	if err != nil {
		return fmt.Errorf("generated parse %q: %w", formula, err)
	}
	truth, err := expression.Truth(MioFormulaLimits{})
	if err != nil {
		return err
	}
	cover, err := expression.SOP(MioFormulaLimits{MaxCubes: 4096})
	if err != nil {
		return err
	}
	for assignment := uint64(0); assignment < 16; assignment++ {
		want := expected.eval(assignment)
		got, _ := truth.Get(assignment)
		covered, coverErr := cover.Evaluate(assignment)
		if coverErr != nil || got != want || covered != want {
			return fmt.Errorf("formula %q assignment %d: truth=%v SOP=%v/%v want=%v", formula, assignment, got, covered, coverErr, want)
		}
	}
	return nil
}

func TestMioFormulaIndependentModelCampaign(t *testing.T) {
	state := uint64(0x6a09e667f3bcc909)
	for iteration := 0; iteration < 5000; iteration++ {
		data := make([]byte, 1+iteration%17)
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := runMioFormulaModel(data); err != nil {
			t.Fatalf("iteration %d: %v", iteration, err)
		}
	}
}

func FuzzMioFormula(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3}, {255, 0, 17, 88}, []byte("formula-model")} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		if err := runMioFormulaModel(data); err != nil {
			t.Fatal(err)
		}
		// Raw bytes exercise rejection and all delimiter/EOF boundaries. A
		// successful parse must still produce self-consistent truth and SOP.
		raw := strings.Map(func(value rune) rune {
			if value > 127 {
				return ' '
			}
			return value
		}, string(data))
		expression, err := ParseMioFormula(raw, []string{"a", "b", "c", "d"}, MioFormulaLimits{MaxNodes: 128, MaxTokens: 256, MaxDepth: 32, MaxCubes: 1024})
		if err != nil {
			return
		}
		if checkErr := expression.Check(); checkErr != nil {
			t.Fatal(checkErr)
		}
		truth, truthErr := expression.Truth(MioFormulaLimits{})
		cover, coverErr := expression.SOP(MioFormulaLimits{MaxCubes: 1024})
		if truthErr != nil || coverErr != nil {
			t.Fatalf("successful parse %q failed derivation: %v, %v", raw, truthErr, coverErr)
		}
		for assignment := uint64(0); assignment < 16; assignment++ {
			value, _ := truth.Get(assignment)
			covered, evalErr := cover.Evaluate(assignment)
			if evalErr != nil || value != covered {
				t.Fatalf("raw %q assignment %d: truth=%v SOP=%v/%v", raw, assignment, value, covered, evalErr)
			}
		}
	})
}
