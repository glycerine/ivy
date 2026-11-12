package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemporaryFileLifecycle(t *testing.T) {
	prefix := filepath.Join(t.TempDir(), "abc-")
	var name string
	f, err := Util_SignalTmpFile(prefix, ".cnf", &name)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".cnf") {
		t.Fatal(name)
	}
	if _, err = f.WriteString("content"); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	Util_SignalTmpFileRemove(name, 1)
	if _, err = os.Stat(name); err != nil {
		t.Fatal(err)
	}
	Util_SignalTmpFileRemove(name, 0)
	if _, err = os.Stat(name); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if nsprintf("%s:%d", "x", 3) != "x:3" {
		t.Fatal("format")
	}
}
func TestSignalSystemExit(t *testing.T) {
	if status := Util_SignalSystem("exit 7"); status != 7<<8 {
		t.Fatal(status)
	}
}
