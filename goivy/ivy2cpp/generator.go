package ivy2cpp

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type Config struct {
	Target    string
	ClassName string
	MainName  string
	OutDir    string
	EmitMain  bool
	Trace     bool
	Stdafx    bool
}

type Output struct {
	Header    string
	Impl      string
	BaseName  string
	ClassName string
}

type Generator struct {
	Mod       *goivy.Module
	Config    Config
	BaseName  string
	ClassName string

	header cppWriter
	impl   cppWriter
	tempID int
}

func Generate(mod *goivy.Module, cfg Config) (*Output, error) {
	if mod == nil {
		return nil, fmt.Errorf("ivy2cpp: nil module")
	}
	target := cfg.Target
	if target == "" {
		target = "impl"
	}
	if target != "impl" && target != "repl" {
		return nil, fmt.Errorf("ivy2cpp: target %q is not supported in v1", target)
	}
	base := moduleBaseName(mod)
	className := cfg.ClassName
	if className == "" {
		className = varName(base)
	}
	g := &Generator{
		Mod:       mod,
		Config:    cfg,
		BaseName:  base,
		ClassName: className,
	}
	if err := g.generate(); err != nil {
		return nil, err
	}
	return &Output{
		Header:    g.header.String(),
		Impl:      g.impl.String(),
		BaseName:  base,
		ClassName: className,
	}, nil
}

func moduleBaseName(mod *goivy.Module) string {
	name := strings.TrimSpace(mod.Name)
	if name == "" {
		name = "ivy"
	}
	name = filepath.Base(name)
	ext := filepath.Ext(name)
	if ext != "" {
		name = strings.TrimSuffix(name, ext)
	}
	if name == "" {
		return "ivy"
	}
	return name
}

func (g *Generator) generate() error {
	g.emitHeader()
	g.emitImpl()
	return nil
}

func (g *Generator) emitHeader() {
	w := &g.header
	w.line("#pragma once")
	w.line("#include <cstdint>")
	w.line("#include <cstdlib>")
	w.line("#include <initializer_list>")
	w.line("#include <iostream>")
	w.line("#include <map>")
	w.line("#include <string>")
	w.line("#include <tuple>")
	w.line("#include <vector>")
	w.blank()
	w.open(fmt.Sprintf("class %s {", g.ClassName))
	w.line("public:")
	w.indent++
	w.linef("typedef %s ivy_class;", g.ClassName)
	w.linef("%s();", g.ClassName)
	w.linef("virtual ~%s();", g.ClassName)
	w.line("virtual void ivy_assert(bool truth, const char *msg);")
	w.line("virtual void ivy_assume(bool truth, const char *msg);")
	w.line("int ___ivy_choose(int rng, const char *name, int id);")
	w.line("void __init();")
	w.blank()
	g.emitSortDecls(w)
	g.emitStateDecls(w)
	g.emitMethodDecls(w)
	w.indent--
	w.close(";")
}

func (g *Generator) emitImpl() {
	w := &g.impl
	if g.Config.Stdafx {
		w.line(`#include "stdafx.h"`)
	}
	w.linef(`#include "%s.h"`, g.BaseName)
	w.blank()
	w.linef("%s::%s() {}", g.ClassName, g.ClassName)
	w.linef("%s::~%s() {}", g.ClassName, g.ClassName)
	w.blank()
	w.open(fmt.Sprintf("void %s::ivy_assert(bool truth, const char *msg) {", g.ClassName))
	w.open("if (!truth) {")
	w.line(`std::cerr << msg << ": assertion failed" << std::endl;`)
	w.line("std::abort();")
	w.close("")
	w.close("")
	w.blank()
	w.open(fmt.Sprintf("void %s::ivy_assume(bool truth, const char *msg) {", g.ClassName))
	w.open("if (!truth) {")
	w.line(`std::cerr << msg << ": assumption failed" << std::endl;`)
	w.line("std::abort();")
	w.close("")
	w.close("")
	w.blank()
	w.open(fmt.Sprintf("int %s::___ivy_choose(int rng, const char *name, int id) {", g.ClassName))
	w.line("(void)rng;")
	w.line("(void)name;")
	w.line("(void)id;")
	w.line("return 0;")
	w.close("")
	w.blank()
	g.emitInit(w)
	g.emitMethods(w)
	if g.Config.Target == "repl" {
		g.emitRepl(w)
	}
}

func (g *Generator) emitSortDecls(w *cppWriter) {
	if g.Mod.Sig == nil {
		return
	}
	for _, name := range g.Mod.SortOrder {
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		switch st := s.(type) {
		case *goivy.LogicEnumeratedSort:
			vals := make([]string, len(st.Extension))
			for i, v := range st.Extension {
				vals[i] = varName(v)
			}
			w.linef("enum %s { %s };", varName(st.Name), strings.Join(vals, ", "))
		case *goivy.RangeSort:
			if st.Name != "" {
				w.linef("typedef long long %s;", varName(st.Name))
			}
		case *goivy.UninterpretedSort:
			w.linef("typedef long long %s;", varName(st.Name))
		}
	}
	destructorNames := insMapKeys(g.Mod.SortDestructors)
	if len(destructorNames) > 0 {
		w.blank()
		for _, name := range destructorNames {
			w.open(fmt.Sprintf("struct %s {", varName(name)))
			for _, d := range g.Mod.SortDestructors.Get(name) {
				if fs, ok := d.CSort.(*goivy.LogicFunctionSort); ok {
					w.linef("%s %s;", cppType(fs.Range()), memName(d.Name))
				}
			}
			w.close(";")
		}
	}
	w.blank()
}

func (g *Generator) emitStateDecls(w *cppWriter) {
	for _, sym := range g.stateSymbols() {
		w.linef("%s %s;", cppType(sym.Sort), varName(sym.Name))
	}
	if len(g.stateSymbols()) > 0 {
		w.blank()
	}
}

type stateSymbol struct {
	Name string
	Sort goivy.Sort
}

func (g *Generator) stateSymbols() []stateSymbol {
	var out []stateSymbol
	seen := map[string]bool{}
	add := func(name string, s goivy.Sort) {
		if name == "" || seen[name] {
			return
		}
		if g.Mod.Sig != nil && g.Mod.Sig.Constructors[name] {
			return
		}
		seen[name] = true
		out = append(out, stateSymbol{Name: name, Sort: s})
	}
	if g.Mod.Relations != nil {
		for k, v := range g.Mod.Relations.All() {
			add(k, v)
		}
	}
	if g.Mod.Functions != nil {
		for k, v := range g.Mod.Functions.All() {
			add(k, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (g *Generator) emitMethodDecls(w *cppWriter) {
	if g.Mod.Actions == nil {
		return
	}
	initActions := g.initialMixinActionNames()
	for name, act := range g.Mod.Actions.All() {
		if initActions[name] {
			continue
		}
		w.line(g.methodSignature(name, act, false) + ";")
	}
}

func (g *Generator) methodSignature(name string, act goivy.Action, qualified bool) string {
	ret := "void"
	typeName := cppType
	if qualified {
		typeName = func(s goivy.Sort) string { return cppQualifiedType(s, g.ClassName) }
	}
	returns := act.GetFormalReturns()
	if len(returns) == 1 {
		ret = typeName(returns[0].CSort)
	}
	fn, err := funName(name)
	if err != nil {
		fn = varName(name)
	}
	if qualified {
		fn = g.ClassName + "::" + fn
	}
	var params []string
	for _, p := range act.GetFormalParams() {
		params = append(params, typeName(p.CSort)+" "+varName(p.Name))
	}
	if len(returns) > 1 {
		for _, r := range returns {
			params = append(params, typeName(r.CSort)+" &"+varName(r.Name))
		}
	}
	return fmt.Sprintf("%s %s(%s)", ret, fn, strings.Join(params, ", "))
}

func (g *Generator) emitInit(w *cppWriter) {
	w.open(fmt.Sprintf("void %s::__init() {", g.ClassName))
	if len(g.Mod.InitialActions) > 0 {
		for _, act := range g.Mod.InitialActions {
			g.emitAction(w, act)
		}
		w.close("")
		w.blank()
		return
	}
	for _, na := range g.Mod.Initializers {
		if act, ok := na.Action.(goivy.Action); ok {
			g.emitAction(w, act)
		}
	}
	if len(g.Mod.Initializers) == 0 && g.Mod.Actions != nil && g.Mod.Mixins != nil {
		for _, mixin := range g.Mod.Mixins.Get("init") {
			if act, ok := g.Mod.Actions.Get2(mixin.Mixer()); ok {
				g.emitAction(w, act)
			}
		}
	}
	w.close("")
	w.blank()
}

func (g *Generator) emitMethods(w *cppWriter) {
	if g.Mod.Actions == nil {
		return
	}
	initActions := g.initialMixinActionNames()
	for name, act := range g.Mod.Actions.All() {
		if initActions[name] {
			continue
		}
		w.open(g.methodSignature(name, act, true) + " {")
		returns := act.GetFormalReturns()
		if len(returns) == 1 && !formalListContains(act.GetFormalParams(), returns[0]) {
			w.linef("%s %s = %s;", cppType(returns[0].CSort), varName(returns[0].Name), cppZeroValue(returns[0].CSort))
		}
		g.emitAction(w, act)
		if len(returns) == 1 {
			w.linef("return %s;", varName(returns[0].Name))
		}
		w.close("")
		w.blank()
	}
}

func formalListContains(formals []*goivy.Const, target *goivy.Const) bool {
	if target == nil {
		return false
	}
	for _, f := range formals {
		if f == target || (f != nil && f.Name == target.Name) {
			return true
		}
	}
	return false
}

func (g *Generator) emitRepl(w *cppWriter) {
	mainName := g.Config.MainName
	if mainName == "" {
		mainName = "main"
	}
	w.line("static void ivy2cpp_dispatch(" + g.ClassName + " &ivy, const std::string &action) {")
	w.indent++
	initActions := g.initialMixinActionNames()
	for name := range g.Mod.PublicActions.All() {
		if initActions[name] {
			continue
		}
		username := strings.TrimPrefix(name, "ext:")
		fn, _ := funName(name)
		args := g.replDispatchArgs(name)
		act, ok := g.Mod.Actions.Get2(name)
		if ok && len(act.GetFormalReturns()) > 1 {
			w.open(fmt.Sprintf(`if (action == "%s") {`, username))
			for _, r := range act.GetFormalReturns() {
				w.linef("%s %s = %s;", cppType(r.CSort), varName(r.Name), cppZeroValue(r.CSort))
				args = append(args, varName(r.Name))
			}
			w.linef("ivy.%s(%s);", fn, strings.Join(args, ", "))
			w.line("return;")
			w.close("")
			continue
		}
		w.linef(`if (action == "%s") { ivy.%s(%s); return; }`, username, fn, strings.Join(args, ", "))
	}
	w.line(`std::cerr << "undefined action: " << action << std::endl;`)
	w.indent--
	w.line("}")
	w.blank()
	w.open(fmt.Sprintf("int %s(int argc, char **argv) {", mainName))
	w.linef("%s ivy;", g.ClassName)
	w.line("ivy.__init();")
	w.line("(void)argc;")
	w.line("(void)argv;")
	w.line("return 0;")
	w.close("")
}

func (g *Generator) replDispatchArgs(name string) []string {
	if g.Mod == nil || g.Mod.Actions == nil {
		return nil
	}
	act, ok := g.Mod.Actions.Get2(name)
	if !ok {
		return nil
	}
	var args []string
	for _, p := range act.GetFormalParams() {
		args = append(args, cppZeroValueInScope(p.CSort, g.ClassName))
	}
	return args
}

func (g *Generator) initialMixinActionNames() map[string]bool {
	out := map[string]bool{}
	if g.Mod == nil || g.Mod.Mixins == nil {
		return out
	}
	for _, mixin := range g.Mod.Mixins.Get("init") {
		out[mixin.Mixer()] = true
	}
	return out
}

func insMapKeys[V any](m *goivy.InsMap[string, V]) []string {
	if m == nil {
		return nil
	}
	var keys []string
	for k := range m.All() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
