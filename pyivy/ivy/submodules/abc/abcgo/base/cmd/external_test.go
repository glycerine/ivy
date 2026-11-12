package cmd

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSynthesisToolRoundtrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX executable fixture")
	}
	f, out := testFrame(t)
	f.SetCurrentNetwork("before")
	directory := t.TempDir()
	binary := filepath.Join(directory, "fake sis")
	script := "#!/bin/sh\n[ \"$1\" = -x ] || exit 5\n[ \"$2\" = -c ] || exit 6\ninput=$(printf '%s' \"$3\" | cut -d \"'\" -f 2)\noutput=$(printf '%s' \"$3\" | cut -d \"'\" -f 4)\ncat \"$input\" > \"$output\"\nprintf transformed >> \"$output\"\n"
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	f.SetFlag("sisunix", binary)
	var transferred string
	f.RegisterCommand(abcmain.Command{Name: "write_blif", Handler: func(_ *abcmain.Frame, args []string) int32 {
		if err := os.WriteFile(args[1], []byte("original\n"), 0600); err != nil {
			t.Error(err)
			return 1
		}
		return 0
	}})
	f.RegisterCommand(abcmain.Command{Name: "read_blif", Handler: func(_ *abcmain.Frame, args []string) int32 {
		data, err := os.ReadFile(args[1])
		if err != nil {
			t.Error(err)
			return 1
		}
		transferred = string(data)
		return 0
	}})
	if status := CmdCommandSis(f, 2, []string{"sis", "sweep"}); status != 0 {
		t.Fatal(status, out.String())
	}
	if !strings.Contains(transferred, "original\ntransformed") {
		t.Fatal(transferred)
	}
}
func TestExternalToolFailureSkipsImport(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX executable fixture")
	}
	f, _ := testFrame(t)
	f.SetCurrentNetwork("before")
	binary := filepath.Join(t.TempDir(), "bad")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexit 9\n"), 0700); err != nil {
		t.Fatal(err)
	}
	f.SetFlag("mvsisunix", binary)
	f.RegisterCommand(abcmain.Command{Name: "write_blif", Handler: func(_ *abcmain.Frame, args []string) int32 { return 0 }})
	f.RegisterCommand(abcmain.Command{Name: "read_blif", Handler: func(_ *abcmain.Frame, args []string) int32 { t.Error("import after failed child"); return 0 }})
	if status := CmdCommandMvsis(f, 2, []string{"mvsis", "sweep"}); status == 0 {
		t.Fatal("lost child error")
	}
}
