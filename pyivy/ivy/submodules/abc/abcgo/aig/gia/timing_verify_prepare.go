package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
)

// giaPrepareBoxVerification reads the specification and collapses both graphs,
// aligning retained black boxes by their original IDs when necessary.
func giaPrepareBoxVerification(impl *Gia_Man_t, path string, sequential int32) (*Gia_Man_t, *Gia_Man_t, error) {
	if path == "" {
		path = impl.pSpec
	}
	if path == "" {
		return nil, nil, fmt.Errorf("specification file is not given")
	}
	if Gia_ManBoxNum(impl) > 0 && impl.pAigExtra == nil {
		return nil, nil, fmt.Errorf("design has no box logic")
	}
	spec, err := Gia_AigerRead(path, 0, 0, 0)
	if err != nil {
		return nil, nil, err
	}
	defer Gia_ManStop(spec)
	if Gia_ManBoxNum(spec) > 0 && spec.pAigExtra == nil {
		return nil, nil, fmt.Errorf("specification has no box logic")
	}
	var present *Vec_Int_t
	if spec.pManTime != nil && tim.Tim_ManBlackBoxNum(spec.pManTime.(*tim.Tim_Man_t)) > 0 && Gia_ManBoxNum(impl) > 0 {
		if Gia_ManBoxNum(spec) < Gia_ManBoxNum(impl) {
			return nil, nil, fmt.Errorf("specification has fewer boxes than implementation")
		}
		if Gia_ManBoxNum(spec) > Gia_ManBoxNum(impl) {
			mask := tim.Tim_ManAlignTwo(spec.pManTime.(*tim.Tim_Man_t), impl.pManTime.(*tim.Tim_Man_t))
			if mask == nil {
				return nil, nil, fmt.Errorf("specification and implementation boxes cannot be aligned")
			}
			present = Vec_IntAlloc(len(mask))
			for _, v := range mask {
				Vec_IntPush(present, int32(v))
			}
		}
	}
	var left, right *Gia_Man_t
	if Gia_ManBoxNum(spec) > 0 {
		left = Gia_ManDupCollapse(spec, spec.pAigExtra, present, sequential)
	} else {
		left = Gia_ManDup(spec)
	}
	if Gia_ManBoxNum(impl) > 0 {
		right = Gia_ManDupCollapse(impl, impl.pAigExtra, nil, sequential)
	} else {
		right = Gia_ManDup(impl)
	}
	return left, right, nil
}
