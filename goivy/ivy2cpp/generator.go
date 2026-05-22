package ivy2cpp

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
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

	// testLoopGenEntries threads the per-run generator locals between
	// emitTestLoopBody (which declares them) and emitTestLoopGenBranch
	// (which dispatches on idx via switch). nil outside of emitTestMain.
	testLoopGenEntries []testGenEntry

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

	// nativeOnceMemo dedups header/impl/inline/encode native bodies
	// (Python `once_memo`, ivy_to_cpp.py:1974, 2270-2274, 2401-2403).
	// Lives on Generator per CLAUDE.md section C — never a package var.
	nativeOnceMemo map[string]bool

	// encodedSorts records sorts whose serialization/encoding is
	// supplied by a `<<< encode <sort> ... >>>` native block. Python
	// uses this to suppress default serializer/Z3 emission for those
	// sorts (ivy_to_cpp.py:2315-2323). Consumers land with their owning
	// TODOs (016, 018, 022); for now we just record the set.
	encodedSorts map[string]bool
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
	// Match the per-target setup the isolate path runs in
	// `prepareModuleForCPP`. Skipping this for the direct-Generate path
	// meant `_generating` and similar test-only state never got
	// declared as a class member.
	prepareModuleForCPP(mod, cfg)
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
	if err := g.emitHeaderNatives(w); err != nil {
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
	if err := g.emitClassMemberNatives(w); err != nil {
		return err
	}
	g.emitDefinitionDecls(w)
	g.emitConstructorDecls(w)
	g.emitMethodDecls(w)
	w.indent--
	w.close(";")
	if err := g.emitInlineNatives(w); err != nil {
		return err
	}
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
	g.emitCallbackThunks(w)
	if err := g.emitImplNatives(w); err != nil {
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
	// Per-enum operator<<, _arg<T>, __ser<T>, __deser<T>. Python
	// ivy_to_cpp.py:2497-2510 (operator<<, __ser) and 2634-2652 (_arg,
	// __deser).
	g.emitEnumSortArgSpecImpls(w)
	w.open(g.constructorSignature(true) + " {")
	g.emitRuntimeConstructorPrelude(w)
	g.emitConstructorParamAssignments(w)
	g.emitCardinalityInitializers(w)
	g.emitProgressCounterInitializers(w)
	if err := g.emitInitNatives(w); err != nil {
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

// cardinalitySortNames returns the union of Sig.Interp keys (Python
// behavior: `il.sig.interp` at ivy_to_cpp.py:2305) and all enumerated
// sort names. Go's compiler does not auto-promote enum sorts into
// Sig.Interp the way Python does, but Python's `__CARD__` table covers
// them because Python eventually places them there. Including enums
// here matches Python's effective behavior end-to-end so `ask_ret` (for
// imported callbacks returning an enum) has its bound available.
func (g *Generator) cardinalitySortNames() []string {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil
	}
	seen := map[string]bool{}
	var names []string
	add := func(name string) {
		if name == "" || name == "bool" || seen[name] {
			return
		}
		seen[name] = true
		names = append(names, name)
	}
	for name := range g.Mod.Sig.Interp {
		add(name)
	}
	for _, name := range g.Mod.SortOrder {
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			continue
		}
		if _, ok := s.(*goivy.LogicEnumeratedSort); ok {
			add(name)
		}
	}
	sort.Strings(names)
	return names
}

func (g *Generator) emitCardinalityDecls(w *cppWriter) {
	names := g.cardinalitySortNames()
	for _, name := range names {
		w.linef("long long __CARD__%s;", varName(name))
	}
	if len(names) > 0 {
		w.blank()
	}
}

func (g *Generator) emitCardinalityInitializers(w *cppWriter) {
	for _, name := range g.cardinalitySortNames() {
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

type testGenEntry struct {
	Class string
	Var   string
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

// emitReplSupport emits the per-classname cmd_reader. Mirrors Python
// emit_repl_boilerplate1a + 2 (ivy_to_cpp.py:4137-4187).
func (g *Generator) emitReplSupport(w *cppWriter) {
	g.emitCmdReader(w)
}

// emitReplMain emits the REPL `main()` body. Mirrors Python
// ivy_to_cpp.py:2701-2852 + emit_repl_boilerplate3 (4228-4241) when
// public actions exist, or emit_repl_boilerplate3server (4243-4263)
// when they don't.
func (g *Generator) emitReplMain(w *cppWriter) {
	mainName := g.Config.MainName
	w.open(fmt.Sprintf("int %s(int argc, char **argv) {", mainName))
	g.emitTestDefaults(w)
	g.emitMainParamSetup(w)
	g.emitConstructFromParams(w)
	g.emitRuntimeArgCapture(w, "ivy")
	w.line("ivy.__init();")
	w.line("ivy.__unlock();")
	if g.hasNonInitPublicActions() {
		// emit_repl_boilerplate3 — interactive REPL.
		w.linef("%s_cmd_reader *cr = new %s_cmd_reader(ivy);", g.ClassName, g.ClassName)
		w.open("while (!cr->eof()) {")
		w.line("cr->read();")
		w.close("")
		w.line("delete cr;")
	} else {
		// emit_repl_boilerplate3server — wait for reader threads.
		w.open("for (unsigned i = 0; true; i++) {")
		w.line("ivy.__lock();")
		w.open("if (i >= ivy.thread_ids.size()) {")
		w.line("ivy.__unlock();")
		w.line("break;")
		w.close("")
		w.line("#ifdef _WIN32")
		w.line("HANDLE tid = ivy.thread_ids[i];")
		w.line("ivy.__unlock();")
		w.line("WaitForSingleObject(tid, INFINITE);")
		w.line("#else")
		w.line("pthread_t tid = ivy.thread_ids[i];")
		w.line("ivy.__unlock();")
		w.line("pthread_join(tid, NULL);")
		w.line("#endif")
		w.close("")
	}
	w.line("return 0;")
	w.close("")
}

// hasNonInitPublicActions reports whether any public action remains
// after stripping initial-mixin entries. Drives REPL vs server-mode
// selection (Python ivy_to_cpp.py:2846-2849).
func (g *Generator) hasNonInitPublicActions() bool {
	if g.Mod == nil || g.Mod.PublicActions == nil {
		return false
	}
	initActions := g.initialMixinActionNames()
	for name := range g.Mod.PublicActions.All() {
		if !initActions[name] {
			return true
		}
	}
	return false
}

func (g *Generator) emitTestMain(w *cppWriter) {
	mainName := g.Config.MainName
	w.open(fmt.Sprintf("int %s(int argc, char **argv) {", mainName))
	g.emitTestDefaults(w)
	g.emitMainParamSetup(w)
	// Multi-run loop. Python emit_repl_boilerplate3test: outer for over
	// runidx, inner for over cycle.
	w.open("for (int runidx = 0; runidx < runs; runidx++) {")
	w.line("initializing = true;")
	g.emitConstructFromParams(w)
	g.emitRuntimeArgCapture(w, "ivy")
	w.line("ivy._generating = false;")
	w.line("ivy.__init();")
	w.line("ivy.__unlock();")
	w.line("initializing = false;")
	g.emitRuntimeBindReaders(w)
	w.blank()
	g.emitTestLoopBody(w)
	w.close("")
	w.line("return 0;")
	w.close("")
}

// emitTestLoopBody emits the body of the per-run test driver, mirroring
// Python `emit_repl_boilerplate3test` (ivy_to_cpp.py:4265-4467):
// build init_gen, weighted action generators, then loop test_iters
// times choosing among generators / readers / timers via select().
//
// The Go port's `gen` base class (ivy_go_z3.hpp) does not have virtual
// generate/execute methods (Python's templated `ivy_z3_gen` does). So
// instead of `vector<gen *>` polymorphism we declare each generator as
// a stack local and dispatch by index through a switch. This keeps
// ivy_go_z3.hpp source-stable.
func (g *Generator) emitTestLoopBody(w *cppWriter) {
	// init_gen sets up the initial state.
	w.line("init_gen my_init_gen(ivy);")
	w.line("my_init_gen.generate(ivy);")
	w.blank()
	w.line("std::vector<double> weights;")
	initActions := g.initialMixinActionNames()
	names := g.publicActionNamesSorted()
	totalweight := 0.0
	var entries []testGenEntry
	for _, name := range names {
		if initActions[name] || isFinalizeName(name) {
			continue
		}
		className := g.actionGeneratorClassName(name)
		genVar := varName(strings.TrimPrefix(name, "ext:")) + "_generator"
		w.linef("%s %s(ivy);", className, genVar)
		weight := g.actionWeight(name)
		w.linef("weights.push_back(%g);", weight)
		totalweight += weight
		entries = append(entries, testGenEntry{Class: className, Var: genVar})
	}
	w.linef("double totalweight = %g;", totalweight)
	w.linef("int num_gens = %d;", len(entries))
	w.blank()
	g.testLoopGenEntries = entries
	defer func() { g.testLoopGenEntries = nil }()
	w.line("#ifdef _WIN32")
	w.line("LARGE_INTEGER freq;")
	w.line("QueryPerformanceFrequency(&freq);")
	w.line("#endif")
	w.line("double frnd = 0.0;")
	w.line("bool do_over = false;")
	w.open("for (int cycle = 0; cycle < test_iters; cycle++) {")
	w.line("double choices = totalweight + 5.0;")
	w.open("if (do_over) {")
	w.line("do_over = false;")
	w.close(" else {")
	w.indent++
	w.line("frnd = choices * (((double)rand()) / (((double)RAND_MAX) + 1.0));")
	w.indent--
	w.line("}")
	w.open("if (frnd < totalweight) {")
	g.emitTestLoopGenBranch(w)
	w.line("continue;")
	w.close("")
	w.blank()
	g.emitTestLoopSelectBranch(w)
	w.close("")
	if g.hasFinalizeExport() {
		w.line("ivy.__lock(); ivy.ext___finalize(); ivy.__unlock();")
	}
	w.line(`__ivy_out << "test_completed" << std::endl;`)
	w.line("#ifdef _WIN32")
	w.line("Sleep(final_ms);")
	w.line("#endif")
	w.open("if (runidx == runs - 1) {")
	w.line("struct timespec ts;")
	w.line("int ms = 50;")
	w.line("ts.tv_sec = ms / 1000;")
	w.line("ts.tv_nsec = (ms % 1000) * 1000000;")
	w.line("nanosleep(&ts, NULL);")
	w.line("exit(0);")
	w.close("")
	w.open("for (unsigned i = 0; i < readers.size(); i++) {")
	w.line("delete readers[i];")
	w.close("")
	w.line("readers.clear();")
	w.open("for (unsigned i = 0; i < timers.size(); i++) {")
	w.line("delete timers[i];")
	w.close("")
	w.line("timers.clear();")
}

func (g *Generator) emitTestLoopGenBranch(w *cppWriter) {
	w.line("int idx = 0;")
	w.line("double sum = 0.0;")
	w.open("while (idx < num_gens - 1) {")
	w.line("sum += weights[idx];")
	w.line("if (frnd < sum) break;")
	w.line("idx++;")
	w.close("")
	w.line("ivy.__lock();")
	w.line("ivy._generating = true;")
	w.line("bool sat = false;")
	// Per-index dispatch (no virtual `gen` API in ivy_go_z3.hpp).
	if len(g.testLoopGenEntries) > 0 {
		w.open("switch (idx) {")
		for i, e := range g.testLoopGenEntries {
			w.linef("case %d: sat = %s.generate(ivy); break;", i, e.Var)
		}
		w.close("")
	}
	w.open("if (sat) {")
	if len(g.testLoopGenEntries) > 0 {
		w.open("switch (idx) {")
		for i, e := range g.testLoopGenEntries {
			w.linef("case %d: %s.execute(ivy); break;", i, e.Var)
		}
		w.close("")
	}
	w.line("ivy._generating = false;")
	w.line("ivy.__unlock();")
	w.line("#ifdef _WIN32")
	w.line("Sleep(sleep_ms);")
	w.line("#endif")
	w.close(" else {")
	w.indent++
	w.line("ivy._generating = false;")
	w.line("ivy.__unlock();")
	w.line("cycle--;")
	w.indent--
	w.line("}")
}

func (g *Generator) emitTestLoopSelectBranch(w *cppWriter) {
	w.line("fd_set rdfds;")
	w.line("FD_ZERO(&rdfds);")
	w.line("int maxfds = 0;")
	w.open("for (unsigned i = 0; i < readers.size(); i++) {")
	w.line("reader *r = readers[i];")
	w.line("int fds = r->fdes();")
	w.open("if (fds >= 0) {")
	w.line("FD_SET(fds, &rdfds);")
	w.close("")
	w.line("if (fds > maxfds) maxfds = fds;")
	w.close("")
	w.line("#ifdef _WIN32")
	w.line("int timer_min = 15;")
	w.line("#else")
	w.line("int timer_min = 5;")
	w.line("#endif")
	w.line("struct timeval timeout;")
	w.line("timeout.tv_sec = timer_min / 1000;")
	w.line("timeout.tv_usec = 1000 * (timer_min % 1000);")
	w.line("#ifdef _WIN32")
	w.line("int foo;")
	w.open("if (readers.size() == 0) {")
	w.line("Sleep(timer_min);")
	w.line("foo = 0;")
	w.close(" else {")
	w.indent++
	w.line("foo = select(maxfds + 1, &rdfds, 0, 0, &timeout);")
	w.indent--
	w.line("}")
	w.line("#else")
	w.line("int foo = select(maxfds + 1, &rdfds, 0, 0, &timeout);")
	w.line("#endif")
	w.open("if (foo < 0) {")
	w.line("#ifdef _WIN32")
	w.line(`std::cerr << "select failed: " << WSAGetLastError() << std::endl; __ivy_exit(1);`)
	w.line("#else")
	w.line(`perror("select failed"); __ivy_exit(1);`)
	w.line("#endif")
	w.close("")
	w.open("if (foo == 0) {")
	w.line("cycle--;")
	w.open("for (unsigned i = 0; i < timers.size(); i++) {")
	w.open("if (timer_min >= timers[i]->ms_delay()) {")
	w.line("cycle++;")
	w.line("break;")
	w.close("")
	w.close("")
	w.open("for (unsigned i = 0; i < timers.size(); i++) {")
	w.line("timers[i]->timeout(timer_min);")
	w.close("")
	w.close(" else {")
	w.indent++
	w.line("int fdc = 0;")
	w.open("for (unsigned i = 0; i < readers.size(); i++) {")
	w.line("reader *r = readers[i];")
	w.line("if (FD_ISSET(r->fdes(), &rdfds)) fdc++;")
	w.close("")
	w.line("int fdi = fdc * (((double)rand()) / (((double)RAND_MAX) + 1.0));")
	w.line("fdc = 0;")
	w.open("for (unsigned i = 0; i < readers.size(); i++) {")
	w.line("reader *r = readers[i];")
	w.open("if (FD_ISSET(r->fdes(), &rdfds)) {")
	w.open("if (fdc == fdi) {")
	w.line("r->read();")
	w.open("if (r->background()) {")
	w.line("cycle--;")
	w.line("do_over = true;")
	w.close("")
	w.line("break;")
	w.close("")
	w.line("fdc++;")
	w.close("")
	w.close("")
	w.indent--
	w.line("}")
}

// isFinalizeName matches both the Python prefixed form ("ext:_finalize")
// and Go's bare form ("_finalize"). The Go compiler does not always
// prepend `ext:` for exported actions, but the runtime helper method
// is still named `ext___finalize` to match Python.
func isFinalizeName(name string) bool {
	return name == "ext:_finalize" || name == "_finalize"
}

func (g *Generator) hasFinalizeExport() bool {
	if g.Mod == nil || g.Mod.PublicActions == nil {
		return false
	}
	for name := range g.Mod.PublicActions.All() {
		if isFinalizeName(name) {
			return true
		}
	}
	return false
}

// actionWeight returns the `<actname>.weight` attribute as a double,
// defaulting to 1.0. Mirrors Python ivy_to_cpp.py:4287-4298.
func (g *Generator) actionWeight(name string) float64 {
	username := strings.TrimPrefix(name, "ext:")
	if g.Mod.Attributes == nil {
		return 1.0
	}
	raw, ok := g.Mod.Attributes[username+".weight"]
	if !ok {
		return 1.0
	}
	type relnamer interface{ Relname() string }
	var s string
	switch v := raw.(type) {
	case string:
		s = v
	case relnamer:
		s = v.Relname()
	default:
		return 1.0
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 1.0
	}
	return f
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
	g.emitMainParamSetup(w)
	g.emitConstructFromParams(w)
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

// emitConstructFromParams emits `ClassName_repl ivy(p__a, p__b, ...);`
// matching Python ivy_to_cpp.py:2838 where the actual parameter
// locals (already filled by emitMainParamSetup) are forwarded.
func (g *Generator) emitConstructFromParams(w *cppWriter) {
	if len(g.Mod.Params) == 0 {
		w.linef("%s ivy;", g.runtimeMainClassName())
		return
	}
	args := make([]string, len(g.Mod.Params))
	for i, p := range g.Mod.Params {
		args[i] = "p__" + varName(p.Name)
	}
	w.linef("%s ivy(%s);", g.runtimeMainClassName(), strings.Join(args, ", "))
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
