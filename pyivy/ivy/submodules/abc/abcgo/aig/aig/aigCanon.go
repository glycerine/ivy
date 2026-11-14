// Semi-canonical AIG library recording, ported from ABC aigCanon.c.
package aig

import (
	"fmt"
	"slices"
	"sync"

	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
)

const RMAN_MAXVARS = 12
const RMAX_MAXWORD = 1 << (RMAN_MAXVARS - 5)

type Aig_VSig_t struct {
	nOnes    int32
	nCofOnes [RMAN_MAXVARS]int32
}
type Aig_Tru_t struct {
	pNext          *Aig_Tru_t
	Id             int32
	nVisits, nVars uint32
	pTruth         []uint32
}
type Aig_RMan_t struct {
	nVars                                              int32
	pAig                                               *Graph
	nBins, nEntries                                    int32
	pBins                                              []*Aig_Tru_t
	pBidec                                             *bdc.Bdc_Man_t
	pTruthInit, pTruth, pTruthTemp                     [RMAX_MAXWORD]uint32
	pMints                                             [2 * RMAN_MAXVARS]Aig_VSig_t
	pPerm                                              [RMAN_MAXVARS]byte
	nVarFuncs                                          [RMAN_MAXVARS + 1]int32
	nTotal, nTtDsd, nTtDsdPart, nTtDsdNot, nUniqueVars int32
}

var s_pRMan *Aig_RMan_t
var canonRecording sync.Mutex

func Aig_RManSortNums(a []int32, n int32) {
	for i := int32(0); i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if a[j] > a[best] {
				best = j
			}
		}
		a[i], a[best] = a[best], a[i]
	}
}
func Aig_RManCompareSigs(a, b *Aig_VSig_t, n int32) int32 {
	// Match little-endian C memcmp of nOnes only, not numeric counts or the
	// cofactor-count fields. This ordering determines canonical identities.
	x, y := uint32(a.nOnes), uint32(b.nOnes)
	for i := uint(0); i < 32; i += 8 {
		aa, bb := byte(x>>i), byte(y>>i)
		if aa < bb {
			return -1
		}
		if aa > bb {
			return 1
		}
	}
	return 0
}
func Aig_RManVarsAreUnique(s []Aig_VSig_t, n int32) int32 {
	for i := int32(0); i < n-1; i++ {
		if Aig_RManCompareSigs(&s[2*i], &s[2*(i+1)], n) == 0 {
			return 0
		}
	}
	return 1
}
func Aig_RManPrintSigs(s []Aig_VSig_t, n int32) {
	for v := int32(0); v < n; v++ {
		fmt.Printf("%2d : ", v)
		for k := int32(0); k < 2; k++ {
			fmt.Printf("%5d  (", s[2*v+k].nOnes)
			for i := int32(0); i < n; i++ {
				fmt.Printf("%4d ", s[2*v+k].nCofOnes[i])
			}
			fmt.Printf(")  ")
		}
		fmt.Println()
	}
}
func Aig_RManPrintUniqueVars(s []Aig_VSig_t, n int32) {
	for i := int32(0); i < n; i++ {
		if Aig_RManCompareSigs(&s[2*i], &s[2*i+1], n) == 0 {
			fmt.Print("=")
		} else {
			fmt.Print("x")
		}
	}
	fmt.Println()
	fmt.Print("0")
	for i := int32(1); i < n; i++ {
		if Aig_RManCompareSigs(&s[2*(i-1)], &s[2*i], n) == 0 {
			fmt.Print("-")
		} else if i < 10 {
			fmt.Printf("%c", '0'+i)
		} else {
			fmt.Printf("%c", 'A'+i-10)
		}
	}
	fmt.Println()
}
func Aig_RManTableHash(t []uint32, n, bins int32, primes []int32) int32 {
	if bins <= 0 {
		panic("invalid canonical hash size")
	}
	var hash uint32
	for i := int32(0); i < canonWords(n); i++ {
		hash ^= t[i] * uint32(primes[i&15])
	}
	return int32(hash % uint32(bins))
}
func Aig_RManComputeVSigs(t []uint32, n int32, s []Aig_VSig_t, aux []uint32) {
	canonWords(n) // Keep the recorder's twelve-variable bound.
	for v := int32(0); v < n; v++ {
		for phase := int32(0); phase < 2; phase++ {
			if phase == 0 {
				kit.Kit_TruthCofactor0New(aux, t, n, v)
			} else {
				kit.Kit_TruthCofactor1New(aux, t, n, v)
			}
			sig := &s[2*v+phase]
			*sig = Aig_VSig_t{}
			sig.nOnes = kit.Kit_TruthCountOnes(aux, n)
			kit.Kit_TruthCountOnesInCofs0(aux, n, sig.nCofOnes[:])
			Aig_RManSortNums(sig.nCofOnes[:], n)
		}
	}
}
func Aig_RManPrintVarProfile(t []uint32, n int32, aux []uint32) {
	for v := int32(0); v < n; v++ {
		var counts [2]int32
		for bit := int32(0); bit < 32*canonWords(n); bit++ {
			counts[(bit>>uint(v))&1] += int32(canonGet(t, bit))
		}
		fmt.Printf("%2d/%2d ", counts[0], counts[1])
	}
	fmt.Println()
}
func Aig_RManSemiCanonicize(out, in []uint32, n int32, perm []byte, s []Aig_VSig_t, returnIn int32) uint32 {
	Aig_RManComputeVSigs(in, n, s, out)
	var phase uint32
	for i := int32(0); i < n; i++ {
		if Aig_RManCompareSigs(&s[2*i], &s[2*i+1], n) > 0 {
			phase |= 1 << uint(i)
			s[2*i], s[2*i+1] = s[2*i+1], s[2*i]
			kit.Kit_TruthChangePhase(in, n, i)
		}
	}
	originalIn, originalOut := in, out
	for {
		changed := false
		for i := int32(0); i < n-1; i++ {
			if Aig_RManCompareSigs(&s[2*i], &s[2*(i+1)], n) <= 0 {
				continue
			}
			changed = true
			perm[i], perm[i+1] = perm[i+1], perm[i]
			s[2*i], s[2*(i+1)] = s[2*(i+1)], s[2*i]
			s[2*i+1], s[2*(i+1)+1] = s[2*(i+1)+1], s[2*i+1]
			kit.Kit_TruthSwapAdjacentVars(out, in, n, i)
			in, out = out, in
		}
		if !changed {
			break
		}
	}
	if returnIn != 0 {
		copy(originalIn[:canonWords(n)], in)
	} else {
		copy(originalOut[:canonWords(n)], in)
	}
	return phase
}
func canonWords(n int32) int32 {
	if n < 0 || n > RMAN_MAXVARS {
		panic("canonical variable count out of range")
	}
	if n <= 5 {
		return 1
	}
	return 1 << uint(n-5)
}
func canonGet(t []uint32, bit int32) uint32 { return t[bit>>5] >> uint(bit&31) & 1 }
func canonSet(t []uint32, bit int32)        { t[bit>>5] |= 1 << uint(bit&31) }
func Aig_RManStart() *Aig_RMan_t {
	p := &Aig_RMan_t{nVars: RMAN_MAXVARS, pAig: NewGraph(), nBins: util.Abc_PrimeCudd(5000)}
	for i := int32(0); i < p.nVars; i++ {
		p.pAig.NewInput()
	}
	p.pBins = make([]*Aig_Tru_t, p.nBins)
	p.pBidec = bdc.Bdc_ManAlloc(&bdc.Bdc_Par_t{NVarsMax: RMAN_MAXVARS})
	return p
}

var canonHashPrimes = []int32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}

func Aig_RManTableLookup(p *Aig_RMan_t, truth []uint32, n int32) **Aig_Tru_t {
	slot := &p.pBins[Aig_RManTableHash(truth, n, p.nBins, canonHashPrimes)]
	for *slot != nil {
		entry := *slot
		// The width tag prevents cross-width truth comparisons and reads past
		// the allocation in the original C lookup.
		if entry.nVars == uint32(n) && slices.Equal(entry.pTruth, truth[:canonWords(n)]) {
			return slot
		}
		slot = &entry.pNext
	}
	return slot
}
func Aig_RManTableResize(p *Aig_RMan_t) {
	old := p.pBins
	p.nBins = util.Abc_PrimeCudd(uint32(3 * p.nBins))
	p.pBins = make([]*Aig_Tru_t, p.nBins)
	var count int32
	for _, head := range old {
		for entry := head; entry != nil; {
			next := entry.pNext
			entry.pNext = nil
			slot := Aig_RManTableLookup(p, entry.pTruth, int32(entry.nVars))
			if *slot != nil {
				panic("duplicate canonical truth entry")
			}
			*slot = entry
			entry = next
			count++
		}
	}
	if count != p.nEntries {
		panic("canonical table entry count mismatch")
	}
}
func Aig_RManTableFindOrAdd(p *Aig_RMan_t, truth []uint32, n int32) int32 {
	slot := Aig_RManTableLookup(p, truth, n)
	if *slot != nil {
		(*slot).nVisits = ((*slot).nVisits + 1) & ((1 << 27) - 1)
		return 0
	}
	if p.nEntries == 3*p.nBins {
		Aig_RManTableResize(p)
		// Reacquire the bucket: slot previously referred to the old table.
		slot = Aig_RManTableLookup(p, truth, n)
	}
	*slot = &Aig_Tru_t{Id: p.nEntries, nVars: uint32(n), nVisits: 1, pTruth: append([]uint32(nil), truth[:canonWords(n)]...)}
	p.nEntries++
	return 1
}
func Aig_RManStop(p *Aig_RMan_t) {
	if p == nil {
		return
	}
	fmt.Printf("Total funcs    = %10d\nFull DSD funcs = %10d\nPart DSD funcs = %10d\nNon- DSD funcs = %10d\nUniq-var funcs = %10d\nUnique   funcs = %10d\nDistribution of functions:\n", p.nTotal, p.nTtDsd, p.nTtDsdPart, p.nTtDsdNot, p.nUniqueVars, p.nEntries)
	for i := int32(5); i <= p.nVars; i++ {
		fmt.Printf("%2d = %8d\n", i, p.nVarFuncs[i])
	}
	bdc.Bdc_ManFree(p.pBidec)
	p.pBins = nil
	p.pBidec = nil
	p.pAig = nil
}
func Aig_RManQuit() error {
	canonRecording.Lock()
	defer canonRecording.Unlock()
	if s_pRMan == nil {
		return nil
	}
	p := s_pRMan
	if err := ioa.Ioa_WriteAiger(p.pAig, fmt.Sprintf("aiglib%02d.aig", p.nVars), 0, 1); err != nil {
		return err
	}
	Aig_RManStop(p)
	s_pRMan = nil
	return nil
}
func Bdc_FunCopyHop(o *bdc.Bdc_Fun_t) Edge {
	edge, ok := bdc.Bdc_FuncCopy(bdc.Bdc_Regular(o)).(Edge)
	if !ok {
		panic("missing canonical decomposition copy")
	}
	if bdc.Bdc_IsComplement(o) != 0 {
		edge = edge.Not()
	}
	return edge
}
func Aig_RManSaveOne(p *Aig_RMan_t, truth []uint32, n int32) {
	if bdc.Bdc_ManDecompose(p.pBidec, truth, nil, n, nil, 1000) < 0 {
		fmt.Println("Decomposition failed.")
		return
	}
	bdc.Bdc_FuncSetCopy(bdc.Bdc_ManFunc(p.pBidec, 0), p.pAig.Const1())
	for i := int32(0); i < n; i++ {
		bdc.Bdc_FuncSetCopy(bdc.Bdc_ManFunc(p.pBidec, i+1), p.pAig.inputs[i])
	}
	for i := n + 1; i < bdc.Bdc_ManNodeNum(p.pBidec); i++ {
		node := bdc.Bdc_ManFunc(p.pBidec, i)
		bdc.Bdc_FuncSetCopy(node, p.pAig.AND(Bdc_FunCopyHop(bdc.Bdc_FuncFanin0(node)), Bdc_FunCopyHop(bdc.Bdc_FuncFanin1(node))))
	}
	p.pAig.AddOutput(Bdc_FunCopyHop(bdc.Bdc_ManRoot(p.pBidec)))
}
func Aig_RManRecord(truth []uint32, nInit int32) {
	if nInit < 0 || nInit > RMAN_MAXVARS {
		fmt.Println("The number of variables is too large.")
		return
	}
	canonRecording.Lock()
	defer canonRecording.Unlock()
	if s_pRMan == nil {
		s_pRMan = Aig_RManStart()
	}
	p := s_pRMan
	p.nTotal++
	network := kit.Kit_DsdDecompose(truth, nInit)
	defer kit.Kit_DsdNtkFree(network)
	prime := kit.Kit_DsdNonDsdPrimeMax(network)
	if prime == nil || kit.Kit_DsdObjFanNum(prime) == 3 {
		p.nTtDsd++
		return
	}
	n := kit.Kit_DsdObjFanNum(prime)
	p.nVarFuncs[n]++
	if n < nInit {
		p.nTtDsdPart++
	} else {
		p.nTtDsdNot++
	}
	words := canonWords(n)
	copy(p.pTruthInit[:words], kit.Kit_DsdObjTruth(prime))
	if p.pTruthInit[0]&1 != 0 {
		for i := int32(0); i < words; i++ {
			p.pTruthInit[i] = ^p.pTruthInit[i]
		}
	}
	copy(p.pTruth[:words], p.pTruthInit[:words])
	for i := int32(0); i < n; i++ {
		p.pPerm[i] = byte(i)
	}
	phase := Aig_RManSemiCanonicize(p.pTruthTemp[:], p.pTruth[:], n, p.pPerm[:], p.pMints[:], 1)
	unique := Aig_RManVarsAreUnique(p.pMints[:], n)
	p.nUniqueVars += unique
	if Aig_RManTableFindOrAdd(p, p.pTruth[:], n) != 0 {
		Aig_RManSaveOne(p, p.pTruth[:], n)
	}
	if unique != 0 {
		for original := int32(0); original < 1<<uint(n); original++ {
			shifted := original ^ int32(phase)
			var canonical int32
			for i := int32(0); i < n; i++ {
				canonical |= ((shifted >> p.pPerm[i]) & 1) << uint(i)
			}
			if canonGet(p.pTruth[:], canonical) != canonGet(p.pTruthInit[:], original) {
				fmt.Println("Verification failed.")
				break
			}
		}
	}
}
