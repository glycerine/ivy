package main

// tinynode runs the tinygo js/wasm build of goivy under Node.js.
//
// It is a command-line producer of the wasm Go Ivy stdout/stderr stream, so
// tests can compare it against python ivy in parser_golden_test.go.

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type stringList []string

func (l *stringList) String() string {
	return strings.Join(*l, ",")
}

func (l *stringList) Set(v string) error {
	*l = append(*l, v)
	return nil
}

type config struct {
	Root        string            `json:"root"`
	IncludeDir  string            `json:"includeDir"`
	GoivyWasm   string            `json:"goivyWasm"`
	WasmExec    string            `json:"wasmExec"`
	Z3JS        string            `json:"z3JS"`
	Z3Wasm      string            `json:"z3Wasm"`
	SMTImports  string            `json:"smtImports"`
	NodeFS      string            `json:"nodeFS"`
	TinyGoWASI  string            `json:"tinyGoWASI"`
	SpecPath    string            `json:"specPath"`
	Params      map[string]string `json:"params"`
	Isolates    []string          `json:"isolates,omitempty"`
	MemoryLimit string            `json:"memoryLimit"`
	GOGC        string            `json:"gogc"`
}

func main() {
	rootDefault := defaultGoivyRoot()
	tinynodeCmdDir := filepath.Join(rootDefault, "cmd", "tinynode")
	staticDefault := filepath.Join(rootDefault, "webvue", "static")
	workerDefault := filepath.Join(rootDefault, "webvue", "src", "workers")

	node := flag.String("node", "node", "Node.js executable")
	script := flag.String("script", filepath.Join(tinynodeCmdDir, "tinynode.mjs"), "Node.js harness script")
	root := flag.String("root", rootDefault, "goivy source root")
	includeDir := flag.String("include-dir", defaultIncludeDir(rootDefault), "Ivy standard-library include directory")
	goivyWasm := flag.String("goivy-wasm", filepath.Join(staticDefault, "goivy-check-tinygo-js.wasm"), "Go Ivy js/wasm file")
	wasmExec := flag.String("wasm-exec", filepath.Join(staticDefault, "wasm_exec_tinygo_0.40.1.js"), "wasm_exec.js matching the tinygo compiler/version")
	z3JS := flag.String("z3-js", filepath.Join(staticDefault, "z3-471-api.js"), "Z3 JavaScript glue")
	z3Wasm := flag.String("z3-wasm", filepath.Join(staticDefault, "z3-471-api.wasm"), "Z3 wasm file")
	smtImports := flag.String("smt-imports", filepath.Join(workerDefault, "smtZ3Imports.js"), "smt_z3 import bridge module")
	nodeFS := flag.String("node-fs", filepath.Join(workerDefault, "goivyNodeFS.js"), "Go wasm fs/process/path host module")
	tinyGoWASI := flag.String("tinygo-wasi", filepath.Join(workerDefault, "goivyTinyGoWasiP1.js"), "TinyGo wasi_snapshot_preview1 host module")
	memoryLimit := flag.String("memory-limit", "3GiB", "GOIVY_WASM_MEMORY_LIMIT for the Go wasm runtime")
	gogc := flag.String("gogc", "50", "GOIVY_WASM_GOGC for the Go wasm runtime")
	keepTemp := flag.Bool("keep-temp", false, "keep the generated tinynode config file")
	var nodeFlags stringList
	var explicitParams stringList
	flag.Var(&nodeFlags, "node-flag", "extra flag passed to Node.js; may be repeated")
	flag.Var(&explicitParams, "param", "Ivy checker parameter key=value; may be repeated")
	flag.Parse()

	params, isolates, specPath, err := parseArgs(flag.Args(), explicitParams)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cfg := config{
		Root:        cleanAbs(*root),
		IncludeDir:  cleanAbs(*includeDir),
		GoivyWasm:   cleanAbs(*goivyWasm),
		WasmExec:    cleanAbs(*wasmExec),
		Z3JS:        cleanAbs(*z3JS),
		Z3Wasm:      cleanAbs(*z3Wasm),
		SMTImports:  cleanAbs(*smtImports),
		NodeFS:      cleanAbs(*nodeFS),
		TinyGoWASI:  cleanAbs(*tinyGoWASI),
		SpecPath:    cleanAbs(specPath),
		Params:      params,
		Isolates:    isolates,
		MemoryLimit: *memoryLimit,
		GOGC:        *gogc,
	}
	if err := validateConfig(cfg, *script); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	tmpDir, err := os.MkdirTemp("", "tinynode-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tinynode: create temp dir: %v\n", err)
		os.Exit(1)
	}
	if !*keepTemp {
		defer os.RemoveAll(tmpDir)
	}

	configPath := filepath.Join(tmpDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tinynode: encode config: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "tinynode: write config: %v\n", err)
		os.Exit(1)
	}
	if *keepTemp {
		fmt.Fprintf(os.Stderr, "tinynode: kept config %s\n", configPath)
	}
	fmt.Printf("tinynode.go:120 wrote config.json to temp dir: '%v'; our script: '%v'\n", tmpDir, *script)

	args := append([]string{}, nodeFlags...)
	if len(nodeFlags) == 0 && nodeSupportsFlag(*node, "--experimental-wasm-exnref") {
		args = append(args, "--experimental-wasm-exnref")
	}
	args = append(args, cleanAbs(*script), configPath)

	cmd := exec.Command(*node, args...)
	cmd.Dir = cfg.Root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "tinynode: run node: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs(args []string, explicitParams []string) (map[string]string, []string, string, error) {
	params := make(map[string]string)
	var isolates []string
	addParam := func(raw string) error {
		k, v, ok := strings.Cut(raw, "=")
		if !ok || k == "" {
			return fmt.Errorf("bad parameter %q; want key=value", raw)
		}
		params[k] = v
		if k == "isolate" {
			isolates = append(isolates, v)
		}
		return nil
	}
	for _, raw := range explicitParams {
		if err := addParam(raw); err != nil {
			return nil, nil, "", fmt.Errorf("bad -param %q; want key=value", raw)
		}
	}

	var specPath string
	for _, arg := range args {
		if strings.Contains(arg, "=") && specPath == "" {
			if err := addParam(arg); err != nil {
				return nil, nil, "", err
			}
			continue
		}
		if specPath != "" {
			return nil, nil, "", fmt.Errorf("unexpected extra argument %q; want key=value params followed by one .ivy file", arg)
		}
		specPath = arg
	}
	if specPath == "" {
		return nil, nil, "", errors.New("usage: tinynode [flags] [key=value ...] file.ivy")
	}
	return params, isolates, specPath, nil
}

func validateConfig(cfg config, script string) error {
	for label, path := range map[string]string{
		"-script":      script,
		"-root":        cfg.Root,
		"-include-dir": cfg.IncludeDir,
		"-goivy-wasm":  cfg.GoivyWasm,
		"-wasm-exec":   cfg.WasmExec,
		"-z3-js":       cfg.Z3JS,
		"-z3-wasm":     cfg.Z3Wasm,
		"-smt-imports": cfg.SMTImports,
		"-node-fs":     cfg.NodeFS,
		"-tinygo-wasi": cfg.TinyGoWASI,
		"file.ivy":     cfg.SpecPath,
	} {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("tinynode: %s %s: %w", label, path, err)
		}
	}
	return nil
}

func nodeSupportsFlag(node, nodeFlag string) bool {
	cmd := exec.Command(node, nodeFlag, "-e", "")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func cleanAbs(path string) string {
	if path == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

func defaultIncludeDir(root string) string {
	parent := filepath.Dir(filepath.Clean(root))
	candidates := []string{
		filepath.Join(parent, "ivy-lang-examples", "ivy", "include"),
		filepath.Join(parent, "pyivy", "ivy", "ivy", "include"),
		filepath.Join(root, "ivy-lang-examples", "ivy", "include"),
	}
	if home := os.Getenv("HOME"); home != "" {
		candidates = append(candidates,
			filepath.Join(home, "ivy", "ivy-lang-examples", "ivy", "include"),
			filepath.Join(home, "ivy", "pyivy", "ivy", "ivy", "include"),
		)
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return filepath.Join(parent, "ivy-lang-examples", "ivy", "include")
}

func defaultGoivyRoot() string {
	if wd, err := os.Getwd(); err == nil {
		for dir := filepath.Clean(wd); ; dir = filepath.Dir(dir) {
			if looksLikeGoivyRoot(dir) {
				return dir
			}
			if parent := filepath.Dir(dir); parent == dir {
				break
			}
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		dir := filepath.Dir(file)
		if filepath.Base(dir) == "tinynode" && looksLikeGoivyRoot(filepath.Dir(dir)) {
			return filepath.Dir(dir)
		}
	}
	return "."
}

func looksLikeGoivyRoot(dir string) bool {
	_, err1 := os.Stat(filepath.Join(dir, "go.mod"))
	_, err2 := os.Stat(filepath.Join(dir, "webvue", "static", "z3-471-api.js"))
	return err1 == nil && err2 == nil
}
