package webui

// DisplayCheckboxes tracks which edges and node labels the user has
// chosen to show or hide in the concept graph display.
type DisplayCheckboxes struct {
	// EdgeDisplay maps relation_name -> class -> visible.
	// Classes are "+", "?", "-", etc. corresponding to
	// "all_to_all", "edge_unknown", "none_to_none".
	EdgeDisplay map[string]map[string]bool `json:"edge_display"`

	// NodeLabelDisplay maps label_name -> display_class -> visible.
	// Display classes are "node_necessarily", "node_necessarily_not", "node_maybe".
	NodeLabelDisplay map[string]map[string]bool `json:"node_label_display"`
}

// NewDisplayCheckboxes creates a DisplayCheckboxes with default all-visible state.
func NewDisplayCheckboxes() *DisplayCheckboxes {
	return &DisplayCheckboxes{
		EdgeDisplay:      make(map[string]map[string]bool),
		NodeLabelDisplay: make(map[string]map[string]bool),
	}
}

// edgeClassFromSymbol maps the display symbols to CSS class names.
var edgeClassFromSymbol = map[string]string{
	"+":  "all_to_all",
	"?":  "edge_unknown",
	"-":  "none_to_none",
	"\u2264": "total", // ≤
}

// EdgeVisible reports whether the edge with the given relation name and
// CSS class should be displayed.
func (d *DisplayCheckboxes) EdgeVisible(relName, cssClass string) bool {
	if d == nil {
		return true
	}
	relMap, ok := d.EdgeDisplay[relName]
	if !ok {
		return true // default visible
	}
	// Try to find by CSS class name directly.
	if v, ok := relMap[cssClass]; ok {
		return v
	}
	return true
}

// NodeLabelVisible reports whether a node label with the given display
// class should be shown.
func (d *DisplayCheckboxes) NodeLabelVisible(labelName, displayClass string) bool {
	if d == nil {
		return true
	}
	lm, ok := d.NodeLabelDisplay[labelName]
	if !ok {
		return true
	}
	if v, ok := lm[displayClass]; ok {
		return v
	}
	return true
}

// SetEdge sets the visibility for a relation + class.
func (d *DisplayCheckboxes) SetEdge(relName, cls string, visible bool) {
	if d.EdgeDisplay[relName] == nil {
		d.EdgeDisplay[relName] = make(map[string]bool)
	}
	d.EdgeDisplay[relName][cls] = visible
}

// SetNodeLabel sets the visibility for a label + display class.
func (d *DisplayCheckboxes) SetNodeLabel(labelName, displayClass string, visible bool) {
	if d.NodeLabelDisplay[labelName] == nil {
		d.NodeLabelDisplay[labelName] = make(map[string]bool)
	}
	d.NodeLabelDisplay[labelName][displayClass] = visible
}
