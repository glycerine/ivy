package ivyutils

import (
	"fmt"
	"sync"
)


// NewParameterRegistry creates a new empty ParameterRegistry.
func NewParameterRegistry() *ParameterRegistry {
	return &ParameterRegistry{
		params: make(map[string]*Parameter),
	}
}

// NewParameterOn creates and registers a parameter on a specific registry.
func NewParameterOn(reg *ParameterRegistry, key string, initVal any) *Parameter {
	p := &Parameter{
		Key:      key,
		Value:    initVal,
		Check:    func(v any) bool { return true },
		Process:  func(v any) any { return v },
		Callback: func(v any) {},
	}
	reg.Register(p)
	return p
}

// NewBooleanParameterOn creates a boolean parameter on a specific registry.
func NewBooleanParameterOn(reg *ParameterRegistry, key string, initVal bool) *Parameter {
	p := &Parameter{
		Key:   key,
		Value: initVal,
		Check: func(v any) bool {
			s, ok := v.(string)
			return ok && (s == "true" || s == "false")
		},
		Process: func(v any) any {
			return v.(string) == "true"
		},
		Callback: func(v any) {},
	}
	reg.Register(p)
	return p
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

// NewParameter creates a parameter (not registered to any registry).
// Use NewParameterOn to register on a specific registry.
func NewParameter(key string, initVal any) *Parameter {
	return &Parameter{
		Key:      key,
		Value:    initVal,
		Check:    func(v any) bool { return true },
		Process:  func(v any) any { return v },
		Callback: func(v any) {},
	}
}

// NewParameterWithOpts creates a parameter with custom check and process functions.
// Not registered to any registry — use NewParameterOn for registration.
func NewParameterWithOpts(key string, initVal any, check func(any) bool, process func(any) any) *Parameter {
	return &Parameter{
		Key:      key,
		Value:    initVal,
		Check:    check,
		Process:  process,
		Callback: func(v any) {},
	}
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
	return &Parameter{
		Key:   key,
		Value: initVal,
		Check: func(v any) bool {
			s, ok := v.(string)
			return ok && (s == "true" || s == "false")
		},
		Process: func(v any) any {
			return v.(string) == "true"
		},
		Callback: func(v any) {},
	}
}

// NewEnumeratedParameter creates a parameter accepting only specified values.
func NewEnumeratedParameter(key string, vals []string, initVal string) *Parameter {
	valSet := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		valSet[v] = struct{}{}
	}
	return &Parameter{
		Key:   key,
		Value: initVal,
		Check: func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			_, found := valSet[s]
			return found
		},
		Process:  func(v any) any { return v },
		Callback: func(v any) {},
	}
}

// Parameterize temporarily sets parameter values. Call Restore() to revert.
type Parameterize struct {
	reg       *ParameterRegistry
	oldValues map[string]any
}

// NewParameterize sets new parameter values on the given registry and saves old ones.
func NewParameterize(reg *ParameterRegistry, values map[string]any) (*Parameterize, error) {
	p := &Parameterize{
		reg:       reg,
		oldValues: make(map[string]any),
	}
	for key, val := range values {
		param, ok := reg.Get(key)
		if !ok {
			return nil, fmt.Errorf("parameter %s undefined", key)
		}
		p.oldValues[key] = param.Get()
		if err := param.Set(val); err != nil {
			// Restore already-set values
			for rKey, rVal := range p.oldValues {
				if rKey != key {
					if rp, ok := reg.Get(rKey); ok {
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
		if param, ok := p.reg.Get(key); ok {
			param.Value = val
		}
	}
}

// SetParameters permanently sets multiple parameters on the given registry.
func SetParameters(reg *ParameterRegistry, values map[string]any) error {
	for key, val := range values {
		param, ok := reg.Get(key)
		if !ok {
			return fmt.Errorf("parameter %s undefined", key)
		}
		if err := param.Set(val); err != nil {
			return err
		}
	}
	return nil
}
