// Ported to Go from ivy_dafny_compiler.py.

// Package dafnygen translates Dafny AST nodes into Ivy IR (intermediate
// representation). This covers variable declarations, assignments,
// control flow (if/while/return), method declarations with
// call-by-value semantics, and logical connectives (and/or/not/implies/iff).
//
// The translation uses a context stack to track module-level type
// information, method-level renaming, expression temporaries, and
// local variable scopes.
package dafnygen

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// Unique-name generators
// ---------------------------------------------------------------------------

// UniqueRenamer produces unique names with a given prefix. It is
// analogous to ivy_utils.UniqueRenamer.
type UniqueRenamer struct {
	prefix string
	count  int64
}

// NewUniqueRenamer creates a renamer with the given prefix.
func NewUniqueRenamer(prefix string) *UniqueRenamer {
	return &UniqueRenamer{prefix: prefix}
}

// Next returns a fresh name.
func (r *UniqueRenamer) Next(hint string) string {
	n := atomic.AddInt64(&r.count, 1) - 1
	if hint == "" {
		return fmt.Sprintf("%s%d", r.prefix, n)
	}
	return fmt.Sprintf("%s%s%d", r.prefix, hint, n)
}

// ---------------------------------------------------------------------------
// Dafny AST types (minimal representation)
// ---------------------------------------------------------------------------

// DafnyType holds a type name string.
type DafnyType struct {
	Rep string
}

func (t *DafnyType) String() string { return t.Rep }

// IsBool reports whether the type is "bool".
func (t *DafnyType) IsBool() bool { return t.Rep == "bool" }

// TypedSymbol is a symbol with an associated type.
type TypedSymbol struct {
	Rep  string
	Type *DafnyType
}

func (s *TypedSymbol) String() string { return s.Rep }

// ---------------------------------------------------------------------------
// Operator mapping
// ---------------------------------------------------------------------------

// OpMap translates Dafny operators to Ivy operators.
var OpMap = map[string]string{
	"==": "=",
}

// MapOp translates an operator rep, returning the original if no mapping exists.
func MapOp(rep string) string {
	if m, ok := OpMap[rep]; ok {
		return m
	}
	return rep
}

// ArithOps are the arithmetic operators whose result type matches
// their argument type (used for type inference).
var ArithOps = map[string]bool{
	"+": true,
	"-": true,
	"*": true,
}

// ---------------------------------------------------------------------------
// Ivy IR nodes (simplified output representation)
// ---------------------------------------------------------------------------

// IvyNode is the interface for all Ivy IR output nodes.
type IvyNode interface {
	IvyString() string
}

// Atom is an Ivy atom: name(args...).
type Atom struct {
	Rep  string
	Args []IvyNode
}

func NewAtom(rep string, args ...IvyNode) *Atom {
	return &Atom{Rep: rep, Args: args}
}

func (a *Atom) IvyString() string {
	if len(a.Args) == 0 {
		return a.Rep
	}
	parts := make([]string, len(a.Args))
	for i, arg := range a.Args {
		parts[i] = arg.IvyString()
	}
	return a.Rep + "(" + strings.Join(parts, ",") + ")"
}

// App is a function application.
type App struct {
	Rep  string
	Sort string // optional sort annotation
	Args []IvyNode
}

func NewApp(rep string, args ...IvyNode) *App {
	return &App{Rep: rep, Args: args}
}

func (a *App) IvyString() string {
	if len(a.Args) == 0 {
		s := a.Rep
		if a.Sort != "" {
			s += ":" + a.Sort
		}
		return s
	}
	parts := make([]string, len(a.Args))
	for i, arg := range a.Args {
		parts[i] = arg.IvyString()
	}
	return a.Rep + "(" + strings.Join(parts, ",") + ")"
}

// AssignAction is lhs := rhs.
type AssignAction struct {
	LHS, RHS IvyNode
}

func NewAssignAction(lhs, rhs IvyNode) *AssignAction {
	return &AssignAction{LHS: lhs, RHS: rhs}
}

func (a *AssignAction) IvyString() string {
	return a.LHS.IvyString() + " := " + a.RHS.IvyString()
}

// Sequence is a list of actions.
type Sequence struct {
	Actions []IvyNode
}

func NewSequence(acts ...IvyNode) *Sequence {
	return &Sequence{Actions: acts}
}

func (s *Sequence) IvyString() string {
	parts := make([]string, len(s.Actions))
	for i, a := range s.Actions {
		parts[i] = a.IvyString()
	}
	return strings.Join(parts, "; ")
}

// AssumeAction is "assume F".
type AssumeAction struct {
	Formula IvyNode
}

func (a *AssumeAction) IvyString() string {
	return "assume " + a.Formula.IvyString()
}

// AssertAction is "assert F".
type AssertAction struct {
	Formula IvyNode
}

func (a *AssertAction) IvyString() string {
	return "assert " + a.Formula.IvyString()
}

// IfAction is "if C { T } else { E }".
type IfAction struct {
	Cond     IvyNode
	ThenBody IvyNode
	ElseBody IvyNode
}

func (a *IfAction) IvyString() string {
	s := "if " + a.Cond.IvyString() + " { " + a.ThenBody.IvyString() + " }"
	if a.ElseBody != nil {
		s += " else { " + a.ElseBody.IvyString() + " }"
	}
	return s
}

// CallAction is "call name(args)".
type CallAction struct {
	Target IvyNode
}

func (a *CallAction) IvyString() string {
	return "call " + a.Target.IvyString()
}

// InstantiateAction is "instantiate atom".
type InstantiateAction struct {
	Target IvyNode
}

func (a *InstantiateAction) IvyString() string {
	return "instantiate " + a.Target.IvyString()
}

// LocalAction introduces local variables.
type LocalAction struct {
	Locals []string
	Body   IvyNode
}

func (a *LocalAction) IvyString() string {
	return "local " + strings.Join(a.Locals, ",") + " { " + a.Body.IvyString() + " }"
}

// LetAction binds definitions.
type LetAction struct {
	Bindings []IvyNode
	Body     IvyNode
}

func (a *LetAction) IvyString() string {
	parts := make([]string, len(a.Bindings))
	for i, b := range a.Bindings {
		parts[i] = b.IvyString()
	}
	return "let " + strings.Join(parts, ", ") + " { " + a.Body.IvyString() + " }"
}

// And is a logical conjunction.
type And struct {
	Terms []IvyNode
}

func NewAnd(terms ...IvyNode) *And {
	return &And{Terms: terms}
}

func (a *And) IvyString() string {
	if len(a.Terms) == 0 {
		return "true"
	}
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = t.IvyString()
	}
	return "(" + strings.Join(parts, " & ") + ")"
}

// Or is a logical disjunction.
type Or struct {
	Terms []IvyNode
}

func NewOr(terms ...IvyNode) *Or {
	return &Or{Terms: terms}
}

func (o *Or) IvyString() string {
	if len(o.Terms) == 0 {
		return "false"
	}
	parts := make([]string, len(o.Terms))
	for i, t := range o.Terms {
		parts[i] = t.IvyString()
	}
	return "(" + strings.Join(parts, " | ") + ")"
}

// Not is logical negation.
type Not struct {
	Body IvyNode
}

func (n *Not) IvyString() string {
	return "~" + n.Body.IvyString()
}

// Implies is logical implication (encoded as ~A | B).
type IvyImplies struct {
	LHS, RHS IvyNode
}

func (i *IvyImplies) IvyString() string {
	return "(" + (&Not{Body: i.LHS}).IvyString() + " | " + i.RHS.IvyString() + ")"
}

// RME is a requires-modifies-ensures clause.
type RME struct {
	Requires IvyNode
	Modifies []string
	Ensures  IvyNode
}

func NewRME(requires IvyNode, modifies []string, ensures IvyNode) *RME {
	return &RME{Requires: requires, Modifies: modifies, Ensures: ensures}
}

func (r *RME) IvyString() string {
	var parts []string
	if r.Requires != nil {
		parts = append(parts, "requires "+r.Requires.IvyString())
	}
	if len(r.Modifies) > 0 {
		parts = append(parts, "modifies "+strings.Join(r.Modifies, ","))
	}
	if r.Ensures != nil {
		parts = append(parts, "ensures "+r.Ensures.IvyString())
	}
	return strings.Join(parts, " ")
}

// RelationDecl declares a boolean relation.
type RelationDecl struct {
	Atom *Atom
}

func (d *RelationDecl) IvyString() string {
	return "relation " + d.Atom.IvyString()
}

// ConstantDecl declares a constant.
type ConstantDecl struct {
	App *App
}

func (d *ConstantDecl) IvyString() string {
	return "individual " + d.App.IvyString()
}

// ActionDecl declares an action.
type ActionDecl struct {
	Name string
	Body IvyNode
}

func (d *ActionDecl) IvyString() string {
	return "action " + d.Name + " = { " + d.Body.IvyString() + " }"
}

// StateDecl declares a state machine.
type StateDecl struct {
	Name string
	Def  IvyNode
}

func (d *StateDecl) IvyString() string {
	return "state " + d.Name + " = " + d.Def.IvyString()
}

// AssertDecl is an assertion declaration.
type AssertDecl struct {
	Formula IvyNode
	Lineno  int
}

func (d *AssertDecl) IvyString() string {
	return "assert " + d.Formula.IvyString()
}

// MacroDecl declares a macro.
type MacroDecl struct {
	Name string
	Def  IvyNode
}

func (d *MacroDecl) IvyString() string {
	return "macro " + d.Name + " = " + d.Def.IvyString()
}

// ---------------------------------------------------------------------------
// Module — accumulates declarations
// ---------------------------------------------------------------------------

// IvyModule collects generated Ivy declarations and macros.
type IvyModule struct {
	Decls  []IvyNode
	Macros map[string]IvyNode
}

// NewIvyModule creates an empty module.
func NewIvyModule() *IvyModule {
	return &IvyModule{Macros: make(map[string]IvyNode)}
}

// Declare appends a declaration.
func (m *IvyModule) Declare(d IvyNode) {
	m.Decls = append(m.Decls, d)
}

// String renders all declarations.
func (m *IvyModule) String() string {
	parts := make([]string, len(m.Decls))
	for i, d := range m.Decls {
		parts[i] = d.IvyString()
	}
	return strings.Join(parts, "\n")
}

// ---------------------------------------------------------------------------
// Translation context
// ---------------------------------------------------------------------------

// ModuleContext holds module-level state for compilation.
type ModuleContext struct {
	Types map[string][][]*DafnyType
	Mod   *IvyModule
}

// NewModuleContext creates a module context.
func NewModuleContext(mod *IvyModule) *ModuleContext {
	return &ModuleContext{
		Types: make(map[string][][]*DafnyType),
		Mod:   mod,
	}
}

// MethodContext holds method-level state.
type MethodContext struct {
	TempRenamer  *UniqueRenamer
	LocalRenamer *UniqueRenamer
	Modifies     []*Atom
	OutParams    []*TypedSymbol
}

// NewMethodContext creates a method context.
func NewMethodContext(methodName string, modifies []*Atom, outParams []*TypedSymbol) *MethodContext {
	return &MethodContext{
		TempRenamer:  NewUniqueRenamer(methodName + ":tmp"),
		LocalRenamer: NewUniqueRenamer("loc:" + methodName + ":"),
		Modifies:     modifies,
		OutParams:    outParams,
	}
}

// ScopeContext tracks local variables within a scope.
type ScopeContext struct {
	Locals    map[string]*TypedSymbol
	NewLocals []*TypedSymbol
	Returns   bool
}

// NewScopeContext creates a scope context, optionally copying parent locals.
func NewScopeContext(parent *ScopeContext) *ScopeContext {
	sc := &ScopeContext{
		Locals: make(map[string]*TypedSymbol),
	}
	if parent != nil {
		for k, v := range parent.Locals {
			sc.Locals[k] = v
		}
	}
	return sc
}

// ExprContext accumulates code emitted during expression compilation.
type ExprContext struct {
	Code []IvyNode
}

// NewExprContext creates an expression context.
func NewExprContext() *ExprContext {
	return &ExprContext{}
}

// ---------------------------------------------------------------------------
// Compiler ties together the contexts for a single compilation pass.
// ---------------------------------------------------------------------------

// Compiler holds all the context state for a compilation pass.
type Compiler struct {
	ModCtx   *ModuleContext
	MethCtx  *MethodContext
	ScopeCtx *ScopeContext
	ExprCtx  *ExprContext
}

// NewCompiler creates a compiler targeting the given module.
func NewCompiler(mod *IvyModule) *Compiler {
	return &Compiler{
		ModCtx: NewModuleContext(mod),
	}
}

// MakeApp creates an Atom (if type is bool) or App from a typed symbol.
func MakeApp(sym *TypedSymbol, args ...IvyNode) IvyNode {
	if sym.Type != nil && sym.Type.IsBool() {
		return NewAtom(sym.Rep, args...)
	}
	return NewApp(sym.Rep, args...)
}

// MakeTemp creates a temporary variable, declares it, and returns an
// IvyNode referring to it.
func (c *Compiler) MakeTemp(typ *DafnyType) IvyNode {
	name := c.MethCtx.TempRenamer.Next("")
	sym := &TypedSymbol{Rep: name, Type: typ}
	c.DeclareVarType(sym)
	c.ScopeCtx.NewLocals = append(c.ScopeCtx.NewLocals, sym)
	return MakeApp(sym)
}

// DeclareVarType emits an Ivy declaration for a typed symbol.
func (c *Compiler) DeclareVarType(sym *TypedSymbol) {
	if sym.Type.IsBool() {
		c.ModCtx.Mod.Declare(&RelationDecl{Atom: NewAtom(sym.Rep)})
	} else {
		app := NewApp(sym.Rep)
		if sym.Type.Rep != "object" {
			app.Sort = sym.Type.Rep
		}
		c.ModCtx.Mod.Declare(&ConstantDecl{App: app})
	}
}

// CompileVarDecl translates variable declarations.
func (c *Compiler) CompileVarDecl(symbols []*TypedSymbol) {
	for _, s := range symbols {
		c.DeclareVarType(s)
	}
}

// CompileAssign translates an assignment statement.
func (c *Compiler) CompileAssign(lhss, rhss []IvyNode) (IvyNode, error) {
	if len(lhss) != len(rhss) {
		return nil, fmt.Errorf("dafnygen: wrong number of values in assignment")
	}
	if len(rhss) == 1 {
		return NewAssignAction(lhss[0], rhss[0]), nil
	}
	// Multi-assignment: use temps
	code := make([]IvyNode, 0, len(rhss)*2)
	temps := make([]IvyNode, len(rhss))
	for i, rhs := range rhss {
		tmp := c.MakeTemp(&DafnyType{Rep: "object"})
		code = append(code, NewAssignAction(tmp, rhs))
		temps[i] = tmp
	}
	for i, lhs := range lhss {
		code = append(code, NewAssignAction(lhs, temps[i]))
	}
	return NewSequence(code...), nil
}

// CompileAssume translates an assume statement.
func (c *Compiler) CompileAssume(formula IvyNode) IvyNode {
	return &AssumeAction{Formula: formula}
}

// CompileAssert translates an assert statement.
func (c *Compiler) CompileAssert(formula IvyNode) IvyNode {
	return &AssertAction{Formula: formula}
}

// CompileIf translates an if statement.
func (c *Compiler) CompileIf(cond, thenBody, elseBody IvyNode) IvyNode {
	return &IfAction{Cond: cond, ThenBody: thenBody, ElseBody: elseBody}
}

// CompileReturn translates a return statement by assigning to output params.
func (c *Compiler) CompileReturn(values []IvyNode) (IvyNode, error) {
	if len(values) != len(c.MethCtx.OutParams) {
		return nil, fmt.Errorf("dafnygen: return value count mismatch: got %d, expected %d",
			len(values), len(c.MethCtx.OutParams))
	}
	code := make([]IvyNode, len(values))
	for i, val := range values {
		code[i] = NewAssignAction(MakeApp(c.MethCtx.OutParams[i]), val)
	}
	if c.ScopeCtx != nil {
		c.ScopeCtx.Returns = true
	}
	if len(code) == 1 {
		return code[0], nil
	}
	return NewSequence(code...), nil
}

// CompileBlock compiles a block of statements within a new scope.
func (c *Compiler) CompileBlock(stmts []IvyNode) IvyNode {
	parent := c.ScopeCtx
	c.ScopeCtx = NewScopeContext(parent)
	defer func() {
		if parent != nil {
			parent.Returns = parent.Returns || c.ScopeCtx.Returns
		}
		c.ScopeCtx = parent
	}()

	result := NewSequence(stmts...)

	if len(c.ScopeCtx.NewLocals) > 0 {
		locals := make([]string, len(c.ScopeCtx.NewLocals))
		for i, l := range c.ScopeCtx.NewLocals {
			locals[i] = l.Rep
		}
		return &LocalAction{Locals: locals, Body: result}
	}
	return result
}

// TranslateSymbol handles translation of a Dafny symbol reference.
func (c *Compiler) TranslateSymbol(rep string) IvyNode {
	mapped := MapOp(rep)
	if c.ScopeCtx != nil {
		if local, ok := c.ScopeCtx.Locals[rep]; ok {
			return MakeApp(local)
		}
	}
	return NewApp(mapped)
}

// TranslateAnd translates a Dafny && to Ivy And.
func TranslateAnd(terms ...IvyNode) IvyNode {
	return NewAnd(terms...)
}

// TranslateOr translates a Dafny || to Ivy Or.
func TranslateOr(terms ...IvyNode) IvyNode {
	return NewOr(terms...)
}

// TranslateNot translates a Dafny ! to Ivy Not.
func TranslateNot(body IvyNode) IvyNode {
	return &Not{Body: body}
}

// TranslateImplies translates ==> to ~A | B.
func TranslateImplies(lhs, rhs IvyNode) IvyNode {
	return &IvyImplies{LHS: lhs, RHS: rhs}
}

// TranslateIff translates <==> to (A & B) | (~A & ~B).
func TranslateIff(lhs, rhs IvyNode) IvyNode {
	return NewOr(
		NewAnd(lhs, rhs),
		NewAnd(&Not{Body: lhs}, &Not{Body: rhs}),
	)
}

// ---------------------------------------------------------------------------
// Preamble for integer support
// ---------------------------------------------------------------------------

// Preamble contains Ivy declarations for integer arithmetic.
const Preamble = `type int
interpret int -> Int
relation (X:int <= Y:int)
relation (X:int < Y:int)
relation (X:int >= Y:int)
relation (X:int > Y:int)
individual (X:int + Y:int) : int
individual (X:int * Y:int) : int
individual (X:int - Y:int) : int
interpret <= -> <=
interpret < -> <
interpret >= -> >=
interpret > -> >
interpret + -> +
interpret - -> -
interpret * -> *
`

// ---------------------------------------------------------------------------
// Symbol substitution (for method parameter renaming)
// ---------------------------------------------------------------------------

// SubstMap maps original names to replacement TypedSymbols.
type SubstMap map[string]*TypedSymbol

// SubstSymbol applies a substitution map to a symbol name.
func SubstSymbol(name string, m SubstMap) string {
	if s, ok := m[name]; ok {
		return s.Rep
	}
	return name
}

// ---------------------------------------------------------------------------
// Type inference helpers
// ---------------------------------------------------------------------------

// InferType returns the type of a symbol in the current scope/module.
func (c *Compiler) InferType(rep string) (*DafnyType, error) {
	if c.ScopeCtx != nil {
		if local, ok := c.ScopeCtx.Locals[rep]; ok {
			return local.Type, nil
		}
	}
	if t, ok := c.ModCtx.Types[rep]; ok {
		// Return the first output type if available
		if len(t) > 1 && len(t[1]) > 0 {
			return t[1][0], nil
		}
	}
	return nil, fmt.Errorf("dafnygen: cannot type term %q", rep)
}

// InferInfixType returns the result type of an infix operation.
func InferInfixType(op string, argType *DafnyType) *DafnyType {
	if ArithOps[op] {
		return argType
	}
	return &DafnyType{Rep: "bool"}
}

// ReturnType returns the return types of a method.
func (c *Compiler) ReturnType(name string) []*DafnyType {
	if t, ok := c.ModCtx.Types[name]; ok && len(t) > 1 {
		return t[1]
	}
	return nil
}
