package mapper

import (
	"fmt"
	"sort"

	abcutil "github.com/glycerine/abcgo/misc/util"
)

func Map_SuperTableCreate(lib *Map_SuperLib_t) *Map_HashTable_t {
	n := abcutil.Abc_PrimeCudd(20000)
	return &Map_HashTableStruct_t_{nBins: n, pBins: make([]*Map_HashEntry_t, n), mmMan: lib.mmEntries}
}

func Map_SuperTableFree(table *Map_HashTable_t) {
	if table == nil {
		return
	}
	clear(table.pBins)
	table.pBins = nil
	table.nEntries = 0
}

func Map_SuperTableInsertC(table *Map_HashTable_t, truth []uint32, gate *Map_Super_t) int32 {
	key := tableKey(truth, table.nBins)
	for entry := table.pBins[key]; entry != nil; entry = entry.pNext {
		if entry.uTruth[0] == truth[0] && entry.uTruth[1] == truth[1] {
			gate.pNext, entry.pGates = entry.pGates, gate
			return 0
		}
	}
	entry := &Map_HashEntryStruct_t_{uTruth: [2]uint32{truth[0], truth[1]}, pGates: gate, pNext: table.pBins[key]}
	table.pBins[key] = entry
	table.nEntries++
	if table.nEntries >= 2*table.nBins {
		Map_SuperTableResize(table)
	}
	return 0
}

func Map_SuperTableInsert(table *Map_HashTable_t, truth []uint32, gate *Map_Super_t, phase uint32) int32 {
	key := tableKey(truth, table.nBins)
	for entry := table.pBins[key]; entry != nil; entry = entry.pNext {
		if entry.uTruth[0] == truth[0] && entry.uTruth[1] == truth[1] {
			return 1
		}
	}
	entry := &Map_HashEntryStruct_t_{uTruth: [2]uint32{truth[0], truth[1]}, uPhase: phase, pGates: gate, pNext: table.pBins[key]}
	table.pBins[key] = entry
	table.nEntries++
	if table.nEntries >= 2*table.nBins {
		Map_SuperTableResize(table)
	}
	return 0
}

func Map_SuperTableLookupC(lib *Map_SuperLib_t, truth []uint32) *Map_Super_t {
	return lookupSuper(lib.tTableC, truth, nil)
}

func Map_SuperTableLookup(table *Map_HashTable_t, truth []uint32, phase *uint32) *Map_Super_t {
	return lookupSuper(table, truth, phase)
}

func lookupSuper(table *Map_HashTable_t, truth []uint32, phase *uint32) *Map_Super_t {
	key := tableKey(truth, table.nBins)
	for entry := table.pBins[key]; entry != nil; entry = entry.pNext {
		if entry.uTruth[0] == truth[0] && entry.uTruth[1] == truth[1] {
			if phase != nil {
				*phase = entry.uPhase
			}
			return entry.pGates
		}
	}
	return nil
}

func tableKey(truth []uint32, size int32) uint32 {
	if size <= 0 {
		return 0
	}
	return (truth[0] + 2003*truth[1]) % uint32(size)
}

func Map_SuperTableResize(table *Map_HashTable_t) {
	newSize := abcutil.Abc_PrimeCudd(uint32(2 * table.nBins))
	bins := make([]*Map_HashEntry_t, newSize)
	for _, head := range table.pBins {
		for entry := head; entry != nil; {
			next := entry.pNext
			key := tableKey(entry.uTruth[:], newSize)
			entry.pNext, bins[key] = bins[key], entry
			entry = next
		}
	}
	table.pBins, table.nBins = bins, newSize
}

func Map_SuperTableCompareSupergates(a, b *Map_Super_t) int {
	if a.nUsed > b.nUsed {
		return -1
	}
	if a.nUsed < b.nUsed {
		return 1
	}
	return 0
}

func Map_SuperTableCompareGatesInList(a, b *Map_Super_t) int {
	if a.Area > b.Area {
		return -1
	}
	if a.Area < b.Area {
		return 1
	}
	return 0
}

func Map_SuperTableSortSupergates(table *Map_HashTable_t, maxSupers int32) {
	supers := make([]*Map_Super_t, 0, max(0, int(maxSupers)))
	for _, head := range table.pBins {
		for entry := head; entry != nil; entry = entry.pNext {
			for gate := entry.pGates; gate != nil; gate = gate.pNext {
				supers = append(supers, gate)
			}
		}
	}
	sort.SliceStable(supers, func(i, j int) bool { return Map_SuperTableCompareSupergates(supers[i], supers[j]) < 0 })
	for i, gate := range supers {
		if i == 10 || gate.nUsed == 0 {
			break
		}
		fmt.Printf("%5d : %5d   A = %5.2f   D = %5.2f   %s\n", gate.nUsed, gate.Num, gate.Area, gate.tDelayMax.Rise, gate.pFormula)
	}
}

func Map_SuperTableSortSupergatesByDelay(table *Map_HashTable_t, _ int32) {
	for _, entry := range table.pBins {
		for ; entry != nil; entry = entry.pNext {
			var gates []*Map_Super_t
			for gate := entry.pGates; gate != nil; gate = gate.pNext {
				gates = append(gates, gate)
			}
			sort.SliceStable(gates, func(i, j int) bool { return Map_SuperTableCompareGatesInList(gates[i], gates[j]) < 0 })
			entry.pGates = nil
			for _, gate := range gates {
				gate.pNext, entry.pGates = entry.pGates, gate
			}
			if entry.pGates != nil {
				entry.pGates.nSupers = uint32(len(gates))
			}
		}
	}
}
