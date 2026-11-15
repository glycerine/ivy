package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestFxInsertRetainsOwnedInterface(t *testing.T) {
	for _, rebuilt := range []bool{false, true} {
		t.Run(fmt.Sprintf("rebuilt%t", rebuilt), func(t *testing.T) {
			p := Gia_ManStart(32)
			defer Gia_ManStop(p)
			a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			x := Gia_ManAppendAnd(p, a, b)
			y, z := Gia_ManAppendAnd(p, x, c), Gia_ManAppendAnd(p, x, c^1)
			Gia_ManAppendCo(p, y)
			Gia_ManAppendCo(p, z^1)
			Gia_ManAppendCo(p, x)
			Gia_ManSetRegNum(p, 1)
			p.pObjs[x>>1].Value = 3
			p.pObjs[y>>1].Value = 4
			p.pObjs[z>>1].Value = 5
			if err := p.SetAigerProperties(1, 1); err != nil {
				t.Fatal(err)
			}
			p.aigerLatchInits = []int32{1}
			p.aigerComment = []byte("extraction interface")
			p.VNamesIn = Vec_PtrAlloc(3)
			p.VNamesOut = Vec_PtrAlloc(3)
			for i := 0; i < 3; i++ {
				Vec_PtrPush(p.VNamesIn, fmt.Sprintf("input%d", i))
				Vec_PtrPush(p.VNamesOut, fmt.Sprintf("output%d", i))
			}
			rows := [][]int32{{3, 0, 2}, {4, 4, 6}, {5, 5, 6}}
			if rebuilt {
				rows = [][]int32{{3, 12}, {4, 4, 12}, {5, 5, 12}, {6, 0, 2}}
			}
			cubes := Vec_WecAlloc(len(rows))
			defer Vec_WecFree(cubes)
			for _, row := range rows {
				v := Vec_WecPushLevel(cubes)
				for _, lit := range row {
					Vec_IntPush(v, lit)
				}
			}
			phases := Vec_StrStart(6)
			defer Vec_StrFree(phases)
			out := Gia_ManFxInsert(p, cubes, phases)
			defer Gia_ManStop(out)
			cofReviewCheckRebuild(t, p, out)
			if out.nConstrs != 1 || out.aigerBadOutputs != 1 || !slices.Equal(out.aigerLatchInits, p.aigerLatchInits) || !slices.Equal(out.aigerComment, p.aigerComment) || !slices.Equal(out.InputNames(), p.InputNames()) || !slices.Equal(out.OutputNames(), p.OutputNames()) {
				t.Fatalf("lost interface: bad%d constraints%d resets%v names%v/%v", out.aigerBadOutputs, out.nConstrs, out.aigerLatchInits, out.InputNames(), out.OutputNames())
			}
			out.aigerLatchInits[0] = 0
			out.aigerComment[0] = 'X'
			Vec_PtrWriteEntry(out.VNamesIn, 0, "changed")
			if p.aigerLatchInits[0] != 1 || string(p.aigerComment) != "extraction interface" || p.InputNames()[0] != "input0" {
				t.Fatal("interface storage aliases source")
			}
		})
	}
}
