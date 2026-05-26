package ivy2go

import (
	"sort"

	"github.com/glycerine/ivy/goivy"
)

// native_thunk.go mirrors ivy2cpp/native_thunk.go. The cpp version
// emits a struct-per-callback-action that captures state and
// invokes the action through a host pointer; Go can do the same
// trick with method values (`obj.Action`), so the emission shape
// would differ. The analysis helpers below — which collect the set
// of action names referenced from native code — are language-
// agnostic and port unchanged.

// collectCallbackActions mirrors ivy2cpp/native_thunk.go:27. Returns
// the sorted list of action names referenced from any native
// antiquote in the module.
func (g *Generator) collectCallbackActions() []string {
	if g == nil {
		return nil
	}
	return collectCallbackActionNames(g.Mod)
}

// collectCallbackActionNames mirrors ivy2cpp/native_thunk.go:34.
func collectCallbackActionNames(mod *goivy.Module) []string {
	if mod == nil || mod.Actions == nil {
		return nil
	}
	used := map[string]bool{}
	// Top-level natives: args[0] is the label, args[1] the code body;
	// args[2:] are the antiquote parameters wrapped in *CompiledNode.
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
	// NativeAction subactions inside action bodies.
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

// unwrapCompiled mirrors ivy2cpp/native_thunk.go:85. Strips a
// CompiledNode wrapper down to the underlying payload.
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

// emitNativeThunkDecls is a no-op placeholder. The Go equivalent of
// cpp's struct-per-thunk would use method values (`obj.<Action>`)
// inlined at each callback site, so there's no per-thunk decl to
// emit. Kept for parallel call-site shape with ivy2cpp.
func (g *Generator) emitNativeThunkDecls(w *goWriter) { _ = w }
