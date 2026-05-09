package main

import "testing"

func TestParseArgsAcceptsIvyCheckStyleParams(t *testing.T) {
	params, isolates, spec, err := parseArgs([]string{"isolate=left_player", "sample.ivy"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if spec != "sample.ivy" {
		t.Fatalf("spec = %q, want sample.ivy", spec)
	}
	if got := params["isolate"]; got != "left_player" {
		t.Fatalf("isolate = %q, want left_player", got)
	}
	if len(isolates) != 1 || isolates[0] != "left_player" {
		t.Fatalf("isolates = %#v, want [left_player]", isolates)
	}
}

func TestParseArgsPreservesRepeatedIsolates(t *testing.T) {
	params, isolates, spec, err := parseArgs([]string{
		"diagnose=true",
		"isolate=left_player",
		"isolate=right_player",
		"sample.ivy",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if spec != "sample.ivy" {
		t.Fatalf("spec = %q, want sample.ivy", spec)
	}
	if got := params["diagnose"]; got != "true" {
		t.Fatalf("diagnose = %q, want true", got)
	}
	if got := params["isolate"]; got != "right_player" {
		t.Fatalf("collapsed isolate = %q, want right_player", got)
	}
	want := []string{"left_player", "right_player"}
	if len(isolates) != len(want) {
		t.Fatalf("isolates = %#v, want %#v", isolates, want)
	}
	for i := range want {
		if isolates[i] != want[i] {
			t.Fatalf("isolates = %#v, want %#v", isolates, want)
		}
	}
}

func TestParseArgsRejectsExtraSpec(t *testing.T) {
	_, _, _, err := parseArgs([]string{"a.ivy", "b.ivy"}, nil)
	if err == nil {
		t.Fatal("expected error for second spec path")
	}
}
