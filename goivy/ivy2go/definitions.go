package ivy2go

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// definitions.go is the literal port of ivy2cpp/definitions.go.
// Definitional axioms (`derived sym(X) = expr`) lower into Go methods
// on *State that compute and return the value. The synthetic
// AssignAction built here is fed through `emitSomeAction` so it picks
// up the standard signature / return handling.

type derivedDefinition struct {
	Name   string
	Head   goivy.Expr
	Params []goivy.Expr
	RHS    goivy.Expr
	Sort   goivy.Sort
}

// derivedDefinitions mirrors ivy2cpp/definitions.go:17.
func (g *Generator) derivedDefinitions() []derivedDefinition {
	if g == nil || g.Mod == nil {
		return nil
	}
	defs := make([]derivedDefinition, 0, len(g.Mod.Definitions))
	for _, lf := range g.Mod.Definitions {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		dd, ok := newDerivedDefinition(def)
		if ok {
			defs = append(defs, dd)
		}
	}
	return defs
}

// nativeDefinitions mirrors ivy2cpp/definitions.go:35.
func (g *Generator) nativeDefinitions() []derivedDefinition {
	if g == nil || g.Mod == nil {
		return nil
	}
	defs := make([]derivedDefinition, 0, len(g.Mod.NativeDefinitions))
	for _, lf := range g.Mod.NativeDefinitions {
		def, ok := lf.Formula.(*goivy.LogicDefinition)
		if !ok || def == nil {
			continue
		}
		dd, ok := newDerivedDefinition(def)
		if ok {
			defs = append(defs, dd)
		}
	}
	return defs
}

// allDefinitions mirrors ivy2cpp/definitions.go:53.
func (g *Generator) allDefinitions() []derivedDefinition {
	defs := g.derivedDefinitions()
	return append(defs, g.nativeDefinitions()...)
}

// newDerivedDefinition mirrors ivy2cpp/definitions.go:58.
func newDerivedDefinition(def *goivy.LogicDefinition) (derivedDefinition, bool) {
	if def == nil || def.Lhs == nil || def.Rhs == nil {
		return derivedDefinition{}, false
	}
	name := goivy.ExprName(def.Defines())
	if name == "" {
		return derivedDefinition{}, false
	}
	var params []goivy.Expr
	if app, ok := def.Lhs.(*goivy.Apply); ok {
		params = append(params, app.Terms...)
	}
	return derivedDefinition{
		Name:   name,
		Head:   def.Defines(),
		Params: params,
		RHS:    def.Rhs,
		Sort:   def.Lhs.NodeSort(),
	}, true
}

// definitionNames mirrors ivy2cpp/definitions.go:79.
func (g *Generator) definitionNames() map[string]bool {
	names := map[string]bool{}
	for _, d := range g.allDefinitions() {
		names[d.Name] = true
	}
	return names
}

// definitionByName mirrors ivy2cpp/definitions.go:91.
func (g *Generator) definitionByName(name string) (derivedDefinition, bool) {
	if name == "" {
		return derivedDefinition{}, false
	}
	for _, d := range g.allDefinitions() {
		if d.Name == name {
			return d, true
		}
	}
	return derivedDefinition{}, false
}

// derivedActionFor mirrors ivy2cpp/definitions.go:127. Builds the
// synthetic AssignAction used to emit a derived definition as a Go
// method via emitSomeAction.
func (g *Generator) derivedActionFor(d derivedDefinition) (string, goivy.Action, error) {
	retval := goivy.NewConst("ret:val", d.Sort)

	subs := map[goivy.NodeKey]goivy.Expr{}
	skolems := make([]*goivy.Const, 0, len(d.Params))
	for _, p := range d.Params {
		pname := goivy.ExprName(p)
		if pname == "" {
			return "", nil, fmt.Errorf("derived definition %s: parameter has empty name (%T)", d.Name, p)
		}
		skol := goivy.NewConst("fml:"+pname, p.NodeSort())
		subs[goivy.Key(p)] = skol
		skolems = append(skolems, skol)
	}

	rhs := d.RHS
	if len(subs) > 0 {
		r, err := goivy.Substitute(d.RHS, subs)
		if err != nil {
			return "", nil, fmt.Errorf("derived definition %s: substitute: %w", d.Name, err)
		}
		rhs = r
	}

	act := goivy.NewAssignAction(retval, rhs)
	act.SetFormalParams(skolems)
	act.SetFormalReturns([]*goivy.Const{retval})
	return d.Name, act, nil
}

// emitDefinitionDecls is the Go counterpart of cpp's same-named
// helper. Go has no forward declarations, so this stays a no-op.
func (g *Generator) emitDefinitionDecls(w *goWriter) { _ = w }

// emitDefinitions mirrors ivy2cpp/definitions.go:172 emitDefinitions.
// Wired into emitActions so derived-definition methods land in the
// same actions.go stream as ordinary actions.
func (g *Generator) emitDefinitions(w *goWriter) {
	for _, d := range g.allDefinitions() {
		name, act, err := g.derivedActionFor(d)
		if err != nil {
			g.errs = append(g.errs, err)
			continue
		}
		g.emitSomeAction(w, name, act)
	}
}
