package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
)

func TestMiniLutExportWideCutsAndPolarity(t *testing.T) {
	var records strings.Builder
	for _, width := range []int{2, 5, 6, 7, 8, 10, 16} {
		for _, mixed := range []bool{false, true} {
			t.Run(fmt.Sprintf("width%d/mixed%t", width, mixed), func(t *testing.T) {
				g := Gia_ManStart(1)
				inputs := make([]int32, width+1)
				for i := range inputs {
					inputs[i] = Gia_ManAppendCi(g)
				}
				head := inputs[0]
				for i := 1; i < width; i++ {
					if i%2 == 0 {
						head = Gia_ManAppendXorReal(g, head, inputs[i])
					} else {
						head = Gia_ManAppendAnd(g, head, inputs[i]^1)
					}
				}
				root := Gia_ManAppendAnd(g, head, inputs[width])
				Gia_ManAppendCo(g, head^1) // complemented head is also a cut leaf
				Gia_ManAppendCo(g, root^1)
				Gia_ManAppendCo(g, root^1) // share an inverter in mixed mode
				if mixed {
					Gia_ManAppendCo(g, root)
				}
				Gia_ManAppendCo(g, inputs[width]^1)
				Gia_ManAppendCo(g, inputs[width]^1) // inverted PI also shares one LUT
				Gia_ManAppendCo(g, 0)
				Gia_ManAppendCo(g, 1)
				Gia_ManSetRegNum(g, 1)
				g.VMapping = Vec_IntStart(int(g.nObjs))
				Vec_IntWriteEntry(g.VMapping, int(head>>1), int32(Vec_IntSize(g.VMapping)))
				Vec_IntPush(g.VMapping, int32(width))
				for _, lit := range inputs[:width] {
					Vec_IntPush(g.VMapping, lit>>1)
				}
				Vec_IntPush(g.VMapping, head>>1)
				Vec_IntWriteEntry(g.VMapping, int(root>>1), int32(Vec_IntSize(g.VMapping)))
				for _, entry := range []int32{2, head >> 1, inputs[width] >> 1, root >> 1} {
					Vec_IntPush(g.VMapping, entry)
				}
				mini := Gia_ManToMiniLut(g)
				if !miniaig.Mini_LutCheck(mini) || miniaig.Mini_LutSize(mini) != width || miniaig.Mini_LutRegNum(mini) != 1 {
					t.Fatal("invalid MiniLUT graph, width or register count")
				}
				// Digest every exported pin and truth word, avoiding enormous
				// diagnostics for 16-input tables in the native comparison.
				digest := uint64(14695981039346656037)
				for id := 0; id < miniaig.Mini_LutNodeNum(mini); id++ {
					for k := 0; k < width; k++ {
						digest = (digest ^ uint64(uint32(miniaig.Mini_LutNodeFanin(mini, id, k)))) * 1099511628211
					}
					for _, word := range miniaig.Mini_LutNodeTruth(mini, id) {
						digest = (digest ^ uint64(word)) * 1099511628211
					}
				}
				fmt.Fprintf(&records, "%d %t %d %d %d %016x\n", width, mixed, miniaig.Mini_LutNodeNum(mini), miniaig.Mini_LutSize(mini), miniaig.Mini_LutRegNum(mini), digest)
				luts := 0
				for id := 2; id < miniaig.Mini_LutNodeNum(mini); id++ {
					if miniaig.Mini_LutNodeIsNode(mini, id) {
						luts++
					}
				}
				wantLUTs := 3 // two mapped nodes and one shared PI inverter
				if mixed {
					wantLUTs++
				}
				if luts != wantLUTs {
					t.Fatalf("%d LUTs, want %d (shared polarity converters)", luts, wantLUTs)
				}
				values := make([]bool, miniaig.Mini_LutNodeNum(mini))
				for mask := 0; mask < 1<<(width+1); mask++ {
					// Interpret exported records independently, including their
					// reordered cut pins and uint32 truth-word boundaries.
					values[0], values[1] = false, true
					pi, po := 0, 0
					hv := mask&1 != 0
					for i := 1; i < width; i++ {
						v := mask&(1<<uint(i)) != 0
						if i%2 == 0 {
							hv = hv != v
						} else {
							hv = hv && !v
						}
					}
					tail := mask&(1<<uint(width)) != 0
					rv := hv && tail
					want := []bool{!hv, !rv, !rv}
					if mixed {
						want = append(want, rv)
					}
					want = append(want, !tail, !tail, false, true)
					for id := 2; id < len(values); id++ {
						switch {
						case miniaig.Mini_LutNodeIsPi(mini, id):
							values[id] = mask&(1<<uint(pi)) != 0
							pi++
						case miniaig.Mini_LutNodeIsPo(mini, id):
							values[id] = values[miniaig.Mini_LutNodeFanin(mini, id, 0)]
							if po >= len(want) || values[id] != want[po] {
								t.Fatalf("assignment %x output %d: got %t, expectations %v", mask, po, values[id], want)
							}
							po++
						default:
							index := 0
							for k := 0; k < width; k++ {
								fanin := miniaig.Mini_LutNodeFanin(mini, id, k)
								if fanin == miniaig.MINI_LUT_NULL {
									break
								}
								if values[fanin] {
									index |= 1 << uint(k)
								}
							}
							truth := miniaig.Mini_LutNodeTruth(mini, id)
							values[id] = truth[index/32]&(uint32(1)<<uint(index%32)) != 0
						}
					}
					if pi != width+1 || po != len(want) {
						t.Fatal("changed interface counts")
					}
				}
			})
		}
	}
	if t.Failed() {
		return
	}
	exe := compileGiaSourceOracle(t, "mini_lut_export.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C MiniLUT export: %v: %s", err, want)
	}
	if records.String() != string(want) {
		t.Fatalf("Go MiniLUT records:\n%sC MiniLUT records:\n%s", records.String(), want)
	}
}
