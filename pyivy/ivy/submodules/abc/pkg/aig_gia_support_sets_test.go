package abc

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"
)

func sameGiaSupports(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !slices.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
func TestGiaSupportSetsNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, c)
	m.AppendCO(y.Not())
	m.AppendCO(GiaConst0)
	m.AppendCO(b)
	raw, err := os.ReadFile("test_vectors/oracle/gia_support_sets.json")
	if err != nil {
		t.Fatal(err)
	}
	var native struct{ Nodes, Outputs [][]int }
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	nodes := []GiaID{0, a.ID(), b.ID(), c.ID(), x.ID(), y.ID()}
	got, err := m.NodeSupports(nodes, 0, 0)
	if err != nil || !sameGiaSupports(got, native.Nodes) {
		t.Fatal(got, err)
	}
	got, err = m.OutputSupports(0, 0)
	if err != nil || !sameGiaSupports(got, native.Outputs) {
		t.Fatal(got, err)
	}
	repeated, err := m.NodeSupports([]GiaID{x.ID(), x.ID()}, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	repeated[0][0] = 99
	if repeated[1][0] != 0 {
		t.Fatal("alias")
	}
}
func TestGiaSupportSetsExtendedLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	buf, _ := m.AppendBuffer(x)
	mux, _ := m.AppendMuxReal(c, buf, a)
	co, _ := m.AppendCO(mux)
	before := slices.Clone(m.objects)
	got, err := m.NodeSupports([]GiaID{buf.ID(), mux.ID(), co.ID()}, 0, 0)
	if err != nil || !reflect.DeepEqual(got, [][]int{{0, 1}, {0, 1, 2}, {0, 1, 2}}) || !slices.Equal(before, m.objects) {
		t.Fatal(got, err)
	}
	for _, limits := range [][2]int{{-1, 0}, {0, -1}, {m.ObjectCount() - 1, 100}, {100, 2}, {100, 7}} {
		got, err := m.NodeSupports([]GiaID{mux.ID()}, limits[0], limits[1])
		if got != nil || err == nil {
			t.Fatal(limits, got, err)
		}
	}
	if got, err := m.NodeSupports([]GiaID{InvalidGiaID}, 0, 0); got != nil || err == nil {
		t.Fatal(got, err)
	}
	var nilManager *GiaManager
	if _, err := nilManager.OutputSupports(0, 0); err == nil {
		t.Fatal("nil")
	}
}
