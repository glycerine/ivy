package abc

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCommandHistoryAdmissionDeduplicationAndFormatting(t *testing.T) {
	history, err := NewCommandHistory(CommandHistoryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"", "abc", "set x", "unset x", "timer", "quit now", "alias x y", "source abc.rc", "source ..\\abc.rc", "history 2", "hi 2", "hi", "query?"} {
		if added, err := history.Add(command); err != nil || added {
			t.Fatal("excluded command", command, added, err)
		}
	}
	for index := 0; index < 12; index++ {
		command := "command-" + string(rune('a'+index))
		if added, err := history.Add(command + "\n"); err != nil || !added {
			t.Fatal(index, added, err)
		}
	}
	if added, err := history.Add("command-c"); err != nil || added {
		t.Fatal("recent duplicate", added, err)
	}
	wantTail := []string{"command-k", "command-l", "command-c"}
	if got, _ := history.Entries(3); !slices.Equal(got, wantTail) {
		t.Fatal(got, wantTail)
	}
	// command-a is now older than the last ten and is appended again.
	if added, err := history.Add("command-a"); err != nil || !added {
		t.Fatal("old duplicate", added, err)
	}
	formatted, err := history.Format(2)
	if err != nil || formatted != "================== Command history ==================\ncommand-c\ncommand-a\n=====================================================\n" {
		t.Fatal(formatted, err)
	}
	copyEntries, _ := history.Entries(100)
	copyEntries[0] = "caller"
	again, _ := history.Entries(100)
	if again[0] == "caller" {
		t.Fatal("entries alias history")
	}
}

func TestCommandHistoryLoadFilesLimitsAndErrors(t *testing.T) {
	history, err := NewCommandHistory(CommandHistoryOptions{MaxEntries: 3, MaxCommandBytes: 5})
	if err != nil {
		t.Fatal(err)
	}
	if err = history.Load([]byte("one\ntwo\n")); err != nil {
		t.Fatal(err)
	}
	if entries, _ := history.Entries(3); !slices.Equal(entries, []string{"one", "two"}) {
		t.Fatal(entries)
	}
	if err = history.Load([]byte("new\n")); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "history.txt")
	if err = history.WriteFile(path, 1); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "two\n" {
		t.Fatal(string(data), err)
	}
	loaded, _ := NewCommandHistory(CommandHistoryOptions{MaxEntries: 3, MaxCommandBytes: 5})
	if err = loaded.ReadFile(path); err != nil {
		t.Fatal(err)
	}
	if entries, _ := loaded.Entries(3); !slices.Equal(entries, []string{"two"}) {
		t.Fatal(entries)
	}
	for _, options := range []CommandHistoryOptions{{Batch: -1}, {Batch: 2}, {MaxEntries: -1}, {MaxCommandBytes: -1}} {
		if got, err := NewCommandHistory(options); got != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(options, got, err)
		}
	}
	limited, _ := NewCommandHistory(CommandHistoryOptions{MaxEntries: 1, MaxCommandBytes: 4})
	if _, err = limited.Add("longer"); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = limited.Add("four"); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.Add("next"); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	before, _ := limited.Entries(1)
	if err = (&CommandHistory{maxEntries: 2, maxCommandBytes: 3}).Load([]byte("okay\n")); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ := limited.Entries(1)
	if !slices.Equal(before, after) {
		t.Fatal("failed operation changed history")
	}
	batch, _ := NewCommandHistory(CommandHistoryOptions{Batch: 1})
	if added, err := batch.Add("ordinary command"); err != nil || added {
		t.Fatal(added, err)
	}
	var absent *CommandHistory
	if _, err = absent.Add("x"); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = absent.Entries(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = history.Entries(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzCommandHistory(f *testing.F) {
	for _, seed := range [][]byte{{}, {1, 2, 3, 4}, {9, 9, 9, 9, 9}, {255, 0, 7, 3}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 200 {
			return
		}
		history, err := NewCommandHistory(CommandHistoryOptions{MaxEntries: 64, MaxCommandBytes: 16})
		if err != nil {
			t.Fatal(err)
		}
		var model []string
		for index, value := range data {
			command := "run-" + string([]byte{'a' + value%12})
			added, addErr := history.Add(command)
			start := max(0, len(model)-10)
			found := -1
			for i := start; i < len(model); i++ {
				if model[i] == command {
					found = i
					break
				}
			}
			wantAdded := found < 0 && len(model) < 64
			if found >= 0 {
				entry := model[found]
				model = append(model[:found], model[found+1:]...)
				model = append(model, entry)
			} else if len(model) < 64 {
				model = append(model, command)
			}
			if addErr != nil && !(len(model) == 64 && !wantAdded && errors.Is(addErr, ErrLimit)) {
				t.Fatal(index, addErr)
			}
			if addErr == nil && added != wantAdded {
				t.Fatal(index, added, wantAdded)
			}
			got, entriesErr := history.Entries(64)
			if entriesErr != nil || !slices.Equal(got, model) {
				t.Fatal(index, got, model, entriesErr)
			}
		}
	})
}
