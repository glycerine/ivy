// ivy_compile.go implements the main ivy_compile entry point.
// This corresponds to Python's ivy_compile function in ivy_compiler.py:2190-2254.
//
// The function runs three separate declaration interpreter passes:
//  1. IvyDomainSetup - processes types, functions, relations, axioms, definitions
//  2. IvyConjectureSetup - processes conjectures and named formulas
//  3. IvyARGSetup - processes exports, delegates, actions, initializers
//
// After the three passes, it runs post-processing:
//   - create_sort_order (topological sort of types)
//   - create_constructor_schemata
//   - fix_constructors
//   - check_definitions
//   - attach_proofs
//   - check_properties
//   - create_conj_actions
//   - handle_temporals
package goivy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// OptMutax controls whether mutable-axiom checking is enabled.
// When true (non-default), axiom symbols are allowed to be modified by actions.
// Corresponds to Python's opt_mutax = iu.BooleanParameter("mutax", False).
var OptMutax = NewBooleanParameter("mutax", false)

// IvyCompile is the main compilation entry point. It takes a list of
// declarations and compiles them into the module.
//
// This corresponds to Python's ivy_compile(decls, mod, create_isolate, **kwargs).
//
// Parameters:
//   - decls: the declaration list from the parser
//   - mod: the module to compile into (if nil, uses a fresh module)
//
// The compilation proceeds in three passes:
//
//	Pass 1 (IvyDomainSetup): types, relations, constants, axioms, definitions
//	Pass 2 (IvyConjectureSetup): conjectures
//	Pass 3 (IvyARGSetup): exports, delegates, actions, initializers
//
// IvyCompile compiles declarations into the module, running all three passes.
// createIsolate controls whether create_isolate is called at the end.
// Python: ivy_compile(decls, mod=None, create_isolate=True, **kwargs)
// When called from ivy_check, pass false because check_module
// calls create_isolate separately for each isolate.
func IvyCompile(decls []Node, mod *Module, createIsolate bool) error {
	if mod == nil {
		mod = New()
	}
	xtracer.Trace("compiler.IvyCompile ENTER decls=%d", len(decls))

	// Python line 2193: check_instantiations(mod, decls)
	if err := CheckInstantiations(mod, decls); err != nil {
		return fmt.Errorf("check instantiations: %w", err)
	}

	// Python line 2194-2195: for name in decls.defined: mod.add_to_hierarchy(name)
	for _, decl := range decls {
		for _, name := range declDefines(decl) {
			mod.AddToHierarchy(name)
		}
	}

	// Process attributes from declarations
	// Python lines 2196-2200
	for _, decl := range decls {
		processAttributes(decl, mod)
	}

	// Collect action signatures for forward reference resolution.
	// Python: TopContext(collect_actions(decls.decls))
	topCtx := CollectActions(decls)

	// Create compiler with the top context
	c := NewFromModule(mod)
	c.TopCtx = topCtx

	// Create ActionsConfig sharing the same IuCfg as the AST config.
	// All action counters (LocalActionCtr, CallActionCtr, ChoiceActionCtr)
	// live on IuCfg and are shared between ast and actions packages,
	// matching Python's single globals in ivy_actions.py.
	actCfg := NewActionsConfig()
	if mod.Cfg != nil && mod.Cfg.AstCfg != nil {
		actCfg.IuCfg = mod.Cfg.AstCfg.IuCfg
	}
	c.ActCfg = actCfg
	if mod.Cfg != nil {
		mod.Cfg.ActCfg = actCfg // store for actions-package callers via mod.Cfg.ActCfg
	}

	// Pass 1: IvyDomainSetup
	// Processes: types, relations, constants, axioms, definitions, etc.
	// Dump last 5 decl types for cross-language comparison
	nDecls := len(decls)
	tailStart := nDecls - 20
	if tailStart < 0 {
		tailStart = 0
	}
	for i := tailStart; i < nDecls; i++ {
		xtracer.Trace("compiler.DeclList i=%d name=%s", i, DeclName(decls[i]))
		//pp("goType=%T", decls[i])
	}
	xtracer.Trace("compiler.DomainSetup ENTER")
	domainInterp := NewDomainSetup(c)
	if err := domainInterp.ProcessDecls(decls); err != nil {
		return fmt.Errorf("domain setup: %w", err)
	}
	// fix_constructors: ensure constructor sorts are properly set
	FixConstructors(mod)
	xtracer.Trace("compiler.DomainSetup EXIT")
	mod.CanonSnapshot("after-domain-setup")

	// Pass 2: IvyConjectureSetup
	// Processes: conjectures, named formulas
	xtracer.Trace("compiler.ConjSetup ENTER")
	conjInterp := NewConjSetup(c)
	if err := conjInterp.ProcessDecls(decls); err != nil {
		return fmt.Errorf("conjecture setup: %w", err)
	}
	xtracer.Trace("compiler.ConjSetup EXIT")
	mod.CanonSnapshot("after-conj-setup")

	// Pass 3: IvyARGSetup
	// Processes: exports, delegates, actions, initializers, progress
	xtracer.Trace("compiler.ARGSetup ENTER")
	argInterp := NewARGSetup(c)
	if err := argInterp.ProcessDecls(decls); err != nil {
		return fmt.Errorf("ARG setup: %w", err)
	}
	xtracer.Trace("compiler.ARGSetup EXIT mod.Mixins.Len=%d", mod.Mixins.Len())
	mod.CanonSnapshot("after-arg-setup")

	// Populate macros: mod.macros = decls.macros (Python ivy_compile.py:2207)
	if mod.Macros == nil {
		mod.Macros = make(map[string]*Definition)
	}
	for _, decl := range decls {
		if md, ok := decl.(*MacroDecl); ok {
			for _, arg := range md.DeclArgs {
				if defn, ok := arg.(*Definition); ok {
					mod.Macros[defn.Defines()] = defn
				}
			}
		}
	}

	// Python lines 2209-2210: remove progress symbols from sig
	// Progress properties are not state symbols — remove from sig.
	// Note: In Python, p.defines() is a latent bug — LabeledFormula has no defines() method.
	// We implement the intended semantics: extract the progress relation's symbol name.
	for _, p := range mod.Progress {
		name := progressDefinesName(p)
		if name != "" && mod.Sig != nil {
			if sym, err := mod.Sig.FindSymbol(name, false); err == nil {
				mod.Sig.RemoveSymbol(name, sym.CSort)
			}
		}
	}

	// Python line 2211: mod.type_check()
	if err := ModuleTypeCheck(mod); err != nil {
		return fmt.Errorf("type check: %w", err)
	}

	// Python lines 2213-2218: type check each action
	for name, action := range mod.Actions.All() {
		CompilerTypeCheckAction(action, mod)
		// Python lines 2216-2218: assertion checks
		if act, ok := action.(ActionsAction); ok {
			if act.GetLineno().Line == 0 {
				pp("no lineno: %s", name)
			}
			if act.GetFormalParams() == nil {
				pp("warning: action %s has no formal_params", name)
			}
		}
	}

	// Python lines 2220-2225: if not iu.version_le(iu.get_string_version(),"1.6"):
	if !VersionLE(mod.Cfg.IuCfg.GetStringVersion(), "1.6") {
		if _, ok := mod.Isolates["this"]; !ok {
			cfg := mod.Cfg.AstCfg
			isol := cfg.NewIsolateDef(
				[]Node{cfg.NewAtom("this"), cfg.NewAtom("this")},
				0,
			)
			mod.Isolates["this"] = isol
		}
	}

	// Python lines 2232-2241: find global objects and add to isolate "with" lists
	addGlobalObjectsToIsolates(mod)

	// Post-processing passes — Python lines 2244-2250
	// Order matches Python: sort_order, constructors, proofs, defs, props, conj_actions, temporals
	if err := CreateSortOrder(mod); err != nil {
		return err
	}
	if err := CreateConstructorSchemata(mod); err != nil {
		return err
	}
	if err := AttachProofs(mod); err != nil {
		return err
	}
	if err := CheckDefinitions(mod); err != nil {
		return err
	}
	if err := CheckProperties(mod); err != nil {
		return err
	}
	if err := CreateConjActions(mod); err != nil {
		return err
	}
	HandleTemporals(mod)

	// Python line 2269-2272: if create_isolate: iso.create_isolate(isolate.get(), mod)
	// When called from ivy_check with create_isolate=False, this is SKIPPED.
	// The create_isolate call mutates mod.Actions (adds ext: prefixed entries),
	// so it must NOT run when the module will be reused for per-isolate checks.
	// Python ivy_compiler.py:2569-2572 — the merge and theory update are
	// INSIDE the if create_isolate block.
	if createIsolate {
		isolateName := ""
		if mod.Cfg != nil {
			isolateName = strings.TrimSpace(mod.Cfg.Isolate)
		}
		if err := CreateIsolate(isolateName, mod); err != nil {
			pp("IvyCompile: CreateIsolate warning: %v", err)
		}

		// Python: im.module.labeled_axioms.extend(im.module.labeled_props)
		if xtracer.Enabled {
			xtracer.Trace("compiler.IvyCompile.merge_props_to_axioms n_axioms=%d n_props=%d", len(mod.LabeledAxioms), len(mod.LabeledProps))
			for i, p := range mod.LabeledProps {
				lbl := ""
				if p.Label != nil {
					lbl = fmt.Sprint(p.Label)
				}
				xtracer.Trace("compiler.IvyCompile.merge_props_to_axioms[%d] label=%s", i, lbl)
			}
		}
		mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)

		// theory_context().__enter__() calls update_theory() which builds
		// the background theory from axioms and definitions, caching the
		// result for BackgroundTheory() calls during verification.
		mod.UpdateTheory()
	}

	xtracer.Trace("compiler.IvyCompile EXIT mod.Mixins.Len=%d", mod.Mixins.Len())
	return nil
}

// declDefines returns the names defined by a declaration.
// Corresponds to Python's decl.defines() which returns a list of (name, lineno) tuples.
func declDefines(decl Node) []string {
	//type ast.DefinerSlice interface {
	//	Defines() []string
	//}
	if d, ok := decl.(DefinerSlice); ok {
		return d.Defines()
	}
	return nil
}

// addGlobalObjectsToIsolates finds objects with "global" attribute and
// adds them to the "with" lists of all isolates.
// Corresponds to Python ivy_compile.py lines 2232-2241.
func addGlobalObjectsToIsolates(mod *Module) {
	cfg := mod.Cfg.AstCfg
	var globalObjects []Node
	for name := range mod.Attributes {
		pc := mod.Cfg.IuCfg.ParentChildName(name)
		p, c := pc[0], pc[1]
		if c == "global" {
			if _, isAlias := mod.Aliases[p]; isAlias {
				continue
			}
			ppc := mod.Cfg.IuCfg.ParentChildName(p)
			pp := ppc[0]
			ppGlobal := mod.Cfg.IuCfg.ComposeNames(pp, "global")
			if pp == "this" || mod.Attributes[ppGlobal] == nil {
				globalObjects = append(globalObjects, cfg.NewAtom(p))
			}
		}
	}
	if len(globalObjects) == 0 {
		return
	}
	for _, iso := range mod.Isolates {
		iso.Elems = append(iso.Elems, globalObjects...)
		iso.WithArgs += len(globalObjects)
	}
}

// TypeCheckAction type-checks a single action.
// Corresponds to Python's type_check_action(action, mod) (ivy_compiler.py:2213-2218).
func CompilerTypeCheckAction(action interface{}, mod *Module) {
	// Type checking validates that all symbols used in the action have
	// consistent sorts. For now, this is a no-op placeholder that will
	// be filled in when the full type checker is ported.
}

// processAttributes extracts attributes from a declaration and stores them
// on the module.
// Corresponds to Python ivy_compile.py:2196-2200:
//
//	mod.attributes[compose_names(name, attribute)] =
//	    decl.common if decl.common is not None and attribute == "common" else "yes"
func processAttributes(decl Node, mod *Module) {
	type attrProvider interface {
		GetAttributes() []Node
	}
	type commonProvider interface {
		GetCommon() Node
	}
	ha, okA := decl.(attrProvider)
	if !okA {
		return
	}
	attrs := ha.GetAttributes()
	if len(attrs) == 0 {
		return
	}
	names := declDefines(decl)
	if len(names) == 0 {
		return
	}
	// Get the common value if available
	var commonVal Node
	if cp, ok := decl.(commonProvider); ok {
		commonVal = cp.GetCommon()
	}
	for _, attrNode := range attrs {
		attribute := compilerExtractSortRep(attrNode)
		if attribute == "" {
			continue
		}
		for _, name := range names {
			key := mod.Cfg.IuCfg.ComposeNames(name, attribute)
			// Python: decl.common if decl.common is not None and attribute == "common" else "yes"
			if commonVal != nil && attribute == "common" {
				mod.Attributes[key] = commonVal
			} else {
				mod.Attributes[key] = "yes"
			}
		}
	}
}

// CollectActions pre-collects all action signatures from declarations
// so that forward references resolve during compilation.
// Corresponds to Python's collect_actions (ivy_compiler.py:1565-1576).
//
// Python collects (formals, formal_returns, keypos) for each action.
// formals = a.args[0].args + a.formal_params (declared args + formal params)
// keypos = index of first KeyArg in formals
func CollectActions(decls []Node) *TopContext {
	xtracer.Trace("compiler.CollectActions ENTER decls=%d", len(decls))
	tc := &TopContext{
		Actions: make(map[string]*ActionInfo),
	}
	for _, decl := range decls {
		switch n := decl.(type) {
		case *ActionDecl:
			for _, arg := range n.DeclArgs {
				ad, isActionDef := arg.(*ActionDef)
				if !isActionDef {
					// Fallback: try Atom for simple action declarations
					if atom, ok := arg.(*Atom); ok {
						tc.Actions[atom.Rep] = &ActionInfo{}
					}
					continue
				}
				name := ""
				if atom, ok := ad.Name.(*Atom); ok {
					name = atom.Rep
				} else {
					name = fmt.Sprint(ad.Name)
				}

				// Collect formals: declared args from the name atom + formal params
				var formals []Node
				if nameAtom, ok := ad.Name.(*Atom); ok {
					for _, t := range nameAtom.Terms {
						formals = append(formals, t)
					}
				}
				formals = append(formals, ad.FormalParams...)

				// Find keypos: index of first KeyArg
				keypos := 0
				for idx, p := range formals {
					if _, isKey := p.(*KeyArg); isKey {
						keypos = idx
						break
					}
				}

				tc.Actions[name] = &ActionInfo{
					FormalAST:    formals,
					FormalRetAST: ad.FormalReturns,
					KeyPos:       keypos,
				}
			}
		case *ExportDecl:
			for _, arg := range n.DeclArgs {
				if atom, ok := arg.(*Atom); ok {
					tc.Actions["ext:"+atom.Rep] = &ActionInfo{}
				}
			}
		}
	}
	xtracer.Trace("compiler.CollectActions EXIT actions=%d", len(tc.Actions))
	return tc
}

// --- Conjecture Setup Pass ---

// ConjSetup is the second compilation pass: processes conjectures.
// Corresponds to Python's IvyConjectureSetup (ivy_compiler.py:1374-1393).
type ConjSetup struct {
	Compiler *Compiler
	lastFact *LabeledFormula
}

func NewConjSetup(c *Compiler) *ConjSetup {
	return &ConjSetup{Compiler: c}
}

func (cs *ConjSetup) ProcessDecls(decls []Node) error {
	for _, decl := range decls {
		name := DeclName(decl)
		//xtracer.Trace("compiler.IvyConjectureSetup.dispatch name=%s\n goType=%T", name, decl)
		xtracer.Trace("compiler.IvyConjectureSetup.dispatch name=%s", name)
		switch n := decl.(type) {
		case *ConjectureDecl:
			// Python: conjecture(self, ax): cax = ax.compile(); self.domain.labeled_conjs.append(cax)
			for _, arg := range n.DeclArgs {
				lf, ok := arg.(*LabeledFormula)
				if !ok {
					pp("ConjSetup: conjecture arg is not LabeledFormula: %T", arg)
					continue
				}
				compiled, err := cs.Compiler.ThingLF(lf)
				if err != nil {
					pp("ConjSetup: compiling conjecture: %v", err)
					continue
				}
				xtracer.Trace("compiler.ConjSetup.conjecture compiled")
				cs.Compiler.Module.LabeledConjs = append(cs.Compiler.Module.LabeledConjs, compiled)
				cs.lastFact = compiled
			}
		case *PropertyDecl:
			xtracer.Trace("compiler.ConjSetup.property ENTER")
			// Python: property(self, p): self.last_fact = None
			cs.lastFact = nil
		case *DefinitionDecl:
			xtracer.Trace("compiler.ConjSetup.definition ENTER")
			// Python: definition(self, p): self.last_fact = None
			cs.lastFact = nil
		case *TheoremDecl:
			xtracer.Trace("compiler.ConjSetup.theorem ENTER")
			// Python: theorem(self, sch): self.last_fact = None
			cs.lastFact = nil
		case *ProofDecl:
			xtracer.Trace("compiler.ConjSetup.proof ENTER")
			// Python: proof(self, pf):
			//   if self.last_fact is None or isinstance(pf, ivy_ast.LabeledFormula): return
			//   self.domain.proofs.append((self.last_fact, pf.compile()))
			if cs.lastFact == nil {
				continue
			}
			for _, arg := range n.DeclArgs {
				if _, isLF := arg.(*LabeledFormula); isLF {
					continue // labeled proof — skip
				}
				// Python: pf.compile() compiles as tactic AST, not as logic formula.
				// Use CompileTactic which handles tactic nodes and returns
				// unknown nodes unchanged — matching Python's default cmpl().
				compiled, err := cs.Compiler.CompileTactic(arg)
				if err != nil {
					pp("ConjSetup: compiling proof: %v", err)
					continue
				}
				cs.Compiler.Module.Proofs = append(cs.Compiler.Module.Proofs, ProofEntry{
					Formula: cs.lastFact,
					Proof:   compiled,
				})
			}
		}
	}
	return nil
}

// --- ARG Setup Pass ---

// ARGSetup is the third compilation pass: processes exports, actions, etc.
type ARGSetup struct {
	Compiler *Compiler
}

func NewARGSetup(c *Compiler) *ARGSetup {
	return &ARGSetup{Compiler: c}
}

func (as *ARGSetup) ProcessDecls(decls []Node) error {
	mod := as.Compiler.Module
	// Count decl types for debugging
	actionCount := 0
	for _, d := range decls {
		if _, ok := d.(*ActionDecl); ok {
			actionCount++
		}
	}
	xtracer.Trace("compiler.ARGSetup.ProcessDecls ENTER total_decls=%d action_decls=%d", len(decls), actionCount)
	for _, decl := range decls {
		name := DeclName(decl)
		xtracer.Trace("compiler.IvyARGSetup.dispatch name=%s", name) // \n goType=%T", name, decl)
		switch n := decl.(type) {
		case *ActionDecl:
			// Python IvyARGSetup.action (ivy_compiler.py:1414-1419):
			//   name = a.args[0].relname
			//   self.mod.actions[name] = compile_action_def(a, self.mod.sig)
			//   self.mod.public_actions.add(name)
			for _, arg := range n.DeclArgs {
				ad, ok := arg.(*ActionDef)
				if !ok {
					continue
				}
				name := ad.Defines()
				xtracer.Trace("compiler.ARGSetup.action ENTER name=%s", name)
				action, err := as.Compiler.CompileAction(ad)
				if err != nil {
					return err
				}
				// Defensive: ensure the registered action carries a source
				// Location, falling back to the ActionDef's Loc set by the
				// parser. Mirrors Python's compile_action_def assertion that
				// the compiled body has lineno after sortify.
				if action.GetLineno() == (Location{}) {
					action.SetLineno(ad.GetLineno())
				}
				AssertEveryActionHasLoc(action, "compiler.IvyARGSetup actname="+name)
				mod.SetAction(name, action)
				mod.PublicActions.Set(name, true)
				xtracer.Trace("compiler.ARGSetup.action EXIT name=%s key=%s", name, name)
			}
		case *MixinDecl:
			xtracer.Trace("compiler.ARGSetup.mixin ENTER")
			// Python IvyARGSetup.mixin (ivy_compiler.py:1422-1425):
			//   self.mod.mixins[m.args[1].relname].append(m)
			for _, arg := range n.DeclArgs {
				switch m := arg.(type) {
				case *MixinBeforeDef:
					mixee := m.Mixee()
					existing, _ := mod.Mixins.Get2(mixee)
					mod.Mixins.Set(mixee, append(existing, m))
					xtracer.Trace("compiler.ARGSetup.mixin mixee=%s mixer=%s kind=%s", mixee, m.Mixer(), "MixinBeforeDef")
				case *MixinAfterDef:
					mixee := m.Mixee()
					existing, _ := mod.Mixins.Get2(mixee)
					mod.Mixins.Set(mixee, append(existing, m))
					xtracer.Trace("compiler.ARGSetup.mixin mixee=%s mixer=%s kind=%s", mixee, m.Mixer(), "MixinAfterDef")
				case *MixinImplementDef:
					mixee := m.Mixee()
					existing, _ := mod.Mixins.Get2(mixee)
					mod.Mixins.Set(mixee, append(existing, m))
					xtracer.Trace("compiler.ARGSetup.mixin mixee=%s mixer=%s kind=%s", mixee, m.Mixer(), "MixinImplementDef")
				}
			}
		case *AssertDecl:
			xtracer.Trace("compiler.ARGSetup._assert ENTER")
			// Python IvyARGSetup._assert (ivy_compiler.py:1426-1428):
			//   self.mod.assertions.append(type(a)(a.args[0], sortify_with_inference(a.args[1])))
			for _, arg := range n.DeclArgs {
				args := arg.Args()
				if len(args) < 2 {
					continue
				}
				compiled, err := as.Compiler.SortifyWithInference(args[1])
				if err != nil {
					pp("ARGSetup: compiling assert: %v", err)
					continue
				}
				acfg := as.Compiler.Module.Cfg.AstCfg
				lf := acfg.NewLabeledFormula(nil, compiled)
				if labeled, ok := arg.(*LabeledFormula); ok {
					lf.Label = labeled.Label
				}
				mod.Assertions = append(mod.Assertions, lf)
			}
		case *IsolateDecl:
			registerIsolateDecl(n.DeclArgs, mod)
		case *IsolateObjectDecl:
			registerIsolateDecl(n.DeclArgs, mod)
		case *ExportDecl:
			// Python IvyARGSetup.export (ivy_compiler.py:1435-1437):
			//   check_is_action(self.mod, exp, exp.exported())
			//   self.mod.exports.append(exp)
			for _, arg := range n.DeclArgs {
				if expDef, ok := arg.(*ExportDef); ok {
					name := expDef.Exported()
					xtracer.Trace("compiler.ARGSetup.export ENTER name=%s", name)
					if err := CheckIsAction(mod, name); err != nil {
						xtracer.Trace("compiler.ARGSetup.export FAIL name=%s err=%v", name, err)
						return err
					}
					mod.Exports = append(mod.Exports, expDef)
					xtracer.Trace("compiler.ARGSetup.export EXIT name=%s", name)
				}
			}
		case *ImportDecl:
			xtracer.Trace("compiler.ARGSetup.import_ ENTER")
			// Python IvyARGSetup.import_ (ivy_compiler.py:1438-1440):
			//   check_is_action(self.mod, imp, imp.imported())
			//   self.mod.imports.append(imp)
			for _, arg := range n.DeclArgs {
				if impDef, ok := arg.(*ImportDef); ok {
					if atom, ok := impDef.Imported.(*Atom); ok {
						name := atom.Relname()
						if err := CheckIsAction(mod, name); err != nil {
							return err
						}
					}
				}
				mod.Imports = append(mod.Imports, arg)
			}
		case *PrivateDecl:
			xtracer.Trace("compiler.ARGSetup.private ENTER")
			// Python IvyARGSetup.private (ivy_compiler.py:1441-1442):
			//   self.mod.privates.add(pvt.privatized())
			for _, arg := range n.DeclArgs {
				if atom, ok := arg.(*Atom); ok {
					mod.Privates[atom.Relname()] = true
				}
			}
		case *DelegateDecl:
			xtracer.Trace("compiler.ARGSetup.delegate ENTER")
			// Python IvyARGSetup.delegate (ivy_compiler.py:1443-1444):
			//   self.mod.delegates.append(exp)
			for _, arg := range n.DeclArgs {
				if dd, ok := arg.(*DelegateDef); ok {
					mod.Delegates = append(mod.Delegates, dd)
				}
			}
		case *NativeDecl:
			xtracer.Trace("compiler.ARGSetup.native ENTER")
			// Python IvyARGSetup.native (ivy_compiler.py:1445-1446):
			//   self.mod.natives.append(compile_native_def(native_def))
			for _, arg := range n.DeclArgs {
				compiled, err := as.Compiler.CompileNativeDef(arg)
				if err != nil {
					pp("ARGSetup: compiling native: %v", err)
					continue
				}
				mod.Natives = append(mod.Natives, compiled)
			}
		case *AttributeDecl:
			xtracer.Trace("compiler.ARGSetup.attribute ENTER")
			// Python IvyARGSetup.attribute (ivy_compiler.py:1447-1461):
			//   self.mod.attributes[lhs.rep] = rhs
			for _, arg := range n.DeclArgs {
				if attrDef, ok := arg.(*AttributeDef); ok {
					if nameAtom, ok := attrDef.Name.(*Atom); ok {
						mod.Attributes[nameAtom.Rep] = attrDef.Value
					}
				}
			}
		case *InitDecl:
			// Matches Python IvyARGSetup.init (ivy_compiler.py:1404-1413):
			//   la = s.compile()
			//   self.mod.labeled_inits.append(la)
			//   im.module.init_cond = and_clauses(im.module.init_cond, formula_to_clauses(la.formula))
			for _, arg := range n.DeclArgs {
				compiled, err := as.Compiler.Thing(arg)
				if err != nil {
					return fmt.Errorf("compiling init: %w", err)
				}
				if compiled != nil {
					xtracer.Trace("compiler.ARGSetup.init compiled sort=%v", compiled.NodeSort())
					//pp("type=%T val=%v", compiled, compiled)
					acfg := as.Compiler.Module.Cfg.AstCfg
					mlf := acfg.NewLabeledFormula(nil, compiled)
					mod.LabeledInits = append(mod.LabeledInits, mlf)
					initClauses := FormulaToClauses(compiled, nil)
					xtracer.Trace("compiler.ARGSetup.init clauses fmlas=%d defs=%d", len(initClauses.Fmlas), len(initClauses.Defs))
					if mod.InitCond == nil {
						mod.InitCond = initClauses
					} else {
						mod.InitCond = AndClausesTyped(mod.InitCond, initClauses)
					}
				}
			}
		case *ProgressDecl:
			// Python IvyARGSetup.progress: progress properties are stored for later
			for _, arg := range n.DeclArgs {
				mod.Progress = append(mod.Progress, arg)
			}
		case *StateDecl:
			xtracer.Trace("compiler.ARGSetup.state ENTER")
			// Python IvyARGSetup.state (ivy_compiler.py:1420-1421):
			//   self.mod.predicates[a.args[0].relname] = a.args[1]
			for _, arg := range n.DeclArgs {
				if lf, ok := arg.(*LabeledFormula); ok {
					if def, ok := lf.Formula.(*Definition); ok {
						key := compilerExtractSortRep(def.Lhs)
						if key != "" {
							mod.Predicates[key] = def.Rhs
						}
					}
				}
			}
		case *ScenarioDecl:
			xtracer.Trace("compiler.ARGSetup.scenario ENTER")
			// Python IvyARGSetup.scenario (ivy_compiler.py:1462-1530)
			for _, arg := range n.DeclArgs {
				if sdef, ok := arg.(*ScenarioDef); ok {
					if err := as.scenario(sdef); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// scenario processes a ScenarioDef during ARGSetup (pass 3).
// Corresponds to Python IvyARGSetup.scenario (ivy_compiler.py:1462-1530).
func (as *ARGSetup) scenario(scen *ScenarioDef) error {
	mod := as.Compiler.Module
	sig := as.Compiler.Sig

	// 1. Build initTokens set from init places
	initTokens := make(map[string]bool)
	if initPL := scen.InitPlaces(); initPL != nil {
		for _, p := range initPL.Elems {
			if atom, ok := p.(*Atom); ok {
				initTokens[atom.Rep] = true
			}
		}
	}

	// 2. Group transitions by action name
	//    Python: transs_by_action[tr.args[2].args[1].args[0].rep]
	//    tr.Action = ScenarioBeforeMixin/AfterMixin, .Def = ActionDef, .Def.Name = Atom
	type transEntry struct {
		tr *ScenarioTransition
	}
	transsByAction := make(map[string][]transEntry)
	for _, tr := range scen.Transitions() {
		var actionName string
		switch m := tr.Action.(type) {
		case *ScenarioBeforeMixin:
			if adef, ok := m.Def.(*ActionDef); ok {
				actionName = adef.Defines()
			}
		case *ScenarioAfterMixin:
			if adef, ok := m.Def.(*ActionDef); ok {
				actionName = adef.Defines()
			}
		}
		if actionName != "" {
			transsByAction[actionName] = append(transsByAction[actionName], transEntry{tr: tr})
		}
	}

	// 3. Create init actions for each place
	//    Python: for (place_name, lineno) in scen.places():
	for _, pi := range scen.Places() {
		sym, err := sig.FindSymbol(pi.Name, false)
		if err != nil {
			return fmt.Errorf("scenario init: %w", err)
		}

		iname := pi.Name + "[init]"
		var rhs Expr
		if initTokens[pi.Name] {
			rhs = &LogicAnd{} // true
		} else {
			rhs = &LogicOr{} // false
		}
		iact := NewAssignAction(sym, rhs)
		iact.SetFormalParams(nil)
		iact.SetFormalReturns(nil)
		iact.SetLineno(scen.GetLineno())
		mod.SetAction(iname, iact)

		// Register MixinAfterDef: place[init] after init
		cfg := mod.Cfg.AstCfg
		mixerAtom := cfg.NewAtom(iname)
		mixeeAtom := cfg.NewAtom("init")
		mdef := cfg.NewMixinAfterDef(mixerAtom, mixeeAtom)
		mixee := mdef.Mixee()
		existing, _ := mod.Mixins.Get2(mixee)
		mod.Mixins.Set(mixee, append(existing, mdef))
		xtracer.Trace("compiler.ARGSetup.mixin mixee=%s mixer=%s kind=%s", mixee, mdef.Mixer(), "MixinAfterDef")
	}

	// 4. For each action's transitions, create mixer actions
	for _, trs := range transsByAction {
		var choices []interface{}
		var params []*Const
		var returns []*Const
		var afters []interface{}
		var mixer *Atom
		var mixee Node

		renameToFirstTransition := func(act ActionsAction, body ActionsAction) ActionsAction {
			aparams := append([]*Const{}, body.GetFormalParams()...)
			aparams = append(aparams, body.GetFormalReturns()...)
			canonical := append([]*Const{}, params...)
			canonical = append(canonical, returns...)

			subst := make(map[NodeKey]Expr)
			for j := 0; j < len(aparams) && j < len(canonical); j++ {
				subst[Key(aparams[j])] = canonical[j]
			}
			if len(subst) == 0 {
				return act
			}
			renamed := SubstituteConstantsAction(act, subst)
			renamed.SetLineno(act.GetLineno())
			return renamed
		}

		for i, te := range trs {
			tr := te.tr
			var scmix Node
			var isAfter bool
			var df *ActionDef

			switch m := tr.Action.(type) {
			case *ScenarioBeforeMixin:
				scmix = m
				isAfter = false
				df = m.Def.(*ActionDef)
			case *ScenarioAfterMixin:
				scmix = m
				isAfter = true
				df = m.Def.(*ActionDef)
			}
			_ = scmix

			// Compile the action body
			body, err := as.Compiler.CompileAction(df)
			if err != nil {
				return fmt.Errorf("scenario compile action: %w", err)
			}

			// Build sequence of place assignments
			var seq []Expr

			fromPL, _ := tr.From.(*PlaceList)
			toPL, _ := tr.To.(*PlaceList)

			if !isAfter {
				// Before: assume sources, set sources false, set targets true, append body
				if fromPL != nil {
					for _, p := range fromPL.Elems {
						if atom, ok := p.(*Atom); ok {
							sym, err := sig.FindSymbol(atom.Rep, false)
							if err != nil {
								return fmt.Errorf("scenario from place: %w", err)
							}
							assm := NewAssumeAction(sym)
							assm.SetLineno(tr.GetLineno())
							seq = append(seq, assm)
						}
					}
				}
				if fromPL != nil {
					for _, p := range fromPL.Elems {
						if atom, ok := p.(*Atom); ok {
							sym, _ := sig.FindSymbol(atom.Rep, false)
							assn := NewAssignAction(sym, &LogicOr{})
							assn.SetLineno(tr.GetLineno())
							seq = append(seq, assn)
						}
					}
				}
				if toPL != nil {
					for _, p := range toPL.Elems {
						if atom, ok := p.(*Atom); ok {
							sym, _ := sig.FindSymbol(atom.Rep, false)
							assn := NewAssignAction(sym, &LogicAnd{})
							assn.SetLineno(tr.GetLineno())
							seq = append(seq, assn)
						}
					}
				}
				seq = append(seq, body)
				seqAction := NewSequence(seq...)
				seqAction.SetLineno(tr.GetLineno())

				if i == 0 {
					params = body.GetFormalParams()
					returns = body.GetFormalReturns()
					switch m := tr.Action.(type) {
					case *ScenarioBeforeMixin:
						mixer = m.Mixer.(*Atom)
						if adef, ok := m.Def.(*ActionDef); ok {
							mixee = adef.Name
						}
					case *ScenarioAfterMixin:
						mixer = m.Mixer.(*Atom)
						if adef, ok := m.Def.(*ActionDef); ok {
							mixee = adef.Name
						}
					}
				} else {
					seqAction = renameToFirstTransition(seqAction, body).(*LogicSequence)
				}

				choices = append(choices, seqAction)
			} else {
				// After: set sources false, set targets true, append body, wrap in IfAction
				if fromPL != nil {
					for _, p := range fromPL.Elems {
						if atom, ok := p.(*Atom); ok {
							sym, _ := sig.FindSymbol(atom.Rep, false)
							assn := NewAssignAction(sym, &LogicOr{})
							assn.SetLineno(tr.GetLineno())
							seq = append(seq, assn)
						}
					}
				}
				if toPL != nil {
					for _, p := range toPL.Elems {
						if atom, ok := p.(*Atom); ok {
							sym, _ := sig.FindSymbol(atom.Rep, false)
							assn := NewAssignAction(sym, &LogicAnd{})
							assn.SetLineno(tr.GetLineno())
							seq = append(seq, assn)
						}
					}
				}
				seq = append(seq, body)
				seqAction := NewSequence(seq...)
				seqAction.SetLineno(tr.GetLineno())

				// IfAction(And(sources...), seq)
				var conds []Expr
				if fromPL != nil {
					for _, p := range fromPL.Elems {
						if atom, ok := p.(*Atom); ok {
							sym, _ := sig.FindSymbol(atom.Rep, false)
							conds = append(conds, sym)
						}
					}
				}
				var condExpr Expr
				if len(conds) == 0 {
					condExpr = &LogicAnd{}
				} else if len(conds) == 1 {
					condExpr = conds[0]
				} else {
					condExpr = &LogicAnd{Terms: conds}
				}
				ifAct := NewIfAction(condExpr, seqAction)
				ifAct.SetLineno(tr.GetLineno())

				if i == 0 {
					params = body.GetFormalParams()
					returns = body.GetFormalReturns()
					switch m := tr.Action.(type) {
					case *ScenarioBeforeMixin:
						mixer = m.Mixer.(*Atom)
						if adef, ok := m.Def.(*ActionDef); ok {
							mixee = adef.Name
						}
					case *ScenarioAfterMixin:
						mixer = m.Mixer.(*Atom)
						if adef, ok := m.Def.(*ActionDef); ok {
							mixee = adef.Name
						}
					}
				} else {
					ifAct = renameToFirstTransition(ifAct, body).(*LogicIfAction)
				}

				afters = append(afters, ifAct)
			}
		}

		// Register before choices
		if len(choices) > 0 {
			choice := BalancedChoice(choices, as.Compiler.ActCfg)
			if act, ok := choice.(ActionsAction); ok {
				if firstAct, ok := choices[0].(ActionsAction); ok {
					act.SetLineno(firstAct.GetLineno())
				}
				act.SetFormalParams(params)
				act.SetFormalReturns(returns)
				if mixer != nil {
					mod.SetAction(mixer.Rep, act)
				}
			}
			if mixer != nil && mixee != nil {
				mdef := mod.Cfg.AstCfg.NewMixinBeforeDef(mixer, mixee)
				mixeeName := mdef.Mixee()
				existingM, _ := mod.Mixins.Get2(mixeeName)
				mod.Mixins.Set(mixeeName, append(existingM, mdef))
				xtracer.Trace("compiler.ARGSetup.mixin mixee=%s mixer=%s kind=%s", mixeeName, mdef.Mixer(), "MixinBeforeDef")
			}
		}

		// Register after sequences
		if len(afters) > 0 {
			var seqExprs []Expr
			for _, a := range afters {
				if act, ok := a.(ActionsAction); ok {
					seqExprs = append(seqExprs, act)
				}
			}
			seqAct := NewSequence(seqExprs...)
			if firstAfter, ok := afters[0].(ActionsAction); ok {
				seqAct.SetLineno(firstAfter.GetLineno())
			}
			seqAct.SetFormalParams(params)
			seqAct.SetFormalReturns(returns)
			if mixer != nil {
				mod.SetAction(mixer.Rep, seqAct)
			}
			if mixer != nil && mixee != nil {
				mdef := mod.Cfg.AstCfg.NewMixinAfterDef(mixer, mixee)
				mixeeName := mdef.Mixee()
				existingM, _ := mod.Mixins.Get2(mixeeName)
				mod.Mixins.Set(mixeeName, append(existingM, mdef))
				xtracer.Trace("compiler.ARGSetup.mixin mixee=%s mixer=%s kind=%s", mixeeName, mdef.Mixer(), "MixinAfterDef")
			}
		}
	}

	return nil
}

// --- Post-processing ---

// FixConstructors ensures constructor sorts are properly set.
// Corresponds to Python's fix_constructors (ivy_compiler.py:1904-1922).
// For zero-arg constructors of structured types, rebuilds their sort
// to include destructor range sorts as domain.
func FixConstructors(mod *Module) {
	xtracer.Trace("compiler.FixConstructors ENTER")
	sig := mod.Sig
	if sig == nil {
		xtracer.Trace("compiler.FixConstructors EXIT")
		return
	}
	for sortname, destrs := range mod.SortDestructors {
		// Skip higher-order: any destructor with len(dom) > 1
		higherOrder := false
		for _, f := range destrs {
			if fs, ok := f.CSort.(*LogicFunctionSort); ok {
				if len(fs.Domain()) > 1 {
					higherOrder = true
					break
				}
			}
		}
		if higherOrder {
			continue
		}

		conss, ok := mod.SortConstructors[sortname]
		if !ok {
			conss = nil
		}

		newCons := make([]*Const, 0, len(conss))
		for _, cons := range conss {
			// Get domain of constructor
			var dom []Sort
			if fs, ok := cons.CSort.(*LogicFunctionSort); ok {
				dom = fs.Domain()
			}
			// If zero-arg constructor but sort has destructors, rebuild
			if len(dom) == 0 && len(destrs) > 0 {
				// new_dom = [f.sort.rng for f in destrs]
				newDomPlusRng := make([]Sort, 0, len(destrs)+1)
				for _, f := range destrs {
					if fs, ok := f.CSort.(*LogicFunctionSort); ok {
						newDomPlusRng = append(newDomPlusRng, fs.Range())
					}
				}
				// Append the constructor's range sort
				var rng Sort
				if fs, ok := cons.CSort.(*LogicFunctionSort); ok {
					rng = fs.Range()
				} else {
					rng = cons.CSort
				}
				newDomPlusRng = append(newDomPlusRng, rng)

				// Remove old, add new, find updated symbol
				sig.RemoveSymbol(cons.Name, cons.CSort)
				newSort := FuncConstSort(newDomPlusRng...)
				sig.AddSymbol(cons.Name, newSort)
				updated, err := sig.FindSymbol(cons.Name, false)
				if err == nil {
					cons = updated
				}
			}
			newCons = append(newCons, cons)
		}
		mod.SortConstructors[sortname] = newCons
	}
	xtracer.Trace("compiler.FixConstructors EXIT")
}

// CreateSortOrder creates a topological ordering of types.
// Corresponds to Python's create_sort_order (ivy_compiler.py:1632-1649).
func CreateSortOrder(mod *Module) error {
	xtracer.Trace("compiler.CreateSortOrder ENTER")
	if len(mod.SortOrder) == 0 {
		xtracer.Trace("compiler.CreateSortOrder EXIT")
		return nil
	}
	// Build arcs: (dependency, sort) for each sort in sort_order
	var arcs [][2]string
	for _, s := range mod.SortOrder {
		deps := mod.SortDependencies(s, false)
		for _, dep := range deps {
			arcs = append(arcs, [2]string{dep, s})
		}
	}
	// Check if already sorted
	number := make(map[string]int)
	for i, x := range mod.SortOrder {
		number[x] = i
	}
	alreadySorted := true
	for _, arc := range arcs {
		x, y := arc[0], arc[1]
		if x == "bool" {
			continue
		}
		nx, okX := number[x]
		ny, okY := number[y]
		if !okX || !okY || nx >= ny {
			alreadySorted = false
			break
		}
	}
	if alreadySorted {
		xtracer.Trace("compiler.CreateSortOrder EXIT")
		return nil
	}
	// Check for cycles using TarjanArcs
	sccs := TarjanArcs(arcs)
	if len(sccs) > 0 {
		xtracer.Trace("compiler.CreateSortOrder EXIT")
		return &IvyError{Msg: fmt.Sprintf("these sorts form a dependency cycle: %s", strings.Join(sccs[0], ","))}
	}
	// Topological sort
	mod.SortOrder = TopologicalSort(mod.SortOrder, arcs, func(s string) string { return s })
	xtracer.Trace("compiler.CreateSortOrder EXIT")
	return nil
}

// CreateConstructorSchemata creates axiom schemata for constructors.
// Corresponds to Python's create_constructor_schemata (ivy_compiler.py:1859-1901).
// Part A: For each structured sort, creates an existence schema.
// Part B: For each constructor, creates a destructor-inverse schema.
// Part C: Validates constructors have destructors.
func CreateConstructorSchemata(mod *Module) error {
	xtracer.Trace("compiler.CreateConstructorSchemata ENTER")
	sig := mod.Sig
	if sig == nil {
		xtracer.Trace("compiler.CreateConstructorSchemata EXIT")
		return nil
	}

	// Part A + B: iterate sort_destructors
	for sortname, destrs := range mod.SortDestructors {
		// Skip higher-order: any destructor with len(dom) > 1
		higherOrder := false
		for _, f := range destrs {
			if fs, ok := f.CSort.(*LogicFunctionSort); ok {
				if len(fs.Domain()) > 1 {
					higherOrder = true
					break
				}
			}
		}
		if higherOrder {
			continue
		}

		sort, err := sig.FindSort(sortname, false)
		if err != nil {
			continue
		}

		// Part A: generic existence schema
		// Y = LogicVariable('Y', sort)
		yVar, err := NewVariable("Y", sort)
		if err != nil {
			continue
		}

		// eqs = [Equals(f(Y), Variable('X'+n, f.sort.rng)) for n,f in enumerate(destrs)]
		eqs := make([]Expr, 0, len(destrs))
		for n, f := range destrs {
			var rng Sort
			if fs, ok := f.CSort.(*LogicFunctionSort); ok {
				rng = fs.Range()
			} else {
				continue
			}
			// f(Y)
			fY, err := f.Call(yVar)
			if err != nil {
				continue
			}
			// Variable('X'+n, f.sort.rng)
			xVar, err := NewVariable(fmt.Sprintf("X%d", n), rng)
			if err != nil {
				continue
			}
			eqs = append(eqs, NewEquals(fY, xVar))
		}

		// fmla = LogicExists([Y], And(*eqs))
		fmla := IvyExists([]*LogicVariable{yVar}, NormalizedAnd(eqs...))

		// name = Atom(compose_names(sortname, 'constr'), [])
		schemaName := mod.Cfg.AstCfg.NewAtom(mod.Cfg.IuCfg.ComposeNames(sortname, "constr"))

		// sch = SchemaBody(fmla)
		// We wrap the formula as the single element (conclusion) of the schema
		cfg := mod.Cfg.AstCfg
		sch := cfg.NewSchemaBody(fmla)

		// goal = LabeledFormula(name, sch)
		goal := cfg.NewLabeledFormula(schemaName, sch)
		xtracer.Trace("compiler.CreateConstructorSchemata schemata.insert key='%s' value=%s", schemaName.Relname(), goal.Canon())
		mod.Schemata.Set(schemaName.Relname(), goal)

		// Part B: per-constructor schema
		conss, ok := mod.SortConstructors[sortname]
		if !ok {
			continue
		}
		for _, cons := range conss {
			// Validate arg count matches destructor count
			var dom []Sort
			if fs, ok := cons.CSort.(*LogicFunctionSort); ok {
				dom = fs.Domain()
			}
			if len(dom) != len(destrs) {
				// Python: raise IvyError(cons, "Constructor {} has wrong number of arguments ...")
				xtracer.Trace("compiler.CreateConstructorSchemata EXIT")
				return NewIvyError(nil, fmt.Sprintf(
					"Constructor %s has wrong number of arguments (got %d, expecting %d)",
					cons.Name, len(dom), len(destrs)))
			}
			// Validate each arg sort matches destructor range sort
			for i, d := range dom {
				if fs, ok := destrs[i].CSort.(*LogicFunctionSort); ok {
					if len(fs.Domain()) != 1 {
						// Python: raise IvyError(cons, "Cannot define constructor ... because field ... has higher type")
						xtracer.Trace("compiler.CreateConstructorSchemata EXIT")
						return NewIvyError(nil, fmt.Sprintf(
							"Cannot define constructor %s for type %s because field %s has higher type",
							cons.Name, sortname, destrs[i].Name))
					}
					if d.String() != fs.Range().String() {
						// Python: raise IvyError(cons, "In constructor ..., argument ... has wrong type ...")
						xtracer.Trace("compiler.CreateConstructorSchemata EXIT")
						return NewIvyError(nil, fmt.Sprintf(
							"In constructor %s, argument %s has wrong type (expecting %s, got %s)",
							cons.Name, destrs[i].Name, fs.Range(), d))
					}
				}
			}

			// xvars = [Variable('X'+n, f.sort.rng) for n,f in enumerate(destrs)]
			xvars := make([]Expr, 0, len(destrs))
			for n, f := range destrs {
				if fs, ok := f.CSort.(*LogicFunctionSort); ok {
					xv, err := NewVariable(fmt.Sprintf("X%d", n), fs.Range())
					if err != nil {
						continue
					}
					xvars = append(xvars, xv)
				}
			}

			// Y = cons(*xvars)
			consY, err := cons.Call(xvars...)
			if err != nil {
				continue
			}

			// eqs = [Equals(f(Y), X_n) for each destructor]
			consEqs := make([]Expr, 0, len(destrs))
			for n, f := range destrs {
				if fs, ok := f.CSort.(*LogicFunctionSort); ok {
					fY, err := f.Call(consY)
					if err != nil {
						continue
					}
					xv, _ := NewVariable(fmt.Sprintf("X%d", n), fs.Range())
					consEqs = append(consEqs, NewEquals(fY, xv))
				}
			}

			// fmla = LogicAnd(*eqs)
			consFmla := NormalizedAnd(consEqs...)

			// name = Atom(compose_names(cons.name, 'constr'), [])
			consSchemaName := mod.Cfg.AstCfg.NewAtom(mod.Cfg.IuCfg.ComposeNames(cons.Name, "constr"))

			// sch = SchemaBody(fmla)
			consSch := cfg.NewSchemaBody(consFmla)

			// goal = LabeledFormula(name, sch)
			consGoal := cfg.NewLabeledFormula(consSchemaName, consSch)
			xtracer.Trace("compiler.CreateConstructorSchemata.nested schemata.insert key='%s' value=%s", consSchemaName.Relname(), consGoal.Canon())
			mod.Schemata.Set(consSchemaName.Relname(), consGoal)
		}
	}

	// Part C: validate constructors have destructors
	// Python: raise IvyError(cons, "Cannot define constructor ... because ... is not a structure type")
	for sortname, conss := range mod.SortConstructors {
		for _, cons := range conss {
			if _, ok := mod.SortDestructors[sortname]; !ok {
				xtracer.Trace("compiler.CreateConstructorSchemata EXIT")
				return NewIvyError(nil, fmt.Sprintf(
					"Cannot define constructor %s for type %s because %s is not a structure type",
					cons.Name, sortname, sortname))
			}
		}
	}
	xtracer.Trace("compiler.CreateConstructorSchemata EXIT")
	return nil
}

// AttachProofs attaches proofs to their corresponding properties.
// Corresponds to Python's attach_proofs (ivy_compiler.py:1672-1694).
func AttachProofs(mod *Module) error {
	xtracer.Trace("compiler.AttachProofs ENTER")
	// Build label → LabeledFormula map from props and conjs
	m := make(map[string]*LabeledFormula)
	for _, lf := range mod.LabeledProps {
		if name := labelName(lf.Label); name != "" {
			m[name] = lf
		}
	}
	for _, lf := range mod.LabeledConjs {
		if name := labelName(lf.Label); name != "" {
			m[name] = lf
		}
	}

	used := make(map[string]bool)
	pfs := mod.Proofs
	mod.Proofs = nil

	// First pass: proofs with non-nil formula
	for _, pf := range pfs {
		if pf.Formula != nil && pf.Formula.Formula != nil {
			mod.Proofs = append(mod.Proofs, pf)
			if name := labelName(pf.Formula.Label); name != "" {
				used[name] = true
			}
		}
	}

	// Second pass: proofs with nil formula (label-only references)
	for _, pf := range pfs {
		if pf.Formula == nil || pf.Formula.Formula != nil {
			continue
		}
		lab := labelName(pf.Formula.Label)
		if lab == "" {
			xtracer.Trace("compiler.AttachProofs EXIT")
			return NewIvyError(pf.Formula, "proof has empty label")
		}
		if used[lab] {
			xtracer.Trace("compiler.AttachProofs EXIT")
			return NewIvyError(pf.Formula, fmt.Sprintf("duplicate proof for label: %s", lab))
		}
		used[lab] = true
		if target, ok := m[lab]; ok {
			mod.Proofs = append(mod.Proofs, ProofEntry{
				Formula: target,
				Proof:   pf.Proof,
			})
		} else if _, ok := mod.Isolates[lab]; ok {
			mod.IsolateProofs[lab] = pf.Proof
		} else {
			xtracer.Trace("compiler.AttachProofs EXIT")
			return NewIvyError(pf.Formula, fmt.Sprintf("no property, conjecture, or isolate for proof label: %s", lab))
		}
	}
	xtracer.Trace("compiler.AttachProofs EXIT")
	return nil
}

// CheckDefinitions validates definitions for cycles and redefinition.
// Corresponds to Python's check_definitions (ivy_compiler.py:1696-1776).
func CheckDefinitions(mod *Module) error {
	xtracer.Trace("compiler.CheckDefinitions ENTER")
	defer xtracer.Trace("compiler.CheckDefinitions EXIT")

	// Get definitions that have no dependence on proofs.
	// stale uses structural keys (Sexp) matching Python's Symbol-as-dict-key semantics.
	stale := make(map[NodeKey]bool)
	withProofs := make(map[int64]bool)
	for _, pe := range mod.Proofs {
		if pe.Formula != nil {
			withProofs[pe.Formula.ID] = true
		}
	}

	props := mod.LabeledProps
	mod.LabeledProps = nil

	for _, prop := range props {
		if logicDef, ok := prop.Formula.(*LogicDefinition); ok {
			defKey := compilerDefinesKey(logicDef)
			if !withProofs[prop.ID] {
				// Check if any used symbols are stale
				hasStale := false
				if expr, ok := prop.Formula.(Expr); ok {
					for _, sym := range UsedSymbolsAst(expr).All() {
						if stale[Key(sym)] {
							hasStale = true
							break
						}
					}
				}
				if !hasStale {
					mod.Definitions = append(mod.Definitions, prop)
					continue
				}
			}
			stale[defKey] = true
		}
		mod.LabeledProps = append(mod.LabeledProps, prop)
	}

	// Check for redefinition and interpreted symbols.
	// Python: checkdef(sym, lf) checks redefinition first, then solver.solver_name(sym).
	// Uses structural keys (Sexp) so symbols with same name but different sorts don't collide.
	defs := make(map[NodeKey]*LabeledFormula)
	checkdef := func(key NodeKey, name string, symObj *Const, lf *LabeledFormula) error {
		// Python: if sym in defs: raise IvyError('redefinition of ...')
		if prev, exists := defs[key]; exists {
			return NewIvyError(lf, fmt.Sprintf("redefinition of %s\n%d from here", name, prev.Lineno()))
		}
		// Python: if solver.solver_name(sym) == None: raise IvyError('definition of interpreted symbol ...')
		if symObj != nil {
			solverN, err := SolverName(symObj, mod.Sig, nil)
			if err != nil {
				return err // z3 builtin clash
			}
			if solverN == "" {
				return NewIvyError(lf, fmt.Sprintf("definition of interpreted symbol %s", name))
			}
		}
		defs[key] = lf
		return nil
	}
	for _, ldf := range mod.Definitions {
		if logicDef, ok := ldf.Formula.(*LogicDefinition); ok {
			defExpr := logicDef.Defines()
			key := Key(defExpr)
			name := defExprName(defExpr)
			var symObj *Const
			if s, ok := defExpr.(*Const); ok {
				symObj = s
			}
			if err := checkdef(key, name, symObj, ldf); err != nil {
				return err
			}
		}
	}
	// Python: for ldf in mod.native_definitions: checkdef(ldf.formula.defines(), ldf)
	for _, ldf := range mod.NativeDefinitions {
		if ldf != nil {
			if logicDef, ok := ldf.Formula.(*LogicDefinition); ok {
				defExpr := logicDef.Defines()
				key := Key(defExpr)
				name := defExprName(defExpr)
				var symObj *Const
				if s, ok := defExpr.(*Const); ok {
					symObj = s
				}
				if err := checkdef(key, name, symObj, ldf); err != nil {
					return err
				}
			}
		}
	}
	// Python: for ldf, term in mod.named: checkdef(term.rep, ldf)
	for _, ne := range mod.Named {
		if sym, ok := ne.Name.(*Const); ok {
			if err := checkdef(Key(sym), sym.Name, sym, ne.Formula); err != nil {
				return err
			}
		}
	}

	// Action interference check (v1.7+).
	// Uses structural NodeKey throughout, matching Python's Symbol-as-dict-key semantics.
	// Python: if iu.version_le("1.7", iu.get_string_version()): ...
	xtracer.Trace("compiler.ActionInterferenceCheck ENTER version=%s", mod.Cfg.IuCfg.GetStringVersion())
	if VersionLE("1.7", mod.Cfg.IuCfg.GetStringVersion()) {
		// Create ActionsConfig with module context so isDestructor() can
		// check mod.DestructorSorts, matching Python's ivy_module.module.destructor_sorts.
		interferenceActCfg := &ActionsConfig{
			Context: NewActionContext(mod),
		}
		// Dump all action keys in insertion order for comparison.
		{
			var allKeys []string
			for name := range mod.Actions.All() {
				allKeys = append(allKeys, name)
			}
			xtracer.Trace("compiler.ActionInterferenceCheck allKeys=%d keys=%s", len(allKeys), strings.Join(allKeys, ","))
		}
		// First loop: build the modified set (order-independent).
		// Python: side_effects = dict(); for action in list(mod.actions.values()):
		//             for sub in action.iter_subactions():
		//                 for s in sub.modifies(): side_effects[s] = sub
		modified := make(map[NodeKey]bool)
		for _, actVal := range mod.Actions.All() {
			if act, ok := actVal.(ActionsAction); ok {
				for _, sub := range act.IterSubactions() {
					mods := Modifies(sub, interferenceActCfg)
					for _, sym := range mods {
						modified[Key(sym)] = true
					}
				}
			}
		}
		// Second loop (xtrace only): iterate in insertion order, deduplicate via set.
		// Python: for name,actval in mod.actions.items():
		//             mod_syms = set(); for sub in actval.iter_subactions(): mod_syms.update(sub.modifies())
		//             if mod_syms: xtracer.trace(...)
		for name, actVal := range mod.Actions.All() {
			if act, ok := actVal.(ActionsAction); ok {
				modSyms := make(map[NodeKey]bool)
				modNames := make(map[string]bool)
				for _, sub := range act.IterSubactions() {
					mods := Modifies(sub, interferenceActCfg)
					for _, sym := range mods {
						modSyms[Key(sym)] = true
						modNames[sym.Name] = true
					}
				}
				if len(modSyms) > 0 {
					var symNames []string
					for n := range modNames {
						symNames = append(symNames, n)
					}
					sort.Strings(symNames)
					xtracer.Trace("compiler.ActionInterferenceCheck action=%s modifies=%d syms=%s", name, len(modSyms), strings.Join(symNames, ","))
				}
			}
		}
		// Build definition map for transitive dep lookup (NodeKey keys).
		interferenceDefMap := make(map[NodeKey]interface{})
		for _, lf := range mod.Definitions {
			if def, ok := lf.Formula.(*LogicDefinition); ok {
				interferenceDefMap[compilerDefinesKey(def)] = def.Rhs
			}
		}
		// Check axioms: no side-effected symbol may appear in axiom deps
		// Python: if not opt_mutax.get(): ...
		optMutax := OptMutax.GetBool()
		if mod.Cfg != nil {
			optMutax = mod.Cfg.OptMutax
		}
		if !optMutax {
			for _, lf := range mod.LabeledAxioms {
				if !lf.IsTemporal() {
					deps := make(map[NodeKey]bool)
					GetSymbolDependencies(interferenceDefMap, deps, lf.Formula)
					for sym := range deps {
						if modified[sym] {
							xtracer.Trace("compiler.ActionInterferenceCheck FAIL axiom immutable sym=%s", sym)
							return NewIvyError(lf, fmt.Sprintf("immutable symbol assigned: %s", sym))
						}
					}
				}
			}
		}
		// Check definitions: LHS must not be modified
		for _, lf := range mod.Definitions {
			if def, ok := lf.Formula.(*LogicDefinition); ok {
				key := compilerDefinesKey(def)
				if modified[key] {
					xtracer.Trace("compiler.ActionInterferenceCheck FAIL defn immutable key=%s", key)
					return NewIvyError(lf, fmt.Sprintf("immutable symbol assigned: %s", key))
				}
			}
		}
	}
	xtracer.Trace("compiler.ActionInterferenceCheck EXIT")

	// Check definition cycles via arcs.
	// Uses structural keys (Sexp) for arc nodes, matching Python's Symbol equality.
	var arcs [][2]string
	dmap := make(map[NodeKey]*LabeledFormula)
	for _, d := range mod.Definitions {
		if logicDef, ok := d.Formula.(*LogicDefinition); ok {
			defKey := compilerDefinesKey(logicDef)
			dmap[defKey] = d
			if rhs, ok := logicDef.Rhs.(Expr); ok {
				for _, sym := range UsedSymbolsAst(rhs).All() {
					arcs = append(arcs, [2]string{string(defKey), string(Key(sym))})
				}
			}
		}
	}
	// Build proof map: formula ID → proof
	pmap := make(map[int64]Node)
	for _, pe := range mod.Proofs {
		if pe.Formula != nil {
			pmap[pe.Formula.ID] = pe.Proof
		}
	}
	sccs := TarjanArcs(arcs)
	// Python line 2028: prover = ivy_proof.ProofChecker(mod.labeled_axioms, [], mod.schemata)
	// Create ONE shared ProofChecker before the SCC loop, matching Python.
	// Python creates this unconditionally; the normalize_goal calls in __init__
	// create LabeledFormulas that advance the LF counter.
	var proofCfg *ProofConfig
	var astCfg *AstConfig
	if mod.Cfg != nil {
		proofCfg = mod.Cfg.ProofCfg
		astCfg = mod.Cfg.AstCfg
	}
	defProver := NewProofChecker(proofCfg, mod, mod.LabeledAxioms, nil, ModuleSchemataToAst(mod.Schemata), astCfg)
	for _, scc := range sccs {
		if len(scc) > 1 {
			return &IvyError{Msg: fmt.Sprintf("these definitions form a dependency cycle: %s", strings.Join(scc, ","))}
		}
		// Singleton SCC with self-loop: requires recursion schema (proof)
		defKey := scc[0]
		if d, ok := dmap[NodeKey(defKey)]; ok {
			proof, hasProof := pmap[d.ID]
			if !hasProof {
				return NewIvyError(d, fmt.Sprintf("definition of %s requires a recursion schema", defKey))
			}
			// Python: prover.admit_definition(d, pmap[d.id])
			if _, err := defProver.AdmitDefinition(d, proof); err != nil {
				return err
			}
		}
	}
	return nil
}

// labelName extracts a string name from a label Node.
func labelName(label Node) string {
	if label == nil {
		return ""
	}
	if atom, ok := label.(*Atom); ok {
		return atom.Relname()
	}
	return fmt.Sprint(label)
}

// definesKey returns a structural identity key for a definition's LHS symbol,
// matching Python's use of Symbol objects as dict keys with recstruct
// equality (name + sort). Uses lg.Key() / Sexp() for structural equivalence.
// Use this for maps that compare definition symbols against each other
// (defs, stale, arcs, dmap).
func compilerDefinesKey(d *LogicDefinition) NodeKey {
	return Key(d.Defines())
}

// defExprName returns a human-readable name from an expression, for error messages.
func defExprName(expr Expr) string {
	if sym, ok := expr.(*Const); ok {
		return sym.Name
	}
	return string(Key(expr))
}

// CreateConjActions creates conjecture actions for runtime verification.
// Corresponds to Python's create_conj_actions (ivy_compiler.py:2404-2451).
// For each conjecture, determines which actions must preserve it.
func CreateConjActions(mod *Module) error {
	xtracer.Trace("compiler.CreateConjActions ENTER")
	// Python: if iu.version_le(iu.get_string_version(), "1.6"): return
	if VersionLE(mod.Cfg.IuCfg.GetStringVersion(), "1.6") {
		xtracer.Trace("compiler.CreateConjActions EXIT")
		return nil
	}

	if mod.ConjActions == nil {
		mod.ConjActions = make(map[string][]string)
	}

	// Build isolate exports and object→isolate mapping
	// Python: myexports[isol.name()] = iso.get_isolate_exports(mod, cg, isol)
	//         objects[x.rep].append(isol) for x in isol.verified()
	type isoEntry struct {
		name string
		def  IsolateDefInterface
	}
	myexports := make(map[string]map[string]bool) // iso name → exported actions
	objects := make(map[string][]isoEntry)        // verified object → isolates
	cg := mod.CallGraph()

	for isoName, isol := range mod.Isolates {
		myexports[isoName] = GetIsolateExports(mod, cg, isol)
		for _, v := range isol.VerifiedNames() {
			objects[v] = append(objects[v], isoEntry{isoName, isol})
		}
	}

	for _, conj := range mod.LabeledConjs {
		if conj.Label == nil {
			continue
		}
		lbl := labelName(conj.Label)
		origLbl := lbl

		// Python: while lbl != 'this' and lbl not in objects: lbl, _ = iu.parent_child_name(lbl)
		for lbl != "this" {
			if _, found := objects[lbl]; found {
				break
			}
			parts := mod.Cfg.IuCfg.ParentChildName(lbl)
			lbl = parts[0]
		}

		var actionSet map[string]bool
		if lbl == "this" {
			// Top-level: all exported actions
			actionSet = make(map[string]bool)
			for _, exp := range mod.Exports {
				actionSet[exp.Exported()] = true
			}
		} else {
			// Isolate-scoped: only that isolate's exports
			actionSet = make(map[string]bool)
			for _, entry := range objects[lbl] {
				for act := range myexports[entry.name] {
					actionSet[act] = true
				}
			}
		}

		actionNames := make([]string, 0, len(actionSet))
		for act := range actionSet {
			actionNames = append(actionNames, act)
		}
		sort.Strings(actionNames)
		mod.ConjActions[origLbl] = actionNames
	}

	doCheckInterference := true
	if mod.Cfg != nil && mod.Cfg.IsolateCfg != nil {
		doCheckInterference = mod.Cfg.IsolateCfg.DoCheckInterference
	}
	if doCheckInterference {
		actionIsos := make(map[string]map[string]bool)
		for ison, actions := range myexports {
			for action := range actions {
				if actionIsos[action] == nil {
					actionIsos[action] = make(map[string]bool)
				}
				actionIsos[action][ison] = true
			}
		}

		isoNames := make([]string, 0, len(mod.Isolates))
		for ison := range mod.Isolates {
			isoNames = append(isoNames, ison)
		}
		sort.Strings(isoNames)
		for _, ison := range isoNames {
			isol := mod.Isolates[ison]
			memo := make(map[string]bool)
			conjs := GetIsolateConjs(mod, isol, false, true)
			exports := myexports[ison]
			roots := reachable(sortedBoolMapKeys(exports), func(x string) []string { return cg[x] })
			for _, conj := range conjs {
				lbl := labelName(conj.Label)
				actions := append([]string(nil), mod.ConjActions[lbl]...)
				sort.Strings(actions)
				for _, action := range actions {
					for _, ison1 := range sortedBoolMapKeys(actionIsos[action]) {
						if ison1 != ison && !memo[action] {
							memo[action] = true
							if roots[action] && !exports[action] {
								for _, victim := range sortedBoolMapKeys(exports) {
									victimRoots := reachable([]string{victim}, func(x string) []string { return cg[x] })
									if victimRoots[action] && action != victim {
										return NewIvyError(conj, fmt.Sprintf("isolate %s depends on invariant %s which might not hold because action %s is called from within action %s, which invalidates the invariant.", ison, lbl, victim, action))
									}
								}
							}
						}
					}
				}
			}
		}
	}

	xtracer.Trace("compiler.CreateConjActions EXIT")
	return nil
}

func sortedBoolMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// HandleTemporals processes temporal properties.
// Corresponds to Python's handle_temporals (ivy_compiler.py:2149-2170).
// Labels each action with the list of isolates in which it is present.
func HandleTemporals(mod *Module) {
	xtracer.Trace("compiler.HandleTemporals ENTER")
	// Get isolate map: action name → list of isolate names
	imap := GetIsolateMap(mod, true, true)
	for actname, action := range mod.Actions.All() {
		if labeler, ok := action.(interface{ SetLabels([]string) }); ok {
			// Use imap[actname] directly: returns nil for missing keys,
			// matching Python's defaultdict(list) returning [] for missing keys.
			labels := imap[actname]
			// Sort labels for deterministic ordering: Python preserves dict
			// insertion order, Go maps iterate randomly. Sorting both sides
			// ensures canonical s-expressions match.
			if len(labels) > 0 {
				sort.Strings(labels)
			}
			labeler.SetLabels(labels)
		} else {
			pp("HandleTemporals: action %s does not support SetLabels", actname)
		}
	}
	xtracer.Trace("compiler.HandleTemporals EXIT")
}

// TheoremToProperty converts a theorem (proved by schema/tactic) into
// a property for checking. Corresponds to Python's theorem_to_property
// (ivy_compiler.py:1786-1826).
//
// If the formula is a SchemaBody, it:
//  1. Extracts vocabulary (sorts, symbols) and renames any that collide with the global signature
//  2. Applies the rename match to the entire goal
//  3. Collects non-explicit/non-definition premises, appends definitions to mod.Definitions
//  4. Checks conclusion is not a Definition
//  5. Builds an implication from premises and conclusion
//
// Otherwise returns the property unchanged.
func TheoremToProperty(goal *LabeledFormula, mod *Module) *LabeledFormula {
	if goal == nil {
		return nil
	}
	if _, ok := goal.Formula.(*SchemaBody); !ok {
		return goal
	}

	// Step A: Extract vocabulary and build rename match
	sig := mod.Sig
	vocab := GoalVocab(goal)
	match := make(map[NodeKey]Expr)

	for _, sort := range vocab.Sorts {
		name := IvySortName(sort)
		if _, exists := sig.Sorts.Get2(name); exists {
			// Name collision — generate unique name
			usedNames := make(map[string]struct{}, sig.Sorts.Len())
			for k, _ := range sig.Sorts.All() {
				usedNames[k] = struct{}{}
			}
			newname := UnusedNameWithBase(name, usedNames)
			newsort := &UninterpretedSort{Name: newname}
			sig.AddSort(newsort)
			match[Key(sort)] = newsort
		} else {
			sig.AddSort(sort)
		}
	}

	for _, sym := range vocab.Symbols {
		if _, exists := sig.Symbols.Get2(sym.Name); exists {
			usedNames := make(map[string]struct{}, sig.Symbols.Len())
			for k := range sig.Symbols.All() {
				usedNames[k] = struct{}{}
			}
			newname := UnusedNameWithBase(sym.Name, usedNames)
			newsym := NewConst(newname, sym.CSort)
			newsym = ApplyMatchFunc(match, newsym)
			sig.AddSymbol(newsym.Name, newsym.CSort)
			match[Key(sym)] = newsym
		} else {
			sig.AddSymbol(sym.Name, sym.CSort)
		}
	}

	// Step B: Apply rename match to entire goal
	prop := goal
	if len(match) > 0 {
		prop = ApplyMatchGoalNode(mod.Cfg.AstCfg, match, goal)
	}

	// Step C: Process premises
	sb := prop.Formula.(*SchemaBody) // re-extract after match
	var prems []Node
	for _, x := range sb.Prems() {
		if lf, ok := x.(*LabeledFormula); ok {
			if lf.IsDefinition {
				mod.Definitions = append(mod.Definitions, PropToDef(lf, mod).(*LabeledFormula))
			} else if !lf.Explicit {
				sub := TheoremToProperty(lf, mod)
				if sub != nil {
					prems = append(prems, sub.Formula)
				}
			}
		}
	}

	// Step D: Check conclusion is not a Definition
	conc := sb.Conc()
	if _, isDef := conc.(*LogicDefinition); isDef {
		panic(fmt.Sprintf("definitional subgoal must be discharged"))
	}

	// Step E: Build formula
	var fmla Node
	if len(prems) > 0 {
		premExprs := exprSlice(prems)
		concExpr := nodeToExpr(conc)
		if len(premExprs) > 0 {
			antecedent := NormalizedAnd(premExprs...)
			impl, err := NewImplies(antecedent, concExpr)
			if err != nil {
				fmla = conc
			} else {
				fmla = impl
			}
		} else {
			fmla = conc
		}
	} else {
		fmla = conc
	}
	acfg := mod.Cfg.AstCfg
	result := acfg.NewLabeledFormula(prop.Label, fmla)
	result.SetLineno(prop.GetLineno())
	return result
}

// exprSlice converts a []ast.Node to []lg.Expr where possible.
func exprSlice(nodes []Node) []Expr {
	result := make([]Expr, 0, len(nodes))
	for _, n := range nodes {
		if e, ok := n.(Expr); ok {
			result = append(result, e)
		}
	}
	return result
}

// nodeToExpr converts an ast.Node to a lg.Expr, returning lg.True if not possible.
func nodeToExpr(n Node) Expr {
	if e, ok := n.(Expr); ok {
		return e
	}
	return True
}

// progressDefinesName extracts the symbol name from a progress item.
// Progress items are either:
//   - lg.Expr (compiled via SortifyWithInference in pass 1) — extract from Apply.Func or Definition LHS
//   - *ast.LabeledFormula (raw from pass 3) — extract from Label
func progressDefinesName(p interface{}) string {
	// Pass 1: compiled lg.Expr
	if expr, ok := p.(Expr); ok {
		return exprDefinesName(expr)
	}
	// Pass 3: raw *ast.LabeledFormula
	if lf, ok := p.(*LabeledFormula); ok {
		if lf.Label != nil {
			if atom, ok := lf.Label.(*Atom); ok {
				return atom.Relname()
			}
		}
		// Try the formula field
		if lf.Formula != nil {
			if expr, ok := lf.Formula.(Expr); ok {
				return exprDefinesName(expr)
			}
		}
	}
	return ""
}

// exprDefinesName extracts the defined symbol name from a compiled lg.Expr.
// Handles: Apply (func name), Definition (LHS name), Symbol (name directly).
func exprDefinesName(expr Expr) string {
	switch e := expr.(type) {
	case *Apply:
		if sym, ok := e.Func.(*Const); ok {
			return sym.Name
		}
	case *LogicDefinition:
		defNode := e.Defines()
		if sym, ok := defNode.(*Const); ok {
			return sym.Name
		}
	case *Const:
		return e.Name
	}
	return ""
}

// registerIsolateDecl registers isolate definitions from an IsolateDecl or
// IsolateObjectDecl into the module's Isolates map. Called from both the
// *ast.IsolateDecl and *ast.IsolateObjectDecl cases in ARGSetup.ProcessDecls.
// Python IvyARGSetup.isolate (ivy_compiler.py:1429-1434):
//
//	self.mod.isolates[iso.name()] = iso.clone(args)
func registerIsolateDecl(declArgs []Node, mod *Module) {
	for _, arg := range declArgs {
		if isoDef, ok := arg.(*IsolateDef); ok {
			name := isoDef.IsoName()
			xtracer.Trace("compiler.ARGSetup.isolate ENTER name=%s", name)
			mod.Isolates[name] = isoDef
			xtracer.Trace("compiler.ARGSetup.isolate EXIT name=%s", name)
		}
	}
}
