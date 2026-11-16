package gia

import (
	"reflect"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
)

func mappingReviewTimingMetadata(mask, owner int) *Gia_Man_t {
	p := Gia_ManStart(1)
	iv := func(value int32) *Vec_Int_t { v := Vec_IntAlloc(1); Vec_IntPush(v, value); return v }
	fv := func(value float32) *Vec_Flt_t { v := Vec_FltAlloc(1); Vec_FltPush(v, value); return v }
	pv := func(value string) *Vec_Ptr_t { v := Vec_PtrAlloc(1); Vec_PtrPush(v, value); return v }
	value := int32(owner * 100)
	prefix := "s"
	if owner == 2 {
		prefix = "d"
	}
	if mask&1 != 0 {
		p.vCiArrs = iv(value + 1)
	}
	if mask&2 != 0 {
		p.vCoReqs = iv(value + 2)
	}
	if mask&4 != 0 {
		p.vCoArrs = iv(value + 3)
	}
	if mask&8 != 0 {
		p.vCoAttrs = iv(value + 4)
	}
	if mask&16 != 0 {
		p.VInArrs = fv(float32(owner) + 0.25)
	}
	if mask&32 != 0 {
		p.VOutReqs = fv(float32(owner) + 0.5)
	}
	if mask&64 != 0 {
		p.VNamesIn = pv(prefix + "-in")
	}
	if mask&128 != 0 {
		p.VNamesOut = pv(prefix + "-out")
	}
	if mask&256 != 0 {
		p.VConfigs = iv(value + 9)
	}
	if mask&512 != 0 {
		p.PCellStr = prefix
	}
	if mask&1024 != 0 {
		p.PManTime = tim.Tim_ManStart(owner, owner)
	}
	if mask&1 != 0 {
		p.PAigExtra = Gia_ManStart(1)
		p.PAigExtra.SetIdentity(prefix+"-extra", "")
	}
	if mask&2 != 0 {
		p.vRegClasses = iv(value + 11)
	}
	if mask&4 != 0 {
		p.vRegInits = iv(value + 12)
	}
	p.And2Delay, p.nAnd2Delay = int32(owner*10+3), int32(owner*10+7)
	p.DefInArrs, p.DefOutReqs = float32(owner)+0.25, float32(owner)+0.5
	return p
}
func mappingReviewTimingPointers(p *Gia_Man_t) []any {
	return []any{p.vCiArrs, p.vCoReqs, p.vCoArrs, p.vCoAttrs, p.VInArrs, p.VOutReqs, p.VNamesIn, p.VNamesOut, p.VConfigs, p.PManTime, p.PAigExtra, p.vRegClasses, p.vRegInits}
}

// Snapshots retain pointer identity, including independently owned empty vectors.
// Normalize typed nil pointers so the table can express absent fields uniformly.
func mappingReviewTimingSnapshot(p *Gia_Man_t) []any {
	fields := mappingReviewTimingPointers(p)
	for i, value := range fields {
		if value == nil || reflect.ValueOf(value).IsNil() {
			fields[i] = nil
		}
	}
	return append(fields, p.PCellStr, p.And2Delay, p.nAnd2Delay, p.DefInArrs, p.DefOutReqs)
}

func TestMappingTimingTransferMasksAndOwnershipGoOnly(t *testing.T) {
	// Groups describe the source contract, not a second transfer implementation.
	// Pointer/string members move together, even when some are absent. Scalar
	// defaults are copied without being cleared; only boxed nAnd2Delay moves.
	groups := []struct {
		present int
		members []int
		scalars []int
		boxed   bool
	}{
		{0x00f, []int{0, 1, 2, 3}, []int{14}, false},
		{0x030, []int{4, 5}, []int{14, 16, 17}, false},
		{0x0c0, []int{6, 7}, nil, false},
		{0x300, []int{8, 13}, nil, false},
		{0x400, []int{9, 10, 11, 12, 15}, nil, true},
	}
	zero := make([]any, 18)
	zero[13], zero[15] = "", int32(0)
	count := 0
	for mask := 0; mask < 2048; mask++ {
		for variant := 0; variant < 3; variant++ {
			src := mappingReviewTimingMetadata(mask, 1)
			dstMask := 0
			if variant == 1 {
				dstMask = 2047
			}
			dst := mappingReviewTimingMetadata(dstMask, 2)
			target := dst
			if variant == 2 {
				target = src
			}
			before := mappingReviewTimingSnapshot(src)
			wantSource := append([]any(nil), before...)
			wantTarget := mappingReviewTimingSnapshot(target)
			for _, group := range groups {
				if mask&group.present == 0 || (group.boxed && target == src) {
					continue
				}
				for _, index := range group.members {
					wantSource[index] = zero[index]
					wantTarget[index] = before[index]
					if target == src {
						wantTarget[index] = zero[index]
					}
				}
				for _, index := range group.scalars {
					wantTarget[index] = before[index]
				}
			}
			Gia_ManTransferTiming(target, src)
			if got := mappingReviewTimingSnapshot(target); !slices.Equal(got, wantTarget) {
				t.Fatalf("mask%x variant%d destination: got%v want%v", mask, variant, got, wantTarget)
			}
			if got := mappingReviewTimingSnapshot(src); !slices.Equal(got, wantSource) {
				t.Fatalf("mask%x variant%d source: got%v want%v", mask, variant, got, wantSource)
			}
			Gia_ManStop(src)
			Gia_ManStop(dst)
			count++
		}
	}
	if count != 6144 {
		t.Fatalf("checked%d configurations", count)
	}
}
