package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func areaReviewGraph(family, signs int) *Gia_Man_t {
	subsets := [][3]int{{0, 1, 2}, {1, 3, 2}, {0, 3, 2}, {0, 1, 3}}
	count := 4
	if family == 0 || family == 2 {
		count = 2
	}
	g := Gia_ManStart(256)
	ci := make([]int32, 6)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(g)
	}
	g.EnableMuxes()
	Gia_ManHashStart(g)
	var roots []int32
	for row := 0; row < count; row++ {
		a, b, c := ci[subsets[row][0]]^int32(signs>>uint((subsets[row][0]+row)%3)&1), ci[subsets[row][1]]^int32(signs>>uint((subsets[row][1]+row)%3)&1), ci[subsets[row][2]]^int32(signs>>uint((subsets[row][2]+row)%3)&1)
		var root int32
		switch {
		case family == 5:
			root = ci[row%6] ^ int32(signs>>uint((row%6)%3)&1)
			for j := 1; j < 6; j++ {
				root = Gia_ManHashAnd(g, root, ci[(row+j)%6]^int32(signs>>uint(((row+j)%6)%3)&1))
			}
		case family == 2 || family == 3 && row&1 != 0:
			x := Gia_ManHashXorReal(g, a, b)
			root = Gia_ManHashXorReal(g, x, c)
		default:
			x := Gia_ManHashAnd(g, a, b)
			root = Gia_ManHashAnd(g, x, c)
		}
		roots = append(roots, root)
	}
	if family == 4 {
		buf := Gia_ManAppendBuf(g, roots[1])
		mux := Gia_ManHashMuxReal(g, ci[4], roots[0], buf)
		unused := Gia_ManHashMuxReal(g, ci[5], ci[0], ci[3])
		Gia_ManObj(g, unused>>1).fMark0 = 1
		roots = append(roots, mux, mux^1, buf)
	}
	for i, root := range roots {
		Gia_ManAppendCo(g, root^int32(signs>>uint(i%3)&1))
	}
	Gia_ManAppendCo(g, ci[4]^int32(signs&1))
	Gia_ManAppendCo(g, roots[0]^1)
	Gia_ManSetRegNum(g, 2)
	Gia_ManHashStop(g)
	return g
}

func areaReviewSnapshot(got *strings.Builder, p *Dam_Man_t, step int) {
	fmt.Fprintf(got, "stage %d %d %d %d %d %d %d %08x sets", step, p.pGia.nObjs, p.nLevelMax, p.nAnds, p.nDivs, p.nGain, p.nGainX, math.Float32bits(vec.Vec_QueTopPriority(p.vQue)))
	for _, obj := range p.pGia.pObjs {
		hand := Dam_ObjHand(p, obj.id)
		if hand == 0 {
			continue
		}
		set := Dam_ObjSet(p, obj.id)
		fmt.Fprintf(got, " %d:%d", obj.id, hand)
		for _, lit := range set[:1+set[0]] {
			fmt.Fprintf(got, ":%d", lit)
		}
	}
	got.WriteString(" divs")
	for id := int32(1); id <= int32(vec.Hash_IntManEntryNum(p.vHash)); id++ {
		hand := Dam_DivHand(p, id)
		fmt.Fprintf(got, " %d:%d:%d:%d:%08x:%d:%d:%d", id, vec.Hash_IntObjData0(p.vHash, id), vec.Hash_IntObjData1(p.vHash, id), vec.Hash_IntObjData2(p.vHash, id), math.Float32bits(p.vCounts[id]), vec.Vec_QueIsMember(p.vQue, int(id)), p.vDivLevR[id], hand)
		if hand > 0 {
			set := Dam_DivSet(p, id)
			for _, node := range set[:1+set[0]] {
				fmt.Fprintf(got, ":%d", node)
			}
		}
	}
	fmt.Fprintf(got, " queue %d %d marks", vec.Vec_QueSize(p.vQue), vec.Vec_QueTop(p.vQue))
	for _, obj := range p.pGia.pObjs {
		fmt.Fprintf(got, " %d", obj.fMark0|obj.fMark1<<1)
	}
	got.WriteString(" visit")
	for _, id := range p.vVisit {
		fmt.Fprintf(got, " %d", id)
	}
	got.WriteString(" levels")
	for _, obj := range p.pGia.pObjs {
		fmt.Fprintf(got, " %d", Gia_ObjLevel(p.pGia, obj))
	}
	got.WriteByte('\n')
}

func areaReviewResult(t *testing.T, got *strings.Builder, src, out *Gia_Man_t) {
	t.Helper()
	fmt.Fprintf(got, "result %d %d %d %d values", out.nObjs, out.nRegs, out.nXors, out.nMuxes)
	for _, obj := range src.pObjs {
		fmt.Fprintf(got, " %d", obj.Value)
	}
	got.WriteString(" nodes")
	for _, obj := range out.pObjs {
		a, b, c, kind := out.Gate(obj.id)
		switch kind {
		case 0, 1:
			fmt.Fprintf(got, " %d", kind)
		case 5, 6:
			fmt.Fprintf(got, " %d:%d", kind, a)
		case 2, 3:
			fmt.Fprintf(got, " %d:%d:%d", kind, a, b)
		case 4:
			fmt.Fprintf(got, " 4:%d:%d:%d", a, b, c)
		default:
			t.Fatalf("invalid kind%d", kind)
		}
	}
	got.WriteString(" truth")
	for mask := 0; mask < 64; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0, mask&32 != 0}
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
		fmt.Fprintf(got, " %d", packed)
	}
	got.WriteByte('\n')
}

func TestAreaBalanceIntermediateNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "balance_area_states.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	cases, updates := 0, 0
	for family := 0; family < 6; family++ {
		for signs := 0; signs < 8; signs++ {
			for arrival := 0; arrival < 2; arrival++ {
				for _, limit := range []int32{0, 1, 2, 4, 8} {
					src := areaReviewGraph(family, signs)
					p := Dam_ManAlloc(src)
					var inputs *Vec_Int_t
					if arrival != 0 {
						inputs = permutation(0, 2, 1, 3, 0, 1)
					}
					p.nLevelMax = Gia_ManSetLevels(src, inputs)
					p.vNodLevR = append([]int32(nil), Vec_IntArray(Gia_ManReverseLevel(src))...)
					Vec_IntFillExtra(src.vLevels, int(3*src.nObjs/2), 0)
					for _, obj := range src.pObjs {
						if Gia_ObjIsMux(src, obj) == 0 {
							obj.fMark0 = 1
						}
						obj.fMark1 = uint8(obj.id & 1)
					}
					fmt.Fprintf(&got, "case %d %d %d %d\n", family, signs, arrival, limit)
					Dam_ManCreatePairs(p, 0)
					areaReviewSnapshot(&got, p, 0)
					for step := int32(0); step < limit && vec.Vec_QueTopPriority(p.vQue) >= 2; step++ {
						div := int32(vec.Vec_QuePop(p.vQue))
						fmt.Fprintf(&got, "extract %d\n", div)
						Dam_ManUpdate(p, div)
						areaReviewSnapshot(&got, p, int(step+1))
						updates++
					}
					out := Dam_ManMultiAig(p)
					areaReviewResult(t, &got, src, out)
					Dam_ManFree(p)
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
		t.Fatalf("line counts C%d Go%d", len(a), len(b))
	}
	t.Logf("%d cases, %d extracted-divisor updates, %d truth rows and exact intermediate DAM states match C", cases, updates, cases*64)
}

func TestAreaDiagnosticsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "balance_area_states.c")
	native, err := testoracle.Command(t, exe, "diagnostics").CombinedOutput()
	if err != nil {
		t.Fatalf("native diagnostics: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		for family := 0; family < 6; family++ {
			for signs := 0; signs < 2; signs++ {
				fmt.Printf("queue %d %d\n", family, signs)
				g := areaReviewGraph(family, signs)
				p := Dam_ManAlloc(g)
				p.nLevelMax = Gia_ManSetLevels(g, nil)
				p.vNodLevR = append([]int32(nil), Vec_IntArray(Gia_ManReverseLevel(g))...)
				Dam_ManCreatePairs(p, 1)
				Dam_PrintQue(p)
				if vec.Vec_QueSize(p.vQue) > 0 {
					Dam_PrintDiv(p, int32(vec.Vec_QueTop(p.vQue)))
				}
				Dam_PrintDiv(p, 0)
				Dam_ManFree(p)
				for mode := int32(0); mode < 4; mode++ {
					fmt.Printf("verbose %d %d %d\n", family, signs, mode)
					Dam_ManAreaBalanceInt(areaReviewGraph(family, signs), nil, 3, mode&1, mode>>1&1)
				}
			}
		}
	})
	// CPU time in C and elapsed time in Go are inherently nondeterministic.
	clock := regexp.MustCompile(`Time\s*=\s*[0-9.]+ sec`)
	wantText, gotText := clock.ReplaceAllString(string(native), "Time=<elapsed>"), clock.ReplaceAllString(got, "Time=<elapsed>")
	if gotText != wantText {
		t.Fatalf("area diagnostics differ from original ABC\nGo:\n%s\nC:\n%s", gotText, wantText)
	}
}
