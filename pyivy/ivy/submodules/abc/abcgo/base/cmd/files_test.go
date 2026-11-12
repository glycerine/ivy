package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirectoryScanAndRename(t *testing.T) {
	f, out := testFrame(t)
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "run_12.txt"), []byte("Property proved.  Time = 2.75 sec\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if status := CmdCommandScanDir(f, 3, []string{"scandir", "-D", directory}); status != 0 || !strings.Contains(out.String(), "12 2.75") {
		t.Fatal(status, out.String())
	}
	for _, name := range []string{"bench2.aig", "bench10.aig"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if status := CmdCommandRenameFiles(f, 9, []string{"renamefiles", "-D", directory, "-E", "aig", "-N", "renamed", "-B", "3"}); status != 0 {
		t.Fatal(status, out.String())
	}
	for _, name := range []string{"renamed05.aig", "renamed13.aig"} {
		if _, err := os.Stat(filepath.Join(directory, name)); err != nil {
			t.Fatal(err)
		}
	}
}
func TestRenameCollisionDoesNotOverwrite(t *testing.T) {
	f, out := testFrame(t)
	directory := t.TempDir()
	for _, name := range []string{"bench1.aig", "other1.aig"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if status := CmdCommandRenameFiles(f, 7, []string{"renamefiles", "-D", directory, "-E", "aig", "-N", "new"}); status == 0 {
		t.Fatal("duplicate destination accepted", out.String())
	}
	for _, name := range []string{"bench1.aig", "other1.aig"} {
		if _, err := os.Stat(filepath.Join(directory, name)); err != nil {
			t.Fatal(err)
		}
	}
}
func TestScriptGenerationQuotesFilenames(t *testing.T) {
	f, out := testFrame(t)
	directory := t.TempDir()
	for _, name := range []string{"one file.aig", "skip.txt", "skip.s"} {
		if err := os.WriteFile(filepath.Join(directory, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if status := CmdCommandScrGen(f, 7, []string{"scrgen", "-D", directory, "-F", "run.s", "-C", "ps; strash"}); status != 0 {
		t.Fatal(status, out.String())
	}
	data, err := os.ReadFile(filepath.Join(directory, "run.s"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "skip.") {
		t.Fatal(string(data))
	}
	lines := strings.SplitN(string(data), "\n", 2)
	commands, err := ParseCommands(lines[1])
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 3 || len(commands[0].Args) != 2 || commands[0].Args[1] != filepath.ToSlash(filepath.Join(directory, "one file.aig")) {
		t.Fatal(commands, string(data))
	}
}
func TestFilenameNumber(t *testing.T) {
	for name, want := range map[string]int32{"bench123.aig": 123, "123.aig": 123, "bench.aig": -2, "bench123": -1} {
		if got := CmfFindNumber(name); got != want {
			t.Fatal(name, got, want)
		}
	}
}
