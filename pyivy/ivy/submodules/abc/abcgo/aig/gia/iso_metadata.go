package gia

// SetIsomorphism copies the object-indexed pairing used by CEC's class
// conversion. A zero entry means no partner.
func (p *Gia_Man_t) SetIsomorphism(pairs []int32) {
	if len(pairs) != int(p.nObjs) {
		panic("gia: isomorphism pair count does not match object count")
	}
	p.pIso = append(p.pIso[:0], pairs...)
}

// Isomorphism returns a copy so callers cannot mutate manager-owned metadata.
func (p *Gia_Man_t) Isomorphism() []int32 {
	return append([]int32(nil), p.pIso...)
}
