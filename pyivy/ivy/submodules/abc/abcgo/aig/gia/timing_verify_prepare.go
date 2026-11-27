package gia

import (
	"errors"
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
)

// giaPrepareBoxVerification reads the specification and collapses both graphs,
// aligning retained black boxes by their original IDs when necessary.
func giaPrepareBoxVerification(impl *Gia_Man_t, path string, sequential int32) (*Gia_Man_t, *Gia_Man_t, error) {
	if path == "" {
		path = impl.PSpec
	}
	if path == "" {
		return nil, nil, errors.New("Spec file is not given. Use standard flow.")
	}
	if Gia_ManBoxNum(impl) > 0 && impl.PAigExtra == nil {
		return nil, nil, errors.New("Design has no box logic. Use standard flow.")
	}
	spec, err := Gia_AigerRead(path, 0, 0, 0)
	if err != nil { // C printed nothing here and crashed (C bug 48)
		return nil, nil, fmt.Errorf("%s\nCannot read spec AIG \"%s\".", GiaAigerReadMessage(err), path)
	}
	defer Gia_ManStop(spec)
	if Gia_ManBoxNum(spec) > 0 && spec.PAigExtra == nil {
		return nil, nil, errors.New("Spec has no box logic. Use standard flow.")
	}
	var present *Vec_Int_t
	if spec.PManTime != nil && tim.Tim_ManBlackBoxNum(spec.PManTime.(*tim.Tim_Man_t)) > 0 && Gia_ManBoxNum(impl) > 0 {
		if Gia_ManBoxNum(spec) < Gia_ManBoxNum(impl) {
			return nil, nil, errors.New("Spec has less boxes than the design. Cannot proceed.")
		}
		if Gia_ManBoxNum(spec) > Gia_ManBoxNum(impl) {
			mask := tim.Tim_ManAlignTwo(spec.PManTime.(*tim.Tim_Man_t), impl.PManTime.(*tim.Tim_Man_t))
			if mask == nil {
				return nil, nil, errors.New("Boxes of spec and design cannot be aligned. Cannot proceed.")
			}
			present = Vec_IntAlloc(len(mask))
			for _, v := range mask {
				Vec_IntPush(present, int32(v))
			}
		}
	}
	var left, right *Gia_Man_t
	if Gia_ManBoxNum(spec) > 0 {
		left = Gia_ManDupCollapse(spec, spec.PAigExtra, present, sequential)
	} else {
		left = Gia_ManDup(spec)
	}
	if Gia_ManBoxNum(impl) > 0 {
		right = Gia_ManDupCollapse(impl, impl.PAigExtra, nil, sequential)
	} else {
		right = Gia_ManDup(impl)
	}
	return left, right, nil
}
