package gia

import (
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func cofReviewGraph(seed uint32) *Gia_Man_t {
	g := Gia_ManStart(256)
	pool := []int32{0, 1}
	rng := seed
	random := func() uint32 { rng = rng*1664525 + 1013904223; return rng >> 8 }
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

func TestCofactorNativeRebuild(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cof_review.c")
	native, err := exec.Command(exe, "--rebuild").CombinedOutput()
	if err != nil {
		t.Fatalf("original C rebuild: %v\n%s", err, native)
	}
	var records strings.Builder
	emit := func(src, out *Gia_Man_t, seed uint32, pivot, clean int32) {
		if out == nil {
			t.Fatalf("nil result seed%d pivot%d", seed, pivot)
		}
		fmt.Fprintf(&records, "dup %d %d %d ci%d co%d regs%d objs%d", seed, pivot, clean, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, out.nObjs)
		for bits := 0; bits < 16; bits++ {
			values := cofReviewValues(out, bits)
			packed := 0
			for i := int32(0); i < Gia_ManCoNum(out); i++ {
				if values[Gia_ManCo(out, i).id] {
					packed |= 1 << i
				}
			}
			fmt.Fprintf(&records, " %d", packed)
		}
		records.WriteByte('\n')
		Gia_ManStop(out)
	}
	for seed := uint32(0); seed < 64; seed++ {
		src := cofReviewGraph(seed)
		Gia_ManHashStop(src)
		Gia_ManSetRegNum(src, int32(seed%3))
		for _, obj := range src.pObjs {
			if Gia_ObjIsCand(obj) == 0 {
				continue
			}
			emit(src, Gia_ManDupCofInt(src, obj.id), seed, obj.id, 0)
			emit(src, Gia_ManDupCof(src, obj.id), seed, obj.id, 1)
		}
		for _, limit := range []int32{0, 1, 2, 4, 1000} {
			emit(src, Gia_ManDupCofAll(src, limit, 0), seed, limit, 2)
		}
		Gia_ManStop(src)
	}
	want, got := strings.Split(strings.TrimSpace(string(native)), "\n"), strings.Split(strings.TrimSpace(records.String()), "\n")
	if len(got) != len(want) {
		t.Fatalf("rebuild rows Go%d C%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record%d\nGo %s\n C %s", i, got[i], want[i])
		}
	}
	t.Logf("%d native cofactor rebuilds match sizes, interfaces and exhaustive outputs", len(got))
}

func cofReviewReportGraph(count int, chain bool) *Gia_Man_t {
	g := Gia_ManStart(int32(count + 32))
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	for i := 0; i < count; i++ {
		other := c
		if i&1 != 0 {
			other = a
		}
		node := Gia_ManAppendAnd(g, root, other)
		if chain {
			root = node
		}
	}
	Gia_ManAppendCo(g, root)
	Gia_ManAppendCo(g, root^1)
	return g
}
func TestCofactorNativeFanioReport(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cof_review.c")
	native, err := exec.Command(exe, "--report").CombinedOutput()
	if err != nil {
		t.Fatalf("C fanio report: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		counts := []int{2, 9, 10, 11, 19, 20, 99, 100, 101, 999, 1000, 10000}
		for _, chain := range []bool{false, true} {
			sizes := counts
			if chain {
				sizes = counts[:7]
			}
			for _, count := range sizes {
				g := cofReviewReportGraph(count, chain)
				p := Cof_ManCreateLogicSimple(g)
				fmt.Printf("report %d %d\n", count, giaBool(chain))
				Cof_ManPrintFanio(p)
				Cof_ManStop(p)
				Gia_ManStop(g)
			}
		}
	})
	if got != string(native) {
		t.Fatalf("fanio report differs\nGo:\n%s\nC:\n%s", got, native)
	}
}

func TestCofactorNativeFanioDriver(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cof_review.c")
	native, err := exec.Command(exe, "--driver").CombinedOutput()
	if err != nil {
		t.Fatalf("C fanio driver: %v\n%s", err, native)
	}
	g := cofReviewGraph(3)
	defer Gia_ManStop(g)
	Gia_ManHashStop(g)
	got := hashReviewStdout(t, func() { Gia_ManPrintFanio(g, 3) })
	// C uses a packed arena and CPU clock. Go reports its own payload and wall
	// clock; the fields and surrounding source report must still be preserved.
	variable := regexp.MustCompile(`Memory for logic network =\s*([0-9.]+) MB  Time =\s*([0-9.]+) sec`)
	for _, text := range []string{string(native), got} {
		fields := variable.FindStringSubmatch(text)
		if fields == nil {
			t.Fatalf("missing source memory/time fields: %s", text)
		}
		for _, field := range fields[1:] {
			n, err := strconv.ParseFloat(field, 64)
			if err != nil || n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
				t.Fatalf("invalid statistic %q", field)
			}
		}
	}
	if variable.ReplaceAllString(got, "<memory/time>") != variable.ReplaceAllString(string(native), "<memory/time>") {
		t.Fatalf("driver report differs\nGo:\n%s\nC:\n%s", got, native)
	}
}
