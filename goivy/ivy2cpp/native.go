package ivy2cpp

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

type nativeBlock struct {
	tag    string
	code   string
	params []goivy.Expr
}

func (g *Generator) emitNativeBlocks(w *cppWriter, tags ...string) error {
	blocks, err := g.nativeBlocks()
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, block := range blocks {
		if !nativeTagMatches(block.tag, tags...) {
			continue
		}
		rendered, err := g.renderNativeTemplate(block.code, block.params)
		if err != nil {
			return err
		}
		if (block.tag == "header" || block.tag == "impl") && seen[rendered] {
			continue
		}
		seen[rendered] = true
		emitNativeLines(w, rendered)
	}
	return nil
}

func (g *Generator) nativeBlocks() ([]nativeBlock, error) {
	if g == nil || g.Mod == nil {
		return nil, nil
	}
	blocks := make([]nativeBlock, 0, len(g.Mod.Natives))
	for _, node := range g.Mod.Natives {
		args := node.Args()
		if len(args) < 2 {
			continue
		}
		codeNode, ok := args[1].(*goivy.NativeCode)
		if !ok {
			return nil, fmt.Errorf("ivy2cpp: native block has code %T", args[1])
		}
		tag, code := splitNativeCode(codeNode.Code)
		params := make([]goivy.Expr, 0, len(args)-2)
		for _, arg := range args[2:] {
			expr, err := nativeExpr(arg)
			if err != nil {
				return nil, err
			}
			params = append(params, expr)
		}
		blocks = append(blocks, nativeBlock{tag: normalizeNativeTag(tag), code: code, params: params})
	}
	return blocks, nil
}

func splitNativeCode(code string) (string, string) {
	parts := strings.SplitN(code, "\n", 2)
	if len(parts) == 2 {
		tag := strings.TrimSpace(parts[0])
		if tag == "" {
			tag = "member"
		}
		return tag, parts[1]
	}
	return "member", code
}

func nativeExpr(node goivy.Node) (goivy.Expr, error) {
	switch n := node.(type) {
	case *goivy.CompiledNode:
		if expr, ok := n.Node.(goivy.Expr); ok {
			return expr, nil
		}
		if nested, ok := n.Node.(goivy.Node); ok {
			return nativeExpr(nested)
		}
		return nil, fmt.Errorf("ivy2cpp: native antiquote is compiled %T, not an expression", n.Node)
	case goivy.Expr:
		return n, nil
	default:
		return nil, fmt.Errorf("ivy2cpp: native antiquote %T is not an expression", node)
	}
}

func nativeTagMatches(tag string, targets ...string) bool {
	tag = normalizeNativeTag(tag)
	for _, target := range targets {
		if tag == normalizeNativeTag(target) {
			return true
		}
	}
	return false
}

func normalizeNativeTag(tag string) string {
	switch strings.TrimSpace(tag) {
	case "once":
		return "header"
	default:
		return strings.TrimSpace(tag)
	}
}
