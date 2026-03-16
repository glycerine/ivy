// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_to_cpp.py action generator code (~lines 1192-1370, 1587-1636).

// This file covers action generator emission (emit_action_gen), action
// method emission (emit_some_action), derived predicate emission
// (emit_derived), and constructor emission (emit_constructor).
package cppgen

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/actions"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// ---------------------------------------------------------------------------
// EmitActionGen — Generate an action generator class.
// ---------------------------------------------------------------------------

// EmitActionGen generates a C++ generator class for a named action.
// The generator class encapsulates:
//   - Member variables for all input parameters and local state
//   - A constructor that sets up Z3 constraints (the precondition)
//   - A generate() method that solves the constraints and assigns values
//   - An execute() method that calls the actual action with generated values
//
// This corresponds to Python emit_action_gen() (lines 1192-1343).
func EmitActionGen(ctx *CppGenContext, header, impl *CodeText, name string, action actions.Action, classname string, mod *module.Module) {
	caname := Varname(name)

	// If there's a before_export override, use it
	if mod != nil && mod.BeforeExport != nil {
		if be, ok := mod.BeforeExport[name]; ok {
			if a, ok := be.(actions.Action); ok {
				action = a
			}
		}
	}

	// Handle external preconditions
	if mod != nil && mod.ExtPreconds != nil {
		if preCond, ok := mod.ExtPreconds[name]; ok {
			origAction := action
			assumeAct := actions.NewAssumeAction(preCond)
			seqAct := actions.NewSequence(actions.WrapAction(assumeAct), actions.WrapAction(origAction))
			seqAct.SetLineno(origAction.GetLineno())
			seqAct.SetFormalParams(origAction.GetFormalParams())
			seqAct.SetFormalReturns(origAction.GetFormalReturns())
			action = seqAct
		}
	}

	// Collect input symbols from formal parameters
	inputs := make([]*lg.Const, 0)
	inputSet := make(map[string]bool)
	for _, p := range action.GetFormalParams() {
		prefixedName := "__" + p.Name
		sym := lg.NewConst(prefixedName, p.CSort)
		if !inputSet[prefixedName] {
			inputs = append(inputs, sym)
			inputSet[prefixedName] = true
		}
	}

	// Generate the header class declaration
	header.Append(fmt.Sprintf("class %s_gen : public gen {\n  public:\n", caname))

	// Declare member variables for input symbols
	for _, sym := range inputs {
		if !strings.HasPrefix(sym.Name, "__ts") {
			declareGenSymbol(ctx, header, sym, classname)
		}
	}

	// Constructor, generate, and execute declarations
	header.Append(fmt.Sprintf("    %s_gen(%s&);\n", caname, classname))
	header.Append(fmt.Sprintf("    bool generate(%s&);\n", classname))
	header.Append(fmt.Sprintf("    void execute(%s&);\n};\n", classname))

	// Generate the constructor implementation
	impl.Append(fmt.Sprintf("%s_gen::%s_gen(%s &obj){\n", caname, caname, classname))
	IndentLevel++

	// Emit signature setup
	EmitSig(ctx, impl)

	// Declare input symbols in the solver
	for _, sym := range inputs {
		emitGenDecl(ctx, impl, sym)
	}

	// Add the precondition constraint
	// In the full implementation, this would compute the reverse image
	// and add it as a Z3 assertion. For now, emit a placeholder.
	Indent(impl)
	impl.Append("// Precondition constraint would be added here via Z3\n")

	IndentLevel--
	impl.Append("}\n")

	// Generate the generate() method
	impl.Append(fmt.Sprintf("bool %s_gen::generate(%s& obj) {\n    push();\n", caname, classname))
	IndentLevel++

	// Set state symbols in solver
	Indent(impl)
	impl.Append("alits.clear();\n")

	// Randomize input symbols
	for _, sym := range inputs {
		if !strings.HasPrefix(sym.Name, "__ts") {
			EmitRandomize(ctx, impl, sym, classname)
		}
	}

	// Solve and evaluate
	impl.Append(`
    // std::cout << slvr << std::endl;
    bool __res = solve();
    if (__res) {
`)
	IndentLevel++

	// Evaluate solved values
	for _, sym := range inputs {
		if !strings.HasPrefix(sym.Name, "__ts") {
			emitGenEval(ctx, impl, sym, classname)
		}
	}

	IndentLevel -= 2
	impl.Append(`
    }
    pop();
    obj.___ivy_gen = this;
    return __res;
}
`)

	// Generate the execute() method
	openScope(impl, fmt.Sprintf("void %s_gen::execute(%s& obj)", caname, classname))

	// Print the action name and parameters
	paramParts := make([]string, 0)
	for _, p := range action.GetFormalParams() {
		paramParts = append(paramParts, fmt.Sprintf(" << %s", Varname(p.Name)))
	}
	if len(paramParts) > 0 {
		codeLine(impl, fmt.Sprintf("__ivy_out << \"> %s(\" %s << \")\" << std::endl",
			strings.Split(name, ":")[len(strings.Split(name, ":"))-1],
			strings.Join(paramParts, " << \",\"")))
	} else {
		codeLine(impl, fmt.Sprintf("__ivy_out << \"> %s\" << std::endl",
			strings.Split(name, ":")[len(strings.Split(name, ":"))-1]))
	}

	// Generate the call
	callParams := make([]string, 0)
	for _, p := range action.GetFormalParams() {
		callParams = append(callParams, Varname(p.Name))
	}
	call := fmt.Sprintf("obj.%s(%s)", caname, strings.Join(callParams, ","))

	returns := action.GetFormalReturns()
	if len(returns) == 0 {
		codeLine(impl, call)
	} else {
		codeLine(impl, fmt.Sprintf("__ivy_out << \"= \" << %s << std::endl", call))
	}

	closeScope(impl, false)
}

// ---------------------------------------------------------------------------
// EmitSomeAction — Emit an action as a C++ method.
// ---------------------------------------------------------------------------

// EmitSomeAction generates a C++ method definition for an action.
// This includes parameter declarations, return value handling,
// and the action body emission.
//
// Corresponds to Python emit_some_action() (lines 1587-1620).
func EmitSomeAction(ctx *CppGenContext, header, impl *CodeText, name string, action actions.Action, classname string, inline bool) {
	// Emit method declaration in header
	if !inline {
		EmitMethodDecl(header, name, action, false, classname, false)
		header.Append(";\n")
	}

	// Emit method definition in impl
	var code CodeText
	EmitMethodDecl(&code, name, action, true, classname, inline)
	code.Append("{\n")
	IndentLevel++

	// Declare return value if needed
	returns := action.GetFormalReturns()
	params := action.GetFormalParams()
	if len(returns) >= 1 {
		retParam := returns[0]
		// Check if return param is also an input param
		isAlsoParam := false
		for _, p := range params {
			if p.Name == retParam.Name {
				isAlsoParam = true
				break
			}
		}
		if !isAlsoParam {
			Indent(&code)
			code.Append(fmt.Sprintf("%s %s;\n",
				CTypeFull(ctx, retParam.CSort, classname),
				Varname(retParam.Name)))
		}
	}

	// Emit action body
	// In the full implementation, this would call action.emit(code).
	// For now, emit a placeholder comment.
	Indent(&code)
	code.Append("// Action body would be emitted here\n")

	// Return the result
	if len(returns) >= 1 {
		Indent(&code)
		code.Append(fmt.Sprintf("return %s;\n", Varname(returns[0].Name)))
	}

	IndentLevel--
	code.Append("}\n")
	impl.Append(code.String())
}

// EmitMethodDecl generates a C++ method declaration/definition header.
func EmitMethodDecl(buf *CodeText, name string, action actions.Action, body bool, classname string, inline bool) {
	returns := action.GetFormalReturns()
	params := action.GetFormalParams()

	// Return type
	retType := "void"
	if len(returns) >= 1 {
		retType = CTypeFull(nil, returns[0].CSort, classname)
	}

	// Method name
	methodName := Varname(name)
	if body && classname != "" && !inline {
		methodName = classname + "::" + methodName
	}

	// Parameter list
	paramParts := make([]string, 0)
	for _, p := range params {
		paramParts = append(paramParts, fmt.Sprintf("%s %s",
			CTypeFull(nil, p.CSort, classname), Varname(p.Name)))
	}

	if inline {
		Indent(buf)
		buf.Append("inline ")
	}
	buf.Append(fmt.Sprintf("%s %s(%s)", retType, methodName, strings.Join(paramParts, ", ")))
}

// ---------------------------------------------------------------------------
// EmitDerived — Emit a derived predicate/function as a C++ method.
// ---------------------------------------------------------------------------

// EmitDerived generates code for a derived predicate or function.
// A derived symbol is defined by an equation: f(x,y) = expr(x,y).
// This is compiled to a C++ method that computes the expression.
//
// Corresponds to Python emit_derived() (lines 1346-1357).
func EmitDerived(ctx *CppGenContext, header, impl *CodeText, defn lg.Node, classname string, inline bool) {
	// Extract the definition components
	type definer interface {
		Defines() lg.Node
	}
	d, ok := defn.(definer)
	if !ok {
		return
	}
	definesSym := d.Defines()
	if definesSym == nil {
		return
	}
	sym, ok := definesSym.(*lg.Const)
	if !ok {
		return
	}

	name := sym.Name
	rng := il.SortRange(sym.CSort)

	// Create a return value symbol
	retval := lg.NewConst("ret:val", rng)

	// Create formal parameters from the definition's variables
	// For a definition f(X,Y) = expr, X and Y become formal params
	dom := il.SortDomain(sym.CSort)
	formalParams := make([]*lg.Const, len(dom))
	for i, s := range dom {
		formalParams[i] = lg.NewConst(fmt.Sprintf("fml:X%d", i), s)
	}

	// Create an assign action: retval := rhs
	// In the full implementation, we'd substitute variables for formals in the RHS.
	// For now, create the action structure.
	assignAct := actions.NewAssignAction(retval, lg.True) // placeholder RHS
	assignAct.SetFormalParams(formalParams)
	assignAct.SetFormalReturns([]*lg.Const{retval})

	EmitSomeAction(ctx, header, impl, name, assignAct, classname, inline)
}

// ---------------------------------------------------------------------------
// EmitConstructor — Emit a constructor for a structured type.
// ---------------------------------------------------------------------------

// EmitConstructor generates code for a type constructor.
// A constructor creates a value of a structured type by assigning
// to each destructor field.
//
// Corresponds to Python emit_constructor() (lines 1359-1370).
func EmitConstructor(ctx *CppGenContext, header, impl *CodeText, cons *lg.Const, classname string, mod *module.Module, inline bool) {
	name := cons.Name
	rng := il.SortRange(cons.CSort)
	retval := lg.NewConst("ret:val", rng)

	// Create formal parameters from the constructor's domain sorts
	dom := il.SortDomain(cons.CSort)
	formalParams := make([]*lg.Const, len(dom))
	for i, s := range dom {
		formalParams[i] = lg.NewConst(fmt.Sprintf("fml:X%d", i), s)
	}

	// Build assignment actions for each destructor
	rngName := il.SortName(rng)
	var assignNodes []lg.Node

	if mod != nil && mod.SortDestructors != nil {
		if destrs, ok := mod.SortDestructors[rngName]; ok {
			for i, d := range destrs {
				if i < len(formalParams) {
					// d(retval) := param_i
					lhs := &lg.Apply{Func: d, Terms: []lg.Node{retval}}
					rhs := formalParams[i]
					asgn := actions.NewAssignAction(lhs, rhs)
					assignNodes = append(assignNodes, actions.WrapAction(asgn))
				}
			}
		}
	}

	seqAct := actions.NewSequence(assignNodes...)
	seqAct.SetFormalParams(formalParams)
	seqAct.SetFormalReturns([]*lg.Const{retval})

	EmitSomeAction(ctx, header, impl, name, seqAct, classname, inline)
}

// ---------------------------------------------------------------------------
// EmitNative — Emit user-provided native code blocks.
// ---------------------------------------------------------------------------

// EmitNative emits a native code block into the output.
// This handles the native { ... } syntax in Ivy that allows embedding
// target-language code directly.
//
// Corresponds to Python emit_native() (lines 1451-1453).
func EmitNative(impl *CodeText, code string, params []lg.Node) {
	// In the full implementation, we'd substitute parameter references
	// in the native code string. For now, emit directly.
	Indent(impl)
	impl.Append(code + "\n")
}

// ---------------------------------------------------------------------------
// EmitInitialAction — Emit the __init() method.
// ---------------------------------------------------------------------------

// EmitInitialAction generates the __init() method that runs all
// initial actions from the module.
//
// Corresponds to Python emit_initial_action() (lines 1637-1646).
func EmitInitialAction(ctx *CppGenContext, header, impl *CodeText, classname string, mod *module.Module) {
	codeLine(header, "void __init()")
	openScope(impl, fmt.Sprintf("void %s::__init()", classname))

	if mod != nil {
		for _, initAct := range mod.InitialActions {
			if act, ok := initAct.(actions.Action); ok {
				// In the full implementation, we'd emit the action body.
				// For now, emit a placeholder.
				_ = act
				Indent(impl)
				impl.Append("// Initial action body would be emitted here\n")
			}
		}
	}

	closeScope(impl, false)
}

// ---------------------------------------------------------------------------
// Helpers for action gen — wrappers around existing functions.
// ---------------------------------------------------------------------------

// declareGenSymbol emits a C++ variable declaration for a generator symbol.
func declareGenSymbol(ctx *CppGenContext, buf *CodeText, sym *lg.Const, classname string) {
	DeclareSymbol(ctx, buf, sym.Name, sym.CSort, "", 0, classname, false, "")
}

// emitGenDecl emits a solver variable declaration for a symbol.
func emitGenDecl(ctx *CppGenContext, impl *CodeText, sym *lg.Const) {
	Indent(impl)
	impl.Append(fmt.Sprintf("mk_const(\"%s\", %s);\n",
		sym.Name, genSortToSMT(sym.CSort)))
}

// emitGenEval emits code to evaluate a solver variable and assign it.
func emitGenEval(ctx *CppGenContext, impl *CodeText, sym *lg.Const, classname string) {
	EmitEval(ctx, impl, sym, "obj", classname)
}

// genSortToSMT returns the SMT-LIB sort name for a sort.
func genSortToSMT(s lg.Sort) string {
	if s == nil {
		return "S"
	}
	switch t := s.(type) {
	case *lg.BooleanSort:
		return "Bool"
	case *lg.UninterpretedSort:
		return t.Name
	case *lg.EnumeratedSort:
		return t.Name
	case *lg.RangeSort:
		return "Int"
	default:
		return "S"
	}
}
