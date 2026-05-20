package ivy2cpp

import (
	"fmt"
	"strconv"
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

func (g *Generator) renderNativeTemplate(code string, params []goivy.Expr) (string, error) {
	fields := strings.Split(code, "`")
	for i := 1; i < len(fields); i += 2 {
		idx, err := strconv.Atoi(fields[i])
		if err != nil {
			return "", fmt.Errorf("bad native antiquote index %q", fields[i])
		}
		if idx < 0 || idx >= len(params) {
			return "", fmt.Errorf("native antiquote index %d out of range", idx)
		}
		prev := fields[i-1]
		var repl string
		switch {
		case strings.HasSuffix(prev, "%"):
			repl, err = g.nativeTypeOf(params[idx])
		case strings.HasSuffix(prev, `"`):
			repl, err = g.nativeZ3Name(params[idx])
		default:
			repl, err = g.nativeReference(params[idx])
		}
		if err != nil {
			return "", err
		}
		fields[i] = repl
	}
	for i := 0; i < len(fields); i += 2 {
		if strings.HasSuffix(fields[i], "%") {
			fields[i] = strings.TrimSuffix(fields[i], "%")
		}
	}
	return strings.Join(fields, ""), nil
}

func (g *Generator) nativeReference(arg goivy.Expr) (string, error) {
	switch a := arg.(type) {
	case *goivy.Const:
		if s, ok := g.sortByName(a.Name); ok {
			return cppType(s), nil
		}
		return varName(a.Name), nil
	case *goivy.LogicVariable:
		return varName(a.Name), nil
	case *goivy.Apply:
		return g.emitExpr(a)
	case *goivy.UninterpretedSort:
		return cppType(a), nil
	case *goivy.LogicEnumeratedSort:
		return cppType(a), nil
	case *goivy.RangeSort:
		return cppType(a), nil
	}
	if rn, ok := arg.(interface{ Relname() string }); ok {
		name := rn.Relname()
		if s, ok := g.sortByName(name); ok {
			return cppType(s), nil
		}
		res := varName(name)
		for _, child := range arg.Args() {
			expr, ok := child.(goivy.Expr)
			if !ok {
				return "", fmt.Errorf("native reference %s has non-expression argument %T", name, child)
			}
			idx, err := g.nativeReference(expr)
			if err != nil {
				return "", err
			}
			res += "[" + idx + "]"
		}
		return res, nil
	}
	return g.emitExpr(arg)
}

func (g *Generator) nativeTypeOf(arg goivy.Expr) (string, error) {
	if rn, ok := arg.(interface{ Relname() string }); ok {
		name := rn.Relname()
		if g.Mod != nil && g.Mod.Actions != nil {
			if _, ok := g.Mod.Actions.Get2(name); ok {
				return "thunk__" + varName(name), nil
			}
		}
	}
	return cppType(arg.NodeSort()), nil
}

func (g *Generator) nativeZ3Name(arg goivy.Expr) (string, error) {
	if v, ok := arg.(*goivy.LogicVariable); ok {
		return sortName(v.VSort), nil
	}
	if rn, ok := arg.(interface{ Relname() string }); ok {
		return rn.Relname(), nil
	}
	if c, ok := arg.(*goivy.Const); ok {
		return c.Name, nil
	}
	return "", fmt.Errorf("cannot emit native z3 name for %T", arg)
}

func (g *Generator) sortByName(name string) (goivy.Sort, bool) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return nil, false
	}
	return g.Mod.Sig.Sorts.Get2(name)
}

func emitNativeLines(w *cppWriter, code string) {
	code = strings.TrimRight(code, " \t\r\n")
	lines := strings.Split(code, "\n")
	base := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := nativeIndent(line)
		if base == -1 || indent < base {
			base = indent
		}
	}
	if base < 0 {
		return
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			w.line("")
			continue
		}
		indent := nativeIndent(line) - base
		if indent < 0 {
			indent = 0
		}
		w.line(strings.Repeat(" ", indent) + strings.TrimSpace(line))
	}
}

func nativeIndent(line string) int {
	indent := 0
	for _, r := range line {
		switch r {
		case ' ':
			indent++
		case '\t':
			indent = ((indent + 8) / 8) * 8
		default:
			return indent
		}
	}
	return indent
}
