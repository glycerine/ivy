// phase7.go implements Phase 7 helper functions for the art package,
// ported from Python ivy_art.py.
package goivy

import (
	"fmt"
	"strings"
)

// RenderRg renders an AnalysisGraph into a string representation for display.
// Shows actual formula strings (not just clause counts), matching the
// information density of Python's render_rg (ivy_art.py).
func RenderRg(rg *AnalysisGraph) string {
	if rg == nil {
		return "(nil graph)"
	}
	var sb strings.Builder

	sb.WriteString("States:\n")
	for _, s := range rg.States {
		label := fmt.Sprintf("  [%d]", s.ID)
		if s.IsBottom() {
			label += " (bottom)"
		} else if s.Clauses != nil {
			openFmla := s.Clauses.ToOpenFormula()
			if and, ok := openFmla.(*And); ok && len(and.Terms) > 0 {
				for _, term := range and.Terms {
					label += fmt.Sprintf("\n    %s", term.String())
				}
			}
		}
		sb.WriteString(label + "\n")
	}

	sb.WriteString("Transitions:\n")
	for _, t := range rg.Transitions {
		lbl := t.Label
		if lbl == "" {
			lbl = "(unlabeled)"
		}
		preID, postID := -1, -1
		if t.Pre != nil {
			preID = t.Pre.ID
		}
		if t.Post != nil {
			postID = t.Post.ID
		}
		sb.WriteString(fmt.Sprintf("  %d --%s--> %d\n", preID, lbl, postID))
	}

	sb.WriteString("Covering:\n")
	for _, cp := range rg.Covering {
		coveredID, coveringID := -1, -1
		if cp.Covered != nil {
			coveredID = cp.Covered.ID
		}
		if cp.Covering != nil {
			coveringID = cp.Covering.ID
		}
		sb.WriteString(fmt.Sprintf("  %d covers %d\n", coveredID, coveringID))
	}

	return sb.String()
}
