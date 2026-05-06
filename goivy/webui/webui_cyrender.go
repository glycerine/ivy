package webui

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"sort"
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

func RenderAnalysisUIARG(ui *AnalysisGraphUI) *WebUICyElements {
	if ui == nil || ui.AG == nil {
		return RenderWebUIARG(nil)
	}
	cy := RenderWebUIARG(ArtToGraphState(ui.AG))
	for i := range cy.Elements {
		ele := &cy.Elements[i]
		if ele.Group != "nodes" {
			continue
		}
		obj, _ := ele.Data["obj"].(string)
		var stateID int
		if _, err := fmt.Sscanf(obj, "state_%d", &stateID); err != nil {
			continue
		}
		entries := ui.GetNodeActions(stateID, "right")
		actions := make([]goivy.NodeAction, 0, len(entries))
		for _, entry := range entries {
			actions = append(actions, goivy.NodeAction{
				Label:  entry.Label,
				Action: entry.Action,
				Args:   entry.Args,
			})
		}
		if len(actions) > 0 {
			ele.Data["actions"] = actions
		}
	}
	return cy
}

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

	nodes := orderedConceptNodes(cs, checks)

	// Build a map of sort name -> color index for per-sort coloring
	// (matches Python tk_graph_ui.py choose_colors).
	sortColorMap := make(map[string]string)
	for i, sortName := range nodes {
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
			for _, sortName := range nodes {
				sortConcept := cs.Domain.Concepts[sortName]
				if sortConcept == nil {
					continue
				}
				if len(c.Sorts) > 0 && len(sortConcept.Sorts) > 0 && c.Sorts[0] == sortConcept.Sorts[0] {
					nKey := fmt.Sprintf("node_label|node_necessarily|%s|%s", sortName, labelName)
					nnKey := fmt.Sprintf("node_label|node_necessarily_not|%s|%s", sortName, labelName)
					if cs.AbstractValue[nKey] {
						if checks != nil && !checks.NodeLabelVisible(labelName, NodeLabelNecessarily) {
							continue
						}
						nodeLabelLines[sortName] = append(nodeLabelLines[sortName], labelName)
					} else if cs.AbstractValue[nnKey] {
						if checks != nil && !checks.NodeLabelVisible(labelName, NodeLabelNecessarilyNot) {
							continue
						}
						nodeLabelLines[sortName] = append(nodeLabelLines[sortName], "~"+labelName)
					} else if checks != nil && !checks.NodeLabelVisible(labelName, NodeLabelMaybe) {
						continue
					}
				}
			}
		}
	}

	edgeTuples := renderConceptGraphEdgeTuples(cs, nodes)
	hiddenByTransitive := make(map[[3]string]bool)
	if checks != nil && cs.AbstractValue != nil {
		hiddenByTransitive = GetTransitiveReduction(checks, cs.AbstractValue, edgeTuples)
	}

	// Add sort nodes.
	for _, sortName := range nodes {
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
		g.Elements[len(g.Elements)-1].Data["cluster"] = conceptCluster(c, sortName)
	}

	// Add binary relations as edges between sort nodes.
	for _, edgeName := range cs.Domain.Edges {
		c := cs.Domain.Concepts[edgeName]
		if c == nil || len(c.Sorts) < 2 {
			continue
		}
		sourceSortName, targetSortName := conceptEdgeEndpoints(cs, nodes, c)
		if sourceSortName == "" || targetSortName == "" {
			continue
		}
		if hiddenByTransitive[[3]string{edgeName, sourceSortName, targetSortName}] {
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
			if hiddenByTransitive[[3]string{comb.Label, comb.Source, comb.Target}] {
				continue
			}
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

func orderedConceptNodes(cs *ConceptSession, checks *DisplayCheckboxes) []string {
	nodes := append([]string{}, cs.Domain.Nodes...)
	sort.Strings(nodes)
	if checks == nil || cs.AbstractValue == nil {
		return nodes
	}
	var order [][2]string
	for _, edge := range renderConceptGraphEdgeTuples(cs, nodes) {
		edgeName, source, target := edge[0], edge[1], edge[2]
		if !checks.EdgeVisible(edgeName, EdgeDisplayAllToAll) || !checks.EdgeVisible(edgeName, EdgeDisplayTransitive) {
			continue
		}
		key := fmt.Sprintf("edge_info|all_to_all|%s|%s|%s", edgeName, source, target)
		if cs.AbstractValue[key] {
			order = append(order, [2]string{source, target})
		}
	}
	if len(order) == 0 {
		return nodes
	}
	return goivy.TopologicalSort(nodes, order, func(s string) string { return s })
}

func renderConceptGraphEdgeTuples(cs *ConceptSession, nodes []string) [][3]string {
	var tuples [][3]string
	seen := make(map[[3]string]bool)
	add := func(edge, source, target string) {
		tuple := [3]string{edge, source, target}
		if edge == "" || source == "" || target == "" || seen[tuple] {
			return
		}
		seen[tuple] = true
		tuples = append(tuples, tuple)
	}
	for _, edgeName := range cs.Domain.Edges {
		c := cs.Domain.Concepts[edgeName]
		if c == nil || len(c.Sorts) < 2 {
			continue
		}
		source, target := conceptEdgeEndpoints(cs, nodes, c)
		add(edgeName, source, target)
	}
	for _, comb := range cs.Domain.Combiners {
		if comb == nil {
			continue
		}
		add(comb.Label, comb.Source, comb.Target)
	}
	sort.Slice(tuples, func(i, j int) bool {
		if tuples[i][0] != tuples[j][0] {
			return tuples[i][0] < tuples[j][0]
		}
		if tuples[i][1] != tuples[j][1] {
			return tuples[i][1] < tuples[j][1]
		}
		return tuples[i][2] < tuples[j][2]
	})
	return tuples
}

func conceptEdgeEndpoints(cs *ConceptSession, nodes []string, c *Concept) (string, string) {
	sourceSortName := ""
	targetSortName := ""
	for _, sn := range nodes {
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
	return sourceSortName, targetSortName
}

func conceptCluster(c *Concept, fallback string) string {
	if c != nil && len(c.Sorts) > 0 && c.Sorts[0] != "" {
		return c.Sorts[0]
	}
	return fallback
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
