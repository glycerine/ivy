package module

// Batch 1.8: Skolemization, Definitions, Parsing, and higher-order helpers.
// Corresponds to ivy_logic_utils.py skolemization and parsing functions.

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// LogicParseError is raised when parsing a logic expression fails.
// Corresponds to Python's class LogicParseError (ivy_logic_utils.py).
type LogicParseError struct {
	Token string
	Msg   string
}

func (e *LogicParseError) Error() string {
	if e.Token != "" {
		return fmt.Sprintf("logic parse error at '%s': %s", e.Token, e.Msg)
	}
	return fmt.Sprintf("logic parse error: %s", e.Msg)
}

// VarToConstant converts a variable to a constant with the given name.
// Corresponds to Python's var_to_constant (ivy_logic_utils.py:1486-1488).
func VarToConstant(v *lg.Variable, name string) lg.Expr {
	sym := lg.NewConst(name, v.VSort)
	return il.Constant(sym)
}

// VarToSkolem converts a variable to a Skolem constant with name prefix+v.name.
// Corresponds to Python's var_to_skolem (ivy_logic_utils.py:1483-1484).
func VarToSkolem(prefix string, v *lg.Variable) lg.Expr {
	return VarToConstant(v, prefix+v.Name)
}

// DualFormula negates a formula after replacing free variables with
// Skolem constants. Corresponds to Python's dual_formula
// (ivy_logic_utils.py:1527-1537).
// If instantiator is non-nil, definition instances are conjoined with the
// negated formula (matching Python's `if instantiator != None` check).
func DualFormula(fmla lg.Expr, skolemizer func(*lg.Variable) lg.Expr, instantiator func([]lg.Expr) *Clauses) lg.Expr {
	if skolemizer == nil {
		skolemizer = func(v *lg.Variable) lg.Expr {
			return VarToSkolem("__", v)
		}
	}
	// Collect used variables in order
	vars := UsedVariablesOrdered(&Clauses{Fmlas: []lg.Expr{fmla}})
	if len(vars) > 0 {
		subs := make(map[string]lg.Expr, len(vars))
		for _, v := range vars {
			subs[v.Name] = skolemizer(v)
		}
		fmla = SubstituteAstByName(fmla, subs)
	}
	fmla = Negate(fmla)
	// Python: if instantiator != None: fmla = And(fmla, clauses_to_formula(insts))
	// Matches Python dual_formula (ivy_logic_utils.py:1567-1577) — unconditional
	// (no len(insts.Fmlas) > 0 guard); Python builds And(fmla, And()) when insts is empty.
	if instantiator != nil {
		gts := il.AppsAst(fmla)
		insts := instantiator(gts)
		fmla = &lg.And{Terms: []lg.Expr{fmla, clausesToFormula(insts)}}
	}
	return fmla
}

// SkolemizeFormula skolemizes leading existential quantifiers in a formula.
// Corresponds to Python's skolemize_formula (ivy_logic_utils.py:1539-1552).
// If instantiator is non-nil, definition instances are conjoined with the
// skolemized formula (matching Python's `if instantiator != None` check).
func SkolemizeFormula(fmla lg.Expr, skolemizer func(*lg.Variable) lg.Expr, instantiator func([]lg.Expr) *Clauses) lg.Expr {
	if skolemizer == nil {
		skolemizer = func(v *lg.Variable) lg.Expr {
			return VarToSkolem("__sk__", v)
		}
	}
	var vs []*lg.Variable
	for {
		if ex, ok := fmla.(*lg.Exists); ok {
			vs = append(vs, ex.Variables...)
			fmla = ex.Body
		} else {
			break
		}
	}
	if len(vs) > 0 {
		subs := make(map[string]lg.Expr, len(vs))
		for _, v := range vs {
			subs[v.Name] = skolemizer(v)
		}
		fmla = SubstituteAstByName(fmla, subs)
	}
	// Python: if instantiator != None: fmla = And(fmla, clauses_to_formula(insts))
	// Matches Python skolemize_formula (ivy_logic_utils.py:1579-1592) — unconditional
	// (no len(insts.Fmlas) > 0 guard); Python builds And(fmla, And()) when insts is empty.
	if instantiator != nil {
		gts := il.AppsAst(fmla)
		insts := instantiator(gts)
		fmla = &lg.And{Terms: []lg.Expr{fmla, clausesToFormula(insts)}}
	}
	return fmla
}

// SkolemizeAst performs full polarity-aware Skolemization on an AST.
// Corresponds to Python's skolemize_ast (ivy_logic_utils.py:1554-1582).
func SkolemizeAst(pos bool, vs []*lg.Variable, usedNames map[string]bool,
	skolems *[]*lg.Const, fmla lg.Expr, prefix string) lg.Expr {

	if il.IsQuantifier(fmla) {
		isExists := il.IsExists(fmla)
		isForall := il.IsForall(fmla)
		// Skolemize: exists in positive position, forall in negative
		if (isExists && pos) || (isForall && !pos) {
			vars := il.BinderVars(fmla)
			body := il.BinderBody(fmla)

			// Collect outer universal variables that appear in body
			usedVars := lu.UsedVariables(body)
			var mvs []*lg.Variable
			for _, w := range vs {
				if _, used := usedVars[lg.Key(w)]; used {
					mvs = append(mvs, w)
				}
			}

			subs := make(map[string]lg.Expr)
			for _, v := range vars {
				name := "@"
				if prefix != "" {
					name += prefix + "."
				}
				name += v.Name
				if usedNames[name] {
					count := 1
					for {
						newName := fmt.Sprintf("%s[%d]", name, count)
						if !usedNames[newName] {
							name = newName
							break
						}
						count++
					}
				}
				usedNames[name] = true

				// Build Skolem function sort
				domSorts := make([]lg.Sort, 0, len(mvs)+1)
				for _, w := range mvs {
					domSorts = append(domSorts, w.VSort)
				}
				domSorts = append(domSorts, v.VSort)
				skSort := il.FuncConstSort(domSorts...)
				skSym := lg.NewConst(name, skSort)
				*skolems = append(*skolems, skSym)

				// Apply Skolem function to outer variables
				if len(mvs) > 0 {
					args := make([]lg.Expr, len(mvs))
					for j, w := range mvs {
						args[j] = w
					}
					subs[v.Name] = il.Apply(skSym, args)
				} else {
					subs[v.Name] = skSym
				}
			}
			return SubstituteAstByName(body, subs)
		}
		// Universal: recurse into body with extended variable scope
		vars := il.BinderVars(fmla)
		body := il.BinderBody(fmla)
		newVs := make([]*lg.Variable, len(vs)+len(vars))
		copy(newVs, vs)
		copy(newVs[len(vs):], vars)
		newBody := SkolemizeAst(pos, newVs, usedNames, skolems, body, prefix)
		return il.CloneBinder(fmla, vars, newBody)
	}

	if _, ok := fmla.(*lg.Not); ok {
		args := il.NodeArgs(fmla)
		newArg := SkolemizeAst(!pos, vs, usedNames, skolems, args[0], prefix)
		return il.CloneNode(fmla, []lg.Expr{newArg})
	}

	if _, ok := fmla.(*lg.Implies); ok {
		args := il.NodeArgs(fmla)
		newLhs := SkolemizeAst(!pos, vs, usedNames, skolems, args[0], prefix)
		newRhs := SkolemizeAst(pos, vs, usedNames, skolems, args[1], prefix)
		return il.CloneNode(fmla, []lg.Expr{newLhs, newRhs})
	}

	args := il.NodeArgs(fmla)
	if len(args) == 0 {
		return fmla
	}
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = SkolemizeAst(pos, vs, usedNames, skolems, a, prefix)
	}
	return il.CloneNode(fmla, newArgs)
}

// WitnessAst substitutes witness values for existentially quantified
// variables (in negative position) or universally quantified variables
// (in positive position).
// Corresponds to Python's witness_ast (ivy_logic_utils.py:1673-1704).
//
// Takes/returns ast.Node so it can handle *ast.TemporalModels (mirroring
// Python's duck-typed generic-branch recursion at line 1704:
// `return fmla.clone([witness_ast(pos,vs,witnesses,arg) for arg in fmla.args])`).
// For lg.Expr inputs, behaves identically to the previous lg.Expr-only
// signature. This mirrors the ast.Node broadening already applied to
// SkolemizeFmla (proof/skolem.go:100).
func WitnessAst(pos bool, vs []*lg.Variable, witnesses map[lg.NodeKey]lg.Expr, fmla ast.Node) (ast.Node, error) {
	xtracer.Trace("ilu.witnessAst ENTER pos=%v type=%s nwitnesses=%d", pos, iu.TypeName(fmla), len(witnesses))
	// TemporalModels — recurse into the wrapped inner formula and rewrap,
	// mirroring Python's generic fmla.clone([... for arg in fmla.args]) branch.
	if tm, ok := fmla.(*ast.TemporalModels); ok {
		xtracer.Trace("ilu.witnessAst branch type=TemporalModels pos=%v", pos)
		innerExpr, ok := tm.Fmla.(lg.Expr)
		if !ok {
			xtracer.Trace("ilu.witnessAst EXIT type=TemporalModels innerNotExpr")
			return tm, nil
		}
		newInner, err := WitnessAst(pos, vs, witnesses, innerExpr)
		if err != nil {
			xtracer.Trace("ilu.witnessAst EXIT type=TemporalModels err=%v", err)
			return nil, err
		}
		result := tm.Clone([]ast.Node{newInner})
		xtracer.Trace("ilu.witnessAst EXIT type=TemporalModels HASH canon=%v", result.Canon())
		return result, nil
	}

	// Everything else in witness_ast's dispatch operates on lg.Expr.
	expr, ok := fmla.(lg.Expr)
	if !ok {
		xtracer.Trace("ilu.witnessAst EXIT notExpr type=%s", iu.TypeName(fmla))
		return fmla, nil
	}

	if il.IsQuantifier(expr) {
		isExists := il.IsExists(expr)
		isForall := il.IsForall(expr)
		// Apply witnesses to: exists in negative, forall in positive
		if (isExists && !pos) || (isForall && pos) {
			vars := il.BinderVars(expr)
			body := il.BinderBody(expr)
			xtracer.Trace("ilu.witnessAst branch type=Quantifier pos=%v isE=%v isA=%v nvars=%d", pos, isExists, isForall, len(vars))

			var newVars []*lg.Variable
			for idx, v := range vars {
				term, found := witnesses[lg.Key(v)]
				xtracer.Trace("ilu.witnessAst quantifierVar v=%s key=%s found=%v", v.Name, string(lg.Key(v)), found)
				if found {
					// Check for capture by subsequent variables
					termVars := lu.UsedVariables(term)
					for _, w := range vars[idx+1:] {
						if _, used := termVars[lg.Key(w)]; used {
							xtracer.Trace("ilu.witnessAst EXIT err=capture var=%s", w.Name)
							return nil, fmt.Errorf("variable %s captured by substitution", w.Name)
						}
					}
					subs := map[lg.NodeKey]lg.Expr{lg.Key(v): term}
					newBody, err := lu.Substitute(body, subs)
					if err != nil {
						xtracer.Trace("ilu.witnessAst EXIT err=substCapture %v", err)
						return nil, fmt.Errorf("variable capture during witness substitution: %v", err)
					}
					body = newBody
				} else {
					newVars = append(newVars, v)
				}
			}
			bodyNode, err := WitnessAst(pos, vs, witnesses, body)
			if err != nil {
				return nil, err
			}
			body = witnessExpr(bodyNode, body)
			if len(newVars) > 0 {
				result := il.CloneBinder(expr, newVars, body)
				xtracer.Trace("ilu.witnessAst EXIT type=Quantifier nnewVars=%d HASH canon=%v", len(newVars), result.Canon())
				return result, nil
			}
			xtracer.Trace("ilu.witnessAst EXIT type=Quantifier allSubstituted HASH canon=%v", body.Canon())
			return body, nil
		}
	}

	if _, ok := expr.(*lg.Not); ok {
		xtracer.Trace("ilu.witnessAst branch type=Not pos=%v", pos)
		args := il.NodeArgs(expr)
		newArg, err := WitnessAst(!pos, vs, witnesses, args[0])
		if err != nil {
			return nil, err
		}
		result := il.CloneNode(expr, []lg.Expr{witnessExpr(newArg, args[0])})
		xtracer.Trace("ilu.witnessAst EXIT type=Not HASH canon=%v", result.Canon())
		return result, nil
	}

	if _, ok := expr.(*lg.Implies); ok {
		xtracer.Trace("ilu.witnessAst branch type=Implies pos=%v", pos)
		args := il.NodeArgs(expr)
		newLhs, err := WitnessAst(!pos, vs, witnesses, args[0])
		if err != nil {
			return nil, err
		}
		newRhs, err := WitnessAst(pos, vs, witnesses, args[1])
		if err != nil {
			return nil, err
		}
		result := il.CloneNode(expr, []lg.Expr{
			witnessExpr(newLhs, args[0]),
			witnessExpr(newRhs, args[1]),
		})
		xtracer.Trace("ilu.witnessAst EXIT type=Implies HASH canon=%v", result.Canon())
		return result, nil
	}

	args := il.NodeArgs(expr)
	if len(args) == 0 {
		xtracer.Trace("ilu.witnessAst EXIT type=%s leaf", iu.TypeName(expr))
		return expr, nil
	}
	xtracer.Trace("ilu.witnessAst branch type=generic pos=%v exprType=%s nargs=%d", pos, iu.TypeName(expr), len(args))
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		na, err := WitnessAst(pos, vs, witnesses, a)
		if err != nil {
			return nil, err
		}
		newArgs[i] = witnessExpr(na, a)
	}
	result := il.CloneNode(expr, newArgs)
	xtracer.Trace("ilu.witnessAst EXIT type=generic exprType=%s HASH canon=%v", iu.TypeName(expr), result.Canon())
	return result, nil
}

// witnessExpr type-asserts a WitnessAst result back to lg.Expr.
// Recursive calls through *ast.TemporalModels can legitimately return
// non-lg.Expr nodes; callers of witnessExpr are only inside lg.Expr
// subtrees where the result must be lg.Expr, so a failed assertion
// would indicate a contract violation. fallback is used as a safety
// net.
func witnessExpr(n ast.Node, fallback lg.Expr) lg.Expr {
	if e, ok := n.(lg.Expr); ok {
		return e
	}
	return fallback
}

// ReskolemizeClauses re-skolemizes clauses by replacing Skolem constants
// (those with '__' in their name) using the given skolemizer.
// Corresponds to Python's reskolemize_clauses (ivy_logic_utils.py:1618-1623).
func ReskolemizeClauses(clauses *Clauses, skolemizer func(*lg.Variable) lg.Expr) *Clauses {
	// Find all constants with '__' in their name
	consts := ConstantsClauses(clauses)
	subs := make(map[lg.NodeKey]lg.Expr)
	for _, c := range consts {
		if strings.Contains(c.Name, "__") {
			v, _ := lg.NewVariable(c.Name, c.CSort)
			subs[lg.Key(c)] = skolemizer(v)
		}
	}
	if len(subs) == 0 {
		return clauses
	}
	return SubstituteConstantsClauses(clauses, subs)
}

// UnusedConstant returns a constant with a name not in the current signature
// or in the given used constants set.
// Corresponds to Python's unused_constant (ivy_logic_utils.py:1625-1635).
func UnusedConstant(sig *il.Sig, usedConstants []*lg.Const, sort lg.Sort) *lg.Const {
	usedNames := make(map[string]bool)
	for _, c := range usedConstants {
		usedNames[c.Name] = true
	}
	for name := range sig.Symbols.All() {
		usedNames[name] = true
	}
	gen := iu.ConstantNameGenerator()
	for {
		name := gen()
		if !usedNames[name] {
			return lg.NewConst(name, sort)
		}
	}
}

// DefinitionInstances returns instantiation clauses for a formula using
// the given instantiator. Corresponds to Python's definition_instances
// (ivy_logic_utils.py:1500-1504).
func DefinitionInstances(fmla lg.Expr, instantiator func([]lg.Expr) *Clauses) *Clauses {
	if instantiator == nil {
		return NewClauses(nil, nil, nil)
	}
	gts := il.AppsAst(fmla)
	return instantiator(gts)
}

// UnfoldDefinitionsClauses adds definition instances to clauses.
// Corresponds to Python's unfold_definitions_clauses
// (ivy_logic_utils.py:1506-1512).
func UnfoldDefinitionsClauses(clauses *Clauses, instantiator func([]lg.Expr) *Clauses) *Clauses {
	if instantiator == nil {
		return clauses
	}
	gts := AppsClauses(clauses)
	insts := instantiator(gts)
	if len(insts.Fmlas) > 0 {
		clauses = AndClausesTyped(clauses, insts)
	}
	return clauses
}

// ApplyGenToClauses returns a function that applies a generator to a
// Clauses or a bare Node. Corresponds to Python's apply_gen_to_clauses
// (ivy_logic_utils.py:129-131).
func ApplyGenToClauses(gen func(lg.Expr) lg.Expr) func(interface{}) interface{} {
	return func(cls interface{}) interface{} {
		switch v := cls.(type) {
		case *Clauses:
			return v.Apply(gen)
		case lg.Expr:
			return gen(v)
		}
		return cls
	}
}

// ApplyFuncToClauses returns a function that applies func to a Clauses
// or a bare Node. Corresponds to Python's apply_func_to_clauses
// (ivy_logic_utils.py:133-135).
func ApplyFuncToClauses(fn func(lg.Expr) lg.Expr) func(interface{}) interface{} {
	return func(cls interface{}) interface{} {
		switch v := cls.(type) {
		case *Clauses:
			return v.Apply(fn)
		case lg.Expr:
			return fn(v)
		}
		return cls
	}
}
