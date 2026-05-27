package goivy

func parser17DeclareAxiom(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, explicit bool, temporal bool, loc Location) {
	lf = parser17AddLabel(cfg, lf, "axiom")
	if explicit {
		lf = addExplicit(lf)
	}
	if temporal {
		lf = addTemporal(lf)
	} else {
		checkNonTemporal(lf)
	}
	d := cfg.NewAxiomDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser17DeclareProperty(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, explicit bool, temporal bool, loc Location, skolem Node, proof Node) {
	lf = parser17AddLabel(cfg, lf, "prop")
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
	if skolem != nil {
		top.declare(cfg.NewNamedDecl(skolem))
	}
	if proof != nil {
		top.declare(cfg.NewProofDecl(proof))
	}
}

func parser17DeclareConjecture(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, loc Location) {
	lf = parser17AddLabel(cfg, lf, "conj")
	d := cfg.NewConjectureDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
}

func parser17DeclareInvariant(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, explicit bool, loc Location, proof Node) {
	lf = parser17AddLabel(cfg, lf, "invar")
	lf.Unprovable = false
	if explicit {
		lf.Explicit = true
	}
	d := cfg.NewConjectureDecl(lf)
	d.SetLineno(loc)
	top.declare(d)
	if proof != nil {
		top.declare(cfg.NewProofDecl(proof))
	}
}

func parser17DeclareUnprovableInvariant(cfg *AstConfig, top *ivyAccum, lf *LabeledFormula, proof Node) {
	lf = parser17AddLabel(cfg, lf, "invar")
	lf.Unprovable = true
	lf.Explicit = true
	d := cfg.NewConjectureDecl(lf)
	if !lf.Unprovable || cfg.CheckUnprovable {
		top.declare(d)
		if proof != nil {
			top.declare(cfg.NewProofDecl(proof))
		}
	}
}

func parser17DeclareAction(cfg *AstConfig, top *ivyAccum, impex Node, isMethod bool, name string, formals []Node, returns []Node, body Node, fallbackLoc Location) {
	lineno := fallbackLoc
	if body != nil {
		if b, ok := body.(interface{ HasLocSet() bool }); ok && b.HasLocSet() {
			lineno = body.GetLineno()
		}
	}

	if isMethod {
		selfArg := cfg.NewApp(cfg.NewSymbol("self", nil))
		selfArg.ASort = cfg.NewThis()
		selfArg.SetLineno(lineno)
		formals = append([]Node{selfArg}, formals...)
	}

	if ca, ok := body.(*CrashAction); ok {
		thisAtom := cfg.NewAtom("this", formals...)
		thisAtom.SetLineno(lineno)
		body = ca.Clone([]Node{thisAtom})
	}

	theAtom := cfg.NewAtom(name)
	theAtom.SetLineno(lineno)
	actdef := cfg.NewActionDef(theAtom, body, formals, returns)
	actdef.SetLineno(lineno)
	decl := cfg.NewActionDecl(actdef)
	decl.SetLineno(lineno)
	top.declare(decl)

	switch impex.(type) {
	case *ExportDecl:
		d := cfg.NewExportDecl(
			cfg.NewExportDef(
				cfg.NewAtom(name),
				cfg.NewAtom(""),
			),
		)
		d.SetLineno(lineno)
		top.declare(d)
	case *ImportDecl:
		d := cfg.NewImportDecl(
			cfg.NewImportDef(
				cfg.NewAtom(name),
				cfg.NewAtom(""),
			),
		)
		d.SetLineno(lineno)
		top.declare(d)
	}
}
