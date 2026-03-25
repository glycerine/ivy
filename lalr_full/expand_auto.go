package lalr_full

// expandAutoInstances implements Python's expand_autoinstances.
// Matches parser.py expand_autoinstances behavior:
// AutoInstanceDecl nodes are collected and removed from the decl list.
// When a type referenced by an autoinstance is encountered in another
// declaration, the autoinstance is expanded inline via do_insts.

import (
	"strings"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/xtracer"
)

type autoKey struct {
	prefix  string
	nparams int
}

func expandAutoInstances(ivy *ivyAccum, decls []ast.Node) []ast.Node {
	xtracer.Trace("parser.expand_auto ENTER decls=%d", len(decls))
	cfg := ivy.astCfg
	if cfg == nil {
		cfg = ast.DefaultAstConfig
	}

	autos := make(map[autoKey][]*ast.Instantiation)
	trefs := make(map[string]bool)
	var result []ast.Node

	for _, decl := range decls {
		if aid, ok := decl.(*ast.AutoInstanceDecl); ok {
			for _, arg := range aid.Args() {
				inst, ok := arg.(*ast.Instantiation)
				if !ok || inst == nil {
					continue
				}
				if inst.Name != nil {
					var nameStr string
					if a, ok := inst.Name.(*ast.Atom); ok {
						nameStr = a.Rep
					}
					pref, parms := extractParametersName(nameStr)
					key := autoKey{pref, len(parms)}
					autos[key] = append(autos[key], inst)
				}
			}
		} else {
			typeNames := getTypeNames(decl)
			for _, tname := range typeNames {
				if trefs[tname] {
					continue
				}
				trefs[tname] = true
				pref, refparms := extractParametersName(tname)
				key := autoKey{pref, len(refparms)}
				for _, inst := range autos[key] {
					var instNameStr string
					if a, ok := inst.Name.(*ast.Atom); ok {
						instNameStr = a.Rep
					}
					_, parms := extractParametersName(instNameStr)
					subst := make(map[string]string)
					for i := 0; i < len(parms) && i < len(refparms); i++ {
						subst[parms[i]] = refparms[i]
					}
					lhs := cfg.NewAtom(tname)
					var rhsArgs []ast.Node
					if sortAtom, ok := inst.Sort.(*ast.Atom); ok {
						for _, a := range sortAtom.Terms {
							if sa, ok := a.(*ast.Atom); ok {
								rep := sa.Rep
								if repl, ok := subst[rep]; ok {
									rep = repl
								}
								rhsArgs = append(rhsArgs, cfg.NewAtom(rep))
							} else {
								rhsArgs = append(rhsArgs, a)
							}
						}
					}
					var rhs ast.Node
					if sortAtom, ok := inst.Sort.(*ast.Atom); ok {
						rhs = ast.NewAtom(sortAtom.Rep, rhsArgs...)
					} else {
						rhs = inst.Sort
					}
					newInst := ast.NewInstantiation(lhs, rhs)
					// Expand via doInsts
					tempAccum := &ivyAccum{
						parent:  ivy,
						modules: ivy.modules,
						macros:  ivy.macros,
						actions: ivy.actions,
					}
					doInsts(tempAccum, []ast.Node{newInst})
					result = append(result, tempAccum.decls...)
				}
			}
			result = append(result, decl)
		}
	}

	xtracer.Trace("parser.expand_auto EXIT decls=%d", len(result))
	return result
}

// extractParametersName splits "foo(X,Y)" into ("foo", ["X","Y"]).
func extractParametersName(name string) (string, []string) {
	idx := strings.Index(name, "(")
	if idx < 0 {
		return name, nil
	}
	pref := name[:idx]
	rest := name[idx+1:]
	rest = strings.TrimSuffix(rest, ")")
	if rest == "" {
		return pref, nil
	}
	return pref, strings.Split(rest, ",")
}

// getTypeNames extracts type names referenced by a declaration.
func getTypeNames(decl ast.Node) []string {
	var names []string
	switch d := decl.(type) {
	case *ast.TypeDecl:
		for _, arg := range d.Args() {
			if td, ok := arg.(*ast.TypeDef); ok {
				if td.Value != nil {
					collectTypeRefs(td.Value, &names)
				}
			}
		}
	}
	// Also collect from defines
	if definer, ok := decl.(interface{ Defines() []string }); ok {
		names = append(names, definer.Defines()...)
	}
	return names
}

func collectTypeRefs(node ast.Node, names *[]string) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.Symbol:
		*names = append(*names, n.Rep)
	case *ast.Atom:
		*names = append(*names, n.Rep)
	}
	for _, arg := range node.Args() {
		collectTypeRefs(arg, names)
	}
}
