package gia

import (
	"fmt"
	"testing"
)

func TestEnableTypedSequentialFrameSemantics(t *testing.T) {
	for _, cofactor := range []bool{false, true} {
		for phase := int32(0); phase < 2; phase++ {
			for frames := int32(1); frames <= 4; frames++ {
				t.Run(fmt.Sprintf("cofactor%t/phase%d/frames%d", cofactor, phase, frames), func(t *testing.T) {
					p := Gia_ManStart(16)
					defer Gia_ManStop(p)
					p.EnableMuxes()
					p.pName, p.PSpec = "stateful", "stateful spec"
					a, b, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
					x := Gia_ManAppendXorReal(p, a, r^phase)
					next := Gia_ManAppendMuxReal(p, b, x, r)
					Gia_ManAppendCo(p, r)
					Gia_ManAppendCo(p, x)
					Gia_ManAppendCo(p, next)
					Gia_ManSetRegNum(p, 1)
					for _, obj := range p.pObjs {
						obj.Value = uint32(1000 + obj.id)
					}
					var q *Gia_Man_t
					if cofactor {
						q = Gia_ManUnrollAndCofactor(p, frames, 1, 0)
						if Vec_IntSize(&p.vCopies) != 0 {
							t.Fatal("cofactoring driver retained its temporary frame map")
						}
					} else {
						q = Gia_ManUnrollInit(p, frames)
						if Vec_IntSize(&p.vCopies) != int(frames*p.nObjs) {
							t.Fatal("unrolling did not publish its complete frame map")
						}
					}
					defer Gia_ManStop(q)
					if q == nil || Gia_ManCiNum(q) != 2*frames || Gia_ManCoNum(q) != 2*frames || q.nRegs != 0 || q.pName != p.pName || q.PSpec != p.PSpec {
						t.Fatal("incorrect frame interface/identity")
					}
					for _, obj := range p.pObjs {
						if obj.Value != uint32(1000+obj.id) {
							t.Fatal("frame transformation clobbered source Value scratch")
						}
					}
					Gia_ManStop(p) // returned graph must own all of its logic
					for mask := 0; mask < 1<<uint(2*frames); mask++ {
						inputs := make([]bool, 2*frames)
						for i := range inputs {
							inputs[i] = mask&(1<<i) != 0
						}
						got, err := q.Evaluate(inputs)
						if err != nil || len(got) != int(2*frames) {
							t.Fatalf("mask%d: outputs %v, error %v", mask, got, err)
						}
						state := false
						for f := 0; f < int(frames); f++ {
							x := inputs[2*f] != (state != (phase != 0))
							if got[2*f] != state || got[2*f+1] != x {
								t.Fatalf("mask%d/frame%d: outputs %v, want %t/%t", mask, f, got[2*f:2*f+2], state, x)
							}
							if inputs[2*f+1] {
								state = x
							}
						}
					}
				})
			}
		}
	}
}

func TestEnableRemovalPreservesUnrecognizedTransitions(t *testing.T) {
	for _, operation := range []struct {
		name string
		copy func(*Gia_Man_t) *Gia_Man_t
	}{{"remove", Gia_ManRemoveEnables}, {"remove2", Gia_ManRemoveEnables2}} {
		for _, kind := range []string{"input", "state", "constant", "and", "mux without hold"} {
			for phase := int32(0); phase < 2; phase++ {
				t.Run(fmt.Sprintf("%s/%s/phase%d", operation.name, kind, phase), func(t *testing.T) {
					p := Gia_ManStart(24)
					defer Gia_ManStop(p)
					var leaves [4]int32
					for i := range leaves {
						leaves[i] = Gia_ManAppendCi(p)
					}
					a, b, c, r := leaves[0], leaves[1], leaves[2], leaves[3]
					var next int32
					switch kind {
					case "input":
						next = a
					case "state":
						next = r
					case "constant":
						next = 0
					case "and":
						next = Gia_ManAppendAnd(p, a, r)
					case "mux without hold":
						next = Gia_ManAppendMux(p, a, b, c)
					}
					Gia_ManAppendCo(p, r)
					Gia_ManAppendCo(p, next^phase)
					Gia_ManSetRegNum(p, 1)
					q := operation.copy(p)
					defer Gia_ManStop(q)
					if operation.name == "remove" {
						assertIDs(t, q.VFlopClasses, 0)
					}
					for mask := 0; mask < 16; mask++ {
						in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
						var want bool
						switch kind {
						case "input":
							want = in[0]
						case "state":
							want = in[3]
						case "and":
							want = in[0] && in[3]
						case "mux without hold":
							want = in[2]
							if in[0] {
								want = in[1]
							}
						}
						want = want != (phase != 0)
						got, err := q.Evaluate(in)
						if err != nil || len(got) != 2 || got[0] != in[3] || got[1] != want {
							t.Fatalf("mask%d: outputs %v (%v), want %t/%t", mask, got, err, in[3], want)
						}
					}
				})
			}
		}
	}
}

func TestEnableTransferFrameMappingBoundaries(t *testing.T) {
	p := Gia_ManStart(8)
	defer Gia_ManStop(p)
	c, d, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	frames, mapped := Gia_ManStart(8), Gia_ManStart(8)
	defer Gia_ManStop(frames)
	defer Gia_ManStop(mapped)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(frames)
		Gia_ManAppendCi(mapped)
	}
	frames.pObjs[1].Value, frames.pObjs[2].Value, frames.pObjs[3].Value, frames.pObjs[4].Value = 0, ^uint32(0), 1, 8
	Vec_IntFill(&p.vCopies, int(3*p.nObjs), -1)
	Gia_ObjSetCopyF(p, 0, Gia_ObjFromLit(p, c), 8)
	Gia_ObjSetCopyF(p, 1, Gia_ObjFromLit(p, c), 8)
	Gia_ObjSetCopyF(p, 0, Gia_ObjFromLit(p, d), 6)
	Gia_ObjSetCopyF(p, 1, Gia_ObjFromLit(p, d), 2)
	Gia_ObjSetCopyF(p, 2, Gia_ObjFromLit(p, d), 4)
	signals := Vec_IntAlloc(4)
	for _, lit := range []int32{c, d, c, r} {
		Vec_IntPush(signals, lit>>1)
	}
	got := Gia_ManTransferFrames(p, frames, 3, mapped, signals)
	assertIDs(t, got, 4, 0) // Value 1 remains the source's constant-one mapping.
	assertIDs(t, signals, c>>1, d>>1, c>>1, r>>1)
}
