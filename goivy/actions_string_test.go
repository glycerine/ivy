package goivy

import (
	"strings"
	"testing"
)

// Python action __str__ (ivy_actions.py), as seen in the
// interp.ApplyAction trace of doc/examples/toy_consensus.ivy:
//   - CallAction prints str(self.args[0]), the call's Atom (args joined
//     with ", ");
//   - LocalAction prints 'local ' + locals + ' ' + bracket_action(body),
//     which adds braces only when the body is not a Sequence;
//   - WhileAction prints 'invariant ' + str(inv.args[0]) (the labeled
//     formula, not the assert action) and bracket_action(body).
func TestActionStringsMatchPython(t *testing.T) {
	cfg := NewAstConfig()
	f := NewConst("f", TopS)
	a, b := NewConst("a", Boolean), NewConst("b", Boolean)
	call := NewCallActionOn(NewActionsConfig(), MustApply(f, a, b))
	call.AstCallee = cfg.NewAtom("f", a, b)
	if got := call.String(); got != "call f(a, b)" {
		t.Fatalf("call: %q", got)
	}

	loc := NewConst("loc:0", Boolean)
	local := NewLocalActionOn(NewActionsConfig(), "", loc, NewSequence(call, call))
	if got := local.String(); got != "local loc:0 {call f(a, b); call f(a, b)}" {
		t.Fatalf("local: %q", got)
	}

	inv := NewAssertAction(a)
	inv.LF = cfg.NewLabeledFormula(cfg.NewAtom("asrt1"), a)
	w := NewWhileAction(b, NewSequence(call), inv)
	got := w.String()
	if !strings.Contains(got, "\ninvariant [asrt1] a\n") || !strings.HasSuffix(got, "\n{call f(a, b)}") {
		t.Fatalf("while: %q", got)
	}

	// ChoiceAction: ''.join('if * ' + bracket_action(x) + '\nelse ' ...)
	// + bracket_action(last) (test/comp_liveness1.ivy).
	ch := NewChoiceActionOn(NewActionsConfig(), NewSequence(call), call)
	if got := ch.String(); got != "if * {call f(a, b)}\nelse {call f(a, b)}" {
		t.Fatalf("choice: %q", got)
	}
}
