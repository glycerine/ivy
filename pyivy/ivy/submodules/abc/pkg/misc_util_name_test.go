package abc

// Standalone tests for src/misc/util/utilNam.c and utilNam.h behavior.

import (
	"errors"
	"slices"
	"strconv"
	"testing"
)

func TestStringNameTableCore(t *testing.T) {
	table, err := NewStringNameTable(StringNameTableOptions{InitialNames: 2, AverageNameBytes: 4, MaxNames: 8, MaxBytes: 64})
	if err != nil {
		t.Fatal(err)
	}
	for index, name := range []string{"alpha", "beta", "gamma"} {
		id, found, err := table.Intern(name)
		if err != nil || found || id != index+1 {
			t.Fatal(index, id, found, err)
		}
	}
	if id, found, err := table.Intern("beta"); err != nil || !found || id != 2 {
		t.Fatal(id, found, err)
	}
	if id, found, err := table.Find("gamma"); err != nil || !found || id != 3 {
		t.Fatal(id, found, err)
	}
	if id, found, err := table.Find("delta"); err != nil || found || id != 0 {
		t.Fatal(id, found, err)
	}
	if id, found, err := table.InternPrefix("alphabet", 5); err != nil || !found || id != 1 {
		t.Fatal(id, found, err)
	}
	if id, found, err := table.FindPrefix("betamax", 4); err != nil || !found || id != 2 {
		t.Fatal(id, found, err)
	}
	if name, found, err := table.Name(0); err != nil || found || name != "" {
		t.Fatal(name, found, err)
	}
	if name, found, err := table.Name(3); err != nil || !found || name != "gamma" {
		t.Fatal(name, found, err)
	}
	names, err := table.Names()
	if err != nil || !slices.Equal(names, []string{"alpha", "beta", "gamma"}) {
		t.Fatal(names, err)
	}
	names[0] = "changed"
	if name, _, _ := table.Name(1); name != "alpha" {
		t.Fatal(name)
	}
	if count, err := table.ObjectCount(); err != nil || count != 4 {
		t.Fatal(count, err)
	}
	if stats, err := table.Stats(); err != nil || stats != (StringNameTableStats{Names: 3, StoredBytes: 21}) {
		t.Fatal(stats, err)
	}
	text, err := table.Render()
	if err != nil || text != "       1 = alpha\n       2 = beta\n       3 = gamma\n" {
		t.Fatal(text, err)
	}
}

func TestStringNameTableMappingReportsAndLifecycle(t *testing.T) {
	first, _ := NewStringNameTable(StringNameTableOptions{})
	second, _ := NewStringNameTable(StringNameTableOptions{})
	for _, name := range []string{"red", "green", "blue", "amber"} {
		_, _, _ = first.Intern(name)
	}
	for _, name := range []string{"blue", "red", "black"} {
		_, _, _ = second.Intern(name)
	}
	mapping, err := first.IDMap(second)
	if err != nil || !slices.Equal(mapping, []int{0, 2, 0, 1, 0}) {
		t.Fatal(mapping, err)
	}
	identity, err := first.IDMap(first)
	if err != nil || !slices.Equal(identity, []int{0, 1, 2, 3, 4}) {
		t.Fatal(identity, err)
	}
	if count, err := first.CommonCount([]int{4, 1, 3, 3}, second); err != nil || count != 3 {
		t.Fatal(count, err)
	}
	if name, found, err := first.FirstUnique([]int{3, 2, 4}, second); err != nil || !found || name != "green" {
		t.Fatal(name, found, err)
	}
	if name, found, err := first.FirstUnique([]int{1, 3}, second); err != nil || found || name != "" {
		t.Fatal(name, found, err)
	}
	if _, err := first.CommonCount([]int{0}, second); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := first.Find("red"); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := first.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestStringNameTableLimitsAndHashes(t *testing.T) {
	for _, options := range []StringNameTableOptions{{InitialNames: -1}, {AverageNameBytes: -1}, {MaxNames: -1}, {MaxBytes: -1}, {InitialNames: 3, MaxNames: 2}, {MaxBytes: 3}} {
		if table, err := NewStringNameTable(options); table != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit)) {
			t.Fatal(options, table, err)
		}
	}
	table, _ := NewStringNameTable(StringNameTableOptions{MaxNames: 1, MaxBytes: 8})
	if _, _, err := table.Intern(""); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err := table.Intern("a\x00b"); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if id, found, err := table.Intern("abc"); err != nil || found || id != 1 {
		t.Fatal(id, found, err)
	}
	if id, found, err := table.Intern("abc"); err != nil || !found || id != 1 {
		t.Fatal(id, found, err)
	}
	if _, _, err := table.Intern("d"); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := table.FindPrefix("abc", 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	tests := []struct {
		text       string
		limit      int
		table      int
		first, jnk int
	}{
		{"abc", -1, 1009, 213, 628},
		{"alphabet", 5, 65521, 19438, 32671},
		{"a\x00b", -1, 257, 213, 19},
		{"a\x00b", 3, 257, 202, 19},
	}
	for _, test := range tests {
		first, err1 := StringNameHash(test.text, test.limit, test.table)
		jenkins, err2 := StringNameHashJenkins(test.text, test.limit, test.table)
		if err1 != nil || err2 != nil || first != test.first || jenkins != test.jnk {
			t.Fatal(test, first, jenkins, err1, err2)
		}
	}
	if _, err := StringNameHash("x", 2, 7); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzStringNameTable(f *testing.F) {
	f.Add([]byte("alpha\x00beta\x00alpha"))
	f.Add([]byte{1, 2, 3, 4, 5, 6})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			data = data[:512]
		}
		table, err := NewStringNameTable(StringNameTableOptions{MaxNames: 257, MaxBytes: 2048})
		if err != nil {
			t.Fatal(err)
		}
		modelNames := []string{""}
		modelIDs := make(map[string]int)
		for index := 0; index < len(data); {
			width := int(data[index]%12) + 1
			index++
			if width > len(data)-index {
				width = len(data) - index
			}
			nameBytes := slices.Clone(data[index : index+width])
			index += width
			for i := range nameBytes {
				nameBytes[i] = 'a' + nameBytes[i]%26
			}
			name := string(nameBytes)
			if name == "" {
				continue
			}
			expectedID, expectedFound := modelIDs[name]
			id, found, err := table.Intern(name)
			if err != nil {
				t.Fatal(err)
			}
			if !expectedFound {
				expectedID = len(modelNames)
				modelIDs[name] = expectedID
				modelNames = append(modelNames, name)
			}
			if id != expectedID || found != expectedFound {
				t.Fatal(id, found, expectedID, expectedFound)
			}
			foundID, exists, err := table.Find(name)
			if err != nil || !exists || foundID != expectedID {
				t.Fatal(foundID, exists, expectedID, err)
			}
		}
		names, err := table.Names()
		if err != nil || !slices.Equal(names, modelNames[1:]) {
			t.Fatal(names, modelNames, err)
		}
		other, _ := NewStringNameTable(StringNameTableOptions{})
		for id := 1; id < len(modelNames); id += 2 {
			_, _, _ = other.Intern(modelNames[id])
		}
		mapping, err := table.IDMap(other)
		if err != nil || len(mapping) != len(modelNames) {
			t.Fatal(mapping, err)
		}
		for id := 1; id < len(modelNames); id++ {
			expected := 0
			if id%2 == 1 {
				expected = (id + 1) / 2
			}
			if mapping[id] != expected {
				t.Fatal(strconv.Itoa(id), mapping[id], expected)
			}
		}
	})
}
