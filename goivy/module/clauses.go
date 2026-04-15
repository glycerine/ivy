package module

import (
	"fmt"
	"strings"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// Clauses wraps a set of formulas and definitions.
// In Python: Clauses(fmlas=[], defs=[], annot=None).
//
// Fmlas stores conjuncts as formulas (not literal clauses).
// Clausal form (Tseitin encoding) can be computed lazily if needed.
type Clauses struct {
	Fmlas  []lg.Expr          // conjuncts (formulas)
	Defs   []*il.Definition   // definitions
	DefIdx map[lg.NodeKey]int // definition index: definesKey() -> index in Defs
	Annot  interface{}        // annotation (for trace reconstruction)
}

// NewClauses constructs a Clauses value. Each formula is first normalized
// via coerceClauseToFormula (strips leading ForAll, matching Python's
// coerce_clause_to_formula), then flattened: any top-level And is expanded
// into its conjuncts (collect_and_list).
// Definitions are indexed by their defining symbol name.
func NewClauses(fmlas []lg.Expr, defs []*il.Definition, annot interface{}) *Clauses {
	coerced := make([]lg.Expr, len(fmlas))
	for i, f := range fmlas {
		coerced[i] = dropUniversals(f)
	}
	flat := collectAndList(coerced)
	idx := make(map[lg.NodeKey]int, len(defs))
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
func definesKey(d *il.Definition) lg.NodeKey {
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
	idx := make(map[lg.NodeKey]int, len(c.DefIdx))
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
	if xtracer.Enabled {
		xtracer.Trace("ops.ToOpenFormula nFmlas=%d nDefs=%d", len(c.Fmlas), len(c.Defs))
	}
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
	return lu.CloseEPR(c.ToOpenFormula())
}

// Conjuncts returns [CloseEPR(c) for c in self.Fmlas].
// Matches Python ivy_logic_utils.py Clauses.conjuncts (lines 63-65).
func (c *Clauses) Conjuncts() []lg.Expr {
	if xtracer.Enabled {
		xtracer.Trace("ops.Conjuncts nFmlas=%d nDefs=%d", len(c.Fmlas), len(c.Defs))
	}
	if len(c.Defs) > 0 {
		panic("Conjuncts requires no definitions")
	}
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
		syms := il.UsedSymbolsAst(f)
		for _, s := range syms {
			name := lg.ExprName(s)
			if len(name) >= 2 && name[0] == '_' && name[1] == '_' {
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

// Symbols yields all symbols used in the Clauses.
// Delegates to the faithful port il.UsedSymbolsAst (via symbols_ilu_ast).
func (c *Clauses) Symbols() map[lg.NodeKey]lg.Expr {
	result := make(map[lg.NodeKey]lg.Expr)
	for _, f := range c.Fmlas {
		for k, v := range il.UsedSymbolsAst(f) {
			result[k] = v
		}
	}
	for _, d := range c.Defs {
		for k, v := range il.UsedSymbolsAst(d) {
			result[k] = v
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

// defToConstraint converts a Definition to a constraint formula.
// Delegates to the faithful port in ivylogic/constraint.go.
func defToConstraint(d *il.Definition) lg.Expr {
	result := il.DefinitionToConstraint(d)
	if xtracer.Enabled {
		lhsSort := d.Lhs.NodeSort()
		xtracer.Trace("module/clauses.go:262 defToConstraint lhsSort=%v resultType=%v", lhsSort, iu.ShortTypeName(result))
		if lhsSort == nil {
			vv("defToConstraint() lhsSort was nil! stack=\n%v\n", stack())
		}
	}
	return result
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
// Matches Python's drop_universals: also unwraps singleton And and
// handles Not by calling dropExistentials on the negated body.
func dropUniversals(f lg.Expr) lg.Expr {
	switch t := f.(type) {
	case *lg.ForAll:
		return dropUniversals(t.Body)
	case *lg.Not:
		return &lg.Not{Body: dropExistentials(t.Body)}
	case *lg.And:
		if len(t.Terms) == 1 {
			return dropUniversals(t.Terms[0])
		}
	}
	return f
}

// dropExistentials strips leading Exists quantifiers from a formula.
// Matches Python's drop_existentials: handles Not by calling
// dropUniversals on the negated body.
func dropExistentials(f lg.Expr) lg.Expr {
	switch t := f.(type) {
	case *lg.Exists:
		return dropExistentials(t.Body)
	case *lg.Not:
		return &lg.Not{Body: dropUniversals(t.Body)}
	}
	return f
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

// isSkolem returns true if the constant name contains "__" (Skolem convention).
// Python: Symbol.is_skolem = lambda self: self.contains('__')
// where contains = lambda self, s: (s in self.name)
func isSkolem(c *lg.Const) bool {
	return strings.Contains(c.Name, "__")
}


// usesSymbolsAST returns true if any of the given symbols occurs in the node.
func usesSymbolsAST(syms map[lg.NodeKey]lg.Expr, node lg.Expr) bool {
	used := il.UsedSymbolsAst(node)
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
func SymPlaceholders(sym *lg.Const) []*lg.Variable {
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
func SymInst(sym *lg.Const) lg.Expr {
	phs := SymPlaceholders(sym)
	if len(phs) == 0 {
		return sym
	}
	args := make([]lg.Expr, len(phs))
	for i, v := range phs {
		args[i] = v
	}
	app, err := lg.NewApply(sym, args...)
	if err != nil {
		panic(fmt.Sprintf("SymInst: NewApply failed for %v: %v", sym, err))
	}
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
