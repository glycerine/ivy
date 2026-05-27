package ivy2go

import (
	"errors"
	"fmt"
	"go/format"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Generator is the per-Module Go code generator. Mirrors
// ivy2cpp/generator.go Generator with C++-specific fields renamed for
// Go output (header/impl pair replaced by per-stream goWriters held in
// Ctx).
type Generator struct {
	Mod           *goivy.Module
	Config        Config
	BaseName      string
	PackageName   string
	StateTypeName string

	Ctx *GoContext

	// One goWriter per output .go file. Each writer's underlying
	// stream lives on Ctx so the finalize step can rebuild the file
	// in import-block + body order.
	types       goWriter
	state       goWriter
	actions     goWriter
	init        goWriter
	runtime     goWriter
	nondet      goWriter
	extensional goWriter
	definitions goWriter
	thunks      goWriter
	native      goWriter
	repl        goWriter
	main        goWriter

	// thunkCtr names anonymous thunk structs (Python thunk_counter,
	// ivy_to_cpp.py:504; see ivy2cpp/generator.go thunkCtr). Each
	// distinct file-scope thunk allocates `thunk__N` and increments.
	thunkCtr int

	// Memoization caches mirroring ivy2cpp/generator.go. These are
	// declared here so subsequent milestones (M2+) can populate them
	// in place without ad-hoc additions to the struct.
	thunkMemo          map[string]string
	exprAliases        map[string]goivy.Expr
	reifyExprCodeAlias map[string]string
	currentReturns     []*goivy.Const
	extRel             map[string]bool
	nativeOnceMemo     map[string]bool
	encodedSorts       map[string]bool
	importCallersCache map[string]bool
	nondetSortStack    map[string]int
	nondetTempCtr      int

	// numberFormatCache memoizes numberFormat() so we don't re-scan
	// module attributes on every trace line. Empty string until
	// computed; readers should call numberFormat() which initializes
	// it on first use.
	numberFormatCache    string
	numberFormatComputed bool

	// lhsContext is set transiently by emitAssign / emitAssignTwoPhase
	// while emitting the LHS expression so goStorageAccess produces
	// an assignable form (raw map index) rather than the getter
	// call. OPEN 061.1.
	lhsContext bool

	errs []error
}

// Generate compiles a goivy.Module into a directory of Go source files.
// Counterpart to ivy2cpp/generator.go Generate.
func Generate(mod *goivy.Module, cfg Config) (*Output, error) {
	if mod == nil {
		return nil, fmt.Errorf("ivy2go: nil module")
	}
	var err error
	cfg, _, err = normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	base := moduleBaseName(mod)
	pkg := cfg.PackageName
	if pkg == "" {
		pkg = goPackageName(base)
	}
	stateType := cfg.StateTypeName
	if stateType == "" {
		stateType = "State"
	}
	// `func main()` is legal only in `package main`. Silently disable
	// EmitMain for any other package to keep emitted files buildable.
	if cfg.EmitMain && pkg != "main" {
		cfg.EmitMain = false
	}

	g := &Generator{
		Mod:           mod,
		Config:        cfg,
		BaseName:      base,
		PackageName:   pkg,
		StateTypeName: stateType,
		Ctx:           NewGoContext(),

		thunkMemo:      map[string]string{},
		exprAliases:    map[string]goivy.Expr{},
		extRel:         map[string]bool{},
		nativeOnceMemo: map[string]bool{},
		encodedSorts:   map[string]bool{},
		// IMPORTANT: importCallersCache is INTENTIONALLY left
		// nil so its first lookup triggers the lazy computation
		// in importCallers(). Pre-populating with an empty map
		// would short-circuit the cache check and permanently
		// hide every import-caller (the trace `< name` lines
		// would never appear).
	}
	g.Ctx.PackageName = pkg
	g.types = newGoWriter(g.Ctx.Types)
	g.state = newGoWriter(g.Ctx.State)
	g.actions = newGoWriter(g.Ctx.Actions)
	g.init = newGoWriter(g.Ctx.Init)
	g.runtime = newGoWriter(g.Ctx.Runtime)
	g.nondet = newGoWriter(g.Ctx.Nondet)
	g.extensional = newGoWriter(g.Ctx.Extensional)
	g.definitions = newGoWriter(g.Ctx.Definitions)
	g.thunks = newGoWriter(g.Ctx.Thunks)
	g.native = newGoWriter(g.Ctx.Native)
	g.repl = newGoWriter(g.Ctx.Repl)
	g.main = newGoWriter(g.Ctx.Main)

	applySessionParameters(mod, cfg)
	prepareModuleForGo(mod, cfg)

	if err := g.generate(); err != nil {
		return nil, err
	}
	files, ferr := g.finalize()
	if ferr != nil {
		return nil, ferr
	}
	out := &Output{
		Files:           files,
		BaseName:        base,
		PackageName:     pkg,
		StateTypeName:   stateType,
		Target:          cfg.RequestedTarget,
		EffectiveTarget: cfg.Target,
		EmitMain:        cfg.EmitMain,
		Config:          cfg,
		LibSpecs:        moduleLibSpecs(mod),
	}
	return out, nil
}

// generate orchestrates emission. Mirrors ivy2cpp/generator.go generate
// but the C++ split into emitHeader + emitImpl is fanned out across
// per-file emit* methods, each writing into its dedicated stream.
func (g *Generator) generate() error {
	if err := g.checkMemberNames(); err != nil {
		return err
	}
	g.emitTypes()
	g.emitState()
	g.emitActions()
	g.emitInitStream()
	g.emitRuntime()
	g.emitNondet()
	g.emitExtensional()
	g.emitDefinitionsStream()
	g.emitThunks()
	g.emitNative()
	// Tick lives in the runtime stream and is needed by every target
	// (test/gen/repl call it; impl/class can ignore it without harm).
	g.emitTick(&g.runtime)
	// REPL helpers are only useful for the repl target. The test
	// target drives actions via actionGen_* in main.go, not by
	// reading commands from stdin — emitting repl.go alongside
	// test would leave unused (and potentially type-broken)
	// dispatch code.
	if g.Config.RequestedTarget != "class" && g.Config.Target == "repl" {
		g.emitRepl()
	}
	if g.Config.EmitMain {
		g.emitMain()
	}
	// Conditional runtime helpers — emitted LAST so any earlier
	// emit step (including emitMain) can request them via
	// Ctx.OnceGlobals.
	g.emitRuntimeHelpersLate(&g.runtime)
	return errors.Join(g.errs...)
}

// checkMemberNames mirrors ivy2cpp/generator.go checkMemberNames
// (Python ivy_to_cpp.py:1830-1834). In Go we forbid the configured
// PackageName from colliding with any symbol name after mangling.
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
	if names[varName(g.PackageName)] {
		return fmt.Errorf("ivy2go: package name %q collides with module symbol", g.PackageName)
	}
	return nil
}

// The emit* methods below are M1 stubs: each writes a marker comment
// into its stream so finalize can produce a non-empty, gofmt-clean
// file when the stream is required. Subsequent milestones (M2-M10)
// replace each stub with real emission logic.

func (g *Generator) emitTypes() {
	g.emitSortDecls(&g.types)
	g.emitConstructorDecls(&g.types)
}

func (g *Generator) emitState() {
	g.emitStateStruct(&g.state)
	g.emitNewState(&g.state)
}

func (g *Generator) emitActions() {
	g.emitMethods(&g.actions)
	g.emitDefinitions(&g.actions)
	g.emitConstructors(&g.actions)
	g.emitActionGenStructs(&g.actions)
}

func (g *Generator) emitInitStream() {
	g.emitInit(&g.init)
}

func (g *Generator) emitRuntime() {
	g.emitRuntimeHelpers(&g.runtime)
	g.emitZ3Support(&g.runtime)
	g.emitSetSolver(&g.runtime)
}

func (g *Generator) emitNondet() {
	g.nondet.line("// nondet.go: nondeterministic choice helpers. Filled in by M4.")
}

func (g *Generator) emitExtensional() {
	g.extensional.line("// extensional.go: extensional relation iteration. Filled in by M4.")
}

func (g *Generator) emitDefinitionsStream() {
	g.definitions.line("// definitions.go: per-stream marker. Definition methods are")
	g.definitions.line("// emitted into the actions stream via emitDefinitions(w).")
}

func (g *Generator) emitThunks() {
	g.thunks.line("// thunk.go: hash-thunk struct definitions. Filled in by M8.")
}

func (g *Generator) emitNative() {
	g.emitNativeBlocks()
}

func (g *Generator) emitRepl() {
	g.emitReplLoop(&g.repl)
	// Tick is emitted from the top-level pipeline (above) for every
	// target, so REPL doesn't re-emit it.
}

func (g *Generator) emitMain() {
	if g.Config.RequestedTarget == "class" {
		// class target produces a library; no main.
		return
	}
	switch g.Config.Target {
	case "repl":
		g.emitReplMain()
	case "test", "gen":
		g.emitTestMain()
	default:
		// impl (and anything else): plain construct + init + exit.
		g.main.open("func main() {")
		stateArgs := g.emitModuleParamSetup()
		g.main.linef("state := New%s(%s)", g.StateTypeName, strings.Join(stateArgs, ", "))
		g.main.line("state.Init()")
		g.main.line("_ = state")
		g.main.close("")
	}
}

// emitReplMain emits a main() that reads commands from stdin via
// runRepl. Used by target=repl.
func (g *Generator) emitReplMain() {
	g.Ctx.AddImport("main", "fmt", "")
	g.Ctx.AddImport("main", "os", "")
	g.main.open("func main() {")
	stateArgs := g.emitModuleParamSetup()
	g.main.linef("state := New%s(%s)", g.StateTypeName, strings.Join(stateArgs, ", "))
	g.main.line("state.Init()")
	g.main.line("if err := runRepl(state, os.Stdin, os.Stdout); err != nil {")
	g.main.line(`	fmt.Fprintln(os.Stderr, err)`)
	g.main.line(`	os.Exit(1)`)
	g.main.line("}")
	g.main.close("")
}

// emitTestMain emits a main() that drives a randomized test loop
// using the per-action actionGen_* structs. Mirrors the role of
// ivy2cpp/generator.go emitTestLoopBody (Python
// emit_repl_boilerplate3test).
//
// Shape:
//
//	func main() {
//	    iters := parseTestItersFlag(<default>)
//	    state := NewState(<module params...>)
//	    state.Init()
//
//	    type actionEntry struct {
//	        name string
//	        gen  interface{ Generate(*State); Close() error }
//	    }
//	    var actions []actionEntry
//	    actions = append(actions, actionEntry{"set_flag", newactionGen_SetFlag(state)})
//	    …
//	    defer func() { for _, a := range actions { _ = a.gen.Close() } }()
//
//	    for i := 0; i < iters; i++ {
//	        idx := ivyChoose(len(actions))
//	        actions[idx].gen.Generate(state)
//	    }
//	    fmt.Println("test_completed")
//	}
func (g *Generator) emitTestMain() {
	// Request parseTestItersFlag up front so emitRuntimeHelpersLate
	// emits it regardless of whether the empty-actions branch fires.
	g.Ctx.OnceGlobals["__need_testflags"] = true
	g.Ctx.AddImport("main", "fmt", "")

	names := g.actionGenNames()

	g.main.open("func main() {")
	g.main.line("applyTestSeedFlag()")
	g.main.linef("iters := parseTestItersFlag(%s)", g.Config.TestIters)
	stateArgs := g.emitModuleParamSetup()
	g.main.linef("state := New%s(%s)", g.StateTypeName, strings.Join(stateArgs, ", "))
	g.main.line("state.Init()")
	g.main.blank()
	if len(names) == 0 {
		// Module has no public actions to drive — just complete.
		g.main.line("_ = iters")
		g.main.line(`fmt.Println("test_completed")`)
		g.main.close("")
		return
	}
	// Action generator slice. Each entry pairs a name (for traces)
	// with the per-action generate(state)bool + execute(state) pair
	// (mirrors ivy2cpp's two-phase generate→execute split).
	g.main.line("// Per-action generator registry.")
	g.main.line("type actionEntry struct {")
	g.main.line("\tname     string")
	g.main.linef("\tgenerate func(*%s) bool", g.StateTypeName)
	g.main.linef("\texecute  func(*%s)", g.StateTypeName)
	g.main.line("\tclose    func() error")
	g.main.line("\tweight   float64")
	g.main.line("}")
	g.main.line("var actions []actionEntry")
	for _, name := range names {
		struc := "actionGen_" + goExportedName(name)
		w := g.actionWeight(name)
		g.main.linef("{")
		g.main.linef("\tg := new%s(state)", struc)
		g.main.linef("\tactions = append(actions, actionEntry{name: %q, generate: g.generate, execute: g.execute, close: g.Close, weight: %g})", name, w)
		g.main.line("}")
	}
	g.main.line("defer func() {")
	g.main.line("\tfor _, a := range actions {")
	g.main.line("\t\t_ = a.close()")
	g.main.line("\t}")
	g.main.line("}()")
	g.main.blank()
	// Mirrors cpp's test-loop structure (ivy2cpp/generator.go:1685+):
	//   choices = totalweight + 5.0   // 5 reserved for IO/select branches
	//   frnd    = choices * Rand() / (RAND_MAX+1.0)
	//   if frnd < totalweight: pick action by prefix-scan, fire
	//   else: cycle-- (select branch is a no-op when no readers)
	//   if !sat after fire: cycle--
	// The +5.0 and rejection-sampling pattern are what make ivy2cpp's
	// random consumption sequence match ivy_to_cpp's — keeping them
	// here is the only way ivy2go's stream stays byte-equivalent under
	// the same chacha8 seed.
	g.main.line("var __totalW float64")
	g.main.line("for _, a := range actions { __totalW += a.weight }")
	g.main.line("__choices := __totalW + 5.0")
	g.main.line("var __frnd float64")
	g.main.line("__doOver := false")
	g.main.line("for i := 0; i < iters; i++ {")
	g.main.line("\t_ = i")
	g.main.line("\tif __doOver {")
	g.main.line("\t\t__doOver = false")
	g.main.line("\t} else {")
	g.main.line("\t\t__frnd = __choices * float64(ivyRand31()) / ivyRandMaxPlus1")
	g.main.line("\t}")
	g.main.line("\tif __frnd < __totalW {")
	g.main.line("\t\tidx := 0")
	g.main.line("\t\tacc := 0.0")
	g.main.line("\t\tfor j := 0; j < len(actions)-1; j++ {")
	g.main.line("\t\t\tacc += actions[j].weight")
	g.main.line("\t\t\tif __frnd < acc { idx = j; break }")
	g.main.line("\t\t\tidx = j + 1")
	g.main.line("\t\t}")
	g.main.line("\t\t// Solver-driven fire/skip: generate() returns false")
	g.main.line("\t\t// when the action's precondition is UNSAT in the")
	g.main.line("\t\t// current state — cpp retries with `cycle--`.")
	g.main.line("\t\tif actions[idx].generate(state) {")
	g.main.line("\t\t\tactions[idx].execute(state)")
	g.main.line("\t\t} else {")
	g.main.line("\t\t\ti--")
	g.main.line("\t\t}")
	g.main.line("\t} else {")
	g.main.line("\t\t// Select-branch no-op (no readers / timers in test")
	g.main.line("\t\t// fixtures). cpp does `select(...); if (foo==0) cycle--`.")
	g.main.line("\t\ti--")
	g.main.line("\t}")
	// cpp's test main does NOT call __tick per iter (no select branch")
	// invokes it for readerless fixtures). Progress/rely checks happen")
	// inside the action body instead. Don't call state.Tick here.")
	g.main.line("}")
	g.main.line(`fmt.Println("test_completed")`)
	g.main.close("")
}

// actionWeight mirrors ivy2cpp/generator.go actionWeight. Returns
// the action's user-declared scheduling weight (default 1.0), used
// by the test-loop random action picker. Reads `<action>.weight`
// from Mod.Attributes.
func (g *Generator) actionWeight(name string) float64 {
	username := strings.TrimPrefix(name, "ext:")
	if g.Mod == nil || g.Mod.Attributes == nil {
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

// finalize composes each stream into a complete .go source file
// (package clause + import block + body) and returns the map keyed by
// filename. Streams with no body content are skipped, except the few
// that ivy2go always emits (state.go, init.go, runtime.go) so the
// emitted package is always self-contained.
func (g *Generator) finalize() (map[string]string, error) {
	type streamEntry struct {
		name   string // basename in Output.Files
		stream string // key in Ctx.Imports
		body   string
		always bool // emit even when body is whitespace-only
	}
	entries := []streamEntry{
		{"types.go", "types", g.Ctx.Types.GetFile(), false},
		{"state.go", "state", g.Ctx.State.GetFile(), true},
		{"actions.go", "actions", g.Ctx.Actions.GetFile(), false},
		{"init.go", "init", g.Ctx.Init.GetFile(), true},
		{"runtime.go", "runtime", g.Ctx.Runtime.GetFile(), true},
		{"nondet.go", "nondet", g.Ctx.Nondet.GetFile(), false},
		{"extensional.go", "extensional", g.Ctx.Extensional.GetFile(), false},
		{"definitions.go", "definitions", g.Ctx.Definitions.GetFile(), false},
		{"thunk.go", "thunks", g.Ctx.Thunks.GetFile(), false},
		{"native.go", "native", g.Ctx.Native.GetFile(), false},
		{"repl.go", "repl", g.Ctx.Repl.GetFile(), false},
		{"main.go", "main", g.Ctx.Main.GetFile(), false},
	}

	files := map[string]string{}
	for _, e := range entries {
		if !e.always && strings.TrimSpace(e.body) == "" {
			continue
		}
		// Auto-add imports when the body references their canonical
		// identifier. Cheaper than threading explicit imports
		// through every emitter that might touch them.
		if strings.Contains(e.body, "big.Int") {
			g.Ctx.AddImport(e.stream, "math/big", "")
		}
		if strings.Contains(e.body, "strconv.") {
			g.Ctx.AddImport(e.stream, "strconv", "")
		}
		var b strings.Builder
		b.WriteString("package ")
		b.WriteString(g.PackageName)
		b.WriteString("\n\n")
		if imp := g.Ctx.renderImportBlock(e.stream); imp != "" {
			b.WriteString(imp)
			b.WriteString("\n")
		}
		b.WriteString(e.body)
		if !strings.HasSuffix(e.body, "\n") {
			b.WriteByte('\n')
		}
		formatted, err := format.Source([]byte(b.String()))
		if err != nil {
			return nil, fmt.Errorf("ivy2go: emitted %s is not valid Go: %w\n---\n%s", e.name, err, b.String())
		}
		files[e.name] = string(formatted)
	}
	return files, nil
}

// moduleBaseName mirrors ivy2cpp/generator.go moduleBaseName.
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

// moduleLibSpecs mirrors ivy2cpp/compile.go moduleLibSpecs but is a
// no-op stub for now; later milestones can populate it if any Go
// tooling wants the spec list.
func moduleLibSpecs(mod *goivy.Module) []string {
	_ = mod
	return nil
}

// ensure deterministic stream ordering when finalize iterates. This is
// here (not in go_context) so the finalize step's order matches the
// stream slice above; sort is imported for normalizeConfig anyway.
var _ = sort.Strings
