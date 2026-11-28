package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"strings"
	"testing"
)

// Two boxes have unequal input/output widths; the second depends on the first.
// The final three COs are the external outputs (including any register inputs).
func boxOrderReviewGraph(signs int, typed, choice bool, regs int32) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, u, v, w := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	var sibling int32
	if choice {
		sibling = Gia_ManAppendAnd(g, a^phase(0), b^phase(1))
	}
	x0 := Gia_ManAppendAnd(g, a^phase(0), b^phase(1))
	x1 := Gia_ManAppendAnd(g, a, b^phase(2))
	y0 := Gia_ManAppendAnd(g, u^phase(3), v^phase(4))
	o0 := Gia_ManAppendAnd(g, w^phase(5), y0^phase(6))
	o1 := Gia_ManAppendAnd(g, w^phase(7), a)
	if typed {
		x1 = Gia_ManAppendXorReal(g, a, b^phase(2))
		g.EnableMuxes()
		y0 = Gia_ManAppendMuxReal(g, u^phase(3), v^phase(4), a)
		o0 = Gia_ManAppendXorReal(g, w^phase(5), y0^phase(6))
		o1 = Gia_ManAppendMuxReal(g, w^phase(7), b, u)
	}
	Gia_ManAppendAnd(g, a^1, b) // unreachable logic must stay absent from the order
	for _, lit := range []int32{x0, x1, y0, o0, o1, b ^ 1} {
		Gia_ManAppendCo(g, lit)
	}
	if choice {
		g.pSibls = make([]int32, g.nObjs)
		g.pSibls[x0>>1] = sibling >> 1
	}
	Gia_ManSetRegNum(g, regs)
	timing := tim.Tim_ManStart(5, 6)
	tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 2, 2, 1, 2, 3, 4}, {1, 1, 1, 5}})
	tim.Tim_ManCreateBox(timing, 0, 2, 2, 2, 0, 0)
	tim.Tim_ManCreateBox(timing, 2, 1, 4, 1, 1, 0)
	g.PManTime = timing
	g.SetIdentity("boxed-review", "boxed-review.aig")
	return g
}

func TestUnnormalizeRetainsBadOutputPartition(t *testing.T) {
	src := boxOrderReviewGraph(5, false, false, 0)
	if err := src.SetAigerProperties(1, 1); err != nil {
		t.Fatal(err)
	}
	out := Gia_ManDupUnnormalize(src)
	if out == nil {
		t.Fatal("unnormalization failed")
	}
	if out.aigerBadOutputs != 1 || out.nConstrs != 1 {
		t.Fatalf("roles bad%d constraints%d, want1/1", out.aigerBadOutputs, out.nConstrs)
	}
}

func TestBoxOrderNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "timing_box_order.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	for signs := 0; signs < 256; signs++ {
		for _, choice := range []bool{false, true} {
			for regs := int32(0); regs < 3; regs++ {
				src := boxOrderReviewGraph(signs, false, choice, regs)
				order := Gia_ManOrderWithBoxes(src)
				fmt.Fprintf(&got, "case %d %d %d order", signs, giaBool(choice), regs)
				for _, id := range vec.Vec_IntArray(order) {
					fmt.Fprintf(&got, " %d", id)
				}
				out := Gia_ManDupUnnormalize(src)
				if out == nil {
					t.Fatal("valid boxed graph rejected")
				}
				fmt.Fprintf(&got, " result %d %d %d %d values", out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				for _, obj := range src.pObjs {
					fmt.Fprintf(&got, " %d", obj.Value)
				}
				got.WriteString(" nodes")
				for _, obj := range out.pObjs {
					a, b, _, kind := out.Gate(obj.id)
					switch kind {
					case 0, 1:
						fmt.Fprintf(&got, " %d", kind)
					case 6:
						fmt.Fprintf(&got, " 6:%d", a)
					case 2:
						fmt.Fprintf(&got, " 2:%d:%d:%d", a, b, Gia_ObjSibl(out, obj.id))
					default:
						t.Fatalf("unexpected gate %d", kind)
					}
				}
				got.WriteString(" truth")
				for mask := 0; mask < 32; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
					want, err := src.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					checkQuantificationOutputs(t, out, inputs, want)
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
				}
				got.WriteByte('\n')
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(string(native), "\n"), strings.Split(got.String(), "\n")
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] {
				t.Fatalf("line %d\nC:  %s\nGo: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("line count C %d Go %d", len(a), len(b))
	}
	t.Log("1536 native boxed graphs and 49152 truth rows match exact source order, copy values, structure and siblings")
}

func TestBoxOrderPartitionAndTypedMatrix(t *testing.T) {
	cases := 0
	for _, signs := range []int{0, 37, 90, 255} {
		for _, typed := range []bool{false, true} {
			for _, choice := range []bool{false, true} {
				for regs := int32(0); regs < 3; regs++ {
					for bad := int32(0); bad <= 3-regs; bad++ {
						for constraints := int32(0); constraints <= 3-regs-bad; constraints++ {
							src := boxOrderReviewGraph(signs, typed, choice, regs)
							if err := src.SetAigerProperties(bad, constraints); err != nil {
								t.Fatal(err)
							}
							out := Gia_ManDupUnnormalize(src)
							if out == nil || out.nRegs != regs {
								t.Fatal("boxed interface lost")
							}
							restored := partitionReviewRoundTrip(t, out, bad, constraints)
							for mask := 0; mask < 32; mask++ {
								inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
								want, err := src.Evaluate(inputs)
								if err != nil {
									t.Fatal(err)
								}
								checkQuantificationOutputs(t, out, inputs, want)
								checkQuantificationOutputs(t, restored, inputs, want)
							}
							cases++
						}
					}
				}
			}
		}
	}
	t.Logf("%d boxed copies, %d rows checked directly and after serialization", cases, cases*32)
}

func BenchmarkBoxOrderAndChain(b *testing.B) {
	g := Gia_ManStart(1)
	a, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, c)
	for i := 0; i < 255; i++ {
		root = Gia_ManAppendAnd(g, a, root)
	}
	Gia_ManAppendCo(g, root)
	g.PManTime = tim.Tim_ManStart(2, 1)
	Gia_ManOrderWithBoxes(g)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vec.Vec_IntFree(Gia_ManOrderWithBoxes(g))
	}
}

func TestUnnormalizeSizesSiblingTableForRetainedNodes(t *testing.T) {
	src := boxOrderReviewGraph(0, false, true, 0)
	out := Gia_ManDupUnnormalize(src)
	if len(out.pSibls) != int(out.nObjs) {
		t.Fatalf("sibling slots %d, retained objects %d (source objects %d)", len(out.pSibls), out.nObjs, src.nObjs)
	}
}

func TestBoxOrderRejectsFutureBoxDependency(t *testing.T) {
	for _, controlOnly := range []bool{false, true} {
		var g *Gia_Man_t
		if !controlOnly {
			g = boxOrderReviewGraph(0, false, false, 0)
			co := Gia_ManCo(g, 0)
			co.iDiff0 = uint32(co.id - Gia_ManCi(g, 4).id)
		} else {
			// Both data inputs are available; only the MUX control belongs to a
			// later box. It must participate in the dependency walk too.
			g = Gia_ManStart(1)
			a, b, u, w := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			g.EnableMuxes()
			root := Gia_ManAppendMuxReal(g, w, a, b)
			Gia_ManAppendCo(g, root)
			Gia_ManAppendCo(g, u)
			Gia_ManAppendCo(g, a)
			timing := tim.Tim_ManStart(4, 3)
			tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, -1, 0)
			tim.Tim_ManCreateBox(timing, 1, 1, 3, 1, -1, 0)
			g.PManTime = timing
		}
		for repeat := 0; repeat < 2; repeat++ {
			if Gia_ManOrderWithBoxes(g) != nil || g.iData2 != 0 {
				t.Fatalf("future dependency accepted: controlOnly%v repeat%d", controlOnly, repeat)
			}
			if Gia_ManDupUnnormalize(g) != nil {
				t.Fatal("invalid box order duplicated")
			}
		}
	}
}

func TestBoxOrderAsymmetricWidths(t *testing.T) {
	g := Gia_ManStart(1)
	a, b, u, v, w := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x0, x1 := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, a, b^1)
	y0 := Gia_ManAppendAnd(g, u, a)
	o0 := Gia_ManAppendAnd(g, v, w)
	cos := []int32{}
	for _, lit := range []int32{x0, x1, y0, o0, v, a ^ 1} {
		cos = append(cos, Gia_ManAppendCo(g, lit)>>1)
	}
	timing := tim.Tim_ManStart(5, 6)
	tim.Tim_ManCreateBox(timing, 0, 2, 2, 1, -1, 0)
	tim.Tim_ManCreateBox(timing, 2, 1, 3, 2, -1, 0)
	g.PManTime = timing
	wantOrder := []int32{0, a >> 1, b >> 1, x0 >> 1, x1 >> 1, cos[0], cos[1], u >> 1, y0 >> 1, cos[2], v >> 1, w >> 1, o0 >> 1, cos[3], cos[4], cos[5]}
	if got := vec.Vec_IntArray(Gia_ManOrderWithBoxes(g)); !slices.Equal(got, wantOrder) {
		t.Fatalf("order %v, want %v", got, wantOrder)
	}
	out := Gia_ManDupUnnormalize(g)
	for mask := 0; mask < 32; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
		want, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		checkQuantificationOutputs(t, out, inputs, want)
	}
}

func TestBoxOrderBufferAndConstantCones(t *testing.T) {
	g := Gia_ManStart(1)
	a := Gia_ManAppendCi(g)
	root := Gia_ManAppendBuf(g, Gia_ManAppendBuf(g, a^1))
	Gia_ManAppendBuf(g, a) // dangling
	co0, co1, co2 := Gia_ManAppendCo(g, root), Gia_ManAppendCo(g, 0), Gia_ManAppendCo(g, 1)
	g.PManTime = tim.Tim_ManStart(1, 3)
	want := []int32{0, a >> 1, 2, root >> 1, co0 >> 1, co1 >> 1, co2 >> 1}
	if got := vec.Vec_IntArray(Gia_ManOrderWithBoxes(g)); !slices.Equal(got, want) {
		t.Fatalf("order %v, want %v", got, want)
	}
}

func TestBoxSelectionTimingAndExtraOutputs(t *testing.T) {
	src := boxOrderReviewGraph(0, false, false, 0)
	timing := src.PManTime.(*tim.Tim_Man_t)
	for _, ids := range [][]int32{{}, {0}, {1}, {0, 1}, {1, 0}} {
		boxes := permutation(ids...)
		outputs := []int{}
		for _, id := range ids {
			if id == 0 {
				outputs = append(outputs, 0, 1)
			} else {
				outputs = append(outputs, 2)
			}
		}
		for diff := int32(0); diff <= 2; diff++ {
			out := Gia_ManUpdateTimMan2(src, boxes, diff).(*tim.Tim_Man_t)
			if tim.Tim_ManCiNum(out) != 2-int(diff)+len(outputs) || tim.Tim_ManCoNum(out) != 3-int(diff)+len(outputs) || tim.Tim_ManBoxNum(out) != len(ids) {
				t.Fatalf("box selection %v diff%d changed timing dimensions", ids, diff)
			}
			for i, id := range ids {
				if tim.Tim_ManBoxInputNum(out, i) != tim.Tim_ManBoxInputNum(timing, int(id)) || tim.Tim_ManBoxOutputNum(out, i) != tim.Tim_ManBoxOutputNum(timing, int(id)) {
					t.Fatal("box width changed")
				}
				// The reduced manager owns its delay table data.
				table := tim.Tim_ManBoxDelayTable(out, i)
				original := tim.Tim_ManBoxDelayTable(timing, int(id))
				old := original[3]
				table[3]++
				if original[3] != old {
					t.Fatal("reduced delay table aliases source")
				}
			}
		}
		extra := Gia_ManStart(1)
		a, b, c := Gia_ManAppendCi(extra), Gia_ManAppendCi(extra), Gia_ManAppendCi(extra)
		for _, lit := range []int32{Gia_ManAppendAnd(extra, a, b^1), c ^ 1, a} {
			Gia_ManAppendCo(extra, lit)
		}
		out := Gia_ManUpdateExtraAig2(timing, extra, boxes)
		if Gia_ManCoNum(out) != int32(len(outputs)) {
			t.Fatal("wrong selected output count")
		}
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			values, err := extra.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			want := make([]bool, len(outputs))
			for i, id := range outputs {
				want[i] = values[id]
			}
			checkQuantificationOutputs(t, out, inputs, want)
		}
	}
}
