package dec

import (
	"github.com/glycerine/abcgo/misc/extra/npn"
)

// Dec_ManStart allocates full four-input NPN
// canonicalization tables. It has no dependency on an ABC process-global frame.
func Dec_ManStart() *Dec_Man_t {
	canon, phases, perms, classes := npn.Extra_Truth4VarNPN()
	return &Dec_Man_t{puCanons: canon, pPhases: phases, pPerms: perms, pMap: classes}
}
func Dec_ManStop(p *Dec_Man_t) {
	if p == nil {
		return
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
