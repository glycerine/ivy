package wln

import (
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Wln_NtkAlloc creates the object-zero sentinel and the source range table.
func Wln_NtkAlloc(name string, nObjsMax int32) *Wln_Ntk_t {
	p := &Wln_Ntk_t{}
	if name != "" {
		p.pName = extra.Extra_FileNameGeneric(name)
	}
	vec.Vec_IntGrow(&p.vCis, 111)
	vec.Vec_IntGrow(&p.vCos, 111)
	vec.Vec_IntGrow(&p.vFfs, 111)
	n := int(nObjsMax) + 1
	vec.Vec_IntGrow(&p.vTypes, n)
	vec.Vec_StrGrow(&p.vSigns, n)
	vec.Vec_IntGrow(&p.vRanges, n)
	vec.Vec_IntPush(&p.vTypes, -1)
	vec.Vec_StrPush(&p.vSigns, 255)
	vec.Vec_IntPush(&p.vRanges, -1)
	p.vFanins = make([]Wln_Vec_t, n)
	p.pRanges = vec.Hash_IntManStart(1000)
	for i := int32(0); i < 65; i++ {
		vec.Hash_Int2ManInsert(p.pRanges, i, i, 0)
	}
	for i := int32(1); i < 64; i++ {
		vec.Hash_Int2ManInsert(p.pRanges, i, 0, 0)
	}
	assert(vec.Hash_IntManEntryNum(p.pRanges) == 128, "Hash_IntManEntryNum(p->pRanges) == 128")
	return p
}

// Wln_NtkFree clears owned allocations, matching the C manager teardown.
func Wln_NtkFree(p *Wln_Ntk_t) {
	if p.pRanges != nil {
		vec.Hash_IntManStop(p.pRanges)
	}
	if p.pManName != nil {
		util.Abc_NamStop(p.pManName)
	}
	*p = Wln_Ntk_t{}
}
