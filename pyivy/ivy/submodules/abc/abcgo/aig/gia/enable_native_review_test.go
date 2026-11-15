package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func enableReviewGraph(kind, phase int32) *Gia_Man_t {
	p := Gia_ManStart(32)
	c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	r := [2]int32{Gia_ManAppendCi(p), Gia_ManAppendCi(p)}
	var next [2]int32
	for i := range r {
		if kind != 0 {
			next[i] = Gia_ManAppendMux(p, c^(phase&1), d^((phase>>1)&1), r[i]^((phase>>2)&1)) ^ ((phase >> 2) & 1)
		} else {
			next[i] = Gia_ManAppendAnd(p, c^(phase&1), r[i]^((phase>>1)&1)) ^ ((phase >> 2) & 1)
		}
	}
	Gia_ManAppendCo(p, r[0])
	Gia_ManAppendCo(p, d)
	for _, v := range next {
		Gia_ManAppendCo(p, v)
	}
	Gia_ManSetRegNum(p, 2)
	return p
}
func enableReviewDump(name string, p *Gia_Man_t) {
	fmt.Printf("%s objects%d cis%d cos%d regs%d\n", name, Gia_ManObjNum(p), Gia_ManCiNum(p), Gia_ManCoNum(p), Gia_ManRegNum(p))
	for mask := 0; mask < 1<<Gia_ManCiNum(p); mask++ {
		in := make([]bool, Gia_ManCiNum(p))
		for i := range in {
			in[i] = mask&(1<<i) != 0
		}
		out, e := p.Evaluate(in)
		if e != nil {
			panic(e)
		}
		fmt.Printf("truth %d", mask)
		for _, v := range out {
			if v {
				fmt.Print(" 1")
			} else {
				fmt.Print(" 0")
			}
		}
		fmt.Println()
	}
	fmt.Print("classes")
	if p.VFlopClasses != nil {
		for _, v := range Vec_IntArray(p.VFlopClasses) {
			fmt.Printf(" %d", v)
		}
	}
	fmt.Println()
}

func enableReviewSupers(p *Gia_Man_t) {
	for id := int32(1); id < p.nObjs; id++ {
		obj := p.pObjs[id]
		if Gia_ObjIsCand(obj) == 0 {
			continue
		}
		for mode := 0; mode < 4; mode++ {
			v := Vec_IntAlloc(8)
			prefix := int32(1)
			if mode == 3 {
				prefix = id
			}
			Vec_IntPush(v, prefix)
			switch mode {
			case 0, 3:
				Gia_CollectSuper_rec(p, obj, v)
			case 1:
				Gia_CollectSuper(p, obj, v)
			case 2:
				Gia_CollectSuper_rec(p, Gia_Not(obj), v)
			}
			fmt.Printf("super %d %d", id, mode)
			for _, value := range Vec_IntArray(v) {
				fmt.Printf(" %d", value)
			}
			fmt.Println()
		}
	}
}

func enableReviewCompose(p, frames *Gia_Man_t, nf int32) {
	clean := Gia_ManCleanup(frames)
	signals := Gia_ManDetectSeqSignalsWithFanout(p, 1, 0)
	mapped := Gia_ManTransferFrames(p, frames, nf, clean, signals)
	fmt.Print("transfer")
	for _, value := range Vec_IntArray(mapped) {
		fmt.Printf(" %d", value)
	}
	fmt.Println()
	Gia_ManStop(clean)
	q := Gia_ManUnrollAndCofactor(p, nf, 1, 0)
	enableReviewDump("cofactored", q)
	Gia_ManStop(q)
	fmt.Printf("copies_after_cofactor %d\n", Vec_IntSize(&p.vCopies))
}

func TestEnableOrdinaryTransformsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "enable_review.c")
	source, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("original enable transforms: %v\n%s", err, source)
	}
	observed := hashReviewStdout(t, func() {
		for kind := int32(0); kind < 2; kind++ {
			for phase := int32(0); phase < 8; phase++ {
				p := enableReviewGraph(kind, phase)
				fmt.Printf("case %d %d\n", kind, phase)
				Gia_ManDetectSeqSignals(p, 1, 1)
				enableReviewSupers(p)
				for _, limit := range []int32{0, 1, 3} {
					s := Gia_ManDetectSeqSignalsWithFanout(p, limit, 1)
					fmt.Printf("signals %d", limit)
					for _, v := range Vec_IntArray(s) {
						fmt.Printf(" %d", v)
					}
					fmt.Println()
				}
				for nf := int32(1); nf <= 3; nf++ {
					fmt.Printf("frames %d\n", nf)
					q := Gia_ManUnrollInit(p, nf)
					enableReviewDump("unrolled", q)
					fmt.Print("copies")
					for f := int32(0); f < nf; f++ {
						for id := int32(0); id < Gia_ManObjNum(p); id++ {
							fmt.Printf(" %d", Gia_ObjCopyF(p, f, Gia_ManObj(p, id)))
						}
					}
					fmt.Println()
					enableReviewCompose(p, q, nf)
					Gia_ManStop(q)
				}
				q := Gia_ManRemoveEnables(p)
				enableReviewDump("removed", q)
				Gia_ManStop(q)
				if kind != 0 {
					q = Gia_ManRemoveEnables2(p)
					enableReviewDump("removed2", q)
					Gia_ManStop(q)
				}
				Gia_ManStop(p)
			}
		}
	})
	got, want := strings.Split(observed, "\n"), strings.Split(string(source), "\n")
	if len(got) != len(want) {
		t.Fatalf("trace lengths: Go %d, C %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("trace line %d: Go %q; C %q", i+1, got[i], want[i])
		}
	}
	t.Run("high frequency", func(t *testing.T) {
		source, err := exec.Command(exe, "high").CombinedOutput()
		if err != nil {
			t.Fatalf("original high-frequency reports: %v\n%s", err, source)
		}
		observed := hashReviewStdout(t, func() {
			p := Gia_ManStart(128)
			defer Gia_ManStop(p)
			c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			var registers [13]int32
			for i := range registers {
				registers[i] = Gia_ManAppendCi(p)
			}
			for _, r := range registers {
				Gia_ManAppendCo(p, Gia_ManAppendMux(p, c, d, r))
			}
			Gia_ManSetRegNum(p, 13)
			Gia_ManDetectSeqSignals(p, 1, 1)
			signals := Gia_ManDetectSeqSignalsWithFanout(p, 10, 1)
			fmt.Print("signals")
			for _, value := range Vec_IntArray(signals) {
				fmt.Printf(" %d", value)
			}
			fmt.Println()
			q := Gia_ManRemoveEnables(p)
			defer Gia_ManStop(q)
			fmt.Print("classes")
			for _, value := range Vec_IntArray(q.VFlopClasses) {
				fmt.Printf(" %d", value)
			}
			fmt.Println()
		})
		if observed != string(source) {
			t.Fatalf("high-frequency report: Go %q; C %q", observed, source)
		}
	})
}
