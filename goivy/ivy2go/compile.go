package ivy2go

import (
	goJSON "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// CompileAndGenerate parses filename and generates a single isolate's
// output. Mirrors ivy2cpp/compile.go CompileAndGenerate.
func CompileAndGenerate(filename string, params map[string]string, cfg Config) (*Output, error) {
	batch, err := CompileAndGenerateAll(filename, params, cfg)
	if err != nil {
		return nil, err
	}
	if len(batch.Outputs) != 1 {
		return nil, fmt.Errorf("ivy2go: expected one generated output, got %d", len(batch.Outputs))
	}
	return batch.Outputs[0], nil
}

// CompileAndGenerateAll parses filename and fans out generation over
// every selected isolate. Mirrors ivy2cpp/compile.go
// CompileAndGenerateAll, with the C++-specific snapshot/restore dance
// reduced to the Go-side equivalent.
func CompileAndGenerateAll(filename string, params map[string]string, cfg Config) (*BatchOutput, error) {
	cfg, ivyParams, err := mergeParams(params, cfg)
	if err != nil {
		return nil, err
	}
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	if isolate := ivyParams["isolate"]; isolate != "" {
		mod.Cfg.Isolate = isolate
	}
	applySessionParameters(mod, cfg)
	sig := goivy.NewSigOn(mod.Cfg.IuCfg)
	if err := goivy.SourceFile(filename, mod, sig, map[string]interface{}{"create_isolate": false}); err != nil {
		return nil, err
	}
	isolates := selectedIsolates(mod, ivyParams["isolate"])
	if len(isolates) == 0 {
		isolates = []string{""}
	}
	batch := &BatchOutput{Config: cfg, ExtraFiles: map[string]string{}}
	for _, isolate := range isolates {
		isoMod := mod.Copy()
		if isoMod.Cfg == nil {
			isoMod.Cfg = goivy.NewConfig()
		}
		isoMod.Cfg.Isolate = isolate
		if cfg.Target == "repl" && languageVersionAtLeast(isoMod, "1.7") {
			if iso, ok := isoMod.Isolates[isolate]; ok && iso != nil && !iso.IsExtract() {
				iso.Kind = "extract"
				iso.WithArgs = len(iso.Elems)
			}
		}
		isoMod.Cfg.IsolateCfg.CompileWithInvariants =
			cfg.Target == "test" && languageVersionAfter(isoMod, "1.7")
		if isolate != "" || languageVersionAtLeast(isoMod, "1.7") {
			if err := goivy.CreateIsolate(isolate, isoMod); err != nil {
				return nil, err
			}
		}
		prepareModuleForGo(isoMod, cfg)
		outCfg := cfg
		if len(isolates) > 1 && isolate != "" {
			suffix := "_" + varName(isolate)
			isoMod.Name = moduleBaseName(isoMod) + suffix
			if outCfg.PackageName != "" {
				outCfg.PackageName += suffix
			}
		}
		out, err := Generate(isoMod, outCfg)
		if err != nil {
			return nil, err
		}
		batch.Outputs = append(batch.Outputs, out)
	}
	return batch, nil
}

// mergeParams routes the params map into Config and a residual map of
// goivy-specific parameters (currently just "isolate"). Mirrors
// ivy2cpp/compile.go mergeParams with the C++-only keys removed
// (compiler, stdafx). No gomodule key — see config.go for why.
func mergeParams(params map[string]string, cfg Config) (Config, map[string]string, error) {
	ivyParams := map[string]string{}
	for k, v := range params {
		switch k {
		case "target":
			if cfg.Target == "" {
				cfg.Target = v
			}
		case "package", "classname":
			// "classname" accepted for argv-compatibility with
			// ivy2cpp invocations; treated as Go package name.
			if cfg.PackageName == "" {
				cfg.PackageName = v
			}
		case "main":
			if cfg.MainName == "" {
				cfg.MainName = v
			}
		case "outdir":
			if cfg.OutDir == "" {
				cfg.OutDir = v
			}
		case "test_iters":
			if cfg.TestIters == "" {
				cfg.TestIters = v
			}
		case "test_runs":
			if cfg.TestRuns == "" {
				cfg.TestRuns = v
			}
		case "trace":
			cfg.Trace = parseBool(v)
		case "build":
			cfg.Build = parseBool(v)
		case "isolate":
			ivyParams[k] = v
		default:
			return cfg, nil, fmt.Errorf("ivy2go: unknown parameter %q", k)
		}
	}
	var err error
	cfg, _, err = normalizeConfig(cfg)
	if err != nil {
		return cfg, nil, err
	}
	return cfg, ivyParams, nil
}

// normalizeConfig validates and defaults the Config. Mirrors
// ivy2cpp/compile.go normalizeConfig minus the compiler-selection
// branch.
func normalizeConfig(cfg Config) (Config, string, error) {
	requested := strings.TrimSpace(cfg.RequestedTarget)
	if requested == "" {
		requested = strings.TrimSpace(cfg.Target)
	}
	if requested == "" {
		requested = "gen"
	}
	switch requested {
	case "impl", "repl", "test", "gen", "class":
	default:
		return cfg, "", fmt.Errorf("ivy2go: target %q is not supported", requested)
	}
	cfg.RequestedTarget = requested
	cfg.Target = requested
	if requested == "class" {
		// "class" target produces a library package — same shape as
		// "repl" but without main().
		cfg.Target = "repl"
		cfg.EmitMain = false
	} else {
		cfg.EmitMain = true
	}
	if cfg.MainName == "" {
		cfg.MainName = "main"
	}
	if cfg.TestIters == "" {
		cfg.TestIters = "100"
	}
	if cfg.TestRuns == "" {
		cfg.TestRuns = "1"
	}
	if cfg.GoivyImportPath == "" {
		cfg.GoivyImportPath = DefaultGoivyImportPath
	}
	return cfg, requested, nil
}

// selectedIsolates mirrors ivy2cpp/compile.go selectedIsolates.
func selectedIsolates(mod *goivy.Module, requested string) []string {
	requested = strings.TrimSpace(requested)
	if requested != "" && requested != "all" {
		return []string{requested}
	}
	if requested == "all" {
		names := make([]string, 0, len(mod.Isolates))
		for name := range mod.Isolates {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}
	if languageVersionAtLeast(mod, "1.7") {
		if _, ok := mod.Isolates["this"]; ok {
			return []string{"this"}
		}
	}
	return []string{""}
}

// languageVersionAtLeast and languageVersionAfter port the helpers in
// ivy2cpp/compile.go of the same names.
func languageVersionAtLeast(mod *goivy.Module, want string) bool {
	if mod == nil || mod.Cfg == nil || mod.Cfg.IuCfg == nil {
		return true
	}
	return compareVersionStrings(mod.Cfg.IuCfg.GetStringVersion(), want) >= 0
}

func languageVersionAfter(mod *goivy.Module, want string) bool {
	if mod == nil || mod.Cfg == nil || mod.Cfg.IuCfg == nil {
		return true
	}
	return compareVersionStrings(mod.Cfg.IuCfg.GetStringVersion(), want) > 0
}

func compareVersionStrings(a, b string) int {
	ap := parseVersionParts(a)
	bp := parseVersionParts(b)
	for i := 0; i < len(ap) || i < len(bp); i++ {
		av, bv := 0, 0
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func parseVersionParts(v string) []int {
	fields := strings.Split(strings.TrimSpace(v), ".")
	out := make([]int, 0, len(fields))
	for _, f := range fields {
		n := 0
		for _, r := range f {
			if r < '0' || r > '9' {
				break
			}
			n = n*10 + int(r-'0')
		}
		out = append(out, n)
	}
	return out
}

// applySessionParameters mirrors ivy2cpp/compile.go applySessionParameters.
// Idempotent; safe to call from both CompileAndGenerateAll and Generate.
func applySessionParameters(mod *goivy.Module, cfg Config) {
	if mod == nil || mod.Cfg == nil {
		return
	}
	if mod.Cfg.ActCfg != nil {
		goivy.SetDeterminize(mod.Cfg.ActCfg, true)
	}
	if mod.Cfg.SolverOpts != nil {
		mod.Cfg.SolverOpts.UseZ3Enums = true
	}
	if mod.Cfg.IsolateCfg != nil {
		isoCfg := mod.Cfg.IsolateCfg
		isoCfg.InterpretAllSorts = true
		isoCfg.ConeOfInfluence = false
		isoCfg.CreateImports = true
		isoCfg.EnforceAxioms = true
		isoCfg.AssumeInvariants = false
		if cfg.Target == "test" {
			isoCfg.IsolateMode = "test"
		} else {
			isoCfg.IsolateMode = "compile"
		}
		if cfg.Target == "gen" {
			isoCfg.FilterSymbols = false
		} else {
			isoCfg.KeepDestructors = true
		}
	}
}

// prepareModuleForGo mirrors ivy2cpp/compile.go prepareModuleForCPP.
// Logic is identical because both packages need the same module
// massaging (the _generating symbol, sort ordering, props→axioms,
// invariants→action assumptions).
func prepareModuleForGo(mod *goivy.Module, cfg Config) {
	_ = cfg
	pruneStateStoresToSignature(mod)
	ensureGeneratingSymbol(mod)
	ensureSortOrderForGo(mod)
	if len(mod.LabeledProps) > 0 {
		mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
		mod.LabeledProps = nil
	}
	addConjsToActions(mod)
}

// pruneStateStoresToSignature mirrors ivy2cpp/compile.go:121. When
// the isolate config requested symbol filtering or cone-of-influence
// pruning, drop module relations / functions whose names are no
// longer in the signature.
func pruneStateStoresToSignature(mod *goivy.Module) {
	if mod == nil || mod.Sig == nil || mod.Cfg == nil || mod.Cfg.IsolateCfg == nil {
		return
	}
	if !mod.Cfg.IsolateCfg.FilterSymbols && !mod.Cfg.IsolateCfg.ConeOfInfluence {
		return
	}
	keep := map[string]bool{}
	for name := range mod.Sig.Symbols.All() {
		keep[name] = true
	}
	if mod.Relations != nil {
		for key := range mod.Relations.All() {
			name := goivy.SymbolNameFromKey(key)
			if !keep[name] {
				mod.Relations.Delkey(key)
			}
		}
	}
	if mod.Functions != nil {
		for key := range mod.Functions.All() {
			name := goivy.SymbolNameFromKey(key)
			if !keep[name] {
				mod.Functions.Delkey(key)
			}
		}
	}
}

// descriptorParamDesc mirrors ivy2cpp/compile.go:20.
type descriptorParamDesc struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// descriptorJSON mirrors ivy2cpp/compile.go:748. Builds a JSON
// descriptor of the emitted module — used by external tooling that
// wants to know the binary name, isolate, and parameter shape per
// generated process.
func descriptorJSON(mod *goivy.Module, cfg Config, outputs []*Output, isolates []string) (string, error) {
	type processDesc struct {
		Binary string                `json:"binary"`
		Name   string                `json:"name"`
		Params []descriptorParamDesc `json:"params"`
	}
	desc := map[string]any{}
	processes := make([]processDesc, 0, len(outputs))
	for i, out := range outputs {
		name := ""
		if i < len(isolates) {
			name = isolates[i]
		}
		processes = append(processes, processDesc{
			Binary: out.BaseName,
			Name:   name,
			Params: describeParams(mod.Params),
		})
	}
	desc["processes"] = processes
	if cfg.Target == "test" {
		desc["test_params"] = []string{"iters", "runs", "seed", "delay", "wait", "modelfile"}
	}
	data, err := goJSON.Marshal(desc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// describeParams mirrors ivy2cpp/compile.go:778.
func describeParams(params []*goivy.Const) []descriptorParamDesc {
	out := make([]descriptorParamDesc, 0, len(params))
	for _, p := range params {
		if p == nil {
			continue
		}
		typ := ""
		if p.CSort != nil {
			typ = p.CSort.String()
		}
		out = append(out, descriptorParamDesc{Name: p.Name, Type: typ})
	}
	return out
}

// addConjsToActions mirrors ivy2cpp/compile.go:714. Walks the
// module's labeled conjectures (invariants) and appends an
// AssertAction for each to the body of every public action. Also
// appends them to InitialActions and registers a synthetic
// `__check_invariants` Initializer so the same checks run at startup.
//
// Idempotent: returns immediately if a `__check_invariants`
// initializer already exists.
func addConjsToActions(mod *goivy.Module) {
	if mod == nil || mod.Actions == nil || mod.PublicActions == nil {
		return
	}
	for _, init := range mod.Initializers {
		if init.Name == "__check_invariants" {
			return
		}
	}
	var asserts []goivy.Expr
	for _, conj := range mod.LabeledConjs {
		if conj == nil {
			continue
		}
		fmla, ok := conj.Formula.(goivy.Expr)
		if !ok {
			continue
		}
		a := goivy.NewAssertAction(fmla)
		a.SetLineno(conj.GetLineno())
		asserts = append(asserts, a)
	}
	if len(asserts) == 0 {
		return
	}
	seq := goivy.NewSequence(asserts...)
	for name, action := range mod.Actions.All() {
		if mod.PublicActions.Get(name) {
			mod.Actions.Set(name, goivy.AppendToAction(action, seq))
		}
	}
	mod.Initializers = append(mod.Initializers, goivy.NamedAction{Name: "__check_invariants", Action: seq})
	initSeq := goivy.NewSequence(asserts...)
	initSeq.SetFormalParams([]*goivy.Const{})
	mod.InitialActions = append(mod.InitialActions, initSeq)
}

func ensureGeneratingSymbol(mod *goivy.Module) {
	if mod == nil || mod.Sig == nil {
		return
	}
	if _, ok := mod.Sig.Symbols.Get2("_generating"); !ok {
		_, _ = mod.Sig.AddSymbol("_generating", goivy.Boolean)
	}
	if mod.Relations != nil {
		if _, exists := mod.Relations.Get2(goivy.RelationKey("_generating", goivy.Boolean)); !exists {
			mod.Relations.Set(goivy.RelationKey("_generating", goivy.Boolean), goivy.Boolean)
		}
	}
}

// ensureSortOrderForGo mirrors ivy2cpp/compile.go ensureSortOrderForCPP.
func ensureSortOrderForGo(mod *goivy.Module) {
	if mod == nil || mod.Sig == nil {
		return
	}
	seen := map[string]bool{}
	for _, name := range mod.SortOrder {
		seen[name] = true
	}
	var add func(goivy.Sort)
	add = func(s goivy.Sort) {
		if s == nil {
			return
		}
		for _, d := range goivy.SortDomain(s) {
			add(d)
		}
		r := goivy.SortRange(s)
		if r != s {
			add(r)
			return
		}
		name := sortName(s)
		if name == "" || name == "bool" || seen[name] {
			return
		}
		if _, ok := mod.Sig.Sorts.Get2(name); !ok {
			_ = mod.Sig.AddSort(s)
		}
		seen[name] = true
		mod.SortOrder = append(mod.SortOrder, name)
	}
	var names []string
	for name := range mod.Sig.Sorts.All() {
		if name != "bool" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if s, ok := mod.Sig.Sorts.Get2(name); ok {
			add(s)
		}
	}
	if mod.Relations != nil {
		for _, s := range mod.Relations.All() {
			add(s)
		}
	}
	if mod.Functions != nil {
		for _, s := range mod.Functions.All() {
			add(s)
		}
	}
	for _, p := range mod.Params {
		if p != nil {
			add(p.CSort)
		}
	}
}

// sortName extracts a sort's name; mirrors the helper in
// ivy2cpp/compile.go (same name, file-local there).
func sortName(s goivy.Sort) string {
	if s == nil {
		return ""
	}
	if n, ok := s.(interface{ GetName() string }); ok {
		return n.GetName()
	}
	return fmt.Sprint(s)
}

// ParseArgs mirrors ivy2cpp/compile.go ParseArgs.
func ParseArgs(args []string) (map[string]string, string, error) {
	params := map[string]string{}
	rest := args
	for len(rest) > 0 && strings.Contains(rest[0], "=") {
		parts := strings.SplitN(rest[0], "=", 2)
		if parts[0] == "" {
			return nil, "", fmt.Errorf("bad parameter %q", rest[0])
		}
		params[parts[0]] = parts[1]
		rest = rest[1:]
	}
	if len(rest) != 1 || !strings.HasSuffix(rest[0], ".ivy") {
		return nil, "", fmt.Errorf("usage: ivy2go [key=value ...] file.ivy")
	}
	return params, rest[0], nil
}

// BuildRequested mirrors ivy2cpp/compile.go BuildRequested.
func BuildRequested(params map[string]string) bool {
	if params == nil {
		return false
	}
	return parseBool(params["build"])
}

// WriteOutput writes a single Output's files to outDir. Mirrors
// ivy2cpp/compile.go WriteOutput, adapted to the multi-file Go package
// layout (per ARCHITECTURE_TODO.md §3.3).
func WriteOutput(out *Output, outDir string) error {
	if out == nil {
		return fmt.Errorf("ivy2go: nil output")
	}
	pkgDir := outputDirectory(outDir, out.BaseName)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return err
	}
	names := make([]string, 0, len(out.Files))
	for name := range out.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(pkgDir, name)
		if err := os.WriteFile(path, []byte(out.Files[name]), 0o644); err != nil {
			return err
		}
	}
	// We intentionally do NOT emit a go.mod. The caller must place
	// outDir inside an existing Go module (or workspace) so the
	// emitted package's `import "github.com/glycerine/ivy/goivy"`
	// resolves naturally. Emitting our own go.mod would force a
	// `replace` (dev-only hack) or a published goivy version (which
	// in-tree dev work doesn't have).
	for name, text := range out.ExtraFiles {
		extraPath := filepath.Join(outputBaseDirectory(outDir), name)
		if err := os.WriteFile(extraPath, []byte(text), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// WriteBatchOutput writes every Output in a BatchOutput. Mirrors
// ivy2cpp/compile.go WriteBatchOutput.
func WriteBatchOutput(batch *BatchOutput, outDir string) error {
	if batch == nil {
		return fmt.Errorf("ivy2go: nil batch output")
	}
	for _, out := range batch.Outputs {
		if err := WriteOutput(out, outDir); err != nil {
			return err
		}
	}
	if len(batch.Outputs) == 0 && len(batch.ExtraFiles) > 0 {
		if err := os.MkdirAll(outputBaseDirectory(outDir), 0o755); err != nil {
			return err
		}
		for name, text := range batch.ExtraFiles {
			extraPath := filepath.Join(outputBaseDirectory(outDir), name)
			if err := os.WriteFile(extraPath, []byte(text), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// outputDirectory returns the per-package directory inside outDir.
// In ivy2go, each Output lands in its own subdirectory (BaseName) so
// multiple isolates don't collide. ivy2cpp emits flat .h/.cpp pairs
// into outDir directly because the basename disambiguates files; the
// Go equivalent needs a directory because Go's compilation unit is the
// package (directory), not the file.
func outputDirectory(outDir, baseName string) string {
	if outDir == "" {
		outDir = "."
	}
	return filepath.Join(outDir, baseName)
}

// outputBaseDirectory returns the parent directory for ExtraFiles
// (e.g., .dsc descriptors that live alongside packages). Mirrors
// ivy2cpp/compile.go outputBaseDirectory.
func outputBaseDirectory(outDir string) string {
	if outDir == "" {
		return "."
	}
	return outDir
}

// parseBool mirrors ivy2cpp/compile.go parseBool.
func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
