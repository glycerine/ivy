package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestQuantNativeMuxCases(t *testing.T) {
	exe := compileGiaSourceOracle(t, "exist_review.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C quantification: %v\n%s", err, native)
	}
	var records strings.Builder
	for mode := 0; mode < 2; mode++ {
		for mask := 0; mask < 8; mask++ {
			for phase := int32(0); phase < 2; phase++ {
				// The source incremental variant crashes after an intermediate constant
				// when eliminating all inputs. Directed scalar tests cover the Go repair.
				if mode == 0 && mask == 0 {
					continue
				}
				g, root := quantReviewMux()
				keep := func(data any, ci int32) int32 { return int32(data.(int)>>ci) & 1 }
				var lit int32
				if mode == 0 {
					lit = Gia_ManQuantExist(g, root^phase, keep, mask)
				} else {
					lit = Gia_ManQuantExist2(g, root^phase, keep, mask)
				}
				fmt.Fprintf(&records, "%d %d %d lit%d objs%d", mode, mask, phase, lit, g.nObjs)
				for bits := 0; bits < 8; bits++ {
					values := cofReviewValues(g, bits)
					fmt.Fprintf(&records, " %d", giaBool(values[lit>>1] != (lit&1 != 0)))
				}
				records.WriteByte('\n')
				Gia_ManStop(g)
			}
		}
	}
	esopReviewCompare(t, native, records.String())
}

func TestQuantNativeSupportTables(t *testing.T) {
	exe := compileGiaSourceOracle(t, "exist_review.c")
	native, err := testoracle.Command(t, exe, "--support").CombinedOutput()
	if err != nil {
		t.Fatalf("C support tables: %v\n%s", err, native)
	}
	var records strings.Builder
	for _, n := range []int{0, 1, 2, 63, 64, 65, 127, 128, 129, 191, 192, 193} {
		for mode := 0; mode < 2; mode++ {
			g := Gia_ManStart(int32(2*n + 16))
			Gia_ManQuantSetSuppStart(g)
			for i := 0; i < n; i++ {
				id := Gia_ManAppendCi(g) >> 1
				Gia_ManQuantSetSuppZero(g)
				if mode == 0 || i%3 != 1 {
					if mode == 1 {
						Gia_ManQuantUpdateCiSupp(g, id)
					} else {
						Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, id))
					}
				}
			}
			for i := 1; i < n; i += 7 {
				Gia_ManAppendAnd(g, 2, int32(2*(i+1)))
			}
			fmt.Fprintf(&records, "support %d %d objs%d words%d count%d vars", n, mode, g.nObjs, g.nSuppWords, g.iSuppPi)
			for _, id := range Vec_IntArray(&g.vSuppVars) {
				fmt.Fprintf(&records, " %d", id)
			}
			records.WriteByte('\n')
			for id := int32(0); id < g.nObjs; id++ {
				fmt.Fprintf(&records, "row %d", id)
				for _, value := range Gia_ManQuantInfoId(g, id) {
					fmt.Fprintf(&records, " %x", value)
				}
				records.WriteByte('\n')
			}
			keep := func(data any, ci int32) int32 { return giaBool((ci+data.(int32))%5 != 0) }
			Gia_ManQuantMarkUsedCis(g, keep, int32(2))
			records.WriteString("mask")
			for _, value := range Gia_ManQuantInfoId(g, 0) {
				fmt.Fprintf(&records, " %x", value)
			}
			records.WriteByte('\n')
			for id := int32(1); id < g.nObjs; id++ {
				fmt.Fprintf(&records, "overlap %d %d\n", id, Gia_ManQuantCheckOverlap(g, id))
			}
			supports := g.iSuppPi
			for bit := int32(0); bit < supports; bit++ {
				g.iSuppPi = bit
				for _, obj := range g.pObjs {
					if Gia_ObjIsAnd(obj) != 0 {
						fmt.Fprintf(&records, "used %d %d %d\n", obj.id, bit, Gia_ManQuantCountUsed(g, obj.id))
					}
				}
			}
			g.iSuppPi = supports
			Gia_ManStop(g)
		}
	}
	esopReviewCompare(t, native, records.String())
	t.Logf("%d exact native support/marker/count records", strings.Count(records.String(), "\n"))
}
