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
