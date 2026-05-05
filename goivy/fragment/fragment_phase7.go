// phase7.go implements Phase 7 helper functions for the fragment package,
// ported from Python ivy_fragment.py.
package fragment

import (
	"fmt"
	"strings"
)

// ShowStratGraph prints the stratification graph nodes and arcs for
// debugging purposes.
// Corresponds to Python's show_strat_graph (ivy_fragment.py).
func ShowStratGraph(nodes map[string]interface{}, arcs []StratArc) string {
	var sb strings.Builder

	sb.WriteString("nodes = {\n")
	for key, node := range nodes {
		sb.WriteString(fmt.Sprintf("  %s : %v\n", key, node))
	}
	sb.WriteString("}\n")

	sb.WriteString("arcs = {\n")
	for _, arc := range arcs {
		sb.WriteString(fmt.Sprintf("  (%s)\n", arc.String()))
	}
	sb.WriteString("}\n")

	return sb.String()
}

// StratArc represents an arc in the stratification graph.
type StratArc struct {
	From  string
	To    string
	Fmla  string
	Line  int
	Extra int // optional: argument position index
}

// String formats a StratArc for display.
func (a StratArc) String() string {
	parts := []string{a.From, a.To, a.Fmla, fmt.Sprint(a.Line)}
	if a.Extra >= 0 {
		parts = append(parts, fmt.Sprint(a.Extra))
	}
	return strings.Join(parts, ",")
}
