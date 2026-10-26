package abc

// Tests for public utility queries from misc/zlib/zutil.c.

import (
	"encoding/json"
	"os"
	"testing"
)

type zlibUtilOracle struct {
	Version string `json:"version"`
	Flags   uint64 `json:"flags"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func TestZlibUtilNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/zlib_util.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle zlibUtilOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Errors) != 9 {
		t.Fatal(len(oracle.Errors), err)
	}
	if got := ZlibVersion(); got != oracle.Version {
		t.Fatal(got, oracle.Version)
	}
	if got := ZlibCompileFlags(); got != oracle.Flags {
		t.Fatal(got, oracle.Flags)
	}
	for _, row := range oracle.Errors {
		if got, valid := ZlibError(row.Code); !valid || got != row.Message {
			t.Fatal(row, got, valid)
		}
	}
}

func TestZlibErrorBounds(t *testing.T) {
	for _, code := range []int{-1000, -7, 3, 1000} {
		if message, valid := ZlibError(code); valid || message != "" {
			t.Fatal(code, message, valid)
		}
	}
}

func FuzzZlibError(f *testing.F) {
	for code := -7; code <= 3; code++ {
		f.Add(code)
	}
	f.Fuzz(func(t *testing.T, code int) {
		message, valid := ZlibError(code)
		wantValid := code >= -6 && code <= 2
		if valid != wantValid || (!valid && message != "") {
			t.Fatal(code, message, valid)
		}
		if valid && message != zlibErrorMessages[code+6] {
			t.Fatal(code, message)
		}
	})
}
