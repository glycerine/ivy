package gia

import (
	"math"
	"testing"
)

func TestEmbeddingMatrixOperations(t *testing.T) {
	m := Emb_ManMatrAlloc(2)
	m[0][0], m[0][1], m[1][0], m[1][1] = 3, 1, 1, 2
	x := []float32{3, 4}
	Emb_ManVecNormal(x, 2)
	if math.Abs(float64(Emb_ManVecMultiplyOne(x, x, 2)-1)) > 1e-6 {
		t.Fatalf("normalization: %v", x)
	}
	out := make([]float32, 2)
	Emb_ManVecMultiply(m, x, 2, out)
	if math.Abs(float64(out[0]-2.6)) > 1e-6 || math.Abs(float64(out[1]-2.2)) > 1e-6 {
		t.Fatalf("matrix product: %v", out)
	}
	Emb_ManVecOrthogonolizeOne([]float32{1, 0}, []float32{3, 4}, 2, out)
	if out[0] != 0 || out[1] != 4 {
		t.Fatalf("orthogonalization: %v", out)
	}
	Emb_ManVecCopyOne(out, x, 2)
	if out[0] != x[0] || out[1] != x[1] {
		t.Fatalf("copy: %v", out)
	}
}

func TestEmbeddingCovarianceAndProjection(t *testing.T) {
	p := &Emb_Man_t{nObjs: 3, nReached: 3, pVecs: []float32{1, 2, 3, 3, 2, 1}}
	Emb_ManComputeCovariance(p, 2)
	if p.pMatr[0][0] != 2 || p.pMatr[0][1] != -2 || p.pMatr[1][0] != -2 || p.pMatr[1][1] != 2 {
		t.Fatalf("covariance = %v", p.pMatr)
	}
	p.pEigen[0][0] = 1
	Emb_ManComputeSolutions(p, 2, 1)
	if got := Emb_ManSol(p, 0); got[0] != -1 || got[1] != 0 || got[2] != 1 {
		t.Fatalf("projection = %v", got)
	}
}

func TestEmbeddingPlacementAndRefinement(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, root)
	p := Emb_ManStart(g)
	Emb_ManSetValue(p)
	p.pSols = make([]float32, p.nObjs*2)
	for i := int32(0); i < p.nObjs; i++ {
		p.pSols[i] = float32(i + 1)
		p.pSols[p.nObjs+i] = float32(p.nObjs - i)
	}
	Emb_ManDerivePlacement(p, 2)
	if len(p.pPlacement) != int(2*p.nObjs) || Emb_ManComputeHPWL(p) <= 0 {
		t.Fatal("invalid placement")
	}
	Emb_ManPlacementRefine(p, 2, 0)
	if len(p.pPlacement) != int(2*p.nObjs) || Emb_ManComputeHPWL(p) <= 0 {
		t.Fatal("invalid refined placement")
	}
}
