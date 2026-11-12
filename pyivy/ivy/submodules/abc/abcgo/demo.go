// Package abcgo exposes ABC's library demonstration workflow.
package abcgo

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// DemoOptions selects the same optional stages as src/demo.c.
type DemoOptions struct{ UseResyn2, PrintStats, Verify bool }

// DefaultDemoOptions matches the original executable's settings.
func DefaultDemoOptions() DemoOptions { return DemoOptions{PrintStats: true, Verify: true} }

// RunDemo runs the original read/balance/rewrite/write/verify command sequence.
// The caller owns frame and must initialize its command modules and close it.
// Synthesis and file-format modules must register their actual commands; missing
// commands are reported by the interpreter rather than simulated here.
func RunDemo(frame *abcmain.Frame, filename string, options DemoOptions) int {
	run := func(command string) bool {
		if cmd.Cmd_CommandExecute(frame, command) == 0 {
			return true
		}
		fmt.Fprintf(frame.Out, "Cannot execute command \"%s\".\n", command)
		return false
	}
	input := demoQuote(filename)
	start := time.Now()
	if !run("read "+input) || !run("balance") {
		return 1
	}
	reading := time.Since(start)
	if options.PrintStats && !run("print_stats") {
		return 1
	}
	start = time.Now()
	sequence := "balance; rewrite -l; rewrite -lz; balance; rewrite -lz; balance"
	if options.UseResyn2 {
		sequence = "balance; rewrite -l; refactor -l; balance; rewrite -l; rewrite -lz; balance; refactor -lz; rewrite -lz; balance"
	}
	if !run(sequence) {
		return 1
	}
	rewriting := time.Since(start)
	if options.PrintStats && !run("print_stats") {
		return 1
	}
	if !run("write_blif result.blif") {
		return 1
	}
	start = time.Now()
	if options.Verify && !run("cec "+input+" result.blif") {
		return 1
	}
	verification := time.Since(start)
	fmt.Fprintf(frame.Out, "Reading = %6.2f sec   Rewriting = %6.2f sec   Verification = %6.2f sec\n", reading.Seconds(), rewriting.Seconds(), verification.Seconds())
	return 0
}

// DemoMain is the CLI entry path. Args excludes the executable name. Unlike the
// C demo's early error returns, every started framework is shut down on failure.
func DemoMain(args []string, out, errOut io.Writer) int {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	if len(args) != 1 {
		fmt.Fprintln(out, "Wrong number of command-line arguments.")
		return 1
	}
	abcmain.Abc_Start()
	defer abcmain.Abc_Stop()
	frame := abcmain.Abc_FrameGetGlobalFrame()
	frame.Out, frame.Err = out, errOut
	cmd.Cmd_Init(frame)
	return RunDemo(frame, args[0], DefaultDemoOptions())
}

// Quote for ABC's command parser, not for an operating-system shell.
func demoQuote(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\r\n;#\\'\"") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
