package super

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"os"
	"strconv"
	"strings"
	"time"
)

func Super_Init(frame *abcmain.Frame) {
	for _, entry := range []struct {
		name    string
		handler func(*abcmain.Frame, int32, []string) int32
	}{{"super", Super_CommandSupergates}, {"super2", Super_CommandSupergatesAnd}} {
		handler := entry.handler
		_ = frame.RegisterCommand(abcmain.Command{Name: entry.name, Group: "SC mapping", Handler: func(f *abcmain.Frame, args []string) int32 { return handler(f, int32(len(args)), args) }})
	}
}
func Super_End(frame *abcmain.Frame) { frame.RemoveCommand("super"); frame.RemoveCommand("super2") }

type commandOptions struct {
	numbers   map[byte]string
	flags     map[byte]bool
	arguments []string
}

func parseCommand(args []string, numeric, flags string) (commandOptions, error) {
	result := commandOptions{numbers: make(map[byte]string), flags: make(map[byte]bool)}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			result.arguments = append(result.arguments, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			result.arguments = append(result.arguments, arg)
			continue
		}
		for j := 1; j < len(arg); j++ {
			flag := arg[j]
			if flag == 'h' {
				return result, fmt.Errorf("help requested")
			}
			if strings.IndexByte(flags, flag) >= 0 {
				result.flags[flag] = !result.flags[flag]
				continue
			}
			if strings.IndexByte(numeric, flag) < 0 {
				return result, fmt.Errorf("unknown option -%c", flag)
			}
			value := arg[j+1:]
			if value == "" {
				i++
				if i == len(args) {
					return result, fmt.Errorf("missing value for -%c", flag)
				}
				value = args[i]
			}
			result.numbers[flag] = value
			break
		}
	}
	return result, nil
}
func commandError(frame *abcmain.Frame, err error, usage string) int32 {
	fmt.Fprintln(frame.Err, err)
	fmt.Fprintln(frame.Err, usage)
	return 1
}
func commandInt(o commandOptions, key byte, value int) (int, error) {
	s, ok := o.numbers[key]
	if !ok {
		return value, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || int64(n) > 1<<31-1 {
		return 0, fmt.Errorf("invalid nonnegative -%c value", key)
	}
	return n, nil
}
func Super_CommandSupergatesAnd(frame *abcmain.Frame, argc int32, args []string) int32 {
	const usage = "usage: super2 [-I inputs] [-L levels] [-v]"
	o, err := parseCommand(args, "IL", "v")
	if err != nil {
		return commandError(frame, err, usage)
	}
	if len(o.arguments) != 0 {
		return commandError(frame, fmt.Errorf("unexpected filename"), usage)
	}
	inputs, err := commandInt(o, 'I', 4)
	if err != nil {
		return commandError(frame, err, usage)
	}
	levels, err := commandInt(o, 'L', 3)
	if err != nil {
		return commandError(frame, err, usage)
	}
	lib, err := GenerateAND(inputs, levels)
	if err != nil {
		return commandError(frame, err, usage)
	}
	if err = Super2_LibWrite(lib); err != nil {
		return commandError(frame, err, usage)
	}
	if o.flags['v'] {
		fmt.Fprintf(frame.Out, "Computed %d AND/inverter supergates\n", len(lib.Gates))
	}
	return 0
}
func Super_CommandSupergates(frame *abcmain.Frame, argc int32, args []string) int32 {
	const usage = "usage: super [-ILNT num] [-DA float] [-E file] [-sov] genlib_file"
	o, err := parseCommand(args, "ILNTDAE", "sov")
	if err != nil {
		return commandError(frame, err, usage)
	}
	if len(o.arguments) != 1 {
		return commandError(frame, fmt.Errorf("exactly one genlib file is required"), usage)
	}
	options := DefaultOptions()
	for _, entry := range []struct {
		key    byte
		target *int
	}{{'I', &options.Inputs}, {'L', &options.Levels}, {'N', &options.MaxGates}} {
		*entry.target, err = commandInt(o, entry.key, *entry.target)
		if err != nil {
			return commandError(frame, err, usage)
		}
	}
	seconds, err := commandInt(o, 'T', 0)
	if err != nil {
		return commandError(frame, err, usage)
	}
	options.TimeLimit = time.Duration(seconds) * time.Second
	for _, entry := range []struct {
		key    byte
		target *float32
	}{{'D', &options.DelayLimit}, {'A', &options.AreaLimit}} {
		if value, ok := o.numbers[entry.key]; ok {
			parsed, e := strconv.ParseFloat(value, 32)
			if e != nil || !finite(float32(parsed)) || parsed <= 0 {
				return commandError(frame, fmt.Errorf("invalid positive -%c value", entry.key), usage)
			}
			*entry.target = float32(parsed)
		}
	}
	if options.Inputs < 2 || options.Inputs > 6 {
		return commandError(frame, fmt.Errorf("super input count must be 2..6"), usage)
	}
	options.SkipInputInverters = !o.flags['s']
	if o.flags['v'] {
		options.Progress = frame.Out
	}
	exclude := o.numbers['E']
	if exclude != "" {
		if _, err = os.ReadFile(exclude); err != nil {
			return commandError(frame, err, usage)
		}
	}
	lib := mio.Mio_LibraryRead(o.arguments[0], "", exclude, 0)
	if lib == nil {
		return commandError(frame, fmt.Errorf("cannot read genlib %s", o.arguments[0]), usage)
	}
	defer mio.Mio_LibraryDelete(lib)
	manager, err := Generate(lib, options)
	if err != nil {
		return commandError(frame, err, usage)
	}
	data, err := manager.MarshalTree()
	if err != nil {
		return commandError(frame, err, usage)
	}
	if err = os.WriteFile(outputBase(manager.Name)+".super", data, 0644); err != nil {
		return commandError(frame, err, usage)
	}
	if o.flags['o'] {
		if err = Super_WriteLibrary(manager); err != nil {
			return commandError(frame, err, usage)
		}
	}
	return 0
}
