package main

import "testing"

func TestParseArgsAcceptsIvyCheckStyleParams(t *testing.T) {
	params, spec, err := parseArgs([]string{"isolate=left_player", "sample.ivy"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if spec != "sample.ivy" {
		t.Fatalf("spec = %q, want sample.ivy", spec)
	}
	if got := params["isolate"]; got != "left_player" {
		t.Fatalf("isolate = %q, want left_player", got)
	}
}

func TestParseArgsRejectsExtraSpec(t *testing.T) {
	_, _, err := parseArgs([]string{"a.ivy", "b.ivy"}, nil)
	if err == nil {
		t.Fatal("expected error for second spec path")
	}
}
