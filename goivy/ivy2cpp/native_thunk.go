package ivy2cpp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Callback thunk emission. Ports Python `create_thunk` and the
// callback-collection pass at ivy_to_cpp.py:1404-1428 and 2187-2200.
//
// A "callback" is an action referenced from inside a native code block
// (either a top-level <<< member|impl|... >>> native, or an inline
// <<< ... >>> NativeAction inside an action body). For each such
// action we emit one file-scope struct named `thunk__<actname>` that
// captures any "prm:"-prefixed formal params at construction and
// invokes the action through the host class pointer on operator().
//
// This is unrelated to the per-assignment hash-thunk closure in
// thunk.go (which serves emit_assign_large).

// collectCallbackActions returns the sorted list of action names
// referenced from any native antiquote in the module. Mirrors Python
// ivy_to_cpp.py:2187-2200.
func (g *Generator) collectCallbackActions() []string {
	if g == nil {
		return nil
	}
	return collectCallbackActionNames(g.Mod)
}

func collectCallbackActionNames(mod *goivy.Module) []string {
	if mod == nil || mod.Actions == nil {
		return nil
	}
	used := map[string]bool{}
	// Top-level natives: args[0] is the label, args[1] the code body;
	// the remaining args are the antiquote parameters. Per
	// CompileNativeDef (goivy/compiler_phase6.go:1086, 1093), each
	// arg is wrapped in a *CompiledNode that holds the compiled
	// goivy.Expr — unwrap before checking.
	for _, n := range mod.Natives {
		args := n.Args()
		if len(args) < 2 {
			continue
		}
		for _, child := range args[2:] {
			node := unwrapCompiled(child)
			if name, ok := callbackActionName(mod, node); ok {
				used[name] = true
			}
		}
	}
	// NativeAction subactions inside action bodies. Python's
	// `actb.iter_subactions()` is `act.IterSubactions()` in Go (see
	// goivy/actions_action.go:1009).
	for _, act := range mod.Actions.All() {
		for _, sub := range act.IterSubactions() {
			na, ok := sub.(*goivy.LogicNativeAction)
			if !ok {
				continue
			}
			for _, p := range na.Params {
				if name, ok := callbackActionName(mod, p); ok {
					used[name] = true
				}
			}
		}
	}
	names := make([]string, 0, len(used))
	for n := range used {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// unwrapCompiled strips a *goivy.CompiledNode wrapper down to the
// underlying Expr/Node payload. Top-level native args go through
// CompileNativeDef which wraps each in a CompiledNode (see
// goivy/compiler_phase6.go:1063, 1086, 1093). Antiquote params on
// LogicNativeAction reach us already unwrapped.
func unwrapCompiled(n goivy.Node) goivy.Node {
	cn, ok := n.(*goivy.CompiledNode)
	if !ok {
		return n
	}
	if inner, ok := cn.Node.(goivy.Node); ok {
		return inner
	}
	return n
}

// emitCallbackThunks writes one struct per callback action at file
// scope in the impl buffer. Called from emitImpl before the impl
// natives are written, matching Python's order at ivy_to_cpp.py:2187
// (thunks) versus 2265 (impl natives).
func (g *Generator) emitCallbackThunks(w *cppWriter) {
	for _, name := range g.collectCallbackActions() {
		act, ok := g.Mod.Actions.Get2(name)
		if !ok {
			continue
		}
		g.emitCallbackThunk(w, name, act)
	}
}

// emitCallbackThunk emits one thunk struct. Direct port of Python
// `create_thunk` (ivy_to_cpp.py:1412-1427):
//
//	struct thunk__name {
//	    Class *__ivy;
//	    <prm: members>
//	    thunk__name(<prm: params>, Class *__ivy)
//	        : __ivy(__ivy), <prm: init list> {}
//	    Return operator()(<non-prm inputs>) const {
//	        [return] __ivy->name(<all formals>);
//	    }
//	};
func (g *Generator) emitCallbackThunk(w *cppWriter, name string, act goivy.Action) {
	tc := "thunk__" + varName(name)
	formals := act.GetFormalParams()
	var prm, inputs []*goivy.Const
	for _, p := range formals {
		if strings.HasPrefix(p.Name, "prm:") {
			prm = append(prm, p)
		} else {
			inputs = append(inputs, p)
		}
	}

	w.open(fmt.Sprintf("struct %s {", tc))
	w.linef("%s *__ivy;", g.ClassName)
	// Captured "prm:" members.
	for _, p := range prm {
		w.linef("%s;", g.cppStorageDecl(p.Name, p.CSort, g.ClassName))
	}
	// Constructor. Python emits `<class> *__ivy` last in the parameter
	// list (`extra = [classname + ' *__ivy']` passed to
	// emit_param_decls); the initializer list always starts with
	// `__ivy(__ivy)` followed by `prm(prm)` for each captured param.
	ctorParams := make([]string, 0, len(prm)+1)
	for _, p := range prm {
		ctorParams = append(ctorParams, g.cppStorageDecl(p.Name, p.CSort, g.ClassName))
	}
	ctorParams = append(ctorParams, fmt.Sprintf("%s *__ivy", g.ClassName))
	inits := []string{"__ivy(__ivy)"}
	for _, p := range prm {
		v := varName(p.Name)
		inits = append(inits, fmt.Sprintf("%s(%s)", v, v))
	}
	w.linef("%s(%s) : %s {}", tc, strings.Join(ctorParams, ", "), strings.Join(inits, ", "))

	// operator() — input params and (possibly) return type.
	ret := "void"
	returns := act.GetFormalReturns()
	if len(returns) > 0 {
		ret = g.cppType(returns[0].CSort)
	}
	opParams := make([]string, 0, len(inputs))
	for _, p := range inputs {
		opParams = append(opParams, g.cppStorageDecl(p.Name, p.CSort, g.ClassName))
	}
	w.open(fmt.Sprintf("%s operator()(%s) const {", ret, strings.Join(opParams, ", ")))
	fn, err := funName(name)
	if err != nil {
		fn = varName(name)
	}
	allArgs := make([]string, 0, len(formals))
	for _, p := range formals {
		allArgs = append(allArgs, varName(p.Name))
	}
	call := fmt.Sprintf("__ivy->%s(%s);", fn, strings.Join(allArgs, ", "))
	w.line("return " + call)
	w.close("")
	w.close(";")
}
