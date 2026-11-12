package bdc

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

var spfdTruths = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

type Bdc_Nod_t struct {
	iFan0g, iFan0n, iFan1g, iFan1n, Type, Weight int32
	Truth                                        uint64
}
type Bdc_Ent_t struct {
	iFan0, iFan1                                    uint32
	fCompl0, fCompl1, fCompl, fExor, fMark0, fMark1 bool
	iNext, iList                                    int32
	Truth                                           uint64
}

const BDC_TERM = 0x1fffffff

func Bdc_CountOnes(t uint64) int32 { return int32(bits.OnesCount64(t)) }
func Bdc_CountSpfd(t, f uint64) int32 {
	return Bdc_CountOnes(^t&^f)*Bdc_CountOnes(t&f) + Bdc_CountOnes(^t&f)*Bdc_CountOnes(t&^f)
}
func Bdc_Cof6(t uint64, v, phase int32) uint64 {
	if v < 0 || v >= 6 {
		panic("bdc: invalid SPFD variable")
	}
	mask := spfdTruths[v]
	if phase != 0 {
		return t&mask | (t&mask)>>(1<<v)
	}
	return t&^mask | (t&^mask)<<(1<<v)
}
func Bdc_SpfdAdjCost(t uint64) int32 {
	var cost int32
	for v := int32(0); v < 6; v++ {
		cost += Bdc_CountOnes(Bdc_Cof6(t, v, 0) ^ Bdc_Cof6(t, v, 1))
	}
	return cost
}
func spfdGate(a, b uint64, kind int32) uint64 {
	if kind == 4 {
		return a ^ b
	}
	if kind&1 != 0 {
		a = ^a
	}
	if kind&2 != 0 {
		b = ^b
	}
	return a & b
}

// SPFDResult retains the beam's level/index references and the selected output
// phase. Level means the original search depth, not a guaranteed minimal DAG.
type SPFDResult struct {
	Levels     [][]Bdc_Nod_t
	Root       Bdc_Nod_t
	Level      int32
	Complement bool
	Constant   *bool
}

func spfdExpression(node Bdc_Nod_t, level int32, levels [][]Bdc_Nod_t) string {
	if level == 0 {
		return string(rune('a' + node.iFan0n))
	}
	child := func(group, index int32) string {
		if group == 0 {
			return string(rune('a' + index))
		}
		return spfdExpression(levels[group][index], group, levels)
	}
	a, b := child(node.iFan0g, node.iFan0n), child(node.iFan1g, node.iFan1n)
	if node.Type&1 != 0 {
		a = "!" + a
	}
	if node.Type&2 != 0 {
		b = "!" + b
	}
	op := "*"
	if node.Type&4 != 0 {
		op = "^"
	}
	return "(" + a + op + b + ")"
}
func (r *SPFDResult) Expression() string {
	if r.Constant != nil {
		if *r.Constant {
			return "1"
		}
		return "0"
	}
	s := spfdExpression(r.Root, r.Level, r.Levels)
	if r.Complement {
		return "!" + s
	}
	return s
}
func Bdc_SpfdPrint_rec(node *Bdc_Nod_t, level int32, levels [][]Bdc_Nod_t) {
	fmt.Fprint(os.Stdout, spfdExpression(*node, level, levels))
}
func Bdc_SpfdPrint(node *Bdc_Nod_t, level int32, levels [][]Bdc_Nod_t, truth uint64) {
	fmt.Fprintf(os.Stdout, "%016X   %016X   %s    %d\n", node.Truth, truth^node.Truth, spfdExpression(*node, level, levels), node.Weight)
}
func DecomposeSPFD(truth uint64, nVars, nCandidates, maxDepth int) (*SPFDResult, error) {
	if nVars < 1 || nVars > 6 || nCandidates < 1 || nCandidates >= 1<<12 || maxDepth < 0 || maxDepth >= 1<<8 {
		return nil, fmt.Errorf("bdc: invalid SPFD search bounds")
	}
	result := &SPFDResult{}
	if truth == 0 || truth == ^uint64(0) {
		v := truth != 0
		result.Constant = &v
		return result, nil
	}
	vars := make([]Bdc_Nod_t, nVars)
	for i := range vars {
		vars[i] = Bdc_Nod_t{iFan0n: int32(i), Truth: spfdTruths[i], Weight: Bdc_CountSpfd(spfdTruths[i], truth)}
		if truth == vars[i].Truth || truth == ^vars[i].Truth {
			result.Root = vars[i]
			result.Complement = truth != vars[i].Truth
			result.Levels = [][]Bdc_Nod_t{vars}
			return result, nil
		}
	}
	result.Levels = append(result.Levels, vars)
	for depth := 1; depth <= maxDepth; depth++ {
		var candidates []Bdc_Nod_t
		add := func(g0, i, g1, j int) {
			a, b := result.Levels[g0][i], result.Levels[g1][j]
			for kind := int32(0); kind < 5; kind++ {
				f := spfdGate(a.Truth, b.Truth, kind)
				candidates = append(candidates, Bdc_Nod_t{iFan0g: int32(g0), iFan0n: int32(i), iFan1g: int32(g1), iFan1n: int32(j), Type: kind, Truth: f, Weight: Bdc_CountSpfd(f, truth)})
			}
		}
		if depth == 1 {
			for i := range vars {
				for j := i + 1; j < len(vars); j++ {
					add(0, i, 0, j)
				}
			}
		} else {
			previous := depth - 1
			for group := 0; group < previous; group++ {
				for i := range result.Levels[group] {
					for j := range result.Levels[previous] {
						add(group, i, previous, j)
					}
				}
			}
			for i := range result.Levels[previous] {
				for j := i + 1; j < len(result.Levels[previous]); j++ {
					add(previous, i, previous, j)
				}
			}
		}
		for _, candidate := range candidates {
			if candidate.Truth == truth || candidate.Truth == ^truth {
				result.Root = candidate
				result.Level = int32(depth)
				result.Complement = candidate.Truth != truth
				return result, nil
			}
		}
		if len(candidates) == 0 {
			break
		}
		if depth > 1 {
			for i, j := 0, len(candidates)-1; i < j; i, j = i+1, j-1 {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
			sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Weight > candidates[j].Weight })
			if len(candidates) > nCandidates {
				candidates = candidates[:nCandidates]
			}
		}
		result.Levels = append(result.Levels, candidates)
	}
	return nil, fmt.Errorf("bdc: SPFD search exhausted depth %d", maxDepth)
}
func Bdc_SpfdDecompose(truth uint64, nVars, nCandidates, maxDepth int32) {
	r, err := DecomposeSPFD(truth, int(nVars), int(nCandidates), int(maxDepth))
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		return
	}
	fmt.Fprintf(os.Stdout, "Function can be implemented at level %d: %s\n", r.Level, r.Expression())
}

func spfdMark(entries []Bdc_Ent_t, entry *Bdc_Ent_t, which bool) int32 {
	if entry.iFan0 == BDC_TERM {
		return 0
	}
	mark := &entry.fMark0
	other := entry.fMark1
	if which {
		mark = &entry.fMark1
		other = entry.fMark0
	}
	if *mark {
		return 0
	}
	*mark = true
	count := int32(0)
	if other {
		count = 1
	}
	return count + spfdMark(entries, &entries[entry.iFan0], which) + spfdMark(entries, &entries[entry.iFan1], which)
}
func Bdc_SpfdMark0(entries []Bdc_Ent_t, e *Bdc_Ent_t) int32 { return spfdMark(entries, e, false) }
func Bdc_SpfdMark1(entries []Bdc_Ent_t, e *Bdc_Ent_t) int32 { return spfdMark(entries, e, true) }
func spfdUnmark(entries []Bdc_Ent_t, e *Bdc_Ent_t, which bool) {
	if e.iFan0 == BDC_TERM {
		return
	}
	mark := &e.fMark0
	if which {
		mark = &e.fMark1
	}
	if !*mark {
		return
	}
	*mark = false
	spfdUnmark(entries, &entries[e.iFan0], which)
	spfdUnmark(entries, &entries[e.iFan1], which)
}
func Bdc_SpfdUnmark0(entries []Bdc_Ent_t, e *Bdc_Ent_t) { spfdUnmark(entries, e, false) }
func Bdc_SpfdUnmark1(entries []Bdc_Ent_t, e *Bdc_Ent_t) { spfdUnmark(entries, e, true) }
func Bdc_SpfdCheckOverlap(entries []Bdc_Ent_t, a, b *Bdc_Ent_t) int32 {
	if Bdc_SpfdMark0(entries, a) != 0 {
		panic("bdc: dirty SPFD marks")
	}
	count := Bdc_SpfdMark1(entries, b)
	Bdc_SpfdUnmark0(entries, a)
	Bdc_SpfdUnmark1(entries, b)
	return count
}
func Bdc_SpfdHashValue(t uint64, size int32) int32 {
	if size <= 0 {
		panic("bdc: invalid hash size")
	}
	primes := [8]uint32{12582917, 25165843, 50331653, 100663319, 201326611, 402653189, 805306457, 1610612741}
	var value uint32
	for i, p := range primes {
		value ^= p * uint32(byte(t>>uint(8*i)))
	}
	return int32(value % uint32(size))
}
func Bdc_SpfdHashLookup(entries []Bdc_Ent_t, size int32, t uint64) *int32 {
	head := &entries[Bdc_SpfdHashValue(t, size)]
	slot := &head.iList
	for *slot != 0 {
		e := &entries[*slot]
		if e.Truth == t {
			return nil
		}
		slot = &e.iNext
	}
	return slot
}

// GenerateSPFDDictionary enumerates canonical-output-phase functions in increasing
// gate count. The bound replaces the C benchmark's eager multi-gigabyte calloc.
func GenerateSPFDDictionary(nVars, maxGates, maxFunctions int) ([]uint64, []int32, error) {
	if nVars < 1 || nVars > 6 || maxGates < 0 || maxFunctions < nVars {
		return nil, nil, fmt.Errorf("bdc: invalid dictionary bounds")
	}
	truths := append([]uint64(nil), spfdTruths[:nVars]...)
	weights := make([]int32, nVars)
	seen := map[uint64]bool{}
	for _, f := range truths {
		seen[f] = true
	}
	if len(truths) >= maxFunctions {
		return truths, weights, nil
	}
	stops := []int{0, nVars}
	for gates := 1; gates <= maxGates; gates++ {
		for k := 0; k < gates; k++ {
			m := gates - 1 - k
			if k > m || m+1 >= len(stops) {
				continue
			}
			for i := stops[k]; i < stops[k+1]; i++ {
				for j := stops[m]; j < stops[m+1]; j++ {
					if k == m && j <= i {
						continue
					}
					for kind := int32(0); kind < 5; kind++ {
						f := spfdGate(truths[i], truths[j], kind)
						if f&1 != 0 {
							f = ^f
						}
						if f == 0 || seen[f] {
							continue
						}
						seen[f] = true
						truths = append(truths, f)
						weights = append(weights, int32(gates))
						if len(truths) >= maxFunctions {
							return truths, weights, nil
						}
					}
				}
			}
		}
		stops = append(stops, len(truths))
	}
	return truths, weights, nil
}
func WriteSPFDDictionary(truthPath, weightPath string, truths []uint64, weights []int32) error {
	if len(truths) != len(weights) {
		return fmt.Errorf("bdc: dictionary lengths differ")
	}
	for _, item := range []struct {
		path string
		data any
	}{{truthPath, truths}, {weightPath, weights}} {
		file, err := os.Create(item.path)
		if err != nil {
			return err
		}
		err = binary.Write(file, binary.LittleEndian, item.data)
		closeErr := file.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
func ReadSPFDDictionary(truthPath, weightPath string) ([]uint64, []int32, error) {
	a, err := os.ReadFile(truthPath)
	if err != nil {
		return nil, nil, err
	}
	b, err := os.ReadFile(weightPath)
	if err != nil {
		return nil, nil, err
	}
	if len(a)%8 != 0 || len(b)%4 != 0 || len(a)/8 != len(b)/4 {
		return nil, nil, fmt.Errorf("bdc: invalid dictionary lengths")
	}
	truths := make([]uint64, len(a)/8)
	weights := make([]int32, len(b)/4)
	for i := range truths {
		truths[i] = binary.LittleEndian.Uint64(a[8*i:])
		weights[i] = int32(binary.LittleEndian.Uint32(b[4*i:]))
	}
	return truths, weights, nil
}
func Bdc_SpfdDecomposeTest__(out *[]int32) []uint64 {
	truths, weights, err := GenerateSPFDDictionary(6, 6, 249999999)
	if err != nil {
		panic(err)
	}
	*out = weights
	if len(truths) < 249999999 {
		if err := WriteSPFDDictionary("func6v6n_bin.txt", "func6v6nW_bin.txt", truths, weights); err != nil {
			panic(err)
		}
	}
	return truths
}
func spfdReadFiles(truthPath, weightPath string, out *[]int32) []uint64 {
	truths, weights, err := ReadSPFDDictionary(truthPath, weightPath)
	if err != nil {
		panic(err)
	}
	*out = weights
	return truths
}
func Bdc_SpfdReadFiles5(out *[]int32) []uint64 {
	return spfdReadFiles("func6v5n_bin.txt", "func6v5nW_bin.txt", out)
}
func Bdc_SpfdReadFiles6(out *[]int32) []uint64 {
	return spfdReadFiles("func6v6n_bin.txt", "func6v6nW_bin.txt", out)
}
func Bdc_SpfdComputeCost(f uint64, i int32, weights []int32) int32 {
	ones := Bdc_CountOnes(f)
	if ones == 0 {
		return -1
	}
	return 7*ones + 10*(8-weights[i])
}
func FindBestSPFD(divisors []uint64, weights []int32, f0, f1 uint64) (uint64, int32, error) {
	if len(divisors) != len(weights) {
		return 0, 0, fmt.Errorf("bdc: dictionary lengths differ")
	}
	bestCost, bestIndex := int32(-1), -1
	var best uint64
	for i, f := range divisors {
		for _, phase := range []uint64{f, ^f} {
			for _, side := range []struct{ forbidden, desired uint64 }{{f0, f1}, {f1, f0}} {
				if phase&side.forbidden != 0 {
					continue
				}
				cost := Bdc_SpfdComputeCost(phase&side.desired, int32(i), weights)
				if cost > bestCost {
					bestCost = cost
					best = phase
					bestIndex = i
				}
			}
		}
	}
	if bestIndex < 0 || bestCost <= 0 {
		return 0, 0, fmt.Errorf("bdc: no separating SPFD divisor")
	}
	return best, weights[bestIndex], nil
}
func Bdc_SpfdFindBest(divisors []uint64, weights []int32, f0, f1 uint64, cost *int32) uint64 {
	f, weight, err := FindBestSPFD(divisors, weights, f0, f1)
	if err != nil {
		panic(err)
	}
	*cost += weight
	return f
}
func DecomposeSPFDWithDivisors(truth uint64, divisors []uint64, weights []int32) (int32, error) {
	if truth == 0 || truth == ^uint64(0) {
		return 0, nil
	}
	f0, f1 := ^truth, truth
	var cost, steps int32
	for f0 != 0 && f1 != 0 {
		f, w, err := FindBestSPFD(divisors, weights, f0, f1)
		if err != nil {
			return 0, err
		}
		f0 &= ^f
		f1 &= ^f
		cost += w
		steps++
		if steps > 64 {
			return 0, fmt.Errorf("bdc: SPFD cover failed to make progress")
		}
	}
	return cost + steps - 1, nil
}
func Bdc_SpfdDecomposeTestOne(truth uint64, divisors []uint64, weights []int32) int32 {
	cost, err := DecomposeSPFDWithDivisors(truth, divisors, weights)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, "Produce solution with cost %d (with adj cost %d).\n", cost, Bdc_SpfdAdjCost(truth))
	return cost
}

// These two exploratory C entry points return before their experimental code.
// Keep their actual no-side-effect behavior; usable searches are exported above.
func Bdc_SpfdDecomposeTest_()  {}
func Bdc_SpfdDecomposeTest44() {}
func SPFDSubsetBenchmark(out io.Writer, large, small int, seed int64) error {
	if large < 0 || small < 0 {
		return fmt.Errorf("bdc: negative benchmark size")
	}
	rng := rand.New(rand.NewSource(seed))
	a, b := make([]uint64, large), make([]uint64, small)
	for i := range a {
		a[i] = rng.Uint64()
	}
	for i := range b {
		b[i] = rng.Uint64()
	}
	start := time.Now()
	var count int
	for _, x := range a {
		for _, y := range b {
			if x&y == y {
				count++
			}
		}
	}
	if _, err := fmt.Fprintf(out, "Total = %d. Time = %s\n", count, time.Since(start)); err != nil {
		return err
	}
	start = time.Now()
	count = 0
	for _, y := range b {
		for _, x := range a {
			if x&y == y {
				count++
			}
		}
	}
	_, err := fmt.Fprintf(out, "Total = %d. Time = %s\n", count, time.Since(start))
	return err
}
func Bdc_SpfdDecomposeTest3() {
	if err := SPFDSubsetBenchmark(os.Stdout, 1<<26, 1<<3, 1); err != nil {
		panic(err)
	}
}
func Bdc_SpfdDecomposeTest() { Bdc_SpfdDecomposeTest3() }
func Bdc_SpfdDecomposeTest8() {
	var weights []int32
	divisors := Bdc_SpfdReadFiles5(&weights)
	truth := uint64(0x513b57150819050f)
	bestCost := int32(1 << 30)
	var best uint64
	for _, f := range divisors {
		if cost := Bdc_SpfdAdjCost(truth ^ f); cost < bestCost {
			bestCost = cost
			best = f
		}
	}
	fmt.Fprintf(os.Stdout, "Best cost = %d. Function = %016X Difference = %016X\n", bestCost, best, best^truth)
}

// Keep a compact textual record useful to command-level clients.
func SPFDExpressionLevels(r *SPFDResult) string {
	var out strings.Builder
	for level, nodes := range r.Levels {
		fmt.Fprintf(&out, "level %d: %d candidates\n", level, len(nodes))
	}
	out.WriteString(r.Expression())
	return out.String()
}
