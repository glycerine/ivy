package cudd

import (
	"fmt"
	"math/bits"
	"os"
	"reflect"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func cuddInitCache(dd *DdManager, cacheSize, maxCacheSize uint32) int32 {
	if dd == nil || cacheSize == 0 || maxCacheSize < cacheSize {
		return 0
	}
	core.Cudd_CacheFlush(dd)
	return 1
}

func cuddCacheInsert(dd *DdManager, op ptruint, f, g, h, data *DdNode) {
	core.Cudd_CacheInsert(dd, op, f, g, h, data, 0)
}

func cuddCacheInsert2(dd *DdManager, op DD_CTFP, f, g, data *DdNode) {
	core.Cudd_CacheInsert(dd, cacheFunctionID(op), f, g, nil, data, 0)
}

func cuddCacheInsert1(dd *DdManager, op DD_CTFP1, f, data *DdNode) {
	core.Cudd_CacheInsert(dd, cacheFunctionID(op), f, nil, nil, data, 0)
}

func cuddCacheLookup(dd *DdManager, op ptruint, f, g, h *DdNode) *DdNode {
	return core.Cudd_CacheLookup(dd, op, f, g, h, 0)
}

func cuddCacheLookupZdd(dd *DdManager, op ptruint, f, g, h *DdNode) *DdNode {
	return core.Cudd_CacheLookup(dd, op, f, g, h, 1)
}

func cuddCacheLookup2(dd *DdManager, op DD_CTFP, f, g *DdNode) *DdNode {
	return core.Cudd_CacheLookup(dd, cacheFunctionID(op), f, g, nil, 0)
}

func cuddCacheLookup1(dd *DdManager, op DD_CTFP1, f *DdNode) *DdNode {
	return core.Cudd_CacheLookup(dd, cacheFunctionID(op), f, nil, nil, 0)
}

func cuddCacheLookup2Zdd(dd *DdManager, op DD_CTFP, f, g *DdNode) *DdNode {
	return core.Cudd_CacheLookup(dd, cacheFunctionID(op), f, g, nil, 1)
}

func cuddCacheLookup1Zdd(dd *DdManager, op DD_CTFP1, f *DdNode) *DdNode {
	return core.Cudd_CacheLookup(dd, cacheFunctionID(op), f, nil, nil, 1)
}

func cuddConstantLookup(dd *DdManager, op ptruint, f, g, h *DdNode) *DdNode {
	return core.Cudd_CacheLookup(dd, op, f, g, h, 2)
}

func cacheFunctionID(op any) uintptr {
	if op == nil {
		return 0
	}
	return reflect.ValueOf(op).Pointer()
}

func cuddCacheProfile(dd *DdManager, fp *os.File) int32 {
	if dd == nil || fp == nil {
		return 0
	}
	_, err := fmt.Fprintf(fp, "Computed cache entries: %d\n", core.Cudd_CacheSize(dd))
	if err != nil {
		return 0
	}
	return 1
}

func cuddCacheResize(dd *DdManager) {
	if dd == nil {
		return
	}
	// The Go map grows dynamically; enforce the configured hard limit by
	// evicting the table as a whole, matching CUDD's cache-flush fallback.
	if uint32(core.Cudd_CacheSize(dd)) >= Cudd_ReadMaxCacheHard(dd) {
		core.Cudd_CacheFlush(dd)
	}
}

func cuddCacheFlush(dd *DdManager) { core.Cudd_CacheFlush(dd) }

func cuddComputeFloorLog2(value uint32) uint32 {
	if value == 0 {
		return 0
	}
	return uint32(bits.Len32(value) - 1)
}
