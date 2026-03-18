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

	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
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

// GetBounds computes the lower and upper bounds for iterating over
// a variable. Returns (lo, hi) strings or a BoundsError.
func GetBounds(ctx *CppGenContext, varSort lg.Sort) ([2]string, error) {
	bds := sortBoundsStr(ctx, varSort)
	if bds[1] == "" {
		return [2]string{}, &BoundsError{
			Msg: fmt.Sprintf("cannot find an upper bound for sort %s", il.SortName(varSort)),
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

// GetBoundExprsFromBody extracts bound expressions for a variable.
// Placeholder: full implementation walks the formula AST.
func GetBoundExprsFromBody(v *lg.Variable, body string, exists bool) []BoundExpr {
	return nil
}

// ---------------------------------------------------------------------------
// emit_assign — Assignment emission.
// ---------------------------------------------------------------------------

// EmitAssign generates C++ code for an assignment action.
func EmitAssign(ctx *CppGenContext, buf *CodeText, lhs, rhs string, freeVars []*lg.Variable) {
	if len(freeVars) == 0 {
		EmitAssignSimple(buf, lhs, rhs)
		return
	}
	openLoop(ctx, buf, freeVars)
	indexArgs := ""
	for _, v := range freeVars {
		indexArgs += fmt.Sprintf("[%s]", Varname(v.Name))
	}
	codeLine(buf, lhs+indexArgs+" = "+rhs+indexArgs)
	closeLoop(buf, freeVars)
}

// EmitAssignSimple generates a simple (scalar) assignment.
func EmitAssignSimple(buf *CodeText, lhs, rhs string) {
	codeLine(buf, lhs+" = "+rhs)
}

// EmitAssignLarge generates an assignment for a "large" (hash-mapped) type.
func EmitAssignLarge(ctx *CppGenContext, buf *CodeText, lhsName string,
	vs []*lg.Variable, exprCode string) {
	thunk := MakeThunk(ctx, buf, vs, exprCode)
	codeLine(buf, Varname(lhsName)+" = "+thunk)
}

// ---------------------------------------------------------------------------
// emit_havoc — Nondeterministic assignment (havoc).
// ---------------------------------------------------------------------------

// EmitHavoc generates code for a havoc action.
func EmitHavoc(buf *CodeText, sym *lg.Symbol) {
	codeLine(buf, fmt.Sprintf("// havoc %s — should have been eliminated", sym.Name))
}

// ---------------------------------------------------------------------------
// emit_sequence — Statement sequences.
// ---------------------------------------------------------------------------

// EmitSequence generates a braced block containing the given action lines.
func EmitSequence(buf *CodeText, actions []string) {
	Indent(buf)
	buf.Append("{\n")
	IndentLevel++
	for _, a := range actions {
		Indent(buf)
		buf.Append(a + "\n")
	}
	IndentLevel--
	Indent(buf)
	buf.Append("}\n")
}

// ---------------------------------------------------------------------------
// emit_assert / emit_assume — Assertion and assumption.
// ---------------------------------------------------------------------------

// EmitAssert generates an assertion check.
func EmitAssert(buf *CodeText, condCode string, locMsg string) {
	codeLine(buf, fmt.Sprintf("ivy_assert(%s, \"%s\")",
		condCode, strings.ReplaceAll(locMsg, "\\", "\\\\")))
}

// EmitAssume generates an assumption check.
func EmitAssume(buf *CodeText, condCode string, locMsg string) {
	codeLine(buf, fmt.Sprintf("ivy_assume(%s, \"%s\")",
		condCode, strings.ReplaceAll(locMsg, "\\", "\\\\")))
}

// ---------------------------------------------------------------------------
// emit_call — Function/action calls.
// ---------------------------------------------------------------------------

// CallArg represents a call argument.
type CallArg struct {
	Code       string
	FormalSort lg.Sort
	ActualSort lg.Sort
}

// EmitCall generates a function/action call.
func EmitCall(buf *CodeText, funcName string, args []CallArg,
	retLHS string, hasReturn bool) {
	argStrs := make([]string, len(args))
	for i, a := range args {
		argStrs[i] = a.Code
	}
	call := fmt.Sprintf("%s(%s)", Varname(funcName), strings.Join(argStrs, ", "))
	if hasReturn && retLHS != "" {
		codeLine(buf, retLHS+" = "+call)
	} else {
		codeLine(buf, call)
	}
}

// ---------------------------------------------------------------------------
// emit_local / local_start / local_end — Local variable scopes.
// ---------------------------------------------------------------------------

// LocalStart opens a new local scope and declares the given parameters.
func LocalStart(ctx *CppGenContext, buf *CodeText, params []*lg.Symbol, nondetID int) {
	Indent(buf)
	buf.Append("{\n")
	IndentLevel++
	for _, p := range params {
		ct := CTypeFull(ctx, p.CSort, "")
		codeLine(buf, ct+" "+Varname(p.Name))
		if nondetID >= 0 {
			MkNondetSym(ctx, buf, p, p.Name, nondetID)
		}
	}
}

// LocalEnd closes a local scope.
func LocalEnd(buf *CodeText) {
	IndentLevel--
	Indent(buf)
	buf.Append("}\n")
}

// EmitLocal generates a local scope action.
func EmitLocal(ctx *CppGenContext, buf *CodeText, params []*lg.Symbol, bodyCode string, uniqueID int) {
	LocalStart(ctx, buf, params, uniqueID)
	buf.Append(bodyCode)
	LocalEnd(buf)
}

// ---------------------------------------------------------------------------
// emit_if — Conditional.
// ---------------------------------------------------------------------------

// EmitIf generates an if/else statement.
func EmitIf(buf *CodeText, condCode string, thenCode string, elseCode string) {
	Indent(buf)
	buf.Append(fmt.Sprintf("if(%s){\n", condCode))
	IndentLevel++
	buf.Append(thenCode)
	IndentLevel--
	Indent(buf)
	buf.Append("}\n")
	if elseCode != "" {
		Indent(buf)
		buf.Append("else {\n")
		IndentLevel++
		buf.Append(elseCode)
		IndentLevel--
		Indent(buf)
		buf.Append("}\n")
	}
}

// ---------------------------------------------------------------------------
// emit_while — While loop.
// ---------------------------------------------------------------------------

// EmitWhile generates a while loop.
func EmitWhile(buf *CodeText, condCode string, condPreamble string, bodyCode string) {
	if condPreamble == "" {
		openScope(buf, "while("+condCode+")")
		buf.Append(bodyCode)
		closeScope(buf, false)
	} else {
		openScope(buf, "while(true)")
		buf.Append(condPreamble)
		openScope(buf, "if("+condCode+")")
		buf.Append(bodyCode)
		closeScope(buf, false)
		openScope(buf, "else")
		codeLine(buf, "break")
		closeScope(buf, false)
		closeScope(buf, false)
	}
}

// ---------------------------------------------------------------------------
// emit_choice — Nondeterministic choice.
// ---------------------------------------------------------------------------

// EmitChoice generates a nondeterministic choice among branches.
func EmitChoice(buf *CodeText, branches []string, uniqueID int) {
	if len(branches) == 1 {
		buf.Append(branches[0])
		return
	}
	tmp := NewTemp(buf, "")
	MkNondet(buf, tmp, len(branches), "___branch", uniqueID)
	for idx, branch := range branches {
		Indent(buf)
		if idx != 0 {
			buf.Append("else ")
		}
		if idx != len(branches)-1 {
			buf.Append(fmt.Sprintf("if(%s == %d)", tmp, idx))
		}
		buf.Append("{\n")
		IndentLevel++
		buf.Append(branch)
		IndentLevel--
		Indent(buf)
		buf.Append("}\n")
	}
}

// ---------------------------------------------------------------------------
// emit_crash / emit_debug
// ---------------------------------------------------------------------------

// EmitCrash generates code for a crash action (no-op).
func EmitCrash(buf *CodeText) {
	// Intentionally empty — matches Python
}

// DebugField represents a named field in a debug event.
type DebugField struct {
	Name string
	Code string
}

// EmitDebug generates code to print a JSON-like debug event.
func EmitDebug(buf *CodeText, event string, fields []DebugField) {
	codeLine(buf, "std::cout << \"{\" << std::endl")
	codeLine(buf, fmt.Sprintf("std::cout << \"    \\\"event\\\" : \\\"%s\\\",\" << std::endl", event))
	for _, f := range fields {
		codeLine(buf, fmt.Sprintf("std::cout << \"    \\\"%s\\\" : \" << %s << \",\" << std::endl",
			f.Name, f.Code))
	}
	codeLine(buf, "std::cout << \"}\" << std::endl")
}

// ---------------------------------------------------------------------------
// emit_native_action — Native code blocks.
// ---------------------------------------------------------------------------

// EmitNativeAction emits a native (pass-through) code block.
func EmitNativeAction(buf *CodeText, code string) {
	Indent(buf)
	buf.Append(code + "\n")
}

// ---------------------------------------------------------------------------
// emit_quant — Quantifier loop emission.
// ---------------------------------------------------------------------------

// EmitQuant generates iteration code for a quantified formula.
func EmitQuant(ctx *CppGenContext, buf *CodeText, vs []*lg.Variable, bodyCode string, exists bool) {
	if len(vs) == 0 {
		buf.Append(bodyCode)
		return
	}

	v0 := vs[0]
	rest := vs[1:]

	tmp := NewTemp(buf, "")
	initVal := "1"
	if exists {
		initVal = "0"
	}
	codeLine(buf, tmp+" = "+initVal)

	bds := sortBoundsStr(ctx, v0.VSort)
	ct := CTypeFull(ctx, v0.VSort, "")
	if ct == "bool" {
		ct = "int"
	}
	Indent(buf)
	buf.Append(fmt.Sprintf("for (%s %s = %s; %s < %s; %s++) {\n",
		ct, v0.Name, bds[0], v0.Name, bds[1], v0.Name))
	IndentLevel++

	var inner CodeText
	EmitQuant(ctx, &inner, rest, bodyCode, exists)
	innerStr := inner.String()

	negStr := "!"
	if exists {
		negStr = ""
	}
	matchVal := "0"
	if exists {
		matchVal = "1"
	}
	Indent(buf)
	buf.Append(fmt.Sprintf("if (%s(%s)) %s = %s;\n",
		negStr, strings.TrimSpace(innerStr), tmp, matchVal))

	IndentLevel--
	Indent(buf)
	buf.Append("}\n")
}

// ---------------------------------------------------------------------------
// emit_some — Existential handling (if-some).
// ---------------------------------------------------------------------------

// EmitSome generates code for an if-some (existential search) construct.
func EmitSome(ctx *CppGenContext, buf *CodeText, vs []*lg.Variable, fmlaCode string,
	resultVar string, paramName string) {
	some := NewTemp(buf, "")
	codeLine(buf, some+" = 0")
	openLoop(ctx, buf, vs)
	openScope(buf, "if("+fmlaCode+")")
	codeLine(buf, Varname(paramName)+" = "+Varname(vs[0].Name))
	codeLine(buf, some+" = 1")
	closeScope(buf, false)
	closeLoop(buf, vs)
}
