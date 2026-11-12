package io

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestReadFormatCommands(t *testing.T) {
	for _, tc := range []struct {
		name    string
		kind    Io_FileType_t
		command func(*abcmain.Frame, []string) int32
	}{
		{"aiger", IO_FILE_AIGER, IoCommandReadAiger}, {"baf", IO_FILE_BAF, IoCommandReadBaf}, {"bblif", IO_FILE_BBLIF, IoCommandReadBblif}, {"blif_mv", IO_FILE_BLIFMV, IoCommandReadBlifMv}, {"bench", IO_FILE_BENCH, IoCommandReadBench}, {"edif", IO_FILE_EDIF, IoCommandReadEdif}, {"eqn", IO_FILE_EQN, IoCommandReadEqn},
	} {
		t.Run(tc.name, func(t *testing.T) {
			formats.Lock()
			old := formats.handlers[tc.kind]
			formats.Unlock()
			defer func() { formats.Lock(); formats.handlers[tc.kind] = old; formats.Unlock() }()
			called := false
			RegisterFormat(tc.kind, func(filename string, check bool) (*abc.Network, error) {
				called = true
				if filename != "source.any" || check {
					return nil, fmt.Errorf("unexpected args")
				}
				return abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0), nil
			}, nil)
			frame := abcmain.NewFrame()
			var errors bytes.Buffer
			frame.Err = &errors
			if tc.command(frame, []string{"read_" + tc.name, "-c", "source.any"}) != 0 || !called || frame.Network() == nil {
				t.Fatalf("read failed: %s", errors.String())
			}
			if tc.command(frame, []string{"read_" + tc.name, "-h"}) != 1 {
				t.Fatal("help accepted as file")
			}
		})
	}
}
