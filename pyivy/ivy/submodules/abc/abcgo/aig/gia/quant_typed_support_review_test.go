package gia

import (
	"fmt"
	"testing"
)

func TestQuantTypedIncrementalSupport(t *testing.T) {
	for _, nvars := range []int{3, 65} {
		for _, kind := range []string{"xor", "mux", "buffer"} {
			for mode := 0; mode < 2; mode++ {
				t.Run(fmt.Sprintf("inputs%d/%s/mode%d", nvars, kind, mode), func(t *testing.T) {
					defer func() {
						if err := recover(); err != nil {
							t.Fatalf("typed support/quantification panicked: %v", err)
						}
					}()
					for phase := int32(0); phase < 2; phase++ {
						for mask := 0; mask < 8; mask++ {
							g := Gia_ManStart(128)
							g.EnableMuxes()
							Gia_ManQuantSetSuppStart(g)
							inputs := make([]int32, nvars)
							for i := range inputs {
								inputs[i] = Gia_ManAppendCi(g)
								Gia_ManQuantSetSuppZero(g)
								Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, inputs[i]>>1))
							}
							a, b, c := inputs[0], inputs[1], inputs[nvars-1]
							var root int32
							switch kind {
							case "xor":
								root = Gia_ManAppendXorReal(g, a, b)
							case "mux":
								root = Gia_ManAppendMuxReal(g, c, a, b)
							case "buffer":
								root = Gia_ManAppendBuf(g, a^1)
							}
							row := Gia_ManQuantInfoId(g, root>>1)
							for i := 0; i < nvars; i++ {
								want := i == 0 || kind != "buffer" && i == 1 || kind == "mux" && i == nvars-1
								got := row[i>>6]&(uint64(1)<<uint(i&63)) != 0
								if got != want {
									t.Fatalf("support bit%d got%t want%t", i, got, want)
								}
							}
							Gia_ManHashAlloc(g)
							keep := func(_ any, ci int32) int32 {
								bit := int32(-1)
								switch ci {
								case 0:
									bit = 0
								case 1:
									bit = 1
								case int32(nvars - 1):
									bit = 2
								}
								if bit < 0 {
									return 1
								}
								return int32(mask>>bit) & 1
							}
							var result int32
							if mode == 0 {
								result = Gia_ManQuantExist(g, root^phase, keep, nil)
							} else {
								result = Gia_ManQuantExist2(g, root^phase, keep, nil)
							}
							Gia_ManAppendCo(g, result)
							for fixed := 0; fixed < 8; fixed++ {
								values := make([]bool, nvars)
								values[0] = fixed&1 != 0
								values[1] = fixed&2 != 0
								values[nvars-1] = fixed&4 != 0
								got, err := g.Evaluate(values)
								if err != nil {
									t.Fatal(err)
								}
								want := false
								for assignment := 0; assignment < 8; assignment++ {
									if assignment&mask != fixed&mask {
										continue
									}
									x, y, z := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
									var value bool
									switch kind {
									case "xor":
										value = x != y
									case "mux":
										if z {
											value = x
										} else {
											value = y
										}
									case "buffer":
										value = !x
									}
									want = want || value != (phase != 0)
								}
								if got[0] != want {
									t.Fatalf("phase%d keep%x input%d got%t want%t", phase, mask, fixed, got[0], want)
								}
							}
							Gia_ManStop(g)
						}
					}
				})
			}
		}
	}
}
