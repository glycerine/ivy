// Ported to Go from ivy_to_lean.py.

// Package leangen generates Lean theorem prover output from Ivy
// specifications. It translates logic sorts, symbol definitions,
// formulas, and actions into Lean's notation using the ivy_pl library
// format.
package leangen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	lg "github.com/glycerine/goivy/logic"
)

// Preamble is the standard Lean file header.
const Preamble = `
import .ivy_pl
namespace ivy
open ivy_logic

`

// Postamble is the standard Lean file footer.
const Postamble = `

end ivy
`

// Generator accumulates Lean output fragments and produces a complete
// Lean file.
type Generator struct {
	buf strings.Builder
}

// NewGenerator creates a new Lean code generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// String returns all accumulated output.
func (g *Generator) String() string {
	return g.buf.String()
}

// Emit appends a string fragment.
func (g *Generator) Emit(s string) {
	g.buf.WriteString(s)
}

// EmitEOL appends a newline.
func (g *Generator) EmitEOL() {
	g.buf.WriteByte('\n')
}

// SortToString converts a logic sort to its Lean representation.
func SortToString(s lg.Sort) (string, error) {
	switch st := s.(type) {
	case *lg.UninterpretedSort:
		return `(sort.ui "` + st.Name + `")`, nil
	case *lg.BooleanSort:
		return "sort.bool", nil
	case *lg.FunctionSort:
		domain := st.Domain()
		parts := make([]string, len(domain))
		for i, d := range domain {
			ds, err := SortToString(d)
			if err != nil {
				return "", err
			}
			parts[i] = ds
		}
		rs, err := SortToString(st.Range())
		if err != nil {
			return "", err
		}
		// Use Unicode cross and arrow
		return "(" + strings.Join(parts, " \u00d7 ") + " \u21a6 " + rs + ")", nil
	default:
		return "", fmt.Errorf("leangen: enumerated sorts not supported yet")
	}
}

// EmitSymbolDef emits a Lean definition for a symbol constant.
func (g *Generator) EmitSymbolDef(name string, s lg.Sort) error {
	ss, err := SortToString(s)
	if err != nil {
		return err
	}
	g.Emit("def " + name + " := mk_cnst \"" + name + "\" " + ss)
	g.EmitEOL()
	return nil
}

// EmitExpr emits the Lean representation of a logic node (formula/term).
func (g *Generator) EmitExpr(f lg.Expr) error {
	switch n := f.(type) {
	case *lg.Variable:
		ss, err := SortToString(n.VSort)
		if err != nil {
			return err
		}
		g.Emit(`("` + n.Name + `",` + ss + ")")
		return nil

	case *lg.Const:
		g.Emit(n.Name)
		return nil

	case *lg.Apply:
		g.Emit(n.Func.String() + "(")
		for i, t := range n.Terms {
			if i > 0 {
				g.Emit(",")
			}
			if err := g.EmitExpr(t); err != nil {
				return err
			}
		}
		g.Emit(")")
		return nil

	case *lg.Eq:
		g.Emit("(fmla.eq ")
		if err := g.EmitExpr(n.T1); err != nil {
			return err
		}
		g.Emit(" ")
		if err := g.EmitExpr(n.T2); err != nil {
			return err
		}
		g.Emit(")")
		return nil

	case *lg.Iff:
		g.Emit("(fmla.eq ")
		if err := g.EmitExpr(n.T1); err != nil {
			return err
		}
		g.Emit(" ")
		if err := g.EmitExpr(n.T2); err != nil {
			return err
		}
		g.Emit(")")
		return nil

	case *lg.Ite:
		g.Emit("(ite_fmla ")
		for _, term := range []lg.Expr{n.Cond, n.Then, n.Else} {
			g.Emit(" ")
			if err := g.EmitExpr(term); err != nil {
				return err
			}
		}
		g.Emit(")")
		return nil

	case *lg.Not:
		g.Emit("\u00ac") // NOT sign
		return g.EmitExpr(n.Body)

	case *lg.And:
		if len(n.Terms) == 0 {
			g.Emit("ltrue")
			return nil
		}
		g.Emit("(")
		for i, t := range n.Terms {
			if i > 0 {
				g.Emit("\u2227") // logical AND
			}
			if err := g.EmitExpr(t); err != nil {
				return err
			}
		}
		g.Emit(")")
		return nil

	case *lg.Or:
		if len(n.Terms) == 0 {
			g.Emit("lfalse")
			return nil
		}
		g.Emit("(")
		for i, t := range n.Terms {
			if i > 0 {
				g.Emit("\u2228") // logical OR
			}
			if err := g.EmitExpr(t); err != nil {
				return err
			}
		}
		g.Emit(")")
		return nil

	case *lg.Implies:
		g.Emit("(")
		if err := g.EmitExpr(n.T1); err != nil {
			return err
		}
		g.Emit(" \u21d2 ") // double arrow
		if err := g.EmitExpr(n.T2); err != nil {
			return err
		}
		g.Emit(")")
		return nil

	case *lg.ForAll:
		g.Emit("\u00ac") // NOT
		for _, v := range n.Variables {
			g.Emit("(fmla.proj ")
			if err := g.EmitExpr(v); err != nil {
				return err
			}
			g.Emit(" ")
		}
		g.Emit("\u00ac") // NOT
		if err := g.EmitExpr(n.Body); err != nil {
			return err
		}
		for range n.Variables {
			g.Emit(")")
		}
		return nil

	case *lg.Exists:
		for _, v := range n.Variables {
			g.Emit("(fmla.proj ")
			if err := g.EmitExpr(v); err != nil {
				return err
			}
			g.Emit(" ")
		}
		if err := g.EmitExpr(n.Body); err != nil {
			return err
		}
		for range n.Variables {
			g.Emit(")")
		}
		return nil

	case *lg.Lambda:
		for _, v := range n.Variables {
			g.Emit("(fmla.lambda ")
			if err := g.EmitExpr(v); err != nil {
				return err
			}
			g.Emit(" ")
		}
		if err := g.EmitExpr(n.Body); err != nil {
			return err
		}
		for range n.Variables {
			g.Emit(")")
		}
		return nil

	default:
		return fmt.Errorf("leangen: unsupported expression type %T", f)
	}
}

// EmitAction emits the Lean representation of an action.
func (g *Generator) EmitAction(a actions.Action) error {
	switch act := a.(type) {
	case *actions.AssignAction:
		g.Emit("    " + fmt.Sprint(act.LHS) + " ::= ")
		return g.EmitExpr(act.RHS)

	case *actions.Sequence:
		g.Emit("(")
		for i, child := range act.Elems {
			if i > 0 {
				g.Emit(";\n")
			}
			childAct, _ := child.(actions.Action)
			if childAct != nil {
				if err := g.EmitAction(childAct); err != nil {
					return err
				}
			}
		}
		g.Emit(")")
		return nil

	case *actions.IfAction:
		g.Emit("(PL.pterm.ite ")
		if err := g.EmitExpr(act.Cond); err != nil {
			return err
		}
		thenAct, _ := act.ThenBody.(actions.Action)
		if thenAct != nil {
			if err := g.EmitAction(thenAct); err != nil {
				return err
			}
		}
		if act.ElseBody != nil {
			elseAct, _ := act.ElseBody.(actions.Action)
			if elseAct != nil {
				if err := g.EmitAction(elseAct); err != nil {
					return err
				}
			}
		} else {
			g.Emit("PL.pterm.skip")
		}
		return nil

	default:
		return fmt.Errorf("leangen: action type %T not supported yet", a)
	}
}

// GenerateProgram generates a complete Lean file from symbols and
// action/export definitions.
func (g *Generator) GenerateProgram(
	symbols []SymbolDef,
	actionMap map[string]actions.Action,
	publicActions []string,
	moduleName string,
) error {
	g.Emit(Preamble)

	for _, sym := range symbols {
		if err := g.EmitSymbolDef(sym.Name, sym.Sort); err != nil {
			return err
		}
	}

	g.Emit("def P := (PL.prog.letrec")
	g.EmitEOL()

	// Sorted action names for deterministic output
	actionNames := make([]string, 0, len(actionMap))
	for name := range actionMap {
		actionNames = append(actionNames, name)
	}
	sort.Strings(actionNames)

	for _, name := range actionNames {
		act := actionMap[name]
		g.Emit("  (@bndng.cons PL.pterm (\"" + name + "\",\n")
		if err := g.EmitAction(act); err != nil {
			return err
		}
		g.EmitEOL()
		g.Emit("    ) ")
		g.EmitEOL()
	}

	g.Emit("  (bndng.default PL.pterm.skip)")
	for range actionNames {
		g.Emit(")")
	}
	g.EmitEOL()

	exports := make([]string, len(publicActions))
	copy(exports, publicActions)
	sort.Strings(exports)

	if len(exports) == 0 {
		g.Emit("PL.pterm.skip")
	} else {
		parts := make([]string, len(exports))
		for i, x := range exports {
			parts[i] = "(PL.pterm.call \"" + x + "\")"
		}
		g.Emit("((" + strings.Join(parts, " + ") + ")**)")
	}

	g.Emit(Postamble)
	return nil
}

// SymbolDef holds a symbol name and its sort for program generation.
type SymbolDef struct {
	Name string
	Sort lg.Sort
}
