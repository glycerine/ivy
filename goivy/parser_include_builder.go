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
		mod, err := importer(name, top)
		if err != nil {
			xtracer.Trace("parser.include ERROR name=%s err=%v", name, err)
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
			xtracer.Trace("parser.using ERROR name=%s err=%v", name, err)
		} else if mod != nil {
			modDeclCount = len(mod.Decls)
			module := &ivyAccum{
				parent:   top,
				decls:    mod.Decls,
				astCfg:   cfg,
				modules:  mod.Modules,
				included: mod.Included,
				objects:  make(map[string]*ivyAccum),
			}
			if module.modules == nil {
				module.modules = make(map[string]*ModuleDecl)
			}
			if module.included == nil {
				module.included = make(map[string]bool)
			}
			pref := cfg.NewAtom(name)
			pref.SetLineno(loc)
			instMod(top, module, pref, map[string]string{}, nil, name, loc)
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
