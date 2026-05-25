package ivy2go

import (
	"fmt"
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
	Tag  string // canonical tag: "go", "go_header", "go_init", "go_inline"
	Body string // raw Go source to insert
}

// emitNativeBlocks distributes any go-tagged native blocks to the
// streams indicated by their tag. Called from generator.go's
// emitNative wrapper.
func (g *Generator) emitNativeBlocks() {
	if g == nil || g.Mod == nil {
		return
	}
	for _, blk := range g.collectNativeGoBlocks() {
		g.dispatchNativeGoBlock(blk)
	}
}

// collectNativeGoBlocks walks g.Mod.Natives and returns those whose
// tag starts with "go".
func (g *Generator) collectNativeGoBlocks() []nativeGoBlock {
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
		tag, body := splitNativeGoCode(codeNode.Code)
		if !strings.HasPrefix(tag, "go") {
			continue
		}
		out = append(out, nativeGoBlock{Tag: tag, Body: body})
	}
	return out
}

// splitNativeGoCode mirrors ivy2cpp/native.go splitNativeCode but
// without the "member" default (Go has no member concept). The first
// non-blank line is the tag; the rest is the body. When the block
// has no tag line, the whole code is treated as a "go" body.
func splitNativeGoCode(code string) (string, string) {
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
	key := blk.Tag + "|" + blk.Body
	if g.nativeOnceMemo[key] {
		return
	}
	g.nativeOnceMemo[key] = true
	body := strings.TrimRight(blk.Body, "\n") + "\n"

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
