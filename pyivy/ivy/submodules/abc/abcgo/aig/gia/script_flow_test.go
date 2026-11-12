package gia

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"strings"
	"testing"
)

type scriptRecorder struct{ commands []string }

func (r *scriptRecorder) Execute(_ *abcmain.Frame, line string) int32 {
	r.commands = append(r.commands, line)
	return 0
}
func TestScriptMappingCommands(t *testing.T) {
	prior := abcmain.Abc_FrameReadGlobalFrame()
	defer abcmain.Abc_FrameSetGlobalFrame(prior)
	frame := abcmain.NewFrame()
	recorder := &scriptRecorder{}
	frame.SetDispatcher(recorder)
	abcmain.Abc_FrameSetGlobalFrame(frame)
	Gia_ManPerformMap(100, 4, 8, 1, 1, 0)
	if len(recorder.commands) != 6 || !strings.Contains(recorder.commands[0], "-k -t") || !strings.Contains(recorder.commands[2], "-S 44 -K 7 -C 16 -t") || recorder.commands[5] != "&put; mfs2 -W 4 -M 500 -C 7000; &get -m" {
		t.Fatalf("commands %q", recorder.commands)
	}
}
