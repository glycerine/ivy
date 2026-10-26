package abc

// Named command flags from src/base/cmd/cmdFlag.c.

import "github.com/glycerine/insdict"

type CommandFlag struct {
	Name, Value string
}

// CommandFlags is the concrete owned flag table used by the future command
// frame. insdict preserves deterministic insertion order for diagnostics and
// command-state serialization.
type CommandFlags struct {
	values     *insdict.Dict[string, string]
	maxEntries int
}

func NewCommandFlags(maxEntries int) (*CommandFlags, error) {
	if maxEntries < 0 {
		return nil, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	return &CommandFlags{values: insdict.NewDictSize[string, string](min(maxEntries, 1024)), maxEntries: maxEntries}, nil
}

func (f *CommandFlags) Read(name string) (string, bool, error) {
	if f == nil || f.values == nil {
		return "", false, ErrInvariant
	}
	value, found := f.values.Get2(name)
	return value, found, nil
}

// Update replaces a flag value without changing its insertion position. A nil
// value means the empty string, matching Cmd_FlagUpdateValue's null-value path.
func (f *CommandFlags) Update(name string, value *string) error {
	if f == nil || f.values == nil {
		return ErrInvariant
	}
	text := ""
	if value != nil {
		text = *value
	}
	if _, found := f.values.Get2(name); !found && f.values.Len() >= f.maxEntries {
		return ErrLimit
	}
	f.values.Set(name, text)
	return nil
}

func (f *CommandFlags) Delete(name string) (bool, error) {
	if f == nil || f.values == nil {
		return false, ErrInvariant
	}
	return f.values.Del(name), nil
}

func (f *CommandFlags) Entries() []CommandFlag {
	if f == nil || f.values == nil {
		return nil
	}
	result := make([]CommandFlag, 0, f.values.Len())
	for name, value := range f.values.All() {
		result = append(result, CommandFlag{Name: name, Value: value})
	}
	return result
}

func (f *CommandFlags) Len() int {
	if f == nil || f.values == nil {
		return 0
	}
	return f.values.Len()
}
