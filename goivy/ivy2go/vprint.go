package ivy2go

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// vprint.go mirrors ivy2cpp/vprint.go. Trace LHS/RHS printing wires
// into Config.Trace and the assignment emitter (assign.go).
// OPEN 058 lands the real emission below.

// numberFormat returns the trace number format flag suffix. Mirrors
// ivy2cpp/vprint.go numberFormat. Returns the empty string until a
// module attribute requests hex tracing.
func (g *Generator) numberFormat() string {
	if g == nil {
		return ""
	}
	if g.numberFormatComputed {
		return g.numberFormatCache
	}
	g.numberFormatComputed = true
	g.numberFormatCache = ""
	// Scan module attributes for "radix" → "hex". Mirrors
	// ivy2cpp/vprint.go's check.
	if g.Mod != nil {
		for k, attr := range g.Mod.Attributes.All() {
			if attr == nil {
				continue
			}
			s := fmt.Sprintf("%s=%v", k, attr)
			if strings.Contains(s, "radix") && strings.Contains(s, "hex") {
				g.numberFormatCache = "hex"
				break
			}
		}
	}
	return g.numberFormatCache
}

// emitTracedLHS emits an fmt.Fprintln call to ivyTraceOut describing
// the assignment `lhs = rhsCode`. Mirrors ivy2cpp/assign.go
// emitTracedLHS, restructured for Go: instead of ostream chaining we
// build a single fmt.Fprintf format string + args.
//
// The trace shape is `write(<lhs-name>(<arg0>,<arg1>),<rhs>)` so it's
// stable across the C++ / Go emission split.
func (g *Generator) emitTracedLHS(w *goWriter, lhs goivy.Expr, rhsCode string) {
	if g == nil || w == nil {
		return
	}
	if lhsHasNamespacedName(lhs) {
		return
	}
	g.Ctx.AddImport("actions", "fmt", "")
	format, args := g.traceFormatFor(lhs, rhsCode)
	if len(args) == 0 {
		w.linef(`fmt.Fprintln(ivyTraceOut, %q)`, format)
		return
	}
	w.linef(`fmt.Fprintf(ivyTraceOut, %q, %s)`, format+"\n", strings.Join(args, ", "))
}

// traceFormatFor walks the LHS to build a format string with %v
// placeholders for any applied arguments and one trailing %v for the
// assigned value (rhsCode). Returns (format, args).
func (g *Generator) traceFormatFor(lhs goivy.Expr, rhsCode string) (string, []string) {
	switch n := lhs.(type) {
	case *goivy.Const:
		return fmt.Sprintf("write(%s,%%v)", escapeFmt(n.Name)), []string{rhsCode}
	case *goivy.Apply:
		name := goivy.ExprName(n.Func)
		var format strings.Builder
		args := []string{}
		format.WriteString("write(")
		format.WriteString(escapeFmt(name))
		if len(n.Terms) > 0 {
			format.WriteString("(")
			for i, t := range n.Terms {
				if i > 0 {
					format.WriteString(",")
				}
				format.WriteString("%v")
				if code, err := g.emitExpr(t); err == nil {
					args = append(args, code)
				} else {
					args = append(args, "nil")
				}
			}
			format.WriteString(")")
		}
		format.WriteString(",%v)")
		args = append(args, rhsCode)
		return format.String(), args
	}
	return "write(?,%v)", []string{rhsCode}
}

// lhsHasNamespacedName mirrors ivy2cpp/assign.go lhsHasNamespacedName.
// Suppresses traces for LHSes whose root symbol carries a `:`-prefixed
// internal namespace (loc:, ext:, etc).
func lhsHasNamespacedName(lhs goivy.Expr) bool {
	return strings.Contains(lhsRootName(lhs), ":")
}

// lhsRootName mirrors ivy2cpp/assign.go lhsRootName.
func lhsRootName(lhs goivy.Expr) string {
	for {
		switch v := lhs.(type) {
		case *goivy.Apply:
			lhs = v.Func
		case *goivy.Const:
			return v.Name
		default:
			return goivy.ExprName(v)
		}
	}
}

// escapeFmt escapes a literal string for use inside an fmt format
// string: `%` becomes `%%`. We don't need to escape much else because
// the format is wrapped in %q at emission time.
func escapeFmt(s string) string {
	return strings.ReplaceAll(s, "%", "%%")
}
