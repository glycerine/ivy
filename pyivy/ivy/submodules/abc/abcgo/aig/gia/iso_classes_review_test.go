package gia

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func isoClassesReviewState(sizes []int) (*Gia_IsoMan_t, [][]int32) {
	total := 0
	for _, size := range sizes {
		total += size
	}
	g := iso2ReviewInputs(total)
	p := Gia_IsoManStart(g)
	p.nEntries, p.nUniques, p.nSingles = int32(total), 17, 13
	p.pUniques[0], p.pUniques[total+1] = 1, 2
	p.pStoreW[0], p.pStoreW[total+1] = 0x123456789abcdef0, 0xfedcba9876543210
	var groups [][]int32
	next := 1
	for _, size := range sizes {
		vec.Vec_IntPush(p.vClasses, int32(next))
		vec.Vec_IntPush(p.vClasses, int32(size))
		var group []int32
		for range size {
			p.pStoreW[next] = uint64(next)<<32 | 0xdeadbeef
			group = append(group, int32(next))
			next++
		}
		groups = append(groups, group)
	}
	vec.Vec_IntPush(p.vClasses2, -99)
	return p, groups
}

func TestIsoClassPartitionAgainstIndependentKeyGroups(t *testing.T) {
	sizes := [][]int{nil, {2}, {3}, {10}, {11}, {13}, {2, 3, 11}}
	for configuration, layout := range sizes {
		for pattern := 0; pattern < 6; pattern++ {
			t.Run(fmt.Sprintf("layout%d/pattern%d", configuration, pattern), func(t *testing.T) {
				p, groups := isoClassesReviewState(layout)
				defer Gia_IsoManStop(p)
				values := make([]uint64, p.nObjs)
				fields := make([][8]uint32, p.nObjs)
				state := uint32(31)
				for id := range values {
					state = state*1664525 + 1013904223
					value := uint32(7)
					switch pattern {
					case 1:
						value = uint32(id)
					case 2:
						if id%2 == 0 {
							value = 0xffffffff
						} else {
							value = 0
						}
					case 3:
						value = []uint32{0xffffffff, 0x80000000, 0, 7}[id%4]
					case 4:
						value = state
					case 5:
						value = uint32(id % 3)
					}
					obj := Gia_ManObj(p.pGia, int32(id))
					obj.Value = value
					obj.fMark0, obj.fMark1 = uint8(id%2), uint8((id+1)%2)
					values[id] = uint64(value)
					fields[id] = isoSignatureReviewFields(obj)
				}
				wantUnique := slices.Clone(p.pUniques)
				var wantClasses []int32
				nextLabel, singles := int32(17), int32(0)
				var refined int32
				for _, group := range groups {
					byKey := map[uint32][]int32{}
					var keys []uint32
					for _, id := range group {
						key := uint32(values[id])
						if _, found := byKey[key]; !found {
							keys = append(keys, key)
						}
						byKey[key] = append(byKey[key], id)
					}
					slices.Sort(keys) // Independent Go ordering of keys, without payload partitioning.
					if len(keys) > 1 {
						refined = 1
					}
					begin := group[0]
					for _, key := range keys {
						ids := byKey[key]
						if len(ids) == 1 {
							wantUnique[ids[0]] = nextLabel
							nextLabel++
							singles++
						} else {
							wantClasses = append(wantClasses, begin, int32(len(ids)))
						}
						begin += int32(len(ids))
					}
				}
				guards := [2]uint64{p.pStoreW[0], p.pStoreW[p.nObjs-1]}
				if got := Gia_IsoSort(p); got != refined {
					t.Fatalf("refined%d want%d", got, refined)
				}
				if !slices.Equal(vec.Vec_IntArray(p.vClasses), wantClasses) || !slices.Equal(p.pUniques, wantUnique) {
					t.Fatalf("classes%v want%v; labels%v want%v", vec.Vec_IntArray(p.vClasses), wantClasses, p.pUniques, wantUnique)
				}
				if p.nUniques != nextLabel || p.nSingles != singles || p.nEntries != p.nObjs-2-singles {
					t.Fatal("singleton/entry accounting differs")
				}
				for _, group := range groups {
					var seen []int32
					var previous uint32
					for i := range group {
						offset := group[0] + int32(i)
						id := int32(Gia_IsoGetItem(p, offset))
						key := Gia_IsoGetValue(p, offset)
						if !slices.Contains(group, id) || slices.Contains(seen, id) || key != uint32(values[id]) || i > 0 && key < previous {
							t.Fatalf("invalid packed partition at%d", offset)
						}
						seen = append(seen, id)
						previous = key
					}
					same := true
					for _, id := range group {
						same = same && values[id] == values[group[0]]
					}
					if same {
						for i, id := range group {
							if Gia_IsoGetItem(p, group[0]+int32(i)) != uint32(id) {
								t.Fatal("uniform class was unnecessarily reordered")
							}
						}
					}
				}
				if guards != [2]uint64{p.pStoreW[0], p.pStoreW[p.nObjs-1]} {
					t.Fatal("sort crossed packed class boundaries")
				}
				checkIsoSignatureReview(t, p.pGia, values, fields, "class partition")
			})
		}
	}
}

func TestIsoTieAssignmentFollowsSourcePolicies(t *testing.T) {
	cases := []struct {
		name          string
		sizes, levels []int
		mode          int
		chosen        []int
	}{
		{"highest level", []int{3, 2, 3, 2, 4}, []int{0, 1, 1, 2, 2}, 0, []int{3, 4}},
		{"first pair", []int{3, 2, 4, 2}, []int{0, 1, 2, 3}, 2, []int{1}},
		{"no pair fallback", []int{3, 4, 3}, []int{0, 1, 2}, 2, []int{2}},
		{"last pair", []int{3, 4, 2}, []int{0, 1, 2}, 3, []int{2}},
		{"last larger class", []int{2, 3, 4}, []int{0, 1, 2}, 3, []int{2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, groups := isoClassesReviewState(tc.sizes)
			defer Gia_IsoManStop(p)
			for i, group := range groups {
				for _, id := range group {
					p.pLevels[id] = int32(tc.levels[i])
				}
			}
			labels := slices.Clone(p.pUniques)
			next, count := p.nUniques, int32(0)
			for _, index := range tc.chosen {
				assigned := groups[index]
				// Source's experimental variants drop the whole larger class after
				// labeling only its first member; they are not complete tie strategies.
				if tc.mode != 0 && len(assigned) > 2 {
					assigned = assigned[:1]
				}
				for _, id := range assigned {
					labels[id] = next
					next++
					count++
				}
			}
			var wantClasses []int32
			for i, group := range groups {
				if !slices.Contains(tc.chosen, i) {
					wantClasses = append(wantClasses, group[0], int32(len(group)))
				}
			}
			packed := slices.Clone(p.pStoreW)
			isoReviewCapture(t, func() {
				switch tc.mode {
				case 0:
					Gia_IsoAssignOneClass(p, 1)
				case 2:
					Gia_IsoAssignOneClass2(p)
				case 3:
					Gia_IsoAssignOneClass3(p)
				}
			})
			if !slices.Equal(vec.Vec_IntArray(p.vClasses), wantClasses) || !slices.Equal(p.pUniques, labels) || !slices.Equal(p.pStoreW, packed) {
				t.Fatal("tie assignment changed source class/label/packed state")
			}
			total := int32(0)
			for _, size := range tc.sizes {
				total += int32(size)
			}
			if p.nUniques != next || p.nSingles != 13+count || p.nEntries != total-count {
				t.Fatal("tie assignment did not accumulate source counters")
			}
		})
	}
}

func TestIsoRefinementVerboseCompletionReports(t *testing.T) {
	for _, forward := range []int32{0, 1} {
		t.Run(fmt.Sprintf("forward%d", forward), func(t *testing.T) {
			quiet := iso2ReviewInputs(1)
			if output := isoReviewCapture(t, func() { Gia_IsoDeriveEquivPos(quiet, forward, 0) }); output != "" {
				t.Fatalf("quiet refinement produced output: %q", output)
			}
			graph := iso2ReviewInputs(1)
			got := isoReviewCapture(t, func() { Gia_IsoDeriveEquivPos(graph, forward, 1) })
			wantIterations := []int{0, 2}
			if forward != 0 {
				wantIterations = []int{0, 1, 2}
			}
			lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
			if len(lines) != len(wantIterations)+6 {
				t.Fatalf("verbose report has%d rows want%d: %q", len(lines), len(wantIterations)+6, got)
			}
			for i, iteration := range wantIterations {
				if !strings.HasPrefix(lines[i], fmt.Sprintf("Iter %4d :", iteration)) {
					t.Errorf("row%d got%q want iteration%d", i, lines[i], iteration)
				}
			}
			for i, name := range []string{"Start", "Simulate", "Refine", "Sort", "Other", "TOTAL"} {
				line := lines[len(wantIterations)+i]
				if !strings.HasPrefix(line, name) || !strings.Contains(line, " sec (") || !strings.HasSuffix(line, " %)") {
					t.Errorf("missing source timing row%s: %q", name, line)
				}
			}
		})
	}
}

func TestIsoClassPrinterKeepsSourceFormat(t *testing.T) {
	p, _ := isoClassesReviewState([]int{2, 3})
	defer Gia_IsoManStop(p)
	want := "The total of 2 classes:\n    0 : (  1,  2)  \n    1 : (  3,  3)  \n"
	if got := isoReviewCapture(t, func() { Gia_IsoPrintClasses(p) }); got != want {
		t.Fatalf("class report got%q want%q", got, want)
	}
}
