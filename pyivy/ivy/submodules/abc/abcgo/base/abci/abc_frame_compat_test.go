package abci

import (
	"testing"

	giapkg "github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestFrameGiaUpdateTransfersNamesAndBackup(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	old := giapkg.Gia_ManStart(4)
	ci := giapkg.Gia_ManAppendCi(old)
	giapkg.Gia_ManAppendCo(old, ci)
	inNames, outNames := vec.Vec_PtrAlloc(1), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(inNames, "input")
	vec.Vec_PtrPush(outNames, "output")
	giapkg.Gia_ManSetInputNames(old, inNames)
	giapkg.Gia_ManSetOutputNames(old, outNames)
	Abc_FrameUpdateGia(frame, old)
	if abcFrameGia(frame) != old {
		t.Fatal("initial GIA was not retained by the frame")
	}

	next := giapkg.Gia_ManStart(4)
	ci = giapkg.Gia_ManAppendCi(next)
	giapkg.Gia_ManAppendCo(next, ci)
	Abc_FrameUpdateGia(frame, next)
	if abcFrameGia(frame) != next || abcFrameResource[*Gia_Man_t](frame, "pGia2") != old {
		t.Fatal("GIA update did not move the old manager into the backup slot")
	}
	if giapkg.Gia_ManInputNames(next) != inNames || giapkg.Gia_ManOutputNames(next) != outNames {
		t.Fatal("GIA update did not transfer compatible terminal names")
	}
	if giapkg.Gia_ManInputNames(old) != nil || giapkg.Gia_ManOutputNames(old) != nil {
		t.Fatal("name ownership was not cleared from the previous manager")
	}
	got := Abc_FrameGetGia(frame)
	if got != next {
		t.Fatal("frame GIA take returned the wrong manager")
	}
	giapkg.Gia_ManStop(got)
	if abcFrameGia(frame) != nil || abcFrameResource[*Gia_Man_t](frame, "pGia2") != nil {
		t.Fatal("frame GIA take did not clear current and backup slots")
	}
}

func TestFrameOpaqueCexAndVerificationState(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	cex := &util.Abc_Cex_t{NPis: 1, NBits: 1, PData: []uint32{1}}
	abcFrameSetCex(frame, cex)
	if abcFrameCex(frame) != cex {
		t.Fatal("counterexample state did not retain the supplied pointer")
	}
	abcFrameSetStatus(frame, 1)
	abcFrameSetNFrames(frame, 7)
	if abcFrameStatus(frame) != 1 || abcFrameNFrames(frame) != 7 {
		t.Fatalf("verification state = (%d,%d), want (1,7)", abcFrameStatus(frame), abcFrameNFrames(frame))
	}
	abcFrameSetCex(frame, nil)
	if abcFrameCex(frame) != nil {
		t.Fatal("counterexample clear did not transfer nil into frame state")
	}
}
