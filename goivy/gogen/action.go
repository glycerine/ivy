// Package gogen generates runnable Go programs from compiled Ivy modules.
//
// This file handles emission of Go code for each action type defined
// in the actions package.
package gogen

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
)

// ActionEmitter emits Go code for Ivy actions.
type ActionEmitter struct {
	// Gen is the parent generator (provides type helpers).
	Gen *Generator

	// w is the output writer.
	w *CodeWriter
}

// NewActionEmitter creates an emitter that writes to w.
func NewActionEmitter(gen *Generator, w *CodeWriter) *ActionEmitter {
	return &ActionEmitter{Gen: gen, w: w}
}

// EmitAction emits Go code for a single action. It dispatches on the
// concrete action type.
func (e *ActionEmitter) EmitAction(act goivy.ActionsAction) {
	if act == nil {
		e.w.Line("// nil action")
		return
	}
	switch a := act.(type) {
	case *goivy.AssignAction:
		e.emitAssign(a)
	case *goivy.Sequence:
		e.emitSequence(a)
	case *goivy.IfAction:
		e.emitIf(a)
	case *goivy.WhileAction:
		e.emitWhile(a)
	case *goivy.CallAction:
		e.emitCall(a)
	case *goivy.AssertAction:
		e.emitAssert(a)
	case *goivy.RequiresAction:
		e.emitRequire(a)
	case *goivy.EnsuresAction:
		e.emitEnsure(a)
	case *goivy.AssumeAction:
		e.emitAssume(a)
	case *goivy.HavocAction:
		e.emitHavoc(a)
	case *goivy.ChoiceAction:
		e.emitChoice(a)
	case *goivy.EnvAction:
		e.emitEnv(a)
	case *goivy.LocalAction:
		e.emitLocal(a)
	case *goivy.LetAction:
		e.emitLet(a)
	case *goivy.NativeAction:
		e.emitNative(a)
	case *goivy.CrashAction:
		e.emitCrash(a)
	case *goivy.SetAction:
		e.emitSet(a)
	case *goivy.BindOldsAction:
		e.emitBindOlds(a)
	case *goivy.ReturnAction:
		e.emitReturn(a)
	case *goivy.IgnoreAction:
		// no-op
	case *goivy.ThunkAction:
		e.emitThunk(a)
	case *goivy.SubgoalAction:
		e.emitSubgoal(a)
	case *goivy.DebugAction:
		e.emitDebug(a)
	case *goivy.AssignFieldAction:
		e.emitAssignField(a)
	case *goivy.NullFieldAction:
		e.emitNullField(a)
	case *goivy.CopyFieldAction:
		e.emitCopyField(a)
	// Note: actions.Ranking is not an Action (Args is a field, not a method)
	// so it cannot appear in this type switch.
	case *goivy.InstantiateAction:
		e.emitInstantiate(a)
	// Note: VarAction is an AST marker node (not an Action), so it cannot
	// appear here. Python: class VarAction(AST): pass
	default:
		e.w.Linef("// unhandled action type %T", act)
	}
}

// emitAssign emits: s.Field = value  (or map assignment for relations).
func (e *ActionEmitter) emitAssign(a *goivy.AssignAction) {
	lhs := e.exprString(a.LHS)
	rhs := e.exprString(a.RHS)
	e.w.Linef("%s = %s", lhs, rhs)
}

// emitSequence emits each child action in order.
func (e *ActionEmitter) emitSequence(a *goivy.Sequence) {
	for _, child := range a.Elems {
		childAct := unwrapToAction(child)
		if childAct != nil {
			e.EmitAction(childAct)
		}
	}
}

// emitIf emits: if cond { ... } else { ... }
func (e *ActionEmitter) emitIf(a *goivy.IfAction) {
	cond := e.exprString(a.Cond)
	e.w.OpenBlock(fmt.Sprintf("if %s {", cond))
	if thenAct := unwrapToAction(a.ThenBody); thenAct != nil {
		e.EmitAction(thenAct)
	}
	if a.ElseBody != nil {
		if elseAct := unwrapToAction(a.ElseBody); elseAct != nil {
			e.w.CloseBlock()
			e.w.OpenBlock("else {")
			e.EmitAction(elseAct)
		}
	}
	e.w.CloseBlock()
}

// emitWhile emits: for cond { body }
func (e *ActionEmitter) emitWhile(a *goivy.WhileAction) {
	cond := e.exprString(a.Cond)
	e.w.OpenBlock(fmt.Sprintf("for %s {", cond))
	if bodyAct := unwrapToAction(a.Body); bodyAct != nil {
		e.EmitAction(bodyAct)
	}
	// Emit invariant checks as runtime assertions inside the loop.
	for i, inv := range a.Invariants {
		invStr := e.exprString(inv)
		e.w.OpenBlock(fmt.Sprintf("if !(%s) {", invStr))
		e.w.Linef(`panic("loop invariant %d violated")`, i)
		e.w.CloseBlock()
	}
	e.w.CloseBlock()
}

// emitCall emits: s.ActionName(args...)
func (e *ActionEmitter) emitCall(a *goivy.CallAction) {
	callee := a.CalleeName()
	goName := GoExportedIdentifier(callee)
	e.w.Linef("s.%s()", goName)
}

// emitAssert emits: if !(cond) { panic("assertion failed: ...") }
func (e *ActionEmitter) emitAssert(a *goivy.AssertAction) {
	cond := e.exprString(a.Formula)
	label := "assertion"
	if a.HasLoc {
		label = a.Loc.String()
	}
	e.w.OpenBlock(fmt.Sprintf("if !(%s) {", cond))
	e.w.Linef(`panic("assertion failed: %s")`, escapeString(label))
	e.w.CloseBlock()
}

// emitRequire emits a precondition check (same shape as assert).
func (e *ActionEmitter) emitRequire(a *goivy.RequiresAction) {
	cond := e.exprString(a.Formula)
	label := "require"
	if a.HasLoc {
		label = a.Loc.String()
	}
	e.w.OpenBlock(fmt.Sprintf("if !(%s) {", cond))
	e.w.Linef(`panic("precondition failed: %s")`, escapeString(label))
	e.w.CloseBlock()
}

// emitEnsure emits a postcondition check (same shape as assert).
func (e *ActionEmitter) emitEnsure(a *goivy.EnsuresAction) {
	cond := e.exprString(a.Formula)
	label := "ensure"
	if a.HasLoc {
		label = a.Loc.String()
	}
	e.w.OpenBlock(fmt.Sprintf("if !(%s) {", cond))
	e.w.Linef(`panic("postcondition failed: %s")`, escapeString(label))
	e.w.CloseBlock()
}

// emitAssume emits a comment for assumptions (not enforced at runtime).
func (e *ActionEmitter) emitAssume(a *goivy.AssumeAction) {
	cond := e.exprString(a.Formula)
	e.w.Linef("// assume: %s", cond)
}

// emitHavoc emits a nondeterministic assignment using a zero value or
// random placeholder.
func (e *ActionEmitter) emitHavoc(a *goivy.HavocAction) {
	target := e.exprString(a.Target)
	sortStr := sortDefaultValue(a.Target.NodeSort())
	e.w.Linef("%s = %s // havoc", target, sortStr)
}

// emitChoice emits nondeterministic choice via switch rand.Intn(N).
func (e *ActionEmitter) emitChoice(a *goivy.ChoiceAction) {
	n := len(a.Branches)
	if n == 0 {
		e.w.Line("// empty choice")
		return
	}
	if n == 1 {
		if act := unwrapToAction(a.Branches[0]); act != nil {
			e.EmitAction(act)
		}
		return
	}
	e.w.OpenBlock(fmt.Sprintf("switch rand.Intn(%d) {", n))
	for i, branch := range a.Branches {
		if i < n-1 {
			e.w.Linef("case %d:", i)
		} else {
			e.w.Line("default:")
		}
		if act := unwrapToAction(branch); act != nil {
			e.EmitAction(act)
		}
	}
	e.w.CloseBlock()
}

// emitEnv emits an environment action (nondeterministic choice of exported
// actions). Same structure as choice.
func (e *ActionEmitter) emitEnv(a *goivy.EnvAction) {
	e.emitChoice(&a.ChoiceAction)
}

// emitLocal emits a block with local variable declarations.
func (e *ActionEmitter) emitLocal(a *goivy.LocalAction) {
	e.w.OpenBlock("{")
	for _, local := range a.Locals {
		name := nodeIdentName(local)
		goType := goTypeForSort(local.NodeSort())
		e.w.Linef("var %s %s", GoIdentifier(name), goType)
	}
	if bodyAct := unwrapToAction(a.Body); bodyAct != nil {
		e.EmitAction(bodyAct)
	}
	e.w.CloseBlock()
}

// emitLet emits let-bindings followed by the body.
func (e *ActionEmitter) emitLet(a *goivy.LetAction) {
	e.w.OpenBlock("{")
	for _, binding := range a.Bindings {
		name := nodeIdentName(binding)
		val := e.exprString(binding)
		e.w.Linef("%s := %s", GoIdentifier(name), val)
	}
	if bodyAct := unwrapToAction(a.Body); bodyAct != nil {
		e.EmitAction(bodyAct)
	}
	e.w.CloseBlock()
}

// emitNative emits a comment with the native code.
func (e *ActionEmitter) emitNative(a *goivy.NativeAction) {
	code := fmt.Sprint(a.Code)
	e.w.Linef("// native: %s", code)
}

// emitCrash emits panic("crash").
func (e *ActionEmitter) emitCrash(a *goivy.CrashAction) {
	e.w.Line(`panic("crash")`)
}

// emitSet emits a set operation on a relation (map assignment).
func (e *ActionEmitter) emitSet(a *goivy.SetAction) {
	lit := e.exprString(a.Lit)
	e.w.Linef("// set: %s", lit)
}

// emitBindOlds emits the inner action (old-value binding is a
// verification concept; at runtime we just execute the body).
func (e *ActionEmitter) emitBindOlds(a *goivy.BindOldsAction) {
	if inner := unwrapToAction(a.Inner); inner != nil {
		e.EmitAction(inner)
	}
}

// emitReturn emits a return statement marker.
func (e *ActionEmitter) emitReturn(a *goivy.ReturnAction) {
	e.w.Line("return")
}

// emitThunk emits a closure-based thunk.
// Python ThunkAction args: [thunkVar, name, type, body, ...].
// In Python, thunks are desugared before code generation; here we emit
// a Go closure that captures the body action.
func (e *ActionEmitter) emitThunk(a *goivy.ThunkAction) {
	args := a.ActionArgs()
	if len(args) >= 4 {
		thunkVar := e.exprString(args[0])
		name := e.exprString(args[1])
		body := e.exprString(args[3])
		e.w.Linef("// thunk %s %s := %s", thunkVar, name, body)
		e.w.Linef("%s = func() { /* %s */ }", thunkVar, body)
	} else {
		e.w.Linef("// thunk (insufficient args: %d)", len(args))
	}
}

// emitSubgoal emits a subgoal assertion.
func (e *ActionEmitter) emitSubgoal(a *goivy.SubgoalAction) {
	fmla := e.exprString(a.Formula)
	e.w.Linef("ivy_assert(%s, %q)", fmla, "subgoal")
}

// emitDebug emits a debug print statement.
func (e *ActionEmitter) emitDebug(a *goivy.DebugAction) {
	e.w.Linef("// debug: %s", a.String())
}

// emitAssignField emits field assignment: obj.field = value.
func (e *ActionEmitter) emitAssignField(a *goivy.AssignFieldAction) {
	e.w.Linef("%s.%s = %s", e.exprString(a.Obj), e.exprString(a.Field), e.exprString(a.Value))
}

// emitNullField emits field nullification: obj.field = nil.
func (e *ActionEmitter) emitNullField(a *goivy.NullFieldAction) {
	e.w.Linef("%s.%s = nil", e.exprString(a.Obj), e.exprString(a.Field))
}

// emitCopyField emits field copy: dst.field = src.field.
func (e *ActionEmitter) emitCopyField(a *goivy.CopyFieldAction) {
	dstField := e.exprString(a.Field)
	srcField := dstField
	if a.SrcField != nil {
		srcField = e.exprString(a.SrcField)
	}
	e.w.Linef("%s.%s = %s.%s", e.exprString(a.Dst), dstField, e.exprString(a.Src), srcField)
}

// emitInstantiate emits schema instantiation (placeholder — Python
// desugars instantiation before code generation).
func (e *ActionEmitter) emitInstantiate(a *goivy.InstantiateAction) {
	e.w.Linef("// instantiate: %s", a.String())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// exprString converts an lg.Expr to its Go expression string.
// This is a simplified version; a full implementation would walk the AST.
func (e *ActionEmitter) exprString(n goivy.Expr) string {
	if n == nil {
		return "nil"
	}
	return ExprToGo(n)
}

// ExprToGo converts a logic node to a Go expression string.
func ExprToGo(n goivy.Expr) string {
	if n == nil {
		return "nil"
	}
	switch v := n.(type) {
	case *goivy.Const:
		return GoIdentifier(v.Name)
	case *goivy.Variable:
		return GoIdentifier(v.Name)
	case *goivy.Apply:
		fname := ExprToGo(v.Func)
		args := make([]string, len(v.Terms))
		for i, t := range v.Terms {
			args[i] = ExprToGo(t)
		}
		return fmt.Sprintf("%s(%s)", fname, strings.Join(args, ", "))
	case *goivy.Eq:
		return fmt.Sprintf("(%s == %s)", ExprToGo(v.T1), ExprToGo(v.T2))
	case *goivy.Not:
		return fmt.Sprintf("!(%s)", ExprToGo(v.Body))
	case *goivy.And:
		if len(v.Terms) == 0 {
			return "true"
		}
		parts := make([]string, len(v.Terms))
		for i, t := range v.Terms {
			parts[i] = ExprToGo(t)
		}
		return "(" + strings.Join(parts, " && ") + ")"
	case *goivy.Or:
		if len(v.Terms) == 0 {
			return "false"
		}
		parts := make([]string, len(v.Terms))
		for i, t := range v.Terms {
			parts[i] = ExprToGo(t)
		}
		return "(" + strings.Join(parts, " || ") + ")"
	case *goivy.Implies:
		return fmt.Sprintf("(!(%s) || (%s))", ExprToGo(v.T1), ExprToGo(v.T2))
	case *goivy.ForAll:
		// Placeholder: forAll requires a helper function
		varNames := make([]string, len(v.Variables))
		for i, va := range v.Variables {
			varNames[i] = va.Name
		}
		return fmt.Sprintf("forAll(/* %s */ func() bool { return %s })",
			strings.Join(varNames, ", "), ExprToGo(v.Body))
	case *goivy.Exists:
		varNames := make([]string, len(v.Variables))
		for i, va := range v.Variables {
			varNames[i] = va.Name
		}
		return fmt.Sprintf("exists(/* %s */ func() bool { return %s })",
			strings.Join(varNames, ", "), ExprToGo(v.Body))
	default:
		// Fall back to String() for types we don't handle specifically.
		return fmt.Sprintf("%v", n)
	}
}

// GoIdentifier converts an Ivy dotted name to a valid Go identifier.
// Dots are replaced with underscores, and the first letter is uppercased
// for exported identifiers.
func GoIdentifier(name string) string {
	if name == "" {
		return "_"
	}
	// Replace dots and colons with underscores.
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, ":", "_")
	// Ensure first character is valid.
	if name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}
	return name
}

// GoExportedIdentifier converts a name to an exported Go identifier.
func GoExportedIdentifier(name string) string {
	id := GoIdentifier(name)
	if len(id) == 0 {
		return "_"
	}
	if id[0] == '_' && len(id) > 1 {
		return strings.ToUpper(id[1:2]) + id[2:]
	}
	return strings.ToUpper(id[0:1]) + id[1:]
}

// goTypeForSort returns a Go type string for a logic sort.
func goTypeForSort(s goivy.Sort) string {
	if s == nil {
		return "interface{}"
	}
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return "bool"
	case *goivy.UninterpretedSort:
		return GoExportedIdentifier(st.Name)
	case *goivy.EnumeratedSort:
		return GoExportedIdentifier(st.Name)
	case *goivy.RangeSort:
		return "int"
	case *goivy.FunctionSort:
		// Functions/relations become maps.
		dom := st.Domain()
		rng := st.Range()
		if len(dom) == 1 {
			return fmt.Sprintf("map[%s]%s", goTypeForSort(dom[0]), goTypeForSort(rng))
		}
		// Multi-argument: use a tuple key.
		keyParts := make([]string, len(dom))
		for i, d := range dom {
			keyParts[i] = goTypeForSort(d)
		}
		return fmt.Sprintf("map[[%d]interface{}]%s", len(dom), goTypeForSort(rng))
	default:
		return "interface{}"
	}
}

// sortDefaultValue returns a Go expression for the zero/default value
// of a given sort.
func sortDefaultValue(s goivy.Sort) string {
	if s == nil {
		return "nil"
	}
	switch s.(type) {
	case *goivy.BooleanSort:
		return "false"
	case *goivy.UninterpretedSort:
		return "0"
	case *goivy.EnumeratedSort:
		return "0"
	case *goivy.RangeSort:
		return "0"
	default:
		return "nil"
	}
}

// unwrapToAction converts an lg.Expr to an Action via type assertion.
func unwrapToAction(n goivy.Expr) goivy.ActionsAction {
	if n == nil {
		return nil
	}
	act, _ := n.(goivy.ActionsAction)
	return act
}

// nodeIdentName extracts a name from a node (Const or Var).
func nodeIdentName(n goivy.Expr) string {
	switch v := n.(type) {
	case *goivy.Const:
		return v.Name
	case *goivy.Variable:
		return v.Name
	default:
		return fmt.Sprint(n)
	}
}

// escapeString escapes special characters for Go string literals.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
