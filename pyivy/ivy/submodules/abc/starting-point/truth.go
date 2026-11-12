package startingpoint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func truthKind(typ string) string {
	switch typ {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "byte", "float32", "float64":
		return "number"
	case "bool":
		return "bool"
	}
	if strings.HasPrefix(typ, "pointer:") || typ == "file" {
		return "pointer"
	}
	return ""
}

func booleanExpression(expr string, types, returns map[string]string) string {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return expr
	}
	var edits []Edit
	var visit func(ast.Expr)
	visit = func(n ast.Expr) {
		switch v := n.(type) {
		case *ast.ParenExpr:
			visit(v.X)
			return
		case *ast.UnaryExpr:
			if v.Op == token.NOT {
				visit(v.X)
			}
			return
		case *ast.BinaryExpr:
			if v.Op == token.LAND || v.Op == token.LOR {
				visit(v.X)
				visit(v.Y)
			}
			return
		}
		kind := ""
		switch v := n.(type) {
		case *ast.Ident:
			kind = truthKind(types[v.Name])
		case *ast.CallExpr:
			if name, ok := v.Fun.(*ast.Ident); ok {
				kind = truthKind(returns[name.Name])
				if kind == "" {
					kind = truthKind(name.Name)
				}
			}
		case *ast.BasicLit:
			if v.Kind == token.INT {
				if value, err := strconv.ParseInt(v.Value, 0, 64); err == nil {
					edits = append(edits, Edit{int(v.Pos()) - 1, int(v.End()) - 1, strconv.FormatBool(value != 0)})
					return
				}
			}
		}
		suffix := ""
		if kind == "number" {
			suffix = " != 0"
		}
		if kind == "pointer" {
			suffix = " != nil"
		}
		if suffix != "" {
			start, end := int(n.Pos())-1, int(n.End())-1
			edits = append(edits, Edit{start, end, "(" + expr[start:end] + suffix + ")"})
		}
	}
	visit(node)
	out, err := ApplyEdits(expr, edits)
	if err != nil {
		return expr
	}
	return out
}

func (p *libraryPass) returnTypes() map[string]string {
	if p.returns != nil {
		return p.returns
	}
	result := map[string]string{}
	for i, t := range p.ts {
		if t.Text != "func" || i+2 >= len(p.ts) || p.ts[i+2].Text != "(" {
			continue
		}
		close := p.pairs[i+2]
		if close+1 >= len(p.ts) {
			continue
		}
		typ := p.ts[close+1].Text
		if typ == "*" && close+2 < len(p.ts) {
			typ = "pointer:" + p.ts[close+2].Text
		}
		result[p.ts[i+1].Text] = typ
	}
	p.returns = result
	return result
}

func (p *libraryPass) rewriteTruthHeaders(source string) (string, error) {
	ts, pairs, err := significantTokens(source)
	if err != nil {
		return "", err
	}
	var edits []Edit
	returns := p.returnTypes()
	for i, t := range ts {
		if t.Text != "if" && t.Text != "for" {
			continue
		}
		start, end := i+1, i+1
		if start >= len(ts) {
			continue
		}
		var semis []int
		for end < len(ts) && ts[end].Text != "{" {
			if ts[end].Text == ";" {
				semis = append(semis, end)
			}
			if close, ok := pairs[end]; ok && close > end {
				end = close
			}
			end++
		}
		if end >= len(ts) || start == end {
			continue
		}
		stop := end
		if t.Text == "for" && len(semis) == 2 {
			start = semis[0] + 1
			stop = semis[1]
		}
		if t.Text == "if" && len(semis) == 1 {
			start = semis[0] + 1
		}
		if start >= stop {
			continue
		}
		begin, finish := ts[start].Start, ts[stop-1].End
		old := source[begin:finish]
		converted := booleanExpression(old, p.types, returns)
		if converted != old && commentsRetained(old, converted) {
			edits = append(edits, Edit{begin, finish, converted})
		}
	}
	out, err := ApplyEdits(source, edits)
	if err != nil {
		return "", err
	}
	return unwrapControlHeaders(out)
}
