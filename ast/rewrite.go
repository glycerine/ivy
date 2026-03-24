package ast

// AST rewriting infrastructure.
// Ported from ivy_ast.py: ast_rewrite, AstRewriteSubstConstants,
// AstRewriteSubstConstantsParams, AstRewriteSubstPrefix, AstRewritePostfix,
// AstRewriteAddParams, plus helpers: parse_name, subst_subscripts,
// base_name_differs, copy_attributes_ast_ref, compose_atoms, rewrite_sort.

import (
	"fmt"
	"regexp"
	"strings"

	iu "github.com/glycerine/goivy/ivyutils"
)

// IvyComposeCharacter is the character used to compose names (e.g., "module.name").
// Python: ivy_compose_character = '.' (for versions > 1.1)
const IvyComposeCharacter = "."

// --- Name tree types for parse_name ---

// NameTree is a parsed representation of a structured name like "a.b[c]".
// Python: parse_name returns a tree of Symbol/Dot/Bracket nodes.
type NameTree interface {
	Unparse() string // reconstruct the string name
	// Subst applies a substitution map to the name tree.
	// root=true means this is the leftmost (root) component.
	Subst(subst map[string]string, root bool) NameTree
	// Apply applies a function to the name components.
	Apply(fun func(string) string, root bool) NameTree
}

// NameSymbol is a leaf name component.
// Python: Symbol in parse_name context (uses .rep, .sort, .subst, .unparse, .apply)
type NameSymbol struct {
	Name string
}

func (ns *NameSymbol) Unparse() string { return ns.Name }

func (ns *NameSymbol) Subst(subst map[string]string, root bool) NameTree {
	if !root {
		return ns
	}
	if repl, ok := subst[ns.Name]; ok {
		if repl == "this" {
			return &NameThis{}
		}
		return &NameSymbol{Name: repl}
	}
	return ns
}

func (ns *NameSymbol) Apply(fun func(string) string, root bool) NameTree {
	if root {
		return &NameSymbol{Name: fun(ns.Name)}
	}
	return ns
}

// NameThis represents the "this" keyword in a name tree.
// Python: This class — This.unparse() returns self (the This object).
type NameThis struct{}

func (nt *NameThis) Unparse() string { return "this" }
func (nt *NameThis) Subst(subst map[string]string, root bool) NameTree {
	return nt
}
func (nt *NameThis) Apply(fun func(string) string, root bool) NameTree {
	return nt
}

// IsNameThis checks if a NameTree is a This node.
func IsNameThis(n NameTree) bool {
	_, ok := n.(*NameThis)
	return ok
}

// NameDot represents a dotted name: lhs.rhs
// Python: Dot(AST) with subst, unparse, apply
type NameDot struct {
	Lhs, Rhs NameTree
}

func (nd *NameDot) Unparse() string {
	return nd.Lhs.Unparse() + IvyComposeCharacter + nd.Rhs.Unparse()
}

func (nd *NameDot) Subst(subst map[string]string, root bool) NameTree {
	lhs := nd.Lhs.Subst(subst, root)
	rhs := nd.Rhs.Subst(subst, false)
	if IsNameThis(lhs) {
		return rhs
	}
	return &NameDot{Lhs: lhs, Rhs: rhs}
}

func (nd *NameDot) Apply(fun func(string) string, root bool) NameTree {
	return &NameDot{
		Lhs: nd.Lhs.Apply(fun, root),
		Rhs: nd.Rhs.Apply(fun, false),
	}
}

// NameBracket represents a bracketed subscript: lhs[rhs]
// Python: Bracket(AST) with subst, unparse, apply
type NameBracket struct {
	Lhs, Rhs NameTree
}

func (nb *NameBracket) Unparse() string {
	lhs := nb.Lhs.Unparse()
	if lhs == "this" {
		lhs = "this"
	}
	return lhs + "[" + nb.Rhs.Unparse() + "]"
}

func (nb *NameBracket) Subst(subst map[string]string, root bool) NameTree {
	return &NameBracket{
		Lhs: nb.Lhs.Subst(subst, root),
		Rhs: nb.Rhs.Subst(subst, true),
	}
}

func (nb *NameBracket) Apply(fun func(string) string, root bool) NameTree {
	return &NameBracket{
		Lhs: nb.Lhs.Apply(fun, root),
		Rhs: nb.Rhs.Apply(fun, true),
	}
}

// symbolCharsParser matches characters that are not [ ] or .
// Python: symbol_chars_parser = re.compile(r'[^\[\]\.]*')
var symbolCharsParser = regexp.MustCompile(`[^\[\].]*`)

// ParseName parses a structured name like "a.b[c]" into a NameTree.
// Python: parse_name(name)
func ParseName(name string) NameTree {
	res, _ := parseNameRecur(name, 0)
	return res
}

func parseNameRecur(name string, pos int) (NameTree, int) {
	loc := symbolCharsParser.FindStringIndex(name[pos:])
	matchStr := ""
	endPos := pos
	if loc != nil {
		matchStr = name[pos+loc[0] : pos+loc[1]]
		endPos = pos + loc[1]
	}
	var pref NameTree = &NameSymbol{Name: matchStr}

	for endPos < len(name) && name[endPos] != ']' {
		if name[endPos] == '[' {
			suff, newPos := parseNameRecur(name, endPos+1)
			pref = &NameBracket{Lhs: pref, Rhs: suff}
			endPos = newPos + 1 // skip ']'
		} else if name[endPos] == '.' {
			suff, newPos := parseNameRecur(name, endPos+1)
			pref = &NameDot{Lhs: pref, Rhs: suff}
			endPos = newPos
		} else {
			break
		}
	}
	return pref, endPos
}

// SubstSubscripts applies a substitution to structured name components.
// Python: subst_subscripts(s, subst)
func SubstSubscripts(s string, subst map[string]string) string {
	if s == "this" || strings.HasPrefix(s, "\"") {
		return s
	}
	tree := ParseName(s)
	result := tree.Subst(subst, true).Unparse()
	return result
}

// BaseNameDiffers checks if the base names of two strings differ.
// Python: base_name_differs(x, y)
func BaseNameDiffers(x, y string) bool {
	return myBaseName(x) != myBaseName(y)
}

func myBaseName(x string) string {
	if x == "this" {
		return x
	}
	return rewriteBaseName(x)
}

func rewriteBaseName(name string) string {
	parts := strings.Split(name, ".")
	return parts[0]
}

// CopyAttributesAstRef copies lineno and sort from source to dest.
// Python: copy_attributes_ast_ref(x, y)
func CopyAttributesAstRef(src, dst Node) {
	dst.SetLineno(src.GetLineno())
}

// ComposeAtoms composes a prefix atom with another atom.
// Python: compose_atoms(pr, atom)
func ComposeAtoms(pr, atom *Atom) *Atom {
	if atom == nil {
		return pr
	}
	var hname string
	if atom.Rep == "this" {
		hname = pr.Rep
	} else {
		hname = composeNames(pr.Rep, atom.Rep)
	}
	args := make([]Node, 0, len(pr.Terms)+len(atom.Terms))
	args = append(args, pr.Terms...)
	args = append(args, atom.Terms...)
	res := NewAtom(hname, args...)
	res.Base = atom.Base // copy_attributes_ast
	res.ASort = atom.ASort
	return res
}

// ComposeAtomsGeneric composes two nodes (Atom or App) faithfully matching
// Python's compose_atoms(pr, atom):
//   hname = pr.rep if isinstance(atom.rep, This) else compose_names(pr.rep, atom.rep)
//   args = pr.args + atom.args
//   res = type(atom)(hname, args)
//   copy_attributes_ast(atom, res)
// The result type matches the second argument's type.
func ComposeAtomsGeneric(pr, atom Node) Node {
	if atom == nil {
		return pr
	}
	prRep := nodeRepStr(pr)
	atomRep := nodeRepStr(atom)
	prArgs := pr.Args()
	atomArgs := atom.Args()

	var hname string
	if atomRep == "this" {
		hname = prRep
	} else {
		hname = composeNames(prRep, atomRep)
	}

	args := make([]Node, 0, len(prArgs)+len(atomArgs))
	args = append(args, prArgs...)
	args = append(args, atomArgs...)

	// Python: res = type(atom)(hname, args) — result is same type as atom
	switch a := atom.(type) {
	case *Atom:
		res := NewAtom(hname, args...)
		res.Base = a.Base
		res.ASort = a.ASort
		return res
	case *App:
		res := NewApp(NewSymbol(hname, nil), args...)
		res.Base = a.Base
		res.ASort = a.ASort
		return res
	default:
		// Fallback: create Atom
		res := NewAtom(hname, args...)
		return res
	}
}

// nodeRepStr extracts the rep string from an Atom or App.
func nodeRepStr(n Node) string {
	switch x := n.(type) {
	case *Atom:
		return x.Rep
	case *App:
		if sym, ok := x.Rep.(*Symbol); ok {
			return sym.Rep
		}
		return fmt.Sprint(x.Rep)
	default:
		return fmt.Sprint(n)
	}
}

func composeNames(names ...string) string {
	var parts []string
	for _, n := range names {
		if n != "" {
			parts = append(parts, n)
		}
	}
	return strings.Join(parts, ".")
}

// --- AstRewriter interface ---

// AstRewriter is the interface for all AST rewrite strategies.
// Python: duck-typed objects with rewrite_name and rewrite_atom methods.
type AstRewriter interface {
	RewriteName(name string) string
	RewriteAtom(atom *Atom, always bool) *Atom
}

// --- Concrete rewriters ---

// AstRewriteSubstConstants substitutes constant names.
// Python: AstRewriteSubstConstants
type AstRewriteSubstConstants struct {
	Subst map[string]Node // maps atom.rep → replacement node
}

func NewAstRewriteSubstConstants(subst map[string]Node) *AstRewriteSubstConstants {
	return &AstRewriteSubstConstants{Subst: subst}
}

func (r *AstRewriteSubstConstants) RewriteName(name string) string {
	return name
}

func (r *AstRewriteSubstConstants) RewriteAtom(atom *Atom, always bool) *Atom {
	if len(atom.Terms) == 0 {
		if repl, ok := r.Subst[atom.Rep]; ok {
			if a, ok := repl.(*Atom); ok {
				return a
			}
		}
	}
	return atom
}

// AstRewriteSubstConstantsParams substitutes constants and parameter subscripts.
// Python: AstRewriteSubstConstantsParams
type AstRewriteSubstConstantsParams struct {
	Subst  map[string]Node
	PSubst map[string]string
}

func NewAstRewriteSubstConstantsParams(subst map[string]Node, psubst map[string]string) *AstRewriteSubstConstantsParams {
	return &AstRewriteSubstConstantsParams{Subst: subst, PSubst: psubst}
}

func (r *AstRewriteSubstConstantsParams) RewriteName(name string) string {
	return SubstSubscripts(name, r.PSubst)
}

func (r *AstRewriteSubstConstantsParams) RewriteAtom(atom *Atom, always bool) *Atom {
	if len(atom.Terms) == 0 {
		if repl, ok := r.Subst[atom.Rep]; ok {
			if a, ok := repl.(*Atom); ok {
				return a
			}
		}
	}
	return atom
}

// AstRewriteSubstPrefix substitutes and prefixes names.
// Python: AstRewriteSubstPrefix
type AstRewriteSubstPrefix struct {
	Subst  map[string]string
	Pref   *Atom           // prefix atom (nil for no prefix)
	ToPref map[string]bool // names that should be prefixed (nil = all)
	Static map[string]bool // static names (get prefix without args)
	Local  bool            // set during SchemaBody rewriting
}

func NewAstRewriteSubstPrefix(subst map[string]string, pref *Atom) *AstRewriteSubstPrefix {
	return &AstRewriteSubstPrefix{Subst: subst, Pref: pref}
}

func (r *AstRewriteSubstPrefix) RewriteName(name string) string {
	return SubstSubscripts(name, r.Subst)
}

func (r *AstRewriteSubstPrefix) PrefixStr(name string, always bool) string {
	if name == "this" && r.Pref != nil {
		return r.Pref.Rep
	}
	if r.Pref == nil {
		return name
	}
	if !always && r.ToPref != nil && !r.ToPref[name] {
		return name
	}
	return composeNames(r.Pref.Rep, name)
}

func (r *AstRewriteSubstPrefix) RewriteAtom(atom *Atom, always bool) *Atom {
	// First handle name rewriting for non-This, non-quoted reps
	if atom.Rep != "this" && !strings.HasPrefix(atom.Rep, "\"") {
		tree := ParseName(atom.Rep)
		newName := tree.Apply(func(x string) string {
			return r.PrefixStr(x, always)
		}, false).Unparse()
		if newName != atom.Rep {
			atom = atom.Rename(newName)
		}
	}
	// Then handle prefixing
	if r.Pref == nil {
		return atom
	}
	if !always && r.ToPref != nil {
		if atom.Rep != "this" {
			parts := strings.Split(atom.Rep, ".")
			if len(parts) > 0 && !r.ToPref[parts[0]] {
				return atom
			}
		}
	}
	thePref := r.Pref
	if r.Static != nil && r.Static[atom.Rep] {
		thePref = NewAtom(thePref.Rep) // no args
	}
	return ComposeAtoms(thePref, atom)
}

// AstRewritePostfix appends a postfix atom.
// Python: AstRewritePostfix
type AstRewritePostfix struct {
	Post *Atom
}

func NewAstRewritePostfix(post *Atom) *AstRewritePostfix {
	return &AstRewritePostfix{Post: post}
}

func (r *AstRewritePostfix) RewriteName(name string) string {
	return name
}

func (r *AstRewritePostfix) RewriteAtom(atom *Atom, always bool) *Atom {
	return ComposeAtoms(atom, r.Post)
}

// AstRewriteAddParams adds parameters to atoms.
// Python: AstRewriteAddParams
type AstRewriteAddParams struct {
	Params []Node
}

func NewAstRewriteAddParams(params []Node) *AstRewriteAddParams {
	return &AstRewriteAddParams{Params: params}
}

func (r *AstRewriteAddParams) RewriteName(name string) string {
	return name
}

func (r *AstRewriteAddParams) RewriteAtom(atom *Atom, always bool) *Atom {
	newArgs := make([]Node, 0, len(atom.Terms)+len(r.Params))
	newArgs = append(newArgs, atom.Terms...)
	newArgs = append(newArgs, r.Params...)
	c := atom.Clone(newArgs).(*Atom)
	return c
}

// RewriteSort rewrites a sort name using a rewriter.
// Python: rewrite_sort(rewrite, orig_sort)
func RewriteSort(rewrite AstRewriter, origSort string) string {
	sort := rewrite.RewriteName(origSort)
	if BaseNameDiffers(sort, origSort) {
		return sort
	}
	sort = rewrite.RewriteAtom(NewAtom(sort), false).Rep
	return sort
}

// AstRewrite performs a deep rewrite of an AST node.
// Python: ast_rewrite(x, rewrite) — handles all AST node types.
func AstRewrite(x Node, rewrite AstRewriter) Node {
	if x == nil {
		return nil
	}
	switch n := x.(type) {
	case *Variable:
		// Python: Variable → resort(rewrite_sort(rewrite, x.sort))
		newSort := RewriteSort(rewrite, n.VSort)
		// Also apply prefix transformation to sort names.
		// Python's rewrite_sort calls rewrite_name which applies subst_subscripts,
		// but for SubstPrefix rewriting, the sort name also needs prefix_str
		// (e.g., alias "t" inside a module becomes "index.t" after expansion).
		if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && newSort != "" {
			newSort = sp.PrefixStr(newSort, false)
		}
		return n.Resort(newSort)

	case *Symbol:
		// Go's parser produces *Symbol where Python produces nullary Atom("x", []).
		// Python's ast_rewrite treats Atom("x",[]) by calling rewrite_name then
		// rewrite_atom. We must do the same for Symbol nodes: first apply name
		// substitution (RewriteName), then apply prefix transformation (RewriteAtom)
		// by converting to a temporary nullary Atom, rewriting it, and converting back.
		newRep := rewrite.RewriteName(n.Rep)
		// Apply prefix transformation via RewriteAtom (same as Atom path).
		// Python: for Atom("this",[]), rewrite_atom applies prefix_str("this") → "index".
		tmpAtom := NewAtom(newRep)
		if n.Sort != nil {
			sortStr := fmt.Sprint(n.Sort)
			tmpAtom.ASort = NewSymbol(RewriteSort(rewrite, sortStr), nil)
		}
		if !BaseNameDiffers(n.Rep, newRep) {
			tmpAtom = rewrite.RewriteAtom(tmpAtom, false)
		}
		// Convert back to Symbol with the rewritten name
		if tmpAtom.Rep != n.Rep || (n.Sort != nil && tmpAtom.ASort != nil && fmt.Sprint(tmpAtom.ASort) != fmt.Sprint(n.Sort)) {
			var newSort Node
			if tmpAtom.ASort != nil {
				newSort = tmpAtom.ASort
			} else if n.Sort != nil {
				newSort = n.Sort
			}
			return NewSymbol(tmpAtom.Rep, newSort)
		}
		return n

	case *Atom:
		// Python: isinstance(x, Atom)
		// Check if rep is a NamedBinder (Python checks isinstance(x.rep, NamedBinder))
		// In Go, Atom.Rep is a string, so this doesn't apply.
		newRep := rewrite.RewriteName(n.Rep)
		newArgs := AstRewriteSlice(n.Terms, rewrite)
		newAtom := NewAtom(newRep, newArgs...)
		CopyAttributesAstRef(n, newAtom)
		if n.ASort != nil {
			sortStr := fmt.Sprint(n.ASort)
			newSortStr := RewriteSort(rewrite, sortStr)
			// Also apply prefix transformation to sort names on Atoms.
			// Python: sort annotations like "t" become "index.t" during module expansion.
			if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && newSortStr != "" {
				newSortStr = sp.PrefixStr(newSortStr, false)
			}
			newAtom.ASort = NewSymbol(newSortStr, nil)
		}
		if BaseNameDiffers(n.Rep, newAtom.Rep) {
			return newAtom
		}
		return rewrite.RewriteAtom(newAtom, false)

	case *App:
		// Python: isinstance(x, App) — treated same as Atom
		// App.Rep is a Node (usually *Symbol), extract the string name
		repStr := ""
		if sym, ok := n.Rep.(*Symbol); ok {
			repStr = sym.Rep
		} else {
			repStr = fmt.Sprint(n.Rep)
		}

		// Check if Rep is a NamedBinder
		if nb, ok := n.Rep.(*NamedBinder); ok {
			newRep := AstRewrite(nb, rewrite)
			newArgs := AstRewriteSlice(n.Terms, rewrite)
			newApp := NewApp(newRep, newArgs...)
			CopyAttributesAstRef(n, newApp)
			return newApp
		}

		newRep := rewrite.RewriteName(repStr)
		newArgs := AstRewriteSlice(n.Terms, rewrite)
		newApp := NewApp(NewSymbol(newRep, nil), newArgs...)
		CopyAttributesAstRef(n, newApp)
		if n.ASort != nil {
			sortStr := fmt.Sprint(n.ASort)
			newApp.ASort = NewSymbol(RewriteSort(rewrite, sortStr), nil)
		}
		if BaseNameDiffers(repStr, newRep) {
			return newApp
		}
		// Convert to Atom for rewrite_atom, then convert back
		appAtom := NewAtom(newRep, newApp.Terms...)
		rewritten := rewrite.RewriteAtom(appAtom, false)
		if rewritten.Rep != newRep {
			return NewApp(NewSymbol(rewritten.Rep, nil), rewritten.Terms...)
		}
		return newApp

	case *Literal:
		// Python: isinstance(x, Literal)
		newAtom := AstRewrite(n.Atom, rewrite)
		return NewLiteral(n.Polarity, newAtom)

	case *Forall:
		// Python: isinstance(x, Quantifier) — Forall is a Quantifier
		newBounds := AstRewriteSlice(n.Bounds, rewrite)
		newBody := AstRewrite(n.Body, rewrite)
		return &Forall{Base: n.Base, Bounds: newBounds, Body: newBody}

	case *Exists:
		// Python: isinstance(x, Quantifier) — Exists is a Quantifier
		newBounds := AstRewriteSlice(n.Bounds, rewrite)
		newBody := AstRewrite(n.Body, rewrite)
		return &Exists{Base: n.Base, Bounds: newBounds, Body: newBody}

	case *NamedBinder:
		// Python: isinstance(x, NamedBinder)
		newBounds := AstRewriteSlice(n.Bounds, rewrite)
		newBody := AstRewrite(n.Body, rewrite)
		return &NamedBinder{Base: n.Base, Name: n.Name, Bounds: newBounds, Body: newBody}

	case *LabeledFormula:
		// Python: isinstance(x, LabeledFormula)
		var arg0 Node = n.Label
		if n.Label == nil {
			if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && sp.Pref != nil {
				arg0 = sp.Pref
			}
		} else {
			if label, ok := n.Label.(*Atom); ok {
				newLabelArgs := AstRewriteSlice(label.Terms, rewrite)
				newLabel := label.Clone(newLabelArgs).(*Atom)
				always := true
				if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && sp.Local {
					always = false
				}
				arg0 = rewrite.RewriteAtom(newLabel, always)
			}
		}
		newFmla := AstRewrite(n.Formula, rewrite)
		// Python: res = x.clone([arg0] + [ast_rewrite(y,rewrite) for y in x.args[1:]])
		// Clone preserves all metadata (temporal, explicit, assumed, etc.).
		res := n.Clone([]Node{arg0, newFmla}).(*LabeledFormula)
		return res

	case *NativeDef:
		// Python: isinstance(x, NativeDef) — treated same as LabeledFormula
		var arg0 Node = nil
		args := n.Args()
		if len(args) > 0 {
			arg0 = args[0]
		}
		if arg0 == nil {
			if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && sp.Pref != nil {
				arg0 = sp.Pref
			}
		} else if label, ok := arg0.(*Atom); ok {
			newLabelArgs := AstRewriteSlice(label.Terms, rewrite)
			newLabel := label.Clone(newLabelArgs).(*Atom)
			always := true
			if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && sp.Local {
				always = false
			}
			arg0 = rewrite.RewriteAtom(newLabel, always)
		}
		var rest []Node
		if len(args) > 1 {
			rest = AstRewriteSlice(args[1:], rewrite)
		}
		newArgs := append([]Node{arg0}, rest...)
		return n.Clone(newArgs)

	case *TypeDef:
		// Python: isinstance(x, TypeDef) — rewrite args, check params
		newArgs := AstRewriteSlice(n.Args(), rewrite)
		return n.Clone(newArgs)

	case *SchemaBody:
		// Python: isinstance(x, SchemaBody) — sets local=True during rewrite
		sp, isSP := rewrite.(*AstRewriteSubstPrefix)
		oldLocal := false
		if isSP {
			oldLocal = sp.Local
			sp.Local = true
		}
		newElems := AstRewriteSlice(n.Elems, rewrite)
		if isSP {
			sp.Local = oldLocal
		}
		return &SchemaBody{Base: n.Base, Elems: newElems}

	case *Tactic:
		// Python: isinstance(x, Tactic) — sets local=True during rewrite
		sp, isSP := rewrite.(*AstRewriteSubstPrefix)
		oldLocal := false
		if isSP {
			oldLocal = sp.Local
			sp.Local = true
		}
		newArgs := AstRewriteSlice(n.Args(), rewrite)
		res := n.Clone(newArgs)
		if isSP {
			sp.Local = oldLocal
		}
		return res

	case *DebugItem:
		// Python: isinstance(x, DebugItem)
		args := n.Args()
		if len(args) >= 2 {
			newArg1 := AstRewrite(args[1], rewrite)
			return n.Clone([]Node{args[0], newArg1})
		}
		return n

	default:
		// Python: hasattr(x, 'rewrite') check, then hasattr(x, 'args') fallback
		if args := x.Args(); args != nil {
			newArgs := AstRewriteSlice(args, rewrite)
			return x.Clone(newArgs)
		}
		return x
	}
}

// AstRewriteSlice rewrites a slice of nodes.
// Python: [ast_rewrite(e, rewrite) for e in x]
func AstRewriteSlice(nodes []Node, rewrite AstRewriter) []Node {
	if nodes == nil {
		return nil
	}
	result := make([]Node, len(nodes))
	for i, n := range nodes {
		result[i] = AstRewrite(n, rewrite)
	}
	return result
}

// --- Convenience functions ---

// SubstPrefixAtomsAst is a convenience for ast_rewrite with AstRewriteSubstPrefix.
// Python: subst_prefix_atoms_ast(ast, subst, pref, to_pref, static=None)
// SubstPrefixAtomsAst matches Python's subst_prefix_atoms_ast exactly:
//
//	po = variables_distinct_ast(pref, ast) if pref else pref
//	return ast_rewrite(ast, AstRewriteSubstPrefix(subst, po, to_pref, static=static))
func SubstPrefixAtomsAst(node Node, subst map[string]string, pref *Atom, toPref map[string]bool, static map[string]bool) Node {
	// Python: po = variables_distinct_ast(pref, ast) if pref else pref
	var po *Atom
	if pref != nil {
		renamed := VariablesDistinctAst(pref, node)
		if a, ok := renamed.(*Atom); ok {
			po = a
		} else {
			po = pref
		}
	}
	if subst == nil {
		subst = map[string]string{}
	}
	rw := &AstRewriteSubstPrefix{
		Subst:  subst,
		Pref:   po,
		ToPref: toPref,
		Static: static,
	}
	return AstRewrite(node, rw)
}

// SubstituteAst substitutes terms for variables in an AST.
// Python: substitute_ast(ast, subs)
func SubstituteAst(node Node, subs map[string]Node) Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *Variable:
		if repl, ok := subs[n.Rep]; ok {
			return repl
		}
		return n
	default:
		args := node.Args()
		if args == nil {
			return node
		}
		newArgs := make([]Node, len(args))
		for i, a := range args {
			newArgs[i] = SubstituteAst(a, subs)
		}
		return node.Clone(newArgs)
	}
}

// SubstituteConstantsAst substitutes constants (nullary atoms) in an AST.
// Python: substitute_constants_ast(ast, subs)
func SubstituteConstantsAst(node Node, subs map[string]Node) Node {
	rw := NewAstRewriteSubstConstants(subs)
	return AstRewrite(node, rw)
}

// SubstituteConstantsAst2 substitutes terms for variables in an AST.
// Here, subs is a dict from string names of variables to terms.
// Unlike SubstituteConstantsAst, this handles name splitting for
// compound names (e.g., "a.b" where "a" is in subs).
// Matches Python ivy_ast.py:1817-1839 substitute_constants_ast2.
func SubstituteConstantsAst2(node Node, subs map[string]Node) Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *Atom:
		if len(n.Terms) == 0 {
			if rep, ok := subs[n.Rep]; ok {
				return rep
			}
			names := iu.SplitName(n.Rep)
			if len(names) > 0 {
				if rep, ok := subs[names[0]]; ok {
					rest := iu.ComposeNames(names[1:]...)
					thing := NewAtom(rest)
					thing.Base = n.Base
					res := &MethodCall{Obj: rep, Method: thing}
					res.Base = n.Base
					return res
				}
			}
			return node
		}
	case *App:
		if len(n.Terms) == 0 {
			repStr := fmt.Sprint(n.Rep)
			if rep, ok := subs[repStr]; ok {
				return rep
			}
			names := iu.SplitName(repStr)
			if len(names) > 0 {
				if rep, ok := subs[names[0]]; ok {
					rest := iu.ComposeNames(names[1:]...)
					thing := NewApp(&Symbol{Rep: rest})
					thing.Base = n.Base
					res := &MethodCall{Obj: rep, Method: thing}
					res.Base = n.Base
					return res
				}
			}
			return node
		}
	}
	newArgs := make([]Node, len(node.Args()))
	for i, a := range node.Args() {
		newArgs[i] = SubstituteConstantsAst2(a, subs)
	}
	res := node.Clone(newArgs)
	CopyAttributesAstRef(node, res)
	return res
}

// --- Variable distinct renaming ---

// UsedVariablesAst collects all Variable nodes in an AST.
// Python: used_variables_ast(ast) — returns a set of Variable objects.
func UsedVariablesAst(node Node) []*Variable {
	var result []*Variable
	seen := make(map[string]bool)
	usedVariablesRec(node, &result, seen)
	return result
}

func usedVariablesRec(node Node, result *[]*Variable, seen map[string]bool) {
	if node == nil {
		return
	}
	if v, ok := node.(*Variable); ok {
		if !seen[v.Rep] {
			seen[v.Rep] = true
			*result = append(*result, v)
		}
		return
	}
	args := node.Args()
	for _, a := range args {
		usedVariablesRec(a, result, seen)
	}
}

// DistinctVariableRenaming creates a renaming map so variables in vars1
// don't clash with variables in vars2.
// Python: distinct_variable_renaming(vars1, vars2)
func DistinctVariableRenaming(vars1, vars2 []*Variable) map[string]Node {
	used := make(map[string]bool)
	for _, v := range vars2 {
		used[v.Rep] = true
	}
	result := make(map[string]Node)
	for _, v := range vars1 {
		if used[v.Rep] {
			newName := v.Rep
			for used[newName] {
				newName = newName + "'"
			}
			used[newName] = true
			result[v.Rep] = NewVariable(newName, v.VSort)
		}
	}
	return result
}

// VariablesDistinctAst renames variables in ast1 so they don't occur in ast2.
// Python: variables_distinct_ast(ast1, ast2)
func VariablesDistinctAst(ast1, ast2 Node) Node {
	if ast1 == nil {
		return nil
	}
	vars1 := UsedVariablesAst(ast1)
	vars2 := UsedVariablesAst(ast2)
	renaming := DistinctVariableRenaming(vars1, vars2)
	if len(renaming) == 0 {
		return ast1
	}
	return SubstituteAst(ast1, renaming)
}

// IsTrue and IsFalse are defined in ast.go
