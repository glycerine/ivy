package goivy

import "github.com/glycerine/ivy/goivy/xtracer"

func parserDeclareInclude(cfg *AstConfig, top *ivyAccum, current *ivyAccum, importer ImporterFunc, name string, loc Location) {
	alreadyIncluded := false
	for cur := current; cur != nil; cur = cur.parent {
		if cur.included[name] {
			alreadyIncluded = true
			break
		}
	}
	if alreadyIncluded {
		return
	}

	if top.included == nil {
		top.included = make(map[string]bool)
	}
	top.included[name] = true

	pref := cfg.NewAtom(name)
	pref.SetLineno(loc)

	xtracer.Trace("parser.include ENTER name=%s", name)
	modDeclCount := 0
	if importer != nil {
		// Python: global parent_object; parent_object = "this"
		top.pendingParentObject = "this"
		mod, err := importer(name, top)
		top.pendingParentObject = ""
		if err != nil {
			// Python: with ASTContext(pref): module = importer(p[3]) -- the
			// error aborts the parse, located at the include if it has no
			// location of its own.
			panic(&parseAbort{err: astContextError(pref, err)})
		} else if mod != nil {
			modDeclCount = len(mod.Decls)
			for _, d := range mod.Decls {
				top.declareAllowRedef(d, true)
			}
			for k, v := range mod.Included {
				top.included[k] = v
			}
			for k, v := range mod.Modules {
				top.modules[k] = v
			}
		}
	}
	xtracer.Trace("parser.include EXIT name=%s decls=%d", name, modDeclCount)
}

func parserDeclareUsing(cfg *AstConfig, top *ivyAccum, importer ImporterFunc, name string, loc Location) {
	xtracer.Trace("parser.using ENTER name=%s", name)
	modDeclCount := 0
	if importer != nil {
		mod, err := importer(name, top)
		if err != nil {
			// Python: module = importer(p[3]) -- no ASTContext; the error
			// aborts the parse as raised.
			panic(&parseAbort{err: err})
		} else if mod != nil {
			modDeclCount = len(mod.Decls)
			module := &ivyAccum{
				parent:   top,
				decls:    mod.Decls,
				astCfg:   cfg,
				modules:  mod.Modules,
				included: mod.Included,
				objects:  make(map[string]interface{}),
				defined:  mod.defined, // Python: module.defined as to_pref
			}
			if module.defined == nil {
				module.defined = make(map[string][]definedEntry)
			}
			if module.modules == nil {
				module.modules = make(map[string]*ModuleDecl)
			}
			if module.included == nil {
				module.included = make(map[string]bool)
			}
			pref := cfg.NewAtom(name)
			pref.SetLineno(loc)
			instMod(top, module, pref, map[string]string{}, nil, nil, name, loc)
			for k, v := range mod.Included {
				top.included[k] = v
			}
			for k, v := range mod.Modules {
				top.modules[k] = v
			}
		}
	}
	xtracer.Trace("parser.using EXIT name=%s decls=%d", name, modDeclCount)
}

