package gia

import (
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
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
func mappingReviewTimingPrint(w io.Writer, name string, p, oldSrc, oldDst *Gia_Man_t) {
	values, source, destination := mappingReviewTimingPointers(p), mappingReviewTimingPointers(oldSrc), mappingReviewTimingPointers(oldDst)
	var fromSrc, fromDst uint32
	for i, value := range values {
		if value == nil || reflect.ValueOf(value).IsNil() {
			continue
		}
		if value == source[i] {
			fromSrc |= 1 << i
		}
		if value == destination[i] {
			fromDst |= 1 << i
		}
	}
	iv := func(v *Vec_Int_t) int32 {
		if v == nil {
			return -1
		}
		return Vec_IntEntry(v, 0)
	}
	fv := func(v *Vec_Flt_t) float32 {
		if v == nil {
			return -1
		}
		return Vec_FltEntry(v, 0)
	}
	pv := func(v *Vec_Ptr_t) string {
		if v == nil {
			return "-"
		}
		return Vec_PtrEntry(v, 0).(string)
	}
	cell := p.PCellStr
	if cell == "" {
		cell = "-"
	}
	cis := -1
	if p.PManTime != nil {
		cis = tim.Tim_ManCiNum(p.PManTime.(*tim.Tim_Man_t))
	}
	extra := "-"
	if p.PAigExtra != nil {
		extra = p.PAigExtra.pName
	}
	fmt.Fprintf(w, "%s %d %d %d %d %.2f %.2f %s %s %d %s %d %s %d %d %d %d %.2f %.2f origin %x %x\n", name, iv(p.vCiArrs), iv(p.vCoReqs), iv(p.vCoArrs), iv(p.vCoAttrs), fv(p.VInArrs), fv(p.VOutReqs), pv(p.VNamesIn), pv(p.VNamesOut), iv(p.VConfigs), cell, cis, extra, iv(p.vRegClasses), iv(p.vRegInits), p.And2Delay, p.nAnd2Delay, p.DefInArrs, p.DefOutReqs, fromSrc, fromDst)
}
func TestMappingNativeTimingTransferMasksAndOwnership(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_timing_transfer_review.c")
	source, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native metadata transfer: %v %s", err, source)
	}
	var got strings.Builder
	for mask := 0; mask < 2048; mask++ {
		for variant := 0; variant < 3; variant++ {
			src := mappingReviewTimingMetadata(mask, 1)
			dstMask := 0
			if variant == 1 {
				dstMask = 2047
			}
			dst := mappingReviewTimingMetadata(dstMask, 2)
			oldSrc, oldDst := *src, *dst
			target := dst
			if variant == 2 {
				target = src
			}
			Gia_ManTransferTiming(target, src)
			fmt.Fprintf(&got, "case %d %d\n", mask, variant)
			mappingReviewTimingPrint(&got, "dst", target, &oldSrc, &oldDst)
			mappingReviewTimingPrint(&got, "src", src, &oldSrc, &oldDst)
			Gia_ManStop(src)
			Gia_ManStop(dst)
		}
	}
	if string(source) != got.String() {
		esopReviewCompare(t, source, got.String())
	}
	t.Log("compared 6,144 source/destination/self-transfer combinations and pointer origins")
}
