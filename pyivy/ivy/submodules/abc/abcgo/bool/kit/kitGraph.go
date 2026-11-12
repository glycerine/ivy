package kit

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func Kit_TruthToGraph(truth []uint32, nVars int32, memory *[]uint32) *Kit_Graph_t {
	if memory == nil {
		var local []uint32
		memory = &local
	}
	phase := Kit_TruthIsop(truth, nVars, memory, 1)
	if phase < 0 || len(*memory) > 1<<16 {
		return nil
	}
	return Kit_SopFactor(*memory, phase, nVars, memory)
}
func Kit_TruthToGraph2(off, on []uint32, nVars int32, memory *[]uint32) *Kit_Graph_t {
	if memory == nil {
		var local []uint32
		memory = &local
	}
	phase := Kit_TruthIsop2(off, on, nVars, memory, 1, 0)
	if phase < 0 || len(*memory) > 1<<16 {
		return nil
	}
	return Kit_SopFactor(*memory, phase, nVars, memory)
}
func Kit_TruthStats(truth []uint32, nVars int32, memory *[]uint32) int32 {
	g := Kit_TruthToGraph(truth, nVars, memory)
	if g == nil {
		return -1
	}
	return Kit_GraphRootLevel(g)<<16 | Kit_GraphNodeNum(g)
}
func Kit_TruthStatsArray(array []uint32, nVars, nFuncs int32) []int32 {
	words := Kit_TruthWordNum(nVars)
	if nFuncs < 0 || int64(words)*int64(nFuncs) > int64(len(array)) {
		panic("kit: short truth-table array")
	}
	out := make([]int32, nFuncs)
	var memory []uint32
	for i := int32(0); i < nFuncs; i++ {
		out[i] = Kit_TruthStats(array[i*words:(i+1)*words], nVars, &memory)
	}
	return out
}
func Kit_TruthFindVarNum(name string) int32 {
	name = filepath.Base(name)
	for i, c := range []byte(name) {
		if c < '0' || c > '9' {
			continue
		}
		end := i
		for end < len(name) && name[end] >= '0' && name[end] <= '9' {
			end++
		}
		n, err := strconv.ParseInt(name[i:end], 10, 32)
		if err != nil {
			return -1
		}
		return int32(n)
	}
	return -1
}

// Kit_TruthTest reads the binary little-endian uint32 tables used by ABC's
// profiling utility. Return I/O and malformed-input errors to the caller.
func Kit_TruthTest(name string) ([]int32, error) {
	started := time.Now()
	nVars := Kit_TruthFindVarNum(name)
	if nVars < 0 || nVars > 15 {
		return nil, fmt.Errorf("kit: no supported variable count in %q", name)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	words := Kit_TruthWordNum(nVars)
	if len(data)%int(4*words) != 0 {
		return nil, fmt.Errorf("kit: truncated binary truth table in %q", name)
	}
	truth := make([]uint32, len(data)/4)
	for i := range truth {
		truth[i] = binary.LittleEndian.Uint32(data[4*i:])
	}
	count := int32(len(truth)) / words
	result := Kit_TruthStatsArray(truth, nVars, count)
	fmt.Printf("Finished processing %d functions with %d variables. Time = %.2f sec\n", count, nVars, time.Since(started).Seconds())
	for i := 0; i < min(5, len(result)); i++ {
		fmt.Printf("Function %3d : AND2 = %3d Lev = %3d\n", i, result[i]&0xffff, result[i]>>16)
	}
	return result, nil
}
func Kit_TruthLitNum(truth []uint32, nVars int32, memory *[]uint32) int32 {
	g := Kit_TruthToGraph(truth, nVars, memory)
	if g == nil {
		return -1
	}
	return 1 + Kit_GraphNodeNum(g)
}
