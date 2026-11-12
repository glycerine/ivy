package cov

// This file ports the manager lifecycle from src/map/cov/covMinMan.c.
// Cube storage is garbage-collected Go memory rather than a C fixed pool.

func Min_ManAlloc(nVars int32) *Min_Man_t {
	if nVars < 0 {
		panic("cov: negative support size")
	}
	p := &Min_Man_t_{ppStore: make([]*Min_Cube_t, nVars+1)}
	Min_ManClean(p, nVars)
	p.pOne0 = Min_CubeAlloc(p)
	p.pOne1 = Min_CubeAlloc(p)
	p.pTemp = Min_CubeAlloc(p)
	p.pBubble = Min_CubeAlloc(p)
	if len(p.pBubble.uData) != 0 {
		p.pBubble.uData[0] = 0
	}
	Min_ManClean(p, 1)
	p.pTriv0[0] = Min_CubeAllocVar(p, 0, 0)
	p.pTriv0[1] = Min_CubeAllocVar(p, 0, 1)
	p.pTriv1[0] = Min_CubeAllocVar(p, 0, 0)
	p.pTriv1[1] = Min_CubeAllocVar(p, 0, 1)
	Min_ManClean(p, nVars)
	return p
}

// Min_ManClean resets per-cover state for a support of nSupp variables. The
// reusable buckets grow on demand because callers sometimes temporarily use a
// one-variable support while retaining the original manager.
func Min_ManClean(p *Min_Man_t, nSupp int32) {
	if p == nil || nSupp < 0 {
		panic("cov: invalid minimizer support")
	}
	p.nVars = nSupp
	p.nWords = (2*nSupp + 31) / 32
	if len(p.ppStore) < int(nSupp)+1 {
		p.ppStore = append(p.ppStore, make([]*Min_Cube_t, int(nSupp)+1-len(p.ppStore))...)
	}
	clear(p.ppStore)
	p.nCubes = 0
}

func Min_ManFree(p *Min_Man_t) {
	if p == nil {
		return
	}
	clear(p.ppStore)
	p.ppStore = nil
	p.pOne0, p.pOne1, p.pTemp, p.pBubble = nil, nil, nil, nil
	p.pTriv0 = [2]*Min_Cube_t{}
	p.pTriv1 = [2]*Min_Cube_t{}
}
