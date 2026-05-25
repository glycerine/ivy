// Compiler transforms parsed AST nodes into semantic logic/action nodes. It
// replaces Python's monkey-patched .compile methods with an explicit Compiler
// visitor/switch dispatch.
//
// This corresponds to Python's ivy_compiler.py.
package goivy

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// typeName is a shim to iu.TypeName, preserving the package-local helper
// name used by existing call sites. iu.TypeName maps Go's "Variable" to
// Python's "Var" so trace output aligns across languages.
func compilerTypeName(v interface{}) string {
	return TypeName(v)
}

// ActionInfo holds metadata about a declared action: its formal parameters,
// formal return values, and key-argument position.
// FormalAST/FormalRetAST hold the AST-level formals before compilation;
// Params/Returns hold the compiled logic-level symbols.
type ActionInfo struct {
	FormalAST    []Node   // AST-level formal parameters (pre-compilation)
	FormalRetAST []Node   // AST-level formal returns (pre-compilation)
	Params       []*Const // compiled formal parameters
	Returns      []*Const // compiled formal returns
	KeyPos       int      // index of first KeyArg in formals
}

// ReturnContext tracks the return values for the current expression compilation.
type ReturnContext struct {
	Values []Expr
	Lineno *Location
}

// ExprContext tracks intermediate code and local symbols generated while
// compiling an expression (e.g., inline action calls in an rhs).
type ExprContext struct {
	Code      []Expr // accumulated action nodes (wrapped)
	LocalSyms []*Const
	Lineno    *Location
	ActCfg    *ActionsConfig // for creating LocalAction in Extract()
}

// CompileInlineCode produces a single action from the accumulated code.
// Matches Python ivy_compiler.py compile_inline_call which wraps multiple
// statements in Sequence/LocalAction.
func (ec *ExprContext) CompileInlineCode() Expr {
	if len(ec.Code) == 0 {
		return nil
	}
	if len(ec.Code) == 1 {
		return ec.Code[0]
	}
	// Multiple code elements → wrap in a Sequence action.
	return NewSequence(ec.Code...)
}

// Extract produces a single action from the accumulated code and local symbols.
// Matches Python ExprContext.extract() (ivy_compiler.py lines 116-123):
//   - Sets lineno on all code items
//   - If 1 code item → return it directly
//   - If multiple items → wrap in LocalAction(*(self.local_syms + [Sequence(*self.code)]))
func (ec *ExprContext) Extract() Expr {
	xtracer.Trace("compiler.ExprContext.Extract ENTER")
	// Set lineno on all code items (Python lines 117-118)
	for _, c := range ec.Code {
		if act, _ := c.(ActionsAction); act != nil && ec.Lineno != nil {
			act.SetLineno(*ec.Lineno)
		}
	}
	// Single code item → return directly (Python lines 119-120)
	if len(ec.Code) == 1 {
		return ec.Code[0]
	}
	// Multiple items → wrap in LocalAction(local_syms..., Sequence(code...))
	args := make([]Expr, 0, len(ec.LocalSyms)+1)
	for _, s := range ec.LocalSyms {
		args = append(args, s)
	}
	args = append(args, NewSequence(ec.Code...))
	res := NewLocalActionOn(ec.ActCfg, "compiler.compile_expression", args...)
	if ec.Lineno != nil {
		res.SetLineno(*ec.Lineno)
	}
	return res
}

// TopContext holds the action metadata used during compilation.
type TopContext struct {
	Actions map[string]*ActionInfo
}

// VariableContext maps variable names to their sorts, used when compiling
// under quantifier scopes.
type VariableContext struct {
	Map map[string]Sort
}

// Compiler transforms AST nodes to logic IR. It carries the compilation
// state that Python stored in global variables.
type Compiler struct {
	// Sig is the current first-order signature.
	Sig *Sig

	// Module is the module being compiled.
	Module *Module

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
	// Pointer shared with Module so all Compiler instances use one chain.
	SigMerkle *MerkleState

	// ActCfg holds the per-session ActionsConfig, threaded from module.
	// Used for creating compiled actions (LocalAction, etc.) with proper
	// counter state. Matches Python's single global local_action_ctr.
	ActCfg *ActionsConfig
}

// SigCheck emits a Merkle-chained HASH trace of the current Sig + Module state.
// The golden test detects divergence via the HASH roots and uses DiffSexp
// on the canon= data to show exactly which sorts/symbols/declarations differ.
//
// IMPORTANT: Only call inside `if xtracer.Enabled { c.SigCheck(...) }` blocks.
// When -tags xtrace_off is set, Enabled is const false, so the compiler
// eliminates the entire block — Canon() string building and Merkle hashing
// have zero production cost.
func (c *Compiler) SigCheck(label string) {
	sigCanon := c.Sig.Canon()
	modCanon := Canonical("")
	if c.Module != nil {
		modCanon = c.Module.Canon()
	}
	combined := Canonical(string(sigCanon) + string(modCanon))
	leaf, root := c.SigMerkle.AddLeaf(combined)
	_ = root
	//xtracer.Trace("compiler.SigCheck@%s HASH leaf=%s root=%s canon=%s", label, leaf, root, string(combined))
	xtracer.Trace("compiler.SigCheck@%s HASH leaf=%s canon=%s", label, leaf, string(combined))
}

// New creates a new Compiler with the given signature and module.
func NewCompiler(sig *Sig, mod *Module) *Compiler {
	c := &Compiler{
		Sig:    sig,
		Module: mod,
		VarCtx: &VariableContext{Map: make(map[string]Sort)},
	}
	if sig == nil {
		if c.Module != nil && c.Module.Cfg != nil && c.Module.Cfg.IuCfg != nil {
			c.Sig = NewSigOn(c.Module.Cfg.IuCfg)
		} else {
			c.Sig = NewSig()
		}
	}
	if mod == nil {
		c.Module = New()
	}
	// Share the Module's Merkle chain so all compilers in a session
	// accumulate into one root, matching Python's module-level sig_merkle.
	if c.Module.SigMerkle != nil {
		c.SigMerkle = c.Module.SigMerkle
	} else {
		c.SigMerkle = &MerkleState{}
	}
	// Ensure ActCfg is always set. IvyCompile seeds mod.Cfg.ActCfg from the
	// AstConfig; compilers created later from the same module reuse it so
	// runtime macro compilation shares Python's action-counter state.
	if c.ActCfg == nil {
		if c.Module != nil && c.Module.Cfg != nil && c.Module.Cfg.ActCfg != nil {
			c.ActCfg = c.Module.Cfg.ActCfg
		} else {
			c.ActCfg = NewActionsConfig()
			if c.Module != nil && c.Module.Cfg != nil && c.Module.Cfg.AstCfg != nil {
				c.ActCfg.IuCfg = c.Module.Cfg.AstCfg.IuCfg
			}
		}
	}
	return c
}

// NewFromModule creates a Compiler using the module's own signature.
func NewFromModule(mod *Module) *Compiler {
	return NewCompiler(mod.Sig, mod)
}

// --- Main dispatch ---

// Cmpl compiles an AST node via the type-specific handler, matching
// Python's cmpl() dispatch. Unlike CompileNode, it does NOT emit the
// "CompileNode ENTER" trace — only the type-specific "CompileNode return case=X"
// trace fires. Use Cmpl where Python calls a.cmpl() (e.g., callee args in
// compile_call's "in actions" path).
func (c *Compiler) Cmpl(node Node) (Expr, error) {
	return c.compileNodeCore(node, false)
}

// CompileNode is the main visitor dispatch. It compiles any AST node to
// the corresponding logic IR node.
//
// CompileNode compiles an AST node, emitting "CompileNode ENTER" before
// dispatching to the type-specific handler. Matches Python's thing() →
// cmpl() path where thing() adds the ENTER trace.
func (c *Compiler) CompileNode(node Node) (Expr, error) {
	return c.compileNodeCore(node, true)
}

func (c *Compiler) compileNodeCore(node Node, emitEnter bool) (Expr, error) {
	if node == nil {
		xtracer.Trace("compiler.CompileNode return case=nil")
		return nil, fmt.Errorf("cannot compile nil node")
	}
	if emitEnter {
		xtracer.Trace(fmt.Sprintf("compiler.CompileNode ENTER type=%s", compilerTypeName(node)))
	}
	switch n := node.(type) {
	// --- Formula operators ---
	case *And:
		xtracer.Trace("compiler.CompileNode return case=And")
		return c.compileAnd(n)
	case *Or:
		xtracer.Trace("compiler.CompileNode return case=Or")
		return c.compileOr(n)
	case *Not:
		xtracer.Trace("compiler.CompileNode return case=Not")
		return c.compileNot(n)
	case *Implies:
		xtracer.Trace("compiler.CompileNode return case=Implies")
		return c.compileImplies(n)
	case *Iff:
		xtracer.Trace("compiler.CompileNode return case=Iff")
		return c.compileIff(n)
	case *Ite:
		xtracer.Trace("compiler.CompileNode return case=Ite")
		return c.compileIte(n)
	case *Definition:
		xtracer.Trace("compiler.CompileNode return case=Definition")
		return c.compileDefinition(n)
	case *Globally:
		xtracer.Trace("compiler.CompileNode return case=Globally")
		return c.compileGlobally(n)
	case *Eventually:
		xtracer.Trace("compiler.CompileNode return case=Eventually")
		return c.compileEventually(n)
	case *WhenOperator:
		xtracer.Trace("compiler.CompileNode return case=WhenOperator")
		return c.compileWhenOperator(n)

	// --- Quantifiers ---
	case *Forall:
		xtracer.Trace("compiler.CompileNode return case=Forall")
		return c.CompileQuantifier(n)
	case *Exists:
		xtracer.Trace("compiler.CompileNode return case=Exists")
		return c.CompileQuantifier(n)

	// --- Terms ---
	case *Atom:
		xtracer.Trace("compiler.CompileNode return case=Atom")
		return c.CompileApp(n, false)
	case *App:
		xtracer.Trace("compiler.CompileNode return case=App")
		return c.compileAppNode(n)
	case *Variable:
		xtracer.Trace("compiler.CompileNode return case=Variable")
		return c.CompileVariable(n)
	case *Old:
		xtracer.Trace("compiler.CompileNode return case=Old")
		return c.compileOld(n)
	case *MethodCall:
		xtracer.Trace("compiler.CompileNode return case=MethodCall")
		return c.compileMethodCall(n)

	// --- Symbol (bare identifier like x, y, used as terms in atoms) ---
	case *Symbol:
		xtracer.Trace("compiler.CompileNode return case=Symbol")
		return c.compileSymbol(n)

	// --- Named binder ---
	case *NamedBinder:
		xtracer.Trace("compiler.CompileNode return case=NamedBinder")
		return c.compileNamedBinder(n)

	// --- Labeled formula ---
	// Python: _labeled_formula_cmpl returns self.clone([...]) — an ivy_ast.LabeledFormula.
	// CompileNode returns lg.Expr, so we compile via CompileLF (which clones with
	// PRESERVE trace) and extract the inner formula. Callers needing the full
	// LabeledFormula (e.g., CompileAssertFormula) should use ThingLF directly.
	case *LabeledFormula:
		xtracer.Trace("compiler.CompileNode return case=LabeledFormula")
		xtracer.Trace("compiler.CompileLabeledFormula ENTER")
		compiled, err := c.CompileLF(n)
		if err != nil {
			return nil, err
		}
		fmla, ok := compiled.Formula.(Expr)
		if !ok {
			return nil, fmt.Errorf("CompileNode LabeledFormula: compiled formula is not lg.Expr (type %T)", compiled.Formula)
		}
		return fmla, nil

	// --- NativeExpr ---
	case *NativeExpr:
		xtracer.Trace("compiler.CompileNode return case=NativeExpr")
		return c.compileNativeExpr(n)

	// --- Trigger ---
	case *Trigger:
		xtracer.Trace("compiler.CompileNode return case=Trigger")
		return c.compileTrigger(n)

	// --- CompiledNode: already-compiled expression wrapper ---
	case *CompiledNode:
		if expr, ok := n.Node.(Expr); ok {
			xtracer.Trace("compiler.CompileNode return case=CompiledNode")
			return expr, nil
		}
		xtracer.Trace("compiler.CompileNode return case=CompiledNode-error")
		return nil, fmt.Errorf("CompiledNode does not contain lg.Expr: %T", n.Node)

	// --- Action AST nodes (from LALR parser) ---
	// Route through CompileActionBody which produces actions.ActionsAction.
	// Only types with explicit .cmpl in Python are listed here;
	// SetAction, HavocAction use other_thing (default).
	// InstantiateDecl is not routed from CompileNode but IS handled
	// in CompileActionBody when reached from other callers.
	// Each handler inside CompileActionBody emits its own trace.
	case *AssignAction, *AssumeAction, *AssertAction,
		*RequiresAction, *EnsuresAction, *SubgoalAction,
		*CrashAction, *ThunkAction,
		*LocalAction,
		*CallAction, *IfAction, *WhileAction,
		*DebugAction, *NativeAction:
		act, err := c.CompileActionBody(node)
		if err != nil {
			return nil, err
		}
		return act, nil

	// --- PatternBasedUpdate ---
	// Python: PatternBasedUpdate.cmpl walks defines/dependencies/patterns and
	// returns a PatternBasedUpdate with compiled symbol/pattern fields.
	case *PatternBasedUpdate:
		xtracer.Trace("compiler.CompileNode return case=PatternBasedUpdate")
		return c.compilePatternBasedUpdate(n)

	// --- Default: Python's AST.cmpl = other_thing ---
	// Handles all other unrecognized AST types.
	default:
		xtracer.Trace(fmt.Sprintf("compiler.CompileNode return case=default type=%s", compilerTypeName(node)))
		return c.OtherThing(node)
	}
}

// compileSymbol compiles a bare symbol node (like `x` in `r(x)`).
// It looks up the name in the signature to find its sort.
func (c *Compiler) compileSymbol(n *Symbol) (Expr, error) {
	name := n.Rep

	// Check variable context first (quantifier-bound variables)
	if sort, ok := c.VarCtx.Map[name]; ok {
		v, err := NewVariable(name, sort)
		return v, err
	}

	// Look up in signature (action parameters, constants, relations)
	entry, ok := c.Sig.Symbols.Get2(name)
	if ok {
		return NewConst(name, entry.Sort), nil
	}

	// Uppercase names are variables (Ivy convention)
	if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
		var sort Sort = TopS
		if n.Sort != nil {
			sortName := compilerExtractSortRep(n.Sort)
			if sortName != "" {
				if s, ok2 := c.Sig.Sorts.Get2(sortName); ok2 {
					sort = s
				}
			}
		}
		v, err := NewVariable(name, sort)
		return v, err
	}

	// Lowercase unresolved names become constants with TopSort
	return NewConst(name, TopS), nil
}

// compilePatternBasedUpdate compiles an ast.PatternBasedUpdate into an
// actions.PatternBasedUpdate. Walks defines/dependencies SymbolLists to
// produce []*lg.Const (looking up sorts in the signature), and compiles
// each UpdatePattern child.
// Corresponds to Python PatternBasedUpdate.cmpl (default compile → clone children).
func (c *Compiler) compilePatternBasedUpdate(n *PatternBasedUpdate) (Expr, error) {
	// Compile defines (SymbolList → []*lg.Const)
	var defines []*Const
	if sl, ok := n.Dfns.(*SymbolList); ok {
		for _, elem := range sl.Elems {
			name := compilerNodeRepStr(elem)
			if name == "" {
				continue
			}
			sym := c.lookupOrCreateConst(name)
			defines = append(defines, sym)
		}
	}

	// Compile dependencies (SymbolList → []*lg.Const)
	var deps []*Const
	if sl, ok := n.Deps.(*SymbolList); ok {
		for _, elem := range sl.Elems {
			name := compilerNodeRepStr(elem)
			if name == "" {
				continue
			}
			sym := c.lookupOrCreateConst(name)
			deps = append(deps, sym)
		}
	}

	// Compile patterns (UpdatePatternList → *actions.UpdatePatternList)
	patList := NewUpdatePatternList()
	if upl, ok := n.Patterns.(*UpdatePatternList); ok {
		for _, elem := range upl.Elems {
			up, ok := elem.(*UpdatePattern)
			if !ok {
				continue
			}
			compiled, err := c.compileUpdatePattern(up)
			if err != nil {
				return nil, err
			}
			patList.Add(compiled)
		}
	}

	result := NewPatternBasedUpdate(defines, deps, patList)
	return result, nil
}

// compileUpdatePattern compiles an ast.UpdatePattern into an actions.UpdatePattern.
//
// Python (ivy_compiler.py:528-532):
//
//	def UpdatePattern_cmpl(self):
//	    with ivy_logic.sig.copy():
//	        return ivy_ast.AST.cmpl(self)
//
// The placeholder symbols compiled below are temporary; they must NOT pollute
// the global signature for the rest of compilation. We follow the same
// sigCopy / restore pattern as compileDefnImpl (compiler.go:1313-1315).
func (c *Compiler) compileUpdatePattern(up *UpdatePattern) (*LogicUpdatePattern, error) {
	sigCopy := c.Sig.Copy()
	savedSig := c.Sig
	c.Sig = sigCopy
	defer func() { c.Sig = savedSig }()

	// Compile placeholders (ConstantDecl → []lg.Expr of *lg.Const).
	// These get added to sigCopy and discarded when the defer fires.
	var placeholders []Expr
	if up.Params != nil {
		for _, p := range up.Params.Args() {
			compiled, err := c.Thing(p)
			if err != nil {
				return nil, err
			}
			placeholders = append(placeholders, compiled)
		}
	}

	// Compile the pattern action
	var patternAction ActionsAction
	if up.Action != nil {
		compiled, err := c.CompileActionBody(up.Action)
		if err != nil {
			return nil, err
		}
		patternAction = compiled
	}

	// Compile requires (precondition formula)
	var precond Expr = True
	if up.Requires != nil {
		compiled, err := c.Thing(up.Requires)
		if err != nil {
			return nil, err
		}
		precond = compiled
	}

	// Compile ensures (transition relation formula)
	var transrel Expr = True
	if up.Ensures != nil {
		compiled, err := c.Thing(up.Ensures)
		if err != nil {
			return nil, err
		}
		transrel = compiled
	}

	return &LogicUpdatePattern{
		Placeholders: placeholders,
		Pattern:      patternAction,
		Precond:      precond,
		TransRel:     transrel,
	}, nil
}

// lookupOrCreateConst resolves a name to a *lg.Const, looking up the sort
// in the signature if available, or using TopS if not found.
func (c *Compiler) lookupOrCreateConst(name string) *Const {
	if entry, ok := c.Sig.Symbols.Get2(name); ok {
		return NewConst(name, entry.Sort)
	}
	return NewConst(name, TopS)
}

// nodeRepStr extracts a string representation (name) from an AST node,
// handling Atom, App, and Symbol types.
func compilerNodeRepStr(n Node) string {
	switch v := n.(type) {
	case *Atom:
		return v.Rep
	case *App:
		if sym, ok := v.Rep.(*Symbol); ok {
			return sym.Rep
		}
	case *Symbol:
		return v.Rep
	}
	return ""
}

// compileGeneric is the fallback for non-sort_infer_root nodes: compile each
// child and clone. For nodes with sort_infer_root, use OtherThing instead.
// Matches the else branch of Python's other_thing():
//
//	return self.clone([a.compile() for a in self.args])
func (c *Compiler) compileGeneric(node Node) (Expr, error) {
	args := node.Args()
	compiled := make([]Node, len(args))
	for i, a := range args {
		r, err := c.Thing(a)
		if err != nil {
			return nil, fmt.Errorf("compiling arg %d of %T: %w", i, node, err)
		}
		compiled[i] = r
	}
	if rank, ok := node.(*Ranking); ok {
		exprs := make([]Expr, 0, len(compiled))
		for _, c := range compiled {
			e, ok := c.(Expr)
			if !ok {
				return nil, fmt.Errorf("compiling Ranking arg of %T produced non-expr %T", node, c)
			}
			exprs = append(exprs, e)
		}
		res := NewRanking(nil, exprs...)
		res.SetLineno(rank.GetLineno())
		return res, nil
	}
	// Python: self.clone([a.compile() for a in self.args])
	result := node.Clone(compiled)
	if expr, ok := result.(Expr); ok {
		return expr, nil
	}
	// Cloned node isn't lg.Expr — extract compiled exprs and combine
	exprs := make([]Expr, 0, len(compiled))
	for _, c := range compiled {
		if e, ok := c.(Expr); ok {
			exprs = append(exprs, e)
		}
	}
	if len(exprs) == 0 {
		return True, nil
	}
	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return &LogicAnd{Terms: exprs}, nil
}

// --- Formula compilation ---

// compileArgs compiles all children of an AST node with no return context.
func (c *Compiler) compileArgs(node Node) ([]Expr, error) {
	xtracer.Trace("compiler.CompileArgs ENTER")
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	defer func() { c.ReturnCtx = saved }()

	args := node.Args()
	result := make([]Expr, len(args))
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

func (c *Compiler) compileAnd(n *And) (Expr, error) {
	xtracer.Trace("compiler.CompileAnd ENTER")
	args, err := c.compileArgs(n)
	if err != nil {
		return nil, err
	}
	return &LogicAnd{Terms: args}, nil
}

func (c *Compiler) compileOr(n *Or) (Expr, error) {
	xtracer.Trace("compiler.CompileOr ENTER")
	args, err := c.compileArgs(n)
	if err != nil {
		return nil, err
	}
	return &LogicOr{Terms: args}, nil
}

func (c *Compiler) compileNot(n *Not) (Expr, error) {
	xtracer.Trace("compiler.CompileNot ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	return &LogicNot{Body: args[0]}, nil
}

func (c *Compiler) compileImplies(n *Implies) (Expr, error) {
	xtracer.Trace("compiler.CompileImplies ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &LogicImplies{T1: args[0], T2: args[1]}, nil
}

func (c *Compiler) compileIff(n *Iff) (Expr, error) {
	xtracer.Trace("compiler.CompileIff ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &LogicIff{T1: args[0], T2: args[1]}, nil
}

func (c *Compiler) compileIte(n *Ite) (Expr, error) {
	xtracer.Trace("compiler.CompileIte ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 3 {
		return nil, err
	}
	return &LogicIte{ISort: args[1].NodeSort(), Cond: args[0], Then: args[1], Else: args[2]}, nil
}

func (c *Compiler) compileDefinition(n *Definition) (Expr, error) {
	xtracer.Trace("compiler.CompileDefinition ENTER\n  (via CompileNode op_pairs path, NOT compile_defn)")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return NewIvyDefinition(args[0], args[1]), nil
}

func (c *Compiler) compileGlobally(n *Globally) (Expr, error) {
	xtracer.Trace("compiler.CompileGlobally ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	empty := ""
	return &LogicGlobally{Environ: &empty, Body: args[0]}, nil
}

func (c *Compiler) compileEventually(n *Eventually) (Expr, error) {
	xtracer.Trace("compiler.CompileEventually ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) == 0 {
		return nil, err
	}
	empty := ""
	return &LogicEventually{Environ: &empty, Body: args[0]}, nil
}

func (c *Compiler) compileWhenOperator(n *WhenOperator) (Expr, error) {
	xtracer.Trace("compiler.CompileWhenOperator ENTER")
	args, err := c.compileArgs(n)
	if err != nil || len(args) < 2 {
		return nil, err
	}
	return &LogicWhenOperator{WSort: args[0].NodeSort(), Name: n.Name, T1: args[0], T2: args[1]}, nil
}

// --- Term compilation ---

// CompileApp compiles function application (Atom). This corresponds to
// Python's compile_app.
func (c *Compiler) CompileApp(n *Atom, old bool) (Expr, error) {
	xtracer.Trace("compiler.compile_app ENTER old=%v", old)
	rep := ResolveAlias(n.Rep, c.Module)

	// Handle boolean literals
	if rep == "true" || rep == "false" {
		if len(n.Terms) > 0 {
			return nil, NewIvyError(n, fmt.Sprintf("%s is not a function", rep))
		}
		if rep == "true" {
			return &LogicAnd{}, nil // empty And = true
		}
		return &LogicOr{}, nil // empty Or = false
	}

	// B4-R5: Python debug print: if any(isinstance(a,ivy_logic.Variable) for a in self.args): print("foo!")
	// lg.Variable satisfies ast.Node, so check for it in the AST args before compilation
	for _, a := range n.Terms {
		if _, ok := a.(*LogicVariable); ok {
			pp("foo!: %s", n)
			break
		}
	}

	// Compile arguments with no return context
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	xtracer.Trace("compiler.CompileApp args loop nTerms=%d", len(n.Terms))
	//pp("rep=%s old=%v", rep, old)
	args := make([]Expr, len(n.Terms))
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
			return nil, NewIvyError(n, "equality requires exactly 2 arguments")
		}
		return &Eq{T1: args[0], T2: args[1]}, nil
	}

	// Look up polymorphic symbol first
	sym, found := FindPolymorphicSymbol(rep, c.Module.Cfg.IuCfg)
	if !found {
		// Look up in signature
		entry, ok := c.Sig.Symbols.Get2(rep)
		if ok {
			sym = NewConst(rep, entry.Sort)
		}
	}

	if sym != nil {
		// Handle numerals with explicit sort annotation
		// B4-R6: Python guards with `sym is not ivy_logic.Equals` here, but
		// Go handles "=" with an early return above, so that guard is already satisfied.
		if IsNumeral(sym) {
			if n.ASort != nil {
				sortName := compilerExtractSortRep(n.ASort)
				if sortName != "S" {
					s, err := c.CmplSort(sortName)
					if err == nil {
						sym = NewConst(sym.Name, s)
					}
				}
			}
		}

		if old {
			sym = NewConst("old_"+sym.Name, sym.CSort)
		}
		if len(args) == 0 {
			return sym, nil
		}
		result, err := NewApply(sym, args...)
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
func (c *Compiler) compileAppNode(n *App) (Expr, error) {
	// If the rep is a Symbol, treat it like an Atom.
	if sym, ok := n.Rep.(*Symbol); ok {
		cfg := c.Module.Cfg.AstCfg
		atom := cfg.NewAtom(sym.Rep, n.Terms...)
		atom.SetLineno(n.GetLineno())
		atom.ASort = n.ASort
		return c.CompileApp(atom, false)
	}
	// NamedBinder-rep branch. Mirrors Python ivy_compiler.py:382-408, 457
	// (App.cmpl == Atom.cmpl == compile_app). Any rep type reaching
	// compile_app emits the same "compile_app ENTER" + args-loop trace
	// sequence; for a NamedBinder rep the rep is resolved via Python's
	// `rep.cmpl()` (line 399), which is a direct cmpl() dispatch without
	// the thing()/Thing() wrapper trace — so use c.Cmpl, not c.Thing.
	xtracer.Trace("compiler.compile_app ENTER old=%v", false)
	xtracer.Trace("compiler.CompileApp args loop nTerms=%d", len(n.Terms))

	// Args FIRST (Python ivy_compiler.py:391-394: with ReturnContext(None):
	// args = [a.compile() for a in self.args]).
	saved := c.ReturnCtx
	c.ReturnCtx = nil
	args := make([]Expr, len(n.Terms))
	for i, a := range n.Terms {
		r, err := c.Thing(a)
		if err != nil {
			c.ReturnCtx = saved
			return nil, err
		}
		args[i] = r
	}
	c.ReturnCtx = saved

	// Then compile the rep (must be NamedBinder here; Python ivy_compiler.py:399).
	repNode, err := c.Cmpl(n.Rep)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return repNode, nil
	}
	return NewApply(repNode, args...)
}

// CompileVariable compiles a Variable AST node to a logic.Variable.
func (c *Compiler) CompileVariable(n *Variable) (Expr, error) {
	xtracer.Trace("compiler.compile_variable ENTER")
	sort, err := c.variableSort(n)
	if err != nil {
		xtracer.Trace("compiler.CompileVariable error\n  var=%s vsort=%s err=%v", n.Rep, n.VSort, err)
		return nil, err
	}
	// If sort is top, check variable context
	if _, isTop := sort.(*TopSort); isTop {
		if s, ok := c.VarCtx.Map[n.Rep]; ok {
			sort = s
		}
	}
	v, err := NewVariable(n.Rep, sort)
	if err != nil {
		return nil, NewIvyError(n, fmt.Sprintf("bad variable: %s", err))
	}
	return v, nil
}

// variableSort resolves the sort of a variable AST node.
func (c *Compiler) variableSort(v *Variable) (Sort, error) {
	if v.VSort == "" {
		return TopS, nil
	}
	return c.CmplSort(v.VSort)
}

// CmplSort resolves a sort name to a logic.Sort, applying alias resolution.
func (c *Compiler) CmplSort(name string) (Sort, error) {
	resolved := ResolveAlias(name, c.Module)
	sort, err := c.Sig.FindSort(resolved, false)
	if err != nil {
		xtracer.Trace("compiler.CmplSort error")
		vv("compiler.CmplSort error: name=%s resolved=%s sigSorts=%v err=%v", name, resolved, c.Sig.SortNames(), err)
	}
	return sort, err
}

// compileOld compiles the Old operator: compile the inner term with old=true.
func (c *Compiler) compileOld(n *Old) (Expr, error) {
	xtracer.Trace("compiler.CompileOld ENTER")
	// The inner term should be an Atom or App
	if atom, ok := n.Term.(*Atom); ok {
		xtracer.Trace("compiler.CompileNode return case=Atom")
		return c.CompileApp(atom, true)
	}
	if app, ok := n.Term.(*App); ok {
		if sym, ok := app.Rep.(*Symbol); ok {
			cfg := c.Module.Cfg.AstCfg
			atom := cfg.NewAtom(sym.Rep, app.Terms...)
			atom.SetLineno(n.GetLineno())
			atom.ASort = app.ASort
			xtracer.Trace("compiler.CompileNode return case=App")
			return c.CompileApp(atom, true)
		}
	}
	// Fallback: compile the inner term normally
	return c.Thing(n.Term)
}

// compileMethodCall compiles obj.method() style calls.
func (c *Compiler) compileMethodCall(n *MethodCall) (Expr, error) {
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
	var methodArgs []Expr
	switch m := n.Method.(type) {
	case *Atom:
		childName = m.Rep
		methodArgs = make([]Expr, len(m.Terms))
		for i, a := range m.Terms {
			r, err := c.Thing(a)
			if err != nil {
				c.ReturnCtx = saved
				return nil, err
			}
			methodArgs[i] = r
		}
	case *Symbol:
		childName = m.Rep
	default:
		c.ReturnCtx = saved
		return nil, NewIvyError(n, fmt.Sprintf("unsupported method node type: %T", n.Method))
	}
	c.ReturnCtx = saved

	sort := base.NodeSort()
	if _, isTop := sort.(*TopSort); isTop {
		return nil, NewIvyError(n, fmt.Sprintf(
			"cannot apply method notation to %s because its type is not inferred", base))
	}

	// Look for the method as a child of the sort
	iuCfg := c.Module.Cfg.IuCfg
	destrName := iuCfg.ComposeNames(IvySortName(sort), childName)
	if c.TopCtx != nil {
		_, inModuleDestr := c.moduleDestructorSymbol(destrName)
		if _, inSig := c.Sig.Symbols.Get2(destrName); !inSig && !inModuleDestr {
			if _, inAct := c.TopCtx.Actions[destrName]; !inAct {
				// Try sibling of the sort
				pc := iuCfg.ParentChildName(IvySortName(sort))
				destrName = iuCfg.ComposeNames(pc[0], childName)
			}
		}
	}

	// Check if it's an action call
	if c.TopCtx != nil {
		if _, ok := c.TopCtx.Actions[destrName]; ok {
			if c.ExprCtx == nil {
				return nil, NewIvyError(n, fmt.Sprintf(
					"call to action %s not allowed outside an action", destrName))
			}
			allArgs := append([]Expr{base}, methodArgs...)
			cfg := c.Module.Cfg.AstCfg
			atom := cfg.NewAtom(destrName)
			atom.SetLineno(n.GetLineno())
			return c.CompileInlineCall(atom, allArgs, true)
		}
	}

	// Find the destructor symbol
	sym, err := c.findSymbol(destrName)
	if err != nil {
		var ok bool
		sym, ok = c.moduleDestructorSymbol(destrName)
		if !ok {
			return nil, err
		}
	}
	allArgs := append([]Expr{base}, methodArgs...)
	if len(allArgs) == 0 {
		return sym, nil
	}
	return NewApply(sym, allArgs...)
}

// compileNamedBinder compiles an AST NamedBinder.
func (c *Compiler) compileNamedBinder(n *NamedBinder) (Expr, error) {
	xtracer.Trace("compiler.CompileNamedBinder ENTER")
	vars := make([]*LogicVariable, len(n.Bounds))
	for i, b := range n.Bounds {
		compiled, err := c.Thing(b)
		if err != nil {
			return nil, err
		}
		v, ok := compiled.(*LogicVariable)
		if !ok {
			return nil, NewIvyError(n, fmt.Sprintf(
				"named binder bound %d is not a variable: %T", i, compiled))
		}
		vars[i] = v
	}
	body, err := c.Thing(n.Body)
	if err != nil {
		return nil, err
	}
	return NewNamedBinder(n.Name, vars, nil, body)
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
func (c *Compiler) CompileLF(lf *LabeledFormula) (*LabeledFormula, error) {
	// Compile label: Python: None if self.label is None else self.label.clone([sortify_with_inference(x) for x in self.label.args])
	var compiledLabel Node
	if lf.Label != nil {
		var newArgs []Node
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
	var compiledFormula Node
	if sb, ok := lf.Formula.(*SchemaBody); ok {
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
	result := lf.Clone([]Node{compiledLabel, compiledFormula}).(*LabeledFormula)
	return result, nil
}

// compileNativeExpr compiles a NativeExpr: compile args, preserve structure.
func (c *Compiler) compileNativeExpr(n *NativeExpr) (Expr, error) {
	xtracer.Trace("compiler.CompileNativeExpr ENTER")
	// Python NativeExpr.cmpl compiles every child, including the NativeCode
	// template, through the normal .compile() dispatch.
	args := n.Args()
	compiled := make([]Expr, len(args))
	for i, a := range args {
		r, err := c.Thing(a)
		if err != nil {
			return nil, err
		}
		compiled[i] = r
	}
	// B3-R3: Preserve all children in a NativeExpr with TopSort, matching Python:
	// res = self.clone([a.compile() for a in self.args]); res.sort = TopS
	return &LogicNativeExpr{CompiledChildren: compiled}, nil
}

// compileTrigger compiles a trigger hint.
func (c *Compiler) compileTrigger(n *Trigger) (Expr, error) {
	args := n.Args()
	if len(args) < 2 {
		return c.Thing(args[0])
	}
	// Compile pattern as-is, but sort-infer the trigger terms
	pattern, err := c.Thing(args[0])
	if err != nil {
		return nil, err
	}
	terms := make([]Expr, len(args)-1)
	for i, a := range args[1:] {
		t, err := c.SortifyWithInference(a)
		if err != nil {
			return nil, err
		}
		terms[i] = t
	}
	// Package as And(pattern, terms...) for now
	all := append([]Expr{pattern}, terms...)
	return &LogicAnd{Terms: all}, nil
}

// --- Quantifier compilation ---

// CompileQuantifier compiles Forall/Exists AST nodes.
func (c *Compiler) CompileQuantifier(node Node) (Expr, error) {
	xtracer.Trace("compiler.compile_quantifier ENTER")
	var bounds []Node
	var body Node
	var isForall bool

	switch n := node.(type) {
	case *Forall:
		bounds = n.Bounds
		body = n.Body
		isForall = true
	case *Exists:
		bounds = n.Bounds
		body = n.Body
		isForall = false
	default:
		return nil, fmt.Errorf("CompileQuantifier: unexpected type %T", node)
	}

	// Compile bound variables
	vars := make([]*LogicVariable, len(bounds))
	for i, b := range bounds {
		v, ok := b.(*Variable)
		if !ok {
			return nil, NewIvyError(node, fmt.Sprintf(
				"quantifier bound %d is not a variable: %T", i, b))
		}
		sort, err := c.variableSort(v)
		if err != nil {
			return nil, err
		}
		lv, err := NewVariable(v.Rep, sort)
		if err != nil {
			return nil, err
		}
		vars[i] = lv
	}

	// Set up variable context for the body
	savedVarCtx := c.VarCtx
	newMap := make(map[string]Sort, len(savedVarCtx.Map)+len(vars))
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
		return IvyForAll(vars, compiled), nil
	}
	return IvyExists(vars, compiled), nil
}

// --- Sort inference ---

// SortInfer resolves TopSort variables in a compiled logic node.
// Matches Python ivy_logic.py sort_infer:
//
//	res = concretize_sorts(term, sort)
//	check_concretely_sorted(res)
//
// The check_concretely_sorted step raises if any used variable or constant
// in the result still has TopSort or polymorphic sort. Mirroring this surfaces
// type-inference bugs at compile time (with a meaningful error) instead of
// letting silent TopSorts propagate to z3bridge.TranslateSort, which can only
// panic.
func (c *Compiler) SortInfer(node Expr) (Expr, error) {
	res, err := ConcretizeSorts(node, nil)
	if err != nil {
		return nil, err
	}
	if err := CheckConcretelySorted(res, nil); err != nil {
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
func (c *Compiler) SortifyWithInference(astNode Node) (Expr, error) {
	xtracer.Trace("compiler.sortify_with_inference ENTER")
	if xtracer.Enabled {
		c.SigCheck("SortifyWithInference")
	}
	// B3-R1: wrap compilation in top_sort_as_default, matching Python
	tsDefault := TopSortAsDefault(c.Sig)
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

// CompileWithSortInference is the adapter used by ast.Schema.GetInstance.
// Python exposes compile_with_sort_inference directly on AST nodes; Go keeps
// the operation on Compiler.
func (c *Compiler) CompileWithSortInference(node Node) (Node, error) {
	return c.SortifyWithInference(node)
}

// CompileConst compiles a constant declaration, adding it to the signature.
func (c *Compiler) CompileConst(v Node, sig *Sig) (*Const, error) {
	xtracer.Trace("compiler.CompileConst ENTER")
	var name string
	var sortArgs []Node
	var sortNode Node

	switch n := v.(type) {
	case *Atom:
		name = n.Rep
		sortArgs = n.Terms
		sortNode = n.ASort
	case *App:
		if sym, ok := n.Rep.(*Symbol); ok {
			name = sym.Rep
		} else {
			name = fmt.Sprint(n.Rep)
		}
		sortArgs = n.Terms
		sortNode = n.ASort
	case *Symbol:
		name = n.Rep
		sortNode = n.Sort
	default:
		return nil, NewIvyError(v, fmt.Sprintf("cannot compile const from %T", v))
	}

	// Determine the range sort
	var rng Sort
	if sortNode != nil {
		sortName := compilerExtractSortRep(sortNode)
		if sortName != "" {
			var err error
			rng, err = c.CmplSort(sortName)
			if err != nil {
				return nil, err
			}
		}
	}
	if rng == nil {
		xtracer.Trace("compiler.CompileConst no_sort calling_default_sort HASH canon=%s", sig.Canon())
		var err error
		rng, err = GetDefaultSort(sig)
		if err != nil {
			return nil, err
		}
	}

	// Get the function sort from arguments
	sort := c.getFunctionSort(sig, sortArgs, rng)

	return c.AddSymbol(name, sort, sig)
}

// getFunctionSort constructs a FunctionSort from arg sorts and range.
func (c *Compiler) getFunctionSort(sig *Sig, args []Node, rng Sort) Sort {
	xtracer.Trace("compiler.GetFunctionSort ENTER")
	if len(args) == 0 {
		return rng
	}
	sorts := make([]Sort, 0, len(args)+1)
	for _, a := range args {
		// Python: arg.compile().get_sort() — compiles every arg uniformly
		compiled, err := c.Thing(a)
		if err != nil {
			sorts = append(sorts, TopS)
		} else {
			sorts = append(sorts, compiled.NodeSort())
		}
	}
	sorts = append(sorts, rng)
	fs, err := NewFunctionSort(sorts...)
	if err != nil {
		return rng
	}
	return fs
}

// AddSymbol adds a symbol with the given name and sort to the signature.
// Python's add_symbol (ivy_logic.py:378) does NOT call add_to_hierarchy;
// hierarchy is populated only from decls.defined in ivy_compile.
func (c *Compiler) AddSymbol(name string, sort Sort, sig *Sig) (*Const, error) {
	sym, err := sig.AddSymbol(name, sort)
	if err != nil {
		return nil, err
	}
	return sym, nil
}

// findSymbol looks up a symbol in the signature.
func (c *Compiler) findSymbol(name string) (*Const, error) {
	// Try polymorphic first
	if sym, ok := FindPolymorphicSymbol(name, c.Module.Cfg.IuCfg); ok {
		return sym, nil
	}
	return c.Sig.FindSymbol(name, false)
}

func (c *Compiler) moduleDestructorSymbol(name string) (*Const, bool) {
	if c == nil || c.Module == nil {
		return nil, false
	}
	if _, ok := c.Module.DestructorSorts[name]; !ok {
		return nil, false
	}
	for _, destrs := range c.Module.SortDestructors.All() {
		for _, destr := range destrs {
			if destr != nil && destr.Name == name {
				return destr, true
			}
		}
	}
	return nil, false
}

// CompileDefn compiles a definition (lhs = rhs) AST node.
// Corresponds to Python's compile_defn.
func (c *Compiler) CompileDefn(df *Definition) (Expr, error) {
	xtracer.Trace("compiler.CompileDefn ENTER")
	//pp("(CompileDefn -> compileDefnImpl isSchema=false) [go side]")
	return c.compileDefnImpl(df, false)
}

// CompileDefnSchema compiles a definition schema (DefinitionSchema variant).
func (c *Compiler) CompileDefnSchema(df *DefinitionSchema) (Expr, error) {
	xtracer.Trace("compiler.CompileDefnSchema ENTER")
	//pp("(CompileDefnSchema -> compileDefnImpl isSchema=true) [go side]")
	return c.compileDefnImpl(&df.Definition, true)
}

func (c *Compiler) compileDefnImpl(df *Definition, isSchema bool) (Expr, error) {
	xtracer.Trace("compiler.CompileDefnImpl ENTER")
	if xtracer.Enabled {
		c.SigCheck("CompileDefnImpl.entry")
	}
	lhs := df.Lhs
	var lhsAtom *Atom
	if a, ok := lhs.(*Atom); ok {
		lhsAtom = a
	}

	sigCopy := c.Sig.Copy()
	savedSig := c.Sig
	c.Sig = sigCopy
	if xtracer.Enabled {
		c.SigCheck("CompileDefnImpl.afterCopy")
	}
	if lhsAtom != nil {
		if _, isPoly := FindPolymorphicSymbol(lhsAtom.Rep, c.Module.Cfg.IuCfg); !isPoly {
			if _, exists := sigCopy.Symbols.Get2(lhsAtom.Rep); !exists {
				if _, err := sigCopy.AddSymbol(lhsAtom.Rep, TopFunctionSort(len(lhsAtom.Terms))); err != nil {
					c.Sig = savedSig
					return nil, err
				}
			}
		}
	}

	// Compile any constant parameters in the LHS and collect variable sort substitutions
	subst := make(map[string]string)
	if lhsAtom != nil {
		for _, p := range lhsAtom.Terms {
			if v, isVar := p.(*Variable); isVar {
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
	xtracer.Trace("compiler.CompileDefnImpl rhs type=%s", compilerTypeName(rhs))
	//pp("isSchema=%v lhs=%v", isSchema, df.Lhs)
	if len(subst) > 0 {
		rhs = SetVariableSorts(rhs, subst)
		xtracer.Trace("compiler.CompileDefnImpl rhs after subst type=%s", compilerTypeName(rhs))
		//pp("subst=%v", subst)
	}

	// Handle SomeExpr on RHS
	// Python: if isinstance(df.args[1], ivy_ast.SomeExpr):
	if someExpr, ok := rhs.(*SomeExpr); ok {
		xtracer.Trace("compiler.CompileDefnImpl SomeExpr branch")
		//pp("isSchema=%v", isSchema)
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
		iteNode := cfg.NewIte(someExpr.Fmla, ifval, elseval)
		eqNode := cfg.NewAtom("=", df.Lhs, iteNode)
		forallNode := cfg.NewForall([]Node{someExpr.Param}, eqNode)

		fmla, err := c.SortifyWithInference(forallNode)
		c.Sig = savedSig
		if err != nil {
			return nil, err
		}

		// Extract from the compiled forall: variables[0], body.args[1].args[0..2]
		var defLhs, someArgs Expr
		if forall, ok := fmla.(*ForAll); ok && len(forall.Variables) > 0 {
			param := forall.Variables[0]
			if eq, ok := forall.Body.(*Eq); ok {
				defLhs = eq.T1
				if ite, ok := eq.T2.(*LogicIte); ok {
					someNode := &LogicSome{
						Params: []Expr{param},
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

		result := NewIvyDefinition(defLhs, someArgs)
		if isSchema {
			return NewIvyDefinitionSchema(defLhs, someArgs), nil
		}
		return result, nil
	}

	// Standard definition: compile as equality lhs = rhs, then apply sort inference
	cfg := c.Module.Cfg.AstCfg
	xtracer.Trace("compiler.CompileDefnImpl standard branch rhs type=%s", compilerTypeName(rhs))
	//pp("isSchema=%v lhs=%v rhs=%v", isSchema, df.Lhs, rhs)
	eqAtom := cfg.NewAtom("=", df.Lhs, rhs)
	xtracer.Trace("compiler.CompileDefnImpl eqAtom nTerms=%d", len(eqAtom.Terms))
	//pp("eqAtom=%v", eqAtom)
	for ti, tt := range eqAtom.Terms {
		if tt == nil {
			xtracer.Trace("compiler.CompileDefnImpl eqAtom.Terms[%d] = nil", ti)
		} else {
			xtracer.Trace("compiler.CompileDefnImpl eqAtom.Terms[%d] type=%s", ti, compilerTypeName(tt))
			//pp("val=%v", tt)
		}
	}
	eqAtom.SetLineno(df.GetLineno())
	compiled, err := c.SortifyWithInference(eqAtom)
	c.Sig = savedSig

	if err != nil {
		return nil, err
	}

	// Extract lhs and rhs from the compiled equality
	if eq, ok := compiled.(*Eq); ok {
		if isSchema {
			return NewIvyDefinitionSchema(eq.T1, eq.T2), nil
		}
		return NewIvyDefinition(eq.T1, eq.T2), nil
	}
	// If sort inference returned the equality as-is, wrap in Definition
	if isSchema {
		return NewIvyDefinitionSchema(compiled, compiled), nil
	}
	return NewIvyDefinition(compiled, compiled), nil
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
//
// This mirrors Python's compile() dispatch for tactic types:
//   - Types with explicit compile overrides in Python bypass thing() — no traces.
//   - Types without overrides go through thing() -> other_thing() — produce
//     Thing/CompileNode/OtherThing traces and do clone([a.compile() for a in self.args]).
func (c *Compiler) CompileTactic(node Node) (Node, error) {
	if node == nil {
		return nil, nil
	}

	// --- Types with EXPLICIT compile overrides in Python ---
	// These bypass thing(), producing no Thing/CompileNode/OtherThing traces.
	switch n := node.(type) {
	case *SchemaInstantiation:
		// Python: compile_schema_instantiation returns self
		return n, nil

	case *AssumeTactic:
		// Python: TacticWithMatch.compile = compile_schema_instantiation → return self
		return n, nil

	case *AssumeGlobalTactic:
		// Python: TacticWithMatch.compile = compile_schema_instantiation → return self
		return n, nil

	case *LetTactic:
		// Python: compile_let_tactic returns self
		return n, nil

	case *WitnessTactic:
		// Python: compile_witness_tactic returns self
		return n, nil

	case *UnfoldTactic:
		// Python: compile_unfold_tactic returns self
		return n, nil

	case *ForgetTactic:
		// Python: compile_forget_tactic returns self
		return n, nil

	case *FunctionTactic:
		// Python: compile_function_tactic returns self
		return n, nil

	case *IfTactic:
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
		condWrapper := c.Module.Cfg.AstCfg.NewCompiledNode(cond)
		return n.Clone([]Node{condWrapper, thenBranch, elseBranch}), nil

	case *PropertyTactic:
		// Python: compile_property_tactic
		// prop = self.args[0] (not compiled)
		// name = self.args[1]: if not NoneAST, compile name args with UnsortedContext
		// proof = self.args[2].compile()
		prop := n.Prop
		name := n.PName
		if _, isNone := name.(*NoneAST); !isNone && name != nil {
			// Compile name's args with unsorted context
			if atom, ok := name.(*Atom); ok {
				compiledTerms := make([]Node, len(atom.Terms))
				for i, arg := range atom.Terms {
					compiled, err := c.Thing(arg)
					if err != nil {
						compiledTerms[i] = arg
						continue
					}
					compiledTerms[i] = c.Module.Cfg.AstCfg.NewCompiledNode(compiled)
				}
				name = &Atom{Base: atom.Base, Rep: atom.Rep, Terms: compiledTerms, ASort: atom.ASort}
			}
		}
		proof, err := c.CompileTactic(n.Proof)
		if err != nil {
			proof = n.Proof
		}
		return &PropertyTactic{Base: n.Base, Prop: prop, PName: name, Proof: proof}, nil

	case *TacticTactic:
		// Python: compile_tactic_tactic returns self.clone(self.args)
		return n.Clone(n.Args()), nil

	case *ProofTactic:
		// Python: compile_proof_tactic compiles label and proof
		proof, err := c.CompileTactic(n.Proof)
		if err != nil {
			proof = n.Proof
		}
		return &ProofTactic{Base: n.Base, TLabel: n.TLabel, Proof: proof}, nil

	// --- Types WITHOUT compile overrides in Python ---
	// These go through thing() -> cmpl() = other_thing() in Python.
	// Produces: Thing ENTER, CompileNode ENTER, CompileNode return default,
	//           OtherThing ENTER, [child compilation], OtherThing return, Thing return.
	// Python other_thing: self.clone([a.compile() for a in self.args])
	default:
		tn := compilerTypeName(node)
		xtracer.Trace(fmt.Sprintf("compiler.Thing ENTER type=%s", tn))
		xtracer.Trace(fmt.Sprintf("compiler.CompileNode ENTER type=%s", tn))
		xtracer.Trace(fmt.Sprintf("compiler.CompileNode return case=default type=%s", tn))
		xtracer.Trace(fmt.Sprintf("compiler.OtherThing ENTER type=%s", tn))
		// Python: self.clone([a.compile() for a in self.args])
		args := node.Args()
		compiled := make([]Node, len(args))
		for i, a := range args {
			ca, err := c.CompileTactic(a)
			if err != nil {
				return nil, err
			}
			compiled[i] = ca
		}
		result := node.Clone(compiled)
		xtracer.Trace(fmt.Sprintf("compiler.OtherThing return type=%s sort_infer_root=False", tn))
		xtracer.Trace(fmt.Sprintf("compiler.Thing return type=%s", tn))
		return result, nil
	}
}

// extractSortRep extracts a string sort name from an AST sort node.
func compilerExtractSortRep(n Node) string {
	if n == nil {
		return ""
	}
	switch s := n.(type) {
	case *Symbol:
		return s.Rep
	case *Atom:
		return s.Rep
	default:
		r := fmt.Sprint(n)
		if r == "" || r == "<nil>" {
			return ""
		}
		return strings.TrimSpace(r)
	}
}
