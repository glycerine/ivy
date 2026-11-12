package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func toolExecutable(f *Abc_Frame_t, name string) (string, error) {
	for _, suffix := range []string{"win", "unix"} {
		candidate := name
		if suffix == "win" {
			candidate += ".exe"
		}
		if value, ok := f.LookupFlag(name + suffix); ok {
			candidate = value
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return exec.LookPath(name)
}

// dispatchIO uses the registered network codec directly. This avoids treating
// internal temporary-file transfers as user commands or creating extra backups.
func dispatchIO(f *Abc_Frame_t, name, filename string) error {
	command, ok := f.LookupCommand(name)
	if !ok {
		return fmt.Errorf("network codec command %q is not registered", name)
	}
	if status := command.Handler(f, []string{name, filename}); status != 0 {
		return fmt.Errorf("%s failed with status %d", name, status)
	}
	return nil
}
func runSynthesisTool(f *Abc_Frame_t, name string, args []string) error {
	if f.Network() == nil {
		return fmt.Errorf("empty network")
	}
	binary, err := toolExecutable(f, name)
	if err != nil {
		return err
	}
	directory, err := os.MkdirTemp("", "abc-"+name+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	input := filepath.Join(directory, "input.blif")
	output := filepath.Join(directory, "output.blif")
	if err := dispatchIO(f, "write_blif", input); err != nil {
		return err
	}
	// SIS and MVSIS have their own command language; the entire script is one
	// process argument, without an intervening shell that could alter its quotes.
	script := "read_blif " + shellQuote(input) + "; " + strings.Join(args, " ") + "; write_blif " + shellQuote(output)
	command := exec.Command(binary, "-x", "-c", script)
	command.Stdin = f.In
	command.Stdout = f.Out
	command.Stderr = f.Err
	if err := command.Run(); err != nil {
		return err
	}
	return dispatchIO(f, "read_blif", output)
}
func synthesisCommand(f *Abc_Frame_t, argc int32, argv []string, name string) int32 {
	args := commandArgs(argc, argv)
	if len(args) < 2 || args[1] == "-h" || args[1] == "-?" {
		return usage(f, name+" command")
	}
	if err := runSynthesisTool(f, name, args[1:]); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	return 0
}
func CmdCommandSis(f *Abc_Frame_t, argc int32, argv []string) int32 {
	return synthesisCommand(f, argc, argv, "sis")
}
func CmdCommandMvsis(f *Abc_Frame_t, argc int32, argv []string) int32 {
	return synthesisCommand(f, argc, argv, "mvsis")
}
func CmdCommandCapo(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	if len(args) == 0 || (len(args) > 1 && (args[1] == "-h" || args[1] == "-?")) {
		return usage(f, "capo options")
	}
	if f.Network() == nil {
		fmt.Fprintln(f.Err, "Empty network.")
		return 1
	}
	binary, err := toolExecutable(f, "capo")
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	file, err := os.CreateTemp("", "abc-capo-*.blif")
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	input := file.Name()
	file.Close()
	defer os.Remove(input)
	if err := dispatchIO(f, "write_blif", input); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	command := exec.Command(binary, append([]string{"-f", input, "-log", "out.txt"}, args[1:]...)...)
	command.Stdin = f.In
	command.Stdout = f.Out
	command.Stderr = f.Err
	if err := command.Run(); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	for i := 1; i < len(args); i++ {
		if args[i] == "-plot" {
			if i+1 == len(args) {
				return usage(f, "capo -plot filename")
			}
			if err := runGnuplot(f, args[i+1]+".plt"); err != nil {
				fmt.Fprintln(f.Err, err)
				return 1
			}
			break
		}
	}
	return 0
}
func runGnuplot(f *Abc_Frame_t, filename string) error {
	binary, err := toolExecutable(f, "gnuplot")
	if err != nil {
		return err
	}
	command := exec.Command(binary, filename)
	command.Stdin = f.In
	command.Stdout = f.Out
	command.Stderr = f.Err
	return command.Run()
}
func Cmd_CommandExecuteGnuplot(f *Abc_Frame_t, filename string) {
	if err := runGnuplot(f, filename); err != nil {
		fmt.Fprintln(f.Err, err)
	}
}
func init() {
	extensionInitializers = append(extensionInitializers, func(f *Abc_Frame_t) {
		Cmd_CommandAdd(f, "Various", "sis", CmdCommandSis, 1)
		Cmd_CommandAdd(f, "Various", "mvsis", CmdCommandMvsis, 1)
		Cmd_CommandAdd(f, "Various", "capo", CmdCommandCapo, 0)
	})
}
