package gia_test

import (
	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"os"
	"os/exec"
	"testing"
)

func TestMappingSATOriginalExamples(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_sat_examples_review.c")
	want, err := exec.Command(executable).CombinedOutput()
	if err != nil {
		t.Fatalf("original examples: %v: %s", err, want)
	}
	output, err := os.CreateTemp(t.TempDir(), "examples")
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = original; output.Close() }()
	output.WriteString("example2\n")
	if_.If_ManSatTest2()
	output.WriteString("example3\n")
	if_.If_ManSatTest3()
	got, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("Go examples %q; original C %q", got, want)
	}
}
