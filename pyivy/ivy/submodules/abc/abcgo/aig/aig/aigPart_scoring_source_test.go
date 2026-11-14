package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestPartitionScoreAt101SupportVariables(t *testing.T) {
	for _, size := range []int{99, 100, 101, 102} {
		support := make([]int32, size)
		for i := range support {
			support[i] = int32(i)
		}
		one := []int32{0, 200, 201, 202, 203, 204, 205, 206, 207, 208}
		bitmap := Aig_ManSuppCharStart(support, 256)
		got := Aig_ManPartitionSmartFindPart([][]int32{support}, [][]int32{{0}}, [][]uint32{bitmap}, 0, one)
		want := int32(0)
		if size >= 101 {
			want = -1 // C's Base2Log(1) is 1, so 100/2 falls below 75.
		}
		if got != want {
			t.Errorf("support size %d: partition %d, want %d", size, got, want)
		}
	}
}

func TestPartitionScoresAgainstCompiledSource(t *testing.T) {
	var input, want strings.Builder
	cases := 0
	for _, size := range []int{0, 1, 31, 32, 33, 99, 100, 101, 102, 103, 127, 128, 200} {
		for _, width := range []int{0, 1, 2, 10, 16, 32, 33, 40, 64} {
			for _, shared := range []int{0, min(3, width, size), min(6, width, size), min(width/2, size), min(width, size)} {
				for _, count := range []int{1, 2} {
					for _, limit := range []int32{-1, 0, 1, 50, 51, 100, 200} {
						supports := make([][]int32, count)
						parts := make([][]int32, count)
						bitmaps := make([][]uint32, count)
						for i := 0; i < size; i++ {
							supports[0] = append(supports[0], int32(i))
						}
						if count == 2 {
							other := 2*size + 1
							for i := 0; i < other/2; i++ {
								supports[1] = append(supports[1], int32(i))
							}
							for i := 0; i < other-other/2; i++ {
								supports[1] = append(supports[1], int32(256+i))
							}
						}
						var one []int32
						for i := 0; i < shared; i++ {
							one = append(one, int32(i))
						}
						for i := 0; i < width-shared; i++ {
							one = append(one, int32(256+i))
						}
						fmt.Fprintf(&input, "512 %d %d %d\n", count, limit, width)
						for i, support := range supports {
							parts[i] = []int32{int32(i)}
							bitmaps[i] = Aig_ManSuppCharStart(support, 512)
							fmt.Fprintf(&input, "%d", len(support))
							for _, value := range support {
								fmt.Fprintf(&input, " %d", value)
							}
							input.WriteByte('\n')
						}
						for _, value := range one {
							fmt.Fprintf(&input, "%d ", value)
						}
						input.WriteByte('\n')
						fmt.Fprintln(&want, Aig_ManPartitionSmartFindPart(supports, parts, bitmaps, limit, one))
						cases++
					}
				}
			}
		}
	}
	got := runAigSourceFixtureInput(t, "partition_score_compare.c", input.String(), "aigPart.c")
	if got != want.String() {
		c, goLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(c) && i < len(goLines); i++ {
			if c[i] != goLines[i] {
				t.Fatalf("case %d: source partition %s, Go %s", i, c[i], goLines[i])
			}
		}
		t.Fatalf("score trace lengths C %d, Go %d", len(c), len(goLines))
	}
	t.Logf("compared %d source partition choices", cases)
}

func TestSmartPartitionBoundaryAgainstCompiledSource(t *testing.T) {
	var input, want strings.Builder
	for _, size := range []int{99, 100, 101, 102, 103} {
		for _, limit := range []int32{-1, 0, 50, 100, 200} {
			g := NewGraph()
			for i := 0; i < 512; i++ {
				g.NewInput()
			}
			large, small := g.Const1(), g.inputs[0]
			for i := 0; i < size; i++ {
				large = g.AND(large, g.inputs[i])
			}
			for i := 0; i < 9; i++ {
				small = g.AND(small, g.inputs[256+i])
			}
			g.AddOutput(large)
			g.AddOutput(small)
			var supports [][]int32
			parts := Aig_ManPartitionSmart(g, limit, 0, &supports)
			fmt.Fprintf(&input, "%d %d\n", size, limit)
			fmt.Fprintf(&want, "parts %d:\n", len(parts))
			seen := map[int32]bool{}
			for i, part := range parts {
				want.WriteString("outputs")
				for _, output := range part {
					if seen[output] {
						t.Fatal("output appears in multiple partitions")
					}
					seen[output] = true
					fmt.Fprintf(&want, " %d", output)
				}
				want.WriteString(" | inputs")
				for _, id := range supports[i] {
					fmt.Fprintf(&want, " %d", id)
				}
				want.WriteByte('\n')
			}
			if len(seen) != 2 || len(parts) != len(supports) {
				t.Fatal("partition/support coverage differs")
			}
		}
	}
	got := runAigSourceFixtureInput(t, "smart_partition_compare.c", input.String(), "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigPart.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		t.Fatalf("smart partition output differs from C: source length %d, Go %d", len(got), want.Len())
	}
}
