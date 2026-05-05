package goivy

// IsQF returns true if the formula is quantifier-free.
func IsQF(n Expr) bool {
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
func IsPrenexUniversal(n Expr) bool {
	if fa, ok := n.(*ForAll); ok {
		return IsPrenexUniversal(fa.Body)
	}
	if neg, ok := n.(*LogicNot); ok {
		return IsPrenexExistential(neg.Body)
	}
	return IsQF(n)
}

// IsPrenexExistential returns true if the formula is in prenex existential form.
func IsPrenexExistential(n Expr) bool {
	if ex, ok := n.(*LogicExists); ok {
		return IsPrenexExistential(ex.Body)
	}
	if neg, ok := n.(*LogicNot); ok {
		return IsPrenexUniversal(neg.Body)
	}
	return IsQF(n)
}

// IsAlternationFree returns true if the formula is prenex universal or
// prenex existential and has no free variables.
func IsAlternationFree(n Expr) bool {
	return IsPrenexUniversal(n) || (IsPrenexExistential(n) && FreeVariables(n).Len() == 0)
}

// IsAE returns true if the formula is in AE form (forall-exists).
func IsAE(n Expr) bool {
	if fa, ok := n.(*ForAll); ok {
		return IsAE(fa.Body)
	}
	if ex, ok := n.(*LogicExists); ok {
		return IsPrenexExistential(ex.Body)
	}
	if neg, ok := n.(*LogicNot); ok {
		return IsEA(neg.Body)
	}
	return IsQF(n)
}

// IsEA returns true if the formula is in EA form (exists-forall).
func IsEA(n Expr) bool {
	if ex, ok := n.(*LogicExists); ok {
		return IsEA(ex.Body)
	}
	if fa, ok := n.(*ForAll); ok {
		return IsPrenexUniversal(fa.Body)
	}
	if neg, ok := n.(*LogicNot); ok {
		return IsAE(neg.Body)
	}
	return IsQF(n)
}

// DropUniversals strips leading universal quantifiers and handles negation.
func IvyDropUniversals(n Expr) Expr {
	if fa, ok := n.(*ForAll); ok {
		return IvyDropUniversals(fa.Body)
	}
	if neg, ok := n.(*LogicNot); ok {
		notBody := DropExistentials(neg.Body)
		return &LogicNot{Body: notBody}
	}
	if and, ok := n.(*LogicAnd); ok && len(and.Terms) == 1 {
		return IvyDropUniversals(and.Terms[0])
	}
	return n
}

// DropExistentials strips leading existential quantifiers and handles negation.
func DropExistentials(n Expr) Expr {
	if ex, ok := n.(*LogicExists); ok {
		return DropExistentials(ex.Body)
	}
	if neg, ok := n.(*LogicNot); ok {
		notBody := IvyDropUniversals(neg.Body)
		return &LogicNot{Body: notBody}
	}
	return n
}

// Subterms yields all subterms of a term (including the term itself).
func Subterms(n Expr) []Expr {
	var result []Expr
	var collect func(Expr)
	collect = func(t Expr) {
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

// IvyLogics is the list of supported logic names.
var IvyLogics = []string{LogicEPR, LogicQF, LogicFO}

// DecidableLogics is the set of decidable logics.
var DecidableLogics = []string{LogicEPR, LogicQF}

// DefaultLogics is the default logic to target.
var DefaultLogics = []string{LogicEPR}

// --- Simplification helpers ---

// SimpAnd simplifies And(x, y) with constant folding.
func SimpAnd(x, y Expr) Expr {
	if IvyIsTrue(x) {
		return y
	}
	if IvyIsFalse(x) {
		return x
	}
	if IvyIsTrue(y) {
		return x
	}
	if IvyIsFalse(y) {
		return y
	}
	return &LogicAnd{Terms: []Expr{x, y}}
}

// SimpOr simplifies Or(x, y) with constant folding.
func SimpOr(x, y Expr) Expr {
	if IvyIsFalse(x) {
		return y
	}
	if IvyIsTrue(x) {
		return x
	}
	if IvyIsFalse(y) {
		return x
	}
	if IvyIsTrue(y) {
		return y
	}
	return &LogicOr{Terms: []Expr{x, y}}
}

// SimpNot simplifies Not(x) with constant folding and double-negation.
func SimpNot(x Expr) Expr {
	if neg, ok := x.(*LogicNot); ok {
		return neg.Body
	}
	if IvyIsTrue(x) {
		return &LogicOr{} // false
	}
	if IvyIsFalse(x) {
		return &LogicAnd{} // true
	}
	return &LogicNot{Body: x}
}

// SimpIte simplifies Ite(i, t, e) with constant folding.
func SimpIte(i, t, e Expr) Expr {
	if t.Equal(e) {
		return t
	}
	if IvyIsTrue(i) {
		return t
	}
	if IvyIsFalse(i) {
		return e
	}
	if IvyIsTrue(t) {
		return SimpOr(i, e)
	}
	if IvyIsFalse(t) {
		return SimpAnd(SimpNot(i), e)
	}
	if IvyIsTrue(e) {
		return SimpOr(SimpNot(i), t)
	}
	if IvyIsFalse(e) {
		return SimpAnd(i, t)
	}
	return &LogicIte{ISort: t.NodeSort(), Cond: i, Then: t, Else: e}
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
func Polar(fmla Expr, pos, pol int) int {
	switch fmla.(type) {
	case *LogicNot:
		return NegatePolarity(pol)
	case *LogicImplies:
		if pos == 1 {
			return pol
		}
		return NegatePolarity(pol)
	case *ForAll, *LogicExists, *LogicAnd, *LogicOr:
		return pol
	case *LogicIte:
		if pos == 0 {
			return -1
		}
		return pol
	}
	return -1
}
