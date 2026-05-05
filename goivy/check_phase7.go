// phase7.go implements the GuiArt analysis-graph GUI entry, ported from
// Python ivy_check.py. The GUI itself is a stub in this Go port; the
// function prints diagnostic info and returns.
package goivy

import (
	"fmt"
	"os"
	"strings"
)

// CheckGuiArtHook is an alias for module.GuiArtHook so check-package callers can
// declare hooks without importing module by name.
type CheckGuiArtHook = GuiArtHook

// GuiArt launches the GUI for an analysis graph. Corresponds to Python's
// gui_art (ivy_check.py:86-102).
//
// Python's gui_art ends with gui.tk.mainloop() then exit(1). In Go, the
// blocking-loop and process-exit responsibilities belong to the caller and
// the registered GuiArtHook (typically supplied by the webui package). When
// no hook is registered, GuiArt performs the same data setup Python does,
// prints a diagnostic summary, and returns nil — matching the previous stub
// behavior but with the data side-effects faithfully ported.
//
// `target` is interface{} to mirror Python's polymorphism: it accepts a
// *art.AnalysisGraph (from ShowCounterexample) or a *MatchHandler (from the
// trace failure path). `isCti` carries the failing-conjecture clauses from
// MatchHandler.IsCti, or nil for non-CTI counterexamples.
func GuiArt(mod *Module, target interface{}, isCti *Clauses) error {
	// Resolve the target to an AnalysisGraph for the data-setup branch.
	// (The hook itself receives the original target unchanged.)
	var otherArt *AnalysisGraph
	switch v := target.(type) {
	case *AnalysisGraph:
		otherArt = v
	case nil:
		otherArt = NewAnalysisGraph(mod)
	default:
		// *MatchHandler or other handler types — leave otherArt nil; the
		// "art" UI branch below will create a fresh graph just like Python.
		_ = v
	}

	// Python: if ivy_ui.default_ui.get() == "art":
	if mod.Cfg.IuCfg != nil && mod.Cfg.IuCfg.DefaultUI == "art" {
		// Python: print("initializers: {}".format(im.module.initializers))
		fmt.Printf("initializers: %d\n", len(mod.Initializers))
		// Python: other_art = ivy_art.AnalysisGraph()
		otherArt = NewAnalysisGraph(mod)
		// Python: other_art.add_initial_state()
		otherArt.AddInitialState(nil, nil)
		// Python: if 'initialize' in im.module.actions:
		if initAct, ok := mod.Actions.Get2("initialize"); ok {
			// Python: print("initialize: {}".format(init_action))
			fmt.Printf("initialize: %v\n", initAct)
			// Python: ag.execute(init_action, None, None, 'initialize')
			//
			// NOTE: Python references undefined `ag` here — almost certainly
			// a bug; the only graph in scope is `other_art`. We use otherArt
			// to make this branch actually run. The Python "art" UI mode is
			// rare in practice (default is "cti"), which is presumably why
			// the typo went unnoticed upstream.
			_, _ = otherArt.Execute(false, initAct, nil, nil, "initialize")
		}
	}

	// Python: agui = gui.add(other_art); gui.tk.update_idletasks();
	//         gui.tk.mainloop(); exit(1)
	//
	// In Go, delegate to the registered hook (typically webui). The hook is
	// a fully-typed module.GuiArtHook — no interface{} unboxing needed.
	if mod.Cfg.GuiArtHook != nil {
		// Pass the original target (which may be a Trace handler) when we
		// have one; otherwise pass the freshly-built otherArt from the "art"
		// UI branch above. Hooks that need a specific concrete type can
		// type-switch themselves.
		if target == nil && otherArt != nil {
			return mod.Cfg.GuiArtHook(mod, otherArt, isCti)
		}
		return mod.Cfg.GuiArtHook(mod, target, isCti)
	}

	// No hook registered: print diagnostic info matching the previous stub.
	fmt.Println("initializers:", len(mod.Initializers))
	if initAct, ok := mod.Actions.Get2("initialize"); ok {
		fmt.Println("initialize:", initAct)
	}
	if isCti != nil {
		fmt.Println("CTI clauses:", isCti)
	}
	fmt.Println("GUI mode not available in CLI; register a check.CheckGuiArtHook or use --trace")
	return nil
}

// Usage prints the usage message and exits.
// Corresponds to Python's usage (ivy_check.py lines 160-162).
func Usage() string {
	return fmt.Sprintf("usage: \n  %s file.ivy", os.Args[0])
}

// ShowAssertions prints all assertion actions found in the module.
// Corresponds to Python's show_assertions (ivy_check.py lines 174-176).
func ShowAssertions(mod *Module) string {
	assertions := FindAssertions("", mod)
	var sb strings.Builder
	for _, a := range assertions {
		sb.WriteString(fmt.Sprintf("%v: %v\n", a.GetLineno(), a))
	}
	return sb.String()
}

// PrintDots prints progress dots to stdout.
// Corresponds to Python's print_dots (ivy_check.py lines 201-203):
//
//	def print_dots(): print('...', end='\n')
func PrintDots() string {
	fmt.Println("...")
	return "...\n"
}

// FilterFcs filters a list of checkers based on the global CheckLineno.
// If CheckLineno is empty, all checkers are returned.
// Corresponds to Python's filter_fcs (ivy_check.py lines 367-371).
func FilterFcs(fcs []Checker, filter func(Checker) bool) []Checker {
	if filter == nil {
		return fcs
	}
	var result []Checker
	for _, fc := range fcs {
		if filter(fc) {
			result = append(result, fc)
		}
	}
	return result
}

// Info is an exception hook that prints exception info and optionally
// starts a debugger. In Go, this is adapted to log the error.
// Corresponds to Python's info (ivy_check.py lines 1008-1020).
func Info(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}
