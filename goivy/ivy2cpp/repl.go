package ivy2cpp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// enumSortsForArgSpecs returns the named, non-numeric, non-encoded enum
// sorts in module sort order. Mirrors Python ivy_to_cpp.py:2419 +
// per-loop filter `sort_name not in encoded_sorts`. These are the enums
// for which Python emits `operator<<`, `_arg<T>`, `__ser<T>`,
// `__deser<T>` (and, for test/gen, `__from_solver`/`__to_solver`/
// `__randomize`).
func (g *Generator) enumSortsForArgSpecs() []*goivy.LogicEnumeratedSort {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	encoded := g.encodedSortSet()
	var out []*goivy.LogicEnumeratedSort
	for _, name := range g.Mod.SortOrder {
		if encoded != nil && encoded[name] {
			continue
		}
		if !g.sortNeededForRuntimeSpecs(name) {
			continue
		}
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		st, ok := s.(*goivy.LogicEnumeratedSort)
		if !ok {
			continue
		}
		if st.Name == "" || len(st.Extension) == 0 {
			continue
		}
		if isNumericEnum(st) {
			continue
		}
		out = append(out, st)
	}
	return out
}

// encodedSortSet returns the encoded-sort set for filtering, or nil if
// none. The accessor lives on Generator via the lazy initializer used
// by native.go.
func (g *Generator) encodedSortSet() map[string]bool {
	if g == nil {
		return nil
	}
	return g.encodedSorts
}

// emitEnumSortArgSpecDecls emits forward declarations of the per-enum
// `operator<<`, `_arg<T>`, `__ser<T>`, `__deser<T>` symbols. Mirrors
// Python ivy_to_cpp.py:2213-2223 — these declarations are placed right
// after `#include "ivy_value.hpp"` so the rest of the impl file can
// resolve them. For gen/test targets, also emits the Z3 solver
// specialization forward declarations (Python lines 2224-2230).
func (g *Generator) emitEnumSortArgSpecDecls(w *cppWriter) {
	enums := g.enumSortsForArgSpecs()
	gateZ3 := g.usesZ3() && len(enums) > 0
	for _, st := range enums {
		cfsname := g.ClassName + "::" + varName(st.Name)
		if g.Config.Target == "test" {
			w.linef("std::ostream &operator <<(std::ostream &s, const %s &t);", cfsname)
		} else {
			w.linef("std::ostream &operator<<(std::ostream &s, const %s &t);", cfsname)
		}
		w.line("template <>")
		w.linef("%s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound);", cfsname, cfsname)
		w.line("template <>")
		if g.Config.Target == "test" {
			w.linef("void  __ser<%s>(ivy_ser &res, const %s&);", cfsname, cfsname)
		} else {
			w.linef("void __ser<%s>(ivy_ser &res, const %s &);", cfsname, cfsname)
		}
		w.line("template <>")
		if g.Config.Target == "test" {
			w.linef("void  __deser<%s>(ivy_deser &inp, %s &res);", cfsname, cfsname)
		} else {
			w.linef("void __deser<%s>(ivy_deser &inp, %s &res);", cfsname, cfsname)
		}
	}
	if gateZ3 {
		if g.Config.Target != "test" {
			w.line("#ifdef Z3PP_H_")
		}
		for _, st := range enums {
			cfsname := g.ClassName + "::" + varName(st.Name)
			w.line("template <>")
			if g.Config.Target == "test" {
				w.linef("void __from_solver<%s>( gen &g, const  z3::expr &v, %s &res);", cfsname, cfsname)
			} else {
				w.linef("void __from_solver<%s>(gen &g, const z3::expr &v, %s &res);", cfsname, cfsname)
			}
			w.line("template <>")
			if g.Config.Target == "test" {
				w.linef("z3::expr __to_solver<%s>( gen &g, const  z3::expr &v, %s &val);", cfsname, cfsname)
			} else {
				w.linef("z3::expr __to_solver<%s>(gen &g, const z3::expr &v, const %s &val);", cfsname, cfsname)
			}
			w.line("template <>")
			if g.Config.Target == "test" {
				w.linef("void __randomize<%s>( gen &g, const  z3::expr &v, const std::string &sort_name);", cfsname)
			} else {
				w.linef("void __randomize<%s>(gen &g, const z3::expr &v, const std::string &sort_name);", cfsname)
			}
		}
		if g.Config.Target != "test" {
			w.line("#endif")
		}
	}
}

// emitEnumSortArgSpecImpls emits the per-enum `operator<<`, `_arg<T>`,
// `__ser<T>`, `__deser<T>` definitions. Mirrors Python
// ivy_to_cpp.py:2497-2510 (operator<<, __ser) and 2634-2652 (_arg,
// __deser).
func (g *Generator) emitEnumSortArgSpecImpls(w *cppWriter) {
	enums := g.enumSortsForArgSpecs()
	if len(enums) == 0 {
		return
	}
	for _, st := range enums {
		g.emitEnumOperatorOut(w, st)
		g.emitEnumSer(w, st)
		if g.Config.Target == "repl" || g.Config.Target == "test" {
			g.emitEnumArg(w, st)
			g.emitEnumDeser(w, st)
		}
	}
	w.blank()
}

func (g *Generator) emitEnumSortOutSerImpls(w *cppWriter) {
	enums := g.enumSortsForArgSpecs()
	for _, st := range enums {
		g.emitEnumOperatorOut(w, st)
		g.emitEnumSer(w, st)
	}
}

func (g *Generator) emitEnumSortArgDeserImpls(w *cppWriter) {
	enums := g.enumSortsForArgSpecs()
	for _, st := range enums {
		g.emitEnumArg(w, st)
		g.emitEnumDeser(w, st)
	}
}

// emitEnumOperatorOut mirrors Python ivy_to_cpp.py:2502-2506.
func (g *Generator) emitEnumOperatorOut(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	if g.Config.Target == "test" {
		w.linef("std::ostream &operator <<(std::ostream &s, const %s &t){", cfsname)
		w.indent++
		for _, sym := range st.Extension {
			w.linef(`if (t == %s::%s) s<<%s;`, g.ClassName, varName(sym), strconv.Quote(sym))
		}
		w.line("return s;")
		w.indent--
		w.line("}")
		return
	}
	w.open(fmt.Sprintf("std::ostream &operator<<(std::ostream &s, const %s &t) {", cfsname))
	for _, sym := range st.Extension {
		w.linef(`if (t == %s::%s) s << %s;`, g.ClassName, varName(sym), strconv.Quote(sym))
	}
	w.line("return s;")
	w.close("")
	w.blank()
}

// emitEnumSer mirrors Python ivy_to_cpp.py:2507-2510.
func (g *Generator) emitEnumSer(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	w.line("template <>")
	if g.Config.Target == "test" {
		w.linef("void  __ser<%s>(ivy_ser &res, const %s&t){", cfsname, cfsname)
		w.indent++
		w.line("__ser(res,(int)t);")
		w.indent--
		w.line("}")
		return
	}
	w.open(fmt.Sprintf("void __ser<%s>(ivy_ser &res, const %s &t) {", cfsname, cfsname))
	w.line("__ser(res, (int)t);")
	w.close("")
	w.blank()
}

// emitEnumArg mirrors Python ivy_to_cpp.py:2639-2646.
func (g *Generator) emitEnumArg(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	w.line("template <>")
	if g.Config.Target == "test" {
		w.linef("%s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound){", cfsname, cfsname)
		w.indent++
		w.line("ivy_value &arg = args[idx];")
		w.line("if (arg.atom.size() == 0 || arg.fields.size() != 0) throw out_of_bounds(idx,arg.pos);")
		for _, sym := range st.Extension {
			w.linef(`if(arg.atom == %s) return %s::%s;`, strconv.Quote(sym), g.ClassName, varName(sym))
		}
		w.line(`throw out_of_bounds("bad value: " + arg.atom,arg.pos);`)
		w.indent--
		w.line("}")
		return
	}
	w.open(fmt.Sprintf("%s _arg<%s>(std::vector<ivy_value> &args, unsigned idx, long long bound) {", cfsname, cfsname))
	if g.Config.Target != "test" {
		w.line("(void)bound;")
	}
	w.line("ivy_value &arg = args[idx];")
	w.line("if (arg.atom.size() == 0 || arg.fields.size() != 0) throw out_of_bounds(idx, arg.pos);")
	for _, sym := range st.Extension {
		w.linef(`if (arg.atom == %s) return %s::%s;`, strconv.Quote(sym), g.ClassName, varName(sym))
	}
	w.line(`throw out_of_bounds("bad value: " + arg.atom, arg.pos);`)
	w.close("")
	w.blank()
}

// emitEnumDeser mirrors Python ivy_to_cpp.py:2647-2652.
func (g *Generator) emitEnumDeser(w *cppWriter, st *goivy.LogicEnumeratedSort) {
	cfsname := g.ClassName + "::" + varName(st.Name)
	w.line("template <>")
	if g.Config.Target == "test" {
		w.linef("void __deser<%s>(ivy_deser &inp, %s &res){", cfsname, cfsname)
		w.indent++
		w.line("int __res;")
		w.line("__deser(inp,__res);")
		w.linef("res = (%s)__res;", cfsname)
		w.indent--
		w.line("}")
		return
	}
	w.open(fmt.Sprintf("void __deser<%s>(ivy_deser &inp, %s &res) {", cfsname, cfsname))
	w.line("int __res;")
	w.line("__deser(inp, __res);")
	w.linef("res = (%s)__res;", cfsname)
	w.close("")
	w.blank()
}

// emitCmdReader emits a per-classname `cmd_reader` subclass of
// `stdin_reader` whose `process()` parses one command line, dispatches
// it to the appropriate public action with `_arg<T>`-converted
// arguments, and catches `syntax_error` / `out_of_bounds` / `bad_arity`.
//
// Mirrors Python `emit_repl_boilerplate1a` (ivy_to_cpp.py:4137-4157),
// the per-action dispatch chain (ivy_to_cpp.py:2677-2697), and
// `emit_repl_boilerplate2` (4160-4187).
func (g *Generator) emitCmdReader(w *cppWriter) {
	reprClass := g.ClassName + "_repl"
	readerClass := g.ClassName + "_cmd_reader"
	if g.Config.Target == "test" {
		readerClass = "cmd_reader"
		if len(g.Mod.Params) == 0 {
			g.emitPythonTestCmdReaderRaw(w, readerClass, reprClass)
			return
		}
	}
	w.open(fmt.Sprintf("class %s : public stdin_reader {", readerClass))
	w.line("int lineno;")
	w.line("public:")
	w.linef("%s &ivy;", reprClass)
	w.blank()
	w.open(fmt.Sprintf("%s(%s &_ivy) : ivy(_ivy) {", readerClass, reprClass))
	w.line("lineno = 1;")
	if g.Config.Target == "test" {
		w.line("if (isatty(fdes()))")
		w.indent++
		w.line(`__ivy_out << "> "; __ivy_out.flush();`)
		w.indent--
	} else {
		w.open("if (isatty(fdes())) {")
		w.line(`__ivy_out << "> ";`)
		w.line("__ivy_out.flush();")
		w.close("")
	}
	w.close("")
	w.blank()
	w.open("virtual void process(const std::string &cmd) {")
	w.line("std::string action;")
	w.line("std::vector<ivy_value> args;")
	w.open("try {")
	w.line("parse_command(cmd, action, args);")
	w.line("ivy.__lock();")
	if g.Config.Target == "test" {
		g.emitPythonTestCmdReaderDispatchChain(w)
		w.open("{")
		w.line(`std::cerr << "undefined action: " << action << std::endl;`)
		w.close("")
	} else {
		g.emitCmdReaderDispatchChain(w)
		w.line(`std::cerr << "undefined action: " << action << std::endl;`)
	}
	w.line("ivy.__unlock();")
	w.close(" catch (syntax_error &err) {")
	w.indent++
	w.line("ivy.__unlock();")
	w.line(`std::cerr << "line " << lineno << ":" << err.pos << ": syntax error" << std::endl;`)
	w.indent--
	w.open("} catch (out_of_bounds &err) {")
	w.line("ivy.__unlock();")
	w.line(`std::cerr << "line " << lineno << ":" << err.pos << ": " << err.txt << " bad value" << std::endl;`)
	w.close("")
	w.open("catch (bad_arity &err) {")
	w.line("ivy.__unlock();")
	w.line(`std::cerr << "action " << err.action << " takes " << err.num << " input parameters" << std::endl;`)
	w.close("")
	if g.Config.Target == "test" {
		w.line("if (isatty(fdes()))")
		w.indent++
		w.line(`__ivy_out << "> "; __ivy_out.flush();`)
		w.indent--
	} else {
		w.open("if (isatty(fdes())) {")
		w.line(`__ivy_out << "> ";`)
		w.line("__ivy_out.flush();")
		w.close("")
	}
	w.line("lineno++;")
	w.close("")
	w.close(";")
	w.blank()
}

func (g *Generator) emitPythonTestCmdReaderRaw(w *cppWriter, readerClass, reprClass string) {
	w.raw("\nclass " + readerClass + ": public stdin_reader {\n")
	w.raw("    int lineno;\n")
	w.raw("public:\n")
	w.raw("    " + reprClass + " &ivy;    \n\n")
	w.raw("    " + readerClass + "(" + reprClass + " &_ivy) : ivy(_ivy) {\n")
	w.raw("        lineno = 1;\n")
	w.raw("        if (isatty(fdes()))\n")
	w.raw("            __ivy_out << \"> \"; __ivy_out.flush();\n")
	w.raw("    }\n\n")
	w.raw("    virtual void process(const std::string &cmd) {\n")
	w.raw("        std::string action;\n")
	w.raw("        std::vector<ivy_value> args;\n")
	w.raw("        try {\n")
	w.raw("            parse_command(cmd,action,args);\n")
	w.raw("            ivy.__lock();\n\n")
	initActions := g.initialMixinActionNames()
	emitted := 0
	for _, name := range g.publicActionNamesSorted() {
		if initActions[name] {
			continue
		}
		username := strings.TrimPrefix(name, "ext:")
		fn, _ := funName(name)
		act, ok := g.Mod.Actions.Get2(name)
		if emitted > 0 {
			w.raw("                else\n    \n")
		}
		w.raw("                if (action == " + strconv.Quote(username) + ") {\n")
		arity := 0
		if ok && act != nil {
			arity = len(act.GetFormalParams())
		}
		w.raw(fmt.Sprintf("                    check_arity(args,%d,action);\n", arity))
		argExprs := []string{}
		if ok && act != nil {
			argExprs = g.emitDispatchArgExprs(act)
		}
		call := "ivy." + fn + "(" + strings.Join(argExprs, ", ") + ")"
		returns := 0
		if ok && act != nil {
			returns = len(act.GetFormalReturns())
		}
		if returns == 1 {
			w.raw("                    __ivy_out  << \"= \" << " + call + " << std::endl;\n")
		} else {
			w.raw("                    " + call + ";\n")
		}
		w.raw("                }\n")
		emitted++
	}
	if emitted > 0 {
		w.raw("                else\n    \n")
	}
	w.raw("            {\n")
	w.raw("                std::cerr << \"undefined action: \" << action << std::endl;\n")
	w.raw("            }\n")
	w.raw("            ivy.__unlock();\n")
	w.raw("        }\n")
	w.raw("        catch (syntax_error& err) {\n")
	w.raw("            ivy.__unlock();\n")
	w.raw("            std::cerr << \"line \" << lineno << \":\" << err.pos << \": syntax error\" << std::endl;\n")
	w.raw("        }\n")
	w.raw("        catch (out_of_bounds &err) {\n")
	w.raw("            ivy.__unlock();\n")
	w.raw("            std::cerr << \"line \" << lineno << \":\" << err.pos << \": \" << err.txt << \" bad value\" << std::endl;\n")
	w.raw("        }\n")
	w.raw("        catch (bad_arity &err) {\n")
	w.raw("            ivy.__unlock();\n")
	w.raw("            std::cerr << \"action \" << err.action << \" takes \" << err.num  << \" input parameters\" << std::endl;\n")
	w.raw("        }\n")
	w.raw("        if (isatty(fdes()))\n")
	w.raw("            __ivy_out << \"> \"; __ivy_out.flush();\n")
	w.raw("        lineno++;\n")
	w.raw("    }\n")
	w.raw("};\n\n")
}

// emitCmdReaderDispatchChain emits the `if (action == "X") { ... }`
// chain. Mirrors Python ivy_to_cpp.py:2677-2697.
func (g *Generator) emitCmdReaderDispatchChain(w *cppWriter) {
	initActions := g.initialMixinActionNames()
	names := g.publicActionNamesSorted()
	for _, name := range names {
		if initActions[name] {
			continue
		}
		username := strings.TrimPrefix(name, "ext:")
		fn, _ := funName(name)
		act, ok := g.Mod.Actions.Get2(name)
		w.open(fmt.Sprintf(`if (action == "%s") {`, username))
		if !ok {
			w.linef("check_arity(args, 0, action);")
			w.linef("ivy.%s();", fn)
			w.line("ivy.__unlock();")
			w.line("return;")
			w.close("")
			continue
		}
		formals := act.GetFormalParams()
		w.linef("check_arity(args, %d, action);", len(formals))
		argExprs := g.emitDispatchArgExprs(act)
		returns := act.GetFormalReturns()
		callExpr := fmt.Sprintf("ivy.%s(%s)", fn, strings.Join(argExprs, ", "))
		if g.Config.Trace {
			g.emitTracePrelude(w, username, argExprs)
		}
		switch len(returns) {
		case 0:
			w.linef("%s;", callExpr)
		case 1:
			retType := g.cppQualifiedType(returns[0].CSort, g.ClassName)
			w.linef("%s __ivy_result = %s;", retType, callExpr)
			w.linef(`__ivy_out << "= " << __ivy_result << std::endl;`)
		default:
			// Multi-return: trailing return-ref args.
			var outNames []string
			extraArgs := make([]string, 0, len(returns))
			for _, r := range returns {
				rname := varName(r.Name)
				w.linef("%s %s = %s;", g.cppQualifiedType(r.CSort, g.ClassName), rname, g.cppZeroValueInScope(r.CSort))
				extraArgs = append(extraArgs, rname)
				outNames = append(outNames, rname)
			}
			callExpr = fmt.Sprintf("ivy.%s(%s)", fn, strings.Join(append(append([]string{}, argExprs...), extraArgs...), ", "))
			w.linef("%s;", callExpr)
			for i, on := range outNames {
				if i == 0 {
					w.linef(`__ivy_out << "= " << %s << std::endl;`, on)
				} else {
					w.linef(`__ivy_out << %s << std::endl;`, on)
				}
			}
		}
		if g.Config.Trace {
			w.linef(`__ivy_out%s << "}" << std::endl;`, g.numberFormat())
		}
		w.line("ivy.__unlock();")
		w.line("return;")
		w.close("")
	}
}

func (g *Generator) emitPythonTestCmdReaderDispatchChain(w *cppWriter) {
	initActions := g.initialMixinActionNames()
	names := g.publicActionNamesSorted()
	emitted := 0
	for _, name := range names {
		if initActions[name] {
			continue
		}
		username := strings.TrimPrefix(name, "ext:")
		fn, _ := funName(name)
		act, ok := g.Mod.Actions.Get2(name)
		prefix := "if"
		if emitted > 0 {
			prefix = "else if"
		}
		w.open(fmt.Sprintf(`%s (action == "%s") {`, prefix, username))
		if !ok {
			w.line("check_arity(args, 0, action);")
			w.linef("ivy.%s();", fn)
			w.close("")
			emitted++
			continue
		}
		formals := act.GetFormalParams()
		w.linef("check_arity(args, %d, action);", len(formals))
		argExprs := g.emitDispatchArgExprs(act)
		returns := act.GetFormalReturns()
		callExpr := fmt.Sprintf("ivy.%s(%s)", fn, strings.Join(argExprs, ", "))
		if g.Config.Trace {
			g.emitTracePrelude(w, username, argExprs)
		}
		switch len(returns) {
		case 0:
			w.linef("%s;", callExpr)
		case 1:
			w.linef(`__ivy_out << "= " << %s << std::endl;`, callExpr)
		default:
			var outNames []string
			extraArgs := make([]string, 0, len(returns))
			for _, r := range returns {
				rname := varName(r.Name)
				w.linef("%s %s = %s;", g.cppQualifiedType(r.CSort, g.ClassName), rname, g.cppZeroValueInScope(r.CSort))
				extraArgs = append(extraArgs, rname)
				outNames = append(outNames, rname)
			}
			callExpr = fmt.Sprintf("ivy.%s(%s)", fn, strings.Join(append(append([]string{}, argExprs...), extraArgs...), ", "))
			w.linef("%s;", callExpr)
			for i, on := range outNames {
				if i == 0 {
					w.linef(`__ivy_out << "= " << %s << std::endl;`, on)
				} else {
					w.linef(`__ivy_out << %s << std::endl;`, on)
				}
			}
		}
		if g.Config.Trace {
			w.linef(`__ivy_out%s << "}" << std::endl;`, g.numberFormat())
		}
		w.close("")
		emitted++
	}
	if emitted > 0 {
		w.line("else")
	}
}

// emitDispatchArgExprs returns the `_arg<T>(args, idx, csortcard)`
// expression for each formal param. Python ivy_to_cpp.py:2680.
func (g *Generator) emitDispatchArgExprs(act goivy.Action) []string {
	formals := act.GetFormalParams()
	exprs := make([]string, 0, len(formals))
	for idx, p := range formals {
		exprs = append(exprs, g.argExprForSort("args", strconv.Itoa(idx), p.CSort))
	}
	return exprs
}

func (g *Generator) argExprForSort(argsExpr, idxExpr string, s goivy.Sort) string {
	return g.argExprForSortBound(argsExpr, idxExpr, s, g.cppSortCardStr(s))
}

func (g *Generator) argExprForSortBound(argsExpr, idxExpr string, s goivy.Sort, bound string) string {
	typ := g.cppQualifiedType(s, g.ClassName)
	return fmt.Sprintf("_arg<%s>(%s, %s, %s)", typ, argsExpr, idxExpr, bound)
}

// emitTracePrelude emits the trace `actname(arg1,arg2) {` line preceding
// an action call. Python ivy_to_cpp.py:2685-2690. The number_format
// prefix is inserted right after `__ivy_out` per Python.
func (g *Generator) emitTracePrelude(w *cppWriter, username string, argExprs []string) {
	nf := g.numberFormat()
	if len(argExprs) == 0 {
		w.linef(`__ivy_out%s << "%s {" << std::endl;`, nf, username)
		return
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`__ivy_out%s << "%s("`, nf, username))
	for i, a := range argExprs {
		if i > 0 {
			b.WriteString(` << ","`)
		}
		b.WriteString(fmt.Sprintf(" << %s", a))
	}
	b.WriteString(` << ") {" << std::endl;`)
	w.line(b.String())
}

// emitValueParser emits a try/catch block that parses `srcExpr` (a C++
// expression yielding a std::string) into a single ivy_value, then
// invokes `_arg<T>(arg_values, 0, csortcard)` and stores the result
// into the param's local variable `p__<name>`. Mirrors Python
// `emit_value_parser` (ivy_to_cpp.py:2858-2870).
//
// `lineno` is prefixed onto the error messages when non-empty, matching
// Python's `"{lineno}parameter ... out of bounds"` interpolation at
// ivy_to_cpp.py:2865-2868. Pass `goivy.Location{}` when no source
// location is available (Python's `lineno=None` branch).
func (g *Generator) emitValueParser(w *cppWriter, p *goivy.Const, srcExpr string, lineno goivy.Location) {
	pname := "p__" + varName(p.Name)
	prefix := escapeString(lineno.String())
	w.open("try {")
	w.line("int pos = 0;")
	w.line("std::vector<ivy_value> arg_values;")
	w.line("arg_values.resize(1);")
	w.linef("arg_values[0] = parse_value(%s, pos);", srcExpr)
	w.linef("%s = %s;", pname, g.argExprForSort("arg_values", "0", p.CSort))
	w.close(" catch (out_of_bounds &) {")
	w.indent++
	w.linef(`std::cerr << "%sparameter %s out of bounds\n";`, prefix, escapeString(p.Name))
	w.line("__ivy_exit(1);")
	w.indent--
	w.open("} catch (syntax_error &) {")
	w.linef(`std::cerr << "%ssyntax error in parameter value %s\n";`, prefix, escapeString(p.Name))
	w.line("__ivy_exit(1);")
	w.close("")
}

// emitMainParamSetup emits the param-handling preamble of every main():
// declares each module parameter, applies defaults from
// g.Mod.ParamDefaults, parses argv into `key=value` pairs (with special
// keys out/iters/runs/seed/delay/wait/modelfile) and positional pos_params,
// validates count, calls srand(), and emits the Winsock init block.
// Mirrors Python ivy_to_cpp.py:2702-2834 + emit_winsock_init.
//
// On return, the local C++ variables `argc`, `argv`, `seed`, `sleep_ms`,
// `final_ms`, and one `p__<name>` per parameter are in scope.
func (g *Generator) emitMainParamSetup(w *cppWriter) {
	// Declare each parameter and apply its default.
	for i, p := range g.Mod.Params {
		w.linef("%s;", g.cppStorageDecl("p__"+p.Name, p.CSort, g.ClassName))
		if i < len(g.Mod.ParamDefaults) && g.Mod.ParamDefaults[i] != nil {
			if _, isFS := p.CSort.(*goivy.LogicFunctionSort); isFS {
				g.errs = append(g.errs, fmt.Errorf("ivy2cpp: can't handle default values for function-sorted parameter %s", p.Name))
			} else {
				defText := paramDefaultText(g.Mod.ParamDefaults[i])
				if defText != "" {
					g.emitValueParser(w, p, strconv.Quote(defText), g.Mod.ParamDefaults[i].GetLineno())
				}
			}
		}
	}
	// argv parsing loop: key=value -> param assignment OR special key.
	w.line("int seed = 1;")

	w.line("std::uint8_t seed32[chacha8c::key_size];")
	w.line(`std::memcpy(seed32, "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456", chacha8c::key_size);`)
	w.line("__chacha8c_rng = chacha8c::NewChaCha8(seed32);")

	w.line("int sleep_ms = 10;")
	w.line("int final_ms = 0;")
	if g.Config.Target != "test" {
		w.line("(void)sleep_ms;")
		w.line("(void)final_ms;")
	}
	w.blank()
	w.line("std::vector<char *> pargs;")
	w.line("pargs.push_back(argv[0]);")
	w.open("for (int i = 1; i < argc; i++) {")
	w.line("std::string arg = argv[i];")
	w.line("size_t p = arg.find('=');")
	if g.Config.Target == "test" {
		w.line("if (p == std::string::npos)")
		w.indent++
		w.line("pargs.push_back(argv[i]);")
		w.indent--
		w.open("else {")
	} else {
		w.open("if (p == std::string::npos) {")
		w.line("pargs.push_back(argv[i]);")
		w.close(" else {")
	}
	w.indent++
	w.line("std::string param = arg.substr(0, p);")
	w.line("std::string value = arg.substr(p + 1);")
	g.emitParamKeyValueDispatch(w)
	w.indent--
	w.line("}")
	w.close("")
	w.line("srand(seed);")
	if g.Config.Target == "test" {
		w.line("if (!__ivy_out.is_open())")
		w.indent++
		w.line("__ivy_out.basic_ios<char>::rdbuf(std::cout.rdbuf());")
		w.indent--
	} else {
		w.open("if (!__ivy_out.is_open()) {")
		w.line("__ivy_out.basic_ios<char>::rdbuf(std::cout.rdbuf());")
		w.close("")
	}
	w.line("argc = pargs.size();")
	w.line("argv = &pargs[0];")
	w.blank()
	g.emitPositionalParamParse(w)
	g.emitWinsockInit(w)
}

// emitParamKeyValueDispatch emits the body of the `else` branch in the
// argv loop: each `key=value` arg dispatches to a default-bearing param
// assignment or to one of the special keys. Python ivy_to_cpp.py:2728-2770.
func (g *Generator) emitParamKeyValueDispatch(w *cppWriter) {
	for i, p := range g.Mod.Params {
		if i >= len(g.Mod.ParamDefaults) || g.Mod.ParamDefaults[i] == nil {
			continue
		}
		if _, isFS := p.CSort.(*goivy.LogicFunctionSort); isFS {
			continue
		}
		w.open(fmt.Sprintf(`if (param == "%s") {`, escapeString(p.Name)))
		g.emitValueParser(w, p, "value", goivy.Location{})
		w.line("continue;")
		w.close("")
	}
	// Special keys: out, iters, runs, seed, delay, wait, modelfile.
	w.open(`if (param == "out") {`)
	w.line("__ivy_out.open(value.c_str());")
	w.open("if (!__ivy_out) {")
	w.line(`std::cerr << "cannot open to write: " << value << std::endl;`)
	w.line("return 1;")
	w.close("")
	w.close("")
	w.linef(`else if (param == "iters") { test_iters = atoi(value.c_str()); }`)
	w.linef(`else if (param == "runs") { runs = atoi(value.c_str()); }`)
	w.linef(`else if (param == "seed") { seed = atoi(value.c_str()); }`)
	w.linef(`else if (param == "delay") { sleep_ms = atoi(value.c_str()); }`)
	w.linef(`else if (param == "wait") { final_ms = atoi(value.c_str()); }`)
	w.open(`else if (param == "modelfile") {`)
	w.line("__ivy_modelfile.open(value.c_str());")
	w.open("if (!__ivy_modelfile) {")
	w.line(`std::cerr << "cannot open to write: " << value << std::endl;`)
	w.line("return 1;")
	w.close("")
	w.close("")
	w.open("else {")
	w.line(`std::cerr << "unknown option: " << param << std::endl;`)
	w.line("return 1;")
	w.close("")
}

// emitPositionalParamParse emits Python's positional-parameter
// extraction (ivy_to_cpp.py:2779-2832): if `argc == npos+2`, the last
// argv is opened as a command file via `_open`/`_dup2`; otherwise we
// require `argc == npos+1`. Then each positional pos_param is parsed
// via parse_value + `_arg<T>`. Function-sorted params use the
// `make_function_app` shape from Python 2803-2826.
func (g *Generator) emitPositionalParamParse(w *cppWriter) {
	posParams := g.positionalParams()
	npos := len(posParams)
	w.linef("if (argc == %d) {", npos+2)
	w.indent++
	w.line("argc--;")
	w.line("int fd = _open(argv[argc], 0);")
	w.open("if (fd < 0) {")
	w.linef(`std::cerr << "cannot open to read: " << argv[argc] << "\n";`)
	w.line("__ivy_exit(1);")
	w.close("")
	w.line("_dup2(fd, 0);")
	w.indent--
	w.line("}")
	w.open(fmt.Sprintf("if (argc != %d) {", npos+1))
	usageNames := make([]string, len(posParams))
	for i, p := range posParams {
		usageNames[i] = escapeString(p.Name)
	}
	usage := strings.Join(usageNames, " ")
	w.linef(`std::cerr << "usage: %s %s\n";`, escapeString(g.ClassName), usage)
	w.line("__ivy_exit(1);")
	w.close("")
	if npos == 0 && g.Config.Target != "test" {
		return
	}
	w.line("std::vector<std::string> args;")
	w.linef("std::vector<ivy_value> arg_values(%d);", npos)
	w.line("for (int i = 1; i < argc; i++) { args.push_back(argv[i]); }")
	for idx, p := range posParams {
		g.emitOnePositionalParam(w, p, idx)
	}
}

// emitOnePositionalParam emits parsing for one positional parameter.
// Python ivy_to_cpp.py:2796-2832.
func (g *Generator) emitOnePositionalParam(w *cppWriter, p *goivy.Const, idx int) {
	pname := "p__" + varName(p.Name)
	w.open("try {")
	w.line("int pos = 0;")
	w.linef("arg_values[%d] = parse_value(args[%d], pos);", idx, idx)
	if fs, isFS := p.CSort.(*goivy.LogicFunctionSort); isFS {
		// Function-sorted param: parse list of {dom0,...,domN,rng} tuples.
		// Python s.sort.dom is the domain without codomain — Go's
		// Domain() has the same semantics, so no slicing needed.
		dom := fs.Domain()
		rng := fs.Range()
		w.linef("ivy_value &arg = arg_values[%d];", idx)
		w.open("if (arg.atom.size()) {")
		w.linef("throw out_of_bounds(%d);", idx)
		w.close("")
		w.open("for (unsigned i = 0; i < arg.fields.size(); i++) {")
		w.open(fmt.Sprintf("if (arg.fields[i].fields.size() != %d) {", 1+len(dom)))
		w.linef("throw out_of_bounds(%d);", idx)
		w.close("")
		// Build LHS: `p__name[arg0][arg1]...` (or ctuple-keyed when needed).
		domArgs := make([]string, len(dom))
		for j, d := range dom {
			domArgs[j] = g.argExprForSortBound("arg.fields[i].fields", strconv.Itoa(j), d, "0")
		}
		lhs := g.functionAppLHS(p, dom, domArgs)
		w.linef("%s = %s;", lhs, g.argExprForSortBound("arg.fields[i].fields", strconv.Itoa(len(dom)), rng, "0"))
		w.close("")
	} else {
		w.linef("%s = %s;", pname, g.argExprForSort("arg_values", strconv.Itoa(idx), p.CSort))
	}
	w.close(" catch (out_of_bounds &) {")
	w.indent++
	w.linef(`std::cerr << "parameter %s out of bounds\n";`, escapeString(p.Name))
	w.line("__ivy_exit(1);")
	w.indent--
	w.open("} catch (syntax_error &) {")
	w.line(`std::cerr << "syntax error in command argument\n";`)
	w.line("__ivy_exit(1);")
	w.close("")
}

// functionAppLHS mirrors Python `make_function_app` (ivy_to_cpp.py:2803-2817):
// for large multi-arg sorts use a ctuple key; otherwise chain `[arg]`
// accesses.
func (g *Generator) functionAppLHS(p *goivy.Const, dom []goivy.Sort, domArgs []string) string {
	base := "p__" + varName(p.Name)
	if isLargeFunctionDomain(g, dom) && len(dom) > 1 {
		return fmt.Sprintf("%s[%s(%s)]", base, cppCTupleNameWith(g, dom, g.ClassName), strings.Join(domArgs, ", "))
	}
	res := base
	for _, a := range domArgs {
		res += "[" + a + "]"
	}
	return res
}

// isLargeFunctionDomain mirrors Python `is_large_type` for function
// sorts (any non-integer-type domain element, or product > largeThresh).
func isLargeFunctionDomain(g *Generator, dom []goivy.Sort) bool {
	for _, d := range dom {
		if !cppIsAnyIntegerType(g, d) {
			return true
		}
	}
	product := 1
	for _, d := range dom {
		c := cppSortCard(g, d)
		if c <= 0 {
			return true
		}
		if product <= largeThresh {
			product *= c
		}
	}
	return product > largeThresh
}

// positionalParams returns module params without defaults — Python
// `pos_params` at ivy_to_cpp.py:2727-2735.
func (g *Generator) positionalParams() []*goivy.Const {
	var out []*goivy.Const
	for i, p := range g.Mod.Params {
		if i < len(g.Mod.ParamDefaults) && g.Mod.ParamDefaults[i] != nil {
			continue
		}
		out = append(out, p)
	}
	return out
}

// paramDefaultText returns the textual form of a param-default AST
// node — `d.rep` in Python (ivy_to_cpp.py:2710). For Go we accept
// either an Atom (Relname) or any node exposing a string Rep.
func paramDefaultText(n goivy.Node) string {
	if n == nil {
		return ""
	}
	if a, ok := n.(*goivy.Atom); ok {
		return a.Relname()
	}
	type relnamer interface{ Relname() string }
	if r, ok := n.(relnamer); ok {
		return r.Relname()
	}
	return fmt.Sprintf("%v", n)
}

// emitWinsockInit emits the Windows-only Winsock 2.2 initialization
// boilerplate from Python ivy_to_cpp.py:emit_winsock_init (4189-4225).
func (g *Generator) emitWinsockInit(w *cppWriter) {
	w.line("#ifdef _WIN32")
	w.open("{")
	w.line("WORD wVersionRequested;")
	w.line("WSADATA wsaData;")
	w.line("int err;")
	w.line("wVersionRequested = MAKEWORD(2, 2);")
	w.line("err = WSAStartup(wVersionRequested, &wsaData);")
	w.open("if (err != 0) {")
	w.line(`printf("WSAStartup failed with error: %d\n", err);`)
	w.line("return 1;")
	w.close("")
	w.open("if (LOBYTE(wsaData.wVersion) != 2 || HIBYTE(wsaData.wVersion) != 2) {")
	w.line(`printf("Could not find a usable version of Winsock.dll\n");`)
	w.line("WSACleanup();")
	w.line("return 1;")
	w.close("")
	w.close("")
	w.line("#endif")
}

// isPlainNumericSort matches uninterpreted sorts that map to a plain
// numeric C++ type (i.e. no native_type, no destructor struct, no
// variant). Used by z3.go to decide which sorts need a numeric
// randomization helper. Mirrors the old `replNeedsNumericParser`
// criterion, kept under its current callers' name.
func (g *Generator) replNeedsNumericParser(s goivy.Sort) bool {
	if g == nil || g.Mod == nil || s == nil {
		return false
	}
	name := sortName(s)
	if name == "" || g.isVariantSuperName(name) {
		return false
	}
	if _, ok := g.nativeTypeName(s, g.ClassName); ok {
		return false
	}
	if _, ok := g.destructorStructName(s); ok {
		return false
	}
	_, ok := s.(*goivy.UninterpretedSort)
	return ok
}

func (g *Generator) publicActionNamesSorted() []string {
	var names []string
	for name := range g.Mod.PublicActions.All() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
