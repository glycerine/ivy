// Module context management.
//
// Go equivalent of Python's Module.__enter__/__exit__ context manager
// and the package-level `module` variable.
package module

import (
	"sync"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

var (
	moduleMu      sync.Mutex
	currentModule *Module
)

// CurrentModule returns the currently active module, or nil if none.
func CurrentModule() *Module {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	return currentModule
}

// Enter sets m as the current module, saving the previous one so that
// Exit can restore it. This is the Go equivalent of Python's
// Module.__enter__.
//
// Usage:
//
//	m.Enter()
//	defer m.Exit()
func (m *Module) Enter() {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	m.prevModule = currentModule
	currentModule = m
}

// Exit restores the previous module that was active before Enter was
// called. This is the Go equivalent of Python's Module.__exit__.
func (m *Module) Exit() {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	currentModule = m.prevModule
	m.prevModule = nil
}

// RelevantDefinitions returns definitions whose defining symbol is
// reachable from the given set of symbol names. It computes the
// transitive closure of symbol dependencies through definition RHS.
//
// Corresponds to Python's relevant_definitions function.
func RelevantDefinitions(m *Module, syms map[string]bool) []*LabeledFormula {
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
	var result []*LabeledFormula
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
func collectSymbolNames(node lg.Node) []string {
	var names []string
	collectSymbolNamesRec(node, &names, make(map[string]bool))
	return names
}

func collectSymbolNamesRec(node lg.Node, names *[]string, seen map[string]bool) {
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
