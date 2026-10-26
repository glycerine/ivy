package abc

// Standalone tests for src/base/main behavior.

import (
	"errors"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func mainTestFrame(t *testing.T) *MainFrame {
	t.Helper()
	frame, err := NewMainFrame("abc-test", MainLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return frame
}

func TestMainBuildInventory(t *testing.T) {
	want := [...]string{"main.c", "mainFrame.c", "mainInit.c", "mainLib.c", "mainReal.c", "libSupport.c", "mainUtils.c"}
	if MainBuildInventory != want {
		t.Fatalf("inventory %v", MainBuildInventory)
	}
	if len(MainSourceInventory) != 13 {
		t.Fatalf("source inventory has %d entries", len(MainSourceInventory))
	}
	seen := make(map[string]bool)
	for _, unit := range MainSourceInventory {
		if unit.Source == "" || unit.TrackedNames < 1 || seen[unit.Source] {
			t.Fatalf("invalid source inventory entry %+v", unit)
		}
		seen[unit.Source] = true
		if unit.Complete != (unit.Source == "module.make") {
			t.Fatalf("overstated disposition %+v", unit)
		}
	}
}

func TestMainFrameLifecycleAndOwnership(t *testing.T) {
	frame := mainTestFrame(t)
	for _, name := range []string{"cmd", "io", "abc"} {
		if err := frame.RegisterSubsystem(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := frame.RegisterSubsystem("cmd"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("duplicate subsystem %v", err)
	}
	if err := frame.Initialize(); err != nil {
		t.Fatal(err)
	}
	if err := frame.SetFlag("progressbar", ""); err != nil {
		t.Fatal(err)
	}
	if enabled, err := frame.FlagEnabled("progressbar"); err != nil || !enabled {
		t.Fatalf("enabled %v %v", enabled, err)
	}
	if old, err := frame.SetNameMode(3); err != nil || old != 0 {
		t.Fatalf("mode %d %v", old, err)
	}
	if err := frame.SetSession(true, true, "spec.aig", "NAND2", 2.5); err != nil {
		t.Fatal(err)
	}

	n1 := MainNetwork{ID: 1, Name: "n1", Kind: MainNetworkStrashed, Mapped: true, PrimaryOutputs: []int8{0, 1, -1}, Payload: []byte{1}}
	n2 := MainNetwork{ID: 2, Name: "n2", Kind: MainNetworkLogic, Mapped: true, Payload: []byte{2}}
	n3 := MainNetwork{ID: 3, Name: "n3", Kind: MainNetworkLogic, Payload: []byte{3}}
	if err := frame.SetCurrentNetwork(n1, 2); err != nil {
		t.Fatal(err)
	}
	if err := frame.SetCurrentNetwork(n2, 2); err != nil {
		t.Fatal(err)
	}
	if err := frame.SetCurrentNetwork(n3, 1); err != nil {
		t.Fatal(err)
	}
	snapshot, err := frame.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Networks) != 2 || snapshot.Networks[0].ID != 3 || snapshot.Networks[1].ID != 2 {
		t.Fatalf("stack %+v", snapshot.Networks)
	}
	if err := frame.SwapCurrentAndBackup(); err != nil {
		t.Fatal(err)
	}
	snapshot, _ = frame.Snapshot()
	if snapshot.Networks[0].ID != 2 {
		t.Fatalf("swap %+v", snapshot.Networks)
	}
	if err := frame.ReplaceCurrentNetwork(n1, true); err != nil {
		t.Fatal(err)
	}
	if value, err := frame.CheckPOConstant(1); err != nil || value != 1 {
		t.Fatalf("constant %d %v", value, err)
	}
	if err := frame.UnmapAllNetworks(); err != nil {
		t.Fatal(err)
	}

	data := []byte{4, 5}
	library := MainLibrary{Name: "cells", Kind: "genlib", Data: data}
	if err := frame.SetLibrary(library); err != nil {
		t.Fatal(err)
	}
	data[0] = 99
	cexBits := []bool{true, false, true}
	cex := MainCounterexample{PIs: 1, Registers: 1, Output: 0, Frame: 1, Bits: cexBits}
	if err := frame.SetVerification(1, 2, &cex); err != nil {
		t.Fatal(err)
	}
	cexBits[0] = false
	if err := frame.AddLog("info", "ready"); err != nil {
		t.Fatal(err)
	}
	snapshot, err = frame.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !snapshot.Initialized || !snapshot.Batch || !snapshot.Bridge || snapshot.NameMode != 3 || snapshot.Libraries[0].Data[0] != 4 || !snapshot.Verification.CEX.Bits[0] {
		t.Fatalf("snapshot %+v", snapshot)
	}
	snapshot.Libraries[0].Data[0] = 55
	snapshot.Networks[0].Payload[0] = 55
	again, _ := frame.Snapshot()
	if again.Libraries[0].Data[0] != 4 || again.Networks[0].Payload[0] == 55 {
		t.Fatal("snapshot aliases frame")
	}
	lib, found, err := frame.TakeLibrary("cells")
	if err != nil || !found || lib.Data[0] != 4 {
		t.Fatalf("take %+v %v %v", lib, found, err)
	}
	if err := frame.ClearVerification(); err != nil {
		t.Fatal(err)
	}
	if err := frame.DeleteAllNetworks(); err != nil {
		t.Fatal(err)
	}
	if err := frame.End(); err != nil {
		t.Fatal(err)
	}
	if err := frame.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := frame.Snapshot(); !errors.Is(err, ErrMainClosed) {
		t.Fatalf("closed %v", err)
	}
}

func TestMainNetworkReplacementModes(t *testing.T) {
	frame := mainTestFrame(t)
	for i := 1; i <= 3; i++ {
		if err := frame.SetCurrentNetwork(MainNetwork{ID: uint64(i), Name: string(rune('a' + i)), Kind: MainNetworkLogic}, 3); err != nil {
			t.Fatal(err)
		}
	}
	before, _ := frame.Snapshot()
	if err := frame.ReplaceCurrentNetwork(MainNetwork{ID: 9, Name: "replacement", Kind: MainNetworkLogic}, true); err != nil {
		t.Fatal(err)
	}
	after, _ := frame.Snapshot()
	if len(after.Networks) != len(before.Networks) || after.Networks[1].ID != before.Networks[1].ID || after.Networks[0].Step != before.Networks[0].Step {
		t.Fatalf("preserve %+v", after.Networks)
	}
	if err := frame.ReplaceCurrentNetwork(MainNetwork{ID: 10, Name: "only", Kind: MainNetworkLogic}, false); err != nil {
		t.Fatal(err)
	}
	after, _ = frame.Snapshot()
	if len(after.Networks) != 1 || after.Networks[0].ID != 10 {
		t.Fatalf("discard %+v", after.Networks)
	}
}

func TestMainLimitsCancellationAndAtomicity(t *testing.T) {
	frame, err := NewMainFrame("v", MainLimits{MaxNetworks: 2, MaxLibraries: 1, MaxFlags: 1, MaxSubsystems: 1, MaxBytes: 32, MaxText: 12, MaxOperations: 8, MaxDepth: 2})
	if err != nil {
		t.Fatal(err)
	}
	if err = frame.SetFlag("a", "b"); err != nil {
		t.Fatal(err)
	}
	before, _ := frame.Snapshot()
	if err = frame.SetFlag("c", "d"); !errors.Is(err, ErrLimit) {
		t.Fatalf("flag limit %v", err)
	}
	if err = frame.SetLibrary(MainLibrary{Name: "toolonglibrary", Kind: "kind"}); !errors.Is(err, ErrLimit) {
		t.Fatalf("text limit %v", err)
	}
	after, _ := frame.Snapshot()
	if !reflect.DeepEqual(before.Flags, after.Flags) || len(after.Libraries) != 0 {
		t.Fatal("failed mutation was not atomic")
	}
	if err = frame.SetCurrentNetwork(MainNetwork{ID: 1, Name: "n", Kind: MainNetworkLogic}, 0); err != nil {
		t.Fatal(err)
	}
	if err = frame.SetCurrentNetwork(MainNetwork{ID: 2, Name: "m", Kind: MainNetworkLogic}, 1); err != nil {
		t.Fatal(err)
	}
	if err = frame.SetCurrentNetwork(MainNetwork{ID: 3, Name: "o", Kind: MainNetworkLogic}, 2); !errors.Is(err, ErrLimit) {
		t.Fatalf("depth %v", err)
	}

	stop := false
	cancelled, _ := NewMainFrame("v", MainLimits{Stop: func() bool { return stop }})
	stop = true
	if err = cancelled.SetFlag("x", "y"); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel flag %v", err)
	}
	if err = cancelled.SetCurrentNetwork(MainNetwork{ID: 1, Name: "n", Kind: MainNetworkLogic}, 0); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel network %v", err)
	}
	if snap, e := cancelled.Snapshot(); !errors.Is(e, ErrCancelled) || len(snap.Networks) != 0 {
		t.Fatalf("cancel snapshot %+v %v", snap, e)
	}

	var nilFrame *MainFrame
	if err = nilFrame.Close(); !errors.Is(err, ErrMainClosed) {
		t.Fatalf("nil close %v", err)
	}
	if _, err = nilFrame.Snapshot(); !errors.Is(err, ErrMainClosed) {
		t.Fatalf("nil snapshot %v", err)
	}
}

func TestMainCounterexampleAndInputValidation(t *testing.T) {
	frame := mainTestFrame(t)
	bad := MainCounterexample{PIs: 2, Registers: 1, Output: 0, Frame: 2, Bits: make([]bool, 6)}
	if err := frame.SetVerification(1, 2, &bad); !errors.Is(err, ErrInvariant) {
		t.Fatalf("short cex %v", err)
	}
	if err := frame.SetCurrentNetwork(MainNetwork{ID: 0, Name: "bad"}, 0); !errors.Is(err, ErrInvariant) {
		t.Fatalf("network %v", err)
	}
	if err := frame.SetSession(false, false, "", "", -1); !errors.Is(err, ErrDimension) {
		t.Fatalf("load %v", err)
	}
	if err := frame.SetFlag("namemode", "bad"); !errors.Is(err, ErrDimension) {
		t.Fatalf("mode %v", err)
	}
	if _, found, _ := frame.Flag("namemode"); found {
		t.Fatal("invalid mode retained")
	}
}

func TestMainRuntimeIsolation(t *testing.T) {
	var nilRuntime *MainRuntime
	if _, err := nilRuntime.Start("nil", MainLimits{}); !errors.Is(err, ErrMainClosed) {
		t.Fatalf("nil runtime start %v", err)
	}
	if nilRuntime.Frame() != nil {
		t.Fatal("nil runtime returned a frame")
	}
	if err := nilRuntime.Stop(); !errors.Is(err, ErrMainClosed) {
		t.Fatalf("nil runtime stop %v", err)
	}
	var first, second MainRuntime
	f1, err := first.Start("one", MainLimits{})
	if err != nil {
		t.Fatal(err)
	}
	f2, err := second.Start("two", MainLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = f1.SetFlag("x", "1"); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := f2.Flag("x"); found {
		t.Fatal("frames leaked state")
	}
	if _, err = first.Start("again", MainLimits{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("double start %v", err)
	}
	if err = first.Stop(); err != nil {
		t.Fatal(err)
	}
	if first.Frame() != nil {
		t.Fatal("runtime retained frame")
	}
	if _, err = f1.Snapshot(); !errors.Is(err, ErrMainClosed) {
		t.Fatalf("old frame %v", err)
	}
	if err = second.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestMainFrameConcurrentReadersAndWriters(t *testing.T) {
	frame := mainTestFrame(t)
	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		worker := worker
		wait.Add(1)
		go func() {
			defer wait.Done()
			for i := 0; i < 200; i++ {
				name := string(rune('a' + (worker+i)%16))
				if worker%2 == 0 {
					if err := frame.SetFlag(name, "1"); err != nil {
						t.Errorf("set: %v", err)
						return
					}
				} else {
					if _, _, err := frame.Flag(name); err != nil {
						t.Errorf("read: %v", err)
						return
					}
					if _, err := frame.Snapshot(); err != nil {
						t.Errorf("snapshot: %v", err)
						return
					}
				}
			}
		}()
	}
	wait.Wait()
}

func TestMainArgumentsAndUtilities(t *testing.T) {
	got, err := ParseMainArguments([]string{"-q", "strash; print_stats", "-s", "-b", "-t", "blif", "-T", "none", "-o", "out.blif", "in.blif"}, MainLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Quiet || !got.SkipSource || !got.Bridge || got.Input != "in.blif" || got.Output != "out.blif" || got.InputType != "blif" || got.OutputType != "none" || len(got.Commands) != 1 {
		t.Fatalf("args %+v", got)
	}
	got, err = ParseMainArguments([]string{"-x", "--", "-literal"}, MainLimits{})
	if err != nil || got.Input != "-literal" || got.InputType != "none" || got.OutputType != "none" {
		t.Fatalf("dash %+v %v", got, err)
	}
	for _, args := range [][]string{{"-z"}, {"-o"}, {"a", "b"}, {"--", "a", "b"}} {
		if _, err = ParseMainArguments(args, MainLimits{}); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if _, err = ParseMainArguments([]string{"-c", "12345"}, MainLimits{MaxText: 4}); !errors.Is(err, ErrLimit) {
		t.Fatalf("text %v", err)
	}
	if prompt, err := MainPrompt(3); err != nil || prompt != "abc 03> " {
		t.Fatalf("prompt %q %v", prompt, err)
	}
	if version, err := MainVersion("abc", "Jan 1", "12:00"); err != nil || version != "abc (compiled Jan 1 12:00)" {
		t.Fatalf("version %q %v", version, err)
	}
	if got := MainBriefDate("Mon Jan 02 13:04:05 UTC 2023"); got != "2-Jan-23 at 1:04 PM" {
		t.Fatalf("date %q", got)
	}
	if got := MainBriefDate("not a date"); got != "not a date" {
		t.Fatalf("bad date %q", got)
	}
}

func TestMainReviewRegressions(t *testing.T) {
	var nilFrame *MainFrame
	if _, err := nilFrame.SetNameMode(1); !errors.Is(err, ErrMainClosed) {
		t.Fatalf("nil name mode %v", err)
	}

	stop := false
	frame, err := NewMainFrame("v", MainLimits{MaxOperations: 8, MaxText: 32, MaxBytes: 1024, Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	first := MainNetwork{ID: 1, Name: "first", Kind: MainNetworkLogic, Payload: []byte{1}}
	if err = frame.SetCurrentNetwork(first, 1); err != nil {
		t.Fatal(err)
	}
	before, err := frame.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	stop = true
	if err = frame.ReplaceCurrentNetwork(MainNetwork{ID: 2, Name: "second", Kind: MainNetworkLogic}, false); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel replace %v", err)
	}
	stop = false
	after, err := frame.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if after.Steps != before.Steps || len(after.Networks) != 1 || after.Networks[0].ID != 1 {
		t.Fatalf("failed replacement mutated frame: before=%+v after=%+v", before, after)
	}
	if err = frame.SetCurrentNetwork(MainNetwork{ID: 1, Name: "changed", Kind: MainNetworkLogic}, 1); !errors.Is(err, ErrInvariant) {
		t.Fatalf("conflicting current identity %v", err)
	}
	if err = frame.SetCurrentNetwork(MainNetwork{ID: 3, Name: "third", Kind: MainNetworkLogic}, 1); err != nil {
		t.Fatal(err)
	}
	if err = frame.SetCurrentNetwork(MainNetwork{ID: 1, Name: "first", Kind: MainNetworkLogic, Payload: []byte{1}}, 2); !errors.Is(err, ErrInvariant) {
		t.Fatalf("backup identity reuse %v", err)
	}

	tight, err := NewMainFrame("v", MainLimits{MaxOperations: 2, MaxText: 5, MaxBytes: 128})
	if err != nil {
		t.Fatal(err)
	}
	if err = tight.SetFlag("a", "1"); err != nil {
		t.Fatal(err)
	}
	if err = tight.SetFlag("b", "1"); err != nil {
		t.Fatal(err)
	}
	if err = tight.SetLibrary(MainLibrary{Name: "c", Kind: "k"}); err != nil {
		t.Fatal(err)
	}
	if _, err = tight.Snapshot(); !errors.Is(err, ErrLimit) {
		t.Fatalf("snapshot preflight %v", err)
	}
	if err = tight.AddLog("abc", "def"); !errors.Is(err, ErrLimit) {
		t.Fatalf("combined log text limit %v", err)
	}
	if err = tight.Close(); err != nil {
		t.Fatal(err)
	}
	if tight.version != "" || tight.specName != "" || tight.drivingCell != "" {
		t.Fatal("close retained owned text")
	}
}

func TestMainArgumentSourceSemantics(t *testing.T) {
	got, err := ParseMainArguments([]string{"-d", "-d", "-m", "12", "-l", "3", "-c", "first", "-F", "script.abc", "-S", "smt", "-t", "bench", "-T", "pla", "input"}, MainLimits{})
	if err != nil {
		t.Fatal(err)
	}
	wantCommands := []string{"first", "source -x script.abc", "smt"}
	if got.Debug || got.MemoryLimitMB != 12 || got.TimeLimitSeconds != 3 || !got.EchoScript || got.SMT || got.Script != "script.abc" || got.Input != "input" || got.InputType != "bench" || got.OutputType != "pla" || !reflect.DeepEqual(got.Commands, wantCommands) {
		t.Fatalf("source-shaped arguments %+v", got)
	}
	got, err = ParseMainArguments([]string{"-S", "smt"}, MainLimits{})
	if err != nil || !got.SMT || !reflect.DeepEqual(got.Commands, []string{"smt"}) {
		t.Fatalf("SMT arguments %+v %v", got, err)
	}
	got, err = ParseMainArguments([]string{"-S", "smt", "-c", "normal"}, MainLimits{})
	if err != nil || got.SMT || !reflect.DeepEqual(got.Commands, []string{"smt", "normal"}) {
		t.Fatalf("last batch mode %+v %v", got, err)
	}
	for _, test := range []struct {
		args   []string
		limits MainLimits
		err    error
	}{
		{[]string{"-t", "json"}, MainLimits{}, ErrUnsupported},
		{[]string{"-T", "blif_mv"}, MainLimits{}, ErrUnsupported},
		{[]string{"-m", "bad"}, MainLimits{}, ErrDimension},
		{[]string{"bad\x00name"}, MainLimits{}, ErrFormat},
		{[]string{"12345"}, MainLimits{MaxBytes: 4}, ErrLimit},
	} {
		if _, err = ParseMainArguments(test.args, test.limits); !errors.Is(err, test.err) {
			t.Fatalf("arguments %q: got %v want %v", test.args, err, test.err)
		}
	}
	if _, err = MainVersion("bad\x00version", "date", "clock"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("version NUL %v", err)
	}
	if _, err = MainVersion(strings.Repeat("x", 990), "date", "clock"); !errors.Is(err, ErrLimit) {
		t.Fatalf("version buffer limit %v", err)
	}
}

func TestMainCopyWorkBudgets(t *testing.T) {
	frame, err := NewMainFrame("v", MainLimits{MaxOperations: 1, MaxBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	if err = frame.SetLibrary(MainLibrary{Name: "lib", Kind: "k", Data: []byte{1, 2}}); !errors.Is(err, ErrLimit) {
		t.Fatalf("library copy work %v", err)
	}
	if err = frame.SetCurrentNetwork(MainNetwork{ID: 1, Name: "n", Kind: MainNetworkLogic, Payload: []byte{1, 2}}, 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("network copy work %v", err)
	}
	cex := MainCounterexample{PIs: 1, Frame: 1, Bits: []bool{false, true}}
	if err = frame.SetVerification(1, 1, &cex); !errors.Is(err, ErrLimit) {
		t.Fatalf("counterexample copy work %v", err)
	}
	snapshot, err := frame.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Networks) != 0 || len(snapshot.Libraries) != 0 || snapshot.Verification.CEX != nil {
		t.Fatalf("failed copies were not atomic: %+v", snapshot)
	}
}

func TestMainIndependentStateModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(1401))
	frame := mainTestFrame(t)
	flags := make(map[string]string)
	order := make([]string, 0)
	networks := make([]uint64, 0)
	for step := 0; step < 5000; step++ {
		switch rng.Intn(4) {
		case 0:
			name := string(rune('a' + rng.Intn(12)))
			value := string(rune('0' + rng.Intn(10)))
			_, found := flags[name]
			if err := frame.SetFlag(name, value); err != nil {
				t.Fatal(err)
			}
			if !found {
				order = append(order, name)
			}
			flags[name] = value
		case 1:
			name := string(rune('a' + rng.Intn(12)))
			deleted, err := frame.DeleteFlag(name)
			if err != nil {
				t.Fatal(err)
			}
			_, found := flags[name]
			if deleted != found {
				t.Fatalf("delete mismatch")
			}
			if found {
				delete(flags, name)
				for i, v := range order {
					if v == name {
						order = append(order[:i], order[i+1:]...)
						break
					}
				}
			}
		case 2:
			id := uint64(step + 1)
			save := rng.Intn(4)
			if err := frame.SetCurrentNetwork(MainNetwork{ID: id, Name: "n", Kind: MainNetworkLogic, Payload: []byte{byte(step)}}, save); err != nil {
				t.Fatal(err)
			}
			keep := min(save, len(networks))
			next := []uint64{id}
			next = append(next, networks[:keep]...)
			networks = next
		case 3:
			snap, err := frame.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			if len(snap.Flags) != len(order) || len(snap.Networks) != len(networks) {
				t.Fatalf("sizes at %d", step)
			}
			for i, name := range order {
				if snap.Flags[i] != (MainFlag{Name: name, Value: flags[name]}) {
					t.Fatalf("flag model at %d", step)
				}
			}
			for i, id := range networks {
				if snap.Networks[i].ID != id {
					t.Fatalf("network model at %d", step)
				}
			}
		}
	}
}

func FuzzMainFrameState(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4})
	f.Add([]byte("abc-main-frame"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		frame, err := NewMainFrame("fuzz", MainLimits{MaxNetworks: 16, MaxFlags: 16, MaxLibraries: 8, MaxBytes: 1 << 16, MaxOperations: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		for i, b := range data {
			switch b % 4 {
			case 0:
				_ = frame.SetFlag(string(rune('a'+b%8)), string(rune('0'+b%10)))
			case 1:
				_, _ = frame.DeleteFlag(string(rune('a' + b%8)))
			case 2:
				_ = frame.SetCurrentNetwork(MainNetwork{ID: uint64(i + 1), Name: "f", Kind: MainNetworkKind(b % 3), Payload: []byte{b}}, int(b%4))
			case 3:
				_ = frame.SetLibrary(MainLibrary{Name: string(rune('a' + b%8)), Kind: "k", Data: []byte{b}})
			}
		}
		snapshot, err := frame.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range snapshot.Networks {
			if err := n.validate(); err != nil {
				t.Fatal(err)
			}
		}
		if err := frame.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
