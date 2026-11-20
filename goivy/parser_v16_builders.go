package goivy

import "github.com/glycerine/ivy/goivy/xtracer"

func parser16Acfg(lex parser16Lexer) *AstConfig {
	return lex.(*parser16LexAdapter).astCfg
}

func tok16Lineno(lex *parser16LexAdapter, tok TokenInfo) Location {
	return Location{
		Filename: xtracer.NormalizeLine(lex.filename),
		Line:     tok.Line,
	}
}

func parser16NodeAt(node Node, loc Location) Node {
	if node != nil {
		node.SetLineno(loc)
	}
	return node
}

func parser16AddLabel(lf *LabeledFormula, pref string) *LabeledFormula {
	xtracer.Trace("parser.addlabel ENTER")
	return lf
}

func parser16ActionFormula(cfg *AstConfig, node Node) Node {
	if lf, ok := node.(*LabeledFormula); ok {
		node = parser16AddLabel(lf, "asrt")
	}
	return checkNonTemporal(node)
}

func parser16DeclareType(cfg *AstConfig, top *ivyAccum, name *Atom, loc Location) {
	scnst := cfg.NewAtom(name.Rep)
	scnst.SetLineno(nodeLineno(name))
	tdfn := cfg.NewTypeDef(scnst, cfg.NewUninterpretedSortAST())
	tdfn.SetLineno(loc)
	top.declare(cfg.NewTypeDecl(tdfn))
}

func parser16DeclareAxiom(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, temporal bool, loc Location) {
	lf = parser16AddLabel(lf, "axiom")
	if temporal {
		lf = addTemporal(lf)
	} else {
		checkNonTemporal(lf)
	}
	d := cfg.NewAxiomDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser16DeclareProperty(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, temporal bool, explicit bool, loc Location) {
	lf = parser16AddLabel(lf, "prop")
	if temporal {
		lf = addTemporal(lf)
	} else {
		checkNonTemporal(lf)
	}
	if explicit {
		lf = addExplicit(lf)
	}
	d := cfg.NewPropertyDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser16DeclareConjecture(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, loc Location) {
	lf = parser16AddLabel(lf, "conj")
	d := cfg.NewConjectureDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser16DeclareInit(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, loc Location) {
	checkNonTemporal(lf)
	d := cfg.NewInitDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

// parser16Pre15 reports Python's iu.get_numeric_version() <= [1,4]: that
// grammar has 'complexact : IF fmla sequence [ELSE action]' and no
// somefmla rules (ivy_parser.py), so no somefmla trace and no fix_if_part.
func parser16Pre15(lex parser16Lexer) bool {
	v := lex.(*parser16LexAdapter).version
	return v[0] < 1 || (v[0] == 1 && v[1] <= 4)
}
