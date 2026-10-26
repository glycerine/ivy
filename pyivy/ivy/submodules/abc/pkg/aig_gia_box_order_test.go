package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaBoxOrderNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_box_order.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Outputs1, Outputs2, Valid, NativeDefined int
		Order                                    []GiaID
		Objects                                  [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		m := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			switch obj[0] {
			case 1:
				_, err = m.AppendCI()
			case 2:
				_, err = m.AppendCO(GiaLit(obj[1]))
			case 3:
				_, err = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			case 5:
				_, err = m.AppendBuffer(GiaLit(obj[1]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		timing, _ := NewTimManager(len(m.cis), len(m.cos), 0)
		timing.AddBox(TimBox{Inputs: 2, FirstOutput: 2, Outputs: row.Outputs1}, TimBoxOptions{})
		timing.AddBox(TimBox{FirstInput: 2, Inputs: 2, FirstOutput: 2 + row.Outputs1, Outputs: row.Outputs2}, TimBoxOptions{})
		before := slices.Clone(m.objects)
		timeBefore := timSnapshot(timing)
		order, err := m.OrderWithBoxes(timing, GiaLUTAnalysisOptions{})
		if row.Valid != 0 {
			if err != nil || !slices.Equal(order, row.Order) {
				t.Fatal(index, order, row.Order, err)
			}
			checkGiaBoxOrderDependencies(t, m, timing, order)
		} else if err == nil || order != nil {
			t.Fatal("bad box order accepted", index, order, err)
		}
		if !slices.Equal(before, m.objects) || !slices.Equal(timeBefore, timSnapshot(timing)) {
			t.Fatal("source changed")
		}
	}
}
func checkGiaBoxOrderDependencies(t *testing.T, m *GiaManager, timing *TimManager, order []GiaID) {
	t.Helper()
	positions := make([]int, m.ObjectCount())
	for i := range positions {
		positions[i] = -1
	}
	for i, id := range order {
		if positions[id] >= 0 {
			t.Fatal("duplicate", id)
		}
		positions[id] = i
	}
	for i, id := range order {
		obj := m.objects[id]
		if obj.IsAnd() || obj.IsCO() {
			children, count, err := m.crossCutFanins(id)
			if err != nil {
				t.Fatal(err)
			}
			for _, child := range children[:count] {
				if positions[child] < 0 || positions[child] >= i {
					t.Fatal("fanin after node", child, id)
				}
			}
		}
	}
	for _, box := range timing.boxes {
		for i := 0; i < box.Outputs; i++ {
			ci := m.cis[box.FirstOutput+i]
			for j := 0; j < box.Inputs; j++ {
				co := m.cos[box.FirstInput+j]
				if positions[co] >= positions[ci] {
					t.Fatal("box output precedes input")
				}
			}
		}
	}
}
func TestGiaBoxOrderLimitsAndDepth(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	for i := 0; i < 10000; i++ {
		root, _ = m.AppendAnd(root, a)
	}
	m.AppendCO(root)
	timing, _ := NewTimManager(2, 1, 0)
	order, err := m.OrderWithBoxes(timing, GiaLUTAnalysisOptions{MaxVisits: m.ObjectCount()})
	if err != nil || len(order) != m.ObjectCount() {
		t.Fatal(len(order), err)
	}
	checkGiaBoxOrderDependencies(t, m, timing, order)
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: m.ObjectCount() - 1}, {MaxObjects: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if order, err := m.OrderWithBoxes(timing, options); order != nil || err == nil {
			t.Fatal(order, err)
		}
	}
	calls := 0
	if order, err := m.OrderWithBoxes(timing, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 10 }}); order != nil || err == nil {
		t.Fatal(order, err)
	}
	m.AppendCI()
	if m.IsNormalized() {
		t.Fatal("interleaved CI accepted")
	}
	if _, err := m.OrderWithBoxes(timing, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("unnormalized graph")
	}
}
func FuzzGiaBoxOrder(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{255, 0, 17, 42})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 25 {
			data = data[:25]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		timing, _ := NewTimManager(len(m.cis), len(m.cos), 0)
		order, err := m.OrderWithBoxes(timing, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkGiaBoxOrderDependencies(t, m, timing, order)
		// Independent reachability set: all CIs, constant, COs and structural cones.
		reached := make([]bool, m.ObjectCount())
		reached[0] = true
		for _, id := range m.cis {
			reached[id] = true
		}
		for _, id := range m.cos {
			reached[id] = true
			driver, _ := m.objects[id].Fanin0(id)
			stack := []GiaID{driver.ID()}
			for len(stack) > 0 {
				next := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if reached[next] {
					continue
				}
				reached[next] = true
				children, n, _ := m.crossCutFanins(next)
				stack = append(stack, children[:n]...)
			}
		}
		actual := make([]bool, m.ObjectCount())
		for _, id := range order {
			actual[id] = true
		}
		if !slices.Equal(reached, actual) {
			t.Fatal("reachability set")
		}
	})
}

func TestGiaBoxOrderTimingLayout(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	m.AppendCI()
	m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	m.AppendCO(root)
	m.AppendCO(a)
	timing, _ := NewTimManager(4, 2, 0)
	timing.AddBox(TimBox{Inputs: 1, FirstOutput: 2, Outputs: 1}, TimBoxOptions{})
	timing.AddBox(TimBox{FirstInput: 0, Inputs: 1, FirstOutput: 3, Outputs: 1}, TimBoxOptions{})
	if order, err := m.OrderWithBoxes(timing, GiaLUTAnalysisOptions{}); order != nil || err == nil {
		t.Fatal("overlapping staged input ranges", order, err)
	}
	missing, _ := NewTimManager(3, 2, 0)
	if _, err := m.OrderWithBoxes(missing, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("terminal counts")
	}
	var absent *GiaManager
	if absent.IsNormalized() {
		t.Fatal("nil normalized")
	}
	if _, err := absent.OrderWithBoxes(timing, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("nil graph")
	}
}
