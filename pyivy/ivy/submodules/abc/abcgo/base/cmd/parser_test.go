package cmd

import (
	"reflect"
	"testing"
)

func TestCommandParsing(t *testing.T) {
	commands, err := ParseCommands(" echo \"a;b\" '' 'x#y'; read file\\ name.blif # trailing;ignored")
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"echo", "a;b", "", "x#y"}, {"read", "file name.blif"}}
	if len(commands) != len(want) {
		t.Fatal(commands)
	}
	for i, command := range commands {
		if !reflect.DeepEqual(command.Args, want[i]) {
			t.Fatal(command.Args, want[i])
		}
	}
	if _, err := ParseCommands("echo 'missing"); err == nil {
		t.Fatal("accepted unmatched quote")
	}
}
func TestShellRawQuotePreservation(t *testing.T) {
	command, rest, err := SplitCommand(`!printf '%s' "file name"; echo next`)
	if err != nil || command.Raw != `!printf '%s' "file name"` || rest != " echo next" {
		t.Fatal(command, rest, err)
	}
}
