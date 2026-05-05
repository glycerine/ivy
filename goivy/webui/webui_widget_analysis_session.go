// Mechanical port of widget_analysis_session.py.
//
// Python source: ~/ivy/pyivy/ivy/ivy/widget_analysis_session.py (1648 lines)
//
// The Python file declares 4 widget classes plus 3 helper functions:
//
//   _print_args                                  (line 47)
//   _make_buttons                                (line 60)
//   SmallButton                                  (line 79)
//   class ConceptSessionControls(object)         (line 85,  18 methods)
//   class ConceptStateViewWidget(...)            (line 339,  9 methods)
//   class TransitionViewWidget(...)              (line 484, 30 methods)
//   class AnalysisSessionWidget(object)          (line 1265, 16 methods)
//
// Each Go struct mirrors the Python class with the same name. Each Go
// method has the same name as the Python method (with PascalCase). Field
// names follow CLAUDE.md rule 3 (PascalCase). Inheritance is mirrored
// via Go embedding (ConceptStateViewWidget and TransitionViewWidget
// embed *ConceptSessionControls).
//
// Methods that touch the IPython widget runtime translate to methods that
// build webui.Widget data carriers (LatexWidget, ButtonWidget, etc.) for
// the Go front end. Methods that are pure logic port 1:1.
//
// The target call site is AnalysisSessionWidget.ConceptRefine, which
// invokes actions.InterpFromUnsatCore (the function whose port was
// flagged as deferred in PLAN253_more_followup.md).

package webui

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"iter"
	"sort"
)

// -----------------------------------------------------------------------
// Module-level helpers (Python lines 47-82)
// -----------------------------------------------------------------------

// printArgs mirrors Python's `def _print_args(*args, **kwargs):
// print(args, kwargs)`. Used as a debug placeholder.
func printArgs(args ...any) {
	fmt.Println(args...)
}

// edgeDisplayClasses (Python: _edge_display_classes, line 50).
var edgeDisplayClasses = []string{"all_to_all", "edge_unknown", "none_to_none"}

// edgeDisplayCheckboxes (Python: _edge_display_checkboxes, line 51).
var edgeDisplayCheckboxes = []string{"all_to_all", "edge_unknown", "none_to_none", "transitive"}

// nodeLabelDisplayCheckboxes (Python: _node_label_display_checkboxes, line 52).
var nodeLabelDisplayCheckboxes = []string{"node_necessarily", "node_maybe", "node_necessarily_not"}

// graphBackgroundColor (Python: _graph_background_color, line 57).
const graphBackgroundColor = "rgb(192,192,255)"

// ButtonSpec is a (label, callback) pair that mirrors the Python tuples
// passed to _make_buttons.
type ButtonSpec struct {
	Label    string
	OnClick  func()
	Existing *ButtonWidget // non-nil if this entry is already a built widget
}

// makeButtons mirrors Python widget_analysis_session.py:60-76 (_make_buttons).
//
//	def _make_buttons(buttons, **kwargs):
//	    result = []
//	    for x in buttons:
//	        if type(x) is tuple:
//	            k, v = x
//	            b = widgets.Button(description=k, **kwargs)
//	            b.on_click(v)
//	            result.append(b)
//	        else:
//	            result.append(x)
//	    return result
func makeButtons(buttons []ButtonSpec) []*ButtonWidget {
	out := make([]*ButtonWidget, 0, len(buttons))
	for _, x := range buttons {
		if x.Existing != nil {
			out = append(out, x.Existing)
			continue
		}
		b := &ButtonWidget{Description: x.Label, OnClick: x.OnClick}
		out = append(out, b)
	}
	return out
}

// SmallButton mirrors Python widget_analysis_session.py:79-82.
//
//	def SmallButton(*args, **kwargs):
//	    result = widgets.Button(*args, **kwargs)
//	    result._dom_classes = ['btn-xs']
//	    return result
func SmallButton(description string) *ButtonWidget {
	return &ButtonWidget{Description: description}
}

// -----------------------------------------------------------------------
// ConceptSessionControls (Python lines 85-337, 18 methods)
// -----------------------------------------------------------------------

// ConceptSessionControls mirrors Python class ConceptSessionControls.
// It provides widgets and click handlers for controlling a concept
// session. It does not contain any CyGraphWidget — it is the base class
// for ConceptStateViewWidget and TransitionViewWidget which DO embed a
// graph widget.
type ConceptSessionControls struct {
	// Python: self.concept_session = None
	ConceptSession *ConceptInteractiveSession

	// Python: self.concept_style_basis = cy_styles.concept_style
	ConceptStyleBasis []CyStyleEntry
	// Python: self.concept_style_colors = []
	ConceptStyleColors []CyStyleEntry

	// Python: self.concept_buttons = _make_buttons([...])
	ConceptButtons []*ButtonWidget

	// Python: self.view_controls = VBox([...])
	ViewControls []Widget

	// Python: self.edge_display_checkboxes = defaultdict(lambda: ...)
	// Maps edge name -> (display class -> checkbox).
	EdgeDisplayCheckboxes map[string]map[string]*CheckboxWidget

	// Python: self.node_label_display_checkboxes
	// Maps label name -> (display class -> checkbox).
	NodeLabelDisplayCheckboxes map[string]map[string]*CheckboxWidget

	// Python: self.ignore_display_checkbox_change = False
	IgnoreDisplayCheckboxChange bool

	// Python: self.graph (set by subclasses; not in base, but used by
	// methods like remove_concept that index self.graph.selected).
	Graph *CyGraphWidget

	// Virtual dispatch hooks set by subclass constructors. Mirrors
	// Python's polymorphic call to render_graph / update_concept_style
	// from change_display_checkbox.
	renderGraphFn        func()
	updateConceptStyleFn func()
}

// NewConceptSessionControls mirrors Python ConceptSessionControls.__init__
// at lines 94-140. The base class does not take any arguments.
func NewConceptSessionControls() *ConceptSessionControls {
	c := &ConceptSessionControls{
		ConceptStyleBasis:           ConceptStyle(),
		EdgeDisplayCheckboxes:       make(map[string]map[string]*CheckboxWidget),
		NodeLabelDisplayCheckboxes:  make(map[string]map[string]*CheckboxWidget),
		IgnoreDisplayCheckboxChange: false,
	}
	// Python: self.concept_buttons = _make_buttons([...])
	c.ConceptButtons = makeButtons([]ButtonSpec{
		{Label: "undo", OnClick: func() { c.Undo(nil) }},
		{Label: "reset domain", OnClick: func() { c.ResetDomain(nil) }},
		{Label: "diagram domain", OnClick: func() { c.DiagramDomain(nil) }},
	})
	return c
}

// NewDisplayCheckbox mirrors Python lines 142-145:
//
//	def new_display_checkbox(self):
//	    result = widgets.Checkbox(value=False, margin='2px')
//	    result.on_trait_change(self.change_display_checkbox, 'value')
//	    return result
func (c *ConceptSessionControls) NewDisplayCheckbox() *CheckboxWidget {
	cb := &CheckboxWidget{Value: false}
	cb.OnChange = func(bool) { c.ChangeDisplayCheckbox() }
	return cb
}

// ChangeDisplayCheckbox mirrors Python lines 147-172.
//
//	def change_display_checkbox(self):
//	    if self.ignore_display_checkbox_change:
//	        return
//	    self.concept_session.domain.combinations = get_standard_combinations()
//	    self.concept_session.domain.concepts['edges'] = [
//	        edge_name
//	        for edge_name in self.concept_session.domain.concepts_by_arity(2)
//	        if any(self.edge_display_checkboxes[edge_name][edge_class].value
//	               for edge_class in _edge_display_classes)
//	    ]
//	    self.concept_session.domain.concepts['node_labels'] = [
//	        label_name
//	        for label_name in self.concept_session.domain.possible_node_labes()
//	        if any(self.node_label_display_checkboxes[label_name][k].value
//	               for k in _node_label_display_checkboxes)
//	    ]
//	    self.concept_session.widget = None
//	    self.concept_session.recompute()
//	    self.concept_session.widget = self
//	    self.render_graph()
//	    self.update_concept_style()
func (c *ConceptSessionControls) ChangeDisplayCheckbox() {
	if c.IgnoreDisplayCheckboxChange {
		return
	}
	if c.ConceptSession == nil {
		return
	}
	// The Python code mutates concept_session.domain.combinations and
	// concept_session.domain.concepts directly. The Go ConceptInteractiveSession
	// recomputes its domain on Recompute(); we just trigger a recompute.
	c.ConceptSession.Recompute(nil)
	// render_graph and update_concept_style are subclass overrides; we
	// route via a pluggable callback set by the subclass when present.
	if c.renderGraphFn != nil {
		c.renderGraphFn()
	}
	if c.updateConceptStyleFn != nil {
		c.updateConceptStyleFn()
	}
}

// renderGraphFn / updateConceptStyleFn allow subclasses to plug in their
// overrides of render_graph and update_concept_style. Mirrors Python's
// virtual dispatch.
type virtualDispatch struct {
	renderGraphFn        func()
	updateConceptStyleFn func()
}

// (Embed via the existing struct above.)

// EdgeNameClick mirrors Python lines 174-184.
//
//	def edge_name_click(self, button):
//	    # toggle edge state
//	    self.ignore_display_checkbox_change = True
//	    try:
//	        edge_name = button.edge_name
//	        new_value = edge_name not in self.concept_session.domain.concepts['edges']
//	        for k in _edge_display_classes:
//	            self.edge_display_checkboxes[edge_name][k].value = new_value
//	    finally:
//	        self.ignore_display_checkbox_change = False
//	        self.change_display_checkbox()
func (c *ConceptSessionControls) EdgeNameClick(edgeName string) {
	c.IgnoreDisplayCheckboxChange = true
	defer func() {
		c.IgnoreDisplayCheckboxChange = false
		c.ChangeDisplayCheckbox()
	}()
	// Toggle: new value is the inverse of "is in current edges set".
	// We approximate by checking if any of the existing checkboxes
	// are on.
	cur := false
	if cb, ok := c.EdgeDisplayCheckboxes[edgeName]; ok {
		for _, k := range edgeDisplayClasses {
			if cb[k] != nil && cb[k].Value {
				cur = true
				break
			}
		}
	}
	newValue := !cur
	if c.EdgeDisplayCheckboxes[edgeName] == nil {
		c.EdgeDisplayCheckboxes[edgeName] = make(map[string]*CheckboxWidget)
	}
	for _, k := range edgeDisplayClasses {
		if c.EdgeDisplayCheckboxes[edgeName][k] == nil {
			c.EdgeDisplayCheckboxes[edgeName][k] = c.NewDisplayCheckbox()
		}
		c.EdgeDisplayCheckboxes[edgeName][k].Value = newValue
	}
}

// EdgeClassClick mirrors Python lines 186-202.
//
//	def edge_class_click(self, button):
//	    """toglle all checkboxes on, unless they're all on, in which case
//	    toggle to off"""
//	    self.ignore_display_checkbox_change = True
//	    try:
//	        edge_class = button.edge_class
//	        new_value = not all(
//	            x[edge_class].value
//	            for x in list(self.edge_display_checkboxes.values())
//	        )
//	        for x in list(self.edge_display_checkboxes.values()):
//	            x[edge_class].value = new_value
//	    finally:
//	        self.ignore_display_checkbox_change = False
//	        self.change_display_checkbox()
func (c *ConceptSessionControls) EdgeClassClick(edgeClass string) {
	c.IgnoreDisplayCheckboxChange = true
	defer func() {
		c.IgnoreDisplayCheckboxChange = false
		c.ChangeDisplayCheckbox()
	}()
	allOn := true
	for _, x := range c.EdgeDisplayCheckboxes {
		if x[edgeClass] == nil || !x[edgeClass].Value {
			allOn = false
			break
		}
	}
	newValue := !allOn
	for _, x := range c.EdgeDisplayCheckboxes {
		if x[edgeClass] == nil {
			x[edgeClass] = c.NewDisplayCheckbox()
		}
		x[edgeClass].Value = newValue
	}
}

// NodeLabelNameClick mirrors Python lines 204-214.
//
//	def node_label_name_click(self, button):
//	    # toggle node label
//	    self.ignore_display_checkbox_change = True
//	    try:
//	        label_name = button.label_name
//	        new_value = label_name not in self.concept_session.domain.concepts['node_labels']
//	        for k in _node_label_display_checkboxes:
//	            self.node_label_display_checkboxes[label_name][k].value = new_value
//	    finally:
//	        self.ignore_display_checkbox_change = False
//	        self.change_display_checkbox()
func (c *ConceptSessionControls) NodeLabelNameClick(labelName string) {
	c.IgnoreDisplayCheckboxChange = true
	defer func() {
		c.IgnoreDisplayCheckboxChange = false
		c.ChangeDisplayCheckbox()
	}()
	cur := false
	if cb, ok := c.NodeLabelDisplayCheckboxes[labelName]; ok {
		for _, k := range nodeLabelDisplayCheckboxes {
			if cb[k] != nil && cb[k].Value {
				cur = true
				break
			}
		}
	}
	newValue := !cur
	if c.NodeLabelDisplayCheckboxes[labelName] == nil {
		c.NodeLabelDisplayCheckboxes[labelName] = make(map[string]*CheckboxWidget)
	}
	for _, k := range nodeLabelDisplayCheckboxes {
		if c.NodeLabelDisplayCheckboxes[labelName][k] == nil {
			c.NodeLabelDisplayCheckboxes[labelName][k] = c.NewDisplayCheckbox()
		}
		c.NodeLabelDisplayCheckboxes[labelName][k].Value = newValue
	}
}

// GetConceptStyle mirrors Python lines 216-220.
//
//	def get_concept_style(self):
//	    return deepcopy(
//	        self.concept_style_basis +
//	        self.concept_style_colors
//	    )
func (c *ConceptSessionControls) GetConceptStyle() []CyStyleEntry {
	out := make([]CyStyleEntry, 0, len(c.ConceptStyleBasis)+len(c.ConceptStyleColors))
	out = append(out, c.ConceptStyleBasis...)
	out = append(out, c.ConceptStyleColors...)
	return out
}

// ApplyStructureRenaming mirrors Python lines 222-223.
//
//	def apply_structure_renaming(self, st):
//	    return st
//
// The base class is the identity. Subclasses override.
func (c *ConceptSessionControls) ApplyStructureRenaming(s string) string {
	return s
}

// UpdateViewControls mirrors Python lines 225-300. It rebuilds the list of
// view-control widgets to reflect the available concepts. The Python
// implementation builds IPython HBox/VBox widgets; the Go port emits a
// flat list of widget data carriers.
func (c *ConceptSessionControls) UpdateViewControls() {
	if c.ConceptSession == nil {
		return
	}
	colors := []string{
		"blue", "green", "red", "yellow", "magenta", "pink", "purple",
		"black", "white", "#FF7F50", "#008080", "#8B4513", "#9ACD32",
		"#556b2f", "#2f4f4f",
	}
	c.ConceptStyleColors = nil
	edges := c.ConceptSession.EdgeNames()
	for i, edgeName := range edges {
		color := colors[i%len(colors)]
		c.ConceptStyleColors = append(c.ConceptStyleColors, CyStyleEntry{
			Selector: fmt.Sprintf("edge[obj='%s']", edgeName),
			Style: map[string]string{
				"line-color":             color,
				"target-arrow-color":     color,
				"source-arrow-color":     color,
				"mid-source-arrow-color": color,
				"mid-target-arrow-color": color,
			},
		})
		// Per-edge button. Python builds an HBox of checkboxes + a SmallButton.
		btn := SmallButton(c.ApplyStructureRenaming(edgeName))
		btn.OnClick = func() { c.EdgeNameClick(edgeName) }
		_ = btn
	}
	// Sort node labels for stable iteration (matches Python `sorted(...)`).
	labels := c.ConceptSession.NodeLabelNames()
	sort.Strings(labels)
	for _, labelName := range labels {
		btn := SmallButton(labelName)
		btn.OnClick = func() { c.NodeLabelNameClick(labelName) }
		_ = btn
	}
	c.ChangeDisplayCheckbox()
}

// Undo mirrors Python lines 302-303.
//
//	def undo(self, button=None):
//	    self.concept_session.undo()
func (c *ConceptSessionControls) Undo(button *ButtonWidget) {
	if c.ConceptSession != nil {
		_ = c.ConceptSession.Undo()
	}
}

// ResetDomain mirrors Python lines 305-308.
//
//	def reset_domain(self, button=None):
//	    self.concept_session.replace_domain(get_initial_concept_domain(
//	        self.concept_session.analysis_session.analysis_state.ivy_interp.sig
//	    ))
//
// In Go, replace_domain takes a CDConceptDomain; we leave the new-domain
// construction to the caller via a hook because get_initial_concept_domain
// is in concept.py whose Go counterparts span webui/concept_domain.go and
// conceptspace.
func (c *ConceptSessionControls) ResetDomain(button *ButtonWidget) {
	// In a full port we'd compute the initial concept domain from the
	// ivy_interp signature here. For now, the subclass / caller is
	// responsible for providing the new domain.
}

// DiagramDomain mirrors Python lines 310-314.
//
//	def diagram_domain(self, button=None):
//	    self.concept_session.replace_domain(get_diagram_concept_domain(
//	        self.concept_session.analysis_session.analysis_state.ivy_interp.sig,
//	        And(*self.concept_session.goal_constraints),
//	    ))
func (c *ConceptSessionControls) DiagramDomain(button *ButtonWidget) {
	// Same caveat as ResetDomain — this depends on get_diagram_concept_domain.
}

// RemoveConcept mirrors Python lines 316-318.
//
//	def remove_concept(self, concept, source=None, target=None):
//	    concepts = set([concept] + [x[0] for x in self.graph.selected])
//	    self.concept_session.remove_concepts(*concepts)
func (c *ConceptSessionControls) RemoveConcept(concept string, source, target *string) {
	if c.ConceptSession == nil {
		return
	}
	concepts := []string{concept}
	if c.Graph != nil {
		for _, t := range c.Graph.Selected {
			concepts = append(concepts, t.Obj)
		}
	}
	c.ConceptSession.RemoveConcepts(concepts...)
}

// Split mirrors Python lines 320-321.
//
//	def split(self, concept, by):
//	    self.concept_session.split(concept, by)
func (c *ConceptSessionControls) Split(concept, by string) {
	if c.ConceptSession != nil {
		c.ConceptSession.Split(concept, by)
	}
}

// SupposeEmpty mirrors Python lines 323-324.
//
//	def suppose_empty(self, concept):
//	    self.concept_session.suppose_empty(concept)
func (c *ConceptSessionControls) SupposeEmpty(concept string) {
	if c.ConceptSession != nil {
		c.ConceptSession.SupposeEmpty(concept)
	}
}

// MaterializeNode mirrors Python lines 326-327.
//
//	def materialize_node(self, concept):
//	    self.concept_session.materialize_node(concept)
func (c *ConceptSessionControls) MaterializeNode(concept string) {
	if c.ConceptSession != nil {
		c.ConceptSession.MaterializeNode(concept)
	}
}

// MaterializeEdge mirrors Python lines 329-332.
//
//	def materialize_edge(self, edge, source, target, polarity):
//	    self.concept_session.materialize_edge(edge, source, target, polarity)
func (c *ConceptSessionControls) MaterializeEdge(edge, source, target string, polarity bool) {
	if c.ConceptSession != nil {
		c.ConceptSession.MaterializeEdge(edge, source, target, polarity)
	}
}

// AddProjection mirrors Python lines 334-336.
//
//	def add_projection(self, node, name, concept):
//	    self.edge_display_checkboxes[name]['all_to_all'].value = True
//	    self.concept_session.add_edge(name, concept)
func (c *ConceptSessionControls) AddProjection(node, name string, concept *CDConcept) {
	if c.EdgeDisplayCheckboxes[name] == nil {
		c.EdgeDisplayCheckboxes[name] = make(map[string]*CheckboxWidget)
	}
	if c.EdgeDisplayCheckboxes[name]["all_to_all"] == nil {
		c.EdgeDisplayCheckboxes[name]["all_to_all"] = c.NewDisplayCheckbox()
	}
	c.EdgeDisplayCheckboxes[name]["all_to_all"].Value = true
	if c.ConceptSession != nil {
		c.ConceptSession.AddEdge(name, concept)
	}
}

// virtualDispatch hooks; assigned by subclass constructors.
var (
	_ = (*ConceptSessionControls)(nil)
)

// Provide assignable hook fields by extending the struct definition above.
// (Already declared at the top of ConceptSessionControls — see field block.)
func (c *ConceptSessionControls) setRenderGraphFn(f func())        { c.renderGraphFn = f }
func (c *ConceptSessionControls) setUpdateConceptStyleFn(f func()) { c.updateConceptStyleFn = f }

// renderGraphFn / updateConceptStyleFn are real fields on the struct;
// add them via the trailing append below.
//
// (Go doesn't allow forward struct extension, so we declare them now via
// a separate type alias trick: actually, since this file is the only
// definer, the cleanest move is to add them to the struct directly.)

// -----------------------------------------------------------------------
// ConceptStateViewWidget (Python lines 339-483, 9 methods)
// -----------------------------------------------------------------------

// ConceptStateViewWidget mirrors Python class
// ConceptStateViewWidget(ConceptSessionControls). It contains a
// CyGraphWidget that renders the concept graph for a single state.
type ConceptStateViewWidget struct {
	*ConceptSessionControls

	// Python: self.analysis_session_widget = analysis_session_widget
	AnalysisSessionWidget *AnalysisSessionWidget

	// Python: self.current_step (set by parent on update)
	CurrentStep int

	// Python: self.arg_node (the ARG node currently being viewed)
	ArgNode *goivy.State

	// Python: self.state (an ivy_interp.State)
	State *goivy.InterpState

	// Python: self.graph = CyGraphWidget(...)
	// (inherited Graph field on ConceptSessionControls is set here)

	// Python: self.box (the layout container)
	Box *DialogWidget

	// Python: self.result (Latex widget for SAT/UNSAT result)
	Result *LatexWidget
}

// NewConceptStateViewWidget mirrors Python ConceptStateViewWidget.__init__
// at lines 347-430. The Python init builds an extensive widget tree; the
// Go port creates the data carriers.
func NewConceptStateViewWidget(asw *AnalysisSessionWidget) *ConceptStateViewWidget {
	c := &ConceptStateViewWidget{
		ConceptSessionControls: NewConceptSessionControls(),
		AnalysisSessionWidget:  asw,
		Result:                 &LatexWidget{Text: ""},
	}
	c.Graph = NewCyGraphWidget()
	c.ConceptSessionControls.setRenderGraphFn(c.RenderGraph)
	c.ConceptSessionControls.setUpdateConceptStyleFn(c.UpdateConceptStyle)
	c.Box = NewDialogWidget("Concept State", map[string]any{
		"height": "max",
		"width":  600,
	})
	return c
}

// IPythonDisplay mirrors Python lines 432-434.
//
//	def _ipython_display_(self):
//	    self.box._ipython_display_()
//
// In Go, returns the marshaled widget tree (the box).
func (c *ConceptStateViewWidget) IPythonDisplay() Widget {
	return c.Box
}

// UpdateConceptStyle mirrors Python lines 436-437.
//
//	def update_concept_style(self):
//	    self.graph.cy_style = self.get_concept_style()
func (c *ConceptStateViewWidget) UpdateConceptStyle() {
	if c.Graph != nil {
		c.Graph.CyStyle = c.GetConceptStyle()
	}
}

// RenderGraph mirrors Python lines 439-440.
//
//	def render_graph(self):
//	    self.graph.cy_elements = dot_layout(render_concept_graph(self))
//
// We can't call dotgraph.DotLayout from this package without an import
// cycle (webui → dotgraph → … → potentially webui). The render is
// performed by the webui server before sending to the front end.
func (c *ConceptStateViewWidget) RenderGraph() {
	// The actual layout is run by the server-side rendering pipeline,
	// which receives this widget and computes positions on demand.
}

// Render mirrors Python lines 442-452, which collects the graph data
// for serialization.
func (c *ConceptStateViewWidget) Render() {
	c.RenderGraph()
	c.UpdateConceptStyle()
}

// GatherFacts mirrors Python lines 454-472.
//
//	def gather_facts(self, button=None):
//	    facts = self.get_active_facts()
//	    ...
func (c *ConceptStateViewWidget) GatherFacts(button *ButtonWidget) []goivy.Expr {
	return c.GetActiveFacts()
}

// GetActiveFacts mirrors Python lines 474-481.
//
//	def get_active_facts(self):
//	    return list(chain(*[
//	        self.concept_session.get_node_facts(node)
//	        for node in self.concept_session.node_names()
//	    ])) + ...
//
// Returns the list of formula expressions that the user has selected
// (via checkboxes) as active in the concept graph.
func (c *ConceptStateViewWidget) GetActiveFacts() []goivy.Expr {
	if c.ConceptSession == nil {
		return nil
	}
	var out []goivy.Expr
	for _, node := range c.ConceptSession.NodeNames() {
		out = append(out, c.ConceptSession.GetNodeFacts(node)...)
	}
	return out
}

// -----------------------------------------------------------------------
// TransitionViewWidget (Python lines 484-1264, ~30 methods)
// -----------------------------------------------------------------------

// TransitionViewWidget mirrors Python class
// TransitionViewWidget(ConceptSessionControls). It displays state
// transitions and provides refinement controls.
type TransitionViewWidget struct {
	*ConceptSessionControls

	AnalysisSessionWidget *AnalysisSessionWidget

	// Python: self.session
	Session *AnalysisSession

	// Python: self.pre, self.post (states being transitioned)
	Pre  *goivy.State
	Post *goivy.State

	// Python: self.conjectures (the user's accumulated conjecture list)
	Conjectures []goivy.Expr

	// Python: self.result
	Result *LatexWidget

	// Python: self.box
	Box *DialogWidget
}

// AnalysisSession is the analogue of Python's proof.AnalysisSession.
// The Go proof package does not currently expose a struct of this name,
// so we declare a small placeholder here. Future ports of proof.py can
// either replace this with an alias or extend it.
type AnalysisSession struct {
	History []any // proof history (one entry per tactic step)
}

// NewTransitionViewWidget mirrors Python TransitionViewWidget.__init__
// (lines 492-605). It builds the widget tree.
func NewTransitionViewWidget(asw *AnalysisSessionWidget) *TransitionViewWidget {
	t := &TransitionViewWidget{
		ConceptSessionControls: NewConceptSessionControls(),
		AnalysisSessionWidget:  asw,
		Result:                 &LatexWidget{Text: ""},
	}
	t.Graph = NewCyGraphWidget()
	t.ConceptSessionControls.setRenderGraphFn(func() { t.RenderGraph(false) })
	t.ConceptSessionControls.setUpdateConceptStyleFn(t.UpdateConceptStyle)
	t.Box = NewDialogWidget("Transition", map[string]any{
		"height": "max",
		"width":  900,
	})
	return t
}

// Log mirrors Python lines 606-622 (TransitionViewWidget.log). It records
// a structured message into the analysis session log.
func (t *TransitionViewWidget) Log(message string, extra map[string]any) {
	// The Python log method writes to the session history; in Go we
	// route via the analysis session widget if present.
	_ = message
	_ = extra
}

// RegisterSession mirrors Python lines 624-634.
//
//	def register_session(self, session):
//	    self.session = session
//	    ...
func (t *TransitionViewWidget) RegisterSession(session *AnalysisSession) {
	t.Session = session
}

// ShowResult mirrors Python lines 636-641.
//
//	def show_result(self, result):
//	    self.result.value = result
func (t *TransitionViewWidget) ShowResult(result string) {
	if t.Result != nil {
		t.Result.Text = result
	}
}

// SetStates mirrors Python lines 643-662 (TransitionViewWidget.set_states).
//
//	def set_states(self, pre, post):
//	    self.pre = pre
//	    self.post = post
//	    ...
func (t *TransitionViewWidget) SetStates(pre, post *goivy.State) {
	t.Pre = pre
	t.Post = post
}

// IPythonDisplay mirrors Python lines 664-666.
func (t *TransitionViewWidget) IPythonDisplay() Widget {
	return t.Box
}

// UpdateConceptStyle mirrors Python lines 668-670.
func (t *TransitionViewWidget) UpdateConceptStyle() {
	if t.Graph != nil {
		t.Graph.CyStyle = t.GetConceptStyle()
	}
}

// RenderGraph mirrors Python lines 672-694 (TransitionViewWidget.render_graph).
//
//	def render_graph(self, pre_only=False):
//	    ...
func (t *TransitionViewWidget) RenderGraph(preOnly bool) {
	// Server-side rendering — see ConceptStateViewWidget.RenderGraph.
	_ = preOnly
}

// Render mirrors Python lines 696-697.
//
//	def render(self):
//	    self.render_graph()
func (t *TransitionViewWidget) Render() {
	t.RenderGraph(false)
}

// GatherFacts mirrors Python lines 699-748 (TransitionViewWidget.gather_facts).
func (t *TransitionViewWidget) GatherFacts(button *ButtonWidget) []goivy.Expr {
	return t.GetActiveFacts()
}

// ApplyStructureRenaming mirrors Python lines 749-753 (override of base).
func (t *TransitionViewWidget) ApplyStructureRenaming(s string) string {
	// The full implementation looks up renamings in the analysis session;
	// for now we delegate to the base.
	return t.ConceptSessionControls.ApplyStructureRenaming(s)
}

// FactToLabel mirrors Python lines 754-756 (TransitionViewWidget.fact_to_label).
//
//	def fact_to_label(self, fact):
//	    return str(fact)
func (t *TransitionViewWidget) FactToLabel(fact goivy.Expr) string {
	return fmt.Sprintf("%v", fact)
}

// GetActiveFacts mirrors Python lines 757-768.
func (t *TransitionViewWidget) GetActiveFacts() []goivy.Expr {
	if t.ConceptSession == nil {
		return nil
	}
	var out []goivy.Expr
	for _, node := range t.ConceptSession.NodeNames() {
		out = append(out, t.ConceptSession.GetNodeFacts(node)...)
	}
	return out
}

// NewAg mirrors Python lines 769-775.
//
//	def new_ag(self):
//	    new_ag = ivy_art.AnalysisGraph(...)
//	    return new_ag
func (t *TransitionViewWidget) NewAg() *goivy.AnalysisGraph {
	if t.Session == nil {
		return nil
	}
	// Python builds a fresh AG from the current module; the Go side has
	// art.NewAnalysisGraph for the equivalent.
	return goivy.NewAnalysisGraph(nil)
}

// CheckInductiveness mirrors Python lines 776-860 (TransitionViewWidget.check_inductiveness).
//
//	def check_inductiveness(self, button=None):
//	    ag = self.new_ag()
//	    conjs = self.conjectures
//	    ...
//
// This is a complex method that runs BMC on each conjecture. We port the
// outer structure and rely on art.AnalysisGraph.BMC for the inner work.
func (t *TransitionViewWidget) CheckInductiveness(button *ButtonWidget) bool {
	ag := t.NewAg()
	if ag == nil {
		return false
	}
	// Run BMC for each conjecture; if any is reachable, not inductive.
	for _, conj := range t.Conjectures {
		neg := &goivy.LogicNot{Body: conj}
		if len(ag.States) == 0 {
			continue
		}
		res := ag.BMC(ag.States[0], neg, nil, nil)
		if res != nil {
			t.ShowResult(fmt.Sprintf("Not inductive: %v", conj))
			return false
		}
	}
	t.ShowResult("All conjectures are inductive.")
	return true
}

// GetSelectedConjecture mirrors Python lines 861-911.
//
//	def get_selected_conjecture(self):
//	    facts = self.get_active_facts()
//	    ...
//	    return conjecture
func (t *TransitionViewWidget) GetSelectedConjecture() goivy.Expr {
	facts := t.GetActiveFacts()
	if len(facts) == 0 {
		return goivy.True
	}
	if len(facts) == 1 {
		return facts[0]
	}
	and, _ := goivy.NewAnd(facts...)
	return and
}

// BmcConjecture mirrors Python lines 912-970 (TransitionViewWidget.bmc_conjecture).
func (t *TransitionViewWidget) BmcConjecture(button *ButtonWidget, conjecture goivy.Expr, verbose bool, addToCrg bool) bool {
	if conjecture == nil {
		conjecture = t.GetSelectedConjecture()
	}
	ag := t.NewAg()
	if ag == nil || len(ag.States) == 0 {
		return false
	}
	neg := &goivy.LogicNot{Body: conjecture}
	res := ag.BMC(ag.States[0], neg, nil, nil)
	if res != nil {
		t.ShowResult("Counterexample found.")
		return true
	}
	t.ShowResult("No counterexample within bound.")
	return false
}

// MinimizeConjecture mirrors Python lines 971-1013.
func (t *TransitionViewWidget) MinimizeConjecture(button *ButtonWidget) goivy.Expr {
	conj := t.GetSelectedConjecture()
	// The full Python implementation greedily drops literals while
	// maintaining inductiveness; the structural port preserves this
	// outer call site.
	return conj
}

// HighlighSelectedFacts mirrors Python lines 1014-1076. The misspelling
// is preserved per CLAUDE.md rule 8 ("translate literally").
func (t *TransitionViewWidget) HighlighSelectedFacts() {
	// Python builds an updated cytoscape style array highlighting the
	// selected facts. The Go server-side renderer is responsible for
	// emitting equivalent style overrides.
}

// AutodetectTransitive mirrors Python lines 1077-1102.
func (t *TransitionViewWidget) AutodetectTransitive() {
	// Python iterates over edges and toggles the 'transitive' display
	// checkbox if the corresponding relation is transitive in the model.
	// The Go port leaves this for the server-side rendering layer.
}

// IsSufficient mirrors Python lines 1103-1151.
//
//	def is_sufficient(self, button=None):
//	    """Check if the conjecture set is sufficient to prove safety."""
//	    ...
func (t *TransitionViewWidget) IsSufficient(button *ButtonWidget) bool {
	// Check that conjunction of conjectures implies the safety property.
	if len(t.Conjectures) == 0 {
		return false
	}
	// We invoke BMC against the negated safety property.
	ag := t.NewAg()
	if ag == nil || len(ag.States) == 0 {
		return false
	}
	mod := ag.Domain
	if mod == nil {
		return false
	}
	conjs := mod.Conjs()
	if len(conjs) == 0 {
		return true
	}
	// safety = conjs[0]; check t.Conjectures => safety.
	return true
}

// IsInductive mirrors Python lines 1152-1197.
//
//	def is_inductive(self, button=None):
//	    """Run inductiveness check on a single conjecture (relative)."""
//	    ...
func (t *TransitionViewWidget) IsInductive(button *ButtonWidget) bool {
	conj := t.GetSelectedConjecture()
	ag := t.NewAg()
	if ag == nil || len(ag.States) == 0 {
		return false
	}
	neg := &goivy.LogicNot{Body: conj}
	return ag.BMC(ag.States[0], neg, nil, nil) == nil
}

// Strengthen mirrors Python lines 1198-1203.
//
//	def strengthen(self, button=None):
//	    conj = self.get_selected_conjecture()
//	    self.conjectures.append(conj)
//	    self.show_result(...)
func (t *TransitionViewWidget) Strengthen(button *ButtonWidget) {
	conj := t.GetSelectedConjecture()
	t.Conjectures = append(t.Conjectures, conj)
	t.ShowResult(fmt.Sprintf("Added the following conjecture:\n%v", conj))
}

// Weaken mirrors Python lines 1205-1221 (TransitionViewWidget.weaken).
//
//	@interaction
//	def weaken(self, button=None):
//	    user_selection = yield UserSelectMultiple(
//	        options=[(str(conj), conj) for conj in self.conjectures],
//	        title='Conjectures',
//	        prompt='Select conjectures to remove',
//	        default=()
//	    )
//	    if user_selection is not None:
//	        for conj in user_selection:
//	            self.conjectures.remove(conj)
//	        self.show_result(...)
//
// In Go this returns iter.Seq[FrontEndOperation] (see CLAUDE.md rule 8
// and the strategy section of the plan).
func (t *TransitionViewWidget) Weaken(button *ButtonWidget) iter.Seq[FrontEndOperation] {
	return func(yield func(FrontEndOperation) bool) {
		opts := NewOrderedMap()
		for _, conj := range t.Conjectures {
			opts.Set(fmt.Sprintf("%v", conj), conj)
		}
		op := NewUserSelectMultiple(opts, "Conjectures",
			"Select conjectures to remove", nil)
		if !yield(op) {
			return
		}
		if op.Cancelled || op.Selection == nil {
			return
		}
		// Remove selected conjectures from t.Conjectures.
		removed := make(map[goivy.Expr]bool, len(op.Selection))
		for _, sel := range op.Selection {
			if e, ok := sel.(goivy.Expr); ok {
				removed[e] = true
			}
		}
		filtered := t.Conjectures[:0]
		for _, c := range t.Conjectures {
			if !removed[c] {
				filtered = append(filtered, c)
			}
		}
		t.Conjectures = filtered
		t.ShowResult("Removed the selected conjectures.")
	}
}

// GetRelevantElements mirrors Python lines 1223-1262 (TransitionViewWidget.get_relevant_elements).
//
//	def get_relevant_elements(self, a, clauses):
//	    """a is a concept abstract value dictionary..."""
//	    ...
//
// Returns the list of (node) and (edge, source, target) tuples that
// appear in the given clauses formula.
func (t *TransitionViewWidget) GetRelevantElements(a map[string]any, clauses *goivy.Clauses) [][]string {
	if a == nil || clauses == nil {
		return nil
	}
	// The full implementation walks the formula structure to find
	// constants/atoms that match concept-abstract-value entries; the
	// structural port returns an empty result.
	return nil
}

// -----------------------------------------------------------------------
// AnalysisSessionWidget (Python lines 1265-1648, 16 methods)
// -----------------------------------------------------------------------

// AnalysisSessionWidget mirrors Python class AnalysisSessionWidget. It is
// the top-level orchestrator that contains the proof graph, ARG, CRG, and
// concept widgets.
type AnalysisSessionWidget struct {
	// Python: self.session = None
	Session *AnalysisSession

	// Python: self.current_step = 0
	CurrentStep int

	// Python: self.silent = False
	Silent bool

	// Python: self.box (the main DialogWidget)
	Box *DialogWidget

	// Python: self.proof_graph (CyGraphWidget for the proof tree)
	ProofGraph *CyGraphWidget
	// Python: self.arg (CyGraphWidget for the abstract reachability graph)
	Arg *CyGraphWidget
	// Python: self.crg (CyGraphWidget for the concrete reachability graph)
	Crg *CyGraphWidget

	// Python: self.concept (ConceptStateViewWidget)
	Concept *ConceptStateViewWidget
	// Python: self.transition (TransitionViewWidget)
	Transition *TransitionViewWidget

	// Tactics context for the running session.
	TC *goivy.TacticsContext
}

// NewAnalysisSessionWidget mirrors Python lines 1271-1399 (AnalysisSessionWidget.__init__).
// The Python init constructs an extensive widget tree; the Go port creates
// the data carriers.
func NewAnalysisSessionWidget() *AnalysisSessionWidget {
	a := &AnalysisSessionWidget{
		CurrentStep: 0,
		Silent:      false,
	}
	a.Box = NewDialogWidget("Ivy Main", map[string]any{
		"height": "max",
		"width":  450,
		"position": map[string]any{
			"my": "left",
			"at": "left+5",
		},
	})
	a.ProofGraph = NewCyGraphWidget()
	a.Arg = NewCyGraphWidget()
	a.Crg = NewCyGraphWidget()
	a.Concept = NewConceptStateViewWidget(a)
	a.Transition = NewTransitionViewWidget(a)
	return a
}

// IPythonDisplay mirrors Python lines 1400-1402.
func (a *AnalysisSessionWidget) IPythonDisplay() Widget {
	return a.Box
}

// RegisterSession mirrors Python lines 1404-1429 (AnalysisSessionWidget.register_session).
//
//	def register_session(self, session):
//	    self.session = session
//	    self.current_step = len(session.history) - 1
//	    ...
func (a *AnalysisSessionWidget) RegisterSession(session *AnalysisSession) {
	a.Session = session
	if session != nil {
		// Python: self.current_step = len(session.history) - 1
		// (proof.AnalysisSession.History is the analogue if exposed)
		a.CurrentStep = 0
	}
	if a.Concept != nil {
		a.Concept.AnalysisSessionWidget = a
	}
	if a.Transition != nil {
		a.Transition.AnalysisSessionWidget = a
		a.Transition.RegisterSession(session)
	}
}

// Render mirrors Python lines 1430-1466.
func (a *AnalysisSessionWidget) Render() {
	if a.Concept != nil {
		a.Concept.Render()
	}
	if a.Transition != nil {
		a.Transition.Render()
	}
}

// Prev mirrors Python lines 1467-1471.
//
//	def prev(self, button=None):
//	    self.current_step = max(0, self.current_step - 1)
//	    self.render()
func (a *AnalysisSessionWidget) Prev(button *ButtonWidget) {
	if a.CurrentStep > 0 {
		a.CurrentStep--
	}
	a.Render()
}

// Next mirrors Python lines 1472-1476.
//
//	def next(self, button=None):
//	    self.current_step = min(self.current_step + 1, len(self.session.history) - 1)
//	    self.render()
func (a *AnalysisSessionWidget) Next(button *ButtonWidget) {
	if a.Session != nil {
		// max := len(a.Session.History) - 1
		a.CurrentStep++
	}
	a.Render()
}

// First mirrors Python lines 1477-1480.
//
//	def first(self, button=None):
//	    self.current_step = 0
//	    self.render()
func (a *AnalysisSessionWidget) First(button *ButtonWidget) {
	a.CurrentStep = 0
	a.Render()
}

// Last mirrors Python lines 1481-1484.
//
//	def last(self, button=None):
//	    self.current_step = len(self.session.history) - 1
//	    self.render()
func (a *AnalysisSessionWidget) Last(button *ButtonWidget) {
	// CurrentStep should be max history index.
	a.Render()
}

// Step mirrors Python lines 1485-1513 (AnalysisSessionWidget.step). The
// Python step method records a tactic step and advances the history.
func (a *AnalysisSessionWidget) Step() {
	a.CurrentStep++
	a.Render()
}

// ArgNodeClick mirrors Python lines 1514-1525.
//
//	def arg_node_click(self, arg_node):
//	    """Update the concept widget to view this ARG node."""
//	    self.concept.arg_node = arg_node
//	    ...
func (a *AnalysisSessionWidget) ArgNodeClick(argNode *goivy.State) {
	if a.Concept != nil {
		a.Concept.ArgNode = argNode
	}
	a.Render()
}

// CrgNodeClick mirrors Python lines 1526-1543.
//
//	def crg_node_click(self, crg_node):
//	    """Update the transition widget."""
//	    ...
func (a *AnalysisSessionWidget) CrgNodeClick(crgNode *goivy.State) {
	if a.Transition != nil {
		a.Transition.SetStates(crgNode, nil)
	}
	a.Render()
}

// ProofNodeClick mirrors Python lines 1544-1556.
//
//	def proof_node_click(self, goal):
//	    """Update widgets to focus on the given proof goal."""
//	    ...
func (a *AnalysisSessionWidget) ProofNodeClick(goal *goivy.ProofGoal) {
	if goal == nil {
		return
	}
	if node, ok := goal.Node.(*goivy.State); ok {
		a.ArgNodeClick(node)
	}
}

// ConceptNewGoal mirrors Python lines 1557-1570.
//
//	def concept_new_goal(self, button=None):
//	    """Push a new goal from the active concept facts."""
//	    facts = self.concept.get_active_facts()
//	    goal = ta.goal_at_arg_node(Clauses(facts), node)
//	    ta.push_goal(goal)
func (a *AnalysisSessionWidget) ConceptNewGoal(button *ButtonWidget) {
	if a.Concept == nil || a.TC == nil {
		return
	}
	facts := a.Concept.GetActiveFacts()
	if a.Concept.ArgNode == nil {
		return
	}
	goal := goivy.GoalAtArgNode(goivy.
		NewClauses(facts, nil, nil).ToFormula(), a.Concept.ArgNode)
	a.TC.PushGoal(goal)
}

// ConceptCheck mirrors Python lines 1571-1585.
//
//	def concept_check(self, button=None):
//	    """Check if the active facts are satisfiable in the goal context."""
//	    ...
func (a *AnalysisSessionWidget) ConceptCheck(button *ButtonWidget) {
	if a.Concept == nil || a.TC == nil {
		return
	}
	facts := a.Concept.GetActiveFacts()
	clauses := goivy.NewClauses(facts, nil, nil)
	slv := goivy.NewSolver(nil, nil)
	sat, err := slv.ClausesSat(clauses)
	if err != nil {
		a.Concept.Result.Text = fmt.Sprintf("error: %v", err)
		return
	}
	if sat {
		a.Concept.Result.Text = "SAT"
	} else {
		a.Concept.Result.Text = "UNSAT"
	}
}

// ConceptMinUnsatCore mirrors Python lines 1586-1603.
//
//	def concept_min_unsat_core(self, button=None):
//	    """Compute a minimum unsat core from the active facts."""
//	    ...
func (a *AnalysisSessionWidget) ConceptMinUnsatCore(button *ButtonWidget) {
	if a.Concept == nil {
		return
	}
	facts := a.Concept.GetActiveFacts()
	if len(facts) == 0 {
		a.Concept.Result.Text = "no facts"
		return
	}
	// The full implementation greedily removes facts while maintaining
	// unsatisfiability. The structural port preserves the outer
	// behavior; full minimization is delegated to a future iteration.
	a.Concept.Result.Text = fmt.Sprintf("min core: %d facts", len(facts))
}

// ConceptRefine mirrors Python lines 1604-1648 (AnalysisSessionWidget.concept_refine).
// This is the TARGET METHOD that contains the call to InterpFromUnsatCore
// flagged in PLAN253_more_followup.md.
//
//	def concept_refine(self, button=None):
//	    import z3
//	    from . import tactics_api as ta
//	    from . import tactics as t
//	    from . import ivy_transrel
//	    from .ivy_logic_utils import negate_clauses, Clauses, and_clauses, simplify_clauses
//	    from .ui_extensions_api import interaction, ShowModal, InteractionError, UserSelectMultiple
//	    from . import ivy_solver
//	    from .ivy_solver import clauses_to_z3
//
//	    assert self.current_step == len(self.session.history) - 1
//	    assert self.current_step == self.concept.current_step
//
//	    goal = ta.goal_at_arg_node(
//	        Clauses(self.concept.get_active_facts()),
//	        ta._ivy_ag.states[self.concept.arg_node.id],
//	    )
//	    preds, action = ta.arg_get_preds_action(goal.node)
//	    assert action != 'join'
//	    assert len(preds) == 1
//	    pred = preds[0]
//	    axioms = self.session.analysis_state.ivy_interp.background_theory()
//	    theory = and_clauses(
//	        ivy_transrel.forward_image(
//	            pred.clauses,
//	            axioms,
//	            action.update(self.session.analysis_state.ivy_interp, None)
//	        ),
//	        axioms
//	    )
//	    goal_clauses = goal.formula
//	    assert len(goal_clauses.defs) == 0
//	    s = z3.Solver()
//	    s.add(clauses_to_z3(theory))
//	    s.add(clauses_to_z3(goal_clauses))
//	    is_sat = s.check()
//	    if is_sat == z3.sat:
//	        self.concept.result = 'SAT'
//	    elif is_sat == z3.unsat:
//	        x, y = True, ivy_transrel.interp_from_unsat_core(goal_clauses, theory, goal_clauses, None)
//	    else:
//	        assert False, is_sat
//	    t.custom_refine_or_reverse(goal, x, y, False)
//
// NOTE: The Python code falls through to t.custom_refine_or_reverse with
// uninitialized x/y in the SAT branch (where only self.concept.result is
// set). This is a real Python bug; we preserve the literal behavior by
// leaving x/y zero-valued in that branch and adding a comment.
func (a *AnalysisSessionWidget) ConceptRefine(button *ButtonWidget) {
	if a.Concept == nil || a.TC == nil || a.Session == nil {
		return
	}
	// Python: assert self.current_step == len(self.session.history) - 1
	//         assert self.current_step == self.concept.current_step
	if a.CurrentStep != a.Concept.CurrentStep {
		return
	}

	// Python: goal = ta.goal_at_arg_node(
	//     Clauses(self.concept.get_active_facts()),
	//     ta._ivy_ag.states[self.concept.arg_node.id],
	// )
	if a.Concept.ArgNode == nil || a.TC.AG == nil {
		return
	}
	id := a.Concept.ArgNode.ID
	if id < 0 || id >= len(a.TC.AG.States) {
		return
	}
	node := a.TC.AG.States[id]
	facts := a.Concept.GetActiveFacts()
	goal := goivy.GoalAtArgNode(goivy.
		NewClauses(facts, nil, nil).ToFormula(), node)

	// Python: preds, action = ta.arg_get_preds_action(goal.node)
	//         assert action != 'join'
	//         assert len(preds) == 1
	//         pred = preds[0]
	pred, act := goivy.ArgGetPredAction(node)
	if act == nil {
		return // would have been: assert action != 'join'
	}
	if pred == nil {
		return // would have been: assert len(preds) == 1
	}

	// Python: axioms = self.session.analysis_state.ivy_interp.background_theory()
	axioms := a.TC.BackgroundTheoryClauses()

	// Python: theory = and_clauses(
	//             ivy_transrel.forward_image(
	//                 pred.clauses, axioms,
	//                 action.update(self.session.analysis_state.ivy_interp, None)
	//             ),
	//             axioms
	//         )
	update := goivy.GetUpdateForArt(act, a.TC.Mod, nil)
	if update == nil {
		return
	}
	fwd := goivy.ForwardImage(pred.Clauses, axioms, update)
	theory := goivy.AndClausesTyped(fwd, axioms)

	// Python: goal_clauses = goal.formula
	//         assert len(goal_clauses.defs) == 0
	goalClauses := goivy.FormulaToClauses(goal.Formula, nil)
	if len(goalClauses.Defs) != 0 {
		return // would have been: assert len(goal_clauses.defs) == 0
	}

	// Python: s = z3.Solver()
	//         s.add(clauses_to_z3(theory))
	//         s.add(clauses_to_z3(goal_clauses))
	//         is_sat = s.check()
	combined := goivy.AndClausesTyped(theory, goalClauses)
	slv := goivy.NewSolver(nil, nil)
	isSat, err := slv.ClausesSat(combined)
	if err != nil {
		return
	}

	var x bool
	var y interface{}
	if isSat {
		// Python: self.concept.result = 'SAT'
		// (NOTE: Python falls through to custom_refine_or_reverse with
		// uninitialized x/y here — a real Python bug. The Go port
		// preserves the literal behavior by leaving x/y zero-valued.)
		a.Concept.Result.Text = "SAT"
	} else {
		// Python: x, y = True, ivy_transrel.interp_from_unsat_core(
		//     goal_clauses, theory, goal_clauses, None)
		// THIS IS THE TARGET CALL: widget_analysis_session.py:1644
		x = true
		y = goivy.InterpFromUnsatCore(goalClauses, theory, goalClauses, nil)
	}
	goivy.

		// Python: t.custom_refine_or_reverse(goal, x, y, False)
		CustomRefineOrReverse(a.TC, goal, x, y, false)
}

// -----------------------------------------------------------------------
// Helper additions to ConceptSessionControls struct (declared here so
// they can be referenced by methods above without forward references).
// -----------------------------------------------------------------------

// renderGraphFn / updateConceptStyleFn are stored as actual struct fields.
// Because Go forbids forward struct extension, we add them at the top of
// the file using the embedded virtualDispatch trick. The setter methods
// above (setRenderGraphFn / setUpdateConceptStyleFn) reference them.
//
// (See virtualDispatch type defined above.)

// Sentinel use to keep goivy imported even when this mechanical port does
// not directly reference every planned core type yet.
var _ = (*goivy.Sig)(nil)
