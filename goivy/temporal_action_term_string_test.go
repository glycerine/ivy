package goivy

import "testing"

// Python ActionTerm.__str__ prints parameters with typed_sym_to_str,
// keeping the fml: prefix (test/liveness_mc1.ivy).
func TestActionTermStringKeepsFmlPrefix(t *testing.T) {
	s := &UninterpretedSort{Name: "t"}
	at := &ActionTerm{Inputs: []*Const{NewConst("fml:x", s)}, Stmt: NewSequence()}
	if got := at.String(); got != "action(fml:x:t){{}}" {
		t.Fatalf("%q", got)
	}
}
