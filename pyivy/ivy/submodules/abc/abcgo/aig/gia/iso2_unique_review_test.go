package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func iso2ReviewInputs(count int) *Gia_Man_t {
	graph := Gia_ManStart(int32(count + 2))
	for range count {
		Gia_ManAppendCi(graph)
	}
	literal := int32(0)
	if count > 0 {
		literal = 2
	}
	Gia_ManAppendCo(graph, literal)
	return graph
}

func TestIso2UniqifyAgainstScalarClassification(t *testing.T) {
	for _, count := range []int{0, 1, 2, 3, 8, 17, 64, 257} {
		for pattern := 0; pattern < 4; pattern++ {
			for _, marked := range []bool{false, true} {
				t.Run(fmt.Sprintf("count%d/pattern%d/marked%t", count, pattern, marked), func(t *testing.T) {
					graph := iso2ReviewInputs(count)
					p := Gia_Iso2ManStart(graph)
					defer Gia_Iso2ManStop(p)
					values := make([]uint32, count)
					state := uint32(count + 1)
					for i := range values {
						state = state*1664525 + 1013904223
						switch pattern {
						case 0:
							values[i] = 0
						case 1:
							values[i] = uint32(i % 5)
						case 2:
							values[i] = uint32(count-i) * 3
						default:
							values[i] = state
							if i%4 == 0 {
								values[i] = 0x80000000
							} else if i%7 == 0 {
								values[i] = 0xffffffff
							}
						}
						obj := Gia_ManCi(graph, int32(i))
						obj.Value, obj.fMark1 = values[i], uint32(i%2)
						if marked && i%11 == 0 {
							obj.fMark0 = 1
						}
					}
					var wantTied []int32
					var singletons []int
					for i, value := range values {
						// A quadratic equality oracle uses no hash table or
						// production classifier, including for value zero.
						tied := marked && i%11 == 0
						for j, other := range values {
							if i != j && value == other {
								tied = true
							}
						}
						if tied {
							wantTied = append(wantTied, int32(i+1))
						} else {
							// Independent insertion order in the existing safe
							// signed-32 comparator domain. C's subtracting
							// comparator can overflow for wide signatures.
							at := len(singletons)
							singletons = append(singletons, i)
							for at > 0 && int32(values[singletons[at-1]]) > int32(value) {
								singletons[at] = singletons[at-1]
								at--
							}
							singletons[at] = i
						}
					}
					labels := make([]int32, count)
					for i := range labels {
						labels[i] = -1
					}
					wantValues := slices.Clone(values)
					for i, input := range singletons {
						label := int32(i + 1)
						labels[input] = label
						wantValues[input] += s_256Primes[label&255]
					}
					if got := Gia_Iso2ManUniqify(p); int(got) != len(singletons) || p.nUniques != int32(len(singletons)+1) {
						t.Fatalf("singleton count%d labels%d want%d", got, p.nUniques, len(singletons))
					}
					if !slices.Equal(vec.Vec_IntArray(p.vTied), wantTied) {
						t.Fatalf("tied order%v want%v", vec.Vec_IntArray(p.vTied), wantTied)
					}
					for i, want := range wantValues {
						obj := Gia_ManCi(graph, int32(i))
						if obj.Value != want || obj.fMark0 != 0 || obj.fMark1 != uint32(i%2) || vec.Vec_IntEntry(p.vUniques, i+1) != labels[i] {
							t.Fatalf("candidate%d value%08x want%08x label%d want%d marks%d/%d", i, obj.Value, want, vec.Vec_IntEntry(p.vUniques, i+1), labels[i], obj.fMark0, obj.fMark1)
						}
					}
				})
			}
		}
	}
}

func TestIso2UniqifyConsumesOnlyCandidateMarks(t *testing.T) {
	graph := iso2ReviewInputs(8)
	p := Gia_Iso2ManStart(graph)
	defer Gia_Iso2ManStop(p)
	for i, value := range []uint32{8, 8, 1, 9, 9, 7, 11, 1} {
		Gia_ManCi(graph, int32(i)).Value = value
	}
	// C uses mark0 as tied-candidate scratch and consumes every marked
	// candidate, including one marked before entry. Other objects are untouched.
	Gia_ManCi(graph, 5).fMark0 = 1
	Gia_ManConst0(graph).fMark0 = 1
	Gia_ManCo(graph, 0).fMark0 = 1
	if count := Gia_Iso2ManUniqify(p); count != 1 {
		t.Fatalf("source singleton count%d want1", count)
	}
	if !slices.Equal(vec.Vec_IntArray(p.vTied), []int32{1, 2, 3, 4, 5, 6, 8}) {
		t.Fatalf("compacted tied order%v", vec.Vec_IntArray(p.vTied))
	}
	for i := int32(0); i < 8; i++ {
		obj := Gia_ManCi(graph, i)
		label := int32(-1)
		if i == 6 {
			label = 1
		}
		if obj.fMark0 != 0 || vec.Vec_IntEntry(p.vUniques, int(Gia_ObjId(graph, obj))) != label {
			t.Fatalf("candidate%d mark%d label%d want%d", i, obj.fMark0, vec.Vec_IntEntry(p.vUniques, int(Gia_ObjId(graph, obj))), label)
		}
	}
	if Gia_ManConst0(graph).fMark0 != 1 || Gia_ManCo(graph, 0).fMark0 != 1 {
		t.Fatal("uniquification consumed marks outside its candidate vector")
	}
	if p.nUniques != 2 || Gia_ManCi(graph, 6).Value != 11+s_256Primes[1] {
		t.Fatal("source label counter/signature salt changed")
	}
}

func TestIso2UniqifyReusesOwnedStorage(t *testing.T) {
	for _, singletons := range []bool{false, true} {
		t.Run(fmt.Sprintf("singletons%t", singletons), func(t *testing.T) {
			graph := iso2ReviewInputs(64)
			p := Gia_Iso2ManStart(graph)
			defer Gia_Iso2ManStop(p)
			resetAndRun := func() {
				vec.Vec_IntClear(p.vTied)
				p.nUniques = 1
				for i := int32(0); i < 64; i++ {
					obj := Gia_ManCi(graph, i)
					id := Gia_ObjId(graph, obj)
					obj.Value, obj.fMark0 = 42, 0
					if singletons {
						obj.Value = uint32(64-i) * 3
					}
					vec.Vec_IntWriteEntry(p.vUniques, int(id), -1)
					vec.Vec_IntPush(p.vTied, id)
				}
				want := int32(0)
				if singletons {
					want = 64
				}
				if got := Gia_Iso2ManUniqify(p); got != want {
					t.Fatalf("singletons%d want%d", got, want)
				}
			}
			resetAndRun()
			if allocs := testing.AllocsPerRun(100, resetAndRun); allocs != 0 {
				t.Fatalf("warmed uniquification allocates%g times, want0", allocs)
			}
		})
	}
}

func BenchmarkIso2ManagerAllocation(b *testing.B) {
	graph := iso2ReviewInputs(64)
	b.ReportAllocs()
	for b.Loop() {
		p := Gia_Iso2ManStart(graph)
		Gia_Iso2ManStop(p)
	}
}
