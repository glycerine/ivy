package webui

// CyStyleEntry is a single Cytoscape.js stylesheet rule.
type CyStyleEntry struct {
	Selector string            `json:"selector"`
	Style    map[string]string `json:"style"`
}

// ConceptStyle returns the Cytoscape stylesheet for concept graphs.
// This is a direct port of cy_styles.py concept_style.
func ConceptStyle() []CyStyleEntry {
	return []CyStyleEntry{
		// base node
		{
			Selector: "node",
			Style: map[string]string{
				"content":            "data(label)",
				"text-wrap":          "wrap",
				"font-size":          "14px",
				"text-outline-width": "3px",
				"text-outline-color": "#888",
				"text-valign":        "center",
				"color":              "#fff",
				"width":              "data(width)",
				"height":             "data(height)",
				"border-color":       "#000",
				"shape":              "data(shape)",
				"background-color":   "#888",
			},
		},
		// node classes
		{
			Selector: "node.non_existing",
			Style:    map[string]string{"display": "none"},
		},
		{
			Selector: "node.subgraph_box",
			Style: map[string]string{
				"content":            "data(label)",
				"shape":              "roundrectangle",
				"background-opacity": "0",
				"border-width":       "2px",
				"border-style":       "dashed",
				"border-color":       "#777",
				"padding":            "18px",
				"events":             "no",
				"z-index":            "0",
			},
		},
		{
			Selector: "node.exactly_one",
			Style:    map[string]string{"border-width": "4px", "border-style": "solid"},
		},
		{
			Selector: "node.at_least_one",
			Style:    map[string]string{"border-width": "8px", "border-style": "double"},
		},
		{
			Selector: "node.at_most_one",
			Style:    map[string]string{"border-width": "3px", "border-style": "dotted"},
		},
		{
			Selector: "node.node_unknown",
			Style:    map[string]string{"border-width": "0px"},
		},
		// base edge
		{
			Selector: "edge",
			Style: map[string]string{
				"target-arrow-shape": "triangle",
				"target-arrow-fill":  "filled",
				"source-arrow-fill":  "filled",
				"text-wrap":          "wrap",
			},
		},
		{
			Selector: "edge[label]",
			Style:    map[string]string{"content": "data(label)"},
		},
		{
			Selector: "edge[text_max_width]",
			Style:    map[string]string{"text-max-width": "data(text_max_width)"},
		},
		// edge classes
		{
			Selector: "edge.none_to_none",
			Style: map[string]string{
				"width":              "4px",
				"line-style":         "dashed",
				"target-arrow-shape": "triangle",
				"target-arrow-fill":  "filled",
				"source-arrow-fill":  "filled",
			},
		},
		{
			Selector: "edge.all_to_all",
			Style:    map[string]string{"width": "4px", "line-style": "solid"},
		},
		{
			Selector: "edge.edge_unknown",
			Style:    map[string]string{"width": "4px", "line-style": "dotted"},
		},
		{
			Selector: "edge.layout_only",
			Style:    map[string]string{"display": "none"},
		},
		{
			Selector: "edge.total",
			Style:    map[string]string{"source-arrow-shape": "circle", "source-arrow-fill": "filled"},
		},
		{
			Selector: "edge.functional",
			Style:    map[string]string{"source-arrow-shape": "square"},
		},
		{
			Selector: "edge.injective",
			Style:    map[string]string{"target-arrow-shape": "triangle-backcurve"},
		},
		{
			Selector: "edge.surjective",
			Style:    map[string]string{"target-arrow-fill": "filled"},
		},
		// selection
		{
			Selector: "node:selected",
			Style:    map[string]string{"overlay-opacity": "0.2"},
		},
		{
			Selector: "edge:selected",
			Style:    map[string]string{"overlay-opacity": "0.2"},
		},
	}
}

// ARGStyle returns the Cytoscape stylesheet for the Abstract Reachability
// Graph. This is a direct port of cy_styles.py arg_style.
func ARGStyle() []CyStyleEntry {
	return []CyStyleEntry{
		// base node
		{
			Selector: "node",
			Style: map[string]string{
				"content":            "data(label)",
				"text-wrap":          "wrap",
				"text-outline-width": "3px",
				"text-outline-color": "#888",
				"text-valign":        "center",
				"color":              "#fff",
				"width":              "50",
				"height":             "50",
				"border-color":       "#000",
				"background-color":   "#888",
			},
		},
		// node classes
		{Selector: "node.state", Style: map[string]string{}},
		{
			Selector: "node.bottom_state",
			Style:    map[string]string{"background-color": "#000", "text-outline-color": "#000"},
		},
		{
			Selector: "node.safe_state",
			Style:    map[string]string{"background-color": "#2f8f46", "text-outline-color": "#1f6b33"},
		},
		{
			Selector: "node.marked_state",
			Style:    map[string]string{"background-color": "#b73535", "text-outline-color": "#7f1f1f"},
		},
		// base edge
		{
			Selector: "edge",
			Style: map[string]string{
				"width":              "4px",
				"line-style":         "solid",
				"edge-text-rotation": "none",
				"text-wrap":          "wrap",
			},
		},
		{
			Selector: "edge[label]",
			Style:    map[string]string{"content": "data(label)"},
		},
		{
			Selector: "edge[text_max_width]",
			Style:    map[string]string{"text-max-width": "data(text_max_width)"},
		},
		{
			Selector: "edge.transition_join",
			Style:    map[string]string{"target-arrow-shape": "triangle-backcurve"},
		},
		{
			Selector: "edge.transition_action",
			Style:    map[string]string{"target-arrow-shape": "triangle"},
		},
		{
			Selector: "edge.layout_only",
			Style:    map[string]string{"display": "none"},
		},
		{
			Selector: "edge.cover",
			Style: map[string]string{
				"content":            "",
				"target-arrow-shape": "triangle",
				"line-style":         "dashed",
			},
		},
		// selection
		{Selector: "node:selected", Style: map[string]string{"overlay-opacity": "0.2"}},
		{Selector: "edge:selected", Style: map[string]string{"overlay-opacity": "0.2"}},
	}
}

// ProofStyle returns the Cytoscape stylesheet for the proof stack.
// This is a direct port of cy_styles.py proof_style.
func ProofStyle() []CyStyleEntry {
	return []CyStyleEntry{
		// base node
		{
			Selector: "node",
			Style: map[string]string{
				"content":            "data(label)",
				"text-wrap":          "wrap",
				"text-outline-width": "3px",
				"text-outline-color": "#888",
				"text-valign":        "center",
				"color":              "#fff",
				"width":              "50",
				"height":             "50",
				"border-color":       "#000",
				"background-color":   "#888",
			},
		},
		// base edge
		{
			Selector: "edge",
			Style:    map[string]string{"target-arrow-shape": "triangle"},
		},
		// node classes
		{
			Selector: "node.refuted",
			Style:    map[string]string{"background-color": "#000", "text-outline-color": "#000"},
		},
		// selection
		{Selector: "node:selected", Style: map[string]string{"overlay-opacity": "0.2"}},
		{Selector: "edge:selected", Style: map[string]string{"overlay-opacity": "0.2"}},
	}
}
