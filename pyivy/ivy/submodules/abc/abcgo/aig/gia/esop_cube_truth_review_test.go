package gia

import (
	"math/rand"
	"testing"
)

func esopReviewLiteralCubeTruth(lits []int32, nvars int) uint64 {
	var truth uint64
	for bits := 0; bits < 1<<nvars; bits++ {
		product := true
		for _, lit := range lits {
			product = product && ((bits&(1<<(lit>>1)) != 0) != (lit&1 != 0))
		}
		if product {
			truth |= 1 << bits
		}
	}
	return truth
}
func esopReviewIDTruth(p *Eso_Man_t, id int32, nvars int) uint64 {
	if id == p.Cube1 {
		return esopReviewLiteralCubeTruth(nil, nvars)
	}
	return esopReviewLiteralCubeTruth(Vec_IntArray(Eso_ManCube(p, id)), nvars)
}
func esopReviewCoverTruth(p *Eso_Man_t, cover *Vec_Int_t, nvars int) uint64 {
	var truth uint64
	for _, id := range Vec_IntArray(cover) {
		truth ^= esopReviewIDTruth(p, id, nvars)
	}
	return truth
}

func TestEsopCubeOperationsIndependentTruth(t *testing.T) {
	pairs := 0
	for nvars := 0; nvars <= 5; nvars++ {
		g := Gia_ManStart(16)
		for i := 0; i < nvars; i++ {
			Gia_ManAppendCi(g)
		}
		p := Eso_ManAlloc(g)
		count := 1
		for i := 0; i < nvars; i++ {
			count *= 3
		}
		ids := make([]int32, count)
		truths := make([]uint64, count)
		for code := range ids {
			lits := esopReviewCube(code, nvars)
			ids[code] = p.Cube1
			if code != 0 {
				ids[code] = p.intern(lits)
			}
			truths[code] = esopReviewLiteralCubeTruth(lits, nvars)
		}
		product, cover, tmp := Vec_IntAlloc(nvars), Vec_IntAlloc(16), Vec_IntAlloc(16)
		for a := range ids {
			for b := range ids {
				if a != 0 && b != 0 {
					id := Eso_ManComputeAnd(p, Eso_ManCube(p, ids[a]), Eso_ManCube(p, ids[b]), product)
					want := truths[a] & truths[b]
					if id < 0 {
						if want != 0 {
							t.Fatalf("nvars%d cubes%d/%d falsely contradictory", nvars, a, b)
						}
					} else if got := esopReviewIDTruth(p, id, nvars); got != want {
						t.Fatalf("nvars%d product%d/%d got%x want%x", nvars, a, b, got, want)
					}
				}
				Eso_ManMinimizeAdd(p, ids[a])
				Eso_ManMinimizeAdd(p, ids[b])
				Eso_ManMinimizeCopy(p, cover)
				want := truths[a] ^ truths[b]
				if got := esopReviewCoverTruth(p, cover, nvars); got != want {
					t.Fatalf("nvars%d XOR%d/%d got%x want%x", nvars, a, b, got, want)
				}
				for phase := int32(0); phase < 2; phase++ {
					Eso_ManTransformOne(p, cover, phase, tmp)
					expected := want
					if phase != 0 {
						expected ^= esopReviewLiteralCubeTruth(nil, nvars)
					}
					if got := esopReviewCoverTruth(p, tmp, nvars); got != expected {
						t.Fatalf("nvars%d phase%d cover%d/%d got%x want%x", nvars, phase, a, b, got, expected)
					}
				}
				pairs++
			}
		}
		rng := rand.New(rand.NewSource(int64(100 + nvars)))
		for sequence := 0; sequence < 128; sequence++ {
			Vec_IntClear(cover)
			var want uint64
			for step := 0; step < 32; step++ {
				for _, id := range Vec_IntArray(cover) {
					Eso_ManMinimizeAdd(p, id)
				}
				code := rng.Intn(count)
				Eso_ManMinimizeAdd(p, ids[code])
				want ^= truths[code]
				Eso_ManMinimizeCopy(p, cover)
				if got := esopReviewCoverTruth(p, cover, nvars); got != want {
					t.Fatalf("nvars%d sequence%d step%d got%x want%x", nvars, sequence, step, got, want)
				}
			}
		}
		Eso_ManStop(p)
		Gia_ManStop(g)
	}
	t.Logf("%d ordered cube pairs and 24576 sequence prefixes checked", pairs)
}

func TestEsopTypedScratchAndResultOwnership(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		g := esopReviewTypedGraph(kind, 5)
		saved := make([]uint32, g.nObjs)
		for i, obj := range g.pObjs {
			obj.Value = uint32(79 + i*19)
			saved[i] = obj.Value
		}
		var cover *Vec_Wec_t
		Eso_ManCompute(g, 0, &cover)
		for i, obj := range g.pObjs {
			if obj.Value != saved[i] {
				t.Fatalf("%s extraction modified Value at%d", kind, i)
			}
		}
		esopReviewCheckCover(t, g, cover)
		out := Eso_ManCompute(g, 0, nil)
		for i := int32(0); i < Gia_ManCiNum(g); i++ {
			if Gia_ManCi(g, i).Value != uint32(Gia_ManCi(out, i).id<<1) {
				t.Fatal("missing CI copy literal")
			}
		}
		cofReviewCheckRebuild(t, g, out)
		before := cofReviewValues(out, 3)
		Gia_ManStop(g)
		after := cofReviewValues(out, 3)
		for i := range before {
			if before[i] != after[i] {
				t.Fatal("result depends on disposed source")
			}
		}
		// The extracted cover and interned cube storage must also survive manager stop.
		if Vec_WecSize(cover) == 0 {
			t.Fatal("lost extracted cover")
		}
		Gia_ManStop(out)
	}
}
