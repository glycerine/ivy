package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestFalseRebuildEmptyHooksPreservesTypedLogic(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		for _, reverse := range []bool{false, true} {
			t.Run(fmt.Sprintf("signs%d/reverse%t", signs, reverse), func(t *testing.T) {
				src := quantificationReviewGraph(signs, true, false)
				if reverse {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				out := Gia_ManFalseRebuild(src, Vec_WecStart(int(src.nObjs)), 0, 0)
				cofReviewCheckRebuild(t, src, out)
				Gia_ManStop(out)
				Gia_ManStop(src)
			})
		}
	}
}
func TestFalseRebuildInterfaceMetadata(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse%t", reverse), func(t *testing.T) {
			src := quantificationReviewGraph(5, true, false)
			if reverse {
				src = Gia_ManDupOrderDfsReverse(src)
			}
			if err := src.SetAigerProperties(1, 1); err != nil {
				t.Fatal(err)
			}
			src.aigerLatchInits = []int32{1}
			src.aigerComment = []byte("false-path interface")
			src.VNamesIn = Vec_PtrAlloc(int(Gia_ManCiNum(src)))
			src.VNamesOut = Vec_PtrAlloc(int(Gia_ManCoNum(src)))
			for i := int32(0); i < Gia_ManCiNum(src); i++ {
				Vec_PtrPush(src.VNamesIn, fmt.Sprintf("input%d", i))
			}
			for i := int32(0); i < Gia_ManCoNum(src); i++ {
				Vec_PtrPush(src.VNamesOut, fmt.Sprintf("output%d", i))
			}
			out := Gia_ManFalseRebuild(src, Vec_WecStart(int(src.nObjs)), 0, 0)
			if out.nConstrs != 1 || out.aigerBadOutputs != 1 || !slices.Equal(out.aigerLatchInits, src.aigerLatchInits) || !slices.Equal(out.aigerComment, src.aigerComment) || !slices.Equal(out.InputNames(), src.InputNames()) || !slices.Equal(out.OutputNames(), src.OutputNames()) {
				t.Fatalf("lost interface bad%d constraints%d resets%v comment%q names%v/%v", out.aigerBadOutputs, out.nConstrs, out.aigerLatchInits, out.aigerComment, out.InputNames(), out.OutputNames())
			}
			out.aigerLatchInits[0] = 0
			out.aigerComment[0] = 'X'
			Vec_PtrWriteEntry(out.VNamesIn, 0, "changed")
			if src.aigerLatchInits[0] != 1 || string(src.aigerComment) != "false-path interface" || src.InputNames()[0] != "input0" {
				t.Fatal("interface storage aliases source")
			}
			Gia_ManStop(out)
			Gia_ManStop(src)
		})
	}
}
