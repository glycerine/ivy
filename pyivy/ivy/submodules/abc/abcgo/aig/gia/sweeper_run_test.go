package gia

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"testing"
)

type flipSweepOutput struct{}

func (flipSweepOutput) Execute(f *abcmain.Frame, command string) int32 {
	graph := abcmain.Abc_FrameReadGia(f).(*Gia_Man_t)
	Gia_ObjFlipFaninC0(Gia_ManPo(graph, 0))
	return 0
}
func TestSweeperRunAndCleanupCommands(t *testing.T) {
	prior := abcmain.Abc_FrameReadGlobalFrame()
	defer abcmain.Abc_FrameSetGlobalFrame(prior)
	frame := abcmain.NewFrame()
	frame.SetDispatcher(flipSweepOutput{})
	abcmain.Abc_FrameSetGlobalFrame(frame)
	p := Gia_SweeperStart(nil)
	a := Gia_ManAppendCi(p)
	id := Gia_SweeperProbeCreate(p, a)
	ids := Vec_IntAlloc(1)
	Vec_IntPush(ids, id)
	if Gia_SweeperRun(p, ids, "flip", 0) != 1 || Gia_SweeperProbeLit(p, id) != a^1 {
		t.Fatal("command output was not grafted")
	}
	p = Gia_SweeperCleanup(p, "flip")
	defer Gia_SweeperStop(p)
	if Gia_SweeperProbeLit(p, id) != a {
		t.Fatal("cleanup command did not preserve modified probe")
	}
}
