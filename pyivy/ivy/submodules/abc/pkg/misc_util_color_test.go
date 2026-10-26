package abc

// Tests for ANSI color diagnostics from misc/util/utilColor.c.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestColorTestANSINative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/util_color.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Length int    `json:"length"`
		SHA256 string `json:"sha256"`
	}
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	output := ColorTestANSI()
	hash := sha256.Sum256([]byte(output))
	if len(output) != oracle.Length || hex.EncodeToString(hash[:]) != oracle.SHA256 {
		t.Fatal(len(output), oracle.Length, hex.EncodeToString(hash[:]), oracle.SHA256)
	}
	if output != ColorTestANSI() {
		t.Fatal("nondeterministic output")
	}
}

func TestColorTestANSIStructure(t *testing.T) {
	output := ColorTestANSI()
	if !strings.HasPrefix(output, "Background       ") || !strings.HasSuffix(output, "\x1b[8mConcealed \x1b[0m\n") {
		t.Fatal("envelope")
	}
	if strings.Count(output, "  Hi  ") != 144 {
		t.Fatal(strings.Count(output, "  Hi  "))
	}
	if strings.Count(output, "Foreground ") != 18 || strings.Count(output, "\x1b[0m") != 150 {
		t.Fatal(strings.Count(output, "Foreground "), strings.Count(output, "\x1b[0m"))
	}
}
