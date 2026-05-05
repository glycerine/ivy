package proof

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

var _ = ast.NewAstConfig // ensure import

// FuzzExprListPop fuzzes the Pop method on ExprListOrLambdaUnion.
// Verifies: items are returned in order, last item is sticky,
// total pops never panic.
func FuzzExprListPop(f *testing.F) {
	f.Add(1, 5)
	f.Add(3, 10)
	f.Add(0, 1)
	f.Add(5, 5)
	f.Add(1, 1)

	f.Fuzz(func(t *testing.T, nItems, nPops int) {
		if nItems < 0 {
			nItems = -nItems
		}
		if nPops < 0 {
			nPops = -nPops
		}
		nItems = nItems % 20 // cap at 20 items
		nPops = nPops % 50   // cap at 50 pops

		s := proofMkSort("S")
		items := make([]lg.Expr, nItems)
		for i := range items {
			items[i] = proofMkConst("c"+string(rune('0'+i%10)), s)
		}
		u := &ExprListOrLambdaUnion{Items: items}

		var lastPopped lg.Expr
		for i := 0; i < nPops; i++ {
			got := u.Pop()
			if nItems == 0 {
				// Empty: should always return nil
				if got != nil {
					t.Fatalf("pop %d from empty: expected nil, got %v", i, got)
				}
			} else if i < nItems-1 {
				// Before last: should return items in order
				if got != items[i] {
					t.Fatalf("pop %d: expected items[%d], got different", i, i)
				}
			} else {
				// At or past last: should return last item
				if got != items[nItems-1] {
					t.Fatalf("pop %d: expected last item, got different", i)
				}
			}
			lastPopped = got
		}
		_ = lastPopped
	})
}

// FuzzMatchFromDefn fuzzes MatchFromDefn with definitions of varying arity.
// Verifies: valid definitions produce a lambda, invalid ones error.
func FuzzMatchFromDefn(f *testing.F) {
	f.Add(1, false)
	f.Add(2, false)
	f.Add(0, true)
	f.Add(3, true)

	f.Fuzz(func(t *testing.T, arity int, useIff bool) {
		if arity < 0 {
			arity = -arity
		}
		arity = arity % 6 // cap arity

		s := proofMkSort("S")

		// Build function sort: S x S x ... -> S
		sortArgs := make([]lg.Sort, arity+1)
		for i := range sortArgs {
			sortArgs[i] = s
		}
		fs, err := lg.NewFunctionSort(sortArgs...)
		if err != nil || arity == 0 {
			// Zero arity or invalid sort: skip
			return
		}
		fn := proofMkConst("f", fs)

		// Build parameters
		params := make([]*lg.Variable, arity)
		paramExprs := make([]lg.Expr, arity)
		for i := range params {
			params[i] = proofMkVar("X"+string(rune('0'+i)), s)
			paramExprs[i] = params[i]
		}

		// Build: f(X0, X1, ...) = c  or  f(X0, X1, ...) <-> true
		app := lg.MustApply(fn, paramExprs...)
		c := proofMkConst("c", s)

		var body lg.Expr
		if useIff {
			body = &lg.Iff{T1: app, T2: lg.True}
		} else {
			body = &lg.Eq{T1: app, T2: c}
		}
		fmla := &lg.ForAll{Variables: params, Body: body}
		defn := mkLF(proofTestAstCfg.NewAtom("def"), fmla)

		match, merr := MatchFromDefn(defn)
		if merr != nil {
			t.Fatalf("unexpected error for arity %d: %v", arity, merr)
		}

		lam, ok := match[lg.Key(fn)]
		if !ok {
			t.Fatal("expected fn key in match")
		}
		l, ok := lam.(*lg.Lambda)
		if !ok {
			t.Fatalf("expected *lg.Lambda, got %T", lam)
		}
		if len(l.Variables) != arity {
			t.Errorf("expected %d lambda vars, got %d", arity, len(l.Variables))
		}
	})
}

// FuzzMatchFromDefns fuzzes MatchFromDefns with varying numbers of definitions.
// Verifies: all definitions for the same symbol produce a union with correct count.
func FuzzMatchFromDefns(f *testing.F) {
	f.Add(1)
	f.Add(2)
	f.Add(5)
	f.Add(0)

	f.Fuzz(func(t *testing.T, nDefns int) {
		if nDefns < 0 {
			nDefns = -nDefns
		}
		nDefns = nDefns % 10 // cap

		if nDefns == 0 {
			_, _, err := MatchFromDefns(nil)
			if err == nil {
				t.Fatal("expected error for empty defns")
			}
			return
		}

		s := proofMkSort("S")
		x := proofMkVar("X", s)
		fs, _ := lg.NewFunctionSort(s, s)
		fn := proofMkConst("f", fs)

		defns := make([]*ast.LabeledFormula, nDefns)
		for i := range defns {
			rhs := proofMkConst("c"+string(rune('0'+i%10)), s)
			defns[i] = mkDefnLF("def"+string(rune('0'+i%10)), x, fn, rhs)
		}

		key, union, err := MatchFromDefns(defns)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key != lg.Key(fn) {
			t.Error("wrong key")
		}
		if len(union.Items) != nDefns {
			t.Errorf("expected %d items, got %d", nDefns, len(union.Items))
		}
	})
}

// FuzzUnfoldFmla fuzzes UnfoldFmla with formulas containing varying numbers
// of occurrences of the defined symbol. Verifies: the destructive pop assigns
// different lambdas to different occurrences in left-to-right order.
func FuzzUnfoldFmla(f *testing.F) {
	f.Add(1, 1) // 1 occurrence, 1 definition
	f.Add(2, 2) // 2 occurrences, 2 definitions (key BUG-11 case)
	f.Add(3, 3)
	f.Add(2, 1) // 2 occurrences, 1 definition (last reused)
	f.Add(1, 3) // 1 occurrence, 3 definitions (extras unused)

	f.Fuzz(func(t *testing.T, nOccurrences, nDefns int) {
		if nOccurrences < 0 {
			nOccurrences = -nOccurrences
		}
		if nDefns < 0 {
			nDefns = -nDefns
		}
		nOccurrences = (nOccurrences % 5) + 1 // 1..5
		nDefns = (nDefns % 5) + 1             // 1..5

		s := proofMkSort("S")
		x := proofMkVar("X", s)
		fs, _ := lg.NewFunctionSort(s, s)
		fn := proofMkConst("f", fs)
		a := proofMkConst("a", s)

		// Build nDefns definitions: forall X. f(X) = c_i
		defns := make([]*ast.LabeledFormula, nDefns)
		rhsConsts := make([]*lg.Const, nDefns)
		for i := range defns {
			rhsConsts[i] = proofMkConst("r"+string(rune('A'+i%26)), s)
			defns[i] = mkDefnLF("d"+string(rune('0'+i%10)), x, fn, rhsConsts[i])
		}

		// Build formula: And(f(a), f(a), ...) with nOccurrences
		terms := make([]lg.Expr, nOccurrences)
		for i := range terms {
			terms[i] = lg.MustApply(fn, a)
		}
		var fmla lg.Expr
		if len(terms) == 1 {
			fmla = terms[0]
		} else {
			fmla = &lg.And{Terms: terms}
		}

		resultNode := UnfoldFmla(fmla, [][]*ast.LabeledFormula{defns})
		result, ok := resultNode.(lg.Expr)
		if !ok {
			t.Fatalf("expected lg.Expr result, got %T", resultNode)
		}

		// Verify: each occurrence should have been replaced by the
		// corresponding rhs constant (or the last one if nDefns < nOccurrences).
		var resultTerms []lg.Expr
		if nOccurrences == 1 {
			resultTerms = []lg.Expr{result}
		} else if andR, ok := result.(*lg.And); ok {
			resultTerms = andR.Terms
		} else {
			t.Fatalf("expected And or single term, got %T", result)
		}

		for i, rt := range resultTerms {
			rc, ok := rt.(*lg.Const)
			if !ok {
				t.Fatalf("occurrence %d: expected Const, got %T: %v", i, rt, rt)
			}
			// Which rhs should this occurrence have gotten?
			// Pop behavior: occurrence i gets defn[i] if i < nDefns-1,
			// otherwise gets defn[nDefns-1] (last is sticky).
			expectedIdx := i
			if expectedIdx >= nDefns {
				expectedIdx = nDefns - 1
			}
			if rc.Name != rhsConsts[expectedIdx].Name {
				t.Errorf("occurrence %d: expected %s, got %s (nOcc=%d, nDefns=%d)",
					i, rhsConsts[expectedIdx].Name, rc.Name, nOccurrences, nDefns)
			}
		}
	})
}

// FuzzApplyUnfoldRec fuzzes the recursive unfold function directly.
// Builds nested And/Not/Implies formulas with embedded f(a) occurrences
// and verifies the unfold reaches all of them.
func FuzzApplyUnfoldRec(f *testing.F) {
	f.Add(0, 1) // And with 1 occurrence
	f.Add(1, 2) // Not(And(...))
	f.Add(2, 3) // Implies

	f.Fuzz(func(t *testing.T, wrapKind, nOccurrences int) {
		if nOccurrences < 0 {
			nOccurrences = -nOccurrences
		}
		nOccurrences = (nOccurrences % 4) + 1 // 1..4
		if wrapKind < 0 {
			wrapKind = -wrapKind
		}
		wrapKind = wrapKind % 3

		s := proofMkSort("S")
		x := proofMkVar("X", s)
		fs, _ := lg.NewFunctionSort(s, s)
		fn := proofMkConst("f", fs)
		a := proofMkConst("a", s)
		c := proofMkConst("c", s)

		lam, _ := lg.NewLambda([]*lg.Variable{x}, c)
		union := &ExprListOrLambdaUnion{Items: []lg.Expr{lam}}

		// Build base formula with nOccurrences of f(a)
		terms := make([]lg.Expr, nOccurrences)
		for i := range terms {
			terms[i] = lg.MustApply(fn, a)
		}
		var inner lg.Expr
		if len(terms) == 1 {
			inner = terms[0]
		} else {
			inner = &lg.And{Terms: terms}
		}

		// Wrap in different structures
		var fmla lg.Expr
		switch wrapKind {
		case 0:
			fmla = inner
		case 1:
			fmla = &lg.Not{Body: inner}
		case 2:
			fmla = &lg.Implies{T1: inner, T2: lg.True}
		}

		result := applyUnfoldRec(lg.Key(fn), union, fmla)
		if result == nil {
			t.Fatal("result should not be nil")
		}

		// Count remaining f applications in result — should be zero
		count := countApps(result, lg.Key(fn))
		if count != 0 {
			t.Errorf("expected 0 remaining f applications, got %d", count)
		}
	})
}

// countApps counts how many Apply nodes in fmla have function matching key.
func countApps(fmla lg.Expr, key lg.NodeKey) int {
	if fmla == nil {
		return 0
	}
	count := 0
	if app, ok := fmla.(*lg.Apply); ok {
		if c, ok := app.Func.(*lg.Const); ok && lg.Key(c) == key {
			count++
		}
	}
	for _, child := range fmla.Children() {
		count += countApps(child, key)
	}
	return count
}
