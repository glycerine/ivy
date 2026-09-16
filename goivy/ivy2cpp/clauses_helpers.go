package ivy2cpp

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// Pure ivylogic helpers ported from Python ivy_to_cpp.py:1012-1175.
// Used by emit_action_gen to massage the reverse_image precondition
// before encoding it for the Z3 solver. Contains no C++ emission logic.
//
// Function -> Python source line:
//
//	isLocalSym                       1012
//	fixDefinition                    1016
//	extractDefinedParameters         1027
//	collectUsedDefinitions           1048
//	minimalFieldReferences           1093
//	minimalFieldSiblings             1126
//	extractInputFields               1142
//	expandFieldReferences            1161

// isLocalSym reports whether sym is a local (non-signature, non-constructor,
// solver-representable) symbol. Mirrors Python `is_local_sym`.
//
// Python:
//
//	def is_local_sym(sym):
//	    sym = il.normalize_symbol(sym)
//	    return (not il.sig.contains_symbol(sym)
//	            and slv.solver_name(il.normalize_symbol(sym)) != None
//	            and sym not in il.sig.constructors)
func isLocalSym(sym *goivy.Const, sig *goivy.Sig) bool {
	if sym == nil || sig == nil {
		return false
	}
	if sig.ContainsSymbol(sym.Name, sym.CSort) {
		return false
	}
	if sig.Constructors[sym.Name] {
		return false
	}
	name, err := goivy.SolverName(sym, sig, nil)
	if err != nil || name == "" {
		return false
	}
	return true
}

// fixDefinition replaces non-variable LHS arguments of df with fresh
// `X__N` skolem variables so the definition's LHS is normalized for
// further processing. Mirrors Python `fix_definition`.
func fixDefinition(df *goivy.LogicDefinition) *goivy.LogicDefinition {
	if df == nil {
		return df
	}
	lhsApp, ok := df.Lhs.(*goivy.Apply)
	if !ok {
		return df
	}
	subs := make(map[goivy.NodeKey]goivy.Expr)
	allVars := true
	for idx, arg := range lhsApp.Terms {
		if _, isVar := arg.(*goivy.LogicVariable); isVar {
			continue
		}
		allVars = false
		v, err := goivy.NewVariable(fmt.Sprintf("X__%d", idx), arg.NodeSort())
		if err != nil {
			return df
		}
		subs[goivy.Key(arg)] = v
	}
	if allVars {
		return df
	}
	newLhs, err := goivy.Substitute(df.Lhs, subs)
	if err != nil {
		return df
	}
	newRhs, err := goivy.Substitute(df.Rhs, subs)
	if err != nil {
		return df
	}
	return &goivy.LogicDefinition{Lhs: newLhs, Rhs: newRhs}
}

// keyOfConst is a small helper that yields the NodeKey for a *Const.
// Centralizing it makes the helpers easier to read.
func keyOfConst(c *goivy.Const) goivy.NodeKey { return goivy.Key(c) }

// extractDefinedParameters strips Eq formulas of the form Eq(input, expr)
// from pre.Fmlas where input is in inputs and does not recur elsewhere.
// Returns the trimmed clauses and the extracted parameter definitions.
// Mirrors Python `extract_defined_parameters`.
func extractDefinedParameters(pre *goivy.Clauses, inputs []*goivy.Const) (*goivy.Clauses, []goivy.Expr) {
	if pre == nil {
		return pre, nil
	}
	inputSet := make(map[goivy.NodeKey]bool, len(inputs))
	for _, in := range inputs {
		if in == nil {
			continue
		}
		inputSet[keyOfConst(in)] = true
	}
	// defmap : input -> formula that defines it.
	defmap := make(map[goivy.NodeKey]goivy.Expr)
	for _, fmla := range pre.Fmlas {
		lhsConst, ok := definedParameterLHS(fmla)
		if !ok {
			continue
		}
		k := keyOfConst(lhsConst)
		if !inputSet[k] {
			continue
		}
		if _, exists := defmap[k]; exists {
			continue
		}
		defmap[k] = fmla
	}
	for _, def := range pre.Defs {
		lhsConst, ok := definedParameterLHS(def)
		if !ok {
			continue
		}
		k := keyOfConst(lhsConst)
		if !inputSet[k] {
			continue
		}
		if _, exists := defmap[k]; exists {
			continue
		}
		defmap[k] = def
	}
	current := pre
	var inpdefs []goivy.Expr
	change := true
	for change {
		change = false
		for inKey, fmla := range defmap {
			recurs := false
			for _, f := range current.Fmlas {
				if f == fmla {
					continue
				}
				if symbolKeysContain(goivy.UsedSymbolsAst(f), inKey) {
					recurs = true
					break
				}
			}
			if !recurs {
				for _, d := range current.Defs {
					if fmla == d {
						continue
					}
					if symbolKeysContain(goivy.UsedSymbolsAst(d.Rhs), inKey) ||
						symbolKeysContain(goivy.UsedSymbolsAst(d.Lhs), inKey) {
						recurs = true
						break
					}
				}
			}
			if recurs {
				continue
			}
			// Drop fmla from current.Fmlas/current.Defs.
			newFmlas := make([]goivy.Expr, 0, len(current.Fmlas))
			for _, f := range current.Fmlas {
				if f == fmla {
					continue
				}
				newFmlas = append(newFmlas, f)
			}
			newDefs := make([]*goivy.IvyDefinition, 0, len(current.Defs))
			for _, d := range current.Defs {
				if fmla == d {
					continue
				}
				newDefs = append(newDefs, d)
			}
			current = goivy.NewClauses(newFmlas, newDefs, current.Annot)
			current = goivy.TrimClauses(current)
			delete(defmap, inKey)
			inpdefs = append(inpdefs, fmla)
			change = true
			break
		}
	}
	// Python `inpdefs.reverse()`.
	for i, j := 0, len(inpdefs)-1; i < j; i, j = i+1, j-1 {
		inpdefs[i], inpdefs[j] = inpdefs[j], inpdefs[i]
	}
	return current, inpdefs
}

func definedParameterLHS(fmla goivy.Expr) (*goivy.Const, bool) {
	switch t := fmla.(type) {
	case *goivy.Eq:
		c, ok := t.T1.(*goivy.Const)
		return c, ok
	case *goivy.LogicIff:
		c, ok := t.T1.(*goivy.Const)
		return c, ok
	case *goivy.LogicDefinition:
		c, ok := t.Defines().(*goivy.Const)
		return c, ok
	default:
		return nil, false
	}
}

func symbolKeysContain(m *goivy.InsMap[goivy.NodeKey, goivy.Expr], k goivy.NodeKey) bool {
	if m == nil {
		return false
	}
	_, ok := m.Get2(k)
	return ok
}

// collectUsedDefinitions walks the RHS of each parameter definition in
// inpdefs and gathers every definition reachable through pre.Defs, plus
// the state symbols those definitions depend on. Mirrors Python
// `collect_used_definitions`.
func collectUsedDefinitions(
	pre *goivy.Clauses,
	inpdefs []goivy.Expr,
	ssyms map[goivy.NodeKey]bool,
) ([]*goivy.LogicDefinition, []*goivy.Const) {
	if pre == nil {
		return nil, nil
	}
	defmap := make(map[goivy.NodeKey]*goivy.LogicDefinition, len(pre.Defs))
	for _, d := range pre.Defs {
		defmap[goivy.Key(d.Defines())] = d
	}
	used := make(map[goivy.NodeKey]bool)
	var res []*goivy.LogicDefinition
	var usyms []*goivy.Const
	var recur func(rhs goivy.Expr)
	recur = func(rhs goivy.Expr) {
		for k, sym := range goivy.UsedSymbolsAst(rhs).All() {
			if used[k] {
				continue
			}
			used[k] = true
			if d, ok := defmap[k]; ok {
				recur(d.Rhs)
				res = append(res, d)
				continue
			}
			if ssyms[k] {
				if c, ok := sym.(*goivy.Const); ok {
					usyms = append(usyms, c)
				}
			}
		}
	}
	for _, inpdef := range inpdefs {
		eq, ok := inpdef.(*goivy.Eq)
		if !ok {
			continue
		}
		recur(eq.T2)
	}
	return res, usyms
}

// minimalFieldReferences walks fmla and, for every apply chain that ends
// at an `inputs` constant via a sequence of destructor unary applies,
// records the maximal-depth reference. After collection, prunes each
// input's set to its minima — references not subsumed by another in the
// same set. Mirrors Python `minimal_field_references`.
func minimalFieldReferences(
	fmla goivy.Expr,
	inputs []*goivy.Const,
	destructorSorts map[string]goivy.Sort,
) map[goivy.NodeKey][]goivy.Expr {
	inpset := make(map[goivy.NodeKey]bool, len(inputs))
	inpConst := make(map[goivy.NodeKey]*goivy.Const, len(inputs))
	for _, in := range inputs {
		if in == nil {
			continue
		}
		k := keyOfConst(in)
		inpset[k] = true
		inpConst[k] = in
	}
	// res maps an input-Symbol (key) to distinct apply references rooted
	// at it, preserving first traversal order.
	resSet := make(map[goivy.NodeKey]map[goivy.NodeKey]bool)
	resOrder := make(map[goivy.NodeKey][]goivy.Expr)
	add := func(rootKey goivy.NodeKey, ref goivy.Expr) {
		if resSet[rootKey] == nil {
			resSet[rootKey] = make(map[goivy.NodeKey]bool)
		}
		k := goivy.Key(ref)
		if resSet[rootKey][k] {
			return
		}
		resSet[rootKey][k] = true
		resOrder[rootKey] = append(resOrder[rootKey], ref)
	}
	isDestrApply := func(f *goivy.Apply) bool {
		c, ok := f.Func.(*goivy.Const)
		if !ok {
			return false
		}
		_, ok = destructorSorts[c.Name]
		return ok && len(f.Terms) == 1
	}
	var fieldRef func(f goivy.Expr) (*goivy.Const, bool)
	fieldRef = func(f goivy.Expr) (*goivy.Const, bool) {
		switch n := f.(type) {
		case *goivy.Apply:
			if isDestrApply(n) {
				return fieldRef(n.Terms[0])
			}
			c, ok := n.Func.(*goivy.Const)
			if !ok {
				return nil, false
			}
			if inpset[keyOfConst(c)] {
				return c, true
			}
		case *goivy.Const:
			if inpset[keyOfConst(n)] {
				return n, true
			}
		}
		return nil, false
	}
	var walk func(f goivy.Expr)
	walk = func(f goivy.Expr) {
		switch n := f.(type) {
		case *goivy.Apply:
			if isDestrApply(n) {
				if root, ok := fieldRef(n.Terms[0]); ok {
					add(keyOfConst(root), n)
					return
				}
			}
		case *goivy.Const:
			if inpset[keyOfConst(n)] {
				add(keyOfConst(n), n)
				return
			}
		}
		for _, c := range f.Children() {
			walk(c)
		}
	}
	walk(fmla)
	// Per-input minima.
	out := make(map[goivy.NodeKey][]goivy.Expr, len(resSet))
	for _, in := range inputs {
		if in == nil {
			continue
		}
		rootKey := keyOfConst(in)
		refs := resOrder[rootKey]
		if len(refs) == 0 {
			continue
		}
		out[rootKey] = filterMinimaRefs(refs)
		_ = inpConst[rootKey] // value kept for potential debug
	}
	return out
}

// filterMinimaRefs keeps references y such that no other reference x in
// the list is strictly a subterm chain of y (Python `lt`).
func filterMinimaRefs(refs []goivy.Expr) []goivy.Expr {
	var ltChain func(x, y goivy.Expr) bool
	ltChain = func(x, y goivy.Expr) bool {
		ap, ok := y.(*goivy.Apply)
		if !ok || len(ap.Terms) != 1 {
			return false
		}
		if x.Equal(ap.Terms[0]) {
			return true
		}
		return ltChain(x, ap.Terms[0])
	}
	out := make([]goivy.Expr, 0, len(refs))
	for _, y := range refs {
		dominated := false
		for _, x := range refs {
			if x == y {
				continue
			}
			if ltChain(x, y) {
				dominated = true
				break
			}
		}
		if !dominated {
			out = append(out, y)
		}
	}
	return out
}

// minimalFieldSiblings expands each input's minimal-field-reference set
// to include every destructor sibling of the field, ensuring all fields
// of the record participate in the solver query. Mirrors Python
// `minimal_field_siblings`.
func minimalFieldSiblings(
	inputs []*goivy.Const,
	mrefs map[goivy.NodeKey][]goivy.Expr,
	sortDestructors *goivy.InsMap[string, []*goivy.Const],
) map[goivy.NodeKey][]goivy.Expr {
	out := make(map[goivy.NodeKey][]goivy.Expr, len(inputs))
	for _, in := range inputs {
		if in == nil {
			continue
		}
		k := keyOfConst(in)
		refs, ok := mrefs[k]
		if !ok || len(refs) == 0 {
			out[k] = []goivy.Expr{in}
			continue
		}
		seen := make(map[goivy.NodeKey]bool)
		var siblings []goivy.Expr
		add := func(e goivy.Expr) {
			key := goivy.Key(e)
			if seen[key] {
				return
			}
			seen[key] = true
			siblings = append(siblings, e)
		}
		for _, f := range refs {
			ap, isApp := f.(*goivy.Apply)
			if isApp && len(ap.Terms) == 1 {
				// f.rep.sort.dom[0] in Python is the receiver sort.
				fc, isConst := ap.Func.(*goivy.Const)
				if !isConst {
					add(in)
					continue
				}
				fs, isFn := fc.CSort.(*goivy.LogicFunctionSort)
				if !isFn || len(fs.Domain()) == 0 {
					add(in)
					continue
				}
				recvSort := fs.Domain()[0]
				destrs, ok := sortDestructors.Get2(sortName(recvSort))
				if !ok {
					add(in)
					continue
				}
				for _, d := range destrs {
					applied, err := goivy.NewApply(d, ap.Terms[0])
					if err != nil {
						applied = goivy.NewApplyUnchecked(d, ap.Terms[0])
					}
					add(applied)
				}
			} else {
				add(in)
			}
		}
		out[k] = siblings
	}
	return out
}

// extractInputFields rewrites pre to use fresh field-symbol constants in
// place of destructor applications rooted at inputs, returning the
// rewritten clauses, the new input list, and the fsyms map from
// fresh-symbol-key to the original field expression. Mirrors Python
// `extract_input_fields`.
func extractInputFields(
	pre *goivy.Clauses,
	inputs []*goivy.Const,
	mod *goivy.Module,
) (*goivy.Clauses, []*goivy.Const, map[goivy.NodeKey]goivy.Expr) {
	if pre == nil || mod == nil {
		return pre, inputs, nil
	}
	mrefs := minimalFieldReferences(pre.ToFormula(), inputs, mod.DestructorSorts)
	mrefs = unionMrefs(mrefs, minimalFieldSiblings(inputs, mrefs, mod.SortDestructors))
	// Python's extract_input_fields actually rebinds `mrefs` to
	// `minimal_field_siblings`, NOT to the union — match that:
	mrefs = minimalFieldSiblings(inputs, minimalFieldReferences(pre.ToFormula(), inputs, mod.DestructorSorts), mod.SortDestructors)

	// Build fsyms (new field-symbol -> field expression) and rfsyms
	// (key of field expression -> new symbol). Python:
	//   fsyms = {Symbol(field_name(y), y.sort): y for refs in mrefs.values() for y in refs}
	//   rfsyms = {v: k for k, v in fsyms.items()}
	type fieldEntry struct {
		sym  *goivy.Const
		expr goivy.Expr
	}
	var ordered []fieldEntry
	rfsyms := make(map[goivy.NodeKey]*goivy.Const)
	for _, in := range inputs {
		if in == nil {
			continue
		}
		refs := mrefs[keyOfConst(in)]
		for _, y := range refs {
			name := fieldSymbolName(y)
			c := goivy.NewConst(name, y.NodeSort())
			if _, dup := rfsyms[goivy.Key(y)]; dup {
				continue
			}
			rfsyms[goivy.Key(y)] = c
			ordered = append(ordered, fieldEntry{sym: c, expr: y})
		}
	}
	fsyms := make(map[goivy.NodeKey]goivy.Expr, len(ordered))
	for _, e := range ordered {
		fsyms[goivy.Key(e.sym)] = e.expr
	}
	// Recursive rewrite replacing field expressions with their fresh symbol.
	destrSorts := mod.DestructorSorts
	isDestrApply := func(f *goivy.Apply) bool {
		c, ok := f.Func.(*goivy.Const)
		if !ok {
			return false
		}
		_, ok = destrSorts[c.Name]
		return ok && len(f.Terms) == 1
	}
	var recur func(f goivy.Expr) goivy.Expr
	recur = func(f goivy.Expr) goivy.Expr {
		if f == nil {
			return nil
		}
		// Replace at this level if f matches a key in rfsyms AND its root
		// is in mrefs or it's a destructor apply.
		if ap, ok := f.(*goivy.Apply); ok {
			matchRoot := false
			if c, ok := ap.Func.(*goivy.Const); ok {
				_, inMrefs := mrefs[keyOfConst(c)]
				if inMrefs {
					matchRoot = true
				}
			}
			if matchRoot || isDestrApply(ap) {
				if c, ok := rfsyms[goivy.Key(f)]; ok {
					return c
				}
			}
		}
		if c, ok := f.(*goivy.Const); ok {
			if repl, ok := rfsyms[goivy.Key(c)]; ok {
				return repl
			}
		}
		// Recurse into children.
		children := f.Children()
		if len(children) == 0 {
			return f
		}
		newArgs := make([]goivy.Node, len(children))
		for i, c := range children {
			newArgs[i] = recur(c).(goivy.Node)
		}
		return f.Clone(newArgs).(goivy.Expr)
	}
	newFmlas := make([]goivy.Expr, len(pre.Fmlas))
	for i, f := range pre.Fmlas {
		newFmlas[i] = recur(f)
	}
	newDefs := make([]*goivy.IvyDefinition, len(pre.Defs))
	for i, d := range pre.Defs {
		nd := &goivy.LogicDefinition{Lhs: recur(d.Lhs), Rhs: recur(d.Rhs)}
		newDefs[i] = nd
	}
	newInputs := make([]*goivy.Const, 0, len(ordered))
	for _, e := range ordered {
		newInputs = append(newInputs, e.sym)
	}
	return goivy.NewClauses(newFmlas, newDefs, pre.Annot), newInputs, fsyms
}

// unionMrefs is a small utility used while reading Python; kept for
// readability though Python ultimately rebinds mrefs to siblings only.
func unionMrefs(a, b map[goivy.NodeKey][]goivy.Expr) map[goivy.NodeKey][]goivy.Expr {
	out := make(map[goivy.NodeKey][]goivy.Expr, len(a)+len(b))
	for k, v := range a {
		out[k] = append(out[k], v...)
	}
	for k, v := range b {
		out[k] = append(out[k], v...)
	}
	return out
}

func fieldSymbolName(f goivy.Expr) string {
	if ap, ok := f.(*goivy.Apply); ok && len(ap.Terms) == 1 {
		recv := fieldSymbolName(ap.Terms[0])
		if c, isConst := ap.Func.(*goivy.Const); isConst {
			return recv + "__" + c.Name
		}
	}
	if c, ok := f.(*goivy.Const); ok {
		return c.Name
	}
	if ap, ok := f.(*goivy.Apply); ok {
		if c, isConst := ap.Func.(*goivy.Const); isConst {
			return c.Name
		}
	}
	return ""
}

// expandFieldReferences inlines definitions of the form sym = expr where
// expr is either a constant or a unary destructor application. Mirrors
// Python `expand_field_references`.
func expandFieldReferences(pre *goivy.Clauses, destrSorts map[string]goivy.Sort) *goivy.Clauses {
	if pre == nil {
		return pre
	}
	defmap := make(map[goivy.NodeKey]goivy.Expr)
	for _, d := range pre.Defs {
		// LHS must be a bare symbol (no args).
		lhsApp, lhsIsApp := d.Lhs.(*goivy.Apply)
		if lhsIsApp && len(lhsApp.Terms) != 0 {
			continue
		}
		var lhsKey goivy.NodeKey
		if lhsIsApp {
			c, ok := lhsApp.Func.(*goivy.Const)
			if !ok {
				continue
			}
			lhsKey = keyOfConst(c)
		} else {
			c, ok := d.Lhs.(*goivy.Const)
			if !ok {
				continue
			}
			lhsKey = keyOfConst(c)
		}
		// RHS must be a constant or a unary destructor application.
		switch rhs := d.Rhs.(type) {
		case *goivy.Const:
			defmap[lhsKey] = rhs
		case *goivy.Apply:
			if len(rhs.Terms) == 0 {
				defmap[lhsKey] = rhs
				continue
			}
			if len(rhs.Terms) == 1 {
				c, ok := rhs.Func.(*goivy.Const)
				if !ok {
					continue
				}
				if _, isDestr := destrSorts[c.Name]; isDestr {
					defmap[lhsKey] = rhs
				}
			}
		}
	}
	var recur func(f goivy.Expr) goivy.Expr
	recur = func(f goivy.Expr) goivy.Expr {
		if f == nil {
			return nil
		}
		// Match Python: if il.is_app(f) and f.rep in defmap.
		if ap, ok := f.(*goivy.Apply); ok {
			if c, isConst := ap.Func.(*goivy.Const); isConst {
				if repl, found := defmap[keyOfConst(c)]; found {
					return recur(repl)
				}
			}
		}
		if c, ok := f.(*goivy.Const); ok {
			if repl, found := defmap[keyOfConst(c)]; found {
				return recur(repl)
			}
		}
		children := f.Children()
		if len(children) == 0 {
			return f
		}
		newArgs := make([]goivy.Node, len(children))
		for i, c := range children {
			newArgs[i] = recur(c).(goivy.Node)
		}
		return f.Clone(newArgs).(goivy.Expr)
	}
	newFmlas := make([]goivy.Expr, len(pre.Fmlas))
	for i, f := range pre.Fmlas {
		newFmlas[i] = recur(f)
	}
	var newDefs []*goivy.IvyDefinition
	for _, d := range pre.Defs {
		nd := &goivy.LogicDefinition{Lhs: d.Lhs, Rhs: recur(d.Rhs)}
		if nd.Lhs.Equal(nd.Rhs) {
			continue
		}
		newDefs = append(newDefs, nd)
	}
	return goivy.NewClauses(newFmlas, newDefs, pre.Annot)
}
