package ivy2cpp

import (
	"os"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

type actionUpdateWorld struct {
	mod      *goivy.Module
	color    goivy.Sort
	node     goivy.Sort
	idx      goivy.Sort
	red      *goivy.Const
	green    *goivy.Const
	saved    *goivy.Const
	flag     *goivy.Const
	marked   *goivy.Const
	obj      *goivy.Const
	otherObj *goivy.Const
	field    *goivy.Const
	srcField *goivy.Const
}

type actionUpdateCase struct {
	name  string
	build func(t *testing.T, w *actionUpdateWorld) goivy.Action
}

func newActionUpdateWorld(t *testing.T) *actionUpdateWorld {
	t.Helper()

	mod := goivy.New()
	mod.Name = "action_update_matrix"
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	node := &goivy.UninterpretedSort{Name: "node"}
	idx := &goivy.RangeSort{
		Name: "idx",
		Lb:   goivy.NumeralBound{Value: "0"},
		Ub:   goivy.NumeralBound{Value: "2"},
	}
	for _, sort := range []goivy.Sort{color, node, idx} {
		if err := mod.Sig.AddSort(sort); err != nil {
			t.Fatalf("AddSort(%s): %v", sort, err)
		}
	}

	red := addActionUpdateSymbol(t, mod, "red", color, false)
	green := addActionUpdateSymbol(t, mod, "green", color, false)
	saved := addActionUpdateSymbol(t, mod, "saved", color, true)
	flag := addActionUpdateSymbol(t, mod, "flag", goivy.Boolean, true)
	marked := addActionUpdateSymbol(t, mod, "marked", goivy.LogicRelationSort([]goivy.Sort{color}), true)
	obj := addActionUpdateSymbol(t, mod, "obj", node, true)
	otherObj := addActionUpdateSymbol(t, mod, "other_obj", node, true)
	field := addActionUpdateSymbol(t, mod, "field", goivy.LogicRelationSort([]goivy.Sort{node, color}), true)
	srcField := addActionUpdateSymbol(t, mod, "src_field", goivy.LogicRelationSort([]goivy.Sort{node, color}), true)

	schemaLabel := mod.Cfg.AstCfg.NewLabeledFormula(goivy.NewConst("schema", goivy.Boolean), goivy.True)
	mod.Schemata.Set("schema", schemaLabel)

	return &actionUpdateWorld{
		mod:      mod,
		color:    color,
		node:     node,
		idx:      idx,
		red:      red,
		green:    green,
		saved:    saved,
		flag:     flag,
		marked:   marked,
		obj:      obj,
		otherObj: otherObj,
		field:    field,
		srcField: srcField,
	}
}

func addActionUpdateSymbol(t *testing.T, mod *goivy.Module, name string, sort goivy.Sort, mutable bool) *goivy.Const {
	t.Helper()
	c, err := mod.Sig.AddSymbol(name, sort)
	if err != nil {
		t.Fatalf("AddSymbol(%s): %v", name, err)
	}
	if mutable {
		if sort == goivy.Boolean || goivy.IsRelationalSort(sort) {
			mod.Relations.Set(name, sort)
		} else {
			mod.Functions.Set(name, sort)
		}
	}
	return c
}

func actionUpdateCases() []actionUpdateCase {
	return []actionUpdateCase{
		{name: "assume", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewAssumeAction(goivy.True)
		}},
		{name: "assert", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewAssertAction(goivy.True)
		}},
		{name: "subgoal", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewSubgoalAction(goivy.True)
		}},
		{name: "requires", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewRequiresAction(goivy.True)
		}},
		{name: "ensures", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewEnsuresAction(goivy.True)
		}},
		{name: "assign_scalar", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewAssignAction(w.saved, w.red)
		}},
		{name: "assign_relation_cell", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewAssignAction(goivy.MustApply(w.marked, w.red), goivy.True)
		}},
		{name: "havoc_scalar", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewHavocAction(w.saved)
		}},
		{name: "set_relation_positive", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewSetAction(goivy.MustApply(w.marked, w.red))
		}},
		{name: "set_relation_negative", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewSetAction(&goivy.LogicNot{Body: goivy.MustApply(w.marked, w.red)})
		}},
		{name: "native", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewNativeAction(w.mod.Cfg.AstCfg.NewNativeCode("// noop"))
		}},
		{name: "debug", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewDebugAction(goivy.NewConst(`"tick"`, goivy.TopS), w.flag)
		}},
		{name: "return", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewReturnAction()
		}},
		{name: "assign_field", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewAssignFieldAction(w.field, w.obj, w.green)
		}},
		{name: "null_field", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewNullFieldAction(w.field, w.obj)
		}},
		{name: "copy_field", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewCopyFieldAction(w.otherObj, w.field, w.obj, w.srcField)
		}},
		{name: "sequence", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewSequence(
				goivy.NewAssignAction(w.saved, w.red),
				goivy.NewAssumeAction(w.flag),
			)
		}},
		{name: "choice", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewChoiceActionOn(w.mod.Cfg.ActCfg,
				goivy.NewAssignAction(w.saved, w.red),
				goivy.NewAssignAction(w.saved, w.green))
		}},
		{name: "choice_determinized", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			w.mod.Cfg.ActCfg.Determinize = true
			return goivy.NewChoiceActionOn(w.mod.Cfg.ActCfg,
				goivy.NewAssignAction(w.saved, w.red),
				goivy.NewAssignAction(w.saved, w.green))
		}},
		{name: "env", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewEnvActionOn(w.mod.Cfg.ActCfg,
				goivy.NewAssignAction(w.saved, w.red),
				goivy.NewAssignAction(w.saved, w.green))
		}},
		{name: "env_determinized", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			w.mod.Cfg.ActCfg.Determinize = true
			return goivy.NewEnvActionOn(w.mod.Cfg.ActCfg,
				goivy.NewAssignAction(w.saved, w.red),
				goivy.NewAssignAction(w.saved, w.green))
		}},
		{name: "if_boolean", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewIfAction(w.flag,
				goivy.NewAssignAction(w.saved, w.red),
				goivy.NewAssignAction(w.saved, w.green))
		}},
		{name: "if_some", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			p := goivy.NewConst("picked", w.color)
			cond := &goivy.SomeCondition{Params: []*goivy.Const{p}, Fmla: goivy.True, Kind: "some"}
			return goivy.NewIfAction(cond,
				goivy.NewAssignAction(w.saved, p),
				goivy.NewAssignAction(w.saved, w.green))
		}},
		{name: "while", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewWhileAction(goivy.False, goivy.NewAssignAction(w.saved, w.red), goivy.True)
		}},
		{name: "local", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			local := goivy.NewConst("local_color", w.color)
			return goivy.NewLocalActionOn(w.mod.Cfg.ActCfg, "test.local",
				local,
				goivy.NewAssignAction(local, w.red))
		}},
		{name: "let", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			alias := goivy.NewConst("alias_color", w.color)
			binding := goivy.NewEqualsNode(alias, w.green)
			return goivy.NewLetAction(binding, goivy.NewAssignAction(w.saved, alias))
		}},
		{name: "call_with_input", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			param := goivy.NewConst("param_color", w.color)
			callee := goivy.NewAssignAction(w.saved, param)
			callee.SetFormalParams([]*goivy.Const{param})
			w.mod.Actions.Set("callee_in", callee)
			return goivy.NewCallActionOn(w.mod.Cfg.ActCfg, goivy.MustApply(goivy.NewConst("callee_in", goivy.TopS), w.red))
		}},
		{name: "call_with_return", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			param := goivy.NewConst("param_color", w.color)
			ret := goivy.NewConst("ret_color", w.color)
			callee := goivy.NewAssignAction(ret, param)
			callee.SetFormalParams([]*goivy.Const{param})
			callee.SetFormalReturns([]*goivy.Const{ret})
			w.mod.Actions.Set("callee_out", callee)
			return goivy.NewCallActionOn(w.mod.Cfg.ActCfg, goivy.MustApply(goivy.NewConst("callee_out", goivy.TopS), w.green), w.saved)
		}},
		{name: "bind_olds", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewBindOldsAction(goivy.NewAssignAction(w.saved, w.red))
		}},
		{name: "crash", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewCrashAction(goivy.NewConst("anything", goivy.TopS))
		}},
		{name: "fail", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewFailAction(goivy.NewAssignAction(w.saved, w.red))
		}},
		{name: "instantiate_schema", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewInstantiateAction(goivy.NewConst("schema", goivy.TopS))
		}},
		{name: "ignore_default_update", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewIgnoreAction()
		}},
		{name: "thunk_default_update", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewThunkAction(goivy.NewConst("handler", goivy.TopS), goivy.NewSequence())
		}},
		{name: "ranking_default_update", build: func(t *testing.T, w *actionUpdateWorld) goivy.Action {
			return goivy.NewRanking(goivy.NewConst("rank_rel", goivy.TopS), w.saved)
		}},
	}
}

func TestGetUpdateForArtCoversLegalActionMatrixWithoutPanics(t *testing.T) {
	for _, tc := range actionUpdateCases() {
		t.Run(tc.name, func(t *testing.T) {
			w := newActionUpdateWorld(t)
			act := tc.build(t, w)
			var upd *goivy.Update
			panicked, panicValue := captureActionUpdatePanic(func() {
				upd = goivy.GetUpdateForArt(act, w.mod, nil)
			})
			if panicked {
				t.Fatalf("GetUpdateForArt panicked for %s: %v", tc.name, panicValue)
			}
			if upd == nil {
				t.Fatalf("GetUpdateForArt returned nil for %s", tc.name)
			}
			if upd.TR == nil {
				t.Fatalf("GetUpdateForArt returned nil TR for %s: %+v", tc.name, upd)
			}
			if upd.Pre == nil {
				t.Fatalf("GetUpdateForArt returned nil Pre for %s: %+v", tc.name, upd)
			}
		})
	}
}

func TestActionGenPlanCoversLegalActionMatrixWithoutFallback(t *testing.T) {
	for _, tc := range actionUpdateCases() {
		t.Run(tc.name, func(t *testing.T) {
			w := newActionUpdateWorld(t)
			act := tc.build(t, w)
			g := &Generator{
				Mod:       w.mod,
				ClassName: "actgen",
				Config:    Config{Target: "gen", ClassName: "actgen"},
			}
			plan := g.buildActionGenPlan(tc.name, act)
			if plan == nil {
				t.Fatalf("buildActionGenPlan returned nil for %s", tc.name)
			}
			if strings.HasPrefix(plan.fallbackReason, "GetUpdate panicked") {
				t.Fatalf("buildActionGenPlan still converted a panic into fallback for %s: %s", tc.name, plan.fallbackReason)
			}
			if plan.fallback {
				t.Fatalf("buildActionGenPlan unexpectedly fell back for %s: %s", tc.name, plan.fallbackReason)
			}
			if plan.origPreClauses == nil {
				t.Fatalf("buildActionGenPlan did not keep original pre clauses for %s", tc.name)
			}
			if plan.preFmla == nil {
				t.Fatalf("buildActionGenPlan did not produce a pre formula for %s", tc.name)
			}
		})
	}
}

func TestActionGenPlanDoesNotRecoverIllegalActionPanics(t *testing.T) {
	w := newActionUpdateWorld(t)
	bad := goivy.NewSequence(goivy.NewConst("not_an_action", goivy.TopS))
	g := &Generator{
		Mod:       w.mod,
		ClassName: "actgen",
		Config:    Config{Target: "gen", ClassName: "actgen"},
	}

	panicked, _ := captureActionUpdatePanic(func() {
		_ = g.buildActionGenPlan("bad", bad)
	})
	if !panicked {
		t.Fatalf("buildActionGenPlan should let illegal action panics propagate instead of falling back")
	}
}

func TestActionGenSourceHasNoRecoverFallback(t *testing.T) {
	b, err := os.ReadFile("action_gen.go")
	if err != nil {
		t.Fatalf("read action_gen.go: %v", err)
	}
	src := string(b)
	for _, forbidden := range []string{"recover()", "GetUpdate panicked"} {
		if strings.Contains(src, forbidden) {
			t.Fatalf("action_gen.go still contains %q", forbidden)
		}
	}
}

func captureActionUpdatePanic(fn func()) (panicked bool, value interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			value = r
		}
	}()
	fn()
	return false, nil
}
