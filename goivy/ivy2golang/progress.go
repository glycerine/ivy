package ivy2golang

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
		return progressDecl{}, false, fmt.Errorf("ivy2golang: unsupported progress declaration %T: %s", expr, expr.String())
	}
	name := goivy.ExprName(def.Defines())
	if name == "" {
		return progressDecl{}, false, fmt.Errorf("ivy2golang: progress declaration has no defined symbol: %s", def.String())
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
		return nil, fmt.Errorf("ivy2golang: progress argument is not a named variable: %T %s", term, term.String())
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
			return relyDecl{}, false, fmt.Errorf("ivy2golang: rely antecedent has no progress symbol: %s", expr.String())
		}
		return relyDecl{Key: key, LHS: imp.T1, RHS: imp.T2, Implied: true}, true, nil
	}
	key := progressKey(expr)
	if key == "" {
		return relyDecl{}, false, fmt.Errorf("ivy2golang: rely declaration has no progress symbol: %s", expr.String())
	}
	return relyDecl{Key: key, LHS: expr, Implied: false}, true, nil
}

func (g *Generator) emitProgressCounterDecls(w *goWriter) {
	for _, p := range g.progressDecls() {
		w.linef("%s %s", goName(p.Name), g.progressCounterType(p))
	}
}

func (g *Generator) progressCounterType(p progressDecl) string {
	if len(p.Vars) == 0 {
		return "int"
	}
	domain := progressDomain(p)
	return "map[" + g.goMapKeyType(domain) + "]int"
}

func (g *Generator) emitProgressStorageInitializers(w *goWriter) {
	for _, p := range g.progressDecls() {
		if len(p.Vars) == 0 {
			continue
		}
		w.linef("ivy.%s = make(%s)", goName(p.Name), g.progressCounterType(p))
	}
}

func (g *Generator) isProgressName(name string) (progressDecl, bool) {
	for _, p := range g.progressDecls() {
		if p.Name == name {
			return p, true
		}
	}
	return progressDecl{}, false
}

func (g *Generator) progressCounterLValue(p progressDecl, obj string) string {
	name := goName(p.Name)
	if obj != "" {
		name = obj + "." + name
	}
	if len(p.Vars) == 0 {
		return name
	}
	args := make([]string, len(p.Vars))
	for i, v := range p.Vars {
		args[i] = goName(v.Name)
	}
	return name + "[" + g.goMapKeyValue(progressDomain(p), args) + "]"
}

func progressDomain(p progressDecl) []goivy.Sort {
	domain := make([]goivy.Sort, len(p.Vars))
	for i, v := range p.Vars {
		domain[i] = v.VSort
	}
	return domain
}

func (g *Generator) emitProgressTick(w *goWriter) {
	progress := g.progressDecls()
	if len(progress) == 0 {
		return
	}
	g.emitProgressTickUpdates(w, progress)
	g.emitProgressRelyChecks(w, progress)
}

func (g *Generator) emitProgressTickUpdates(w *goWriter, progress []progressDecl) {
	for _, p := range progress {
		g.withProgressLoops(w, p, func() {
			cond, err := g.emitExpr(p.Cond)
			if err != nil {
				g.unsupported(w, "unsupported progress condition %s: %s", p.Name, err.Error())
				return
			}
			lhs := g.progressCounterLValue(p, "ivy")
			w.open("if " + cond + " {")
			w.linef("%s = 0", lhs)
			w.close(" else {")
			w.indent++
			w.linef("%s = %s + 1", lhs, lhs)
			w.indent--
			w.line("}")
		})
	}
}

func (g *Generator) emitProgressRelyChecks(w *goWriter, progress []progressDecl) {
	relyByKey := map[string][]relyDecl{}
	for _, r := range g.relyDecls() {
		relyByKey[r.Key] = append(relyByKey[r.Key], r)
	}
	for _, p := range progress {
		relies := relyByKey[p.Name]
		if hasBareRely(relies) {
			continue
		}
		g.withProgressLoops(w, p, func() {
			maxt := g.nextTemp("__tmp")
			w.linef("%s := 0", maxt)
			for _, r := range relies {
				if r.Implied {
					g.emitRelyMax(w, maxt, p, r)
				}
			}
			lhs := g.progressCounterLValue(p, "ivy")
			w.open(fmt.Sprintf("if %s > timeout {", maxt))
			w.linef("%s = 0", lhs)
			w.close("")
			w.linef("ivyCheckProgress(%s, %s)", lhs, maxt)
		})
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

func (g *Generator) emitRelyMax(w *goWriter, maxt string, p progressDecl, r relyDecl) {
	aliases := progressRelyAliases(p, r)
	extras := extraRelyVars(r.RHS, aliases)
	renamed := make([]*goivy.LogicVariable, len(extras))
	for i, v := range extras {
		renamed[i] = &goivy.LogicVariable{Name: v.Name + "__", VSort: v.VSort}
	}
	g.pushScope()
	opened := 0
	for _, rv := range renamed {
		vals, ok := g.finiteValueExprs(rv.VSort)
		if !ok {
			g.unsupported(w, "unsupported rely variable %s:%s", goName(rv.Name), sortName(rv.VSort))
			g.popScope()
			return
		}
		g.addLocal(rv.Name)
		w.open(fmt.Sprintf("for _, %s := range []%s{%s} {", goName(rv.Name), g.goScalarType(rv.VSort), strings.Join(vals, ", ")))
		opened++
	}
	subs := map[goivy.NodeKey]goivy.Expr{}
	for _, v := range goivy.VariablesAstList(r.LHS) {
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
		g.popScope()
		return
	}
	rhs, err := g.emitExpr(substituted)
	if err != nil {
		g.unsupported(w, "unsupported rely expression for %s: %s", p.Name, err.Error())
		for i := 0; i < opened; i++ {
			w.close("")
		}
		g.popScope()
		return
	}
	w.open(fmt.Sprintf("if %s > %s {", rhs, maxt))
	w.linef("%s = %s", maxt, rhs)
	w.close("")
	for i := 0; i < opened; i++ {
		w.close("")
	}
	g.popScope()
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
	for _, v := range goivy.VariablesAstList(rhs) {
		if aliases[v.Name] != nil {
			continue
		}
		out = append(out, v)
	}
	return out
}

func (g *Generator) withProgressLoops(w *goWriter, p progressDecl, body func()) {
	g.pushScope()
	opened := 0
	for _, v := range p.Vars {
		vals, ok := g.finiteValueExprs(v.VSort)
		if !ok {
			g.unsupported(w, "unsupported progress variable %s:%s", goName(v.Name), sortName(v.VSort))
			g.popScope()
			return
		}
		g.addLocal(v.Name)
		w.open(fmt.Sprintf("for _, %s := range []%s{%s} {", goName(v.Name), g.goScalarType(v.VSort), strings.Join(vals, ", ")))
		opened++
	}
	body()
	for i := 0; i < opened; i++ {
		w.close("")
	}
	g.popScope()
}
