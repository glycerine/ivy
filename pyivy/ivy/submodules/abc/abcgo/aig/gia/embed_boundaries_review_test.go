package gia

import (
	"math"
	"testing"
)

func TestEmbeddingRepeatedSeedsCountUniqueVertices(t *testing.T) {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	p := Emb_ManStartSimple(g)
	Emb_ManSetValue(p)
	starts := Vec_IntAlloc(4)
	for _, id := range []int32{Emb_ManCi(p, 0).hHandle, Emb_ManCo(p, 0).hHandle, Emb_ManCi(p, 0).hHandle, Emb_ManCo(p, 0).hHandle} {
		Vec_IntPush(starts, id)
	}
	dist := make([]float32, p.nObjs)
	Emb_ManFindDistances(p, starts, dist)
	if p.nReached != 4 || p.nDistMax != 3 {
		t.Fatalf("repeated seeds: reached %d vertices in %d levels, want 4 in 3", p.nReached, p.nDistMax)
	}
	if dist[Emb_ManCi(p, 0).Value] != 0 || dist[Emb_ManCo(p, 0).Value] != 0 || dist[Emb_ManCi(p, 1).Value] != 2 {
		t.Fatalf("incorrect multisource distances: %v", dist)
	}
}

func TestEmbeddingZeroVectorNormalizationIsFinite(t *testing.T) {
	v := []float32{0, 0, 0, 7}
	Emb_ManVecNormal(v, 3)
	for i, x := range v {
		want := float32(0)
		if i == 3 {
			want = 7
		}
		if x != want {
			t.Fatalf("zero-vector normalization changed element %d to %g, want %g", i, x, want)
		}
	}
}

func TestEmbeddingZeroCovarianceProducesFinitePlacement(t *testing.T) {
	p := &Emb_Man_t{nObjs: 3, nReached: 3, pVecs: make([]float32, 9)}
	Emb_ManComputeCovariance(p, 3)
	Gia_ManRandom(1)
	Emb_ManComputeEigenvectors(p, 3, 2)
	for i := 0; i < 2; i++ {
		for _, v := range p.pEigen[i] {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				t.Fatalf("zero covariance produced nonfinite eigenvector %d: %v", i, p.pEigen[i])
			}
		}
	}
	Emb_ManComputeSolutions(p, 3, 2)
	for _, v := range p.pSols {
		if v != 0 {
			t.Fatalf("zero input coordinates projected to %g", v)
		}
	}
	Emb_ManDerivePlacement(p, 2)
	for d := 0; d < 2; d++ {
		seen := make(map[uint16]bool)
		for id := 0; id < 3; id++ {
			coordinate := p.pPlacement[2*id+d]
			if coordinate > GIA_PLACE_SIZE || seen[coordinate] {
				t.Fatalf("degenerate axis %d has invalid placement %v", d, p.pPlacement)
			}
			seen[coordinate] = true
		}
	}
}

func TestEmbeddingNormalizationFiniteMagnitudeRange(t *testing.T) {
	for _, magnitude := range []float32{math.SmallestNonzeroFloat32, 1e-30, 1e-22, 1, 1e30, math.MaxFloat32} {
		v := []float32{magnitude, -magnitude, 7}
		Emb_ManVecNormal(v, 2)
		want := float32(1 / math.Sqrt(2))
		if math.Abs(float64(v[0]-want)) > 1e-6 || math.Abs(float64(v[1]+want)) > 1e-6 || v[2] != 7 {
			t.Errorf("magnitude %g: normalized vector %v, want [%g %g 7]", magnitude, v, want, -want)
		}
	}
	v := make([]float32, 2048)
	for i := range v {
		v[i] = 1e-20
	}
	Emb_ManVecNormal(v, int32(len(v)))
	var lengthSquared float64
	for _, x := range v {
		lengthSquared += float64(x) * float64(x)
	}
	if math.Abs(lengthSquared-1) > 1e-6 {
		t.Fatalf("accumulated subnormal squares: normalized length squared=%g", lengthSquared)
	}
}

func TestEmbeddingConstantOutputRankDeficientPipeline(t *testing.T) {
	g := Gia_ManStart(1)
	Gia_ManAppendCo(g, 0)
	p := Emb_ManStartSimple(g)
	Emb_ManSetValue(p)
	Gia_ManRandom(1)
	Emb_ManComputeDimensions(p, 3)
	Emb_ManComputeCovariance(p, 3)
	Emb_ManComputeEigenvectors(p, 3, 2)
	Emb_ManComputeSolutions(p, 3, 2)
	for _, value := range p.pSols {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			t.Fatalf("rank-deficient constant-output embedding produced %v", p.pSols)
		}
	}
	Emb_ManDerivePlacement(p, 2)
	for axis := 0; axis < 2; axis++ {
		if p.pPlacement[axis] == p.pPlacement[2+axis] || p.pPlacement[axis] > GIA_PLACE_SIZE || p.pPlacement[2+axis] > GIA_PLACE_SIZE {
			t.Fatalf("invalid rank-deficient placement: %v", p.pPlacement)
		}
	}
}
