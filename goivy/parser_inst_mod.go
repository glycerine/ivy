package goivy

// Module instantiation logic.
// Cleanroom port from Python ivy_parser.py: do_insts() and inst_mod().
//
// Python source: ivy_parser.py lines 203-201.
// These functions expand `instantiate` declarations by looking up
// module definitions, building substitutions, and rewriting each
// declaration in the module body with the substitution and prefix.

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// doInsts expands instantiation declarations.
// Matches Python do_insts(ivy, insts) at ivy_parser.py:203.
//
// For each instantiation:
//   - Look up the module definition via stackLookup
//   - If found, expand it via instMod
//   - If not found, collect as unexpanded
//
// Any unexpanded instantiations are re-declared as InstantiateDecl.
func doInsts(ivy *ivyAccum, insts []Node) {
	xtracer.Trace("parser.do_insts ENTER")
	cfg := ivy.astCfg
	var others []Node

	for _, instantiation := range insts {
		inst, ok := instantiation.(*AstInstantiation)
		if !ok {
			others = append(others, instantiation)
			continue
		}

		// Python: pref, inst = instantiation.args
		pref := inst.Name // may be nil
		modCall := inst.Sort

		// Get the module name from the call atom
		var modName string
		var actualArgs []Node
		if a, ok := modCall.(*Atom); ok {
			modName = a.Rep
			actualArgs = a.Terms
		}

		// Python: defn = stack_lookup(inst.relname)
		defn := stackLookup(ivy, modName)
		if defn == nil {
			others = append(others, instantiation)
			continue
		}

		// If there's a prefix, declare an ObjectDecl for it
		// Python: if pref != None: ivy.declare(ObjectDecl(pref))
		if pref != nil {
			ivy.declare(cfg.NewObjectDecl(pref))
		}

		// Get formal parameters from the module definition
		// Python: fparams = defn.args[0].args
		// defn is a ModuleDecl whose first arg is a Definition
		// The Definition's LHS is an Atom whose Terms are the formal params
		var formalParams []Node
		var moduleBody Node
		for _, arg := range defn.Args() {
			if d, ok := arg.(*AstDefinition); ok {
				if lhs, ok := d.Lhs.(*Atom); ok {
					formalParams = lhs.Terms
				}
				moduleBody = d.Rhs
				break
			}
		}

		// Check parameter count
		if len(actualArgs) != len(formalParams) {
			fmt.Printf("warning: wrong number of arguments to module %s: expected %d, got %d\n",
				modName, len(formalParams), len(actualArgs))
			others = append(others, instantiation)
			continue
		}

		// Build substitutions
		// Python: subst = dict((x.rep,y.rep) for x,y in zip(fparams,aparams) if not isinstance(y,Variable))
		// Python: vsubst = dict((x.rep,y) for x,y in zip(fparams,aparams) if isinstance(y,Variable))
		subst := make(map[string]string)
		vsubst := make(map[string]*AstVariable)
		for i := 0; i < len(formalParams) && i < len(actualArgs); i++ {
			formalName := nodeRep(formalParams[i])
			if formalName == "" {
				continue
			}
			if v, ok := actualArgs[i].(*AstVariable); ok {
				vsubst[formalName] = v
			} else {
				subst[formalName] = nodeRep(actualArgs[i])
			}
		}

		// Python: module = defn.args[1] — the module body is an ivyAccum (Ivy instance)
		var modAccum *ivyAccum
		if ma, ok := moduleBody.(*ivyAccum); ok {
			modAccum = ma
		} else if seq, ok := moduleBody.(*AstSequence); ok {
			// Backward compat: old-style Sequence bodies (before B2 migration)
			modAccum = &ivyAccum{decls: seq.Stmts}
		} else {
			others = append(others, instantiation)
			continue
		}

		var prefAtom *Atom
		if pref != nil {
			if a, ok := pref.(*Atom); ok {
				prefAtom = a
			}
		}

		// Python: inst_mod(ivy, module, pref, subst, vsubst, modname=inst.relname, lineno=instantiation.lineno)
		instLineno := inst.GetLineno()
		instMod(ivy, modAccum, prefAtom, subst, vsubst, modName, instLineno)

		// Python: if pref is None: ivy.objects.update(module.objects)
		if pref == nil {
			if ivy.objects == nil {
				ivy.objects = make(map[string]interface{})
			}
			for k, v := range modAccum.objects {
				ivy.objects[k] = v
			}
		}
	}

	if len(others) > 0 {
		ivy.declare(cfg.NewInstantiateDecl(others...))
	}
	xtracer.Trace("parser.do_insts EXIT decls=%d", len(others))
}

// instMod expands a module definition into an accumulator.
// Matches Python inst_mod(ivy, module, pref, subst, vsubst, modname, lineno)
// at ivy_parser.py:135-201 EXACTLY.
// The module parameter is the ivyAccum that was parsed for the module body,
// matching Python where module is the Ivy class instance stored in Definition.Rhs.
func instMod(ivy *ivyAccum, module *ivyAccum, pref *Atom, subst map[string]string, vsubst map[string]*AstVariable, modname string, lineno ...Location) {
	xtracer.Trace("parser.inst_mod ENTER name=%s", modname)

	// Python line 154: set_always_clone_with_fresh_id(True)
	// Python line 218: set_always_clone_with_fresh_id(False)
	cfg := ivy.astCfg
	xtracer.Trace("ast.LF.instMod SET_FRESH cfg=global")
	cfg.SetAlwaysCloneWithFreshID(true)
	defer func() {
		xtracer.Trace("ast.LF.instMod CLEAR_FRESH cfg=global")
		cfg.SetAlwaysCloneWithFreshID(false)
	}()

	// Python line 140-141: save = ivy.attributes
	//                       ivy.attributes = tuple(x for x in ivy.attributes if x == "common")
	saveAttrs := ivy.attributes
	ivy.attributes = filterCommonAttrs(ivy.attributes)

	// Python: module.defined — tracks which names the module defines
	// Use the module's defined map directly; fall back to collecting from decls
	// if the map isn't populated (for backward compat).
	defined := collectDefined(module.decls)

	// Python lines 142-145: static = module.static.copy()
	//   for name,dfs in module.defined.items():
	//       if any((df[1] is TypeDecl) or (df[1] is DestructorDecl) for df in dfs):
	//           static.add(name)
	static := collectStatic(module.decls)

	// Extract optional lineno parameter
	var refLineno Location
	if len(lineno) > 0 {
		refLineno = lineno[0]
	}

	// Python lines 146-159: inner function spaa(decl, subst, pref)
	spaa := func(decl Node, subst map[string]string, spPref *Atom) Node {
		xtracer.Trace("parser.spaa ENTER decl=%v pref=%v", decl.Canon(), spPref.Canon())

		localSubst := subst
		// Python: if modname is not None and pref is not None and isinstance(decl, ModuleDecl):
		//             subst = subst.copy()
		//             p, c = iu.parent_child_name(modname)
		//             subst[c] = pref.rep
		if modname != "" && spPref != nil {
			if _, ok := decl.(*ModuleDecl); ok {
				xtracer.Trace("parser.spaa substituting with pref.rep='%s'", pref.Rep)
				localSubst = make(map[string]string, len(subst)+1)
				for k, v := range subst {
					localSubst[k] = v
				}
				pc := cfg.IuCfg.ParentChildName(modname)
				c := pc[1]
				localSubst[c] = spPref.Rep
			}
		}
		// Python: if lineno is not None: set_reference_lineno(lineno)
		if refLineno != (Location{}) {
			cfg.SetReferenceLineno(refLineno)
		}
		res := SubstPrefixAtomsAst(decl, localSubst, spPref, defined, static)
		// Python: if lineno is not None: set_reference_lineno(None)
		if refLineno != (Location{}) {
			cfg.SetReferenceLineno(Location{})
		}
		xtracer.Trace("parser.spaa EXIT res=%v", res.Canon())
		return res
	}

	//xtracer.Trace("parser.inst_mod.body name=%s ndecls=%d\n pref=%v", modname, len(module.decls), pref)
	xtracer.Trace("parser.inst_mod.body name=%s ndecls=%d", modname, len(module.decls))
	for di, decl := range module.decls {
		//xtracer.Trace("parser.inst_mod.iter name=%s i=%d decl=%s\n goType=%T", modname, di, ast.DeclName(decl), decl)
		xtracer.Trace("parser.inst_mod.iter name=%s i=%d decl=%s", modname, di, DeclName(decl))
		// Python line 161: dpref = pref.clone([]) if pref is not None and "common" in decl.attributes else pref
		dpref := pref
		dvsubst := vsubst
		if pref != nil && declHasCommonAttribute(decl) {
			if cloned, ok := pref.Clone(nil).(*Atom); ok {
				dpref = cloned
			}
		}
		// Python line 162: dvsubst = dict() if "common" in decl.attributes else vsubst
		if declHasCommonAttribute(decl) {
			dvsubst = nil
		}

		var idecl Node

		if _, ok := decl.(*AttributeDecl); ok {
			// Python lines 163-171: special handling for AttributeDecl
			if len(dvsubst) > 0 {
				// Python: variable renaming path for AttributeDecl
				map1 := DistinctVariableRenaming(UsedVariablesAst(dpref), UsedVariablesAst(decl))
				vpref := substAtomVars(dpref, map1)
				vvsubst := buildVVSubst(dvsubst, map1)
				idecl = composeAttributeDecl(cfg, decl.(*AttributeDecl), vpref)
				idecl = SubstituteConstantsAst(idecl, vvsubst)
			} else {
				// Python: idecl = AttributeDecl(*[x.clone([compose_atoms(dpref,x.args[0]),x.args[1]]) for x in decl.args])
				idecl = composeAttributeDecl(cfg, decl.(*AttributeDecl), dpref)
			}
		} else if len(dvsubst) > 0 {
			// Python lines 172-177: variable substitution path
			map1 := DistinctVariableRenaming(UsedVariablesAst(dpref), UsedVariablesAst(decl))
			vpref := substAtomVars(dpref, map1)
			vvsubst := buildVVSubst(dvsubst, map1)
			idecl = spaa(decl, subst, vpref)
			idecl = SubstituteConstantsAst2(idecl, vvsubst)
		} else {
			// Python line 179: idecl = spaa(decl, subst, dpref)
			idecl = spaa(decl, subst, dpref)
		}

		// Python lines 180-183: common field handling
		if db := GetDeclBase(decl); db != nil {
			if idb := GetDeclBase(idecl); idb != nil {
				if db.Common != nil {
					commonName := ""
					if a, ok := db.Common.(*Atom); ok {
						commonName = a.Rep
					}
					if pref != nil {
						if commonName == "this" {
							idb.Common = cfg.NewAtom(pref.Rep)
						} else {
							idb.Common = cfg.NewAtom(cfg.IuCfg.ComposeNames(pref.Rep, commonName))
						}
					} else {
						idb.Common = db.Common
					}
				} else {
					idb.Common = nil
				}
			}
		}

		// Python line 188: idecl.attributes = decl.attributes
		if db := GetDeclBase(decl); db != nil {
			if idb := GetDeclBase(idecl); idb != nil {
				idb.Attributes = db.Attributes
			}
		}

		// Python lines 189-198: declare based on type
		//xtracer.Trace("parser.inst_mod.declare name=%s\n goType=%T origType=%T", ast.DeclName(idecl), idecl, decl)
		xtracer.Trace("parser.inst_mod.declare name=%s", DeclName(idecl))
		if _, ok := idecl.(*ObjectDecl); ok {
			ivy.declare(idecl)
			var objName string
			if len(idecl.Args()) > 0 {
				if a, ok := idecl.Args()[0].(*Atom); ok {
					objName = a.Rep
				}
			}
			// Python: ivy.set_object_defined(idecl.args[0].rep, module.get_object_defined(idecl.args[0].rep))
			moduleDefined := getObjectDefined(module, objName)
			setObjectDefined(ivy, objName, moduleDefined)
		} else if instDecl, ok := idecl.(*InstantiateDecl); ok {
			// Python lines 192-196: recursive expansion with attribute propagation
			oldAttrs := ivy.attributes
			if idb := GetDeclBase(idecl); idb != nil {
				for _, attr := range idb.Attributes {
					if a, ok := attr.(*Atom); ok {
						ivy.attributes = append(ivy.attributes, a.Rep)
					}
				}
			}
			doInsts(ivy, instDecl.Args())
			ivy.attributes = oldAttrs
		} else {
			ivy.declare(idecl)
		}
	}

	// Python line 199: ivy.attributes = save
	ivy.attributes = saveAttrs
	xtracer.Trace("parser.inst_mod EXIT name=%s", modname)
}

// InstModSubst applies inst_mod with a substitution map and no prefix.
// Matches Python: ivy = Ivy(); inst_mod(ivy, module, None, subst, dict())
// Used by theory compilation to substitute type parameter 't' with the actual sort name.
func InstModSubst(decls []Node, subst map[string]string, cfg *AstConfig) []Node {
	// Wrap parsed decls — Python uses read_module result directly, no Ivy() call
	module := &ivyAccum{
		decls:    decls,
		astCfg:   cfg,
		modules:  make(map[string]*ModuleDecl),
		macros:   make(map[string]Node),
		actions:  make(map[string]Node),
		included: make(map[string]bool),
	}

	// Python: ivy = Ivy() — this one correctly traces parser.__init__
	ivy := newIvyAccum(nil, "")
	ivy.astCfg = cfg
	instMod(ivy, module, nil, subst, nil, "")
	return ivy.decls
}

// substAtomVars applies a variable renaming map to an Atom, returning the renamed Atom.
// Python: vpref = substitute_ast(dpref, map1)
func substAtomVars(pref *Atom, renaming map[string]Node) *Atom {
	if pref == nil || len(renaming) == 0 {
		return pref
	}
	renamed := SubstituteAst(pref, renaming)
	if a, ok := renamed.(*Atom); ok {
		return a
	}
	return pref
}

// buildVVSubst creates the variable-variable substitution map.
// Python: vvsubst = dict((x, map1[y.rep]) for x, y in dvsubst.items())
func buildVVSubst(dvsubst map[string]*AstVariable, map1 map[string]Node) map[string]Node {
	vvsubst := make(map[string]Node, len(dvsubst))
	for x, y := range dvsubst {
		if renamed, ok := map1[y.Rep]; ok {
			vvsubst[x] = renamed
		} else {
			vvsubst[x] = y
		}
	}
	return vvsubst
}

// composeAttributeDecl composes an AttributeDecl with a prefix.
// Python: AttributeDecl(*[x.clone([compose_atoms(dpref, x.args[0]), x.args[1]]) for x in decl.args])
func composeAttributeDecl(cfg *AstConfig, decl *AttributeDecl, pref *Atom) Node {
	if pref == nil {
		return decl
	}
	var newArgs []Node
	for _, arg := range decl.Args() {
		argArgs := arg.Args()
		if len(argArgs) >= 2 {
			var composed Node
			if a, ok := argArgs[0].(*Atom); ok {
				composed = ComposeAtoms(pref, a)
			} else {
				composed = argArgs[0]
			}
			newArgArgs := []Node{composed, argArgs[1]}
			if len(argArgs) > 2 {
				newArgArgs = append(newArgArgs, argArgs[2:]...)
			}
			newArgs = append(newArgs, arg.Clone(newArgArgs))
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	return cfg.NewAttributeDecl(newArgs...)
}

// getObjectDefined matches Python Ivy.get_object_defined (ivy_parser.py:375-381):
//
//	def get_object_defined(self, name):
//	    if name in self.defined:
//	        x = self.defined[name][0]
//	        if len(x) >= 3:
//	            return x[2]
//	    return None
func getObjectDefined(ivy *ivyAccum, name string) map[string][]definedEntry {
	xtracer.Trace("parser.get_object_defined ENTER")
	if ivy == nil {
		return nil
	}
	// Python: if name in self.defined: — uses `in` check, does NOT auto-vivify
	if ivy.defined == nil {
		return nil
	}
	entries, ok := ivy.defined[name]
	if !ok || len(entries) == 0 {
		return nil
	}
	return entries[0].ObjectDefined
}

// setObjectDefined matches Python Ivy.set_object_defined (ivy_parser.py:383-391):
//
//	def set_object_defined(self, name, defined):
//	    if defined is not None:
//	        defined = defaultdict(list, ((k, v.copy()) for k, v in defined.items()))
//	    if name in self.defined:
//	        self.defined[name] = [(x[0], x[1], defined) for x in self.defined[name]]
func setObjectDefined(ivy *ivyAccum, name string, moduleDefined map[string][]definedEntry) {
	xtracer.Trace("parser.set_object_defined ENTER")
	if moduleDefined != nil {
		// Deep copy: Python's defaultdict(list, ((k, v.copy()) ...))
		copied := make(map[string][]definedEntry, len(moduleDefined))
		for k, v := range moduleDefined {
			entryCopy := make([]definedEntry, len(v))
			copy(entryCopy, v)
			copied[k] = entryCopy
		}
		moduleDefined = copied
	}
	// Python: if name in self.defined: — uses `in` check, does NOT auto-vivify
	if ivy.defined == nil {
		return
	}
	entries, ok := ivy.defined[name]
	if !ok || len(entries) == 0 {
		return
	}
	for i := range entries {
		entries[i].ObjectDefined = moduleDefined
	}
	ivy.defined[name] = entries
}

// stackLookup searches the accumulator's modules map for a module definition.
// Matches Python stack_lookup(name) at ivy_parser.py:117-122.
// In the full Python parser, this searches a stack of Ivy objects.
// For the LALR parser, we search the single accumulator's modules.
func stackLookup(ivy *ivyAccum, name string) *ModuleDecl {
	xtracer.Trace("parser.stack_lookup ENTER")
	// Walk the parent chain, matching Python's stack_lookup which
	// iterates reversed(stack) to search from innermost to outermost scope.
	for cur := ivy; cur != nil; cur = cur.parent {
		if md, ok := cur.modules[name]; ok {
			return md
		}
	}
	return nil
}

// collectDefined collects names defined by a list of declarations.
// Matches Python module.defined — a defaultdict(list) mapping name → [(lineno, cls)].
// We return a simple set of names for SubstPrefixAtomsAst's toPref parameter.
func collectDefined(decls []Node) map[string]bool {
	defined := make(map[string]bool)
	for _, d := range decls {
		if definer, ok := d.(interface{ Defines() []string }); ok {
			for _, name := range definer.Defines() {
				defined[name] = true
			}
		}
	}
	return defined
}

// collectStatic collects "static" names (types and destructors).
// Matches Python module.static + the TypeDecl/DestructorDecl check in inst_mod.
func collectStatic(decls []Node) map[string]bool {
	static := make(map[string]bool)
	for _, d := range decls {
		switch d.(type) {
		case *TypeDecl, *DestructorDecl:
			if definer, ok := d.(interface{ Defines() []string }); ok {
				for _, name := range definer.Defines() {
					static[name] = true
				}
			}
		}
	}
	return static
}

// nodeRep extracts the string representation from a node.
func nodeRep(n Node) string {
	switch x := n.(type) {
	case *Atom:
		return x.Rep
	case *Symbol:
		return x.Rep
	case *AstVariable:
		return x.Rep
	case *App:
		if x.Rep != nil {
			if s, ok := x.Rep.(*Symbol); ok {
				return s.Rep
			}
		}
	}
	return fmt.Sprint(n)
}
