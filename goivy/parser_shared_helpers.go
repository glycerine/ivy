package goivy

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// parser17Acfg extracts the *AstConfig from the v1.7 lexer for grammar actions.
func parser17Acfg(lex parser17Lexer) *AstConfig {
	return lex.(*parser17LexAdapter).astCfg
}

// getLineno returns a Location for the current token position.
// Matches Python's get_lineno(p, n).
func getLineno(lex *parser17LexAdapter) Location {
	return Location{
		Filename: xtracer.NormalizeLine(lex.filename),
		Line:     lex.prevTok.Line,
	}
}

// normalizeFilename replaces the include directory path with <IVY_INCLUDE>
// so canonical strings can match across install locations.
func normalizeFilename(f string) string {
	stdDir := NewIvyUtilsConfig().GetStdIncludeDir()
	if stdDir == "" {
		return f
	}
	baseDir := filepath.Dir(stdDir)
	if baseDir != "" && !strings.HasSuffix(baseDir, string(filepath.Separator)) {
		baseDir += string(filepath.Separator)
	}
	if strings.HasPrefix(f, baseDir) {
		return "<IVY_INCLUDE>/" + f[len(baseDir):]
	}
	return f
}

// newLabel generates a unique label with the given prefix, matching Python newlabel().
func newLabel(cfg *AstConfig, pref string) *Atom {
	xtracer.Trace("parser.newlabel ENTER")
	cfg.LabelCounter++
	return cfg.NewAtom(fmt.Sprintf("%s%d", pref, cfg.LabelCounter))
}

// parser17AddLabel adds a label to a LabeledFormula if it does not have one.
// Matches Python addlabel() for the v1.7+ grammar.
func parser17AddLabel(cfg *AstConfig, lf *LabeledFormula, pref string) *LabeledFormula {
	xtracer.Trace("parser.addlabel ENTER")
	if lf.Label != nil {
		return lf
	}
	res := cfg.NewLabeledFormula(newLabel(cfg, pref), lf.Formula)
	if lf.HasLocSet() {
		res.SetLineno(lf.GetLineno())
	}
	return res
}

// parser17MkLF wraps a node in a LabeledFormula with no label.
func parser17MkLF(cfg *AstConfig, x Node) *LabeledFormula {
	xtracer.Trace("parser.mk_lf ENTER")
	lf := cfg.NewLabeledFormula(nil, x)
	if x != nil && x.GetLineno().Line > 0 {
		lf.SetLineno(x.GetLineno())
	}
	return lf
}

// checkNonTemporal validates that a formula does not contain temporal operators.
func checkNonTemporal(x Node) Node {
	xtracer.Trace("parser.check_non_temporal ENTER")
	if lf, ok := x.(*LabeledFormula); ok {
		checkNonTemporal(lf.Formula)
		return x
	}
	if HasTemporal(x) {
		fmt.Printf("warning: non-temporal formula expected\n")
	}
	return x
}

// addUnprovable marks a LabeledFormula as unprovable if cond is non-nil.
func addUnprovable(lf *LabeledFormula, cond Node) *LabeledFormula {
	xtracer.Trace("parser.addunprovable ENTER")
	if cond != nil {
		lf.Unprovable = true
	}
	return lf
}

// addTemporal marks a LabeledFormula as temporal.
func addTemporal(lf *LabeledFormula) *LabeledFormula {
	xtracer.Trace("parser.addtemporal ENTER")
	t := true
	lf.Temporal = &t
	return lf
}

// addExplicit marks a LabeledFormula as explicit.
func addExplicit(lf *LabeledFormula) *LabeledFormula {
	xtracer.Trace("parser.addexplicit ENTER")
	lf.Explicit = true
	return lf
}

// nodeLineno extracts the normalized Location from a node.
func nodeLineno(n Node) Location {
	if n == nil {
		return Location{}
	}
	loc := n.GetLineno()
	loc.Filename = xtracer.NormalizeLine(loc.Filename)
	return loc
}

// parser17AtypeToString extracts the string sort name from an atype Node.
func parser17AtypeToString(n Node) string {
	switch v := n.(type) {
	case *Symbol:
		return v.Rep
	case *This:
		return "this"
	default:
		return fmt.Sprint(n)
	}
}

// atypeToAtom converts an atype Node into an Atom.
func atypeToAtom(cfg *AstConfig, n Node) *Atom {
	switch v := n.(type) {
	case *Symbol:
		return cfg.NewAtom(v.Rep)
	case *This:
		return cfg.NewAtom("this")
	default:
		return cfg.NewAtom(fmt.Sprint(n))
	}
}

// makeMixinName generates a unique mixin name.
func makeMixinName(cfg *AstConfig, atom *Atom, suffix string) *Atom {
	xtracer.Trace("parser.make_mixin_name ENTER")
	cfg.LabelCounter++
	rep := strings.ReplaceAll(atom.Rep, ".", "_")
	return cfg.NewAtom(fmt.Sprintf("%s[%s%d]", rep, suffix, cfg.LabelCounter))
}

// handleMixin declares a mixin.
func handleMixin(cfg *AstConfig, kind string, mixer *Atom, mixee *Atom, ivy *ivyAccum) {
	xtracer.Trace("parser.handle_mixin ENTER")
	var m Node
	switch kind {
	case "before":
		m = cfg.NewMixinBeforeDef(mixer, mixee)
	case "after":
		m = cfg.NewMixinAfterDef(mixer, mixee)
	case "implement":
		m = cfg.NewMixinImplementDef(mixer, mixee)
	default:
		m = cfg.NewMixinBeforeDef(mixer, mixee)
	}
	ivy.declare(cfg.NewMixinDecl(m))
}

// stackActionLookup searches the accumulator stack for an action definition.
func stackActionLookup(ivy *ivyAccum, name string) (Node, int) {
	xtracer.Trace("parser.stack_action_lookup ENTER")
	params := 0
	for cur := ivy; cur != nil; cur = cur.parent {
		if cur.isModule {
			break
		}
		params += len(cur.params)
		if ad, ok := cur.actions[name]; ok {
			return ad, params
		}
	}
	return nil, 0
}

// inferActionParams looks up the matching action definition and infers
// missing formal parameters and returns.
func inferActionParams(ivy *ivyAccum, actname string, formals []Node, returns []Node) ([]Node, []Node) {
	xtracer.Trace("parser.infer_action_params ENTER")
	mixee, numParams := stackActionLookup(ivy, actname)
	if mixee == nil {
		return formals, returns
	}
	mixeeDef, ok := mixee.(*ActionDef)
	if !ok {
		return formals, returns
	}
	mixeeCommon := hasAttributeStr(mixeeDef.Attributes, "common")
	ivyCommon := hasAttributeStr(ivy.attributes, "common")
	if mixeeCommon != ivyCommon {
		return formals, returns
	}
	mformals, mreturns := mixeeDef.Formals()
	start := numParams + len(formals)
	if start < len(mformals) {
		formals = append(formals, mformals[start:]...)
	}
	rstart := len(returns)
	if rstart < len(mreturns) {
		returns = append(returns, mreturns[rstart:]...)
	}
	return formals, returns
}

// handleBeforeAfter processes before/after/implement action declarations.
func handleBeforeAfter(cfg *AstConfig, kind string, atom *Atom, action Node, ivy *ivyAccum, optargs []Node, optreturns []Node) {
	xtracer.Trace("parser.handle_before_after ENTER")
	mixer := makeMixinName(cfg, atom, kind)
	optargs, optreturns = inferActionParams(ivy, atom.Rep, optargs, optreturns)
	df := cfg.NewActionDef(mixer, action, optargs, optreturns)
	df.SetLineno(atom.GetLineno())
	ivy.declare(cfg.NewActionDecl(df))
	handleMixin(cfg, kind, mixer, atom, ivy)
}

// parseNativequote parses a native code block and extracts backtick references.
func parseNativequote(cfg *AstConfig, raw string, lex *parser17LexAdapter) (string, []Node) {
	xtracer.Trace("parser.parse_nativequote ENTER")
	s := raw
	if len(s) >= 6 && strings.HasPrefix(s, "<<<") && strings.HasSuffix(s, ">>>") {
		s = s[3 : len(s)-3]
	}
	fields := strings.Split(s, "`")
	var bqs []Node
	for idx, f := range fields {
		if idx%2 == 1 {
			if f == "this" {
				bqs = append(bqs, cfg.NewAtom("this"))
			} else {
				bqs = append(bqs, cfg.NewAtom(f))
			}
		}
	}
	var parts []string
	for idx, f := range fields {
		if idx%2 == 0 {
			parts = append(parts, f)
		} else {
			parts = append(parts, fmt.Sprintf("%d", idx/2))
		}
	}
	text := strings.Join(parts, "`")
	_ = getLineno(lex)
	return text, bqs
}

// methcall matches Python methcall(lhs, rhs).
func methcall(cfg *AstConfig, lhs, rhs Node) Node {
	xtracer.Trace("parser.methcall ENTER")
	switch l := lhs.(type) {
	case *App:
		if len(l.Terms) == 0 {
			return ComposeAtomsGeneric(l, rhs)
		}
	case *Atom:
		if len(l.Terms) == 0 {
			return ComposeAtomsGeneric(l, rhs)
		}
	}
	return cfg.NewMethodCall(lhs, rhs)
}

// fixIfPart handles the `some` condition case in if/while actions.
func fixIfPart(cond Node, part Node) Node {
	xtracer.Trace("parser.fix_if_part ENTER")
	var params []Node
	switch s := cond.(type) {
	case *Some:
		params = s.Params
	case *SomeMin:
		params = s.Params
	case *SomeMax:
		params = s.Params
	}
	if params != nil {
		subst := make(map[string]string)
		for _, p := range params {
			rep := NodeRep(p)
			if len(rep) > 4 {
				subst[rep[4:]] = rep
			}
		}
		if len(subst) > 0 {
			part = SubstPrefixAtomsAst(part, subst, nil, nil, nil)
		}
	}
	return part
}

func objectArgSortName(pr Node) string {
	var sort Node
	switch x := pr.(type) {
	case *Variable:
		return x.VSort
	case *Atom:
		sort = x.ASort
	case *App:
		sort = x.ASort
	}
	switch s := sort.(type) {
	case nil:
		return ""
	case *Symbol:
		return s.Rep
	case *This:
		return s.Relname()
	default:
		return fmt.Sprint(s)
	}
}

// createObject expands an object declaration with prefix substitution.
func createObject(cfg *AstConfig, top *ivyAccum, name *Atom, objectargs []Node, module *ivyAccum, lineno Location, continuation bool) {
	xtracer.Trace("parser.create_object ENTER name=%s", name.Rep)

	var prefargs []Node
	for idx, pr := range objectargs {
		vname := fmt.Sprintf("V%d", idx)
		prefargs = append(prefargs, cfg.NewVariable(vname, objectArgSortName(pr)))
	}

	pref := cfg.NewAtom(name.Rep, prefargs...)
	pref.SetLineno(lineno)

	if !continuation {
		top.declare(cfg.NewObjectDecl(pref))
		setObjectDefined(top, name.Rep, module.defined)
	}

	vsubst := make(map[string]*Variable)
	for i, pr := range objectargs {
		if i < len(prefargs) {
			prName := nodeRep(pr)
			if v, ok := prefargs[i].(*Variable); ok {
				vsubst[prName] = v
			}
		}
	}

	instMod(top, module, pref, map[string]string{}, vsubst, "", lineno)

	xtracer.Trace("parser.create_object EXIT name=%s", name.Rep)
}

// TokenInfo carries both the string value and source location of a terminal token.
type TokenInfo struct {
	Val  string
	Line int
}

func tokLineno(lex *parser17LexAdapter, tok TokenInfo) Location {
	return Location{
		Filename: xtracer.NormalizeLine(lex.filename),
		Line:     tok.Line,
	}
}

// lowerVarStmts transforms VarAction declarations into LocalAction scopes.
func lowerVarStmts(stmts []Node) []Node {
	return LowerVarStatements(stmts)
}

// parser17MakeSequence matches Python's sequence construction for Ivy 1.7+.
func parser17MakeSequence(cfg *AstConfig, stmts []Node) Node {
	if len(stmts) == 0 {
		return cfg.NewSequence()
	}
	if len(stmts) == 1 {
		return stmts[0]
	}
	stmts = lowerVarStmts(stmts)
	return cfg.NewSequence(stmts...)
}

// parser17StmtToSeq matches Python stmt_to_seq() used by around declarations.
func parser17StmtToSeq(cfg *AstConfig, stmts []Node) Node {
	xtracer.Trace("parser.stmt_to_seq ENTER")
	stmts = lowerVarStmts(stmts)
	if len(stmts) == 1 {
		return stmts[0]
	}
	res := cfg.NewSequence(stmts...)
	if len(stmts) > 0 {
		res.SetLineno(stmts[0].GetLineno())
	}
	return res
}
