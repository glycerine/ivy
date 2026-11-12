/*////////////////////////////////////////////////////////////////////////////

ABC: System for Sequential Synthesis and Verification

http://www.eecs.berkeley.edu/~alanmi/abc/

Copyright (c) The Regents of the University of California. All rights reserved.

Permission is hereby granted, without written agreement and without license or
royalty fees, to use, copy, modify, and distribute this software and its
documentation for any purpose, provided that the above copyright notice and
the following two paragraphs appear in all copies of this software.

IN NO EVENT SHALL THE UNIVERSITY OF CALIFORNIA BE LIABLE TO ANY PARTY FOR
DIRECT, INDIRECT, SPECIAL, INCIDENTAL, OR CONSEQUENTIAL DAMAGES ARISING OUT OF
THE USE OF THIS SOFTWARE AND ITS DOCUMENTATION, EVEN IF THE UNIVERSITY OF
CALIFORNIA HAS BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

THE UNIVERSITY OF CALIFORNIA SPECIFICALLY DISCLAIMS ANY WARRANTIES, INCLUDING,
BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE. THE SOFTWARE PROVIDED HEREUNDER IS ON AN "AS IS" BASIS,
AND THE UNIVERSITY OF CALIFORNIA HAS NO OBLIGATION TO PROVIDE MAINTENANCE,
SUPPORT, UPDATES, ENHANCEMENTS, OR MODIFICATIONS.

////////////////////////////////////////////////////////////////////////////*/

package abcmain

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func strconvQuote(text string) string { return strconv.Quote(text) }
func firstArg(args []string) string {
	if len(args) == 0 {
		return "abc"
	}
	return args[0]
}

// Abc_RealMain is the command interpreter entry point; args includes argv[0].
func Abc_RealMain(args []string) int32 {
	f := Abc_FrameGetGlobalFrame()
	defer Abc_Stop()
	f.Binary = firstArg(args)
	if len(args) == 0 {
		args = []string{f.Binary}
	}
	flags := flag.NewFlagSet(f.Binary, flag.ContinueOnError)
	flags.SetOutput(f.Err)
	var commands []string
	batch, interactiveAfter, quiet := false, false, false
	for _, name := range []string{"c", "q", "C", "Q", "S"} {
		name := name
		flags.Func(name, "commands", func(value string) error {
			commands = append(commands, value)
			batch = true
			quiet = name == "q" || name == "Q"
			interactiveAfter = name == "C" || name == "Q"
			if name == "S" {
				commands[len(commands)-1] = "read_smt " + strconvQuote(value)
			}
			return nil
		})
	}
	flags.Func("f", "script", func(value string) error {
		commands = append(commands, "source "+strconvQuote(value))
		batch = true
		return nil
	})
	flags.Func("F", "echoed script", func(value string) error {
		commands = append(commands, "source -e "+strconvQuote(value))
		batch = true
		return nil
	})
	skipSource := flags.Bool("s", false, "skip startup scripts")
	help := flags.Bool("h", false, "usage")
	output := flags.String("o", "", "output file")
	inputType := flags.String("t", "", "input format")
	outputType := flags.String("T", "", "output format")
	noIO := flags.Bool("x", false, "disable initial/final IO")
	bridge := flags.Bool("b", false, "bridge mode")
	debug := flags.Bool("d", false, "debug output")
	memory := flags.Int64("m", 0, "memory limit in MiB")
	cpu := flags.Int64("l", 0, "CPU time limit in seconds")
	if err := flags.Parse(args[1:]); err != nil {
		return 1
	}
	if *help {
		Abc_UtilsPrintHello(f)
		Abc_UtilsPrintUsage(f, f.Binary)
		return 1
	}
	if *inputType != "" && TypeCheck(f, *inputType) == 0 {
		return 1
	}
	if *outputType != "" && TypeCheck(f, *outputType) == 0 {
		return 1
	}
	if len(flags.Args()) > 1 {
		Abc_UtilsPrintUsage(f, f.Binary)
		return 1
	}
	if *memory < 0 || *cpu < 0 {
		fmt.Fprintln(f.Err, "abcmain: resource limits cannot be negative")
		return 1
	}
	if *memory > 0 || *cpu > 0 {
		if f.Limits == nil {
			fmt.Fprintln(f.Err, "abcmain: OS resource limits require a ResourceLimits adapter")
			return 1
		}
		if *memory > 0 {
			if err := f.Limits.SetMemoryLimit(*memory << 20); err != nil {
				fmt.Fprintln(f.Err, err)
				return 1
			}
		}
		if *cpu > 0 {
			if err := f.Limits.SetCPULimit(time.Duration(*cpu) * time.Second); err != nil {
				fmt.Fprintln(f.Err, err)
				return 1
			}
		}
	}
	if *debug {
		f.SetFlag("debug", "1")
	}
	f.BridgeMode = *bridge
	if *inputType != "" || *outputType != "" || *noIO || len(flags.Args()) != 0 || *output != "" {
		batch = true
	}
	f.BatchMode = batch
	if !*skipSource {
		if status := f.SourceResourceFiles(); status != 0 {
			return status
		}
	}
	if *bridge {
		if err := f.Import("stdin-bridge", f.In); err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
	}
	if !*noIO && *inputType != "none" && len(flags.Args()) == 1 {
		command := "read"
		if *inputType != "" {
			command += "_" + *inputType
		}
		if status := f.Execute(command + " " + strconvQuote(flags.Arg(0))); status != 0 {
			return normalizeExit(status)
		}
	}
	if len(commands) != 0 {
		line := strings.Join(commands, " ; ")
		if !quiet {
			fmt.Fprintf(f.Out, "ABC command line: %q.\n\n", line)
		}
		if status := f.Execute(line); status > 0 {
			return status
		} else if status < 0 && !interactiveAfter {
			return 0
		}
	}
	if !*noIO && *outputType != "none" && *output != "" {
		command := "write"
		if *outputType != "" {
			command += "_" + *outputType
		}
		if status := f.Execute(command + " " + strconvQuote(*output)); status != 0 {
			return normalizeExit(status)
		}
	}
	if batch && !interactiveAfter {
		return 0
	}
	f.BatchMode = false
	Abc_UtilsPrintHello(f)
	for {
		f.mu.RLock()
		step := f.steps
		f.mu.RUnlock()
		fmt.Fprintf(f.Out, "abc %02d> ", step)
		line, err := f.ReadLine()
		if line != "" {
			f.AppendHistory(line)
			status := f.Execute(line)
			if status == -1 || status == -2 {
				return 0
			}
		}
		if err == io.EOF {
			return 0
		}
		if err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
	}
}
func normalizeExit(status int32) int32 {
	if status < 0 {
		return 0
	}
	return status
}
