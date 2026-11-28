package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"sort"
	"strings"
	"testing"
)

func TestAigBalancingHelpersNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "balance_pair_helpers.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	for mode := 0; mode < 4; mode++ {
		g := Gia_ManStart(256)
		inputs := make([]int32, 12)
		for i := range inputs {
			inputs[i] = Gia_ManAppendCi(g)
		}
		Gia_ManCleanLevels(g, g.nObjs)
		for i, lit := range inputs {
			level := 0
			switch mode {
			case 1:
				level = i % 3
			case 2:
				level = i / 3
			case 3:
				level = (12 - i) % 4
			}
			Gia_ObjSetLevelId(g, lit>>1, int32(level))
		}
		Gia_ManHashStart(g)
		for i := 0; i < 12; i++ {
			for j := i + 1; j < 12; j++ {
				for phase := int32(0); phase < 4; phase++ {
					if (i+j+int(phase))%3 == 0 {
						lit := Gia_ManHashAnd(g, inputs[i]^(phase&1), inputs[j]^((phase>>1)&1))
						Gia_ObjSetGateLevel(g, Gia_ManObj(g, lit>>1))
					}
				}
			}
		}
		for n := 2; n <= 12; n++ {
			for rotation := 0; rotation < 12; rotation++ {
				literals := make([]int32, n)
				for i := range literals {
					literals[i] = inputs[(rotation+i)%12] ^ int32((rotation+i)&1)
				}
				sort.SliceStable(literals, func(i, j int) bool { return Gia_ObjLevelId(g, literals[i]>>1) > Gia_ObjLevelId(g, literals[j]>>1) })
				v := permutation(literals...)
				Gia_ManPrepareLastTwo(g, v)
				fmt.Fprintf(&got, "prepare %d %d %d", mode, n, rotation)
				for _, lit := range Vec_IntArray(v) {
					fmt.Fprintf(&got, " %d", lit)
				}
				got.WriteByte('\n')
				for phase := int32(0); phase < 2; phase++ {
					copy := permutation(literals...)
					chosen := Gia_ManFindSharedNode(g, copy, inputs[(rotation+3)%12]^phase)
					fmt.Fprintf(&got, "find %d %d %d %d %d", mode, n, rotation, phase, chosen)
					for _, lit := range Vec_IntArray(copy) {
						fmt.Fprintf(&got, " %d", lit)
					}
					got.WriteByte('\n')
				}
			}
		}
	}
	for mode := 0; mode < 2; mode++ {
		for _, size := range []int{8, 64} {
			for extra := 0; extra < 3; extra++ {
				for leafRefs := 0; leafRefs < 2; leafRefs++ {
					for strict := int32(0); strict < 2; strict++ {
						g := Gia_ManStart(256)
						inputs := make([]int32, size)
						for i := range inputs {
							inputs[i] = Gia_ManAppendCi(g)
						}
						appendGate := Gia_ManAppendAnd
						if mode == 1 {
							appendGate = Gia_ManAppendXorReal
						}
						base := appendGate(g, inputs[0], inputs[1])
						root := base
						for _, lit := range inputs[2:] {
							root = appendGate(g, root, lit)
						}
						Gia_ManAppendCo(g, root)
						for i := 0; i < extra; i++ {
							Gia_ManAppendCo(g, base)
						}
						if leafRefs != 0 {
							Gia_ManAppendCo(g, inputs[0])
							Gia_ManAppendCo(g, inputs[1])
						}
						Gia_ManCreateRefs(g)
						Gia_ManSuperCollect(g, Gia_ManObj(g, root>>1), strict)
						literals := append([]int32(nil), Vec_IntArray(g.vSuper)...)
						fmt.Fprintf(&got, "collect %d %d %d %d %d", mode, size, extra, leafRefs, strict)
						for _, lit := range literals {
							fmt.Fprintf(&got, " %d", lit)
						}
						got.WriteByte('\n')
						state := uint64(77)
						for row := 0; row < 128; row++ {
							bits := make([]bool, size)
							for i := range bits {
								state = state*6364136223846793005 + 1442695040888963407
								bits[i] = state>>63 != 0
							}
							values, _, err := g.evaluateObjects(bits)
							if err != nil {
								t.Fatal(err)
							}
							value := mode == 0
							for _, lit := range literals {
								x := values[lit>>1] != (lit&1 != 0)
								if mode == 0 {
									value = value && x
								} else {
									value = value != x
								}
							}
							if value != values[root>>1] {
								t.Fatal("supergate collection changed Boolean expression")
							}
						}
					}
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(string(native), "\n"), strings.Split(got.String(), "\n")
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] {
				t.Fatalf("line%d\nC: %s\nGo: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("line counts C%d Go%d", len(a), len(b))
	}
	t.Log("528 last-pair preparations, 1056 shared-node selections and 48 strict/refcount/size collector cases match C")
}
