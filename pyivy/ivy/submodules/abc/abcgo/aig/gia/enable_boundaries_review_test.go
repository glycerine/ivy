package gia

import (
	"os/exec"
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEnableSupergateKeepsTypedNodesAtomic(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendXorReal(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	super := Vec_IntAlloc(8)
	Gia_CollectSuper_rec(p, Gia_ObjFromLit(p, y), super)
	assertIDs(t, super, c>>1, x>>1)
	Gia_CollectSuper(p, Gia_ObjFromLit(p, x), super)
	assertIDs(t, super, x>>1)
}

func TestEnableDetectionRejectsFalseTypedMux(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	p.EnableMuxes()
	c, d, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	a, b := Gia_ManAppendAnd(p, c, r), Gia_ManAppendAnd(p, c^1, d)
	root := Gia_ManAppendMuxReal(p, d, a^1, b^1)
	Gia_ManAppendCo(p, root)
	Gia_ManSetRegNum(p, 1)
	if got := giaDetectSeqSignals(p); got.haveEnable != 0 {
		t.Fatalf("typed MUX has no hold-state branch but was classified as an enable: %d", got.haveEnable)
	}
}

func TestEnableRecursiveSupergateSharedCone(t *testing.T) {
	const marker = "ABCGO_ENABLE_SUPER_SHARED_CHILD"
	if os.Getenv(marker) == "1" {
		p := Gia_ManStart(128)
		defer Gia_ManStop(p)
		root := Gia_ManAppendCi(p)
		for i := 0; i < 60; i++ {
			root = Gia_ManAppendBuf(p, root)
		}
		v := Vec_IntAlloc(1)
		Gia_CollectSuper_rec(p, Gia_ObjFromLit(p, root), v)
		assertIDs(t, v, 1)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestEnableRecursiveSupergateSharedCone$")
	cmd.Env = append(os.Environ(), marker+"=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shared DAG must not be expanded as an exponential tree: %v (context %v)\n%s", err, ctx.Err(), out)
	}
}

func TestEnableSignalReportUsesObjectCount(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("oversized frequency buffer panicked: %v", r)
		}
	}()
	p := Gia_ManStart(8)
	defer Gia_ManStop(p)
	Gia_ManAppendCi(p)
	Gia_ManCreateRefs(p)
	freq := make([]int32, p.nObjs+2)
	freq[1], freq[p.nObjs], freq[p.nObjs+1] = 2, 99, 99
	got := hashReviewStdout(t, func() { Gia_ManPrintSignals(p, freq, "test") })
	if !strings.HasPrefix(got, "test (total = 1  driven = 2)\n") {
		t.Fatalf("frequency buffer tail leaked into the report: %q", got)
	}
}

func TestEnableRemovalRejectsFalseTypedMuxPatterns(t *testing.T) {
	for _, operation := range []struct {
		name string
		copy func(*Gia_Man_t) *Gia_Man_t
	}{{"remove", Gia_ManRemoveEnables}, {"remove2", Gia_ManRemoveEnables2}} {
		for _, kind := range []string{"xor root", "mux root", "xor branch"} {
			for phase := int32(0); phase < 2; phase++ {
				t.Run(operation.name+"/"+kind+string(rune('0'+phase)), func(t *testing.T) {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("valid typed graph panicked: %v", r)
						}
					}()
					p := Gia_ManStart(24)
					defer Gia_ManStop(p)
					c, d, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
					a, b := Gia_ManAppendAnd(p, c, r), Gia_ManAppendAnd(p, c^1, d)
					var root int32
					switch kind {
					case "xor root":
						root = Gia_ManAppendXorReal(p, a^1, b^1)
					case "mux root":
						p.EnableMuxes()
						root = Gia_ManAppendMuxReal(p, d, a^1, b^1)
					case "xor branch":
						x := Gia_ManAppendXorReal(p, c, r)
						root = Gia_ManAppendAnd(p, x^1, b^1)
					}
					Gia_ManAppendCo(p, r)
					Gia_ManAppendCo(p, root^phase)
					Gia_ManSetRegNum(p, 1)
					q := operation.copy(p)
					defer Gia_ManStop(q)
					for mask := 0; mask < 8; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
						ctrl, data, state := inputs[0], inputs[1], inputs[2]
						left, right := ctrl && state, !ctrl && data
						var want bool
						switch kind {
						case "xor root":
							want = !left != !right
						case "mux root":
							want = !right
							if data {
								want = !left
							}
						case "xor branch":
							want = !(ctrl != state) && !right
						}
						want = want != (phase != 0)
						got, err := q.Evaluate(inputs)
						if err != nil || len(got) != 2 || got[0] != state || got[1] != want {
							t.Errorf("mask%d: outputs %v, error %v, want %t/%t", mask, got, err, state, want)
						}
					}
				})
			}
		}
	}
}

func TestEnableRemovalPreservesLogicalPortOrder(t *testing.T) {
	for _, which := range []string{"inputs", "outputs", "both"} {
		t.Run(which, func(t *testing.T) {
			p := Gia_ManStart(8)
			defer Gia_ManStop(p)
			a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			Gia_ManAppendCo(p, a^1)
			Gia_ManAppendCo(p, b)
			if which != "outputs" {
				Vec_IntWriteEntry(p.vCis, 0, b>>1)
				Vec_IntWriteEntry(p.vCis, 1, a>>1)
				Gia_ObjSetCioId(Gia_ManObj(p, a>>1), 1)
				Gia_ObjSetCioId(Gia_ManObj(p, b>>1), 0)
			}
			if which != "inputs" {
				first, second := Gia_ManCo(p, 0), Gia_ManCo(p, 1)
				Vec_IntWriteEntry(p.vCos, 0, second.id)
				Vec_IntWriteEntry(p.vCos, 1, first.id)
				Gia_ObjSetCioId(first, 1)
				Gia_ObjSetCioId(second, 0)
			}
			q := Gia_ManRemoveEnables(p)
			defer Gia_ManStop(q)
			for mask := 0; mask < 4; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0}
				want, originalErr := p.Evaluate(inputs)
				got, err := q.Evaluate(inputs)
				if err != nil || originalErr != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
					t.Errorf("mask%d: got %v (%v); original %v (%v)", mask, got, err, want, originalErr)
				}
			}
		})
	}
}
