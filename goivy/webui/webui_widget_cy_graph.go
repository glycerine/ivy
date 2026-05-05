// Mechanical port of widget_cy_graph.py.
//
// Python source: ~/ivy/pyivy/ivy/ivy/widget_cy_graph.py (222 lines)
//
// The Python file declares two classes:
//   - CyGraphWidget (lines 26-115)        — IPython DOMWidget for cytoscape
//   - WebUICyElements   (lines 118-222)        — graph element accumulator
//
// The WebUICyElements class in widget_cy_graph.py is a verbatim duplicate of
// the canonical class in cy_elements.py, which is already ported to Go as
// art.CyElements (see art/cyrender.go). Re-porting it here would create a
// namespace collision with no benefit, so this file ports only the
// CyGraphWidget class. Calls that use WebUICyElements use art.CyElements.

package webui

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/art"
)

// CyGraphWidget mirrors widget_cy_graph.py:CyGraphWidget. In Python this
// is an IPython DOMWidget that wraps a Cytoscape.js graph; in Go it is a
// data carrier that the webui server serializes to its JSON protocol.
type CyGraphWidget struct {
	// Trait fields (Python class attributes)
	ViewModule string
	ViewName   string

	// Synced traits — these mirror Python's _cy_elements, cy_style,
	// cy_layout, selected, info_area attributes.
	cyElements *art.CyElements // Python: _cy_elements
	CyStyle    []CyStyleEntry  // Python: cy_style (defined in cystyles.go)
	CyLayout   any             // Python: cy_layout (any layout config)
	Selected   []CyTuple       // Python: selected (see _ele_to_tuple)
	InfoArea   any             // Python: info_area

	// Local fields set in __init__
	BackgroundColor string

	// Object reference dictionaries (Python: _obj_to_key, _key_to_obj).
	// These maintain Go-side identity for Python user objects so that
	// callbacks can still resolve them after a JSON round-trip.
	objToKey map[any]string
	keyToObj map[string]any
}

// CyTuple is a graph-element identity tuple as returned by
// CyGraphWidget._ele_to_tuple. For nodes it has length 1; for edges,
// length 3 (obj, source_obj, target_obj).
type CyTuple struct {
	Obj       string
	SourceObj string // empty for nodes
	TargetObj string // empty for nodes
}

// NewCyGraphWidget mirrors CyGraphWidget.__init__(self, **kwargs).
func NewCyGraphWidget() *CyGraphWidget {
	return &CyGraphWidget{
		ViewModule:      "nbextensions/ivy/js/widget_cy_graph",
		ViewName:        "CyGraphView",
		BackgroundColor: "rgb(192,192,255)",
		objToKey:        make(map[any]string),
		keyToObj:        make(map[string]any),
	}
}

func (*CyGraphWidget) widget() {}

// WebUICyElements is the property getter for self._cy_elements.
// Mirrors Python: cy_elements = property(lambda self: self._cy_elements).
func (w *CyGraphWidget) WebUICyElements() *art.CyElements {
	return w.cyElements
}

// SetCyElements is the property setter for self._cy_elements.
// Mirrors Python lines 47-61:
//
//	@cy_elements.setter
//	def cy_elements(self, value):
//	    assert type(value) is WebUICyElements
//	    elements = value.elements
//	    value.elements = None       # prevents future use of this graph
//	    if self._cy_elements != elements:
//	        self.selected = []      # clear selection
//	        self._cy_elements = elements
//
// In Go we cannot null out the source WebUICyElements' Elements slice the way
// Python does (would mutate a foreign value), but we still clear the
// selection on assignment as the Python code does.
func (w *CyGraphWidget) SetCyElements(value *art.CyElements) {
	if value == nil {
		return
	}
	if w.cyElements != value {
		w.Selected = nil // Python: self.selected = []
		w.cyElements = value
	}
}

// eleToTuple mirrors Python CyGraphWidget._ele_to_tuple at lines 63-67.
// For node elements returns (obj,); for edges returns
// (obj, source_obj, target_obj).
func (w *CyGraphWidget) eleToTuple(ele art.CyElement) CyTuple {
	if ele.Group == "nodes" {
		return CyTuple{Obj: asString(ele.Data["obj"])}
	}
	return CyTuple{
		Obj:       asString(ele.Data["obj"]),
		SourceObj: asString(ele.Data["source_obj"]),
		TargetObj: asString(ele.Data["target_obj"]),
	}
}

// Elements mirrors the Python `elements` property at lines 69-75:
//
//	@property
//	def elements(self):
//	    return [self._ele_to_tuple(ele) for ele in self.cy_elements]
//
// Returns the underlying graph elements as identity tuples.
func (w *CyGraphWidget) Elements() []CyTuple {
	if w.cyElements == nil {
		return nil
	}
	out := make([]CyTuple, 0, len(w.cyElements.Elements))
	for _, ele := range w.cyElements.Elements {
		out = append(out, w.eleToTuple(ele))
	}
	return out
}

// HandleCyMsg mirrors Python CyGraphWidget._handle_cy_msg at lines 77-80.
// In Python this is a callback registered via on_msg; in Go we expose it
// as a method that the webui server invokes when the front end sends a
// message. The Python implementation:
//
//	def _handle_cy_msg(self, _, content):
//	    content = self._trait_from_json(content)
//	    if content['type'] == 'callback':
//	        content['callback'](*content['args'])
//
// dispatches a callback by the registered key. In Go the callback table
// is held in keyToObj.
func (w *CyGraphWidget) HandleCyMsg(content map[string]any) {
	content = w.traitFromJSON(content).(map[string]any)
	if t, _ := content["type"].(string); t == "callback" {
		if cb, ok := content["callback"].(func(...any)); ok {
			args, _ := content["args"].([]any)
			cb(args...)
		}
	}
}

// ExecuteNewCell mirrors Python CyGraphWidget.execute_new_cell at lines 82-89:
//
//	def execute_new_cell(self, code):
//	    self.send({
//	        "method": "execute_new_cell",
//	        "code": code,
//	    })
//
// In Go, this returns the message dictionary that the webui server flushes
// to the front end as JSON. The Python self.send is the IPython equivalent.
func (w *CyGraphWidget) ExecuteNewCell(code string) map[string]any {
	return map[string]any{
		"method": "execute_new_cell",
		"code":   code,
	}
}

// _object_prefix from widget_cy_graph.py:20.
const objectPrefix = "CY_OBJECT_"

// _object_key from widget_cy_graph.py:17-18:
//
//	def _object_key(x):
//	    return str(id(x))
//
// In Go, the equivalent is the address of the value (for pointers) or
// the string form of the value. The webui front end never inspects this
// key, it only round-trips it.
func objectKey(x any) string {
	return asString(x)
}

// _is_user_object from widget_cy_graph.py:22-23:
//
//	def _is_user_object(x):
//	    return str(type(x)).startswith('<class ')
//
// In Python this checks whether x is an instance of a user-defined class.
// In Go we approximate by checking whether x is a pointer to a struct.
// This helper is only used by traitToJSON below.
func isUserObject(x any) bool {
	if x == nil {
		return false
	}
	// All pointers and structs in Go count as "user objects" here;
	// primitive scalars (string, numeric, bool) do not.
	switch x.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return false
	}
	return true
}

// traitToJSON mirrors Python CyGraphWidget._trait_to_json at lines 94-105.
// It substitutes a key for any callable or user object so that the
// front end can round-trip the reference back to Go.
func (w *CyGraphWidget) traitToJSON(x any) any {
	if _, isFn := x.(func(...any)); isFn || isUserObject(x) {
		k, ok := w.objToKey[x]
		if !ok {
			k = objectKey(x)
			w.objToKey[x] = k
			w.keyToObj[k] = x
		}
		return objectPrefix + k
	}
	return x
}

// traitFromJSON mirrors Python CyGraphWidget._trait_from_json at lines 107-115.
// If x is a string starting with the object prefix, it resolves the
// reference back to the original Go object.
func (w *CyGraphWidget) traitFromJSON(x any) any {
	if s, ok := x.(string); ok && len(s) > len(objectPrefix) && s[:len(objectPrefix)] == objectPrefix {
		return w.keyToObj[s[len(objectPrefix):]]
	}
	return x
}

// asString returns a stable string representation of x. Used for object
// identity keys and for converting WebUICyElement data values to strings.
func asString(x any) string {
	if x == nil {
		return ""
	}
	if s, ok := x.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", x)
}
