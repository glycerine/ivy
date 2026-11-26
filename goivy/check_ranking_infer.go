package goivy

import (
	"errors"
	"fmt"
	"strings"
)

var errRankingModelExported = errors.New("ranking inference model exported")

func exportRankingInference(mod *Module, goal *LabeledFormula, model *NormalProgram, formula Expr) error {
	tasks, triggers, order := rankingTaskDefinitions(goal, formula)
	if len(order) == 0 {
		return fmt.Errorf("ranking_infer requires a work task")
	}
	for _, suffix := range order {
		for _, required := range []string{"work_created", "work_progress"} {
			if tasks[suffix][required] == nil {
				return fmt.Errorf("ranking_infer requires a definition of %s%s", required, suffix)
			}
		}
	}

	suffix := order[0]
	start := triggers[suffix]["work_start"]
	progress := tasks[suffix]["work_progress"]
	if start == nil {
		return fmt.Errorf("ranking_infer needs a work_start definition or a globally reaction")
	}
	negated, ok := start.T2.(*LogicNot)
	if !ok {
		return fmt.Errorf("ranking_infer needs a globally reaction of the form P -> eventually Q")
	}
	reaction, ok := negated.Body.(*LogicImplies)
	if !ok {
		return fmt.Errorf("ranking_infer needs a globally reaction of the form P -> eventually Q")
	}
	eventual, ok := reaction.T2.(*LogicEventually)
	if !ok {
		return fmt.Errorf("ranking_infer needs a globally reaction of the form P -> eventually Q")
	}
	predicates := []vmtPredicate{
		{name: ".p", tag: "react_p", parameters: eqLHSArgs(start), body: reaction.T1},
		{name: ".q", tag: "react_q", parameters: eqLHSArgs(start), body: eventual.Body},
		{name: ".r", tag: "react_r", parameters: eqLHSArgs(progress), body: progress.T2},
	}
	var premises []Node
	for _, premise := range GoalPrems(goal) {
		work := false
		if lf, ok := premise.(*LabeledFormula); ok && GoalIsProperty(lf) {
			if formula, ok := lf.Formula.(Expr); ok {
				for _, symbol := range UsedSymbolsAST(formula).All() {
					if strings.HasPrefix(ExprName(symbol), "work_") {
						work = true
						break
					}
				}
			}
		}
		if !work {
			premises = append(premises, premise)
		}
	}
	cfg := mod.Cfg.AstCfg
	filtered := CloneGoal(cfg, goal, premises, cfg.NewTemporalModels(model, True))
	component := temporalProofModule(mod, filtered, model)
	vocabulary := GoalVocab(filtered)
	withSymbols := NewWithSymbols(component.Sig, vocabulary.Symbols)
	withSymbols.Enter()
	defer withSymbols.Exit()
	withSorts := NewWithSorts(component.Sig, vocabulary.Sorts)
	withSorts.Enter()
	defer withSorts.Exit()
	leaveTheory := component.TheoryContext()
	defer leaveTheory()
	if err := exportVMTModel("vmt", component, false, predicates); err != nil {
		return err
	}
	return errRankingModelExported
}
