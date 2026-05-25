package ivy2go

// Config holds generator options. Mirrors ivy2cpp/generator.go Config
// with two substitutions per ARCHITECTURE_TODO.md §3.2:
//   - ClassName     → PackageName  (Go has packages, not classes).
//   - Compiler      → deleted      (Go has one official toolchain).
//   - Stdafx        → deleted      (Windows PCH artefact).
//   - GoModule      → new          (emit go.mod when non-empty).
//   - GoivyImport*  → new          (lets tests/forks override the import path).
type Config struct {
	// Target is the resolved generation target: "impl", "class",
	// "repl", "test", or "gen". When empty, defaults to "gen" per
	// normalizeConfig (mirroring ivy2cpp).
	Target string

	// RequestedTarget is the original user-provided target string
	// before any normalization (e.g., "class" → target=repl with
	// EmitMain=false).
	RequestedTarget string

	// PackageName is the Go package name the emitted files declare.
	// If empty, defaults to varName(BaseName). Counterpart to
	// ivy2cpp Config.ClassName.
	PackageName string

	// StateTypeName is the exported name of the state struct in the
	// emitted package. Defaults to "State".
	StateTypeName string

	// MainName, OutDir, TestIters, TestRuns mirror ivy2cpp/Config
	// fields with the same semantics.
	MainName  string
	OutDir    string
	TestIters string
	TestRuns  string

	// Build, EmitMain, Trace mirror ivy2cpp/Config flags.
	Build    bool
	EmitMain bool
	Trace    bool

	// HostOS overrides the build-host detection. Empty string falls
	// back to runtime.GOOS; tests set "windows"/"linux" to exercise
	// both branches of any os-conditional emission.
	HostOS string

	// GoModule, when non-empty, is written as `module <path>` in a
	// generated go.mod. Empty (default) skips go.mod emission —
	// the caller is assumed to provide one (e.g., via go.work).
	GoModule string

	// GoivyImportPath is the import path the emitted programs use to
	// reach goivy at runtime. Defaults to
	// "github.com/glycerine/ivy/goivy" (see config.go DefaultGoivyImportPath).
	// Override for tests/forks.
	GoivyImportPath string
}

// DefaultGoivyImportPath is the canonical import path emitted programs
// use to reach goivy.
const DefaultGoivyImportPath = "github.com/glycerine/ivy/goivy"

// Output is the result of a single Generate call. Mirrors
// ivy2cpp/generator.go Output with the header/impl split replaced by a
// Files map keyed by filename (per ARCHITECTURE_TODO.md §3.2).
type Output struct {
	// Files is the set of emitted .go files keyed by basename. For
	// example, {"types.go": "...", "state.go": "..."}.
	Files map[string]string

	// BaseName is the module's base name (e.g., "pingpong" for
	// pingpong.ivy). Counterpart to ivy2cpp Output.BaseName.
	BaseName string

	// PackageName is the Go package name the emitted files declare.
	PackageName string

	// StateTypeName is the exported state struct name in the
	// emitted package.
	StateTypeName string

	// Target / EffectiveTarget mirror ivy2cpp.
	Target          string
	EffectiveTarget string

	// EmitMain indicates that a main.go was emitted.
	EmitMain bool

	// Config is the resolved configuration used during generation.
	Config Config

	// ExtraFiles carries auxiliary files (e.g., .dsc descriptors)
	// that live alongside the package directory rather than inside
	// it. Mirrors ivy2cpp Output.ExtraFiles.
	ExtraFiles map[string]string

	// LibSpecs is informational only (Go build does not consume it).
	// Carried for parity with ivy2cpp Output.LibSpecs and for any
	// future tooling that wants to record Ivy lib spec declarations.
	LibSpecs []string
}

// BatchOutput is the result of CompileAndGenerateAll when fanned out
// over multiple isolates. Mirrors ivy2cpp/compile.go BatchOutput.
type BatchOutput struct {
	Outputs    []*Output
	ExtraFiles map[string]string
	Config     Config
}
