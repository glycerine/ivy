package abcmain

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// SplitCommands handles quoting, escaped characters, comments and command
// separators. Component command packages may install a richer Dispatcher while
// retaining the same frame-owned registry and Dispatch implementation.
func SplitCommands(line string) ([][]string, error) {
	var result [][]string
	var args []string
	var word strings.Builder
	started := false
	var quote rune
	escape := false
	flushWord := func() {
		if started {
			args = append(args, word.String())
			word.Reset()
			started = false
		}
	}
	flushCommand := func() {
		flushWord()
		if len(args) != 0 {
			result = append(result, args)
			args = nil
		}
	}
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if escape {
			if c != '\n' {
				word.WriteRune(c)
				started = true
			}
			escape = false
			continue
		}
		if c == '\\' && quote != '\'' {
			escape = true
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
			} else {
				word.WriteRune(c)
			}
			started = true
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
			started = true
		case ' ', '\t', '\r':
			flushWord()
		case ';', '\n':
			flushCommand()
		case '#':
			if !started {
				for i < len(runes) && runes[i] != '\n' {
					i++
				}
				flushCommand()
			} else {
				word.WriteRune(c)
			}
		default:
			word.WriteRune(c)
			started = true
		}
	}
	if escape {
		return nil, errors.New("abcmain: trailing command escape")
	}
	if quote != 0 {
		return nil, errors.New("abcmain: unmatched command quote")
	}
	flushCommand()
	return result, nil
}
func executeDefault(frame *Frame, line string, depth int) int32 {
	if depth > 20 {
		fmt.Fprintln(frame.Err, "abcmain: alias expansion limit exceeded")
		return 1
	}
	commands, err := SplitCommands(line)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	for _, args := range commands {
		aliasDepth := depth
		for {
			alias, ok := frame.LookupAlias(args[0])
			if !ok {
				break
			}
			if aliasDepth >= 20 {
				fmt.Fprintln(frame.Err, "abcmain: alias expansion limit exceeded")
				return 1
			}
			aliasDepth++
			if len(alias) == 0 {
				fmt.Fprintln(frame.Err, "abcmain: empty alias")
				return 1
			}
			args = append(alias, args[1:]...)
		}
		status := frame.Dispatch(args)
		if status != 0 {
			return status
		}
	}
	return 0
}
func registerBuiltins(frame *Frame) {
	add := func(name string, handler func(*Frame, []string) int32) {
		if err := frame.RegisterCommand(Command{Name: name, Group: "Basic", Handler: handler}); err != nil {
			panic(err)
		}
	}
	add("echo", func(f *Frame, args []string) int32 { fmt.Fprintln(f.Out, strings.Join(args[1:], " ")); return 0 })
	add("set", func(f *Frame, args []string) int32 {
		if len(args) == 1 {
			flags := f.FlagList()
			keys := make([]string, 0, len(flags))
			for key := range flags {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Fprintf(f.Out, "%s %s\n", key, flags[key])
			}
			return 0
		}
		f.SetFlag(args[1], strings.Join(args[2:], " "))
		return 0
	})
	add("unset", func(f *Frame, args []string) int32 {
		if len(args) < 2 {
			fmt.Fprintln(f.Err, "usage: unset name ...")
			return 1
		}
		for _, name := range args[1:] {
			f.DeleteFlag(name)
		}
		return 0
	})
	add("alias", func(f *Frame, args []string) int32 {
		if len(args) == 1 {
			aliases := f.AliasList()
			names := make([]string, 0, len(aliases))
			for name := range aliases {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Fprintf(f.Out, "%s %s\n", name, strings.Join(aliases[name], " "))
			}
			return 0
		}
		if len(args) == 2 {
			alias, exists := f.LookupAlias(args[1])
			if !exists {
				return 1
			}
			fmt.Fprintln(f.Out, strings.Join(alias, " "))
			return 0
		}
		f.RegisterAlias(args[1], args[2:])
		return 0
	})
	add("unalias", func(f *Frame, args []string) int32 {
		if len(args) < 2 {
			return 1
		}
		for _, name := range args[1:] {
			f.RemoveAlias(name)
		}
		return 0
	})
	add("help", func(f *Frame, args []string) int32 {
		for _, command := range f.CommandList() {
			fmt.Fprintln(f.Out, command.Name)
		}
		return 0
	})
	add("version", func(f *Frame, args []string) int32 { fmt.Fprintln(f.Out, f.Version); return 0 })
	add("history", func(f *Frame, args []string) int32 {
		history := f.History()
		start := 0
		if len(args) > 1 {
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 0 {
				return 1
			}
			if n < len(history) {
				start = len(history) - n
			}
		}
		for i := start; i < len(history); i++ {
			fmt.Fprintf(f.Out, "%d: %s\n", i+1, history[i])
		}
		return 0
	})
	add("quit", func(f *Frame, args []string) int32 {
		if len(args) > 1 && args[1] == "-f" {
			return -2
		}
		return -1
	})
	add("exit", func(f *Frame, args []string) int32 { return -1 })
	add("undo", func(f *Frame, args []string) int32 {
		if f.BackupCount() == 0 {
			fmt.Fprintln(f.Err, "No backup network is available.")
			return 1
		}
		f.SwapCurrentAndBackup()
		return 0
	})
	add("source", sourceCommand)
	add("time", func(f *Frame, args []string) int32 {
		f.mu.RLock()
		last, total := f.TimeCommand, f.TimeTotal
		f.mu.RUnlock()
		fmt.Fprintf(f.Out, "time = %s, total = %s\n", last, total)
		return 0
	})
}
func sourceCommand(frame *Frame, args []string) int32 {
	silent, echo := false, false
	var path string
	for _, arg := range args[1:] {
		switch arg {
		case "-s":
			silent = true
		case "-e":
			echo = true
		default:
			if path != "" {
				fmt.Fprintln(frame.Err, "usage: source [-s] [-e] file")
				return 1
			}
			path = arg
		}
	}
	if path == "" {
		fmt.Fprintln(frame.Err, "usage: source [-s] [-e] file")
		return 1
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(frame.Err, err)
			return 1
		}
		path = home + path[1:]
	}
	file, err := os.Open(path)
	if err != nil {
		if silent && os.IsNotExist(err) {
			return 0
		}
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	defer file.Close()
	if err := frame.BeginSource(); err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	defer frame.EndSource()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), ABC_MAX_STR)
	for scanner.Scan() {
		line := scanner.Text()
		if echo {
			fmt.Fprintln(frame.Out, line)
		}
		if status := frame.Execute(line); status != 0 {
			return status
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	return 0
}
func (f *Frame) BackupCount() int { f.mu.RLock(); defer f.mu.RUnlock(); return len(f.backups) }
