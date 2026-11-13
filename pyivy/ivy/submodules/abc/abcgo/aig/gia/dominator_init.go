package gia

import "github.com/glycerine/abcgo/misc/vec"

// Gia_ManResetDoms creates or resets the object-indexed dominator array.
func Gia_ManResetDoms(p *Gia_Man_t) {
	if p == nil {
		panic("gia: nil manager")
	}
	if p.vDoms == nil {
		p.vDoms = vec.Vec_IntAlloc(0)
	}
	vec.Vec_IntFill(p.vDoms, int(p.nObjs), -1)
}
