// Package printer provides functions for printing Ivy modules and their
// components in a human-readable format.
//
// Ported from ivy_printer.py.
package printer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/module"
)

// LabeledFmlasToStr formats a slice of labeled formulas with a keyword prefix.
// Each formula is printed on its own line with the keyword, optional label in
// brackets, and the formula body.
func LabeledFmlasToStr(kwd string, lfmlas []*module.LabeledFormula) string {
	var b strings.Builder
	for _, f := range lfmlas {
		b.WriteString(kwd)
		b.WriteByte(' ')
		if f.Label != nil {
			b.WriteString(fmt.Sprintf("[%v] ", f.Label))
		}
		b.WriteString(fmt.Sprint(f.Formula))
		b.WriteByte('\n')
	}
	return b.String()
}

// FormatModule formats an Ivy module's declarations and actions as a string.
// It prints schemata, axioms, properties, init, conjectures, definitions,
// interpretations, initializers, actions, and exports.
func FormatModule(mod *module.Module) string {
	var b strings.Builder

	// Signature
	if mod.Sig != nil {
		b.WriteString(fmt.Sprintf("%v\n", mod.Sig))
	}

	// Schemata
	names := sortedKeys(mod.Schemata)
	for _, x := range names {
		y := mod.Schemata[x]
		b.WriteString(fmt.Sprintf("schema [%s]%v\n", x, y))
	}

	// Labeled formula sections
	type section struct {
		kwd  string
		list []*module.LabeledFormula
	}
	sections := []section{
		{"axiom", mod.LabeledAxioms},
		{"property", mod.LabeledProps},
		{"init", mod.LabeledInits},
		{"conjecture", mod.LabeledConjs},
		{"definition", mod.Definitions},
	}
	for _, sec := range sections {
		b.WriteString(LabeledFmlasToStr(sec.kwd, sec.list))
	}

	// Interpretations
	interpNames := sortedInterpKeys(mod)
	for _, tn := range interpNames {
		interps := mod.Interps[tn]
		for _, interp := range interps {
			b.WriteString(fmt.Sprintf("interp %s -> %v\n", tn, interp))
		}
	}

	// Initializers
	for _, na := range mod.Initializers {
		if act, ok := na.Action.(actions.Action); ok {
			s := fmt.Sprintf("after init {%s}", act.String())
			b.WriteString(s)
			b.WriteByte('\n')
		}
	}

	// Actions
	actionNames := sortedActionNames(mod)
	for _, name := range actionNames {
		actIface := mod.Actions[name]
		if act, ok := actIface.(actions.Action); ok {
			b.WriteString(actions.ActionDefToStr(name, act))
			b.WriteByte('\n')
		}
	}

	// Exports
	publicNames := sortedPublicActions(mod)
	for _, x := range publicNames {
		b.WriteString(fmt.Sprintf("export %s\n", x))
	}

	return b.String()
}

// PrintModule prints the module to stdout. This mirrors the Python function
// print_module which calls print() on each section.
func PrintModule(mod *module.Module) {
	fmt.Print(FormatModule(mod))
}

// --- helper functions ---

// sortedKeys returns the sorted keys from a string-keyed map.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedInterpKeys returns sorted interpretation type names from a module.
func sortedInterpKeys(mod *module.Module) []string {
	keys := make([]string, 0, len(mod.Interps))
	for k := range mod.Interps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedActionNames returns sorted action names from a module.
func sortedActionNames(mod *module.Module) []string {
	keys := make([]string, 0, len(mod.Actions))
	for k := range mod.Actions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedPublicActions returns sorted public action names from a module.
func sortedPublicActions(mod *module.Module) []string {
	keys := make([]string, 0, len(mod.PublicActions))
	for k := range mod.PublicActions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
