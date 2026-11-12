package io

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestWriteFormatCommands(t *testing.T) {
	for _, tc := range []struct {
		name    string
		kind    Io_FileType_t
		command func(*abcmain.Frame, []string) int32
	}{
		{"baf", IO_FILE_BAF, IoCommandWriteBaf}, {"bblif", IO_FILE_BBLIF, IoCommandWriteBblif}, {"blif_mv", IO_FILE_BLIFMV, IoCommandWriteBlifMv}, {"dot", IO_FILE_DOT, IoCommandWriteDot}, {"eqn", IO_FILE_EQN, IoCommandWriteEqn}, {"gml", IO_FILE_GML, IoCommandWriteGml}, {"smv", IO_FILE_SMV, IoCommandWriteSmv},
	} {
		t.Run(tc.name, func(t *testing.T) {
			formats.Lock()
			old := formats.handlers[tc.kind]
			formats.Unlock()
			defer func() { formats.Lock(); formats.handlers[tc.kind] = old; formats.Unlock() }()
			called := false
			RegisterFormat(tc.kind, nil, func(n *abc.Network, filename string) error {
				called = true
				if n == nil || filename != "output.any" {
					return fmt.Errorf("unexpected args")
				}
				return nil
			})
			frame := abcmain.NewFrame()
			var errors bytes.Buffer
			frame.Err = &errors
			frame.ReplaceCurrentNetwork(abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0))
			if tc.command(frame, []string{"write_" + tc.name, "output.any"}) != 0 || !called {
				t.Fatalf("write failed: %s", errors.String())
			}
			if tc.command(frame, []string{"write_" + tc.name, "-h"}) != 1 {
				t.Fatal("help accepted as file")
			}
		})
	}
}
