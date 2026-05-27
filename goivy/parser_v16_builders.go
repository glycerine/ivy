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

func parser16ActionFormula(node Node) Node {
	node = checkNonTemporal(node)
	lf, ok := node.(*LabeledFormula)
	if !ok || lf.Label != nil {
		return node
	}
	if lf.Formula != nil {
		if lf.Formula.GetLineno().Line == 0 && lf.GetLineno().Line > 0 {
			lf.Formula.SetLineno(lf.GetLineno())
		}
		return lf.Formula
	}
	return node
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
