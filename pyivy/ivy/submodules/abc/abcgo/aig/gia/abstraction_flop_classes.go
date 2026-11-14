package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

// AbstractionFlopClasses returns a copy of the register class map, or nil
// when no map has been installed.
func (p *Gia_Man_t) AbstractionFlopClasses() []int32 {
	if p.VFlopClasses == nil {
		return nil
	}
	values := vec.Vec_IntArray(p.VFlopClasses)
	return append(make([]int32, 0, len(values)), values...)
}

// SetAbstractionFlopClasses installs a copy of one class per register.
// A nil slice clears the map.
func (p *Gia_Man_t) SetAbstractionFlopClasses(classes []int32) error {
	if classes == nil {
		p.VFlopClasses = nil
		return nil
	}
	if len(classes) != int(Gia_ManRegNum(p)) {
		return fmt.Errorf("expected %d flop classes, got %d", Gia_ManRegNum(p), len(classes))
	}
	p.VFlopClasses = vec.Vec_IntAllocArrayCopy(classes, len(classes))
	return nil
}
