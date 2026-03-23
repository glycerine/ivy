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

	"github.com/glycerine/goivy/ast"
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
			ivy.declare(ast.NewObjectDecl(pref))
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

		// Get the module body's declarations
		// The body is either a Sequence (our lalr_full type) or something else
		var bodyDecls []ast.Node
		if seq, ok := moduleBody.(*ast.Sequence); ok {
			bodyDecls = seq.Stmts
		}

		// Python: module = defn.args[1]
		// inst_mod(ivy, module, pref, subst, vsubst, modname=inst.relname, ...)
		var prefAtom *ast.Atom
		if pref != nil {
			if a, ok := pref.(*ast.Atom); ok {
				prefAtom = a
			}
		}
		instMod(ivy, bodyDecls, prefAtom, subst, vsubst, modName)

		// Python: if pref is None: ivy.objects.update(module.objects)
		// (object tracking — deferred for now)
	}

	if len(others) > 0 {
		ivy.declare(ast.NewInstantiateDecl(others...))
	}
	xtracer.Trace("parser.do_insts EXIT decls=%d", len(others))
}

// instMod expands a module definition into an accumulator.
// Matches Python inst_mod(ivy, module, pref, subst, vsubst, modname, lineno)
// at ivy_parser.py:135-201.
//
// For each declaration in the module body:
//   - Apply subst_prefix_atoms_ast with the substitution and prefix
//   - Apply variable substitution if any
//   - Declare the result into the accumulator
func instMod(ivy *ivyAccum, bodyDecls []ast.Node, pref *ast.Atom, subst map[string]string, vsubst map[string]*ast.Variable, modname string) {
	xtracer.Trace("parser.inst_mod ENTER name=%s", modname)

	// Build defined names set from the module body
	// Python: module.defined — tracks which names the module defines
	defined := collectDefined(bodyDecls)

	// Build static set
	// Python: module.static + names where df is TypeDecl or DestructorDecl
	static := collectStatic(bodyDecls)

	for _, decl := range bodyDecls {
		// Python: dpref = pref.clone([]) if pref is not None and "common" in decl.attributes else pref
		// For now, use pref directly (common attribute handling deferred)
		dpref := pref

		if len(vsubst) > 0 {
			// Python path with variable substitution:
			// map1 = distinct_variable_renaming(used_variables_ast(dpref), used_variables_ast(decl))
			// vpref = substitute_ast(dpref, map1)
			// vvsubst = dict((x, map1[y.rep]) for x,y in vsubst.items())
			// idecl = spaa(decl, subst, vpref)
			// idecl = substitute_constants_ast2(idecl, vvsubst)
			//
			// For now, do the simple substitution without variable renaming
			idecl := ast.SubstPrefixAtomsAst(decl, subst, dpref, defined, static)
			vsub := make(map[string]ast.Node)
			for k, v := range vsubst {
				vsub[k] = v
			}
			idecl = ast.SubstituteConstantsAst2(idecl, vsub)
			declareInstDecl(ivy, idecl)
		} else {
			// Python: idecl = spaa(decl, subst, dpref)
			// spaa calls subst_prefix_atoms_ast
			xtracer.Trace("parser.spaa ENTER")
			idecl := ast.SubstPrefixAtomsAst(decl, subst, dpref, defined, static)
			declareInstDecl(ivy, idecl)
		}
	}

	xtracer.Trace("parser.inst_mod EXIT name=%s", modname)
}

// declareInstDecl declares an instantiated declaration, handling
// recursive InstantiateDecl expansion.
// Matches the isinstance checks in Python inst_mod (lines 189-198).
func declareInstDecl(ivy *ivyAccum, idecl ast.Node) {
	if _, ok := idecl.(*ast.ObjectDecl); ok {
		ivy.declare(idecl)
		// Python also: ivy.set_object_defined(...)
	} else if instDecl, ok := idecl.(*ast.InstantiateDecl); ok {
		// Recursive expansion
		doInsts(ivy, instDecl.Args())
	} else {
		ivy.declare(idecl)
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
