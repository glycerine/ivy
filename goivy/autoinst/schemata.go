package autoinst

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
)

// ExpandSchemata expands axiom schemata into axioms by matching schema
// premises against the sort constants and functions.
//
// Python: ivy_auto_inst.py:100-200 (uses match_schema_prems generator)
func ExpandSchemata(m *goivy.Module, sortConstants map[string][]*goivy.Const, funs map[string]bool) []*goivy.LabeledFormula {
	var result []*goivy.LabeledFormula

	if m.Schemata == nil {
		return result
	}

	// Build initial match with all known sorts
	match := NewMatch()
	if m.Sig != nil {
		for sortName, _ := range m.Sig.Sorts.All() {
			match.Add(sortName, sortName)
		}
	}

	// For each schema, try matching premises
	for name, lf := range m.Schemata.All() {
		// Skip recursive/inductive schemata
		if len(name) >= 4 && (name[:4] == "rec[" || name[:4] == "lep[" || name[:4] == "ind[") {
			continue
		}

		schema, ok := extractSchemaNode(lf)
		if !ok || schema == nil {
			continue
		}

		children := schema.Children()
		if len(children) < 2 {
			continue
		}

		conc := children[len(children)-1]
		prems := children[:len(children)-1]

		// Collect bound sorts
		boundSorts := make(map[string]bool)
		for _, prem := range prems {
			if us, ok := prem.(*goivy.UninterpretedSort); ok {
				boundSorts[us.Name] = true
			}
		}

		// Match premises and instantiate conclusion
		MatchSchemaPrems(prems, sortConstants, funs, match, boundSorts, func(mp map[string]interface{}) {
			// Build substitution map
			subs := make(map[string]goivy.Expr)
			for k, v := range mp {
				switch val := v.(type) {
				case *goivy.Const:
					subs[k] = val
				case goivy.Expr:
					subs[k] = val
				case string:
					subs[k] = goivy.NewConst(val, nil)
				}
			}
			inst := goivy.SubstituteByName(conc, subs)
			result = append(result, m.Cfg.AstCfg.NewLabeledFormula(nil, inst))
		})
	}

	return result
}

// MatchSchemaPrems recursively matches schema premises against available
// constants and functions. Calls callback for each successful complete match.
//
// Python: ivy_auto_inst.py match_schema_prems generator
func MatchSchemaPrems(
	prems []goivy.Expr,
	sortConstants map[string][]*goivy.Const,
	funs map[string]bool,
	match *Match,
	boundSorts map[string]bool,
	callback func(map[string]interface{}),
) {
	if len(prems) == 0 {
		callback(match.CopyMap())
		return
	}

	// Process the last premise
	prem := prems[len(prems)-1]
	remainingPrems := prems[:len(prems)-1]

	switch p := prem.(type) {
	case *goivy.UninterpretedSort:
		// Match to known sorts
		for sortName := range sortConstants {
			match.Push()
			if match.Unify(p.Name, sortName) {
				MatchSchemaPrems(remainingPrems, sortConstants, funs, match, boundSorts, callback)
			}
			match.Pop()
		}

	case *goivy.LogicVariable:
		// Match to constants of the appropriate sort
		sortKey := p.VSort.String()
		// Match Python defaultdict auto-vivification
		if _, ok := sortConstants[sortKey]; !ok {
			sortConstants[sortKey] = nil
		}
		consts := sortConstants[sortKey]
		for _, c := range consts {
			match.Push()
			if match.Unify(p.Name, c) {
				MatchSchemaPrems(remainingPrems, sortConstants, funs, match, boundSorts, callback)
			}
			match.Pop()
		}

	case *goivy.Const:
		if goivy.IsFunctionSort(p.CSort) {
			// Match to function symbols
			for funName := range funs {
				match.Push()
				if match.Unify(p.Name, goivy.NewConst(funName, p.CSort)) {
					MatchSchemaPrems(remainingPrems, sortConstants, funs, match, boundSorts, callback)
				}
				match.Pop()
			}
		} else {
			// Match to constants
			sortKey := ""
			if p.CSort != nil {
				sortKey = p.CSort.String()
			}
			// Match Python defaultdict auto-vivification
			if _, ok := sortConstants[sortKey]; !ok {
				sortConstants[sortKey] = nil
			}
			consts := sortConstants[sortKey]
			for _, c := range consts {
				match.Push()
				if match.Unify(p.Name, c) {
					MatchSchemaPrems(remainingPrems, sortConstants, funs, match, boundSorts, callback)
				}
				match.Pop()
			}
		}

	default:
		// Skip unrecognized premise types
		MatchSchemaPrems(remainingPrems, sortConstants, funs, match, boundSorts, callback)
	}
}

// extractSchemaNode tries to get a lg.Expr from a schema interface{}.
func extractSchemaNode(lf interface{}) (goivy.Expr, bool) {
	switch t := lf.(type) {
	case *goivy.LabeledFormula:
		if n, ok := t.Formula.(goivy.Expr); ok {
			return n, true
		}
		return nil, false
	case goivy.Expr:
		return t, true
	}
	return nil, false
}

// GetTrigger finds a trigger expression in a formula that covers all bound variables.
//
// Python: ivy_mc.py:674-684, also used in ivy_auto_inst.py
func GetTrigger(expr goivy.Expr, vars []*goivy.LogicVariable) goivy.Expr {
	if goivy.IsQuantifier(expr) || goivy.IsVariable(expr) {
		return nil
	}

	for _, child := range goivy.NodeArgs(expr) {
		r := GetTrigger(child, vars)
		if r != nil {
			return r
		}
	}

	if goivy.IsApp(expr) || isEqNode(expr) {
		exprVars := goivy.FreeVariablesList(expr)
		if containsAllVars(exprVars, vars) {
			return expr
		}
	}
	return nil
}

func isEqNode(n goivy.Expr) bool {
	_, ok := n.(*goivy.Eq)
	return ok
}

func containsAllVars(have []*goivy.LogicVariable, need []*goivy.LogicVariable) bool {
	haveSet := make(map[string]bool, len(have))
	for _, v := range have {
		haveSet[v.Name] = true
	}
	for _, v := range need {
		if !haveSet[v.Name] {
			return false
		}
	}
	return true
}

// unused avoids "imported and not used" errors
var _ = fmt.Sprint
