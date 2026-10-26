package abc

// Tests for src/misc/util/utilSignal.c.

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type utilSignalOracle struct {
	System []int `json:"system"`
	Temp   []int `json:"temp"`
}

func TestUtilSignalNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/util_signal.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle utilSignalOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.System) != 2 || len(oracle.Temp) != 7 {
		t.Fatal(err, oracle)
	}
	if result := UtilSignalSystem("exit 0"); result != oracle.System[0] {
		t.Fatal(result)
	}
	if result := UtilSignalSystem("exit 7"); result != oracle.System[1] {
		t.Fatal(result)
	}
	if result := UtilSignalSystem(""); result != 0 {
		t.Fatal(result)
	}
	directory := t.TempDir()
	prefix := filepath.Join(directory, "abc-go-signal-")
	first, firstName, err := CreateSignalTempFile(prefix, ".tmp")
	if err != nil {
		t.Fatal(err)
	}
	second, secondName, err := CreateSignalTempFile(prefix, ".tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	defer os.Remove(firstName)
	defer os.Remove(secondName)
	if boolInt(firstName != secondName) != oracle.Temp[0] || boolInt(strings.HasPrefix(firstName, prefix)) != oracle.Temp[1] || boolInt(strings.HasSuffix(firstName, ".tmp")) != oracle.Temp[2] {
		t.Fatal(firstName, secondName)
	}
	info, err := first.Stat()
	if err != nil || boolInt(info.Mode().Perm() == 0600) != oracle.Temp[3] {
		t.Fatal(info, err)
	}
	if _, err = first.WriteString("abc"); err != nil {
		t.Fatal(err)
	}
	if _, err = first.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 3)
	count, err := first.Read(buffer)
	if err != nil || boolInt(count == 3 && string(buffer) == "abc") != oracle.Temp[4] {
		t.Fatal(count, string(buffer), err)
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	if err = second.Close(); err != nil {
		t.Fatal(err)
	}
	if err = RemoveSignalTempFile(firstName, true); err != nil {
		t.Fatal(err)
	}
	_, keptErr := os.Stat(firstName)
	if boolInt(keptErr == nil) != oracle.Temp[5] {
		t.Fatal(keptErr)
	}
	if err = RemoveSignalTempFile(firstName, false); err != nil {
		t.Fatal(err)
	}
	if err = RemoveSignalTempFile(secondName, false); err != nil {
		t.Fatal(err)
	}
	_, firstErr := os.Stat(firstName)
	_, secondErr := os.Stat(secondName)
	if boolInt(os.IsNotExist(firstErr) && os.IsNotExist(secondErr)) != oracle.Temp[6] {
		t.Fatal(firstErr, secondErr)
	}
	if err = RemoveSignalTempFile(firstName, false); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	badFile, badName, err := CreateSignalTempFile(filepath.Join(directory, "missing", "prefix"), ".tmp")
	if err == nil || badFile != nil || badName != "" {
		t.Fatal(badFile, badName, err)
	}
	if err = RemoveSignalTempFile(filepath.Join(directory, "missing"), true); err != nil {
		t.Fatal(err)
	}
}

func FuzzSignalTempFile(f *testing.F) {
	f.Add([]byte("prefix"), []byte(".tmp"), []byte("payload"), false)
	f.Fuzz(func(t *testing.T, prefixBytes, suffixBytes, payload []byte, leave bool) {
		if len(prefixBytes) > 32 {
			prefixBytes = prefixBytes[:32]
		}
		if len(suffixBytes) > 32 {
			suffixBytes = suffixBytes[:32]
		}
		if len(payload) > 256 {
			payload = payload[:256]
		}
		prefix := filepath.Join(t.TempDir(), hex.EncodeToString(prefixBytes))
		suffix := hex.EncodeToString(suffixBytes)
		file, name, err := CreateSignalTempFile(prefix, suffix)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
			t.Fatal(name, prefix, suffix)
		}
		if count, err := file.Write(payload); err != nil || count != len(payload) {
			t.Fatal(count, err)
		}
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
		if err = RemoveSignalTempFile(name, leave); err != nil {
			t.Fatal(err)
		}
		_, err = os.Stat(name)
		if leave && err != nil || !leave && !os.IsNotExist(err) {
			t.Fatal(leave, err)
		}
		if leave {
			if err = os.Remove(name); err != nil {
				t.Fatal(err)
			}
		}
	})
}
