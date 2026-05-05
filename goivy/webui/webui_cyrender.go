package webui

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
)

// -----------------------------------------------------------------------
// Type aliases — these types now live in art/cyrender.go.
// Existing webui code continues to compile via these aliases.
// -----------------------------------------------------------------------

type WebUINodeAction = goivy.NodeAction
type WebUICyElements = goivy.CyElements
type WebUICyElement = goivy.CyElement
type WebUIAnalysisGraphState = goivy.AnalysisGraphState
type WebUIARGNode = goivy.ARGNode
type WebUIARGTransition = goivy.ARGTransition
type WebUIARGCover = goivy.ARGCover

// Forwarding functions — delegate to art package.

var NewWebUICyElements = goivy.NewCyElements
var NewWebUIAnalysisGraphState = goivy.NewAnalysisGraphState
var RenderWebUIARG = goivy.RenderARG

// -----------------------------------------------------------------------
// Proof goal rendering — stays in webui (only used by webui).
// -----------------------------------------------------------------------

// WebUIProofGoal is a lightweight proof goal for rendering.
type WebUIProofGoal struct {
	ID       int
	Label    string
	Refuted  bool
	Info     string
	ParentID int // -1 if no parent
}

// WebUIProofStack holds goals for rendering.
type WebUIProofStack struct {
	Goals []WebUIProofGoal
}

// RenderProofStack converts a WebUIProofStack into Cytoscape elements.
func RenderProofStack(stack *WebUIProofStack) *WebUICyElements {
	g := goivy.NewCyElements()
	if stack == nil {
		return g
	}
	for _, goal := range stack.Goals {
		classes := []string{"proof_goal"}
		info := fmt.Sprintf("Goal %d", goal.ID)
		if goal.Refuted {
			classes = append(classes, "refuted")
			info += " (refuted)"
		}
		g.AddNode(
			fmt.Sprintf("goal_%d", goal.ID),
			goal.Label,
			classes,
			info,
			goal.Info,
			nil,
			"ellipse",
		)
	}
	for _, goal := range stack.Goals {
		if goal.ParentID >= 0 {
			g.AddEdge(
				fmt.Sprintf("parent_%d_%d", goal.ID, goal.ParentID),
				fmt.Sprintf("goal_%d", goal.ID),
				fmt.Sprintf("goal_%d", goal.ParentID),
				"",
				[]string{"proof_edge"},
				"",
				"",
			)
		}
	}
	return g
}

// -----------------------------------------------------------------------
// Concept graph rendering — stays in webui (depends on ConceptSession).
// -----------------------------------------------------------------------

// RenderConceptGraph converts a ConceptSession into Cytoscape elements.
// Matches Python's cy_render.render_concept_graph:
//   - Only sort nodes (Domain.Nodes) become graph nodes
//   - Binary relations (Domain.Edges) become graph edges between sort nodes
//   - Unary relations (Domain.NodeLabels) become label text inside sort nodes
//
// If checks is nil all edges are shown.
func RenderConceptGraph(cs *ConceptSession, checks *DisplayCheckboxes) *WebUICyElements {
	g := goivy.NewCyElements()
	if cs == nil || cs.Domain == nil {
		return g
	}

	// Build a map of sort name -> color index for per-sort coloring
	// (matches Python tk_graph_ui.py choose_colors).
	sortColorMap := make(map[string]string)
	for i, sortName := range cs.Domain.Nodes {
		sortColorMap[sortName] = goivy.SortColors[(i+1)%len(goivy.SortColors)]
	}

	// Build node label lines: for each sort node, collect applicable unary relations.
	nodeLabelLines := make(map[string][]string)
	if len(cs.AbstractValue) > 0 {
		for _, labelName := range cs.Domain.NodeLabels {
			c := cs.Domain.Concepts[labelName]
			if c == nil {
				continue
			}
			for _, sortName := range cs.Domain.Nodes {
				sortConcept := cs.Domain.Concepts[sortName]
				if sortConcept == nil {
					continue
				}
				if len(c.Sorts) > 0 && len(sortConcept.Sorts) > 0 && c.Sorts[0] == sortConcept.Sorts[0] {
					nKey := fmt.Sprintf("node_label|node_necessarily|%s|%s", sortName, labelName)
					nnKey := fmt.Sprintf("node_label|node_necessarily_not|%s|%s", sortName, labelName)
					if cs.AbstractValue[nKey] {
						nodeLabelLines[sortName] = append(nodeLabelLines[sortName], labelName)
					} else if cs.AbstractValue[nnKey] {
						nodeLabelLines[sortName] = append(nodeLabelLines[sortName], "~"+labelName)
					}
				}
			}
		}
	}

	// Add sort nodes.
	for _, sortName := range cs.Domain.Nodes {
		c := cs.Domain.Concepts[sortName]
		if c == nil {
			continue
		}
		cls := conceptNodeClass(cs, sortName)
		labelParts := []string{sortName}
		labelParts = append(labelParts, nodeLabelLines[sortName]...)
		label := strings.Join(labelParts, "\n")
		shortInfo := sortName
		longInfo := c.Formula
		g.AddNodeWithColor(sortName, label, []string{cls}, shortInfo, longInfo, nil, "octagon", sortColorMap[sortName])
	}

	// Add binary relations as edges between sort nodes.
	for _, edgeName := range cs.Domain.Edges {
		c := cs.Domain.Concepts[edgeName]
		if c == nil || len(c.Sorts) < 2 {
			continue
		}
		sourceSortName := ""
		targetSortName := ""
		for _, sn := range cs.Domain.Nodes {
			sc := cs.Domain.Concepts[sn]
			if sc == nil || len(sc.Sorts) == 0 {
				continue
			}
			if sourceSortName == "" && sc.Sorts[0] == c.Sorts[0] {
				sourceSortName = sn
			}
			if targetSortName == "" && sc.Sorts[0] == c.Sorts[1] {
				targetSortName = sn
			}
		}
		if sourceSortName == "" || targetSortName == "" {
			continue
		}
		if _, ok := g.NodeID[sourceSortName]; !ok {
			continue
		}
		if _, ok := g.NodeID[targetSortName]; !ok {
			continue
		}
		edgeCls := "edge_unknown"
		if cs.AbstractValue != nil {
			noneKey := fmt.Sprintf("edge_info|none_to_none|%s|%s|%s", edgeName, sourceSortName, targetSortName)
			allKey := fmt.Sprintf("edge_info|all_to_all|%s|%s|%s", edgeName, sourceSortName, targetSortName)
			if cs.AbstractValue[noneKey] {
				edgeCls = "none_to_none"
			} else if cs.AbstractValue[allKey] {
				edgeCls = "all_to_all"
			}
		}
		if checks != nil && !checks.EdgeVisible(edgeName, edgeCls) {
			continue
		}
		shortInfo := fmt.Sprintf("%s(%s, %s)", edgeName, sourceSortName, targetSortName)
		g.AddEdge(edgeName, sourceSortName, targetSortName, edgeName, []string{edgeCls}, shortInfo, "")
	}

	// Also add any explicit combiners.
	for _, comb := range cs.Domain.Combiners {
		if comb.Source != "" && comb.Target != "" {
			key := comb.Label + "|" + comb.Source + "|" + comb.Target
			if _, exists := g.EdgeID[key]; exists {
				continue
			}
			if _, ok := g.NodeID[comb.Source]; !ok {
				continue
			}
			if _, ok := g.NodeID[comb.Target]; !ok {
				continue
			}
			edgeCls := conceptEdgeClass(cs, comb)
			if checks != nil && !checks.EdgeVisible(comb.Label, edgeCls) {
				continue
			}
			shortInfo := fmt.Sprintf("%s(%s, %s)", comb.Label, comb.Source, comb.Target)
			g.AddEdge(comb.Label, comb.Source, comb.Target, comb.Label, []string{edgeCls}, shortInfo, "")
		}
	}
	return g
}

// conceptNodeClass determines the CSS class for a concept node.
func conceptNodeClass(cs *ConceptSession, name string) string {
	av := cs.AbstractValue
	if av == nil {
		return "node_unknown"
	}
	noneKey := "node_info|none|" + name
	atLeastKey := "node_info|at_least_one|" + name
	atMostKey := "node_info|at_most_one|" + name
	if av[noneKey] {
		return "non_existing"
	}
	if av[atLeastKey] && av[atMostKey] {
		return "exactly_one"
	}
	if av[atLeastKey] {
		return "at_least_one"
	}
	if av[atMostKey] {
		return "at_most_one"
	}
	return "node_unknown"
}

// conceptEdgeClass determines the CSS class for a concept edge.
func conceptEdgeClass(cs *ConceptSession, comb *ConceptCombiner) string {
	av := cs.AbstractValue
	if av == nil {
		return "edge_unknown"
	}
	key := func(kind string) string {
		return fmt.Sprintf("edge_info|%s|%s|%s|%s", kind, comb.Label, comb.Source, comb.Target)
	}
	if av[key("none_to_none")] {
		return "none_to_none"
	}
	if av[key("all_to_all")] {
		return "all_to_all"
	}
	return "edge_unknown"
}

// conceptShape returns the Cytoscape shape for a concept.
func conceptShape(name string) string {
	return "octagon"
}
