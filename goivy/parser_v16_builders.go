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
	if temporal {
		lf = addTemporal(lf)
	} else {
		checkNonTemporal(lf)
	}
	d := cfg.NewAxiomDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser16DeclareProperty(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, temporal bool, loc Location) {
	if temporal {
		lf = addTemporal(lf)
	} else {
		checkNonTemporal(lf)
	}
	d := cfg.NewPropertyDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser16DeclareConjecture(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, loc Location) {
	d := cfg.NewConjectureDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser16DeclareInit(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, loc Location) {
	d := cfg.NewInitDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}
