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

	// thunkCtr names anonymous thunk structs (Python `thunk_counter`,
	// ivy_to_cpp.py:504). Each makeThunk call allocates `__thunk__N`
	// and increments. See thunk.go.
	thunkCtr int

	exprAliases    map[string]goivy.Expr
	currentReturns []*goivy.Const
	errs           []error

	// extRel caches the result of extensionalRelations(). nil before
	// computation; non-nil after the first call (may be empty).
	// Mirrors Python ivy_to_cpp.py:1912-1913 `the_extensional_relations`,
	// but lives on the Generator per goivy/CLAUDE.md section C.
	extRel map[string]bool

	// ptypeCache caches the (param_types, return_types) computed by
	// annotateAction (Python ivy_to_cpp.py:1479-1517). Python attaches
	// these to the action object; Go cannot monkey-patch *Action so the
	// cache lives here per goivy/CLAUDE.md section C.
	ptypeCache map[string]ptypeCacheEntry
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
	g.emitConstructorDecls(w)
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
	g.emitDestructorImpls(w)
	g.emitVariantImpls(w)
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
	g.emitConstructors(w)
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
	g.emitDestructorStructHash(w, destructors)
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
	g.emitVariantWrapperDecl(w, name)
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
	w.line("size_t __hash() const { return hash_space::hash<long long>()(__value); }")
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

// allStateSymbols mirrors Python ivy_to_cpp.py:33-35 `all_state_symbols`.
// It returns every symbol in the signature that is neither a constructor
// nor a solver-interpreted symbol (the result of SolverName is "" for
// interpreted symbols, matching Python's `slv.solver_name(...) is None`).
// Unlike stateSymbols below, it does NOT filter destructors or derived
// definitions — Python filters those at consumer sites (`sym_is_member`,
// `lhs.rep.name not in destructor_sorts`, etc).
//
// Multiple symbols can share a name when polymorphic; deduplicate by
// name so callers can match relation names directly.
func (g *Generator) allStateSymbols() []stateSymbol {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []stateSymbol
	for _, sym := range g.Mod.Sig.AllSymbols() {
		name := sym.Name
		if name == "" || seen[name] {
			continue
		}
		if g.Mod.Sig.Constructors[name] {
			continue
		}
		// Python's `slv.solver_name(il.normalize_symbol(s)) != None`. Treat
		// an error as "non-interpreted" (Python's IvyError path raises
		// rather than excludes; at compile time we don't want to mask it).
		n, err := goivy.SolverName(sym, g.Mod.Sig, nil)
		if err == nil && n == "" {
			continue
		}
		seen[name] = true
		out = append(out, stateSymbol{Name: name, Sort: sym.CSort})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
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
		g.emitMethodDeclLine(w, name, act)
	}
}

// emitMethodDeclLine emits the forward declaration line for a method.
// Mirrors the `emit_method_decl(header,...)` + `header.append(';\n')`
// pair inside Python emit_some_action (ivy_to_cpp.py:1595-1597).
func (g *Generator) emitMethodDeclLine(w *cppWriter, name string, act goivy.Action) {
	w.line(g.methodSignature(name, act, false, false) + ";")
}

// methodSignature emits the C++ method signature for action act.
// Mirrors Python emit_method_decl + emit_param_decls_with_inouts +
// emit_param_decls (ivy_to_cpp.py:1551-1571 + 1542-1549 + 1534-1540).
//
//   - qualified=true emits the body header form (ClassName::name) and
//     suppresses the "virtual " keyword (Python `body=True`).
//   - inline=true suppresses "virtual " on the declaration form too
//     (Python `inline=True`, used by native code emission).
func (g *Generator) methodSignature(name string, act goivy.Action, qualified, inline bool) string {
	className := ""
	if qualified {
		className = g.ClassName
	}

	ptypes, rtypes := g.getParamTypes(name, act)
	formals := act.GetFormalParams()
	returns := act.GetFormalReturns()

	// Return type. Python lines 1561-1564: void if no returns, else
	// ctype(rs[0].sort, classname, ptype=rtypes[0]). ReturnRefType.Make
	// returns "void", which is the same as the no-returns case.
	ret := "void"
	if len(returns) > 0 {
		ret = rtypes[0].Make(g.cppQualifiedType(returns[0].CSort, className))
	}

	// Multi-return validation (Python lines 1565-1567): every secondary
	// return must be ReturnRefType.
	for i := 1; i < len(rtypes); i++ {
		if _, ok := rtypes[i].(ReturnRefType); !ok {
			g.errs = append(g.errs, fmt.Errorf("ivy2cpp: cannot handle multiple output in exported actions: %s", name))
		}
	}

	// "virtual " on the declaration form for non-gen targets (Python
	// lines 1559-1560).
	if !qualified && g.Config.Target != "gen" && !inline {
		ret = "virtual " + ret
	}

	fn, err := funName(name)
	if err != nil {
		fn = varName(name)
	}
	if qualified {
		fn = g.ClassName + "::" + fn
	}

	// Positional input parameters. Function-sorted params use sym_decl
	// (Python emit_param_decls ternary at line 1539); ptype wrappers do
	// not apply to function-sorted parameters.
	var params []string
	for i, p := range formals {
		if _, isFS := p.CSort.(*goivy.LogicFunctionSort); isFS {
			params = append(params, g.cppStorageDecl(p.Name, p.CSort, className))
			continue
		}
		typ := ptypes[i].Make(cppScalarTypeWith(g, p.CSort, className))
		params = append(params, typ+" "+varName(p.Name))
	}

	// Trailing output params for ReturnRefType returns whose Pos lies
	// beyond the input slots (Python emit_param_decls_with_inouts at
	// ivy_to_cpp.py:1542-1549). Each gets RefType.
	for i, r := range returns {
		rrt, ok := rtypes[i].(ReturnRefType)
		if !ok || rrt.Pos < len(formals) {
			continue
		}
		typ := RefType{}.Make(g.cppQualifiedType(r.CSort, className))
		params = append(params, typ+" "+varName(r.Name))
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
		g.emitSomeAction(w, name, act)
	}
}

// emitSomeAction emits one C++ method definition (signature + body) for
// any action carrying formal_params / formal_returns. Mirrors Python
// emit_some_action (ivy_to_cpp.py:1592-1625) for the inline=False path.
//
// Used by emitMethods (ordinary actions), emitDefinitions (derived
// definitions from TODO 010), and emitConstructors (sort constructors
// from TODO 010).
//
// Note: the primary-return local is initialized with cppZeroValue, NOT
// the Python `mk_nondet_sym` call. The nondet-init divergence is owned
// by TODO 014.
func (g *Generator) emitSomeAction(w *cppWriter, name string, act goivy.Action) {
	w.open(g.methodSignature(name, act, true, false) + " {")
	returns := act.GetFormalReturns()
	_, rtypes := g.getParamTypes(name, act)
	// When the primary return is a ReturnRefType, its storage IS
	// an input parameter — no synthetic local, no return statement.
	// Otherwise the primary return is by value and needs both a
	// synthetic local (unless it's also a formal param, in which
	// case it aliases the input slot) and a trailing return.
	firstIsReturnRef := false
	if len(rtypes) >= 1 {
		_, firstIsReturnRef = rtypes[0].(ReturnRefType)
	}
	prevReturns := g.currentReturns
	g.currentReturns = returns
	if len(returns) >= 1 && !firstIsReturnRef && !formalListContains(act.GetFormalParams(), returns[0]) {
		w.linef("%s %s = %s;", g.cppType(returns[0].CSort), varName(returns[0].Name), g.cppZeroValue(returns[0].CSort))
	}
	g.emitAction(w, act)
	g.currentReturns = prevReturns
	if len(returns) >= 1 && !firstIsReturnRef {
		w.linef("return %s;", varName(returns[0].Name))
	}
	w.close("")
	w.blank()
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
