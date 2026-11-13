package gia

// Gia_ManHasStaticFanout reports whether the static fanout map is installed.
func Gia_ManHasStaticFanout(p *Gia_Man_t) bool {
	return p != nil && p.vFanout != nil && p.vFanoutNums != nil
}
