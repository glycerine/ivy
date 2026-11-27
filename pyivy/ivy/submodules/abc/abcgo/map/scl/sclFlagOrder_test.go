package scl

import (
	"bytes"
	"strings"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
)

// With several non-integer values the integer options are checked in a
// fixed order, so the same option is always the one reported.
func TestIntegerOptionErrorIsDeterministic(t *testing.T) {
	for _, c := range []struct{ command, want string }{
		{"upsize -I 1.5 -J 2.5 -W 3.5 -R 4.5 -N 5.5 -B 6.5", "-I requires an integer"},
		{"dnsize -I 1.5 -J 2.5 -N 3.5 -D 4.5 -X 5.5", "-I requires an integer"},
	} {
		for run := 0; run < 20; run++ {
			f := abcmain.NewFrame()
			var errs bytes.Buffer
			f.Out, f.Err = &bytes.Buffer{}, &errs
			Scl_Init(f)
			if f.Execute(c.command) == 0 {
				t.Fatalf("%s accepted", c.command)
			}
			f.Close()
			if !strings.Contains(errs.String(), c.want) {
				t.Fatalf("run %d: %s reported %q, want %q", run, c.command, errs.String(), c.want)
			}
		}
	}
}
