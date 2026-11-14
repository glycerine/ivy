package gia

// Gia_ManPacking returns the manager's borrowed packing vector, or nil when
// no packing metadata is attached. The vector remains owned by the manager.
func Gia_ManPacking(p *Gia_Man_t) *Vec_Int_t { return p.VPacking }
