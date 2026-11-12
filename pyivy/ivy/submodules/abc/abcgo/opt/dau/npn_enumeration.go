package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Dau_InsertFunction(manager *Abc_TtHieMan_t, truth []uint64, nodes, inputs, baseVars, n int32, memory *vec.Vec_Mem_t, support *vec.Vec_Int_t, fronts int32, start time.Time) int32 {
	newVars := util.Abc_TtMinBase(truth, nil, n, inputs)
	perm := make([]byte, 16)
	Abc_TtCanonicizeWrap(Abc_TtCanonicizeAda, manager, truth, newVars, perm, 99)
	before := vec.Vec_MemEntryNum(memory)
	vec.Vec_MemHashInsert(memory, truth)
	if before == vec.Vec_MemEntryNum(memory) {
		return 0
	}
	vec.Vec_IntPush(support, nodes<<16|newVars)
	if vec.Vec_IntSize(support)%1000000 == 0 {
		Dau_TablesSave(inputs, baseVars, memory, support, fronts, start)
	}
	return 1
}

// FunctionEnumeration exposes the real enumeration without forcing file output.
type FunctionEnumeration struct {
	Truths      []uint64
	NodeSupport []int32
	Frontiers   []int
	Steps       uint64
}

func EnumerateFunctions(inputs, n, maxNodes int32, useTwo, reduce bool) (*FunctionEnumeration, error) {
	if n < 3 || n > inputs || inputs > 6 || maxNodes < 0 {
		return nil, fmt.Errorf("dau: enumeration requires3<=vars<=inputs<=6 and nonnegative nodes")
	}
	manager := Abc_TtHieManStart(inputs, 5)
	defer Abc_TtHieManStop(manager)
	memory := vec.Vec_MemAlloc(1, 16)
	vec.Vec_MemHashAlloc(memory, 1<<16)
	support := vec.Vec_IntAlloc(1 << 16)
	vec.Vec_MemHashInsert(memory, []uint64{0})
	vec.Vec_IntPush(support, 0)
	vec.Vec_MemHashInsert(memory, []uint64{^npnVariables[0]})
	vec.Vec_IntPush(support, 1)
	limits := make([]int, maxNodes+2)
	limits[0] = 1
	limits[1] = 2
	result := &FunctionEnumeration{}
	start := time.Now()
	for nodes := int32(1); nodes <= maxNodes; nodes++ {
		expand := !(reduce && nodes == maxNodes)
		offer := func(t uint64, vars int32) {
			Dau_InsertFunction(manager, []uint64{t}, nodes, inputs, n, vars, memory, support, 0, start)
		}
		for entry := limits[nodes-1]; entry < limits[nodes]; entry++ {
			truth := vec.Vec_MemReadEntry(memory, entry)[0]
			vars := vec.Vec_IntEntry(support, entry) & 15
			result.Steps += npnOfferCrossbars(truth, vars, inputs, Dau_CountSymms2(truth, vars), expand, false, offer)
		}
		if useTwo && nodes > 2 && expand {
			for entry := limits[nodes-2]; entry < limits[nodes-1]; entry++ {
				truth := vec.Vec_MemReadEntry(memory, entry)[0]
				vars := vec.Vec_IntEntry(support, entry) & 15
				for v := int32(0); v < vars; v++ {
					for u := int32(0); u < vars; u++ {
						if u == v {
							continue
						}
						c0, c1 := util.Abc_Tt6Cofactor0(truth, v), util.Abc_Tt6Cofactor1(truth, v)
						c00, c01 := util.Abc_Tt6Cofactor0(c0, u), util.Abc_Tt6Cofactor1(c0, u)
						c10, c11 := util.Abc_Tt6Cofactor0(c1, u), util.Abc_Tt6Cofactor1(c1, u)
						gates := [5]uint64{npnVariables[v] & npnVariables[u], npnVariables[v] &^ npnVariables[u], ^npnVariables[v] & npnVariables[u], npnVariables[v] | npnVariables[u], npnVariables[v] ^ npnVariables[u]}
						for first := 0; first < 5; first++ {
							for second := first + 1; second < 5; second++ {
								a, b := gates[first]&c01|^gates[first]&c00, gates[first]&c11|^gates[first]&c10
								offer(gates[second]&b|^gates[second]&a, vars)
							}
						}
					}
				}
			}
		}
		limits[nodes+1] = vec.Vec_IntSize(support)
		if Dau_CountFuncs(support, int32(limits[nodes]), int32(limits[nodes+1]), n) == 0 {
			limits = limits[:nodes+2]
			break
		}
	}
	result.Frontiers = limits
	result.NodeSupport = append([]int32(nil), vec.Vec_IntArray(support)...)
	for i := 0; i < vec.Vec_MemEntryNum(memory); i++ {
		result.Truths = append(result.Truths, vec.Vec_MemReadEntry(memory, i)[0])
	}
	return result, nil
}
func Dau_FunctionEnum(inputs, n, maxNodes, useTwo, reduce, verbose int32) {
	result, err := EnumerateFunctions(inputs, n, maxNodes, useTwo != 0, reduce != 0)
	if err != nil {
		panic(err)
	}
	memory := vec.Vec_MemAlloc(1, 16)
	for _, truth := range result.Truths {
		vec.Vec_MemPush(memory, []uint64{truth})
	}
	support := vec.Vec_IntAllocArrayCopy(result.NodeSupport, len(result.NodeSupport))
	start := time.Now()
	for nodes := 0; nodes+1 < len(result.Frontiers); nodes++ {
		first := result.Frontiers[nodes]
		if nodes == 0 {
			first = 0
		}
		Dau_PrintStats(int32(nodes), inputs, n, support, int32(first), int32(result.Frontiers[nodes+1]), result.Steps, 0, start)
	}
	Dau_TablesSave(inputs, n, memory, support, int32(len(result.Truths)), start)
}
