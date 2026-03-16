// smtlib.go provides SMT-LIB2 output for Ivy clauses.
// Corresponds to Python's ivy_smtlib.py.
package solver

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
)

// ClausesToSMTLIB2 converts a Clauses set to an SMT-LIB2 string.
// Corresponds to Python's clauses_to_smtlib.
func ClausesToSMTLIB2(clauses *clauseops.Clauses) string {
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
func formulaToSMTLIB2(node lg.Node) string {
	if node == nil {
		return "true"
	}
	switch n := node.(type) {
	case *lg.And:
		if len(n.Terms) == 0 {
			return "true"
		}
		args := make([]string, len(n.Terms))
		for i, t := range n.Terms {
			args[i] = formulaToSMTLIB2(t)
		}
		return "(and " + strings.Join(args, " ") + ")"
	case *lg.Or:
		if len(n.Terms) == 0 {
			return "false"
		}
		args := make([]string, len(n.Terms))
		for i, t := range n.Terms {
			args[i] = formulaToSMTLIB2(t)
		}
		return "(or " + strings.Join(args, " ") + ")"
	case *lg.Not:
		return "(not " + formulaToSMTLIB2(n.Body) + ")"
	case *lg.Implies:
		return "(=> " + formulaToSMTLIB2(n.T1) + " " + formulaToSMTLIB2(n.T2) + ")"
	case *lg.Iff:
		return "(= " + formulaToSMTLIB2(n.T1) + " " + formulaToSMTLIB2(n.T2) + ")"
	case *lg.Eq:
		return "(= " + formulaToSMTLIB2(n.T1) + " " + formulaToSMTLIB2(n.T2) + ")"
	case *lg.Const:
		return "|" + n.Name + "|"
	case *lg.Var:
		return "|" + n.VName + "|"
	case *lg.ForAll:
		return "(forall (...) " + formulaToSMTLIB2(n.Body) + ")"
	case *lg.Exists:
		return "(exists (...) " + formulaToSMTLIB2(n.Body) + ")"
	default:
		return fmt.Sprint(node)
	}
}
