package goivy

import (
	"fmt"
	"sort"
)

// CloneNode clones a logic node, replacing its children with the given args.
func CloneNode(n Expr, args []Expr) Expr {
	switch t := n.(type) {
	case *Const:
		return t // constants are immutable
	case *LogicVariable:
		return t // variables are immutable
	case ActionsAction:
		return t.ActionClone(args)
	case *Apply:
		if len(args) > 0 {
			return CloneApplyTerms(t, args)
		}
		return t
	case *Eq:
		if len(args) == 2 {
			return &Eq{T1: args[0], T2: args[1]}
		}
		return t
	case *LogicNot:
		if len(args) == 1 {
			return &LogicNot{Body: args[0]}
		}
		return t
	case *LogicAnd:
		return &LogicAnd{Terms: args}
	case *LogicOr:
		return &LogicOr{Terms: args}
	case *LogicImplies:
		if len(args) == 2 {
			return &LogicImplies{T1: args[0], T2: args[1]}
		}
		return t
	case *LogicIff:
		if len(args) == 2 {
			return &LogicIff{T1: args[0], T2: args[1]}
		}
		return t
	case *LogicIte:
		if len(args) == 3 {
			// Python's Ite.__init__ recomputes sort from t_then.sort
			return &LogicIte{ISort: args[1].NodeSort(), Cond: args[0], Then: args[1], Else: args[2]}
		}
		return t
	case *ForAll:
		if len(args) == 1 {
			return &ForAll{Variables: t.Variables, Body: args[0]}
		}
		return t
	case *RawForAll:
		if len(args) == 1 {
			return &RawForAll{Variables: t.Variables, Body: args[0]}
		}
		return t
	case *LogicExists:
		if len(args) == 1 {
			return &LogicExists{Variables: t.Variables, Body: args[0]}
		}
		return t
	case *Lambda:
		if len(args) == 1 {
			return &Lambda{Variables: t.Variables, Body: args[0]}
		}
		return t
	case *LogicNamedBinder:
		if len(args) == 1 {
			return &LogicNamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: args[0]}
		}
		return t
	case *LogicGlobally:
		if len(args) == 1 {
			return &LogicGlobally{Environ: t.Environ, Body: args[0]}
		}
		return t
	case *LogicEventually:
		if len(args) == 1 {
			return &LogicEventually{Environ: t.Environ, Body: args[0]}
		}
		return t
	case *LogicWhenOperator:
		if len(args) == 2 {
			// Python's WhenOperator.__init__ recomputes sort from t1.sort
			return &LogicWhenOperator{WSort: args[0].NodeSort(), Name: t.Name, T1: args[0], T2: args[1]}
		}
		return t
	case *Cond:
		if len(args) == 2 {
			// Python's Cond.__init__ recomputes sort from t2.sort
			return &Cond{CSort: args[1].NodeSort(), T1: args[0], T2: args[1]}
		}
		return t
	case *IvyDefinition:
		if len(args) == 2 {
			return &IvyDefinition{Lhs: args[0], Rhs: args[1]}
		}
		return t
	case *LogicSome:
		// Python's Some extends AST, so clone = type(self)(*args) replaces all args.
		// Python's Some.args = (params_node, fmla_node, [if_val, [else_val]])
		// Go's Some.Children() = params... + fmla + [if_val] + [else_val]
		// To match Python clone semantics, we reconstruct from all args.
		nParams := len(t.Params)
		if len(args) >= nParams+1 {
			s := &LogicSome{
				Params: args[:nParams],
				Fmla:   args[nParams],
			}
			if len(args) > nParams+1 {
				s.IfVal = args[nParams+1]
			}
			if len(args) > nParams+2 {
				s.ElseVal = args[nParams+2]
			}
			return s
		}
		return t
	}
	return n
}

// CloneBinder clones a binder with new variables and body.
func CloneBinder(n Expr, vars []*LogicVariable, body Expr) Expr {
	switch t := n.(type) {
	case *ForAll:
		return &ForAll{Variables: deduplicateAndSortVars(vars), Body: body}
	case *RawForAll:
		return &RawForAll{Variables: deduplicateAndSortVars(vars), Body: body}
	case *LogicExists:
		return &LogicExists{Variables: deduplicateAndSortVars(vars), Body: body}
	case *Lambda:
		return &Lambda{Variables: vars, Body: body}
	case *LogicNamedBinder:
		return &LogicNamedBinder{Name: t.Name, Variables: vars, Environ: t.Environ, Body: body}
	case *LogicSome:
		// Python: clone_binder(vs, body) → Some(*(vs + self.args[1:]))
		// Replaces params with vs. IGNORES body parameter — keeps original
		// fmla, if_val, else_val from self.args[1:]. This matches Python's
		// Some.clone_binder exactly (the body param is discarded).
		params := make([]Expr, len(vars))
		for i, v := range vars {
			params[i] = v
		}
		_ = body // Python discards the body parameter for Some
		return &LogicSome{Params: params, Fmla: t.Fmla, IfVal: t.IfVal, ElseVal: t.ElseVal}
	}
	return n
}

// BinderVars returns the bound variables of a binder node.
func BinderVars(n Expr) []*LogicVariable {
	switch t := n.(type) {
	case *ForAll:
		return t.Variables
	case *RawForAll:
		return t.Variables
	case *LogicExists:
		return t.Variables
	case *Lambda:
		return t.Variables
	case *LogicNamedBinder:
		return t.Variables
	case *LogicSome:
		// Some's Params are the bound variables
		vars := make([]*LogicVariable, 0, len(t.Params))
		for _, p := range t.Params {
			if v, ok := p.(*LogicVariable); ok {
				vars = append(vars, v)
			}
		}
		return vars
	}
	return nil
}

// BinderBody returns the body of a binder node.
func BinderBody(n Expr) Expr {
	switch t := n.(type) {
	case *ForAll:
		return t.Body
	case *RawForAll:
		return t.Body
	case *LogicExists:
		return t.Body
	case *Lambda:
		return t.Body
	case *LogicNamedBinder:
		return t.Body
	case *LogicSome:
		return t.Fmla
	}
	return nil
}

// NodeArgs returns the arguments of a node (mimics Python's .args property).
func NodeArgs(n Expr) []Expr {
	switch t := n.(type) {
	case *Const:
		return nil
	case *LogicVariable:
		return nil
	case *Apply:
		return t.Terms
	case *Eq:
		return []Expr{t.T1, t.T2}
	case *LogicNot:
		return []Expr{t.Body}
	case *LogicAnd:
		return t.Terms
	case *LogicOr:
		return t.Terms
	case *LogicImplies:
		return []Expr{t.T1, t.T2}
	case *LogicIff:
		return []Expr{t.T1, t.T2}
	case *LogicIte:
		return []Expr{t.Cond, t.Then, t.Else}
	case *ForAll:
		return []Expr{t.Body}
	case *RawForAll:
		return []Expr{t.Body}
	case *LogicExists:
		return []Expr{t.Body}
	case *Lambda:
		return []Expr{t.Body}
	case *LogicNamedBinder:
		return []Expr{t.Body}
	case *LogicGlobally:
		return []Expr{t.Body}
	case *LogicEventually:
		return []Expr{t.Body}
	case *LogicWhenOperator:
		return []Expr{t.T1, t.T2}
	case *Cond:
		return []Expr{t.T1, t.T2}
	case *IvyDefinition:
		return []Expr{t.Lhs, t.Rhs}
	case *LogicLiteral:
		return []Expr{t.Atom}
	}
	return n.Children()
}

// ForAll creates a ForAll node, or returns the body if vars is empty.
func IvyForAll(vs []*LogicVariable, body Expr) Expr {
	if len(vs) == 0 {
		return body
	}
	return &ForAll{Variables: deduplicateAndSortVars(vs), Body: body}
}

// Exists creates an Exists node, or returns the body if vars is empty.
func IvyExists(vs []*LogicVariable, body Expr) Expr {
	if len(vs) == 0 {
		return body
	}
	return &LogicExists{Variables: deduplicateAndSortVars(vs), Body: body}
}

// CloseFormula universally quantifies over all free variables.
func CloseFormula(fmla Expr) Expr {
	fvs := FreeVariablesList(fmla)
	if len(fvs) == 0 {
		return fmla
	}
	return IvyForAll(fvs, fmla)
}

// IsGroundFormula returns true if a formula contains no free variables.
func IsGroundFormula(fmla Expr) bool {
	fvs := FreeVariablesList(fmla)
	return len(fvs) == 0
}

// Extensionality generates an extensionality axiom for a list of destructors.
// Given destructors d1:S→T1, d2:S→T2, ..., returns:
// forall X:S, Y:S. (d1(X) = d1(Y) & d2(X) = d2(Y) & ...) -> X = Y
func Extensionality(destrs []*Const) Expr {
	if len(destrs) == 0 {
		return &LogicOr{} // false
	}
	// Get the sort from the first destructor's domain
	fs, ok := destrs[0].CSort.(*LogicFunctionSort)
	if !ok || fs.Arity() == 0 {
		return &LogicOr{}
	}
	sort := fs.Domain()[0]
	x, _ := NewVariable("X", sort)
	y, _ := NewVariable("Y", sort)

	var conjuncts []Expr
	for _, d := range destrs {
		dfs, ok := d.CSort.(*LogicFunctionSort)
		if !ok {
			continue
		}
		dom := dfs.Domain()
		if len(dom) == 0 {
			continue
		}
		// Create extra variables for multi-arg destructors
		var extraVars []*LogicVariable
		for i := 1; i < len(dom); i++ {
			v, _ := NewVariable(varName(i-1), dom[i])
			extraVars = append(extraVars, v)
		}

		// Build d(X, v0, v1, ...) and d(Y, v0, v1, ...)
		argsX := []Expr{x}
		argsY := []Expr{y}
		for _, v := range extraVars {
			argsX = append(argsX, v)
			argsY = append(argsY, v)
		}

		appX := MustApply(d, argsX...)
		appY := MustApply(d, argsY...)
		eq := &Eq{T1: appX, T2: appY}

		if len(extraVars) > 0 {
			conjuncts = append(conjuncts, &ForAll{Variables: extraVars, Body: eq})
		} else {
			conjuncts = append(conjuncts, eq)
		}
	}

	antecedent := &LogicAnd{Terms: conjuncts}
	consequent := &Eq{T1: x, T2: y}
	return &LogicImplies{T1: antecedent, T2: consequent}
}

func varName(idx int) string {
	return fmt.Sprintf("V%d", idx)
}

// PartialFunction returns a formula stating that rel is a partial function:
// forall X, Y, Z. (rel(X,Y) & rel(X,Z)) -> Y = Z
func PartialFunction(rel *Const) Expr {
	fs, ok := rel.CSort.(*LogicFunctionSort)
	if !ok || fs.Arity() < 2 {
		return &LogicAnd{} // true
	}
	dom := fs.Domain()
	x, _ := NewVariable("X", dom[0])
	y, _ := NewVariable("Y", dom[1])
	z, _ := NewVariable("Z", dom[1])

	relXY := MustApply(rel, x, y)
	relXZ := MustApply(rel, x, z)
	premise := &LogicAnd{Terms: []Expr{relXY, relXZ}}
	conclusion := &Eq{T1: y, T2: z}
	body := &LogicImplies{T1: premise, T2: conclusion}
	return &ForAll{Variables: []*LogicVariable{x, y, z}, Body: body}
}

// VariableUniqifier alpha-converts formulas so all bound variables are unique.
type VariableUniqifier struct {
	rn     *UniqueRenamer
	InvMap map[NodeKey]*LogicVariable // renamed var → original var
}

// NewVariableUniqifier creates a new uniqifier, reserving the given names.
func NewVariableUniqifier(used []string) *VariableUniqifier {
	return &VariableUniqifier{
		rn:     NewUniqueRenamer("", used),
		InvMap: make(map[NodeKey]*LogicVariable),
	}
}

// Uniquify alpha-converts a formula so all bound variables have unique names.
func (vu *VariableUniqifier) Uniquify(fmla Expr) Expr {
	vmap := make(map[NodeKey]*LogicVariable)
	return vu.rec(fmla, vmap)
}

func (vu *VariableUniqifier) rec(fmla Expr, vmap map[NodeKey]*LogicVariable) Expr {
	if IsBinder(fmla) {
		vars := BinderVars(fmla)
		body := BinderBody(fmla)

		// Save old bindings
		type saved struct {
			v   *LogicVariable
			old *LogicVariable
		}
		var obs []saved
		for _, v := range vars {
			if old, ok := vmap[Key(v)]; ok {
				obs = append(obs, saved{v, old})
			}
		}

		// Create new variable names
		newVars := make([]*LogicVariable, len(vars))
		for i, v := range vars {
			newName := vu.rn.Rename(v.Name)
			nv, _ := NewVariable(newName, v.VSort)
			newVars[i] = nv
			vmap[Key(v)] = nv
			vu.InvMap[Key(nv)] = v
		}

		newBody := vu.rec(body, vmap)
		result := CloneBinder(fmla, newVars, newBody)

		// Restore old bindings
		for _, v := range vars {
			delete(vmap, Key(v))
		}
		for _, o := range obs {
			vmap[Key(o.v)] = o.old
		}

		return result
	}

	if v, ok := fmla.(*LogicVariable); ok {
		if mapped, exists := vmap[Key(v)]; exists {
			return mapped
		}
		// Free variable — assign a new unique name
		newName := vu.rn.Rename(v.Name)
		nv, _ := NewVariable(newName, v.VSort)
		vmap[Key(v)] = nv
		vu.InvMap[Key(nv)] = v
		return nv
	}

	args := NodeArgs(fmla)
	if len(args) == 0 {
		return fmla
	}
	newArgs := make([]Expr, len(args))
	for i, a := range args {
		newArgs[i] = vu.rec(a, vmap)
	}
	return CloneNode(fmla, newArgs)
}

// Undo reverses the renaming applied by this uniqifier.
func (vu *VariableUniqifier) Undo(fmla Expr) Expr {
	subs := make(map[NodeKey]Expr)
	for k, v := range vu.InvMap {
		subs[k] = v
	}
	result, _ := Substitute(fmla, subs)
	return result
}

// AlphaAvoid alpha-converts a formula so that bound variable names do not
// clash with the given set of variables.
func AlphaAvoid(fmla Expr, vs []*LogicVariable) Expr {
	vu := NewVariableUniqifier(nil)
	// Reserve names of vs and free variables
	for _, v := range vs {
		vu.rn.Rename(v.Name)
	}
	fvs := FreeVariablesList(fmla)
	vmap := make(map[NodeKey]*LogicVariable)
	for _, v := range fvs {
		vu.rn.Rename(v.Name)
		vmap[Key(v)] = v // preserve free variable
	}
	for _, v := range vs {
		vmap[Key(v)] = v
	}
	return vu.rec(fmla, vmap)
}

// AlphaAvoidMap alpha-converts a formula so that bound variable names do not
// clash with any named symbol in vs. vs is a map[lg.NodeKey]lg.Expr (e.g.,
// from MatchRhsVars) which may contain Variables, Consts, and Sort values.
// Only Variables have names that can clash with bound variables — other
// node types in vs are still reserved (if they have names) to prevent
// accidental shadowing.
//
// Corresponds to Python's alpha_avoid called with match_rhs_vars(match).
func AlphaAvoidMap(fmla Expr, vs map[NodeKey]Expr) Expr {
	vu := NewVariableUniqifier(nil)
	// Reserve names of all values in vs
	for _, v := range vs {
		if vv, ok := v.(*LogicVariable); ok {
			vu.rn.Rename(vv.Name)
		} else if c, ok := v.(*Const); ok {
			vu.rn.Rename(c.Name)
		} else if us, ok := v.(*UninterpretedSort); ok {
			vu.rn.Rename(us.Name)
		}
	}
	fvs := FreeVariablesList(fmla)
	vmap := make(map[NodeKey]*LogicVariable)
	for _, v := range fvs {
		vu.rn.Rename(v.Name)
		vmap[Key(v)] = v // preserve free variable
	}
	// Preserve variable values in vs
	for _, v := range vs {
		if vv, ok := v.(*LogicVariable); ok {
			vmap[Key(vv)] = vv
		}
	}
	return vu.rec(fmla, vmap)
}

// NormalizeOps converts conjunctions and disjunctions to binary ops
// and quantifiers to single-variable quantifiers.
func NormalizeOps(fmla Expr) Expr {
	args := NodeArgs(fmla)
	newArgs := make([]Expr, len(args))
	for i, a := range args {
		newArgs[i] = NormalizeOps(a)
	}

	switch fmla.(type) {
	case *LogicAnd, *LogicOr:
		if len(newArgs) == 0 {
			return CloneNode(fmla, nil)
		}
		return makeBin(fmla, newArgs[0], newArgs[1:])
	case *ForAll:
		vars := BinderVars(fmla)
		sort.Slice(vars, func(i, j int) bool { return vars[i].Name < vars[j].Name })
		return makeQuant(fmla, vars, newArgs[0])
	case *LogicExists:
		vars := BinderVars(fmla)
		sort.Slice(vars, func(i, j int) bool { return vars[i].Name < vars[j].Name })
		return makeQuant(fmla, vars, newArgs[0])
	}

	return CloneNode(fmla, newArgs)
}

func makeBin(proto Expr, first Expr, rest []Expr) Expr {
	if len(rest) == 0 {
		return first
	}
	combined := CloneNode(proto, []Expr{first, rest[0]})
	// For And/Or with more than 2, we need binary nesting
	if len(rest) > 1 {
		return makeBin(proto, combined, rest[1:])
	}
	return combined
}

func makeQuant(proto Expr, vars []*LogicVariable, body Expr) Expr {
	if len(vars) == 0 {
		return body
	}
	inner := makeQuant(proto, vars[1:], body)
	return CloneBinder(proto, vars[0:1], inner)
}

// ASTMatch performs structural matching of x against pattern y.
// Placeholders in y can match any subterm; successful matches are
// recorded in subst.
func ASTMatch(x, y Expr, placeholders map[NodeKey]Expr, subst map[NodeKey]Expr) bool {
	// Type must match
	if typeTag(x) != typeTag(y) {
		return false
	}

	// Variable or constant placeholder
	if IsVariable(y) || IsConstant(y) {
		if placeholders != nil && placeholders[Key(y)] != nil {
			if prev, ok := subst[Key(y)]; ok {
				return x.Equal(prev)
			}
			subst[Key(y)] = x
			return true
		}
		return x.Equal(y)
	}

	// Apply: check func and args
	if appX, ok := x.(*Apply); ok {
		appY := y.(*Apply)
		if !appX.Func.Equal(appY.Func) {
			return false
		}
		return astMatchLists(NodeArgs(x), NodeArgs(y), placeholders, subst)
	}

	// Literal
	if litX, ok := x.(*LogicLiteral); ok {
		litY := y.(*LogicLiteral)
		if litX.Polarity != litY.Polarity {
			return false
		}
		return ASTMatch(litX.Atom, litY.Atom, placeholders, subst)
	}

	// Generic: match args
	xArgs := NodeArgs(x)
	yArgs := NodeArgs(y)
	return astMatchLists(xArgs, yArgs, placeholders, subst)
}

func astMatchLists(xs, ys []Expr, placeholders map[NodeKey]Expr, subst map[NodeKey]Expr) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if !ASTMatch(xs[i], ys[i], placeholders, subst) {
			return false
		}
	}
	return true
}

func typeTag(n Expr) string {
	switch n.(type) {
	case *LogicVariable:
		return "Var"
	case *Const:
		return "Const"
	case *Apply:
		return "Apply"
	case *Eq:
		return "Eq"
	case *LogicNot:
		return "Not"
	case *LogicAnd:
		return "And"
	case *LogicOr:
		return "Or"
	case *LogicImplies:
		return "Implies"
	case *LogicIff:
		return "Iff"
	case *LogicIte:
		return "Ite"
	case *ForAll:
		return "ForAll"
	case *LogicExists:
		return "Exists"
	case *Lambda:
		return "Lambda"
	case *LogicNamedBinder:
		return "NamedBinder"
	case *LogicGlobally:
		return "Globally"
	case *LogicEventually:
		return "Eventually"
	case *LogicWhenOperator:
		return "WhenOperator"
	case *Cond:
		return "Cond"
	case *IvyDefinition:
		return "IvyDefinition"
	case *LogicLiteral:
		return "Literal"
	case *LogicSome:
		return "Some"
	default:
		return "unknown"
	}
}

// LabelTemporal labels temporal operators with a given label string.
func LabelTemporal(fmla Expr, label string) Expr {
	switch t := fmla.(type) {
	case *LogicGlobally:
		return &LogicGlobally{Environ: &label, Body: LabelTemporal(t.Body, label)}
	case *LogicEventually:
		return &LogicEventually{Environ: &label, Body: LabelTemporal(t.Body, label)}
	case *LogicWhenOperator:
		return &LogicWhenOperator{
			WSort: t.WSort,
			Name:  t.Name,
			T1:    LabelTemporal(t.T1, label),
			T2:    LabelTemporal(t.T2, label),
		}
	case *LogicNamedBinder:
		return &LogicNamedBinder{
			Name:      t.Name,
			Variables: t.Variables,
			Environ:   &label,
			Body:      LabelTemporal(t.Body, label),
		}
	case *Apply:
		newFunc := LabelTemporal(t.Func, label)
		newArgs := make([]Expr, len(t.Terms))
		for i, a := range t.Terms {
			newArgs[i] = LabelTemporal(a, label)
		}
		return MustApply(newFunc, newArgs...)
	}
	args := NodeArgs(fmla)
	newArgs := make([]Expr, len(args))
	for i, a := range args {
		newArgs[i] = LabelTemporal(a, label)
	}
	return CloneNode(fmla, newArgs)
}
