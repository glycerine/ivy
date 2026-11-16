package if_

import (
	"fmt"
	"slices"
	"testing"
)

func TestCluThreeLutConnectionRestrictions(t *testing.T) {
	const n = 10
	for family := 0; family < 2; family++ {
		for mode := int32(0); mode < 3; mode++ {
			for phase := 0; phase < 4; phase++ {
				t.Run(fmt.Sprintf("family%d/mode%d/phase%d", family, mode, phase), func(t *testing.T) {
					// Parallel: AND(ab) XOR AND(cdef) XOR AND(ghij).
					// Serial:   AND(abc) XOR AND(defghij).
					// Both initial bound sets are ghij. The remaining function has seven
					// variables, with that bound LUT's output at position zero.
					value := func(x int) bool {
						if family == 0 {
							return (x&3 == 3) != (x>>2&15 == 15) != (x>>6&15 == 15)
						}
						return (x&7 == 7) != (x>>3 == 127)
					}
					signedMask := 0
					if phase&2 != 0 {
						signedMask = 0x155
					}
					truth := make([]uint64, 16)
					for x := 0; x < 1<<n; x++ {
						if value(x^signedMask) != (phase&1 != 0) {
							truth[x>>6] |= 1 << uint(x&63)
						}
					}
					before := slices.Clone(truth)
					p := &If_Man_t{pPars: &If_Par_t{NLutSize: n, NStructType: mode}}
					var root, second If_Grp_t
					var f0, f1, f2 uint64
					first := If_CluCheck3(p, truth, n, 4, 4, 4, &root, &second, &f0, &f1, &f2)
					// With four inputs per LUT there is no spare pin: parallel factors
					// cannot absorb the first output, and the serial factor needs it.
					rejected := family == 0 && mode == 2 || family == 1 && mode == 1
					if rejected {
						if first.nVars != 0 {
							t.Fatalf("forbidden structure accepted: first%+v second%+v", first, second)
						}
						if !slices.Equal(truth, before) {
							t.Fatal("rejected search modified truth")
						}
						return
					}
					if first.nVars != 4 || first.nMyu != 2 || second.nVars != 4 || second.nMyu != 2 || root.nVars != 4 {
						t.Fatalf("groups: first%+v second%+v root%+v", first, second, root)
					}
					if !slices.Equal(first.pVars[:4], []int32{6, 7, 8, 9}) {
						t.Fatalf("first pins%v", first.pVars[:4])
					}
					gotPins := slices.Clone(second.pVars[:4])
					slices.Sort(gotPins)
					wantPins := []int32{2, 3, 4, 5}
					if family == 1 {
						wantPins = []int32{3, 4, 5, 10}
					}
					if !slices.Equal(gotPins, wantPins) {
						t.Fatalf("second pins%v want set%v", second.pVars[:4], wantPins)
					}
					// Source moves the second result to iVarStart: zero for allowed/
					// required, one for forbidden. Remap the first result to sentinel10.
					wantRoot := []int32{11, 10, 0, 1}
					if family == 1 {
						wantRoot = []int32{11, 0, 1, 2}
					} else if mode == 1 {
						wantRoot = []int32{10, 11, 0, 1}
					}
					if !slices.Equal(root.pVars[:4], wantRoot) {
						t.Fatalf("root pins%v want%v", root.pVars[:4], wantRoot)
					}
					if phase == 0 {
						// Bound functions are four-input AND. Root pin order above gives
						// x0 XOR x1 XOR AND(x2,x3), or x0 XOR AND(x1,x2,x3).
						wantRootTruth := uint64(0x9666966696669666)
						if family == 1 {
							wantRootTruth = 0x6aaa6aaa6aaa6aaa
						}
						if f1 != 0x8000800080008000 || f2 != 0x8000800080008000 || f0 != wantRootTruth {
							t.Fatalf("truth words%016x/%016x/%016x", f0, f1, f2)
						}
					}
					evaluate := func(g *If_Grp_t, table uint64, x int, firstBit, secondBit uint64) uint64 {
						index := 0
						for pin, v := range g.pVars[:g.nVars] {
							bit := uint64(x >> v & 1)
							if v == n {
								bit = firstBit
							} else if v == n+1 {
								bit = secondBit
							}
							index |= int(bit) << pin
						}
						return table >> index & 1
					}
					for x := 0; x < 1<<n; x++ {
						a := evaluate(&first, f1, x, 0, 0)
						b := evaluate(&second, f2, x, a, 0)
						got := evaluate(&root, f0, x, a, b) != 0
						want := value(x^signedMask) != (phase&1 != 0)
						if got != want {
							t.Fatalf("assignment%d got%t want%t", x, got, want)
						}
					}
					// Optional output arguments must not change the selected first group.
					if got := If_CluCheck3(p, truth, n, 4, 4, 4, nil, nil, nil, nil, nil); got != first {
						t.Fatalf("selection-only call changed group: %+v", got)
					}
					if !slices.Equal(truth, before) {
						t.Fatal("search modified truth")
					}
				})
			}
		}
	}
}

func TestCluThreeLutTwoOutputFallback(t *testing.T) {
	// Four distinct two-input cofactors selected by an explicit four-input
	// two-bit code. No three- or four-variable set admits a single bound
	// output, even with one shared variable (proved by enumeration below).
	// Checking size three matters: the source retries that smaller group
	// before taking its two-output fallback.
	value := func(x int) bool {
		functions := [4]uint64{6, 8, 14, 9}
		classes := [16]int{0, 0, 3, 2, 2, 3, 3, 1, 2, 1, 3, 0, 0, 3, 1, 3}
		class := classes[x>>2]
		return functions[class]>>uint(x&3)&1 != 0
	}
	for mask := 0; mask < 64; mask++ {
		var bound, free []int
		for pin := 0; pin < 6; pin++ {
			if mask>>pin&1 != 0 {
				bound = append(bound, pin)
			} else {
				free = append(free, pin)
			}
		}
		if len(bound) != 3 && len(bound) != 4 {
			continue
		}
		rows := make([]uint8, 1<<len(bound))
		for b := range rows {
			for a := 0; a < 1<<len(free); a++ {
				x := 0
				for i, pin := range bound {
					x |= (b >> i & 1) << pin
				}
				for i, pin := range free {
					x |= (a >> i & 1) << pin
				}
				if value(x) {
					rows[b] |= 1 << a
				}
			}
		}
		distinct := map[uint8]bool{}
		for _, row := range rows {
			distinct[row] = true
		}
		if len(distinct) <= 2 {
			t.Fatalf("fixture unexpectedly has a disjoint decomposition on%v", bound)
		}
		for shared := range bound {
			compatible := true
			for phase := 0; phase < 2; phase++ {
				classes := map[uint8]bool{}
				for assignment, row := range rows {
					if assignment>>shared&1 == phase {
						classes[row] = true
					}
				}
				compatible = compatible && len(classes) <= 2
			}
			if compatible {
				t.Fatalf("fixture unexpectedly shares variable%d in%v", bound[shared], bound)
			}
		}
	}
	for signedMask := 0; signedMask < 64; signedMask++ {
		for phase := 0; phase < 2; phase++ {
			t.Run(fmt.Sprintf("signs%d/phase%d", signedMask, phase), func(t *testing.T) {
				var truth uint64
				for x := 0; x < 64; x++ {
					if value(x^signedMask) != (phase != 0) {
						truth |= 1 << x
					}
				}
				input := []uint64{truth}
				var root, second If_Grp_t
				var f0, f1, f2 uint64
				first := If_CluCheck3(nil, input, 6, 4, 4, 4, &root, &second, &f0, &f1, &f2)
				if first.nVars != 4 || first.nMyu < 3 || first.nMyu > 4 || second != first || root.nVars != 4 {
					t.Fatalf("fallback groups first%+v second%+v root%+v", first, second, root)
				}
				if root.pVars[0] != 6 || root.pVars[1] != 7 {
					t.Fatalf("root lacks leading result sentinels: %+v", root)
				}
				evaluate := func(g *If_Grp_t, table uint64, values []bool) bool {
					index := 0
					for pin, v := range g.pVars[:g.nVars] {
						if values[v] {
							index |= 1 << pin
						}
					}
					return table>>index&1 != 0
				}
				for x := 0; x < 64; x++ {
					var values [8]bool
					for pin := 0; pin < 6; pin++ {
						values[pin] = x>>pin&1 != 0
					}
					values[6] = evaluate(&first, f1, values[:])
					values[7] = evaluate(&second, f2, values[:])
					got := evaluate(&root, f0, values[:])
					if want := value(x^signedMask) != (phase != 0); got != want {
						t.Fatalf("assignment%d got%t want%t", x, got, want)
					}
				}
				if input[0] != truth {
					t.Fatal("fallback changed input truth")
				}
			})
		}
	}
}
