package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManTransferTiming in giaIf.c transfers each metadata group when any
// member is present, including clearing absent members at the destination.
func TestTimingTransferPartialGroups(t *testing.T) {
	src, dst := Gia_ManStart(4), Gia_ManStart(4)
	src.vCoReqs = vec.Vec_IntAlloc(1)
	src.vOutReqs = vec.Vec_FltAlloc(1)
	src.vNamesOut = vec.Vec_PtrAlloc(1)
	src.vConfigs = vec.Vec_IntAlloc(1)
	src.DefInArrs, src.DefOutReqs, src.And2Delay = 2, 3, 4
	dst.vCiArrs, dst.vCoArrs, dst.vCoAttrs = vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	dst.vInArrs = vec.Vec_FltAlloc(1)
	dst.vNamesIn = vec.Vec_PtrAlloc(1)
	dst.pCellStr = "old"
	coReqs, outReqs, names, configs := src.vCoReqs, src.vOutReqs, src.vNamesOut, src.vConfigs
	Gia_ManTransferTiming(dst, src)
	if dst.vCoReqs != coReqs || dst.vCiArrs != nil || dst.vCoArrs != nil || dst.vCoAttrs != nil || src.vCoReqs != nil {
		t.Fatal("integer timing group did not transfer as a group")
	}
	if dst.vOutReqs != outReqs || dst.vInArrs != nil || src.vOutReqs != nil || dst.DefInArrs != 2 || dst.DefOutReqs != 3 || dst.And2Delay != 4 {
		t.Fatal("float timing group did not transfer as a group")
	}
	if dst.vNamesOut != names || dst.vNamesIn != nil || src.vNamesOut != nil || dst.vConfigs != configs || dst.pCellStr != "" || src.vConfigs != nil {
		t.Fatal("name or cell group did not transfer as a group")
	}
}

// C checks dst == src only after the groups above; box timing is retained.
func TestTimingTransferSelfMatchesSourceOrder(t *testing.T) {
	g := Gia_ManStart(4)
	g.vCiArrs = vec.Vec_IntAlloc(1)
	g.vInArrs = vec.Vec_FltAlloc(1)
	g.vNamesOut = vec.Vec_PtrAlloc(1)
	g.vConfigs = vec.Vec_IntAlloc(1)
	g.pCellStr = "cell"
	g.pManTime = tim.Tim_ManStart(0, 0)
	g.pAigExtra = Gia_ManStart(1)
	g.vRegClasses, g.vRegInits = vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	g.nAnd2Delay = 8
	timing, extra, classes, inits := g.pManTime, g.pAigExtra, g.vRegClasses, g.vRegInits
	Gia_ManTransferTiming(g, g)
	if g.vCiArrs != nil || g.vInArrs != nil || g.vNamesOut != nil || g.vConfigs != nil || g.pCellStr != "" {
		t.Fatal("self transfer did not clear pre-box metadata")
	}
	if g.pManTime != timing || g.pAigExtra != extra || g.vRegClasses != classes || g.vRegInits != inits || g.nAnd2Delay != 8 {
		t.Fatal("self transfer changed box metadata")
	}
}

func TestTimingTransferBoxGroupRequiresManager(t *testing.T) {
	src, dst := Gia_ManStart(4), Gia_ManStart(4)
	src.pAigExtra, src.vRegClasses, src.vRegInits = Gia_ManStart(1), vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	src.nAnd2Delay = 6
	dst.pManTime, dst.pAigExtra = tim.Tim_ManStart(0, 0), Gia_ManStart(1)
	dst.vRegClasses, dst.vRegInits, dst.nAnd2Delay = vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1), 9
	oldTime, oldExtra, oldClasses, oldInits := dst.pManTime, dst.pAigExtra, dst.vRegClasses, dst.vRegInits
	Gia_ManTransferTiming(dst, src)
	if dst.pManTime != oldTime || dst.pAigExtra != oldExtra || dst.vRegClasses != oldClasses || dst.vRegInits != oldInits || dst.nAnd2Delay != 9 || src.pAigExtra == nil {
		t.Fatal("box metadata moved without a timing manager")
	}
	src.pManTime = tim.Tim_ManStart(0, 0)
	timing, extra, classes, inits := src.pManTime, src.pAigExtra, src.vRegClasses, src.vRegInits
	Gia_ManTransferTiming(dst, src)
	if dst.pManTime != timing || dst.pAigExtra != extra || dst.vRegClasses != classes || dst.vRegInits != inits || dst.nAnd2Delay != 6 {
		t.Fatal("box metadata not transferred")
	}
	if src.pManTime != nil || src.pAigExtra != nil || src.vRegClasses != nil || src.vRegInits != nil || src.nAnd2Delay != 0 {
		t.Fatal("box metadata ownership not cleared")
	}
}
