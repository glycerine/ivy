package aig

import (
	"github.com/glycerine/abcgo/aig/ioa"
	"math/rand"
	"os"
	"testing"
)

func TestCanonSignatureByteOrder(t *testing.T) {
	a, b := Aig_VSig_t{nOnes: 256}, Aig_VSig_t{nOnes: 1}
	if Aig_RManCompareSigs(&a, &b, 1) >= 0 {
		t.Fatal("must preserve little-endian memcmp ordering")
	}
	x := []int32{3, 1, 9, 2}
	Aig_RManSortNums(x, 4)
	for i, v := range []int32{9, 3, 2, 1} {
		if x[i] != v {
			t.Fatal(x)
		}
	}
}
func TestCanonPhasePermutationRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(573))
	for n := int32(1); n <= 9; n++ {
		for trial := 0; trial < 30; trial++ {
			original := make([]uint32, canonWords(n))
			for bit := int32(0); bit < 1<<uint(n); bit++ {
				if r.Intn(2) != 0 {
					canonSet(original, bit)
				}
			}
			if n < 5 {
				for bit := int32(1) << uint(n); bit < 32; bit++ {
					if canonGet(original, bit&((1<<uint(n))-1)) != 0 {
						canonSet(original, bit)
					}
				}
			}
			for _, returnIn := range []int32{0, 1} {
				in, out := append([]uint32(nil), original...), make([]uint32, len(original))
				perm := make([]byte, n)
				for i := range perm {
					perm[i] = byte(i)
				}
				sigs := make([]Aig_VSig_t, 2*n)
				phase := Aig_RManSemiCanonicize(out, in, n, perm, sigs, returnIn)
				result := out
				if returnIn != 0 {
					result = in
				}
				for bit := int32(0); bit < 1<<uint(n); bit++ {
					var src int32
					for v, p := range perm {
						src |= ((bit >> uint(v)) & 1) << p
					}
					src ^= int32(phase)
					if canonGet(result, bit) != canonGet(original, src) {
						t.Fatal(n, trial, returnIn, bit, phase, perm)
					}
				}
				for i := int32(0); i < n; i++ {
					if Aig_RManCompareSigs(&sigs[2*i], &sigs[2*i+1], n) > 0 {
						t.Fatal("phase ordering")
					}
					if i > 0 && Aig_RManCompareSigs(&sigs[2*(i-1)], &sigs[2*i], n) > 0 {
						t.Fatal("permutation ordering")
					}
				}
			}
		}
	}
}
func TestCanonSignaturesAndHash(t *testing.T) {
	table := []uint32{0x88888888}
	s := make([]Aig_VSig_t, 4)
	Aig_RManComputeVSigs(table, 2, s, make([]uint32, 1))
	if s[0].nOnes != 0 || s[1].nOnes != 16 || s[2].nOnes != 0 || s[3].nOnes != 16 {
		t.Fatal(s)
	}
	if Aig_RManVarsAreUnique(s, 2) != 0 {
		t.Fatal("AND inputs are symmetric")
	}
	primes := make([]int32, 16)
	primes[0] = 1291
	if got := Aig_RManTableHash(table, 2, 5003, primes); got != int32((uint32(0x88888888)*uint32(primes[0]))%5003) {
		t.Fatal(got)
	}
}

func TestCanonHashTableResizeAndOwnership(t *testing.T) {
	p := &Aig_RMan_t{nBins: 2, pBins: make([]*Aig_Tru_t, 2)}
	for i := uint32(0); i < 40; i++ {
		truth := []uint32{i * 2654435761}
		if Aig_RManTableFindOrAdd(p, truth, 5) != 1 {
			t.Fatal("new entry")
		}
		truth[0] = 0
	}
	if p.nEntries != 40 || p.nBins <= 2 {
		t.Fatal("resize did not run")
	}
	for i := uint32(0); i < 40; i++ {
		truth := []uint32{i * 2654435761}
		slot := Aig_RManTableLookup(p, truth, 5)
		if *slot == nil || (*slot).Id != int32(i) || (*slot).pTruth[0] != truth[0] {
			t.Fatal("entry lost at resize", i)
		}
		if Aig_RManTableFindOrAdd(p, truth, 5) != 0 || (*slot).nVisits != 2 {
			t.Fatal("visit count")
		}
	}
	if Aig_RManTableFindOrAdd(p, []uint32{0}, 4) != 1 {
		t.Fatal("different variable counts collided")
	}
}

func TestCanonBidecompositionGraph(t *testing.T) {
	p := Aig_RManStart()
	defer Aig_RManStop(p)
	for _, truth := range []uint32{0, ^uint32(0), 0x88888888, 0x96969696, 0xe8e8e8e8} {
		before := p.pAig.OutputCount()
		Aig_RManSaveOne(p, []uint32{truth}, 3)
		if p.pAig.OutputCount() != before+1 {
			t.Fatal("missing recorded output")
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := make([]bool, RMAN_MAXVARS)
			for v := 0; v < 3; v++ {
				inputs[v] = assignment&(1<<v) != 0
			}
			outputs, err := p.pAig.Evaluate(inputs)
			if err != nil || outputs[before] != (truth>>uint(assignment)&1 != 0) {
				t.Fatal(truth, assignment, outputs, err)
			}
		}
	}
	if err := p.pAig.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestCanonRecordsPrimeAndDisjointFunctions(t *testing.T) {
	previous := s_pRMan
	s_pRMan = nil
	defer func() { Aig_RManStop(s_pRMan); s_pRMan = previous }()
	var truth uint32
	for a := uint32(0); a < 32; a++ {
		count := 0
		for v := uint(0); v < 4; v++ {
			if a>>v&1 != 0 {
				count++
			}
		}
		if count >= 2 {
			truth |= 1 << a
		}
	}
	Aig_RManRecord([]uint32{truth}, 4)
	if s_pRMan == nil || s_pRMan.nEntries != 1 || s_pRMan.nTtDsdNot != 1 || s_pRMan.pAig.OutputCount() != 1 {
		t.Fatal("prime was not recorded", s_pRMan)
	}
	Aig_RManRecord([]uint32{truth}, 5)
	if s_pRMan.nEntries != 1 || s_pRMan.nTtDsdPart != 1 {
		t.Fatal("partial prime deduplication")
	}
	Aig_RManRecord([]uint32{0x80008000}, 4)
	if s_pRMan.nTtDsd != 1 {
		t.Fatal("AND should be fully DSD")
	}
	if s_pRMan.nTotal != 3 {
		t.Fatal("record count")
	}
}

func TestCanonQuitWritesLibraryAndKeepsStateOnError(t *testing.T) {
	t.Chdir(t.TempDir())
	previous := s_pRMan
	s_pRMan = Aig_RManStart()
	defer func() {
		if s_pRMan != nil {
			Aig_RManStop(s_pRMan)
		}
		s_pRMan = previous
	}()
	Aig_RManSaveOne(s_pRMan, []uint32{0x88888888}, 3)
	if err := os.Mkdir("aiglib12.aig", 0755); err != nil {
		t.Fatal(err)
	}
	if err := Aig_RManQuit(); err == nil || s_pRMan == nil {
		t.Fatal("failed write discarded recording state")
	}
	if err := os.Remove("aiglib12.aig"); err != nil {
		t.Fatal(err)
	}
	if err := Aig_RManQuit(); err != nil {
		t.Fatal(err)
	}
	if s_pRMan != nil {
		t.Fatal("successful quit retained manager")
	}
	model, err := ioa.Ioa_ReadAiger("aiglib12.aig", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Inputs) != RMAN_MAXVARS || len(model.Outputs) != 1 {
		t.Fatal("incorrect recorded library shape")
	}
	if err := model.Check(); err != nil {
		t.Fatal(err)
	}
}
