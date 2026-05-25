package ivy2go

import (
	"fmt"
	"strings"
)

// GoText is the buffered code stream used by goWriter. It mirrors
// ivy2cpp/cpp_context.go's CppText: appends typed code chunks and lets
// callers flatten them to a final string via GetFile.
type GoText struct {
	Code []any
}

func NewGoText() *GoText {
	return &GoText{}
}

func (t *GoText) Close() {}

func (t *GoText) Write(code any) {
	if t == nil {
		panic("ivy2go: write to nil GoText")
	}
	t.Code = append(t.Code, code)
}

func (t *GoText) GetFile() string {
	if t == nil {
		return ""
	}
	var b strings.Builder
	for _, code := range t.Code {
		b.WriteString(fmt.Sprint(code))
	}
	return b.String()
}

// GoContext holds one named stream per output .go file in the emitted
// package directory. Unlike ivy2cpp/CppContext (which has just Globals
// and Impls), ivy2go has one stream per generated file because the
// emitted Go package is multi-file (per ARCHITECTURE_TODO.md §3.3).
//
// Per-stream conventions:
//   - Types       → types.go        — sort declarations, struct types.
//   - State       → state.go        — State struct + NewState + helpers.
//   - Actions     → actions.go      — methods on *State for each action.
//   - Init        → init.go         — (*State).Init() body.
//   - Runtime     → runtime.go      — package helpers: ivyAssert, etc.
//   - Nondet      → nondet.go       — choose/havoc helpers.
//   - Extensional → extensional.go  — extensional relation iteration.
//   - Definitions → definitions.go  — pure-function definitional axioms.
//   - Thunks      → thunk.go        — hash-thunk struct types.
//   - Native      → native.go       — user-supplied native Go blocks.
//   - Repl        → repl.go         — REPL command reader (target=repl).
//   - Main        → main.go         — func main() (when EmitMain).
//
// Each stream's import block is collected separately so emitted files
// import only what they actually use.
type GoContext struct {
	Types       *GoText
	State       *GoText
	Actions     *GoText
	Init        *GoText
	Runtime     *GoText
	Nondet      *GoText
	Extensional *GoText
	Definitions *GoText
	Thunks      *GoText
	Native      *GoText
	Repl        *GoText
	Main        *GoText

	// Imports collected per stream. Keyed by stream name ("types",
	// "state", …) and then by import path. A non-empty value is the
	// optional alias (e.g., `goivy` for "github.com/glycerine/ivy/goivy").
	Imports map[string]map[string]string

	// PackageName is the Go package the generated files declare.
	PackageName string

	// OnceGlobals deduplicates emission of helpers that may be
	// requested multiple times (e.g., ite helpers per sort).
	OnceGlobals map[string]bool

	tempCounter int
}

func NewGoContext() *GoContext {
	return &GoContext{
		Types:       NewGoText(),
		State:       NewGoText(),
		Actions:     NewGoText(),
		Init:        NewGoText(),
		Runtime:     NewGoText(),
		Nondet:      NewGoText(),
		Extensional: NewGoText(),
		Definitions: NewGoText(),
		Thunks:      NewGoText(),
		Native:      NewGoText(),
		Repl:        NewGoText(),
		Main:        NewGoText(),
		Imports:     map[string]map[string]string{},
		OnceGlobals: map[string]bool{},
	}
}

// AddImport records that the named stream needs to import path. alias
// may be empty for an unaliased import.
func (c *GoContext) AddImport(stream, path, alias string) {
	if c.Imports[stream] == nil {
		c.Imports[stream] = map[string]string{}
	}
	if existing, ok := c.Imports[stream][path]; ok && existing != "" && alias == "" {
		return
	}
	c.Imports[stream][path] = alias
}

// AddOnceGlobal writes code to Runtime if the same code hasn't been
// written before. The dedup key is the code text itself. Mirrors
// ivy2cpp/cpp_context.go AddOnceGlobal.
func (c *GoContext) AddOnceGlobal(code any) {
	key := fmt.Sprint(code)
	if c.OnceGlobals[key] {
		return
	}
	c.OnceGlobals[key] = true
	c.Runtime.Write(code)
}

// GetTemp returns a fresh per-context temporary variable name. Mirrors
// ivy2cpp CppContext.GetTemp.
func (c *GoContext) GetTemp() string {
	res := fmt.Sprintf("__temp__%d", c.tempCounter)
	c.tempCounter++
	return res
}

// renderImportBlock produces a Go import block for the given stream.
// Returns "" when the stream has no recorded imports.
func (c *GoContext) renderImportBlock(stream string) string {
	imports := c.Imports[stream]
	if len(imports) == 0 {
		return ""
	}
	keys := make([]string, 0, len(imports))
	for k := range imports {
		keys = append(keys, k)
	}
	sortStrings(keys)

	var stdlib, third []string
	for _, k := range keys {
		if strings.Contains(k, ".") {
			third = append(third, k)
		} else {
			stdlib = append(stdlib, k)
		}
	}

	var b strings.Builder
	b.WriteString("import (\n")
	for _, k := range stdlib {
		writeImportLine(&b, k, imports[k])
	}
	if len(stdlib) > 0 && len(third) > 0 {
		b.WriteString("\n")
	}
	for _, k := range third {
		writeImportLine(&b, k, imports[k])
	}
	b.WriteString(")\n")
	return b.String()
}

func writeImportLine(b *strings.Builder, path, alias string) {
	b.WriteString("\t")
	if alias != "" {
		b.WriteString(alias)
		b.WriteString(" ")
	}
	b.WriteString("\"")
	b.WriteString(path)
	b.WriteString("\"\n")
}

// sortStrings is a tiny shim so go_context.go doesn't pull "sort" for
// what is otherwise a one-line use. Insertion sort is fine for the
// handful of imports per stream.
func sortStrings(xs []string) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j-1] > xs[j]; j-- {
			xs[j-1], xs[j] = xs[j], xs[j-1]
		}
	}
}
