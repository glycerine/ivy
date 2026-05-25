package ivy2go

import (
	"errors"
	"fmt"
	"go/format"
	"path/filepath"
	"sort"
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
	currentReturns     []*goivy.Const
	extRel             map[string]bool
	nativeOnceMemo     map[string]bool
	encodedSorts       map[string]bool
	importCallersCache map[string]bool

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

		thunkMemo:          map[string]string{},
		exprAliases:        map[string]goivy.Expr{},
		extRel:             map[string]bool{},
		nativeOnceMemo:     map[string]bool{},
		encodedSorts:       map[string]bool{},
		importCallersCache: map[string]bool{},
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
	g.emitInit()
	g.emitRuntime()
	g.emitNondet()
	g.emitExtensional()
	g.emitDefinitions()
	g.emitThunks()
	g.emitNative()
	if g.Config.Target == "repl" || g.Config.Target == "test" {
		g.emitRepl()
	}
	if g.Config.EmitMain {
		g.emitMain()
	}
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
	g.state.linef("// state.go: %s struct + NewState. Filled in by M5.", g.StateTypeName)
	g.state.linef("type %s struct{}", g.StateTypeName)
	g.state.blank()
	g.state.linef("func New%s() *%s { return &%s{} }", g.StateTypeName, g.StateTypeName, g.StateTypeName)
}

func (g *Generator) emitActions() {
	g.actions.line("// actions.go: action methods on *State. Filled in by M4.")
}

func (g *Generator) emitInit() {
	g.init.line("// init.go: (*State).Init body. Filled in by M5.")
	g.init.linef("func (s *%s) Init() {}", g.StateTypeName)
}

func (g *Generator) emitRuntime() {
	g.runtime.line("// runtime.go: package-level helpers. Filled in by M3+.")
}

func (g *Generator) emitNondet() {
	g.nondet.line("// nondet.go: nondeterministic choice helpers. Filled in by M4.")
}

func (g *Generator) emitExtensional() {
	g.extensional.line("// extensional.go: extensional relation iteration. Filled in by M4.")
}

func (g *Generator) emitDefinitions() {
	g.definitions.line("// definitions.go: definitional axioms. Filled in by M4.")
}

func (g *Generator) emitThunks() {
	g.thunks.line("// thunk.go: hash-thunk struct definitions. Filled in by M8.")
}

func (g *Generator) emitNative() {
	g.native.line("// native.go: user-supplied native Go blocks. Filled in by M10.")
}

func (g *Generator) emitRepl() {
	g.repl.line("// repl.go: REPL command reader. Filled in by M7.")
}

func (g *Generator) emitMain() {
	g.main.line("// main.go: program entry point. Filled in by M5+.")
	g.main.open("func main() {")
	g.main.linef("_ = New%s()", g.StateTypeName)
	g.main.close("")
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
