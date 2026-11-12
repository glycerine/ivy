package cmd

import (
	"bytes"
	abcmain "github.com/glycerine/abcgo/base/main"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testFrame(t *testing.T) (*Abc_Frame_t, *bytes.Buffer) {
	t.Helper()
	f := abcmain.NewFrame()
	out := new(bytes.Buffer)
	f.Out = out
	f.Err = out
	Cmd_Init(f)
	t.Cleanup(func() { Cmd_End(f) })
	return f, out
}
func TestCommandAliasesAndAutoexec(t *testing.T) {
	f, out := testFrame(t)
	if status := Cmd_CommandExecute(f, `alias hi echo hello; hi "two words"`); status != 0 {
		t.Fatal(status, out.String())
	}
	if !strings.Contains(out.String(), "hello two words") {
		t.Fatal(out.String())
	}
	f.SetFlag("autoexec", "echo automatic")
	if status := Cmd_CommandExecute(f, "echo primary"); status != 0 {
		t.Fatal(status)
	}
	if strings.Count(out.String(), "automatic") != 1 {
		t.Fatal(out.String())
	}
}
func TestAliasLoopIsBounded(t *testing.T) {
	f, out := testFrame(t)
	f.RegisterAlias("a", []string{"b"})
	f.RegisterAlias("b", []string{"a"})
	if status := Cmd_CommandExecute(f, "a"); status == 0 {
		t.Fatal("recursive aliases succeeded", out.String())
	}
}
func TestSourceAndQuotedSemicolons(t *testing.T) {
	f, out := testFrame(t)
	file := filepath.Join(t.TempDir(), "commands.abc")
	if err := os.WriteFile(file, []byte("echo 'a;b'\necho second\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if status := Cmd_CommandExecute(f, "source "+shellQuote(file)); status != 0 {
		t.Fatal(status, out.String())
	}
	if !strings.Contains(out.String(), "a;b") || !strings.Contains(out.String(), "second") {
		t.Fatal(out.String())
	}
	if f.SourceDepth() != 0 {
		t.Fatal("source depth leaked")
	}
}
func TestRunStarterFailures(t *testing.T) {
	file := filepath.Join(t.TempDir(), "jobs")
	if err := os.WriteFile(file, []byte("printf first\nprintf second\nexit 7\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RunStarter(file, "", "", 3, &out, &out); err == nil {
		t.Fatal("lost child failure")
	}
	if !strings.Contains(out.String(), "first") || !strings.Contains(out.String(), "second") {
		t.Fatal(out.String())
	}
	if err := RunStarter(file, "", "", 1, &out, &out); err == nil {
		t.Fatal("one core would deadlock source implementation")
	}
}
func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	if got := expandHome("~"); got != home {
		t.Fatalf("%s != %s", got, home)
	}
}
func TestHistoryLog(t *testing.T) {
	f, _ := testFrame(t)
	file := filepath.Join(t.TempDir(), "history")
	if status := Cmd_CommandExecute(f, "set history "+shellQuote(file)); status != 0 {
		t.Fatal(status)
	}
	Cmd_CommandExecute(f, "echo logged")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "echo logged") {
		t.Fatal(string(data))
	}
}
func TestAliasPrefixLoopIsBounded(t *testing.T) {
	f, out := testFrame(t)
	f.RegisterAlias("a", []string{"a; echo never"})
	args := []string{"a"}
	count := int32(1)
	loops := int32(0)
	if status := CmdApplyAlias(f, &count, &args, &loops); status == 0 {
		t.Fatal("prefix recursion succeeded", out.String())
	}
}
func TestEndClearsSharedCommandState(t *testing.T) {
	f, _ := testFrame(t)
	f.RegisterCommand(abcmain.Command{Name: "extension", Handler: func(*abcmain.Frame, []string) int32 { return 0 }})
	f.RegisterAlias("alias", []string{"echo"})
	f.SetFlag("custom", "value")
	f.AppendHistory("custom command")
	Cmd_End(f)
	if len(f.CommandList()) != 0 || len(f.AliasList()) != 0 || len(f.FlagList()) != 0 || len(f.History()) != 0 {
		t.Fatal("command teardown leaked shared state")
	}
	Cmd_Init(f)
	if _, ok := f.LookupCommand("echo"); !ok {
		t.Fatal("cannot initialize after teardown")
	}
}
