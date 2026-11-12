package dec

import (
	"github.com/glycerine/abcgo/misc/extra/npn"
	"github.com/glycerine/abcgo/misc/mvc"
)

// Dec_ManStart allocates per-manager factoring scratch and full four-input NPN
// canonicalization tables. It has no dependency on an ABC process-global frame.
func Dec_ManStart() *Dec_Man_t {
	canon, phases, perms, classes := npn.Extra_Truth4VarNPN()
	return &Dec_Man_t{pMvcMem: mvc.Mvc_ManagerStart(), vCubes: make([]int32, 0, 8), vLits: make([]int32, 0, 8), puCanons: canon, pPhases: phases, pPerms: perms, pMap: classes}
}
func Dec_ManStop(p *Dec_Man_t) {
	if p == nil {
		return
	}
	if manager, ok := p.pMvcMem.(*mvc.Mvc_Manager_t); ok {
		mvc.Mvc_ManagerFree(manager)
	}
	*p = Dec_Man_t{}
}

// CanonicalTruth returns the canonical representative and ABC phase/permutation
// indices, followed by the NPN equivalence-class number for a four-input truth.
func (p *Dec_Man_t) CanonicalTruth(truth uint16) (uint16, byte, byte, byte) {
	if len(p.puCanons) != 1<<16 {
		panic("dec: manager stopped")
	}
	return p.puCanons[truth], p.pPhases[truth], p.pPerms[truth], p.pMap[truth]
}
