package dau

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/bool/lucky"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"time"
)

var npnVariables = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func npnCheckVars(n int32) {
	if n < 0 || n > 6 {
		panic("dau: truth requires zero to six variables")
	}
}
func Dau_CountSymms(t uint64, n int32) int32 {
	npnCheckVars(n)
	count := int32(0)
	for i := int32(0); i < n; i++ {
		for j := i + 1; j < n; j++ {
			if util.Abc_Tt6SwapVars(t, i, j) == t {
				count++
			}
		}
	}
	return count
}
func Dau_CountSymms2(t uint64, n int32) int32 {
	npnCheckVars(n)
	mask := int32(0)
	for i := int32(0); i < n; i++ {
		for j := i + 1; j < n; j++ {
			if util.Abc_Tt6SwapVars(t, i, j) == t {
				mask |= 1 << j
			}
		}
	}
	return mask
}
func Dau_CountCompl1(t uint64, v, n int32) int32 {
	npnCheckVars(n)
	flipped := util.Abc_Tt6Flip(t, v)
	if flipped == ^t {
		return 1
	}
	for k := int32(0); k < n; k++ {
		if k != v && flipped == util.Abc_Tt6Flip(t, k) {
			return 1
		}
	}
	return 0
}
func Dau_CountCompl(t uint64, n int32) int32 {
	count := int32(0)
	for v := int32(0); v < n; v++ {
		count += Dau_CountCompl1(t, v, n)
	}
	return count
}

// EnumerateNPN enumerates source representatives in increasing normalized truth
// order. The returned table indexes functions whose highest minterm is zero.
func EnumerateNPN(n int32) (table, representatives []uint32, err error) {
	if n < 1 || n > 5 {
		return nil, nil, fmt.Errorf("dau: table enumeration requires 1..5 variables")
	}
	minterms := 1 << n
	count := uint64(1) << uint(minterms-1)
	table = make([]uint32, int(count))
	seen := make([]uint64, int((count+63)/64))
	mask := uint64(1)<<uint(minterms) - 1
	perm := make([]int, n)
	used := make([]bool, n)
	for current := uint64(0); current < count; current++ {
		if seen[current/64]>>(current%64)&1 != 0 {
			continue
		}
		representatives = append(representatives, uint32(current))
		var visit func(int)
		visit = func(level int) {
			if level < int(n) {
				for v := 0; v < int(n); v++ {
					if !used[v] {
						used[v] = true
						perm[level] = v
						visit(level + 1)
						used[v] = false
					}
				}
				return
			}
			for phase := 0; phase < minterms; phase++ {
				var transformed uint64
				for assignment := 0; assignment < minterms; assignment++ {
					original := 0
					for v := 0; v < int(n); v++ {
						original |= ((assignment >> v & 1) ^ (phase >> v & 1)) << perm[v]
					}
					transformed |= (current >> uint(original) & 1) << uint(assignment)
				}
				if transformed >= count {
					transformed = mask &^ transformed
				}
				seen[transformed/64] |= 1 << (transformed % 64)
				table[transformed] = uint32(current)
			}
		}
		visit(0)
	}
	return table, representatives, nil
}
func Dau_TruthEnum(n int32) {
	table, reps, err := EnumerateNPN(n)
	if err != nil {
		panic(err)
	}
	log := (1 << n) - 2
	name := fmt.Sprintf("tableW%d.data", log)
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	err = binary.Write(f, binary.LittleEndian, table)
	closeErr := f.Close()
	if err != nil {
		panic(err)
	}
	if closeErr != nil {
		panic(closeErr)
	}
	fmt.Printf("Computed %d NPN classes of %d variables.\n", len(reps), n)
}
func Dau_ReadFile(name string, sizeWords int32) []uint32 {
	if sizeWords < 0 {
		panic("dau: negative table size")
	}
	table := make([]uint32, int(sizeWords)*2)
	f, err := os.Open(name)
	if err != nil {
		fmt.Printf("Cannot open input file %q.\n", name)
		return table
	}
	defer f.Close()
	data := make([]byte, len(table)*4)
	_, err = io.ReadFull(f, data)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		panic(err)
	}
	for i := range table {
		table[i] = binary.LittleEndian.Uint32(data[4*i:])
	}
	fmt.Printf("Finished reading file %q.\n", name)
	return table
}
func Dau_AddFunction(t uint64, n int32, table []uint32, classes, partial *vec.Vec_Int_t) int32 {
	mask := uint64(1)<<uint(1<<n) - 1
	if t>>uint((1<<n)-1)&1 != 0 {
		t = ^t
	}
	index := t & mask
	rep := table[index] & 0x7fffffff
	entry := table[rep]
	if entry>>31 != 0 {
		return 0
	}
	vec.Vec_IntPush(classes, int32(entry))
	truth := []uint64{t}
	if util.Abc_TtSupportSize(truth, n) < n {
		vec.Vec_IntPush(partial, int32(entry))
	}
	table[rep] = entry | 0x80000000
	return int32(entry)
}

func Dau_ExactNpnForClasses(memory *vec.Vec_Mem_t, support *vec.Vec_Int_t, n, nInputs int32) *vec.Vec_Wrd_t {
	if n > 6 {
		panic("dau: exact classes require at most six variables")
	}
	result := vec.Vec_WrdStart(vec.Vec_IntSize(support))
	info := lucky.NewPermInfo(n)
	aux, minimal := make([]uint64, 1024), make([]uint64, 1024)
	for i, entry := range vec.Vec_IntArray(support) {
		if entry&15 > n {
			continue
		}
		truth := []uint64{vec.Vec_MemReadEntry(memory, i)[0]}
		lucky.SimpleMinimal(truth, aux, minimal, info, n)
		vec.Vec_WrdWriteEntry(result, i, truth[0])
	}
	return result
}
func Dau_ExactNpnPrint(memory *vec.Vec_Mem_t, support *vec.Vec_Int_t, n, nInputs, maxNodes int32) {
	canons := vec.Vec_WrdArray(Dau_ExactNpnForClasses(memory, support, n, nInputs))
	seen := map[uint64]bool{}
	total := 0
	fmt.Println("Final results:")
	for nodes := int32(0); nodes <= maxNodes; nodes++ {
		semi, newClasses := 0, 0
		for i, e := range vec.Vec_IntArray(support) {
			if e&15 > n || e>>16 != nodes {
				continue
			}
			semi++
			if !seen[canons[i]] {
				seen[canons[i]] = true
				newClasses++
			}
		}
		total += newClasses
		fmt.Printf("Nodes = %2d.  Semi-canonical = %8d.  Canonical = %8d.  Total = %8d.\n", nodes, semi, newClasses, total)
	}
}
func Dau_TablesSave(nInputs, n int32, memory *vec.Vec_Mem_t, support *vec.Vec_Int_t, fronts int32, start time.Time) {
	name := fmt.Sprintf("npn%d%d.ttd", nInputs, n)
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	for i := 0; i < vec.Vec_MemEntryNum(memory); i++ {
		if err = binary.Write(f, binary.LittleEndian, vec.Vec_MemReadEntry(memory, i)[:1<<max(0, nInputs-6)]); err != nil {
			f.Close()
			panic(err)
		}
	}
	err = binary.Write(f, binary.LittleEndian, vec.Vec_IntArray(support))
	closeErr := f.Close()
	if err != nil {
		panic(err)
	}
	if closeErr != nil {
		panic(closeErr)
	}
	fmt.Printf("Dumped file %q with %10d classes after exploring %10d frontiers. Time = %.2f sec\n", name, vec.Vec_IntSize(support), fronts, time.Since(start).Seconds())
}
func Dau_DumpFuncs(memory *vec.Vec_Mem_t, support *vec.Vec_Int_t, n, maxNodes int32) {
	if n != 4 && n != 5 || maxNodes < 0 {
		panic("dau: dump expects four or five variables")
	}
	files := make([]*os.File, maxNodes+1)
	defer func() {
		for _, f := range files {
			if f != nil {
				f.Close()
			}
		}
	}()
	for nodes := range files {
		var err error
		files[nodes], err = os.Create(fmt.Sprintf("func%d_min%d.tt", n, nodes))
		if err != nil {
			panic(err)
		}
	}
	for i, e := range vec.Vec_IntArray(support) {
		if e&15 != n {
			continue
		}
		nodes := e >> 16
		if nodes > maxNodes {
			panic("dau: dump node count exceeds file range")
		}
		truth := vec.Vec_MemReadEntry(memory, i)[0]
		if n == 4 {
			if _, err := fmt.Fprintf(files[nodes], "%04x\n", uint16(truth)); err != nil {
				panic(err)
			}
		} else {
			if _, err := fmt.Fprintf(files[nodes], "%08x\n", uint32(truth)); err != nil {
				panic(err)
			}
		}
	}
}
func Dau_CountFuncs(support *vec.Vec_Int_t, first, last, n int32) int32 {
	count := int32(0)
	for _, e := range vec.Vec_IntArray(support)[first:last] {
		if e&15 <= n {
			count++
		}
	}
	return count
}
func Dau_PrintStats(nodes, nInputs, n int32, support *vec.Vec_Int_t, first, last int32, steps uint64, countTwo int32, start time.Time) int32 {
	fresh := Dau_CountFuncs(support, first, last, n)
	fmt.Printf("N =%2d | C =%12d New%d =%10d All%d =%10d | New%d =%8d All%d =%8d Two =%6d %.2f sec\n", nodes, steps, nInputs, last-first, nInputs, last, n, fresh, n, Dau_CountFuncs(support, 0, last, n), countTwo, time.Since(start).Seconds())
	return fresh
}

// npnOfferCrossbars preserves source candidate order, including both phases of
// AND gates and the two possible replacements of a double-crossbar XOR.
func npnOfferCrossbars(truth uint64, support, inputs, symmetric int32, expand bool, reverse bool, offer func(uint64, int32)) uint64 {
	var steps uint64
	variable := func(v int32) uint64 {
		if reverse {
			return npnVariables[inputs-1-v]
		}
		return npnVariables[v]
	}
	for v := int32(0); v < support; v++ {
		if symmetric>>v&1 != 0 {
			continue
		}
		index := v
		if reverse {
			index = inputs - 1 - v
		}
		c0, c1 := util.Abc_Tt6Cofactor0(truth, index), util.Abc_Tt6Cofactor1(truth, index)
		gate := func(g uint64, both bool, newSupport int32) {
			offer(g&c1|^g&c0, newSupport)
			steps++
			if both {
				offer(g&c0|^g&c1, newSupport)
				steps++
			}
		}
		if expand && support < inputs {
			gate(variable(v)&variable(support), true, support+1)
			gate(variable(v)^variable(support), false, support+1)
		}
		if expand {
			for k := int32(0); k < support; k++ {
				if k == v || symmetric>>k&1 != 0 && k != v+1 {
					continue
				}
				gate(variable(v)&variable(k), true, support)
				gate(variable(v)&^variable(k), true, support)
				gate(variable(v)^variable(k), false, support)
			}
		}
		for k := int32(0); k < support; k++ {
			if k == v {
				continue
			}
			for j := k + 1; j < support; j++ {
				if j == v {
					continue
				}
				gate(variable(j)&variable(k), true, support)
				gate(variable(j)&^variable(k), true, support)
				gate(^variable(j)&variable(k), true, support)
				gate(^variable(j)&^variable(k), true, support)
				gate(variable(j)^variable(k), !reverse, support)
			}
		}
	}
	return steps
}
func EnumerateNetworks(n int32, table []uint32) ([][]int32, [][]int32) {
	if n < 2 || n > 5 {
		panic("dau: network table enumeration requires2..5variables")
	}
	levels, partial := make([]*vec.Vec_Int_t, 32), make([]*vec.Vec_Int_t, 32)
	for i := range levels {
		levels[i] = vec.Vec_IntAlloc(16)
		partial[i] = vec.Vec_IntAlloc(16)
	}
	inv := uint32(uint64(1)<<uint(1<<(n-1)) - 1)
	table[0] |= 0x80000000
	table[inv] |= 0x80000000
	vec.Vec_IntPushTwo(levels[0], 0, int32(inv))
	vec.Vec_IntPushTwo(partial[0], 0, int32(inv))
	for nodes := 1; nodes < 32; nodes++ {
		for _, entry := range vec.Vec_IntArray(levels[nodes-1]) {
			truth := uint64(uint32(entry)) * 0x100000001
			support := util.Abc_TtSupportSize([]uint64{truth}, n)
			npnOfferCrossbars(truth, support, n, 0, true, true, func(t uint64, _ int32) { Dau_AddFunction(t, n, table, levels[nodes], partial[nodes]) })
		}
		if vec.Vec_IntSize(levels[nodes]) == 0 {
			break
		}
	}
	a, b := make([][]int32, 0), make([][]int32, 0)
	for i := range levels {
		a = append(a, append([]int32(nil), vec.Vec_IntArray(levels[i])...))
		b = append(b, append([]int32(nil), vec.Vec_IntArray(partial[i])...))
		if i > 0 && len(a[i]) == 0 {
			break
		}
	}
	return a, b
}
func Dau_NetworkEnum(n int32) {
	if n < 2 || n > 5 {
		panic("dau: network enumeration requires2..5variables")
	}
	sizeLog := (1 << n) - 2
	table := Dau_ReadFile(fmt.Sprintf("tableW%d.data", sizeLog), 1<<sizeLog)
	levels, partial := EnumerateNetworks(n, table)
	for nodes := range levels {
		fmt.Printf("Nodes = %2d. New = %6d. Partial = %6d.\n", nodes, len(levels[nodes]), len(partial[nodes]))
	}
}
func Dau_NetworkEnumTest() { Dau_NetworkEnum(4) }
