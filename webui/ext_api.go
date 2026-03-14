package webui

// Full port of ui_extensions_api.py — extension point system for
// registering callbacks that add actions to the verification UI.

import (
	"fmt"
	"sync"
)

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

// --- Standard extension points (Python: module-level declarations) ---

// ArgNodeActions is the extension point for ARG node context menu actions
// (Python: arg_node_actions).
var ArgNodeActions = NewExtensionPoint("arg_node_actions")

// GoalNodeActions is the extension point for proof goal node context menu actions
// (Python: goal_node_actions).
var GoalNodeActions = NewExtensionPoint("goal_node_actions")

// --- Default extension registrations ---

func init() {
	// Register execute_actions: lists all available actions for an ARG node
	// (Python: @arg_node_actions.register def execute_actions(s)).
	ArgNodeActions.Register(func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error) {
		// Stub: real implementation returns sorted list of actions from
		// analysis_state.ivy_ag.actions.
		return []ExtensionAction{
			{Label: "execute_actions (stub)", Callback: nil},
		}, nil
	})

	// Register try_conjectures: lists conjectures that are not yet implied
	// (Python: @arg_node_actions.register def try_conjectures(s)).
	ArgNodeActions.Register(func(ctx interface{}, args ...interface{}) ([]ExtensionAction, error) {
		// Stub: real implementation checks each conjecture against
		// the background theory and state clauses.
		return []ExtensionAction{
			{Label: "try_conjectures (stub)", Callback: nil},
		}, nil
	})
}

// --- Front-end operation types (Python: FrontEndOperation hierarchy) ---

// FrontEndOperation represents an async operation with the front end.
type FrontEndOperation interface {
	Submit(onDone func(result interface{}))
}

// InteractionError represents an error during user interaction.
type InteractionError struct {
	Message string
}

func (e *InteractionError) Error() string {
	return e.Message
}

// --- Convenience registration functions ---

// RegisterArgNewGoal registers the "new goal" action
// (Python: @arg_node_actions.action('new goal')).
func RegisterArgNewGoal() {
	ArgNodeActions.Action("new goal", func(args ...interface{}) error {
		// Stub: push_new_goal(true_clauses(), arg_node(node.id))
		return nil
	})
}

// RegisterArgRecalculate registers the "recalculate" action
// (Python: @arg_node_actions.action('recalculate')).
func RegisterArgRecalculate() {
	ArgNodeActions.Action("recalculate", func(args ...interface{}) error {
		// Stub: recalculate_facts(node, arg_get_conjuncts(arg_get_pred(node)))
		return nil
	})
}

// RegisterArgCheckCover registers the "check cover" action
// (Python: @arg_node_actions.action('check cover')).
func RegisterArgCheckCover() {
	ArgNodeActions.Action("check cover", func(args ...interface{}) error {
		// Stub: check_cover(arg_node(node.id), arg_node(by.id))
		return nil
	})
}

// RegisterArgRemoveFacts registers the "remove facts" action
// (Python: @arg_node_actions.action('remove facts')).
func RegisterArgRemoveFacts() {
	ArgNodeActions.Action("remove facts", func(args ...interface{}) error {
		// Stub: remove_facts(arg_node(node.id), *selected_facts)
		return nil
	})
}

// RegisterArgJoin registers the "join with selection" action
// (Python: @arg_node_actions.action('join with selection')).
func RegisterArgJoin() {
	ArgNodeActions.Action("join with selection", func(args ...interface{}) error {
		// Stub: join2(arg_node(node.id), arg_node(selection.id))
		return nil
	})
}
