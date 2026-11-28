package gia

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"github.com/glycerine/abcgo/misc/tim"
	"slices"
	"strings"
	"testing"
)

func unshuffleReviewGraph(signs int, kind string) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, u, r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	var x, y int32
	switch kind {
	case "and":
		x = Gia_ManAppendAnd(g, a^phase(0), u^phase(1))
		y = Gia_ManAppendAnd(g, x, r0^phase(2))
	case "xor":
		x = Gia_ManAppendXorReal(g, a^phase(0), u^phase(1))
		y = Gia_ManAppendAnd(g, x, r0^phase(2))
	case "mux":
		g.EnableMuxes()
		x = Gia_ManAppendMuxReal(g, u^phase(1), a^phase(0), b)
		y = Gia_ManAppendXorReal(g, x, r0^phase(2))
	case "buffer":
		x = Gia_ManAppendBuf(g, u^phase(1))
		y = Gia_ManAppendBuf(g, x^phase(2))
	case "simple":
		g.fGiaSimple = 1
		x = Gia_ManAppendAnd(g, u, u^phase(0))
		y = Gia_ManAppendAnd(g, x, r0^phase(2))
	default:
		panic(kind)
	}
	z := Gia_ManAppendAnd(g, y, r1)
	for _, lit := range []int32{y, x ^ phase(3), a ^ 1, r0, z} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, 2)
	timing := tim.Tim_ManStart(5, 5)
	tim.Tim_ManCreateBox(timing, 0, 1, 4, 1, -1, 0)
	g.PManTime = timing
	g.SetIdentity("unshuffle", "unshuffle.aig")
	return g
}

func TestUnshuffleRetainsBadOutputPartition(t *testing.T) {
	src := unshuffleReviewGraph(0, "and")
	if err := src.SetAigerProperties(1, 1); err != nil {
		t.Fatal(err)
	}
	out := Gia_ManDupUnshuffleInputs(src)
	if out.aigerBadOutputs != 1 || out.nConstrs != 1 {
		t.Fatalf("roles bad%d constraints%d, want1/1", out.aigerBadOutputs, out.nConstrs)
	}
}

func TestUnshuffleRetainsInternalGateKinds(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer", "simple"} {
		t.Run(kind, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Fatalf("unshuffle rejected supported %s graph: %v", kind, err)
				}
			}()
			src := unshuffleReviewGraph(5, kind)
			out := Gia_ManDupUnshuffleInputs(src)
			for mask := 0; mask < 32; mask++ {
				original := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				inputs := []bool{original[0], original[1], original[3], original[4], original[2]}
				want, err := src.Evaluate(original)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}

func unshufflePartitionReviewGraph(pis, regs, boxes, signs int, kind string) *Gia_Man_t {
	g := Gia_ManStart(1)
	ci := make([]int32, pis+regs+boxes)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	a, b, c := ci[0], ci[len(ci)-1], ci[pis]
	gate := func() int32 {
		switch kind {
		case "and":
			return Gia_ManAppendAnd(g, a^phase(0), b^phase(1))
		case "xor":
			return Gia_ManAppendXorReal(g, a^phase(0), b^phase(1))
		case "mux":
			g.EnableMuxes()
			return Gia_ManAppendMuxReal(g, c^phase(2), a^phase(0), b^phase(1))
		case "buffer":
			return Gia_ManAppendBuf(g, a^phase(0))
		case "simple":
			g.fGiaSimple = 1
			return Gia_ManAppendAnd(g, a, a^phase(0))
		default:
			panic(kind)
		}
	}
	x, duplicate := gate(), gate()
	y := Gia_ManAppendAnd(g, x, c^phase(2))
	outputs := []int32{y, x ^ phase(3), duplicate}
	for _, r := range ci[pis+boxes:] {
		outputs = append(outputs, Gia_ManAppendAnd(g, y, r))
	}
	for _, lit := range outputs {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, int32(regs))
	timing := tim.Tim_ManStart(len(ci), len(outputs))
	tim.Tim_ManCreateBox(timing, 0, 1, pis+regs, boxes, -1, 0)
	g.PManTime = timing
	g.StartRepresentatives()
	Gia_ObjSetRepr(g, duplicate>>1, x>>1)
	g.PNexts = Gia_ManDeriveNexts(g)
	g.SetIdentity("unshuffle-partition", "unshuffle-partition.aig")
	return g
}

func unshuffleReviewPermutation(pis, regs, boxes int) []int {
	var order []int
	for i := 0; i < pis; i++ {
		order = append(order, i)
	}
	for i := pis + boxes; i < pis+boxes+regs; i++ {
		order = append(order, i)
	}
	for i := pis; i < pis+boxes; i++ {
		order = append(order, i)
	}
	return order
}

func TestUnshufflePartitionMatrix(t *testing.T) {
	cases, rows := 0, 0
	for pis := 1; pis <= 3; pis++ {
		for regs := 1; regs <= 3; regs++ {
			for boxes := 1; boxes <= 3; boxes++ {
				order := unshuffleReviewPermutation(pis, regs, boxes)
				for _, kind := range []string{"and", "xor", "mux", "buffer", "simple"} {
					for _, signs := range []int{0, 5, 10, 15} {
						for bad := int32(0); bad <= 3; bad++ {
							for constraints := int32(0); constraints <= 3-bad; constraints++ {
								src := unshufflePartitionReviewGraph(pis, regs, boxes, signs, kind)
								if err := src.SetAigerProperties(bad, constraints); err != nil {
									t.Fatal(err)
								}
								out := Gia_ManDupUnshuffleInputs(src)
								if out.nObjs != src.nObjs || out.nRegs != src.nRegs || Gia_ManIsNormalized(out) == 0 {
									t.Fatalf("copy layout changed for %s", kind)
								}
								if name, spec := out.Identity(); name != "unshuffle-partition" || spec != "unshuffle-partition.aig" {
									t.Fatal("identity lost")
								}
								for dest, source := range order {
									if src.pObjs[source+1].Value != uint32((dest+1)<<1) {
										t.Fatalf("CI source%d maps to %d, want dest%d", source, src.pObjs[source+1].Value, dest)
									}
								}
								for i := 1 + len(order); i < int(src.nObjs); i++ {
									if src.pObjs[i].Value != uint32(i<<1) {
										t.Fatal("internal/CO copy map changed")
									}
								}
								if !slices.Equal(src.PReprs, out.PReprs) || !slices.Equal(src.PNexts, out.PNexts) {
									t.Fatal("equivalence classes changed")
								}
								restored := partitionReviewRoundTrip(t, out, bad, constraints)
								original := make([]bool, len(order))
								inputs := make([]bool, len(order))
								for mask := 0; mask < 1<<len(order); mask++ {
									for i := range original {
										original[i] = mask>>i&1 != 0
									}
									for i, source := range order {
										inputs[i] = original[source]
									}
									want, err := src.Evaluate(original)
									if err != nil {
										t.Fatal(err)
									}
									checkQuantificationOutputs(t, out, inputs, want)
									checkQuantificationOutputs(t, restored, inputs, want)
									rows++
								}
								cases++
							}
						}
					}
				}
			}
		}
	}
	t.Logf("%d boxed-input permutations, %d rows checked directly and after serialization", cases, rows)
}

func TestUnshuffleNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "timing_unshuffle.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\\n%s", err, native)
	}
	// The source prints this unconditional informational line before each copy.
	native = bytes.ReplaceAll(native, []byte("Warning: Unshuffled CI order to be correct AIG with boxes.\n"), nil)
	var got strings.Builder
	cases, rows := 0, 0
	for pis := 1; pis <= 3; pis++ {
		for regs := 1; regs <= 3; regs++ {
			for boxes := 1; boxes <= 3; boxes++ {
				for signs := 0; signs < 16; signs++ {
					src := unshufflePartitionReviewGraph(pis, regs, boxes, signs, "and")
					out := Gia_ManDupUnshuffleInputs(src)
					fmt.Fprintf(&got, "case %d %d %d %d result %d %d values", pis, regs, boxes, signs, out.nObjs, out.nRegs)
					for _, obj := range src.pObjs {
						fmt.Fprintf(&got, " %d", obj.Value)
					}
					got.WriteString(" repr")
					for _, r := range out.PReprs {
						fmt.Fprintf(&got, " %d", r.iRepr)
					}
					got.WriteString(" next")
					for _, next := range out.PNexts {
						fmt.Fprintf(&got, " %d", next)
					}
					got.WriteString(" nodes")
					for _, obj := range out.pObjs {
						a, b, _, kind := out.Gate(obj.id)
						switch kind {
						case 0, 1:
							fmt.Fprintf(&got, " %d", kind)
						case 2:
							fmt.Fprintf(&got, " 2:%d:%d", a, b)
						case 6:
							fmt.Fprintf(&got, " 6:%d", a)
						default:
							t.Fatalf("unexpected kind %d", kind)
						}
					}
					got.WriteString(" truth")
					inputs := make([]bool, Gia_ManCiNum(out))
					for mask := 0; mask < 1<<len(inputs); mask++ {
						for i := range inputs {
							inputs[i] = mask>>i&1 != 0
						}
						values, err := out.Evaluate(inputs)
						if err != nil {
							t.Fatal(err)
						}
						packed := 0
						for i, value := range values {
							if value {
								packed |= 1 << i
							}
						}
						fmt.Fprintf(&got, " %d", packed)
						rows++
					}
					got.WriteByte('\n')
					cases++
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
		t.Fatalf("line count C%d Go%d", len(a), len(b))
	}
	t.Logf("%d native cases, %d input/state rows, exact structure, copy maps and representatives", cases, rows)
}
