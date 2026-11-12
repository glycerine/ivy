package amap

// Amap_ManStart creates the state owned by one mapping operation. Node
// addresses are stable Go allocations retained in objects; temporary cuts and
// best-match arenas are separate so exact-area remapping can reset them.
func Amap_ManStart(nNodes int32) *Amap_Man_t {
	if nNodes < 0 {
		panic("amap: negative node estimate")
	}
	p := &Amap_Man_t_{fEpsilonInternal: 0.01, sourceCopies: make(map[int]*Amap_Obj_t)}
	p.objects = make([]*Amap_Obj_t, 0, nNodes+1)
	p.vPis = Vec_PtrAlloc(100)
	p.vPos = Vec_PtrAlloc(100)
	p.vObjs = Vec_PtrAlloc(nNodes + 1)
	p.vTemp = Vec_IntAlloc(100)
	p.vCuts0 = Vec_PtrAlloc(100)
	p.vCuts1 = Vec_PtrAlloc(100)
	p.vCuts2 = Vec_PtrAlloc(100)
	p.vTempP = Vec_PtrAlloc(100)
	return p
}

// Amap_ManStop releases manager-owned containers and drops all graph nodes.
// The library belongs to the caller and is never destroyed here.
func Amap_ManStop(p *Amap_Man_t) {
	if p == nil {
		return
	}
	Vec_PtrFree(p.vPis)
	Vec_PtrFree(p.vPos)
	Vec_PtrFree(p.vObjs)
	Vec_PtrFree(p.vCuts0)
	Vec_PtrFree(p.vCuts1)
	Vec_PtrFree(p.vCuts2)
	Vec_PtrFree(p.vTempP)
	Vec_IntFree(p.vTemp)
	*p = Amap_Man_t_{}
}
