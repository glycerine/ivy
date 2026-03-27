// Package compiler transforms AST nodes (ast/ package) into logic IR nodes
// (logic/ and ivylogic/ packages). It replaces the Python monkey-patching
// pattern in ivy_compiler.py with a Compiler struct using a visitor/switch
// dispatch.
//
// This corresponds to Python's ivy_compiler.py.
package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/typeinfer"
	"github.com/glycerine/goivy/xtracer"
)

// typeName returns the bare struct name for a value, stripping pointer and
// package prefix. e.g. *ast.Variable → "Variable". Matches Python's
// type(self).__name__ output for cross-language trace comparison.
func typeName(v interface{}) string {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// ActionInfo holds metadata about a declared action: its formal parameters,
// formal return values, and key-argument position.
// FormalAST/FormalRetAST hold the AST-level formals before compilation;
// Params/Returns hold the compiled logic-level symbols.
type ActionInfo struct {
	FormalAST    []ast.Node   // AST-level formal parameters (pre-compilation)
	FormalRetAST []ast.Node   // AST-level formal returns (pre-compilation)
	Params       []*lg.Symbol // compiled formal parameters
	Returns      []*lg.Symbol // compiled formal returns
	KeyPos       int          // index of first KeyArg in formals
}

// ReturnContext tracks the return values for the current expression compilation.
type ReturnContext struct {
	Values []lg.Expr
	Lineno *ast.Location
}

// ExprContext tracks intermediate code and local symbols generated while
// compiling an expression (e.g., inline action calls in an rhs).
type ExprContext struct {
	Code      []lg.Expr // accumulated action nodes (wrapped)
	LocalSyms []*lg.Symbol
	Lineno    *ast.Location
}

// CompileInlineCode produces a single action from the accumulated code.
// Matches Python ivy_compiler.py compile_inline_call which wraps multiple
// statements in Sequence/LocalAction.
func (ec *ExprContext) CompileInlineCode() lg.Expr {
	if len(ec.Code) == 0 {
		return nil
	}
	if len(ec.Code) == 1 {
		return ec.Code[0]
	}
	// Multiple code elements → wrap in a Sequence action.
	return actions.WrapAction(actions.NewSequence(ec.Code...))
}

// Extract produces a single action from the accumulated code and local symbols.
// Matches Python ExprContext.extract() (ivy_compiler.py lines 116-123):
//   - Sets lineno on all code items
//   - If 1 code item → return it directly
//   - If multiple items → wrap in LocalAction(*(self.local_syms + [Sequence(*self.code)]))
func (ec *ExprContext) Extract() lg.Expr {
	xtracer.Trace("compiler.ExprContext.Extract ENTER")
	// Set lineno on all code items (Python lines 117-118)
	for _, c := range ec.Code {
		if act := actions.UnwrapAction(c); act != nil && ec.Lineno != nil {
			act.SetLineno(*ec.Lineno)
		}
	}
	// Single code item → return directly (Python lines 119-120)
	if len(ec.Code) == 1 {
		return ec.Code[0]
	}
	// Multiple items → wrap in LocalAction(local_syms..., Sequence(code...))
	args := make([]lg.Expr, 0, len(ec.LocalSyms)+1)
	for _, s := range ec.LocalSyms {
		args = append(args, s)
	}
	args = append(args, actions.WrapAction(actions.NewSequence(ec.Code...)))
	res := actions.NewLocalAction(args...)
	if ec.Lineno != nil {
		res.SetLineno(*ec.Lineno)
	}
	return actions.WrapAction(res)
}

// TopContext holds the action metadata used during compilation.
type TopContext struct {
	Actions map[string]*ActionInfo
}

// VariableContext maps variable names to their sorts, used when compiling
// under quantifier scopes.
type VariableContext struct {
	Map map[string]lg.Sort
}

// Compiler transforms AST nodes to logic IR. It carries the compilation
// state that Python stored in global variables.
type Compiler struct {
	// Sig is the current first-order signature.
	Sig *il.Sig

	// Module is the module being compiled.
	Module *module.Module

	// ReturnCtx tracks the return context for expression compilation.
	// nil means "no return context".
	ReturnCtx *ReturnContext

	// ExprCtx tracks the expression compilation context.
	// nil when not inside an expression context.
	ExprCtx *ExprContext

	// TopCtx holds action metadata for the current top-level scope.
	TopCtx *TopContext

	// VarCtx maps variable names to sorts within quantifier scopes.
	VarCtx *VariableContext

	// SigMerkle is a rolling Merkle hash for Sig conformance auditing.
	// Each SigCheck() call feeds the Sig's canon into this chain.
	SigMerkle iu.MerkleState
}

// SigCheck emits a Merkle-chained HASH trace of the current Sig + Module state.
// The golden test detects divergence via the HASH roots and uses DiffSexp
// on the canon= data to show exactly which sorts/symbols/declarations differ.
//
// IMPORTANT: Only call inside `if xtracer.Enabled { c.SigCheck(...) }` blocks.
// When -tags xtracer_off is set, Enabled is const false, so the compiler
// eliminates the entire block — Canon() string building and Merkle hashing
// have zero production cost.
func (c *Compiler) SigCheck(label string) {
	sigCanon := c.Sig.Canon()
	modCanon := iu.Canonical("")
	if c.Module != nil {
		modCanon = c.Module.Canon()
	}
	combined := iu.Canonical(string(sigCanon) + string(modCanon))
	leaf, root := c.SigMerkle.AddLeaf(combined)
	xtracer.Trace("compiler.SigCheck@%s HASH leaf=%s root=%s canon=%s", label, leaf, root, string(combined))
}

// New creates a new Compiler with the given signature and module.
func New(sig *il.Sig, mod *module.Module) *Compiler {
	c := &Compiler{
		Sig:    sig,
		Module: mod,
		VarCtx: &VariableContext{Map: make(map[string]lg.Sort)},
	}
	if sig == nil {
		c.Sig = il.NewSig()
	}
	if mod == nil {
		c.Module = module.New()
	}
	return c
}

// NewFromModule creates a Compiler using the module's own signature.
func NewFromModule(mod *module.Module) *Compiler {
	return New(mod.Sig, mod)
}

// --- Main dispatch ---

// CompileNode is the main visitor dispatch. It compiles any AST node to
// the corresponding logic IR node.
func (c *Compiler) CompileNode(node ast.Node) (lg.Expr, error) {
	if node == nil {
		xtracer.Trace("compiler.CompileNode return case=nil")
		return nil, fmt.Errorf("cannot compile nil node")
	}
	xtracer.Trace(fmt.Sprintf("compiler.CompileNode ENTER type=%s", typeName(node)))
	switch n := node.(type) {
	// --- Formula operators ---
	case *ast.And:
		xtracer.Trace("compiler.CompileNode return case=And")
		return c.compileAnd(n)
	case *ast.Or:
		xtracer.Trace("compiler.CompileNode return case=Or")
		return c.compileOr(n)
	case *ast.Not:
		xtracer.Trace("compiler.CompileNode return case=Not")
		return c.compileNot(n)
	case *ast.Implies:
		xtracer.Trace("compiler.CompileNode return case=Implies")
		return c.compileImplies(n)
	case *ast.Iff:
		xtracer.Trace("compiler.CompileNode return case=Iff")
		return c.compileIff(n)
	case *ast.Ite:
		xtracer.Trace("compiler.CompileNode return case=Ite")
		return c.compileIte(n)
	case *ast.Definition:
		xtracer.Trace("compiler.CompileNode return case=Definition")
		return c.compileDefinition(n)
	case *ast.Globally:
		xtracer.Trace("compiler.CompileNode return case=Globally")
		return c.compileGlobally(n)
	case *ast.Eventually:
		xtracer.Trace("compiler.CompileNode return case=Eventually")
		return c.compileEventually(n)
	case *ast.WhenOperator:
		xtracer.Trace("compiler.CompileNode return case=WhenOperator")
		return c.compileWhenOperator(n)

	// --- Quantifiers ---
	case *ast.Forall:
		xtracer.Trace("compiler.CompileNode return case=Forall")
		return c.CompileQuantifier(n)
	case *ast.Exists:
		xtracer.Trace("compiler.CompileNode return case=Exists")
		return c.CompileQuantifier(n)

	// --- Terms ---
	case *ast.Atom:
		xtracer.Trace("compiler.CompileNode return case=Atom")
		return c.CompileApp(n, false)
	case *ast.App:
		xtracer.Trace("compiler.CompileNode return case=App")
		return c.compileAppNode(n)
	case *ast.Variable:
		xtracer.Trace("compiler.CompileNode return case=Variable")
		return c.CompileVariable(n)
	case *ast.Old:
		xtracer.Trace("compiler.CompileNode return case=Old")
		return c.compileOld(n)
	case *ast.MethodCall:
		xtracer.Trace("compiler.CompileNode return case=MethodCall")
		return c.compileMethodCall(n)

	// --- Symbol (bare identifier like x, y, used as terms in atoms) ---
	case *ast.Symbol:
		xtracer.Trace("compiler.CompileNode return case=Symbol")
		return c.compileSymbol(n)

	// --- Named binder ---
	case *ast.NamedBinder:
		xtracer.Trace("compiler.CompileNode return case=NamedBinder")
		return c.compileNamedBinder(n)

	// --- Labeled formula ---
	case *ast.LabeledFormula:
		xtracer.Trace("compiler.CompileNode return case=LabeledFormula")
		return c.compileLabeledFormula(n)

	// --- NativeExpr ---
	case *ast.NativeExpr:
		xtracer.Trace("compiler.CompileNode return case=NativeExpr")
		return c.compileNativeExpr(n)

	// --- Trigger ---
	case *ast.Trigger:
		xtracer.Trace("compiler.CompileNode return case=Trigger")
		return c.compileTrigger(n)

	// --- CompiledNode: already-compiled expression wrapper ---
	case *ast.CompiledNode:
		if expr, ok := n.Node.(lg.Expr); ok {
			xtracer.Trace("compiler.CompileNode return case=CompiledNode")
			return expr, nil
		}
		xtracer.Trace("compiler.CompileNode return case=CompiledNode-error")
		return nil, fmt.Errorf("CompiledNode does not contain lg.Expr: %T", n.Node)

	// --- Action AST nodes (from LALR parser) ---
	// These have sort_infer_root = True in Python. Route through
	// CompileActionBody which produces actions.Action (satisfies lg.Expr).
	case *ast.AssignAction, *ast.SetAction, *ast.HavocAction,
		*ast.AssumeAction, *ast.AssertAction:
		act, err := c.CompileActionBody(node)
		if err != nil {
			xtracer.Trace("compiler.CompileNode return case=Action-error")
			return nil, err
		}
		xtracer.Trace("compiler.CompileNode return case=Action")
		return actions.WrapAction(act), nil

	// --- Default: Python's AST.cmpl = other_thing ---
	// Handles all other unrecognized AST types.
	default:
		xtracer.Trace(fmt.Sprintf("compiler.CompileNode return case=default type=%s", typeName(node)))
		return c.OtherThing(node)
	}
}

// compileSymbol compiles a bare symbol node (like `x` in `r(x)`).
// It looks up the name in the signature to find its sort.
func (c *Compiler) compileSymbol(n *ast.Symbol) (lg.Expr, error) {
	name := n.Rep

	// Check variable context first (quantifier-bound variables)
	if sort, ok := c.VarCtx.Map[name]; ok {
		v, err := lg.NewVariable(name, sort)
		return v, err
	}

	// Look up in signature (action parameters, constants, relations)
	entry, ok := c.Sig.Symbols[name]
	if ok {
		return lg.NewSymbol(name, entry.Sort), nil
	}

	// Uppercase names are variables (Ivy convention)
	if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
		var sort lg.Sort = lg.TopS
		if n.Sort != nil {
			sortName := extractSortName(n.Sort)
			if sortName != "" {
				if s, ok2 := c.Sig.Sorts[sortName]; ok2 {
					sort = s
				}
			}
		}
		v, err := lg.NewVariable(name, sort)
		return v, err
	}

	// Lowercase unresolved names become constants with TopSort
	return lg.NewSymbol(name, lg.TopS), nil
}

// compileGeneric is the fallback for non-sort_infer_root nodes: compile each
// child and clone. For nodes with sort_infer_root, use OtherThing instead.
// Matches the else branch of Python's other_thing():
//
//	return self.clone([a.compile() for a in self.args])
func (c *Compiler) compileGeneric(node ast.Node) (lg.Expr, error) {
	args := node.Args()
	compiled := make([]ast.Node, len(args))
	for i, a := range args {
		r, err := c.Thing(a)
		if err != nil {
			return nil, fmt.Errorf("compiling arg %d of %T: %w", i, node, err)
		}
		compiled[i] = r
	}
	// Python: self.clone([a.compile() for a in self.args])
	result := node.Clone(compiled)
	if expr, ok := result.(lg.Expr); ok {
		return expr, nil
	}
	// Cloned node isn't lg.Expr — extract compiled exprs and combine
	exprs := make([]lg.Expr, 0, len(compiled))
	for _, c := range compiled {
		if e, ok := c.(lg.Expr); ok {
			exprs = append(exprs, e)
		}
	}
	if len(exprs) == 0 {
		return lg.True, nil
	}
	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return &lg.And{Terms: exprs}, nil
}

// --- Formula compilation ---

// compileArgs compiles all children of an AST node with no return context.
func (c *Compiler) compileArgs(node ast.Node) ([]lg.Expr, error) {
	xtracer.Trace("compiler.CompileArgs ENTER")
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	defer func() { c.ReturnCtx = saved }()

	args := node.Args()
	result := make([]lg.Expr, len(args))
	for i, a := range args {
		r, err := c.Thing(a)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	xtracer.Trace("compiler.CompileArgs return")
	return result, nil
}

func (c *Compiler) compileAnd(n *ast.And) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileAnd ENTER")
	args, err := c.compileArgs(n)
	if err != nil {
		return nil, err
	}
	return &lg.And{Terms: args}, nil
}

func (c *Compiler) compileOr(n *ast.Or) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileOr ENTER")
	args, err := c.compileArgs(n)
	if err != nil {
		return nil, err
	}
	return &lg.Or{Terms: args}, nil
}

func (c *Compiler) compileNot(n *ast.Not) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileNot ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	return &lg.Not{Body: args[0]}, nil
}

func (c *Compiler) compileImplies(n *ast.Implies) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileImplies ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &lg.Implies{T1: args[0], T2: args[1]}, nil
}

func (c *Compiler) compileIff(n *ast.Iff) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileIff ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &lg.Iff{T1: args[0], T2: args[1]}, nil
}

func (c *Compiler) compileIte(n *ast.Ite) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileIte ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 3 {
		return nil, err
	}
	return &lg.Ite{ISort: args[1].NodeSort(), Cond: args[0], Then: args[1], Else: args[2]}, nil
}

func (c *Compiler) compileDefinition(n *ast.Definition) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileDefinition ENTER\n  (via CompileNode op_pairs path, NOT compile_defn)")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return il.NewDefinition(args[0], args[1]), nil
}

func (c *Compiler) compileGlobally(n *ast.Globally) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileGlobally ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	empty := ""
	return &lg.Globally{Environ: &empty, Body: args[0]}, nil
}

func (c *Compiler) compileEventually(n *ast.Eventually) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileEventually ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	empty := ""
	return &lg.Eventually{Environ: &empty, Body: args[0]}, nil
}

func (c *Compiler) compileWhenOperator(n *ast.WhenOperator) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileWhenOperator ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &lg.WhenOperator{WSort: args[0].NodeSort(), Name: n.Name, T1: args[0], T2: args[1]}, nil
}

// --- Term compilation ---

// CompileApp compiles function application (Atom). This corresponds to
// Python's compile_app.
func (c *Compiler) CompileApp(n *ast.Atom, old bool) (lg.Expr, error) {
	xtracer.Trace("compiler.compile_app ENTER old=%v", old)
	rep := ResolveAlias(n.Rep, c.Module)

	// Handle boolean literals
	if rep == "true" || rep == "false" {
		if len(n.Terms) > 0 {
			return nil, lg.NewIvyError(n, fmt.Sprintf("%s is not a function", rep))
		}
		if rep == "true" {
			return &lg.And{}, nil // empty And = true
		}
		return &lg.Or{}, nil // empty Or = false
	}

	// B4-R5: Python debug print: if any(isinstance(a,ivy_logic.Variable) for a in self.args): print("foo!")
	// lg.Variable satisfies ast.Node, so check for it in the AST args before compilation
	for _, a := range n.Terms {
		if _, ok := a.(*lg.Variable); ok {
			pp("foo!: %s", n)
			break
		}
	}

	// Compile arguments with no return context
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	xtracer.Trace("compiler.CompileApp args loop nTerms=%d\n  rep=%s old=%v", len(n.Terms), rep, old)
	args := make([]lg.Expr, len(n.Terms))
	for i, a := range n.Terms {
		r, err := c.Thing(a)
		if err != nil {
			xtracer.Trace("compiler.CompileApp args loop error at i=%d\n  err=%v", i, err)
			c.ReturnCtx = saved
			return nil, err
		}
		args[i] = r
	}
	c.ReturnCtx = saved

	// Handle inline action calls in rhs of assignment
	if c.ExprCtx != nil && c.TopCtx != nil {
		if _, ok := c.TopCtx.Actions[rep]; ok {
			return c.CompileInlineCall(n, args, false)
		}
	}

	// Try to find the symbol
	if rep == "=" {
		// Equality
		if len(args) != 2 {
			return nil, lg.NewIvyError(n, "equality requires exactly 2 arguments")
		}
		return &lg.Eq{T1: args[0], T2: args[1]}, nil
	}

	// Look up polymorphic symbol first
	sym, found := il.FindPolymorphicSymbol(rep)
	if !found {
		// Look up in signature
		entry, ok := c.Sig.Symbols[rep]
		if ok {
			sym = lg.NewSymbol(rep, entry.Sort)
		}
	}

	if sym != nil {
		// Handle numerals with explicit sort annotation
		// B4-R6: Python guards with `sym is not ivy_logic.Equals` here, but
		// Go handles "=" with an early return above, so that guard is already satisfied.
		if il.IsNumeral(sym) {
			if n.ASort != nil {
				sortName := extractSortName(n.ASort)
				if sortName != "S" {
					s, err := c.CmplSort(sortName)
					if err == nil {
						sym = lg.NewSymbol(sym.Name, s)
					}
				}
			}
		}

		if old {
			sym = lg.NewSymbol("old_"+sym.Name, sym.CSort)
		}
		if len(args) == 0 {
			return sym, nil
		}
		result, err := lg.NewApply(sym, args...)
		if err != nil {
			// Sort mismatch in application: try field reference as fallback
			return c.CompileFieldReference(rep, args, n.GetLineno(), old)
		}
		return result, nil
	}

	// Symbol not found: try field reference
	return c.CompileFieldReference(rep, args, n.GetLineno(), old)
}

// compileAppNode compiles an App node (function application with a Node rep).
func (c *Compiler) compileAppNode(n *ast.App) (lg.Expr, error) {
	// If the rep is a Symbol, treat it like an Atom.
	if sym, ok := n.Rep.(*ast.Symbol); ok {
		cfg := c.Module.Cfg.AstCfg
		atom := cfg.NewAtom(sym.Rep, n.Terms...)
		atom.SetLineno(n.GetLineno())
		atom.ASort = n.ASort
		return c.CompileApp(atom, false)
	}
	// If the rep is a NamedBinder, compile it and apply.
	repNode, err := c.Thing(n.Rep)
	if err != nil {
		return nil, err
	}

	saved := c.ReturnCtx
	c.ReturnCtx = nil
	args := make([]lg.Expr, len(n.Terms))
	for i, a := range n.Terms {
		r, err := c.Thing(a)
		if err != nil {
			c.ReturnCtx = saved
			return nil, err
		}
		args[i] = r
	}
	c.ReturnCtx = saved

	if len(args) == 0 {
		return repNode, nil
	}
	return lg.NewApply(repNode, args...)
}

// CompileVariable compiles a Variable AST node to a logic.Variable.
func (c *Compiler) CompileVariable(n *ast.Variable) (lg.Expr, error) {
	xtracer.Trace("compiler.compile_variable ENTER")
	sort, err := c.variableSort(n)
	if err != nil {
		xtracer.Trace("compiler.CompileVariable error\n  var=%s vsort=%s err=%v", n.Rep, n.VSort, err)
		return nil, err
	}
	// If sort is top, check variable context
	if _, isTop := sort.(*lg.TopSort); isTop {
		if s, ok := c.VarCtx.Map[n.Rep]; ok {
			sort = s
		}
	}
	v, err := lg.NewVariable(n.Rep, sort)
	if err != nil {
		return nil, lg.NewIvyError(n, fmt.Sprintf("bad variable: %s", err))
	}
	return v, nil
}

// variableSort resolves the sort of a variable AST node.
func (c *Compiler) variableSort(v *ast.Variable) (lg.Sort, error) {
	if v.VSort == "" || v.VSort == "S" {
		return lg.TopS, nil
	}
	return c.CmplSort(v.VSort)
}

// CmplSort resolves a sort name to a logic.Sort, applying alias resolution.
func (c *Compiler) CmplSort(name string) (lg.Sort, error) {
	resolved := ResolveAlias(name, c.Module)
	sort, err := c.Sig.FindSort(resolved, false)
	if err != nil {
		xtracer.Trace("compiler.CmplSort error\n  name=%s resolved=%s sigSorts=%v err=%v", name, resolved, c.Sig.SortNames(), err)
	}
	return sort, err
}

// compileOld compiles the Old operator: compile the inner term with old=true.
func (c *Compiler) compileOld(n *ast.Old) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileOld ENTER")
	// The inner term should be an Atom or App
	if atom, ok := n.Term.(*ast.Atom); ok {
		return c.CompileApp(atom, true)
	}
	if app, ok := n.Term.(*ast.App); ok {
		if sym, ok := app.Rep.(*ast.Symbol); ok {
			cfg := c.Module.Cfg.AstCfg
			atom := cfg.NewAtom(sym.Rep, app.Terms...)
			atom.SetLineno(n.GetLineno())
			atom.ASort = app.ASort
			return c.CompileApp(atom, true)
		}
	}
	// Fallback: compile the inner term normally
	return c.Thing(n.Term)
}

// compileMethodCall compiles obj.method() style calls.
func (c *Compiler) compileMethodCall(n *ast.MethodCall) (lg.Expr, error) {
	xtracer.Trace("compiler.compile_method_call ENTER")
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	base, err := c.Thing(n.Obj)
	if err != nil {
		c.ReturnCtx = saved
		return nil, err
	}

	// Method is typically an Atom or App
	var childName string
	var methodArgs []lg.Expr
	switch m := n.Method.(type) {
	case *ast.Atom:
		childName = m.Rep
		methodArgs = make([]lg.Expr, len(m.Terms))
		for i, a := range m.Terms {
			r, err := c.Thing(a)
			if err != nil {
				c.ReturnCtx = saved
				return nil, err
			}
			methodArgs[i] = r
		}
	case *ast.Symbol:
		childName = m.Rep
	default:
		c.ReturnCtx = saved
		return nil, lg.NewIvyError(n, fmt.Sprintf("unsupported method node type: %T", n.Method))
	}
	c.ReturnCtx = saved

	sort := base.NodeSort()
	if _, isTop := sort.(*lg.TopSort); isTop {
		return nil, lg.NewIvyError(n, fmt.Sprintf(
			"cannot apply method notation to %s because its type is not inferred", base))
	}

	// Look for the method as a child of the sort
	destrName := iu.ComposeNames(il.SortName(sort), childName)
	if c.TopCtx != nil {
		if _, inSig := c.Sig.Symbols[destrName]; !inSig {
			if _, inAct := c.TopCtx.Actions[destrName]; !inAct {
				// Try sibling of the sort
				pc := iu.ParentChildName(il.SortName(sort))
				destrName = iu.ComposeNames(pc[0], childName)
			}
		}
	}

	// Check if it's an action call
	if c.TopCtx != nil {
		if _, ok := c.TopCtx.Actions[destrName]; ok {
			if c.ExprCtx == nil {
				return nil, lg.NewIvyError(n, fmt.Sprintf(
					"call to action %s not allowed outside an action", destrName))
			}
			allArgs := append([]lg.Expr{base}, methodArgs...)
			cfg := c.Module.Cfg.AstCfg
			atom := cfg.NewAtom(destrName)
			atom.SetLineno(n.GetLineno())
			return c.CompileInlineCall(atom, allArgs, true)
		}
	}

	// Find the destructor symbol
	sym, err := c.findSymbol(destrName)
	if err != nil {
		return nil, err
	}
	allArgs := append([]lg.Expr{base}, methodArgs...)
	if len(allArgs) == 0 {
		return sym, nil
	}
	return lg.NewApply(sym, allArgs...)
}

// compileNamedBinder compiles an AST NamedBinder.
func (c *Compiler) compileNamedBinder(n *ast.NamedBinder) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileNamedBinder ENTER")
	vars := make([]*lg.Variable, len(n.Bounds))
	for i, b := range n.Bounds {
		compiled, err := c.Thing(b)
		if err != nil {
			return nil, err
		}
		v, ok := compiled.(*lg.Variable)
		if !ok {
			return nil, lg.NewIvyError(n, fmt.Sprintf(
				"named binder bound %d is not a variable: %T", i, compiled))
		}
		vars[i] = v
	}
	body, err := c.Thing(n.Body)
	if err != nil {
		return nil, err
	}
	return lg.NewNamedBinder(n.Name, vars, nil, body)
}

// compileLabeledFormula compiles a LabeledFormula AST node.
func (c *Compiler) compileLabeledFormula(n *ast.LabeledFormula) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileLabeledFormula ENTER")
	var label lg.Expr
	if n.Label != nil {
		l, err := c.SortifyWithInference(n.Label)
		if err != nil {
			label = nil // label compilation failure is not fatal
		} else {
			label = l
		}
	}

	var fmla lg.Expr
	var err error
	if _, ok := n.Formula.(*ast.SchemaBody); ok {
		fmla, err = c.Thing(n.Formula)
	} else {
		fmla, err = c.SortifyWithInference(n.Formula)
	}
	if err != nil {
		return nil, err
	}

	// Return the compiled labeled formula as a pair wrapped in a Definition
	// (the Definition type serves as a general container).
	// The caller will typically extract label and formula separately.
	if label != nil {
		return il.NewDefinition(label, fmla), nil
	}
	return fmla, nil
}

// CompileLF compiles a LabeledFormula by cloning it with compiled children.
// Matches Python's LabeledFormula.cmpl (ivy_compiler.py:411-414):
//
//	self.clone([
//	    None if self.label is None else self.label.clone([sortify_with_inference(x) for x in self.label.args]),
//	    self.formula.compile() if isinstance(self.formula, SchemaBody) else sortify_with_inference(self.formula)
//	])
//
// This calls lf.Clone() which emits the PRESERVE/FRESH trace and properly
// manages the LfCounter, matching Python's compile() → clone() path.
func (c *Compiler) CompileLF(lf *ast.LabeledFormula) (*ast.LabeledFormula, error) {
	// Compile label: Python: None if self.label is None else self.label.clone([sortify_with_inference(x) for x in self.label.args])
	var compiledLabel ast.Node
	if lf.Label != nil {
		var newArgs []ast.Node
		for _, arg := range lf.Label.Args() {
			compiled, err := c.SortifyWithInference(arg)
			if err != nil {
				return nil, err
			}
			newArgs = append(newArgs, compiled)
		}
		compiledLabel = lf.Label.Clone(newArgs)
	}

	// Compile formula: Python: self.formula.compile() if isinstance(self.formula, SchemaBody) else sortify_with_inference(self.formula)
	var compiledFormula ast.Node
	if sb, ok := lf.Formula.(*ast.SchemaBody); ok {
		// Python: self.formula.compile() → compile_schema_body (direct, no thing() wrapper)
		// SchemaBody.compile = compile_schema_body is a direct assignment in Python,
		// so there are NO Thing ENTER/CompileNode ENTER traces for the SchemaBody.
		compiled, err := c.CompileSchemaBody(sb)
		if err != nil {
			return nil, err
		}
		compiledFormula = compiled // *ast.SchemaBody is ast.Node
	} else {
		f, err := c.SortifyWithInference(lf.Formula)
		if err != nil {
			return nil, err
		}
		compiledFormula = f
	}

	// Clone preserving ID and metadata — triggers PRESERVE trace
	// Matches Python: self.clone([compiledLabel, compiledFormula])
	result := lf.Clone([]ast.Node{compiledLabel, compiledFormula}).(*ast.LabeledFormula)
	return result, nil
}

// compileNativeExpr compiles a NativeExpr: compile args, preserve structure.
func (c *Compiler) compileNativeExpr(n *ast.NativeExpr) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileNativeExpr ENTER")
	// NativeExpr compilation: compile children, result has TopSort.
	args := n.Args()
	compiled := make([]lg.Expr, len(args))
	for i, a := range args {
		r, err := c.Thing(a)
		if err != nil {
			return nil, err
		}
		compiled[i] = r
	}
	// B3-R3: Preserve all children in a NativeExpr with TopSort, matching Python:
	// res = self.clone([a.compile() for a in self.args]); res.sort = TopS
	return &lg.NativeExpr{CompiledChildren: compiled}, nil
}

// compileTrigger compiles a trigger hint.
func (c *Compiler) compileTrigger(n *ast.Trigger) (lg.Expr, error) {
	args := n.Args()
	if len(args) < 2 {
		return c.Thing(args[0])
	}
	// Compile pattern as-is, but sort-infer the trigger terms
	pattern, err := c.Thing(args[0])
	if err != nil {
		return nil, err
	}
	terms := make([]lg.Expr, len(args)-1)
	for i, a := range args[1:] {
		t, err := c.SortifyWithInference(a)
		if err != nil {
			return nil, err
		}
		terms[i] = t
	}
	// Package as And(pattern, terms...) for now
	all := append([]lg.Expr{pattern}, terms...)
	return &lg.And{Terms: all}, nil
}

// --- Quantifier compilation ---

// CompileQuantifier compiles Forall/Exists AST nodes.
func (c *Compiler) CompileQuantifier(node ast.Node) (lg.Expr, error) {
	xtracer.Trace("compiler.compile_quantifier ENTER")
	var bounds []ast.Node
	var body ast.Node
	var isForall bool

	switch n := node.(type) {
	case *ast.Forall:
		bounds = n.Bounds
		body = n.Body
		isForall = true
	case *ast.Exists:
		bounds = n.Bounds
		body = n.Body
		isForall = false
	default:
		return nil, fmt.Errorf("CompileQuantifier: unexpected type %T", node)
	}

	// Compile bound variables
	vars := make([]*lg.Variable, len(bounds))
	for i, b := range bounds {
		v, ok := b.(*ast.Variable)
		if !ok {
			return nil, lg.NewIvyError(node, fmt.Sprintf(
				"quantifier bound %d is not a variable: %T", i, b))
		}
		sort, err := c.variableSort(v)
		if err != nil {
			return nil, err
		}
		lv, err := lg.NewVariable(v.Rep, sort)
		if err != nil {
			return nil, err
		}
		vars[i] = lv
	}

	// Set up variable context for the body
	savedVarCtx := c.VarCtx
	newMap := make(map[string]lg.Sort, len(savedVarCtx.Map)+len(vars))
	for k, v := range savedVarCtx.Map {
		newMap[k] = v
	}
	for _, v := range vars {
		newMap[v.Name] = v.VSort
	}
	c.VarCtx = &VariableContext{Map: newMap}

	compiled, err := c.Thing(body)

	c.VarCtx = savedVarCtx

	if err != nil {
		return nil, err
	}

	if isForall {
		return il.ForAll(vars, compiled), nil
	}
	return il.Exists(vars, compiled), nil
}

// --- Sort inference ---

// SortInfer resolves TopSort variables in a compiled logic node.
// Matches Python ivy_logic.py sort_infer:
//
//	res = concretize_sorts(term, sort)
//	check_concretely_sorted(res)
func (c *Compiler) SortInfer(node lg.Expr) (lg.Expr, error) {
	res, err := typeinfer.ConcretizeSorts(node, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SortifyWithInference compiles an AST node and applies sort inference.
// Python: def sortify_with_inference(ast):
//
//	with top_sort_as_default():
//	    res = ast.compile()
//	with ASTContext(ast):
//	    res = sort_infer(res)
//	return res
func (c *Compiler) SortifyWithInference(astNode ast.Node) (lg.Expr, error) {
	xtracer.Trace("compiler.sortify_with_inference ENTER")
	if xtracer.Enabled {
		c.SigCheck("SortifyWithInference")
	}
	// B3-R1: wrap compilation in top_sort_as_default, matching Python
	tsDefault := il.TopSortAsDefault(c.Sig)
	tsDefault.Enter()
	res, err := c.Thing(astNode)
	tsDefault.Exit()
	if err != nil {
		return nil, err
	}
	// B4-R3: Python wraps sort_infer in ASTContext(ast) to attach location on error
	// with ASTContext(ast): res = sort_infer(res)
	result, err := c.SortInfer(res)
	if err != nil {
		loc := astNode.GetLineno()
		return nil, fmt.Errorf("at %v: %w", loc, err)
	}
	return result, nil
}

// CompileConst compiles a constant declaration, adding it to the signature.
func (c *Compiler) CompileConst(v ast.Node, sig *il.Sig) (*lg.Symbol, error) {
	xtracer.Trace("compiler.CompileConst ENTER")
	var name string
	var sortArgs []ast.Node
	var sortNode ast.Node

	switch n := v.(type) {
	case *ast.Atom:
		name = n.Rep
		sortArgs = n.Terms
		sortNode = n.ASort
	case *ast.App:
		if sym, ok := n.Rep.(*ast.Symbol); ok {
			name = sym.Rep
		} else {
			name = fmt.Sprint(n.Rep)
		}
		sortArgs = n.Terms
		sortNode = n.ASort
	case *ast.Symbol:
		name = n.Rep
		sortNode = n.Sort
	default:
		return nil, lg.NewIvyError(v, fmt.Sprintf("cannot compile const from %T", v))
	}

	// Determine the range sort
	var rng lg.Sort
	if sortNode != nil {
		sortName := extractSortName(sortNode)
		if sortName != "" {
			var err error
			rng, err = c.CmplSort(sortName)
			if err != nil {
				return nil, err
			}
		}
	}
	if rng == nil {
		// Default sort
		if sig.DefaultSort != nil {
			rng = sig.DefaultSort
		} else {
			rng = lg.TopS
		}
	}

	// Get the function sort from arguments
	sort := c.getFunctionSort(sig, sortArgs, rng)

	return c.AddSymbol(name, sort, sig)
}

// getFunctionSort constructs a FunctionSort from arg sorts and range.
func (c *Compiler) getFunctionSort(sig *il.Sig, args []ast.Node, rng lg.Sort) lg.Sort {
	xtracer.Trace("compiler.GetFunctionSort ENTER")
	if len(args) == 0 {
		return rng
	}
	sorts := make([]lg.Sort, 0, len(args)+1)
	for _, a := range args {
		// Python: arg.compile().get_sort() — compiles every arg uniformly
		compiled, err := c.Thing(a)
		if err != nil {
			sorts = append(sorts, lg.TopS)
		} else {
			sorts = append(sorts, compiled.NodeSort())
		}
	}
	sorts = append(sorts, rng)
	fs, err := lg.NewFunctionSort(sorts...)
	if err != nil {
		return rng
	}
	return fs
}

// AddSymbol adds a symbol with the given name and sort to the signature.
func (c *Compiler) AddSymbol(name string, sort lg.Sort, sig *il.Sig) (*lg.Symbol, error) {
	sym, err := sig.AddSymbol(name, sort)
	if err != nil {
		return nil, err
	}
	// Also add to hierarchy
	c.Module.AddToHierarchy(name)
	return sym, nil
}

// findSymbol looks up a symbol in the signature.
func (c *Compiler) findSymbol(name string) (*lg.Symbol, error) {
	// Try polymorphic first
	if sym, ok := il.FindPolymorphicSymbol(name); ok {
		return sym, nil
	}
	return c.Sig.FindSymbol(name, false)
}

// CompileDefn compiles a definition (lhs = rhs) AST node.
// Corresponds to Python's compile_defn.
func (c *Compiler) CompileDefn(df *ast.Definition) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileDefn ENTER\n  (CompileDefn -> compileDefnImpl isSchema=false) [go side; at %v]", fileLine(1))
	return c.compileDefnImpl(df, false)
}

// CompileDefnSchema compiles a definition schema (DefinitionSchema variant).
func (c *Compiler) CompileDefnSchema(df *ast.DefinitionSchema) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileDefnSchema ENTER\n  (CompileDefnSchema -> compileDefnImpl isSchema=true) [go side; at %v]", fileLine(1))
	return c.compileDefnImpl(&df.Definition, true)
}

func (c *Compiler) compileDefnImpl(df *ast.Definition, isSchema bool) (lg.Expr, error) {
	xtracer.Trace("compiler.CompileDefnImpl ENTER")
	if xtracer.Enabled {
		c.SigCheck("CompileDefnImpl.entry")
	}
	lhs := df.Lhs
	var lhsAtom *ast.Atom
	if a, ok := lhs.(*ast.Atom); ok {
		lhsAtom = a
	}

	sigCopy := c.Sig.Copy()
	savedSig := c.Sig
	c.Sig = sigCopy
	if xtracer.Enabled {
		c.SigCheck("CompileDefnImpl.afterCopy")
	}

	// Compile any constant parameters in the LHS and collect variable sort substitutions
	subst := make(map[string]string)
	if lhsAtom != nil {
		for _, p := range lhsAtom.Terms {
			if v, isVar := p.(*ast.Variable); isVar {
				if v.VSort != "" { // wrong: && v.VSort != "S" {
					subst[v.Rep] = v.VSort
				}
			} else {
				c.CompileConst(p, sigCopy)
			}
		}
	}

	// Apply variable sort substitutions to RHS if needed
	rhs := df.Rhs
	xtracer.Trace("compiler.CompileDefnImpl rhs type=%s\n  isSchema=%v lhs=%v", typeName(rhs), isSchema, df.Lhs)
	if len(subst) > 0 {
		rhs = ast.SetVariableSorts(rhs, subst)
		xtracer.Trace("compiler.CompileDefnImpl rhs after subst type=%s\n  subst=%v", typeName(rhs), subst)
	}

	// Handle SomeExpr on RHS
	// Python: if isinstance(df.args[1], ivy_ast.SomeExpr):
	if someExpr, ok := rhs.(*ast.SomeExpr); ok {
		xtracer.Trace("compiler.CompileDefnImpl SomeExpr branch\n  isSchema=%v", isSchema)
		// Build: forall(params, lhs = ite(fmla, ifval, elseval))
		ifval := someExpr.IfValue
		if ifval == nil {
			ifval = someExpr.Param
		}
		elseval := someExpr.ElseVal
		if elseval == nil {
			elseval = ifval
		}

		cfg := c.Module.Cfg.AstCfg
		iteNode := &ast.Ite{Cond: someExpr.Fmla, Then: ifval, Else: elseval}
		eqNode := cfg.NewAtom("=", df.Lhs, iteNode)
		forallNode := &ast.Forall{Bounds: []ast.Node{someExpr.Param}, Body: eqNode}

		fmla, err := c.SortifyWithInference(forallNode)
		c.Sig = savedSig
		if err != nil {
			return nil, err
		}

		// Extract from the compiled forall: variables[0], body.args[1].args[0..2]
		var defLhs, someArgs lg.Expr
		if forall, ok := fmla.(*lg.ForAll); ok && len(forall.Variables) > 0 {
			param := forall.Variables[0]
			if eq, ok := forall.Body.(*lg.Eq); ok {
				defLhs = eq.T1
				if ite, ok := eq.T2.(*lg.Ite); ok {
					someNode := &il.Some{
						Params: []lg.Expr{param},
						Fmla:   ite.Cond,
					}
					if someExpr.IfValue != nil {
						someNode.IfVal = ite.Then
					}
					if someExpr.ElseVal != nil {
						someNode.ElseVal = ite.Else
					}
					someArgs = someNode
				} else {
					someArgs = eq.T2
				}
			}
		}
		if defLhs == nil {
			defLhs = fmla
			someArgs = fmla
		}

		result := il.NewDefinition(defLhs, someArgs)
		if isSchema {
			return il.NewDefinitionSchema(defLhs, someArgs), nil
		}
		return result, nil
	}

	// Standard definition: compile as equality lhs = rhs, then apply sort inference
	cfg := c.Module.Cfg.AstCfg
	xtracer.Trace("compiler.CompileDefnImpl standard branch rhs type=%s\n  isSchema=%v lhs=%v rhs=%v", typeName(rhs), isSchema, df.Lhs, rhs)
	eqAtom := cfg.NewAtom("=", df.Lhs, rhs)
	xtracer.Trace("compiler.CompileDefnImpl eqAtom nTerms=%d\n  eqAtom=%v", len(eqAtom.Terms), eqAtom)
	for ti, tt := range eqAtom.Terms {
		if tt == nil {
			xtracer.Trace("compiler.CompileDefnImpl eqAtom.Terms[%d] = nil", ti)
		} else {
			xtracer.Trace("compiler.CompileDefnImpl eqAtom.Terms[%d] type=%s\n  val=%v", ti, typeName(tt), tt)
		}
	}
	eqAtom.SetLineno(df.GetLineno())
	compiled, err := c.SortifyWithInference(eqAtom)
	c.Sig = savedSig

	if err != nil {
		return nil, err
	}

	// Extract lhs and rhs from the compiled equality
	if eq, ok := compiled.(*lg.Eq); ok {
		if isSchema {
			return il.NewDefinitionSchema(eq.T1, eq.T2), nil
		}
		return il.NewDefinition(eq.T1, eq.T2), nil
	}
	// If sort inference returned the equality as-is, wrap in Definition
	if isSchema {
		return il.NewDefinitionSchema(compiled, compiled), nil
	}
	return il.NewDefinition(compiled, compiled), nil
}

// --- Tactic/proof compilation ---
//
// Corresponds to Python ivy_compiler.py:912-1008.
// Most tactic compile methods are identity (return self unchanged).
// A few compile sub-parts: IfTactic (condition), PropertyTactic (prop, name, proof),
// ProofTactic (label, proof), ComposeTactics (each sub-tactic).

// CompileTactic compiles a tactic/proof AST node. Unlike CompileNode which
// produces lg.Expr, this returns an ast.Node since tactics remain as AST
// nodes for the proof checker to process later.
func (c *Compiler) CompileTactic(node ast.Node) (ast.Node, error) {
	xtracer.Trace("compiler.CompileTactic ENTER")
	if node == nil {
		return nil, nil
	}
	switch n := node.(type) {
	case *ast.SchemaInstantiation:
		// Python: compile_schema_instantiation returns self (no-op in current Python)
		return n, nil

	case *ast.AssumeTactic:
		// No compilation needed
		return n, nil

	case *ast.AssumeGlobalTactic:
		// No compilation needed
		return n, nil

	case *ast.LetTactic:
		// Python: compile_let_tactic returns self (no-op in current Python)
		return n, nil

	case *ast.WitnessTactic:
		// Python: compile_witness_tactic returns self
		return n, nil

	case *ast.UnfoldTactic:
		// Python: compile_unfold_tactic returns self
		return n, nil

	case *ast.ForgetTactic:
		// Python: compile_forget_tactic returns self
		return n, nil

	case *ast.FunctionTactic:
		// Python: compile_function_tactic returns self
		return n, nil

	case *ast.ShowGoalsTactic:
		return n, nil

	case *ast.DeferGoalTactic:
		return n, nil

	case *ast.NullTactic:
		return n, nil

	case *ast.SpoilTactic:
		return n, nil

	case *ast.Tactic:
		return n, nil

	case *ast.IfTactic:
		// Python: compile_if_tactic compiles condition with sort inference,
		// then recursively compiles both branches.
		cond, err := c.SortifyWithInference(n.Cond)
		if err != nil {
			// If sort inference fails, fall back to compiling normally
			cond, err = c.Thing(n.Cond)
			if err != nil {
				return n, nil // return unchanged on error
			}
		}
		thenBranch, err := c.CompileTactic(n.Then)
		if err != nil {
			return n, nil
		}
		elseBranch, err := c.CompileTactic(n.Else)
		if err != nil {
			return n, nil
		}
		// Wrap the compiled condition as an AST node for Clone
		condWrapper := &ast.CompiledNode{Node: cond}
		return n.Clone([]ast.Node{condWrapper, thenBranch, elseBranch}), nil

	case *ast.PropertyTactic:
		// Python: compile_property_tactic
		// prop = self.args[0] (not compiled)
		// name = self.args[1]: if not NoneAST, compile name args with UnsortedContext
		// proof = self.args[2].compile()
		prop := n.Prop
		name := n.PName
		if _, isNone := name.(*ast.NoneAST); !isNone && name != nil {
			// Compile name's args with unsorted context
			if atom, ok := name.(*ast.Atom); ok {
				compiledTerms := make([]ast.Node, len(atom.Terms))
				for i, arg := range atom.Terms {
					compiled, err := c.Thing(arg)
					if err != nil {
						compiledTerms[i] = arg
						continue
					}
					compiledTerms[i] = &ast.CompiledNode{Node: compiled}
				}
				name = &ast.Atom{Base: atom.Base, Rep: atom.Rep, Terms: compiledTerms, ASort: atom.ASort}
			}
		}
		proof, err := c.CompileTactic(n.Proof)
		if err != nil {
			proof = n.Proof
		}
		return &ast.PropertyTactic{Base: n.Base, Prop: prop, PName: name, Proof: proof}, nil

	case *ast.TacticTactic:
		// Python: compile_tactic_tactic returns self.clone(self.args)
		return n.Clone(n.Args()), nil

	case *ast.ProofTactic:
		// Python: compile_proof_tactic compiles label and proof
		proof, err := c.CompileTactic(n.Proof)
		if err != nil {
			proof = n.Proof
		}
		return &ast.ProofTactic{Base: n.Base, TLabel: n.TLabel, Proof: proof}, nil

	case *ast.ComposeTactics:
		// Recursively compile each sub-tactic
		compiledTactics := make([]ast.Node, len(n.Tactics))
		for i, t := range n.Tactics {
			ct, err := c.CompileTactic(t)
			if err != nil {
				compiledTactics[i] = t
				continue
			}
			compiledTactics[i] = ct
		}
		return &ast.ComposeTactics{Base: n.Base, Tactics: compiledTactics}, nil

	default:
		// For unknown tactic types, return unchanged
		return n, nil
	}
}

// extractSortName extracts a string sort name from an AST sort node.
func extractSortName(n ast.Node) string {
	if n == nil {
		return ""
	}
	switch s := n.(type) {
	case *ast.Symbol:
		return s.Rep
	case *ast.Atom:
		return s.Rep
	default:
		r := fmt.Sprint(n)
		if r == "" || r == "<nil>" {
			return ""
		}
		return strings.TrimSpace(r)
	}
}
