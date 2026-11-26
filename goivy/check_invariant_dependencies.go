package goivy

import (
	"fmt"
	"sort"
	"strings"
)

func checkInvariantDependencyCycles(mod *Module) error {
	if len(mod.InvariantDependencies) == 0 {
		return nil
	}
	var arcs [][2]string
	for name, deps := range mod.InvariantDependencies {
		for _, dep := range deps {
			arcs = append(arcs, [2]string{name, dep})
		}
	}
	var cycles []string
	for _, cycle := range TarjanArcs(arcs) {
		sort.Strings(cycle)
		cycles = append(cycles, strings.Join(cycle, ", "))
	}
	if len(cycles) == 0 {
		return nil
	}
	sort.Strings(cycles)
	return &IvyError{Msg: "zero-delay invariant dependency cycle: " + cycles[0]}
}

func invariantDependencyAssumptions(mod *Module) map[string]Expr {
	if len(mod.InvariantDependencies) == 0 {
		return nil
	}
	available := make(map[string]Expr, len(mod.AssumedInvs)+len(mod.LabeledConjs))
	for _, invariants := range [][]*LabeledFormula{mod.AssumedInvs, mod.LabeledConjs} {
		for _, lf := range invariants {
			if name := lfLabelName(lf); name != "" {
				if formula, ok := lf.Formula.(Expr); ok {
					available[name] = formula
				}
			}
		}
	}
	return available
}

func invariantDependencyHypotheses(mod *Module, lf *LabeledFormula, available map[string]Expr) ([]Expr, error) {
	var missing []string
	var hypotheses []Expr
	for _, name := range mod.InvariantDependencies[lfLabelName(lf)] {
		if formula, ok := available[name]; ok {
			hypotheses = append(hypotheses, formula)
		} else {
			missing = append(missing, name)
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return nil, NewIvyError(lf, fmt.Sprintf("zero-delay dependency %s is not visible in this isolate", strings.Join(missing, ", ")))
	}
	return hypotheses, nil
}

func invariantWithDependencies(mod *Module, lf *LabeledFormula, available map[string]Expr) (*LabeledFormula, error) {
	if available == nil {
		return lf, nil
	}
	hypotheses, err := invariantDependencyHypotheses(mod, lf, available)
	if len(hypotheses) == 0 {
		return lf, err
	}
	formula := &LogicImplies{T1: &LogicAnd{Terms: hypotheses}, T2: lf.Formula.(Expr)}
	return lf.Clone([]Node{lf.Label, formula}).(*LabeledFormula), nil
}
