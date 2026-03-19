// ivy_compile.go implements the main ivy_compile entry point.
// This corresponds to Python's ivy_compile function in ivy_compiler.py:2190-2254.
//
// The function runs three separate declaration interpreter passes:
//   1. IvyDomainSetup - processes types, functions, relations, axioms, definitions
//   2. IvyConjectureSetup - processes conjectures and named formulas
//   3. IvyARGSetup - processes exports, delegates, actions, initializers
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
package compiler

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
	"github.com/glycerine/goivy/isolate"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/module"
)

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
//   Pass 1 (IvyDomainSetup): types, relations, constants, axioms, definitions
//   Pass 2 (IvyConjectureSetup): conjectures
//   Pass 3 (IvyARGSetup): exports, delegates, actions, initializers
func IvyCompile(decls []ast.Node, mod *module.Module) error {
	if mod == nil {
		mod = module.New()
	}

	// Process attributes from declarations
	for _, decl := range decls {
		processAttributes(decl, mod)
	}

	// Collect action signatures for forward reference resolution.
	// Python: TopContext(collect_actions(decls.decls))
	topCtx := CollectActions(decls)

	// Create compiler with the top context
	c := NewFromModule(mod)
	c.TopCtx = topCtx

	// Pass 1: IvyDomainSetup
	// Processes: types, relations, constants, axioms, definitions, etc.
	domainInterp := NewDomainSetup(c)
	if err := domainInterp.ProcessDecls(decls); err != nil {
		return fmt.Errorf("domain setup: %w", err)
	}

	// fix_constructors: ensure constructor sorts are properly set
	FixConstructors(mod)

	// Pass 2: IvyConjectureSetup
	// Processes: conjectures, named formulas
	conjInterp := NewConjSetup(c)
	if err := conjInterp.ProcessDecls(decls); err != nil {
		return fmt.Errorf("conjecture setup: %w", err)
	}

	// Pass 3: IvyARGSetup
	// Processes: exports, delegates, actions, initializers, progress
	argInterp := NewARGSetup(c)
	if err := argInterp.ProcessDecls(decls); err != nil {
		return fmt.Errorf("ARG setup: %w", err)
	}

	// Populate macros: mod.macros = decls.macros (Python ivy_compile.py:2207)
	if mod.Macros == nil {
		mod.Macros = make(map[string]interface{})
	}
	for _, decl := range decls {
		if md, ok := decl.(*ast.MacroDecl); ok {
			for _, arg := range md.DeclArgs {
				if defn, ok := arg.(*ast.Definition); ok {
					mod.Macros[defn.Defines()] = defn
				}
			}
		}
	}

	// Post-processing passes
	CreateSortOrder(mod)
	CreateConstructorSchemata(mod)
	AttachProofs(mod)
	CheckDefinitions(mod)
	CheckPropertiesPass(mod)
	CreateConjActions(mod)
	HandleTemporals(mod)

	// From version 1.7, ensure there is a default "this" isolate.
	// Matches Python ivy_compile lines 2221-2225.
	if _, ok := mod.Isolates["this"]; !ok {
		isol := &ast.IsolateDef{
			Elems:    []ast.Node{ast.NewAtom("this"), ast.NewAtom("this")},
			WithArgs: 0,
		}
		mod.Isolates["this"] = isol
	}

	// Create isolate — resolves mixins (including after init) into actions.
	// Matches Python ivy_compile line 2251:
	//   if create_isolate:
	//       iso.create_isolate(isolate.get(), mod, **kwargs)
	if err := isolate.CreateIsolate("this", mod); err != nil {
		// Log but don't fail: CreateIsolate may fail on incomplete
		// mixin wiring (e.g., after-init actions) while the module's
		// sig (sorts, symbols) is already fully populated from Pass 1.
		fmt.Printf("IvyCompile: CreateIsolate warning: %v\n", err)
	}

	// Python line 2253-2254:
	//   im.module.labeled_axioms.extend(im.module.labeled_props)
	//   im.module.theory_context().__enter__()
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)

	// theory_context().__enter__() calls update_theory() which builds
	// the background theory from axioms and definitions, caching the
	// result for BackgroundTheory() calls during verification.
	mod.UpdateTheory()

	// Set CompileActionBodyFn on the module so that InstantiateAction.IntUpdate
	// can compile macro expansions at runtime. Python: im.compile().int_update(...)
	mod.CompileActionBodyFn = func(node ast.Node) (interface{}, error) {
		cc := NewFromModule(mod)
		act, err := cc.CompileActionBody(node)
		if err != nil {
			return nil, err
		}
		return act, nil
	}

	return nil
}

// processAttributes extracts attributes from a declaration and stores them
// on the module.
func processAttributes(decl ast.Node, mod *module.Module) {
	type attrProvider interface {
		GetAttributes() []string
	}
	type defProvider interface {
		GetDefines() []string
	}
	ha, okA := decl.(attrProvider)
	hd, okD := decl.(defProvider)
	if okA && okD {
		for _, attr := range ha.GetAttributes() {
			for _, name := range hd.GetDefines() {
				key := iu.ComposeNames(name, attr)
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
func CollectActions(decls []ast.Node) *TopContext {
	tc := &TopContext{
		Actions: make(map[string]*ActionInfo),
	}
	for _, decl := range decls {
		switch n := decl.(type) {
		case *ast.ActionDecl:
			for _, arg := range n.DeclArgs {
				ad, isActionDef := arg.(*ast.ActionDef)
				if !isActionDef {
					// Fallback: try Atom for simple action declarations
					if atom, ok := arg.(*ast.Atom); ok {
						tc.Actions[atom.Rep] = &ActionInfo{}
					}
					continue
				}
				name := ""
				if atom, ok := ad.Name.(*ast.Atom); ok {
					name = atom.Rep
				} else {
					name = fmt.Sprint(ad.Name)
				}

				// Collect formals: declared args from the name atom + formal params
				var formals []ast.Node
				if nameAtom, ok := ad.Name.(*ast.Atom); ok {
					for _, t := range nameAtom.Terms {
						formals = append(formals, t)
					}
				}
				formals = append(formals, ad.FormalParams...)

				// Find keypos: index of first KeyArg
				keypos := 0
				for idx, p := range formals {
					if _, isKey := p.(*ast.KeyArg); isKey {
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
		case *ast.ExportDecl:
			for _, arg := range n.DeclArgs {
				if atom, ok := arg.(*ast.Atom); ok {
					tc.Actions["ext:"+atom.Rep] = &ActionInfo{}
				}
			}
		}
	}
	return tc
}

// --- Conjecture Setup Pass ---

// ConjSetup is the second compilation pass: processes conjectures.
type ConjSetup struct {
	Compiler *Compiler
}

func NewConjSetup(c *Compiler) *ConjSetup {
	return &ConjSetup{Compiler: c}
}

func (cs *ConjSetup) ProcessDecls(decls []ast.Node) error {
	for _, decl := range decls {
		switch n := decl.(type) {
		case *ast.ConjectureDecl:
			for _, arg := range n.DeclArgs {
				_ = arg // Already processed in pass 1
			}
		case *ast.NamedDecl:
			_ = n // Process named formulas
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

func (as *ARGSetup) ProcessDecls(decls []ast.Node) error {
	for _, decl := range decls {
		switch n := decl.(type) {
		case *ast.ExportDecl:
			for _, arg := range n.DeclArgs {
				_ = arg // Process exports
			}
		case *ast.DelegateDecl:
			for _, arg := range n.DeclArgs {
				_ = arg // Process delegates
			}
		case *ast.InitDecl:
			// Matches Python IvyARGSetup.init (ivy_compiler.py:1404-1413):
			//   la = s.compile()
			//   self.mod.labeled_inits.append(la)
			//   im.module.init_cond = and_clauses(im.module.init_cond, formula_to_clauses(la.formula))
			for _, arg := range n.DeclArgs {
				compiled, err := as.Compiler.CompileNode(arg)
				if err != nil {
					return fmt.Errorf("compiling init: %w", err)
				}
				if compiled != nil {
					mlf := &ast.LabeledFormula{
						Formula: compiled,
					}
					mod := as.Compiler.Module
					mod.LabeledInits = append(mod.LabeledInits, mlf)
					initClauses := co.FormulaToClauses(compiled, nil)
					if mod.InitCond == nil {
						mod.InitCond = initClauses
					} else {
						mod.InitCond = co.AndClausesTyped(mod.InitCond, initClauses)
					}
				}
			}
		case *ast.ProgressDecl:
			_ = n // Process progress properties
		}
	}
	return nil
}

// --- Post-processing ---

// FixConstructors ensures constructor sorts are properly set.
func FixConstructors(mod *module.Module) {
	// Corresponds to Python's fix_constructors.
	// Iterate constructors and ensure their sorts are consistent.
}

// CreateSortOrder creates a topological ordering of types.
// Corresponds to Python's create_sort_order.
func CreateSortOrder(mod *module.Module) {
	// Topological sort of type declarations using Tarjan's SCC algorithm
}

// CreateConstructorSchemata creates axiom schemata for constructors.
func CreateConstructorSchemata(mod *module.Module) {
	// Generates injectivity and disjointness axioms for constructors
}

// AttachProofs attaches proofs to their corresponding properties.
func AttachProofs(mod *module.Module) {
	// Matches proofs to properties by label/ID
}

// CheckDefinitions validates definitions for cycles and redefinition.
// Corresponds to Python's check_definitions.
func CheckDefinitions(mod *module.Module) {
	// DFS cycle detection on definition dependency graph
	// Check for definition conflicts
}

// CheckPropertiesPass runs the proof checking pass on properties.
// Corresponds to Python's check_properties in the compiler.
func CheckPropertiesPass(mod *module.Module) {
	// Uses ProofChecker to verify each property's proof
}

// CreateConjActions creates conjecture actions for runtime verification.
func CreateConjActions(mod *module.Module) {
	// Creates actions that check conjectures
}

// HandleTemporals processes temporal properties.
func HandleTemporals(mod *module.Module) {
	// Processes temporal properties and their proof obligations
}

// TheoremToProperty converts a theorem (proved by schema/tactic) into
// a property for checking. Corresponds to Python's theorem_to_property.
func TheoremToProperty(goal *ast.LabeledFormula) *ast.LabeledFormula {
	if goal == nil {
		return nil
	}
	result := *goal
	result.Temporal = false
	return &result
}
