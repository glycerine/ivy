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
	src.VOutReqs = vec.Vec_FltAlloc(1)
	src.VNamesOut = vec.Vec_PtrAlloc(1)
	src.VConfigs = vec.Vec_IntAlloc(1)
	src.DefInArrs, src.DefOutReqs, src.And2Delay = 2, 3, 4
	dst.vCiArrs, dst.vCoArrs, dst.vCoAttrs = vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	dst.VInArrs = vec.Vec_FltAlloc(1)
	dst.VNamesIn = vec.Vec_PtrAlloc(1)
	dst.PCellStr = "old"
	coReqs, outReqs, names, configs := src.vCoReqs, src.VOutReqs, src.VNamesOut, src.VConfigs
	Gia_ManTransferTiming(dst, src)
	if dst.vCoReqs != coReqs || dst.vCiArrs != nil || dst.vCoArrs != nil || dst.vCoAttrs != nil || src.vCoReqs != nil {
		t.Fatal("integer timing group did not transfer as a group")
	}
	if dst.VOutReqs != outReqs || dst.VInArrs != nil || src.VOutReqs != nil || dst.DefInArrs != 2 || dst.DefOutReqs != 3 || dst.And2Delay != 4 {
		t.Fatal("float timing group did not transfer as a group")
	}
	if dst.VNamesOut != names || dst.VNamesIn != nil || src.VNamesOut != nil || dst.VConfigs != configs || dst.PCellStr != "" || src.VConfigs != nil {
		t.Fatal("name or cell group did not transfer as a group")
	}
}

// C checks dst == src only after the groups above; box timing is retained.
func TestTimingTransferSelfMatchesSourceOrder(t *testing.T) {
	g := Gia_ManStart(4)
	g.vCiArrs = vec.Vec_IntAlloc(1)
	g.VInArrs = vec.Vec_FltAlloc(1)
	g.VNamesOut = vec.Vec_PtrAlloc(1)
	g.VConfigs = vec.Vec_IntAlloc(1)
	g.PCellStr = "cell"
	g.PManTime = tim.Tim_ManStart(0, 0)
	g.PAigExtra = Gia_ManStart(1)
	g.vRegClasses, g.vRegInits = vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	g.nAnd2Delay = 8
	timing, extra, classes, inits := g.PManTime, g.PAigExtra, g.vRegClasses, g.vRegInits
	Gia_ManTransferTiming(g, g)
	if g.vCiArrs != nil || g.VInArrs != nil || g.VNamesOut != nil || g.VConfigs != nil || g.PCellStr != "" {
		t.Fatal("self transfer did not clear pre-box metadata")
	}
	if g.PManTime != timing || g.PAigExtra != extra || g.vRegClasses != classes || g.vRegInits != inits || g.nAnd2Delay != 8 {
		t.Fatal("self transfer changed box metadata")
	}
}

func TestTimingTransferBoxGroupRequiresManager(t *testing.T) {
	src, dst := Gia_ManStart(4), Gia_ManStart(4)
	src.PAigExtra, src.vRegClasses, src.vRegInits = Gia_ManStart(1), vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	src.nAnd2Delay = 6
	dst.PManTime, dst.PAigExtra = tim.Tim_ManStart(0, 0), Gia_ManStart(1)
	dst.vRegClasses, dst.vRegInits, dst.nAnd2Delay = vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1), 9
	oldTime, oldExtra, oldClasses, oldInits := dst.PManTime, dst.PAigExtra, dst.vRegClasses, dst.vRegInits
	Gia_ManTransferTiming(dst, src)
	if dst.PManTime != oldTime || dst.PAigExtra != oldExtra || dst.vRegClasses != oldClasses || dst.vRegInits != oldInits || dst.nAnd2Delay != 9 || src.PAigExtra == nil {
		t.Fatal("box metadata moved without a timing manager")
	}
	src.PManTime = tim.Tim_ManStart(0, 0)
	timing, extra, classes, inits := src.PManTime, src.PAigExtra, src.vRegClasses, src.vRegInits
	Gia_ManTransferTiming(dst, src)
	if dst.PManTime != timing || dst.PAigExtra != extra || dst.vRegClasses != classes || dst.vRegInits != inits || dst.nAnd2Delay != 6 {
		t.Fatal("box metadata not transferred")
	}
	if src.PManTime != nil || src.PAigExtra != nil || src.vRegClasses != nil || src.vRegInits != nil || src.nAnd2Delay != 0 {
		t.Fatal("box metadata ownership not cleared")
	}
}
