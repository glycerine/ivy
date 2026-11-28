package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"math"
	"strings"
	"testing"
)

func TestEmbeddingEigenvectorsSeparateDistantVariances(t *testing.T) {
	// These random offsets start near the dominant axis. The source's angle
	// stopping condition accepts that first eigenvector after one iteration.
	for _, small := range []float32{1, 1e-12, 1e-20} {
		for _, offset := range []int{3770, 4539, 7701, 8510} {
			p := &Emb_Man_t{pMatr: Emb_ManMatrAlloc(3), pEigen: Emb_ManMatrAlloc(3)}
			p.pMatr[0][0], p.pMatr[1][1], p.pMatr[2][2] = 1e9, small, small*0.1
			Gia_ManRandom(1)
			for i := 0; i < offset; i++ {
				Gia_ManRandom(0)
			}
			Emb_ManComputeEigenvectors(p, 3, 2)
			first, second := p.pEigen[0], p.pEigen[1]
			var dot, norm, variance float64
			for i := 0; i < 3; i++ {
				dot += float64(first[i]) * float64(second[i])
				norm += float64(second[i]) * float64(second[i])
				variance += float64(second[i]) * float64(second[i]) * float64(p.pMatr[i][i])
			}
			if !(math.Abs(dot) <= 1e-5 && math.Abs(norm-1) <= 1e-5 && math.Abs(variance-float64(small)) <= 0.01*float64(small)) {
				t.Errorf("offset %d: secondary axis %v has dot=%g norm=%g variance=%g, want orthogonal unit axis with variance %g", offset, second, dot, norm, variance, small)
			}
		}
	}
}

func TestEmbeddingEigenvectorsOriginalDeflationWitness(t *testing.T) {
	exe := compileGiaSourceOracle(t, "embedding_eigen.c")
	out, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C eigenvector witness: %v: %s", err, out)
	}
	reader := strings.NewReader(string(out))
	var original [6]float64
	for i := range original {
		if _, err := fmt.Fscan(reader, &original[i]); err != nil {
			t.Fatalf("C eigenvector field %d: %v: %s", i, err, out)
		}
	}
	p := &Emb_Man_t{pMatr: Emb_ManMatrAlloc(3), pEigen: Emb_ManMatrAlloc(3)}
	p.pMatr[0][0], p.pMatr[1][1], p.pMatr[2][2] = 1e9, 1, 0.1
	Gia_ManRandom(1)
	for i := 0; i < 7701; i++ {
		Gia_ManRandom(0)
	}
	Emb_ManComputeEigenvectors(p, 3, 2)
	var sourceVariance, correctedVariance float64
	for i := 0; i < 3; i++ {
		if !(math.Abs(float64(p.pEigen[0][i])-original[i]) <= 1e-6) {
			t.Fatal("principal eigenvector changed from original C")
		}
		sourceVariance += original[3+i] * original[3+i] * float64(p.pMatr[i][i])
		correctedVariance += float64(p.pEigen[1][i]) * float64(p.pEigen[1][i]) * float64(p.pMatr[i][i])
	}
	t.Logf("secondary variance: original C %.9g; corrected Go %.9g; expected 1", sourceVariance, correctedVariance)
	if !(math.Abs(correctedVariance-1) <= 0.01) {
		t.Fatalf("corrected secondary variance=%g, want 1", correctedVariance)
	}
}

func TestEmbeddingEigenvectorsRotatedSpectra(t *testing.T) {
	basis := [4][4]float64{{1, 1, 1, 1}, {1, -1, 1, -1}, {1, 1, -1, -1}, {1, -1, -1, 1}}
	for _, spectrum := range [][4]float64{{64, 16, 4, 0}, {9, 9, 1, 0}, {16, 4, 0, 0}, {16, 0, 0, 0}, {0, 0, 0, 0}} {
		for _, offset := range []int{0, 7, 31, 7701} {
			p := &Emb_Man_t{pMatr: Emb_ManMatrAlloc(4), pEigen: Emb_ManMatrAlloc(4)}
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					for k := 0; k < 4; k++ {
						p.pMatr[i][j] += float32(basis[i][k] * spectrum[k] * basis[j][k] / 4)
					}
				}
			}
			Gia_ManRandom(1)
			for i := 0; i < offset; i++ {
				Gia_ManRandom(0)
			}
			Emb_ManComputeEigenvectors(p, 4, 3)
			for axis := 0; axis < 3; axis++ {
				v := p.pEigen[axis]
				var norm, variance float64
				for i := 0; i < 4; i++ {
					norm += float64(v[i]) * float64(v[i])
					for j := 0; j < 4; j++ {
						variance += float64(v[i]) * float64(p.pMatr[i][j]) * float64(v[j])
					}
				}
				validNorm := math.Abs(norm-1) <= 1e-5 || (spectrum[axis] == 0 && norm == 0)
				if !validNorm || !(math.Abs(variance-spectrum[axis]) <= 0.05*math.Max(1, spectrum[axis])) {
					t.Errorf("spectrum %v offset %d axis %d: variance=%g norm=%g vector=%v", spectrum, offset, axis, variance, norm, v)
				}
				for previous := 0; previous < axis; previous++ {
					var dot float64
					for i := 0; i < 4; i++ {
						dot += float64(v[i]) * float64(p.pEigen[previous][i])
					}
					if !(math.Abs(dot) <= 1e-5) {
						t.Errorf("spectrum %v offset %d axes %d/%d: dot=%g", spectrum, offset, previous, axis, dot)
					}
				}
			}
		}
	}
}

func TestEmbeddingEigenvectorsHighDimensionLowRank(t *testing.T) {
	const dims = 100
	p := &Emb_Man_t{pMatr: Emb_ManMatrAlloc(dims), pEigen: Emb_ManMatrAlloc(dims)}
	// Two dense, orthogonal constant/sign-alternating directions. All other
	// directions have zero variance, despite every matrix entry being nonzero.
	for i := 0; i < dims; i++ {
		for j := 0; j < dims; j++ {
			p.pMatr[i][j] = 0.12
			if i%2 == j%2 {
				p.pMatr[i][j] = 0.20
			}
		}
	}
	Gia_ManRandom(1)
	Emb_ManComputeEigenvectors(p, dims, 3)
	for axis, target := range []float64{16, 4, 0} {
		var norm, variance float64
		for i, v := range p.pEigen[axis] {
			norm += float64(v) * float64(v)
			for j, w := range p.pEigen[axis] {
				variance += float64(v) * float64(p.pMatr[i][j]) * float64(w)
			}
		}
		validNorm := math.Abs(norm-1) <= 1e-5 || (target == 0 && norm == 0)
		if !validNorm || !(math.Abs(variance-target) <= 0.01) {
			t.Errorf("axis %d: norm=%g variance=%g, want %g", axis, norm, variance, target)
		}
		for previous := 0; previous < axis; previous++ {
			var dot float64
			for i, v := range p.pEigen[axis] {
				dot += float64(v) * float64(p.pEigen[previous][i])
			}
			if !(math.Abs(dot) <= 1e-5) {
				t.Errorf("axes %d/%d: dot=%g", previous, axis, dot)
			}
		}
	}
}
