package webui

import (
	"fmt"
	"strings"
)

// NodeAction describes a context-menu action available on a graph node.
type NodeAction struct {
	Label string `json:"label"`
	// Action is the identifier sent back to the server when chosen.
	Action string `json:"action"`
}

// CyElements is a collection of Cytoscape.js graph elements (nodes + edges)
// ready to be serialized to JSON and consumed by the browser.
type CyElements struct {
	Elements []CyElement       `json:"elements"`
	NodeID   map[string]string `json:"-"` // obj key -> cy node id
	EdgeID   map[string]string `json:"-"` // obj key -> cy edge id
}

// CyElement is a single Cytoscape.js element (node or edge).
type CyElement struct {
	Group   string                 `json:"group"`
	Data    map[string]interface{} `json:"data"`
	Classes string                 `json:"classes,omitempty"`
	Locked  bool                   `json:"locked,omitempty"`
}

// NewCyElements creates an empty CyElements container.
func NewCyElements() *CyElements {
	return &CyElements{
		NodeID: make(map[string]string),
		EdgeID: make(map[string]string),
	}
}

// AddNode adds a node element.
func (g *CyElements) AddNode(obj, label string, classes []string, shortInfo, longInfo string, actions []NodeAction, shape string) {
	nid := fmt.Sprintf("n%d", len(g.NodeID))
	g.NodeID[obj] = nid
	data := map[string]interface{}{
		"id":         nid,
		"obj":        obj,
		"label":      label,
		"short_info": shortInfo,
		"long_info":  longInfo,
		"shape":      shape,
	}
	if len(actions) > 0 {
		data["actions"] = actions
	}

	// Compute width heuristic: 10px per character, minimum 50.
	maxLine := 0
	for _, line := range strings.Split(label, "\n") {
		if len(line) > maxLine {
			maxLine = len(line)
		}
	}
	w := maxLine * 10
	if w < 50 {
		w = 50
	}
	h := 50
	data["width"] = w
	data["height"] = h

	g.Elements = append(g.Elements, CyElement{
		Group:   "nodes",
		Data:    data,
		Classes: strings.Join(classes, " "),
	})
}

// AddNodeWithColor adds a node element with an explicit border color.
// Matches Python's per-sort coloring (tk_graph_ui.py choose_colors).
func (g *CyElements) AddNodeWithColor(obj, label string, classes []string, shortInfo, longInfo string, actions []NodeAction, shape, borderColor string) {
	nid := fmt.Sprintf("n%d", len(g.NodeID))
	g.NodeID[obj] = nid
	data := map[string]interface{}{
		"id":         nid,
		"obj":        obj,
		"label":      label,
		"short_info": shortInfo,
		"long_info":  longInfo,
		"shape":      shape,
	}
	if len(actions) > 0 {
		data["actions"] = actions
	}
	if borderColor != "" {
		data["border_color"] = borderColor
	}

	// Compute width heuristic: 10px per character, minimum 50.
	maxLine := 0
	for _, line := range strings.Split(label, "\n") {
		if len(line) > maxLine {
			maxLine = len(line)
		}
	}
	w := maxLine * 10
	if w < 50 {
		w = 50
	}
	// Height grows with number of label lines
	lines := strings.Count(label, "\n") + 1
	h := 30 + lines*20
	if h < 50 {
		h = 50
	}
	data["width"] = w
	data["height"] = h

	g.Elements = append(g.Elements, CyElement{
		Group:   "nodes",
		Data:    data,
		Classes: strings.Join(classes, " "),
	})
}

// AddEdge adds an edge element.  sourceObj and targetObj must have been
// added as nodes already.
func (g *CyElements) AddEdge(obj, sourceObj, targetObj, label string, classes []string, shortInfo, longInfo string) {
	eid := fmt.Sprintf("e%d", len(g.EdgeID))
	key := obj + "|" + sourceObj + "|" + targetObj
	g.EdgeID[key] = eid
	srcID := g.NodeID[sourceObj]
	tgtID := g.NodeID[targetObj]
	g.Elements = append(g.Elements, CyElement{
		Group: "edges",
		Data: map[string]interface{}{
			"id":         eid,
			"source":     srcID,
			"target":     tgtID,
			"obj":        obj,
			"source_obj": sourceObj,
			"target_obj": targetObj,
			"label":      label,
			"short_info": shortInfo,
			"long_info":  longInfo,
		},
		Classes: strings.Join(classes, " "),
	})
}

// AnalysisGraphState holds the ARG data needed for rendering.
// It wraps the art.AnalysisGraph indirectly so we can test without
// pulling in the full art/module/z3 dependency chain.
type AnalysisGraphState struct {
	States      []ARGNode
	Transitions []ARGTransition
	Covering    []ARGCover
}

// ARGNode is a lightweight representation of an ARG state for rendering.
type ARGNode struct {
	ID       int
	Label    string
	IsBottom bool
	Info     string
}

// ARGTransition is a lightweight transition for rendering.
type ARGTransition struct {
	SourceID int
	TargetID int
	Label    string
	IsJoin   bool
}

// ARGCover is a lightweight cover relation for rendering.
type ARGCover struct {
	CoveredID  int
	CoveringID int
}

// NewAnalysisGraphState creates an empty ARG state.
func NewAnalysisGraphState() *AnalysisGraphState {
	return &AnalysisGraphState{}
}

// RenderARG converts an AnalysisGraphState into Cytoscape elements.
func RenderARG(ag *AnalysisGraphState) *CyElements {
	g := NewCyElements()
	if ag == nil {
		return g
	}

	// Add state nodes.
	for _, st := range ag.States {
		classes := []string{"state"}
		if st.IsBottom {
			classes = []string{"bottom_state"}
		}
		g.AddNode(
			fmt.Sprintf("state_%d", st.ID),
			st.Label,
			classes,
			st.Info,
			st.Info,
			nil,
			"ellipse",
		)
	}

	// Add transition edges.
	for _, tr := range ag.Transitions {
		cls := "transition_action"
		if tr.IsJoin {
			cls = "transition_join"
		}
		g.AddEdge(
			fmt.Sprintf("tr_%d_%d", tr.SourceID, tr.TargetID),
			fmt.Sprintf("state_%d", tr.SourceID),
			fmt.Sprintf("state_%d", tr.TargetID),
			tr.Label,
			[]string{cls},
			tr.Label,
			tr.Label,
		)
	}

	// Add covering edges.
	for _, cv := range ag.Covering {
		g.AddEdge(
			fmt.Sprintf("cover_%d_%d", cv.CoveredID, cv.CoveringID),
			fmt.Sprintf("state_%d", cv.CoveredID),
			fmt.Sprintf("state_%d", cv.CoveringID),
			"",
			[]string{"cover"},
			"",
			"",
		)
	}
	return g
}

// ProofGoal is a lightweight proof goal for rendering.
type ProofGoal struct {
	ID       int
	Label    string
	Refuted  bool
	Info     string
	ParentID int // -1 if no parent
}

// ProofStack holds goals for rendering.
type ProofStack struct {
	Goals []ProofGoal
}

// RenderProofStack converts a ProofStack into Cytoscape elements.
func RenderProofStack(stack *ProofStack) *CyElements {
	g := NewCyElements()
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

// sortColors matches Python's tk_graph_ui.py line_colors palette.
// Nodes are colored by sort index to visually distinguish types.
var sortColors = []string{
	"#000000", // black
	"#0000ff", // blue
	"#ff0000", // red
	"#008000", // green
	"#8b2252", // VioletRed4
	"#ff4040", // brown1
	"#528b8b", // DarkSlateGray4
	"#000080", // navy
	"#8b0a50", // DeepPink4
	"#556b2f", // DarkOliveGreen4
	"#551a8b", // purple4
	"#8b6969", // RosyBrown4
	"#4a708b", // SkyBlue4
	"#8b3626", // tomato4
	"#8fbc8f", // DarkSeaGreen4
	"#b23aee", // DarkOrchid2
	"#cd6600", // DarkOrange3
	"#00688b", // DeepSkyBlue4
	"#ff6a6a", // IndianRed1
	"#8b1a1a", // maroon4
}

// RenderConceptGraph converts a ConceptSession into Cytoscape elements.
// Matches Python's cy_render.render_concept_graph:
//   - Only sort nodes (Domain.Nodes) become graph nodes
//   - Binary relations (Domain.Edges) become graph edges between sort nodes
//   - Unary relations (Domain.NodeLabels) become label text inside sort nodes
// If checks is nil all edges are shown.
func RenderConceptGraph(cs *ConceptSession, checks *DisplayCheckboxes) *CyElements {
	g := NewCyElements()
	if cs == nil || cs.Domain == nil {
		return g
	}

	// Build a map of sort name -> color index for per-sort coloring
	// (matches Python tk_graph_ui.py choose_colors)
	sortColorMap := make(map[string]string)
	for i, sortName := range cs.Domain.Nodes {
		sortColorMap[sortName] = sortColors[i%len(sortColors)]
	}

	// Build node label lines: for each sort node, collect applicable unary relations.
	// Python only renders these when abstract_value has been computed (via Z3)
	// and the label's value (necessarily/maybe/necessarily_not) is checked in
	// node_label_display_checkboxes. Without an abstract value, no labels shown.
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
					// Check abstract value for this label on this node
					nKey := fmt.Sprintf("node_label|node_necessarily|%s|%s", sortName, labelName)
					nnKey := fmt.Sprintf("node_label|node_necessarily_not|%s|%s", sortName, labelName)
					if cs.AbstractValue[nKey] {
						nodeLabelLines[sortName] = append(nodeLabelLines[sortName], labelName)
					} else if cs.AbstractValue[nnKey] {
						nodeLabelLines[sortName] = append(nodeLabelLines[sortName], "~"+labelName)
					}
					// "maybe" labels are only shown if checkbox says so (skip for now)
				}
			}
		}
	}

	// Add sort nodes only (not edges/labels as separate nodes).
	// Python: only concepts['nodes'] become visible graph nodes.
	for _, sortName := range cs.Domain.Nodes {
		c := cs.Domain.Concepts[sortName]
		if c == nil {
			continue
		}
		cls := conceptNodeClass(cs, sortName)
		// Build label: sort name + any node label lines (only when abstract value exists)
		labelParts := []string{sortName}
		labelParts = append(labelParts, nodeLabelLines[sortName]...)
		label := strings.Join(labelParts, "\n")
		shortInfo := sortName
		longInfo := c.Formula
		// Python ivy_graph.py get_shape always returns 'octagon'
		g.AddNodeWithColor(sortName, label, []string{cls}, shortInfo, longInfo, nil, "octagon", sortColorMap[sortName])
	}

	// Add binary relations as edges between sort nodes.
	// Python: concepts['edges'] become graph edges connecting sort nodes.
	for _, edgeName := range cs.Domain.Edges {
		c := cs.Domain.Concepts[edgeName]
		if c == nil || len(c.Sorts) < 2 {
			continue
		}
		// Find source and target sort nodes
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
		// Check if both source and target nodes were added
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

	// Also add any explicit combiners (for backward compat / interactive sessions).
	for _, comb := range cs.Domain.Combiners {
		if comb.Source != "" && comb.Target != "" {
			// Skip if already added as an edge above
			key := comb.Label + "|" + comb.Source + "|" + comb.Target
			if _, exists := g.EdgeID[key]; exists {
				continue
			}
			// Skip if source/target nodes don't exist in graph
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

// conceptNodeClass determines the CSS class for a concept node based on
// the abstract value.
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
// Python ivy_graph.py get_shape() always returns 'octagon'.
// The cy_render.py version with __ID/ellipse is for a different context (leader_demo).
func conceptShape(name string) string {
	return "octagon"
}
