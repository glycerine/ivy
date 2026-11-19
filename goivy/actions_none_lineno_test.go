package goivy

import "testing"

// Python postfix_action/prefix_action copy self.lineno whenever the
// attribute exists, even when it is None (l2s sets model.init.lineno =
// None), and pretty_lineno prints str(None) = "None" for it. The
// initializer listing then shows "(internal) init" (test/liveness1.ivy).
func TestNoneLinenoPropagatesLikePython(t *testing.T) {
	a := NewAssumeAction(True)
	if got := PrettyActionLineno(a); got != "(internal) " {
		t.Fatalf("no lineno attribute: %q", got)
	}
	a.SetLineno(Location{}) // Python: a.lineno = None
	for _, res := range []ActionsAction{
		PostfixAction(a, []ActionsAction{NewAssumeAction(True)}),
		PrefixAction(a, []ActionsAction{NewAssumeAction(True)}),
	} {
		if !actionHasLocSet(res) {
			t.Fatalf("%T did not copy a None lineno", res)
		}
		if got := PrettyActionLineno(res); got != "None" {
			t.Fatalf("None lineno prints %q", got)
		}
	}
}
