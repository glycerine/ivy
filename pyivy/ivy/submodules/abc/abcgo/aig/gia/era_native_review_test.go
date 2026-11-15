package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func eraReviewGraph(ni, nr, seed int32) *Gia_Man_t {
	p := Gia_ManStart(256)
	var inputs, regs []int32
	for i := int32(0); i < ni; i++ {
		inputs = append(inputs, Gia_ManAppendCi(p))
	}
	for i := int32(0); i < nr; i++ {
		regs = append(regs, Gia_ManAppendCi(p))
	}
	Gia_ManHashAlloc(p)
	a, b := int32(1), int32(0)
	if ni != 0 {
		a, b = inputs[0], inputs[ni-1]
	}
	next := append([]int32(nil), regs...)
	for i := int32(0); i < nr && i < 3; i++ {
		switch (seed + i) % 4 {
		case 0:
			next[i] = regs[i] ^ 1
		case 1:
			next[i] = a
		case 2:
			next[i] = Gia_ManHashAnd(p, regs[i]^1, b^(seed&1))
		case 3:
			next[i] = regs[(i+1)%nr]
		}
	}
	Gia_ManAppendCo(p, regs[0])
	j := int32(0)
	if nr > 1 {
		j = 1
	}
	Gia_ManAppendCo(p, Gia_ManHashAnd(p, a, regs[j]^1)^(seed&1))
	for _, v := range next {
		Gia_ManAppendCo(p, v)
	}
	Gia_ManHashStop(p)
	Gia_ManSetRegNum(p, nr)
	return p
}

func eraReviewInspect(g *Gia_Man_t, limit, miter int32) {
	p := Gia_ManEraCreate(g)
	defer Gia_ManEraFree(p)
	p.nBins, p.pBins = 3, make([]int32, 3)
	s := Gia_ManEraCreateState(p)
	p.pBins[Gia_ManEraStateHash(s.pData, p.nWordsDat, p.nBins)] = s.Num
	result := int32(1)
	for p.iCurState < int32(len(p.vStates))-1 {
		if int32(len(p.vStates)) >= limit {
			result = -1
			break
		}
		p.iCurState++
		s = Gia_ManEraState(p, p.iCurState)
		if p.iCurState > 1 && s.iPrev == 0 {
			continue
		}
		Gia_ManInsertState(p, s)
		Gia_ManPerformOneIter(p)
		if Gia_ManAnalyzeResult(p, s, miter, 1) != 0 && miter != 0 {
			result = 0
			break
		}
	}
	pending := int32(0)
	if p.pStateNew != nil {
		pending = p.pStateNew.Num
	}
	fmt.Printf("manager %d %d %d %d %d %d %d\n", result, p.iCurState, len(p.vStates), p.nBins, p.nWordsSim, pending, Gia_ManCountDepth(p))
	for _, s := range p.vStates[1:] {
		cond := s.Cond
		if s == p.pStateNew {
			cond = 0
		}
		fmt.Printf("state %d %d %d", s.Num, cond, s.iPrev)
		for _, word := range s.pData {
			fmt.Printf(" %d", word)
		}
		found := int32(-1)
		Gia_ManEraHashFind(p, s, &found)
		fmt.Printf(" found%d\n", found)
	}
	fmt.Print("trace")
	if p.vBugTrace != nil {
		for _, v := range Vec_IntArray(p.vBugTrace) {
			fmt.Printf(" %d", v)
		}
	}
	fmt.Println()
	fmt.Print("transitions")
	for _, v := range Vec_IntArray(p.vStgDump) {
		fmt.Printf(" %d", v)
	}
	fmt.Println()
}

func TestExplicitReachabilityNativeStateTables(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_review.c")
	cmd := exec.Command(exe)
	cmd.Dir = t.TempDir()
	native, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("original C: %v\n%s", err, native)
	}
	goText := hashReviewStdout(t, func() {
		for _, ni := range []int32{0, 1, 3, 5, 6} {
			for _, nr := range []int32{1, 3, 33, 65} {
				for seed := int32(0); seed < 4; seed++ {
					for _, limit := range []int32{2, 4, 100} {
						for miter := int32(0); miter < 2; miter++ {
							p := eraReviewGraph(ni, nr, seed)
							fmt.Printf("case %d %d %d %d %d\n", ni, nr, seed, limit, miter)
							result := Gia_ManCollectReachable(p, limit, miter, 0, 0)
							fmt.Printf("result %d\n", result)
							eraReviewInspect(p, limit, miter)
							Gia_ManStop(p)
						}
					}
				}
			}
		}
	})
	// Time formatting is checked separately; compare all algorithmic records here.
	records := func(s string) []string {
		var rows []string
		for _, line := range strings.Split(s, "\n") {
			for _, prefix := range []string{"case ", "result ", "manager ", "state ", "trace", "transitions"} {
				if strings.HasPrefix(line, prefix) {
					rows = append(rows, line)
					break
				}
			}
		}
		return rows
	}
	want, got := records(string(native)), records(goText)
	if len(want) != len(got) {
		t.Fatalf("record counts: Go %d, C %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record %d\nGo: %s\n C: %s", i, got[i], want[i])
		}
	}
	t.Logf("480 native configurations; %d matching algorithmic records", len(got))
}
