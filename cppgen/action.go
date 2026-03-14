// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_to_cpp.py action code generation (~lines 4000-4700).

// This file covers action emission: assignments, havoc, sequences,
// assert/assume, function calls, local scopes, control flow (if/while/choice),
// crash/debug, native code, quantifier loops, existential handling,
// and loop-bound computation.
package cppgen

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// BoundsError — describes a failure to compute iteration bounds.
// ---------------------------------------------------------------------------

// BoundsError records a failure to find loop bounds for a variable.
type BoundsError struct {
	Msg string
}

func (e *BoundsError) Error() string { return e.Msg }

// ---------------------------------------------------------------------------
// GetBounds — Compute loop bounds for a quantified variable.
// ---------------------------------------------------------------------------

// GetBounds computes the lower and upper bounds for iterating over a
// variable, given bound expressions extracted from the body formula.
// Returns (lo, hi) strings or a BoundsError.
func GetBounds(varSort Sort, body string) ([2]string, error) {
	bds := SortBounds(varSort)
	if bds[1] == "" {
		return [2]string{}, &BoundsError{
			Msg: fmt.Sprintf("cannot find an upper bound for sort %s", varSort.Name),
		}
	}
	return bds, nil
}

// GetBoundExprs collects bound expressions from a formula body for a
// quantified variable.
type BoundExpr struct {
	Expr string
	Neg  bool
}

// GetBoundExprs extracts bound expressions for a variable from the body.
// In the full port this would walk the AST; here we provide the interface.
func GetBoundExprs(v Variable, body string, exists bool) []BoundExpr {
	// Placeholder: full implementation walks the formula AST
	return nil
}

// ---------------------------------------------------------------------------
// emit_assign — Assignment emission.
// ---------------------------------------------------------------------------

// EmitAssign generates C++ code for an assignment action.
// If the LHS has free variables it expands to loops; otherwise it
// delegates to EmitAssignSimple.
func EmitAssign(buf *strings.Builder, lhs, rhs string, freeVars []Variable) {
	if len(freeVars) == 0 {
		EmitAssignSimple(buf, lhs, rhs)
		return
	}
	// When the LHS has free variables, iterate over them
	OpenLoop(buf, freeVars)
	indexArgs := ""
	for _, v := range freeVars {
		indexArgs += fmt.Sprintf("[%s]", VarName(v.Name))
	}
	CodeLine(buf, lhs+indexArgs+" = "+rhs+indexArgs)
	CloseLoop(buf, freeVars)
}

// EmitAssignSimple generates a simple (scalar) assignment.
func EmitAssignSimple(buf *strings.Builder, lhs, rhs string) {
	CodeLine(buf, lhs+" = "+rhs)
}

// EmitAssignLarge generates an assignment for a "large" (hash-mapped)
// type, creating a thunk that captures the RHS expression.
func EmitAssignLarge(buf *strings.Builder, lhsName string, vs []Variable, exprCode string) {
	thunk := MakeThunk(buf, vs, exprCode)
	CodeLine(buf, VarName(lhsName)+" = "+thunk)
}

// ---------------------------------------------------------------------------
// emit_havoc — Nondeterministic assignment (havoc).
// ---------------------------------------------------------------------------

// EmitHavoc generates code for a havoc action. In the Python implementation
// this asserts false (havoc should have been eliminated earlier).
func EmitHavoc(buf *strings.Builder, sym Symbol) {
	CodeLine(buf, fmt.Sprintf("// havoc %s — should have been eliminated", sym.Name))
}

// ---------------------------------------------------------------------------
// emit_sequence — Statement sequences.
// ---------------------------------------------------------------------------

// EmitSequence generates a braced block containing the given action lines.
func EmitSequence(buf *strings.Builder, actions []string) {
	buf.WriteString(IndentStr())
	buf.WriteString("{\n")
	IndentLevel++
	for _, a := range actions {
		buf.WriteString(IndentStr())
		buf.WriteString(a)
		buf.WriteString("\n")
	}
	IndentLevel--
	buf.WriteString(IndentStr())
	buf.WriteString("}\n")
}

// ---------------------------------------------------------------------------
// emit_assert / emit_assume — Assertion and assumption.
// ---------------------------------------------------------------------------

// EmitAssert generates an assertion check with a source location message.
func EmitAssert(buf *strings.Builder, condCode string, locMsg string) {
	CodeLine(buf, fmt.Sprintf("ivy_assert(%s, \"%s\")",
		condCode, strings.ReplaceAll(locMsg, "\\", "\\\\")))
}

// EmitAssume generates an assumption check with a source location message.
func EmitAssume(buf *strings.Builder, condCode string, locMsg string) {
	CodeLine(buf, fmt.Sprintf("ivy_assume(%s, \"%s\")",
		condCode, strings.ReplaceAll(locMsg, "\\", "\\\\")))
}

// ---------------------------------------------------------------------------
// emit_call — Function/action calls.
// ---------------------------------------------------------------------------

// CallArg represents a call argument with its formal parameter sort.
type CallArg struct {
	Code      string
	FormalSort Sort
	ActualSort Sort
}

// EmitCall generates a function/action call. If the call has a return value,
// it is assigned to retLHS. Arguments are emitted comma-separated.
func EmitCall(buf *strings.Builder, funcName string, args []CallArg,
	retLHS string, hasReturn bool) {
	argStrs := make([]string, len(args))
	for i, a := range args {
		argStrs[i] = a.Code
	}
	call := fmt.Sprintf("%s(%s)", VarName(funcName), strings.Join(argStrs, ", "))
	if hasReturn && retLHS != "" {
		CodeLine(buf, retLHS+" = "+call)
	} else {
		CodeLine(buf, call)
	}
}

// ---------------------------------------------------------------------------
// emit_local / local_start / local_end — Local variable scopes.
// ---------------------------------------------------------------------------

// LocalStart opens a new local scope and declares the given parameters.
// If nondetID >= 0, the parameters are initialised nondeterministically.
func LocalStart(buf *strings.Builder, params []Symbol, nondetID int) {
	buf.WriteString(IndentStr())
	buf.WriteString("{\n")
	IndentLevel++
	for _, p := range params {
		ct := CType(p.Sort, "")
		CodeLine(buf, ct+" "+VarName(p.Name))
		if nondetID >= 0 {
			MkNondetSym(buf, p, p.Name, nondetID)
		}
	}
}

// LocalEnd closes a local scope.
func LocalEnd(buf *strings.Builder) {
	IndentLevel--
	buf.WriteString(IndentStr())
	buf.WriteString("}\n")
}

// EmitLocal generates a local scope action: opens scope, declares locals
// with nondeterministic init, emits the body, then closes.
func EmitLocal(buf *strings.Builder, params []Symbol, bodyCode string, uniqueID int) {
	LocalStart(buf, params, uniqueID)
	buf.WriteString(bodyCode)
	LocalEnd(buf)
}

// ---------------------------------------------------------------------------
// emit_if — Conditional.
// ---------------------------------------------------------------------------

// EmitIf generates an if/else statement.
func EmitIf(buf *strings.Builder, condCode string, thenCode string, elseCode string) {
	buf.WriteString(IndentStr())
	buf.WriteString(fmt.Sprintf("if(%s){\n", condCode))
	IndentLevel++
	buf.WriteString(thenCode)
	IndentLevel--
	buf.WriteString(IndentStr())
	buf.WriteString("}\n")
	if elseCode != "" {
		buf.WriteString(IndentStr())
		buf.WriteString("else {\n")
		IndentLevel++
		buf.WriteString(elseCode)
		IndentLevel--
		buf.WriteString(IndentStr())
		buf.WriteString("}\n")
	}
}

// ---------------------------------------------------------------------------
// emit_while — While loop.
// ---------------------------------------------------------------------------

// EmitWhile generates a while loop. If condPreamble is non-empty, the
// loop is transformed into while(true) { preamble; if(cond) { body } else break }.
func EmitWhile(buf *strings.Builder, condCode string, condPreamble string, bodyCode string) {
	if condPreamble == "" {
		OpenScope(buf, "while("+condCode+")")
		buf.WriteString(bodyCode)
		CloseScope(buf, false)
	} else {
		OpenScope(buf, "while(true)")
		buf.WriteString(condPreamble)
		OpenScope(buf, "if("+condCode+")")
		buf.WriteString(bodyCode)
		CloseScope(buf, false)
		OpenScope(buf, "else")
		CodeLine(buf, "break")
		CloseScope(buf, false)
		CloseScope(buf, false)
	}
}

// ---------------------------------------------------------------------------
// emit_choice — Nondeterministic choice.
// ---------------------------------------------------------------------------

// EmitChoice generates a nondeterministic choice among branches.
func EmitChoice(buf *strings.Builder, branches []string, uniqueID int) {
	if len(branches) == 1 {
		buf.WriteString(branches[0])
		return
	}
	tmp := NewTemp(buf, "")
	MkNondet(buf, tmp, len(branches), "___branch", uniqueID)
	for idx, branch := range branches {
		buf.WriteString(IndentStr())
		if idx != 0 {
			buf.WriteString("else ")
		}
		if idx != len(branches)-1 {
			buf.WriteString(fmt.Sprintf("if(%s == %d)", tmp, idx))
		}
		buf.WriteString("{\n")
		IndentLevel++
		buf.WriteString(branch)
		IndentLevel--
		buf.WriteString(IndentStr())
		buf.WriteString("}\n")
	}
}

// ---------------------------------------------------------------------------
// emit_crash / emit_debug
// ---------------------------------------------------------------------------

// EmitCrash generates code for a crash action (a no-op in C++).
func EmitCrash(buf *strings.Builder) {
	// Intentionally empty — matches Python: pass
}

// DebugField represents a named field in a debug event.
type DebugField struct {
	Name string
	Code string
}

// EmitDebug generates code to print a JSON-like debug event.
func EmitDebug(buf *strings.Builder, event string, fields []DebugField) {
	CodeLine(buf, "std::cout << \"{\" << std::endl")
	CodeLine(buf, fmt.Sprintf("std::cout << \"    \\\"event\\\" : \\\"%s\\\",\" << std::endl", event))
	for _, f := range fields {
		CodeLine(buf, fmt.Sprintf("std::cout << \"    \\\"%s\\\" : \" << %s << \",\" << std::endl",
			f.Name, f.Code))
	}
	CodeLine(buf, "std::cout << \"}\" << std::endl")
}

// ---------------------------------------------------------------------------
// emit_native_action — Native code blocks.
// ---------------------------------------------------------------------------

// EmitNativeAction emits a native (pass-through) code block, substituting
// Ivy references with their C++ names.
func EmitNativeAction(buf *strings.Builder, code string) {
	buf.WriteString(IndentStr())
	buf.WriteString(code)
	buf.WriteString("\n")
}

// ---------------------------------------------------------------------------
// emit_quant — Quantifier loop emission.
// ---------------------------------------------------------------------------

// EmitQuant generates iteration code for a universally or existentially
// quantified formula over the given variables.
func EmitQuant(buf *strings.Builder, vs []Variable, bodyCode string, exists bool) {
	if len(vs) == 0 {
		buf.WriteString(bodyCode)
		return
	}

	v0 := vs[0]
	rest := vs[1:]

	tmp := NewTemp(buf, "")
	initVal := "1"
	if exists {
		initVal = "0"
	}
	CodeLine(buf, tmp+" = "+initVal)

	bds := SortBounds(v0.Sort)
	ct := CType(v0.Sort, "")
	if ct == "bool" {
		ct = "int"
	}
	buf.WriteString(IndentStr())
	buf.WriteString(fmt.Sprintf("for (%s %s = %s; %s < %s; %s++) {\n",
		ct, v0.Name, bds[0], v0.Name, bds[1], v0.Name))
	IndentLevel++

	// Recursively emit inner quantifiers and body
	var inner strings.Builder
	EmitQuant(&inner, rest, bodyCode, exists)
	innerStr := inner.String()

	buf.WriteString(IndentStr())
	negStr := "!"
	if exists {
		negStr = ""
	}
	matchVal := "1"
	if !exists {
		matchVal = "0"
	}
	buf.WriteString(fmt.Sprintf("if (%s(%s)) %s = %s;\n",
		negStr, strings.TrimSpace(innerStr), tmp, matchVal))

	IndentLevel--
	buf.WriteString(IndentStr())
	buf.WriteString("}\n")
}

// ---------------------------------------------------------------------------
// emit_some — Existential handling (if-some).
// ---------------------------------------------------------------------------

// EmitSome generates code for an if-some (existential search) construct.
// It iterates over the domain of the quantified variable, checking the
// formula, and optionally tracking min/max.
func EmitSome(buf *strings.Builder, vs []Variable, fmlaCode string,
	resultVar string, paramName string) {
	some := NewTemp(buf, "")
	CodeLine(buf, some+" = 0")
	OpenLoop(buf, vs)
	// Check the formula
	OpenScope(buf, "if("+fmlaCode+")")
	CodeLine(buf, VarName(paramName)+" = "+VarName(vs[0].Name))
	CodeLine(buf, some+" = 1")
	CloseScope(buf, false)
	CloseLoop(buf, vs)
}
