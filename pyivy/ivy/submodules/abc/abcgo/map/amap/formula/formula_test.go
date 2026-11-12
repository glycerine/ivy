package formula

import (
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
)

func TestParsePrecedenceAndLibertyOperators(t *testing.T) {
	// AND binds tighter than XOR, which binds tighter than OR.
	e, err := Parse("A + B ^ C * D")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := e.String(), "(A + (B ^ (C * D)))"; got != want {
		t.Fatalf("tree = %s, want %s", got, want)
	}
	if got, want := e.Variables(), []string{"A", "B", "C", "D"}; !equalStrings(got, want) {
		t.Fatalf("variables = %v, want %v", got, want)
	}
	for _, tc := range []struct {
		vars map[string]bool
		want bool
	}{
		{map[string]bool{"A": false, "B": true, "C": true, "D": true}, false},
		{map[string]bool{"A": false, "B": true, "C": true, "D": false}, true},
		{map[string]bool{"A": true, "B": false, "C": true, "D": true}, true},
	} {
		got, err := e.Eval(tc.vars)
		if err != nil || got != tc.want {
			t.Fatalf("Eval(%v) = %v, %v; want %v", tc.vars, got, err, tc.want)
		}
	}
}

func TestImplicitAndPostfixNegation(t *testing.T) {
	e, err := Parse("!(A B') + C")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		vars map[string]bool
		want bool
	}{
		{map[string]bool{"A": true, "B": false, "C": false}, false},
		{map[string]bool{"A": true, "B": true, "C": false}, true},
		{map[string]bool{"A": true, "B": false, "C": true}, true},
	} {
		got, err := e.Eval(tc.vars)
		if err != nil || got != tc.want {
			t.Fatalf("Eval(%v) = %v, %v; want %v", tc.vars, got, err, tc.want)
		}
	}
}

func TestToHopMatchesExpressionTruthTable(t *testing.T) {
	manager := hop.Hop_ManStart()
	defer hop.Hop_ManStop(manager)
	e, err := Parse("(A + B) * (A' ^ C)")
	if err != nil {
		t.Fatal(err)
	}
	vars := map[string]*hop.Hop_Obj_t{
		"A": hop.Hop_IthVar(manager, 0),
		"B": hop.Hop_IthVar(manager, 1),
		"C": hop.Hop_IthVar(manager, 2),
	}
	root, err := e.ToHop(manager, vars)
	if err != nil {
		t.Fatal(err)
	}
	truth := hop.Hop_ManConvertAigToTruth(manager, root, 3, nil, 0)[0]
	for assignment := 0; assignment < 8; assignment++ {
		values := map[string]bool{
			"A": assignment&1 != 0,
			"B": assignment&2 != 0,
			"C": assignment&4 != 0,
		}
		want, err := e.Eval(values)
		if err != nil {
			t.Fatal(err)
		}
		got := truth&(1<<uint(assignment)) != 0
		if got != want {
			t.Fatalf("assignment %03b: Hop=%v, expression=%v", assignment, got, want)
		}
	}
}

func TestParseErrorsAreActionable(t *testing.T) {
	for _, input := range []string{"", "A +", "(A", "A )", "A ? B", "!"} {
		_, err := Parse(input)
		if err == nil || !strings.Contains(err.Error(), "formula at byte") {
			t.Errorf("Parse(%q) error = %v, want byte-positioned error", input, err)
		}
	}
	if _, err := (&Expr{kind: 'v', name: "missing"}).Eval(nil); err == nil {
		t.Fatal("missing variable evaluated without error")
	}
}

func equalStrings(a, b []string) bool {
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
