package ivyutils

import (
	"fmt"
	"sync"
)

// Registry holds all registered parameters by key.
var Registry = &ParameterRegistry{
	params: make(map[string]*Parameter),
}

// ParameterRegistry stores parameters by name.
type ParameterRegistry struct {
	mu     sync.RWMutex
	params map[string]*Parameter
}

// Get retrieves a parameter by key.
func (r *ParameterRegistry) Get(key string) (*Parameter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.params[key]
	return p, ok
}

// Register adds a parameter to the registry.
func (r *ParameterRegistry) Register(p *Parameter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.params[p.Key] = p
}

// Parameter holds a named value with optional check and process functions.
type Parameter struct {
	Key      string
	Value    any
	Check    func(any) bool
	Process  func(any) any
	Callback func(any)
}

// NewParameter creates and registers a parameter.
func NewParameter(key string, initVal any) *Parameter {
	p := &Parameter{
		Key:      key,
		Value:    initVal,
		Check:    func(v any) bool { return true },
		Process:  func(v any) any { return v },
		Callback: func(v any) {},
	}
	Registry.Register(p)
	return p
}

// NewParameterWithOpts creates a parameter with custom check and process functions.
func NewParameterWithOpts(key string, initVal any, check func(any) bool, process func(any) any) *Parameter {
	p := &Parameter{
		Key:      key,
		Value:    initVal,
		Check:    check,
		Process:  process,
		Callback: func(v any) {},
	}
	Registry.Register(p)
	return p
}

// Get returns the current value.
func (p *Parameter) Get() any {
	return p.Value
}

// GetBool returns the value as bool, defaulting to false.
func (p *Parameter) GetBool() bool {
	if b, ok := p.Value.(bool); ok {
		return b
	}
	return false
}

// GetString returns the value as string, defaulting to "".
func (p *Parameter) GetString() string {
	if s, ok := p.Value.(string); ok {
		return s
	}
	return ""
}

// Set sets the value after validation.
func (p *Parameter) Set(newVal any) error {
	if !p.Check(newVal) {
		return fmt.Errorf("bad parameter value: %s=%v", p.Key, newVal)
	}
	p.Value = p.Process(newVal)
	p.Callback(p.Value)
	return nil
}

// SetCallback sets a function called after value changes.
func (p *Parameter) SetCallback(cb func(any)) {
	p.Callback = cb
}

// NewBooleanParameter creates a parameter accepting "true"/"false" strings.
func NewBooleanParameter(key string, initVal bool) *Parameter {
	return NewParameterWithOpts(key, initVal,
		func(v any) bool {
			s, ok := v.(string)
			return ok && (s == "true" || s == "false")
		},
		func(v any) any {
			return v.(string) == "true"
		},
	)
}

// NewEnumeratedParameter creates a parameter accepting only specified values.
func NewEnumeratedParameter(key string, vals []string, initVal string) *Parameter {
	valSet := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		valSet[v] = struct{}{}
	}
	return NewParameterWithOpts(key, initVal,
		func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			_, found := valSet[s]
			return found
		},
		func(v any) any { return v },
	)
}

// Parameterize temporarily sets parameter values. Call Restore() to revert.
type Parameterize struct {
	oldValues map[string]any
}

// NewParameterize sets new parameter values and saves old ones.
func NewParameterize(values map[string]any) (*Parameterize, error) {
	p := &Parameterize{
		oldValues: make(map[string]any),
	}
	for key, val := range values {
		param, ok := Registry.Get(key)
		if !ok {
			return nil, fmt.Errorf("parameter %s undefined", key)
		}
		p.oldValues[key] = param.Get()
		if err := param.Set(val); err != nil {
			// Restore already-set values
			for rKey, rVal := range p.oldValues {
				if rKey != key {
					if rp, ok := Registry.Get(rKey); ok {
						rp.Value = rVal
					}
				}
			}
			return nil, err
		}
	}
	return p, nil
}

// Restore reverts parameters to their saved values.
func (p *Parameterize) Restore() {
	for key, val := range p.oldValues {
		if param, ok := Registry.Get(key); ok {
			param.Value = val
		}
	}
}

// SetParameters permanently sets multiple parameters from a map.
func SetParameters(values map[string]any) error {
	for key, val := range values {
		param, ok := Registry.Get(key)
		if !ok {
			return fmt.Errorf("parameter %s undefined", key)
		}
		if err := param.Set(val); err != nil {
			return err
		}
	}
	return nil
}
