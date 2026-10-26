package abc

// Go-only tests for src/demo.c, src/generic.h, src/starter.c, and
// src/template.c.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDefaultDemoRewritePlan(t *testing.T) {
	got, err := DefaultDemoRewritePlan("design.blif", RootExampleLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"read design.blif",
		"balance",
		"print_stats",
		"balance; rewrite -l; rewrite -lz; balance; rewrite -lz; balance",
		"print_stats",
		"write_blif result.blif",
		"cec design.blif result.blif",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("plan = %#v", got)
	}
	got[0] = "changed"
	again, err := DefaultDemoRewritePlan("design.blif", RootExampleLimits{})
	if err != nil || again[0] != want[0] {
		t.Fatal("plan aliases state", again, err)
	}
}

func TestDemoRewritePlanOptionsLimitsAndCancellation(t *testing.T) {
	got, err := DemoRewritePlan("x.aig", true, false, false, RootExampleLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"read x.aig",
		"balance",
		"balance; rewrite -l; refactor -l; balance; rewrite -l; rewrite -lz; balance; refactor -lz; rewrite -lz; balance",
		"write_blif result.blif",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("plan = %#v", got)
	}
	if _, err = DefaultDemoRewritePlan("x", RootExampleLimits{MaxCommands: 6}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = DefaultDemoRewritePlan(strings.Repeat("x", 1000), RootExampleLimits{}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = DefaultDemoRewritePlan("x", RootExampleLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err = DefaultDemoRewritePlan("x", RootExampleLimits{MaxCommands: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = DefaultDemoRewritePlan("x\x00y", RootExampleLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestStarterCommandPlan(t *testing.T) {
	data := []byte("# comment\n read skipped\n\talso skipped\nfirst\r\nsecond\nthird")
	got, err := StarterCommandPlan(data, 4, RootExampleLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"first", "second", "third"}; !slices.Equal(got, want) {
		t.Fatalf("commands = %#v", got)
	}
	data[0] = 'x'
	if !slices.Equal(got, []string{"first", "second", "third"}) {
		t.Fatal("commands alias input")
	}
	if _, err = StarterCommandPlan(nil, 0, RootExampleLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = StarterCommandPlan([]byte("one\ntwo\n"), 2, RootExampleLimits{MaxCommands: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = StarterCommandPlan([]byte("long\n"), 2, RootExampleLimits{MaxLineBytes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = StarterCommandPlan([]byte("one\n"), 2, RootExampleLimits{MaxInputBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = StarterCommandPlan([]byte("one\n"), 2, RootExampleLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err = StarterCommandPlan([]byte("one\x00two\n"), 2, RootExampleLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func rootStarterModel(data []byte, maxLine int) ([]string, bool) {
	if slices.Contains(data, byte(0)) {
		return nil, false
	}
	commands := make([]string, 0)
	start := 0
	for start < len(data) {
		end := start
		for end < len(data) && data[end] != '\n' {
			end++
		}
		if end < len(data) {
			end++
		}
		line := data[start:end]
		start = end
		if len(line) > maxLine {
			return nil, false
		}
		if len(line) == 0 || line[0] == '\n' || line[0] == '\r' || line[0] == '\t' || line[0] == ' ' || line[0] == '#' {
			continue
		}
		last := len(line)
		if last > 0 && line[last-1] == '\n' {
			last--
		}
		if last > 0 && line[last-1] == '\r' {
			last--
		}
		commands = append(commands, string(line[:last]))
	}
	return commands, true
}

func rootDemoModel(input string, resyn2, stats, verify bool) []string {
	commands := []string{"read " + input, "balance"}
	if stats {
		commands = append(commands, "print_stats")
	}
	synthesis := "balance; rewrite -l; rewrite -lz; balance; rewrite -lz; balance"
	if resyn2 {
		synthesis = "balance; rewrite -l; refactor -l; balance; rewrite -l; rewrite -lz; balance; refactor -lz; rewrite -lz; balance"
	}
	commands = append(commands, synthesis)
	if stats {
		commands = append(commands, "print_stats")
	}
	commands = append(commands, "write_blif result.blif")
	if verify {
		commands = append(commands, "cec "+input+" result.blif")
	}
	return commands
}

func TestDemoRewritePlanIndependentModel5000(t *testing.T) {
	state := uint64(0xbb67ae8584caa73b)
	for execution := 0; execution < 5000; execution++ {
		state = state*6364136223846793005 + 1442695040888963407
		input := strings.Repeat(string(byte('a'+state%26)), int(state%80))
		resyn2, stats, verify := state&1 != 0, state&2 != 0, state&4 != 0
		got, err := DemoRewritePlan(input, resyn2, stats, verify, RootExampleLimits{})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		if want := rootDemoModel(input, resyn2, stats, verify); !slices.Equal(got, want) {
			t.Fatalf("execution %d got=%q want=%q", execution, got, want)
		}
	}
}

func TestStarterCommandPlanIndependentModel5000(t *testing.T) {
	state := uint64(0x6a09e667f3bcc909)
	for execution := 0; execution < 5000; execution++ {
		length := execution % 96
		data := make([]byte, length)
		for index := range data {
			state ^= state << 13
			state ^= state >> 7
			state ^= state << 17
			alphabet := "abc #\t\r\n;09"
			data[index] = alphabet[state%uint64(len(alphabet))]
		}
		want, valid := rootStarterModel(data, 31)
		got, err := StarterCommandPlan(data, 2, RootExampleLimits{MaxInputBytes: 128, MaxLineBytes: 31, MaxCommands: 128, MaxCommandBytes: 31})
		if (err == nil) != valid {
			t.Fatalf("execution %d valid=%v err=%v data=%q", execution, valid, err, data)
		}
		if valid && !slices.Equal(got, want) {
			t.Fatalf("execution %d got=%q want=%q", execution, got, want)
		}
	}
}

func FuzzStarterCommandPlan(f *testing.F) {
	f.Add([]byte("one\ntwo\n"))
	f.Add([]byte("# skip\r\ncommand\r\n next\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			return
		}
		want, valid := rootStarterModel(data, 127)
		got, err := StarterCommandPlan(data, 2, RootExampleLimits{MaxInputBytes: 512, MaxLineBytes: 127, MaxCommands: 512, MaxCommandBytes: 127})
		if (err == nil) != valid {
			t.Fatalf("valid=%v err=%v data=%q", valid, err, data)
		}
		if valid && !slices.Equal(got, want) {
			t.Fatalf("got=%q want=%q", got, want)
		}
	})
}

func FuzzDemoRewritePlan(f *testing.F) {
	f.Add("design.blif", uint8(7))
	f.Add("", uint8(0))
	f.Fuzz(func(t *testing.T, input string, flags uint8) {
		if len(input) > 512 {
			return
		}
		resyn2, stats, verify := flags&1 != 0, flags&2 != 0, flags&4 != 0
		got, err := DemoRewritePlan(input, resyn2, stats, verify, RootExampleLimits{MaxCommandBytes: 600})
		if strings.IndexByte(input, 0) >= 0 {
			if !errors.Is(err, ErrDimension) {
				t.Fatalf("NUL input error=%v", err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if want := rootDemoModel(input, resyn2, stats, verify); !slices.Equal(got, want) {
			t.Fatalf("got=%q want=%q", got, want)
		}
	})
}
