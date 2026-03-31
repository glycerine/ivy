// cyrender.go defines Cytoscape.js rendering types for the analysis graph.
// These types were originally in webui/cyrender.go but live here because
// art is the fundamental package that owns the AnalysisGraph and its
// rendering. The webui package imports these from art.
//
// Corresponds to Python's cy_elements.py (CyElements) and the ARG
// rendering portion of ivy_art.py (render_rg).
package art

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/clauseops"
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
func (g *CyElements) AddNode(obj, label string, classes []string, shortInfo string, longInfo interface{}, actions []NodeAction, shape string) {
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
func (g *CyElements) AddNodeWithColor(obj, label string, classes []string, shortInfo string, longInfo interface{}, actions []NodeAction, shape, borderColor string) {
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

// AddEdge adds an edge element. sourceObj and targetObj must have been
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

// -----------------------------------------------------------------------
// ARG rendering types
// -----------------------------------------------------------------------

// AnalysisGraphState holds the ARG data needed for rendering.
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

// SortColors matches Python's tk_graph_ui.py line_colors palette.
// Nodes are colored by sort index to visually distinguish types.
var SortColors = []string{
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

// -----------------------------------------------------------------------
// ConceptGraphView — interface for Python's sg.current
// -----------------------------------------------------------------------

// ConceptGraphView represents the concept graph object returned by
// the standard_graph callback. In Python, this is the GraphUI object
// whose .current attribute has set_state() and set_concrete() methods.
// The Go interface names the methods that art.ConceptGraph calls.
type ConceptGraphView interface {
	// SetGraphState sets the clause state on the concept graph.
	// Python: sg.current.set_state(and_clauses(clauses, bg))
	SetGraphState(clauses *clauseops.Clauses)
	// SetGraphConcrete clears/sets the concrete state.
	// Python: sg.current.set_concrete([])
	SetGraphConcrete(concrete []interface{})
}
