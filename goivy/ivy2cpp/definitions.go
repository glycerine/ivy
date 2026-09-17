package ivy2cpp

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

type derivedDefinition struct {
	Name   string
	Head   goivy.Expr
	Params []goivy.Expr
	RHS    goivy.Expr
	Sort   goivy.Sort
}

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

func (g *Generator) allDefinitions() []derivedDefinition {
	defs := g.derivedDefinitions()
	return append(defs, g.nativeDefinitions()...)
}

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

func (g *Generator) definitionNames() map[string]bool {
	names := map[string]bool{}
	for _, d := range g.allDefinitions() {
		names[d.Name] = true
	}
	return names
}

// definitionByName looks up a derived (or native) definition by its
// head symbol name. Returns ({}, false) when no such definition exists.
// Used by matchExtensionalBoundExprs to unfold definitions while
// hunting for an extensional-relation atom (Python ivy_to_cpp.py:3370).
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

func (g *Generator) isDefinitionName(name string) bool {
	if name == "" {
		return false
	}
	for _, d := range g.allDefinitions() {
		if d.Name == name {
			return true
		}
	}
	return false
}

// derivedActionFor builds the synthetic AssignAction used to emit a
// derived definition as a C++ method. Mirrors Python emit_derived
// (ivy_to_cpp.py:1351-1362) up to the emit_some_action call:
//
//   - retval = Symbol("ret:val", sort)
//   - skolemize bound Variables X_i → Constants "fml:X_i"
//   - substitute Variables→Constants in RHS
//   - action = AssignAction(retval, rhs); formal_params = skolems;
//     formal_returns = [retval].
//
// Returns the method name and the synthetic action ready for
// emitSomeAction / emitMethodDeclLine.
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

func (g *Generator) emitDefinitionDecls(w *cppWriter) {
	defs := g.allDefinitions()
	for _, d := range defs {
		name, act, err := g.derivedActionFor(d)
		if err != nil {
			g.errs = append(g.errs, err)
			continue
		}
		g.emitMethodDeclLine(w, name, act)
	}
}

func (g *Generator) emitDefinitions(w *cppWriter) {
	for _, d := range g.allDefinitions() {
		name, act, err := g.derivedActionFor(d)
		if err != nil {
			g.errs = append(g.errs, err)
			continue
		}
		g.emitSomeAction(w, name, act)
	}
}
