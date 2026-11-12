package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func shellQuote(value string) string {
	if runtime.GOOS == "windows" {
		return "\"" + strings.ReplaceAll(value, "\"", "\\\"") + "\""
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
func executeShell(f *Abc_Frame_t, text string) int32 {
	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		command = exec.Command("cmd", "/C", text)
	} else {
		if text == "" {
			command = exec.Command("/bin/sh")
		} else {
			command = exec.Command("/bin/sh", "-c", text)
		}
	}
	command.Stdin, command.Stdout, command.Stderr = f.In, f.Out, f.Err
	if err := command.Run(); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	return 0
}
func cmdCheckShellEscape(f *Abc_Frame_t, count int32, args []string) int32 {
	if count <= 0 || len(args) == 0 || !strings.HasPrefix(args[0], "!") {
		return 0
	}
	text := strings.TrimPrefix(args[0], "!")
	for _, arg := range args[1:count] {
		text += " " + shellQuote(arg)
	}
	executeShell(f, text)
	return 1
}
func CmdCommandDispatch(f *Abc_Frame_t, count *int32, args *[]string) int32 {
	if int(*count) > len(*args) || *count < 0 {
		panic("invalid command argc")
	}
	arguments := (*args)[:*count]
	raw := make([]string, len(arguments))
	for i, arg := range arguments {
		raw[i] = shellQuote(arg)
	}
	if len(arguments) > 0 && strings.HasPrefix(arguments[0], "!") {
		raw[0] = strings.TrimPrefix(arguments[0], "!")
		return executeShell(f, strings.Join(raw, " "))
	}
	return executeParsed(f, ParsedCommand{Args: arguments, Raw: strings.Join(raw, " ")}, new(int), false)
}
func CmdSplitLine(f *Abc_Frame_t, text string, count *int32, args *[]string) string {
	command, rest, err := SplitCommand(text)
	if err != nil {
		fmt.Fprintln(f.Err, err)
		*count = -1
		*args = nil
		return ""
	}
	*args = command.Args
	*count = int32(len(command.Args))
	return rest
}
func CmdApplyAlias(f *Abc_Frame_t, count *int32, args *[]string, loops *int32) int32 {
	if *count < 0 || int(*count) > len(*args) {
		return 1
	}
	current := append([]string(nil), (*args)[:*count]...)
	for len(current) > 0 {
		parts, exists := f.LookupAlias(current[0])
		if !exists {
			break
		}
		*loops++
		if *loops > 200 {
			fmt.Fprintln(f.Err, "** cmd warning: alias loop")
			return 1
		}
		commands, err := ParseCommands(strings.Join(parts, " "))
		if err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
		if len(commands) == 0 {
			current = nil
			break
		}
		for _, command := range commands[:len(commands)-1] {
			budget := int(*loops)
			status := executeParsed(f, command, &budget, true)
			*loops = int32(budget)
			if status != 0 {
				return status
			}
		}
		last := commands[len(commands)-1].Args
		self := len(last) > 0 && last[0] == current[0]
		current = append(last, current[1:]...)
		if self {
			break
		}
	}
	*args = current
	*count = int32(len(current))
	return 0
}
func CmdHistorySubstitution(_ *Abc_Frame_t, line string, changed *int32) string {
	*changed = 0
	return line
}
func searchPaths(value string) []string {
	command, _, err := SplitCommand(value)
	if err != nil {
		return strings.Fields(value)
	}
	var paths []string
	for _, field := range command.Args {
		paths = append(paths, filepath.SplitList(field)...)
	}
	return paths
}
func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}
func CmdFileOpen(f *Abc_Frame_t, name, mode string, realName *string, silent int32) *os.File {
	if name == "-" {
		if mode == "w" || mode == "a" {
			if realName != nil {
				*realName = "stdout"
			}
			return os.Stdout
		}
		if realName != nil {
			*realName = "stdin"
		}
		return os.Stdin
	}
	name = expandHome(name)
	candidates := []string{name}
	if mode == "r" && !filepath.IsAbs(name) {
		for _, flag := range []string{"open_path", "lib_path"} {
			if value, ok := f.LookupFlag(flag); ok {
				for _, dir := range searchPaths(value) {
					candidates = append(candidates, filepath.Join(expandHome(dir), name))
				}
			}
		}
	}
	var last error
	for _, candidate := range candidates {
		var file *os.File
		var err error
		switch mode {
		case "r", "rb":
			file, err = os.Open(candidate)
		case "w", "wb":
			file, err = os.Create(candidate)
		case "a", "ab":
			file, err = os.OpenFile(candidate, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		case "r+":
			file, err = os.OpenFile(candidate, os.O_RDWR, 0)
		case "w+":
			file, err = os.OpenFile(candidate, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		default:
			err = fmt.Errorf("unsupported file mode %q", mode)
		}
		if err == nil {
			if realName != nil {
				*realName = candidate
			}
			return file
		}
		last = err
	}
	if realName != nil {
		*realName = name
	}
	if silent == 0 {
		fmt.Fprintln(f.Err, last)
	}
	return nil
}
func CmdFreeArgv(count int32, args []string) {
	if count < 0 || int(count) > len(args) {
		panic("invalid argv size")
	}
	clear(args[:count])
}
func CmdDupArgv(count int32, args []string) []string {
	if count < 0 || int(count) > len(args) {
		panic("invalid argv size")
	}
	return append([]string(nil), args[:count]...)
}
func CmdAddToArgv(count int32, args []string) []string {
	return append([]string{"read"}, CmdDupArgv(count, args)...)
}
func CmdCommandFree(command *Abc_Command) { *command = Abc_Command{} }
func CmdCommandPrint(f *Abc_Frame_t, all, details int32) {
	commands := f.CommandList()
	sort.SliceStable(commands, func(i, j int) bool {
		if commands[i].Group == commands[j].Group {
			return commands[i].Name < commands[j].Name
		}
		return commands[i].Group < commands[j].Group
	})
	group := ""
	for _, command := range commands {
		if all == 0 && strings.HasPrefix(command.Group, "_") {
			continue
		}
		if command.Group != group {
			group = command.Group
			fmt.Fprintf(f.Out, "\n%s commands:\n", group)
		}
		fmt.Fprintf(f.Out, " %-20s", command.Name)
		if details != 0 {
			fmt.Fprintln(f.Out)
			command.Handler(f, []string{command.Name, "-h"})
		}
	}
	fmt.Fprintln(f.Out)
}
func CmdCommandPrintCompare(a, b *Abc_Command) int32 {
	if cmp := strings.Compare(a.Group, b.Group); cmp != 0 {
		return int32(cmp)
	}
	return int32(strings.Compare(a.Name, b.Name))
}
func CmdNamePrintCompare(a, b *string) int32 { return int32(strings.Compare(*a, *b)) }
func CmdPrintTable(f *Abc_Frame_t, aliases int32) {
	if aliases != 0 {
		table := f.AliasList()
		var keys []string
		for key := range table {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			CmdCommandAliasPrint(f, &Abc_Alias{Name: key, Args: table[key]})
		}
		return
	}
	flags := f.FlagList()
	var keys []string
	for key := range flags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(f.Out, "%-20s %s\n", key, flags[key])
	}
}
