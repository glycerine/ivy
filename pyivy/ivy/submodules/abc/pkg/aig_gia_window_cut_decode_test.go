package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaWindowCutDecodeNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_window_cut_decode.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Leaves, Nodes []GiaID
		Cuts          []GiaWindowCut
		Selected      []int
		Window        [][]GiaID
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 144 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		actual, err := DecodeGiaWindowCuts(row.Leaves, row.Nodes, row.Cuts, row.Selected, GiaLUTAnalysisOptions{})
		if err != nil || !equalGiaIDLevels(actual, row.Window) {
			t.Fatal(index, actual, row.Window, err)
		}
	}
}
func TestGiaWindowDecodeToTiming(t *testing.T) {
	m, n, nodes, expected := buildGiaUnitWindowTest(t)
	cuts := []GiaWindowCut{{Node: 0, Leaves: [2]uint64{3, 0}}, {Node: 1, Leaves: [2]uint64{12, 0}}, {Node: 2, Nodes: [2]uint64{3, 0}}, {Node: 3, Leaves: [2]uint64{1, 0}, Nodes: [2]uint64{4, 0}}, {Node: 4, Nodes: [2]uint64{10, 0}}}
	window, err := DecodeGiaWindowCuts([]GiaID{1, 2, 3, 4}, nodes, cuts, []int{0, 1, 2, 3, 4}, GiaLUTAnalysisOptions{})
	if err != nil || !equalGiaIDLevels(window, expected) {
		t.Fatal(window, err)
	}
	r, err := n.EvaluateUnitWindow(m, nodes, window, []GiaID{9}, 0, GiaLUTAnalysisOptions{})
	if err != nil || r.Accepted || !slices.Equal(r.Clause, []SatLit{9, 7, 5, 1}) {
		t.Fatal(r, err)
	}
}
func TestGiaWindowCutDecodeRejectsAtomically(t *testing.T) {
	leaves, nodes := []GiaID{1, 2}, []GiaID{3}
	good := GiaWindowCut{Node: 0, Leaves: [2]uint64{3, 0}}
	cuts := []GiaWindowCut{good, {Node: -1}}
	for _, selected := range [][]int{nil, {-1}, {2}, {0, 1}} {
		if result, err := DecodeGiaWindowCuts(leaves, nodes, cuts, selected, GiaLUTAnalysisOptions{}); result != nil || err == nil {
			t.Fatal(result, err)
		}
	}
	invalid := good
	invalid.Leaves[1] = 1
	if result, err := DecodeGiaWindowCuts(leaves, nodes, []GiaWindowCut{invalid}, []int{0}, GiaLUTAnalysisOptions{}); result != nil || err == nil {
		t.Fatal(result, err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxObjects: 4}, {MaxVisits: 2}, {MaxObjects: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		if result, err := DecodeGiaWindowCuts(leaves, nodes, []GiaWindowCut{good}, []int{0}, options); result != nil || err == nil {
			t.Fatal(result, err)
		}
	}
	result, err := DecodeGiaWindowCuts(leaves, nodes, []GiaWindowCut{good}, []int{0}, GiaLUTAnalysisOptions{MaxObjects: 5, MaxVisits: 3})
	if err != nil {
		t.Fatal(err)
	}
	result[0][0] = 99
	if leaves[0] != 1 {
		t.Fatal("input alias")
	}
	calls := 0
	if result, err := DecodeGiaWindowCuts(leaves, nodes, []GiaWindowCut{good}, []int{0}, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 3 }}); result != nil || err == nil {
		t.Fatal(result, err)
	}
}
func FuzzGiaWindowCutDecode(f *testing.F) {
	f.Add(uint8(0), uint64(0))
	f.Add(uint8(255), uint64(0xffffffffffffffff))
	f.Add(uint8(3), uint64(0x123456789abcdef))
	f.Fuzz(func(t *testing.T, flags uint8, seed uint64) {
		leaves := make([]GiaID, int(seed%129))
		nodes := make([]GiaID, 1+int((seed>>8)%128))
		for i := range leaves {
			leaves[i] = GiaID(1000 + i)
		}
		for i := range nodes {
			nodes[i] = GiaID(2000 + i)
		}
		next := func() uint64 { seed ^= seed << 13; seed ^= seed >> 7; seed ^= seed << 17; return seed }
		cuts := make([]GiaWindowCut, 3)
		for i := range cuts {
			cuts[i].Node = int(next() % uint64(len(nodes)))
			cuts[i].Leaves = [2]uint64{next(), next()}
			cuts[i].Nodes = [2]uint64{next(), next()}
			for bit := len(leaves); bit < 128; bit++ {
				cuts[i].Leaves[bit/64] &= ^(uint64(1) << uint(bit%64))
			}
			for bit := len(nodes); bit < 128; bit++ {
				cuts[i].Nodes[bit/64] &= ^(uint64(1) << uint(bit%64))
			}
		}
		if flags&128 != 0 {
			cuts[1].Leaves[1] |= uint64(1) << 63
		}
		selected := []int{0, 1, 2, int(flags % 3)}
		expected := make([][]GiaID, len(nodes))
		invalid := false
		for _, id := range selected {
			cut := cuts[id]
			values := make([]GiaID, 0)
			for category := 0; category < 2; category++ {
				source, mask := leaves, cut.Leaves
				if category == 1 {
					source, mask = nodes, cut.Nodes
				}
				for bit := 0; bit < 128; bit++ {
					if mask[bit/64]>>(uint(bit)%64)&1 == 0 {
						continue
					}
					if bit >= len(source) {
						invalid = true
						continue
					}
					values = append(values, source[bit])
				}
			}
			expected[cut.Node] = values
		}
		options := GiaLUTAnalysisOptions{}
		limited := flags&64 != 0
		if limited {
			options.MaxVisits = 1
		}
		result, err := DecodeGiaWindowCuts(leaves, nodes, cuts, selected, options)
		if invalid || limited {
			if err == nil || result != nil {
				t.Fatal("invalid partial decode", result, err)
			}
		} else if err != nil || !equalGiaIDLevels(result, expected) {
			t.Fatal("bit-scan model", result, expected, err)
		}
	})
}
