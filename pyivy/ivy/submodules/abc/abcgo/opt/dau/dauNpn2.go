package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"sort"
	"time"
)

// Dtt_Man_t retains node/phase implementations and optional multiplicities.
// Go maps replace the source's enormous dense marks and temporary hash bins.
type dttRound struct {
	Steps, Multis uint64
	New, Total    int
}
type Dtt_Man_t struct {
	rounds            []dttRound
	nVars             int32
	perms, comps      []int32
	marks             map[uint32]bool
	fanins            [][2]int32
	truths, truthNpns []uint32
	configs, classes  []int32
	funNodes          [][]int32
	temp, temp2       *vec.Vec_Int_t
	FunMask, CmpMask  uint32
	Counts            [32]int32
	nClasses          int32
	table             []uint32
	nodes, times      []int32
	visited           []bool
	visitedClasses    []int32
	hash              map[uint32]bool
}

func npnPermSchedule(n int32) []int32 {
	if n < 1 {
		panic("dau: permutation schedule requires variables")
	}
	if n == 1 {
		return []int32{0}
	}
	if n == 2 {
		return []int32{0, 0}
	}
	factorial := 1
	for i := int32(2); i <= n; i++ {
		factorial *= int(i)
	}
	previous := npnPermSchedule(n - 1)
	out := make([]int32, 0, factorial)
	for k := 0; k < factorial/int(n)/2; k++ {
		for i := n - 1; i > 0; i-- {
			out = append(out, i-1)
		}
		out = append(out, previous[2*k]+1)
		for i := int32(0); i < n-1; i++ {
			out = append(out, i)
		}
		out = append(out, previous[2*k+1])
	}
	return out
}
func npnCompSchedule(n int32) []int32 {
	out := make([]int32, 0, 1<<n)
	for k := int32(0); k < n; k++ {
		out = append(out, k)
		for i := 1; i < 1<<k; i++ {
			out = append(out, out[i-1])
		}
	}
	return append(out, n-1)
}
func Dau_ReadFile2(name string, size int32) []uint32 {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return nil
	}
	return Dau_ReadFile(name, size)
}
func Dtt_ManRenum(n int32, table []uint32, classes *int32) {
	count := uint32(0)
	limit := uint64(1) << uint((1<<n)-1)
	for i := uint64(0); i < limit; i++ {
		entry := table[i]
		if uint64(entry) == i {
			table[i] = count
			count++
		} else {
			if uint64(entry) >= i {
				panic("dau: nonminimal NPN class table")
			}
			table[i] = table[entry]
		}
	}
	*classes = int32(count)
	fmt.Printf("The total number of NPN classes = %d.\n", count)
}
func Dtt_ManLoadClasses(n int32, classes *int32) []uint32 {
	log := (1 << n) - 2
	name := fmt.Sprintf("tableW%d.data", log)
	table := Dau_ReadFile2(name, 1<<log)
	if table == nil {
		Dau_TruthEnum(n)
		table = Dau_ReadFile2(name, 1<<log)
	}
	Dtt_ManRenum(n, table, classes)
	return table
}
func Dtt_ManAlloc(n, multi int32) *Dtt_Man_t {
	if n < 3 || n > 5 {
		panic("dau: Dtt requires3..5variables")
	}
	p := &Dtt_Man_t{nVars: n, perms: npnPermSchedule(n), comps: npnCompSchedule(n), marks: map[uint32]bool{}, hash: map[uint32]bool{}, temp: vec.Vec_IntAlloc(4000), temp2: vec.Vec_IntAlloc(4000), funNodes: make([][]int32, 16), FunMask: ^uint32(0), CmpMask: 1 << 31}
	if n < 5 {
		p.FunMask = uint32(1)<<uint(1<<n) - 1
		p.CmpMask = uint32(1) << uint((1<<n)-1)
	}
	if multi != 0 {
		p.table = Dtt_ManLoadClasses(n, &p.nClasses)
		p.nodes = make([]int32, p.nClasses)
		p.times = make([]int32, p.nClasses)
		p.visited = make([]bool, p.nClasses)
	}
	return p
}
func Dtt_ManFree(p *Dtt_Man_t) {
	if p != nil {
		*p = Dtt_Man_t{}
	}
}
func (p *Dtt_Man_t) normalize(truth uint32) uint32 {
	if truth&p.CmpMask != 0 {
		truth = ^truth
	}
	return truth & p.FunMask
}
func Dtt_ManAddVisited(p *Dtt_Man_t, truth uint32, n int32) {
	class := p.table[p.normalize(truth)]
	if p.nodes[class] < n {
		return
	}
	if p.nodes[class] != n {
		panic("dau: visited class at wrong level")
	}
	if !p.visited[class] {
		p.visited[class] = true
		p.visitedClasses = append(p.visitedClasses, int32(class))
	}
}
func Dtt_ManProcessVisited(p *Dtt_Man_t) {
	for _, class := range p.visitedClasses {
		if !p.visited[class] {
			panic("dau: missing visited class")
		}
		p.visited[class] = false
		p.times[class]++
	}
	p.visitedClasses = p.visitedClasses[:0]
}
func Dtt_ManHashKey(_ *Dtt_Man_t, truth uint32) uint32 {
	return (truth&255)*1699 + (truth>>8&255)*5147 + (truth>>16&255)*7103 + (truth>>24)*8147
}
func Dtt_ManCheckHash(p *Dtt_Man_t, truth uint32) int32 {
	if p.hash[truth] {
		return 0
	}
	p.hash[truth] = true
	return 1
}
func Dtt_ManCollect(p *Dtt_Man_t, truth uint32, functions *vec.Vec_Int_t) *vec.Vec_Int_t {
	vec.Vec_IntClear(functions)
	t := uint64(truth) * 0x100000001
	for _, perm := range p.perms {
		for _, comp := range p.comps {
			normalized := t
			if t&uint64(p.CmpMask) != 0 {
				normalized = ^t
			}
			if Dtt_ManCheckHash(p, uint32(normalized)) != 0 {
				vec.Vec_IntPush(functions, int32(normalized))
			}
			t = util.Abc_Tt6Flip(t, comp)
		}
		t = util.Abc_Tt6SwapAdjacent(t, perm)
	}
	if t != uint64(truth)*0x100000001 {
		panic("dau: incomplete transformation schedule")
	}
	clear(p.hash)
	return functions
}
func Dtt_ManGetFun(p *Dtt_Man_t, truth uint32, n int32) int32 {
	index := p.normalize(truth)
	if !p.marks[index] {
		return 0
	}
	if p.table == nil {
		return 1
	}
	class := p.table[index]
	if p.nodes[class] < n {
		return 1
	}
	if p.nodes[class] != n {
		panic("dau: class found at future level")
	}
	return dauBool(p.visited[class])
}
func Dtt_ManSetFun(p *Dtt_Man_t, truth uint32) { p.marks[p.normalize(truth)] = true }
func Dtt_ManAddFunction(p *Dtt_Man_t, n, faninI, faninJ, kind int32, truth uint32) {
	functions := vec.Vec_IntArray(Dtt_ManCollect(p, truth, p.temp2))
	minimum := ^uint32(0)
	for _, f := range functions {
		minimum = min(minimum, uint32(f))
	}
	id := int32(len(p.truths))
	nodes := (p.configs[faninI] >> 3 & 15) + (p.configs[faninJ] >> 3 & 15) + 1
	if nodes >= 32 {
		panic("dau: implementation exceeds source node-count range")
	}
	for int(n) >= len(p.funNodes) {
		p.funNodes = append(p.funNodes, nil)
	}
	p.funNodes[n] = append(p.funNodes[n], id)
	p.fanins = append(p.fanins, [2]int32{faninI, faninJ})
	p.truths = append(p.truths, truth)
	p.configs = append(p.configs, nodes<<3|kind)
	p.classes = append(p.classes, int32(len(p.truthNpns)))
	p.truthNpns = append(p.truthNpns, minimum)
	for _, f := range functions {
		Dtt_ManSetFun(p, uint32(f))
	}
	p.Counts[nodes]++
	if p.table != nil {
		p.nodes[p.table[p.normalize(truth)]] = n
	}
}
func Dtt_PrintStats(nodes, n int32, levels [][]int32, steps uint64, start time.Time, delay int32, multis uint64) int32 {
	fresh := len(levels[nodes])
	total := 1
	for _, level := range levels {
		total += len(level)
	}
	kind := 'N'
	if delay != 0 {
		kind = 'D'
	}
	extra := 0
	if nodes == 0 {
		extra = 1
	}
	fmt.Printf("%c =%2d | C =%12d | New%d =%10d All%d =%10d | Multi =%10d | Time = %.2f sec\n", kind, nodes, steps, n, fresh+extra, n, total, multis, time.Since(start).Seconds())
	return int32(fresh)
}
func Dtt_PrintDistrib(p *Dtt_Man_t) {
	fmt.Println("NPN classes for each node count (N):")
	for n, count := range p.Counts {
		if count != 0 {
			fmt.Printf("N = %2d : NPN = %6d\n", n, count)
		}
	}
}
func Dtt_PrintMulti2(p *Dtt_Man_t) {
	for n := int32(0); n <= 7; n++ {
		fmt.Printf("n=%d : ", n)
		for i, count := range p.nodes {
			if count == n {
				fmt.Printf("%d ", p.times[i])
			}
		}
		fmt.Println()
	}
}
func Dtt_PrintMulti1(p *Dtt_Man_t) {
	for n := int32(0); n < 16; n++ {
		var times []int32
		for i, count := range p.nodes {
			if count == n {
				times = append(times, p.times[i])
			}
		}
		if len(times) == 0 {
			break
		}
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
		fmt.Printf("n=%d : ", n)
		for i := 0; i < len(times); {
			j := i + 1
			for j < len(times) && times[j] == times[i] {
				j++
			}
			fmt.Printf("%d=%d ", times[i], j-i)
			i = j
		}
		fmt.Println()
	}
}
func Dtt_PrintMulti(p *Dtt_Man_t) {
	for n := int32(0); n < 13; n++ {
		var bins [15]int
		total, count := int32(0), int32(0)
		for i, level := range p.nodes {
			if level == n {
				value := p.times[i]
				log := 0
				for x := value - 1; x > 0; x >>= 1 {
					log++
				}
				if log >= len(bins) {
					panic("dau: multiplicity exceeds source range")
				}
				bins[log]++
				total += value
				count++
			}
		}
		if count == 0 {
			break
		}
		fmt.Printf("n=%2d : All = %7d Ave = %6.2f ", n, count, float64(total)/float64(count))
		for _, count := range bins {
			if count == 0 {
				fmt.Printf("%6s", "")
			} else {
				fmt.Printf("%6d", count)
			}
		}
		fmt.Println()
	}
}

func Dtt_FindNP(p *Dtt_Man_t, fun, goal, npn uint32, np, kind *int32, npOut int32) {
	permutation := [5]int32{0, 1, 2, 3, 4}
	var negative [5]int32
	t := uint64(npn) * 0x100000001
	for _, perm := range p.perms {
		for _, comp := range p.comps {
			if Dtt_Check(fun, goal, uint32(t), kind) != 0 {
				if goal == 0 && fun == ^uint32(t) {
					if npOut == 0 {
						Dtt_ProcessType(kind, 3)
					} else {
						Dtt_ProcessType(kind, 2)
					}
				}
				*np = 0
				for j := 0; j < 5; j++ {
					if npOut == 0 {
						*np |= (negative[j]<<3 | permutation[j]) << uint(4*j)
					} else {
						index := npOut >> uint(4*j) & 7
						*np |= (permutation[index] | (negative[index]^(npOut>>uint(4*j+3)&1))<<3) << uint(4*j)
					}
				}
				return
			}
			t = util.Abc_Tt6Flip(t, comp)
			negative[comp] ^= 1
		}
		t = util.Abc_Tt6SwapAdjacent(t, perm)
		permutation[perm], permutation[perm+1] = permutation[perm+1], permutation[perm]
	}
	panic("dau: NP transform not found")
}
func Dtt_DumpLibrary(p *Dtt_Man_t, name string) {
	functions := append([]uint32(nil), p.truthNpns...)
	sort.Slice(functions, func(i, j int) bool { return int32(functions[i]) < int32(functions[j]) })
	unique := functions[:0]
	for _, f := range functions {
		if len(unique) == 0 || unique[len(unique)-1] != f {
			unique = append(unique, f)
		}
	}
	functions = unique
	index := map[uint32]int32{}
	for i, f := range functions {
		index[f] = int32(i)
	}
	library := make([][]*Dtt_FunImpl_t, len(functions))
	for i, truth := range p.truths {
		if i < 2 {
			continue
		}
		fanin1, fanin2 := p.fanins[i][0], p.fanins[i][1]
		impl := &Dtt_FunImpl_t{Type: p.configs[i] & 7, FI1: index[p.truthNpns[fanin1]], FI2: index[p.truthNpns[fanin2]]}
		target := index[p.truthNpns[i]]
		repeated := false
		for _, old := range library[target] {
			if old.FI1 == impl.FI1 && old.FI2 == impl.FI2 || old.FI2 == impl.FI1 && old.FI1 == impl.FI2 {
				repeated = true
				break
			}
		}
		if repeated {
			continue
		}
		var output int32
		Dtt_FindNP(p, p.truthNpns[i], 0, truth, &output, &impl.Type, 0)
		Dtt_FindNP(p, p.truths[fanin2], truth, p.truthNpns[fanin1], &impl.NP1, &impl.Type, output)
		Dtt_FindNP(p, p.truths[fanin2], 0, p.truthNpns[fanin2], &impl.NP2, &impl.Type, output)
		library[target] = append(library[target], impl)
	}
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for i, truth := range functions {
		if i < 2 {
			continue
		}
		for _, impl := range library[i] {
			Dtt_MakeFormula(truth, impl, library, 0x43210, true, f)
		}
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("Dumped file %q.\n", name)
}

// EnumerateLF implements L(f)/D(f) enumeration. Multiplicity-enabled managers
// retain per-class visit counts; the returned manager owns all result arrays.
func EnumerateLF(n, maxNodes int32, delay, multi bool) *Dtt_Man_t {
	return enumerateLF(Dtt_ManAlloc(n, dauBool(multi)), maxNodes, delay)
}
func enumerateLF(p *Dtt_Man_t, maxNodes int32, delay bool) *Dtt_Man_t {
	n := p.nVars
	p.fanins = append(p.fanins, [2]int32{}, [2]int32{})
	p.truths = append(p.truths, 0, uint32(npnVariables[0]))
	p.configs = append(p.configs, 0, 0)
	p.classes = append(p.classes, 0, 1)
	p.truthNpns = append(p.truthNpns, 0, uint32(npnVariables[0]))
	Dtt_ManSetFun(p, 0)
	p.funNodes[0] = append(p.funNodes[0], 1)
	for i := int32(0); i < n; i++ {
		Dtt_ManSetFun(p, uint32(npnVariables[i]))
	}
	p.Counts[0] = 2
	p.rounds = append(p.rounds, dttRound{New: 2, Total: 2})
	var steps, multis uint64
	for level := int32(1); level <= maxNodes; level++ {
		for int(level) >= len(p.funNodes) {
			p.funNodes = append(p.funNodes, nil)
		}
		for i, j := int32(0), level-1; i < level; i, j = i+1, j-1+dauBool(delay) {
			if i > j {
				continue
			}
			for firstIndex, first := range p.funNodes[i] {
				functions := vec.Vec_IntArray(Dtt_ManCollect(p, p.truths[first], p.temp))
				start := 0
				if i == j {
					start = firstIndex
				}
				for _, second := range p.funNodes[j][start:] {
					b := p.truths[second]
					for _, function := range functions {
						a := uint32(function)
						truths := [5]uint32{b & a, b &^ a, ^b & a, b | a, b ^ a}
						for kind, truth := range truths {
							if Dtt_ManGetFun(p, truth, level) == 0 {
								Dtt_ManAddFunction(p, level, first, second, int32(kind), truth)
							}
						}
						steps += 5
						if p.table != nil {
							for _, truth := range truths {
								Dtt_ManAddVisited(p, truth, level)
							}
						}
					}
					multis++
					if p.table != nil {
						Dtt_ManProcessVisited(p)
					}
				}
			}
		}
		p.rounds = append(p.rounds, dttRound{Steps: steps, Multis: multis, New: len(p.funNodes[level]), Total: len(p.truths)})
		if len(p.funNodes[level]) == 0 {
			break
		}
	}
	return p
}
func Dtt_EnumerateLf(n, maxNodes, delay, multi, verbose int32, name string) {
	start := time.Now()
	p := EnumerateLF(n, maxNodes, delay != 0, multi != 0)
	defer Dtt_ManFree(p)
	kind := 'N'
	if delay != 0 {
		kind = 'D'
	}
	for level, round := range p.rounds {
		fmt.Printf("%c =%2d | C =%12d | New%d =%10d All%d =%10d | Multi =%10d | Time = %.2f sec\n", kind, level, round.Steps, n, round.New, n, round.Total, round.Multis, time.Since(start).Seconds())
	}
	if delay != 0 {
		Dtt_PrintDistrib(p)
	}
	if delay == 0 && name != "" {
		Dtt_DumpLibrary(p, name)
	}
}
