package ivylogic

import (
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// mkCond returns a Boolean Symbol for use as a condition in tests.
func mkCond() lg.Expr {
	return lg.NewConst("cond", lg.Boolean)
}

// mkX returns a Boolean Symbol "x" for use in tests.
func mkX() lg.Expr {
	return lg.NewConst("x", lg.Boolean)
}

// mkY returns a Boolean Symbol "y" for use in tests.
func mkY() lg.Expr {
	return lg.NewConst("y", lg.Boolean)
}

func TestSimpIte_TrueThen(t *testing.T) {
	// Ite(cond, true, x) → Or(cond, x)
	cond := mkCond()
	x := mkX()
	result := SimpIte(cond, &lg.And{}, x)
	or, ok := result.(*lg.Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T: %v", result, result)
	}
	if len(or.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(or.Terms))
	}
	if !or.Terms[0].Equal(cond) {
		t.Errorf("expected first term to be cond, got %v", or.Terms[0])
	}
	if !or.Terms[1].Equal(x) {
		t.Errorf("expected second term to be x, got %v", or.Terms[1])
	}
}

func TestSimpIte_FalseThen(t *testing.T) {
	// Ite(cond, false, x) → And(Not(cond), x)
	cond := mkCond()
	x := mkX()
	result := SimpIte(cond, &lg.Or{}, x)
	and, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T: %v", result, result)
	}
	if len(and.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(and.Terms))
	}
	notCond, ok := and.Terms[0].(*lg.Not)
	if !ok {
		t.Fatalf("expected first term to be *lg.Not, got %T", and.Terms[0])
	}
	if !notCond.Body.Equal(cond) {
		t.Errorf("expected Not(cond), got Not(%v)", notCond.Body)
	}
	if !and.Terms[1].Equal(x) {
		t.Errorf("expected second term to be x, got %v", and.Terms[1])
	}
}

func TestSimpIte_TrueElse(t *testing.T) {
	// Ite(cond, x, true) → Or(Not(cond), x)
	cond := mkCond()
	x := mkX()
	result := SimpIte(cond, x, &lg.And{})
	or, ok := result.(*lg.Or)
	if !ok {
		t.Fatalf("expected *lg.Or, got %T: %v", result, result)
	}
	if len(or.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(or.Terms))
	}
	notCond, ok := or.Terms[0].(*lg.Not)
	if !ok {
		t.Fatalf("expected first term to be *lg.Not, got %T", or.Terms[0])
	}
	if !notCond.Body.Equal(cond) {
		t.Errorf("expected Not(cond), got Not(%v)", notCond.Body)
	}
	if !or.Terms[1].Equal(x) {
		t.Errorf("expected second term to be x, got %v", or.Terms[1])
	}
}

func TestSimpIte_FalseElse(t *testing.T) {
	// Ite(cond, x, false) → And(cond, x)
	cond := mkCond()
	x := mkX()
	result := SimpIte(cond, x, &lg.Or{})
	and, ok := result.(*lg.And)
	if !ok {
		t.Fatalf("expected *lg.And, got %T: %v", result, result)
	}
	if len(and.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(and.Terms))
	}
	if !and.Terms[0].Equal(cond) {
		t.Errorf("expected first term to be cond, got %v", and.Terms[0])
	}
	if !and.Terms[1].Equal(x) {
		t.Errorf("expected second term to be x, got %v", and.Terms[1])
	}
}

func TestSimpIte_TrueCond(t *testing.T) {
	// Ite(true, x, y) → x
	x := mkX()
	y := mkY()
	result := SimpIte(&lg.And{}, x, y)
	if !result.Equal(x) {
		t.Errorf("expected x, got %v", result)
	}
}

func TestSimpIte_FalseCond(t *testing.T) {
	// Ite(false, x, y) → y
	x := mkX()
	y := mkY()
	result := SimpIte(&lg.Or{}, x, y)
	if !result.Equal(y) {
		t.Errorf("expected y, got %v", result)
	}
}

func TestSimpIte_EqualBranches(t *testing.T) {
	// Ite(cond, x, x) → x
	cond := mkCond()
	x := mkX()
	x2 := lg.NewConst("x", lg.Boolean)
	result := SimpIte(cond, x, x2)
	if !result.Equal(x) {
		t.Errorf("expected x, got %v", result)
	}
}

func TestSimpIte_NoSimplification(t *testing.T) {
	// Ite(cond, x, y) → Ite(cond, x, y)
	cond := mkCond()
	x := mkX()
	y := mkY()
	result := SimpIte(cond, x, y)
	ite, ok := result.(*lg.Ite)
	if !ok {
		t.Fatalf("expected *lg.Ite, got %T: %v", result, result)
	}
	if !ite.Cond.Equal(cond) {
		t.Errorf("expected cond=%v, got %v", cond, ite.Cond)
	}
	if !ite.Then.Equal(x) {
		t.Errorf("expected then=%v, got %v", x, ite.Then)
	}
	if !ite.Else.Equal(y) {
		t.Errorf("expected else=%v, got %v", y, ite.Else)
	}
}

