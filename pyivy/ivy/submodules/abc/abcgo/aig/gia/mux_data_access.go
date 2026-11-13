package gia

// Gia_ManHasMuxData reports whether explicit MUX control storage exists.
func Gia_ManHasMuxData(p *Gia_Man_t) bool { return p != nil && p.pMuxes != nil }
