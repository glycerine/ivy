package ivy2go

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// tick.go is the literal port of ivy2cpp/tick.go. Progress / rely
// declarations are how Ivy programs express liveness obligations:
// each `progress P(X) <-> cond` declares a counter that resets when
// `cond` holds and increments on every Tick otherwise; each
// `rely R -> bound` declares an upper bound on how long P can grow.
// At Tick time we advance counters and call ivyCheckProgress to
// panic when a bound is exceeded.

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

// progressDecls mirrors ivy2cpp/tick.go:24.
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

// progressDeclFrom mirrors ivy2cpp/tick.go:52.
func (g *Generator) progressDeclFrom(item any) (progressDecl, bool, error) {
	expr, ok := item.(goivy.Expr)
	if !ok {
		return progressDecl{}, false, nil
	}
	def, ok := expr.(*goivy.LogicDefinition)
	if !ok {
		return progressDecl{}, false, fmt.Errorf("ivy2go: unsupported progress declaration %T: %s", expr, expr.String())
	}
	name := goivy.ExprName(def.Defines())
	if name == "" {
		return progressDecl{}, false, fmt.Errorf("ivy2go: progress declaration has no defined symbol: %s", def.String())
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

// progressTermVariable mirrors ivy2cpp/tick.go:78.
func progressTermVariable(term goivy.Expr) (*goivy.LogicVariable, error) {
	if v, ok := term.(*goivy.LogicVariable); ok {
		return v, nil
	}
	name := goivy.ExprName(term)
	if name == "" {
		return nil, fmt.Errorf("ivy2go: progress argument is not a named variable: %T %s", term, term.String())
	}
	return &goivy.LogicVariable{Name: name, VSort: term.NodeSort()}, nil
}

// progressKey mirrors ivy2cpp/tick.go:89.
func progressKey(expr goivy.Expr) string {
	return goivy.ExprName(expr)
}

// relyDecls mirrors ivy2cpp/tick.go:93.
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

// relyDeclFrom mirrors ivy2cpp/tick.go:111.
func relyDeclFrom(expr goivy.Expr) (relyDecl, bool, error) {
	if expr == nil {
		return relyDecl{}, false, nil
	}
	if imp, ok := expr.(*goivy.LogicImplies); ok {
		key := progressKey(imp.T1)
		if key == "" {
			return relyDecl{}, false, fmt.Errorf("ivy2go: rely antecedent has no progress symbol: %s", expr.String())
		}
		return relyDecl{Key: key, LHS: imp.T1, RHS: imp.T2, Implied: true}, true, nil
	}
	key := progressKey(expr)
	if key == "" {
		return relyDecl{}, false, fmt.Errorf("ivy2go: rely declaration has no progress symbol: %s", expr.String())
	}
	return relyDecl{Key: key, LHS: expr, Implied: false}, true, nil
}

// needsTickMax mirrors ivy2cpp/tick.go:129.
func (g *Generator) needsTickMax() bool {
	for _, r := range g.relyDecls() {
		if r.Implied {
			return true
		}
	}
	return false
}

// hasBareRely mirrors ivy2cpp/tick.go:265.
func hasBareRely(relies []relyDecl) bool {
	for _, r := range relies {
		if !r.Implied {
			return true
		}
	}
	return false
}

// progressCounterStorage mirrors ivy2cpp/tick.go:163. Picks the Go
// storage shape for a progress counter — a scalar `int` for
// zero-arg counters, an n-D `int` array when every domain sort is
// integer-typed with a known cardinality, or a `map[K]int` when
// either condition fails.
func progressCounterStorage(g *Generator, domain []goivy.Sort) goFunctionStorage {
	st := goFunctionStorage{Kind: goStorageScalar, Domain: domain, RangeType: "int", Type: "int"}
	if len(domain) == 0 {
		return st
	}
	dims := make([]int, len(domain))
	product := 1
	allCards := true
	allIntegerLike := true
	for i, d := range domain {
		card := goSortCard(g, d)
		if card <= 0 {
			allCards = false
		} else {
			dims[i] = card
			if product <= largeThresh {
				product *= card
			}
		}
		if !goIsAnyIntegerType(g, d) {
			allIntegerLike = false
		}
	}
	if allCards && allIntegerLike && product <= largeThresh {
		st.Kind = goStorageArray
		st.Dims = dims
		st.Type = goArrayPrefix(dims) + "int"
		return st
	}
	st.Kind = goStorageHashThunk
	st.KeyType = goCTupleNameWith(g, domain)
	if len(domain) == 1 {
		st.Type = fmt.Sprintf("map[%s]int", g.goType(domain[0]))
	} else {
		st.Type = fmt.Sprintf("map[%s]int", st.KeyType)
	}
	return st
}

// progressCounterDecl mirrors ivy2cpp/tick.go:148. Emits a single
// Go struct field declaration for a progress counter.
func (g *Generator) progressCounterDecl(p progressDecl) string {
	field := goExportedName(p.Name)
	if len(p.Vars) == 0 {
		return field + " int"
	}
	domain := make([]goivy.Sort, len(p.Vars))
	for i, v := range p.Vars {
		domain[i] = v.VSort
	}
	st := progressCounterStorage(g, domain)
	return field + " " + st.Type
}

// progressCounterLValue mirrors ivy2cpp/tick.go:383. Returns the Go
// expression that addresses the counter cell for one iteration of
// the progress loops.
func (g *Generator) progressCounterLValue(p progressDecl) string {
	field := "s." + goExportedName(p.Name)
	if len(p.Vars) == 0 {
		return field
	}
	args := make([]string, len(p.Vars))
	for i, v := range p.Vars {
		args[i] = goIdent(v.Name)
	}
	domain := make([]goivy.Sort, len(p.Vars))
	for i, v := range p.Vars {
		domain[i] = v.VSort
	}
	st := progressCounterStorage(g, domain)
	switch st.Kind {
	case goStorageArray:
		return field + g.compactArrayIndexSuffix(domain, args)
	case goStorageHashThunk:
		if len(args) == 1 {
			return field + "[" + args[0] + "]"
		}
		return field + "[" + st.KeyType + "{" + strings.Join(args, ", ") + "}]"
	}
	return field
}

// progressRelyAliases mirrors ivy2cpp/tick.go:334.
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

// progressExprArgs mirrors ivy2cpp/tick.go:348.
func progressExprArgs(expr goivy.Expr) []goivy.Expr {
	if app, ok := expr.(*goivy.Apply); ok {
		return app.Terms
	}
	return nil
}

// extraRelyVars mirrors ivy2cpp/tick.go:355.
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

// emitProgressCounterDecls mirrors ivy2cpp/tick.go:138. Returns the
// per-progress-counter field declarations for inclusion in the State
// struct.
func (g *Generator) progressCounterFields() []string {
	progress := g.progressDecls()
	if len(progress) == 0 {
		return nil
	}
	out := make([]string, 0, len(progress))
	for _, p := range progress {
		out = append(out, g.progressCounterDecl(p))
	}
	return out
}

// openProgressLoops mirrors ivy2cpp/tick.go:366.
func (g *Generator) openProgressLoops(w *goWriter, p progressDecl) int {
	opened := 0
	for _, v := range p.Vars {
		header, _, err := g.loopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "unsupported progress variable %s: %s", goIdent(v.Name), err.Error())
			for i := 0; i < opened; i++ {
				w.line("}")
			}
			return 0
		}
		w.line(header)
		opened++
	}
	return opened
}

// closeProgressLoops closes `opened` `for` loops with a trailing
// "}" each, mirroring the cpp closeAssignmentLoops call sites.
func (g *Generator) closeProgressLoops(w *goWriter, opened int) {
	for i := 0; i < opened; i++ {
		w.line("}")
	}
}

// emitTick mirrors ivy2cpp/tick.go:198. Writes a `(s *State) Tick(timeout int)`
// method that updates every progress counter and, for each implied
// rely declaration, calls ivyCheckProgress to enforce the bound.
func (g *Generator) emitTick(w *goWriter) {
	progress := g.progressDecls()
	if len(progress) == 0 {
		w.linef("// Tick advances any timed / progress-tracked state. No progress decls in this module.")
		w.linef("func (s *%s) Tick(timeout int) { _ = timeout }", g.StateTypeName)
		w.blank()
		return
	}
	g.Ctx.OnceGlobals["__need_progress_check"] = true
	w.linef("// Tick advances progress counters and checks rely bounds.")
	w.linef("// Mirrors ivy2cpp/tick.go emitTick.")
	w.linef("func (s *%s) Tick(timeout int) {", g.StateTypeName)
	w.line("\t_ = timeout")
	g.emitProgressTickUpdates(w, progress)
	g.emitProgressRelyChecks(w, progress)
	w.line("}")
	w.blank()
}

// emitProgressTickUpdates mirrors ivy2cpp/tick.go:220.
func (g *Generator) emitProgressTickUpdates(w *goWriter, progress []progressDecl) {
	for _, p := range progress {
		opened := g.openProgressLoops(w, p)
		cond, err := g.emitExpr(p.Cond)
		if err != nil {
			g.unsupported(w, "unsupported progress condition %s: %s", p.Name, err.Error())
			g.closeProgressLoops(w, opened)
			continue
		}
		lhs := g.progressCounterLValue(p)
		w.linef("if %s { %s = 0 } else { %s = %s + 1 }", cond, lhs, lhs, lhs)
		g.closeProgressLoops(w, opened)
	}
}

// emitProgressRelyChecks mirrors ivy2cpp/tick.go:235.
func (g *Generator) emitProgressRelyChecks(w *goWriter, progress []progressDecl) {
	relyByKey := map[string][]relyDecl{}
	for _, r := range g.relyDecls() {
		relyByKey[r.Key] = append(relyByKey[r.Key], r)
	}
	for i, p := range progress {
		relies := relyByKey[p.Name]
		if hasBareRely(relies) {
			continue
		}
		opened := g.openProgressLoops(w, p)
		// Use a per-counter unique temp name so multiple progress
		// decls don't collide on `__maxt`.
		maxt := fmt.Sprintf("__maxt_%d", i)
		w.linef("%s := 0", maxt)
		for _, r := range relies {
			if !r.Implied {
				continue
			}
			g.emitRelyMax(w, maxt, p, r)
		}
		lhs := g.progressCounterLValue(p)
		w.linef("if %s > timeout {", maxt)
		w.linef("\t%s = 0", lhs)
		w.line("}")
		w.linef("ivyCheckProgress(%s, %s)", lhs, maxt)
		g.closeProgressLoops(w, opened)
	}
}

// emitRelyMax mirrors ivy2cpp/tick.go:274. Renames free variables
// on the rely RHS that don't appear on the LHS (alpha-rename with
// "__" suffix) to prevent capture by the outer progress loop's
// variables, then emits `if v := <rhs>; v > maxt { maxt = v }`.
// Progress-counter references in the rely RHS (e.g. `helper` when
// `helper` is the name of another progress decl) are routed through
// `s.<Counter>` via the alias map so they hit the State field.
func (g *Generator) emitRelyMax(w *goWriter, maxt string, p progressDecl, r relyDecl) {
	aliases := progressRelyAliases(p, r)
	extras := extraRelyVars(r.RHS, aliases)
	renamed := make([]*goivy.LogicVariable, len(extras))
	for i, v := range extras {
		renamed[i] = &goivy.LogicVariable{Name: v.Name + "__", VSort: v.VSort}
	}
	opened := 0
	for _, rv := range renamed {
		header, _, err := g.loopHeaderForVar(rv)
		if err != nil {
			g.unsupported(w, "unsupported rely variable %s: %s", goIdent(rv.Name), err.Error())
			g.closeProgressLoops(w, opened)
			return
		}
		w.line(header)
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
		g.closeProgressLoops(w, opened)
		return
	}
	// Substitute progress-counter Consts with uniquely-named fake
	// Consts so emitExpr emits a known token we can text-replace
	// post-emission. The lowering path strips the dot from
	// "s.<Field>" via goIdent, so we use a sentinel name and rewrite
	// the emitted string. Same trick thunk.go uses for env captures.
	progressNames := map[string]bool{}
	for _, pp := range g.progressDecls() {
		progressNames[pp.Name] = true
	}
	progressSubs := map[goivy.NodeKey]goivy.Expr{}
	progressTokens := map[string]string{} // sentinel → s.<Field>
	for _, sym := range goivy.UsedSymbolsAst(substituted).All() {
		c, ok := sym.(*goivy.Const)
		if !ok || !progressNames[c.Name] {
			continue
		}
		sentinel := "__ivy_progress_" + c.Name
		progressSubs[goivy.Key(c)] = &goivy.Const{Name: sentinel, CSort: c.CSort}
		progressTokens[goIdent(sentinel)] = "s." + goExportedName(c.Name)
	}
	if len(progressSubs) > 0 {
		if next, serr := goivy.Substitute(substituted, progressSubs); serr == nil {
			substituted = next
		}
	}
	rhs, err := g.emitExpr(substituted)
	if err != nil {
		g.unsupported(w, "unsupported rely expression for %s: %s", p.Name, err.Error())
		g.closeProgressLoops(w, opened)
		return
	}
	for token, replacement := range progressTokens {
		rhs = strings.ReplaceAll(rhs, token, replacement)
	}
	w.linef("if v := %s; v > %s { %s = v }", rhs, maxt, maxt)
	g.closeProgressLoops(w, opened)
}
