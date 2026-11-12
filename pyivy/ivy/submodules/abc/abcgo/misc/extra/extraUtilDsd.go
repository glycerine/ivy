package extra

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/bool/rsb"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
	"os"
	"slices"
	"sync"
)

/**CFile****************************************************************

  FileName    [extraUtilDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [File management utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilDsd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

const DSD_CLASS_NUM = 595

type Sdm_Dsd_t struct {
	nVars, nAnds, nClauses int32
	uTruth                 uint64
	pStr                   string
}
type sdmSignature struct {
	weight, support byte
	cofactors       [6]uint16
}
type sdmGeometry struct {
	classCounts      [DSD_CLASS_NUM][6][2]byte
	classOnes        [DSD_CLASS_NUM]byte
	permutations     [720][6]byte
	permutationIndex map[uint32]int32
	permuted         []uint64
	buckets          map[sdmSignature][]int32
}
type Sdm_Man_t_ struct {
	mu               sync.Mutex
	geometry         *sdmGeometry
	pHash            map[uint64]int32
	vConfgRes        []int32
	queryCache       map[uint64]int32
	nCountDsd        [DSD_CLASS_NUM]uint64
	nNonDsd, nAllDsd uint64
}
type Sdm_Man_t = Sdm_Man_t_

var sdmGeometryOnce sync.Once
var sdmSharedGeometry *sdmGeometry
var sdmTruthMasks = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
var sdmSwapMasks = [5][3]uint64{{0x9999999999999999, 0x2222222222222222, 0x4444444444444444}, {0xc3c3c3c3c3c3c3c3, 0x0c0c0c0c0c0c0c0c, 0x3030303030303030}, {0xf00ff00ff00ff00f, 0x00f000f000f000f0, 0x0f000f000f000f00}, {0xff0000ffff0000ff, 0x0000ff000000ff00, 0x00ff000000ff0000}, {0xffff00000000ffff, 0x00000000ffff0000, 0x0000ffff00000000}}

func sdmSwapAdjacent(truth uint64, v int32) uint64 {
	m := sdmSwapMasks[v]
	shift := uint(1) << uint(v)
	return truth&m[0] | (truth&m[1])<<shift | (truth&m[2])>>shift
}
func sdmPermSchedule(n int) []int32 {
	if n == 1 {
		return []int32{0}
	}
	if n == 2 {
		return []int32{0, 0}
	}
	previous := sdmPermSchedule(n - 1)
	fact := len(previous) * n
	groups := fact / n / 2
	result := make([]int32, 0, fact)
	for k := 0; k < groups; k++ {
		for i := n - 1; i > 0; i-- {
			result = append(result, int32(i-1))
		}
		result = append(result, previous[2*k]+1)
		for i := 0; i < n-1; i++ {
			result = append(result, int32(i))
		}
		result = append(result, previous[2*k+1])
	}
	return result
}
func sdmTruthSignature(truth uint64) sdmSignature {
	ones := bits.OnesCount64(truth)
	result := sdmSignature{weight: byte(min(ones, 64-ones))}
	var direct, complement [6]uint16
	for v := int32(0); v < 6; v++ {
		if util.Abc_Tt6HasVar(truth, v) != 0 {
			result.support++
		}
		a, b := bits.OnesCount64(util.Abc_Tt6Cofactor0(truth, v)), bits.OnesCount64(util.Abc_Tt6Cofactor1(truth, v))
		direct[v] = uint16(min(a, b)<<7 | max(a, b))
		complement[v] = uint16(min(64-a, 64-b)<<7 | max(64-a, 64-b))
	}
	slices.Sort(direct[:])
	slices.Sort(complement[:])
	result.cofactors = direct
	if slices.Compare(complement[:], direct[:]) < 0 {
		result.cofactors = complement
	}
	return result
}
func sdmGetGeometry() *sdmGeometry { sdmGeometryOnce.Do(sdmInitGeometry); return sdmSharedGeometry }
func sdmInitGeometry() {
	g := &sdmGeometry{permutationIndex: map[uint32]int32{}, permuted: make([]uint64, DSD_CLASS_NUM*720), buckets: map[sdmSignature][]int32{}}
	schedule := sdmPermSchedule(6)
	permutation := [6]byte{0, 1, 2, 3, 4, 5}
	for k, swap := range schedule {
		permutation[swap], permutation[swap+1] = permutation[swap+1], permutation[swap]
		g.permutations[k] = permutation
		var packed uint32
		for v, label := range permutation {
			packed |= uint32(label) << uint(3*v)
		}
		g.permutationIndex[packed] = int32(k)
	}
	for class, entry := range s_DsdClass6 {
		g.classCounts[class] = sdmCofactorCounts(entry.uTruth)
		g.classOnes[class] = byte(bits.OnesCount64(entry.uTruth))
		value := entry.uTruth
		for k, swap := range schedule {
			value = sdmSwapAdjacent(value, swap)
			g.permuted[class*720+k] = value
		}
		signature := sdmTruthSignature(entry.uTruth)
		g.buckets[signature] = append(g.buckets[signature], int32(class))
	}
	sdmSharedGeometry = g
}
func sdmCofactorCounts(truth uint64) (counts [6][2]byte) {
	for v, mask := range sdmTruthMasks {
		counts[v][0] = byte(2 * bits.OnesCount64(truth&^mask))
		counts[v][1] = byte(2 * bits.OnesCount64(truth&mask))
	}
	return
}

// Cofactor populations are exact necessary conditions, not a heuristic match.
// They fix asymmetric phase bits before the full truth-word comparison.
func sdmPhaseCandidates(g *sdmGeometry, class int32, perm int, counts [6][2]byte, ones int) uint64 {
	var candidates uint64
	for output := 0; output < 2; output++ {
		population := int(g.classOnes[class])
		if output != 0 {
			population = 64 - population
		}
		if population != ones {
			continue
		}
		fixed, free := uint32(0), uint32(0)
		valid := true
		for v, old := range g.permutations[perm] {
			a, b := g.classCounts[class][old][0], g.classCounts[class][old][1]
			if output != 0 {
				a, b = 64-a, 64-b
			}
			zero := a == counts[v][0] && b == counts[v][1]
			one := b == counts[v][0] && a == counts[v][1]
			if !zero && !one {
				valid = false
				break
			}
			if zero && one {
				free |= 1 << uint(v)
			} else if one {
				fixed |= 1 << uint(v)
			}
		}
		if !valid {
			continue
		}
		for subset := free; ; subset = (subset - 1) & free {
			candidates |= uint64(1) << uint(fixed|subset)
			if subset == 0 {
				break
			}
		}
	}
	return candidates
}
func sdmFlipMask(value uint64, mask uint32) uint64 {
	for mask != 0 {
		v := int32(bits.TrailingZeros32(mask))
		wordMask := sdmTruthMasks[v]
		shift := uint(1) << uint(v)
		value = (value&wordMask)>>shift | (value&^wordMask)<<shift
		mask &= mask - 1
	}
	return value
}
func (g *sdmGeometry) permutationForMask(mask uint32) (int32, bool) {
	var labels [6]byte
	var used uint32
	for i := range labels {
		v := byte(mask >> uint(3*i) & 7)
		if v == 6 {
			return 0, false
		}
		labels[i] = v
		if v != 7 {
			if used&(1<<v) != 0 {
				return 0, false
			}
			used |= 1 << v
		}
	}
	next := byte(0)
	var packed uint32
	for i, v := range labels {
		if v == 7 {
			for used&(1<<next) != 0 {
				next++
			}
			v = next
			used |= 1 << v
		}
		packed |= uint32(v) << uint(3*i)
	}
	index, ok := g.permutationIndex[packed]
	return index, ok
}

// Read the original little-endian file format: N truth words, then N configs.
func sdmReadHashTable(path string) (map[uint64]int32, []int32, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(data) == 0 || len(data)%12 != 0 {
		return nil, nil, fmt.Errorf("DSD table must contain complete 12-byte records")
	}
	n := len(data) / 12
	index := make(map[uint64]int32, n)
	configs := make([]int32, n)
	g := sdmGetGeometry()
	for i := 0; i < n; i++ {
		truth := binary.LittleEndian.Uint64(data[8*i:])
		config := binary.LittleEndian.Uint32(data[8*n+4*i:])
		class := config >> 17
		perm := config >> 6 & 1023
		if class >= DSD_CLASS_NUM || perm >= 720 || truth&1 != 0 {
			return nil, nil, fmt.Errorf("invalid DSD table record %d", i)
		}
		value := sdmFlipMask(g.permuted[int(class)*720+int(perm)], config&63)
		if config&(1<<16) != 0 {
			value = ^value
		}
		if value != truth {
			return nil, nil, fmt.Errorf("DSD configuration mismatch at record %d", i)
		}
		if _, exists := index[truth]; exists {
			return nil, nil, fmt.Errorf("duplicate DSD truth at record %d", i)
		}
		index[truth] = int32(i)
		configs[i] = int32(config)
	}
	return index, configs, nil
}

// BuildHashTable can materialize the complete original orbit table when callers
// need it. Managers avoid its large allocation when no data file is present by
// performing exact, signature-filtered orbit searches and caching used queries.
func Sdm_ManBuildHashTable() (map[uint64]int32, []int32, error) {
	index, configs, err := sdmReadHashTable("dsdfuncs6.dat")
	if err == nil {
		return index, configs, nil
	}
	if !os.IsNotExist(err) {
		return nil, nil, err
	}
	index, configs = sdmBuildHashOrbits(sdmGetGeometry(), DSD_CLASS_NUM)
	return index, configs, nil
}
func sdmBuildHashOrbits(g *sdmGeometry, limit int) (map[uint64]int32, []int32) {
	index := make(map[uint64]int32)
	var configs []int32
	for class := 0; class < limit; class++ {
		for perm := 0; perm < 720; perm++ {
			value := g.permuted[class*720+perm]
			for phase := uint32(0); phase < 64; phase++ {
				if phase != 0 {
					value = sdmFlipMask(value, phase^(phase-1))
				}
				truth, output := value, uint32(0)
				if truth&1 != 0 {
					truth = ^truth
					output = 1
				}
				if _, exists := index[truth]; !exists {
					index[truth] = int32(len(configs))
					configs = append(configs, int32(uint32(class)<<17|output<<16|uint32(perm)<<6|phase))
				}
			}
		}
	}
	return index, configs
}
func Sdm_ManPrecomputePerms(p *Sdm_Man_t) {
	if p.geometry != nil {
		panic("DSD manager already initialized")
	}
	p.geometry = sdmGetGeometry()
	p.queryCache = map[uint64]int32{}
	index, configs, err := sdmReadHashTable("dsdfuncs6.dat")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	p.pHash, p.vConfgRes = index, configs
}
func Sdm_ManAlloc() *Sdm_Man_t { p := &Sdm_Man_t{}; Sdm_ManPrecomputePerms(p); return p }
func Sdm_ManFree(p *Sdm_Man_t) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pHash = nil
	p.vConfgRes = nil
	p.queryCache = nil
	p.geometry = nil
}
func sdmCheckDsd6(p *Sdm_Man_t, truth uint64) int32 {
	output := int32(truth & 1)
	if output != 0 {
		truth = ^truth
	}
	if p.geometry == nil {
		panic("closed DSD manager")
	}
	if p.pHash != nil {
		index, exists := p.pHash[truth]
		if !exists {
			return -1
		}
		return p.vConfgRes[index] ^ (output << 16)
	}
	if cached, exists := p.queryCache[truth]; exists {
		if cached < 0 {
			return -1
		}
		return cached ^ (output << 16)
	}
	config := int32(-1)
	counts := sdmCofactorCounts(truth)
	ones := bits.OnesCount64(truth)
search:
	for _, class := range p.geometry.buckets[sdmTruthSignature(truth)] {
		for perm := 0; perm < 720; perm++ {
			candidates := sdmPhaseCandidates(p.geometry, class, perm, counts, ones)
			for candidates != 0 {
				phase := uint32(bits.TrailingZeros64(candidates))
				candidates &= candidates - 1
				value := sdmFlipMask(p.geometry.permuted[int(class)*720+perm], phase)
				normalized, compl := value, uint32(0)
				if value&1 != 0 {
					normalized = ^value
					compl = 1
				}
				if normalized == truth {
					config = int32(uint32(class)<<17 | compl<<16 | uint32(perm)<<6 | phase)
					break search
				}
			}
		}
	}
	if len(p.queryCache) >= 1<<20 {
		clear(p.queryCache)
	}
	p.queryCache[truth] = config
	if config < 0 {
		return -1
	}
	return config ^ (output << 16)
}
func Sdm_ManCheckDsd6(p *Sdm_Man_t, truth uint64) int32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return sdmCheckDsd6(p, truth)
}
func Sdm_ManComputeFunc(p *Sdm_Man_t, lit0, lit1 int32, cut []int32, mask, xor int32) int32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if lit0 < 0 || lit1 < 0 || lit0 >= 2*DSD_CLASS_NUM || lit1 >= 2*DSD_CLASS_NUM || mask <= 1 {
		panic("invalid DSD composition")
	}
	permutation, ok := p.geometry.permutationForMask(uint32(mask) & 0x3ffff)
	if !ok {
		panic("invalid DSD permutation mask")
	}
	a := s_DsdClass6[lit0>>1].uTruth
	b := p.geometry.permuted[int(lit1>>1)*720+int(permutation)]
	b = sdmFlipMask(b, uint32(mask)>>18&63)
	if lit0&1 != 0 {
		a = ^a
	}
	if lit1&1 != 0 {
		b = ^b
	}
	truth := a & b
	if xor != 0 {
		truth = a ^ b
	}
	p.nAllDsd++
	config := sdmCheckDsd6(p, truth)
	if config < 0 {
		p.nNonDsd++
		return -1
	}
	class := config >> 17
	if cut != nil {
		if len(cut) < 1 || cut[0] < 0 || cut[0] > 6 || len(cut) < int(cut[0])+1 {
			panic("invalid DSD cut")
		}
		leaves := [6]int32{-1, -1, -1, -1, -1, -1}
		perm := p.geometry.permutations[(config&0xffff)>>6]
		for i := int32(0); i < cut[0]; i++ {
			leaves[perm[i]] = cut[i+1] ^ ((config >> uint(i)) & 1)
		}
		count := s_DsdClass6[class].nVars
		if len(cut) < int(count)+1 {
			panic("short DSD cut output")
		}
		for i := int32(0); i < count; i++ {
			if leaves[i] < 0 {
				panic("DSD cut misses a variable")
			}
		}
		cut[0] = count
		copy(cut[1:], leaves[:count])
	}
	p.nCountDsd[class]++
	return class*2 + (config >> 16 & 1)
}
func Sdm_ManReadDsdVarNum(p *Sdm_Man_t, class int32) int32    { return s_DsdClass6[class].nVars }
func Sdm_ManReadDsdAndNum(p *Sdm_Man_t, class int32) int32    { return s_DsdClass6[class].nAnds }
func Sdm_ManReadDsdClauseNum(p *Sdm_Man_t, class int32) int32 { return s_DsdClass6[class].nClauses }
func Sdm_ManReadDsdTruth(p *Sdm_Man_t, class int32) uint64    { return s_DsdClass6[class].uTruth }
func Sdm_ManReadDsdStr(p *Sdm_Man_t, class int32) string      { return s_DsdClass6[class].pStr }
func Sdm_ManReadCnfCosts(p *Sdm_Man_t, costs []int32, count int32) {
	if count != DSD_CLASS_NUM {
		panic("invalid DSD cost table size")
	}
	for i := int32(2); i < count; i++ {
		costs[i] = s_DsdClass6[i].nClauses
	}
}
func Sdm_ManPrintPerm(mask uint32) {
	for i := 0; i < 6; i++ {
		fmt.Printf("%d ", mask>>uint(3*i)&7)
	}
	fmt.Print("  ")
}
func Sdm_ManPrintDsdStats(p *Sdm_Man_t, verbose int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	absent := 0
	for i, count := range p.nCountDsd {
		if count == 0 {
			absent++
			continue
		}
		if verbose != 0 {
			fmt.Printf("%5d  :  %-20s   %8d\n", i, s_DsdClass6[i].pStr, count)
		}
	}
	fmt.Printf("Unused classes = %d (%.2f %%).  Non-DSD cuts = %d (%.2f %%).\n", absent, 100*float64(absent)/DSD_CLASS_NUM, p.nNonDsd, 100*float64(p.nNonDsd)/float64(max(uint64(1), p.nAllDsd)))
}

var sdmSingletonMu sync.Mutex
var s_SdmMan *Sdm_Man_t

func Sdm_ManRead() *Sdm_Man_t {
	sdmSingletonMu.Lock()
	defer sdmSingletonMu.Unlock()
	if s_SdmMan == nil {
		s_SdmMan = Sdm_ManAlloc()
	}
	s_SdmMan.mu.Lock()
	clear(s_SdmMan.nCountDsd[:])
	s_SdmMan.nNonDsd = 0
	s_SdmMan.nAllDsd = 0
	s_SdmMan.mu.Unlock()
	return s_SdmMan
}
func Sdm_ManQuit() {
	sdmSingletonMu.Lock()
	defer sdmSingletonMu.Unlock()
	Sdm_ManFree(s_SdmMan)
	s_SdmMan = nil
}
func Sdm_ManCanRead() int32 {
	_, _, err := sdmReadHashTable("dsdfuncs6.dat")
	if err == nil || os.IsNotExist(err) {
		return 1
	}
	return 0
}

func Sdm_ManTest() {
	p := Sdm_ManAlloc()
	defer Sdm_ManFree(p)
	mask := int32(0x3ffff ^ ((7 ^ 0) << 6) ^ ((7 ^ 1) << 9))
	cut := []int32{4, 10, 20, 30, 40, 0, 0}
	Sdm_ManComputeFunc(p, 4, 6, cut, mask, 0)
}
func sdmCollectDivisors(p *Sdm_Man_t, truth uint64, divisors [][]uint64, seen map[uint64]bool) {
	if seen[truth] {
		return
	}
	seen[truth] = true
	config := Sdm_ManCheckDsd6(p, truth)
	if config >= 0 && config>>17 < 2 {
		return
	}
	count := 0
	for v := int32(0); v < 6; v++ {
		if util.Abc_Tt6HasVar(truth, v) == 0 {
			continue
		}
		sdmCollectDivisors(p, util.Abc_Tt6Cofactor0(truth, v), divisors, seen)
		sdmCollectDivisors(p, util.Abc_Tt6Cofactor1(truth, v), divisors, seen)
		count++
	}
	if config >= 0 && count >= 2 && count <= 4 {
		if truth&1 != 0 {
			truth = ^truth
		}
		divisors[count] = append(divisors[count], truth)
	}
}
func Sdm_ManDivCollect_rec(truth uint64, divisors [][]uint64) [][]uint64 {
	if len(divisors) < 7 {
		panic("short DSD divisor array")
	}
	p := Sdm_ManRead()
	sdmCollectDivisors(p, truth, divisors, map[uint64]bool{})
	return divisors
}
func Sdm_ManDivTest() {
	truth := uint64(0x3f3fffff37003700)
	p := Sdm_ManAlloc()
	defer Sdm_ManFree(p)
	divisors := make([][]uint64, 7)
	sdmCollectDivisors(p, truth, divisors, map[uint64]bool{})
	var all []uint64
	elementary := Extra_TruthElementary(6)
	for v := int32(0); v < 6; v++ {
		words := elementary[v]
		all = append(all, uint64(words[0])|uint64(words[1])<<32)
	}
	for n := 2; n <= 4; n++ {
		slices.Sort(divisors[n])
		divisors[n] = slices.Compact(divisors[n])
		all = append(all, divisors[n]...)
	}
	for i := range all {
		fmt.Printf("%2d : \n", i)
	}
	manager := rsb.Rsb_ManAlloc(6, 200, 3, 1)
	defer rsb.Rsb_ManFree(manager)
	var off, on uint64
	if rsb.Rsb_ManPerformResub6(manager, 6, truth, all, &off, &on, 1) != 0 {
		fmt.Println("Decomposition exits.")
	}
}
