package abc

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

func timBoxWords(b TimBox) [9]int {
	black := 0
	if b.Black {
		black = 1
	}
	return [9]int{b.ID, b.TravID, b.FirstInput, b.Inputs, b.FirstOutput, b.Outputs, b.DelayTable, b.Copy, black}
}
func TestTimBoxesNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_boxes.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Counts [3]int
		Boxes  [][9]int
		CI, CO [][6]int64
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 72 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		m, _ := NewTimManager(10, 12, 0)
		for _, b := range row.Boxes {
			id, err := m.AddBox(TimBox{FirstInput: b[2], Inputs: b[3], FirstOutput: b[4], Outputs: b[5], DelayTable: b[6], Black: b[8] != 0}, TimBoxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			m.SetBoxCopy(id, b[7])
		}
		m.InitPIArrivalAll(1.25)
		m.InitPORequiredAll(7.5)
		m.IncrementTraversal()
		m.MarkBox(0, false, false)
		m.MarkBox(0, true, true)
		m.MarkBox(1, false, true)
		m.MarkBox(1, true, false)
		pi, po := m.PrimaryCounts()
		black, _, _ := m.BlackBoxCounts()
		if [3]int{pi, po, black} != row.Counts {
			t.Fatal(index, pi, po, black)
		}
		for i, want := range row.Boxes {
			actual, err := m.Box(i)
			if err != nil || timBoxWords(actual) != want {
				t.Fatal(index, actual, want, err)
			}
		}
		for i, want := range row.CI {
			actual, _ := m.CI(i)
			if timTerminalWords(actual) != want {
				t.Fatal(index, "CI", i, actual, want)
			}
		}
		for i, want := range row.CO {
			actual, _ := m.CO(i)
			if timTerminalWords(actual) != want {
				t.Fatal(index, "CO", i, actual, want)
			}
		}
	}
}
func TestTimBoxBoundsCloneAndUnsupported(t *testing.T) {
	m, _ := NewTimManager(4, 4, 8)
	spec := TimBox{FirstInput: 1, Inputs: 2, FirstOutput: 1, Outputs: 2, DelayTable: -1}
	if _, err := m.AddBox(spec, TimBoxOptions{MaxPins: 4, MaxBoxes: 1}); err != nil {
		t.Fatal(err)
	}
	before := timSnapshot(m)
	if _, err := m.AddBox(spec, TimBoxOptions{MaxBoxes: 1}); err == nil || !slices.Equal(before, timSnapshot(m)) {
		t.Fatal("partial box write", err)
	}
	for _, bad := range []TimBox{{Inputs: 0, Outputs: 1}, {Inputs: 1, Outputs: 0}, {FirstInput: -1, Inputs: 1, Outputs: 1}, {FirstInput: 4, Inputs: 1, Outputs: 1}, {FirstOutput: 4, Inputs: 1, Outputs: 1}} {
		if _, err := m.AddBox(bad, TimBoxOptions{}); err == nil {
			t.Fatal(bad)
		}
	}
	m.IncrementTraversal()
	if _, err := m.CIArrival(1); err == nil {
		t.Fatal("boxed propagation silently accepted")
	}
	if _, err := m.CORequired(1); err == nil {
		t.Fatal("boxed required silently accepted")
	}
	clone := m.Clone()
	clone.SetBoxCopy(0, 99)
	b, _ := m.Box(0)
	if b.Copy != 0 {
		t.Fatal("box clone alias")
	}
	clone.MarkBox(0, true, true)
	if m.cis[1].TravID != 0 {
		t.Fatal("terminal clone alias")
	}
	if m.BoxForCI(-1) != -1 || m.BoxForCO(4) != -1 || m.BoxForCI(1) != 0 {
		t.Fatal("box lookup")
	}
	m.epoch = 0
	m.MarkBox(0, false, true)
	if m.cos[1].TravID != -1 {
		t.Fatal("previous zero label")
	}
	if _, err := m.FindBoxFromCI(-1); err == nil {
		t.Fatal("negative CI lookup")
	}
	var missing *TimManager
	if _, err := missing.FindBoxFromCI(0); err == nil {
		t.Fatal("nil CI lookup")
	}
}

func TestTimBoxFindNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_box_find.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows   int
		SHA256 string
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	rows := 0
	for boxes := 0; boxes <= 4; boxes++ {
		for pis := 0; pis <= 2; pis++ {
			for seed := 0; seed < 4; seed++ {
				outputs, total := make([]int, boxes), pis
				for box := 0; box < boxes; box++ {
					outputs[box] = 1 + (seed+2*box)%3
					total += outputs[box]
				}
				if total == 0 {
					continue
				}
				m, err := NewTimManager(total, boxes, 0)
				if err != nil {
					t.Fatal(err)
				}
				first := pis
				for box := 0; box < boxes; box++ {
					if _, err = m.AddBox(TimBox{FirstInput: box, Inputs: 1, FirstOutput: first, Outputs: outputs[box], DelayTable: -1}, TimBoxOptions{}); err != nil {
						t.Fatal(err)
					}
					first += outputs[box]
				}
				for query := 0; query < total; query++ {
					result, err := m.FindBoxFromCI(query)
					if err != nil {
						t.Fatal(err)
					}
					fmt.Fprintf(&output, "%d,%d,%d,%d,%d,%d\n", boxes, pis, seed, total, query, result)
					rows++
				}
			}
		}
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(output.String())))
	if rows != fixture.Rows || digest != fixture.SHA256 {
		t.Fatal(rows, digest, fixture)
	}
}
func FuzzTimBoxAssociations(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{255, 255, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 48 {
			data = data[:48]
		}
		m, _ := NewTimManager(8, 8, 16)
		ciBox, coBox := make([]int, 8), make([]int, 8)
		ciPos, coPos := make([]int, 8), make([]int, 8)
		for i := range ciBox {
			ciBox[i] = -1
			coBox[i] = -1
			ciPos[i] = -1
			coPos[i] = -1
		}
		for i := 0; i+2 < len(data); i += 3 {
			ins, outs := 1+int(data[i]%3), 1+int(data[i+1]%3)
			firstIn, firstOut := int(data[i+2])%(9-ins), int(data[i])%(9-outs)
			id, err := m.AddBox(TimBox{FirstInput: firstIn, Inputs: ins, FirstOutput: firstOut, Outputs: outs}, TimBoxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for j := 0; j < ins; j++ {
				coBox[firstIn+j] = id
				coPos[firstIn+j] = j
			}
			for j := 0; j < outs; j++ {
				ciBox[firstOut+j] = id
				ciPos[firstOut+j] = j
			}
		}
		for i := 0; i < 8; i++ {
			ci, _ := m.CI(i)
			co, _ := m.CO(i)
			if ci.Box != ciBox[i] || ci.Position != ciPos[i] || co.Box != coBox[i] || co.Position != coPos[i] {
				t.Fatal("association model", i, ci, co)
			}
		}
	})
}
