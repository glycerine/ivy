package cmd

import (
	"fmt"
	"strings"
)

// ParsedCommand keeps raw shell text as well as argv. In particular, shell
// escapes must not be reconstructed from argv and lose quoting of filenames.
type ParsedCommand struct {
	Args []string
	Raw  string
}

func SplitCommand(text string) (ParsedCommand, string, error) {
	var command ParsedCommand
	var word strings.Builder
	quote := byte(0)
	started := false
	rawStart := -1
	flush := func() {
		if started {
			command.Args = append(command.Args, word.String())
			word.Reset()
			started = false
		}
	}
	for i := 0; i < len(text); i++ {
		c := text[i]
		if rawStart < 0 && c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			rawStart = i
		}
		if quote == 0 && (c == ';' || c == '#') {
			flush()
			if rawStart >= 0 {
				command.Raw = strings.TrimSpace(text[rawStart:i])
			}
			if c == '#' {
				return command, "", nil
			}
			return command, text[i+1:], nil
		}
		if c == '\\' && quote != '\'' && i+1 < len(text) {
			next := text[i+1]
			if next == '\n' {
				i++
				continue
			}
			if next == '\\' || next == '\'' || next == '"' || next == ';' || next == '#' || next == ' ' || next == '\t' {
				word.WriteByte(next)
				started = true
				i++
				continue
			}
		}
		if c == '\'' || c == '"' {
			if quote == 0 {
				quote = c
				started = true
				continue
			}
			if quote == c {
				quote = 0
				continue
			}
		}
		if quote == 0 && (c == ' ' || c == '\t' || c == '\n' || c == '\r') {
			flush()
			continue
		}
		word.WriteByte(c)
		started = true
	}
	if quote != 0 {
		return ParsedCommand{}, "", fmt.Errorf("unbalanced %c quote", quote)
	}
	flush()
	if rawStart >= 0 {
		command.Raw = strings.TrimSpace(text[rawStart:])
	}
	return command, "", nil
}
func ParseCommands(text string) ([]ParsedCommand, error) {
	var commands []ParsedCommand
	for text != "" {
		command, rest, err := SplitCommand(text)
		if err != nil {
			return nil, err
		}
		if len(command.Args) > 0 {
			commands = append(commands, command)
		}
		text = rest
	}
	return commands, nil
}
