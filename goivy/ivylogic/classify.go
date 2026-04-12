package ivylogic

import (
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
)

// IsQF returns true if the formula is quantifier-free.
func IsQF(n lg.Expr) bool {
	if IsQuantifier(n) {
		return false
	}
	for _, c := range n.Children() {
		if !IsQF(c) {
			return false
		}
	}
	return true
}

// IsPrenexUniversal returns true if the formula is in prenex universal form:
// forall X. forall Y. ... QF-body
// Also handles negation (negated prenex existential is prenex universal).
func IsPrenexUniversal(n lg.Expr) bool {
	if fa, ok := n.(*lg.ForAll); ok {
		return IsPrenexUniversal(fa.Body)
	}
	if neg, ok := n.(*lg.Not); ok {
		return IsPrenexExistential(neg.Body)
	}
	return IsQF(n)
}

// IsPrenexExistential returns true if the formula is in prenex existential form.
func IsPrenexExistential(n lg.Expr) bool {
	if ex, ok := n.(*lg.Exists); ok {
		return IsPrenexExistential(ex.Body)
	}
	if neg, ok := n.(*lg.Not); ok {
		return IsPrenexUniversal(neg.Body)
	}
	return IsQF(n)
}

// IsAlternationFree returns true if the formula is prenex universal or
// prenex existential and has no free variables.
func IsAlternationFree(n lg.Expr) bool {
	return IsPrenexUniversal(n) || (IsPrenexExistential(n) && lu.FreeVariables(n).Len() == 0)
}

// IsAE returns true if the formula is in AE form (forall-exists).
func IsAE(n lg.Expr) bool {
	if fa, ok := n.(*lg.ForAll); ok {
		return IsAE(fa.Body)
	}
	if ex, ok := n.(*lg.Exists); ok {
		return IsPrenexExistential(ex.Body)
	}
	if neg, ok := n.(*lg.Not); ok {
		return IsEA(neg.Body)
	}
	return IsQF(n)
}

// IsEA returns true if the formula is in EA form (exists-forall).
func IsEA(n lg.Expr) bool {
	if ex, ok := n.(*lg.Exists); ok {
		return IsEA(ex.Body)
	}
	if fa, ok := n.(*lg.ForAll); ok {
		return IsPrenexUniversal(fa.Body)
	}
	if neg, ok := n.(*lg.Not); ok {
		return IsAE(neg.Body)
	}
	return IsQF(n)
}

// DropUniversals strips leading universal quantifiers and handles negation.
func DropUniversals(n lg.Expr) lg.Expr {
	if fa, ok := n.(*lg.ForAll); ok {
		return DropUniversals(fa.Body)
	}
	if neg, ok := n.(*lg.Not); ok {
		notBody := DropExistentials(neg.Body)
		return &lg.Not{Body: notBody}
	}
	if and, ok := n.(*lg.And); ok && len(and.Terms) == 1 {
		return DropUniversals(and.Terms[0])
	}
	return n
}

// DropExistentials strips leading existential quantifiers and handles negation.
func DropExistentials(n lg.Expr) lg.Expr {
	if ex, ok := n.(*lg.Exists); ok {
		return DropExistentials(ex.Body)
	}
	if neg, ok := n.(*lg.Not); ok {
		notBody := DropUniversals(neg.Body)
		return &lg.Not{Body: notBody}
	}
	return n
}

// Subterms yields all subterms of a term (including the term itself).
func Subterms(n lg.Expr) []lg.Expr {
	var result []lg.Expr
	var collect func(lg.Expr)
	collect = func(t lg.Expr) {
		result = append(result, t)
		for _, c := range t.Children() {
			collect(c)
		}
	}
	collect(n)
	return result
}

// Logic names.
const (
	LogicEPR = "epr"
	LogicQF  = "qf"
	LogicFO  = "fo"
)

// Logics is the list of supported logic names.
var Logics = []string{LogicEPR, LogicQF, LogicFO}

// DecidableLogics is the set of decidable logics.
var DecidableLogics = []string{LogicEPR, LogicQF}

// DefaultLogics is the default logic to target.
var DefaultLogics = []string{LogicEPR}

// --- Simplification helpers ---

// SimpAnd simplifies And(x, y) with constant folding.
func SimpAnd(x, y lg.Expr) lg.Expr {
	if IsTrue(x) {
		return y
	}
	if IsFalse(x) {
		return x
	}
	if IsTrue(y) {
		return x
	}
	if IsFalse(y) {
		return y
	}
	return &lg.And{Terms: []lg.Expr{x, y}}
}

// SimpOr simplifies Or(x, y) with constant folding.
func SimpOr(x, y lg.Expr) lg.Expr {
	if IsFalse(x) {
		return y
	}
	if IsTrue(x) {
		return x
	}
	if IsFalse(y) {
		return x
	}
	if IsTrue(y) {
		return y
	}
	return &lg.Or{Terms: []lg.Expr{x, y}}
}

// SimpNot simplifies Not(x) with constant folding and double-negation.
func SimpNot(x lg.Expr) lg.Expr {
	if neg, ok := x.(*lg.Not); ok {
		return neg.Body
	}
	if IsTrue(x) {
		return &lg.Or{} // false
	}
	if IsFalse(x) {
		return &lg.And{} // true
	}
	return &lg.Not{Body: x}
}

// SimpIte simplifies Ite(i, t, e) with constant folding.
func SimpIte(i, t, e lg.Expr) lg.Expr {
	if t.Equal(e) {
		return t
	}
	if IsTrue(i) {
		return t
	}
	if IsFalse(i) {
		return e
	}
	if IsTrue(t) {
		return SimpOr(i, e)
	}
	if IsFalse(t) {
		return SimpAnd(SimpNot(i), e)
	}
	if IsTrue(e) {
		return SimpOr(SimpNot(i), t)
	}
	if IsFalse(e) {
		return SimpAnd(i, t)
	}
	return &lg.Ite{ISort: t.NodeSort(), Cond: i, Then: t, Else: e}
}

// NegatePolarity negates a polarity value (0 ↔ 1, nil stays nil).
// Returns -1 for "both" (equivalent to Python's None).
func NegatePolarity(pol int) int {
	if pol < 0 {
		return pol
	}
	return 1 - pol
}

// Polar returns the polarity of the pos-th argument of fmla assuming
// fmla has polarity pol. Returns -1 for "both polarities".
func Polar(fmla lg.Expr, pos, pol int) int {
	switch fmla.(type) {
	case *lg.Not:
		return NegatePolarity(pol)
	case *lg.Implies:
		if pos == 1 {
			return pol
		}
		return NegatePolarity(pol)
	case *lg.ForAll, *lg.Exists, *lg.And, *lg.Or:
		return pol
	case *lg.Ite:
		if pos == 0 {
			return -1
		}
		return pol
	}
	return -1
}
