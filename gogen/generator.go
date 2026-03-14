package gogen

import (
	"fmt"

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

	// Package declaration.
	w.Linef("package %s", g.PackageName)
	w.BlankLine()

	// Sort/type declarations.
	if g.Module != nil {
		EmitSortDecls(w, g.Module)
	}

	// Emit helper functions that were referenced during expression emission.
	g.emitHelpers()

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

// EmitStateStruct emits a Go struct declaration for the module's state,
// using the Relations and Functions maps.
func (g *Generator) EmitStateStruct(name string) {
	w := g.Writer
	w.OpenBlock(fmt.Sprintf("type %s struct {", name))
	if g.Module != nil {
		for symName, sort := range g.Module.Relations {
			fieldName := goExportedName(symName)
			fieldType := StateFieldType(symName, sort)
			w.Linef("%s %s", fieldName, fieldType)
		}
		for symName, sort := range g.Module.Functions {
			fieldName := goExportedName(symName)
			fieldType := StateFieldType(symName, sort)
			w.Linef("%s %s", fieldName, fieldType)
		}
	}
	w.CloseBlock()
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
