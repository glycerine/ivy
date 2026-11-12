package cmd

import (
	"bufio"
	"errors"
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type builtin struct {
	name, group string
	handler     Cmd_CommandFuncType
	changes     bool
}

var basicCommands = []builtin{
	{"time", "Basic", CmdCommandTime, false}, {"echo", "Basic", CmdCommandEcho, false}, {"quit", "Basic", CmdCommandQuit, false},
	{"abcrc", "Basic", CmdCommandAbcrc, false}, {"history", "Basic", CmdCommandHistory, false}, {"alias", "Basic", CmdCommandAlias, false},
	{"unalias", "Basic", CmdCommandUnalias, false}, {"help", "Basic", CmdCommandHelp, false}, {"source", "Basic", CmdCommandSource, false},
	{"set", "Basic", CmdCommandSetVariable, false}, {"unset", "Basic", CmdCommandUnsetVariable, false}, {"undo", "Basic", CmdCommandUndo, false},
	{"recall", "Basic", CmdCommandRecall, false}, {"empty", "Basic", CmdCommandEmpty, false}, {"version", "Basic", CmdCommandVersion, false},
}

// Extensions add static built-in definitions; all live registrations still
// reside exclusively in Frame. No component keeps a second command registry.
var extensionInitializers []func(*Abc_Frame_t)

func init() {
	abcmain.Abc_FrameAddInitializer(&abcmain.Abc_FrameInitializer_t{Init: Cmd_Init, Destroy: Cmd_End})
}
func Cmd_Init(f *Abc_Frame_t) {
	if _, installed := f.Dispatcher.(Dispatcher); installed {
		return
	}
	f.SetDispatcher(Dispatcher{})
	for _, definition := range basicCommands {
		entry := definition
		if err := f.RegisterCommand(abcmain.Command{Name: entry.name, Group: entry.group, ChangesNetwork: entry.changes, Handler: func(frame *abcmain.Frame, args []string) int32 { return entry.handler(frame, int32(len(args)), args) }}); err != nil {
			panic(err)
		}
	}
	for _, initialize := range extensionInitializers {
		initialize(f)
	}
	if _, enabled := f.LookupFlag("history_file"); enabled {
		Cmd_HistoryRead(f)
	}
}

const streamResourceName = "cmd:owned-streams"

type commandStreams struct {
	frame                    *Abc_Frame_t
	out, err, history        *os.File
	originalOut, originalErr io.Writer
}

func (s *commandStreams) Close() error {
	var failures []error
	for _, file := range []*os.File{s.out, s.err, s.history} {
		if file != nil {
			if err := file.Close(); err != nil {
				failures = append(failures, err)
			}
		}
	}
	s.frame.Out, s.frame.Err = s.originalOut, s.originalErr
	s.out, s.err, s.history = nil, nil, nil
	return errors.Join(failures...)
}
func streams(f *Abc_Frame_t) *commandStreams {
	if state, ok := f.Resource(streamResourceName).(*commandStreams); ok {
		return state
	}
	state := &commandStreams{frame: f, originalOut: f.Out, originalErr: f.Err}
	f.SetResource(streamResourceName, state)
	return state
}
func Cmd_End(f *Abc_Frame_t) {
	if state, ok := f.TakeResource(streamResourceName).(*commandStreams); ok {
		if err := state.Close(); err != nil {
			fmt.Fprintln(f.Err, err)
		}
	}
	for _, entry := range basicCommands {
		f.RemoveCommand(entry.name)
	}
	if _, ours := f.Dispatcher.(Dispatcher); ours {
		f.SetDispatcher(nil)
	}
}
func commandArgs(count int32, args []string) []string {
	if count < 0 || int(count) > len(args) {
		panic("invalid command argc")
	}
	return args[:count]
}
func usage(f *Abc_Frame_t, text string) int32 { fmt.Fprintln(f.Err, "usage: "+text); return 1 }
func CmdCommandDispatchBasicById(id Cmd_CommandId_t, f *Abc_Frame_t, count int32, args []string, result *int32) int32 {
	var handler Cmd_CommandFuncType
	switch id {
	case CMD_COMMAND_ALIAS:
		handler = CmdCommandAlias
	case CMD_COMMAND_SOURCE:
		handler = CmdCommandSource
	case CMD_COMMAND_SET:
		handler = CmdCommandSetVariable
	default:
		return 0
	}
	*result = handler(f, count, args)
	return 1
}
func CmdCommandTime(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	clear := false
	if len(args) == 2 && args[1] == "-c" {
		clear = true
	} else if len(args) != 1 {
		return usage(f, "time [-c]")
	}
	elapsed, total := f.CommandTiming(true)
	if !clear {
		fmt.Fprintf(f.Out, "elapse: %.2f seconds, total: %.2f seconds\n", elapsed.Seconds(), total.Seconds())
	}
	return 0
}
func CmdCommandEcho(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	newline := true
	start := 1
	if len(args) > 1 && args[1] == "-n" {
		newline = false
		start++
	}
	if len(args) > 1 && args[1] == "-h" {
		return usage(f, "echo [-n] string ...")
	}
	for _, arg := range args[start:] {
		fmt.Fprintf(f.Out, "%s ", arg)
	}
	if newline {
		fmt.Fprintln(f.Out)
	}
	return 0
}
func CmdCommandQuit(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) == 1 {
		return -1
	}
	if len(args) == 2 && args[1] == "-s" {
		return -2
	}
	return usage(f, "quit [-s]")
}
func CmdCommandAbcrc(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) != 1 {
		return usage(f, "abcrc")
	}
	seen := map[string]bool{}
	candidates := []string{"abc.rc", filepath.Join("..", "abc.rc"), filepath.Join("..", "..", "abc.rc")}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "abc.rc"))
	}
	for _, path := range candidates {
		absolute, err := filepath.Abs(path)
		if err != nil || seen[absolute] {
			continue
		}
		seen[absolute] = true
		if _, err := os.Stat(path); err == nil {
			return CmdCommandSource(f, 3, []string{"source", "-s", path})
		}
	}
	return 0
}
func CmdCommandHistory(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	limit := 20
	if len(args) > 2 {
		return usage(f, "history [num]")
	}
	if len(args) == 2 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 0 {
			return usage(f, "history [num]")
		}
		limit = n
	}
	Cmd_HistoryPrint(f, int32(limit))
	return 0
}
func CmdCommandAlias(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) == 1 {
		CmdPrintTable(f, 1)
		return 0
	}
	if args[1] == "-h" {
		return usage(f, "alias [name [value ...]]")
	}
	if len(args) == 2 {
		parts, exists := f.LookupAlias(args[1])
		if !exists {
			fmt.Fprintf(f.Err, "alias %q is not defined\n", args[1])
			return 1
		}
		CmdCommandAliasPrint(f, &Abc_Alias{Name: args[1], Args: parts})
		return 0
	}
	CmdCommandAliasAdd(f, args[1], int32(len(args)-2), args[2:])
	return 0
}
func CmdCommandUnalias(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) < 2 || args[1] == "-h" {
		return usage(f, "unalias name ...")
	}
	for _, name := range args[1:] {
		f.RemoveAlias(name)
	}
	return 0
}
func CmdCommandHelp(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	all, details := int32(0), int32(0)
	for _, arg := range args[1:] {
		switch arg {
		case "-a":
			all ^= 1
		case "-d":
			details ^= 1
		default:
			return usage(f, "help [-a] [-d]")
		}
	}
	CmdCommandPrint(f, all, details)
	return 0
}
func CmdCommandSource(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	silent, echo, prompt, interactive := false, false, false, false
	var path string
	for _, arg := range args[1:] {
		switch arg {
		case "-s":
			silent = !silent
		case "-x", "-e":
			echo = !echo
		case "-p":
			prompt = !prompt
		case "-i":
			interactive = true
		default:
			if strings.HasPrefix(arg, "-") && arg != "-" || path != "" {
				return usage(f, "source [-s] [-x] [-p] [-i] file")
			}
			path = arg
		}
	}
	if path == "" {
		return usage(f, "source [-s] [-x] [-p] [-i] file")
	}
	if err := f.BeginSource(); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	defer f.EndSource()
	var reader io.Reader
	var file *os.File
	realName := path
	if path == "-" {
		reader = f.In
	} else {
		file = CmdFileOpen(f, path, "r", &realName, boolInt(silent))
		if file == nil {
			if silent {
				return 0
			}
			return 1
		}
		defer file.Close()
		reader = file
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	continuation := ""
	for {
		if prompt {
			value, exists := f.LookupFlag("prompt")
			if !exists {
				value = "abc> "
			}
			fmt.Fprint(f.Out, value)
		}
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if echo {
			fmt.Fprintf(f.Out, "abc - > %s\n", line)
		}
		if strings.HasSuffix(line, "\\") {
			continuation += strings.TrimSuffix(line, "\\")
			continue
		}
		line = continuation + line
		continuation = ""
		if interactive {
			Cmd_HistoryAddCommand(f, line)
		}
		if status := executeLine(f, line, false, new(int)); status != 0 {
			if status > 0 {
				fmt.Fprintf(f.Err, "** cmd error: aborting 'source %s'\n", realName)
			}
			return status
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	if continuation != "" {
		return executeLine(f, continuation, false, new(int))
	}
	if interactive {
		return -1
	}
	return 0
}
func boolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func CmdCommandSetVariable(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) == 1 {
		CmdPrintTable(f, 0)
		return 0
	}
	if len(args) > 3 || args[1] == "-h" {
		return usage(f, "set [name [value]]")
	}
	value := ""
	if len(args) == 3 {
		value = args[2]
	}
	if args[1] == "abcout" || args[1] == "abcerr" || args[1] == "history" {
		state := streams(f)
		var file *os.File
		if value != "" && value != "-" {
			file = CmdFileOpen(f, value, "w", nil, 0)
			if file == nil {
				return 1
			}
		}
		switch args[1] {
		case "abcout":
			if state.out != nil {
				state.out.Close()
			}
			state.out = file
			f.Out = state.originalOut
			if file != nil {
				f.Out = file
			}
		case "abcerr":
			if state.err != nil {
				state.err.Close()
			}
			state.err = file
			f.Err = state.originalErr
			if file != nil {
				f.Err = file
			}
		case "history":
			if state.history != nil {
				state.history.Close()
			}
			state.history = file
		}
	}
	f.SetFlag(args[1], value)
	return 0
}
func CmdCommandUnsetVariable(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) < 2 || args[1] == "-h" {
		return usage(f, "unset name ...")
	}
	for _, name := range args[1:] {
		f.DeleteFlag(name)
	}
	return 0
}
func CmdCommandUndo(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) != 1 {
		return usage(f, "undo")
	}
	return CmdCommandRecall(f, 1, []string{"recall"})
}
func CmdCommandRecall(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	step := int32(-1)
	if len(args) > 2 {
		return usage(f, "recall [step]")
	}
	if len(args) == 2 {
		n, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil || n < 0 {
			return usage(f, "recall [step]")
		}
		step = int32(n)
	}
	if f.Network() == nil && step != 0 {
		fmt.Fprintln(f.Out, "Empty network.")
		return 0
	}
	if err := f.RecallNetwork(step); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	return 0
}
func CmdCommandEmpty(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) != 1 {
		return usage(f, "empty")
	}
	f.DeleteNetworks()
	return 0
}
func CmdCommandVersion(f *Abc_Frame_t, count int32, args []string) int32 {
	args = commandArgs(count, args)
	if len(args) != 1 {
		return usage(f, "version")
	}
	fmt.Fprintln(f.Out, f.Version)
	return 0
}
