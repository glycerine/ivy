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

func TestScriptFlow2Commands(t *testing.T) {
	prior := abcmain.Abc_FrameReadGlobalFrame()
	defer abcmain.Abc_FrameSetGlobalFrame(prior)
	frame := abcmain.NewFrame()
	recorder := &scriptRecorder{}
	frame.SetDispatcher(recorder)
	abcmain.Abc_FrameSetGlobalFrame(frame)
	Gia_ManPerformFlow2(1, 100001, 0, 4, 8, 1, 1, 1, 0)
	joined := strings.Join(recorder.commands, "\n")
	if !strings.Contains(joined, "&synch2 -K 4 -C 500;  &if -mt -K 4 -C 8") || !strings.Contains(joined, "&sopb -R 10 -C 4") || strings.Count(joined, "&sopb") != 1 || strings.Count(joined, "&load") != 2 {
		t.Fatalf("unexpected flow: %s", joined)
	}
}

func TestFlow3Commands(t *testing.T) {
	for _, tc := range []struct {
		lib, rec, mfs        bool
		first, second, third string
	}{
		{false, false, false, "&if -C 8 -K 4", "&if -g -K 6; &dch -f", "&if -g -K 6; &synch2"},
		{true, true, true, "&if -C 8;       &save", "&if -y -K 6; &dch -f", "&if -y -K 6; &synch2"},
	} {
		commands := giaFlow3Commands(4, 8, tc.mfs, tc.lib, tc.rec)
		for i, fragment := range []string{tc.first, tc.second, tc.third} {
			if !strings.Contains(commands[i], fragment) {
				t.Fatalf("command %d: %q does not contain %q", i, commands[i], fragment)
			}
		}
		if strings.Contains(commands[1], "&mfs;") != tc.mfs || strings.Contains(commands[2], "&mfs;") != tc.mfs {
			t.Fatalf("incorrect MFS option: %q", commands)
		}
	}
}
