package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func frontierReviewBuffers() *Gia_Man_t {
	g := Gia_ManStart(32)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendBuf(g, a)
	z := Gia_ManAppendBuf(g, b)
	// These dangling buffers force slot reuse before the final use of a.
	for i := 0; i < 3; i++ {
		Gia_ManAppendBuf(g, b)
	}
	y := Gia_ManAppendAnd(g, a, z)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, y)
	return g
}

func TestFrontierBufferSlotReuse(t *testing.T) {
	g := frontierReviewBuffers()
	defer Gia_ManStop(g)
	out := Gia_ManFront(g)
	defer Gia_ManStop(out)
	Gia_ManFrontTransform(out)
	cofReviewCheckRebuild(t, g, out)
}

func TestFrontierBufferReferenceScratch(t *testing.T) {
	g := frontierReviewBuffers()
	defer Gia_ManStop(g)
	if got := Gia_ManCrossCutSimple(g); got != 4 {
		t.Errorf("cross-cut=%d, want4 (a, x, b, z before releasing b)", got)
	}
	for id, obj := range g.pObjs {
		if obj.Value != 0 {
			t.Errorf("object%d retains reference count%d after complete sweep", id, obj.Value)
		}
	}
}

// last-use intervals provide an independent cross-cut count without reference
// increments/decrements or the frontier allocator.
func frontierReviewLastUseCut(g *Gia_Man_t) int32 {
	last := make([]int32, g.nObjs)
	for i := range last {
		last[i] = -1
	}
	for id := int32(1); id < g.nObjs; id++ {
		a, b, c, kind := g.Gate(id)
		var deps []int32
		switch kind {
		case 2, 3:
			deps = []int32{a, b}
		case 4:
			deps = []int32{a, b, c}
		case 5, 6:
			deps = []int32{a}
		}
		for _, lit := range deps {
			last[lit>>1] = id
		}
	}
	var maximum int32
	for at := int32(0); at < g.nObjs; at++ {
		var live int32
		for id := int32(0); id <= at; id++ {
			if last[id] >= at && last[id] > id {
				live++
			}
		}
		maximum = max(maximum, live)
	}
	return maximum
}

func frontierReviewGraph(seed uint32) *Gia_Man_t {
	g := Gia_ManStart(128)
	pool := []int32{0}
	rng := seed
	next := func() uint32 { rng = rng*1664525 + 1013904223; return rng >> 8 }
	for i := 0; i < 4; i++ {
		pool = append(pool, Gia_ManAppendCi(g))
	}
	for i := 0; i < 24; i++ {
		a, b := int(next())%len(pool), int(next())%len(pool)
		if a == b {
			b = (b + 1) % len(pool)
		}
		x, y := pool[a]^int32(next()&1), pool[b]^int32(next()&1)
		pool = append(pool, Gia_ManAppendAnd(g, x, y))
	}
	Gia_ManAppendCo(g, pool[len(pool)-1])
	Gia_ManAppendCo(g, pool[len(pool)-3]^1)
	Gia_ManAppendCo(g, 0)
	Gia_ManAppendCi(g)
	Gia_ManSetRegNum(g, 1)
	return g
}

func TestFrontierNativeSlotsAndRoundTrip(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frontier_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C frontier: %v\n%s", err, native)
	}
	var got strings.Builder
	dump := func(g *Gia_Man_t, tag string) {
		fmt.Fprintf(&got, "%s %d %d", tag, g.nObjs, g.nFront)
		for _, o := range g.pObjs {
			fmt.Fprintf(&got, " [%d %d %d %d %d %d]", o.iDiff0, o.iDiff1, o.fCompl0, o.fCompl1, o.fTerm, o.Value)
		}
		got.WriteByte('\n')
	}
	for seed := uint32(0); seed < 64; seed++ {
		g := frontierReviewGraph(seed)
		cut := Gia_ManCrossCutSimple(g)
		if want := frontierReviewLastUseCut(g); cut != want {
			t.Fatalf("seed%d cut%d want%d", seed, cut, want)
		}
		fmt.Fprintf(&got, "seed %d cut %d\n", seed, cut)
		out := Gia_ManFront(g)
		dump(out, "front")
		Gia_ManFrontTransform(out)
		dump(out, "restored")
		cofReviewCheckRebuild(t, g, out)
		Gia_ManStop(out)
		Gia_ManStop(g)
	}
	esopReviewCompare(t, native, got.String())
}

func TestFrontierSlotAllocator(t *testing.T) {
	for n := 1; n <= 8; n++ {
		for mask := 0; mask < (1<<n)-1; mask++ {
			for start := 0; start < n; start++ {
				front := make([]byte, n)
				want, distance := -1, n
				for i := range front {
					front[i] = byte(mask >> i & 1)
					if front[i] == 0 && (i-start+n)%n < distance {
						want, distance = i, (i-start+n)%n
					}
				}
				if got := Gia_ManFrontFindNext(front, int32(n), int32(start)); int(got) != want {
					t.Fatalf("n%d mask%d start%d slot%d want%d", n, mask, start, got, want)
				}
				for i, used := range front {
					if used != byte(mask>>i&1)|byte(giaBool(i == want)) {
						t.Fatal("allocator changed an unrelated slot")
					}
				}
			}
		}
	}
	for _, tc := range []struct {
		front []byte
		n, at int32
	}{{nil, 0, 0}, {[]byte{0}, 2, 0}, {[]byte{0}, 1, -1}, {[]byte{0}, 1, 1}, {[]byte{1, 1}, 2, 0}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid/full frontier did not panic")
				}
			}()
			Gia_ManFrontFindNext(tc.front, tc.n, tc.at)
		}()
	}
}

func TestFrontierNativeBufferCounterexample(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frontier_review.c")
	native, err := exec.Command(exe, "--buffer").CombinedOutput()
	if err != nil {
		t.Fatalf("C buffer probe: %v\n%s", err, native)
	}
	rows := strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(rows) != 4 {
		t.Fatalf("expected four source truth rows, got%s", native)
	}
	wrong := 0
	for bits, row := range rows {
		var input, before, after int
		if n, err := fmt.Sscanf(row, "buffer %d %d %d", &input, &before, &after); err != nil || n != 3 || input != bits {
			t.Fatalf("invalid native row%q", row)
		}
		want := bits & 1
		if bits == 3 {
			want |= 2
		}
		if before != want {
			t.Fatalf("native input graph differs from formula at%d", bits)
		}
		if after != want {
			wrong++
		}
	}
	if wrong == 0 {
		t.Fatal("original C did not reproduce the demonstrated buffer defect")
	}
	g := frontierReviewBuffers()
	defer Gia_ManStop(g)
	out := Gia_ManFront(g)
	defer Gia_ManStop(out)
	Gia_ManFrontTransform(out)
	cofReviewCheckRebuild(t, g, out)
}
