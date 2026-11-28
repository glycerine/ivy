package gia

import (
	"context"
	"github.com/glycerine/abcgo/internal/testoracle"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEmbeddingPivotStaysInOutputComponent(t *testing.T) {
	for _, cluster := range []bool{false, true} {
		for offset := 0; offset < 64; offset++ {
			g := Gia_ManStart(1)
			a, b, unused := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
			start := Emb_ManStartSimple
			if cluster {
				start = Emb_ManStart
			}
			p := start(g)
			Emb_ManSetValue(p)
			Gia_ManRandom(1)
			for i := 0; i < offset; i++ {
				Gia_ManRandom(0)
			}
			Emb_ManComputeDimensions(p, 5)
			if p.nReached != 4 {
				t.Errorf("cluster=%t offset=%d: reached %d vertices, want the four output-connected vertices", cluster, offset, p.nReached)
				continue
			}
			for _, id := range []int32{0, Gia_ObjValue(Gia_ManObj(g, unused>>1))} {
				if Emb_ManVec(p, 0)[id] != embInfinity {
					t.Fatalf("isolated vertex %d was embedded as part of the output component", id)
				}
			}
			Emb_ManComputeCovariance(p, 5)
			if p.pMatr[0][0] <= 0 {
				t.Fatal("nontrivial output component lost all geometric variation")
			}
		}
	}
}

func TestEmbeddingConnectedDimensionsMatchOriginal(t *testing.T) {
	exe := compileGiaSourceOracle(t, "embedding_dimensions.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C dimensions: %v: %s", err, want)
	}
	var values []float64
	for cluster := 0; cluster < 2; cluster++ {
		for offset := 0; offset < 8; offset++ {
			g := Gia_ManStart(1)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			x := Gia_ManAppendAnd(g, a, b)
			y := Gia_ManAppendAnd(g, x, c)
			Gia_ManAppendCo(g, y)
			Gia_ManAppendCo(g, 0)
			start := Emb_ManStartSimple
			if cluster != 0 {
				start = Emb_ManStart
			}
			p := start(g)
			Emb_ManSetValue(p)
			Gia_ManRandom(1)
			for i := 0; i < offset; i++ {
				Gia_ManRandom(0)
			}
			Emb_ManComputeDimensions(p, 4)
			Emb_ManComputeCovariance(p, 4)
			Emb_ManComputeEigenvectors(p, 4, 1)
			Emb_ManComputeSolutions(p, 4, 1)
			values = append(values, float64(cluster), float64(offset), float64(p.nObjs), float64(p.nReached))
			appendVector := func(v []float32) {
				for _, x := range v {
					values = append(values, float64(x))
				}
			}
			appendVector(p.pVecs)
			for _, row := range p.pMatr {
				appendVector(row)
			}
			appendVector(p.pEigen[0])
			appendVector(p.pSols)
		}
	}
	fields := strings.Fields(string(want))
	if len(fields) != len(values) {
		t.Fatalf("source emitted %d fields, Go emitted %d", len(fields), len(values))
	}
	for i, value := range values {
		expected, err := strconv.ParseFloat(fields[i], 64)
		if err != nil || math.IsNaN(expected) || math.IsNaN(value) || math.IsInf(expected, 0) || math.IsInf(value, 0) || math.Abs(value-expected) > 1e-5 {
			t.Fatalf("field %d: Go %g, source %q (%v)", i, value, fields[i], err)
		}
	}
}

func TestEmbeddingPivotJoinsWireOnlyComponents(t *testing.T) {
	g := Gia_ManStart(1)
	Gia_ManAppendCo(g, Gia_ManAppendCi(g))
	Gia_ManAppendCo(g, Gia_ManAppendCi(g))
	p := Emb_ManStartSimple(g)
	Emb_ManSetValue(p)
	Gia_ManRandom(1)
	Emb_ManComputeDimensions(p, 4)
	if p.nReached != p.nObjs {
		t.Fatalf("joined wire components: reached %d of %d vertices", p.nReached, p.nObjs)
	}
	Emb_ManComputeCovariance(p, 4)
	if p.pMatr[0][0] <= 0 {
		t.Fatal("joined wire components lost all geometric variation")
	}
}

// Exercise rejection sampling in a child process: a regression must fail
// within a deadline rather than hanging the whole test suite indefinitely.
func TestEmbeddingDegeneratePivotTerminates(t *testing.T) {
	if kind := os.Getenv("ABCGO_EMBED_PIVOT_CHILD"); kind != "" {
		g := Gia_ManStart(1)
		expected := int32(1)
		if kind == "wire" {
			Gia_ManAppendCo(g, Gia_ManAppendCi(g))
			expected = 2
		}
		p := Emb_ManStartSimple(g)
		Emb_ManSetValue(p)
		Gia_ManRandom(1)
		Emb_ManComputeDimensions(p, 4)
		if p.nReached != expected {
			t.Fatalf("%s graph: reached %d vertices, want %d", kind, p.nReached, expected)
		}
		Emb_ManComputeCovariance(p, 4)
		Emb_ManComputeEigenvectors(p, 4, 2)
		Emb_ManComputeSolutions(p, 4, 2)
		for _, value := range p.pSols {
			if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
				t.Fatalf("nonfinite %s graph embedding: %v", kind, p.pSols)
			}
		}
		Emb_ManDerivePlacement(p, 2)
		if len(p.pPlacement) != int(2*p.nObjs) {
			t.Fatal("missing degenerate placement")
		}
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"wire", "empty"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := testoracle.CommandContext(t, ctx, exe, "-test.run=^TestEmbeddingDegeneratePivotTerminates$")
			cmd.Env = append(os.Environ(), "ABCGO_EMBED_PIVOT_CHILD="+kind)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%s pivot pipeline: %v (context %v): %s", kind, err, ctx.Err(), out)
			}
		})
	}
}
