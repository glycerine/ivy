package webui

// This file supplements graph_model.go with convenience methods.
// The main DisplayCheckboxes type is defined in graph_model.go.

// SetEdge is a convenience wrapper for SetEdgeCheckbox.
func (d *DisplayCheckboxes) SetEdge(relName, cls string, visible bool) {
	d.SetEdgeCheckbox(relName, cls, visible)
}

// SetNodeLabel is a convenience wrapper for SetNodeLabelCheckbox.
func (d *DisplayCheckboxes) SetNodeLabel(labelName, displayClass string, visible bool) {
	d.SetNodeLabelCheckbox(labelName, displayClass, visible)
}
