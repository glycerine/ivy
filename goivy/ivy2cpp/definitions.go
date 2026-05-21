package ivy2cpp

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type derivedDefinition struct {
	Name   string
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

func (g *Generator) emitDefinitionDecls(w *cppWriter) {
	defs := g.allDefinitions()
	for _, d := range defs {
		w.line(g.definitionSignature(d, false) + ";")
	}
	if len(defs) > 0 {
		w.blank()
	}
}

func (g *Generator) emitDefinitions(w *cppWriter) {
	for _, d := range g.allDefinitions() {
		w.open(g.definitionSignature(d, true) + " {")
		rhs, err := g.emitExpr(d.RHS)
		if err != nil {
			g.unsupported(w, "unsupported definition %s rhs: %s", d.Name, err.Error())
		} else {
			w.linef("return %s;", rhs)
		}
		w.close("")
		w.blank()
	}
}

func (g *Generator) definitionSignature(d derivedDefinition, qualified bool) string {
	className := ""
	fn, err := funName(d.Name)
	if err != nil {
		fn = varName(d.Name)
	}
	if qualified {
		fn = g.ClassName + "::" + fn
		className = g.ClassName
	}
	params := make([]string, 0, len(d.Params))
	for _, p := range d.Params {
		name := goivy.ExprName(p)
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("__arg%d", len(params))
		}
		params = append(params, g.cppQualifiedType(p.NodeSort(), className)+" "+varName(name))
	}
	return fmt.Sprintf("%s %s(%s)", g.cppQualifiedType(d.Sort, className), fn, strings.Join(params, ", "))
}
