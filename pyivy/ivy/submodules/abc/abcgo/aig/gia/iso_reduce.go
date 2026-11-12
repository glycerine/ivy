package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
	"time"
)

// Gia_ManIsoReduce groups structurally isomorphic outputs and keeps one
// representative per class. Dual-output mode groups consecutive pairs.
func Gia_ManIsoReduce(initial *Gia_Man_t, outputGroups, inputPerms **Vec_Ptr_t, estimate, dual, verbose, veryVerbose int32) *Gia_Man_t {
	start := time.Now()
	if outputGroups != nil {
		*outputGroups = nil
	}
	if inputPerms != nil {
		*inputPerms = vec.Vec_PtrStart(int(Gia_ManPoNum(initial)))
	}
	graph := initial
	if dual != 0 {
		if Gia_ManPoNum(initial)&1 != 0 {
			panic("gia: dual output count must be even")
		}
		if Gia_ManPoNum(initial) == 2 {
			return Gia_ManDup(initial)
		}
		transformed := Gia_ManTransformMiter(initial)
		graph = Gia_ManSeqStructSweep(transformed, 1, 1, 0)
		Gia_ManStop(transformed)
		defer Gia_ManStop(graph)
	} else if Gia_ManPoNum(initial) == 1 {
		return Gia_ManDup(initial)
	}
	candidates := Gia_IsoDeriveEquivPos(graph, 1, veryVerbose)
	if candidates == nil {
		return nil
	}
	var used int32
	nontrivial := Vec_IntCountNonTrivial(candidates, &used)
	fmt.Printf("Reduced %d outputs to %d candidate classes (%d outputs are in %d non-trivial classes).  Time = %.2f sec\n", Gia_ManPoNum(graph), vec.Vec_PtrSize(candidates), used, nontrivial, time.Since(start).Seconds())
	if estimate != 0 {
		return Gia_ManDup(initial)
	}
	var groups [][]int32
	finished := 0
	for _, item := range vec.Vec_PtrArray(candidates) {
		candidate := vec.Vec_IntArray(item.(*Vec_Int_t))
		if len(candidate) < 2 {
			groups = append(groups, append([]int32(nil), candidate...))
			finished += len(candidate)
			continue
		}
		groupStart := len(groups)
		signatures := make(map[string]int)
		for _, output := range candidate {
			finished++
			var permutation *Vec_Int_t
			signature := Gia_ManIsoFindString(graph, output, 0, &permutation)
			if signature == nil {
				return nil
			}
			if inputPerms != nil {
				vec.Vec_PtrWriteEntry(*inputPerms, int(output), permutation)
			}
			key := string(vec.Vec_StrArray(signature))
			index, ok := signatures[key]
			if !ok {
				index = len(groups) - groupStart
				signatures[key] = index
				groups = append(groups, nil)
			}
			groups[groupStart+index] = append(groups[groupStart+index], output)
		}
	}
	if finished != int(Gia_ManPoNum(graph)) {
		panic("gia: output classes do not cover all outputs")
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i][0] < groups[j][0] })
	remain := make([]int32, 0, len(groups)*2)
	resultGroups := vec.Vec_PtrAlloc(len(groups))
	for _, group := range groups {
		vec.Vec_PtrPush(resultGroups, vec.Vec_IntAllocArrayCopy(group, len(group)))
		if dual != 0 {
			remain = append(remain, 2*group[0], 2*group[0]+1)
		} else {
			remain = append(remain, group[0])
		}
	}
	result := Gia_ManDupCones(initial, remain, int32(len(remain)), 0)
	nontrivial = Vec_IntCountNonTrivial(resultGroups, &used)
	if dual == 0 {
		fmt.Printf("Reduced %d outputs to %d equivalence classes (%d outputs are in %d non-trivial classes).  ", Gia_ManPoNum(initial), len(groups), used, nontrivial)
	} else {
		fmt.Printf("Reduced %d dual outputs to %d dual outputs.  ", Gia_ManPoNum(initial)/2, Gia_ManPoNum(result)/2)
	}
	fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
	if verbose != 0 {
		for _, group := range groups {
			if len(group) > 1 {
				fmt.Println(group)
			}
		}
	}
	if outputGroups != nil {
		*outputGroups = resultGroups
	}
	return result
}
