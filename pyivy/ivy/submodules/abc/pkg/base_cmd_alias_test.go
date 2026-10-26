package abc

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestCommandAliasesOwnershipLookupFormatAndOrder(t *testing.T) {
	aliases, err := NewCommandAliases(3)
	if err != nil {
		t.Fatal(err)
	}
	first := []string{"read", "one.aig"}
	if err = aliases.Add("r", first); err != nil {
		t.Fatal(err)
	}
	first[0] = "caller"
	if err = aliases.Add("p", []string{"print_stats"}); err != nil {
		t.Fatal(err)
	}
	if got, err := aliases.Lookup("r"); err != nil || got != "read" {
		t.Fatal(got, err)
	}
	if got, err := aliases.Lookup("unknown"); err != nil || got != "unknown" {
		t.Fatal(got, err)
	}
	alias, found, err := aliases.Alias("r")
	if err != nil || !found || !slices.Equal(alias.Args, []string{"read", "one.aig"}) {
		t.Fatal(alias, found, err)
	}
	alias.Args[0] = "changed"
	again, _, _ := aliases.Alias("r")
	if again.Args[0] != "read" {
		t.Fatal("alias accessor aliases storage")
	}
	formatted, err := FormatCommandAlias(again)
	if err != nil || formatted != "r               read one.aig\n" {
		t.Fatal(formatted, err)
	}
	want := []CommandAlias{{Name: "r", Args: []string{"read", "one.aig"}}, {Name: "p", Args: []string{"print_stats"}}}
	if got := aliases.Entries(); !reflect.DeepEqual(got, want) {
		t.Fatal(got, want)
	}
	entries := aliases.Entries()
	entries[0].Args[0] = "changed"
	if got, _ := aliases.Lookup("r"); got != "read" {
		t.Fatal("entries alias storage")
	}
	if deleted, err := aliases.Delete("r"); err != nil || !deleted {
		t.Fatal(deleted, err)
	}
	if err = aliases.Add("r", []string{"read_new"}); err != nil {
		t.Fatal(err)
	}
	want = []CommandAlias{{Name: "p", Args: []string{"print_stats"}}, {Name: "r", Args: []string{"read_new"}}}
	if got := aliases.Entries(); !reflect.DeepEqual(got, want) {
		t.Fatal(got, want)
	}
}

func TestCommandAliasesErrorsAndCapacity(t *testing.T) {
	if aliases, err := NewCommandAliases(-1); aliases != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(aliases, err)
	}
	aliases, err := NewCommandAliases(1)
	if err != nil {
		t.Fatal(err)
	}
	if err = aliases.Add("", []string{"x"}); err != nil {
		t.Fatal("empty but non-null name", err)
	}
	before := aliases.Entries()
	for _, test := range []struct {
		name string
		args []string
		want error
	}{{"empty", nil, ErrDimension}, {"", []string{"replacement"}, ErrInvariant}, {"full", []string{"x"}, ErrLimit}} {
		if err = aliases.Add(test.name, test.args); !errors.Is(err, test.want) || !reflect.DeepEqual(aliases.Entries(), before) {
			t.Fatal(test, err, aliases.Entries())
		}
	}
	if _, err = FormatCommandAlias(CommandAlias{Name: "x"}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	var absent *CommandAliases
	if err = absent.Add("x", []string{"x"}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = absent.Lookup("x"); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = absent.Delete("x"); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzCommandAliases(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 1, 2}, {0, 1, 2, 0, 3, 4, 1, 1, 0}, {2, 7, 9}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 300 {
			return
		}
		aliases, err := NewCommandAliases(16)
		if err != nil {
			t.Fatal(err)
		}
		var model []CommandAlias
		find := func(name string) int {
			for i := range model {
				if model[i].Name == name {
					return i
				}
			}
			return -1
		}
		for i := 0; i+2 < len(data); i += 3 {
			name := string([]byte{data[i+1] % 8})
			index := find(name)
			switch data[i] % 3 {
			case 0:
				args := []string{string([]byte{data[i+2]})}
				err = aliases.Add(name, args)
				switch {
				case index >= 0:
					if !errors.Is(err, ErrInvariant) {
						t.Fatal("duplicate", err)
					}
				case len(model) >= 16:
					if !errors.Is(err, ErrLimit) {
						t.Fatal("limit", err)
					}
				default:
					if err != nil {
						t.Fatal(err)
					}
					model = append(model, CommandAlias{Name: name, Args: args})
				}
			case 1:
				deleted, deleteErr := aliases.Delete(name)
				if deleteErr != nil || deleted != (index >= 0) {
					t.Fatal(deleted, index, deleteErr)
				}
				if index >= 0 {
					model = append(model[:index], model[index+1:]...)
				}
			case 2:
				got, lookupErr := aliases.Lookup(name)
				want := name
				if index >= 0 {
					want = model[index].Args[0]
				}
				if lookupErr != nil || got != want {
					t.Fatal(got, want, lookupErr)
				}
			}
			if got := aliases.Entries(); !reflect.DeepEqual(got, model) && !(len(got) == 0 && len(model) == 0) {
				t.Fatal(got, model)
			}
		}
	})
}
