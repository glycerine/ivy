// Command whatwasm inspects a WebAssembly binary and reports:
//   - C++ exception handling model (native-exnref, js-mediated, or none)
//   - Runtime environment (WASIp1, js/wasm, bare, etc.)
//   - WebAssembly features present (from target_features section + wazero probing)
//   - Section inventory and import/export summary
//
// Usage: whatwasm <path-to-wasm-binary>
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"

	cristalbase64 "github.com/cristalhq/base64"
	"github.com/glycerine/blake3"
)

// WASM binary section IDs per spec.
const (
	sectionCustom    = 0
	sectionType      = 1
	sectionImport    = 2
	sectionFunction  = 3
	sectionTable     = 4
	sectionMemory    = 5
	sectionGlobal    = 6
	sectionExport    = 7
	sectionStart     = 8
	sectionElement   = 9
	sectionCode      = 10
	sectionData      = 11
	sectionDataCount = 12 // bulk memory indicator
	sectionTag       = 13 // exception handling indicator
)

// Import/export kind bytes per spec.
const (
	kindFunction = 0
	kindTable    = 1
	kindMemory   = 2
	kindGlobal   = 3
)

type sectionHeader struct {
	id   byte
	size uint32
	off  int // byte offset of section content in original data
}

type wasmImport struct {
	module string
	name   string
	kind   byte
}

type wasmExport struct {
	name string
	kind byte
}

type wasmBinary struct {
	path    string
	size    int64
	version uint32
	b3      string

	sections       []sectionHeader
	customSections map[string][]byte // section name → raw bytes after name field

	imports []wasmImport
	exports []wasmExport

	targetFeatures      map[string]bool // from target_features custom section; true='+'
	hasTagSection       bool            // section 13 → exception handling in use
	hasDataCountSection bool            // section 12 → bulk memory in use
}

type probeResult struct {
	required bool // binary fails to compile without this feature
	probed   bool // false when feature has no wazero bit
}

// featureEntry describes one wasm-opt feature flag and its wazero equivalent.
// binaryName is the name used in the WASM binary's target_features custom section
// (Binaryen's internal canonical name); empty means it matches wasmOptName.
type featureEntry struct {
	wasmOptName string
	binaryName  string // Binaryen canonical name in target_features section
	wazeroBit   api.CoreFeatures
	hasBit      bool
}

func (fe featureEntry) lookupName() string {
	if fe.binaryName != "" {
		return fe.binaryName
	}
	return fe.wasmOptName
}

// knownFeatures is the ordered list of wasm-opt feature flags to report.
// Features without a wazero bit are still reported from target_features / section analysis.
var knownFeatures = []featureEntry{
	{"exception-handling", "", experimental.CoreFeaturesExceptionHandling, true},
	{"reference-types", "", api.CoreFeatureReferenceTypes, true},
	{"bulk-memory", "", api.CoreFeatureBulkMemoryOperations, true},
	{"bulk-memory-opt", "", 0, false},
	{"multivalue", "", api.CoreFeatureMultiValue, true},
	{"mutable-globals", "", api.CoreFeatureMutableGlobal, true},
	{"sign-ext", "", api.CoreFeatureSignExtensionOps, true},
	{"nontrapping-float-to-int", "nontrapping-fptoint", api.CoreFeatureNonTrappingFloatToIntConversion, true},
	{"simd128", "", api.CoreFeatureSIMD, true},
	{"tail-call", "", experimental.CoreFeaturesTailCall, true},
	{"threads", "atomics", experimental.CoreFeaturesThreads, true},
	{"extended-const", "", experimental.CoreFeaturesExtendedConst, true},
	{"gc", "", 0, false},
	{"memory64", "", 0, false},
	{"relaxed-simd", "", 0, false},
	{"strings", "", 0, false},
	{"multimemory", "", 0, false},
	{"stack-switching", "", 0, false},
	{"shared-everything", "", 0, false},
	{"fp16", "", 0, false},
	{"custom-descriptors", "", 0, false},
	{"relaxed-atomics", "", 0, false},
	{"custom-page-sizes", "", 0, false},
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: whatwasm <path-to-wasm-binary>\n")
		os.Exit(1)
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	wb, err := parseBinary(path, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	fmt.Fprintf(os.Stderr, "probing features with wazero (may take a moment for large binaries)...\n")
	probes, wazeroErr := probeFeatures(ctx, data)
	printReport(wb, probes, wazeroErr)
}

// ── Binary parsing ────────────────────────────────────────────────────────────

func readLEB128(data []byte, pos int) (uint32, int, error) {
	result := uint32(0)
	shift := uint(0)
	for {
		if pos >= len(data) {
			return 0, pos, fmt.Errorf("unexpected EOF at pos %d reading LEB128", pos)
		}
		b := data[pos]
		pos++
		result |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			return result, pos, nil
		}
		shift += 7
		if shift >= 35 {
			return 0, pos, fmt.Errorf("LEB128 overflow")
		}
	}
}

func readString(data []byte, pos int) (string, int, error) {
	n, newPos, err := readLEB128(data, pos)
	if err != nil {
		return "", pos, err
	}
	end := newPos + int(n)
	if end > len(data) {
		return "", newPos, fmt.Errorf("string length %d overruns data at pos %d", n, newPos)
	}
	return string(data[newPos:end]), end, nil
}

func skipLimits(data []byte, pos int) (int, error) {
	if pos >= len(data) {
		return pos, fmt.Errorf("EOF reading limits flags")
	}
	flags := data[pos]
	pos++
	_, pos, err := readLEB128(data, pos) // min
	if err != nil {
		return pos, err
	}
	if flags&1 != 0 {
		_, pos, err = readLEB128(data, pos) // max
		if err != nil {
			return pos, err
		}
	}
	return pos, nil
}

func skipImportDesc(data []byte, pos int, kind byte) (int, error) {
	switch kind {
	case kindFunction:
		_, newPos, err := readLEB128(data, pos)
		return newPos, err
	case kindTable:
		if pos >= len(data) {
			return pos, fmt.Errorf("EOF reading table reftype")
		}
		return skipLimits(data, pos+1)
	case kindMemory:
		return skipLimits(data, pos)
	case kindGlobal:
		if pos+1 >= len(data) {
			return pos, fmt.Errorf("EOF reading global type")
		}
		return pos + 2, nil // valtype + mutable flag
	default:
		return pos, fmt.Errorf("unknown import kind %d", kind)
	}
}

func parseImports(wb *wasmBinary, data []byte) error {
	count, pos, err := readLEB128(data, 0)
	if err != nil {
		return err
	}
	for i := uint32(0); i < count; i++ {
		mod, newPos, err := readString(data, pos)
		if err != nil {
			return fmt.Errorf("import %d module: %w", i, err)
		}
		pos = newPos
		name, newPos, err := readString(data, pos)
		if err != nil {
			return fmt.Errorf("import %d name: %w", i, err)
		}
		pos = newPos
		if pos >= len(data) {
			return fmt.Errorf("EOF reading import %d kind", i)
		}
		kind := data[pos]
		pos++
		pos, err = skipImportDesc(data, pos, kind)
		if err != nil {
			return fmt.Errorf("import %d desc: %w", i, err)
		}
		wb.imports = append(wb.imports, wasmImport{module: mod, name: name, kind: kind})
	}
	return nil
}

func parseExports(wb *wasmBinary, data []byte) error {
	count, pos, err := readLEB128(data, 0)
	if err != nil {
		return err
	}
	for i := uint32(0); i < count; i++ {
		name, newPos, err := readString(data, pos)
		if err != nil {
			return fmt.Errorf("export %d name: %w", i, err)
		}
		pos = newPos
		if pos >= len(data) {
			return fmt.Errorf("EOF reading export %d kind", i)
		}
		kind := data[pos]
		pos++
		_, pos, err = readLEB128(data, pos) // index
		if err != nil {
			return fmt.Errorf("export %d index: %w", i, err)
		}
		wb.exports = append(wb.exports, wasmExport{name: name, kind: kind})
	}
	return nil
}

func parseTargetFeatures(wb *wasmBinary, data []byte) {
	count, pos, err := readLEB128(data, 0)
	if err != nil {
		return
	}
	for i := uint32(0); i < count; i++ {
		if pos >= len(data) {
			return
		}
		prefix := data[pos]
		pos++
		name, newPos, err := readString(data, pos)
		if err != nil {
			return
		}
		pos = newPos
		wb.targetFeatures[name] = prefix == '+'
	}
}

func parseBinary(path string, data []byte) (*wasmBinary, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("file too small (%d bytes) to be a WASM binary", len(data))
	}
	if string(data[:4]) != "\x00asm" {
		return nil, fmt.Errorf("not a WASM binary: bad magic bytes %02x %02x %02x %02x", data[0], data[1], data[2], data[3])
	}
	version := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24

	abs, err := filepath.Abs(path)
	panicOn(err)
	wb := &wasmBinary{
		path:           abs,
		version:        version,
		customSections: make(map[string][]byte),
		targetFeatures: make(map[string]bool),
	}
	fi, err := os.Stat(path)
	if err == nil {
		wb.size = fi.Size()
	} else {
		panic(fmt.Sprintf("could not stat path '%v': '%v'\n", path, err))
	}

	b3, err := Blake3OfFile(path)
	if err != nil {
		panic(fmt.Sprintf("Blake3OfFile on path '%v' error: %v\n", path, err))
	}
	wb.b3 = b3

	pos := 8
	for pos < len(data) {
		if pos >= len(data) {
			break
		}
		sectionID := data[pos]
		pos++
		sectionSize, newPos, err := readLEB128(data, pos)
		if err != nil {
			break
		}
		pos = newPos
		contentStart := pos
		contentEnd := pos + int(sectionSize)
		if contentEnd > len(data) {
			break
		}
		wb.sections = append(wb.sections, sectionHeader{
			id: sectionID, size: sectionSize, off: contentStart,
		})
		switch sectionID {
		case sectionTag:
			wb.hasTagSection = true
		case sectionDataCount:
			wb.hasDataCountSection = true
		case sectionImport:
			_ = parseImports(wb, data[contentStart:contentEnd])
		case sectionExport:
			_ = parseExports(wb, data[contentStart:contentEnd])
		case sectionCustom:
			name, after, err := readString(data, contentStart)
			if err == nil {
				wb.customSections[name] = data[after:contentEnd]
				if name == "target_features" {
					parseTargetFeatures(wb, data[after:contentEnd])
				}
			}
		}
		pos = contentEnd
	}
	return wb, nil
}

// ── Analysis ──────────────────────────────────────────────────────────────────

func detectEnvironment(wb *wasmBinary) (env string, evidence string) {
	hasWASI1, hasWASI0, hasWASI2, hasEnv, hasWbindgen, hasGoJS := false, false, false, false, false, false
	for _, imp := range wb.imports {
		switch {
		case imp.module == "wasi_snapshot_preview1":
			hasWASI1 = true
		case imp.module == "wasi_snapshot_preview0" || imp.module == "wasi_unstable":
			hasWASI0 = true
		case strings.HasPrefix(imp.module, "wasi:"):
			hasWASI2 = true
		case imp.module == "env":
			hasEnv = true
		case imp.module == "go" || imp.module == "gojs":
			hasGoJS = true
		}
		if strings.HasPrefix(imp.name, "__wbindgen") {
			hasWbindgen = true
		}
	}
	switch {
	case hasWASI2:
		return "WASIp2 (component model)", `import module with "wasi:" prefix`
	case hasWASI1:
		return "WASIp1", `import module "wasi_snapshot_preview1"`
	case hasWASI0:
		return "WASIp0 (preview0/unstable)", `import module "wasi_snapshot_preview0" or "wasi_unstable"`
	case hasWbindgen:
		return "js/wasm (wasm-bindgen/Rust)", `"__wbindgen_*" imports`
	case hasGoJS:
		return `js/wasm (Go runtime)`, `import module "gojs" or "go"`
	case hasEnv:
		return "js/wasm (Emscripten)", `import module "env" with no WASI imports`
	default:
		return "bare (no environment imports)", "no WASI or env imports found"
	}
}

// detectExceptionModel replicates the logic from z3wasm/z3_wazero_test.go.
func detectExceptionModel(wb *wasmBinary) (model string, jsImports []string) {
	for _, imp := range wb.imports {
		if imp.module != "env" {
			continue
		}
		switch {
		case imp.name == "__cxa_throw":
			jsImports = append(jsImports, imp.name)
		case imp.name == "__resumeException":
			jsImports = append(jsImports, imp.name)
		case strings.HasPrefix(imp.name, "__cxa_find_matching_catch"):
			jsImports = append(jsImports, imp.name)
		}
	}
	if len(jsImports) > 0 {
		return "js-mediated (Emscripten legacy)", jsImports
	}
	if wb.hasTagSection {
		return "native-exnref", nil
	}
	return "none", nil
}

// ── Wazero feature probing ────────────────────────────────────────────────────

func allFeatureBits() api.CoreFeatures {
	return api.CoreFeaturesV2 |
		experimental.CoreFeaturesExceptionHandling |
		experimental.CoreFeaturesTailCall |
		experimental.CoreFeaturesThreads |
		experimental.CoreFeaturesExtendedConst
}

func probeFeatures(ctx context.Context, data []byte) (map[string]probeResult, error) {
	results := make(map[string]probeResult)
	full := allFeatureBits()

	// First: compile with all features to establish baseline.
	rt := wazero.NewRuntimeWithConfig(ctx,
		wazero.NewRuntimeConfigInterpreter().WithCoreFeatures(full))
	compiled, err := rt.CompileModule(ctx, data)
	if err != nil {
		_ = rt.Close(ctx)
		return results, fmt.Errorf("compile with all features: %w", err)
	}
	_ = compiled.Close(ctx)
	_ = rt.Close(ctx)

	// Probe each feature: remove its bit and see if compilation fails.
	for _, fe := range knownFeatures {
		if !fe.hasBit {
			results[fe.wasmOptName] = probeResult{probed: false}
			continue
		}
		withoutBit := full &^ fe.wazeroBit
		rt2 := wazero.NewRuntimeWithConfig(ctx,
			wazero.NewRuntimeConfigInterpreter().WithCoreFeatures(withoutBit))
		compiled2, err2 := rt2.CompileModule(ctx, data)
		required := err2 != nil
		if compiled2 != nil {
			_ = compiled2.Close(ctx)
		}
		_ = rt2.Close(ctx)
		results[fe.wasmOptName] = probeResult{required: required, probed: true}
	}
	return results, nil
}

// ── Report output ─────────────────────────────────────────────────────────────

var sectionNames = map[byte]string{
	0: "Custom", 1: "Type", 2: "Import", 3: "Function",
	4: "Table", 5: "Memory", 6: "Global", 7: "Export",
	8: "Start", 9: "Element", 10: "Code", 11: "Data",
	12: "DataCount", 13: "Tag",
}

func sectionLabel(id byte) string {
	if n, ok := sectionNames[id]; ok {
		return fmt.Sprintf("%d:%s", id, n)
	}
	return fmt.Sprintf("%d:Unknown", id)
}

func humanSize(n int64) string {
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	case n >= 1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func uniqueImportModules(wb *wasmBinary) []string {
	seen := make(map[string]bool)
	var order []string
	for _, imp := range wb.imports {
		if !seen[imp.module] {
			seen[imp.module] = true
			order = append(order, imp.module)
		}
	}
	return order
}

func printReport(wb *wasmBinary, probes map[string]probeResult, wazeroErr error) {
	fmt.Printf("=== WebAssembly Binary Analysis ===\n")
	fmt.Printf("File:    %s  (%v bytes ; %s)\n", filepath.Base(wb.path), wb.size, humanSize(wb.size))
	fmt.Printf("Path:    %s\n", wb.path)
	fmt.Printf("Version: %d\n", wb.version)
	fmt.Printf("b3: %v\n", wb.b3)

	fmt.Println()

	// Environment
	env, envEvidence := detectEnvironment(wb)
	fmt.Printf("ENVIRONMENT\n")
	fmt.Printf("  %s\n", env)
	fmt.Printf("  Evidence: %s\n", envEvidence)
	fmt.Println()

	// Exception handling
	exModel, jsImps := detectExceptionModel(wb)
	fmt.Printf("C++ EXCEPTION HANDLING\n")
	fmt.Printf("  %s\n", exModel)
	if len(jsImps) > 0 {
		fmt.Printf("  JS-mediated imports: %s\n", strings.Join(jsImps, ", "))
	} else if wb.hasTagSection {
		fmt.Printf("  Evidence: Tag section (§13) present; no __cxa_throw/__resumeException imports\n")
	}
	fmt.Println()

	// Feature table
	fmt.Printf("WEBASSEMBLY FEATURES\n")
	fmt.Printf("  (target_features = authoritative; wazero-probe may have false negatives for\n")
	fmt.Printf("   features checked at instantiation time rather than compile time)\n")
	if wazeroErr != nil {
		fmt.Printf("  (wazero probing unavailable: %v)\n", wazeroErr)
	}
	fmt.Printf("  %-28s  %-16s  %-14s  %s\n",
		"Feature", "target_features", "section-hint", "wazero-probe")
	fmt.Printf("  %s\n", strings.Repeat("─", 80))

	sectionHint := func(name string) string {
		switch name {
		case "exception-handling":
			if wb.hasTagSection {
				return "Tag §13"
			}
		case "bulk-memory":
			if wb.hasDataCountSection {
				return "DataCount §12"
			}
		}
		return ""
	}

	for _, fe := range knownFeatures {
		tfEnabled := false
		tfVal := "n/a"
		if v, ok := wb.targetFeatures[fe.lookupName()]; ok {
			tfEnabled = v
			if v {
				tfVal = "[+]"
			} else {
				tfVal = "[-]"
			}
		}
		hint := sectionHint(fe.wasmOptName)
		probeStr := "—"
		if r, ok := probes[fe.wasmOptName]; ok {
			switch {
			case !r.probed:
				probeStr = "(no wazero bit)"
			case r.required:
				probeStr = "required"
			case tfEnabled:
				// target_features says used but wazero compile-phase didn't catch it:
				// wazero validates this feature lazily (at instantiation, not compile time)
				probeStr = "not required (lazy check)"
			default:
				probeStr = "not required"
			}
		}
		fmt.Printf("  %-28s  %-16s  %-14s  %s\n",
			fe.wasmOptName, tfVal, hint, probeStr)
	}
	fmt.Println()

	// Sections inventory
	fmt.Printf("SECTIONS\n")
	seen := make(map[byte]bool)
	var secLabels []string
	for _, s := range wb.sections {
		if !seen[s.id] {
			seen[s.id] = true
			secLabels = append(secLabels, sectionLabel(s.id))
		}
	}
	fmt.Printf("  %s\n", strings.Join(secLabels, "  "))
	if len(wb.customSections) > 0 {
		names := make([]string, 0, len(wb.customSections))
		for n := range wb.customSections {
			names = append(names, n)
		}
		sort.Strings(names)
		fmt.Printf("  Custom section names: %s\n", strings.Join(names, ", "))
	}
	fmt.Println()

	// Imports
	fmt.Printf("IMPORTS  (%d total)\n", len(wb.imports))
	modOrder := uniqueImportModules(wb)
	modCounts := make(map[string]int)
	for _, imp := range wb.imports {
		modCounts[imp.module]++
	}
	for _, mod := range modOrder {
		fmt.Printf("  %-32s %d items\n", mod+":", modCounts[mod])
	}
	fmt.Println()

	// Exports
	funcCount := 0
	var notableExports []string
	notableSet := map[string]bool{
		"_start": true, "_initialize": true, "malloc": true, "free": true,
		"main": true, "memory": true,
	}
	for _, exp := range wb.exports {
		if exp.kind == kindFunction {
			funcCount++
		}
		if notableSet[exp.name] {
			notableExports = append(notableExports, exp.name)
		}
	}
	fmt.Printf("EXPORTS  (%d total, %d functions)\n", len(wb.exports), funcCount)
	if len(notableExports) > 0 {
		fmt.Printf("  Notable: %s\n", strings.Join(notableExports, ", "))
	}
	// Show first 5 function exports
	shown := 0
	var firstFuncs []string
	for _, exp := range wb.exports {
		if exp.kind == kindFunction {
			firstFuncs = append(firstFuncs, exp.name)
			shown++
			if shown >= 5 {
				break
			}
		}
	}
	if len(firstFuncs) > 0 {
		suffix := ""
		if funcCount > 5 {
			suffix = fmt.Sprintf(", ... (%d total)", funcCount)
		}
		fmt.Printf("  First funcs: %s%s\n", strings.Join(firstFuncs, ", "), suffix)
	}
}

func Blake3OfFile(path string) (blake3sum string, err error) {

	sum, _, err1 := blake3.HashFile(path)
	if err1 != nil {
		return "", err1
	}

	blake3sum = "blake3.33B-" + cristalbase64.URLEncoding.EncodeToString(sum[:33])
	return
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
