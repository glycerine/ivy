package wlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCodeMax4(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	Wlc_GenerateCodeMax4(4)
	data, err := os.ReadFile(filepath.Join("max4.v"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, part := range []string{"input [3:0] a, b, c, d;", "wire [8:0] A = a;", "module comp0", "module comp1", "2'b00"} {
		if !strings.Contains(text, part) {
			t.Errorf("generated Verilog missing %q", part)
		}
	}
}
