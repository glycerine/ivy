package abci

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"strings"
	"testing"
)

func TestRegisteredCommandHelp(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	Abc_Init(frame)
	commands := frame.CommandList()
	if len(commands) < 300 {
		t.Fatalf("only %d commands registered", len(commands))
	}
	for _, command := range commands {
		t.Run(command.Name, func(t *testing.T) {
			defer func() {
				if value := recover(); value != nil {
					t.Fatalf("help panicked: %v", value)
				}
			}()
			previous := util.PrintBridge
			var output strings.Builder
			util.PrintBridge = func(text string) { output.WriteString(text) }
			defer func() { util.PrintBridge = previous }()
			status := command.Handler(frame, []string{command.Name, "-h"})
			if status != 0 && status != 1 && !((command.Name == "quit" || command.Name == "exit") && status == -1) {
				t.Errorf("help returned status %d", status)
			}
			if frame.Network() != nil {
				t.Error("help created a network")
			}
		})
	}
}
