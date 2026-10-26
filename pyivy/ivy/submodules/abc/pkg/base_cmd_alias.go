package abc

// Command aliases from src/base/cmd/cmdAlias.c.

import (
	"fmt"
	"slices"
	"strings"

	"github.com/glycerine/insdict"
)

type CommandAlias struct {
	Name string
	Args []string
}

type CommandAliases struct {
	values     *insdict.Dict[string, CommandAlias]
	maxEntries int
}

func NewCommandAliases(maxEntries int) (*CommandAliases, error) {
	if maxEntries < 0 {
		return nil, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	return &CommandAliases{values: insdict.NewDictSize[string, CommandAlias](min(maxEntries, 1024)), maxEntries: maxEntries}, nil
}

// Add owns name and args. The source asserts that aliases are unique; Go
// reports duplicate names and empty expansions as checked errors.
func (a *CommandAliases) Add(name string, args []string) error {
	if a == nil || a.values == nil {
		return ErrInvariant
	}
	if len(args) == 0 {
		return ErrDimension
	}
	if _, found := a.values.Get2(name); found {
		return ErrInvariant
	}
	if a.values.Len() >= a.maxEntries {
		return ErrLimit
	}
	copyArgs := slices.Clone(args)
	a.values.Set(name, CommandAlias{Name: name, Args: copyArgs})
	return nil
}

func (a *CommandAliases) Alias(name string) (CommandAlias, bool, error) {
	if a == nil || a.values == nil {
		return CommandAlias{}, false, ErrInvariant
	}
	alias, found := a.values.Get2(name)
	if !found {
		return CommandAlias{}, false, nil
	}
	alias.Args = slices.Clone(alias.Args)
	return alias, true, nil
}

// Lookup returns an alias's first expansion token or the original command when
// no alias exists.
func (a *CommandAliases) Lookup(command string) (string, error) {
	alias, found, err := a.Alias(command)
	if err != nil || !found {
		return command, err
	}
	if len(alias.Args) == 0 {
		return "", ErrInvariant
	}
	return alias.Args[0], nil
}

func (a *CommandAliases) Delete(name string) (bool, error) {
	if a == nil || a.values == nil {
		return false, ErrInvariant
	}
	return a.values.Del(name), nil
}

func (a *CommandAliases) Entries() []CommandAlias {
	if a == nil || a.values == nil {
		return nil
	}
	result := make([]CommandAlias, 0, a.values.Len())
	for _, alias := range a.values.All() {
		alias.Args = slices.Clone(alias.Args)
		result = append(result, alias)
	}
	return result
}

func FormatCommandAlias(alias CommandAlias) (string, error) {
	if len(alias.Args) == 0 {
		return "", ErrDimension
	}
	var result strings.Builder
	_, _ = fmt.Fprintf(&result, "%-15s", alias.Name)
	for _, argument := range alias.Args {
		result.WriteByte(' ')
		result.WriteString(argument)
	}
	result.WriteByte('\n')
	return result.String(), nil
}
