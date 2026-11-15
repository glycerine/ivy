package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func cofReviewGraph(seed uint32) *Gia_Man_t {
	g := Gia_ManStart(256)
	pool := []int32{0, 1}
	random := func() uint32 { seed = seed*1664525 + 1013904223; return seed >> 8 }
	edge := func() int32 { lit := pool[int(random())%len(pool)]; return lit ^ int32(random()&1) }
	for i := 0; i < 4; i++ {
		pool = append(pool, Gia_ManAppendCi(g))
	}
	Gia_ManHashAlloc(g)
	for i := 0; i < 12; i++ {
		a, b, c := edge(), edge(), edge()
		var out int32
		switch random() % 3 {
		case 0:
			out = Gia_ManHashAnd(g, a, b)
		case 1:
			out = Gia_ManHashXor(g, a, b)
		case 2:
			out = Gia_ManHashMux(g, a, b, c)
		}
		pool = append(pool, out)
	}
	for i := 0; i < 3; i++ {
		Gia_ManAppendCo(g, pool[len(pool)-1-i]^int32(seed&1))
	}
	return g
}
func cofReviewInspect(seed uint32) {
	g := cofReviewGraph(seed)
	defer Gia_ManStop(g)
	p := Cof_ManCreateLogicSimple(g)
	defer Cof_ManStop(p)
	Gia_ManLevelNum(g)
	Cof_ManResetTravId(p)
	fmt.Printf("case %d objects%d nodes%d\n", seed, Cof_ManObjNum(p), Cof_ManNodeNum(p))
	var candidates []*Cof_Obj_t
	for _, obj := range p.objects {
		fmt.Printf("object %d term%d fanins%d fanouts%d mux%d", obj.Id, Cof_ObjIsTerm(obj), Cof_ObjFaninNum(obj), Cof_ObjFanoutNum(obj), obj.nFanoutsM)
		for i, fanin := range obj.fanins {
			fmt.Printf(" in%d/%d", fanin.Id, giaBool(obj.complements[i]))
		}
		for _, fanout := range obj.fanouts {
			fmt.Printf(" out%d", fanout.Id)
		}
		fmt.Println()
		if Cof_ObjIsCi(obj) != 0 || Cof_ObjIsNode(obj) != 0 {
			candidates = append(candidates, obj)
		}
	}
	for i, a := range candidates {
		roots := []*Cof_Obj_t{a}
		fwd, rev, supp := Cof_ManTfoSize(p, roots, 1), Cof_ManTfiSize(p, roots, 1), Cof_ManSuppSize(p, roots, 1)
		fmt.Printf("single %d %d %d %d mffc%d\n", a.Id, fwd, rev, supp, Cof_ObjMffcSize(a))
		for _, b := range candidates[i+1:] {
			roots := []*Cof_Obj_t{a, b}
			fwd, rev, supp := Cof_ManTfoSize(p, roots, 2), Cof_ManTfiSize(p, roots, 2), Cof_ManSuppSize(p, roots, 2)
			fmt.Printf("pair %d %d %d %d %d\n", a.Id, b.Id, fwd, rev, supp)
		}
		for phase := int32(0); phase < 2; phase++ {
			removed := Cof_ManCountRemoved(p, a, phase)
			fmt.Printf("removed %d %d %d", a.Id, phase, removed)
			for _, obj := range p.objects {
				if Cof_ObjIsTravIdCurrent(p, obj) != 0 {
					fmt.Printf(" %d/%d", obj.Id, obj.iLit)
				}
			}
			fmt.Println()
		}
	}
	for limit := 1; limit <= len(candidates)+1; limit++ {
		top := Cof_ManCollectHighFanout(p, int32(limit))
		fmt.Printf("top %d", limit)
		for _, obj := range Vec_PtrArray(top) {
			fmt.Printf(" %d", obj.(*Cof_Obj_t).Id)
		}
		fmt.Println()
	}
	Gia_ManHashStop(g)
}
func TestCofactorNativeNetworkAnalysis(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cof_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("original C cofactor analysis: %v\n%s", err, native)
	}
	goText := hashReviewStdout(t, func() {
		for seed := uint32(0); seed < 32; seed++ {
			cofReviewInspect(seed)
		}
	})
	want, got := strings.Split(strings.TrimSpace(string(native)), "\n"), strings.Split(strings.TrimSpace(goText), "\n")
	if len(got) != len(want) {
		t.Fatalf("record counts Go%d C%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record%d\nGo %s\n C %s", i, got[i], want[i])
		}
	}
	t.Logf("32 native graphs; %d normalized analysis records match", len(got))
}
