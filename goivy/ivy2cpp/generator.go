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
	// HostOS overrides the build-host detection used for the header
	// preamble. Python `ivy_to_cpp.py:1948` checks `platform.system()`
	// at codegen time to gate `WIN32_LEAN_AND_MEAN`/`<windows.h>`. Empty
	// string falls back to runtime.GOOS; tests set "windows"/"linux" to
	// exercise both branches.
	HostOS string
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
	Ctx       *CppContext

	header cppWriter
	impl   cppWriter
	tempID int

	// thunkCtr names anonymous thunk structs (Python `thunk_counter`,
	// ivy_to_cpp.py:504). Each distinct file-scope thunk allocates
	// `__thunk__N` and increments. See thunk.go.
	thunkCtr        int
	thunkDefs       cppWriter
	thunkMemo       map[string]string
	fileScopeThunks bool

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

	// importCallersCache memoizes find_import_callers (ivy_to_cpp.py:1888-
	// 1897). Populated on first access by importCallers(). nil before that;
	// non-nil (possibly empty) after.
	importCallersCache map[string]bool

	// numberFormatCache memoizes the result of numberFormat() so we don't
	// re-scan module attributes on every trace line. Empty string when not
	// yet computed; readers should use numberFormat() which initializes it.
	numberFormatCache    string
	numberFormatComputed bool
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
		Ctx:       NewCppContext(),
		// Python make_thunk writes through the impl-level `thunks`
		// buffer. Unit tests that call makeThunk directly leave this
		// false and get the legacy inline writer for focused white-box
		// assertions; full Generate uses the faithful file-scope path.
		fileScopeThunks: true,
	}
	g.header = newCPPWriter(g.Ctx.Globals)
	g.impl = newCPPWriter(g.Ctx.Impls)
	applySessionParameters(mod, cfg)
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
	if err := g.checkMemberNames(); err != nil {
		return err
	}
	if err := g.emitHeader(); err != nil {
		return err
	}
	if err := g.emitImpl(); err != nil {
		return err
	}
	return errors.Join(g.errs...)
}

// checkMemberNames mirrors Python check_member_names
// (ivy_to_cpp.py:1830-1834). Reject any module whose generated C++ class name
// would collide with a member name derived from a signature symbol, sort, or
// action (after varName lowering, matching Python's varname mapping).
func (g *Generator) checkMemberNames() error {
	if g == nil || g.Mod == nil {
		return nil
	}
	names := map[string]bool{}
	if g.Mod.Sig != nil {
		for name := range g.Mod.Sig.Symbols.All() {
			names[varName(name)] = true
		}
		for name := range g.Mod.Sig.Sorts.All() {
			names[varName(name)] = true
		}
	}
	if g.Mod.Actions != nil {
		for name := range g.Mod.Actions.All() {
			names[varName(name)] = true
		}
	}
	if names[g.ClassName] {
		return fmt.Errorf(
			"ivy2cpp: cannot create C++ class %s with member %s.\n"+
				"Use command line option classname=... to change the class name",
			g.ClassName, g.ClassName)
	}
	return nil
}

func (g *Generator) unsupported(w *cppWriter, format string, args ...any) {
	err := fmt.Errorf("ivy2cpp: "+format, args...)
	g.errs = append(g.errs, err)
	w.linef("/* %s */", escapeComment(err.Error()))
}

func (g *Generator) nextTemp(prefix string) string {
	id := g.tempID
	g.tempID++
	return fmt.Sprintf("%s%d", prefix, id)
}

func (g *Generator) emitHeader() error {
	w := &g.header
	g.emitRuntimeHeaderPreamble(w)
	if err := g.emitHeaderNatives(w); err != nil {
		return err
	}
	g.emitRuntimeHeaderForwardDecls(w)
	w.blank()
	if g.Config.Target == "test" {
		w.linef("class %s {", g.ClassName)
		w.raw("  public:\n")
		w.indent++
	} else {
		w.open(fmt.Sprintf("class %s {", g.ClassName))
		w.line("public:")
		w.indent++
	}
	w.linef("typedef %s ivy_class;", g.ClassName)
	g.emitRuntimeClassMembers(w)
	w.line("int ___ivy_choose(int rng,const char *name,int id);")
	if g.Config.Target != "gen" {
		w.line("virtual void ivy_assert(bool,const char *){}")
		w.line("virtual void ivy_assume(bool,const char *){}")
		w.line("virtual void ivy_check_progress(int,int){}")
	}
	g.emitSortDecls(w)
	g.emitCTupleDecls(w)
	g.emitCTupleHashDecls(w)
	g.emitStateDecls(w)
	g.emitProgressCounterDecls(w)
	g.emitCardinalityDecls(w)
	if err := g.emitClassMemberNatives(w); err != nil {
		return err
	}
	if g.Config.Target == "test" {
		w.raw("    " + g.constructorSignature(false) + ";\n")
	} else {
		w.line(g.constructorSignature(false) + ";")
	}
	if g.Config.Target == "test" {
		w.raw("void __init();\n")
	} else {
		w.line("void __init();")
	}
	g.emitDefinitionDecls(w)
	g.emitConstructorDecls(w)
	g.emitMethodDecls(w)
	w.line("void __tick(int timeout);")
	w.indent--
	w.close(";")
	g.emitVariantEqualityForwardDecls(w)
	g.emitDestructorEqualityInlines(w)
	g.emitVariantEqualityInlines(w)
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
	if g.Config.Target == "repl" || g.Config.Target == "test" {
		g.emitRuntimeInstallMethods(w)
	}
	g.emitRuntimeLockMethods(w)
	g.emitCallbackThunks(w)
	if g.Config.Target == "test" {
		g.emitZ3Boilerplate1(w)
	}
	g.emitRuntimeValueIncludes(w)
	g.emitCTupleEqualities(w)
	if err := g.emitImplNatives(w); err != nil {
		return err
	}
	if g.Config.Target == "test" {
		g.emitCPPTypeImpls(w)
	} else if g.usesZ3() {
		if err := g.emitZ3Support(w); err != nil {
			return err
		}
	} else {
		g.emitCPPTypeImpls(w)
	}
	g.emitVariantImpls(w)
	if g.usesZ3() && g.Config.Target != "test" {
		// to_solver_class<hash_thunk<D,R>> specializations for every
		// hash_thunk-backed domain (single-arg and ctuple). Python
		// ivy_to_cpp.py:2673 → emit_all_ctuples_to_solver.
		g.emitAllCtuplesToSolver(w)
		// The init_gen / action_gen classes go AFTER the variant and
		// destructor __from_solver specializations so the action_gen
		// body can deduce the right template specialization.
		if err := g.emitZ3GeneratorClasses(w); err != nil {
			return err
		}
	}
	g.emitRuntimeChoose(w)
	var methodSection cppWriter
	mw := &methodSection
	g.emitInit(mw)
	g.emitDefinitions(mw)
	g.emitConstructors(mw)
	g.emitMethods(mw)
	g.emitTick(mw)
	w.raw(g.thunkDefs.String())
	w.raw(methodSection.String())
	var body cppWriter
	bw := &body
	if g.Config.Target == "test" {
		bw.open(g.constructorSignature(true) + "{")
	} else {
		bw.open(g.constructorSignature(true) + " {")
	}
	g.emitRuntimeConstructorPrelude(bw)
	g.emitConstructorParamAssignments(bw)
	g.emitCardinalityInitializers(bw)
	if err := g.emitInitNatives(bw); err != nil {
		return err
	}
	if !g.usesZ3() {
		if err := g.emitOneInitialState(bw); err != nil {
			return err
		}
	}
	bw.close("")
	w.raw(body.String())
	g.emitRuntimeDestructor(w)
	if g.Config.Target == "test" {
		// Python's target=test path emits generator classes before the
		// parser/Z3 helper definitions, then splits serializers around
		// the REPL subclass.
	} else {
		g.emitDestructorImpls(w)
	}
	if g.usesZ3() && g.Config.Target == "test" {
		if err := g.emitZ3GeneratorClasses(w); err != nil {
			return err
		}
		g.emitDestructorOutSerImpls(w)
		g.emitEnumSortOutSerImpls(w)
	}
	// Per-enum operator<<, _arg<T>, __ser<T>, __deser<T>. Python
	// emits the definitions after class methods and the runtime
	// destructor; only forward declarations live near ivy_value.hpp.
	if g.Config.Target != "test" {
		g.emitEnumSortArgSpecImpls(w)
	}
	var tail cppWriter
	bw = &tail
	if g.runtimeUsesReplSubclass() {
		g.emitRuntimeReplSubclass(bw)
	}
	switch g.Config.Target {
	case "repl":
		g.emitReplSupport(bw)
		if g.Config.EmitMain {
			g.emitReplMain(bw)
		}
	case "test":
		g.emitAllCtuplesToSolver(bw)
		g.emitDestructorArgDeserZ3Impls(bw)
		g.emitEnumSortArgDeserImpls(bw)
		g.emitZ3SolverConversions(bw)
		g.emitReplSupport(bw)
		if g.Config.EmitMain {
			g.emitTestMain(bw)
		}
	case "gen":
		if g.Config.EmitMain {
			g.emitGenMain(bw)
		}
	}
	w.raw(tail.String())
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
	if g.Mod.Sig == nil || g.Mod.Sig.Symbols.Len() == 0 {
		return
	}
	emitted := map[string]bool{}
	visiting := map[string]bool{}
	emittedDestructorStructs := map[string]bool{}
	emittedVariantSupers := map[string]bool{}
	emittedIntClass := false

	var emitOne func(name string, force bool)
	emitOne = func(name string, force bool) {
		if name == "" || name == "bool" || emitted[name] {
			return
		}
		if visiting[name] {
			return
		}
		visiting[name] = true
		for _, dep := range g.sortDeclDependencyNames(name) {
			emitOne(dep, true)
		}
		visiting[name] = false

		if g.isVariantSuperName(name) {
			emitted[name] = true
			g.emitVariantSuperStruct(w, name)
			emittedVariantSupers[name] = true
			return
		}
		if nt, ok := g.nativeTypeForSort(name); ok {
			emitted[name] = true
			g.emitNativeTypeDecl(w, name, nt)
			return
		}
		if g.Mod.SortDestructors != nil {
			if _, ok := g.Mod.SortDestructors.Get2(name); ok {
				emitted[name] = true
				g.emitDestructorStruct(w, name)
				emittedDestructorStructs[name] = true
				return
			}
		}
		if g.isPlainVariantSubtypeName(name) {
			emitted[name] = true
			return
		}
		if g.isVariantSubtypeName(name) {
			emitted[name] = true
			g.emitVariantLeafStruct(w, name)
			return
		}
		s, ok := g.Mod.Sig.Sorts.Get2(name)
		if !ok {
			emitted[name] = true
			return
		}
		if !force && !g.sortNeededForGeneratedDecl(name) {
			return
		}
		emitted[name] = true
		if it, ok := g.cppInterpType(s); ok {
			if it.Kind == cppInterpBV && it.primitiveType() != "" && (!g.usesZ3() || g.Config.Target == "test") {
				return
			}
			if it.Kind == cppInterpIntBV && !emittedIntClass {
				g.emitIntClassDecl(w)
				emittedIntClass = true
			}
			g.emitCPPTypeDecl(w, s, it)
			return
		}
		if _, interpreted := g.Mod.Sig.Interp[name]; interpreted {
			return
		}
		switch st := s.(type) {
		case *goivy.LogicEnumeratedSort:
			if isNumericEnum(st) {
				return
			}
			vals := make([]string, len(st.Extension))
			for i, v := range st.Extension {
				vals[i] = varName(v)
			}
			if g.Config.Target == "test" {
				w.linef("enum %s{%s};", varName(st.Name), strings.Join(vals, ","))
				break
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

	for _, name := range g.Mod.SortOrder {
		emitOne(name, false)
	}
	if g.Mod.SortDestructors != nil {
		destructorNames := insMapKeys(g.Mod.SortDestructors)
		if len(destructorNames) > 0 {
			w.blank()
			for _, name := range destructorNames {
				if emittedDestructorStructs[name] {
					continue
				}
				emitOne(name, true)
			}
		}
	}
	for _, name := range g.Mod.SortOrder {
		if !g.isVariantSuperName(name) || emittedVariantSupers[name] {
			continue
		}
		g.emitVariantSuperStruct(w, name)
		emittedVariantSupers[name] = true
	}
	if g.Config.Target != "test" {
		w.blank()
	}
}

func (g *Generator) sortDeclDependencyNames(name string) []string {
	if g == nil || g.Mod == nil || name == "" {
		return nil
	}
	seen := map[string]bool{}
	var deps []string
	for _, dep := range g.Mod.SortDependencies(name, false) {
		if dep == "" || dep == "bool" || dep == name || seen[dep] {
			continue
		}
		seen[dep] = true
		deps = append(deps, dep)
	}
	return deps
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
	w.close(";")
}

func (g *Generator) emitDestructorEqualityInlines(w *cppWriter) {
	for _, name := range g.destructorSortNames() {
		destrs := g.Mod.SortDestructors.Get(name)
		typeName := g.ClassName + "::" + varName(name)
		w.open(fmt.Sprintf("inline bool operator ==(const %s &s, const %s &t) {", typeName, typeName))
		parts := make([]string, 0, len(destrs))
		for _, d := range destrs {
			fs, ok := d.CSort.(*goivy.LogicFunctionSort)
			if !ok {
				continue
			}
			domain := fs.Domain()
			if len(domain) > 0 {
				domain = domain[1:]
			}
			field := varName(memName(d.Name))
			st := cppFunctionStorageFor(g, domain, fs.Range(), "")
			if st.Kind == cppStorageArray || st.Kind == cppStorageHashThunk {
				continue
			}
			parts = append(parts, fmt.Sprintf("(s.%s == t.%s)", field, field))
		}
		if len(parts) == 0 {
			w.line("return true;")
		} else {
			w.linef("return (%s);", strings.Join(parts, " && "))
		}
		w.close("")
	}
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
		w.open("size_t __hash() const {")
		w.line("size_t hv = 0;")
		for i, s := range dom {
			w.linef("hv += hash_space::hash<%s>()(arg%d);", cppHashType(g, s), i)
		}
		w.line("return hv;")
		w.close("")
		w.close(";")
		w.blank()
	}
}

func (g *Generator) emitCTupleHashDecls(w *cppWriter) {
	for _, dom := range g.cppCTuples() {
		name := cppCTupleLocalNameWith(g, dom)
		hashName := "hash__" + name
		qualified := g.ClassName + "::" + name
		w.open(fmt.Sprintf("class %s {", hashName))
		w.line("public:")
		w.indent++
		w.open(fmt.Sprintf("size_t operator()(const %s &__s) const {", qualified))
		hashParts := make([]string, len(dom))
		for i, s := range dom {
			hashParts[i] = fmt.Sprintf("hash_space::hash<%s>()(__s.arg%d)", cppHashType(g, s), i)
		}
		w.linef("return %s;", strings.Join(hashParts, "+"))
		w.close("")
		w.indent--
		w.close(";")
		w.blank()
	}
}

func (g *Generator) emitCTupleEqualities(w *cppWriter) {
	for _, dom := range g.cppCTuples() {
		name := cppCTupleLocalNameWith(g, dom)
		qualified := g.ClassName + "::" + name
		w.open(fmt.Sprintf("bool operator==(const %s &x, const %s &y) {", qualified, qualified))
		eqParts := make([]string, len(dom))
		for i := range dom {
			eqParts[i] = fmt.Sprintf("x.arg%d == y.arg%d", i, i)
		}
		w.linef("return %s;", strings.Join(eqParts, " && "))
		w.close("")
	}
	if len(g.cppCTuples()) > 0 {
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

func (g *Generator) isPlainVariantSubtypeName(name string) bool {
	if !g.isVariantSubtypeName(name) {
		return false
	}
	if g.Mod == nil {
		return false
	}
	if _, ok := g.Mod.NativeTypes[name]; ok {
		return false
	}
	if g.Mod.SortDestructors != nil {
		if _, ok := g.Mod.SortDestructors.Get2(name); ok {
			return false
		}
	}
	if g.Mod.Sig != nil {
		if _, ok := g.Mod.Sig.Interp[name]; ok {
			return false
		}
	}
	return true
}

func (g *Generator) emitStateDecls(w *cppWriter) {
	for _, sym := range g.stateSymbols() {
		w.linef("%s;", g.cppStorageDecl(sym.Name, sym.Sort, ""))
	}
	if len(g.stateSymbols()) > 0 && g.Config.Target != "test" {
		w.blank()
	}
}

// cardinalitySortNames returns Sig.Interp keys, matching Python
// ivy_to_cpp.py:2305.
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
		if g.isPlainVariantSubtypeName(name) {
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
			if !g.shouldInitializeCardinality(name) {
				continue
			}
			card := cppSortCard(g, s)
			if card > 0 {
				w.linef("__CARD__%s = %d;", varName(name), card)
				continue
			}
			w.linef("__CARD__%s = 0;", varName(name))
		}
	}
}

func (g *Generator) shouldInitializeCardinality(name string) bool {
	return g.sortNeededForRuntimeSpecs(name)
}

func (g *Generator) sortNeededForGeneratedDecl(name string) bool {
	if g.sortNeededForRuntimeSpecs(name) {
		return true
	}
	if g.Mod.Actions != nil {
		for _, act := range g.Mod.Actions.All() {
			if g.exprReferencesSortName(act, name, map[goivy.NodeKey]bool{}) {
				return true
			}
		}
	}
	return false
}

func (g *Generator) sortNeededForRuntimeSpecs(name string) bool {
	for _, sym := range g.stateSymbols() {
		if g.sortDependencyReferencesName(sym.Sort, name, map[string]bool{}) {
			return true
		}
	}
	for _, p := range g.Mod.Params {
		if g.sortDependencyReferencesName(p.CSort, name, map[string]bool{}) {
			return true
		}
	}
	if g.Mod.Actions != nil {
		for _, act := range g.Mod.Actions.All() {
			for _, p := range act.GetFormalParams() {
				if g.sortDependencyReferencesName(p.CSort, name, map[string]bool{}) {
					return true
				}
			}
			for _, r := range act.GetFormalReturns() {
				if g.sortDependencyReferencesName(r.CSort, name, map[string]bool{}) {
					return true
				}
			}
		}
	}
	for _, d := range g.allDefinitions() {
		if g.sortDependencyReferencesName(d.Sort, name, map[string]bool{}) {
			return true
		}
		for _, p := range d.Params {
			if p != nil && g.sortDependencyReferencesName(p.NodeSort(), name, map[string]bool{}) {
				return true
			}
		}
	}
	return false
}

func (g *Generator) exprReferencesSortName(e goivy.Expr, name string, seen map[goivy.NodeKey]bool) bool {
	if e == nil || name == "" {
		return false
	}
	key := goivy.Key(e)
	if seen[key] {
		return false
	}
	seen[key] = true
	if g.sortDependencyReferencesName(e.NodeSort(), name, map[string]bool{}) {
		return true
	}
	for _, child := range e.Children() {
		if g.exprReferencesSortName(child, name, seen) {
			return true
		}
	}
	return false
}

func (g *Generator) sortDependencyReferencesName(s goivy.Sort, name string, seen map[string]bool) bool {
	if sortReferencesName(s, name) {
		return true
	}
	sortText := sortName(s)
	if sortText == "" || seen[sortText] {
		return false
	}
	seen[sortText] = true
	if g == nil || g.Mod == nil {
		return false
	}
	for _, sub := range g.Mod.Variants[sortText] {
		if g.sortDependencyReferencesName(sub, name, seen) {
			return true
		}
	}
	if g.Mod.SortDestructors == nil {
		return false
	}
	for _, d := range g.Mod.SortDestructors.Get(sortText) {
		if fs, ok := d.CSort.(*goivy.LogicFunctionSort); ok {
			if g.sortDependencyReferencesName(fs.Range(), name, seen) {
				return true
			}
		}
	}
	return false
}

func sortReferencesName(s goivy.Sort, name string) bool {
	if s == nil || name == "" {
		return false
	}
	if sortName(s) == name {
		return true
	}
	if fs, ok := s.(*goivy.LogicFunctionSort); ok {
		for _, d := range fs.Domain() {
			if sortReferencesName(d, name) {
				return true
			}
		}
		return sortReferencesName(fs.Range(), name)
	}
	return false
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
	if g == nil || g.Mod == nil {
		return nil
	}
	seen := map[string]bool{}
	knownSig := map[string]bool{}
	var out []stateSymbol
	add := func(name string, s goivy.Sort) {
		if name == "" || seen[name] {
			return
		}
		if g.Mod.Sig != nil && g.Mod.Sig.Constructors[name] {
			return
		}
		if g.isSortConstructorName(name) {
			return
		}
		seen[name] = true
		out = append(out, stateSymbol{Name: name, Sort: s})
	}
	if g.Mod.Sig != nil {
		for _, sym := range g.Mod.Sig.AllSymbols() {
			name := sym.Name
			if name != "" {
				knownSig[name] = true
			}
			if name == "" || seen[name] {
				continue
			}
			if g.Mod.Sig.Constructors[name] || g.isSortConstructorName(name) {
				continue
			}
			// Python's `slv.solver_name(il.normalize_symbol(s)) != None`. Treat
			// an error as "non-interpreted" (Python's IvyError path raises
			// rather than excludes; at compile time we don't want to mask it).
			n, err := goivy.SolverName(sym, g.Mod.Sig, nil)
			if err == nil && n == "" {
				continue
			}
			add(name, sym.CSort)
		}
	}
	if g.Mod.Relations != nil {
		for key, s := range g.Mod.Relations.All() {
			name := goivy.SymbolNameFromKey(key)
			if !knownSig[name] {
				add(name, s)
			}
		}
	}
	if g.Mod.Functions != nil {
		for key, s := range g.Mod.Functions.All() {
			name := goivy.SymbolNameFromKey(key)
			if !knownSig[name] {
				add(name, s)
			}
		}
	}
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
		if g.isSortConstructorName(name) {
			return
		}
		seen[name] = true
		out = append(out, stateSymbol{Name: name, Sort: s})
	}
	for _, sym := range g.allStateSymbols() {
		add(sym.Name, sym.Sort)
	}
	return out
}

func (g *Generator) isSortConstructorName(name string) bool {
	if g == nil || g.Mod == nil || name == "" {
		return false
	}
	if g.Mod.ConstructorSorts != nil {
		if _, ok := g.Mod.ConstructorSorts[name]; ok {
			return true
		}
	}
	if g.Mod.SortConstructors != nil {
		for _, conss := range g.Mod.SortConstructors.All() {
			for _, cons := range conss {
				if cons != nil && cons.Name == name {
					return true
				}
			}
		}
	}
	return false
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
	fnClassName := ""
	if qualified {
		fnClassName = g.ClassName
	}
	ptypes, rtypes := g.getParamTypes(name, act)
	formals := act.GetFormalParams()
	returns := act.GetFormalReturns()

	// Return type. Python lines 1561-1564: void if no returns, else
	// ctype(rs[0].sort, classname, ptype=rtypes[0]). ReturnRefType.Make
	// returns "void", which is the same as the no-returns case.
	ret := "void"
	if len(returns) > 0 {
		retClassName := ""
		if qualified || g.isVariantSuperName(sortName(returns[0].CSort)) {
			retClassName = g.ClassName
		}
		ret = rtypes[0].Make(g.cppQualifiedType(returns[0].CSort, retClassName))
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
		fn = fnClassName + "::" + fn
	}

	// Positional input parameters. Function-sorted params use sym_decl
	// (Python emit_param_decls ternary at line 1539); ptype wrappers do
	// not apply to function-sorted parameters.
	var params []string
	for i, p := range formals {
		if _, isFS := p.CSort.(*goivy.LogicFunctionSort); isFS {
			params = append(params, g.cppStorageDecl(p.Name, p.CSort, ""))
			continue
		}
		typ := ptypes[i].Make(cppScalarTypeWith(g, p.CSort, ""))
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
		typ := RefType{}.Make(g.cppQualifiedType(r.CSort, ""))
		params = append(params, typ+" "+varName(r.Name))
	}

	return fmt.Sprintf("%s %s(%s)", ret, fn, strings.Join(params, ", "))
}

func (g *Generator) emitInit(w *cppWriter) {
	if g.Config.Target == "test" {
		w.open(fmt.Sprintf("void %s::__init(){", g.ClassName))
	} else {
		w.open(fmt.Sprintf("void %s::__init() {", g.ClassName))
	}
	if len(g.Mod.InitialActions) > 0 {
		for _, act := range g.Mod.InitialActions {
			g.emitAction(w, act)
		}
		w.close("")
		if g.Config.Target != "test" {
			w.blank()
		}
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
	if g.Config.Target != "test" {
		w.blank()
	}
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
func (g *Generator) emitSomeAction(w *cppWriter, name string, act goivy.Action) {
	openSig := g.methodSignature(name, act, true, false) + " {"
	if g.Config.Target == "test" {
		openSig = g.methodSignature(name, act, true, false) + "{"
	}
	w.open(openSig)
	returns := act.GetFormalReturns()
	_, rtypes := g.getParamTypes(name, act)
	// Python emit_some_action (ivy_to_cpp.py:1604-1607): for imported
	// actions in test target, emit a `< name(args)` trace prologue at
	// the top of the body, and (when opt_trace) wrap the body in `{`
	// and `}` braces.
	traceImportCaller := g.importCallers()[name]
	if traceImportCaller {
		g.emitTraceActionPrologue(w, name, act.GetFormalParams())
		if g.Config.Trace {
			w.linef(`__ivy_out%s << "{" << std::endl;`, g.numberFormat())
		}
	}
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
		w.linef("%s %s;", g.cppQualifiedType(returns[0].CSort, g.ClassName), varName(returns[0].Name))
		g.mkNondetSym(w, returns[0], returns[0].Name, 0)
	}
	g.emitAction(w, act)
	g.currentReturns = prevReturns
	if traceImportCaller && g.Config.Trace {
		w.linef(`__ivy_out%s << "}" << std::endl;`, g.numberFormat())
	}
	if len(returns) >= 1 && !firstIsReturnRef {
		w.linef("return %s;", varName(returns[0].Name))
	}
	w.close("")
	if g.Config.Target != "test" {
		w.blank()
	}
}

// importCallers mirrors Python find_import_callers (ivy_to_cpp.py:1888-1897).
// For target=test, imported wrappers are named `imp__foo`; Python strips that
// prefix and records both `foo` and `ext:foo` so the trace lands on the caller
// action before precondition assertions. Direct non-isolate imports are kept as
// their bare names for unit-level Generate callers.
func (g *Generator) importCallers() map[string]bool {
	if g.importCallersCache != nil {
		return g.importCallersCache
	}
	out := map[string]bool{}
	if g.Config.Target == "test" && g.Mod != nil {
		for _, imp := range g.Mod.Imports {
			impDef, ok := imp.(*goivy.ImportDef)
			if !ok {
				continue
			}
			if atom, ok := impDef.Scope.(*goivy.Atom); ok && atom.Relname() != "" {
				continue
			}
			name := ""
			if atom, ok := impDef.Imported.(*goivy.Atom); ok {
				name = atom.Relname()
			}
			if name == "" {
				continue
			}
			if _, ok := g.Mod.Actions.Get2(name); !ok {
				continue
			}
			caller := name
			if strings.HasPrefix(caller, "imp__") {
				caller = strings.TrimPrefix(caller, "imp__")
			} else {
				caller = strings.TrimPrefix(caller, "ext:")
			}
			out["ext:"+caller] = true
			out[caller] = true
		}
	}
	g.importCallersCache = out
	return out
}

// emitTraceActionPrologue emits Python trace_action (ivy_to_cpp.py:1576-1590):
//
//	__ivy_out << "< name" << "(" << p0 << "," << p1 << ")" << std::endl;
//
// The leading `ext:` is stripped per Python lines 1578-1579. The
// number_format prefix (g.numberFormat()) is injected between
// `__ivy_out` and the first `<<` literal to match Python.
func (g *Generator) emitTraceActionPrologue(w *cppWriter, name string, formals []*goivy.Const) {
	display := strings.TrimPrefix(name, "ext:")
	var b strings.Builder
	if g.Config.Target == "test" {
		b.WriteString(fmt.Sprintf(`__ivy_out%s  << "< %s"`, g.numberFormat(), display))
	} else {
		b.WriteString(fmt.Sprintf(`__ivy_out%s << "< %s"`, g.numberFormat(), display))
	}
	if len(formals) > 0 {
		b.WriteString(` << "("`)
		for i, p := range formals {
			if i > 0 {
				b.WriteString(` << ","`)
			}
			b.WriteString(fmt.Sprintf(" << %s", varName(p.Name)))
		}
		b.WriteString(` << ")"`)
	}
	b.WriteString(" << std::endl;")
	w.line(b.String())
}

// numberFormat returns the std::ostream manipulator prefix that Python
// inserts after `__ivy_out` on every trace line. Python sets
// `number_format = ' << std::hex << std::showbase '` when the module
// attribute `radix == "16"`, else the empty string
// (ivy_to_cpp.py:1935-1938). Cached on Generator.
func (g *Generator) numberFormat() string {
	if g.numberFormatComputed {
		return g.numberFormatCache
	}
	g.numberFormatComputed = true
	if g.Mod == nil {
		return ""
	}
	val, ok := g.Mod.Attributes["radix"]
	if !ok {
		return ""
	}
	rep := ""
	switch v := val.(type) {
	case string:
		rep = v
	case interface{ Relname() string }:
		rep = v.Relname()
	case goivy.Expr:
		rep = string(v.Sexp())
	}
	if rep == "16" {
		g.numberFormatCache = " << std::hex << std::showbase"
	}
	return g.numberFormatCache
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
	if len(g.Mod.Params) == 0 {
		g.emitPythonZeroParamTestMain(w)
		return
	}
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
	w.line("ivy.__unlock();")
	w.line("initializing = false;")
	g.emitRuntimeBindReaders(w)
	w.blank()
	g.emitTestLoopBody(w)
	w.close("")
	w.line("return 0;")
	w.close("")
}

func (g *Generator) emitPythonZeroParamTestMain(w *cppWriter) {
	mainName := g.Config.MainName
	var genLines strings.Builder
	names := g.publicActionNamesSorted()
	initActions := g.initialMixinActionNames()
	totalweight := 0.0
	numGens := 0
	for _, name := range names {
		if initActions[name] || isFinalizeName(name) {
			continue
		}
		className := g.actionGeneratorClassName(name)
		weight := g.actionWeight(name)
		genLines.WriteString(fmt.Sprintf("        generators.push_back(new %s(ivy));\n", className))
		genLines.WriteString(fmt.Sprintf("        weights.push_back(%s);\n", pythonFloatLiteral(weight)))
		totalweight += weight
		numGens++
	}
	finalizeLine := ""
	if g.hasFinalizeExport() {
		finalizeLine = "    ivy.__lock(); ivy.ext___finalize(); ivy.__unlock();\n"
	}
	w.raw(fmt.Sprintf(`

int %s(int argc, char **argv){
        int test_iters = %s;
        int runs = %s;

    int seed = 1;
    std::uint8_t seed32[chacha8c::key_size] = {0};
    std::memcpy(seed32, &seed, sizeof(seed));
    __chacha8c_rng.Seed(seed32);

    int sleep_ms = 10;
    int final_ms = 0; 
    
    std::vector<char *> pargs; // positional args
    pargs.push_back(argv[0]);
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        size_t p = arg.find('=');
        if (p == std::string::npos)
            pargs.push_back(argv[i]);
        else {
            std::string param = arg.substr(0,p);
            std::string value = arg.substr(p+1);

            if (param == "out") {
                __ivy_out.open(value.c_str());
                if (!__ivy_out) {
                    std::cerr << "cannot open to write: " << value << std::endl;
                    return 1;
                }
            }
            else if (param == "iters") {
                test_iters = atoi(value.c_str());
            }
            else if (param == "runs") {
                runs = atoi(value.c_str());
            }
            else if (param == "seed") {
                seed = atoi(value.c_str());
            }
            else if (param == "delay") {
                sleep_ms = atoi(value.c_str());
            }
            else if (param == "wait") {
                final_ms = atoi(value.c_str());
            }
            else if (param == "modelfile") {
                __ivy_modelfile.open(value.c_str());
                if (!__ivy_modelfile) {
                    std::cerr << "cannot open to write: " << value << std::endl;
                    return 1;
                }
            }
            else {
                std::cerr << "unknown option: " << param << std::endl;
                return 1;
            }
        }
    }
    srand(seed);
    if (!__ivy_out.is_open())
        __ivy_out.basic_ios<char>::rdbuf(std::cout.rdbuf());
    argc = pargs.size();
    argv = &pargs[0];
    if (argc == 2){
        argc--;
        int fd = _open(argv[argc],0);
        if (fd < 0){
            std::cerr << "cannot open to read: " << argv[argc] << "\n";
            __ivy_exit(1);
        }
        _dup2(fd, 0);
    }
    if (argc != 1){
        std::cerr << "usage: %s \n";
        __ivy_exit(1);
    }
    std::vector<std::string> args;
    std::vector<ivy_value> arg_values(0);
    for(int i = 1; i < argc;i++){args.push_back(argv[i]);}

#ifdef _WIN32
    // Boilerplate from windows docs

    {
        WORD wVersionRequested;
        WSADATA wsaData;
        int err;

    /* Use the MAKEWORD(lowbyte, highbyte) macro declared in Windef.h */
        wVersionRequested = MAKEWORD(2, 2);

        err = WSAStartup(wVersionRequested, &wsaData);
        if (err != 0) {
            /* Tell the user that we could not find a usable */
            /* Winsock DLL.                                  */
            printf("WSAStartup failed with error: %%d\n", err);
            return 1;
        }

    /* Confirm that the WinSock DLL supports 2.2.*/
    /* Note that if the DLL supports versions greater    */
    /* than 2.2 in addition to 2.2, it will still return */
    /* 2.2 in wVersion since that is the version we      */
    /* requested.                                        */

        if (LOBYTE(wsaData.wVersion) != 2 || HIBYTE(wsaData.wVersion) != 2) {
            /* Tell the user that we could not find a usable */
            /* WinSock DLL.                                  */
            printf("Could not find a usable version of Winsock.dll\n");
            WSACleanup();
            return 1;
        }
    }
#endif
    for(int runidx = 0; runidx < runs; runidx++) {
    initializing = true;
    %s_repl ivy;
    for(unsigned i = 0; i < argc; i++) {ivy.__argv.push_back(argv[i]);}
    ivy._generating = false;

        ivy.__unlock();
        initializing = false;
        for(int rdridx = 0; rdridx < readers.size(); rdridx++) {
            readers[rdridx]->bind();
        }
                    
        init_gen my_init_gen(ivy);
        my_init_gen.generate(ivy);
        std::vector<gen *> generators;
        std::vector<double> weights;

%s        double totalweight = %s;
        int num_gens = %d;


#ifdef _WIN32
    LARGE_INTEGER freq;
    QueryPerformanceFrequency(&freq);
#endif
    double frnd = 0.0;
    bool do_over = false;
    for(int cycle = 0; cycle < test_iters; cycle++) {

//        std::cout << "totalweight = " << totalweight << std::endl;
//        double choices = totalweight + readers.size() + timers.size();
        double choices = totalweight + 5.0;
        if (do_over) {
           do_over = false;
        }  else {
            frnd = choices * (((double)rand())/(((double)RAND_MAX)+1.0));
        }
        // std::cout << "frnd = " << frnd << std::endl;
        if (frnd < totalweight) {
            int idx = 0;
            double sum = 0.0;
            while (idx < num_gens-1) {
                sum += weights[idx];
                if (frnd < sum)
                    break;
                idx++;
            }
            gen &g = *generators[idx];
            ivy.__lock();
#ifdef _WIN32
            LARGE_INTEGER before;
            QueryPerformanceCounter(&before);
#endif
            ivy._generating = true;
            bool sat = g.generate(ivy);
#ifdef _WIN32
            LARGE_INTEGER after;
            QueryPerformanceCounter(&after);
//            __ivy_out << "idx: " << idx << " sat: " << sat << " time: " << (((double)(after.QuadPart-before.QuadPart))/freq.QuadPart) << std::endl;
#endif
            if (sat){
                g.execute(ivy);
                ivy._generating = false;
                ivy.__unlock();
#ifdef _WIN32
                Sleep(sleep_ms);
#endif
            }
            else {
                ivy._generating = false;
                ivy.__unlock();
                cycle--;
            }
            continue;
        }


        fd_set rdfds;
        FD_ZERO(&rdfds);
        int maxfds = 0;

        for (unsigned i = 0; i < readers.size(); i++) {
            reader *r = readers[i];
            int fds = r->fdes();
            if (fds >= 0) {
                FD_SET(fds,&rdfds);
            }
            if (fds > maxfds)
                maxfds = fds;
        }

#ifdef _WIN32
        int timer_min = 15;
#else
        int timer_min = 5;
#endif

        struct timeval timeout;
        timeout.tv_sec = timer_min/1000;
        timeout.tv_usec = 1000 * (timer_min %% 1000);

#ifdef _WIN32
        int foo;
        if (readers.size() == 0){  // winsock can't handle empty fdset!
            Sleep(timer_min);
            foo = 0;
        }
        else
            foo = select(maxfds+1,&rdfds,0,0,&timeout);
#else
        int foo = select(maxfds+1,&rdfds,0,0,&timeout);
#endif

        if (foo < 0)
#ifdef _WIN32
            {std::cerr << "select failed: " << WSAGetLastError() << std::endl; __ivy_exit(1);}
#else
            {perror("select failed"); __ivy_exit(1);}
#endif
        
        if (foo == 0){
           // std::cout << "TIMEOUT\n";            
           cycle--;
           for (unsigned i = 0; i < timers.size(); i++){
               if (timer_min >= timers[i]->ms_delay()) {
                   cycle++;
                   break;
               }
           }
           for (unsigned i = 0; i < timers.size(); i++)
               timers[i]->timeout(timer_min);
        }
        else {
            int fdc = 0;
            for (unsigned i = 0; i < readers.size(); i++) {
                reader *r = readers[i];
                if (FD_ISSET(r->fdes(),&rdfds))
                    fdc++;
            }
            // std::cout << "fdc = " << fdc << std::endl;
            int fdi = fdc * (((double)rand())/(((double)RAND_MAX)+1.0));
            fdc = 0;
            for (unsigned i = 0; i < readers.size(); i++) {
                reader *r = readers[i];
                if (FD_ISSET(r->fdes(),&rdfds)) {
                    if (fdc == fdi) {
                        // std::cout << "reader = " << i << std::endl;
                        r->read();
                        if (r->background()) {
                           cycle--;
                           do_over = true;
                        }
                        break;
                    }
                    fdc++;

                }
            }
        }            
    }
%s    
#ifdef _WIN32
                Sleep(final_ms);  // HACK: wait for late responses
#endif
    __ivy_out << "test_completed" << std::endl;
    if (runidx == runs-1) {
        struct timespec ts;
        int ms = 50;
        ts.tv_sec = ms/1000;
        ts.tv_nsec = (ms %% 1000) * 1000000;
        nanosleep(&ts,NULL);
        exit(0);
    }
    for (unsigned i = 0; i < readers.size(); i++)
        delete readers[i];
    readers.clear();
    for (unsigned i = 0; i < timers.size(); i++)
        delete timers[i];
    timers.clear();


    }
    return 0;
}
`, mainName, g.Config.TestIters, g.Config.TestRuns, g.ClassName, g.ClassName, genLines.String(), pythonFloatLiteral(totalweight), numGens, finalizeLine))
}

// emitTestLoopBody emits the body of the per-run test driver, mirroring
// Python `emit_repl_boilerplate3test` (ivy_to_cpp.py:4265-4467):
// build init_gen, weighted action generators, then loop test_iters
// times choosing among generators / readers / timers via select().
func (g *Generator) emitTestLoopBody(w *cppWriter) {
	w.line("init_gen my_init_gen(ivy);")
	w.line("my_init_gen.generate(ivy);")
	w.line("std::vector<gen *> generators;")
	w.line("std::vector<double> weights;")
	w.blank()
	names := g.publicActionNamesSorted()
	initActions := g.initialMixinActionNames()
	totalweight := 0.0
	numGens := 0
	for _, name := range names {
		if initActions[name] || isFinalizeName(name) {
			continue
		}
		className := g.actionGeneratorClassName(name)
		w.linef("generators.push_back(new %s(ivy));", className)
		weight := g.actionWeight(name)
		w.linef("weights.push_back(%s);", pythonFloatLiteral(weight))
		totalweight += weight
		numGens++
	}
	w.linef("double totalweight = %s;", pythonFloatLiteral(totalweight))
	w.linef("int num_gens = %d;", numGens)
	w.blank()
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
	w.line("#ifdef _WIN32")
	w.line("Sleep(final_ms);")
	w.line("#endif")
	w.line(`__ivy_out << "test_completed" << std::endl;`)
	w.open("if (runidx == runs - 1) {")
	w.line("struct timespec ts;")
	w.line("int ms = 50;")
	w.line("ts.tv_sec = ms / 1000;")
	w.line("ts.tv_nsec = (ms % 1000) * 1000000;")
	w.line("nanosleep(&ts, NULL);")
	w.line("exit(0);")
	w.close("")
	if g.Config.Target == "test" {
		w.line("for (unsigned i = 0; i < readers.size(); i++)")
		w.indent++
		w.line("delete readers[i];")
		w.indent--
	} else {
		w.open("for (unsigned i = 0; i < readers.size(); i++) {")
		w.line("delete readers[i];")
		w.close("")
	}
	w.line("readers.clear();")
	if g.Config.Target == "test" {
		w.line("for (unsigned i = 0; i < timers.size(); i++)")
		w.indent++
		w.line("delete timers[i];")
		w.indent--
	} else {
		w.open("for (unsigned i = 0; i < timers.size(); i++) {")
		w.line("delete timers[i];")
		w.close("")
	}
	w.line("timers.clear();")
}

func (g *Generator) emitTestLoopGenBranch(w *cppWriter) {
	w.line("int idx = 0;")
	w.line("double sum = 0.0;")
	w.open("while (idx < num_gens-1) {")
	w.line("sum += weights[idx];")
	w.line("if (frnd < sum)")
	w.indent++
	w.line("break;")
	w.indent--
	w.line("idx++;")
	w.close("")
	w.line("gen &g = *generators[idx];")
	w.line("ivy.__lock();")
	w.line("#ifdef _WIN32")
	w.line("LARGE_INTEGER before;")
	w.line("QueryPerformanceCounter(&before);")
	w.line("#endif")
	w.line("ivy._generating = true;")
	w.line("bool sat = g.generate(ivy);")
	w.line("#ifdef _WIN32")
	w.line("LARGE_INTEGER after;")
	w.line("QueryPerformanceCounter(&after);")
	w.line("#endif")
	w.open("if (sat){")
	w.line("g.execute(ivy);")
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
	if g.Config.Target == "test" {
		w.close("")
		w.line("else")
		w.indent++
		w.line("foo = select(maxfds + 1, &rdfds, 0, 0, &timeout);")
		w.indent--
	} else {
		w.close(" else {")
		w.indent++
		w.line("foo = select(maxfds + 1, &rdfds, 0, 0, &timeout);")
		w.indent--
		w.line("}")
	}
	w.line("#else")
	w.line("int foo = select(maxfds + 1, &rdfds, 0, 0, &timeout);")
	w.line("#endif")
	if g.Config.Target == "test" {
		w.line("if (foo < 0)")
		w.line("#ifdef _WIN32")
		w.line(`{std::cerr << "select failed: " << WSAGetLastError() << std::endl; __ivy_exit(1);}`)
		w.line("#else")
		w.line(`{perror("select failed"); __ivy_exit(1);}`)
		w.line("#endif")
	} else {
		w.open("if (foo < 0) {")
		w.line("#ifdef _WIN32")
		w.line(`std::cerr << "select failed: " << WSAGetLastError() << std::endl; __ivy_exit(1);`)
		w.line("#else")
		w.line(`perror("select failed"); __ivy_exit(1);`)
		w.line("#endif")
		w.close("")
	}
	w.open("if (foo == 0) {")
	w.line("cycle--;")
	w.open("for (unsigned i = 0; i < timers.size(); i++) {")
	w.open("if (timer_min >= timers[i]->ms_delay()) {")
	w.line("cycle++;")
	w.line("break;")
	w.close("")
	w.close("")
	if g.Config.Target == "test" {
		w.line("for (unsigned i = 0; i < timers.size(); i++)")
		w.indent++
		w.line("timers[i]->timeout(timer_min);")
		w.indent--
	} else {
		w.open("for (unsigned i = 0; i < timers.size(); i++) {")
		w.line("timers[i]->timeout(timer_min);")
		w.close("")
	}
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

// isFinalizeName matches Python's special generated-tester finalizer hook.
// A bare action named "_finalize" is just an ordinary action; only the
// external action name gets substituted into the target=test FINALIZE slot.
func isFinalizeName(name string) bool {
	return name == "ext:_finalize"
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

func pythonFloatLiteral(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	lower := strings.ToLower(s)
	if strings.Contains(s, ".") || strings.Contains(lower, "e") || strings.Contains(lower, "nan") || strings.Contains(lower, "inf") {
		return s
	}
	return s + ".0"
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
	if g.Config.Target != "test" {
		w.line("(void)test_iters;")
		w.line("(void)runs;")
	}
}

func (g *Generator) emitGeneratorInvocations(w *cppWriter) {
	w.line("init_gen my_init_gen(ivy);")
	w.line("my_init_gen.generate(ivy);")
	initActions := g.initialMixinActionNames()
	for name := range g.Mod.PublicActions.All() {
		if initActions[name] || isFinalizeName(name) {
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
