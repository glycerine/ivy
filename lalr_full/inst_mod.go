package lalr_full

// Module instantiation logic.
// Cleanroom port from Python ivy_parser.py: do_insts() and inst_mod().
//
// Python source: ivy_parser.py lines 203-201.
// These functions expand `instantiate` declarations by looking up
// module definitions, building substitutions, and rewriting each
// declaration in the module body with the substitution and prefix.

import (
	"fmt"
	"os"

	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/xtracer"
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
func doInsts(ivy *ivyAccum, insts []ast.Node) {
	xtracer.Trace("parser.do_insts ENTER")
	cfg := ivy.astCfg
	var others []ast.Node

	for _, instantiation := range insts {
		inst, ok := instantiation.(*ast.Instantiation)
		if !ok {
			others = append(others, instantiation)
			continue
		}

		// Python: pref, inst = instantiation.args
		pref := inst.Name // may be nil
		modCall := inst.Sort

		// Get the module name from the call atom
		var modName string
		var actualArgs []ast.Node
		if a, ok := modCall.(*ast.Atom); ok {
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
		var formalParams []ast.Node
		var moduleBody ast.Node
		for _, arg := range defn.Args() {
			if d, ok := arg.(*ast.Definition); ok {
				if lhs, ok := d.Lhs.(*ast.Atom); ok {
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
		vsubst := make(map[string]*ast.Variable)
		for i := 0; i < len(formalParams) && i < len(actualArgs); i++ {
			formalName := nodeRep(formalParams[i])
			if formalName == "" {
				continue
			}
			if v, ok := actualArgs[i].(*ast.Variable); ok {
				vsubst[formalName] = v
			} else {
				subst[formalName] = nodeRep(actualArgs[i])
			}
		}

		// Python: module = defn.args[1] — the module body is an ivyAccum (Ivy instance)
		var modAccum *ivyAccum
		if ma, ok := moduleBody.(*ivyAccum); ok {
			modAccum = ma
		} else if seq, ok := moduleBody.(*ast.Sequence); ok {
			// Backward compat: old-style Sequence bodies (before B2 migration)
			modAccum = &ivyAccum{decls: seq.Stmts}
		} else {
			others = append(others, instantiation)
			continue
		}

		var prefAtom *ast.Atom
		if pref != nil {
			if a, ok := pref.(*ast.Atom); ok {
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
func instMod(ivy *ivyAccum, module *ivyAccum, pref *ast.Atom, subst map[string]string, vsubst map[string]*ast.Variable, modname string, lineno ...ast.Location) {
	xtracer.Trace("parser.inst_mod ENTER name=%s", modname)

	// Python line 154: set_always_clone_with_fresh_id(True)
	// Python line 218: set_always_clone_with_fresh_id(False)
	cfg := ivy.astCfg
	fmt.Fprintf(os.Stderr, "instMod SET_FRESH cfgp=%p\n", cfg)
	xtracer.Trace("ast.LF.instMod SET_FRESH cfg=global")
	cfg.SetAlwaysCloneWithFreshID(true)
	defer func() {
		fmt.Fprintf(os.Stderr, "instMod CLEAR_FRESH cfgp=%p\n", cfg)
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
	var refLineno ast.Location
	if len(lineno) > 0 {
		refLineno = lineno[0]
	}

	// Python lines 146-159: inner function spaa(decl, subst, pref)
	spaa := func(decl ast.Node, subst map[string]string, spPref *ast.Atom) ast.Node {
		xtracer.Trace("parser.spaa ENTER")
		localSubst := subst
		// Python: if modname is not None and pref is not None and isinstance(decl, ModuleDecl):
		//             subst = subst.copy()
		//             p, c = iu.parent_child_name(modname)
		//             subst[c] = pref.rep
		if modname != "" && spPref != nil {
			if _, ok := decl.(*ast.ModuleDecl); ok {
				localSubst = make(map[string]string, len(subst)+1)
				for k, v := range subst {
					localSubst[k] = v
				}
				pc := iu.ParentChildName(modname)
				c := pc[1]
				localSubst[c] = spPref.Rep
			}
		}
		// Python: if lineno is not None: set_reference_lineno(lineno)
		if refLineno != (ast.Location{}) {
			cfg.SetReferenceLineno(refLineno)
		}
		res := ast.SubstPrefixAtomsAst(decl, localSubst, spPref, defined, static)
		// Python: if lineno is not None: set_reference_lineno(None)
		if refLineno != (ast.Location{}) {
			cfg.SetReferenceLineno(ast.Location{})
		}
		return res
	}

	for _, decl := range module.decls {
		// Python line 161: dpref = pref.clone([]) if pref is not None and "common" in decl.attributes else pref
		dpref := pref
		dvsubst := vsubst
		if pref != nil && declHasCommonAttribute(decl) {
			if cloned, ok := pref.Clone(nil).(*ast.Atom); ok {
				dpref = cloned
			}
		}
		// Python line 162: dvsubst = dict() if "common" in decl.attributes else vsubst
		if declHasCommonAttribute(decl) {
			dvsubst = nil
		}

		var idecl ast.Node

		if _, ok := decl.(*ast.AttributeDecl); ok {
			// Python lines 163-171: special handling for AttributeDecl
			if len(dvsubst) > 0 {
				// Python: variable renaming path for AttributeDecl
				map1 := ast.DistinctVariableRenaming(ast.UsedVariablesAst(dpref), ast.UsedVariablesAst(decl))
				vpref := substAtomVars(dpref, map1)
				vvsubst := buildVVSubst(dvsubst, map1)
				idecl = composeAttributeDecl(cfg, decl.(*ast.AttributeDecl), vpref)
				idecl = ast.SubstituteConstantsAst(idecl, vvsubst)
			} else {
				// Python: idecl = AttributeDecl(*[x.clone([compose_atoms(dpref,x.args[0]),x.args[1]]) for x in decl.args])
				idecl = composeAttributeDecl(cfg, decl.(*ast.AttributeDecl), dpref)
			}
		} else if len(dvsubst) > 0 {
			// Python lines 172-177: variable substitution path
			map1 := ast.DistinctVariableRenaming(ast.UsedVariablesAst(dpref), ast.UsedVariablesAst(decl))
			vpref := substAtomVars(dpref, map1)
			vvsubst := buildVVSubst(dvsubst, map1)
			idecl = spaa(decl, subst, vpref)
			idecl = ast.SubstituteConstantsAst2(idecl, vvsubst)
		} else {
			// Python line 179: idecl = spaa(decl, subst, dpref)
			idecl = spaa(decl, subst, dpref)
		}

		// Python lines 180-183: common field handling
		if db := ast.GetDeclBase(decl); db != nil {
			if idb := ast.GetDeclBase(idecl); idb != nil {
				if db.Common != nil {
					commonName := ""
					if a, ok := db.Common.(*ast.Atom); ok {
						commonName = a.Rep
					}
					if pref != nil {
						if commonName == "this" {
							idb.Common = cfg.NewAtom(pref.Rep)
						} else {
							idb.Common = cfg.NewAtom(iu.ComposeNames(pref.Rep, commonName))
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
		if db := ast.GetDeclBase(decl); db != nil {
			if idb := ast.GetDeclBase(idecl); idb != nil {
				idb.Attributes = db.Attributes
			}
		}

		// Python lines 189-198: declare based on type
		if _, ok := idecl.(*ast.ObjectDecl); ok {
			ivy.declare(idecl)
			var objName string
			if len(idecl.Args()) > 0 {
				if a, ok := idecl.Args()[0].(*ast.Atom); ok {
					objName = a.Rep
				}
			}
			// Python: ivy.set_object_defined(idecl.args[0].rep, module.get_object_defined(idecl.args[0].rep))
			moduleDefined := getObjectDefined(module, objName)
			setObjectDefined(ivy, objName, moduleDefined)
		} else if instDecl, ok := idecl.(*ast.InstantiateDecl); ok {
			// Python lines 192-196: recursive expansion with attribute propagation
			oldAttrs := ivy.attributes
			if idb := ast.GetDeclBase(idecl); idb != nil {
				for _, attr := range idb.Attributes {
					if a, ok := attr.(*ast.Atom); ok {
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

// substAtomVars applies a variable renaming map to an Atom, returning the renamed Atom.
// Python: vpref = substitute_ast(dpref, map1)
func substAtomVars(pref *ast.Atom, renaming map[string]ast.Node) *ast.Atom {
	if pref == nil || len(renaming) == 0 {
		return pref
	}
	renamed := ast.SubstituteAst(pref, renaming)
	if a, ok := renamed.(*ast.Atom); ok {
		return a
	}
	return pref
}

// buildVVSubst creates the variable-variable substitution map.
// Python: vvsubst = dict((x, map1[y.rep]) for x, y in dvsubst.items())
func buildVVSubst(dvsubst map[string]*ast.Variable, map1 map[string]ast.Node) map[string]ast.Node {
	vvsubst := make(map[string]ast.Node, len(dvsubst))
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
func composeAttributeDecl(cfg *ast.AstConfig, decl *ast.AttributeDecl, pref *ast.Atom) ast.Node {
	if pref == nil {
		return decl
	}
	var newArgs []ast.Node
	for _, arg := range decl.Args() {
		argArgs := arg.Args()
		if len(argArgs) >= 2 {
			var composed ast.Node
			if a, ok := argArgs[0].(*ast.Atom); ok {
				composed = ast.ComposeAtoms(pref, a)
			} else {
				composed = argArgs[0]
			}
			newArgArgs := []ast.Node{composed, argArgs[1]}
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
	if entries, ok := ivy.defined[name]; ok && len(entries) > 0 {
		return entries[0].ObjectDefined
	}
	return nil
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
	if ivy.defined == nil {
		return
	}
	if entries, ok := ivy.defined[name]; ok {
		for i := range entries {
			entries[i].ObjectDefined = moduleDefined
		}
		ivy.defined[name] = entries
	}
}

// stackLookup searches the accumulator's modules map for a module definition.
// Matches Python stack_lookup(name) at ivy_parser.py:117-122.
// In the full Python parser, this searches a stack of Ivy objects.
// For the LALR parser, we search the single accumulator's modules.
func stackLookup(ivy *ivyAccum, name string) *ast.ModuleDecl {
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
func collectDefined(decls []ast.Node) map[string]bool {
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
func collectStatic(decls []ast.Node) map[string]bool {
	static := make(map[string]bool)
	for _, d := range decls {
		switch d.(type) {
		case *ast.TypeDecl, *ast.DestructorDecl:
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
func nodeRep(n ast.Node) string {
	switch x := n.(type) {
	case *ast.Atom:
		return x.Rep
	case *ast.Symbol:
		return x.Rep
	case *ast.Variable:
		return x.Rep
	case *ast.App:
		if x.Rep != nil {
			if s, ok := x.Rep.(*ast.Symbol); ok {
				return s.Rep
			}
		}
	}
	return fmt.Sprint(n)
}
