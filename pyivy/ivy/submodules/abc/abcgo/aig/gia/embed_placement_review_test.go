package gia

import (
	"fmt"
	"math"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestEmbeddingPlacementPreservesZeroCoordinateOrder(t *testing.T) {
	p := &Emb_Man_t{nObjs: 3, pSols: []float32{0, -1, 1, 0, 1, -1}}
	Emb_ManDerivePlacement(p, 2)
	want := []uint16{10922, 10922, 0, 21844, 21844, 0}
	if !slices.Equal(p.pPlacement, want) {
		t.Fatalf("placement %v, want zero ranked between negative and positive: %v", p.pPlacement, want)
	}
}

func TestEmbeddingPlacementMatchesOriginalNonzeroAxes(t *testing.T) {
	exe := compileGiaSourceOracle(t, "embedding_placement.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C placement: %v: %s", err, want)
	}
	var records strings.Builder
	for cluster := 0; cluster < 2; cluster++ {
		for regs := int32(0); regs < 2; regs++ {
			g := Gia_ManStart(1)
			var inputs [4]int32
			for i := range inputs {
				inputs[i] = Gia_ManAppendCi(g)
			}
			root := inputs[0]
			for i := 0; i < 12; i++ {
				root = Gia_ManAppendAnd(g, root, inputs[(i+1)%4]^int32(i&1))
			}
			Gia_ManAppendCo(g, root)
			Gia_ManAppendCo(g, inputs[0])
			Gia_ManAppendCo(g, inputs[3])
			Gia_ManSetRegNum(g, regs)
			start := Emb_ManStartSimple
			if cluster != 0 {
				start = Emb_ManStart
			}
			p := start(g)
			Emb_ManSetValue(p)
			p.pSols = make([]float32, 2*p.nObjs)
			for i := int32(0); i < p.nObjs; i++ {
				p.pSols[i] = float32((i*7)%17 + 1)
				p.pSols[p.nObjs+i] = float32((i*11)%19 + 1)
			}
			record := func() {
				fmt.Fprintf(&records, " %.0f", Emb_ManComputeHPWL(p))
				for _, value := range p.pPlacement {
					fmt.Fprintf(&records, " %d", value)
				}
			}
			Emb_ManDerivePlacement(p, 2)
			fmt.Fprintf(&records, "%d %d %d", cluster, regs, p.nObjs)
			record()
			for iter := 0; iter < 3; iter++ {
				Emb_ManPlacementRefine(p, 1, 0)
				records.WriteString(" refine")
				record()
			}
			records.WriteByte('\n')
		}
	}
	if records.String() != string(want) {
		t.Fatalf("Go placement:\n%sC placement:\n%s", records.String(), want)
	}
}

func TestEmbeddingPlacementFlatNonzeroAxesStayFinite(t *testing.T) {
	p := &Emb_Man_t{nObjs: 4, pSols: []float32{5, 5, 5, 5, -3, -3, -3, -3}}
	Emb_ManDerivePlacement(p, 2)
	for _, value := range p.pSols {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			t.Fatalf("flat finite axis became nonfinite: %v", p.pSols)
		}
	}
	for axis := 0; axis < 2; axis++ {
		seen := make(map[uint16]bool)
		for id := 0; id < 4; id++ {
			value := p.pPlacement[2*id+axis]
			if value > GIA_PLACE_SIZE || seen[value] {
				t.Fatalf("invalid flat-axis placement: %v", p.pPlacement)
			}
			seen[value] = true
		}
	}
}
