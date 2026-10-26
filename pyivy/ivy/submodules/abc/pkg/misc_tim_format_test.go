package abc

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestTimStatsAndFormatting(t *testing.T) {
	m := makeTimRebuildManager(t)
	stats, err := m.Stats(17)
	if err != nil {
		t.Fatal(err)
	}
	want := TimManagerStats{
		CIs: 8, COs: 8, PIs: 2, POs: 2, Boxes: 4,
		BlackBoxes: 2, BlackInputs: 2, BlackOutputs: 4, And2Delay: 17,
		Kinds: []TimBoxKindStats{
			{DelayTable: 0, Count: 1, Inputs: 2, Outputs: 1},
			{DelayTable: 1, Count: 1, Inputs: 1, Outputs: 2},
			{DelayTable: 2, Count: 1, Inputs: 2, Outputs: 1},
			{DelayTable: 3, Count: 1, Inputs: 1, Outputs: 2},
		},
	}
	if !reflect.DeepEqual(stats, want) {
		t.Fatal(stats, want)
	}
	stats.Kinds[0].Count = 99
	again, _ := m.Stats(17)
	if again.Kinds[0].Count != 1 {
		t.Fatal("stats aliases manager")
	}
	copies, err := m.FormatBoxCopies()
	if err != nil || copies != "-1 2 -1 0\n" {
		t.Fatal(copies, err)
	}
	formatted, err := m.Format()
	if err != nil {
		t.Fatal(err)
	}
	formattedAgain, _ := m.Format()
	if formatted != formattedAgain || !strings.HasPrefix(formatted, "TIMING MANAGER PI=2 CI=8 PO=2 CO=8 Box=4\n") || strings.Count(formatted, "\n") != 25 {
		t.Fatal(formatted)
	}
	if digest := fmt.Sprintf("%x", sha256.Sum256([]byte(formatted))); digest != "ad2832ef5ecdc26448aeac14375f05cb7552623986de452ca528436e34582aa7" {
		t.Fatal("format changed", digest)
	}
	before := timManagerWords(m)
	m.Format()
	m.FormatBoxCopies()
	m.Stats(0)
	if !reflect.DeepEqual(before, timManagerWords(m)) {
		t.Fatal("diagnostics mutate manager")
	}
}

func TestTimFormattingEmptyAndErrors(t *testing.T) {
	m, _ := NewTimManager(0, 0, 1)
	formatted, err := m.Format()
	if err != nil || formatted != "TIMING MANAGER PI=0 CI=0 PO=0 CO=0 Box=0\n" {
		t.Fatal(formatted, err)
	}
	if copies, err := m.FormatBoxCopies(); err != nil || copies != "\n" {
		t.Fatal(copies, err)
	}
	var missing *TimManager
	if _, err := missing.Stats(0); err == nil {
		t.Fatal("nil stats")
	}
	if value, err := missing.Format(); err == nil || value != "" {
		t.Fatal(value, err)
	}
	if value, err := missing.FormatBoxCopies(); err == nil || value != "" {
		t.Fatal(value, err)
	}
}

func FuzzTimFormatting(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(3))
	f.Add([]byte{255, 0, 17}, uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) == 0 {
			return
		}
		boxes := int(flags & 3)
		m, _ := NewTimManager(boxes+1, boxes+1, 16)
		for box := 0; box < boxes; box++ {
			m.AddBox(TimBox{FirstInput: box, Inputs: 1, FirstOutput: box + 1, Outputs: 1, DelayTable: int(int8(data[box%len(data)])), Black: data[(box+1)%len(data)]&1 != 0}, TimBoxOptions{})
			m.SetBoxCopy(box, int(int8(data[(box+2)%len(data)])))
		}
		before := timManagerWords(m)
		first, err := m.Format()
		if err != nil {
			t.Fatal(err)
		}
		second, _ := m.Format()
		if first != second || !reflect.DeepEqual(before, timManagerWords(m)) {
			t.Fatal("nondeterministic or mutating format")
		}
		stats, err := m.Stats(int(int8(flags)))
		if err != nil || stats.Boxes != boxes || stats.PIs != 1 || stats.POs != 1 {
			t.Fatal(stats, err)
		}
		for i := 1; i < len(stats.Kinds); i++ {
			if stats.Kinds[i-1].DelayTable >= stats.Kinds[i].DelayTable {
				t.Fatal("unsorted kinds", stats.Kinds)
			}
		}
	})
}
