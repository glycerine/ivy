package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func giaSupportRankGraph(t *testing.T, sizes []int) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	inputs := make([]GiaLit, 4)
	for i := range inputs {
		var e error
		inputs[i], e = m.AppendCI()
		if e != nil {
			t.Fatal(e)
		}
	}
	roots := []GiaLit{GiaConst0, inputs[0]}
	for i := 1; i < 4; i++ {
		r, e := m.AppendAnd(roots[len(roots)-1], inputs[i])
		if e != nil {
			t.Fatal(e)
		}
		roots = append(roots, r)
	}
	for _, size := range sizes {
		if _, e := m.AppendCO(roots[size]); e != nil {
			t.Fatal(e)
		}
	}
	return m
}
func TestGiaSupportRankingNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_support_rank.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []GiaSupportRanking
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 25 {
		t.Fatal(e)
	}
	for n, row := range rows {
		sizes := make([]int, n)
		for k := range sizes {
			sizes[k] = (k*7 + n) % 5
		}
		m := giaSupportRankGraph(t, sizes)
		got, e := m.RankOutputSupports(0, 0)
		if e != nil || got.Maximum != row.Maximum || got.Largest != row.Largest || !slices.Equal(got.Order, row.Order) {
			t.Fatal(n, got, row, e)
		}
	}
	var m *GiaManager
	if _, e = m.RankOutputSupports(0, 0); e == nil {
		t.Fatal("nil")
	}
	if _, e = giaSupportRankGraph(t, []int{4}).RankOutputSupports(0, 1); e == nil {
		t.Fatal("entry limit")
	}
}
func FuzzGiaSupportRanking(f *testing.F) {
	f.Add([]byte{0, 4, 3, 1, 2, 4, 4, 0, 3, 1, 2, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		sizes := make([]int, len(data))
		for i, v := range data {
			sizes[i] = int(v % 5)
		}
		m := giaSupportRankGraph(t, sizes)
		got, e := m.RankOutputSupports(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		maximum, largest := 1, -1
		for i, size := range sizes {
			maximum = max(maximum, size)
			if largest < 0 || size > sizes[largest] {
				largest = i
			}
		}
		if got.Maximum != maximum || got.Largest != largest || len(got.Order) != len(sizes) {
			t.Fatal(got, sizes)
		}
		seen := make([]bool, len(sizes))
		previous := 5
		for _, index := range got.Order {
			if index < 0 || index >= len(sizes) || seen[index] || sizes[index] > previous {
				t.Fatal(got, sizes)
			}
			seen[index] = true
			previous = sizes[index]
		}
	})
}
