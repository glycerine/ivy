package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func dfsReviewGraph(seed uint32) *Gia_Man_t {
	p := Gia_ManStart(64)
	next := func() uint32 { seed = 1664525*seed + 1013904223; return seed }
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(p)
	}
	for i := 0; i < 12; i++ {
		a := int32(1 + next()%uint32(p.nObjs-1))
		b := int32(1 + next()%uint32(p.nObjs-1))
		if a == b {
			b = b%(p.nObjs-1) + 1
		}
		pa, pb := int32(next()&1), int32(next()&1)
		Gia_ManAppendAnd(p, 2*a+pa, 2*b+pb)
	}
	for i := 0; i < 6; i++ {
		a := int32(1 + next()%17)
		Gia_ManAppendCo(p, 2*a+int32(next()&1))
	}
	Gia_ManSetRegNum(p, 2)
	return p
}

func TestDfsSupportMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "dfs_support_review.c")
	source, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("original DFS: %v\n%s", err, source)
	}
	var out strings.Builder
	emit := func(name string, v *Vec_Int_t) {
		fmt.Fprint(&out, name)
		for _, id := range Vec_IntArray(v) {
			fmt.Fprintf(&out, " %d", id)
		}
		out.WriteByte('\n')
	}
	marks := func(name string, p *Gia_Man_t) {
		fmt.Fprint(&out, name)
		for id := int32(0); id < p.nObjs; id++ {
			if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
				fmt.Fprintf(&out, " %d", id)
			}
		}
		out.WriteByte('\n')
	}
	for seed := uint32(1); seed <= 32; seed++ {
		p, v := dfsReviewGraph(seed), Vec_IntAlloc(32)
		for mask := 0; mask < 64; mask++ {
			var roots []int32
			for i := int32(0); i < 6; i++ {
				if mask&(1<<i) != 0 {
					roots = append(roots, Gia_ObjId(p, Gia_ManCo(p, i)))
				}
			}
			n := int32(len(roots))
			fmt.Fprintf(&out, "case %d %d\n", seed, mask)
			Gia_ManCollectCis(p, roots, n, v)
			emit("cis", v)
			fmt.Fprintf(&out, "supp %d\n", Gia_ManSuppSize(p, roots, n))
			marks("suppmarks", p)
			fmt.Fprintf(&out, "cone %d\n", Gia_ManConeSize(p, roots, n))
			marks("conemarks", p)
			emit("nodescis", Gia_ManCollectNodesCis(p, roots, n))
			Gia_ManIncrementTravId(p)
			Gia_ManCollectAnds(p, roots, n, v, nil)
			emit("ands", v)
			Gia_ManCollectAnds(p, roots, n, v, nil)
			emit("shared", v)
			leaves := Vec_IntAlloc(2)
			Vec_IntPush(leaves, 6)
			Vec_IntPush(leaves, 7)
			Gia_ManIncrementTravId(p)
			Gia_ManCollectAnds(p, roots, n, v, leaves)
			emit("leaves", v)
		}
		for mask := 0; mask < 16; mask++ {
			var pos []int32
			for i := int32(0); i < 4; i++ {
				if mask&(1<<i) != 0 {
					pos = append(pos, i)
				}
			}
			fmt.Fprintf(&out, "seq %d %d\n", seed, mask)
			emit("objects", Gia_ManCollectSeq(p, pos, int32(len(pos))))
		}
		Gia_ManStaticFanoutStart(p)
		for i := int32(6); i < 18; i++ {
			for j := int32(6); j < 18; j++ {
				roots := Vec_IntAlloc(2)
				Vec_IntPush(roots, i)
				Vec_IntPush(roots, j)
				fmt.Fprintf(&out, "direction %d %d %d\n", seed, i, j)
				Gia_ManCollectTfi(p, roots, v)
				emit("tfi", v)
				Gia_ManCollectTfo(p, roots, v)
				emit("tfo", v)
			}
		}
		emit("all", Gia_ManCollectAndsAll(p))
		for i := int32(1); i < 18; i++ {
			fmt.Fprintf(&out, "one %d %d\n", i, Gia_ManSuppSizeOne(p, Gia_ManObj(p, i)))
		}
		levels := Gia_ManLevelize(p)
		for i := 0; i < vec.Vec_VecSize(levels); i++ {
			fmt.Fprintf(&out, "level %d", i)
			for _, entry := range vec.Vec_PtrArray(vec.Vec_VecEntry(levels, i)) {
				fmt.Fprintf(&out, " %d", Gia_ObjId(p, entry.(*Gia_Obj_t)))
			}
			out.WriteByte('\n')
		}
		emit("reverse", Gia_ManOrderReverse(p))
		reverse := Gia_ManLevelizeR(p)
		for i := 0; i < Vec_WecSize(reverse); i++ {
			emit(fmt.Sprintf("rlevel %d", i), Vec_WecEntry(reverse, i))
		}
		Gia_ManStop(p)
	}
	got, want := strings.Split(out.String(), "\n"), strings.Split(string(source), "\n")
	if len(got) != len(want) {
		t.Fatalf("trace lengths: Go %d, C %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("trace line %d: Go %q; C %q", i+1, got[i], want[i])
		}
	}
	t.Run("diagnostics", func(t *testing.T) {
		source, err := testoracle.Command(t, exe, "reports").CombinedOutput()
		if err != nil {
			t.Fatalf("original reports: %v\n%s", err, source)
		}
		p := Gia_ManStart(8)
		defer Gia_ManStop(p)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
		got := hashReviewStdout(t, func() {
			Gia_ManCollectTest(p)
			if Gia_ManSuppSizeTest(p) != 1 {
				t.Error("expected one node with support <=16")
			}
			Gia_ManCollectSeqTest(p)
		})
		// Clock values vary, but labels, padding and progress output must match.
		elapsed := regexp.MustCompile(`[0-9]+\.[0-9]{2}`)
		if normalized := elapsed.ReplaceAllString(got, "0.00"); normalized != string(source) {
			t.Fatalf("reports: Go %q; C %q", normalized, source)
		}
	})
}

func TestDfsSupportPreservesSourceTraversalMarks(t *testing.T) {
	for _, empty := range []bool{false, true} {
		name := "output root"
		if empty {
			name = "no roots"
		}
		t.Run(name, func(t *testing.T) {
			p := Gia_ManStart(8)
			defer Gia_ManStop(p)
			a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			x := Gia_ManAppendAnd(p, a, b)
			Gia_ManAppendCo(p, x)
			co := Gia_ObjId(p, Gia_ManCo(p, 0))
			roots, want := []int32{co}, int32(2)
			if empty {
				roots, want = nil, 0
			}
			if got := Gia_ManSuppSize(p, roots, int32(len(roots))); got != want {
				t.Fatalf("support = %d, want %d", got, want)
			}
			if Gia_ObjIsTravIdCurrentId(p, 0) == 0 {
				t.Error("support wrapper must premark constant, even without roots")
			}
			if Gia_ObjIsTravIdCurrentId(p, co) != 0 {
				t.Error("support wrapper must visit the CO's driver without marking the CO")
			}
		})
	}
}

func TestDfsRecursiveHelpersShareCallerState(t *testing.T) {
	p := Gia_ManStart(8)
	defer Gia_ManStop(p)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManStaticFanoutStart(p)
	for _, tc := range []struct {
		name string
		call func(*Vec_Int_t)
		want []int32
	}{
		{"CIs", func(v *Vec_Int_t) { Gia_ManCollectCis_rec(p, Gia_ObjFromLit(p, y), v) }, []int32{c >> 1}},
		{"nodes/CIs", func(v *Vec_Int_t) { Gia_ManCollectNodesCis_rec(p, Gia_ObjFromLit(p, y), v) }, []int32{c >> 1, y >> 1}},
		{"ANDs", func(v *Vec_Int_t) { Gia_ManCollectAnds_rec(p, y>>1, v) }, []int32{y >> 1}},
		{"TFI", func(v *Vec_Int_t) { Gia_ManCollectTfi_rec(p, y>>1, v) }, []int32{y >> 1}},
		{"TFO", func(v *Vec_Int_t) { Gia_ManCollectTfo_rec(p, y>>1, v) }, []int32{y >> 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			Gia_ManIncrementTravId(p)
			Gia_ObjSetTravIdCurrentId(p, 0)
			Gia_ObjSetTravIdCurrentId(p, x>>1)
			epoch := p.nTravIds
			v := Vec_IntAlloc(8)
			Vec_IntPush(v, -7)
			tc.call(v)
			tc.call(v)
			want := append([]int32{-7}, tc.want...)
			assertIDs(t, v, want...)
			if p.nTravIds != epoch {
				t.Fatal("recursive helper changed the caller's epoch")
			}
		})
	}
	for _, tc := range []struct {
		name string
		call func() int32
	}{
		{"support", func() int32 { return Gia_ManSuppSize_rec(p, Gia_ObjFromLit(p, y)) }},
		{"cone", func() int32 { return Gia_ManConeSize_rec(p, Gia_ObjFromLit(p, y)) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			Gia_ManIncrementTravId(p)
			Gia_ObjSetTravIdCurrentId(p, 0)
			Gia_ObjSetTravIdCurrentId(p, x>>1)
			epoch := p.nTravIds
			if first, second := tc.call(), tc.call(); first != 1 || second != 0 {
				t.Fatalf("shared counts = %d/%d, want 1/0", first, second)
			}
			if p.nTravIds != epoch {
				t.Fatal("recursive counter changed the caller's epoch")
			}
		})
	}
}
