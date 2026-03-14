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
		Locked:  true,
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

// RenderConceptGraph converts a ConceptSession into Cytoscape elements.
// If checks is nil all edges are shown.
func RenderConceptGraph(cs *ConceptSession, checks *DisplayCheckboxes) *CyElements {
	g := NewCyElements()
	if cs == nil || cs.Domain == nil {
		return g
	}

	for name, c := range cs.Domain.Concepts {
		cls := conceptNodeClass(cs, name)
		label := name
		shape := conceptShape(name)
		g.AddNode(name, label, []string{cls}, name, c.Formula, nil, shape)
	}

	for _, comb := range cs.Domain.Combiners {
		if comb.Source != "" && comb.Target != "" {
			edgeCls := conceptEdgeClass(cs, comb)
			if checks != nil && !checks.EdgeVisible(comb.Label, edgeCls) {
				continue
			}
			info := fmt.Sprintf("%s(%s, %s)", comb.Label, comb.Source, comb.Target)
			g.AddEdge(comb.Label, comb.Source, comb.Target, comb.Label, []string{edgeCls}, info, info)
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
func conceptShape(name string) string {
	prefix := strings.SplitN(name, "!", 2)[0]
	switch prefix {
	case "__ID":
		return "octagon"
	default:
		return "ellipse"
	}
}
