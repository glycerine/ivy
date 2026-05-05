package ast

import (
	"fmt"
	"strings"
	"sync/atomic"

	iu "github.com/glycerine/ivy/goivy/ivyutils"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// debugNextDefinitionDeclSn is a debug-only atomic counter for tagging
// DefinitionDecl instances with a unique serial number at creation time.
// This is intentionally a package-level var (debug-only, exempt per CLAUDE.md rule 10).
var debugNextDefinitionDeclSn atomic.Int64

// --- Declaration types ---

// LabeledFormula associates a label with a formula (used in axioms, properties, etc.).
//
// Source location lives ONLY in Base.Loc (set via SetLineno / read via GetLineno).
// Python's lf.lineno is a LocationTuple — there is no separate int field — so the
// Go port keeps Base.Loc as the single source of truth. Use LinenoLine() when you
// need just the integer line number.
type LabeledFormula struct {
	Base
	Label        Node // label (may be nil)
	Formula      Node
	ID           int64
	Temporal     *bool // tristate: nil = not set, true = temporal, false = non-temporal
	Explicit     bool
	IsDefinition bool
	Assumed      bool
	Unprovable   bool
	// Annot holds the annotation for trace reconstruction, or nil.
	// In Python this is a dynamically-attached attribute (lf.annot).
	// It carries (action, annotation) pair context for proof checking.
	Annot interface{}
	// TraceHook is a diagnostic hook closure attached by tactics (e.g. l2s).
	// The concrete type is check.TraceHookFn; the field is interface{} only
	// because ast cannot import check (cycle). Mirrors Python's
	// dynamically-attached lf.trace_hook attribute (ivy_l2s.py:88, 1311, 1313).
	TraceHook interface{}
}

// Lineno returns the line number from the node's Location.
// Equivalent to lf.GetLineno().Line. Matches Python lf.lineno's
// integer-coercion semantics for callers that need just the line.
func (lf *LabeledFormula) Lineno() int { return lf.Base.Loc.Line }

// BoolPtr returns a pointer to a bool value.
func BoolPtr(b bool) *bool { return &b }

// IsTemporal returns true if the Temporal field is set and true.
func (lf *LabeledFormula) IsTemporal() bool {
	return lf.Temporal != nil && *lf.Temporal
}

func (cfg *AstConfig) NewLabeledFormula(label, formula Node) *LabeledFormula {
	id := cfg.NextLFID()
	lf := &LabeledFormula{
		Label:   label,
		Formula: formula,
		ID:      id,
	}
	lf.Cfg = cfg
	xtracer.Trace("ast.LF.__init__ id=%d counter=%d", id, cfg.LfCounter)
	return lf
}

func (lf *LabeledFormula) Args() []Node { return []Node{lf.Label, lf.Formula} }
func (lf *LabeledFormula) Clone(args []Node) Node {
	cfg := lf.Cfg
	if cfg == nil {
		panic("ast: Clone called on node with nil AstConfig — node was not created via cfg.NewFoo()")
	}
	// Python: clone() calls AST.clone() → __init__ (suppressed via _in_clone),
	// then if not always_clone_with_fresh_id: decrements counter and
	// restores original ID, emitting LF.clone PRESERVE.
	// If always_clone_with_fresh_id: keeps fresh ID, emitting LF.clone FRESH.
	c := lf.cloneInternal(args)
	if cfg.AlwaysCloneWithFreshID {
		xtracer.Trace("ast.LF.clone FRESH origid=%d newid=%d counter=%d", lf.ID, c.ID, cfg.LfCounter)
	} else {
		// Python: lf_counter -= 1; res.id = self.id
		cfg.LfCounter--
		c.ID = lf.ID
		xtracer.Trace("ast.LF.clone PRESERVE origid=%d counter=%d", c.ID, cfg.LfCounter)
	}
	return c
}

// cloneInternal creates a copy with a fresh ID, no tracing.
// Used by both Clone() and CloneWithFreshID() to avoid double-tracing.
func (lf *LabeledFormula) cloneInternal(args []Node) *LabeledFormula {
	cfg := lf.Cfg
	id := cfg.NextLFID()
	return &LabeledFormula{
		Base:         Base{Cfg: cfg, Loc: lf.Base.Loc, HasLoc: lf.Base.HasLoc},
		Label:        args[0],
		Formula:      args[1],
		ID:           id,
		Temporal:     lf.Temporal,
		Explicit:     lf.Explicit,
		IsDefinition: lf.IsDefinition,
		Assumed:      lf.Assumed,
		Unprovable:   lf.Unprovable,
		Annot:        lf.Annot,
		TraceHook:    lf.TraceHook,
	}
}

func (lf *LabeledFormula) CloneWithFreshID(args []Node) *LabeledFormula {
	cfg := lf.Cfg
	if cfg == nil {
		panic("ast: Clone called on node with nil AstConfig — node was not created via cfg.NewFoo()")
	}
	// Python: clone_with_fresh_id() calls AST.clone() → __init__ (emits LF.__init__),
	// keeps fresh ID. Does NOT emit LF.clone PRESERVE.
	c := lf.cloneInternal(args)
	xtracer.Trace("ast.LF.__init__ id=%d counter=%d", c.ID, cfg.LfCounter)
	return c
}
func (lf *LabeledFormula) String() string {
	if lf.Label != nil {
		return "[" + fmt.Sprint(lf.Label) + "] " + fmt.Sprint(lf.Formula)
	}
	return fmt.Sprint(lf.Formula)
}
func (lf *LabeledFormula) LabelName() string {
	if lf.Label == nil {
		return ""
	}
	if a, ok := lf.Label.(*Atom); ok {
		return a.Relname()
	}
	return fmt.Sprint(lf.Label)
}

// LabelForTrace mirrors Python's `%s goal.label` — includes Atom args
// in parentheses (full __repr__), not just the relation name. Use this
// for XTRACE lines that correspond to Python sites emitting `%s goal.label`.
// For name-only needs (map keys, equality), use LabelName().
func (lf *LabeledFormula) LabelForTrace() string {
	if lf.Label == nil {
		return ""
	}
	return fmt.Sprint(lf.Label)
}

func (lf *LabeledFormula) Rename(s string) *LabeledFormula {
	newLabel := lf.Label
	if a, ok := lf.Label.(*Atom); ok {
		newLabel = a.Rename(s)
	}
	return lf.Clone([]Node{newLabel, lf.Formula}).(*LabeledFormula)
}

// Decl is the interface for all declaration nodes that embed DeclBase.
// Matches Python's Decl base class with .args, .attributes, .common fields.
type Decl interface {
	Node
	GetDeclBase() *DeclBase
}

// DeclBase is the base for all declaration nodes.
// In Python, Decl has args and attributes. In Go, each specific Decl type
// has its own fields. We provide a common DeclBase.
//
// NOTE: Clone methods on Decl types intentionally do NOT copy Attributes
// or Common. This matches Python's Decl.__init__, which always resets
// self.attributes=() and self.common=None on clone. Callers that need
// these fields (e.g., inst_mod) explicitly restore them after cloning.
type DeclBase struct {
	Base
	DeclArgs   []Node
	Attributes []Node
	Common     Node // optional common block; Python: decl.common (None or 'this')
}

func (r *DeclBase) Canon() iu.Canonical {
	s := fmt.Sprintf("(declBase base:%v", r.Base.Canon())
	if len(r.DeclArgs) > 0 {
		s += " declArgs:["
		for i, d := range r.DeclArgs {
			_ = i
			s += fmt.Sprintf("%v ", d.Canon())
		}
		s += "]"
	}
	if len(r.Attributes) > 0 {
		s += " attributes:["
		for i, d := range r.Attributes {
			_ = i
			s += fmt.Sprintf("%v ", d.Canon())
		}
		s += "]"
	}
	s += fmt.Sprintf(" common:%v)", r.Common.Canon())
	return iu.Canonical(s)
}

// canonFields returns flattened fields from DeclBase for inclusion in
// parent Canon() output, promoting Base fields inline.
func (d *DeclBase) canonFields() string {
	return fmt.Sprintf("%v declArgs:%v attributes:%v common:%v",
		d.Base.canonFields(), SliceCanon(d.DeclArgs), attrSliceCanon(d.Attributes), nodeCanon(d.Common))
}

// attrSliceCanon returns canonical form for attribute nodes as bare strings.
// Python attributes are plain strings; their canon is just the string itself (no quotes).
func attrSliceCanon(attrs []Node) string {
	if len(attrs) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, a := range attrs {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if atom, ok := a.(*Atom); ok {
			sb.WriteString(atom.Rep)
		} else {
			sb.WriteString(fmt.Sprint(a))
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (d *DeclBase) GetDeclBase() *DeclBase { return d }
func (d *DeclBase) Args() []Node           { return d.DeclArgs }
func (d *DeclBase) GetAttributes() []Node  { return d.Attributes }
func (d *DeclBase) SetAttributes(a []Node) { d.Attributes = a }
func (d *DeclBase) GetCommon() Node        { return d.Common }
func (d *DeclBase) SetCommon(c Node)       { d.Common = c }

// GetDeclBase extracts the DeclBase from any node that embeds it.
// Returns nil if the node is not a declaration.
func GetDeclBase(n Node) *DeclBase {
	if d, ok := n.(Decl); ok {
		return d.GetDeclBase()
	}
	return nil
}

// Defines returns the names defined by this declaration, by iterating DeclArgs
// and collecting defines from each arg. Specific decl types may override.
// Corresponds to Python Decl.defines() which returns [(name, lineno), ...].
func (d *DeclBase) Defines() []string {
	var names []string
	for _, arg := range d.DeclArgs {
		// Try the arg's own Defines() method (e.g. TypeDef, EnumeratedSort)
		//type DefinerSlice interface {
		//	Defines() []string
		//}
		if df, ok := arg.(DefinerSlice); ok {
			names = append(names, df.Defines()...)
			continue
		}
		// Fall back to extracting the rep/relname from specific arg types.
		// This matches Python's polymorphic defines() dispatch in ivy_ast.py.
		switch a := arg.(type) {
		case *AstDefinition:
			// Python: Definition.defines() returns self.args[0].rep
			if n := a.Defines(); n != "" {
				names = append(names, n)
			}
		case *Atom:
			// Python: ConstantDecl.defines() / RelationDecl.defines()
			// filter out polymorphic symbols like <, <=, +, *, etc.
			if a.Rep != "" {
				if _, poly := iu.PolymorphicSymbols[a.Rep]; !poly {
					names = append(names, a.Rep)
				}
			}
		case *App:
			// Python: App.rep is used for defines() — matches ConstantDecl args
			if a.Rep != nil {
				if s, ok := a.Rep.(*Symbol); ok && s.Rep != "" {
					if _, poly := iu.PolymorphicSymbols[s.Rep]; !poly {
						names = append(names, s.Rep)
					}
				}
			}
		case *ActionDef:
			if n := a.Defines(); n != "" {
				names = append(names, n)
			}
		case *LabeledFormula:
			if a.Label != nil {
				if rep := NodeRep(a.Label); rep != "" {
					names = append(names, rep)
				}
			}
		case DefinerStr:
			// Catch-all for types with Defines() string (e.g. Schema)
			if n := a.Defines(); n != "" {
				names = append(names, n)
			}
		}
	}
	return names
}

// ModuleDecl declares a module.
type ModuleDecl struct {
	DeclBase
	FormalParams []Node // formal parameters for instantiation
	BodyDecls    []Node // parsed body declarations for instantiation
}

func (r *ModuleDecl) Canon() iu.Canonical {
	// FormalParams and BodyDecls are Go-only convenience fields not present
	// in Python's ModuleDecl. Omit from canon to match Python.
	return iu.Canonical(fmt.Sprintf("(moduleDecl%v)",
		r.DeclBase.canonFields()))
}

func (cfg *AstConfig) NewModuleDecl(args ...Node) *ModuleDecl {
	d := &ModuleDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ModuleDecl) Clone(args []Node) Node {
	return &ModuleDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ModuleDecl) String() string {
	parts := make([]string, len(d.DeclArgs))
	for i, a := range d.DeclArgs {
		parts[i] = fmt.Sprint(a)
	}
	return "module " + strings.Join(parts, ", ")
}

// MacroDecl declares a macro.
type MacroDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewMacroDecl(args ...Node) *MacroDecl {
	d := &MacroDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *MacroDecl) Clone(args []Node) Node {
	return &MacroDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *MacroDecl) String() string { return "macro" }

// ObjectDecl declares an object.
type ObjectDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewObjectDecl(args ...Node) *ObjectDecl {
	d := &ObjectDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ObjectDecl) Clone(args []Node) Node {
	return &ObjectDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ObjectDecl) String() string { return "object" }

// ActionDecl declares one or more actions.
type ActionDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewActionDecl(args ...Node) *ActionDecl {
	d := &ActionDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ActionDecl) Clone(args []Node) Node {
	return &ActionDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ActionDecl) String() string { return "action" }

// ActionDef defines an action with formal params and returns.
type ActionDef struct {
	Base
	Name          Node
	Body          Node
	FormalParams  []Node
	FormalReturns []Node
	Attributes    []string // copied from ActionDecl during declare; see ivy_parser.py:344
}

// NewActionDef creates an ActionDef, renaming formals with "fml:" prefix
// and substituting those names into the body to prevent name capture.
// Matches Python ActionDef.__init__ (ivy_ast.py:1375-1383).
func (cfg *AstConfig) NewActionDef(name, body Node, params, returns []Node) *ActionDef {
	fmlParams := prefixNodes(params, "fml:")
	fmlReturns := prefixNodes(returns, "fml:")
	if len(params) > 0 || len(returns) > 0 {
		subst := make(map[string]string)
		for i, p := range params {
			subst[NodeRep(p)] = NodeRep(fmlParams[i])
		}
		for i, r := range returns {
			subst[NodeRep(r)] = NodeRep(fmlReturns[i])
		}
		body = SubstPrefixAtomsAst(body, subst, nil, nil, nil)
	}
	d := &ActionDef{Name: name, Body: body, FormalParams: fmlParams, FormalReturns: fmlReturns}
	d.Cfg = cfg
	return d
}

// prefixNodes applies Prefix(s) to each node, returning prefixed copies.
func prefixNodes(nodes []Node, s string) []Node {
	if len(nodes) == 0 {
		return nil
	}
	result := make([]Node, len(nodes))
	for i, n := range nodes {
		switch a := n.(type) {
		case *Atom:
			// Python: Atom.prefix() returns another Atom (type-preserving clone).
			// Atom and App are sibling classes in Python, not parent-child.
			result[i] = a.Prefix(s)
		case *App:
			result[i] = a.Prefix(s)
		case *Variable:
			// Python: Variable.prefix creates an App with prefixed rep, preserving sort
			repSym := &Symbol{Rep: s + a.Rep}
			repSym.Cfg = a.Cfg
			app := &App{Rep: repSym}
			app.Cfg = a.Cfg
			if a.VSort != "" {
				sortSym := &Symbol{Rep: a.VSort}
				sortSym.Cfg = a.Cfg
				app.ASort = sortSym
			}
			result[i] = app
		default:
			result[i] = n
		}
	}
	return result
}

// NodeRep extracts the name string from an AST node.
func NodeRep(n Node) string {
	switch a := n.(type) {
	case *Atom:
		return a.Rep
	case *App:
		return a.Relname()
	case *Variable:
		return a.Rep
	case *Symbol:
		return a.Rep
	default:
		return fmt.Sprint(n)
	}
}

// Args returns all child nodes for rewriting: Name, Body, FormalParams..., FormalReturns...
// The layout is: [Name, Body, params..., returns...]
// Clone reconstructs from this layout.
func (a *ActionDef) Args() []Node {
	result := []Node{a.Name, a.Body}
	result = append(result, a.FormalParams...)
	result = append(result, a.FormalReturns...)
	return result
}

func (a *ActionDef) Clone(args []Node) Node {
	name := args[0]
	body := args[1]
	rest := args[2:]
	nParams := len(a.FormalParams)
	var params, returns []Node
	if nParams <= len(rest) {
		params = rest[:nParams]
		returns = rest[nParams:]
	}
	return &ActionDef{Base: a.Base, Name: name, Body: body,
		FormalParams: params, FormalReturns: returns}
}
func (a *ActionDef) String() string {
	parts := make([]string, len(a.FormalParams))
	for i, p := range a.FormalParams {
		parts[i] = fmt.Sprint(p)
	}
	return fmt.Sprint(a.Name) + "(" + strings.Join(parts, ",") + ") = " + fmt.Sprint(a.Body)
}
func (a *ActionDef) Defines() string {
	return NodeRep(a.Name)
}

// Formals returns unprefixed (original) params and returns by stripping "fml:".
// Matches Python ActionDef.formals() (ivy_ast.py:1406-1408).
func (a *ActionDef) Formals() (params []Node, returns []Node) {
	params = dropPrefixNodes(a.FormalParams, "fml:")
	returns = dropPrefixNodes(a.FormalReturns, "fml:")
	return
}

// dropPrefixNodes strips a prefix from each node.
func dropPrefixNodes(nodes []Node, s string) []Node {
	if len(nodes) == 0 {
		return nil
	}
	result := make([]Node, len(nodes))
	for i, n := range nodes {
		switch a := n.(type) {
		case *Atom:
			result[i] = a.DropPrefix(s)
		case *App:
			result[i] = a.DropPrefix(s)
		default:
			result[i] = n
		}
	}
	return result
}

// Rewrite applies an AST rewriter to the ActionDef's body and formals.
// Matches Python ActionDef.rewrite() (ivy_ast.py:1408-1416).
// Returns Node to satisfy AstRewritable interface — Python dispatches via
// hasattr(x, 'rewrite') at ivy_ast.py:1729.
func (a *ActionDef) Rewrite(rw AstRewriter) Node {
	xtracer.Trace("ActionDef.rewrite ENTER nParams=%d nReturns=%d", len(a.FormalParams), len(a.FormalReturns))
	// Python: res = self.clone(ast_rewrite(self.args, rewrite))
	// self.args = [atom, action] — only Name and Body, NOT formals.
	rewrittenNameBody := AstRewriteSlice([]Node{a.Name, a.Body}, rw)
	// Clone expects [Name, Body, params..., returns...]. Pass old formals
	// through; they'll be overwritten by rewriteParams below.
	allArgs := make([]Node, 0, 2+len(a.FormalParams)+len(a.FormalReturns))
	allArgs = append(allArgs, rewrittenNameBody...)
	allArgs = append(allArgs, a.FormalParams...)
	allArgs = append(allArgs, a.FormalReturns...)
	res := a.Clone(allArgs).(*ActionDef)
	// Python: res.formal_params = [rewrite_param(p, rewrite) for p in self.formal_params]
	res.FormalParams = rewriteParams(a.FormalParams, rw)
	res.FormalReturns = rewriteParams(a.FormalReturns, rw)
	xtracer.Trace("ActionDef.rewrite EXIT nParams=%d nReturns=%d", len(res.FormalParams), len(res.FormalReturns))
	return res
}

// rewriteParams applies rewriteParam to each param node.
// Matches Python: [rewrite_param(p, rewrite) for p in self.formal_params]
func rewriteParams(params []Node, rw AstRewriter) []Node {
	if len(params) == 0 {
		return nil
	}
	result := make([]Node, len(params))
	for i, p := range params {
		result[i] = rewriteParam(p, rw)
	}
	return result
}

// rewriteParam matches Python rewrite_param (ivy_ast.py:1421-1424):
//
//	res = type(p)(p.rep)
//	res.sort = rewrite_sort(rewrite, p.sort)
//
// Creates a fresh node with same rep (no args/terms) and rewrites the sort.
// Does NOT do a full recursive AstRewrite — Python doesn't either.
func rewriteParam(p Node, rw AstRewriter) Node {
	switch n := p.(type) {
	case *App:
		repStr := NodeRep(n.Rep)
		repSym := &Symbol{Rep: repStr}
		repSym.Cfg = n.Cfg
		res := &App{Rep: repSym}
		res.Cfg = n.Cfg
		if n.ASort != nil {
			sortStr := fmt.Sprint(n.ASort)
			newSort := RewriteSort(rw, sortStr, n.Cfg)
			ss := &Symbol{Rep: newSort}
			ss.Cfg = n.Cfg
			res.ASort = ss
		}
		return res
	case *Atom:
		res := &Atom{Rep: n.Rep}
		res.Cfg = n.Cfg
		if n.ASort != nil {
			sortStr := fmt.Sprint(n.ASort)
			newSort := RewriteSort(rw, sortStr, n.Cfg)
			ss := &Symbol{Rep: newSort}
			ss.Cfg = n.Cfg
			res.ASort = ss
		}
		return res
	default:
		// Fallback — formal params should always be App or Atom
		return AstRewrite(p, rw)
	}
}

// RelationDecl declares a relation.
type RelationDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewRelationDecl(args ...Node) *RelationDecl {
	d := &RelationDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *RelationDecl) Clone(args []Node) Node {
	return &RelationDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *RelationDecl) String() string { return "relation" }

// ConstantDecl declares a constant (individual).
type ConstantDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewConstantDecl(args ...Node) *ConstantDecl {
	d := &ConstantDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ConstantDecl) Clone(args []Node) Node {
	return &ConstantDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ConstantDecl) String() string { return "individual" }

// ParameterDecl declares a parameter.
type ParameterDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewParameterDecl(args ...Node) *ParameterDecl {
	d := &ParameterDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ParameterDecl) Clone(args []Node) Node {
	return &ParameterDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ParameterDecl) String() string { return "parameter" }

// FreshConstantDecl declares a fresh constant.
type FreshConstantDecl struct {
	ConstantDecl
}

func (cfg *AstConfig) NewFreshConstantDecl(cd ConstantDecl) *FreshConstantDecl {
	f := &FreshConstantDecl{ConstantDecl: cd}
	f.Cfg = cfg
	return f
}

// DestructorDecl declares a destructor.
type DestructorDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewDestructorDecl(args ...Node) *DestructorDecl {
	d := &DestructorDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *DestructorDecl) Clone(args []Node) Node {
	return &DestructorDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *DestructorDecl) String() string { return "destructor" }

// ConstructorDecl declares a constructor.
type ConstructorDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewConstructorDecl(args ...Node) *ConstructorDecl {
	d := &ConstructorDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ConstructorDecl) Clone(args []Node) Node {
	return &ConstructorDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ConstructorDecl) String() string { return "constructor" }

// TypeDecl declares a type.
type TypeDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewTypeDecl(args ...Node) *TypeDecl {
	d := &TypeDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *TypeDecl) Clone(args []Node) Node {
	return &TypeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *TypeDecl) String() string { return "type" }

// TypeDef defines a named type.
type TypeDef struct {
	Base
	Name   Node
	Value  Node
	Finite bool
}

func (cfg *AstConfig) NewTypeDef(name, value Node) *TypeDef {
	d := &TypeDef{Name: name, Value: value}
	d.Cfg = cfg
	return d
}

func (t *TypeDef) Args() []Node { return []Node{t.Name, t.Value} }
func (t *TypeDef) Clone(args []Node) Node {
	return &TypeDef{Base: t.Base, Name: args[0], Value: args[1], Finite: t.Finite}
}
func (t *TypeDef) String() string {
	prefix := ""
	if t.Finite {
		prefix = "finite "
	}
	return prefix + fmt.Sprint(t.Name) + " = " + fmt.Sprint(t.Value)
}

type DefinerSlice interface {
	Defines() []string
}

type DefinerStr interface {
	Defines() string
}

func (t *TypeDef) Defines() []string {
	var syms []string
	if rep := NodeRep(t.Name); rep != "" {
		syms = append(syms, rep)
	}
	// Add names defined by the value (e.g., enum elements)
	if d, ok := t.Value.(DefinerSlice); ok {
		syms = append(syms, d.Defines()...)
	}
	return syms
}

// GhostTypeDef is a ghost (specification-only) type definition.
type GhostTypeDef struct {
	TypeDef
}

func (cfg *AstConfig) NewGhostTypeDef(td TypeDef) *GhostTypeDef {
	g := &GhostTypeDef{TypeDef: td}
	g.Cfg = cfg
	return g
}

func (g *GhostTypeDef) Clone(args []Node) Node {
	return &GhostTypeDef{TypeDef: *g.TypeDef.Clone(args).(*TypeDef)}
}

// VariantDecl declares a variant type.
type VariantDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewVariantDecl(args ...Node) *VariantDecl {
	d := &VariantDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *VariantDecl) Clone(args []Node) Node {
	return &VariantDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *VariantDecl) String() string { return "variant" }

// VariantDef defines a variant of a type.
type VariantDef struct {
	Base
	Name  Node
	VSort Node
}

func (cfg *AstConfig) NewVariantDef(name, sort Node) *VariantDef {
	d := &VariantDef{Name: name, VSort: sort}
	d.Cfg = cfg
	return d
}

func (v *VariantDef) Args() []Node { return []Node{v.Name, v.VSort} }
func (v *VariantDef) Clone(args []Node) Node {
	return &VariantDef{Base: v.Base, Name: args[0], VSort: args[1]}
}
func (v *VariantDef) String() string { return fmt.Sprint(v.Name) + " of " + fmt.Sprint(v.VSort) }

// AxiomDecl declares an axiom.
type AxiomDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewAxiomDecl(args ...Node) *AxiomDecl {
	d := &AxiomDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *AxiomDecl) Clone(args []Node) Node {
	return &AxiomDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *AxiomDecl) String() string { return "axiom" }

// PropertyDecl declares a property (provable axiom).
type PropertyDecl struct {
	AxiomDecl
}

func (cfg *AstConfig) NewPropertyDecl(args ...Node) *PropertyDecl {
	d := &PropertyDecl{AxiomDecl: AxiomDecl{DeclBase: DeclBase{DeclArgs: args}}}
	d.Cfg = cfg
	return d
}

func (d *PropertyDecl) Clone(args []Node) Node {
	return &PropertyDecl{AxiomDecl: *d.AxiomDecl.Clone(args).(*AxiomDecl)}
}
func (d *PropertyDecl) String() string { return "property" }

// ConjectureDecl declares a conjecture.
type ConjectureDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewConjectureDecl(args ...Node) *ConjectureDecl {
	d := &ConjectureDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ConjectureDecl) Clone(args []Node) Node {
	return &ConjectureDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ConjectureDecl) String() string { return "conjecture" }

// ProofDecl declares a proof.
type AstProofDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewProofDecl(args ...Node) *AstProofDecl {
	d := &AstProofDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *AstProofDecl) Clone(args []Node) Node {
	return &AstProofDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *AstProofDecl) String() string { return "proof" }

// NamedDecl declares a named entity.
type NamedDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewNamedDecl(args ...Node) *NamedDecl {
	d := &NamedDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *NamedDecl) Clone(args []Node) Node {
	return &NamedDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *NamedDecl) String() string { return "named" }

// SchemaDecl declares a proof schema.
type SchemaDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewSchemaDecl(args ...Node) *SchemaDecl {
	d := &SchemaDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *SchemaDecl) Clone(args []Node) Node {
	return &SchemaDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *SchemaDecl) String() string { return "schema" }

// SchemaBody holds premises and a conclusion.
type SchemaBody struct {
	Base
	Elems []Node // premises + conclusion (last elem)
}

func (cfg *AstConfig) NewSchemaBody(elems ...Node) *SchemaBody {
	d := &SchemaBody{Elems: elems}
	d.Cfg = cfg
	return d
}

func (s *SchemaBody) Args() []Node           { return s.Elems }
func (s *SchemaBody) Clone(args []Node) Node { return &SchemaBody{Base: s.Base, Elems: args} }
func (s *SchemaBody) String() string         { return "{...}" }
func (s *SchemaBody) Prems() []Node {
	if len(s.Elems) <= 1 {
		return nil
	}
	// Return a copy so callers can append without overwriting the
	// conclusion element via the shared backing array. Matches
	// Python's list(g.formula.prems()) which returns a fresh list.
	src := s.Elems[:len(s.Elems)-1]
	cp := make([]Node, len(src))
	copy(cp, src)
	return cp
}
func (s *SchemaBody) Conc() Node {
	if len(s.Elems) == 0 {
		return nil
	}
	return s.Elems[len(s.Elems)-1]
}

// Schema wraps a Definition for schema declarations.
// Matches Python's Schema(AST) from ivy_actions.py.
type AstSchema struct {
	Base
	Defn      Node   // the Definition
	Fresh     []Node // fresh variables
	Instances []Node // instantiation records
}

func (cfg *AstConfig) NewSchema(defn Node) *AstSchema {
	d := &AstSchema{Defn: defn}
	d.Cfg = cfg
	return d
}

func (s *AstSchema) Args() []Node { return []Node{s.Defn} }
func (s *AstSchema) Clone(args []Node) Node {
	ns := &AstSchema{Base: s.Base, Fresh: s.Fresh, Instances: s.Instances}
	if len(args) > 0 {
		ns.Defn = args[0]
	}
	return ns
}
func (s *AstSchema) String() string {
	res := fmt.Sprintf("%v", s.Defn)
	if len(s.Fresh) > 0 {
		res += " fresh "
		for i, f := range s.Fresh {
			if i > 0 {
				res += ","
			}
			res += fmt.Sprint(f)
		}
	}
	return res
}
func (s *AstSchema) Defines() string {
	if d, ok := s.Defn.(*AstDefinition); ok {
		return d.Defines()
	}
	return ""
}

// SchemaCompiler is the interface needed by Schema.GetInstance to compile AST to logic.
// Python: compile_with_sort_inference is monkey-patched onto AST; we use an explicit interface.
type SchemaCompiler interface {
	CompileWithSortInference(node Node) (Node, error)
}

// SchemaClauseConverter converts logic formulas to clauses.
// Python: formula_to_clauses(fmla)
type SchemaClauseConverter interface {
	FormulaToClauses(fmla Node) (Node, error)
}

// GetInstance creates an instance of this schema with the given parameters.
// Python: Schema.get_instance(self, params, to_clauses=True)
func (s *AstSchema) GetInstance(params []Node, compiler SchemaCompiler, clauseConverter SchemaClauseConverter, toClauses bool) (Node, error) {
	defn, ok := s.Defn.(*AstDefinition)
	if !ok {
		return nil, fmt.Errorf("schema defn is not a Definition")
	}
	// defn.Lhs is the atom with formal parameters
	var lhsArgs []Node
	if atom, ok := defn.Lhs.(*Atom); ok {
		lhsArgs = atom.Terms
	}
	if len(params) != len(lhsArgs) {
		return nil, fmt.Errorf("schema parameter count mismatch: expected %d, got %d", len(lhsArgs), len(params))
	}
	// Build substitution: formal param name → actual param name
	subst := make(map[string]string)
	for i, formal := range lhsArgs {
		var formalName string
		switch f := formal.(type) {
		case *Atom:
			formalName = f.Rep
		case *Symbol:
			formalName = f.Rep
		case *Variable:
			formalName = f.Rep
		default:
			formalName = fmt.Sprint(formal)
		}
		var actualName string
		switch a := params[i].(type) {
		case *Atom:
			actualName = a.Rep
		case *Symbol:
			actualName = a.Rep
		case *Variable:
			actualName = a.Rep
		default:
			actualName = fmt.Sprint(params[i])
		}
		subst[formalName] = actualName
	}
	// Rewrite the body with the substitution
	rewriter := NewAstRewriteSubstPrefix(subst, nil)
	rewrittenBody := AstRewrite(defn.Rhs, rewriter)

	// Compile with sort inference
	fmla, err := compiler.CompileWithSortInference(rewrittenBody)
	if err != nil {
		return nil, err
	}

	if toClauses && clauseConverter != nil {
		return clauseConverter.FormulaToClauses(fmla)
	}
	return fmla, nil
}

// Instantiate adds an instance to this schema's instance list.
// Python: Schema.instantiate(self, params)
func (s *AstSchema) Instantiate(params []Node, compiler SchemaCompiler) {
	inst, err := s.GetInstance(params, compiler, nil, false)
	if err == nil {
		s.Instances = append(s.Instances, inst)
	}
}

// TheoremDecl declares a theorem.
type TheoremDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewTheoremDecl(args ...Node) *TheoremDecl {
	d := &TheoremDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *TheoremDecl) Clone(args []Node) Node {
	return &TheoremDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *TheoremDecl) String() string { return "theorem" }

// DerivedDecl declares a derived relation/function.
type DerivedDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewDerivedDecl(args ...Node) *DerivedDecl {
	d := &DerivedDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *DerivedDecl) Clone(args []Node) Node {
	return &DerivedDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *DerivedDecl) String() string { return "derived" }

// Defines returns the names defined by this derived declaration.
// Python: DerivedDecl.defines() returns [(c.formula.defines(), lineno(c.formula)) for c in self.args]
// Each arg is a LabeledFormula; we extract the formula's defines (the LHS name).
func (d *DerivedDecl) Defines() []string {
	var names []string
	for _, arg := range d.DeclArgs {
		if lf, ok := arg.(*LabeledFormula); ok && lf.Formula != nil {
			if defn, ok := lf.Formula.(*AstDefinition); ok {
				if n := defn.Defines(); n != "" {
					names = append(names, n)
				}
			}
		}
	}
	return names
}

// DefinitionDecl declares a definition.
type DefinitionDecl struct {
	DeclBase
	Sn int64 // debug serial number, assigned at creation
}

func (cfg *AstConfig) NewDefinitionDecl(args ...Node) *DefinitionDecl {
	sn := debugNextDefinitionDeclSn.Add(1)
	d := &DefinitionDecl{DeclBase: DeclBase{DeclArgs: args}, Sn: sn}
	d.Cfg = cfg
	return d
}

func (d *DefinitionDecl) Clone(args []Node) Node {
	sn := debugNextDefinitionDeclSn.Add(1)
	c := &DefinitionDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}, Sn: sn}
	return c
}
func (d *DefinitionDecl) String() string { return "definition" }

// ProgressDecl declares a progress property.
type ProgressDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewProgressDecl(args ...Node) *ProgressDecl {
	d := &ProgressDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ProgressDecl) Clone(args []Node) Node {
	return &ProgressDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ProgressDecl) String() string { return "progress" }

// RelyDecl declares a rely condition.
type RelyDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewRelyDecl(args ...Node) *RelyDecl {
	d := &RelyDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *RelyDecl) Clone(args []Node) Node {
	return &RelyDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *RelyDecl) String() string { return "rely" }

// MixOrdDecl declares a mixin ordering.
type MixOrdDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewMixOrdDecl(args ...Node) *MixOrdDecl {
	d := &MixOrdDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *MixOrdDecl) Clone(args []Node) Node {
	return &MixOrdDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *MixOrdDecl) String() string { return "mixord" }

// ConceptDecl declares a concept.
type ConceptDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewConceptDecl(args ...Node) *ConceptDecl {
	d := &ConceptDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ConceptDecl) Clone(args []Node) Node {
	return &ConceptDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ConceptDecl) String() string { return "concept" }

// InitDecl declares initialization code.
type InitDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewInitDecl(args ...Node) *InitDecl {
	d := &InitDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *InitDecl) Clone(args []Node) Node {
	return &InitDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *InitDecl) String() string { return "init" }

// StateDecl declares state variables.
type StateDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewStateDecl(args ...Node) *StateDecl {
	d := &StateDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *StateDecl) Clone(args []Node) Node {
	return &StateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *StateDecl) String() string { return "state" }

// UpdateDecl declares an update.
type UpdateDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewUpdateDecl(args ...Node) *UpdateDecl {
	d := &UpdateDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *UpdateDecl) Clone(args []Node) Node {
	return &UpdateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *UpdateDecl) String() string { return "update" }

// AssertDecl declares an assertion.
type AssertDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewAssertDecl(args ...Node) *AssertDecl {
	d := &AssertDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *AssertDecl) Clone(args []Node) Node {
	return &AssertDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *AssertDecl) String() string { return "assert" }

// InterpretDecl interprets a type.
type InterpretDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewInterpretDecl(args ...Node) *InterpretDecl {
	d := &InterpretDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *InterpretDecl) Clone(args []Node) Node {
	return &InterpretDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *InterpretDecl) String() string { return "interpret" }

// Defines returns names defined by this interpret declaration.
// Matches Python InterpretDecl.defines() (ivy_ast.py:1108-1117).
func (d *InterpretDecl) Defines() []string {
	var res []string
	// label name (for version > 1.6)
	if len(d.DeclArgs) > 0 {
		if lf, ok := d.DeclArgs[0].(*LabeledFormula); ok {
			if lf.Label != nil {
				if la, ok := lf.Label.(*Atom); ok && la.Rep != "" {
					res = append(res, la.Rep)
				}
			}
			// If the RHS of the formula is a Range, add its non-numeric args.
			// Python: for arg in rhs.args: if not arg.rep.isdigit(): ...
			// Use nodeRep to handle both *Atom and *App nodes.
			if lf.Formula != nil {
				if imp, ok := lf.Formula.(*AstImplies); ok {
					if rng, ok := imp.T2.(*AstRange); ok {
						for _, arg := range rng.Args() {
							repStr := NodeRep(arg)
							if repStr == "" {
								continue
							}
							isDigit := true
							for _, c := range repStr {
								if c < '0' || c > '9' {
									isDigit = false
									break
								}
							}
							if !isDigit {
								res = append(res, repStr)
							}
						}
					}
				}
			}
		}
	}
	return res
}

// --- Mixin declarations ---

// MixinDecl declares mixin relationships.
type MixinDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewMixinDecl(args ...Node) *MixinDecl {
	d := &MixinDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *MixinDecl) Clone(args []Node) Node {
	return &MixinDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *MixinDecl) String() string { return "mixin" }

// MixinBeforeDef defines "X before Y".
type MixinBeforeDef struct {
	Base
	MixerNode Node
	MixeeNode Node
}

func (cfg *AstConfig) NewMixinBeforeDef(mixer, mixee Node) *MixinBeforeDef {
	m := &MixinBeforeDef{MixerNode: mixer, MixeeNode: mixee}
	m.Cfg = cfg
	return m
}

func (m *MixinBeforeDef) Args() []Node { return []Node{m.MixerNode, m.MixeeNode} }
func (m *MixinBeforeDef) Clone(args []Node) Node {
	return &MixinBeforeDef{Base: m.Base, MixerNode: args[0], MixeeNode: args[1]}
}
func (m *MixinBeforeDef) String() string {
	return fmt.Sprint(m.MixerNode) + " before " + fmt.Sprint(m.MixeeNode)
}
func (m *MixinBeforeDef) Mixer() string { return nodeRelname(m.MixerNode) }
func (m *MixinBeforeDef) Mixee() string { return nodeRelname(m.MixeeNode) }
func (m *MixinBeforeDef) IsAfter() bool { return false }

// MixinImplementDef defines "X implement Y".
type MixinImplementDef struct {
	Base
	MixerNode Node
	MixeeNode Node
}

func (cfg *AstConfig) NewMixinImplementDef(mixer, mixee Node) *MixinImplementDef {
	m := &MixinImplementDef{MixerNode: mixer, MixeeNode: mixee}
	m.Cfg = cfg
	return m
}

func (m *MixinImplementDef) Args() []Node { return []Node{m.MixerNode, m.MixeeNode} }
func (m *MixinImplementDef) Clone(args []Node) Node {
	return &MixinImplementDef{Base: m.Base, MixerNode: args[0], MixeeNode: args[1]}
}
func (m *MixinImplementDef) String() string {
	return fmt.Sprint(m.MixerNode) + " implement " + fmt.Sprint(m.MixeeNode)
}
func (m *MixinImplementDef) Mixer() string { return nodeRelname(m.MixerNode) }
func (m *MixinImplementDef) Mixee() string { return nodeRelname(m.MixeeNode) }
func (m *MixinImplementDef) IsAfter() bool { return false }

// MixinAfterDef defines "X after Y".
type MixinAfterDef struct {
	Base
	MixerNode Node // the mixer action AST node
	MixeeNode Node // the mixee (target) action AST node
}

func (cfg *AstConfig) NewMixinAfterDef(mixer, mixee Node) *MixinAfterDef {
	m := &MixinAfterDef{MixerNode: mixer, MixeeNode: mixee}
	m.Cfg = cfg
	return m
}

func (m *MixinAfterDef) Args() []Node { return []Node{m.MixerNode, m.MixeeNode} }
func (m *MixinAfterDef) Clone(args []Node) Node {
	return &MixinAfterDef{Base: m.Base, MixerNode: args[0], MixeeNode: args[1]}
}
func (m *MixinAfterDef) String() string {
	return fmt.Sprint(m.MixerNode) + " after " + fmt.Sprint(m.MixeeNode)
}

// Mixer returns the mixer action name (implements isolate.MixinDef).
func (m *MixinAfterDef) Mixer() string { return nodeRelname(m.MixerNode) }

// Mixee returns the mixee (target) action name (implements isolate.MixinDef).
func (m *MixinAfterDef) Mixee() string { return nodeRelname(m.MixeeNode) }

// IsAfter returns true — this is an after-mixin (implements isolate.MixinDef).
func (m *MixinAfterDef) IsAfter() bool { return true }

// nodeRelname extracts a relname string from an AST node.
func nodeRelname(n Node) string {
	if a, ok := n.(*Atom); ok {
		return a.Rep
	}
	return fmt.Sprint(n)
}

// --- Isolate declarations ---

// IsolateDecl declares an isolate.
type IsolateDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewIsolateDecl(args ...Node) *IsolateDecl {
	d := &IsolateDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *IsolateDecl) Clone(args []Node) Node {
	return &IsolateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *IsolateDecl) String() string { return "isolate" }

// Defines returns the names defined by this isolate declaration.
// Matches Python IsolateDecl.defines() (ivy_ast.py:1149-1150).
func (d *IsolateDecl) Defines() []string {
	var names []string
	for _, arg := range d.DeclArgs {
		if idef, ok := arg.(*IsolateDef); ok {
			if len(idef.Elems) > 0 {
				if rep := NodeRep(idef.Elems[0]); rep != "" {
					names = append(names, rep)
				}
			}
		}
	}
	return names
}

// IsolateDef defines an isolate with verified and present components.
type IsolateDef struct {
	Base
	Elems    []Node // args[0] = name, rest = verified + present
	WithArgs int    // number of "with" args at end
	// Trusted marks this as a trusted (unverified) isolate.
	// In Python, this is tracked via isinstance(idef, TrustedIsolateDef).
	// In Go, we use a field since the typed map[string]*IsolateDef
	// cannot store the TrustedIsolateDef subtype.
	Trusted  bool
	IsObject bool // Python: df.is_object — marks isolate as created from object body
}

func (cfg *AstConfig) NewIsolateDef(elems []Node, withArgs int) *IsolateDef {
	i := &IsolateDef{Elems: elems, WithArgs: withArgs}
	i.Cfg = cfg
	return i
}

func (i *IsolateDef) Args() []Node { return i.Elems }
func (i *IsolateDef) Clone(args []Node) Node {
	return &IsolateDef{Base: i.Base, Elems: args, WithArgs: i.WithArgs}
}
func (i *IsolateDef) IsoName() string {
	if len(i.Elems) > 0 {
		if a, ok := i.Elems[0].(*Atom); ok {
			return a.Relname()
		}
	}
	return ""
}
func (i *IsolateDef) Verified() []Node {
	end := len(i.Elems) - i.WithArgs
	if end > 1 {
		return i.Elems[1:end]
	}
	return nil
}
func (i *IsolateDef) Present() []Node {
	start := len(i.Elems) - i.WithArgs
	if start >= 0 && start < len(i.Elems) {
		return i.Elems[start:]
	}
	return nil
}

// Params returns the isolate parameters (terms of the name atom).
// Python: def params(self): return self.args[0].args
func (i *IsolateDef) Params() []Node {
	if len(i.Elems) == 0 {
		return nil
	}
	if a, ok := i.Elems[0].(*Atom); ok {
		return a.Terms
	}
	return nil
}

func (i *IsolateDef) String() string {
	parts := make([]string, len(i.Elems))
	for j, e := range i.Elems {
		parts[j] = fmt.Sprint(e)
	}
	return strings.Join(parts, ", ")
}

// VerifiedNames returns the relnames of verified components.
// Matches Python: set(a.relname for a in isolate.verified())
func (i *IsolateDef) VerifiedNames() []string {
	nodes := i.Verified()
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if a, ok := n.(*Atom); ok {
			names = append(names, a.Relname())
		} else if s, ok := n.(*Symbol); ok {
			names = append(names, s.Rep)
		}
	}
	return names
}

// PresentNames returns the relnames of present (with) components.
// Matches Python: set(a.relname for a in isolate.present())
func (i *IsolateDef) PresentNames() []string {
	nodes := i.Present()
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if a, ok := n.(*Atom); ok {
			names = append(names, a.Relname())
		} else if s, ok := n.(*Symbol); ok {
			names = append(names, s.Rep)
		}
	}
	return names
}

// IsExtract returns false — IsolateDef is an isolate, not an extract.
func (i *IsolateDef) IsExtract() bool {
	return false
}

// TrustedIsolateDef is a trusted (unverified) isolate.
type TrustedIsolateDef struct{ IsolateDef }

func (t *TrustedIsolateDef) Clone(args []Node) Node {
	return &TrustedIsolateDef{IsolateDef: *t.IsolateDef.Clone(args).(*IsolateDef)}
}

// ExtractDef is an extraction target.
type ExtractDef struct{ IsolateDef }

func (cfg *AstConfig) NewExtractDef(idef IsolateDef) *ExtractDef {
	e := &ExtractDef{IsolateDef: idef}
	e.Cfg = cfg
	return e
}

func (e *ExtractDef) Clone(args []Node) Node {
	return &ExtractDef{IsolateDef: *e.IsolateDef.Clone(args).(*IsolateDef)}
}

// ProcessDef is a process definition.
type ProcessDef struct{ ExtractDef }

func (p *ProcessDef) Clone(args []Node) Node {
	return &ProcessDef{ExtractDef: *p.ExtractDef.Clone(args).(*ExtractDef)}
}

func (cfg *AstConfig) NewProcessDef(edef ExtractDef) *ProcessDef {
	p := &ProcessDef{ExtractDef: edef}
	p.Cfg = cfg
	return p
}

// --- Export/Import declarations ---

// ExportDecl declares exports.
type ExportDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewExportDecl(args ...Node) *ExportDecl {
	d := &ExportDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ExportDecl) Clone(args []Node) Node {
	return &ExportDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ExportDecl) String() string { return "export" }

// ExportDef defines what is exported.
type ExportDef struct {
	Base
	ExportedNode Node
	ScopeNode    Node
}

func (cfg *AstConfig) NewExportDef(exported, scope Node) *ExportDef {
	e := &ExportDef{ExportedNode: exported, ScopeNode: scope}
	e.Cfg = cfg
	return e
}

func (e *ExportDef) Args() []Node { return []Node{e.ExportedNode, e.ScopeNode} }
func (e *ExportDef) Clone(args []Node) Node {
	return &ExportDef{Base: e.Base, ExportedNode: args[0], ScopeNode: args[1]}
}
func (e *ExportDef) String() string { return fmt.Sprint(e.ExportedNode) }

// Exported returns the exported action name (implements isolate exporter interface).
// Matches Python ExportDef.exported() → self.args[0].relname
func (e *ExportDef) Exported() string {
	if a, ok := e.ExportedNode.(*Atom); ok {
		return a.Relname()
	}
	return fmt.Sprint(e.ExportedNode)
}

// Scope returns the scope name (implements isolate exporter interface).
// Empty string means global scope.
func (e *ExportDef) Scope() string {
	if e.ScopeNode == nil {
		return ""
	}
	if a, ok := e.ScopeNode.(*Atom); ok {
		r := a.Relname()
		if r == "" {
			return ""
		}
		return r
	}
	s := fmt.Sprint(e.ScopeNode)
	if s == "" || s == "<nil>" {
		return ""
	}
	return s
}

// ImportDecl declares imports.
type ImportDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewImportDecl(args ...Node) *ImportDecl {
	d := &ImportDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ImportDecl) Clone(args []Node) Node {
	return &ImportDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ImportDecl) String() string { return "import" }

// ImportDef defines what is imported.
type ImportDef struct {
	Base
	Imported Node
	Scope    Node
}

func (cfg *AstConfig) NewImportDef(imported, scope Node) *ImportDef {
	i := &ImportDef{Imported: imported, Scope: scope}
	i.Cfg = cfg
	return i
}

func (i *ImportDef) Args() []Node { return []Node{i.Imported, i.Scope} }
func (i *ImportDef) Clone(args []Node) Node {
	return &ImportDef{Base: i.Base, Imported: args[0], Scope: args[1]}
}
func (i *ImportDef) String() string { return fmt.Sprint(i.Imported) }

// PrivateDecl declares private members.
type PrivateDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewPrivateDecl(args ...Node) *PrivateDecl {
	d := &PrivateDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *PrivateDecl) Clone(args []Node) Node {
	return &PrivateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *PrivateDecl) String() string { return "private" }

// AliasDecl declares a type alias.
type AliasDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewAliasDecl(args ...Node) *AliasDecl {
	d := &AliasDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *AliasDecl) Clone(args []Node) Node {
	return &AliasDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *AliasDecl) String() string { return "alias" }

// DelegateDecl declares delegation.
type DelegateDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewDelegateDecl(args ...Node) *DelegateDecl {
	d := &DelegateDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *DelegateDecl) Clone(args []Node) Node {
	return &DelegateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *DelegateDecl) String() string { return "delegate" }

// DelegateDef defines a delegation.
type DelegateDef struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewDelegateDef(elems []Node) *DelegateDef {
	d := &DelegateDef{Elems: elems}
	d.Cfg = cfg
	return d
}

func (d *DelegateDef) Args() []Node           { return d.Elems }
func (d *DelegateDef) Clone(args []Node) Node { return &DelegateDef{Base: d.Base, Elems: args} }
func (d *DelegateDef) String() string         { return "delegate" }

// Delegated returns the delegated action name.
// Matches Python DelegateDef.delegated() → self.args[0].relname.
func (d *DelegateDef) Delegated() string {
	if len(d.Elems) > 0 {
		return nodeRelname(d.Elems[0])
	}
	return ""
}

// Delegee returns the delegee (target) action name.
// Matches Python DelegateDef.delegee() → self.args[1].relname.
func (d *DelegateDef) Delegee() string {
	if len(d.Elems) > 1 {
		return nodeRelname(d.Elems[1])
	}
	return ""
}

// ImplementTypeDecl declares a type implementation.
type ImplementTypeDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewImplementTypeDecl(args ...Node) *ImplementTypeDecl {
	d := &ImplementTypeDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ImplementTypeDecl) Clone(args []Node) Node {
	return &ImplementTypeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ImplementTypeDecl) String() string { return "implementtype" }

// --- Native code ---

// NativeCode holds embedded native code strings.
type NativeCode struct {
	Base
	Code string
}

func (cfg *AstConfig) NewNativeCode(code string) *NativeCode {
	d := &NativeCode{Code: code}
	d.Cfg = cfg
	return d
}

func (n *NativeCode) Args() []Node           { return nil }
func (n *NativeCode) Clone(args []Node) Node { return &NativeCode{Base: n.Base, Code: n.Code} }
func (n *NativeCode) String() string         { return n.Code }

// NativeType wraps a native type expression.
type NativeType struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewNativeType(elems ...Node) *NativeType {
	n := &NativeType{Elems: elems}
	n.Cfg = cfg
	return n
}

func (n *NativeType) Args() []Node           { return n.Elems }
func (n *NativeType) Clone(args []Node) Node { return &NativeType{Base: n.Base, Elems: args} }
func (n *NativeType) String() string         { return "<<<...>>>" }

// NativeExpr wraps a native expression.
type AstNativeExpr struct {
	Base
	Elems []Node
	ASort Node
}

func (n *AstNativeExpr) Args() []Node { return n.Elems }
func (n *AstNativeExpr) Clone(args []Node) Node {
	return &AstNativeExpr{Base: n.Base, Elems: args, ASort: n.ASort}
}
func (n *AstNativeExpr) String() string { return "<<<...>>>" }

func (cfg *AstConfig) NewNativeExpr(elems []Node) *AstNativeExpr {
	n := &AstNativeExpr{Elems: elems}
	n.Cfg = cfg
	return n
}

// NativeDef defines a native block.
type NativeDef struct {
	Base
	Elems []Node
}

func (n *NativeDef) Args() []Node           { return n.Elems }
func (n *NativeDef) Clone(args []Node) Node { return &NativeDef{Base: n.Base, Elems: args} }
func (n *NativeDef) String() string {
	if len(n.Elems) > 0 {
		res := ""
		if n.Elems[0] != nil {
			res = "[" + fmt.Sprint(n.Elems[0]) + "] "
		}
		for _, s := range n.Elems[1:] {
			res += " " + fmt.Sprint(s)
		}
		return res
	}
	return "native"
}

func (cfg *AstConfig) NewNativeDef(elems []Node) *NativeDef {
	n := &NativeDef{Elems: elems}
	n.Cfg = cfg
	return n
}

// NativeDecl declares native code.
type NativeDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewNativeDecl(args ...Node) *NativeDecl {
	d := &NativeDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *NativeDecl) Clone(args []Node) Node {
	return &NativeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *NativeDecl) String() string { return "native" }

// --- Attribute declarations ---

// AttributeDef defines an attribute.
type AttributeDef struct {
	Base
	Name  Node
	Value Node
}

func (cfg *AstConfig) NewAttributeDef(name, value Node) *AttributeDef {
	d := &AttributeDef{Name: name, Value: value}
	d.Cfg = cfg
	return d
}

func (a *AttributeDef) Args() []Node { return []Node{a.Name, a.Value} }
func (a *AttributeDef) Clone(args []Node) Node {
	return &AttributeDef{Base: a.Base, Name: args[0], Value: args[1]}
}
func (a *AttributeDef) String() string {
	return "attribute " + fmt.Sprint(a.Name) + " = " + fmt.Sprint(a.Value)
}

// AttributeDecl declares attributes.
type AttributeDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewAttributeDecl(args ...Node) *AttributeDecl {
	d := &AttributeDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *AttributeDecl) Clone(args []Node) Node {
	return &AttributeDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *AttributeDecl) String() string { return "attribute" }

// --- Instantiation ---

// Instantiation represents "name : sort" instantiation.
type AstInstantiation struct {
	Base
	Name Node // may be nil
	Sort Node
}

func (cfg *AstConfig) NewInstantiation(name, sort Node) *AstInstantiation {
	d := &AstInstantiation{Name: name, Sort: sort}
	d.Cfg = cfg
	return d
}

func (i *AstInstantiation) Args() []Node { return []Node{i.Name, i.Sort} }
func (i *AstInstantiation) Clone(args []Node) Node {
	return &AstInstantiation{Base: i.Base, Name: args[0], Sort: args[1]}
}
func (i *AstInstantiation) String() string {
	if i.Name != nil {
		return fmt.Sprint(i.Name) + " : " + fmt.Sprint(i.Sort)
	}
	return fmt.Sprint(i.Sort)
}

// InstantiateDecl declares instantiation.
type InstantiateDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewInstantiateDecl(args ...Node) *InstantiateDecl {
	d := &InstantiateDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *InstantiateDecl) Clone(args []Node) Node {
	return &InstantiateDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *InstantiateDecl) String() string { return "instantiate" }

// AutoInstanceDecl declares automatic instantiation.
type AutoInstanceDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewAutoInstanceDecl(args ...Node) *AutoInstanceDecl {
	d := &AutoInstanceDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *AutoInstanceDecl) Clone(args []Node) Node {
	return &AutoInstanceDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *AutoInstanceDecl) String() string { return "autoinstance" }

// StateDef defines a state.
type StateDef struct {
	Base
	Name  string
	State Node
}

func (cfg *AstConfig) NewStateDef(name string, state Node) *StateDef {
	d := &StateDef{Name: name, State: state}
	d.Cfg = cfg
	return d
}

func (s *StateDef) Args() []Node {
	a := &Atom{Rep: s.Name}
	a.Cfg = s.Cfg
	return []Node{a, s.State}
}
func (s *StateDef) Clone(args []Node) Node {
	return &StateDef{Base: s.Base, Name: s.Name, State: args[1]}
}
func (s *StateDef) String() string { return s.Name + " = " + fmt.Sprint(s.State) }

// Renaming holds rename mappings.
type AstRenaming struct {
	Base
	Elems []Node
}

func (r *AstRenaming) Args() []Node           { return r.Elems }
func (r *AstRenaming) Clone(args []Node) Node { return &AstRenaming{Base: r.Base, Elems: args} }
func (r *AstRenaming) String() string {
	parts := make([]string, len(r.Elems))
	for i, e := range r.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return "<" + strings.Join(parts, ",") + ">"
}

// --- Scenario types ---

// ScenarioDecl declares a scenario.
type ScenarioDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewScenarioDecl(args ...Node) *ScenarioDecl {
	d := &ScenarioDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *ScenarioDecl) Clone(args []Node) Node {
	return &ScenarioDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *ScenarioDecl) String() string { return "scenario" }

// PlaceList is a comma-separated list of places.
type PlaceList struct {
	Base
	Elems []Node
}

func (p *PlaceList) Args() []Node           { return p.Elems }
func (p *PlaceList) Clone(args []Node) Node { return &PlaceList{Base: p.Base, Elems: args} }

func (cfg *AstConfig) NewPlaceList(elems []Node) *PlaceList {
	p := &PlaceList{Elems: elems}
	p.Cfg = cfg
	return p
}

func (p *PlaceList) String() string {
	parts := make([]string, len(p.Elems))
	for i, e := range p.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return strings.Join(parts, ",")
}

// ScenarioTransition represents a state transition.
type ScenarioTransition struct {
	Base
	From   Node
	To     Node
	Action Node
}

func (s *ScenarioTransition) Args() []Node { return []Node{s.From, s.To, s.Action} }
func (s *ScenarioTransition) Clone(args []Node) Node {
	return &ScenarioTransition{Base: s.Base, From: args[0], To: args[1], Action: args[2]}
}
func (s *ScenarioTransition) String() string {
	return fmt.Sprint(s.From) + "->" + fmt.Sprint(s.To) + " : " + fmt.Sprint(s.Action)
}

func (cfg *AstConfig) NewScenarioTransition(from, to, action Node) *ScenarioTransition {
	s := &ScenarioTransition{From: from, To: to, Action: action}
	s.Cfg = cfg
	return s
}

// ScenarioDef defines a scenario.
type ScenarioDef struct {
	Base
	Elems []Node
}

func (s *ScenarioDef) Args() []Node           { return s.Elems }
func (s *ScenarioDef) Clone(args []Node) Node { return &ScenarioDef{Base: s.Base, Elems: args} }

func (cfg *AstConfig) NewScenarioDef(elems []Node) *ScenarioDef {
	s := &ScenarioDef{Elems: elems}
	s.Cfg = cfg
	return s
}

func (s *ScenarioDef) String() string { return "scenario{...}" }

// InitPlaces returns the initial place list (Elems[0]).
// Python: scen.args[0] — the PlaceList from "-> places"
func (s *ScenarioDef) InitPlaces() *PlaceList {
	if len(s.Elems) == 0 {
		return nil
	}
	if pl, ok := s.Elems[0].(*PlaceList); ok {
		return pl
	}
	return nil
}

// Transitions returns Elems[1:] cast to ScenarioTransition.
func (s *ScenarioDef) Transitions() []*ScenarioTransition {
	var result []*ScenarioTransition
	for _, elem := range s.Elems[1:] {
		if tr, ok := elem.(*ScenarioTransition); ok {
			result = append(result, tr)
		}
	}
	return result
}

// PlaceInfo holds a place name and its source location.
type PlaceInfo struct {
	Name   string
	Lineno Location
}

// Places returns unique place names from init + all transitions.
// Matches Python ScenarioDef.places() (ivy_ast.py:1453-1464).
func (s *ScenarioDef) Places() []PlaceInfo {
	done := make(map[string]bool)
	var places []Node
	if init := s.InitPlaces(); init != nil {
		places = append(places, init.Elems...)
	}
	for _, tr := range s.Transitions() {
		if from, ok := tr.From.(*PlaceList); ok {
			places = append(places, from.Elems...)
		}
		if to, ok := tr.To.(*PlaceList); ok {
			places = append(places, to.Elems...)
		}
	}
	var res []PlaceInfo
	for _, pl := range places {
		if atom, ok := pl.(*Atom); ok {
			if !done[atom.Rep] {
				res = append(res, PlaceInfo{Name: atom.Rep, Lineno: atom.GetLineno()})
				done[atom.Rep] = true
			}
		}
	}
	return res
}

// DefineInfo holds a mixer name and its source location.
type DefineInfo struct {
	Name   string
	Lineno Location
}

// Defines returns mixer names + places.
// Matches Python ScenarioDef.defines() (ivy_ast.py:1465-1474).
func (s *ScenarioDef) Defines() []DefineInfo {
	var res []DefineInfo
	done := make(map[string]bool)
	for _, tr := range s.Transitions() {
		// tr.Action is ScenarioBeforeMixin or ScenarioAfterMixin
		// mixer = tr.args[2].args[0].rep in Python
		var mixer Node
		switch m := tr.Action.(type) {
		case *ScenarioBeforeMixin:
			mixer = m.Mixer
		case *ScenarioAfterMixin:
			mixer = m.Mixer
		}
		if mixer != nil {
			if atom, ok := mixer.(*Atom); ok {
				if !done[atom.Rep] {
					done[atom.Rep] = true
					res = append(res, DefineInfo{Name: atom.Rep, Lineno: atom.GetLineno()})
				}
			}
		}
	}
	for _, pi := range s.Places() {
		res = append(res, DefineInfo{Name: pi.Name, Lineno: pi.Lineno})
	}
	return res
}

// ScenarioBeforeMixin wraps a "before" mixin in a scenario transition.
// Python: ScenarioBeforeMixin(ScenarioMixin) (ivy_ast.py:1438-1440)
type ScenarioBeforeMixin struct {
	Base
	Mixer Node // Atom with generated mixer name (args[0])
	Def   Node // ActionDef (args[1])
}

func (s *ScenarioBeforeMixin) Args() []Node { return []Node{s.Mixer, s.Def} }
func (s *ScenarioBeforeMixin) Clone(args []Node) Node {
	return &ScenarioBeforeMixin{Base: s.Base, Mixer: args[0], Def: args[1]}
}
func (s *ScenarioBeforeMixin) String() string { return "before " + fmt.Sprint(s.Def) }

func (cfg *AstConfig) NewScenarioBeforeMixin(mixer, def Node) *ScenarioBeforeMixin {
	s := &ScenarioBeforeMixin{Mixer: mixer, Def: def}
	s.Cfg = cfg
	return s
}

// ScenarioAfterMixin wraps an "after" mixin in a scenario transition.
// Python: ScenarioAfterMixin(ScenarioMixin) (ivy_ast.py:1442-1444)
type ScenarioAfterMixin struct {
	Base
	Mixer Node // Atom with generated mixer name (args[0])
	Def   Node // ActionDef (args[1])
}

func (s *ScenarioAfterMixin) Args() []Node { return []Node{s.Mixer, s.Def} }
func (s *ScenarioAfterMixin) Clone(args []Node) Node {
	return &ScenarioAfterMixin{Base: s.Base, Mixer: args[0], Def: args[1]}
}
func (s *ScenarioAfterMixin) String() string { return "after " + fmt.Sprint(s.Def) }

func (cfg *AstConfig) NewScenarioAfterMixin(mixer, def Node) *ScenarioAfterMixin {
	s := &ScenarioAfterMixin{Mixer: mixer, Def: def}
	s.Cfg = cfg
	return s
}

// PrivateDef marks a symbol as private.
// Python: ivy_ast.py:1224-1228
type PrivateDef struct {
	Base
	Elems []Node // args
}

func (p *PrivateDef) Args() []Node { return p.Elems }
func (p *PrivateDef) Clone(args []Node) Node {
	return &PrivateDef{Base: p.Base, Elems: args}
}
func (p *PrivateDef) Privatized() string {
	if len(p.Elems) > 0 {
		if r, ok := p.Elems[0].(interface{ Relname() string }); ok {
			return r.Relname()
		}
	}
	return ""
}
func (p *PrivateDef) String() string {
	if len(p.Elems) > 0 {
		return "private " + fmt.Sprint(p.Elems[0])
	}
	return "private"
}

// ImplementTypeDef represents a type implementation declaration.
// Python: ivy_ast.py:1266-1274
type ImplementTypeDef struct {
	Base
	Elems []Node // args: [implemented, implementer]
}

func (d *ImplementTypeDef) Args() []Node { return d.Elems }
func (d *ImplementTypeDef) Clone(args []Node) Node {
	return &ImplementTypeDef{Base: d.Base, Elems: args}
}
func (d *ImplementTypeDef) Implemented() string {
	if len(d.Elems) > 0 {
		if r, ok := d.Elems[0].(interface{ Relname() string }); ok {
			return r.Relname()
		}
	}
	return ""
}
func (d *ImplementTypeDef) Implementer() string {
	if len(d.Elems) > 1 {
		if r, ok := d.Elems[1].(interface{ Relname() string }); ok {
			return r.Relname()
		}
	}
	return ""
}
func (d *ImplementTypeDef) String() string {
	return d.Implemented() + " with " + d.Implementer()
}

func (cfg *AstConfig) NewImplementTypeDef(elems []Node) *ImplementTypeDef {
	d := &ImplementTypeDef{Elems: elems}
	d.Cfg = cfg
	return d
}

// IsolateObjectDecl is an isolate for an object (no defines).
type IsolateObjectDecl struct {
	IsolateDecl
}

func (cfg *AstConfig) NewIsolateObjectDecl(base IsolateDecl) *IsolateObjectDecl {
	d := &IsolateObjectDecl{IsolateDecl: base}
	d.Cfg = cfg
	return d
}

func (d *IsolateObjectDecl) Clone(args []Node) Node {
	return &IsolateObjectDecl{IsolateDecl: *d.IsolateDecl.Clone(args).(*IsolateDecl)}
}

// Defines returns nothing — IsolateObjectDecl does not define names.
// Matches Python IsolateObjectDecl.defines() which returns [].
func (d *IsolateObjectDecl) Defines() []string { return nil }

// SubclassDecl declares a subclass (like object but with a supertype).
// Python: top : top SUBCLASS objsym OF atype EQ LCB optdotdotdot top RCB objectend
type SubclassDecl struct {
	DeclBase
}

func (cfg *AstConfig) NewSubclassDecl(args ...Node) *SubclassDecl {
	d := &SubclassDecl{DeclBase: DeclBase{DeclArgs: args}}
	d.Cfg = cfg
	return d
}

func (d *SubclassDecl) Clone(args []Node) Node {
	return &SubclassDecl{DeclBase: DeclBase{Base: d.Base, DeclArgs: args}}
}
func (d *SubclassDecl) String() string { return "subclass" }

// PatternBasedUpdate represents an update declaration.
// Python: PatternBasedUpdate(SymbolList(*dfns), SymbolList(*deps), UpdatePatternList(*patterns))
type AstPatternBasedUpdate struct {
	Base
	Dfns     Node // SymbolList of defined symbols
	Deps     Node // SymbolList of dependency symbols
	Patterns Node // UpdatePatternList of patterns
}

func (cfg *AstConfig) NewPatternBasedUpdate(dfns, deps, patterns Node) *AstPatternBasedUpdate {
	d := &AstPatternBasedUpdate{Dfns: dfns, Deps: deps, Patterns: patterns}
	d.Cfg = cfg
	return d
}

func (p *AstPatternBasedUpdate) Args() []Node { return []Node{p.Dfns, p.Deps, p.Patterns} }
func (p *AstPatternBasedUpdate) Clone(args []Node) Node {
	return &AstPatternBasedUpdate{Base: p.Base, Dfns: args[0], Deps: args[1], Patterns: args[2]}
}
func (p *AstPatternBasedUpdate) String() string {
	return "update " + fmt.Sprint(p.Dfns) + " from " + fmt.Sprint(p.Deps)
}

// UpdatePattern represents a single update pattern.
// Python: UpdatePattern(ConstantDecl(*params), action, requires, ensures)
type AstUpdatePattern struct {
	Base
	Params   Node // ConstantDecl of params
	Action   Node // the action body
	Requires Node // requires formula
	Ensures  Node // ensures formula
}

func (cfg *AstConfig) NewUpdatePattern(params, action, requires, ensures Node) *AstUpdatePattern {
	d := &AstUpdatePattern{Params: params, Action: action, Requires: requires, Ensures: ensures}
	d.Cfg = cfg
	return d
}

func (u *AstUpdatePattern) Args() []Node { return []Node{u.Params, u.Action, u.Requires, u.Ensures} }
func (u *AstUpdatePattern) Clone(args []Node) Node {
	return &AstUpdatePattern{Base: u.Base, Params: args[0], Action: args[1], Requires: args[2], Ensures: args[3]}
}
func (u *AstUpdatePattern) String() string { return "params ... in ... -> ..." }

// UpdatePatternList holds a list of update patterns.
type AstUpdatePatternList struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewUpdatePatternList(elems ...Node) *AstUpdatePatternList {
	d := &AstUpdatePatternList{Elems: elems}
	d.Cfg = cfg
	return d
}

func (u *AstUpdatePatternList) Args() []Node { return u.Elems }
func (u *AstUpdatePatternList) Clone(args []Node) Node {
	return &AstUpdatePatternList{Base: u.Base, Elems: args}
}
func (u *AstUpdatePatternList) String() string { return "updatepatterns" }

// SymbolList holds a list of symbol names.
// Python: SymbolList(*names) where names are strings.
type AstSymbolList struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewSymbolList(elems ...Node) *AstSymbolList {
	d := &AstSymbolList{Elems: elems}
	d.Cfg = cfg
	return d
}

func (s *AstSymbolList) Args() []Node           { return s.Elems }
func (s *AstSymbolList) Clone(args []Node) Node { return &AstSymbolList{Base: s.Base, Elems: args} }
func (s *AstSymbolList) String() string {
	parts := make([]string, len(s.Elems))
	for i, e := range s.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return strings.Join(parts, ",")
}

// RME represents requires/modifies/ensures state expressions.
// Python: RME(requires, modifies, ensures)
type AstRME struct {
	Base
	RequiresFmla Node   // requires formula
	ModifiesList []Node // modifies list (nil = *, empty = {})
	EnsuresFmla  Node   // ensures formula
}

func (cfg *AstConfig) NewRME(requires Node, modifies []Node, ensures Node) *AstRME {
	d := &AstRME{RequiresFmla: requires, ModifiesList: modifies, EnsuresFmla: ensures}
	d.Cfg = cfg
	return d
}

func (r *AstRME) Args() []Node { return []Node{r.RequiresFmla, r.EnsuresFmla} }
func (r *AstRME) Clone(args []Node) Node {
	return &AstRME{Base: r.Base, RequiresFmla: args[0], ModifiesList: r.ModifiesList, EnsuresFmla: args[1]}
}
func (r *AstRME) String() string { return "{requires ... modifies ... ensures ...}" }

// NamedSpace wraps a literal in a concept space expression.
// Python: NamedSpace(Literal(polarity, atom))
type AstNamedSpace struct {
	Base
	Lit Node // a Literal
}

func (cfg *AstConfig) NewNamedSpace(lit Node) *AstNamedSpace {
	d := &AstNamedSpace{Lit: lit}
	d.Cfg = cfg
	return d
}

func (n *AstNamedSpace) Args() []Node           { return []Node{n.Lit} }
func (n *AstNamedSpace) Clone(args []Node) Node { return &AstNamedSpace{Base: n.Base, Lit: args[0]} }
func (n *AstNamedSpace) String() string         { return fmt.Sprint(n.Lit) }

// ProductSpace represents a product of concept space expressions.
// Python: ProductSpace([expr1, expr2, ...])
type AstProductSpace struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewProductSpace(elems ...Node) *AstProductSpace {
	d := &AstProductSpace{Elems: elems}
	d.Cfg = cfg
	return d
}

func (p *AstProductSpace) Args() []Node           { return p.Elems }
func (p *AstProductSpace) Clone(args []Node) Node { return &AstProductSpace{Base: p.Base, Elems: args} }
func (p *AstProductSpace) String() string {
	parts := make([]string, len(p.Elems))
	for i, e := range p.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return strings.Join(parts, " * ")
}

// SumSpace represents a sum of concept space expressions.
// Python: SumSpace([expr1, expr2, ...])
type AstSumSpace struct {
	Base
	Elems []Node
}

func (cfg *AstConfig) NewSumSpace(elems ...Node) *AstSumSpace {
	d := &AstSumSpace{Elems: elems}
	d.Cfg = cfg
	return d
}

func (s *AstSumSpace) Args() []Node           { return s.Elems }
func (s *AstSumSpace) Clone(args []Node) Node { return &AstSumSpace{Base: s.Base, Elems: args} }
func (s *AstSumSpace) String() string {
	parts := make([]string, len(s.Elems))
	for i, e := range s.Elems {
		parts[i] = fmt.Sprint(e)
	}
	return strings.Join(parts, " + ")
}

// DeclName returns the Python-compatible dispatch name for a declaration node.
// Matches Python's decl.name() used in IvyDeclInterp.__call__.
func DeclName(decl Node) string {
	switch decl.(type) {
	case *TypeDecl:
		return "type"
	case *AxiomDecl:
		return "axiom"
	case *PropertyDecl:
		return "property"
	case *ConjectureDecl:
		return "conjecture"
	case *RelationDecl:
		return "relation"
	case *ConstantDecl:
		return "individual"
	case *DerivedDecl:
		return "derived"
	case *DefinitionDecl:
		return "definition"
	case *ActionDecl:
		return "action"
	case *InitDecl:
		return "init"
	case *ObjectDecl:
		return "object"
	case *ModuleDecl:
		return "module"
	case *VariantDecl:
		return "variant"
	case *ExportDecl:
		return "export"
	case *ImportDecl:
		return "import"
	case *IsolateDecl:
		return "isolate"
	case *InterpretDecl:
		return "interpret"
	case *MixinDecl:
		return "mixin"
	case *DelegateDecl:
		return "delegate"
	case *NativeDecl:
		return "native"
	case *AliasDecl:
		return "alias"
	case *AttributeDecl:
		return "attribute"
	case *ProgressDecl:
		return "progress"
	case *PrivateDecl:
		return "private"
	case *SchemaDecl:
		return "schema"
	case *InstantiateDecl:
		return "instantiate"
	case *AstProofDecl:
		return "proof"
	case *NamedDecl:
		return "named"
	case *TheoremDecl:
		return "theorem"
	case *AssertDecl:
		return "_assert"
	case *ParameterDecl:
		return "parameter"
	case *DestructorDecl:
		return "destructor"
	case *ConstructorDecl:
		return "constructor"
	case *ConceptDecl:
		return "concept"
	case *RelyDecl:
		return "rely"
	case *MixOrdDecl:
		return "mixord"
	case *UpdateDecl:
		return "update"
	case *ScenarioDecl:
		return "scenario"
	case *ImplementTypeDecl:
		return "implementtype"
	case *AutoInstanceDecl:
		return "autoinstance"
	case *IsolateObjectDecl:
		return "isolate"
	default:
		return fmt.Sprintf("unknown(%T)", decl)
	}
}
