package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var testDsdVariables = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func testDsdEval(text string) (uint64, int) {
	if text[0] == '!' {
		value, n := testDsdEval(text[1:])
		return ^value, n + 1
	}
	if text[0] == '0' {
		return 0, 1
	}
	if text[0] >= 'a' && text[0] <= 'f' {
		return testDsdVariables[text[0]-'a'], 1
	}
	end := byte(')')
	if text[0] == '[' {
		end = ']'
	} else if text[0] == '<' {
		end = '>'
	}
	var values []uint64
	pos := 1
	for text[pos] != end {
		value, n := testDsdEval(text[pos:])
		values = append(values, value)
		pos += n
	}
	var result uint64
	switch text[0] {
	case '(':
		result = ^uint64(0)
		for _, v := range values {
			result &= v
		}
	case '[':
		for _, v := range values {
			result ^= v
		}
	case '<':
		if len(values) != 3 {
			panic("mux arity")
		}
		result = values[2]&values[1] | ^values[2]&values[0]
	default:
		panic("bad DSD operator")
	}
	return result, pos + 1
}
func TestDsdClassRepresentatives(t *testing.T) {
	for i, entry := range s_DsdClass6 {
		got, n := testDsdEval(entry.pStr)
		if n != len(entry.pStr) || got != entry.uTruth {
			t.Fatal(i, entry.pStr, got, entry.uTruth)
		}
		if sdmTruthSignature(got).support != byte(entry.nVars) {
			t.Fatal("class support", i)
		}
	}
	g := sdmGetGeometry()
	if len(g.permutationIndex) != 720 || g.permutations[719] != [6]byte{0, 1, 2, 3, 4, 5} {
		t.Fatal("permutation schedule")
	}
	for class, entry := range s_DsdClass6 {
		if g.permuted[class*720+719] != entry.uTruth {
			t.Fatal("cycle", class)
		}
	}
}
func testDsdConfigTruth(g *sdmGeometry, config int32) uint64 {
	value := sdmFlipMask(g.permuted[int(config>>17)*720+int(config>>6&1023)], uint32(config)&63)
	if config&(1<<16) != 0 {
		value = ^value
	}
	return value
}
func TestDsdClassesAndTransformedQueries(t *testing.T) {
	p := Sdm_ManAlloc()
	defer Sdm_ManFree(p)
	for class, entry := range s_DsdClass6 {
		config := Sdm_ManCheckDsd6(p, entry.uTruth)
		if config < 0 || config>>17 != int32(class) || testDsdConfigTruth(p.geometry, config) != entry.uTruth {
			t.Fatal("class lookup", class, config)
		}
	}
	rng := rand.New(rand.NewSource(13))
	for i := 0; i < 200; i++ {
		class := rng.Intn(DSD_CLASS_NUM)
		perm := rng.Intn(720)
		phase := uint32(rng.Intn(64))
		truth := sdmFlipMask(p.geometry.permuted[class*720+perm], phase)
		if i%2 != 0 {
			truth = ^truth
		}
		config := Sdm_ManCheckDsd6(p, truth)
		if config < 0 || config>>17 != int32(class) || testDsdConfigTruth(p.geometry, config) != truth {
			t.Fatal(class, perm, phase, config)
		}
	}
}
func TestDsdBinaryTableValidation(t *testing.T) {
	g := sdmGetGeometry()
	configs := []int32{0, 1 << 17, 2 << 17}
	data := make([]byte, 12*len(configs))
	for i, config := range configs {
		truth := testDsdConfigTruth(g, config)
		if truth&1 != 0 {
			truth = ^truth
			config ^= 1 << 16
			configs[i] = config
		}
		binary.LittleEndian.PutUint64(data[8*i:], truth)
		binary.LittleEndian.PutUint32(data[8*len(configs)+4*i:], uint32(config))
	}
	path := filepath.Join(t.TempDir(), "table")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	index, loaded, err := sdmReadHashTable(path)
	if err != nil || len(index) != 3 || len(loaded) != 3 {
		t.Fatal(index, loaded, err)
	}
	data[0] ^= 2
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sdmReadHashTable(path); err == nil {
		t.Fatal("invalid record accepted")
	}
	if err := os.WriteFile(path, []byte{0}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sdmReadHashTable(path); err == nil {
		t.Fatal("truncated table accepted")
	}
}

func TestDsdFlipMaskBitOracle(t *testing.T) {
	const truth = uint64(0x83a19c67f50d24be)
	for mask := uint32(0); mask < 64; mask++ {
		var want uint64
		for bit := uint32(0); bit < 64; bit++ {
			if truth&(uint64(1)<<uint(bit^mask)) != 0 {
				want |= uint64(1) << uint(bit)
			}
		}
		if got := sdmFlipMask(truth, mask); got != want {
			t.Fatal(mask, got, want)
		}
	}
}
func testDsdEvaluateLiteral(literal int32, cut []int32) uint64 {
	class := s_DsdClass6[literal>>1]
	var truth uint64
	for bit := 0; bit < 64; bit++ {
		assignment := 0
		for v := int32(0); v < class.nVars; v++ {
			leaf := cut[v+1]
			value := (bit >> uint(leaf>>1)) & 1
			value ^= int(leaf & 1)
			assignment |= value << uint(v)
		}
		value := class.uTruth >> uint(assignment) & 1
		value ^= uint64(literal & 1)
		truth |= value << uint(bit)
	}
	return truth
}
func TestDsdCompositionAndCutRemapping(t *testing.T) {
	p := Sdm_ManAlloc()
	defer Sdm_ManFree(p)
	// Swap positions 0/1 so the two literal-a inputs represent a and b.
	mask := int32(1 | 0<<3 | 2<<6 | 3<<9 | 4<<12 | 5<<15)
	for _, tc := range []struct {
		a, b, xor int32
		truth     uint64
	}{{2, 2, 0, testDsdVariables[0] & testDsdVariables[1]}, {2, 2, 1, testDsdVariables[0] ^ testDsdVariables[1]}, {3, 2, 0, ^testDsdVariables[0] & testDsdVariables[1]}, {0, 2, 0, 0}, {1, 2, 0, testDsdVariables[1]}} {
		cut := []int32{2, 0, 2, 0, 0, 0, 0}
		literal := Sdm_ManComputeFunc(p, tc.a, tc.b, cut, mask, tc.xor)
		if literal < 0 || testDsdEvaluateLiteral(literal, cut) != tc.truth {
			t.Fatal(tc, literal, cut)
		}
	}
	costs := make([]int32, DSD_CLASS_NUM)
	costs[0], costs[1] = -1, -1
	Sdm_ManReadCnfCosts(p, costs, DSD_CLASS_NUM)
	if costs[0] != -1 || costs[1] != -1 || costs[2] != 3 {
		t.Fatal("cost slots")
	}
	if p.nAllDsd != 5 || p.nNonDsd != 0 {
		t.Fatal("stats")
	}
}
func TestDsdOrbitMaterializationAndCache(t *testing.T) {
	g := sdmGetGeometry()
	index, configs := sdmBuildHashOrbits(g, 4)
	if len(index) != 82 || len(configs) != 82 {
		t.Fatal(len(index), len(configs))
	}
	for truth, id := range index {
		if testDsdConfigTruth(g, configs[id]) != truth {
			t.Fatal("materialized config")
		}
	}
	p := &Sdm_Man_t{geometry: g, pHash: index, vConfgRes: configs}
	truth := testDsdVariables[0] & testDsdVariables[1]
	if config := Sdm_ManCheckDsd6(p, truth); config>>17 != 2 {
		t.Fatal(config)
	}
}
func TestDsdSingletonResetAndDivisors(t *testing.T) {
	p := Sdm_ManRead()
	p.mu.Lock()
	p.nCountDsd[0] = 3
	p.nAllDsd = 5
	p.nNonDsd = 2
	p.mu.Unlock()
	if q := Sdm_ManRead(); q != p || q.nAllDsd != 0 || q.nNonDsd != 0 || q.nCountDsd[0] != 0 {
		t.Fatal("singleton reset")
	}
	divisors := make([][]uint64, 7)
	Sdm_ManDivCollect_rec(testDsdVariables[0]&testDsdVariables[1]&testDsdVariables[2], divisors)
	if len(divisors[3]) != 1 || len(divisors[2]) == 0 {
		t.Fatal("divisor collection", divisors)
	}
	Sdm_ManQuit()
	Sdm_ManQuit()
}

// Compiled original-C oracle: all595 metadata records, its720 permutation
// vectors, and every595*720 precomputed truth word, in original schedule order.
func TestDsdOriginalCPermutationOracle(t *testing.T) {
	g := sdmGetGeometry()
	hash := sha256.New()
	for _, entry := range s_DsdClass6 {
		fmt.Fprintf(hash, "%d,%d,%d,%016x,%s\n", entry.nVars, entry.nAnds, entry.nClauses, entry.uTruth, entry.pStr)
	}
	for _, perm := range g.permutations {
		hash.Write(perm[:])
	}
	record := make([]byte, 8)
	for _, truth := range g.permuted {
		binary.LittleEndian.PutUint64(record, truth)
		hash.Write(record)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "583f6f7012857410eeab1d7bbaf76765fa5fe0919a6fec60643d7b3625dfb834" {
		t.Fatal(got)
	}
}
func TestDsdFullOrbitMaterialization(t *testing.T) {
	g := sdmGetGeometry()
	index, configs := sdmBuildHashOrbits(g, DSD_CLASS_NUM)
	if len(index) != 2866420 || len(configs) != 2866420 {
		t.Fatal("original complete DSD6 table count", len(index), len(configs))
	}
	for class := range s_DsdClass6 {
		truth := s_DsdClass6[class].uTruth
		normalized := truth
		if normalized&1 != 0 {
			normalized = ^normalized
		}
		id, exists := index[normalized]
		if !exists || configs[id]>>17 != int32(class) {
			t.Fatal("class absent from full materialization", class)
		}
	}
	// Differentially compare the exact lazy search with the complete original
	// orbit map, including first-config tie ordering and non-DSD misses.
	p := &Sdm_Man_t{geometry: g, queryCache: map[uint64]int32{}}
	seen := 0
	for truth, id := range index {
		if got := Sdm_ManCheckDsd6(p, truth); got != configs[id] {
			t.Fatal("lazy/materialized config mismatch", truth, got, configs[id])
		}
		seen++
		if seen == 100 {
			break
		}
	}
	rng := rand.New(rand.NewSource(72))
	for i := 0; i < 100; i++ {
		truth := rng.Uint64() &^ 1
		want := int32(-1)
		if id, exists := index[truth]; exists {
			want = configs[id]
		}
		if got := Sdm_ManCheckDsd6(p, truth); got != want {
			t.Fatal("lazy miss mismatch", truth, got, want)
		}
	}

}

func TestDsdConcurrentQueries(t *testing.T) {
	p := Sdm_ManAlloc()
	defer Sdm_ManFree(p)
	var group sync.WaitGroup
	failures := make(chan int32, 4)
	for worker := 0; worker < 4; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for phase := uint32(0); phase < 64; phase++ {
				truth := sdmFlipMask(s_DsdClass6[2].uTruth, phase)
				config := Sdm_ManCheckDsd6(p, truth)
				if config < 0 || config>>17 != 2 || testDsdConfigTruth(p.geometry, config) != truth {
					failures <- config
					return
				}
			}
		}()
	}
	group.Wait()
	close(failures)
	for result := range failures {
		t.Fatal(result)
	}
}
func TestDsdPartialMasksAndClosedManager(t *testing.T) {
	g := sdmGetGeometry()
	identity := uint32(0 | 1<<3 | 2<<6 | 3<<9 | 4<<12 | 5<<15)
	full, ok := g.permutationForMask(identity)
	if !ok || full != 719 {
		t.Fatal("identity index", full, ok)
	}
	partial := uint32(0x3ffff) ^ ((7 ^ 0) << 6) ^ ((7 ^ 1) << 9)
	index, ok := g.permutationForMask(partial)
	if !ok {
		t.Fatal("partial mask")
	}
	perm := g.permutations[index]
	if perm[2] != 0 || perm[3] != 1 {
		t.Fatal(perm)
	}
	for _, mask := range []uint32{0, 6, 2} {
		if _, ok := g.permutationForMask(mask); ok {
			t.Fatal("invalid mask", mask)
		}
	}
	p := Sdm_ManAlloc()
	cut := []int32{4, 0, 2, 4, 6, 0, 0}
	literal := Sdm_ManComputeFunc(p, 4, 6, cut, int32(partial), 0)
	want := (testDsdVariables[0] & testDsdVariables[1]) & (testDsdVariables[2] ^ testDsdVariables[3])
	if literal < 0 || testDsdEvaluateLiteral(literal, cut) != want {
		t.Fatal(literal, cut)
	}
	Sdm_ManFree(p)
	Sdm_ManFree(p)
	defer func() {
		if recover() == nil {
			t.Error("closed manager accepted")
		}
	}()
	Sdm_ManCheckDsd6(p, 0)
}
