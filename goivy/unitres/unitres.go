// Ported to Go from ivy_unitres.py.

// Package unitres implements a unit resolution engine with literal
// hash-consing, literal indexing (subsumption, unification), an
// equational theory via congruence closure, and incremental
// push/pop support.
package unitres

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/congclos"
	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/resolution"
)

// ---------- Literal type ----------

// Literal is a polarity (0=negative, 1=positive) paired with an Atom.
type Literal struct {
	Polarity int
	Atom     *resolution.Atom
}

// NewLiteral creates a Literal.
func NewLiteral(polarity int, atom *resolution.Atom) *Literal {
	return &Literal{Polarity: polarity, Atom: atom}
}

// String renders the literal. Negative literals are prefixed with "~".
func (l *Literal) String() string {
	prefix := ""
	if l.Polarity == 0 {
		prefix = "~"
	}
	args := make([]string, len(l.Atom.Args))
	for i, a := range l.Atom.Args {
		args[i] = a.String()
	}
	if len(args) == 0 {
		return fmt.Sprintf("%s%s", prefix, l.Atom.RelName)
	}
	return fmt.Sprintf("%s%s(%s)", prefix, l.Atom.RelName, strings.Join(args, ", "))
}

// Invert returns the negation of this literal.
func (l *Literal) Invert() *Literal {
	return &Literal{Polarity: 1 - l.Polarity, Atom: l.Atom}
}

// LitEqual returns true if two literals are syntactically equal.
func LitEqual(a, b *Literal) bool {
	if a.Polarity != b.Polarity {
		return false
	}
	return AtomEqual(a.Atom, b.Atom)
}

// AtomEqual returns true if two atoms are syntactically equal.
func AtomEqual(a, b *resolution.Atom) bool {
	if a.RelName != b.RelName || len(a.Args) != len(b.Args) {
		return false
	}
	for i := range a.Args {
		if !a.Args[i].Equal(b.Args[i]) {
			return false
		}
	}
	return true
}

// ---------- Term/Literal helpers ----------

// rep returns the name of a Var or Const.
func rep(n logic.Expr) string {
	switch t := n.(type) {
	case *logic.Variable:
		return t.Name
	case *logic.Const:
		return t.Name
	default:
		return n.String()
	}
}

// isVar returns true if the node is a *logic.Variable.
func isVar(n logic.Expr) bool {
	_, ok := n.(*logic.Variable)
	return ok
}

// isConst returns true if the node is a *logic.Const.
func isConst(n logic.Expr) bool {
	_, ok := n.(*logic.Const)
	return ok
}

// isEqualityLit returns true if the literal is a positive equality.
func isEqualityLit(lit *Literal) bool {
	return lit.Polarity == 1 && lit.Atom.RelName == "="
}

// isDisequalityLit returns true if the literal is a negative equality.
func isDisequalityLit(lit *Literal) bool {
	return lit.Polarity == 0 && lit.Atom.RelName == "="
}

// isTautEqualityLit returns true if the literal is x=x (positive).
func isTautEqualityLit(lit *Literal) bool {
	if lit.Polarity != 1 || lit.Atom.RelName != "=" {
		return false
	}
	if len(lit.Atom.Args) != 2 {
		return false
	}
	return lit.Atom.Args[0].Equal(lit.Atom.Args[1])
}

// isTrueLit returns true if the literal is trivially true.
func isTrueLit(lit *Literal) bool {
	// A literal with atom = And() (empty And = True) and polarity 1
	if lit.Polarity == 1 && len(lit.Atom.Args) == 1 {
		if a, ok := lit.Atom.Args[0].(*logic.And); ok && len(a.Terms) == 0 {
			return true
		}
	}
	return false
}

// isTautLit returns true if the literal is tautological.
func isTautLit(lit *Literal) bool {
	return isTrueLit(lit) || isTautEqualityLit(lit)
}

// isVacEqualityLit returns true if the literal is ~(x=x).
func isVacEqualityLit(lit *Literal) bool {
	if lit.Polarity != 0 || lit.Atom.RelName != "=" {
		return false
	}
	if len(lit.Atom.Args) != 2 {
		return false
	}
	return lit.Atom.Args[0].Equal(lit.Atom.Args[1])
}

// isVacLit returns true if the literal is vacuously false.
func isVacLit(lit *Literal) bool {
	return isVacEqualityLit(lit)
}

// isGroundLit returns true if the literal contains no variables.
func isGroundLit(lit *Literal) bool {
	for _, t := range lit.Atom.Args {
		if isVar(t) {
			return false
		}
	}
	return true
}

// isGroundClause returns true if all literals in the clause are ground.
func isGroundClause(cl []*Literal) bool {
	for _, lit := range cl {
		if !isGroundLit(lit) {
			return false
		}
	}
	return true
}

// isGroundEqualityLit returns true for a positive ground equality.
func isGroundEqualityLit(lit *Literal) bool {
	return isEqualityLit(lit) && isGroundLit(lit)
}

// swapArgsLit swaps the two args of a binary literal.
func swapArgsLit(lit *Literal) *Literal {
	if len(lit.Atom.Args) != 2 {
		return lit
	}
	return &Literal{
		Polarity: lit.Polarity,
		Atom:     resolution.NewAtom(lit.Atom.RelName, lit.Atom.Args[1], lit.Atom.Args[0]),
	}
}

// ---------- Hash-consing ----------

// LitConsing provides hash-consed literal ids.
type LitConsing struct {
	literals []*Literal
	litHash  map[litKey]int
}

type litKey struct {
	polarity int
	relname  string
	argsKey  string // concatenation of rep(arg) separated by \x00
}

func makeLitKey(lit *Literal) litKey {
	parts := make([]string, len(lit.Atom.Args))
	for i, a := range lit.Atom.Args {
		parts[i] = rep(a)
	}
	return litKey{
		polarity: lit.Polarity,
		relname:  lit.Atom.RelName,
		argsKey:  strings.Join(parts, "\x00"),
	}
}

// NewLitConsing creates a new LitConsing.
func NewLitConsing() *LitConsing {
	return &LitConsing{litHash: make(map[litKey]int)}
}

// LitID returns a unique id for a literal (hash-consing).
func (lc *LitConsing) LitID(lit *Literal) int {
	key := makeLitKey(lit)
	if id, ok := lc.litHash[key]; ok {
		return id
	}
	id := len(lc.literals)
	lc.literals = append(lc.literals, lit)
	lc.litHash[key] = id
	return id
}

// ---------- Canonize ----------

// CanonizeLiteral renames variables to canonical constants __v0, __v1, ...
// Returns the canonized literal and the substitution mapping old var names to new constants.
func CanonizeLiteral(lit *Literal) (*Literal, map[string]*logic.Const) {
	subs := make(map[string]*logic.Const)
	terms := make([]logic.Expr, len(lit.Atom.Args))
	for i, t := range lit.Atom.Args {
		if isVar(t) {
			name := rep(t)
			if _, ok := subs[name]; !ok {
				subs[name] = logic.NewConst(fmt.Sprintf("__v%d", i), logic.TopS)
			}
			terms[i] = subs[name]
		} else {
			terms[i] = t
		}
	}
	return NewLiteral(lit.Polarity, resolution.NewAtom(lit.Atom.RelName, terms...)), subs
}

// CanonizeLiteralVars renames variables to canonical variables V0, V1, ...
func CanonizeLiteralVars(lit *Literal) *Literal {
	subs := make(map[string]logic.Expr)
	terms := make([]logic.Expr, len(lit.Atom.Args))
	for i, t := range lit.Atom.Args {
		if isVar(t) {
			name := rep(t)
			if _, ok := subs[name]; !ok {
				v, _ := logic.NewVariable(fmt.Sprintf("V%d", i), t.NodeSort())
				subs[name] = v
			}
			terms[i] = subs[name]
		} else {
			terms[i] = t
		}
	}
	return NewLiteral(lit.Polarity, resolution.NewAtom(lit.Atom.RelName, terms...))
}

// CanonizeLiteralUnique renames variables to canonical variables W0, W1, ...
func CanonizeLiteralUnique(lit *Literal) *Literal {
	subs := make(map[string]logic.Expr)
	terms := make([]logic.Expr, len(lit.Atom.Args))
	for i, t := range lit.Atom.Args {
		if isVar(t) {
			name := rep(t)
			if _, ok := subs[name]; !ok {
				v, _ := logic.NewVariable(fmt.Sprintf("W%d", i), t.NodeSort())
				subs[name] = v
			}
			terms[i] = subs[name]
		} else {
			terms[i] = t
		}
	}
	return NewLiteral(lit.Polarity, resolution.NewAtom(lit.Atom.RelName, terms...))
}

// ---------- Substitution helpers ----------

// SubstituteLit applies a variable substitution to a literal.
// subs maps variable names to replacement terms.
func SubstituteLit(lit *Literal, subs resolution.Env) *Literal {
	terms := make([]logic.Expr, len(lit.Atom.Args))
	changed := false
	for i, t := range lit.Atom.Args {
		if isVar(t) {
			if repl, ok := subs[rep(t)]; ok {
				terms[i] = repl
				changed = true
				continue
			}
		}
		terms[i] = t
	}
	if !changed {
		return lit
	}
	return NewLiteral(lit.Polarity, resolution.NewAtom(lit.Atom.RelName, terms...))
}

// SubstituteConstantsLit substitutes constants by name in a literal.
func SubstituteConstantsLit(lit *Literal, subs map[string]logic.Expr) *Literal {
	terms := make([]logic.Expr, len(lit.Atom.Args))
	changed := false
	for i, t := range lit.Atom.Args {
		if isConst(t) {
			if repl, ok := subs[rep(t)]; ok {
				terms[i] = repl
				changed = true
				continue
			}
		}
		terms[i] = t
	}
	if !changed {
		return lit
	}
	return NewLiteral(lit.Polarity, resolution.NewAtom(lit.Atom.RelName, terms...))
}

// SubstituteConstantsClause substitutes constants in each literal of a clause.
func SubstituteConstantsClause(cl []*Literal, subs map[string]logic.Expr) []*Literal {
	result := make([]*Literal, len(cl))
	for i, lit := range cl {
		result[i] = SubstituteConstantsLit(lit, subs)
	}
	return result
}

// ---------- Literal indexing ----------

// Index is a trie-like structure for indexing literals by polarity, relname, then arguments.
// index[polarity][relname] -> nested IndexNode
type Index [2]map[string]*IndexNode

// IndexNode is a node in the literal index trie.
type IndexNode struct {
	Units    []int        // unit queue indices stored here
	Watching []IndexEntry // (clause_idx, lit_idx) pairs watching here
	Children map[string]*IndexNode
}

// IndexEntry is a (clause_index, lit_index) pair.
type IndexEntry struct {
	ClauseIdx int
	LitIdx    int
}

// NewIndex creates a new empty index.
func NewIndex() *Index {
	return &Index{
		make(map[string]*IndexNode),
		make(map[string]*IndexNode),
	}
}

func newIndexNode() *IndexNode {
	return &IndexNode{Children: make(map[string]*IndexNode)}
}

// getOrCreate descends into child[key], creating if needed.
func (n *IndexNode) getOrCreate(key string) *IndexNode {
	if child, ok := n.Children[key]; ok {
		return child
	}
	child := newIndexNode()
	n.Children[key] = child
	return child
}

// indexLookup navigates the index to the leaf for the given literal.
func indexLookup(idx *Index, lit *Literal) *IndexNode {
	polMap := idx[lit.Polarity]
	rn := lit.Atom.RelName
	node, ok := polMap[rn]
	if !ok {
		node = newIndexNode()
		polMap[rn] = node
	}
	for _, t := range lit.Atom.Args {
		key := "V"
		if !isVar(t) {
			key = rep(t)
		}
		node = node.getOrCreate(key)
	}
	return node
}

// ---------- Equational theory methods ----------

// findTerm returns the representative of term under the equational theory.
func (ur *UnitRes) findTerm(term logic.Expr) logic.Expr {
	if ur.EquationalTheory == nil {
		return term
	}
	return ur.EquationalTheory.Find(term)
}

// groundMatch yields keys in index.Children whose representative matches term's representative.
func (ur *UnitRes) groundMatch(term logic.Expr, children map[string]*IndexNode) []string {
	if ur.EquationalTheory == nil {
		return []string{rep(term)}
	}
	trep := ur.EquationalTheory.Find(term)
	var result []string
	for key := range children {
		if ur.EquationalTheory.FindByName(key) == trep {
			result = append(result, key)
		}
	}
	return result
}

// ---------- Index search (generators as slices) ----------

// findSubsumedRec finds index nodes subsumed by the given terms.
func (ur *UnitRes) findSubsumedRec(node *IndexNode, terms []logic.Expr, idx int) []*IndexNode {
	if idx >= len(terms) {
		return []*IndexNode{node}
	}
	t := terms[idx]
	var results []*IndexNode
	if isVar(t) {
		for _, child := range node.Children {
			results = append(results, ur.findSubsumedRec(child, terms, idx+1)...)
		}
	} else {
		for _, key := range ur.groundMatch(t, node.Children) {
			if child, ok := node.Children[key]; ok {
				results = append(results, ur.findSubsumedRec(child, terms, idx+1)...)
			}
		}
	}
	return results
}

// findSubsumingRec finds index nodes that subsume the given terms.
func (ur *UnitRes) findSubsumingRec(node *IndexNode, terms []logic.Expr, idx int) []*IndexNode {
	if idx >= len(terms) {
		return []*IndexNode{node}
	}
	t := terms[idx]
	var results []*IndexNode
	if vChild, ok := node.Children["V"]; ok {
		results = append(results, ur.findSubsumingRec(vChild, terms, idx+1)...)
	}
	if !isVar(t) {
		for _, key := range ur.groundMatch(t, node.Children) {
			if child, ok := node.Children[key]; ok {
				results = append(results, ur.findSubsumingRec(child, terms, idx+1)...)
			}
		}
	}
	return results
}

// findUnifyingRec finds index nodes that unify with the given terms.
func (ur *UnitRes) findUnifyingRec(node *IndexNode, terms []logic.Expr, idx int) []*IndexNode {
	if idx >= len(terms) {
		return []*IndexNode{node}
	}
	t := terms[idx]
	var results []*IndexNode
	if isVar(t) || (isConst(t) && strings.HasPrefix(rep(t), "__v")) {
		for _, child := range node.Children {
			results = append(results, ur.findUnifyingRec(child, terms, idx+1)...)
		}
	} else {
		if vChild, ok := node.Children["V"]; ok {
			results = append(results, ur.findUnifyingRec(vChild, terms, idx+1)...)
		}
		for _, key := range ur.groundMatch(t, node.Children) {
			if child, ok := node.Children[key]; ok {
				results = append(results, ur.findUnifyingRec(child, terms, idx+1)...)
			}
		}
	}
	return results
}

// FindSubsumed finds index nodes whose literals are subsumed by lit.
func (ur *UnitRes) FindSubsumed(idx *Index, lit *Literal) []*IndexNode {
	polMap := idx[lit.Polarity]
	node, ok := polMap[lit.Atom.RelName]
	if !ok {
		return nil
	}
	return ur.findSubsumedRec(node, lit.Atom.Args, 0)
}

// FindSubsuming finds index nodes whose literals subsume lit.
func (ur *UnitRes) FindSubsuming(idx *Index, lit *Literal) []*IndexNode {
	polMap := idx[lit.Polarity]
	node, ok := polMap[lit.Atom.RelName]
	if !ok {
		return nil
	}
	return ur.findSubsumingRec(node, lit.Atom.Args, 0)
}

// FindUnifying finds index nodes whose literals unify with lit.
func (ur *UnitRes) FindUnifying(idx *Index, lit *Literal) []*IndexNode {
	polMap := idx[lit.Polarity]
	node, ok := polMap[lit.Atom.RelName]
	if !ok {
		return nil
	}
	return ur.findUnifyingRec(node, lit.Atom.Args, 0)
}

// ---------- Literal representation under equational theory ----------

// litRep returns the literal with its arguments replaced by their
// representatives in the equational theory.
func (ur *UnitRes) litRep(lit *Literal) *Literal {
	if ur.EquationalTheory == nil {
		return lit
	}
	terms := make([]logic.Expr, len(lit.Atom.Args))
	for i, a := range lit.Atom.Args {
		if isVar(a) {
			terms[i] = a
		} else {
			terms[i] = ur.EquationalTheory.Find(a)
		}
	}
	return NewLiteral(lit.Polarity, resolution.NewAtom(lit.Atom.RelName, terms...))
}

// ---------- Subsumption ----------

// termSubsume tries to match term1 to term2 (env only operates on term1 variables).
func termSubsume(term1, term2 logic.Expr, env map[string]logic.Expr) bool {
	if isConst(term1) {
		if !isConst(term2) || rep(term1) != rep(term2) {
			return false
		}
		return true
	}
	// term1 is a variable
	name := rep(term1)
	if prev, ok := env[name]; ok {
		return prev.Equal(term2)
	}
	env[name] = term2
	return true
}

// litSubsume tries to make lit1 subsume lit2 (env only operates on lit1 variables).
func litSubsume(lit1, lit2 *Literal, env map[string]logic.Expr) bool {
	if lit1.Polarity != lit2.Polarity || lit1.Atom.RelName != lit2.Atom.RelName ||
		len(lit1.Atom.Args) != len(lit2.Atom.Args) {
		return false
	}
	for i := range lit1.Atom.Args {
		if !termSubsume(lit1.Atom.Args[i], lit2.Atom.Args[i], env) {
			return false
		}
	}
	return true
}

// atomSubsume checks if atom at1 subsumes at2 (at1 is more general).
func atomSubsume(at1, at2 *resolution.Atom) bool {
	env := make(map[string]logic.Expr)
	if at1.RelName != at2.RelName || len(at1.Args) != len(at2.Args) {
		return false
	}
	for i := range at1.Args {
		if !termSubsume(at1.Args[i], at2.Args[i], env) {
			return false
		}
	}
	return true
}

// litSubsumeModEq checks subsumption modulo the equational theory.
func (ur *UnitRes) litSubsumeModEq(lit1, lit2 *Literal, env map[string]logic.Expr) bool {
	return litSubsume(ur.litRep(lit1), ur.litRep(lit2), env)
}

// ---------- Simplify / Tautology ----------

// rewriteClause substitutes variable v with term t in each literal of a clause.
func rewriteClause(cl []*Literal, v logic.Expr, t logic.Expr) []*Literal {
	subs := resolution.Env{rep(v): t}
	result := make([]*Literal, len(cl))
	for i, lit := range cl {
		result[i] = SubstituteLit(lit, subs)
	}
	return result
}

// SimplifyClause simplifies a clause by rewriting equalities and removing
// vacuous/tautological literals.
func SimplifyClause(cl []*Literal) []*Literal {
	for _, lit := range cl {
		if lit.Polarity == 0 && lit.Atom.RelName == "=" && len(lit.Atom.Args) == 2 {
			for i := 0; i < 2; i++ {
				if isVar(lit.Atom.Args[i]) {
					cl = rewriteClause(cl, lit.Atom.Args[i], lit.Atom.Args[1-i])
					break
				}
			}
		}
	}
	if anyTaut(cl) {
		// Return a single tautological literal.
		return []*Literal{NewLiteral(1, resolution.NewAtom("=",
			logic.NewConst("__true", logic.TopS),
			logic.NewConst("__true", logic.TopS)))}
	}
	return removeDuplicatesAndVac(cl)
}

func anyTaut(cl []*Literal) bool {
	for _, lit := range cl {
		if isTautLit(lit) {
			return true
		}
	}
	return false
}

func removeDuplicatesAndVac(cl []*Literal) []*Literal {
	var result []*Literal
	for _, lit := range cl {
		if isVacLit(lit) {
			continue
		}
		dup := false
		for _, existing := range result {
			if LitEqual(lit, existing) {
				dup = true
				break
			}
		}
		if !dup {
			result = append(result, lit)
		}
	}
	return result
}

// IsTautology returns true if the clause is tautological.
func IsTautology(cl []*Literal) bool {
	for _, lit := range cl {
		if isTautLit(lit) {
			return true
		}
		for _, lit2 := range cl {
			if lit.Polarity != lit2.Polarity && AtomEqual(lit.Atom, lit2.Atom) {
				return true
			}
		}
	}
	return false
}

// ---------- keep_atom / keep_lit ----------

// isSkolemName returns true if the name contains "__" (Skolem convention).
func isSkolemName(name string) bool {
	return strings.Contains(name, "__")
}

// keepAtom returns false if the atom or any of its arguments is Skolem.
func keepAtom(atom *resolution.Atom) bool {
	if isSkolemName(atom.RelName) {
		return false
	}
	for _, t := range atom.Args {
		if isConst(t) && isSkolemName(rep(t)) {
			return false
		}
	}
	return true
}

// keepLit returns keepAtom(lit.Atom).
func keepLit(lit *Literal) bool {
	return keepAtom(lit.Atom)
}

// ---------- UnitRes ----------

// UnitRes performs unit resolution with an equational theory.
type UnitRes struct {
	Clauses      [][]*Literal // multi-literal clauses
	index        *Index
	UnitQueue    []*Literal // unit literals (propagated + pending)
	Subsumed     []int      // indices of subsumed clauses
	Unsat        bool
	UsedUnits    int // number of units already propagated
	Stack        []stackFrame
	clausesGen   []int // generation per clause
	unitQueueGen []int // generation per unit
	unitIDs      map[int]bool

	DetectedSpecializations []*resolution.Atom

	EquationalTheory *congclos.CongClos
	unitTermIndex    map[string][]int // maps term rep -> unit queue indices

	litConsing *LitConsing

	// Verbose controls debug printing. Python: verbose global.
	Verbose bool
	// NewSpecialization mirrors Python's new_specialization = True.
	NewSpecialization bool
}

type stackFrame struct {
	numClauses  int
	numUnits    int
	usedUnits   int
	numSubsumed int
	unsat       bool
}

// NewUnitRes creates a UnitRes and adds all initial clauses.
func NewUnitRes(clauses [][]*Literal) *UnitRes {
	ur := &UnitRes{
		index:             NewIndex(),
		unitIDs:           make(map[int]bool),
		EquationalTheory:  congclos.New(),
		unitTermIndex:     make(map[string][]int),
		litConsing:        NewLitConsing(),
		NewSpecialization: true, // Python default
	}

	for _, cl := range clauses {
		ur.AddClause(cl, 0)
	}
	return ur
}

// Push saves the current state for later Pop.
func (ur *UnitRes) Push() {
	ur.Stack = append(ur.Stack, stackFrame{
		numClauses:  len(ur.Clauses),
		numUnits:    len(ur.UnitQueue),
		usedUnits:   ur.UsedUnits,
		numSubsumed: len(ur.Subsumed),
		unsat:       ur.Unsat,
	})
	ur.EquationalTheory.Push()
}

// Pop restores the state saved by the most recent Push.
func (ur *UnitRes) Pop() {
	if len(ur.Stack) == 0 {
		return
	}
	frame := ur.Stack[len(ur.Stack)-1]
	ur.Stack = ur.Stack[:len(ur.Stack)-1]

	// Deindex units that were added after the push
	for i := frame.numUnits; i < len(ur.UnitQueue); i++ {
		ur.deindexUnitTerms(i)
	}

	ur.Clauses = ur.Clauses[:frame.numClauses]
	ur.clausesGen = ur.clausesGen[:frame.numClauses]
	ur.UnitQueue = ur.UnitQueue[:frame.numUnits]
	ur.unitQueueGen = ur.unitQueueGen[:frame.numUnits]
	ur.UsedUnits = frame.usedUnits
	ur.Subsumed = ur.Subsumed[:frame.numSubsumed]
	ur.Unsat = frame.unsat

	ur.EquationalTheory.Pop()
}

// ---------- Internal methods ----------

func (ur *UnitRes) unitSubsumedBasic(lit *Literal) bool {
	subsuming := ur.FindSubsuming(ur.index, lit)
	for _, node := range subsuming {
		for _, litIdx := range node.Units {
			lit2 := ur.UnitQueue[litIdx]
			if ur.litSubsumeModEq(lit2, lit, make(map[string]logic.Expr)) {
				return true
			}
		}
	}
	return false
}

func (ur *UnitRes) unitSubsumed(lit *Literal) bool {
	if ur.unitSubsumedBasic(lit) {
		return true
	}
	if isDisequalityLit(lit) && len(lit.Atom.Args) == 2 {
		return ur.unitSubsumedBasic(swapArgsLit(lit))
	}
	return false
}

func (ur *UnitRes) addClauseBasic(cl []*Literal, gen int) {
	n := len(cl)
	if n == 0 {
		ur.Unsat = true
		ur.UsedUnits = len(ur.UnitQueue)
		return
	}
	if n == 1 {
		lit := ur.litRep(CanonizeLiteralVars(cl[0]))
		if isTautLit(lit) || ur.unitSubsumed(lit) {
			return
		}
		litid := ur.litConsing.LitID(lit)
		node := indexLookup(ur.index, lit)
		node.Units = append(node.Units, len(ur.UnitQueue))
		ur.UnitQueue = append(ur.UnitQueue, lit)
		ur.unitQueueGen = append(ur.unitQueueGen, gen)
		ur.unitIDs[litid] = true
		if ur.Verbose {
			fmt.Printf("added %s %d\n", lit, litid)
		}
		return
	}
	// multi-literal clause
	i := len(ur.Clauses)
	ur.Clauses = append(ur.Clauses, cl)
	ur.clausesGen = append(ur.clausesGen, gen)
	ur.reindex(i)
}

// AddClause adds a clause and applies hyper binary resolution with transitivity.
func (ur *UnitRes) AddClause(cl []*Literal, gen int) {
	ur.addClauseBasic(cl, gen)
	// hyper binary resolution with transitivity axiom
	if len(cl) == 2 && isGroundClause(cl) {
		for i := 0; i < 2; i++ {
			lhs, rhs := cl[i], cl[1-i]
			if isDisequalityLit(lhs) && isEqualityLit(rhs) && len(lhs.Atom.Args) == 2 {
				for j := 0; j < 2; j++ {
					subs := map[string]logic.Expr{rep(lhs.Atom.Args[j]): lhs.Atom.Args[1-j]}
					newRHS := SubstituteConstantsLit(rhs, subs)
					if !LitEqual(rhs, newRHS) {
						newCl := []*Literal{lhs, newRHS}
						if ur.Verbose {
							fmt.Printf("applied transitivity: %v\n", cl)
						}
						ur.addClauseBasic(newCl, gen)
					}
				}
			}
		}
	}
}

func (ur *UnitRes) getWatching(lit *Literal) *IndexNode {
	return indexLookup(ur.index, lit)
}

func (ur *UnitRes) deindex(i int) {
	for _, lit := range ur.Clauses[i] {
		node := ur.getWatching(lit)
		// Remove (i, j) entries from Watching
		var kept []IndexEntry
		for _, e := range node.Watching {
			if e.ClauseIdx != i {
				kept = append(kept, e)
			}
		}
		node.Watching = kept
	}
}

func (ur *UnitRes) reindex(i int) {
	for j, lit := range ur.Clauses[i] {
		node := ur.getWatching(lit)
		node.Watching = append(node.Watching, IndexEntry{ClauseIdx: i, LitIdx: j})
	}
}

func (ur *UnitRes) indexUnitTerms(i int) {
	used := make(map[string]bool)
	for _, t := range ur.UnitQueue[i].Atom.Args {
		if !isVar(t) {
			name := rep(t)
			if !used[name] {
				used[name] = true
				ur.unitTermIndex[name] = append(ur.unitTermIndex[name], i)
			}
		}
	}
	// Also index by relname
	rn := ur.UnitQueue[i].Atom.RelName
	ur.unitTermIndex[rn] = append(ur.unitTermIndex[rn], i)
}

func (ur *UnitRes) deindexUnitTerms(i int) {
	if i >= len(ur.UnitQueue) {
		return
	}
	used := make(map[string]bool)
	for _, t := range ur.UnitQueue[i].Atom.Args {
		if !isVar(t) {
			name := rep(t)
			if !used[name] {
				used[name] = true
				ur.unitTermIndex[name] = removeInt(ur.unitTermIndex[name], i)
			}
		}
	}
	rn := ur.UnitQueue[i].Atom.RelName
	ur.unitTermIndex[rn] = removeInt(ur.unitTermIndex[rn], i)
}

func removeInt(s []int, val int) []int {
	for i, v := range s {
		if v == val {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (ur *UnitRes) updateEquationalTheory(lit *Literal) {
	if lit.Polarity == 1 && lit.Atom.RelName == "=" && len(lit.Atom.Args) == 2 {
		t0, t1 := lit.Atom.Args[0], lit.Atom.Args[1]
		if isConst(t0) && isConst(t1) {
			ur.EquationalTheory.Union(t0, t1)
			if ur.Verbose {
				fmt.Printf("merged %s %s\n", t0, t1)
			}
		}
	}
}

func (ur *UnitRes) unitSubsumedByUsed(lit *Literal) bool {
	subsuming := ur.FindSubsuming(ur.index, lit)
	for _, node := range subsuming {
		for _, litIdx := range node.Units {
			if litIdx < ur.UsedUnits {
				lit2 := ur.UnitQueue[litIdx]
				if ur.litSubsumeModEq(lit2, lit, make(map[string]logic.Expr)) {
					return true
				}
			}
		}
	}
	return false
}

func (ur *UnitRes) subsumedByUsedLit(a1, a2 interface{}, clause []*Literal) bool {
	for _, lit := range clause {
		if ur.unitSubsumedByUsed(lit) {
			return true
		}
		inv := &Literal{Polarity: 1 - lit.Polarity, Atom: lit.Atom}
		if ur.unitSubsumedByUsed(inv) {
			return true
		}
	}
	return false
}

func (ur *UnitRes) allowEqs(lit *Literal, eqs []*resolution.Atom, isUnit bool, other interface{}) bool {
	if lit.Atom.RelName == "=" {
		return len(eqs) == 0
	}
	allSpec := true
	for _, atom := range eqs {
		if len(atom.Args) < 2 || !strings.HasPrefix(rep(atom.Args[0]), "__v") || strings.HasPrefix(rep(atom.Args[1]), "__v") {
			allSpec = false
			break
		}
	}
	if allSpec {
		if !ur.NewSpecialization || isUnit {
			return true
		}
		ur.DetectedSpecializations = append(ur.DetectedSpecializations, eqs...)
	}
	return len(eqs) == 0
}

// PropagateEquality propagates a positive ground equality literal.
func (ur *UnitRes) PropagateEquality(lit *Literal, gen int) {
	lit = ur.litRep(lit)
	if isTautLit(lit) {
		return
	}
	t0, t1 := lit.Atom.Args[0], lit.Atom.Args[1]
	if rep(t0) < rep(t1) {
		t0, t1 = t1, t0
	}
	for _, litIdx := range copyInts(ur.unitTermIndex[rep(t0)]) {
		lit2 := ur.UnitQueue[litIdx]
		subs := map[string]logic.Expr{rep(t0): t1}
		lit3 := SubstituteConstantsLit(lit2, subs)
		if !LitEqual(lit2, lit3) {
			newCl := []*Literal{lit3}
			if ur.Verbose {
				fmt.Printf("rewrite! %s,%s -> %v\n", lit, lit2, newCl)
			}
			newGen := maxInt(gen+1, ur.unitQueueGen[litIdx])
			ur.AddClause(newCl, newGen)
			if ur.Unsat {
				return
			}
		}
	}
	ur.updateEquationalTheory(lit)
	for _, litIdx := range copyInts(ur.unitTermIndex[rep(t1)]) {
		lit2 := ur.UnitQueue[litIdx]
		if ur.Verbose {
			fmt.Printf("re-propagate: %s\n", lit2)
		}
		ur.PropagateLit(lit2, maxInt(gen+1, ur.unitQueueGen[litIdx]), nil)
		if ur.Unsat {
			return
		}
	}
}

// PropagateLit performs unit resolution using a single literal.
func (ur *UnitRes) PropagateLit(lit *Literal, gen int, specs map[string]logic.Expr) {
	if ur.EquationalTheory != nil && isGroundEqualityLit(lit) {
		ur.PropagateEquality(lit, gen)
		return
	}

	// For equality literals, also consider the symmetric version
	lits := []*Literal{lit}
	if lit.Atom.RelName == "=" && len(lit.Atom.Args) == 2 {
		lits = append(lits, NewLiteral(lit.Polarity,
			resolution.NewAtom("=", lit.Atom.Args[1], lit.Atom.Args[0])))
	}

	for _, lit := range lits {
		keep := keepLit(lit)
		// Find clauses that might resolve with the negation of lit
		negLit := &Literal{Polarity: 1 - lit.Polarity, Atom: lit.Atom}
		indices := ur.FindUnifying(ur.index, negLit) // snapshot
		lit = ur.litRep(CanonizeLiteralUnique(lit))

		for _, idxNode := range indices {
			// Process watching list
			wl := make([]IndexEntry, len(idxNode.Watching))
			copy(wl, idxNode.Watching) // copy to avoid mutation during iteration
			for _, entry := range wl {
				i, j := entry.ClauseIdx, entry.LitIdx
				if i >= len(ur.Clauses) {
					continue
				}
				cl := ur.Clauses[i]
				if j >= len(cl) {
					continue
				}
				lit2 := ur.litRep(cl[j])
				if lit2.Polarity != 1-lit.Polarity || lit2.Atom.RelName != lit.Atom.RelName {
					continue
				}
				// Skip trivially universal equalities
				if isEqualityLit(lit2) && len(lit2.Atom.Args) == 2 &&
					isVar(lit2.Atom.Args[0]) && isVar(lit2.Atom.Args[1]) {
					continue
				}
				match, subs, eqs := resolution.MGUEq(lit.Atom, lit2.Atom)
				if match && ur.allowEqs(lit, eqs, false, cl) {
					if atomSubsume(lit.Atom, lit2.Atom) {
						ur.deindex(i)
						ur.Subsumed = append(ur.Subsumed, i)
					}
					var newCl []*Literal
					for k, lit1 := range cl {
						if k != j {
							newCl = append(newCl, SubstituteLit(lit1, subs))
						}
					}
					for _, eq := range eqs {
						newCl = append(newCl, NewLiteral(0, eq))
					}
					if !ur.NewSpecialization && specs != nil {
						newCl = SubstituteConstantsClause(newCl, specs)
					}
					newCl = SimplifyClause(newCl)
					if !IsTautology(newCl) && !ur.subsumedByUsedLit(lit, cl, newCl) {
						if ur.Verbose {
							fmt.Printf("%s,%v -> %v\n", lit, cl, newCl)
						}
						newGen := maxInt(gen+1, ur.clausesGen[i])
						ur.AddClause(newCl, newGen)
						if ur.Unsat {
							return
						}
					}
				}
			}

			if keep {
				ur.resolveUnits(lit, idxNode.Units, gen, false)
				if ur.Unsat {
					return
				}
			}
		}

		// If not keeping, resolve against units indexed by relname
		if !keep {
			others := make([]int, 0)
			for _, idx := range ur.unitTermIndex[lit.Atom.RelName] {
				if ur.UnitQueue[idx].Polarity != lit.Polarity {
					others = append(others, idx)
				}
			}
			ur.resolveUnits(lit, others, gen, true)
			if ur.Unsat {
				return
			}
		}
	}
}

func (ur *UnitRes) resolveUnits(lit *Literal, units []int, gen int, allowUnitDiseqs bool) {
	unitsCopy := copyInts(units)
	for _, litIdx := range unitsCopy {
		if litIdx >= len(ur.UnitQueue) {
			continue
		}
		lit2 := ur.litRep(ur.UnitQueue[litIdx])
		match, _, eqs := resolution.MGUEq(lit.Atom, lit2.Atom)
		if match && (ur.allowEqs(lit, eqs, true, nil) ||
			(allowUnitDiseqs && len(eqs) == 1 && keepAtom(eqs[0]))) {
			var newCl []*Literal
			for _, eq := range eqs {
				newCl = append(newCl, NewLiteral(0, eq))
			}
			if !ur.subsumedByUsedLit(lit, lit2, newCl) {
				if ur.Verbose {
					fmt.Printf("units resolve! %s,%s -> %v\n", lit, lit2, newCl)
				}
				newGen := maxInt(gen+1, ur.unitQueueGen[litIdx])
				ur.AddClause(newCl, newGen)
				if ur.Unsat {
					return
				}
			}
		}
	}
}

// Propagate runs unit propagation to a fixed point.
func (ur *UnitRes) Propagate(specs map[string]logic.Expr) {
	for ur.UsedUnits < len(ur.UnitQueue) {
		litNum := ur.UsedUnits
		lit := ur.UnitQueue[litNum]
		gen := ur.unitQueueGen[litNum]
		ur.PropagateLit(lit, gen, specs)
		if ur.UsedUnits < litNum+1 {
			ur.UsedUnits = litNum + 1
		}
		// Maintain invariant: all used units have terms indexed
		for j := litNum; j < ur.UsedUnits; j++ {
			ur.indexUnitTerms(j)
		}
	}
}

// UsedUnitLiterals returns the unit literals that have been propagated.
func (ur *UnitRes) UsedUnitLiterals() []*Literal {
	if ur.UsedUnits > len(ur.UnitQueue) {
		return ur.UnitQueue
	}
	return ur.UnitQueue[:ur.UsedUnits]
}

// ---------- Helpers ----------

func copyInts(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
