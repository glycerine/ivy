package logic

import (
	"math/rand"
	"strings"
	"testing"
)

// randomExpr generates a deterministic random Expr tree from a seed.
func randomExpr(seed uint64, maxDepth int) Expr {
	rng := rand.New(rand.NewSource(int64(seed)))
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	Y, _ := NewVariable("Y", S)
	Z, _ := NewVariable("Z", S)
	vars := []*Variable{X, Y, Z}
	fs, _ := NewFunctionSort(S, Boolean)
	sym := NewSymbol("f", fs)
	return randNode(rng, S, vars, sym, maxDepth)
}

func randNode(rng *rand.Rand, s Sort, vars []*Variable, sym *Symbol, depth int) Expr {
	if depth <= 0 {
		return vars[rng.Intn(len(vars))]
	}
	switch rng.Intn(16) {
	case 0: // Variable
		return vars[rng.Intn(len(vars))]
	case 1: // Symbol
		return sym
	case 2: // Apply
		a, _ := NewApply(sym, vars[rng.Intn(len(vars))])
		if a == nil {
			return vars[0]
		}
		return a
	case 3: // Eq
		v1 := vars[rng.Intn(len(vars))]
		v2 := vars[rng.Intn(len(vars))]
		e, _ := NewEq(v1, v2)
		if e == nil {
			return vars[0]
		}
		return e
	case 4: // Not
		body := randBoolNode(rng, s, vars, sym, depth-1)
		n, _ := NewNot(body)
		if n == nil {
			return body
		}
		return n
	case 5: // And
		n := rng.Intn(3)
		terms := make([]Expr, n)
		for i := range terms {
			terms[i] = randBoolNode(rng, s, vars, sym, depth-1)
		}
		a, _ := NewAnd(terms...)
		if a == nil {
			return &And{}
		}
		return a
	case 6: // Or
		n := rng.Intn(3)
		terms := make([]Expr, n)
		for i := range terms {
			terms[i] = randBoolNode(rng, s, vars, sym, depth-1)
		}
		o, _ := NewOr(terms...)
		if o == nil {
			return &Or{}
		}
		return o
	case 7: // Implies
		t1 := randBoolNode(rng, s, vars, sym, depth-1)
		t2 := randBoolNode(rng, s, vars, sym, depth-1)
		imp, _ := NewImplies(t1, t2)
		if imp == nil {
			return t1
		}
		return imp
	case 8: // Iff
		t1 := randBoolNode(rng, s, vars, sym, depth-1)
		t2 := randBoolNode(rng, s, vars, sym, depth-1)
		iff, _ := NewIff(t1, t2)
		if iff == nil {
			return t1
		}
		return iff
	case 9: // ForAll
		v := vars[rng.Intn(len(vars))]
		body := randBoolNode(rng, s, vars, sym, depth-1)
		fa, _ := NewForAll([]*Variable{v}, body)
		if fa == nil {
			return body
		}
		return fa
	case 10: // Exists
		v := vars[rng.Intn(len(vars))]
		body := randBoolNode(rng, s, vars, sym, depth-1)
		ex, _ := NewExists([]*Variable{v}, body)
		if ex == nil {
			return body
		}
		return ex
	case 11: // Lambda
		v := vars[rng.Intn(len(vars))]
		body := randNode(rng, s, vars, sym, depth-1)
		lam, _ := NewLambda([]*Variable{v}, body)
		if lam == nil {
			return body
		}
		return lam
	case 12: // Globally
		body := randBoolNode(rng, s, vars, sym, depth-1)
		g, _ := NewGlobally(nil, body)
		if g == nil {
			return body
		}
		return g
	case 13: // Eventually
		body := randBoolNode(rng, s, vars, sym, depth-1)
		e, _ := NewEventually(nil, body)
		if e == nil {
			return body
		}
		return e
	case 14: // Definition
		v1 := vars[rng.Intn(len(vars))]
		v2 := vars[rng.Intn(len(vars))]
		return NewDefinition(v1, v2)
	case 15: // NamedBinder
		v := vars[rng.Intn(len(vars))]
		body := randBoolNode(rng, s, vars, sym, depth-1)
		nb, _ := NewNamedBinder("nb", []*Variable{v}, nil, body)
		if nb == nil {
			return body
		}
		return nb
	}
	return vars[0]
}

// randBoolNode generates an Expr guaranteed to be Boolean-sorted.
func randBoolNode(rng *rand.Rand, s Sort, vars []*Variable, sym *Symbol, depth int) Expr {
	if depth <= 0 {
		// Base case: simple equality
		v1 := vars[rng.Intn(len(vars))]
		v2 := vars[rng.Intn(len(vars))]
		e, _ := NewEq(v1, v2)
		if e == nil {
			return &And{} // true
		}
		return e
	}
	switch rng.Intn(5) {
	case 0:
		v1 := vars[rng.Intn(len(vars))]
		v2 := vars[rng.Intn(len(vars))]
		e, _ := NewEq(v1, v2)
		if e != nil {
			return e
		}
		return &And{}
	case 1:
		body := randBoolNode(rng, s, vars, sym, depth-1)
		n, _ := NewNot(body)
		if n != nil {
			return n
		}
		return body
	case 2:
		n := rng.Intn(3)
		terms := make([]Expr, n)
		for i := range terms {
			terms[i] = randBoolNode(rng, s, vars, sym, depth-1)
		}
		a, _ := NewAnd(terms...)
		if a != nil {
			return a
		}
		return &And{}
	case 3:
		n := rng.Intn(3)
		terms := make([]Expr, n)
		for i := range terms {
			terms[i] = randBoolNode(rng, s, vars, sym, depth-1)
		}
		o, _ := NewOr(terms...)
		if o != nil {
			return o
		}
		return &Or{}
	case 4:
		t1 := randBoolNode(rng, s, vars, sym, depth-1)
		t2 := randBoolNode(rng, s, vars, sym, depth-1)
		imp, _ := NewImplies(t1, t2)
		if imp != nil {
			return imp
		}
		return t1
	}
	return &And{}
}

// balanced checks that open/close characters are balanced in s.
func balanced(s string, open, close byte) bool {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == open {
			count++
		} else if s[i] == close {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

func FuzzSexpBalancedParens(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(42))
	f.Add(uint64(12345))
	f.Add(uint64(999999))
	f.Fuzz(func(t *testing.T, seed uint64) {
		node := randomExpr(seed, 5)
		sexp := string(node.Sexp())

		if !balanced(sexp, '(', ')') {
			t.Errorf("Unbalanced parens in: %s", sexp)
		}
		if !balanced(sexp, '[', ']') {
			t.Errorf("Unbalanced brackets in: %s", sexp)
		}
	})
}

func FuzzSexpDeterministic(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(42))
	f.Add(uint64(12345))
	f.Fuzz(func(t *testing.T, seed uint64) {
		node := randomExpr(seed, 4)
		s1 := string(node.Sexp())
		s2 := string(node.Sexp())
		if s1 != s2 {
			t.Errorf("Non-deterministic sexp:\n  call1: %s\n  call2: %s", s1, s2)
		}
	})
}

func FuzzSexpCanonEquality(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(42))
	f.Add(uint64(12345))
	f.Fuzz(func(t *testing.T, seed uint64) {
		node := randomExpr(seed, 4)
		sexp := string(node.Sexp())
		// Verify Canon() == Sexp() via type switch on all concrete types.
		var canon string
		switch n := node.(type) {
		case *Variable:
			canon = string(n.Canon())
		case *Symbol:
			canon = string(n.Canon())
		case *Apply:
			canon = string(n.Canon())
		case *Eq:
			canon = string(n.Canon())
		case *Not:
			canon = string(n.Canon())
		case *And:
			canon = string(n.Canon())
		case *Or:
			canon = string(n.Canon())
		case *Implies:
			canon = string(n.Canon())
		case *Iff:
			canon = string(n.Canon())
		case *ForAll:
			canon = string(n.Canon())
		case *Exists:
			canon = string(n.Canon())
		case *Lambda:
			canon = string(n.Canon())
		case *Globally:
			canon = string(n.Canon())
		case *Eventually:
			canon = string(n.Canon())
		case *NamedBinder:
			canon = string(n.Canon())
		case *Definition:
			canon = string(n.Canon())
		default:
			t.Fatalf("unknown type: %T", node)
		}
		if sexp != canon {
			t.Errorf("Canon != Sexp for %T:\n  sexp:  %s\n  canon: %s", node, sexp, canon)
		}
	})
}

func FuzzSexpNonEmpty(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(42))
	f.Add(uint64(12345))
	f.Fuzz(func(t *testing.T, seed uint64) {
		node := randomExpr(seed, 3)
		sexp := string(node.Sexp())
		if len(sexp) == 0 {
			t.Error("Empty sexp")
		}
		if sexp[0] != '(' {
			t.Errorf("Sexp does not start with '(': %s", sexp)
		}
		if !strings.HasSuffix(sexp, ")") {
			t.Errorf("Sexp does not end with ')': %s", sexp)
		}
	})
}
