// Package ivydump provides analysis graph serialization and Ivy source generation.
// This is a port of Python's ivy_dump.py (60 lines).
//
// The Python version uses pickle to load .a2g files and converts them to .ivy.
// The Go version uses JSON for serialization (pickle is Python-specific).
// It generates Ivy source text from an analysis graph.
package ivydump

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"io"
	"strings"
)

// DumpToIvy writes an analysis graph as Ivy source to the writer.
// Corresponds to Python's ivy_dump_file().
func DumpToIvy(w io.Writer, ag *goivy.AnalysisGraph) error {
	fmt.Fprintln(w, "#lang ivy1.7")
	fmt.Fprintln(w)

	if ag.Domain == nil {
		return nil
	}
	mod := ag.Domain

	// Dump relations
	for name, sort := range mod.Relations.All() {
		fmt.Fprintf(w, "relation %s : %s\n", name, sort)
	}
	if mod.Relations.Len() > 0 {
		fmt.Fprintln(w)
	}

	// Dump functions
	for name, sort := range mod.Functions.All() {
		fmt.Fprintf(w, "function %s : %s\n", name, sort)
	}
	if mod.Functions.Len() > 0 {
		fmt.Fprintln(w)
	}

	// Dump axioms
	for _, lf := range mod.LabeledAxioms {
		if lf.Formula != nil {
			fmt.Fprintf(w, "axiom %s\n", lf.Formula)
		}
	}
	if len(mod.LabeledAxioms) > 0 {
		fmt.Fprintln(w)
	}

	// Dump initial states
	for _, state := range ag.States {
		if state.Pred == nil && state.Clauses != nil {
			fmt.Fprintf(w, "init %s\n", state.Clauses.ToFormula())
		}
	}

	// Dump actions
	for name, actIface := range mod.Actions.All() {
		fmt.Fprintf(w, "action %s = {\n", name)
		if act, ok := actIface.(interface{ String() string }); ok {
			lines := strings.Split(act.String(), "\n")
			for _, line := range lines {
				fmt.Fprintf(w, "    %s\n", line)
			}
		}
		fmt.Fprintln(w, "}")
		fmt.Fprintln(w)
	}

	return nil
}

// DumpFormulaToIvy converts a formula to Ivy syntax string.
func DumpFormulaToIvy(fmla goivy.Expr) string {
	if fmla == nil {
		return "true"
	}
	return fmla.String()
}
