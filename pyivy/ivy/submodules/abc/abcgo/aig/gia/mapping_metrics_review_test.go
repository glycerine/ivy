package gia

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func mappingReviewOrdinary(seed uint32, mode int) *Gia_Man_t {
	next := func() uint32 { seed = seed*1664525 + 1013904223; return seed }
	p := Gia_ManStart(64)
	lits := make([]int32, 4, 16)
	for i := range lits {
		lits[i] = Gia_ManAppendCi(p)
	}
	var roots []int32
	for i := 0; i < 6; i++ {
		a, b := int(next()%uint32(4+2*i)), int(next()%uint32(3+2*i))
		if b >= a {
			b++
		}
		al, bl := lits[a]^int32(next()&1), lits[b]^int32(next()&1)
		x, y := Gia_ManAppendAnd(p, al, bl), Gia_ManAppendAnd(p, al, bl)
		lits = append(lits, x, y)
		roots = append(roots, x, y)
	}
	for i, root := range roots {
		Gia_ManAppendCo(p, root^int32(i&1))
	}
	p.pSibls = make([]int32, p.nObjs)
	for i := 0; i < len(roots); i += 2 {
		p.pSibls[roots[i+1]>>1] = roots[i] >> 1
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for i, root := range roots {
		leaves := []int32{1, 2, 3, 4}
		obj := p.pObjs[root>>1]
		if mode%3 == 1 || (mode%3 == 2 && i&1 != 0) {
			leaves = []int32{Gia_ObjFaninId0(obj, obj.id), Gia_ObjFaninId1(obj, obj.id)}
		}
		Vec_IntWriteEntry(p.VMapping, int(obj.id), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(leaves)))
		for _, id := range leaves {
			Vec_IntPush(p.VMapping, id)
		}
		Vec_IntPush(p.VMapping, obj.id)
	}
	p.SetIdentity("mapping_review", "")
	return p
}
func mappingReviewArray(w io.Writer, name string, data []int32) {
	fmt.Fprint(w, name)
	for _, v := range data {
		fmt.Fprintf(w, " %d", v)
	}
	fmt.Fprintln(w)
}
func mappingReviewMetrics(w io.Writer, p *Gia_Man_t) {
	var levels []int32
	depth := Gia_ManLutLevel(p, &levels)
	fmt.Fprintf(w, "metrics %d %d %d %d\n", Gia_ManLutNum(p), Gia_ManLutFaninCount(p), Gia_ManLutSizeMax(p), depth)
	mappingReviewArray(w, "levels", levels)
	var luts, edges, maximum int32
	Gia_ManLutParams(p, &luts, &edges, &maximum)
	fmt.Fprintf(w, "params %d %d %d\n", luts, edges, maximum)
	Gia_ManSetRefsMapped(p)
	mappingReviewArray(w, "refs", p.Refs)
	Gia_ManSetLutRefs(p)
	mappingReviewArray(w, "lutrefs", p.pLutRefs)
	label, visited := Vec_StrStart(int(p.nObjs)), Vec_IntAlloc(32)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			count := Gia_ManComputeOverlap2One(p, id, label, visited)
			fmt.Fprintf(w, "cone %d %d %d", id, Gia_ManComputeOverlapOne(p, id), count)
			mappingReviewArray(w, " visited", Vec_IntArray(visited))
		}
	}
	fmt.Fprintf(w, "overlap %d %d\n", Gia_ManComputeOverlap(p), Gia_ManComputeOverlap2(p))
	fmt.Fprintf(w, "choice %d\n", Gia_ManChoiceLevel(p))
	mappingReviewArray(w, "choicelevels", Vec_IntArray(p.vLevels))
}
func TestMappingNativeMetricsAndChoiceLevels(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_metrics_review.c")
	comparisons := 0
	for seed := 1; seed <= 32; seed++ {
		for mode := 0; mode < 3; mode++ {
			source, err := exec.Command(exe, strconv.Itoa(seed), strconv.Itoa(mode)).CombinedOutput()
			if err != nil {
				t.Fatalf("native seed%d mode%d: %v %s", seed, mode, err, source)
			}
			p := mappingReviewOrdinary(uint32(seed), mode)
			var got strings.Builder
			mappingReviewMetrics(&got, p)
			got.WriteString(hashReviewStdout(t, func() { Gia_ManPrintMappingStats(p, ""); Gia_ManPrintLutStats(p) }))
			Gia_ManStop(p)
			if string(source) != got.String() {
				t.Logf("seed%d mode%d", seed, mode)
				esopReviewCompare(t, source, got.String())
			}
			comparisons++
		}
	}
	t.Logf("compared %d complete ordinary-AND metrics/choice/report states", comparisons)
}
func mappingReviewAverageGraph() *Gia_Man_t {
	p := Gia_ManStart(256)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	for i := 0; i < 200; i++ {
		root := x
		if i < 13 {
			root = y
		}
		Gia_ManAppendCo(p, root)
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, row := range [][]int32{{x >> 1, 2, a >> 1, b >> 1, x >> 1}, {y >> 1, 2, x >> 1, c >> 1, y >> 1}} {
		Vec_IntWriteEntry(p.VMapping, int(row[0]), int32(Vec_IntSize(p.VMapping)))
		for _, v := range row[1:] {
			Vec_IntPush(p.VMapping, v)
		}
	}
	return p
}
func TestMappingNativeAveragePrecision(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_metrics_review.c")
	source, err := exec.Command(exe, "1", "0", "1").CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v %s", err, source)
	}
	p := mappingReviewAverageGraph()
	defer Gia_ManStop(p)
	got := hashReviewStdout(t, func() { Gia_ManPrintMappingStats(p, "") })
	if string(source) != got {
		t.Fatalf("Go average report%q; C%q", got, source)
	}
}
func TestMappingNativeEmptyPackingReport(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_metrics_review.c")
	source, err := exec.Command(exe, "1", "0", "2").CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v %s", err, source)
	}
	p := Gia_ManStart(1)
	defer Gia_ManStop(p)
	p.VPacking = Vec_IntStart(1)
	got := hashReviewStdout(t, func() { Gia_ManPrintPackingStats(p) })
	if string(source) != got {
		t.Fatalf("Go packing report%q; C%q", got, source)
	}
}
