// Module context management.
//
// Go equivalent of Python's Module.__enter__/__exit__ context manager
// and the package-level `module` variable.
package module

import (
	"strings"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// Enter sets m as the current module on its ModCfg, saving the previous
// one so that Exit can restore it. This is the Go equivalent of Python's
// Module.__enter__.
//
// Panics if m.ModCfg is nil — caller must set it.
//
// Usage:
//
//	m.Enter()
//	defer m.Exit()
func (m *Module) Enter() {
	cfg := m.Cfg
	if cfg == nil {
		panic("module.Enter: Cfg is nil — caller must set Cfg before calling Enter")
	}
	m.prevModule = cfg.CurrentModule
	// Python: self.old_sig = il.sig (save the previous module's sig)
	if m.prevModule != nil {
		m.oldSig = m.prevModule.Sig
	}
	cfg.CurrentModule = m
	// Python: ivy_solver.clear() — clear cached Z3 values when changing sig
	if cfg.SolverClearFn != nil {
		cfg.SolverClearFn()
	}
}

// Exit restores the previous module that was active before Enter was
// called. This is the Go equivalent of Python's Module.__exit__.
//
// Panics if m.ModCfg is nil.
func (m *Module) Exit() {
	cfg := m.Cfg
	if cfg == nil {
		panic("module.Exit: Cfg is nil — caller must set Cfg before calling Exit")
	}
	// Python: il.sig = self.old_sig (restore the previous sig)
	cfg.CurrentModule = m.prevModule
	m.prevModule = nil
	m.oldSig = nil
}

// RelevantDefinitions returns definitions whose defining symbol is
// reachable from the given set of symbol names. It computes the
// transitive closure of symbol dependencies through definition RHS.
//
// Corresponds to Python's relevant_definitions function.
func RelevantDefinitions(m *Module, syms map[string]bool) []*ast.LabeledFormula {
	// Build a map from defining symbol name to definition RHS symbols.
	defMap := make(map[string][]string)
	defSymMap := make(map[string]bool)
	for _, ldf := range m.Definitions {
		if def, ok := ldf.Formula.(*il.Definition); ok {
			defName := def.Defines().String()
			defSymMap[defName] = true
			// Collect symbols used in the RHS.
			rhsSyms := collectSymbolNames(def.Rhs)
			defMap[defName] = rhsSyms
		}
	}

	// Compute reachable set from syms through defMap.
	reachable := make(map[string]bool)
	var stack []string
	for s := range syms {
		stack = append(stack, s)
	}
	for len(stack) > 0 {
		sym := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable[sym] {
			continue
		}
		reachable[sym] = true
		if deps, ok := defMap[sym]; ok {
			for _, d := range deps {
				if !reachable[d] {
					stack = append(stack, d)
				}
			}
		}
	}

	// Filter definitions to those whose defining symbol is reachable.
	var result []*ast.LabeledFormula
	for _, ldf := range m.Definitions {
		if def, ok := ldf.Formula.(*il.Definition); ok {
			defName := def.Defines().String()
			if reachable[defName] {
				result = append(result, ldf)
			}
		}
	}
	return result
}

// SortDependencyGraph computes the full sort dependency graph for all
// sorts in the module. Returns a map from sort name to the list of sort
// names it depends on.
func (m *Module) SortDependencyGraph() map[string][]string {
	result := make(map[string][]string)
	for _, sname := range m.SortOrder {
		deps := m.SortDependencies(sname, true)
		if len(deps) > 0 {
			result[sname] = deps
		}
	}
	return result
}

// collectSymbolNames returns the names of all constant symbols in a node.
func collectSymbolNames(node lg.Expr) []string {
	var names []string
	collectSymbolNamesRec(node, &names, make(map[string]bool))
	return names
}

func collectSymbolNamesRec(node lg.Expr, names *[]string, seen map[string]bool) {
	switch t := node.(type) {
	case *lg.Const:
		if !seen[t.Name] {
			seen[t.Name] = true
			*names = append(*names, t.Name)
		}
		return
	case *lg.Apply:
		// Explicitly walk Func since Children() returns only Terms.
		collectSymbolNamesRec(t.Func, names, seen)
	}
	for _, c := range node.Children() {
		collectSymbolNamesRec(c, names, seen)
	}
}

// --- Package-level functions corresponding to Python module-level functions ---

// FindAction looks up an action by name in the current module.
// Corresponds to Python's module-level find_action (ivy_module.py:349-350).
func FindAction(cfg *Config, name string) (Action, bool) {
	if cfg == nil || cfg.CurrentModule == nil {
		return nil, false
	}
	return cfg.CurrentModule.FindAction(name)
}

// BackgroundTheory returns the background theory from the current module.
// Corresponds to Python's module-level background_theory (ivy_module.py:346-347).
func BackgroundTheory(cfg *Config, symbols map[string]bool) *co.Clauses {
	if cfg == nil || cfg.CurrentModule == nil {
		return co.NewClauses(nil, nil, nil)
	}
	return cfg.CurrentModule.BackgroundTheory(symbols)
}

// Logics returns the active logic names, checking current module first,
// then Config.CompleteLogic, then il.DefaultLogics.
// Corresponds to Python's module-level logics() function (ivy_module.py:355-358).
func Logics(cfg *Config) []string {
	if cfg != nil && cfg.CurrentModule != nil {
		return cfg.CurrentModule.GetLogics()
	}
	if cfg != nil && cfg.CompleteLogic != "" {
		return strings.Split(cfg.CompleteLogic, ",")
	}
	return il.DefaultLogics
}
