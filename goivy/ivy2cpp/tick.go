package ivy2cpp

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type progressDecl struct {
	Name string
	LHS  goivy.Expr
	Cond goivy.Expr
	Vars []*goivy.LogicVariable
}

type relyDecl struct {
	Key     string
	LHS     goivy.Expr
	RHS     goivy.Expr
	Implied bool
}

func (g *Generator) progressDecls() []progressDecl {
	if g == nil || g.Mod == nil {
		return nil
	}
	var out []progressDecl
	seen := map[string]bool{}
	for _, item := range g.Mod.Progress {
		p, ok, err := g.progressDeclFrom(item)
		if err != nil {
			g.errs = append(g.errs, err)
			continue
		}
		if !ok || p.Name == "" {
			continue
		}
		key := progressKey(p.LHS)
		if key == "" {
			key = p.Name
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}

func (g *Generator) progressDeclFrom(item any) (progressDecl, bool, error) {
	expr, ok := item.(goivy.Expr)
	if !ok {
		return progressDecl{}, false, nil
	}
	def, ok := expr.(*goivy.LogicDefinition)
	if !ok {
		return progressDecl{}, false, fmt.Errorf("ivy2cpp: unsupported progress declaration %T: %s", expr, expr.String())
	}
	name := goivy.ExprName(def.Defines())
	if name == "" {
		return progressDecl{}, false, fmt.Errorf("ivy2cpp: progress declaration has no defined symbol: %s", def.String())
	}
	var vars []*goivy.LogicVariable
	if app, ok := def.Lhs.(*goivy.Apply); ok {
		for _, term := range app.Terms {
			v, err := progressTermVariable(term)
			if err != nil {
				return progressDecl{}, false, err
			}
			vars = append(vars, v)
		}
	}
	return progressDecl{Name: name, LHS: def.Lhs, Cond: def.Rhs, Vars: vars}, true, nil
}

func progressTermVariable(term goivy.Expr) (*goivy.LogicVariable, error) {
	if v, ok := term.(*goivy.LogicVariable); ok {
		return v, nil
	}
	name := goivy.ExprName(term)
	if name == "" {
		return nil, fmt.Errorf("ivy2cpp: progress argument is not a named variable: %T %s", term, term.String())
	}
	return &goivy.LogicVariable{Name: name, VSort: term.NodeSort()}, nil
}

func progressKey(expr goivy.Expr) string {
	return goivy.ExprName(expr)
}

func (g *Generator) relyDecls() []relyDecl {
	if g == nil || g.Mod == nil {
		return nil
	}
	var out []relyDecl
	for _, expr := range g.Mod.Rely {
		r, ok, err := relyDeclFrom(expr)
		if err != nil {
			g.errs = append(g.errs, err)
			continue
		}
		if ok {
			out = append(out, r)
		}
	}
	return out
}

func relyDeclFrom(expr goivy.Expr) (relyDecl, bool, error) {
	if expr == nil {
		return relyDecl{}, false, nil
	}
	if imp, ok := expr.(*goivy.LogicImplies); ok {
		key := progressKey(imp.T1)
		if key == "" {
			return relyDecl{}, false, fmt.Errorf("ivy2cpp: rely antecedent has no progress symbol: %s", expr.String())
		}
		return relyDecl{Key: key, LHS: imp.T1, RHS: imp.T2, Implied: true}, true, nil
	}
	key := progressKey(expr)
	if key == "" {
		return relyDecl{}, false, fmt.Errorf("ivy2cpp: rely declaration has no progress symbol: %s", expr.String())
	}
	return relyDecl{Key: key, LHS: expr, Implied: false}, true, nil
}

func (g *Generator) needsTickMax() bool {
	for _, r := range g.relyDecls() {
		if r.Implied {
			return true
		}
	}
	return false
}

func (g *Generator) emitProgressCounterDecls(w *cppWriter) {
	progress := g.progressDecls()
	for _, p := range progress {
		w.linef("%s;", g.progressCounterDecl(p))
	}
	if len(progress) > 0 {
		w.blank()
	}
}

func (g *Generator) progressCounterDecl(p progressDecl) string {
	if len(p.Vars) == 0 {
		return "int " + varName(p.Name)
	}
	domain := make([]goivy.Sort, len(p.Vars))
	for i, v := range p.Vars {
		domain[i] = v.VSort
	}
	st := progressCounterStorage(g, domain)
	if st.Kind == cppStorageArray {
		return fmt.Sprintf("int %s%s", varName(p.Name), cppArraySuffix(st.Dims))
	}
	return fmt.Sprintf("%s %s", st.Type, varName(p.Name))
}

func progressCounterStorage(g *Generator, domain []goivy.Sort) cppFunctionStorage {
	st := cppFunctionStorage{Kind: cppStorageScalar, Domain: domain, RangeType: "int", Type: "int"}
	if len(domain) == 0 {
		return st
	}
	dims := make([]int, len(domain))
	product := 1
	allCards := true
	allIntegerLike := true
	for i, d := range domain {
		card := cppArrayDim(g, d)
		if card <= 0 {
			allCards = false
		} else {
			dims[i] = card
			if product <= largeThresh {
				product *= card
			}
		}
		if !cppIsAnyIntegerType(g, d) {
			allIntegerLike = false
		}
	}
	if allCards && allIntegerLike && product <= largeThresh {
		st.Kind = cppStorageArray
		st.Dims = dims
		st.Type = "int" + cppArraySuffix(dims)
		return st
	}
	st.Kind = cppStorageHashThunk
	st.KeyType = cppCTupleNameWith(g, domain, "")
	st.Type = fmt.Sprintf("hash_thunk<%s,int>", st.KeyType)
	return st
}

func (g *Generator) emitTick(w *cppWriter) {
	w.open(fmt.Sprintf("void %s::__tick(int __timeout) {", g.ClassName))
	progress := g.progressDecls()
	if len(progress) == 0 {
		w.close("")
		w.blank()
		return
	}
	g.emitProgressTickUpdates(w, progress)
	g.emitProgressRelyChecks(w, progress)
	w.close("")
	w.blank()
}

func (g *Generator) emitProgressTickUpdates(w *cppWriter, progress []progressDecl) {
	for _, p := range progress {
		opened := g.openProgressLoops(w, p)
		cond, err := g.emitExpr(p.Cond)
		if err != nil {
			g.unsupported(w, "unsupported progress condition %s: %s", p.Name, err.Error())
			g.closeAssignmentLoops(w, opened)
			continue
		}
		lhs := g.progressCounterLValue(p)
		w.linef("%s = %s ? 0 : %s + 1;", lhs, cond, lhs)
		g.closeAssignmentLoops(w, opened)
	}
}

func (g *Generator) emitProgressRelyChecks(w *cppWriter, progress []progressDecl) {
	relyByKey := map[string][]relyDecl{}
	for _, r := range g.relyDecls() {
		relyByKey[r.Key] = append(relyByKey[r.Key], r)
	}
	for _, p := range progress {
		relies := relyByKey[p.Name]
		if hasBareRely(relies) {
			continue
		}
		opened := g.openProgressLoops(w, p)
		maxt := g.nextTemp("__tmp")
		w.linef("int %s;", maxt)
		w.linef("%s = 0;", maxt)
		for _, r := range relies {
			if !r.Implied {
				continue
			}
			g.emitRelyMax(w, maxt, p, r)
		}
		lhs := g.progressCounterLValue(p)
		w.linef("if (%s > __timeout)", maxt)
		w.indent++
		w.linef("%s = 0;", lhs)
		w.indent--
		w.linef("ivy_check_progress(%s, %s);", lhs, maxt)
		g.closeAssignmentLoops(w, opened)
	}
}

func hasBareRely(relies []relyDecl) bool {
	for _, r := range relies {
		if !r.Implied {
			return true
		}
	}
	return false
}

func (g *Generator) emitRelyMax(w *cppWriter, maxt string, p progressDecl, r relyDecl) {
	aliases := progressRelyAliases(p, r)
	extras := extraRelyVars(r.RHS, aliases)
	// Python (ivy_to_cpp.py:1798-1802): free variables on the rely RHS
	// that don't appear on the rely LHS are alpha-renamed by appending
	// "__" to their name, to prevent capture by the outer progress
	// loop's variables.
	renamed := make([]*goivy.LogicVariable, len(extras))
	for i, v := range extras {
		renamed[i] = &goivy.LogicVariable{Name: v.Name + "__", VSort: v.VSort}
	}
	opened := 0
	for _, rv := range renamed {
		header, err := g.loopHeaderForVar(rv)
		if err != nil {
			g.unsupported(w, "unsupported rely variable %s:%s", varName(rv.Name), err.Error())
			for i := 0; i < opened; i++ {
				w.close("")
			}
			return
		}
		w.open(header)
		opened++
	}
	// Build a single-pass substitution map (mirrors Python's
	// substitute_ast call at ivy_to_cpp.py:1804). We can't reuse
	// g.exprAliases here because emitExpr chains alias lookups
	// recursively, which would re-substitute the freshly-substituted
	// progress variable into the renamed extra (e.g. C -> D -> D__).
	subs := map[goivy.NodeKey]goivy.Expr{}
	for _, v := range goivy.FreeVariablesList(r.LHS) {
		if repl, ok := aliases[v.Name]; ok {
			subs[goivy.Key(v)] = repl
		}
	}
	for i, orig := range extras {
		subs[goivy.Key(orig)] = renamed[i]
	}
	substituted, err := goivy.Substitute(r.RHS, subs)
	if err != nil {
		g.unsupported(w, "unsupported rely substitution for %s: %s", p.Name, err.Error())
		for i := 0; i < opened; i++ {
			w.close("")
		}
		return
	}
	rhs, err := g.emitExpr(substituted)
	if err != nil {
		g.unsupported(w, "unsupported rely expression for %s: %s", p.Name, err.Error())
		for i := 0; i < opened; i++ {
			w.close("")
		}
		return
	}
	w.linef("%s = std::max(%s, %s);", maxt, maxt, rhs)
	for i := 0; i < opened; i++ {
		w.close("")
	}
}

func progressRelyAliases(p progressDecl, r relyDecl) map[string]goivy.Expr {
	aliases := map[string]goivy.Expr{}
	rvars := progressExprArgs(r.LHS)
	for i, rv := range rvars {
		if i >= len(p.Vars) {
			break
		}
		if name := goivy.ExprName(rv); name != "" {
			aliases[name] = p.Vars[i]
		}
	}
	return aliases
}

func progressExprArgs(expr goivy.Expr) []goivy.Expr {
	if app, ok := expr.(*goivy.Apply); ok {
		return app.Terms
	}
	return nil
}

func extraRelyVars(rhs goivy.Expr, aliases map[string]goivy.Expr) []*goivy.LogicVariable {
	var out []*goivy.LogicVariable
	for _, v := range goivy.FreeVariablesList(rhs) {
		if aliases[v.Name] != nil {
			continue
		}
		out = append(out, v)
	}
	return out
}

func (g *Generator) openProgressLoops(w *cppWriter, p progressDecl) int {
	opened := 0
	for _, v := range p.Vars {
		header, err := g.loopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "unsupported progress variable %s:%s", varName(v.Name), err.Error())
			for i := 0; i < opened; i++ {
				w.close("")
			}
			return 0
		}
		w.open(header)
		opened++
	}
	return opened
}

func (g *Generator) progressCounterLValue(p progressDecl) string {
	return g.progressCounterLValueWithObj(p, "")
}

func (g *Generator) progressCounterLValueWithObj(p progressDecl, obj string) string {
	name := varName(p.Name)
	if obj != "" {
		name = obj + "." + name
	}
	if len(p.Vars) == 0 {
		return name
	}
	args := make([]string, len(p.Vars))
	for i, v := range p.Vars {
		args[i] = varName(v.Name)
	}
	domain := make([]goivy.Sort, len(p.Vars))
	for i, v := range p.Vars {
		domain[i] = v.VSort
	}
	st := progressCounterStorage(g, domain)
	if st.Kind == cppStorageHashThunk && len(args) > 1 {
		return fmt.Sprintf("%s[%s(%s)]", name, cppCTupleLocalNameWith(g, domain), strings.Join(args, ", "))
	}
	return name + cppIndexSuffix(args)
}
