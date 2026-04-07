package ivylogic

import (
	"fmt"

	lg "github.com/glycerine/ivy/goivy/logic"
	lu "github.com/glycerine/ivy/goivy/logicutil"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
)

// CloneNode clones a logic node, replacing its children with the given args.
func CloneNode(n lg.Expr, args []lg.Expr) lg.Expr {
	switch t := n.(type) {
	case *lg.Const:
		return t // constants are immutable
	case *lg.Variable:
		return t // variables are immutable
	case *lg.Apply:
		if len(args) > 0 {
			app, err := lg.NewApply(t.Func, args...)
			if err != nil {
				panic(fmt.Sprintf("CloneNode: NewApply failed: %v", err))
			}
			return app
		}
		return t
	case *lg.Eq:
		if len(args) == 2 {
			return &lg.Eq{T1: args[0], T2: args[1]}
		}
		return t
	case *lg.Not:
		if len(args) == 1 {
			return &lg.Not{Body: args[0]}
		}
		return t
	case *lg.And:
		return &lg.And{Terms: args}
	case *lg.Or:
		return &lg.Or{Terms: args}
	case *lg.Implies:
		if len(args) == 2 {
			return &lg.Implies{T1: args[0], T2: args[1]}
		}
		return t
	case *lg.Iff:
		if len(args) == 2 {
			return &lg.Iff{T1: args[0], T2: args[1]}
		}
		return t
	case *lg.Ite:
		if len(args) == 3 {
			// Python's Ite.__init__ recomputes sort from t_then.sort
			return &lg.Ite{ISort: args[1].NodeSort(), Cond: args[0], Then: args[1], Else: args[2]}
		}
		return t
	case *lg.ForAll:
		if len(args) == 1 {
			return &lg.ForAll{Variables: t.Variables, Body: args[0]}
		}
		return t
	case *lg.Exists:
		if len(args) == 1 {
			return &lg.Exists{Variables: t.Variables, Body: args[0]}
		}
		return t
	case *lg.Lambda:
		if len(args) == 1 {
			return &lg.Lambda{Variables: t.Variables, Body: args[0]}
		}
		return t
	case *lg.NamedBinder:
		if len(args) == 1 {
			return &lg.NamedBinder{Name: t.Name, Variables: t.Variables, Environ: t.Environ, Body: args[0]}
		}
		return t
	case *lg.Globally:
		if len(args) == 1 {
			return &lg.Globally{Environ: t.Environ, Body: args[0]}
		}
		return t
	case *lg.Eventually:
		if len(args) == 1 {
			return &lg.Eventually{Environ: t.Environ, Body: args[0]}
		}
		return t
	case *lg.WhenOperator:
		if len(args) == 2 {
			// Python's WhenOperator.__init__ recomputes sort from t1.sort
			return &lg.WhenOperator{WSort: args[0].NodeSort(), Name: t.Name, T1: args[0], T2: args[1]}
		}
		return t
	case *lg.Cond:
		if len(args) == 2 {
			// Python's Cond.__init__ recomputes sort from t2.sort
			return &lg.Cond{CSort: args[1].NodeSort(), T1: args[0], T2: args[1]}
		}
		return t
	case *Definition:
		if len(args) == 2 {
			return &Definition{Lhs: args[0], Rhs: args[1]}
		}
		return t
	case *Some:
		// Python's Some extends AST, so clone = type(self)(*args) replaces all args.
		// Python's Some.args = (params_node, fmla_node, [if_val, [else_val]])
		// Go's Some.Children() = params... + fmla + [if_val] + [else_val]
		// To match Python clone semantics, we reconstruct from all args.
		nParams := len(t.Params)
		if len(args) >= nParams+1 {
			s := &Some{
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
func CloneBinder(n lg.Expr, vars []*lg.Variable, body lg.Expr) lg.Expr {
	switch t := n.(type) {
	case *lg.ForAll:
		return &lg.ForAll{Variables: vars, Body: body}
	case *lg.Exists:
		return &lg.Exists{Variables: vars, Body: body}
	case *lg.Lambda:
		return &lg.Lambda{Variables: vars, Body: body}
	case *lg.NamedBinder:
		return &lg.NamedBinder{Name: t.Name, Variables: vars, Environ: t.Environ, Body: body}
	case *Some:
		// Python: clone_binder(vs, body) → Some(*(vs + self.args[1:]))
		// Replaces params with vs. IGNORES body parameter — keeps original
		// fmla, if_val, else_val from self.args[1:]. This matches Python's
		// Some.clone_binder exactly (the body param is discarded).
		params := make([]lg.Expr, len(vars))
		for i, v := range vars {
			params[i] = v
		}
		_ = body // Python discards the body parameter for Some
		return &Some{Params: params, Fmla: t.Fmla, IfVal: t.IfVal, ElseVal: t.ElseVal}
	}
	return n
}

// BinderVars returns the bound variables of a binder node.
func BinderVars(n lg.Expr) []*lg.Variable {
	switch t := n.(type) {
	case *lg.ForAll:
		return t.Variables
	case *lg.Exists:
		return t.Variables
	case *lg.Lambda:
		return t.Variables
	case *lg.NamedBinder:
		return t.Variables
	case *Some:
		// Some's Params are the bound variables
		vars := make([]*lg.Variable, 0, len(t.Params))
		for _, p := range t.Params {
			if v, ok := p.(*lg.Variable); ok {
				vars = append(vars, v)
			}
		}
		return vars
	}
	return nil
}

// BinderBody returns the body of a binder node.
func BinderBody(n lg.Expr) lg.Expr {
	switch t := n.(type) {
	case *lg.ForAll:
		return t.Body
	case *lg.Exists:
		return t.Body
	case *lg.Lambda:
		return t.Body
	case *lg.NamedBinder:
		return t.Body
	case *Some:
		return t.Fmla
	}
	return nil
}

// NodeArgs returns the arguments of a node (mimics Python's .args property).
func NodeArgs(n lg.Expr) []lg.Expr {
	switch t := n.(type) {
	case *lg.Const:
		return nil
	case *lg.Variable:
		return nil
	case *lg.Apply:
		return t.Terms
	case *lg.Eq:
		return []lg.Expr{t.T1, t.T2}
	case *lg.Not:
		return []lg.Expr{t.Body}
	case *lg.And:
		return t.Terms
	case *lg.Or:
		return t.Terms
	case *lg.Implies:
		return []lg.Expr{t.T1, t.T2}
	case *lg.Iff:
		return []lg.Expr{t.T1, t.T2}
	case *lg.Ite:
		return []lg.Expr{t.Cond, t.Then, t.Else}
	case *lg.ForAll:
		return []lg.Expr{t.Body}
	case *lg.Exists:
		return []lg.Expr{t.Body}
	case *lg.Lambda:
		return []lg.Expr{t.Body}
	case *lg.NamedBinder:
		return []lg.Expr{t.Body}
	case *lg.Globally:
		return []lg.Expr{t.Body}
	case *lg.Eventually:
		return []lg.Expr{t.Body}
	case *lg.WhenOperator:
		return []lg.Expr{t.T1, t.T2}
	case *lg.Cond:
		return []lg.Expr{t.T1, t.T2}
	case *Definition:
		return []lg.Expr{t.Lhs, t.Rhs}
	case *Literal:
		return []lg.Expr{t.Atom}
	}
	return n.Children()
}

// ForAll creates a ForAll node, or returns the body if vars is empty.
func ForAll(vs []*lg.Variable, body lg.Expr) lg.Expr {
	if len(vs) == 0 {
		return body
	}
	return &lg.ForAll{Variables: vs, Body: body}
}

// Exists creates an Exists node, or returns the body if vars is empty.
func Exists(vs []*lg.Variable, body lg.Expr) lg.Expr {
	if len(vs) == 0 {
		return body
	}
	return &lg.Exists{Variables: vs, Body: body}
}

// CloseFormula universally quantifies over all free variables.
func CloseFormula(fmla lg.Expr) lg.Expr {
	fvs := lu.FreeVariablesList(fmla)
	if len(fvs) == 0 {
		return fmla
	}
	return &lg.ForAll{Variables: fvs, Body: fmla}
}

// IsGroundFormula returns true if a formula contains no free variables.
func IsGroundFormula(fmla lg.Expr) bool {
	fvs := lu.FreeVariablesList(fmla)
	return len(fvs) == 0
}

// Extensionality generates an extensionality axiom for a list of destructors.
// Given destructors d1:S→T1, d2:S→T2, ..., returns:
// forall X:S, Y:S. (d1(X) = d1(Y) & d2(X) = d2(Y) & ...) -> X = Y
func Extensionality(destrs []*lg.Const) lg.Expr {
	if len(destrs) == 0 {
		return &lg.Or{} // false
	}
	// Get the sort from the first destructor's domain
	fs, ok := destrs[0].CSort.(*lg.FunctionSort)
	if !ok || fs.Arity() == 0 {
		return &lg.Or{}
	}
	sort := fs.Domain()[0]
	x, _ := lg.NewVariable("X", sort)
	y, _ := lg.NewVariable("Y", sort)

	var conjuncts []lg.Expr
	for _, d := range destrs {
		dfs, ok := d.CSort.(*lg.FunctionSort)
		if !ok {
			continue
		}
		dom := dfs.Domain()
		if len(dom) == 0 {
			continue
		}
		// Create extra variables for multi-arg destructors
		var extraVars []*lg.Variable
		for i := 1; i < len(dom); i++ {
			v, _ := lg.NewVariable(varName(i-1), dom[i])
			extraVars = append(extraVars, v)
		}

		// Build d(X, v0, v1, ...) and d(Y, v0, v1, ...)
		argsX := []lg.Expr{x}
		argsY := []lg.Expr{y}
		for _, v := range extraVars {
			argsX = append(argsX, v)
			argsY = append(argsY, v)
		}

		appX := &lg.Apply{Func: d, Terms: argsX}
		appY := &lg.Apply{Func: d, Terms: argsY}
		eq := &lg.Eq{T1: appX, T2: appY}

		if len(extraVars) > 0 {
			conjuncts = append(conjuncts, &lg.ForAll{Variables: extraVars, Body: eq})
		} else {
			conjuncts = append(conjuncts, eq)
		}
	}

	antecedent := &lg.And{Terms: conjuncts}
	consequent := &lg.Eq{T1: x, T2: y}
	return &lg.Implies{T1: antecedent, T2: consequent}
}

func varName(idx int) string {
	return fmt.Sprintf("V%d", idx)
}

// PartialFunction returns a formula stating that rel is a partial function:
// forall X, Y, Z. (rel(X,Y) & rel(X,Z)) -> Y = Z
func PartialFunction(rel *lg.Const) lg.Expr {
	fs, ok := rel.CSort.(*lg.FunctionSort)
	if !ok || fs.Arity() < 2 {
		return &lg.And{} // true
	}
	dom := fs.Domain()
	x, _ := lg.NewVariable("X", dom[0])
	y, _ := lg.NewVariable("Y", dom[1])
	z, _ := lg.NewVariable("Z", dom[1])

	relXY := &lg.Apply{Func: rel, Terms: []lg.Expr{x, y}}
	relXZ := &lg.Apply{Func: rel, Terms: []lg.Expr{x, z}}
	premise := &lg.And{Terms: []lg.Expr{relXY, relXZ}}
	conclusion := &lg.Eq{T1: y, T2: z}
	body := &lg.Implies{T1: premise, T2: conclusion}
	return &lg.ForAll{Variables: []*lg.Variable{x, y, z}, Body: body}
}

// VariableUniqifier alpha-converts formulas so all bound variables are unique.
type VariableUniqifier struct {
	rn     *iu.UniqueRenamer
	InvMap map[lg.NodeKey]*lg.Variable // renamed var → original var
}

// NewVariableUniqifier creates a new uniqifier, reserving the given names.
func NewVariableUniqifier(used []string) *VariableUniqifier {
	return &VariableUniqifier{
		rn:     iu.NewUniqueRenamer("", used),
		InvMap: make(map[lg.NodeKey]*lg.Variable),
	}
}

// Uniquify alpha-converts a formula so all bound variables have unique names.
func (vu *VariableUniqifier) Uniquify(fmla lg.Expr) lg.Expr {
	vmap := make(map[lg.NodeKey]*lg.Variable)
	return vu.rec(fmla, vmap)
}

func (vu *VariableUniqifier) rec(fmla lg.Expr, vmap map[lg.NodeKey]*lg.Variable) lg.Expr {
	if IsBinder(fmla) {
		vars := BinderVars(fmla)
		body := BinderBody(fmla)

		// Save old bindings
		type saved struct {
			v   *lg.Variable
			old *lg.Variable
		}
		var obs []saved
		for _, v := range vars {
			if old, ok := vmap[lg.Key(v)]; ok {
				obs = append(obs, saved{v, old})
			}
		}

		// Create new variable names
		newVars := make([]*lg.Variable, len(vars))
		for i, v := range vars {
			newName := vu.rn.Rename(v.Name)
			nv, _ := lg.NewVariable(newName, v.VSort)
			newVars[i] = nv
			vmap[lg.Key(v)] = nv
			vu.InvMap[lg.Key(nv)] = v
		}

		newBody := vu.rec(body, vmap)
		result := CloneBinder(fmla, newVars, newBody)

		// Restore old bindings
		for _, v := range vars {
			delete(vmap, lg.Key(v))
		}
		for _, o := range obs {
			vmap[lg.Key(o.v)] = o.old
		}

		return result
	}

	if v, ok := fmla.(*lg.Variable); ok {
		if mapped, exists := vmap[lg.Key(v)]; exists {
			return mapped
		}
		// Free variable — assign a new unique name
		newName := vu.rn.Rename(v.Name)
		nv, _ := lg.NewVariable(newName, v.VSort)
		vmap[lg.Key(v)] = nv
		vu.InvMap[lg.Key(nv)] = v
		return nv
	}

	args := NodeArgs(fmla)
	if len(args) == 0 {
		return fmla
	}
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = vu.rec(a, vmap)
	}
	return CloneNode(fmla, newArgs)
}

// Undo reverses the renaming applied by this uniqifier.
func (vu *VariableUniqifier) Undo(fmla lg.Expr) lg.Expr {
	subs := make(map[lg.NodeKey]lg.Expr)
	for k, v := range vu.InvMap {
		subs[k] = v
	}
	result, _ := lu.Substitute(fmla, subs)
	return result
}

// AlphaAvoid alpha-converts a formula so that bound variable names do not
// clash with the given set of variables.
func AlphaAvoid(fmla lg.Expr, vs []*lg.Variable) lg.Expr {
	vu := NewVariableUniqifier(nil)
	// Reserve names of vs and free variables
	for _, v := range vs {
		vu.rn.Rename(v.Name)
	}
	fvs := lu.FreeVariablesList(fmla)
	vmap := make(map[lg.NodeKey]*lg.Variable)
	for _, v := range fvs {
		vu.rn.Rename(v.Name)
		vmap[lg.Key(v)] = v // preserve free variable
	}
	for _, v := range vs {
		vmap[lg.Key(v)] = v
	}
	return vu.rec(fmla, vmap)
}

// NormalizeOps converts conjunctions and disjunctions to binary ops
// and quantifiers to single-variable quantifiers.
func NormalizeOps(fmla lg.Expr) lg.Expr {
	args := NodeArgs(fmla)
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = NormalizeOps(a)
	}

	switch fmla.(type) {
	case *lg.And, *lg.Or:
		if len(newArgs) == 0 {
			return CloneNode(fmla, nil)
		}
		return makeBin(fmla, newArgs[0], newArgs[1:])
	case *lg.ForAll:
		vars := BinderVars(fmla)
		return makeQuant(fmla, vars, newArgs[0])
	case *lg.Exists:
		vars := BinderVars(fmla)
		return makeQuant(fmla, vars, newArgs[0])
	}

	return CloneNode(fmla, newArgs)
}

func makeBin(proto lg.Expr, first lg.Expr, rest []lg.Expr) lg.Expr {
	if len(rest) == 0 {
		return first
	}
	combined := CloneNode(proto, []lg.Expr{first, rest[0]})
	// For And/Or with more than 2, we need binary nesting
	if len(rest) > 1 {
		return makeBin(proto, combined, rest[1:])
	}
	return combined
}

func makeQuant(proto lg.Expr, vars []*lg.Variable, body lg.Expr) lg.Expr {
	if len(vars) == 0 {
		return body
	}
	inner := makeQuant(proto, vars[1:], body)
	return CloneBinder(proto, vars[0:1], inner)
}

// ASTMatch performs structural matching of x against pattern y.
// Placeholders in y can match any subterm; successful matches are
// recorded in subst.
func ASTMatch(x, y lg.Expr, placeholders map[lg.NodeKey]lg.Expr, subst map[lg.NodeKey]lg.Expr) bool {
	// Type must match
	if typeTag(x) != typeTag(y) {
		return false
	}

	// Variable or constant placeholder
	if IsVariable(y) || IsConstant(y) {
		if placeholders != nil && placeholders[lg.Key(y)] != nil {
			if prev, ok := subst[lg.Key(y)]; ok {
				return x.Equal(prev)
			}
			subst[lg.Key(y)] = x
			return true
		}
		return x.Equal(y)
	}

	// Apply: check func and args
	if appX, ok := x.(*lg.Apply); ok {
		appY := y.(*lg.Apply)
		if !appX.Func.Equal(appY.Func) {
			return false
		}
		return astMatchLists(NodeArgs(x), NodeArgs(y), placeholders, subst)
	}

	// Literal
	if litX, ok := x.(*Literal); ok {
		litY := y.(*Literal)
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

func astMatchLists(xs, ys []lg.Expr, placeholders map[lg.NodeKey]lg.Expr, subst map[lg.NodeKey]lg.Expr) bool {
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

func typeTag(n lg.Expr) string {
	switch n.(type) {
	case *lg.Variable:
		return "Var"
	case *lg.Const:
		return "Const"
	case *lg.Apply:
		return "Apply"
	case *lg.Eq:
		return "Eq"
	case *lg.Not:
		return "Not"
	case *lg.And:
		return "And"
	case *lg.Or:
		return "Or"
	case *lg.Implies:
		return "Implies"
	case *lg.Iff:
		return "Iff"
	case *lg.Ite:
		return "Ite"
	case *lg.ForAll:
		return "ForAll"
	case *lg.Exists:
		return "Exists"
	case *lg.Lambda:
		return "Lambda"
	case *lg.NamedBinder:
		return "NamedBinder"
	case *lg.Globally:
		return "Globally"
	case *lg.Eventually:
		return "Eventually"
	case *lg.WhenOperator:
		return "WhenOperator"
	case *lg.Cond:
		return "Cond"
	case *Definition:
		return "Definition"
	case *Literal:
		return "Literal"
	case *Some:
		return "Some"
	default:
		return "unknown"
	}
}

// LabelTemporal labels temporal operators with a given label string.
func LabelTemporal(fmla lg.Expr, label string) lg.Expr {
	switch t := fmla.(type) {
	case *lg.Globally:
		return &lg.Globally{Environ: &label, Body: LabelTemporal(t.Body, label)}
	case *lg.Eventually:
		return &lg.Eventually{Environ: &label, Body: LabelTemporal(t.Body, label)}
	case *lg.WhenOperator:
		return &lg.WhenOperator{
			WSort: t.WSort,
			Name:  t.Name,
			T1:    LabelTemporal(t.T1, label),
			T2:    LabelTemporal(t.T2, label),
		}
	case *lg.NamedBinder:
		return &lg.NamedBinder{
			Name:      t.Name,
			Variables: t.Variables,
			Environ:   &label,
			Body:      LabelTemporal(t.Body, label),
		}
	case *lg.Apply:
		newFunc := LabelTemporal(t.Func, label)
		newArgs := make([]lg.Expr, len(t.Terms))
		for i, a := range t.Terms {
			newArgs[i] = LabelTemporal(a, label)
		}
		return &lg.Apply{Func: newFunc, Terms: newArgs}
	}
	args := NodeArgs(fmla)
	newArgs := make([]lg.Expr, len(args))
	for i, a := range args {
		newArgs[i] = LabelTemporal(a, label)
	}
	return CloneNode(fmla, newArgs)
}
