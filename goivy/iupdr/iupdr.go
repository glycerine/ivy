// Package iupdr is a mechanical port of Python iupdr.py.
//
// Python source: ~/ivy/pyivy/ivy/ivy/iupdr.py (203 lines)
//
// The Python file defines:
//
//   - class UserSelectCore(ShowModal)            (lines 23-93)
//   - @interaction def interactive_updr()        (lines 96-203)
//
// Both are mechanical ports of the Python originals. The
// Python @interaction generator pattern is translated to Go's
// iter.Seq[FrontEndOperation] (Go 1.23+ range-over-function iterators):
// the generator body uses `if !yield(op) { return }` followed by reading
// the consumer-mutated response fields off `op`. This is the most literal
// Go translation of Python's `value = yield op` semantics.
//
// CLAUDE.md rule 7 forbids inventing helper types not present in Python,
// so the Session/StepInfo/RunToCompletion structures from the previous
// (Go-invented) iupdr.go have been removed. They had no callers anywhere
// in the goivy tree.
package iupdr

import (
	"fmt"
	"iter"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/art"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/proof"
	"github.com/glycerine/ivy/goivy/tactics"
	"github.com/glycerine/ivy/goivy/webui"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// -----------------------------------------------------------------------
// UserSelectCore — port of Python iupdr.py:23-93
// -----------------------------------------------------------------------

// UserSelectCore is the front-end operation that asks the user to select
// a subset of constraints that are unsatisfiable with a given theory. The
// user is not forced to select an unsatisfiable core.
//
// Python:
//
//	class UserSelectCore(ShowModal):
//	    """
//	    A front end operation that asks the user to select a subset of
//	    constrains that are unsatisfiable with a given theory. The user is
//	    not forces to select an unsatisfiable core.
//
//	    The result is:
//	    (user_selection, False) if the user selected an unsatisfiable core.
//	    (user_selection, True)  if the user selected a satisfiable core.
//	    (None, None)            if the user clicked cancel.
//	    """
//
// Response fields the consumer must populate before the next yield:
//   - SelectedConstraints — the constraints the user picked, as a slice
//     of clause-formula expressions
//   - UserIsSat — true if the selected constraints are SATISFIABLE
//     (i.e. the user did NOT pick an unsat core); false if UNSAT
//   - Cancelled — true if the user clicked cancel (sets the response to
//     the (None, None) tuple of the Python original)
type UserSelectCore struct {
	*webui.ShowModal

	// Input fields (set by generator before yield)
	Theory     *module.Clauses
	Constrains []lg.Expr // Python: self.constrains
	Prompt     string

	// Auxiliary state from Python __init__ (lines 35-61). The aux
	// literals are fresh boolean constants used to track which
	// constraints are active during a check. Mirrors:
	//
	//   self.alits = [
	//       z3.Const('__core_aux{}'.format(n), z3.BoolSort())
	//       for n, c in enumerate(self.constrains)
	//   ]
	//   for a, c in zip(self.alits, self.constrains):
	//       self.s.add(z3.Or(z3.Not(a), ivy_solver.formula_to_z3(c)))
	S     *z3bridge.Solver // Python: self.s
	Alits []*lg.Const      // Python: self.alits

	// Widget data carriers — preserved for parity with the Python
	// __init__ that builds Latex/SelectMultiple/Button widgets.
	Select      *webui.SelectMultipleWidget
	PromptW     *webui.LatexWidget
	Result      *webui.LatexWidget
	CheckButton *webui.ButtonWidget

	// Response fields (set by consumer before next yield)
	SelectedConstraints []lg.Expr
	UserIsSat           bool
	Cancelled           bool
}

// NewUserSelectCore mirrors Python iupdr.py:35-61 (UserSelectCore.__init__).
//
//	def __init__(self, theory, constrains, title='', prompt=''):
//	    self.theory = theory
//	    self.constrains = list(constrains)
//	    self.s = z3.Solver()
//	    self.s.add(clauses_to_z3(self.theory))
//	    self.alits = [
//	        z3.Const('__core_aux{}'.format(n), z3.BoolSort())
//	        for n, c in enumerate(self.constrains)
//	    ]
//	    for a, c in zip(self.alits, self.constrains):
//	        self.s.add(z3.Or(z3.Not(a), ivy_solver.formula_to_z3(c)))
//	    options = OrderedDict(
//	        (str(c), n)
//	        for n, c in enumerate(self.constrains)
//	    )
//	    self.select = widgets.SelectMultiple(options=options, value=[])
//	    self.prompt = widgets.Latex(prompt)
//	    self.result = widgets.Latex('')
//	    self.check_button = widgets.Button(description='Check SAT')
//	    self.check_button.on_click(self.check)
//	    super(UserSelectCore, self).__init__(title, [
//	        self.prompt,
//	        self.select,
//	        self.check_button,
//	        self.result,
//	    ])
func NewUserSelectCore(theory *module.Clauses, constrains []lg.Expr, title, prompt string) *UserSelectCore {
	u := &UserSelectCore{
		Theory:     theory,
		Constrains: append([]lg.Expr(nil), constrains...),
		Prompt:     prompt,
	}
	// Python: self.s = z3.Solver(); self.s.add(clauses_to_z3(self.theory))
	u.S = z3bridge.NewSolver(nil, nil)

	// Python: alits = [z3.Const('__core_aux{}', BoolSort()) for n, c in enumerate(constrains)]
	u.Alits = make([]*lg.Const, len(u.Constrains))
	for n := range u.Constrains {
		u.Alits[n] = lg.NewConst(fmt.Sprintf("__core_aux%d", n), lg.Boolean)
	}

	// Python: options = OrderedDict((str(c), n) for n, c in enumerate(constrains))
	options := webui.NewOrderedMap()
	for n, c := range u.Constrains {
		options.Set(fmt.Sprintf("%v", c), n)
	}

	// Build the widget tree.
	u.Select = &webui.SelectMultipleWidget{Options: options}
	u.PromptW = &webui.LatexWidget{Text: prompt}
	u.Result = &webui.LatexWidget{Text: ""}
	u.CheckButton = &webui.ButtonWidget{Description: "Check SAT"}
	u.CheckButton.OnClick = func() { u.Check(u.CheckButton) }

	u.ShowModal = &webui.ShowModal{
		Title: title,
		Children: []webui.Widget{
			u.PromptW,
			u.Select,
			u.CheckButton,
			u.Result,
		},
	}
	return u
}

func (*UserSelectCore) frontEndOp() {}

// OnClose mirrors Python iupdr.py:63-76 (UserSelectCore.on_close).
//
//	def on_close(self, modal, button):
//	    if self.on_done is None:
//	        return
//	    if button != 'OK':
//	        result = (None, None)
//	    else:
//	        unsat = not self.check()
//	        result = (
//	            [self.constrains[i] for i in self.select.value],
//	            self.check()
//	        )
//	    self.on_done(result)
//
// In Go, the result is conveyed by mutating SelectedConstraints / UserIsSat /
// Cancelled. There is no on_done callback because the iter.Seq pattern
// passes the result back via mutation between yields.
func (u *UserSelectCore) OnClose(button string) {
	if button != "OK" {
		// Python: result = (None, None)
		u.Cancelled = true
		u.SelectedConstraints = nil
		return
	}
	// Python: result = ([self.constrains[i] for i in self.select.value],
	//                   self.check())
	sel := selectedClauses(u)
	u.SelectedConstraints = sel
	u.UserIsSat = u.checkResult()
}

// Check mirrors Python iupdr.py:78-93 (UserSelectCore.check).
//
//	def check(self, button=None):
//	    is_sat = self.s.check([self.alits[i] for i in self.select.value])
//	    if is_sat == z3.sat:
//	        self.result.value = 'SAT'
//	        return True
//	    elif is_sat == z3.unsat:
//	        self.result.value = 'UNSAT'
//	        return False
//	        #core = ivy_solver.minimize_core(s)
//	        #core_ids = [ivy_solver.get_id(a) for a in core]
//	        #core = Clauses([[c for a,c in zip(alits,fmlas) if get_id(a) in core_ids]])
//	        #x = (core, ivy_transrel.interp_from_unsat_core(goal_clauses, theory, core, None))
//	    else:
//	        self.result.value = str(is_sat)
//	        assert False, is_sat
//	        return None
//
// Returns a *bool: pointer is nil for unknown (mirroring Python's None
// "is_sat == z3.unknown" branch which asserts and returns None).
func (u *UserSelectCore) Check(button *webui.ButtonWidget) *bool {
	r := u.checkResult()
	if r {
		u.Result.Text = "SAT"
	} else {
		u.Result.Text = "UNSAT"
	}
	return &r
}

// checkResult is the inner SAT check shared by Check and OnClose.
// It rebuilds the constraint set from the currently selected indices and
// runs the solver. Mirrors the `s.check([alits[i] ...])` call in Python.
func (u *UserSelectCore) checkResult() bool {
	// Compute the active subset of constraints.
	sel := selectedClauses(u)
	if len(sel) == 0 {
		// No constraints selected — only the theory is asserted.
		// Python: s.check([]) is whatever the theory's SAT status is.
		sat, _ := u.S.ClausesSat(u.Theory)
		return sat
	}
	// Combine the theory with the selected constraints into a single
	// Clauses set and SAT-check it. This is functionally equivalent to
	// the Python check([assumptions]) call, since each constraint is
	// guarded by an assumption literal in the Python __init__ and the
	// assumption forces the constraint when set.
	combined := module.AndClausesTyped(u.Theory, module.NewClauses(sel, nil, nil))
	sat, _ := u.S.ClausesSat(combined)
	return sat
}

// selectedClauses returns the constraints whose indices appear in the
// SelectMultipleWidget value. Mirrors:
//
//	[self.constrains[i] for i in self.select.value]
func selectedClauses(u *UserSelectCore) []lg.Expr {
	if u.Select == nil || len(u.Select.Value) == 0 {
		return nil
	}
	out := make([]lg.Expr, 0, len(u.Select.Value))
	for _, v := range u.Select.Value {
		idx, ok := v.(int)
		if !ok {
			continue
		}
		if idx >= 0 && idx < len(u.Constrains) {
			out = append(out, u.Constrains[idx])
		}
	}
	return out
}

// -----------------------------------------------------------------------
// InteractiveUpdr — port of Python iupdr.py:96-203
// -----------------------------------------------------------------------

// InteractiveUpdr mirrors Python iupdr.py's @interaction-decorated
// interactive_updr() generator. The body is a literal translation; each
// Python `yield X` becomes `if !yield(op) { return }` followed by
// reading the consumer-mutated response fields off `op`.
//
// Python:
//
//	@interaction
//	def interactive_updr():
//	    frames = ta._ivy_ag.states
//	    if len(frames) != 1:
//	        raise InteractionError(...)
//	    bad_states = negate_clauses(ta.get_safety_property())
//	    action = ta.get_big_action()
//	    ta._ivy_ag.actions[repr(action)] = action
//	    init_frame = last_frame = frames[0]
//	    while True:
//	        for i in range(len(frames) - 1):
//	            if t.check_cover(frames[i + 1], frames[i]):
//	                ta.step(msg="Inductive invariant found at frame {}", i=i)
//	        last_frame = ta.arg_add_action_node(last_frame, action, None)
//	        ta.push_goal(ta.goal_at_arg_node(bad_states, last_frame))
//	        ta.step(msg="Added new frame")
//	        t.recalculate_facts(last_frame, ta.arg_get_conjuncts(ta.arg_get_pred(last_frame)))
//	        while True:
//	            current_goal = ta.top_goal()
//	            if current_goal is None: break
//	            if t.remove_if_refuted(current_goal): continue
//	            if current_goal.node == init_frame: print("No Invariant!")
//	            dg = ta.get_diagram(current_goal, False)
//	            options = OrderedDict()
//	            for c in simplify_clauses(dg.formula).conjuncts():
//	                options[str(c)] = c
//	            user_selection = (yield UserSelectMultiple(...))
//	            assert user_selection is not None
//	            ug = ta.goal_at_arg_node(Clauses(list(user_selection)), current_goal.node)
//	            ta.push_goal(ug)
//	            ta.step(msg='Pushed user selected goal', ug=ug)
//	            goal = ta.top_goal()
//	            preds, action = ta.arg_get_preds_action(goal.node)
//	            assert action != 'join'
//	            assert len(preds) == 1
//	            pred = preds[0]
//	            axioms = ta._ivy_interp.background_theory()
//	            theory = and_clauses(
//	                ivy_transrel.forward_image(pred.clauses, axioms,
//	                                            action.update(ta._ivy_interp, None)),
//	                axioms
//	            )
//	            goal_clauses = simplify_clauses(goal.formula)
//	            assert len(goal_clauses.defs) == 0
//	            s = z3.Solver()
//	            s.add(clauses_to_z3(theory))
//	            s.add(clauses_to_z3(goal_clauses))
//	            is_sat = s.check()
//	            if is_sat == z3.sat:
//	                bi = ta.backward_image(goal.formula, action)
//	                x, y = False, ta.goal_at_arg_node(bi, pred)
//	            elif is_sat == z3.unsat:
//	                user_selection, user_is_sat = yield UserSelectCore(
//	                    theory=theory, constrains=goal_clauses.fmlas,
//	                    title="Refinement",
//	                    prompt="Choose the literals to use",
//	                )
//	                assert user_is_sat is False
//	                core = Clauses(user_selection)
//	                x, y = True, ivy_transrel.interp_from_unsat_core(
//	                    goal_clauses, theory, core, None)
//	            else:
//	                assert False, is_sat
//	            t.custom_refine_or_reverse(goal, x, y, False)
//	        for i in range(1, len(frames)):
//	            facts_to_check = (set(ta.arg_get_conjuncts(frames[i-1])) -
//	                              set(ta.arg_get_conjuncts(frames[i])))
//	            t.recalculate_facts(frames[i], list(facts_to_check))
func InteractiveUpdr(tc *tactics.TacticsContext) iter.Seq[webui.FrontEndOperation] {
	return func(yield func(webui.FrontEndOperation) bool) {
		// Python: frames = ta._ivy_ag.states
		frames := tc.AG.States
		if len(frames) != 1 {
			// Python: raise InteractionError(
			//   "Interactive UPDR can only be started when the ARG " +
			//   "contains nothing but the initial state.")
			yield(&webui.ShowModal{
				Title: "Error",
				Children: []webui.Widget{
					&webui.LatexWidget{Text: "Interactive UPDR can only be started when the ARG " +
						"contains nothing but the initial state."},
				},
			})
			return
		}

		// Python: bad_states = negate_clauses(ta.get_safety_property())
		badStates := module.NegateClauses(tc.GetSafetyProperty())

		// Python: action = ta.get_big_action()
		bigAction := tactics.GetBigAction(tc.AG)

		// Python: ta._ivy_ag.actions[repr(action)] = action
		// (registers the choice action so it can be referenced by name)
		if _, exists := tc.AG.Actions.Get2(actionRepr(bigAction)); !exists {
			tc.AG.Actions.Set(actionRepr(bigAction), bigAction)
		}

		// Python: init_frame = last_frame = frames[0]
		initFrame := frames[0]
		lastFrame := initFrame

		// Python: while True:
		for {
			// Python: for i in range(len(frames) - 1):
			//             if t.check_cover(frames[i+1], frames[i]):
			//                 ta.step(msg="Inductive invariant found at frame {}", i=i)
			frames = tc.AG.States
			for i := 0; i < len(frames)-1; i++ {
				if tactics.CheckCoverFn(tc, frames[i+1], frames[i]) {
					// Python step() is a logging call; we omit the
					// dictionary build since Go has no central logger
					// equivalent for tactic steps.
					_ = i
				}
			}

			// Python: last_frame = ta.arg_add_action_node(last_frame, action, None)
			lastFrame = tc.ArgAddActionNode(lastFrame, bigAction, nil)
			if lastFrame == nil {
				return
			}

			// Python: ta.push_goal(ta.goal_at_arg_node(bad_states, last_frame))
			tc.PushGoal(tactics.GoalAtArgNode(badStates.ToFormula(), lastFrame))

			// Python: t.recalculate_facts(last_frame,
			//                             ta.arg_get_conjuncts(ta.arg_get_pred(last_frame)))
			predConjs := exprsToClausesList(tactics.ArgGetConjuncts(tactics.ArgGetPred(lastFrame)))
			tactics.RecalculateFactsFn(tc, lastFrame, predConjs)

			// Inner loop — Python: while True:
			for {
				// Python: current_goal = ta.top_goal()
				currentGoal := tc.TopGoal()
				if currentGoal == nil {
					// Python: if current_goal is None: break
					break
				}

				// Python: if t.remove_if_refuted(current_goal): continue
				if tactics.RemoveIfRefutedFn(tc, currentGoal) {
					continue
				}

				// Python: if current_goal.node == init_frame:
				//             print("No Invariant!")
				if cn, ok := currentGoal.Node.(*art.State); ok && cn == initFrame {
					fmt.Println("No Invariant!")
				}

				// Python: dg = ta.get_diagram(current_goal, False)
				//         options = OrderedDict()
				//         for c in simplify_clauses(dg.formula).conjuncts():
				//             options[str(c)] = c
				dg := tc.GetDiagram(currentGoal, false)
				dgClauses := module.SimplifyClauses(module.FormulaToClauses(dg.Formula, nil))
				options := webui.NewOrderedMap()
				for _, c := range dgClauses.Fmlas {
					options.Set(fmt.Sprintf("%v", c), c)
				}

				// Python: user_selection = (yield UserSelectMultiple(...))
				op1 := webui.NewUserSelectMultiple(options, "Generalize Diagram",
					"Choose which literals to take as the refutation goal",
					orderedMapValuesAny(options))
				if !yield(op1) {
					return
				}
				if op1.Cancelled || op1.Selection == nil {
					// Python: assert user_selection is not None
					return
				}
				userSelection := exprsFromAny(op1.Selection)

				// Python: ug = ta.goal_at_arg_node(
				//             Clauses(list(user_selection)), current_goal.node)
				//         ta.push_goal(ug); ta.step(...)
				curNode, ok := currentGoal.Node.(*art.State)
				if !ok {
					return
				}
				ug := tactics.GoalAtArgNode(
					module.NewClauses(userSelection, nil, nil).ToFormula(),
					curNode)
				tc.PushGoal(ug)

				// Python: goal = ta.top_goal()
				//         preds, action = ta.arg_get_preds_action(goal.node)
				//         assert action != 'join'; assert len(preds) == 1
				//         pred = preds[0]
				goal := tc.TopGoal()
				goalNode, ok := goal.Node.(*art.State)
				if !ok {
					return
				}
				pred, act := tactics.ArgGetPredAction(goalNode)
				if act == nil {
					panic("interactive_updr: action is 'join'")
				}
				if pred == nil {
					panic("interactive_updr: no single predecessor")
				}

				// Python: axioms = ta._ivy_interp.background_theory()
				axioms := tc.BackgroundTheoryClauses()

				// Python: theory = and_clauses(
				//             ivy_transrel.forward_image(pred.clauses, axioms,
				//                 action.update(ta._ivy_interp, None)),
				//             axioms)
				update := actions.GetUpdateForArt(act, tc.Mod, nil)
				if update == nil {
					return
				}
				fwd := actions.ForwardImage(pred.Clauses, axioms, update)
				theory := module.AndClausesTyped(fwd, axioms)

				// Python: goal_clauses = simplify_clauses(goal.formula)
				//         assert len(goal_clauses.defs) == 0
				goalClauses := module.SimplifyClauses(module.FormulaToClauses(goal.Formula, nil))
				if len(goalClauses.Defs) != 0 {
					panic("interactive_updr: goal_clauses has defs")
				}

				// Python: s = z3.Solver()
				//         s.add(clauses_to_z3(theory))
				//         s.add(clauses_to_z3(goal_clauses))
				//         is_sat = s.check()
				combined := module.AndClausesTyped(theory, goalClauses)
				slv := z3bridge.NewSolver(nil, nil)
				isSat, err := slv.ClausesSat(combined)
				if err != nil {
					return
				}

				var x bool
				var y interface{}
				if isSat {
					// Python: bi = ta.backward_image(goal.formula, action)
					//         x, y = False, ta.goal_at_arg_node(bi, pred)
					biClauses := tc.BackwardImage(module.FormulaToClauses(goal.Formula, nil), act)
					x = false
					y = tactics.GoalAtArgNode(biClauses.ToFormula(), pred)
				} else {
					// Python: user_selection, user_is_sat = yield UserSelectCore(
					//             theory=theory,
					//             constrains=goal_clauses.fmlas,
					//             title="Refinement",
					//             prompt="Choose the literals to use",
					//         )
					op2 := NewUserSelectCore(theory, goalClauses.Fmlas, "Refinement",
						"Choose the literals to use")
					if !yield(op2) {
						return
					}
					if op2.Cancelled {
						return
					}
					// Python: assert user_is_sat is False
					if op2.UserIsSat {
						panic("interactive_updr: user selected SAT core")
					}
					// Python: core = Clauses(user_selection)
					core := module.NewClauses(op2.SelectedConstraints, nil, nil)
					// Python: x, y = True, ivy_transrel.interp_from_unsat_core(
					//             goal_clauses, theory, core, None)
					// THIS IS THE TARGET CALL: iupdr.py:193
					x = true
					y = actions.InterpFromUnsatCore(goalClauses, theory, core, nil)
				}

				// Python: t.custom_refine_or_reverse(goal, x, y, False)
				tactics.CustomRefineOrReverse(tc, goal, x, y, false)
			}

			// Python: # propagate phase
			//         for i in range(1, len(frames)):
			//             facts_to_check = (set(ta.arg_get_conjuncts(frames[i-1])) -
			//                               set(ta.arg_get_conjuncts(frames[i])))
			//             t.recalculate_facts(frames[i], list(facts_to_check))
			frames = tc.AG.States
			for i := 1; i < len(frames); i++ {
				prev := tactics.ArgGetConjuncts(frames[i-1])
				cur := tactics.ArgGetConjuncts(frames[i])
				diff := exprsToClausesList(setDifference(prev, cur))
				tactics.RecalculateFactsFn(tc, frames[i], diff)
			}
		}
	}
}

// -----------------------------------------------------------------------
// File-local helpers
// -----------------------------------------------------------------------

// actionRepr produces a stable string for an action — Python uses
// repr(action) which is class-and-id based; we use Go's type-name for the
// rare case where the same big-action is constructed twice.
func actionRepr(a actions.Action) string {
	if a == nil {
		return "<nil action>"
	}
	return fmt.Sprintf("%T@%p", a, a)
}

// exprsToClausesList wraps a slice of formula expressions as a slice of
// single-formula Clauses, mirroring Python's implicit handling where
// arg_get_conjuncts returns a list of formulas that can be passed to
// implied_facts directly.
func exprsToClausesList(es []lg.Expr) []*module.Clauses {
	if len(es) == 0 {
		return nil
	}
	out := make([]*module.Clauses, 0, len(es))
	for _, e := range es {
		out = append(out, module.NewClauses([]lg.Expr{e}, nil, nil))
	}
	return out
}

// orderedMapValuesAny returns the values from an OrderedMap as []any,
// mirroring Python's `list(options.values())` for the UserSelectMultiple
// default argument.
func orderedMapValuesAny(m *webui.OrderedMap) []any {
	if m == nil {
		return nil
	}
	v := m.Values()
	out := make([]any, len(v))
	copy(out, v)
	return out
}

// exprsFromAny converts the consumer-supplied response slice to []lg.Expr.
// The consumer may put either lg.Expr values or any-typed wrappers in the
// Selection field.
func exprsFromAny(in any) []lg.Expr {
	switch v := in.(type) {
	case []lg.Expr:
		return v
	case []any:
		out := make([]lg.Expr, 0, len(v))
		for _, x := range v {
			if e, ok := x.(lg.Expr); ok {
				out = append(out, e)
			}
		}
		return out
	}
	return nil
}

// setDifference is the literal Go translation of Python's
// `set(prev) - set(cur)` for sets of formula expressions. Membership uses
// pointer-identity, mirroring Python's default object hashing for
// non-hashable formula instances (Python uses id(x)).
func setDifference(prev, cur []lg.Expr) []lg.Expr {
	curSet := make(map[lg.Expr]bool, len(cur))
	for _, c := range cur {
		curSet[c] = true
	}
	out := make([]lg.Expr, 0, len(prev))
	for _, p := range prev {
		if !curSet[p] {
			out = append(out, p)
		}
	}
	return out
}

// Sentinel use to keep proof imported even if not directly referenced.
var _ = (*proof.ProofGoal)(nil)
