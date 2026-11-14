package gia

import (
	"fmt"
	"math"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func embeddingPartialComponentReview(t *testing.T) (*Emb_Man_t, int32) {
	t.Helper()
	g := Gia_ManStart(1)
	a, b, unused := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, root)
	Gia_ManAppendCo(g, 0)
	p := Emb_ManStartSimple(g)
	Emb_ManSetValue(p)
	Gia_ManRandom(1)
	Emb_ManComputeDimensions(p, 4)
	if p.nReached != p.nObjs-1 {
		t.Fatalf("reached %d of %d vertices, want only unused CI excluded", p.nReached, p.nObjs)
	}
	return p, Gia_ObjValue(Gia_ManObj(g, unused>>1))
}

func TestEmbeddingDistancesMatchOriginalAndGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "embedding_distances.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C embedding BFS: %v: %s", err, want)
	}
	var records strings.Builder
	for cluster := 0; cluster < 2; cluster++ {
		for regs := int32(0); regs < 2; regs++ {
			g := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b)
			y := Gia_ManAppendAnd(g, a, x)
			Gia_ManAppendCo(g, y)
			Gia_ManAppendCo(g, a)
			Gia_ManSetRegNum(g, regs)
			start := Emb_ManStartSimple
			if cluster != 0 {
				start = Emb_ManStart
			}
			p := start(g)
			Emb_ManSetValue(p)
			starts := Vec_IntAlloc(2)
			dist := make([]float32, p.nObjs)
			for first := int32(0); first < p.nObjs; first++ {
				for second := first; second < p.nObjs; second++ {
					Vec_IntClear(starts)
					Vec_IntPush(starts, first)
					if second != first {
						Vec_IntPush(starts, second)
					}
					for i := range dist {
						dist[i] = -1
					}
					far := Emb_ManFindDistances(p, starts, dist)
					fmt.Fprintf(&records, "%d %d %d %d %d %d %d", cluster, regs, first, second, p.nReached, p.nDistMax, far.Value)
					for _, d := range dist {
						fmt.Fprintf(&records, " %.0f", d)
					}
					records.WriteByte('\n')
				}
			}
		}
	}
	if records.String() != string(want) {
		t.Fatalf("Go BFS:\n%sC BFS:\n%s", records.String(), want)
	}
}

func TestEmbeddingDimensionsMarkUnreachableVertices(t *testing.T) {
	p, unused := embeddingPartialComponentReview(t)
	for d := int32(0); d < 4; d++ {
		if got := Emb_ManVec(p, d)[unused]; got != embInfinity {
			t.Fatalf("dimension %d: unreachable input distance=%g, want infinity sentinel", d, got)
		}
	}
}

func TestEmbeddingCovarianceExcludesUnreachableVertices(t *testing.T) {
	p, unused := embeddingPartialComponentReview(t)
	raw := slices.Clone(p.pVecs)
	Emb_ManComputeCovariance(p, 4)
	for i := int32(0); i < 4; i++ {
		for j := int32(0); j < 4; j++ {
			// Compute the centered sum from raw moments over the reachable
			// population, independently of the production centering loop.
			var sx, sy, sxy float64
			for id := int32(0); id < p.nObjs; id++ {
				if id == unused {
					continue
				}
				x, y := float64(raw[i*p.nObjs+id]), float64(raw[j*p.nObjs+id])
				sx += x
				sy += y
				sxy += x * y
			}
			want := sxy - sx*sy/float64(p.nReached)
			if got := float64(p.pMatr[i][j]); math.Abs(got-want) > 1e-4 {
				t.Fatalf("covariance[%d][%d]=%g, reachable-population sum=%g", i, j, got, want)
			}
		}
		if Emb_ManVec(p, i)[unused] != 0 {
			t.Fatal("unreachable vertex acquired a centered coordinate")
		}
	}
}
