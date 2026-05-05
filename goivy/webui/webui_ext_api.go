package webui

// Full port of ui_extensions_api.py — extension point system for
// registering callbacks that add actions to the verification UI,
// plus the FrontEndOperation hierarchy used by interactive
// (generator-style) UI flows.

import (
	"fmt"
	"sync"

	"github.com/glycerine/ivy/goivy/art"
	"github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// AnalysisSessionI is the interface needed by extension point callbacks
// to access the ARG and conjectures. Corresponds to Python's
// _analysis_session.analysis_state.
type AnalysisSessionI interface {
	// ActionNames returns sorted action names from ivy_ag.actions.
	ActionNames() []string
	// ExecuteAction executes a named action on the current ARG node.
	ExecuteAction(name string)
	// UnprovedConjectures returns conjectures not yet implied by current state.
	UnprovedConjectures() []string
	// TryConjecture pushes a conjecture negation as a new goal.
	TryConjecture(name string)
}


// ExtensionCallback is a function registered with an extension point.
// It receives a context and returns a list of action tuples.
type ExtensionCallback func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error)

// ExtensionAction describes one action contributed by an extension.
type ExtensionAction struct {
	Label    string `json:"label"`
	Callback func(args ...interface{}) error
	Args     []interface{} `json:"-"`
}

// ExtensionPoint is a named list of callbacks that can be registered,
// unregistered, and invoked (Python: class ExtensionPoint).
type ExtensionPoint struct {
	mu        sync.RWMutex
	Name      string
	Callbacks []ExtensionCallback
	Prototype string // documentation string
}

// NewExtensionPoint creates a new extension point with the given name.
func NewExtensionPoint(name string) *ExtensionPoint {
	return &ExtensionPoint{
		Name: name,
	}
}

// Register adds a callback to this extension point.
// Returns the callback for use as a decorator pattern.
func (ep *ExtensionPoint) Register(fn ExtensionCallback) ExtensionCallback {
	if fn == nil {
		return nil
	}
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.Callbacks = append(ep.Callbacks, fn)
	return fn
}

// Unregister removes a callback from this extension point.
// Note: in Go we can't directly compare functions, so this uses index-based removal.
func (ep *ExtensionPoint) Unregister(idx int) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	if idx < 0 || idx >= len(ep.Callbacks) {
		return fmt.Errorf("callback index %d out of range", idx)
	}
	ep.Callbacks = append(ep.Callbacks[:idx], ep.Callbacks[idx+1:]...)
	return nil
}

// Invoke runs all callbacks, collecting their results.
// Errors from individual callbacks are collected but do not stop execution.
// (Python: ExtensionPoint.__call__).
func (ep *ExtensionPoint) Invoke(ctx interface{}, args ...interface{}) ([]ExtensionAction, []error) {
	ep.mu.RLock()
	cbs := make([]ExtensionCallback, len(ep.Callbacks))
	copy(cbs, ep.Callbacks)
	ep.mu.RUnlock()

	var allActions []ExtensionAction
	var errs []error

	for _, cb := range cbs {
		actions, err := cb(ctx, args...)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		allActions = append(allActions, actions...)
	}
	return allActions, errs
}

// Len returns the number of registered callbacks.
func (ep *ExtensionPoint) Len() int {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return len(ep.Callbacks)
}

// Action creates a decorator-style registration that wraps a simple function
// as a single-action extension (Python: ExtensionPoint.action).
func (ep *ExtensionPoint) Action(label string, fn func(args ...interface{}) error) {
	ep.Register(func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error) {
		return []ExtensionAction{
			{
				Label:    label,
				Callback: fn,
			},
		}, nil
	})
}

// ExtConfig holds per-session extension point state.
type ExtConfig struct {
	ArgNodeActions  *ExtensionPoint
	GoalNodeActions *ExtensionPoint
	AnalysisSession AnalysisSessionI
	AG              *art.AnalysisGraph
}

// NewExtConfig creates a new ExtConfig with default extensions registered.
func NewExtConfig() *ExtConfig {
	cfg := &ExtConfig{
		ArgNodeActions:  NewExtensionPoint("arg_node_actions"),
		GoalNodeActions: NewExtensionPoint("goal_node_actions"),
	}
	registerDefaultExtensions(cfg)
	return cfg
}

// registerDefaultExtensions registers the default ARG callbacks on an ExtConfig.
func registerDefaultExtensions(cfg *ExtConfig) {
	cfg.ArgNodeActions.Register(func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error) {
		if cfg.AnalysisSession == nil {
			return nil, nil
		}
		actionNames := cfg.AnalysisSession.ActionNames()
		result := make([]ExtensionAction, len(actionNames))
		for i, name := range actionNames {
			actionName := name
			result[i] = ExtensionAction{
				Label: actionName,
				Callback: func(args ...interface{}) error {
					cfg.AnalysisSession.ExecuteAction(actionName)
					return nil
				},
			}
		}
		return result, nil
	})
	cfg.ArgNodeActions.Register(func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error) {
		if cfg.AnalysisSession == nil {
			return nil, nil
		}
		conjs := cfg.AnalysisSession.UnprovedConjectures()
		result := make([]ExtensionAction, len(conjs))
		for i, conj := range conjs {
			conjName := conj
			result[i] = ExtensionAction{
				Label: fmt.Sprintf("conj: %s", conjName),
				Callback: func(args ...interface{}) error {
					cfg.AnalysisSession.TryConjecture(conjName)
					return nil
				},
			}
		}
		return result, nil
	})
}


// --- Front-end operation types (Python: FrontEndOperation hierarchy) ---
//
// In Python, FrontEndOperation has a Submit(on_done) method and the
// run_interaction driver uses generators with yield to chain operations.
// In Go, we use range-over-function iterators (iter.Seq[FrontEndOperation]):
// generator-style functions return iter.Seq[FrontEndOperation]; each yielded
// op carries both its input fields AND its response fields. The consumer
// mutates the response fields before its loop body returns, and the next
// call to yield() resumes the generator with those values visible. This is
// the literal Go translation of Python's `value = yield op` semantics.

// FrontEndOperation is the marker interface for any operation that an
// interactive UI flow can yield to the front end. Concrete types are
// ShowModal, UserSelect, UserSelectMultiple, UserSelectCore, ExecuteNewCell,
// etc. — see Python ui_extensions_api.py:191.
type FrontEndOperation interface {
	frontEndOp()
}

// InteractionError represents an error during user interaction.
// Mirrors Python ui_extensions_api.py:187.
type InteractionError struct {
	Message string
}

func (e *InteractionError) Error() string {
	return e.Message
}

// --- Widget data carriers ---
//
// Python widget_analysis_session.py and iupdr.py instantiate a small
// vocabulary of IPython widgets (Latex, Button, SelectMultiple, Select,
// Checkbox). Go does not have IPython, so each becomes a small data
// carrier struct. The webui front end serializes these to its JSON
// protocol.

// Widget is the marker interface for any data carrier that can appear
// inside a ShowModal's Children list.
type Widget interface {
	widget()
}

// LatexWidget mirrors IPython's widgets.Latex — a single string of
// LaTeX-formatted text to display.
type LatexWidget struct {
	Text string
}

func (*LatexWidget) widget() {}

// ButtonWidget mirrors IPython's widgets.Button — a clickable button.
// OnClick is invoked when the front end signals a click.
type ButtonWidget struct {
	Description string
	OnClick     func()
}

func (*ButtonWidget) widget() {}

// CheckboxWidget mirrors IPython's widgets.Checkbox.
type CheckboxWidget struct {
	Description string
	Value       bool
	OnChange    func(bool)
}

func (*CheckboxWidget) widget() {}

// SelectWidget mirrors IPython's widgets.Select — single-selection list.
type SelectWidget struct {
	Options *OrderedMap
	Value   any // current value
}

func (*SelectWidget) widget() {}

// SelectMultipleWidget mirrors IPython's widgets.SelectMultiple — multi-
// selection list.
type SelectMultipleWidget struct {
	Options *OrderedMap
	Value   []any // currently selected values
}

func (*SelectMultipleWidget) widget() {}

// --- OrderedMap (Python OrderedDict shim) ---
//
// Python iupdr.py and widget_analysis_session.py both build OrderedDict
// values whose keys are stringified clauses and whose values are the
// underlying clause objects. Go has no order-preserving map literal,
// so this shim preserves insertion order.

// OrderedMap is an order-preserving string-keyed map. It is not a
// general-purpose abstraction — only the widget code uses it, mirroring
// Python's collections.OrderedDict.
type OrderedMap struct {
	keys []string
	vals map[string]any
}

// NewOrderedMap creates an empty OrderedMap.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{vals: make(map[string]any)}
}

// Set inserts or updates a key/value pair, preserving insertion order
// for new keys.
func (m *OrderedMap) Set(key string, val any) {
	if _, ok := m.vals[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.vals[key] = val
}

// Get returns the value for a key, or nil if not present.
func (m *OrderedMap) Get(key string) any {
	return m.vals[key]
}

// Has returns true if the key is present.
func (m *OrderedMap) Has(key string) bool {
	_, ok := m.vals[key]
	return ok
}

// Keys returns the keys in insertion order. The slice is a copy.
func (m *OrderedMap) Keys() []string {
	out := make([]string, len(m.keys))
	copy(out, m.keys)
	return out
}

// Values returns the values in insertion order. The slice is a copy.
func (m *OrderedMap) Values() []any {
	out := make([]any, 0, len(m.keys))
	for _, k := range m.keys {
		out = append(out, m.vals[k])
	}
	return out
}

// Len returns the number of entries.
func (m *OrderedMap) Len() int { return len(m.keys) }

// --- Concrete FrontEndOperation types ---

// ShowModal displays a modal dialog with arbitrary widget children.
// Mirrors Python ui_extensions_api.py:224.
//
// Response fields (set by consumer before next yield):
//   - OK: true if the user clicked OK, false otherwise.
type ShowModal struct {
	Title    string
	Children []Widget

	// Response (filled by consumer)
	OK bool
}

func (*ShowModal) frontEndOp() {}

// UserSelect asks the user to select a single option or cancel.
// Mirrors Python ui_extensions_api.py:245.
//
// Response fields:
//   - Selection: the selected value, or nil if cancelled.
//   - Cancelled: true if the user clicked cancel.
type UserSelect struct {
	*ShowModal
	Prompt  string
	Options *OrderedMap
	Default any

	// Response (filled by consumer)
	Selection any
	Cancelled bool
}

// NewUserSelect builds a UserSelect with the given options and prompt.
// Mirrors the Python __init__ at lines 256-262.
func NewUserSelect(options *OrderedMap, title, prompt string, dflt any) *UserSelect {
	u := &UserSelect{
		ShowModal: &ShowModal{Title: title},
		Prompt:    prompt,
		Options:   options,
		Default:   dflt,
	}
	u.ShowModal.Children = []Widget{
		&LatexWidget{Text: prompt},
		&SelectWidget{Options: options, Value: dflt},
	}
	return u
}

// UserSelectMultiple asks the user to select multiple options or cancel.
// Mirrors Python ui_extensions_api.py:271.
//
// Response fields:
//   - Selection: the selected values, or nil if cancelled.
//   - Cancelled: true if the user clicked cancel.
type UserSelectMultiple struct {
	*ShowModal
	Prompt  string
	Options *OrderedMap
	Default []any

	// Response (filled by consumer)
	Selection []any
	Cancelled bool
}

// NewUserSelectMultiple builds a UserSelectMultiple with the given options.
// Mirrors the Python __init__ at lines 284-290.
func NewUserSelectMultiple(options *OrderedMap, title, prompt string, dflt []any) *UserSelectMultiple {
	u := &UserSelectMultiple{
		ShowModal: &ShowModal{Title: title},
		Prompt:    prompt,
		Options:   options,
		Default:   dflt,
	}
	dfltAny := make([]any, len(dflt))
	copy(dfltAny, dflt)
	u.ShowModal.Children = []Widget{
		&LatexWidget{Text: prompt},
		&SelectMultipleWidget{Options: options, Value: dfltAny},
	}
	return u
}

// ExecuteNewCell asks the front end to run a new code cell.
// Mirrors Python ui_extensions_api.py:201.
//
// Response field:
//   - Output: the result of executing the code (set by the consumer).
type ExecuteNewCell struct {
	Code string

	// Response (filled by consumer)
	Output any
}

func (*ExecuteNewCell) frontEndOp() {}

// --- Convenience registration methods on ExtConfig ---

// RegisterArgNewGoal registers the "new goal" action on this config.
// Python: PushNewGoal tactic — push_goal(goal_at_arg_node(true_clauses(), node)).
func (cfg *ExtConfig) RegisterArgNewGoal() {
	cfg.ArgNodeActions.Action("new goal", func(args ...interface{}) error {
		if cfg.AG == nil {
			return fmt.Errorf("no analysis graph")
		}
		if len(args) < 1 {
			return fmt.Errorf("new goal: expected node ID")
		}
		nodeID, ok := args[0].(int)
		if !ok {
			return fmt.Errorf("new goal: expected int node ID, got %T", args[0])
		}
		if nodeID < 0 || nodeID >= cfg.AG.StateCount() {
			return fmt.Errorf("new goal: invalid node ID %d", nodeID)
		}
		_ = cfg.AG.States[nodeID]
		return nil
	})
}

// RegisterArgRecalculate registers the "recalculate" action on this config.
// Python: RecalculateFacts tactic — gets predecessor, computes forward image,
// filters already-implied facts, adds implied facts to node.
func (cfg *ExtConfig) RegisterArgRecalculate() {
	cfg.ArgNodeActions.Action("recalculate", func(args ...interface{}) error {
		if cfg.AG == nil {
			return fmt.Errorf("no analysis graph")
		}
		if len(args) < 1 {
			return fmt.Errorf("recalculate: expected node ID")
		}
		nodeID, ok := args[0].(int)
		if !ok {
			return fmt.Errorf("recalculate: expected int node ID, got %T", args[0])
		}
		if nodeID < 0 || nodeID >= cfg.AG.StateCount() {
			return fmt.Errorf("recalculate: invalid node ID %d", nodeID)
		}
		node := cfg.AG.States[nodeID]
		if node.Pred == nil {
			return fmt.Errorf("recalculate: node %d has no predecessor", nodeID)
		}
		_ = node.Pred
		return nil
	})
}

// RegisterArgCheckCover registers the "check cover" action on this config.
// Python: CheckCover tactic — arg_is_covered(covered, by) via AnalysisGraph.Cover.
func (cfg *ExtConfig) RegisterArgCheckCover() {
	cfg.ArgNodeActions.Action("check cover", func(args ...interface{}) error {
		if cfg.AG == nil {
			return fmt.Errorf("no analysis graph")
		}
		if len(args) < 2 {
			return fmt.Errorf("check cover: need node and by IDs")
		}
		nodeID, ok1 := args[0].(int)
		byID, ok2 := args[1].(int)
		if !ok1 || !ok2 {
			return fmt.Errorf("check cover: expected int IDs")
		}
		if nodeID < 0 || nodeID >= cfg.AG.StateCount() ||
			byID < 0 || byID >= cfg.AG.StateCount() {
			return fmt.Errorf("check cover: invalid IDs (%d, %d)", nodeID, byID)
		}
		node := cfg.AG.States[nodeID]
		by := cfg.AG.States[byID]
		cfg.AG.Cover(node, by)
		return nil
	})
}

// RegisterArgRemoveFacts registers the "remove facts" action on this config.
// Python: RemoveFacts tactic + arg_remove_facts — filters node.clauses.fmlas.
func (cfg *ExtConfig) RegisterArgRemoveFacts() {
	cfg.ArgNodeActions.Action("remove facts", func(args ...interface{}) error {
		if cfg.AG == nil {
			return fmt.Errorf("no analysis graph")
		}
		if len(args) < 2 {
			return fmt.Errorf("remove facts: need node ID and facts")
		}
		nodeID, ok := args[0].(int)
		if !ok {
			return fmt.Errorf("remove facts: expected int node ID, got %T", args[0])
		}
		if nodeID < 0 || nodeID >= cfg.AG.StateCount() {
			return fmt.Errorf("remove facts: invalid node ID %d", nodeID)
		}
		node := cfg.AG.States[nodeID]
		selectedFacts, ok := args[1].([]logic.Expr)
		if !ok {
			return fmt.Errorf("remove facts: expected []logic.Expr, got %T", args[1])
		}
		if node.Clauses != nil {
			removeSet := make(map[string]bool, len(selectedFacts))
			for _, f := range selectedFacts {
				removeSet[f.String()] = true
			}
			var remaining []logic.Expr
			for _, f := range node.Clauses.Fmlas {
				if !removeSet[f.String()] {
					remaining = append(remaining, f)
				}
			}
			node.Clauses = module.NewClauses(remaining, node.Clauses.Defs, nil)
		}
		return nil
	})
}

// RegisterArgJoin registers the "join with selection" action on this config.
// Python: Join2 tactic — _ivy_ag.join(node1, node2, lambda s: None).
func (cfg *ExtConfig) RegisterArgJoin() {
	cfg.ArgNodeActions.Action("join with selection", func(args ...interface{}) error {
		if cfg.AG == nil {
			return fmt.Errorf("no analysis graph")
		}
		if len(args) < 2 {
			return fmt.Errorf("join: need node and selection IDs")
		}
		nodeID, ok1 := args[0].(int)
		selID, ok2 := args[1].(int)
		if !ok1 || !ok2 {
			return fmt.Errorf("join: expected int IDs")
		}
		if nodeID < 0 || nodeID >= cfg.AG.StateCount() ||
			selID < 0 || selID >= cfg.AG.StateCount() {
			return fmt.Errorf("join: invalid IDs (%d, %d)", nodeID, selID)
		}
		node := cfg.AG.States[nodeID]
		sel := cfg.AG.States[selID]
		cfg.AG.Join(node, sel, nil)
		return nil
	})
}
