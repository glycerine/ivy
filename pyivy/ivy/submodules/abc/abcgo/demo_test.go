package abcgo

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func demoFrame(t *testing.T, failAt int) (*abcmain.Frame, *[][]string, *bytes.Buffer) {
	t.Helper()
	frame := abcmain.NewFrame()
	cmd.Cmd_Init(frame)
	var calls [][]string
	output := new(bytes.Buffer)
	frame.Out, frame.Err = output, output
	for _, name := range []string{"read", "balance", "rewrite", "refactor", "print_stats", "write_blif", "cec"} {
		if err := frame.RegisterCommand(abcmain.Command{Name: name, Handler: func(_ *abcmain.Frame, args []string) int32 {
			calls = append(calls, append([]string(nil), args...))
			if failAt > 0 && len(calls) == failAt {
				return 1
			}
			return 0
		}}); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { frame.Close() })
	return frame, &calls, output
}

func TestDemoDefaultWorkflow(t *testing.T) {
	frame, calls, output := demoFrame(t, 0)
	filename := "a b;'quote\\.blif"
	if code := RunDemo(frame, filename, DefaultDemoOptions()); code != 0 {
		t.Fatal(code, output.String())
	}
	want := [][]string{{"read", filename}, {"balance"}, {"print_stats"}, {"balance"}, {"rewrite", "-l"}, {"rewrite", "-lz"}, {"balance"}, {"rewrite", "-lz"}, {"balance"}, {"print_stats"}, {"write_blif", "result.blif"}, {"cec", filename, "result.blif"}}
	if !reflect.DeepEqual(*calls, want) {
		t.Fatalf("calls %q; want %q", *calls, want)
	}
	if !strings.Contains(output.String(), "Reading = ") || !strings.Contains(output.String(), "Verification = ") {
		t.Fatal(output.String())
	}
}

func TestDemoResyn2WithoutOptionalStages(t *testing.T) {
	frame, calls, output := demoFrame(t, 0)
	if RunDemo(frame, "input.blif", DemoOptions{UseResyn2: true}) != 0 {
		t.Fatal(output.String())
	}
	var names []string
	for _, args := range *calls {
		names = append(names, strings.Join(args, " "))
	}
	want := []string{"read input.blif", "balance", "balance", "rewrite -l", "refactor -l", "balance", "rewrite -l", "rewrite -lz", "balance", "refactor -lz", "rewrite -lz", "balance", "write_blif result.blif"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("%q", names)
	}
}

func TestDemoStopsAtEveryFailure(t *testing.T) {
	for fail := 1; fail <= 12; fail++ {
		frame, calls, output := demoFrame(t, fail)
		if RunDemo(frame, "input.blif", DefaultDemoOptions()) != 1 {
			t.Fatalf("failure %d ignored", fail)
		}
		if len(*calls) != fail {
			t.Fatalf("failure %d: %d calls", fail, len(*calls))
		}
		if !strings.Contains(output.String(), "Cannot execute command") || strings.Contains(output.String(), "Reading = ") {
			t.Fatal(output.String())
		}
	}
}

type demoCloser struct{ count int }

func (c *demoCloser) Close() error { c.count++; return nil }

func TestDemoMainClosesFrameworkOnSuccessAndFailure(t *testing.T) {
	for _, fail := range []int{0, 1, 6, 12} {
		frame, _, output := demoFrame(t, fail)
		closer := new(demoCloser)
		frame.SetResource("demo-test", closer)
		abcmain.Abc_FrameSetGlobalFrame(frame)
		code := DemoMain([]string{"input.blif"}, output, output)
		if (code == 0) != (fail == 0) {
			t.Fatalf("fail %d returned %d", fail, code)
		}
		if closer.count != 1 || abcmain.Abc_FrameReadGlobalFrame() != nil {
			t.Fatalf("cleanup fail %d, count %d", fail, closer.count)
		}
	}
}

func TestDemoMainRejectsArgumentsBeforeStarting(t *testing.T) {
	if abcmain.Abc_FrameReadGlobalFrame() != nil {
		t.Fatal("unexpected global frame")
	}
	for _, args := range [][]string{nil, {"a", "b"}} {
		var out bytes.Buffer
		if DemoMain(args, &out, &out) != 1 || out.String() != "Wrong number of command-line arguments.\n" {
			t.Fatal(out.String())
		}
		if abcmain.Abc_FrameReadGlobalFrame() != nil {
			t.Fatal("started on invalid arguments")
		}
	}
}

// Argument quoting is independent of the host shell and must survive the ABC
// parser itself, including an empty name and literal line breaks.
func TestDemoFilenameRemainsSingleArgument(t *testing.T) {
	for _, filename := range []string{"", "plain.blif", "with spaces.blif", "semi;colon", "hash#name", "single'quote", `double"quote`, `back\slash`, "line\nbreak", "tab\tname"} {
		frame, calls, output := demoFrame(t, 0)
		if RunDemo(frame, filename, DefaultDemoOptions()) != 0 {
			t.Fatal(filename, output.String())
		}
		if got := (*calls)[0]; !reflect.DeepEqual(got, []string{"read", filename}) {
			t.Fatalf("read %q: %q", filename, got)
		}
		if got := (*calls)[len(*calls)-1]; !reflect.DeepEqual(got, []string{"cec", filename, "result.blif"}) {
			t.Fatalf("cec %q: %q", filename, got)
		}
	}
}
