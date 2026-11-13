package gia

import "github.com/glycerine/abcgo/misc/vec"

// ClockDomains returns the register partitions attached to this manager.
// The returned vectors belong to the manager and must not be modified.
func (p *Gia_Man_t) ClockDomains() []*vec.Vec_Int_t {
	if p.vClockDoms == nil {
		return nil
	}
	domains := make([]*vec.Vec_Int_t, vec.Vec_VecSize(p.vClockDoms))
	for i := range domains {
		domains[i] = vec.Vec_VecEntryInt(p.vClockDoms, i)
	}
	return domains
}
