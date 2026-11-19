package goivy

import "testing"

// Python AssertAction.assert_to_assume uses mykind = self.kind if set,
// else type(self). apply_assert_proof sets sga.kind = type(self), so a
// SubgoalAction produced by proving an assert is converted to an assume
// whenever AssertAction is in kinds. goivy ignored the kind and only
// converted for an explicit "subgoal" (test/creport2.ivy).
func TestAssertToAssumeSubgoalUsesItsKind(t *testing.T) {
	sga := NewSubgoalAction(NewConst("p", Boolean))
	sga.SubgoalKind = "assert"
	if _, ok := AssertToAssume(sga, makeKindSet("assert"), nil).(*LogicAssumeAction); !ok {
		t.Fatal("a subgoal of kind assert should become an assume for kinds {assert}")
	}
	if _, ok := AssertToAssume(sga, makeKindSet("require"), nil).(*LogicAssumeAction); ok {
		t.Fatal("a subgoal of kind assert must stay for kinds {require}")
	}
	plain := NewSubgoalAction(NewConst("p", Boolean)) // no kind: type(self)
	if _, ok := AssertToAssume(plain, makeKindSet("assert"), nil).(*LogicAssumeAction); ok {
		t.Fatal("a subgoal without kind is its own class, not assert")
	}
}
