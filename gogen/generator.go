package gogen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// Generator is the top-level Go code generator. It coordinates type
// emission, expression translation, and action code generation.
type Generator struct {
	// Module is the Ivy module being compiled.
	Module *module.Module

	// ExprEmitter handles formula/expression translation.
	Expr *ExprEmitter

	// Writer is the main code output buffer.
	Writer *CodeWriter

	// PackageName is the Go package name for the generated code.
	PackageName string
}

// NewGenerator creates a new Go code generator for the given module.
func NewGenerator(mod *module.Module, pkgName string) *Generator {
	return &Generator{
		Module:      mod,
		Expr:        NewExprEmitter(),
		Writer:      NewCodeWriter(),
		PackageName: pkgName,
	}
}

// Generate produces the complete Go source for the module.
func (g *Generator) Generate() (string, error) {
	w := g.Writer

	pkg := g.PackageName
	if pkg == "" {
		pkg = "main"
	}

	// 1. Package declaration.
	w.Linef("package %s", pkg)
	w.BlankLine()

	// 2. Imports.
	imports := g.collectImports()
	if len(imports) > 0 {
		w.OpenBlock("import (")
		for _, imp := range imports {
			w.Linef(`"%s"`, imp)
		}
		w.CloseBlock()
		w.BlankLine()
	}

	// 3. Sort/type declarations.
	if g.Module != nil {
		w.Line("// === Sort Types ===")
		w.BlankLine()
		EmitSortDecls(w, g.Module)
		w.BlankLine()
	}

	// 4. Quantifier helpers for finite sorts.
	g.emitQuantifierHelpers(w)

	// 5. Emit helper functions (ite, forAll, exists) referenced during expression emission.
	g.emitHelpers()

	// 6. State struct.
	g.emitStateStruct(w)

	// 7. NewState constructor.
	g.emitNewState(w)

	// 8. Actions as methods on *State.
	g.emitActionMethods(w)

	// 9. Init method.
	g.emitInit(w)

	// 10. Main function (only for package main).
	if pkg == "main" {
		g.emitMain(w)
	}

	return w.String(), nil
}

// emitHelpers emits all helper functions (forAll, exists, ite) that
// were referenced during expression translation.
func (g *Generator) emitHelpers() {
	w := g.Writer

	// Ite helpers.
	EmitIteHelper(w)
	w.BlankLine()

	// forAll and exists helpers for each referenced sort.
	for _, s := range g.Expr.HelperSorts {
		EmitForAllHelper(w, s)
		w.BlankLine()
		EmitExistsHelper(w, s)
		w.BlankLine()
	}
}

// collectImports determines what standard library imports are needed.
func (g *Generator) collectImports() []string {
	imports := []string{"fmt", "math/rand"}
	return imports
}

// emitQuantifierHelpers emits forAll_X and exists_X helpers for each
// enumerated sort in the module.
func (g *Generator) emitQuantifierHelpers(w *CodeWriter) {
	if g.Module == nil || g.Module.Sig == nil {
		return
	}
	hasAny := false
	for _, sortName := range g.Module.SortOrder {
		s, ok := g.Module.Sig.Sorts[sortName]
		if !ok {
			continue
		}
		if _, ok := s.(*lg.EnumeratedSort); !ok {
			continue
		}
		if !hasAny {
			w.Line("// === Quantifier Helpers ===")
			w.BlankLine()
			hasAny = true
		}
		EmitForAllHelper(w, s)
		w.BlankLine()
		EmitExistsHelper(w, s)
		w.BlankLine()
	}
}

// stateFields returns sorted (name, goType) pairs for all state variables.
func (g *Generator) stateFields() [][2]string {
	if g.Module == nil {
		return nil
	}
	var fields [][2]string
	seen := make(map[string]bool)

	// Relations
	relNames := sortedKeys(g.Module.Relations)
	for _, name := range relNames {
		s := g.Module.Relations[name]
		goName := goExportedName(name)
		goType := StateFieldType(name, s)
		if !seen[goName] {
			fields = append(fields, [2]string{goName, goType})
			seen[goName] = true
		}
	}

	// Functions
	fnNames := sortedKeys(g.Module.Functions)
	for _, name := range fnNames {
		s := g.Module.Functions[name]
		goName := goExportedName(name)
		goType := StateFieldType(name, s)
		if !seen[goName] {
			fields = append(fields, [2]string{goName, goType})
			seen[goName] = true
		}
	}

	return fields
}

// emitStateStruct emits the State struct.
func (g *Generator) emitStateStruct(w *CodeWriter) {
	w.Line("// === State ===")
	w.BlankLine()
	w.OpenBlock("type State struct {")
	for _, f := range g.stateFields() {
		w.Linef("%s %s", f[0], f[1])
	}
	w.CloseBlock()
	w.BlankLine()
}

// emitNewState emits the NewState() constructor with map initialization.
func (g *Generator) emitNewState(w *CodeWriter) {
	fields := g.stateFields()
	w.OpenBlock("func NewState() *State {")
	w.OpenBlock("return &State{")
	for _, f := range fields {
		if strings.HasPrefix(f[1], "map[") {
			w.Linef("%s: make(%s),", f[0], f[1])
		}
	}
	w.CloseBlock()
	w.CloseBlock()
	w.BlankLine()
}

// emitActionMethods emits each named action as a method on *State.
func (g *Generator) emitActionMethods(w *CodeWriter) {
	if g.Module == nil {
		return
	}
	w.Line("// === Actions ===")
	w.BlankLine()

	actionNames := sortedKeysIface(g.Module.Actions)

	for _, name := range actionNames {
		actIface := g.Module.Actions[name]
		act, ok := actIface.(actions.Action)
		if !ok {
			w.Linef("// skipped action %q: not an actions.Action", name)
			w.BlankLine()
			continue
		}
		goName := GoExportedIdentifier(name)

		// Build parameter list from formal params.
		params := formatFormalParams(act.GetFormalParams())
		returns := formatFormalReturns(act.GetFormalReturns())

		sig := fmt.Sprintf("func (s *State) %s(%s)", goName, params)
		if returns != "" {
			sig += " " + returns
		}
		w.OpenBlock(sig + " {")
		emitter := NewActionEmitter(g, w)
		emitter.EmitAction(act)
		w.CloseBlock()
		w.BlankLine()
	}
}

// emitInit emits the Init() method from module initializers.
func (g *Generator) emitInit(w *CodeWriter) {
	w.Line("// === Initialization ===")
	w.BlankLine()
	w.OpenBlock("func (s *State) Init() {")

	if g.Module != nil {
		for _, init := range g.Module.Initializers {
			act, ok := init.Action.(actions.Action)
			if !ok {
				w.Linef("// skipped initializer %q: not an actions.Action", init.Name)
				continue
			}
			w.Linef("// initializer: %s", init.Name)
			emitter := NewActionEmitter(g, w)
			emitter.EmitAction(act)
		}
	}

	w.CloseBlock()
	w.BlankLine()
}

// emitMain emits a main() function.
func (g *Generator) emitMain(w *CodeWriter) {
	w.Line("// === Main ===")
	w.BlankLine()
	w.OpenBlock("func main() {")
	w.Line("s := NewState()")
	w.Line("s.Init()")
	w.Line(`fmt.Println("Initialized")`)
	// Suppress unused import warning for rand.
	w.Line("_ = rand.Intn")
	w.CloseBlock()
}

// EmitStateStruct emits a Go struct declaration for the module's state,
// using the Relations and Functions maps. (Legacy API for direct struct name.)
func (g *Generator) EmitStateStruct(name string) {
	w := g.Writer
	w.OpenBlock(fmt.Sprintf("type %s struct {", name))
	if g.Module != nil {
		for symName, s := range g.Module.Relations {
			fieldName := goExportedName(symName)
			fieldType := StateFieldType(symName, s)
			w.Linef("%s %s", fieldName, fieldType)
		}
		for symName, s := range g.Module.Functions {
			fieldName := goExportedName(symName)
			fieldType := StateFieldType(symName, s)
			w.Linef("%s %s", fieldName, fieldType)
		}
	}
	w.CloseBlock()
}

// formatFormalParams formats formal parameters as a Go parameter list.
func formatFormalParams(params []*lg.Symbol) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		goName := GoIdentifier(p.Name)
		goType := GoType(p.CSort)
		parts[i] = goName + " " + goType
	}
	return strings.Join(parts, ", ")
}

// formatFormalReturns formats formal return parameters.
func formatFormalReturns(params []*lg.Symbol) string {
	if len(params) == 0 {
		return ""
	}
	if len(params) == 1 {
		return GoType(params[0].CSort)
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = GoType(p.CSort)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// sortedKeys returns sorted keys from a map[string]lg.Sort.
func sortedKeys(m map[string]lg.Sort) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedKeysIface returns sorted keys from a map[string]interface{}.
func sortedKeysIface(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// RangeValues is a runtime helper that generates a slice of ints
// from lo to hi (inclusive). This function is emitted into generated code.
func EmitRangeValuesHelper(w *CodeWriter) {
	w.OpenBlock("func rangeValues(lo, hi int) []int {")
	w.Line("r := make([]int, 0, hi-lo+1)")
	w.OpenBlock("for i := lo; i <= hi; i++ {")
	w.Line("r = append(r, i)")
	w.CloseBlock()
	w.Line("return r")
	w.CloseBlock()
}

// GoTypeForNode returns the Go type string for an arbitrary logic Node's sort.
func GoTypeForNode(n lg.Node) string {
	if n == nil {
		return "interface{}"
	}
	return GoType(n.NodeSort())
}
