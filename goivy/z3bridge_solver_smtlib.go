// smtlib.go provides SMT-LIB2 output for Ivy clauses.
// Corresponds to Python's ivy_smtlib.py.
package goivy

import (
	"fmt"
	"strings"
)

// ClausesToSMTLIB2 converts a Clauses set to an SMT-LIB2 string.
// Corresponds to Python's clauses_to_smtlib.
func ClausesToSMTLIB2(clauses *Clauses) string {
	if clauses == nil {
		return "(assert true)\n(check-sat)\n"
	}

	var sb strings.Builder
	sb.WriteString("; SMT-LIB2 encoding of Ivy clauses\n")
	sb.WriteString("(set-logic ALL)\n")

	// Output each formula as an assertion
	for i, f := range clauses.Fmlas {
		sb.WriteString(fmt.Sprintf("; formula %d\n", i))
		sb.WriteString(fmt.Sprintf("(assert %s)\n", formulaToSMTLIB2(f)))
	}

	// Output definitions
	for _, d := range clauses.Defs {
		sb.WriteString(fmt.Sprintf("; definition\n"))
		sb.WriteString(fmt.Sprintf("(assert %s)\n", formulaToSMTLIB2(d.Lhs)))
	}

	sb.WriteString("(check-sat)\n")
	return sb.String()
}

// formulaToSMTLIB2 converts a formula to SMT-LIB2 syntax (simplified).
func formulaToSMTLIB2(node Expr) string {
	if node == nil {
		return "true"
	}
	switch n := node.(type) {
	case *LogicAnd:
		if len(n.Terms) == 0 {
			return "true"
		}
		args := make([]string, len(n.Terms))
		for i, t := range n.Terms {
			args[i] = formulaToSMTLIB2(t)
		}
		return "(and " + strings.Join(args, " ") + ")"
	case *LogicOr:
		if len(n.Terms) == 0 {
			return "false"
		}
		args := make([]string, len(n.Terms))
		for i, t := range n.Terms {
			args[i] = formulaToSMTLIB2(t)
		}
		return "(or " + strings.Join(args, " ") + ")"
	case *LogicNot:
		return "(not " + formulaToSMTLIB2(n.Body) + ")"
	case *LogicImplies:
		return "(=> " + formulaToSMTLIB2(n.T1) + " " + formulaToSMTLIB2(n.T2) + ")"
	case *LogicIff:
		return "(= " + formulaToSMTLIB2(n.T1) + " " + formulaToSMTLIB2(n.T2) + ")"
	case *Eq:
		return "(= " + formulaToSMTLIB2(n.T1) + " " + formulaToSMTLIB2(n.T2) + ")"
	case *Const:
		return "|" + n.Name + "|"
	case *LogicVariable:
		return "|" + n.Name + "|"
	case *ForAll:
		return "(forall (...) " + formulaToSMTLIB2(n.Body) + ")"
	case *RawForAll:
		return "(forall (...) " + formulaToSMTLIB2(n.Body) + ")"
	case *LogicExists:
		return "(exists (...) " + formulaToSMTLIB2(n.Body) + ")"
	default:
		return fmt.Sprint(node)
	}
}
