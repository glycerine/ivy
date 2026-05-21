package ivy2cpp

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type Config struct {
	Target          string
	RequestedTarget string
	ClassName       string
	MainName        string
	OutDir          string
	Compiler        string
	TestIters       string
	TestRuns        string
	Build           bool
	EmitMain        bool
	Trace           bool
	Stdafx          bool
}

type Output struct {
	Header          string
	Impl            string
	BaseName        string
	ClassName       string
	Target          string
	EffectiveTarget string
	EmitMain        bool
	Config          Config
	ExtraFiles      map[string]string
	LibSpecs        []string
}

type Generator struct {
	Mod       *goivy.Module
	Config    Config
	BaseName  string
	ClassName string

	header cppWriter
	impl   cppWriter
	tempID int

	exprAliases    map[string]goivy.Expr
	currentReturns []*goivy.Const
	errs           []error
}

func Generate(mod *goivy.Module, cfg Config) (*Output, error) {
	if mod == nil {
		return nil, fmt.Errorf("ivy2cpp: nil module")
	}
	if cfg.Target == "" && cfg.RequestedTarget == "" {
		cfg.Target = "impl"
	}
	var err error
	cfg, _, err = normalizeConfig(cfg)
	if err != nil {
		return nil, err
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
		Header:          g.header.String(),
		Impl:            g.impl.String(),
		BaseName:        base,
		ClassName:       className,
		Target:          cfg.RequestedTarget,
		EffectiveTarget: cfg.Target,
		EmitMain:        cfg.EmitMain,
		Config:          cfg,
		LibSpecs:        moduleLibSpecs(mod),
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
	if err := g.emitHeader(); err != nil {
		return err
	}
	if err := g.emitImpl(); err != nil {
		return err
	}
	return errors.Join(g.errs...)
}

func (g *Generator) unsupported(w *cppWriter, format string, args ...any) {
	err := fmt.Errorf("ivy2cpp: "+format, args...)
	g.errs = append(g.errs, err)
	w.linef("/* %s */", escapeComment(err.Error()))
}

func (g *Generator) nextTemp(prefix string) string {
	g.tempID++
	return fmt.Sprintf("%s%d", prefix, g.tempID)
}

func (g *Generator) emitHeader() error {
	w := &g.header
	w.line("#pragma once")
	g.emitRuntimeHeaderPreamble(w)
	if err := g.emitNativeBlocks(w, "header"); err != nil {
		return err
	}
	w.blank()
	w.open(fmt.Sprintf("class %s {", g.ClassName))
	w.line("public:")
	w.indent++
	w.linef("typedef %s ivy_class;", g.ClassName)
	g.emitRuntimeClassMembers(w)
	if g.Config.Target != "gen" {
		w.line("virtual void ivy_assert(bool truth, const char *msg) {}")
		w.line("virtual void ivy_assume(bool truth, const char *msg) {}")
		w.line("virtual void ivy_check_progress(int guarantee_ticks, int assume_ticks) {}")
	}
	w.line("int ___ivy_choose(int rng, const char *name, int id);")
	w.line("void __init();")
	w.line("void __tick(int timeout);")
	w.blank()
	g.emitSortDecls(w)
	w.line(g.constructorSignature(false) + ";")
	w.blank()
	g.emitCTupleDecls(w)
	g.emitCardinalityDecls(w)
	g.emitStateDecls(w)
	g.emitProgressCounterDecls(w)
	if err := g.emitNativeBlocks(w, "member"); err != nil {
		return err
	}
	g.emitDefinitionDecls(w)
	g.emitMethodDecls(w)
	w.indent--
	w.close(";")
	return nil
}

func (g *Generator) emitImpl() error {
	w := &g.impl
	if g.Config.Stdafx {
		w.line(`#include "stdafx.h"`)
	}
	w.linef(`#include "%s.h"`, g.BaseName)
	w.blank()
	g.emitRuntimeImplPreamble(w)
	if err := g.emitNativeBlocks(w, "impl"); err != nil {
		return err
	}
	if g.usesZ3() {
		if err := g.emitZ3Support(w); err != nil {
			return err
		}
	} else {
		g.emitCPPTypeImpls(w)
	}
	w.open(g.constructorSignature(true) + " {")
	g.emitRuntimeConstructorPrelude(w)
	g.emitConstructorParamAssignments(w)
	g.emitCardinalityInitializers(w)
	g.emitProgressCounterInitializers(w)
	if err := g.emitNativeBlocks(w, "init"); err != nil {
		return err
	}
	if !g.usesZ3() {
		if err := g.emitOneInitialState(w); err != nil {
			return err
		}
	}
	w.close("")
	g.emitRuntimeMethods(w)
	g.emitInit(w)
	g.emitDefinitions(w)
	g.emitMethods(w)
	g.emitTick(w)
	if g.runtimeUsesReplSubclass() {
		g.emitRuntimeReplSubclass(w)
	}
	switch g.Config.Target {
	case "repl":
		g.emitReplSupport(w)
		if g.Config.EmitMain {
			g.emitReplMain(w)
		}
	case "test":
		if g.Config.EmitMain {
			g.emitTestMain(w)
		}
	case "gen":
		if g.Config.EmitMain {
			g.emitGenMain(w)
		}
	}
	return nil
}

func (g *Generator) constructorSignature(qualified bool) string {
	name := g.ClassName
	className := ""
	if qualified {
		name = g.ClassName + "::" + g.ClassName
		className = g.ClassName
	}
	params := make([]string, 0, len(g.Mod.Params))
	for _, p := range g.Mod.Params {
		params = append(params, g.cppStorageDecl(p.Name, p.CSort, className))
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(params, ", "))
}

func (g *Generator) emitConstructorParamAssignments(w *cppWriter) {
	for _, p := range g.Mod.Params {
		name := varName(p.Name)
		w.linef("this->%s = %s;", name, name)
	}
}

func (g *Generator) emitSortDecls(w *cppWriter) {
	if g.Mod.Sig == nil {
		return
	}
	emittedDestructorStructs := map[string]bool{}
	emittedVariantSupers := map[string]bool{}
	emittedIntClass := false
	for _, name := range g.Mod.SortOrder {
		if g.isVariantSuperName(name) {
			continue
		}
		if nt, ok := g.nativeTypeForSort(name); ok {
			g.emitNativeTypeDecl(w, name, nt)
			continue
		}
		if _, ok := g.Mod.SortDestructors.Get2(name); ok {
			g.emitDestructorStruct(w, name)
			emittedDestructorStructs[name] = true
			continue
		}
		if g.isVariantSubtypeName(name) {
			g.emitVariantLeafStruct(w, name)
			continue
		}
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		if it, ok := g.cppInterpType(s); ok {
			if it.Kind == cppInterpIntBV && !emittedIntClass {
				g.emitIntClassDecl(w)
				emittedIntClass = true
			}
			g.emitCPPTypeDecl(w, s, it)
			continue
		}
		switch st := s.(type) {
		case *goivy.LogicEnumeratedSort:
			if isNumericEnum(st) {
				continue
			}
			vals := make([]string, len(st.Extension))
			for i, v := range st.Extension {
				vals[i] = varName(v)
			}
			w.linef("enum %s { %s };", varName(st.Name), strings.Join(vals, ", "))
		case *goivy.RangeSort:
			if st.Name != "" {
				w.linef("typedef %s %s;", g.cppType(st), varName(st.Name))
			}
		case *goivy.UninterpretedSort:
			if typ := g.cppType(st); typ != "int" {
				w.linef("typedef %s %s;", typ, varName(st.Name))
			}
		}
	}
	destructorNames := insMapKeys(g.Mod.SortDestructors)
	if len(destructorNames) > 0 {
		w.blank()
		for _, name := range destructorNames {
			if emittedDestructorStructs[name] {
				continue
			}
			g.emitDestructorStruct(w, name)
		}
	}
	for _, name := range g.Mod.SortOrder {
		if !g.isVariantSuperName(name) || emittedVariantSupers[name] {
			continue
		}
		g.emitVariantSuperStruct(w, name)
		emittedVariantSupers[name] = true
	}
	w.blank()
}

func (g *Generator) emitDestructorStruct(w *cppWriter, name string) {
	destructors := g.Mod.SortDestructors.Get(name)
	w.open(fmt.Sprintf("struct %s {", varName(name)))
	for _, d := range destructors {
		if fs, ok := d.CSort.(*goivy.LogicFunctionSort); ok {
			domain := fs.Domain()
			if len(domain) > 0 {
				domain = domain[1:]
			}
			w.linef("%s;", g.cppFunctionStorageDecl(memName(d.Name), domain, fs.Range(), ""))
		}
	}
	g.emitDestructorStructComparators(w, name, destructors)
	g.emitDestructorStructWriter(w, name, destructors)
	w.close(";")
}

func (g *Generator) emitCTupleDecls(w *cppWriter) {
	for _, dom := range g.cppCTuples() {
		name := cppCTupleLocalNameWith(g, dom)
		w.open(fmt.Sprintf("struct %s {", name))
		for i, s := range dom {
			w.linef("%s arg%d;", g.cppType(s), i)
		}
		w.linef("%s() {}", name)
		params := make([]string, len(dom))
		inits := make([]string, len(dom))
		for i, s := range dom {
			params[i] = fmt.Sprintf("const %s &arg%d", g.cppType(s), i)
			inits[i] = fmt.Sprintf("arg%d(arg%d)", i, i)
		}
		w.linef("%s(%s) : %s {}", name, strings.Join(params, ", "), strings.Join(inits, ", "))
		hashParts := make([]string, len(dom))
		for i, s := range dom {
			hashParts[i] = fmt.Sprintf("hash_space::hash<%s>()(arg%d)", cppHashType(g, s), i)
		}
		w.linef("size_t __hash() const { return %s; }", strings.Join(hashParts, " + "))
		w.open(fmt.Sprintf("bool operator==(const %s &other) const {", name))
		eqParts := make([]string, len(dom))
		for i := range dom {
			eqParts[i] = fmt.Sprintf("arg%d == other.arg%d", i, i)
		}
		w.linef("return %s;", strings.Join(eqParts, " && "))
		w.close("")
		w.close(";")
		w.blank()
	}
}

func (g *Generator) emitVariantSuperStruct(w *cppWriter, name string) {
	variants := g.Mod.Variants[name]
	typeName := varName(name)
	w.open(fmt.Sprintf("struct %s {", typeName))
	w.line("int __tag;")
	for _, v := range variants {
		vname := varName(sortName(v))
		if vname == "" {
			continue
		}
		w.linef("%s __%s;", vname, vname)
	}
	w.linef("%s() : __tag(-1) {}", typeName)
	for i, v := range variants {
		vname := varName(sortName(v))
		if vname == "" {
			continue
		}
		w.linef("%s(const %s &value) : __tag(%d), __%s(value) {}", typeName, vname, i, vname)
	}
	g.emitVariantSuperComparators(w, typeName, variants)
	g.emitVariantSuperWriter(w, typeName, variants)
	w.close(";")
}

func (g *Generator) emitVariantLeafStruct(w *cppWriter, name string) {
	typeName := varName(name)
	w.open(fmt.Sprintf("struct %s {", typeName))
	w.line("long long __value;")
	w.linef("%s(long long value = 0) : __value(value) {}", typeName)
	w.line("operator long long() const { return __value; }")
	w.open(fmt.Sprintf("bool operator==(const %s &other) const {", typeName))
	w.line("return __value == other.__value;")
	w.close("")
	w.open(fmt.Sprintf("bool operator<(const %s &other) const {", typeName))
	w.line("return __value < other.__value;")
	w.close("")
	w.open(fmt.Sprintf("friend std::ostream &operator<<(std::ostream &out, const %s &value) {", typeName))
	w.line("out << value.__value;")
	w.line("return out;")
	w.close("")
	w.close(";")
}

func (g *Generator) emitVariantSuperComparators(w *cppWriter, typeName string, variants []goivy.Sort) {
	w.open(fmt.Sprintf("bool operator==(const %s &other) const {", typeName))
	w.open("if (__tag != other.__tag) {")
	w.line("return false;")
	w.close("")
	w.open("switch (__tag) {")
	for i, v := range variants {
		vname := varName(sortName(v))
		if vname == "" {
			continue
		}
		w.linef("case %d: return __%s == other.__%s;", i, vname, vname)
	}
	w.line("default: return true;")
	w.close("")
	w.close("")
	w.open(fmt.Sprintf("bool operator<(const %s &other) const {", typeName))
	w.line("if (__tag < other.__tag) return true;")
	w.line("if (other.__tag < __tag) return false;")
	w.open("switch (__tag) {")
	for i, v := range variants {
		vname := varName(sortName(v))
		if vname == "" {
			continue
		}
		w.linef("case %d: return __%s < other.__%s;", i, vname, vname)
	}
	w.line("default: return false;")
	w.close("")
	w.close("")
}

func (g *Generator) emitVariantSuperWriter(w *cppWriter, typeName string, variants []goivy.Sort) {
	w.open(fmt.Sprintf("friend std::ostream &operator<<(std::ostream &out, const %s &value) {", typeName))
	w.open("switch (value.__tag) {")
	for i, v := range variants {
		vname := varName(sortName(v))
		if vname == "" {
			continue
		}
		w.linef("case %d: out << value.__%s; return out;", i, vname)
	}
	w.line(`default: out << "<none>"; return out;`)
	w.close("")
	w.close("")
}

func (g *Generator) emitDestructorStructComparators(w *cppWriter, name string, destructors []*goivy.Const) {
	typeName := varName(name)
	w.open(fmt.Sprintf("bool operator==(const %s &other) const {", typeName))
	comparisons := destructorFieldComparisons(destructors, " == ")
	if len(comparisons) == 0 {
		w.line("return true;")
	} else {
		w.linef("return %s;", strings.Join(comparisons, " && "))
	}
	w.close("")
	w.open(fmt.Sprintf("bool operator<(const %s &other) const {", typeName))
	for _, d := range destructors {
		if _, ok := d.CSort.(*goivy.LogicFunctionSort); !ok {
			continue
		}
		field := varName(memName(d.Name))
		w.linef("if (%s < other.%s) return true;", field, field)
		w.linef("if (other.%s < %s) return false;", field, field)
	}
	w.line("return false;")
	w.close("")
}

func (g *Generator) emitDestructorStructWriter(w *cppWriter, name string, destructors []*goivy.Const) {
	typeName := varName(name)
	w.open(fmt.Sprintf("friend std::ostream &operator<<(std::ostream &out, const %s &value) {", typeName))
	w.line(`out << "{";`)
	w.line("bool first = true;")
	for _, d := range destructors {
		if _, ok := d.CSort.(*goivy.LogicFunctionSort); !ok {
			continue
		}
		field := varName(memName(d.Name))
		w.line(`if (!first) out << ",";`)
		w.line("first = false;")
		w.linef(`out << "%s:" << value.%s;`, field, field)
	}
	w.line(`out << "}";`)
	w.line("return out;")
	w.close("")
}

func destructorFieldComparisons(destructors []*goivy.Const, op string) []string {
	var out []string
	for _, d := range destructors {
		if _, ok := d.CSort.(*goivy.LogicFunctionSort); !ok {
			continue
		}
		field := varName(memName(d.Name))
		out = append(out, field+op+"other."+field)
	}
	return out
}

func (g *Generator) isVariantSuperName(name string) bool {
	if g == nil || g.Mod == nil || len(g.Mod.Variants) == 0 {
		return false
	}
	return len(g.Mod.Variants[name]) > 0
}

func (g *Generator) isVariantSubtypeName(name string) bool {
	if g == nil || g.Mod == nil || len(g.Mod.Variants) == 0 || name == "" {
		return false
	}
	for _, variants := range g.Mod.Variants {
		for _, s := range variants {
			if sortName(s) == name {
				return true
			}
		}
	}
	return false
}

func (g *Generator) emitStateDecls(w *cppWriter) {
	for _, sym := range g.stateSymbols() {
		w.linef("%s;", g.cppStorageDecl(sym.Name, sym.Sort, ""))
	}
	if len(g.stateSymbols()) > 0 {
		w.blank()
	}
}

func (g *Generator) emitCardinalityDecls(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || len(g.Mod.Sig.Interp) == 0 {
		return
	}
	names := make([]string, 0, len(g.Mod.Sig.Interp))
	for name := range g.Mod.Sig.Interp {
		if name != "" && name != "bool" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		w.linef("long long __CARD__%s;", varName(name))
	}
	if len(names) > 0 {
		w.blank()
	}
}

func (g *Generator) emitCardinalityInitializers(w *cppWriter) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil || len(g.Mod.Sig.Interp) == 0 {
		return
	}
	names := make([]string, 0, len(g.Mod.Sig.Interp))
	for name := range g.Mod.Sig.Interp {
		if name != "" && name != "bool" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if s, ok := g.Mod.Sig.Sorts.Get2(name); ok {
			card := cppSortCard(g, s)
			if card > 0 {
				w.linef("__CARD__%s = %d;", varName(name), card)
				continue
			}
		}
		w.linef("__CARD__%s = 0;", varName(name))
	}
}

type stateSymbol struct {
	Name string
	Sort goivy.Sort
}

func (g *Generator) stateSymbols() []stateSymbol {
	var out []stateSymbol
	seen := map[string]bool{}
	defNames := g.definitionNames()
	add := func(name string, s goivy.Sort) {
		if name == "" || seen[name] {
			return
		}
		if defNames[name] {
			return
		}
		if g.Mod.DestructorSorts != nil {
			if _, ok := g.Mod.DestructorSorts[name]; ok {
				return
			}
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
	className := ""
	if qualified {
		className = g.ClassName
	}
	returns := act.GetFormalReturns()
	if len(returns) == 1 {
		ret = g.cppQualifiedType(returns[0].CSort, className)
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
		params = append(params, g.cppStorageDecl(p.Name, p.CSort, className))
	}
	if len(returns) > 1 {
		for _, r := range returns {
			params = append(params, g.cppQualifiedType(r.CSort, className)+" &"+varName(r.Name))
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
		prevReturns := g.currentReturns
		g.currentReturns = returns
		if len(returns) == 1 && !formalListContains(act.GetFormalParams(), returns[0]) {
			w.linef("%s %s = %s;", g.cppType(returns[0].CSort), varName(returns[0].Name), g.cppZeroValue(returns[0].CSort))
		}
		g.emitAction(w, act)
		g.currentReturns = prevReturns
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

func (g *Generator) emitReplSupport(w *cppWriter) {
	g.emitReplParsers(w)
	w.line("static void ivy2cpp_dispatch(" + g.ClassName + " &ivy, const std::string &action, const std::vector<std::string> &args) {")
	w.indent++
	initActions := g.initialMixinActionNames()
	for name := range g.Mod.PublicActions.All() {
		if initActions[name] {
			continue
		}
		username := strings.TrimPrefix(name, "ext:")
		fn, _ := funName(name)
		act, ok := g.Mod.Actions.Get2(name)
		if ok {
			w.open(fmt.Sprintf(`if (action == "%s") {`, username))
			w.linef("ivy2cpp_check_arity(args, %d, action);", len(act.GetFormalParams()))
			args := g.emitReplDispatchArgs(w, act)
			returns := act.GetFormalReturns()
			switch len(returns) {
			case 0:
				w.linef("ivy.%s(%s);", fn, strings.Join(args, ", "))
			case 1:
				w.linef("%s __ivy_result = ivy.%s(%s);", g.cppQualifiedType(returns[0].CSort, g.ClassName), fn, strings.Join(args, ", "))
				g.emitReplWriteOutputs(w, []string{"__ivy_result"})
			default:
				var outNames []string
				for _, r := range returns {
					rname := varName(r.Name)
					w.linef("%s %s = %s;", g.cppQualifiedType(r.CSort, g.ClassName), rname, g.cppZeroValueInScope(r.CSort))
					args = append(args, rname)
					outNames = append(outNames, rname)
				}
				w.linef("ivy.%s(%s);", fn, strings.Join(args, ", "))
				g.emitReplWriteOutputs(w, outNames)
			}
			w.line("return;")
			w.close("")
			continue
		}
		w.linef(`if (action == "%s") { ivy2cpp_check_arity(args, 0, action); ivy.%s(); return; }`, username, fn)
	}
	w.line(`std::cerr << "undefined action: " << action << std::endl;`)
	w.indent--
	w.line("}")
	w.blank()
}

func (g *Generator) emitReplMain(w *cppWriter) {
	mainName := g.Config.MainName
	w.open(fmt.Sprintf("int %s(int argc, char **argv) {", mainName))
	g.emitTestDefaults(w)
	g.emitRuntimeOutputSetup(w)
	g.emitConstructDefaultObject(w)
	g.emitRuntimeArgCapture(w, "ivy")
	w.line("ivy.__init();")
	w.line("ivy.__unlock();")
	w.line("std::string line;")
	w.line("std::string action;")
	w.line("std::vector<std::string> args;")
	w.open("if (isatty(0)) {")
	w.line(`__ivy_out << "> ";`)
	w.line("__ivy_out.flush();")
	w.close("")
	w.open("while (std::getline(std::cin, line)) {")
	w.open("if (line.empty()) {")
	w.line("continue;")
	w.close("")
	w.line("ivy.__lock();")
	w.open("try {")
	w.line("ivy2cpp_parse_command(line, action, args);")
	w.line("ivy2cpp_dispatch(ivy, action, args);")
	w.line("ivy.__unlock();")
	w.close(" catch (const std::exception &err) {")
	w.line("ivy.__unlock();")
	w.line("std::cerr << err.what() << std::endl;")
	w.close("")
	w.open("if (isatty(0)) {")
	w.line(`__ivy_out << "> ";`)
	w.line("__ivy_out.flush();")
	w.close("")
	w.close("")
	w.line("return 0;")
	w.close("")
}

func (g *Generator) emitTestMain(w *cppWriter) {
	mainName := g.Config.MainName
	w.open(fmt.Sprintf("int %s(int argc, char **argv) {", mainName))
	g.emitTestDefaults(w)
	w.line("int seed = 1;")
	w.line("int sleep_ms = 10;")
	w.line("int final_ms = 0;")
	w.line("(void)sleep_ms;")
	w.line("(void)final_ms;")
	w.line("srand(seed);")
	g.emitRuntimeOutputSetup(w)
	w.line("initializing = true;")
	g.emitConstructDefaultObject(w)
	g.emitRuntimeArgCapture(w, "ivy")
	w.line("ivy.__unlock();")
	w.line("initializing = false;")
	g.emitRuntimeBindReaders(w)
	w.line("gen g;")
	w.line("ivy2cpp_setup(g);")
	w.line("ivy2cpp_randomize(g, ivy);")
	g.emitGeneratorInvocations(w)
	w.line(`__ivy_out << "test_completed" << std::endl;`)
	w.line("return 0;")
	w.close("")
}

func (g *Generator) emitGenMain(w *cppWriter) {
	mainName := g.Config.MainName
	w.open(fmt.Sprintf("static void ivy2cpp_generate(%s &ivy) {", g.ClassName))
	w.line("gen g;")
	w.line("ivy2cpp_setup(g);")
	w.line("ivy2cpp_randomize(g, ivy);")
	g.emitGeneratorInvocations(w)
	w.close("")
	w.blank()
	w.open(fmt.Sprintf("int %s(int argc, char **argv) {", mainName))
	g.emitTestDefaults(w)
	g.emitRuntimeOutputSetup(w)
	g.emitConstructDefaultObject(w)
	g.emitRuntimeArgCapture(w, "ivy")
	w.line("ivy.__unlock();")
	w.line("ivy2cpp_generate(ivy);")
	w.line("return 0;")
	w.close("")
}

func (g *Generator) emitTestDefaults(w *cppWriter) {
	w.linef("int test_iters = %s;", g.Config.TestIters)
	w.linef("int runs = %s;", g.Config.TestRuns)
	w.line("(void)test_iters;")
	w.line("(void)runs;")
}

func (g *Generator) emitGeneratorInvocations(w *cppWriter) {
	w.line("init_gen my_init_gen(ivy);")
	w.line("my_init_gen.generate(ivy);")
	initActions := g.initialMixinActionNames()
	for name := range g.Mod.PublicActions.All() {
		if initActions[name] {
			continue
		}
		className := g.actionGeneratorClassName(name)
		genVar := varName(strings.TrimPrefix(name, "ext:")) + "_generator"
		w.linef("%s %s(ivy);", className, genVar)
		w.open(fmt.Sprintf("if (%s.generate(ivy)) {", genVar))
		w.linef("%s.execute(ivy);", genVar)
		w.close("")
	}
}

func (g *Generator) emitConstructDefaultObject(w *cppWriter) {
	if args := g.constructorDefaultArgs(); len(args) == 0 {
		w.linef("%s ivy;", g.runtimeMainClassName())
	} else {
		w.linef("%s ivy{%s};", g.runtimeMainClassName(), strings.Join(args, ", "))
	}
}

func (g *Generator) constructorDefaultArgs() []string {
	args := make([]string, 0, len(g.Mod.Params))
	for _, p := range g.Mod.Params {
		args = append(args, g.cppZeroValueInScope(p.CSort))
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

func (g *Generator) hasInitialMixinActions() bool {
	if g == nil || g.Mod == nil || g.Mod.Actions == nil || g.Mod.Mixins == nil {
		return false
	}
	return len(g.Mod.Mixins.Get("init")) > 0
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
