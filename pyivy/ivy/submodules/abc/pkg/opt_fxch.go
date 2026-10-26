package abc

// Fast cube-hashing extraction from src/opt/fxch/Fxch.h, Fxch.c,
// FxchDiv.c, FxchMan.c, and FxchSCHashTable.c.

import (
	"encoding/binary"
	"math/bits"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// FXCHCube is one SOP cube. Literals use Abc_Var2Lit encoding.
type FXCHCube struct {
	Output   int
	Literals []int
}

type FXCHOptions struct {
	ObjectIDMax, MaxExtractions, MaxCubes, MaxLiterals int
	Stop                                               func() bool
}

type FXCHResult struct {
	Cubes                 []FXCHCube
	Extracted, Variables  int
	InitialLiterals, Lits int
}

type FXCHSubCube struct {
	ID                       uint32
	Cube, Literal0, Literal1 int
}

type fxchGroupedCube struct {
	outputs  []int
	literals []int
}

type fxchCandidate struct {
	encoded     []int
	terms       [][]int
	weight      float64
	singleCubes []int
	pairs       [][2]int
}

// FXCHManager owns a cover and all extraction state.
type FXCHManager struct {
	cubes     []FXCHCube
	objectMax int
	levels    []int
	hashKeys  []uint32
	divisors  []fxchCandidate
	extracted int
	closed    bool
	options   FXCHOptions
}

func fxchVar(lit int) int { return lit >> 1 }
func fxchNot(lit int) int { return lit ^ 1 }
func fxchMake(value int, c bool) int {
	if c {
		return value<<1 | 1
	}
	return value << 1
}

func FXCHCountOnes(value uint32) int { return bits.OnesCount32(value) }

func fxchCloneCover(cubes []FXCHCube) []FXCHCube {
	r := make([]FXCHCube, len(cubes))
	for i := range cubes {
		r[i] = FXCHCube{Output: cubes[i].Output, Literals: slices.Clone(cubes[i].Literals)}
	}
	return r
}

func fxchNormalizeOptions(o FXCHOptions, n int) (FXCHOptions, error) {
	if o.ObjectIDMax < 0 || o.MaxExtractions < 0 || o.MaxCubes < 0 || o.MaxLiterals < 0 {
		return o, ErrDimension
	}
	if o.ObjectIDMax == 0 {
		o.ObjectIDMax = n
	}
	if o.MaxCubes == 0 {
		o.MaxCubes = 1 << 20
	}
	if o.MaxLiterals == 0 {
		o.MaxLiterals = 1 << 24
	}
	if o.ObjectIDMax > int(^uint(0)>>1)>>1 {
		return o, ErrLimit
	}
	return o, nil
}

func fxchValidateCover(cubes []FXCHCube, o FXCHOptions) (int, error) {
	if len(cubes) > o.MaxCubes {
		return 0, ErrLimit
	}
	lits := 0
	for _, cube := range cubes {
		if cube.Output < 0 || cube.Output >= o.ObjectIDMax {
			return 0, ErrDimension
		}
		last := -1
		lastVar := -1
		for _, lit := range cube.Literals {
			if lit < 0 || fxchVar(lit) >= o.ObjectIDMax || lit <= last || fxchVar(lit) == lastVar {
				return 0, ErrDimension
			}
			last = lit
			lastVar = fxchVar(lit)
			lits++
			if lits > o.MaxLiterals {
				return 0, ErrLimit
			}
		}
	}
	return lits, nil
}

func NewFXCHManager(cubes []FXCHCube, options FXCHOptions) (*FXCHManager, error) {
	o, err := fxchNormalizeOptions(options, options.ObjectIDMax)
	if err != nil {
		return nil, err
	}
	if o.ObjectIDMax == 0 {
		return nil, ErrDimension
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	if _, err = fxchValidateCover(cubes, o); err != nil {
		return nil, err
	}
	m := &FXCHManager{cubes: fxchCloneCover(cubes), objectMax: o.ObjectIDMax, options: o}
	m.computeLevels()
	m.generateHashKeys()
	return m, nil
}

func (m *FXCHManager) Close() {
	if m != nil {
		m.closed = true
		m.cubes = nil
		m.divisors = nil
		m.levels = nil
		m.hashKeys = nil
	}
}
func (m *FXCHManager) Cover() ([]FXCHCube, error) {
	if m == nil || m.closed {
		return nil, ErrInvariant
	}
	return fxchCloneCover(m.cubes), nil
}

func (m *FXCHManager) computeLevels() {
	m.levels = make([]int, m.objectMax)
	if len(m.cubes) == 0 {
		return
	}
	output, first := m.cubes[0].Output, 0
	for index, c := range m.cubes {
		if output != c.Output {
			m.levels[output] += index - first
			output, first = c.Output, index
		}
		l := 0
		for _, x := range c.Literals {
			if fxchVar(x) < len(m.levels) && m.levels[fxchVar(x)] > l {
				l = m.levels[fxchVar(x)]
			}
		}
		if c.Output < len(m.levels) && l > m.levels[c.Output] {
			m.levels[c.Output] = l
		}
	}
}

func fxchRandom(x uint32) uint32 { x ^= x << 13; x ^= x >> 17; x ^= x << 5; return x }
func (m *FXCHManager) generateHashKeys() {
	x := uint32(0x9e3779b9)
	m.hashKeys = make([]uint32, 2*m.objectMax)
	for i := range m.hashKeys {
		x = fxchRandom(x)
		m.hashKeys[i] = x & 0x3ffffff
	}
}

func FXCHGroupCubes(cubes []FXCHCube) ([]FXCHCube, [][]int, error) {
	result := make([]FXCHCube, 0, len(cubes))
	outputs := make([][]int, 0, len(cubes))
	index := make(map[string]int)
	for _, c := range cubes {
		key := fxchKey(c.Literals)
		at, ok := index[key]
		if !ok {
			index[key] = len(result)
			result = append(result, FXCHCube{Literals: slices.Clone(c.Literals)})
			outputs = append(outputs, []int{c.Output})
			continue
		}
		if !slices.Contains(outputs[at], c.Output) {
			outputs[at] = append(outputs[at], c.Output)
		}
	}
	return result, outputs, nil
}

func FXCHUngroupCubes(grouped []FXCHCube, outputs [][]int) ([]FXCHCube, error) {
	if len(grouped) != len(outputs) {
		return nil, ErrDimension
	}
	var r []FXCHCube
	for i, c := range grouped {
		for _, out := range outputs[i] {
			if out < 0 {
				return nil, ErrDimension
			}
			r = append(r, FXCHCube{Output: out, Literals: slices.Clone(c.Literals)})
		}
	}
	sort.SliceStable(r, func(i, j int) bool { return r[i].Output < r[j].Output })
	return r, nil
}

func fxchKey(v []int) string {
	var b strings.Builder
	for _, x := range v {
		b.WriteString(strconv.Itoa(x))
		b.WriteByte(',')
	}
	return b.String()
}
func fxchContains(c []int, x int) bool { _, ok := slices.BinarySearch(c, x); return ok }
func fxchRemove(c []int, values []int) []int {
	r := make([]int, 0, len(c))
	for _, x := range c {
		if !slices.Contains(values, x) {
			r = append(r, x)
		}
	}
	return r
}
func fxchIntersection(a, b []int) []int {
	r := make([]int, 0)
	for _, x := range a {
		if fxchContains(b, x) {
			r = append(r, x)
		}
	}
	return r
}

func (m *FXCHManager) candidateLevel(encoded []int) int {
	l := 0
	for _, e := range encoded {
		v := fxchVar(fxchVar(e))
		if v < len(m.levels) && m.levels[v] > l {
			l = m.levels[v]
		}
	}
	if l > 800 {
		return 800
	}
	return l
}

func (m *FXCHManager) addCandidate(list *[]fxchCandidate, byKey map[string]int, encoded []int, terms [][]int, increment float64, single bool, cube int, pair [2]int) {
	key := fxchKey(encoded)
	i, ok := byKey[key]
	if !ok {
		i = len(*list)
		byKey[key] = i
		penalty := .0009
		if single {
			penalty = .001
		}
		*list = append(*list, fxchCandidate{encoded: slices.Clone(encoded), terms: terms, weight: -float64(len(encoded)) + .9 - penalty*float64(m.candidateLevel(encoded))})
	}
	c := &(*list)[i]
	c.weight += increment
	if single {
		c.singleCubes = append(c.singleCubes, cube)
	} else {
		c.pairs = append(c.pairs, pair)
	}
}

func (m *FXCHManager) rebuildDivisors() error {
	if m == nil || m.closed {
		return ErrInvariant
	}
	list := make([]fxchCandidate, 0)
	byKey := make(map[string]int)
	for i, c := range m.cubes {
		if m.options.Stop != nil && m.options.Stop() {
			return ErrCancelled
		}
		for a := 0; a < len(c.Literals); a++ {
			for b := a + 1; b < len(c.Literals); b++ {
				x, y := c.Literals[a], c.Literals[b]
				encoded := []int{fxchMake(fxchNot(x), false), fxchMake(fxchNot(y), true)}
				m.addCandidate(&list, byKey, encoded, [][]int{{x, y}}, 1, true, i, [2]int{})
			}
		}
	}
	for i := 0; i < len(m.cubes); i++ {
		for j := i + 1; j < len(m.cubes); j++ {
			a, b := m.cubes[i], m.cubes[j]
			if a.Output != b.Output {
				continue
			}
			common := fxchIntersection(a.Literals, b.Literals)
			da, db := fxchRemove(a.Literals, common), fxchRemove(b.Literals, common)
			var encoded []int
			var definition [][]int
			if len(da) == 1 && len(db) == 1 {
				x, y := da[0], db[0]
				if x > y {
					x, y = y, x
				}
				encoded = []int{fxchMake(x, false), fxchMake(y, true)}
				// A two-literal divisor is represented as z = !x & !y. A
				// double-cube occurrence therefore uses !z = x | y.
				definition = [][]int{{fxchNot(x), fxchNot(y)}}
			} else if len(da) == 2 && len(db) == 2 {
				raw := []int{fxchMake(da[0], false), fxchMake(db[0], true), fxchMake(da[1], false), fxchMake(db[1], true)}
				normalized, _, normalizeErr := FXCHNormalizeDivisor(raw)
				if normalizeErr != nil {
					continue
				}
				encoded = normalized
				left, right := FXCHSeparateDivisor(encoded)
				definition = [][]int{left, right}
			} else if len(da) == 1 && len(db) == 2 || len(da) == 2 && len(db) == 1 {
				one, two := da, db
				if len(one) == 2 {
					one, two = two, one
				}
				x, y := one[0], -1
				if two[0] == fxchNot(x) {
					y = two[1]
				} else if two[1] == fxchNot(x) {
					y = two[0]
				}
				if y < 0 {
					continue
				}
				if x > y {
					x, y = y, x
				}
				encoded = []int{fxchMake(x, false), fxchMake(y, true)}
				definition = [][]int{{fxchNot(x), fxchNot(y)}}
			} else {
				continue
			}
			m.addCandidate(&list, byKey, encoded, definition, float64(len(common)+len(encoded)-1), false, -1, [2]int{i, j})
		}
	}
	m.divisors = list
	return nil
}

func (m *FXCHManager) bestDivisor() int {
	best := -1
	for i := range m.divisors {
		if m.divisors[i].weight > 0 && (best < 0 || m.divisors[i].weight > m.divisors[best].weight) {
			best = i
		}
	}
	return best
}

func fxchCanonical(c FXCHCube) FXCHCube {
	sort.Ints(c.Literals)
	c.Literals = slices.Compact(c.Literals)
	return c
}

func fxchSimplifyCover(cubes []FXCHCube) []FXCHCube {
	dead := make([]bool, len(cubes))
	for i := range cubes {
		if dead[i] {
			continue
		}
		for j := range cubes {
			if i == j || dead[j] || cubes[i].Output != cubes[j].Output || len(cubes[i].Literals) > len(cubes[j].Literals) {
				continue
			}
			subset := true
			for _, lit := range cubes[i].Literals {
				if !fxchContains(cubes[j].Literals, lit) {
					subset = false
					break
				}
			}
			if subset {
				dead[j] = true
			}
		}
	}
	result := cubes[:0]
	for i := range cubes {
		if !dead[i] {
			result = append(result, cubes[i])
		}
	}
	return result
}

func (m *FXCHManager) Update(divisor int) error {
	if m == nil || m.closed || divisor < 0 || divisor >= len(m.divisors) {
		return ErrDimension
	}
	if m.options.Stop != nil && m.options.Stop() {
		return ErrCancelled
	}
	work := fxchCloneCover(m.cubes)
	c := m.divisors[divisor]
	z := m.objectMax
	if z >= int(^uint(0)>>1)>>1 {
		return ErrLimit
	}
	if len(work)+len(c.terms) > m.options.MaxCubes {
		return ErrLimit
	}
	if len(c.terms) == 1 {
		seen := make(map[int]bool)
		for _, ci := range c.singleCubes {
			if seen[ci] || ci >= len(work) {
				continue
			}
			seen[ci] = true
			work[ci].Literals = append(fxchRemove(work[ci].Literals, c.terms[0]), fxchMake(z, false))
			work[ci] = fxchCanonical(work[ci])
		}
	} else {
		dead := make([]bool, len(work))
		used := make([]bool, len(work))
		for _, p := range c.pairs {
			if used[p[0]] || used[p[1]] || work[p[0]].Output != work[p[1]].Output {
				continue
			}
			used[p[0]], used[p[1]] = true, true
			common := fxchIntersection(work[p[0]].Literals, work[p[1]].Literals)
			polarity := len(c.encoded) == 2
			work[p[0]].Literals = append(common, fxchMake(z, polarity))
			work[p[0]] = fxchCanonical(work[p[0]])
			dead[p[1]] = true
		}
		n := work[:0]
		for i := range work {
			if !dead[i] {
				n = append(n, work[i])
			}
		}
		work = n
	}
	for _, term := range c.terms {
		work = append(work, fxchCanonical(FXCHCube{Output: z, Literals: slices.Clone(term)}))
	}
	work = fxchSimplifyCover(work)
	lits := 0
	for _, x := range work {
		lits += len(x.Literals)
	}
	if lits > m.options.MaxLiterals {
		return ErrLimit
	}
	m.cubes = work
	m.objectMax++
	m.extracted++
	m.levels = append(m.levels, 1)
	m.generateHashKeys()
	return nil
}

func (m *FXCHManager) Extract() error {
	for m.options.MaxExtractions == 0 || m.extracted < m.options.MaxExtractions {
		if err := m.rebuildDivisors(); err != nil {
			return err
		}
		best := m.bestDivisor()
		if best < 0 {
			return nil
		}
		if err := m.Update(best); err != nil {
			return err
		}
	}
	return nil
}

func FXCHFastExtract(cubes []FXCHCube, options FXCHOptions) (FXCHResult, error) {
	o, err := fxchNormalizeOptions(options, options.ObjectIDMax)
	if err != nil {
		return FXCHResult{}, err
	}
	initial, err := fxchValidateCover(cubes, o)
	if err != nil {
		return FXCHResult{}, err
	}
	m, err := NewFXCHManager(cubes, o)
	if err != nil {
		return FXCHResult{}, err
	}
	defer m.Close()
	if err = m.Extract(); err != nil {
		return FXCHResult{}, err
	}
	result := fxchCloneCover(m.cubes)
	sort.SliceStable(result, func(i, j int) bool { return result[i].Output < result[j].Output })
	lits := 0
	for _, c := range result {
		lits += len(c.Literals)
	}
	return FXCHResult{Cubes: result, Extracted: m.extracted, Variables: m.objectMax, InitialLiterals: initial, Lits: lits}, nil
}

func FXCHSeparateDivisor(divisor []int) (first, second []int) {
	for _, x := range divisor {
		if x&1 == 0 {
			first = append(first, fxchVar(x))
		} else {
			second = append(second, fxchVar(x))
		}
	}
	if len(divisor) == 4 {
		sort.Ints(first)
		sort.Ints(second)
	}
	return
}

// FXCHNormalizeDivisor preserves Fxch_DivNormalize's mux divisor spelling.
func FXCHNormalizeDivisor(divisor []int) ([]int, bool, error) {
	if len(divisor) != 4 {
		return nil, false, ErrDimension
	}
	l := slices.Clone(divisor)
	a0, b0, a1, b1 := -1, -1, -1, -1
	if (l[0]&1) != (l[1]&1) && l[0]>>2 == l[1]>>2 {
		if (l[2] & 1) == (l[3] & 1) {
			return nil, false, ErrInvariant
		}
		a0, b0 = fxchVar(l[0]), fxchVar(l[1])
		if (l[0] & 1) == (l[2] & 1) {
			a1, b1 = fxchVar(l[2]), fxchVar(l[3])
		} else {
			a1, b1 = fxchVar(l[3]), fxchVar(l[2])
		}
	} else if (l[1]&1) != (l[2]&1) && l[1]>>2 == l[2]>>2 {
		if (l[0] & 1) == (l[3] & 1) {
			return nil, false, ErrInvariant
		}
		a0, b0 = fxchVar(l[1]), fxchVar(l[2])
		if (l[1] & 1) == (l[0] & 1) {
			a1, b1 = fxchVar(l[0]), fxchVar(l[3])
		} else {
			a1, b1 = fxchVar(l[3]), fxchVar(l[0])
		}
	} else if (l[2]&1) != (l[3]&1) && l[2]>>2 == l[3]>>2 {
		if (l[0] & 1) == (l[1] & 1) {
			return nil, false, ErrInvariant
		}
		a0, b0 = fxchVar(l[2]), fxchVar(l[3])
		if (l[2] & 1) == (l[0] & 1) {
			a1, b1 = fxchVar(l[0]), fxchVar(l[1])
		} else {
			a1, b1 = fxchVar(l[1]), fxchVar(l[0])
		}
	} else {
		return nil, false, ErrInvariant
	}
	if a0 != fxchNot(b0) {
		return nil, false, ErrInvariant
	}
	if a0&1 != 0 {
		a0, b0, a1, b1 = b0, a0, b1, a1
	}
	complement := false
	if a1&1 != 0 {
		a1, b1, complement = fxchNot(a1), fxchNot(b1), true
	}
	return []int{fxchMake(a0, false), fxchMake(b0, true), fxchMake(a1, false), fxchMake(b1, true)}, complement, nil
}

func FXCHDivisorNotConstantOne(divisor []int) bool {
	return len(divisor) != 2 || fxchVar(divisor[0]) != fxchNot(fxchVar(divisor[1]))
}

func FXCHRemoveDivisorLiterals(cube0, cube1, divisor []int) ([]int, []int, bool, int, error) {
	a, b := slices.Clone(cube0), slices.Clone(cube1)
	removedPositive, removedNegative, removedSecond := 0, 0, 0
	remove := func(p *[]int, x int) bool {
		i := slices.Index(*p, x)
		if i < 0 {
			return false
		}
		*p = append((*p)[:i], (*p)[i+1:]...)
		return true
	}
	for _, d := range divisor {
		x := fxchVar(d)
		if remove(&a, x) {
			if x&1 != 0 {
				removedNegative++
			} else {
				removedPositive++
			}
		}
	}
	for _, d := range divisor {
		if remove(&b, fxchVar(d)) {
			removedSecond++
		}
	}
	if len(divisor) == 2 {
		for _, d := range divisor {
			x := fxchNot(fxchVar(d))
			remove(&a, x)
			remove(&b, x)
		}
	}
	removed := removedPositive + removedNegative + removedSecond
	compl := false
	if len(divisor) == 4 {
		x := []int{fxchVar(divisor[0]), fxchVar(divisor[1]), fxchVar(divisor[2]), fxchVar(divisor[3])}
		compl = x[0] == fxchNot(x[1]) && x[2] == fxchNot(x[3]) && removedNegative == 1
		if x[0] == fxchNot(x[1]) && removed == 2 {
			compl = true
			for index, d := range divisor {
				x := fxchVar(d)
				if index > 1 {
					x = fxchNot(x)
				}
				if remove(&a, x) {
					removed++
				}
			}
			for index, d := range divisor {
				x := fxchVar(d)
				if index > 1 {
					x = fxchNot(x)
				}
				if remove(&b, x) {
					removed++
				}
			}
		}
	}
	if removed != len(divisor) {
		return nil, nil, false, removed, ErrInvariant
	}
	return a, b, compl, removed, nil
}

// FXCHMurmur32 is the source MurmurHash3_x86_32 helper.
func FXCHMurmur32(value, seed uint32) uint32 {
	var bytes [4]byte
	binary.LittleEndian.PutUint32(bytes[:], value)
	k := binary.LittleEndian.Uint32(bytes[:])
	k *= 0xcc9e2d51
	k = bits.RotateLeft32(k, 15)
	k *= 0x1b873593
	h := seed ^ k
	h = bits.RotateLeft32(h, 13)
	h = h*5 + 0xe6546b64
	h ^= 4
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	return h ^ (h >> 16)
}

type FXCHSubCubeHash struct {
	bins    [][]FXCHSubCube
	mask    uint32
	entries int
}

func NewFXCHSubCubeHash(entries int) (*FXCHSubCubeHash, error) {
	if entries < 0 {
		return nil, ErrDimension
	}
	if entries > 1<<24 {
		return nil, ErrLimit
	}
	size := 1
	for size < entries+1 {
		size <<= 1
	}
	return &FXCHSubCubeHash{bins: make([][]FXCHSubCube, size), mask: uint32(size - 1)}, nil
}
func (h *FXCHSubCubeHash) Insert(s FXCHSubCube) error {
	if h == nil || len(h.bins) == 0 || s.Cube < 0 || uint64(s.Cube) > uint64(^uint32(0)) || s.Literal0 < 0 || s.Literal0 > 0xffff || s.Literal1 < 0 || s.Literal1 > 0xffff {
		return ErrDimension
	}
	b := FXCHMurmur32(s.ID, 0x9747b28c) & h.mask
	h.bins[b] = append(h.bins[b], s)
	h.entries++
	return nil
}
func (h *FXCHSubCubeHash) Remove(s FXCHSubCube) error {
	if h == nil || len(h.bins) == 0 {
		return ErrDimension
	}
	b := FXCHMurmur32(s.ID, 0x9747b28c) & h.mask
	for i, x := range h.bins[b] {
		if x == s {
			h.bins[b] = append(h.bins[b][:i], h.bins[b][i+1:]...)
			h.entries--
			return nil
		}
	}
	return ErrInvariant
}
func (h *FXCHSubCubeHash) Entries() int {
	if h == nil {
		return 0
	}
	return h.entries
}
func (h *FXCHSubCubeHash) MemoryBytes() int {
	if h == nil {
		return 0
	}
	n := len(h.bins) * 24
	for _, b := range h.bins {
		n += cap(b) * 16
	}
	return n
}
