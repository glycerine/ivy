package cmd

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"math"
	"strconv"
	"strings"
)

func Cmd_CommandNameToId(name string) Cmd_CommandId_t {
	switch name {
	case "alias":
		return CMD_COMMAND_ALIAS
	case "source":
		return CMD_COMMAND_SOURCE
	case "set":
		return CMD_COMMAND_SET
	case "read_blif":
		return CMD_COMMAND_READ_BLIF
	case "read_aiger":
		return CMD_COMMAND_READ_AIGER
	case "strash":
		return CMD_COMMAND_STRASH
	case "print_stats":
		return CMD_COMMAND_PRINT_STATS
	case "pdr":
		return CMD_COMMAND_PDR
	case "write_aiger_cex":
		return CMD_COMMAND_WRITE_AIGER_CEX
	}
	return CMD_COMMAND_UNKNOWN
}
func Cmd_CommandIsDefined(f *Abc_Frame_t, name string) int32 {
	_, exists := f.LookupCommand(name)
	if exists {
		return 1
	}
	return 0
}
func Cmd_CommandAdd(f *Abc_Frame_t, group, name string, handler Cmd_CommandFuncType, changes int32) {
	if handler == nil {
		panic("nil command handler")
	}
	if _, exists := f.LookupCommand(name); exists {
		fmt.Fprintf(f.Err, "Cmd warning: redefining '%s'\n", name)
	}
	if err := f.RegisterCommand(abcmain.Command{Name: name, Group: group, ChangesNetwork: changes != 0, Handler: func(frame *abcmain.Frame, args []string) int32 { return handler(frame, int32(len(args)), args) }}); err != nil {
		panic(err)
	}
}

// StatisticsProvider is implemented by mapped-network adapters without making
// the command interpreter depend on concrete network or gate-library packages.
type StatisticsProvider interface {
	CommandStatistics() (inputs, outputs, nodes int, area float64, depth int)
}

func Cmd_CommandHandleSpecial(f *Abc_Frame_t, text string) int32 {
	pi, po, nodes, depth := 0, 0, 0, 0
	area := 0.0
	if provider, ok := f.Network().(StatisticsProvider); ok {
		pi, po, nodes, area, depth = provider.CommandStatistics()
	}
	if strings.Contains(text, "#PS") {
		fmt.Fprintf(f.Out, "pi=%d po=%d fn=%d ma=%.1f de=%d\n", pi, po, nodes, area, depth)
		return 1
	}
	if strings.Contains(text, "#CEC") {
		fmt.Fprintf(f.Out, "proofStatus=%d\n", f.Status)
		return 1
	}
	if !strings.Contains(text, "#ASSERT") {
		return 0
	}
	at := strings.LastIndexByte(text, '=')
	if at < 0 {
		return 0
	}
	want, err := strconv.ParseFloat(strings.TrimSpace(text[at+1:]), 64)
	if err != nil {
		return 0
	}
	got := 0.0
	epsilon := 0.0
	switch {
	case strings.Contains(text, "_PI_"):
		got = float64(pi)
	case strings.Contains(text, "_PO_"):
		got = float64(po)
	case strings.Contains(text, "_NODE_"):
		got = float64(nodes)
	case strings.Contains(text, "_AREA_"):
		got = area
		epsilon = 1
	case strings.Contains(text, "_DEPTH_"):
		got = float64(depth)
	default:
		return 0
	}
	passed := false
	if strings.Contains(text, "==") {
		passed = math.Abs(got-want) <= epsilon
	} else if strings.Contains(text, "<=") {
		passed = got <= want+epsilon
	} else {
		return 0
	}
	if passed {
		fmt.Fprintln(f.Out, "succeeded")
	} else {
		fmt.Fprintln(f.Out, "failed")
	}
	return 1
}

// Dispatcher installs the complete command language on Frame's shared registry.
type Dispatcher struct{}

func (Dispatcher) Execute(f *abcmain.Frame, text string) int32 {
	return executeLine(f, text, true, new(int))
}
func Cmd_CommandExecute(f *Abc_Frame_t, text string) int32 {
	return executeLine(f, text, true, new(int))
}
func executeLine(f *Abc_Frame_t, text string, history bool, expansions *int) int32 {
	if history && f.SourceDepth() == 0 {
		Cmd_HistoryAddCommand(f, text)
		if state, ok := f.Resource("cmd:owned-streams").(*commandStreams); ok && state.history != nil {
			fmt.Fprintln(state.history, text)
		}
	}
	for text != "" {
		if strings.HasPrefix(strings.TrimSpace(text), "#") && Cmd_CommandHandleSpecial(f, strings.TrimSpace(text)) != 0 {
			return 0
		}
		command, rest, err := SplitCommand(text)
		if err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
		text = rest
		if len(command.Args) == 0 {
			continue
		}
		if status := executeParsed(f, command, expansions, true); status != 0 {
			return status
		}
	}
	return 0
}
func executeParsed(f *Abc_Frame_t, command ParsedCommand, expansions *int, aliases bool) int32 {
	if aliases {
		if parts, ok := f.LookupAlias(command.Args[0]); ok {
			*expansions++
			if *expansions > 200 {
				fmt.Fprintln(f.Err, "** cmd warning: alias loop")
				return 1
			}
			commands, err := ParseCommands(strings.Join(parts, " "))
			if err != nil {
				fmt.Fprintln(f.Err, err)
				return 1
			}
			if len(commands) == 0 {
				return 0
			}
			last := len(commands) - 1
			commands[last].Args = append(commands[last].Args, command.Args[1:]...)
			for _, arg := range command.Args[1:] {
				commands[last].Raw += " " + shellQuote(arg)
			}
			for i, next := range commands {
				allow := !(i == 0 && len(next.Args) > 0 && next.Args[0] == command.Args[0])
				if status := executeParsed(f, next, expansions, allow); status != 0 {
					return status
				}
			}
			return 0
		}
	}
	if strings.HasPrefix(command.Args[0], "!") {
		return executeShell(f, strings.TrimSpace(strings.TrimPrefix(command.Raw, "!")))
	}
	args := command.Args
	if _, exists := f.LookupCommand(args[0]); !exists && len(args) == 1 && strings.Contains(args[0], ".") {
		args = append([]string{"read"}, args...)
	}
	status := f.Dispatch(args)
	if status == 0 {
		if automatic, exists := f.LookupFlag("autoexec"); exists && automatic != "" && f.EnterAutoexec() {
			status = executeLine(f, automatic, false, new(int))
			f.LeaveAutoexec()
		}
	}
	return status
}
