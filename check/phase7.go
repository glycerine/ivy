// phase7.go implements Phase 7 helper functions for the check package,
// ported from Python ivy_check.py.
package check

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/module"
)

// GuiArt launches the GUI for an analysis graph.
// In the Python version this opens a Tk UI for interactive exploration.
// In Go, this is a skeletal implementation that prints diagnostic info.
// Corresponds to Python's gui_art (ivy_check.py lines 85-101).
func GuiArt(mod *module.Module, otherArt *art.AnalysisGraph) error {
	if otherArt == nil {
		otherArt = art.NewAnalysisGraph(mod)
	}
	fmt.Println("initializers:", len(mod.Initializers))
	if initAct, ok := mod.Actions.Get2("initialize"); ok {
		fmt.Println("initialize:", initAct)
	}
	// In the full implementation, this would open a Tk GUI.
	// For now, just report that the GUI is not available in CLI mode.
	fmt.Println("GUI mode not available in CLI; use --trace for text output")
	return nil
}

// Usage prints the usage message and exits.
// Corresponds to Python's usage (ivy_check.py lines 160-162).
func Usage() string {
	return fmt.Sprintf("usage: \n  %s file.ivy", os.Args[0])
}

// ShowAssertions prints all assertion actions found in the module.
// Corresponds to Python's show_assertions (ivy_check.py lines 174-176).
func ShowAssertions(mod *module.Module) string {
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
//	def print_dots(): print('...', end=' ')
func PrintDots() string {
	fmt.Print("... ")
	return "... "
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
