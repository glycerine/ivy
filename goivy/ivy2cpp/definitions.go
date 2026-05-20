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
		Sort:   def.Rhs.NodeSort(),
	}, true
}

func (g *Generator) definitionNames() map[string]bool {
	names := map[string]bool{}
	for _, d := range g.derivedDefinitions() {
		names[d.Name] = true
	}
	return names
}

func (g *Generator) isDefinitionName(name string) bool {
	if name == "" {
		return false
	}
	for _, d := range g.derivedDefinitions() {
		if d.Name == name {
			return true
		}
	}
	return false
}

func (g *Generator) emitDefinitionDecls(w *cppWriter) {
	for _, d := range g.derivedDefinitions() {
		w.line(g.definitionSignature(d, false) + ";")
	}
	if len(g.derivedDefinitions()) > 0 {
		w.blank()
	}
}

func (g *Generator) emitDefinitions(w *cppWriter) {
	for _, d := range g.derivedDefinitions() {
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
	typeName := cppType
	fn, err := funName(d.Name)
	if err != nil {
		fn = varName(d.Name)
	}
	if qualified {
		typeName = func(s goivy.Sort) string { return cppQualifiedType(s, g.ClassName) }
		fn = g.ClassName + "::" + fn
	}
	params := make([]string, 0, len(d.Params))
	for _, p := range d.Params {
		name := goivy.ExprName(p)
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("__arg%d", len(params))
		}
		params = append(params, typeName(p.NodeSort())+" "+varName(name))
	}
	return fmt.Sprintf("%s %s(%s)", typeName(d.Sort), fn, strings.Join(params, ", "))
}
