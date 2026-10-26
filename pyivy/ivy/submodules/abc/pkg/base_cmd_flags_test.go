package abc

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func commandFlagString(value string) *string { return &value }

func TestCommandFlagsUpdateReadDeleteAndOrder(t *testing.T) {
	flags, err := NewCommandFlags(3)
	if err != nil {
		t.Fatal(err)
	}
	if value, found, err := flags.Read("missing"); err != nil || found || value != "" {
		t.Fatal(value, found, err)
	}
	if err = flags.Update("alpha", commandFlagString("one")); err != nil {
		t.Fatal(err)
	}
	if err = flags.Update("beta", nil); err != nil {
		t.Fatal(err)
	}
	if err = flags.Update("gamma", commandFlagString("three")); err != nil {
		t.Fatal(err)
	}
	if err = flags.Update("alpha", commandFlagString("updated")); err != nil {
		t.Fatal(err)
	}
	want := []CommandFlag{{"alpha", "updated"}, {"beta", ""}, {"gamma", "three"}}
	if got := flags.Entries(); !reflect.DeepEqual(got, want) {
		t.Fatal(got, want)
	}
	if err = flags.Update("overflow", nil); !errors.Is(err, ErrLimit) || !reflect.DeepEqual(flags.Entries(), want) {
		t.Fatal("limit or atomicity", err, flags.Entries())
	}
	if deleted, err := flags.Delete("beta"); err != nil || !deleted {
		t.Fatal(deleted, err)
	}
	if deleted, err := flags.Delete("beta"); err != nil || deleted {
		t.Fatal(deleted, err)
	}
	if err = flags.Update("beta", commandFlagString("last")); err != nil {
		t.Fatal(err)
	}
	want = []CommandFlag{{"alpha", "updated"}, {"gamma", "three"}, {"beta", "last"}}
	if got := flags.Entries(); !reflect.DeepEqual(got, want) || flags.Len() != 3 {
		t.Fatal(got, flags.Len())
	}
	if value, found, err := flags.Read("beta"); err != nil || !found || value != "last" {
		t.Fatal(value, found, err)
	}
	copyEntries := flags.Entries()
	copyEntries[0].Value = "caller"
	if value, _, _ := flags.Read("alpha"); value != "updated" {
		t.Fatal("entries alias table")
	}
}

func TestCommandFlagsErrorsAndExactLimit(t *testing.T) {
	if flags, err := NewCommandFlags(-1); flags != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(flags, err)
	}
	flags, err := NewCommandFlags(1)
	if err != nil {
		t.Fatal(err)
	}
	if err = flags.Update("", nil); err != nil {
		t.Fatal("empty but non-null key", err)
	}
	if err = flags.Update("", commandFlagString("replacement")); err != nil {
		t.Fatal("overwrite at exact limit", err)
	}
	var absent *CommandFlags
	if _, _, err = absent.Read("x"); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = absent.Update("x", nil); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = absent.Delete("x"); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzCommandFlags(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 'a', '1'}, {0, 'a', '1', 0, 'b', '2', 1, 'a', 0}, {2, 0, 0}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 300 {
			return
		}
		flags, err := NewCommandFlags(16)
		if err != nil {
			t.Fatal(err)
		}
		var model []CommandFlag
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
				value := string([]byte{data[i+2]})
				err = flags.Update(name, &value)
				if index >= 0 {
					model[index].Value = value
				} else if len(model) < 16 {
					model = append(model, CommandFlag{name, value})
				} else if !errors.Is(err, ErrLimit) {
					t.Fatal("missing limit", err)
				}
			case 1:
				deleted, deleteErr := flags.Delete(name)
				if deleteErr != nil || deleted != (index >= 0) {
					t.Fatal(deleted, index, deleteErr)
				}
				if index >= 0 {
					model = append(model[:index], model[index+1:]...)
				}
			case 2:
				value, found, readErr := flags.Read(name)
				if readErr != nil || found != (index >= 0) || (index >= 0 && value != model[index].Value) {
					t.Fatal(value, found, index, readErr)
				}
			}
			if got := flags.Entries(); !slices.Equal(got, model) {
				t.Fatal(got, model)
			}
		}
	})
}
