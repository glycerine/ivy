package io

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"testing"
)

func TestIOCommandRegistration(t *testing.T) {
	frame := abcmain.NewFrame()
	Io_Init(frame)
	for _, name := range []string{"read_blif", "write_blif", "read_aiger", "write_cex", "&write_cnf", "write_status"} {
		if _, ok := frame.LookupCommand(name); !ok {
			t.Fatalf("command %q missing", name)
		}
	}
}
