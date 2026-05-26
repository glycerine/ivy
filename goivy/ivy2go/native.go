package ivy2go

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// native.go mirrors ivy2cpp/native.go. Walks the module's native code
// blocks and emits any block tagged "go" / "go_header" / "go_init" /
// "go_inline" verbatim into the corresponding stream. Non-Go-tagged
// blocks (e.g. plain "cpp", "header") are silently skipped so the
// same Ivy source can carry blocks for both targets.

// nativeGoBlock is the parsed form of a single Ivy native block whose
// tag indicates Go emission.
type nativeGoBlock struct {
	Tag    string        // canonical tag: "go", "go_header", "go_init", "go_inline"
	Body   string        // raw Go source to insert (after antiquote substitution)
	Params []goivy.Expr  // antiquote substitution params (`0`, `1`, ...)
}

// emitNativeBlocks distributes any go-tagged native blocks to the
// streams indicated by their tag. Called from generator.go's
// emitNative wrapper.
func (g *Generator) emitNativeBlocks() {
	if g == nil || g.Mod == nil {
		return
	}
	for _, blk := range g.buildNativeBlocks() {
		g.dispatchNativeGoBlock(blk)
	}
}

// buildNativeBlocks walks g.Mod.Natives and returns those whose
// tag starts with "go".
func (g *Generator) buildNativeBlocks() []nativeGoBlock {
	var out []nativeGoBlock
	for _, node := range g.Mod.Natives {
		args := node.Args()
		if len(args) < 2 {
			continue
		}
		codeNode, ok := args[1].(*goivy.NativeCode)
		if !ok {
			continue
		}
		tag, body := splitNativeCode(codeNode.Code)
		if !strings.HasPrefix(tag, "go") {
			continue
		}
		// Trailing args (after [0] meta, [1] code) are antiquote
		// substitution params. Lift them into Expr where possible.
		params := make([]goivy.Expr, 0, len(args)-2)
		for _, p := range args[2:] {
			if expr, ok := p.(goivy.Expr); ok {
				params = append(params, expr)
				continue
			}
			if cn, ok := p.(*goivy.CompiledNode); ok {
				if expr, ok := cn.Node.(goivy.Expr); ok {
					params = append(params, expr)
					continue
				}
			}
		}
		out = append(out, nativeGoBlock{Tag: tag, Body: body, Params: params})
	}
	return out
}

// renderNativeTemplate substitutes `` `N` ``-delimited antiquotes
// in body with one of three flavours, selected by the trailing
// character of the preceding text (mirrors ivy2cpp/native.go
// renderNativeTemplate):
//
//   - default     → emitExpr(params[N])         — Go-source value reference
//   - prefix '%'  → goType(params[N].NodeSort()) — Go type expression
//   - prefix '"'  → strconv.Quote(name(params[N])) — Z3 name (quoted string)
//
// The prefix character is consumed (dropped from the preceding
// text) when matched, so a template like ` %`0` ` produces
// `<typeOf(p0)>` with no `%` left in the output.
//
// Antiquote indices that are non-numeric or out of range produce a
// `/*ivy2go: …*/` marker so the emitted Go fails cleanly at compile
// time with a clear message.
func (g *Generator) renderNativeTemplate(body string, params []goivy.Expr) string {
	if !strings.Contains(body, "`") {
		return body
	}
	fields := strings.Split(body, "`")
	for i := 1; i < len(fields); i += 2 {
		idx, err := strconv.Atoi(fields[i])
		if err != nil {
			fields[i] = "/*ivy2go: bad antiquote index `" + fields[i] + "`*/"
			continue
		}
		if idx < 0 || idx >= len(params) {
			fields[i] = "/*ivy2go: antiquote index out of range*/"
			continue
		}
		// Inspect the trailing char of the preceding field to pick
		// a flavour.
		prev := fields[i-1]
		flavour := byte(0)
		if len(prev) > 0 {
			c := prev[len(prev)-1]
			if c == '%' || c == '"' {
				flavour = c
			}
		}
		var sub string
		switch flavour {
		case '%':
			sub = g.goType(params[idx].NodeSort())
			fields[i-1] = prev[:len(prev)-1]
		case '"':
			// Z3-name: the user has already opened with `"`; we
			// emit the bare identifier so the surrounding "..." is
			// still well-formed.
			sub = goivy.ExprName(params[idx])
		default:
			code, err := g.emitExpr(params[idx])
			if err != nil {
				fields[i] = "/*ivy2go: antiquote emit error*/"
				continue
			}
			sub = code
		}
		fields[i] = sub
	}
	return strings.Join(fields, "")
}

// splitNativeCode mirrors ivy2cpp/native.go splitNativeCode but
// without the "member" default (Go has no member concept). The first
// non-blank line is the tag; the rest is the body. When the block
// has no tag line, the whole code is treated as a "go" body.
func splitNativeCode(code string) (string, string) {
	// Trim any leading blank line(s) the parser may have prepended.
	for strings.HasPrefix(code, "\n") {
		code = code[1:]
	}
	parts := strings.SplitN(code, "\n", 2)
	if len(parts) == 2 {
		tag := strings.TrimSpace(parts[0])
		if tag == "" {
			return "go", parts[1]
		}
		return tag, parts[1]
	}
	return "go", code
}

// dispatchNativeGoBlock routes a block to the right stream based on
// its tag. The mapping is:
//
//   - "go"         → native.go    (free package-level declarations)
//   - "go_header"  → types.go     (prepended type/import-adjacent decls)
//   - "go_init"    → init.go      (additions to (*State).Init body)
//   - "go_inline"  → runtime.go   (inline helper functions)
//   - other "go_X" → native.go    (with an unknown-tag marker)
//
// The blocks are emitted verbatim; we trust the user's Go source per
// ARCHITECTURE_TODO.md §3.5.10 (native blocks are trusted source).
// nativeOnceMemo dedups identical bodies.
func (g *Generator) dispatchNativeGoBlock(blk nativeGoBlock) {
	// Substitute antiquotes before dedup so different param bindings
	// produce different keys (and hence different emissions).
	rendered := g.renderNativeTemplate(blk.Body, blk.Params)
	key := blk.Tag + "|" + rendered
	if g.nativeOnceMemo[key] {
		return
	}
	g.nativeOnceMemo[key] = true
	body := strings.TrimRight(rendered, "\n") + "\n"

	switch blk.Tag {
	case "go":
		g.native.raw(body)
	case "go_header":
		g.types.raw(body)
	case "go_init":
		g.init.raw(body)
	case "go_inline":
		g.runtime.raw(body)
	default:
		g.native.linef("// unknown native tag %q (emitted into native.go):", blk.Tag)
		g.native.raw(body)
	}
	// Ensure a trailing newline so subsequent emitter calls start
	// clean.
	if !strings.HasSuffix(body, "\n") {
		g.native.blank()
	}
}

// Silence unused-import diagnostics in degenerate code paths.
var _ = fmt.Sprintf
