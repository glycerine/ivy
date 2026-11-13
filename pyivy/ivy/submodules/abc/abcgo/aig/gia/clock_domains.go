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

// SetClockDomains replaces the register-index partitions owned by the graph.
// A nil slice removes the vector; an empty non-nil slice sets an empty vector.
func (p *Gia_Man_t) SetClockDomains(domains [][]int32) {
	if domains == nil {
		p.vClockDoms = nil
		return
	}
	v := vec.Vec_VecAlloc(len(domains))
	vec.Vec_VecExpandInt(v, len(domains)-1)
	for i, domain := range domains {
		for _, reg := range domain {
			vec.Vec_VecPushInt(v, i, reg)
		}
	}
	p.vClockDoms = v
}
