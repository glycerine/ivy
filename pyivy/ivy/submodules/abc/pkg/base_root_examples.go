package abc

// Stand-alone workflows and source dispositions from src/demo.c,
// src/generic.h, src/starter.c, and src/template.c.

// Inventory and callers:
//
//   - demo.c defines one stand-alone main and locally declares the four ABC
//     entry points it calls. It is built only by the README's external static-
//     library example; no module.make includes it. DemoRewritePlan preserves
//     its command construction, including both synthesis branches.
//   - starter.c defines Abc_UtilStrsav, Abc_RunThread, and a stand-alone main,
//     plus its counter and mutex. It has no repository caller or module.make
//     entry. The maintained command-package implementation is the separate
//     src/base/cmd/cmdStarter.c, registered by src/base/cmd/cmd.c.
//   - generic.h declares no types, values, or functions and has no includers.
//   - template.c defines two placeholder structures and three Xyz_ManTry
//     functions. It has no callers; its body is the scaffold used by the
//     developed src/opt/dau/dauCore.c.

import (
	"bytes"
	"strings"
)

// RootExampleLimits bounds the reusable, side-effect-free parts of the root
// example programs. Zero values select conservative defaults.
type RootExampleLimits struct {
	MaxInputBytes   int
	MaxLineBytes    int
	MaxCommands     int
	MaxCommandBytes int
	Stop            func() bool
}

func normalizeRootExampleLimits(limits RootExampleLimits) (RootExampleLimits, error) {
	if limits.MaxInputBytes < 0 || limits.MaxLineBytes < 0 || limits.MaxCommands < 0 || limits.MaxCommandBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxInputBytes == 0 {
		limits.MaxInputBytes = 1 << 20
	}
	if limits.MaxLineBytes == 0 {
		limits.MaxLineBytes = 999
	}
	if limits.MaxCommands == 0 {
		limits.MaxCommands = 1000
	}
	if limits.MaxCommandBytes == 0 {
		limits.MaxCommandBytes = 999
	}
	return limits, nil
}

func appendRootCommand(commands []string, command string, limits RootExampleLimits) ([]string, error) {
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if len(command) > limits.MaxCommandBytes || len(commands) >= limits.MaxCommands {
		return nil, ErrLimit
	}
	return append(commands, command), nil
}

// DemoRewritePlan returns the command sequence assembled by demo.c. It does
// not execute commands or access files; execution belongs to the command
// package and the embedding application.
func DemoRewritePlan(input string, useResyn2, printStats, verify bool, limits RootExampleLimits) ([]string, error) {
	limits, err := normalizeRootExampleLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	// C argv strings cannot contain an embedded NUL. Reject the additional
	// state admitted by a Go string instead of producing a truncated command.
	if strings.IndexByte(input, 0) >= 0 {
		return nil, ErrDimension
	}
	commands := make([]string, 0, 7)
	for _, command := range []string{"read " + input, "balance"} {
		commands, err = appendRootCommand(commands, command, limits)
		if err != nil {
			return nil, err
		}
	}
	if printStats {
		commands, err = appendRootCommand(commands, "print_stats", limits)
		if err != nil {
			return nil, err
		}
	}
	synthesis := "balance; rewrite -l; rewrite -lz; balance; rewrite -lz; balance"
	if useResyn2 {
		synthesis = "balance; rewrite -l; refactor -l; balance; rewrite -l; rewrite -lz; balance; refactor -lz; rewrite -lz; balance"
	}
	commands, err = appendRootCommand(commands, synthesis, limits)
	if err != nil {
		return nil, err
	}
	if printStats {
		commands, err = appendRootCommand(commands, "print_stats", limits)
		if err != nil {
			return nil, err
		}
	}
	commands, err = appendRootCommand(commands, "write_blif result.blif", limits)
	if err != nil {
		return nil, err
	}
	if verify {
		commands, err = appendRootCommand(commands, "cec "+input+" result.blif", limits)
		if err != nil {
			return nil, err
		}
	}
	return commands, nil
}

// DefaultDemoRewritePlan matches demo.c's compiled defaults.
func DefaultDemoRewritePlan(input string, limits RootExampleLimits) ([]string, error) {
	return DemoRewritePlan(input, false, true, true, limits)
}

// StarterCommandPlan parses the command-file portion of starter.c. Commands
// whose first byte is whitespace or '#' are skipped, as in the source. The
// stand-alone program's system calls and busy-waiting thread launcher are not
// library semantics and are deliberately not reproduced.
func StarterCommandPlan(data []byte, cpus int, limits RootExampleLimits) ([]string, error) {
	limits, err := normalizeRootExampleLimits(limits)
	if err != nil {
		return nil, err
	}
	if cpus <= 0 {
		return nil, ErrDimension
	}
	if len(data) > limits.MaxInputBytes {
		return nil, ErrLimit
	}
	// fgets/system operate on NUL-terminated strings. Embedded NUL input is
	// not a representable command-file line for this checked planner.
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, ErrDimension
	}
	commands := make([]string, 0)
	for len(data) > 0 {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		end := bytes.IndexByte(data, '\n')
		var line []byte
		if end < 0 {
			line, data = data, nil
		} else {
			line, data = data[:end+1], data[end+1:]
		}
		if len(line) > limits.MaxLineBytes {
			return nil, ErrLimit
		}
		if len(line) == 0 || line[0] == '\n' || line[0] == '\r' || line[0] == '\t' || line[0] == ' ' || line[0] == '#' {
			continue
		}
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		commands, err = appendRootCommand(commands, string(line), limits)
		if err != nil {
			return nil, err
		}
	}
	return commands, nil
}

// generic.h is an empty declaration scaffold: it declares no types, values,
// or functions and has no includers. template.c is a three-function scaffold
// with no callers; its developed copy is src/opt/dau/dauCore.c. No Go symbols
// are created for either file because doing so would manufacture an API.
