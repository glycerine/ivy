package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strings"
	"testing"
)

func TestFxNativeFactoredCovers(t *testing.T) {
	exe := compileGiaSourceOracle(t, "fx_factor_review.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C factor: %v\n%s", err, native)
	}
	var got strings.Builder
	for n := 3; n <= 8; n++ {
		for seed := uint32(1); seed <= 16; seed++ {
			for phase := 0; phase < 2; phase++ {
				state, mask := seed, seed&((1<<n)-1)
				var sop strings.Builder
				for row := 0; row < 3; row++ {
					sop.WriteString([]string{"1-", "01", "00"}[row])
					for i := 2; i < n; i++ {
						state = state*1664525 + 1013904223
						sop.WriteByte("-01"[state>>16%3])
					}
					fmt.Fprintf(&sop, " %d\n", phase)
				}
				g := Gia_ManStart(128)
				leaves := Vec_IntAlloc(n)
				for i := 0; i < n; i++ {
					Vec_IntPush(leaves, Gia_ManAppendCi(g)^int32(mask>>i&1))
				}
				Gia_ManHashAlloc(g)
				lit := Gia_ManFactorNode(g, sop.String(), leaves)
				fmt.Fprintf(&got, "n%d seed%d phase%d lit%d objs%d ", n, seed, phase, lit, g.nObjs)
				for input := 0; input < 1<<n; input++ {
					want := false
					for _, row := range strings.Split(strings.TrimSuffix(sop.String(), "\n"), "\n") {
						term := true
						for i := 0; i < n; i++ {
							if row[i] != '-' && int(row[i]-'0') != (input^int(mask))>>i&1 {
								term = false
							}
						}
						want = want || term
					}
					want = want != (phase == 0)
					values := cofReviewValues(g, input)
					actual := values[lit>>1] != (lit&1 != 0)
					if actual != want {
						t.Fatalf("n%d seed%d phase%d input%d", n, seed, phase, input)
					}
					if actual {
						got.WriteByte('1')
					} else {
						got.WriteByte('0')
					}
				}
				got.WriteByte('\n')
				Gia_ManAppendCo(g, lit)
				g.pObjs[lit>>1].Value = uint32(n)
				cubes := Vec_WecAlloc(4)
				phases := Vec_StrStart(n + 1)
				Vec_StrWriteEntry(phases, n, byte((1-phase)^int(lit&1)))
				for _, row := range strings.Split(strings.TrimSuffix(sop.String(), "\n"), "\n") {
					cube := Vec_WecPushLevel(cubes)
					Vec_IntPush(cube, int32(n))
					var literals []int32
					for i := 0; i < n; i++ {
						if row[i] != '-' {
							v := i
							if i == 0 {
								v = n + 1
							}
							sign := int32(mask >> i & 1)
							if row[i] == '0' {
								sign ^= 1
							}
							literals = append(literals, int32(2*v)|sign)
						}
					}
					slices.Sort(literals)
					for _, v := range literals {
						Vec_IntPush(cube, v)
					}
				}
				cube := Vec_WecPushLevel(cubes)
				Vec_IntPush(cube, int32(n+1))
				Vec_IntPush(cube, 0)
				out := Gia_ManFxInsert(g, cubes, phases)
				cofReviewCheckRebuild(t, g, out)
				fmt.Fprintf(&got, "insert n%d seed%d phase%d and%d ", n, seed, phase, Gia_ManAndNum(out))
				for input := 0; input < 1<<n; input++ {
					values := cofReviewValues(out, input)
					obj := Gia_ManCo(out, 0)
					if values[Gia_ObjFaninId0p(out, obj)] != (obj.fCompl0 != 0) {
						got.WriteByte('1')
					} else {
						got.WriteByte('0')
					}
				}
				got.WriteByte('\n')
				Vec_WecFree(cubes)
				Vec_StrFree(phases)
				Gia_ManStop(out)
				Vec_IntFree(leaves)
				Gia_ManStop(g)
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}
