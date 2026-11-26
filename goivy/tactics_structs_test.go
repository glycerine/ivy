package goivy

import "testing"

func makePairStruct(t *testing.T) (*Module, *Const, *Const, *Const, *Const) {
	t.Helper()
	mod := New()
	pair := &UninterpretedSort{Name: "pair"}
	selector, err := NewFunctionSort(pair, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	constructor, err := NewFunctionSort(Boolean, Boolean, pair)
	if err != nil {
		t.Fatal(err)
	}
	left, right := NewConst("left", selector), NewConst("right", selector)
	makePair := NewConst("make_pair", constructor)
	value := NewConst("cell", pair)
	mod.Sig.Sorts.Set(pair.Name, pair)
	mod.DestructorSorts[left.Name], mod.DestructorSorts[right.Name] = pair, pair
	mod.ConstructorSorts[makePair.Name] = pair
	mod.SortDestructors.Set(pair.Name, []*Const{left, right})
	mod.SortConstructors.Set(pair.Name, []*Const{makePair})
	return mod, value, makePair, left, right
}

func TestConstructorNormalizationSharesFieldsAndPreservesBinderScope(t *testing.T) {
	mod, pair, ctor, left, right := makePairStruct(t)
	mod.Attributes.Set("constructors", mod.Cfg.AstCfg.NewAtom("true"))
	condition := &LogicAnd{Terms: []Expr{NewApplyUnchecked(left, pair), NewApplyUnchecked(right, pair)}}
	flag := NewConst("flag", Boolean)
	update := computeUpdate(mod, NewAssignAction(flag, condition))
	count := 0
	for _, formula := range update.TR.Fmlas {
		for _, symbol := range UsedSymbolsAST(formula).All() {
			switch ExprName(symbol) {
			case left.Name, right.Name:
				t.Fatalf("untranslated selector in ground update: %s", formula)
			case ctor.Name:
				count++
			}
		}
	}
	if count != 1 {
		t.Fatalf("constructor occurs in %d update clauses, want one", count)
	}
	variable := &LogicVariable{Name: "X", VSort: pair.CSort}
	bound := &LogicExists{Variables: []*LogicVariable{variable}, Body: NewApplyUnchecked(left, variable)}
	clauses := NewClauses([]Expr{bound}, nil, nil)
	if got := eliminateStructDestructors(mod, clauses); got != clauses {
		t.Fatal("constructor normalization changed a bound selector")
	}
}

func TestFlattenStructsPreservesParallelAssignment(t *testing.T) {
	mod, cell, ctor, left, right := makePairStruct(t)
	body := NewAssignAction(cell, NewApplyUnchecked(ctor, NewApplyUnchecked(right, cell), NewApplyUnchecked(left, cell)))
	program := &NormalProgram{Bindings: []*ActionTermBinding{{Name: "swap", Action: &ActionTerm{Stmt: body}}}, Calls: []string{"swap"}}
	cfg := mod.Cfg.AstCfg
	goal := cfg.NewLabeledFormula(cfg.NewAtom("swap"), cfg.NewTemporalModels(program, True))
	checker := NewProofChecker(mod.Cfg.ProofCfg, mod, nil, nil, nil)
	goals, err := FlattenStructsTactic(checker, []*LabeledFormula{goal}, nil)
	if err != nil {
		t.Fatal(err)
	}
	model := GoalConc(goals[0]).(*TemporalModels).Model.(*NormalProgram)
	updated := computeUpdate(mod, model.Bindings[0].Action.Stmt)
	leftField, rightField := NewConst("cell.left", Boolean), NewConst("cell.right", Boolean)
	pre := NewClauses([]Expr{leftField, &LogicNot{Body: rightField}}, nil, nil)
	clauses := AndClausesTyped(pre, updated.TR)
	solver := NewSolver(mod, mod.Cfg.SolverOpts)
	defer solver.Close()
	if model, err := solver.GetModelClauses(clauses); err != nil || model == nil {
		t.Fatalf("flattened swap has no transition: %v", err)
	}
	bad := &LogicOr{Terms: []Expr{NewConst(ActionNewName(leftField.Name), Boolean), &LogicNot{Body: NewConst(ActionNewName(rightField.Name), Boolean)}}}
	if model, err := solver.GetModelClauses(AndClausesTyped(clauses, NewClauses([]Expr{bad}, nil, nil))); err != nil || model != nil {
		t.Fatalf("flattened swap allows a wrong result: %v", err)
	}
}

func TestMacroExpandExpandsDefinitionAcrossTemporalModel(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	typ := &UninterpretedSort{Name: "item"}
	sig, err := NewFunctionSort(typ, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	base, alias := NewConst("base", sig), NewConst("alias", sig)
	variable := &LogicVariable{Name: "X", VSort: typ}
	definition := cfg.NewLabeledFormula(cfg.NewAtom("alias_definition"), NewDefinition(NewApplyUnchecked(alias, variable), NewApplyUnchecked(base, variable)))
	definition.IsDefinition = true
	actual := NewConst("sample", typ)
	use := NewApplyUnchecked(alias, actual)
	program := &NormalProgram{Invars: []*LabeledFormula{cfg.NewLabeledFormula(cfg.NewAtom("invariant"), use)}}
	goal := CloneGoal(cfg, cfg.NewLabeledFormula(cfg.NewAtom("goal"), True), []Node{definition}, cfg.NewTemporalModels(program, use))
	checker := NewProofChecker(mod.Cfg.ProofCfg, mod, nil, nil, nil)
	goals, err := MacroExpandTactic(checker, []*LabeledFormula{goal}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(GoalPrems(goals[0])) != 0 {
		t.Fatal("expanded definition is still a premise")
	}
	model := GoalConc(goals[0]).(*TemporalModels)
	expected := NewApplyUnchecked(base, actual)
	if !model.Fmla.(Expr).Equal(expected) || !model.Model.(*NormalProgram).Invars[0].Formula.(Expr).Equal(expected) {
		t.Fatalf("macro remained in the transformed model: %v", model)
	}
}

func TestFlattenStructsPreservesIndexedFieldUpdates(t *testing.T) {
	mod := New()
	outer := &UninterpretedSort{Name: "outer"}
	inner := &UninterpretedSort{Name: "inner"}
	index := &UninterpretedSort{Name: "index"}
	for _, sort := range []*UninterpretedSort{outer, inner, index} {
		mod.Sig.Sorts.Set(sort.Name, sort)
	}
	fieldSort, err := NewFunctionSort(outer, inner)
	if err != nil {
		t.Fatal(err)
	}
	indexedSort, err := NewFunctionSort(inner, index, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	entries := NewConst("entries", fieldSort)
	active := NewConst("active", indexedSort)
	mod.SortDestructors.Set(outer.Name, []*Const{entries})
	mod.SortDestructors.Set(inner.Name, []*Const{active})
	mod.DestructorSorts[entries.Name] = outer
	mod.DestructorSorts[active.Name] = inner
	state := NewConst("state", outer)
	i, j := NewConst("i", index), NewConst("j", index)
	field := NewApplyUnchecked(active, NewApplyUnchecked(entries, state), i)
	program := &NormalProgram{Init: NewAssignAction(field, True)}
	ast := mod.Cfg.AstCfg
	goal := ast.NewLabeledFormula(ast.NewAtom("update"), ast.NewTemporalModels(program, field))
	proof := NewProofChecker(mod.Cfg.ProofCfg, mod, nil, nil, nil)
	goals, err := FlattenStructsTactic(proof, []*LabeledFormula{goal}, nil)
	if err != nil {
		t.Fatal(err)
	}
	temporal := GoalConc(goals[0]).(*TemporalModels)
	flatSort, err := NewFunctionSort(index, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	flat := NewConst("state.entries.active", flatSort)
	if !temporal.Fmla.(Expr).Equal(NewApplyUnchecked(flat, i)) {
		t.Fatalf("flattened field = %s", temporal.Fmla)
	}
	updated := computeUpdate(mod, temporal.Model.(*NormalProgram).Init)
	solver := NewSolver(mod, mod.Cfg.SolverOpts)
	defer solver.Close()
	next := NewConst(ActionNewName(flat.Name), flatSort)
	prior := &LogicAnd{Terms: []Expr{&LogicNot{Body: &Eq{T1: i, T2: j}}, &LogicNot{Body: NewApplyUnchecked(flat, j)}}}
	contradiction := &LogicOr{Terms: []Expr{&LogicNot{Body: NewApplyUnchecked(next, i)}, NewApplyUnchecked(next, j)}}
	clauses := AndClausesTyped(FormulaToClauses(prior, nil), updated.TR)
	if model, err := solver.GetModelClauses(clauses); err != nil || model == nil {
		t.Fatalf("no transition for indexed field assignment: %v", err)
	}
	if model, err := solver.GetModelClauses(AndClausesTyped(clauses, FormulaToClauses(contradiction, nil))); err != nil || model != nil {
		t.Fatalf("indexed update changed the wrong index: %v", err)
	}
}
