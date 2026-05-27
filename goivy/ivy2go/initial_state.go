package ivy2go

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
	"github.com/glycerine/ivy/goivy/smt"
)

// initialStateConstraints mirrors ivy2cpp/initial_state.go:23. Collects
// InitCond formulas + labeled axioms + relevant definitions into a
// single Clauses set the solver can chew on to derive a consistent
// initial state. Used by `emitOneInitialState` when the module has
// any axiom- or formula-driven initialization (post the `after init`
// imperative actions, which `emitAfterInitActions` handles
// separately).
type initialStateConstraints struct {
	Formulas []goivy.Expr
	Used     map[string]bool
}

type initialDomainValue struct {
	Expr  goivy.Expr
	Go    string
	Z3Int string
}

func (g *Generator) initialStateConstraints() (*initialStateConstraints, error) {
	if g == nil || g.Mod == nil {
		return &initialStateConstraints{Used: map[string]bool{}}, nil
	}
	if err := g.checkInitialStateParameters("initial condition", g.Mod.LabeledInits); err != nil {
		return nil, err
	}
	if err := g.checkInitialStateParameters("axiom", g.Mod.LabeledAxioms); err != nil {
		return nil, err
	}
	var formulas []goivy.Expr
	if g.Mod.InitCond != nil && !g.Mod.InitCond.IsTrue() {
		formulas = append(formulas, g.Mod.InitCond.Fmlas...)
	}
	for _, lf := range g.Mod.LabeledAxioms {
		if f, ok := lf.Formula.(goivy.Expr); ok && f != nil {
			formulas = append(formulas, f)
		}
	}
	used := usedSymbolNames(formulas)
	for _, lf := range goivy.RelevantDefinitions(g.Mod, used) {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		formulas = append(formulas, goivy.DefinitionToConstraint(def))
	}
	return &initialStateConstraints{
		Formulas: formulas,
		Used:     usedSymbolNames(formulas),
	}, nil
}

// initial_state.go mirrors ivy2cpp/initial_state.go. M5 implements the
// nondeterministic (impl/repl/class) path: every unconstrained scalar
// gets ivyChoose; named maps stay empty until first write.
//
// The solver-driven path for target=test/gen is deferred to M9 and
// will plug into the same emitOneInitialState hook.

// emitOneInitialState mirrors ivy2cpp/initial_state.go:101. When the
// module has axiom- or InitCond-formula constraints, run goivy.Solver
// once to derive a consistent initial value for each constrained state
// symbol, then emit per-symbol assignments to those derived values.
// Unconstrained symbols fall back to `emitDefaultInitialState` (nondet
// `ivyChoose` per sort).
//
// On failure (solver error / inconsistent constraints) the function
// records an error on g.errs and leaves the entire state at nondet
// defaults; the user's program will compile and run, just without the
// axiom-guaranteed invariants.
func (g *Generator) emitOneInitialState(w *goWriter) {
	if g == nil || g.Mod == nil {
		return
	}
	constraints, err := g.initialStateConstraints()
	if err != nil {
		g.errs = append(g.errs, fmt.Errorf("ivy2go: initial state constraints: %w", err))
		for _, sym := range g.stateSymbols() {
			g.emitDefaultInitialState(w, sym)
		}
		return
	}
	var slv *goivy.Solver
	var model *smt.Model
	if len(constraints.Formulas) > 0 {
		slv = goivy.NewSolver(g.Mod, nil)
		mr, err := slv.GetModelClauses(goivy.NewClauses(constraints.Formulas, nil, nil))
		if err != nil {
			g.errs = append(g.errs, fmt.Errorf("ivy2go: initial state solver failed: %w", err))
		} else if mr == nil || mr.Model == nil {
			g.errs = append(g.errs, fmt.Errorf("ivy2go: axioms and/or initial condition are inconsistent"))
		} else {
			model = mr.Model
		}
	}
	for _, sym := range g.stateSymbols() {
		if g.isParamName(sym.Name) {
			continue
		}
		if !constraints.Used[sym.Name] || model == nil {
			g.emitDefaultInitialState(w, sym)
			continue
		}
		if err := g.emitSolvedInitialState(w, slv, model, sym); err != nil {
			g.errs = append(g.errs, fmt.Errorf("ivy2go: %w", err))
			g.emitDefaultInitialState(w, sym)
		}
	}
	if slv != nil {
		_ = slv.Close()
	}
}

// emitDefaultInitialState mirrors ivy2cpp/initial_state.go:134.
// Emits a nondet initialization for a state symbol when no solver
// model constrains its value.
func (g *Generator) emitDefaultInitialState(w *goWriter, sym stateSymbol) {
	field := "s." + goExportedName(sym.Name)
	g.mkNondetSymbolStorage(w, field, sym.Sort, "init", 0)
}

// emitSolvedInitialState mirrors ivy2cpp/initial_state.go:142. For a
// scalar symbol, emits `s.<Field> = <value>` using the solver-derived
// value. For a function-sorted symbol, enumerates the domain tuples
// and emits per-cell assignments.
func (g *Generator) emitSolvedInitialState(w *goWriter, slv *goivy.Solver, model *smt.Model, sym stateSymbol) error {
	if slv == nil || model == nil {
		g.emitDefaultInitialState(w, sym)
		return nil
	}
	field := "s." + goExportedName(sym.Name)
	fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
	if !ok || len(fs.Domain()) == 0 {
		term := initialStateSymbolTerm(sym, nil)
		value, err := g.initialModelValue(slv, model, term, initialStateRange(sym.Sort))
		if err != nil {
			return fmt.Errorf("initial value for %s: %w", sym.Name, err)
		}
		w.linef("%s = %s", field, value)
		return nil
	}
	tuples, ok := g.initialDomainTuples(fs.Domain())
	if !ok {
		return fmt.Errorf("ivy2go: cannot enumerate initial-state domain of %s", sym.Name)
	}
	st := goFunctionStorageFor(g, fs.Domain(), fs.Range())
	for _, tuple := range tuples {
		args := make([]goivy.Expr, len(tuple))
		goArgs := make([]string, len(tuple))
		for i, v := range tuple {
			args[i] = v.Expr
			goArgs[i] = v.Go
		}
		term := initialStateSymbolTerm(sym, args)
		value, err := g.initialModelValue(slv, model, term, fs.Range())
		if err != nil {
			return fmt.Errorf("initial value for %s: %w", sym.Name, err)
		}
		switch st.Kind {
		case goStorageArray:
			lhs := field + g.compactArrayIndexSuffix(fs.Domain(), goArgs)
			w.linef("%s = %s", lhs, value)
		case goStorageHashThunk:
			// Lazily allocate the map then assign the cell.
			w.linef("if %s == nil { %s = %s{} }", field, field, st.Type)
			keyType := goCTupleNameWith(g, fs.Domain())
			if len(goArgs) == 1 {
				w.linef("%s[%s] = %s", field, goArgs[0], value)
			} else {
				w.linef("%s[%s{%s}] = %s", field, keyType, strings.Join(goArgs, ", "), value)
			}
		default:
			// Scalar fallback (shouldn't reach for function-sorted).
			w.linef("%s = %s", field, value)
		}
	}
	return nil
}

// initialModelValue mirrors ivy2cpp/initial_state.go:205. Translates
// a state-symbol term to its Z3 form, evaluates it under the model,
// and converts the result to a Go expression of sort `s`.
func (g *Generator) initialModelValue(slv *goivy.Solver, model *smt.Model, term goivy.Expr, s goivy.Sort) (string, error) {
	zterm, err := slv.FormulaToZ3(term)
	if err != nil {
		return "", err
	}
	value, ok := model.Eval(zterm, true)
	if !ok {
		return "", fmt.Errorf("model did not evaluate %s", term.String())
	}
	return g.modelValueToGo(value, s)
}

// modelValueToGo mirrors ivy2cpp/initial_state.go:217 modelValueToCpp.
// Converts a Z3 model value into a Go literal of the appropriate sort.
func (g *Generator) modelValueToGo(value smt.Z3Expr, s goivy.Sort) (string, error) {
	text := stripZ3Bars(strings.TrimSpace(value.String()))
	switch st := s.(type) {
	case *goivy.BooleanSort:
		switch text {
		case "true", "1":
			return "true", nil
		case "false", "0":
			return "false", nil
		default:
			return "", fmt.Errorf("expected bool model value, got %q", text)
		}
	case *goivy.LogicEnumeratedSort:
		if isNumericEnum(st) {
			for i, name := range st.Extension {
				if text == name || text == fmt.Sprintf("%s_%d", z3SortName(st), i) {
					return name, nil
				}
			}
			if idx, ok := parseTrailingModelIndex(text); ok && idx >= 0 && idx < len(st.Extension) {
				return st.Extension[idx], nil
			}
			return "", fmt.Errorf("expected %s model value, got %q", sortName(s), text)
		}
		for i, name := range st.Extension {
			if text == name || text == fmt.Sprintf("%s_%d", z3SortName(st), i) {
				return goExportedName(name), nil
			}
		}
		if idx, ok := parseTrailingModelIndex(text); ok && idx >= 0 && idx < len(st.Extension) {
			return goExportedName(st.Extension[idx]), nil
		}
		return "", fmt.Errorf("expected %s model value, got %q", sortName(s), text)
	default:
		if it, ok := g.goInterpType(s); ok && it.Kind == goInterpBV {
			return g.bvModelValueToGo(text, it)
		}
		typeName := g.goType(s)
		if n, err := strconv.ParseInt(text, 10, 64); err == nil {
			if typeName == "int" || typeName == "" {
				return strconv.FormatInt(n, 10), nil
			}
			return fmt.Sprintf("%s(%d)", typeName, n), nil
		}
		if idx, ok := parseTrailingModelIndex(text); ok {
			if typeName == "int" || typeName == "" {
				return strconv.Itoa(idx), nil
			}
			return fmt.Sprintf("%s(%d)", typeName, idx), nil
		}
		return "", fmt.Errorf("unsupported model value %q for sort %s", text, sortName(s))
	}
}

func (g *Generator) bvModelValueToGo(text string, it goInterpType) (string, error) {
	decimal, ok := parseBVModelValueText(text)
	if !ok {
		return "", fmt.Errorf("unsupported bitvector model value %q", text)
	}
	switch {
	case it.Bits <= 32:
		return fmt.Sprintf("(uint32(%s) & %s)", decimal, bvMask(it.Bits)), nil
	case it.Bits <= 64:
		return fmt.Sprintf("(uint64(%s) & %s)", decimal, bvMask(it.Bits)), nil
	default:
		return g.wideBVLiteralExpr(decimal, it.Bits), nil
	}
}

func parseBVModelValueText(text string) (string, bool) {
	text = stripZ3Bars(strings.TrimSpace(text))
	if text == "" {
		return "", false
	}
	if strings.HasPrefix(text, "#x") || strings.HasPrefix(text, "#X") {
		return parseBigIntString(text[2:], 16)
	}
	if strings.HasPrefix(text, "#b") || strings.HasPrefix(text, "#B") {
		return parseBigIntString(text[2:], 2)
	}
	if strings.HasPrefix(text, "(_ bv") && strings.HasSuffix(text, ")") {
		body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "(_ bv"), ")"))
		fields := strings.Fields(body)
		if len(fields) != 2 {
			return "", false
		}
		return parseBigIntString(fields[0], 10)
	}
	return parseBigIntString(text, 10)
}

func parseBigIntString(text string, base int) (string, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}
	n, ok := new(big.Int).SetString(text, base)
	if !ok || n.Sign() < 0 {
		return "", false
	}
	return n.String(), true
}

// initialDomainTuples mirrors ivy2cpp/initial_state.go:269.
func (g *Generator) initialDomainTuples(domain []goivy.Sort) ([][]initialDomainValue, bool) {
	tuples := [][]initialDomainValue{{}}
	for _, s := range domain {
		vals, ok := g.initialDomainValues(s)
		if !ok {
			return nil, false
		}
		var next [][]initialDomainValue
		for _, tuple := range tuples {
			for _, v := range vals {
				cp := append([]initialDomainValue{}, tuple...)
				cp = append(cp, v)
				next = append(next, cp)
			}
		}
		tuples = next
	}
	return tuples, true
}

// initialDomainValues mirrors ivy2cpp/initial_state.go:289. Enumerates
// the values of a finite-cardinality sort (bool / enum / range).
func (g *Generator) initialDomainValues(s goivy.Sort) ([]initialDomainValue, bool) {
	switch st := s.(type) {
	case *goivy.BooleanSort:
		return []initialDomainValue{
			{Expr: goivy.NewConst("false", goivy.Boolean), Go: "false", Z3Int: "0"},
			{Expr: goivy.NewConst("true", goivy.Boolean), Go: "true", Z3Int: "1"},
		}, true
	case *goivy.LogicEnumeratedSort:
		vals := make([]initialDomainValue, 0, len(st.Extension))
		for i, name := range st.Extension {
			goLit := goExportedName(name)
			if isNumericEnum(st) {
				goLit = name
			}
			vals = append(vals, initialDomainValue{
				Expr:  goivy.NewConst(name, st),
				Go:    goLit,
				Z3Int: strconv.Itoa(i),
			})
		}
		return vals, true
	default:
		if rs, ok := g.rangeSortFor(s); ok {
			loText, hiText, ok := numericRangeBounds(rs)
			if !ok {
				return nil, false
			}
			lo, _ := strconv.Atoi(loText)
			hi, _ := strconv.Atoi(hiText)
			typeName := g.goType(s)
			vals := make([]initialDomainValue, 0, hi-lo+1)
			for i := lo; i <= hi; i++ {
				text := strconv.Itoa(i)
				goLit := text
				if typeName != "" && typeName != "int" {
					goLit = fmt.Sprintf("%s(%d)", typeName, i)
				}
				vals = append(vals, initialDomainValue{
					Expr:  goivy.NewConst(text, s),
					Go:    goLit,
					Z3Int: text,
				})
			}
			return vals, true
		}
		return nil, false
	}
}

// emitScalarChoice writes a single nondet initialization for a scalar
// state symbol. Mirrors Python's `mk_nondet_sym` (ivy_to_cpp.py:3019).
func (g *Generator) emitScalarChoice(w *goWriter, sym stateSymbol) {
	field := "s." + goExportedName(sym.Name)
	switch sym.Sort.(type) {
	case *goivy.BooleanSort:
		w.linef("%s = ivyChoose(2) == 1", field)
	default:
		if expr, ok := g.rangeChoiceExpr(sym.Sort, fmt.Sprintf("ivyChoose(%d)", goSortCard(g, sym.Sort))); ok {
			w.linef("%s = %s", field, expr)
			return
		}
		card := goSortCard(g, sym.Sort)
		if card <= 0 {
			// Fall back to zero-initialisation when cardinality is
			// unknown; Go's zero value is sensible for scalars.
			return
		}
		typeName := g.goType(sym.Sort)
		w.linef("%s = %s(ivyChoose(%d))", field, typeName, card)
	}
	_ = fmt.Sprintf
}

// isFunctionSort returns true when s is a function sort. Local helper
// so callers don't need to type-assert inline.
func isFunctionSort(s goivy.Sort) bool {
	_, ok := s.(*goivy.LogicFunctionSort)
	return ok
}

// checkInitialStateParameters mirrors ivy2cpp/initial_state.go:56.
// Returns an error when an initial-state formula references a
// module parameter that was stripped from the signature — such
// formulas would silently fail to constrain the right symbol.
func (g *Generator) checkInitialStateParameters(kind string, lfs []*goivy.LabeledFormula) error {
	if g == nil || g.Mod == nil || len(g.Mod.Params) == 0 {
		return nil
	}
	params := make(map[string]bool, len(g.Mod.Params))
	for _, p := range g.Mod.Params {
		if p != nil && p.Name != "" {
			params[p.Name] = true
		}
	}
	for _, lf := range lfs {
		f, ok := lf.Formula.(goivy.Expr)
		if !ok || f == nil {
			continue
		}
		for name := range usedSymbolNames([]goivy.Expr{f}) {
			if !params[name] {
				continue
			}
			line := ""
			if lf != nil && lf.Lineno() > 0 {
				line = fmt.Sprintf(" at line %d", lf.Lineno())
			}
			return fmt.Errorf("ivy2go: %s%s depends on stripped parameter %q", kind, line, name)
		}
	}
	return nil
}

// usedSymbolNames mirrors ivy2cpp/initial_state.go:85. Returns the
// set of named symbols referenced by any of `formulas`.
func usedSymbolNames(formulas []goivy.Expr) map[string]bool {
	used := map[string]bool{}
	for _, f := range formulas {
		if f == nil {
			continue
		}
		for _, sym := range goivy.UsedSymbolsAst(f).All() {
			name := goivy.ExprName(sym)
			if name != "" {
				used[name] = true
			}
		}
	}
	return used
}

// cloneSortEnv mirrors ivy2cpp/initial_state.go:408. Returns a
// shallow copy of a sort-environment map.
func cloneSortEnv(env map[string]goivy.Sort) map[string]goivy.Sort {
	out := make(map[string]goivy.Sort, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

// isParamName mirrors ivy2cpp/initial_state.go:635. Reports whether
// `name` matches one of the module's stripped parameter names.
func (g *Generator) isParamName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	for _, p := range g.Mod.Params {
		if p != nil && p.Name == name {
			return true
		}
	}
	return false
}

// initialStateRange mirrors ivy2cpp/initial_state.go:178. Returns
// the range sort of a function sort, or the sort itself for scalars.
func initialStateRange(s goivy.Sort) goivy.Sort {
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		return fs.Range()
	}
	return s
}

// initialStateSymbolTerm mirrors ivy2cpp/initial_state.go:185.
// Builds the term that references a state symbol — a bare Const for
// scalars, an Apply with the given args for function-sorted symbols.
func initialStateSymbolTerm(sym stateSymbol, args []goivy.Expr) goivy.Expr {
	c := goivy.NewConst(sym.Name, sym.Sort)
	fs, ok := sym.Sort.(*goivy.LogicFunctionSort)
	if !ok {
		return c
	}
	if len(args) == 0 && len(fs.Domain()) == 0 {
		app, err := goivy.NewApply(c)
		if err == nil {
			return app
		}
		return c
	}
	app, err := goivy.NewApply(c, args...)
	if err != nil {
		return c
	}
	return app
}

// parseTrailingModelIndex mirrors ivy2cpp/initial_state.go:257.
// Parses the trailing `_N` suffix Z3 attaches to model identifiers.
func parseTrailingModelIndex(s string) (int, bool) {
	pos := strings.LastIndexByte(s, '_')
	if pos < 0 || pos+1 >= len(s) {
		return 0, false
	}
	n, err := strconv.Atoi(s[pos+1:])
	if err != nil {
		return 0, false
	}
	return n, true
}
