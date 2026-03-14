// Package compiler transforms AST nodes (ast/ package) into logic IR nodes
// (logic/ and ivylogic/ packages). It replaces the Python monkey-patching
// pattern in ivy_compiler.py with a Compiler struct using a visitor/switch
// dispatch.
//
// This corresponds to Python's ivy_compiler.py.
package compiler

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// ActionInfo holds metadata about a declared action: its formal parameters,
// formal return values, and key-argument position.
type ActionInfo struct {
	Params  []*lg.Const
	Returns []*lg.Const
	KeyPos  int
}

// ReturnContext tracks the return values for the current expression compilation.
type ReturnContext struct {
	Values []lg.Node
	Lineno *ast.Location
}

// ExprContext tracks intermediate code and local symbols generated while
// compiling an expression (e.g., inline action calls in an rhs).
type ExprContext struct {
	Code      []lg.Node // accumulated action nodes (wrapped)
	LocalSyms []*lg.Const
	Lineno    *ast.Location
}

// Extract produces a single action from the accumulated code.
func (ec *ExprContext) Extract() lg.Node {
	// TODO: when actions package is better integrated, produce
	// LocalAction / Sequence as appropriate.
	if len(ec.Code) == 1 {
		return ec.Code[0]
	}
	// Placeholder: just return the last code element.
	if len(ec.Code) == 0 {
		return nil
	}
	return ec.Code[len(ec.Code)-1]
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
func (c *Compiler) CompileNode(node ast.Node) (lg.Node, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot compile nil node")
	}
	switch n := node.(type) {
	// --- Formula operators ---
	case *ast.And:
		return c.compileAnd(n)
	case *ast.Or:
		return c.compileOr(n)
	case *ast.Not:
		return c.compileNot(n)
	case *ast.Implies:
		return c.compileImplies(n)
	case *ast.Iff:
		return c.compileIff(n)
	case *ast.Ite:
		return c.compileIte(n)
	case *ast.Definition:
		return c.compileDefinition(n)
	case *ast.Globally:
		return c.compileGlobally(n)
	case *ast.Eventually:
		return c.compileEventually(n)
	case *ast.WhenOperator:
		return c.compileWhenOperator(n)

	// --- Quantifiers ---
	case *ast.Forall:
		return c.CompileQuantifier(n)
	case *ast.Exists:
		return c.CompileQuantifier(n)

	// --- Terms ---
	case *ast.Atom:
		return c.CompileApp(n, false)
	case *ast.App:
		return c.compileAppNode(n)
	case *ast.Variable:
		return c.CompileVariable(n)
	case *ast.Old:
		return c.compileOld(n)
	case *ast.MethodCall:
		return c.compileMethodCall(n)

	// --- Named binder ---
	case *ast.NamedBinder:
		return c.compileNamedBinder(n)

	// --- Labeled formula ---
	case *ast.LabeledFormula:
		return c.compileLabeledFormula(n)

	// --- NativeExpr ---
	case *ast.NativeExpr:
		return c.compileNativeExpr(n)

	// --- Trigger ---
	case *ast.Trigger:
		return c.compileTrigger(n)

	// --- Sort-inference root nodes ---
	// For nodes that have a sort_infer_root property in Python,
	// we compile children and do sort inference on the result.
	default:
		return c.compileGeneric(node)
	}
}

// compileGeneric is the fallback: compile each child and clone.
func (c *Compiler) compileGeneric(node ast.Node) (lg.Node, error) {
	args := node.Args()
	compiled := make([]lg.Node, len(args))
	for i, a := range args {
		r, err := c.CompileNode(a)
		if err != nil {
			return nil, fmt.Errorf("compiling arg %d of %T: %w", i, node, err)
		}
		compiled[i] = r
	}
	if len(compiled) == 0 {
		return lg.True, nil
	}
	// For unknown nodes, return the first compiled arg or wrap as And.
	if len(compiled) == 1 {
		return compiled[0], nil
	}
	return &lg.And{Terms: compiled}, nil
}

// --- Formula compilation ---

// compileArgs compiles all children of an AST node with no return context.
func (c *Compiler) compileArgs(node ast.Node) ([]lg.Node, error) {
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	defer func() { c.ReturnCtx = saved }()

	args := node.Args()
	result := make([]lg.Node, len(args))
	for i, a := range args {
		r, err := c.CompileNode(a)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}

func (c *Compiler) compileAnd(n *ast.And) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil {
		return nil, err
	}
	return &lg.And{Terms: args}, nil
}

func (c *Compiler) compileOr(n *ast.Or) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil {
		return nil, err
	}
	return &lg.Or{Terms: args}, nil
}

func (c *Compiler) compileNot(n *ast.Not) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	return &lg.Not{Body: args[0]}, nil
}

func (c *Compiler) compileImplies(n *ast.Implies) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &lg.Implies{T1: args[0], T2: args[1]}, nil
}

func (c *Compiler) compileIff(n *ast.Iff) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &lg.Iff{T1: args[0], T2: args[1]}, nil
}

func (c *Compiler) compileIte(n *ast.Ite) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 3 {
		return nil, err
	}
	return &lg.Ite{ISort: args[1].NodeSort(), Cond: args[0], Then: args[1], Else: args[2]}, nil
}

func (c *Compiler) compileDefinition(n *ast.Definition) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return il.NewDefinition(args[0], args[1]), nil
}

func (c *Compiler) compileGlobally(n *ast.Globally) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	empty := ""
	return &lg.Globally{Environ: &empty, Body: args[0]}, nil
}

func (c *Compiler) compileEventually(n *ast.Eventually) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	empty := ""
	return &lg.Eventually{Environ: &empty, Body: args[0]}, nil
}

func (c *Compiler) compileWhenOperator(n *ast.WhenOperator) (lg.Node, error) {
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &lg.WhenOperator{WSort: args[0].NodeSort(), Name: n.Name, T1: args[0], T2: args[1]}, nil
}

// --- Term compilation ---

// CompileApp compiles function application (Atom). This corresponds to
// Python's compile_app.
func (c *Compiler) CompileApp(n *ast.Atom, old bool) (lg.Node, error) {
	rep := ResolveAlias(n.Rep, c.Module)

	// Handle boolean literals
	if rep == "true" || rep == "false" {
		if len(n.Terms) > 0 {
			return nil, &lg.IvyError{Msg: fmt.Sprintf("%s is not a function", rep)}
		}
		if rep == "true" {
			return &lg.And{}, nil // empty And = true
		}
		return &lg.Or{}, nil // empty Or = false
	}

	// Compile arguments with no return context
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	args := make([]lg.Node, len(n.Terms))
	for i, a := range n.Terms {
		r, err := c.CompileNode(a)
		if err != nil {
			c.ReturnCtx = saved
			return nil, err
		}
		args[i] = r
	}
	c.ReturnCtx = saved

	// Handle inline action calls in rhs of assignment
	if c.ExprCtx != nil && c.TopCtx != nil {
		if _, ok := c.TopCtx.Actions[rep]; ok {
			return c.CompileInlineCall(n, args)
		}
	}

	// Try to find the symbol
	if rep == "=" {
		// Equality
		if len(args) != 2 {
			return nil, &lg.IvyError{Msg: "equality requires exactly 2 arguments"}
		}
		return &lg.Eq{T1: args[0], T2: args[1]}, nil
	}

	// Look up polymorphic symbol first
	sym, found := il.FindPolymorphicSymbol(rep)
	if !found {
		// Look up in signature
		entry, ok := c.Sig.Symbols[rep]
		if ok {
			sym = lg.NewConst(rep, entry.Sort)
		}
	}

	if sym != nil {
		// Handle numerals with explicit sort annotation
		if il.IsNumeral(sym) {
			if n.ASort != nil {
				sortName := extractSortName(n.ASort)
				if sortName != "S" {
					s, err := c.Sig.FindSort(sortName, false)
					if err == nil {
						sym = lg.NewConst(sym.Name, s)
					}
				}
			}
		}

		if old {
			sym = lg.NewConst("old_"+sym.Name, sym.CSort)
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
func (c *Compiler) compileAppNode(n *ast.App) (lg.Node, error) {
	// If the rep is a Symbol, treat it like an Atom.
	if sym, ok := n.Rep.(*ast.Symbol); ok {
		atom := ast.NewAtom(sym.Rep, n.Terms...)
		atom.SetLineno(n.GetLineno())
		atom.ASort = n.ASort
		return c.CompileApp(atom, false)
	}
	// If the rep is a NamedBinder, compile it and apply.
	repNode, err := c.CompileNode(n.Rep)
	if err != nil {
		return nil, err
	}

	saved := c.ReturnCtx
	c.ReturnCtx = nil
	args := make([]lg.Node, len(n.Terms))
	for i, a := range n.Terms {
		r, err := c.CompileNode(a)
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

// CompileVariable compiles a Variable AST node to a logic.Var.
func (c *Compiler) CompileVariable(n *ast.Variable) (lg.Node, error) {
	sort, err := c.variableSort(n)
	if err != nil {
		return nil, err
	}
	// If sort is top, check variable context
	if _, isTop := sort.(*lg.TopSort); isTop {
		if s, ok := c.VarCtx.Map[n.Rep]; ok {
			sort = s
		}
	}
	v, err := lg.NewVar(n.Rep, sort)
	if err != nil {
		return nil, &lg.IvyError{Msg: fmt.Sprintf("bad variable: %s", err)}
	}
	return v, nil
}

// variableSort resolves the sort of a variable AST node.
func (c *Compiler) variableSort(v *ast.Variable) (lg.Sort, error) {
	if v.VSort == nil {
		return lg.TopS, nil
	}
	sortName := extractSortName(v.VSort)
	if sortName == "" {
		return lg.TopS, nil
	}
	return c.CmplSort(sortName)
}

// CmplSort resolves a sort name to a logic.Sort, applying alias resolution.
func (c *Compiler) CmplSort(name string) (lg.Sort, error) {
	resolved := ResolveAlias(name, c.Module)
	return c.Sig.FindSort(resolved, false)
}

// compileOld compiles the Old operator: compile the inner term with old=true.
func (c *Compiler) compileOld(n *ast.Old) (lg.Node, error) {
	// The inner term should be an Atom or App
	if atom, ok := n.Term.(*ast.Atom); ok {
		return c.CompileApp(atom, true)
	}
	if app, ok := n.Term.(*ast.App); ok {
		if sym, ok := app.Rep.(*ast.Symbol); ok {
			atom := ast.NewAtom(sym.Rep, app.Terms...)
			atom.SetLineno(n.GetLineno())
			atom.ASort = app.ASort
			return c.CompileApp(atom, true)
		}
	}
	// Fallback: compile the inner term normally
	return c.CompileNode(n.Term)
}

// compileMethodCall compiles obj.method() style calls.
func (c *Compiler) compileMethodCall(n *ast.MethodCall) (lg.Node, error) {
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	base, err := c.CompileNode(n.Obj)
	if err != nil {
		c.ReturnCtx = saved
		return nil, err
	}

	// Method is typically an Atom or App
	var childName string
	var methodArgs []lg.Node
	switch m := n.Method.(type) {
	case *ast.Atom:
		childName = m.Rep
		methodArgs = make([]lg.Node, len(m.Terms))
		for i, a := range m.Terms {
			r, err := c.CompileNode(a)
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
		return nil, &lg.IvyError{Msg: fmt.Sprintf("unsupported method node type: %T", n.Method)}
	}
	c.ReturnCtx = saved

	sort := base.NodeSort()
	if _, isTop := sort.(*lg.TopSort); isTop {
		return nil, &lg.IvyError{Msg: fmt.Sprintf(
			"cannot apply method notation to %s because its type is not inferred", base)}
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
				return nil, &lg.IvyError{Msg: fmt.Sprintf(
					"call to action %s not allowed outside an action", destrName)}
			}
			allArgs := append([]lg.Node{base}, methodArgs...)
			atom := ast.NewAtom(destrName)
			atom.SetLineno(n.GetLineno())
			return c.CompileInlineCall(atom, allArgs)
		}
	}

	// Find the destructor symbol
	sym, err := c.findSymbol(destrName)
	if err != nil {
		return nil, err
	}
	allArgs := append([]lg.Node{base}, methodArgs...)
	if len(allArgs) == 0 {
		return sym, nil
	}
	return lg.NewApply(sym, allArgs...)
}

// compileNamedBinder compiles an AST NamedBinder.
func (c *Compiler) compileNamedBinder(n *ast.NamedBinder) (lg.Node, error) {
	vars := make([]*lg.Var, len(n.Bounds))
	for i, b := range n.Bounds {
		compiled, err := c.CompileNode(b)
		if err != nil {
			return nil, err
		}
		v, ok := compiled.(*lg.Var)
		if !ok {
			return nil, &lg.IvyError{Msg: fmt.Sprintf(
				"named binder bound %d is not a variable: %T", i, compiled)}
		}
		vars[i] = v
	}
	body, err := c.CompileNode(n.Body)
	if err != nil {
		return nil, err
	}
	return lg.NewNamedBinder(n.Name, vars, nil, body)
}

// compileLabeledFormula compiles a LabeledFormula AST node.
func (c *Compiler) compileLabeledFormula(n *ast.LabeledFormula) (lg.Node, error) {
	var label lg.Node
	if n.Label != nil {
		l, err := c.SortifyWithInference(n.Label)
		if err != nil {
			label = nil // label compilation failure is not fatal
		} else {
			label = l
		}
	}

	var fmla lg.Node
	var err error
	if _, ok := n.Formula.(*ast.SchemaBody); ok {
		fmla, err = c.CompileNode(n.Formula)
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

// compileNativeExpr compiles a NativeExpr: compile args, preserve structure.
func (c *Compiler) compileNativeExpr(n *ast.NativeExpr) (lg.Node, error) {
	// NativeExpr compilation: compile children, result has TopSort.
	args := n.Args()
	compiled := make([]lg.Node, len(args))
	for i, a := range args {
		r, err := c.CompileNode(a)
		if err != nil {
			return nil, err
		}
		compiled[i] = r
	}
	// Return the first compiled result (or a const representing the native expr).
	if len(compiled) == 0 {
		return lg.NewConst("native", lg.TopS), nil
	}
	return compiled[0], nil
}

// compileTrigger compiles a trigger hint.
func (c *Compiler) compileTrigger(n *ast.Trigger) (lg.Node, error) {
	args := n.Args()
	if len(args) < 2 {
		return c.CompileNode(args[0])
	}
	// Compile pattern as-is, but sort-infer the trigger terms
	pattern, err := c.CompileNode(args[0])
	if err != nil {
		return nil, err
	}
	terms := make([]lg.Node, len(args)-1)
	for i, a := range args[1:] {
		t, err := c.SortifyWithInference(a)
		if err != nil {
			return nil, err
		}
		terms[i] = t
	}
	// Package as And(pattern, terms...) for now
	all := append([]lg.Node{pattern}, terms...)
	return &lg.And{Terms: all}, nil
}

// --- Quantifier compilation ---

// CompileQuantifier compiles Forall/Exists AST nodes.
func (c *Compiler) CompileQuantifier(node ast.Node) (lg.Node, error) {
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
	vars := make([]*lg.Var, len(bounds))
	for i, b := range bounds {
		v, ok := b.(*ast.Variable)
		if !ok {
			return nil, &lg.IvyError{Msg: fmt.Sprintf(
				"quantifier bound %d is not a variable: %T", i, b)}
		}
		sort, err := c.variableSort(v)
		if err != nil {
			return nil, err
		}
		lv, err := lg.NewVar(v.Rep, sort)
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

	compiled, err := c.CompileNode(body)

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

// SortInfer performs sort inference on a compiled logic node.
// TODO: integrate with typeinfer.InferSorts when ready.
func (c *Compiler) SortInfer(node lg.Node) (lg.Node, error) {
	// Stub: return the node as-is for now.
	// Full implementation will call typeinfer.InferSorts.
	return node, nil
}

// SortifyWithInference compiles an AST node and applies sort inference.
func (c *Compiler) SortifyWithInference(astNode ast.Node) (lg.Node, error) {
	// In Python: with top_sort_as_default(): res = ast.compile()
	// then: res = sort_infer(res)
	res, err := c.CompileNode(astNode)
	if err != nil {
		return nil, err
	}
	return c.SortInfer(res)
}

// CompileConst compiles a constant declaration, adding it to the signature.
func (c *Compiler) CompileConst(v ast.Node, sig *il.Sig) (*lg.Const, error) {
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
		return nil, &lg.IvyError{Msg: fmt.Sprintf("cannot compile const from %T", v)}
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
	if len(args) == 0 {
		return rng
	}
	sorts := make([]lg.Sort, 0, len(args)+1)
	for _, a := range args {
		v, ok := a.(*ast.Variable)
		if ok {
			s, err := c.variableSort(v)
			if err != nil {
				sorts = append(sorts, lg.TopS)
			} else {
				sorts = append(sorts, s)
			}
		} else {
			// Compile the arg and use its sort
			compiled, err := c.CompileNode(a)
			if err != nil {
				sorts = append(sorts, lg.TopS)
			} else {
				sorts = append(sorts, compiled.NodeSort())
			}
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
func (c *Compiler) AddSymbol(name string, sort lg.Sort, sig *il.Sig) (*lg.Const, error) {
	sym, err := sig.AddSymbol(name, sort)
	if err != nil {
		return nil, err
	}
	// Also add to hierarchy
	c.Module.AddToHierarchy(name)
	return sym, nil
}

// findSymbol looks up a symbol in the signature.
func (c *Compiler) findSymbol(name string) (*lg.Const, error) {
	// Try polymorphic first
	if sym, ok := il.FindPolymorphicSymbol(name); ok {
		return sym, nil
	}
	return c.Sig.FindSymbol(name, false)
}

// CompileDefn compiles a definition (lhs = rhs) AST node.
// Corresponds to Python's compile_defn.
func (c *Compiler) CompileDefn(df *ast.Definition) (lg.Node, error) {
	// Check if any args are non-variables (need fresh signature scope)
	lhs := df.Lhs
	var lhsAtom *ast.Atom
	if a, ok := lhs.(*ast.Atom); ok {
		lhsAtom = a
	}

	sigCopy := c.Sig.Copy()
	savedSig := c.Sig
	c.Sig = sigCopy

	// Compile any constant parameters in the LHS
	if lhsAtom != nil {
		for _, p := range lhsAtom.Terms {
			if _, isVar := p.(*ast.Variable); !isVar {
				c.CompileConst(p, sigCopy)
			}
		}
	}

	// Compile as an equality: lhs = rhs, then apply sort inference
	eqAtom := ast.NewAtom("=", df.Lhs, df.Rhs)
	eqAtom.SetLineno(df.GetLineno())
	compiled, err := c.SortifyWithInference(eqAtom)
	c.Sig = savedSig

	if err != nil {
		return nil, err
	}

	// Extract lhs and rhs from the compiled equality
	if eq, ok := compiled.(*lg.Eq); ok {
		return il.NewDefinition(eq.T1, eq.T2), nil
	}
	// If sort inference returned the equality as-is, wrap in Definition
	return il.NewDefinition(compiled, compiled), nil
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
