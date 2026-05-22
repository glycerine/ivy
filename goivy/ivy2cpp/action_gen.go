package ivy2cpp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// action_gen.go ports Python's emit_action_gen (ivy_to_cpp.py:1197-1348),
// converting an action into a Z3-backed generator class that asserts the
// action's reverse_image precondition, sets the pre-state, randomizes
// inputs, solves, and reads back the inputs from the model.
//
// The work is split into two phases:
//
//   buildActionGenPlan(name, act)  -- pure analysis (no emission).
//   emitActionGen(w, plan)         -- emits the constructor / generate /
//                                     execute bodies for one plan.
//
// Class headers are emitted by emitActionGenClassHeader(w, plan).
//
// On any analysis failure (panic during GetUpdate, missing Update, etc.)
// the plan reports a fallback and the emitter falls back to the weaker
// pre-M4 generator that randomizes inputs directly without consulting
// the solver. This keeps the build resilient while we extend coverage.

// actionGenPlan captures everything we need to emit a single action's
// solver-backed generator class.
type actionGenPlan struct {
	name           string
	className      string
	origAct        goivy.Action
	act            goivy.Action // possibly wrapped by before_export / ext_preconds
	inputs         []*goivy.Const
	fsyms          map[goivy.NodeKey]goivy.Expr
	paramDefs      []goivy.Expr
	oldPreClauses  *goivy.Clauses // pre-AND of relevant_definitions / variant_axioms
	origPreClauses *goivy.Clauses
	preFmla        goivy.Expr
	used           *goivy.InsMap[goivy.NodeKey, goivy.Expr]
	fallback       bool
	fallbackReason string
}

// buildActionGenPlan runs the Python flow at ivy_to_cpp.py:1197-1239 to
// produce a normalized precondition and the matching set of solver
// inputs. Returns a plan whose .fallback flag indicates whether the
// caller should fall back to the weak generator.
func (g *Generator) buildActionGenPlan(name string, act goivy.Action) *actionGenPlan {
	plan := &actionGenPlan{
		name:      name,
		className: g.actionGeneratorClassName(name),
		origAct:   act,
		act:       act,
	}

	// Python: action = im.module.before_export.get(name, action)
	if g.Mod.BeforeExport != nil {
		if be, ok := g.Mod.BeforeExport.Get2(name); ok && be != nil {
			plan.act = be
		}
	}

	// Python: if name in im.module.ext_preconds:
	//             action = ia.Sequence(ia.AssumeAction(im.module.ext_preconds[name]),action)
	if g.Mod.ExtPreconds != nil {
		if pre, ok := g.Mod.ExtPreconds[name]; ok && pre != nil {
			plan.act = goivy.NewSequence(goivy.NewAssumeAction(pre), exprOfAction(plan.act))
		}
	}

	// Compute the action's update. Wrap in recover() because GetUpdate
	// can panic on action subtypes whose Update method isn't fully
	// implemented yet (TODO 014/026 territory).
	var upd *goivy.Update
	func() {
		defer func() {
			if r := recover(); r != nil {
				plan.fallback = true
				plan.fallbackReason = fmt.Sprintf("GetUpdate panicked: %v", r)
			}
		}()
		upd = goivy.GetUpdateForArt(plan.act, g.Mod, nil)
	}()
	if plan.fallback {
		return plan
	}
	if upd == nil {
		plan.fallback = true
		plan.fallbackReason = "GetUpdate returned nil"
		return plan
	}

	// pre = tr.reverse_image(true_clauses, true_clauses, upd)
	truePre := goivy.TrueClauses(nil)
	preClauses := goivy.ReverseImage(truePre, truePre, upd)
	plan.origPreClauses = preClauses
	preClauses = goivy.TrimClauses(preClauses)
	preClauses = expandFieldReferences(preClauses, g.Mod.DestructorSorts)

	// Collect inputs from used local symbols + formal params (prefixed __).
	var inputs []*goivy.Const
	inputSet := map[goivy.NodeKey]bool{}
	for _, sym := range goivy.UsedSymbolsClauses(preClauses).All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if !isLocalSym(c, g.Mod.Sig) {
			continue
		}
		if goivy.IsNumeral(c) {
			continue
		}
		k := goivy.Key(c)
		if inputSet[k] {
			continue
		}
		inputSet[k] = true
		inputs = append(inputs, c)
	}
	for _, p := range plan.act.GetFormalParams() {
		pp := goivy.NewConst("__"+p.Name, p.CSort)
		k := goivy.Key(pp)
		if inputSet[k] {
			continue
		}
		inputSet[k] = true
		inputs = append(inputs, pp)
	}

	// Field extraction + defined parameters.
	preClauses, inputs, plan.fsyms = extractInputFields(preClauses, inputs, g.Mod)
	plan.oldPreClauses = preClauses
	preClauses, plan.paramDefs = extractDefinedParameters(preClauses, inputs)

	// AND with relevant definitions + variant axioms.
	usedNames := make(map[string]bool)
	for _, sym := range goivy.UsedSymbolsClauses(preClauses).All() {
		if c, ok := sym.(*goivy.Const); ok && c.Name != "" {
			usedNames[c.Name] = true
		}
	}
	rdefs := goivy.RelevantDefinitions(g.Mod, usedNames)
	var rdefFmlas []goivy.Expr
	for _, lf := range rdefs {
		if def, ok := lf.Formula.(*goivy.LogicDefinition); ok {
			fixed := fixDefinition(def)
			rdefFmlas = append(rdefFmlas, goivy.DefinitionToConstraint(fixed))
		}
	}
	if len(rdefFmlas) > 0 {
		preClauses = goivy.AndClausesTyped(preClauses, goivy.NewClauses(rdefFmlas, nil, nil))
	}
	varAxioms := g.Mod.VariantAxioms()
	if len(varAxioms) > 0 {
		preClauses = goivy.AndClausesTyped(preClauses, goivy.NewClauses(varAxioms, nil, nil))
	}
	plan.preFmla = preClauses.ToFormula()
	plan.used = goivy.UsedSymbolsAst(plan.preFmla)
	plan.inputs = inputs

	// Validate: no numerals of uninterpreted sort (Python raises here at
	// ivy_to_cpp.py:1243-1244). The Go side propagates as a Generator
	// error so callers can decide whether to keep the gen target.
	for _, sym := range plan.used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if !goivy.IsNumeral(c) {
			continue
		}
		if _, ok := c.CSort.(*goivy.UninterpretedSort); ok {
			err := fmt.Errorf("ivy2cpp: cannot compile numeral %s of uninterpreted sort %s", c.Name, c.CSort)
			g.errs = append(g.errs, err)
			plan.fallback = true
			plan.fallbackReason = err.Error()
			return plan
		}
	}
	return plan
}

// exprOfAction is a small adapter: NewSequence takes Expr arguments, but
// Action and Expr are distinct interfaces. Most action types embed an
// Expr-implementing Base, so a runtime type assertion is enough.
func exprOfAction(a goivy.Action) goivy.Expr {
	if e, ok := a.(goivy.Expr); ok {
		return e
	}
	return nil
}

// emitActionGenClassHeader emits the class declaration (members + method
// decls) for one action_gen. Mirrors Python ivy_to_cpp.py:1246-1259.
func (g *Generator) emitActionGenClassHeader(w *cppWriter, plan *actionGenPlan) {
	className := plan.className
	w.open(fmt.Sprintf("class %s : public gen {", className))
	w.line("public:")
	w.indent++
	w.linef("%s(%s &obj);", className, g.ClassName)
	w.linef("bool generate(%s &obj);", g.ClassName)
	w.linef("void execute(%s &obj);", g.ClassName)
	if plan.fallback {
		// Weak generator: declare the formal params as members like the
		// pre-M4 path did, so the existing execute body still compiles.
		for _, p := range plan.origAct.GetFormalParams() {
			w.linef("%s %s;", g.cppQualifiedType(p.CSort, g.ClassName), varName(p.Name))
		}
	} else {
		// Strong generator: declare each input root as a member,
		// filtering __ts, defidx, and '*>' per Python.
		decld := map[goivy.NodeKey]bool{}
		for _, sym := range plan.inputs {
			var rootExpr goivy.Expr = sym
			if mapped, ok := plan.fsyms[goivy.Key(sym)]; ok {
				rootExpr = exprRoot(mapped)
			}
			rootConst, isConst := exprAsConst(rootExpr)
			if !isConst {
				continue
			}
			k := goivy.Key(rootConst)
			if decld[k] {
				continue
			}
			decld[k] = true
			if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
				continue
			}
			if _, defidx := plan.oldPreClauses.DefIdx[goivy.Key(sym)]; defidx {
				continue
			}
			w.linef("%s %s;", g.cppQualifiedType(rootConst.CSort, g.ClassName), varName(rootConst.Name))
		}
	}
	w.indent--
	w.close(";")
	w.blank()
}

// emitActionGen emits the constructor, generate(), and execute() bodies
// for one plan. Mirrors Python ivy_to_cpp.py:1260-1347.
func (g *Generator) emitActionGen(w *cppWriter, plan *actionGenPlan) {
	if plan.fallback {
		g.emitWeakActionGenerator(w, plan)
		return
	}
	className := plan.className

	// Constructor: ivy2cpp_setup + emit_decl for syms/used-'*>' + add(SMT-LIB).
	w.open(fmt.Sprintf("%s::%s(%s &obj) {", className, className, g.ClassName))
	w.line("(void)obj;")
	w.line("ivy2cpp_setup(*this);")
	emitDeclSet := make(map[goivy.NodeKey]bool)
	for _, sym := range plan.inputs {
		k := goivy.Key(sym)
		if emitDeclSet[k] {
			continue
		}
		emitDeclSet[k] = true
		if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
			continue
		}
		if _, defidx := plan.oldPreClauses.DefIdx[k]; defidx {
			continue
		}
		g.emitDeclSolver(w, stateSymbol{Name: sym.Name, Sort: sym.CSort})
	}
	for _, sym := range plan.used.All() {
		c, ok := sym.(*goivy.Const)
		if !ok {
			continue
		}
		if c.Name != "*>" {
			continue
		}
		k := goivy.Key(c)
		if emitDeclSet[k] {
			continue
		}
		emitDeclSet[k] = true
		g.emitDeclSolver(w, stateSymbol{Name: c.Name, Sort: c.CSort})
	}
	// Emit the precondition assertion in SMT-LIB textual form.
	if smt, ok := g.formulaToSmtlib(plan.preFmla); ok {
		w.linef("add(std::string(\"(assert \") + %s + std::string(\")\"));", strconv.Quote(smt))
	} else {
		w.line("// ivy2cpp: failed to translate precondition to SMT-LIB; falling back to no constraint")
	}
	w.close("")

	// generate() body.
	w.open(fmt.Sprintf("bool %s::generate(%s &obj) {", className, g.ClassName))
	w.line("push();")
	w.line("cpptype_prepare(*this);")
	w.linef("ivy2cpp_progress(*this, %s);", strconv.Quote(className))
	// emit_set for state symbols referenced by the precondition.
	preUsed := goivy.UsedSymbolsAst(plan.preFmla)
	defNames := preDefinedNames(plan.oldPreClauses)
	for _, sym := range g.stateSymbols() {
		if !preUsedContains(preUsed, sym.Name) {
			continue
		}
		if defNames[sym.Name] {
			continue
		}
		g.emitSetSolver(w, sym, "obj")
	}
	w.line("alits.clear();")
	// emit_randomize for each input.
	defedParams := plan.defedParamSet()
	for _, sym := range plan.inputs {
		if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
			continue
		}
		if _, defidx := plan.oldPreClauses.DefIdx[goivy.Key(sym)]; defidx {
			continue
		}
		st := stateSymbol{Name: sym.Name, Sort: sym.CSort}
		if err := g.emitRandomizeSolver(w, st); err != nil {
			w.linef("// ivy2cpp: emitRandomize skipped for %q (%v)", sym.Name, err)
		}
	}
	w.line("bool __res = solve();")
	w.open("if (__res) {")
	for _, sym := range plan.inputs {
		if strings.HasPrefix(sym.Name, "__ts") || sym.Name == "*>" {
			continue
		}
		if _, defidx := plan.oldPreClauses.DefIdx[goivy.Key(sym)]; defidx {
			continue
		}
		if defedParams[goivy.Key(sym)] {
			continue
		}
		st := stateSymbol{Name: sym.Name, Sort: sym.CSort}
		if err := g.emitEvalSolver(w, st, ""); err != nil {
			w.linef("// ivy2cpp: emitEvalSolver skipped for %q (%v)", sym.Name, err)
		}
	}
	// emit_defined_inputs (Python ivy_to_cpp.py:1316). For each
	// parameter-def Eq(input, expr) stripped from the precondition by
	// extract_defined_parameters, recompute the input value from C++
	// state in the post-solve branch.
	if len(plan.paramDefs) > 0 {
		ssyms := make(map[string]bool)
		for _, sym := range g.stateSymbols() {
			ssyms[sym.Name] = true
		}
		g.emitDefinedInputs(w, plan.paramDefs, plan.fsyms, ssyms)
	}
	w.close("")
	w.line("cpptype_cleanup(*this);")
	w.line("pop();")
	w.line("obj.___ivy_gen = this;")
	w.line("return __res;")
	w.close("")

	// execute() body: trace + call action with formal params.
	g.emitActionGenExecute(w, plan)
}

// emitWeakActionGenerator emits the pre-M4 fallback shape: randomize
// state, check(), randomize inputs at C++ level, return true. Used when
// the analysis path cannot produce a precondition for this action.
func (g *Generator) emitWeakActionGenerator(w *cppWriter, plan *actionGenPlan) {
	className := plan.className
	act := plan.origAct
	w.open(fmt.Sprintf("%s::%s(%s &obj) {", className, className, g.ClassName))
	w.line("(void)obj;")
	w.line("ivy2cpp_setup(*this);")
	if plan.fallbackReason != "" {
		w.linef("// ivy2cpp: action_gen fallback (%s)", plan.fallbackReason)
	}
	w.close("")
	w.open(fmt.Sprintf("bool %s::generate(%s &obj) {", className, g.ClassName))
	w.line("obj.___ivy_gen = this;")
	w.linef("ivy2cpp_progress(*this, %s);", strconv.Quote(className))
	w.line("ivy2cpp_randomize(*this, obj);")
	w.open("if (!check()) {")
	w.line("return false;")
	w.close("")
	for _, p := range act.GetFormalParams() {
		nm := varName(p.Name)
		if value, ok := g.z3RandomValueExprFrom(p.CSort, "*this"); ok {
			w.linef("this->%s = %s;", nm, value)
		} else {
			w.linef("this->%s = %s;", nm, g.cppZeroValueInScope(p.CSort))
		}
	}
	w.line("return true;")
	w.close("")
	g.emitActionGenExecute(w, plan)
}

// emitActionGenExecute writes the execute() body for plan.act, mirroring
// Python's lines 1328-1346: trace the call and invoke the method on obj.
func (g *Generator) emitActionGenExecute(w *cppWriter, plan *actionGenPlan) {
	className := plan.className
	act := plan.origAct
	w.open(fmt.Sprintf("void %s::execute(%s &obj) {", className, g.ClassName))
	fn, err := funName(plan.name)
	if err != nil {
		fn = varName(plan.name)
	}
	args := make([]string, 0, len(act.GetFormalParams())+len(act.GetFormalReturns()))
	for _, p := range act.GetFormalParams() {
		args = append(args, "this->"+varName(p.Name))
	}
	returns := act.GetFormalReturns()
	if len(returns) == 1 {
		w.linef("(void)obj.%s(%s);", fn, strings.Join(args, ", "))
	} else {
		for _, r := range returns {
			nm := varName(r.Name)
			w.linef("%s %s = %s;", g.cppQualifiedType(r.CSort, g.ClassName), nm, g.cppZeroValueInScope(r.CSort))
			args = append(args, nm)
		}
		w.linef("obj.%s(%s);", fn, strings.Join(args, ", "))
	}
	w.close("")
	w.blank()
}

// formulaToSmtlib translates fmla to a Z3 expression and returns the
// SMT-LIB textual form after Python's sanitization. Returns ok=false if
// the translator errors out.
func (g *Generator) formulaToSmtlib(fmla goivy.Expr) (string, bool) {
	if fmla == nil {
		return "true", true
	}
	solver := goivy.NewSolver(g.Mod, nil)
	z3expr, err := solver.FormulaToZ3(fmla)
	if err != nil {
		return "", false
	}
	return cleanSmtlib(z3expr.String()), true
}

// preDefinedNames returns the set of defining-symbol names for each
// definition in `clauses`. Used to skip state symbols that the
// precondition itself defines (Python's `sym not in old_pre_clauses.defidx`).
func preDefinedNames(clauses *goivy.Clauses) map[string]bool {
	out := make(map[string]bool)
	if clauses == nil {
		return out
	}
	for _, d := range clauses.Defs {
		def := d.Defines()
		if c, ok := def.(*goivy.Const); ok {
			out[c.Name] = true
		}
	}
	return out
}

// preUsedContains reports whether the precondition's used-symbols set
// references a symbol with the given name.
func preUsedContains(used *goivy.InsMap[goivy.NodeKey, goivy.Expr], name string) bool {
	if used == nil {
		return false
	}
	for _, sym := range used.All() {
		if c, ok := sym.(*goivy.Const); ok && c.Name == name {
			return true
		}
	}
	return false
}

// defedParamSet returns a set of NodeKey for the LHS of each paramDef.
func (p *actionGenPlan) defedParamSet() map[goivy.NodeKey]bool {
	out := make(map[goivy.NodeKey]bool, len(p.paramDefs))
	for _, pd := range p.paramDefs {
		eq, ok := pd.(*goivy.Eq)
		if !ok {
			continue
		}
		if c, ok := eq.T1.(*goivy.Const); ok {
			out[goivy.Key(c)] = true
		}
	}
	return out
}

// exprRoot peels destructor applications and returns the leaf expression
// (Python's `get_root` inside emit_action_gen).
func exprRoot(f goivy.Expr) goivy.Expr {
	for {
		ap, ok := f.(*goivy.Apply)
		if !ok || len(ap.Terms) != 1 {
			return f
		}
		f = ap.Terms[0]
	}
}

// exprAsConst extracts a *Const root from an expression (for member
// declaration). Returns false when the expression isn't reducible to a
// bare constant.
func exprAsConst(f goivy.Expr) (*goivy.Const, bool) {
	switch n := f.(type) {
	case *goivy.Const:
		return n, true
	case *goivy.Apply:
		if c, ok := n.Func.(*goivy.Const); ok && len(n.Terms) == 0 {
			return c, true
		}
	}
	return nil, false
}
