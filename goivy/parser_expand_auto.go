package goivy

// expandAutoInstances implements Python's expand_autoinstances.
// Matches parser.py expand_autoinstances behavior (ivy_parser.py:3615-3648):
// AutoInstanceDecl nodes are collected and removed from the decl list.
// When a type referenced by an autoinstance is encountered in another
// declaration, the autoinstance is expanded inline via do_insts.

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// TypeNames matches Python's TypeNames class (ivy_parser.py:3601-3613).
// It collects type names referenced by declarations, recursively extracting
// parameter names from parameterized type references.
type TypeNames struct {
	namelist []string
	nameset  map[string]bool
}

func newTypeNames() *TypeNames {
	xtracer.Trace("parser.TypeNames.__init__ ENTER")
	return &TypeNames{nameset: make(map[string]bool)}
}

// add matches Python TypeNames.add (ivy_parser.py:3606-3613).
// It recursively extracts parameter names from parameterized type references.
func (tn *TypeNames) add(tname string) {
	xtracer.Trace("parser.TypeNames.add ENTER tname=%s", tname)
	if tn.nameset[tname] {
		return
	}
	_, refparms := extractParametersName(tname)
	for _, rp := range refparms {
		tn.add(rp)
	}
	tn.namelist = append(tn.namelist, tname)
	tn.nameset[tname] = true
}

type autoKey struct {
	prefix  string
	nparams int
}

// expandAutoInstances implements Python expand_autoinstances (ivy_parser.py:3615-3648).
// It operates on the ivyAccum in-place, matching Python's mutation of ivy.decls.
func expandAutoInstances(ivy *ivyAccum) {
	xtracer.Trace("parser.expand_autoinstances ENTER")
	autos := make(map[autoKey][]*Instantiation)
	trefs := make(map[string]bool)
	decls := ivy.decls
	xtracer.Trace("parser.expand_auto ENTER decls=%d", len(decls))
	cfg := ivy.astCfg

	// Python: ivy.decls = []
	ivy.decls = nil

	for idx, decl := range decls {
		xtracer.Trace("parser.expand_auto.decl ENTER i=%d decl=%s", idx, parserTraceNodeType(decl))
		if aid, ok := decl.(*AutoInstanceDecl); ok {
			// Python: for inst in decl.args:
			//             if len(inst.args) == 2:
			for _, arg := range aid.Args() {
				inst, ok := arg.(*Instantiation)
				if !ok || inst == nil {
					continue
				}
				// Python: if len(inst.args) == 2: — inst has Name and Sort
				if inst.Name != nil && inst.Sort != nil {
					var nameStr string
					if a, ok := inst.Name.(*Atom); ok {
						nameStr = a.Rep
					}
					pref, parms := extractParametersName(nameStr)
					key := autoKey{pref, len(parms)}
					autos[key] = append(autos[key], inst)
				}
			}
		} else {
			// Python: drefs = TypeNames()
			//         decl.get_type_names(drefs)
			drefs := newTypeNames()
			getTypeNamesFromDecl(decl, drefs)

			for _, tname := range drefs.namelist {
				if trefs[tname] {
					continue
				}
				trefs[tname] = true
				pref, refparms := extractParametersName(tname)
				key := autoKey{pref, len(refparms)}
				// Match Python defaultdict auto-vivification
				if _, ok := autos[key]; !ok {
					autos[key] = nil
				}
				for _, inst := range autos[key] {
					var instNameStr string
					if a, ok := inst.Name.(*Atom); ok {
						instNameStr = a.Rep
					}
					_, parms := extractParametersName(instNameStr)

					// Python: subst = dict(list(zip(parms, refparms)))
					subst := make(map[string]string)
					for i := 0; i < len(parms) && i < len(refparms); i++ {
						subst[parms[i]] = refparms[i]
					}

					// Python: lhs = Atom(tname, [])
					lhs := cfg.NewAtom(tname)

					// Python: rhs = inst.args[1].clone([Atom(subst.get(a.rep,a.rep),[]) for a in inst.args[1].args])
					var rhsArgs []Node
					for _, a := range inst.Sort.Args() {
						rep := nodeRep(a)
						if repl, ok := subst[rep]; ok {
							rep = repl
						}
						rhsArgs = append(rhsArgs, cfg.NewAtom(rep))
					}
					rhs := inst.Sort.Clone(rhsArgs)

					// Python: newinst = LogicInstantiation(lhs, rhs)
					newInst := cfg.NewInstantiation(lhs, rhs)

					// Python: if hasattr(decl,"lineno"): newinst.lineno = decl.lineno
					newInst.SetLineno(decl.GetLineno())

					// Python: do_insts(ivy, [newinst])
					doInsts(ivy, []Node{newInst})
				}
			}
			// Python: ivy.decls.append(decl)
			ivy.decls = append(ivy.decls, decl)
		}
	}

	result := ivy.decls
	xtracer.Trace("parser.expand_auto EXIT decls=%d", len(result))
}

func parserTraceNodeType(n Node) string {
	if n == nil {
		return "None"
	}
	typ := reflect.TypeOf(n)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ.Name()
}

// extractParametersName splits "foo(X,Y)" into ("foo", ["X","Y"]).
// Matches Python iu.extract_parameters_name.
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

// nodeSort extracts the sort string from an AST node, matching Python's
// hasattr(c, 'sort') check. Returns "" if no sort is set.
func nodeSort(n Node) string {
	switch x := n.(type) {
	case *Symbol:
		if x.Sort != nil {
			return fmt.Sprint(x.Sort)
		}
	case *Variable:
		return x.VSort
	case *Atom:
		if x.ASort != nil {
			return fmt.Sprint(x.ASort)
		}
	case *App:
		if x.ASort != nil {
			return fmt.Sprint(x.ASort)
		}
	}
	return ""
}

// ttermTypeNames matches Python tterm_type_names (ivy_ast.py:983-988).
// Extracts sort type names from a typed term and its args.
func ttermTypeNames(c Node, names *TypeNames) {
	// Python: if hasattr(c,'sort'): names.add(c.sort)
	if sortStr := nodeSort(c); sortStr != "" {
		names.add(sortStr)
	}
	// Python: for arg in c.args: if hasattr(arg,'sort'): names.add(arg.sort)
	for _, arg := range c.Args() {
		if sortStr := nodeSort(arg); sortStr != "" {
			names.add(sortStr)
		}
	}
}

// getTypeNamesFromDecl dispatches to the appropriate type-specific
// get_type_names method. Matches Python's polymorphic decl.get_type_names(names).
//
// Python classes with get_type_names overrides:
//   - Decl base: no-op (ivy_ast.py:587)
//   - ConstantDecl: tterm_type_names for each arg (ivy_ast.py:996)
//   - ParameterDecl: tterm_type_names(self.mysym(), names) (ivy_ast.py:1009)
//   - ActionDecl (FunctionDecl): formal_params + formal_returns + body (ivy_ast.py:1067)
//   - TypeDecl: if StructSort, tterm_type_names on fields (ivy_ast.py:1095)
//   - AliasDecl: names.add(c.args[1].rep) (ivy_ast.py:1244)
//   - Action base: iter subactions for LocalAction (ivy_actions.py:293)
func getTypeNamesFromDecl(decl Node, names *TypeNames) {
	switch d := decl.(type) {
	case *ConstantDecl:
		// Python ConstantDecl.get_type_names (ivy_ast.py:996-998):
		//     for c in self.args: tterm_type_names(c, names)
		getTypeNamesFromConstantArgs(d.Args(), names)
	case *FreshConstantDecl:
		// Python FreshConstantDecl inherits ConstantDecl.get_type_names.
		getTypeNamesFromConstantArgs(d.Args(), names)
	case *DestructorDecl:
		// Python DestructorDecl inherits ConstantDecl.get_type_names.
		getTypeNamesFromConstantArgs(d.Args(), names)
	case *ConstructorDecl:
		// Python ConstructorDecl inherits ConstantDecl.get_type_names.
		getTypeNamesFromConstantArgs(d.Args(), names)
	case *ParameterDecl:
		// Python ParameterDecl.get_type_names (ivy_ast.py:1009-1010):
		//     tterm_type_names(self.mysym(), names)
		// mysym = self.args[0].args[0] if isinstance(self.args[0], Definition) else self.args[0]
		args := d.Args()
		if len(args) > 0 {
			mysym := args[0]
			if defn, ok := mysym.(*Definition); ok {
				if defnArgs := defn.Args(); len(defnArgs) > 0 {
					mysym = defnArgs[0]
				}
			}
			ttermTypeNames(mysym, names)
		}
	case *TypeDecl:
		// Python TypeDecl.get_type_names (ivy_ast.py:1095-1100):
		//     for c in self.args:
		//         t = c.args[1]
		//         if isinstance(t, StructSort):
		//             for s in t.args: tterm_type_names(s, names)
		for _, arg := range d.Args() {
			if td, ok := arg.(*TypeDef); ok && td.Value != nil {
				if ss, ok := td.Value.(*StructSort); ok {
					for _, s := range ss.Args() {
						ttermTypeNames(s, names)
					}
				}
			}
		}
	case *ActionDecl:
		// Python ActionDecl.get_type_names (ivy_ast.py:1067-1073):
		//     for c in self.args:
		//         for tt in c.formal_params: tterm_type_names(tt, names)
		//         for tt in c.formal_returns: tterm_type_names(tt, names)
		//         c.args[1].get_type_names(names)
		for _, arg := range d.Args() {
			if ad, ok := arg.(*ActionDef); ok {
				for _, fp := range ad.FormalParams {
					ttermTypeNames(fp, names)
				}
				for _, fr := range ad.FormalReturns {
					ttermTypeNames(fr, names)
				}
				// Python: c.args[1].get_type_names(names) — the body
				if ad.Body != nil {
					getTypeNamesFromAction(ad.Body, names)
				}
			}
		}
	case *AliasDecl:
		// Python AliasDecl.get_type_names (ivy_ast.py:1244-1246):
		//     for c in self.args: names.add(c.args[1].rep)
		for _, arg := range d.Args() {
			argArgs := arg.Args()
			if len(argArgs) >= 2 {
				if a, ok := argArgs[1].(*Atom); ok {
					names.add(a.Rep)
				}
			}
		}
	}
	// Default: Decl base class has no-op get_type_names (ivy_ast.py:587)
}

func getTypeNamesFromConstantArgs(args []Node, names *TypeNames) {
	for _, arg := range args {
		ttermTypeNames(arg, names)
	}
}

// getTypeNamesFromAction matches Python Action.get_type_names (ivy_actions.py:293-297):
//
//	def get_type_names(self, names):
//	    for a in self.iter_subactions():
//	        if isinstance(a, LocalAction):
//	            for c in a.args[:-1]:
//	                ivy_ast.tterm_type_names(c, names)
func getTypeNamesFromAction(action Node, names *TypeNames) {
	// Walk action tree looking for LocalAction nodes.
	// Python's iter_subactions yields self and all nested sub-actions.
	walkActions(action, func(a Node) {
		if la, ok := a.(*LocalAction); ok {
			// Python: for c in a.args[:-1] — all args except the last (the body)
			elems := la.Elems
			if len(elems) > 1 {
				for _, c := range elems[:len(elems)-1] {
					ttermTypeNames(c, names)
				}
			}
		}
	})
}

// walkActions walks an action tree, calling fn for each action node.
// Matches Python Action.iter_subactions() which yields self and recursively
// yields from each child action.
func walkActions(node Node, fn func(Node)) {
	if node == nil {
		return
	}
	fn(node)
	for _, arg := range node.Args() {
		walkActions(arg, fn)
	}
}
