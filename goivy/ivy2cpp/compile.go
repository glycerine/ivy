package ivy2cpp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type BatchOutput struct {
	Outputs    []*Output
	ExtraFiles map[string]string
	Config     Config
}

type descriptorParamDesc struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func CompileAndGenerate(filename string, params map[string]string, cfg Config) (*Output, error) {
	batch, err := CompileAndGenerateAll(filename, params, cfg)
	if err != nil {
		return nil, err
	}
	if len(batch.Outputs) != 1 {
		return nil, fmt.Errorf("ivy2cpp: expected one generated output, got %d", len(batch.Outputs))
	}
	return batch.Outputs[0], nil
}

func CompileAndGenerateAll(filename string, params map[string]string, cfg Config) (*BatchOutput, error) {
	rawClassName, explicitClassName := compileClassNameParam(params, cfg)
	cfg, ivyParams, err := mergeParams(params, cfg)
	if err != nil {
		return nil, err
	}
	if explicitClassName {
		cfg.ClassName = varName(rawClassName)
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
		// Python ivy_to_cpp.py:4612-4618 rewrites a non-extract isolate
		// named on the command line into an ExtractDef for repl in 1.7+.
		if cfg.Target == "repl" && languageVersionAtLeast(isoMod, "1.7") {
			if iso, ok := isoMod.Isolates[isolate]; ok && iso != nil && !iso.IsExtract() {
				iso.Kind = "extract"
				iso.WithArgs = len(iso.Elems)
			}
		}
		// Python ivy_to_cpp.py:4620-4622 — compile_with_invariants is
		// true only for the test target in language versions after 1.7.
		isoMod.Cfg.IsolateCfg.CompileWithInvariants =
			cfg.Target == "test" && languageVersionAfter(isoMod, "1.7")
		cppIface := snapshotCPPInterface(isoMod)
		if isolate != "" || languageVersionAtLeast(isoMod, "1.7") {
			if err := goivy.CreateIsolate(isolate, isoMod); err != nil {
				return nil, err
			}
			pruneStateStoresToSignature(isoMod)
			restoreCPPInterface(isoMod, cppIface)
		}
		prepareModuleForCPP(isoMod, cfg)
		outCfg := cfg
		baseName := moduleBaseName(isoMod)
		if explicitClassName {
			baseName = rawClassName
			outCfg.ClassName = varName(baseName)
		}
		if len(isolates) > 1 && isolate != "" {
			baseName = baseName + "_" + isolate
			outCfg.ClassName = varName(baseName)
		}
		isoMod.Name = baseName
		out, err := Generate(isoMod, outCfg)
		if err != nil {
			return nil, err
		}
		out.BaseName = baseName
		batch.Outputs = append(batch.Outputs, out)
	}
	if cfg.RequestedTarget == "repl" || cfg.RequestedTarget == "test" {
		if dsc, err := descriptorJSON(mod, cfg, batch.Outputs, isolates); err != nil {
			return nil, err
		} else if dsc != "" {
			name := moduleBaseName(mod)
			if explicitClassName {
				name = rawClassName
			}
			batch.ExtraFiles[name+".dsc"] = dsc
		}
	}
	for _, out := range batch.Outputs {
		if len(batch.ExtraFiles) > 0 {
			if out.ExtraFiles == nil {
				out.ExtraFiles = map[string]string{}
			}
			for k, v := range batch.ExtraFiles {
				out.ExtraFiles[k] = v
			}
		}
	}
	return batch, nil
}

func compileClassNameParam(params map[string]string, cfg Config) (string, bool) {
	if params != nil {
		if raw := strings.TrimSpace(params["classname"]); raw != "" {
			return raw, true
		}
	}
	if raw := strings.TrimSpace(cfg.ClassName); raw != "" {
		return raw, true
	}
	return "", false
}

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

func mergeParams(params map[string]string, cfg Config) (Config, map[string]string, error) {
	ivyParams := map[string]string{}
	for k, v := range params {
		switch k {
		case "target":
			if cfg.Target == "" {
				cfg.Target = v
			}
		case "classname":
			if cfg.ClassName == "" {
				cfg.ClassName = v
			}
		case "main":
			if cfg.MainName == "" {
				cfg.MainName = v
			}
		case "outdir":
			if cfg.OutDir == "" {
				cfg.OutDir = v
			}
		case "compiler":
			if cfg.Compiler == "" {
				cfg.Compiler = v
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
		case "stdafx":
			cfg.Stdafx = parseBool(v)
		case "build":
			cfg.Build = parseBool(v)
		case "isolate":
			ivyParams[k] = v
		default:
			return cfg, nil, fmt.Errorf("ivy2cpp: unknown parameter %q", k)
		}
	}
	var err error
	cfg, _, err = normalizeConfig(cfg)
	if err != nil {
		return cfg, nil, err
	}
	return cfg, ivyParams, nil
}

func BuildRequested(params map[string]string) bool {
	if params == nil {
		return false
	}
	return parseBool(params["build"])
}

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
		return nil, "", fmt.Errorf("usage: ivy2cpp [key=value ...] file.ivy")
	}
	return params, rest[0], nil
}

func WriteOutput(out *Output, outDir string) error {
	if out == nil {
		return fmt.Errorf("ivy2cpp: nil output")
	}
	dir := outputDirectory(outDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, out.BaseName+".h"), []byte(out.Header), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, out.BaseName+".cpp"), []byte(out.Impl), 0o644); err != nil {
		return err
	}
	for name, text := range out.ExtraFiles {
		if err := os.WriteFile(filepath.Join(outputBaseDirectory(outDir), name), []byte(text), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func WriteBatchOutput(batch *BatchOutput, outDir string) error {
	if batch == nil {
		return fmt.Errorf("ivy2cpp: nil batch output")
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
			if err := os.WriteFile(filepath.Join(outputBaseDirectory(outDir), name), []byte(text), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

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
		return cfg, "", fmt.Errorf("ivy2cpp: target %q is not supported", requested)
	}
	compiler := strings.TrimSpace(cfg.Compiler)
	if compiler == "" {
		compiler = "default"
	}
	switch compiler {
	case "default", "g++", "cl":
	default:
		return cfg, "", fmt.Errorf("ivy2cpp: compiler %q is not supported", compiler)
	}
	cfg.RequestedTarget = requested
	cfg.Target = requested
	if requested == "class" {
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
	cfg.Compiler = compiler
	return cfg, requested, nil
}

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

type cppInterfaceSnapshot struct {
	Actions         map[string]cppActionSignature
	CallbackActions map[string]goivy.Action
	ActionOrder     []string
	SortOrder       []string
	Sorts           map[string]goivy.Sort
	Interps         map[string]interface{}
	SortDestructors map[string][]*goivy.Const
	DestructorSorts map[string]goivy.Sort
}

type cppActionSignature struct {
	Params  []*goivy.Const
	Returns []*goivy.Const
}

func snapshotCPPInterface(mod *goivy.Module) cppInterfaceSnapshot {
	snap := cppInterfaceSnapshot{
		Actions:         map[string]cppActionSignature{},
		CallbackActions: map[string]goivy.Action{},
		Sorts:           map[string]goivy.Sort{},
		Interps:         map[string]interface{}{},
		SortDestructors: map[string][]*goivy.Const{},
		DestructorSorts: map[string]goivy.Sort{},
	}
	if mod == nil {
		return snap
	}
	snap.SortOrder = append([]string(nil), mod.SortOrder...)
	if mod.Actions != nil {
		for name, act := range mod.Actions.All() {
			snap.ActionOrder = append(snap.ActionOrder, name)
			snap.Actions[name] = cppActionSignature{
				Params:  append([]*goivy.Const(nil), act.GetFormalParams()...),
				Returns: append([]*goivy.Const(nil), act.GetFormalReturns()...),
			}
		}
		for _, name := range collectCallbackActionNames(mod) {
			if act, ok := mod.Actions.Get2(name); ok {
				snap.CallbackActions[name] = act
			}
		}
	}
	if mod.Sig != nil {
		for name, sort := range mod.Sig.Sorts.All() {
			snap.Sorts[name] = sort
		}
		for name, sort := range mod.Sig.Interp {
			snap.Interps[name] = sort
		}
	}
	if mod.SortDestructors != nil {
		for name, destrs := range mod.SortDestructors.All() {
			snap.SortDestructors[name] = append([]*goivy.Const(nil), destrs...)
		}
	}
	for name, sort := range mod.DestructorSorts {
		snap.DestructorSorts[name] = sort
	}
	return snap
}

func restoreCPPInterface(mod *goivy.Module, snap cppInterfaceSnapshot) {
	if mod == nil {
		return
	}
	if mod.Actions != nil {
		if len(snap.CallbackActions) > 0 {
			oldActions := mod.Actions
			nextActions := goivy.NewInsMap[string, goivy.Action]()
			added := map[string]bool{}
			for _, name := range snap.ActionOrder {
				if act, ok := oldActions.Get2(name); ok {
					nextActions.Set(name, act)
					added[name] = true
					continue
				}
				if act, ok := snap.CallbackActions[name]; ok {
					nextActions.Set(name, act)
					added[name] = true
				}
			}
			for name, act := range oldActions.All() {
				if added[name] {
					continue
				}
				nextActions.Set(name, act)
			}
			mod.Actions = nextActions
		}
	}
	for name, sig := range snap.Actions {
		restoreActionSignature(mod, name, sig)
		restoreActionSignature(mod, "ext:"+name, sig)
	}
	for _, sig := range snap.Actions {
		for _, p := range sig.Params {
			restoreSortForCPPInterface(mod, snap, p.CSort, map[string]bool{})
		}
		for _, r := range sig.Returns {
			restoreSortForCPPInterface(mod, snap, r.CSort, map[string]bool{})
		}
	}
}

func restoreActionSignature(mod *goivy.Module, name string, sig cppActionSignature) {
	if mod == nil || mod.Actions == nil || len(sig.Params)+len(sig.Returns) == 0 {
		return
	}
	act, ok := mod.Actions.Get2(name)
	if !ok || act == nil {
		return
	}
	if len(act.GetFormalParams()) == 0 && len(sig.Params) > 0 {
		act.SetFormalParams(append([]*goivy.Const(nil), sig.Params...))
	}
	if len(act.GetFormalReturns()) == 0 && len(sig.Returns) > 0 {
		act.SetFormalReturns(append([]*goivy.Const(nil), sig.Returns...))
	}
}

func restoreSortForCPPInterface(mod *goivy.Module, snap cppInterfaceSnapshot, sort goivy.Sort, seen map[string]bool) {
	if mod == nil || mod.Sig == nil || sort == nil {
		return
	}
	if fs, ok := sort.(*goivy.LogicFunctionSort); ok {
		for _, d := range fs.Domain() {
			restoreSortForCPPInterface(mod, snap, d, seen)
		}
		restoreSortForCPPInterface(mod, snap, fs.Range(), seen)
		return
	}
	name := sortName(sort)
	if name == "" || name == "bool" || seen[name] {
		return
	}
	seen[name] = true
	if s, ok := snap.Sorts[name]; ok {
		if _, exists := mod.Sig.Sorts.Get2(name); !exists {
			mod.Sig.Sorts.Set(name, s)
		}
		ensureSortOrderName(mod, snap, name)
	}
	if interp, ok := snap.Interps[name]; ok {
		if mod.Sig.Interp == nil {
			mod.Sig.Interp = map[string]interface{}{}
		}
		if _, exists := mod.Sig.Interp[name]; !exists {
			mod.Sig.Interp[name] = interp
		}
	}
	if destrs, ok := snap.SortDestructors[name]; ok {
		if mod.SortDestructors == nil {
			mod.SortDestructors = goivy.NewInsMap[string, []*goivy.Const]()
		}
		if _, exists := mod.SortDestructors.Get2(name); !exists {
			mod.SortDestructors.Set(name, append([]*goivy.Const(nil), destrs...))
		}
		if mod.DestructorSorts == nil {
			mod.DestructorSorts = map[string]goivy.Sort{}
		}
		if ds, ok := snap.DestructorSorts[name]; ok {
			mod.DestructorSorts[name] = ds
		}
		for _, d := range destrs {
			if fs, ok := d.CSort.(*goivy.LogicFunctionSort); ok {
				restoreSortForCPPInterface(mod, snap, fs.Range(), seen)
			}
		}
	}
}

func ensureSortOrderName(mod *goivy.Module, snap cppInterfaceSnapshot, name string) {
	for _, existing := range mod.SortOrder {
		if existing == name {
			return
		}
	}
	orderIdx := len(snap.SortOrder)
	for i, n := range snap.SortOrder {
		if n == name {
			orderIdx = i
			break
		}
	}
	insertAt := len(mod.SortOrder)
	for i, existing := range mod.SortOrder {
		existingIdx := len(snap.SortOrder)
		for j, n := range snap.SortOrder {
			if n == existing {
				existingIdx = j
				break
			}
		}
		if existingIdx > orderIdx {
			insertAt = i
			break
		}
	}
	mod.SortOrder = append(mod.SortOrder, "")
	copy(mod.SortOrder[insertAt+1:], mod.SortOrder[insertAt:])
	mod.SortOrder[insertAt] = name
}

func prepareModuleForCPP(mod *goivy.Module, cfg Config) {
	ensureGeneratingSymbol(mod)
	ensureSortOrderForCPP(mod)
	if len(mod.LabeledProps) > 0 {
		mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
		mod.LabeledProps = nil
	}
	if len(mod.LabeledConjs) > 0 {
		addConjsToActions(mod)
	}
}

func ensureGeneratingSymbol(mod *goivy.Module) {
	if mod == nil || mod.Sig == nil {
		return
	}
	// Python ivy_to_cpp.py adds _generating to the signature for every
	// target before code generation, and the runtime constructor initializes
	// it with ___ivy_choose even for impl/repl outputs.
	if _, ok := mod.Sig.Symbols.Get2("_generating"); !ok {
		_, _ = mod.Sig.AddSymbol("_generating", goivy.Boolean)
	}
	if mod.Relations != nil {
		if _, exists := mod.Relations.Get2(goivy.RelationKey("_generating", goivy.Boolean)); !exists {
			mod.Relations.Set(goivy.RelationKey("_generating", goivy.Boolean), goivy.Boolean)
		}
	}
}

func ensureSortOrderForCPP(mod *goivy.Module) {
	if mod == nil || mod.Sig == nil {
		return
	}
	seen := map[string]bool{}
	for _, name := range mod.SortOrder {
		seen[name] = true
	}
	addSort := func(s goivy.Sort) {}
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
	addSort = add
	var names []string
	for name := range mod.Sig.Sorts.All() {
		if name != "bool" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		if s, ok := mod.Sig.Sorts.Get2(name); ok {
			addSort(s)
		}
	}
	for _, s := range mod.Relations.All() {
		addSort(s)
	}
	for _, s := range mod.Functions.All() {
		addSort(s)
	}
	for _, p := range mod.Params {
		if p != nil {
			addSort(p.CSort)
		}
	}
}

// applySessionParameters mirrors the per-session flag setup that
// Python ivy_to_cpp.py:main_int runs at lines 4514-4530 (and the
// implicit slv.set_use_native_enums(True) default). Idempotent so it
// is safe to call from both CompileAndGenerateAll and Generate.
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
	if mod.CompCfg == nil {
		mod.CompCfg = goivy.NewCompilerConfig(mod.Cfg)
	}
	goivy.SetVerifyingOnMod(mod, false)
}

func addConjsToActions(mod *goivy.Module) {
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
	data, err := json.Marshal(desc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

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

func outputBaseDirectory(outDir string) string {
	if outDir == "" {
		return "."
	}
	return outDir
}

func outputDirectory(outDir string) string {
	base := outputBaseDirectory(outDir)
	buildDir := "build"
	if _, err := os.Stat(buildDir); err == nil {
		return filepath.Join(base, buildDir)
	}
	return base
}

func moduleLibSpecs(mod *goivy.Module) []string {
	if mod == nil {
		return nil
	}
	var specs []string
	for key, value := range mod.Attributes.All() {
		parts := strings.Split(key, ".")
		if len(parts) == 0 || parts[len(parts)-1] != "libspec" {
			continue
		}
		var text string
		if node, ok := value.(goivy.Node); ok {
			text = goivy.NodeRep(node)
		} else {
			text = fmt.Sprint(value)
		}
		text = strings.Trim(text, `"`)
		for _, item := range strings.Split(text, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				specs = append(specs, item)
			}
		}
	}
	sort.Strings(specs)
	return specs
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
