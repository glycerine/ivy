package goivy

import (
	"testing"
)

func mkWhileActionForTest(sortName string) *LogicWhileAction {
	sortT := actionsMkSort(sortName)
	ltSym := NewConst("<", LogicRelationSort([]Sort{sortT, sortT}))
	xSym := NewConst("x", sortT)
	boundSym := NewConst("bound", sortT)
	cond, _ := NewApply(ltSym, xSym, boundSym)
	body := NewAssignAction(xSym, xSym)
	return NewWhileAction(cond, body)
}

func constCard(n int) CardFunc {
	return func(s Sort) int { return n }
}

func TestUnrollLoops_Nil(t *testing.T) {
	result := UnrollLoops(nil, constCard(3))
	if result != nil {
		t.Errorf("expected nil, got %T", result)
	}
}

func TestUnrollLoops_NoWhile(t *testing.T) {
	act := NewAssumeAction(True)
	result := UnrollLoops(act, constCard(3))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.(*LogicAssumeAction); !ok {
		t.Errorf("expected *LogicAssumeAction, got %T", result)
	}
}

func TestUnrollLoops_SimpleWhile(t *testing.T) {
	wa := mkWhileActionForTest("T")
	result := UnrollLoops(wa, constCard(3))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	top, ok := result.(*LogicIfAction)
	if !ok {
		t.Fatalf("expected *LogicIfAction at top, got %T", result)
	}
	depth := countIfNesting(top)
	// card=3 means 3 iterations + 1 base case = 4 nested IfActions
	if depth != 4 {
		t.Errorf("expected 4 nested IfActions, got %d", depth)
	}
}

func TestUnrollLoops_WhileInSequence(t *testing.T) {
	assume1 := NewAssumeAction(True)
	wa := mkWhileActionForTest("T")
	assume2 := NewAssumeAction(True)
	seq := NewSequence(assume1, wa, assume2)

	result := UnrollLoops(seq, constCard(2))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	s, ok := result.(*LogicSequence)
	if !ok {
		t.Fatalf("expected *LogicSequence, got %T", result)
	}
	if len(s.Elems) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(s.Elems))
	}
	if _, ok := s.Elems[0].(*LogicAssumeAction); !ok {
		t.Errorf("elem 0: expected *LogicAssumeAction, got %T", s.Elems[0])
	}
	if _, ok := s.Elems[1].(*LogicIfAction); !ok {
		t.Errorf("elem 1: expected *LogicIfAction (unrolled while), got %T", s.Elems[1])
	}
	if _, ok := s.Elems[2].(*LogicAssumeAction); !ok {
		t.Errorf("elem 2: expected *LogicAssumeAction, got %T", s.Elems[2])
	}
}

func TestUnrollLoops_WhileInIf(t *testing.T) {
	wa := mkWhileActionForTest("T")
	assume := NewAssumeAction(True)
	ifAct := NewIfAction(True, wa, assume)

	result := UnrollLoops(ifAct, constCard(2))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	top, ok := result.(*LogicIfAction)
	if !ok {
		t.Fatalf("expected *LogicIfAction, got %T", result)
	}
	// The then-branch should now be an IfAction (unrolled while), not a WhileAction
	if _, ok := top.ThenBody.(*LogicIfAction); !ok {
		t.Errorf("then-branch: expected *LogicIfAction (unrolled while), got %T", top.ThenBody)
	}
	// The else-branch should still be an AssumeAction
	if _, ok := top.ElseBody.(*LogicAssumeAction); !ok {
		t.Errorf("else-branch: expected *LogicAssumeAction, got %T", top.ElseBody)
	}
}

func TestUnrollLoops_NestedWhile(t *testing.T) {
	inner := mkWhileActionForTest("T")
	outer := NewWhileAction(inner.Cond, inner)

	result := UnrollLoops(outer, constCard(2))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Both loops should be unrolled; no WhileActions remain
	if containsWhile(result) {
		t.Error("unrolled result should not contain any WhileAction")
	}
}

func TestUnrollLoops_NotEqCondition(t *testing.T) {
	sortT := actionsMkSort("T")
	xSym := NewConst("x", sortT)
	boundSym := NewConst("bound", sortT)
	eq := &Eq{T1: xSym, T2: boundSym}
	cond := &LogicNot{Body: eq}

	body := NewSequence()
	wa := NewWhileAction(cond, body)

	result := UnrollLoops(wa, constCard(2))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.(*LogicIfAction); !ok {
		t.Errorf("expected *LogicIfAction, got %T", result)
	}
}

func TestUnrollLoops_AndCondition(t *testing.T) {
	sortT := actionsMkSort("T")
	ltSym := NewConst("<", LogicRelationSort([]Sort{sortT, sortT}))
	xSym := NewConst("x", sortT)
	boundSym := NewConst("bound", sortT)
	ltCond, _ := NewApply(ltSym, xSym, boundSym)
	flag := NewConst("flag", Boolean)
	andCond := &LogicAnd{Terms: []Expr{ltCond, flag}}

	wa := NewWhileAction(andCond, NewSequence())

	result := UnrollLoops(wa, constCard(2))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result.(*LogicIfAction); !ok {
		t.Errorf("expected *LogicIfAction, got %T", result)
	}
}

func TestUnrollLoops_UnknownCondition(t *testing.T) {
	flag := NewConst("flag", Boolean)
	wa := NewWhileAction(flag, NewSequence())

	var receivedSort Sort
	card := CardFunc(func(s Sort) int {
		receivedSort = s
		return 1
	})

	result := UnrollLoops(wa, card)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if receivedSort != nil {
		t.Errorf("expected nil sort for unknown condition, got %v", receivedSort)
	}
}

func TestUnrollLoops_FormalsPreserved(t *testing.T) {
	wa := mkWhileActionForTest("T")
	params := []*Const{actionsMkConst("p")}
	returns := []*Const{actionsMkConst("r")}
	wa.SetFormalParams(params)
	wa.SetFormalReturns(returns)

	result := UnrollLoops(wa, constCard(2))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	fp := result.(ActionsAction).GetFormalParams()
	fr := result.(ActionsAction).GetFormalReturns()
	if len(fp) != 1 || fp[0].Name != "p" {
		t.Errorf("formal params not preserved: %v", fp)
	}
	if len(fr) != 1 || fr[0].Name != "r" {
		t.Errorf("formal returns not preserved: %v", fr)
	}
}

func TestUnrollLoops_PanicsOnLargeCard(t *testing.T) {
	wa := mkWhileActionForTest("T")
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for cardinality > 100")
		}
	}()
	UnrollLoops(wa, constCard(200))
}

func TestUnrollLoops_PanicsOnNegativeCard(t *testing.T) {
	wa := mkWhileActionForTest("T")
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for negative cardinality")
		}
	}()
	UnrollLoops(wa, constCard(-1))
}

func TestUnrollLoops_VerifyNesting(t *testing.T) {
	wa := mkWhileActionForTest("T")
	result := UnrollLoops(wa, constCard(2))

	// Expected structure for card=2:
	// IfAction(cond,
	//   Sequence(body,
	//     IfAction(cond,
	//       Sequence(body,
	//         IfAction(cond,
	//           AssumeAction(LogicOr{}))))))
	//
	// 3 IfActions total: 2 iterations + 1 base case

	if1, ok := result.(*LogicIfAction)
	if !ok {
		t.Fatalf("level 1: expected *LogicIfAction, got %T", result)
	}
	seq1, ok := if1.ThenBody.(*LogicSequence)
	if !ok {
		t.Fatalf("level 1 then: expected *LogicSequence, got %T", if1.ThenBody)
	}
	if len(seq1.Elems) != 2 {
		t.Fatalf("level 1 seq: expected 2 elements, got %d", len(seq1.Elems))
	}

	if2, ok := seq1.Elems[1].(*LogicIfAction)
	if !ok {
		t.Fatalf("level 2: expected *LogicIfAction, got %T", seq1.Elems[1])
	}
	seq2, ok := if2.ThenBody.(*LogicSequence)
	if !ok {
		t.Fatalf("level 2 then: expected *LogicSequence, got %T", if2.ThenBody)
	}
	if len(seq2.Elems) != 2 {
		t.Fatalf("level 2 seq: expected 2 elements, got %d", len(seq2.Elems))
	}

	if3, ok := seq2.Elems[1].(*LogicIfAction)
	if !ok {
		t.Fatalf("base case: expected *LogicIfAction, got %T", seq2.Elems[1])
	}
	assumeAct, ok := if3.ThenBody.(*LogicAssumeAction)
	if !ok {
		t.Fatalf("base case then: expected *LogicAssumeAction, got %T", if3.ThenBody)
	}
	if _, ok := assumeAct.Formula.(*LogicOr); !ok {
		t.Errorf("base case assume: expected *lg.Or (false), got %T", assumeAct.Formula)
	}
}

func TestUnrollLoops_ChoiceAction(t *testing.T) {
	cfg := NewActionsConfig()
	wa := mkWhileActionForTest("T")
	assume := NewAssumeAction(True)
	choice := NewChoiceActionOn(cfg, wa, assume)

	result := UnrollLoops(choice, constCard(2))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ch, ok := result.(*LogicChoiceAction)
	if !ok {
		t.Fatalf("expected *LogicChoiceAction, got %T", result)
	}
	if len(ch.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(ch.Branches))
	}
	if _, ok := ch.Branches[0].(*LogicIfAction); !ok {
		t.Errorf("branch 0: expected *LogicIfAction (unrolled while), got %T", ch.Branches[0])
	}
	if _, ok := ch.Branches[1].(*LogicAssumeAction); !ok {
		t.Errorf("branch 1: expected *LogicAssumeAction, got %T", ch.Branches[1])
	}
}

func TestUnrollLoops_ZeroCard(t *testing.T) {
	wa := mkWhileActionForTest("T")
	result := UnrollLoops(wa, constCard(0))
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// card=0 means just the base case: if(cond, assume(false))
	ifAct, ok := result.(*LogicIfAction)
	if !ok {
		t.Fatalf("expected *LogicIfAction, got %T", result)
	}
	if _, ok := ifAct.ThenBody.(*LogicAssumeAction); !ok {
		t.Errorf("then-body: expected *LogicAssumeAction, got %T", ifAct.ThenBody)
	}
}

// --- helpers ---

func countIfNesting(act ActionsAction) int {
	ifAct, ok := act.(*LogicIfAction)
	if !ok {
		return 0
	}
	count := 1
	if seq, ok := ifAct.ThenBody.(*LogicSequence); ok {
		for _, e := range seq.Elems {
			if child, ok := e.(ActionsAction); ok {
				n := countIfNesting(child)
				if n > 0 {
					count += n
				}
			}
		}
	} else if child, ok := ifAct.ThenBody.(ActionsAction); ok {
		count += countIfNesting(child)
	}
	return count
}

func containsWhile(act ActionsAction) bool {
	if _, ok := act.(*LogicWhileAction); ok {
		return true
	}
	for _, arg := range act.ActionArgs() {
		if child, ok := arg.(ActionsAction); ok {
			if containsWhile(child) {
				return true
			}
		}
	}
	return false
}
