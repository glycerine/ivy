// tactics_auto_inst.go — Port of Python ivy_auto_inst.py.
// Implements the auto_inst proof tactic: eager, trigger-based instantiation
// of axiom schemata, adding new labeled premises to the proof goal.
package goivy

import (
	"fmt"
)

// AutoInstTactic implements the "auto_inst" proof tactic.
// Mirrors Python ivy_auto_inst.py:auto_inst (lines 319-352).
//
// For each Trigger tactic decl, it pattern-matches the trigger against
// formulas in the goal, instantiates the named axiom, and adds the instance
// as a new premise. With no trigger decls the goal is returned unchanged.
func AutoInstTactic(pc ProofCheckerInterface, goals []*LabeledFormula, pf Node) ([]*LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, fmt.Errorf("auto_inst: no proof goals")
	}
	goal := goals[0]
	conc := GoalConc(goal)

	// Build axiom map keyed by label name.
	// Python: axmap = dict((ax.label.rep, ax) for ax in im.module.labeled_axioms)
	axiomMap := make(map[string]*LabeledFormula)
	for _, ax := range pc.GetAxioms() {
		axiomMap[ax.LabelName()] = ax
	}

	// Process tactic decls — only *Trigger nodes are valid.
	// Python: for decl in proof.tactic_decls: if isinstance(decl, ivy_ast.Trigger): ...
	var triggers []TriggerAxiom
	if tt, ok := pf.(*TacticTactic); ok {
		for _, decl := range tt.TacticDeclsList() {
			trig, isTrig := decl.(*Trigger)
			if !isTrig {
				return nil, fmt.Errorf("auto_inst: tactic does not take this type of argument: %T", decl)
			}
			axname := ""
			if a, ok := trig.Pattern.(*Atom); ok {
				axname = a.Relname()
			}
			ax, found := axiomMap[axname]
			if !found {
				return nil, fmt.Errorf("auto_inst: property %s not found", axname)
			}
			var trigExprs []Expr
			for _, t := range trig.Terms {
				if e, ok := t.(Expr); ok {
					trigExprs = append(trigExprs, e)
				}
			}
			triggers = append(triggers, TriggerAxiom{
				Triggers: trigExprs,
				Axiom:    ax,
			})
		}
	}

	// Collect formulas to search for trigger matches.
	// Python: fmlas = [ipr.goal_conc(g) for g in ipr.goal_prems(goal) if ipr.is_goal(g)] + [conc]
	var fmlas []Expr
	for _, prem := range GoalPrems(goal) {
		if subgoal, ok := prem.(*LabeledFormula); ok {
			if subGoalConc := GoalConc(subgoal); subGoalConc != nil {
				if e, ok := subGoalConc.(Expr); ok {
					fmlas = append(fmlas, e)
				}
			}
		}
	}
	if e, ok := conc.(Expr); ok {
		fmlas = append(fmlas, e)
	}

	// Instantiate axioms by matching triggers against goal formulas.
	instances := InstantiateAxioms2(pc.GetModule(), fmlas, triggers)

	if len(instances) == 0 {
		// No new instances — return goal unchanged.
		// Python: return [goal] + decls[1:]
		return append([]*LabeledFormula{goal}, goals[1:]...), nil
	}

	// Add instantiated axioms as new premises with fresh label names.
	// Python: UniqueRenamer + ax.clone([Atom(name), fmla])
	prems := make([]Node, len(GoalPrems(goal)))
	copy(prems, GoalPrems(goal))

	usedNames := make(map[string]bool)
	for _, p := range prems {
		if lf, ok := p.(*LabeledFormula); ok {
			usedNames[lf.LabelName()] = true
		}
	}
	for n := range axiomMap {
		usedNames[n] = true
	}

	counter := 0
	cfg := pc.GetAstCfg()
	for _, inst := range instances {
		axLabel := inst.Axiom.LabelName()
		name := axLabel
		for usedNames[name] {
			counter++
			name = fmt.Sprintf("%s_%d", axLabel, counter)
		}
		usedNames[name] = true

		newLabel := cfg.NewAtom(name)
		newLF := &LabeledFormula{
			Label:    newLabel,
			Formula:  inst.Formula,
			Explicit: false,
		}
		newLF.Cfg = cfg
		prems = append(prems, newLF)
	}

	newGoal := CloneGoal(cfg, goal, prems, conc)
	return append([]*LabeledFormula{newGoal}, goals[1:]...), nil
}

// RegisterAutoInstTactics registers the auto_inst tactic on the proof config.
// Mirrors Python ivy_auto_inst.py:354 (ipr.register_tactic('auto_inst', auto_inst)).
func RegisterAutoInstTactics(proofCfg *ProofConfig) {
	proofCfg.RegisterTactic("auto_inst", AutoInstTactic)
}
