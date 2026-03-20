package clauseops

import (
	"fmt"
	"iter"
	"strings"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
)

// Clauses wraps a set of formulas and definitions.
// In Python: Clauses(fmlas=[], defs=[], annot=None).
//
// Fmlas stores conjuncts as formulas (not literal clauses).
// Clausal form (Tseitin encoding) can be computed lazily if needed.
type Clauses struct {
	Fmlas  []lg.Expr          // conjuncts (formulas)
	Defs   []*il.Definition   // definitions
	DefIdx map[string]int     // definition index: defines().String() -> index in Defs
	Annot  interface{}        // annotation (for trace reconstruction)
}

// NewClauses constructs a Clauses value. The formulas are flattened:
// any top-level And is expanded into its conjuncts (collect_and_list).
// Definitions are indexed by their defining symbol name.
func NewClauses(fmlas []lg.Expr, defs []*il.Definition, annot interface{}) *Clauses {
	flat := collectAndList(fmlas)
	idx := make(map[string]int, len(defs))
	for i, d := range defs {
		key := definesKey(d)
		idx[key] = i
	}
	return &Clauses{
		Fmlas:  flat,
		Defs:   defs,
		DefIdx: idx,
		Annot:  annot,
	}
}

// definesKey returns a structural identity key for the symbol defined by
// a Definition. Uses Sexp() to include both name and sort, matching
// Python's structural equality on Symbol objects used as defidx keys.
func definesKey(d *il.Definition) string {
	return lg.Key(d.Defines())
}

// IsFalse returns true if any formula is the logical False constant (empty Or).
func (c *Clauses) IsFalse() bool {
	for _, f := range c.Fmlas {
		if il.IsFalse(f) {
			return true
		}
	}
	return false
}

// IsTrue returns true if all formulas are the logical True constant (empty And),
// or there are no formulas.
func (c *Clauses) IsTrue() bool {
	for _, f := range c.Fmlas {
		if !il.IsTrue(f) {
			return false
		}
	}
	return true
}

// Copy returns a shallow copy of the Clauses.
func (c *Clauses) Copy() *Clauses {
	fmlas := make([]lg.Expr, len(c.Fmlas))
	copy(fmlas, c.Fmlas)
	defs := make([]*il.Definition, len(c.Defs))
	copy(defs, c.Defs)
	idx := make(map[string]int, len(c.DefIdx))
	for k, v := range c.DefIdx {
		idx[k] = v
	}
	// Python's Clauses.copy() drops the annotation:
	//   def copy(self): return Clauses(list(self.fmlas), list(self.defs))
	return &Clauses{
		Fmlas:  fmlas,
		Defs:   defs,
		DefIdx: idx,
		Annot:  nil,
	}
}

// ToOpenFormula converts the Clauses to a single formula:
// And(def1.ToConstraint(), def2.ToConstraint(), ..., fmla1, fmla2, ...).
func (c *Clauses) ToOpenFormula() lg.Expr {
	conjuncts := make([]lg.Expr, 0, len(c.Defs)+len(c.Fmlas))
	for _, d := range c.Defs {
		conjuncts = append(conjuncts, defToConstraint(d))
	}
	conjuncts = append(conjuncts, c.Fmlas...)
	return &lg.And{Terms: conjuncts}
}

// ToFormula converts to a closed formula by universally quantifying
// over all free variables.
func (c *Clauses) ToFormula() lg.Expr {
	return il.CloseFormula(c.ToOpenFormula())
}

// Conjuncts returns [CloseEPR(c) for c in self.Fmlas].
// Matches Python ivy_logic_utils.py Clauses.conjuncts (lines 63-65).
func (c *Clauses) Conjuncts() []lg.Expr {
	result := make([]lg.Expr, len(c.Fmlas))
	for i, f := range c.Fmlas {
		result[i] = lu.CloseEPR(f)
	}
	return result
}

// IsUniversalFirstOrder returns true if there are no definitions
// and no skolem symbols in the formulas.
func (c *Clauses) IsUniversalFirstOrder() bool {
	if len(c.Defs) > 0 {
		return false
	}
	for _, f := range c.Fmlas {
		syms := usedSymbolsAST(f)
		for _, s := range syms {
			if isSkolem(s.(*lg.Symbol)) {
				return false
			}
		}
	}
	return true
}

// String returns a human-readable representation.
func (c *Clauses) String() string {
	var b strings.Builder
	b.WriteString("Clauses{")
	if len(c.Defs) > 0 {
		b.WriteString("defs=[")
		for i, d := range c.Defs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(d.String())
		}
		b.WriteString("], ")
	}
	b.WriteString("fmlas=[")
	for i, f := range c.Fmlas {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(f.String())
	}
	b.WriteString("]}")
	return b.String()
}

// Equal returns true if two Clauses have the same formulas and definitions.
func (c *Clauses) Equal(other *Clauses) bool {
	if len(c.Fmlas) != len(other.Fmlas) || len(c.Defs) != len(other.Defs) {
		return false
	}
	for i := range c.Fmlas {
		if !c.Fmlas[i].Equal(other.Fmlas[i]) {
			return false
		}
	}
	for i := range c.Defs {
		if !c.Defs[i].Equal(other.Defs[i]) {
			return false
		}
	}
	return true
}

// Symbols yields all constant symbols used in the Clauses.
func (c *Clauses) Symbols() map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
	for _, f := range c.Fmlas {
		for s, node := range usedSymbolsAST(f) {
			result[s] = node
		}
	}
	for _, d := range c.Defs {
		for s, node := range usedSymbolsAST(d) {
			result[s] = node
		}
	}
	return result
}

// Apply applies a function to all subformulas and definitions,
// returning a new Clauses.
func (c *Clauses) Apply(fn func(lg.Expr) lg.Expr) *Clauses {
	fmlas := make([]lg.Expr, len(c.Fmlas))
	for i, f := range c.Fmlas {
		fmlas[i] = fn(f)
	}
	defs := make([]*il.Definition, len(c.Defs))
	for i, d := range c.Defs {
		result := fn(d)
		if nd, ok := result.(*il.Definition); ok {
			defs[i] = nd
		} else {
			// If fn returns something that's not a Definition, wrap it.
			// This shouldn't normally happen.
			defs[i] = d
		}
	}
	return NewClauses(fmlas, defs, c.Annot)
}

// --- Constructors ---

// TrueClauses returns a Clauses representing logical True (empty clause set = tautology).
func TrueClauses(annot interface{}) *Clauses {
	return NewClauses(nil, nil, annot)
}

// FalseClauses returns a Clauses representing logical False.
func FalseClauses(annot interface{}) *Clauses {
	return NewClauses([]lg.Expr{lg.False}, nil, annot)
}

// FormulaToClauses wraps a single formula as a Clauses.
// If the formula is an And or Or with one element, it is unwrapped.
func FormulaToClauses(f lg.Expr, annot interface{}) *Clauses {
	// Unwrap single-element And/Or
	f = unwrapSingleton(f)
	// Drop universals (strip leading ForAll)
	f = dropUniversals(f)
	return NewClauses([]lg.Expr{f}, nil, annot)
}

// --- helpers ---

// defToConstraint converts a Definition to a constraint formula:
// ForAll vars. (lhs = rhs) or ForAll vars. Iff(lhs, rhs) for Boolean.
func defToConstraint(d *il.Definition) lg.Expr {
	// Python: d.to_constraint() produces ForAll(vars, Iff(lhs, rhs)) for bool,
	// or ForAll(vars, Eq(lhs, rhs)) for non-bool.
	// We produce an Iff for Boolean sort, Eq otherwise.
	lhs := d.Lhs
	rhs := d.Rhs
	var constraint lg.Expr
	if lg.SortEqual(rhs.NodeSort(), lg.Boolean) {
		constraint = &lg.Iff{T1: lhs, T2: rhs}
	} else {
		constraint = &lg.Eq{T1: lhs, T2: rhs}
	}
	// Universally quantify over free variables in the constraint
	fvs := lu.FreeVariablesList(constraint)
	if len(fvs) > 0 {
		constraint = &lg.ForAll{Variables: fvs, Body: constraint}
	}
	return constraint
}

// collectAndList flattens a list of formulas: any top-level And is expanded.
func collectAndList(fmlas []lg.Expr) []lg.Expr {
	var result []lg.Expr
	for _, f := range fmlas {
		if a, ok := f.(*lg.And); ok {
			if len(a.Terms) > 0 {
				result = append(result, collectAndList(a.Terms)...)
			}
			// Empty And() = True → skip (adds no conjuncts)
		} else {
			result = append(result, f)
		}
	}
	return result
}

// dropUniversals strips leading ForAll quantifiers from a formula.
func dropUniversals(f lg.Expr) lg.Expr {
	for {
		switch t := f.(type) {
		case *lg.ForAll:
			f = t.Body
			continue
		case *lg.And:
			if len(t.Terms) == 1 {
				f = t.Terms[0]
				continue
			}
		}
		return f
	}
}

// unwrapSingleton unwraps a single-element And or Or.
func unwrapSingleton(f lg.Expr) lg.Expr {
	switch t := f.(type) {
	case *lg.And:
		if len(t.Terms) == 1 {
			return t.Terms[0]
		}
	case *lg.Or:
		if len(t.Terms) == 1 {
			return t.Terms[0]
		}
	}
	return f
}

// isSkolem returns true if the constant name starts with "__" (Skolem convention).
func isSkolem(c *lg.Symbol) bool {
	return len(c.Name) >= 2 && c.Name[0] == '_' && c.Name[1] == '_'
}

// usedSymbolsAST returns the set of constant symbols used in an AST node.
// This matches Python's used_symbols_ast: it yields the function symbols
// of applications, plus recurses into arguments.
func usedSymbolsAST(node lg.Expr) map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
	symbolsASTRec(node, result)
	return result
}

func symbolsASTRec(node lg.Expr, result map[lg.NodeKey]lg.Expr) {
	switch t := node.(type) {
	case *lg.Symbol:
		result[lg.Key(t)] = t
	case *lg.Apply:
		if c, ok := t.Func.(*lg.Symbol); ok {
			result[lg.Key(c)] = c
		} else {
			symbolsASTRec(t.Func, result)
		}
		for _, arg := range t.Terms {
			symbolsASTRec(arg, result)
		}
		return
	}
	for _, c := range node.Children() {
		symbolsASTRec(c, result)
	}
}

// IterSymbolsAST yields symbols from the AST in depth-first traversal order.
// This matches Python's symbols_ast generator: yield func head first (for Apply),
// then recurse into args. Unlike UsedSymbolsAST, this preserves encounter order
// and does not use a map internally.
func IterSymbolsAST(node lg.Expr) iter.Seq[*lg.Symbol] {
	return func(yield func(*lg.Symbol) bool) {
		iterSymbolsRec(node, yield)
	}
}

func iterSymbolsRec(node lg.Expr, yield func(*lg.Symbol) bool) bool {
	switch t := node.(type) {
	case *lg.Symbol:
		return yield(t)
	case *lg.Apply:
		if c, ok := t.Func.(*lg.Symbol); ok {
			if !yield(c) {
				return false
			}
		} else {
			if !iterSymbolsRec(t.Func, yield) {
				return false
			}
		}
		for _, arg := range t.Terms {
			if !iterSymbolsRec(arg, yield) {
				return false
			}
		}
		return true
	}
	for _, c := range node.Children() {
		if !iterSymbolsRec(c, yield) {
			return false
		}
	}
	return true
}

// usesSymbolsAST returns true if any of the given symbols occurs in the node.
func usesSymbolsAST(syms map[lg.NodeKey]lg.Expr, node lg.Expr) bool {
	used := usedSymbolsAST(node)
	for s := range syms {
		if _, ok := used[s]; ok {
			return true
		}
	}
	return false
}

// Negate negates a formula with double-negation elimination.
func Negate(f lg.Expr) lg.Expr {
	if n, ok := f.(*lg.Not); ok {
		return n.Body
	}
	return &lg.Not{Body: f}
}

// IsTrue returns true if the node is logical True (empty And).
func IsTrue(n lg.Expr) bool {
	return il.IsTrue(n)
}

// IsFalse returns true if the node is logical False (empty Or).
func IsFalse(n lg.Expr) bool {
	return il.IsFalse(n)
}

// SymPlaceholders returns placeholder variables V0, V1, ... for each
// domain sort of the given symbol's function sort.
func SymPlaceholders(sym *lg.Symbol) []*lg.Variable {
	dom := il.SortDomain(sym.CSort)
	if len(dom) == 0 {
		return nil
	}
	result := make([]*lg.Variable, len(dom))
	for i, s := range dom {
		v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), s)
		result[i] = v
	}
	return result
}

// SymInst instantiates a symbol with placeholder variables.
// For relations, returns Apply(sym, V0, V1, ...).
// For functions, returns Apply(sym, V0, V1, ...).
func SymInst(sym *lg.Symbol) lg.Expr {
	phs := SymPlaceholders(sym)
	if len(phs) == 0 {
		return sym
	}
	args := make([]lg.Expr, len(phs))
	for i, v := range phs {
		args[i] = v
	}
	app := &lg.Apply{Func: sym, Terms: args}
	return app
}

// EqAtom returns the equality atom x == y.
func EqAtom(x, y lg.Expr) lg.Expr {
	return &lg.Eq{T1: x, T2: y}
}

// EqLit returns an equality literal (positive): x == y.
// Returns an il.Literal with polarity 1.
func EqLit(x, y lg.Expr) *il.Literal {
	return il.NewLiteral(1, EqAtom(x, y))
}
