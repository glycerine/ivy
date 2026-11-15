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

func TestFalsePathDriverInterfaceOwnership(t *testing.T) {
	for _, driver := range []string{"one", "batch", "outputs", "all"} {
		for _, reverse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reverse%t", driver, reverse), func(t *testing.T) {
				g := Gia_ManStart(64)
				g.EnableMuxes()
				a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
				x := Gia_ManAppendAnd(g, a, b)
				f := Gia_ManAppendAnd(g, x, a^1)
				var side int32
				if driver == "one" {
					side = Gia_ManAppendAnd(g, c, d)
				} else {
					z := Gia_ManAppendXorReal(g, c, d)
					side = Gia_ManAppendMuxReal(g, a, z, b)
				}
				Gia_ManAppendCo(g, f)
				Gia_ManAppendCo(g, side^1)
				Gia_ManAppendCo(g, side)
				Gia_ManSetRegNum(g, 1)
				if reverse {
					g = Gia_ManDupOrderDfsReverse(g)
				}
				defer Gia_ManStop(g)
				if err := g.SetAigerProperties(1, 1); err != nil {
					t.Fatal(err)
				}
				g.SetIdentity("false driver", "false_driver.aig")
				g.aigerLatchInits = []int32{1}
				g.aigerComment = []byte("driver ownership")
				g.VNamesIn = Vec_PtrAlloc(4)
				g.VNamesOut = Vec_PtrAlloc(3)
				for i := 0; i < 4; i++ {
					Vec_PtrPush(g.VNamesIn, fmt.Sprintf("ci%d", i))
				}
				for i := 0; i < 3; i++ {
					Vec_PtrPush(g.VNamesOut, fmt.Sprintf("co%d", i))
				}
				Gia_ManLevelNum(g)
				var out *Gia_Man_t
				hashReviewStdout(t, func() {
					switch driver {
					case "one":
						out = Gia_ManCheckOne(g, 0, Gia_ObjFaninId0p(g, Gia_ManCo(g, 0)), 0, 0, 0)
					case "batch":
						out = Gia_ManCheckFalse2(g, 32, 0, 0, 0)
					case "outputs":
						out = Gia_ManCheckFalse(g, 32, 0, 0, 0)
					case "all":
						out = Gia_ManCheckFalseAll(g, 32, 0, 0, 0)
					}
				})
				if out == nil {
					t.Fatal("missing transformed graph")
				}
				defer Gia_ManStop(out)
				cofReviewCheckRebuild(t, g, out)
				if out.nConstrs != 1 || out.aigerBadOutputs != 1 || !slices.Equal(out.aigerLatchInits, g.aigerLatchInits) || !slices.Equal(out.aigerComment, g.aigerComment) || !slices.Equal(out.InputNames(), g.InputNames()) || !slices.Equal(out.OutputNames(), g.OutputNames()) {
					t.Fatal("driver lost unchanged interface metadata")
				}
				if name, spec := out.Identity(); name != "false driver" || spec != "false_driver.aig" {
					t.Fatal("driver lost identity")
				}
				out.aigerLatchInits[0] = 0
				out.aigerComment[0] = 'X'
				Vec_PtrWriteEntry(out.VNamesIn, 0, "changed")
				if g.aigerLatchInits[0] != 1 || string(g.aigerComment) != "driver ownership" || g.InputNames()[0] != "ci0" {
					t.Fatal("driver output aliases source metadata")
				}
			})
		}
	}
}
