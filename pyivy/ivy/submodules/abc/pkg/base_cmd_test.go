package abc

// Standalone Go tests for the command-processing port.

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestCmdStableIDsAndBuildInventory(t *testing.T) {
	names := []string{"", "alias", "source", "set", "read_blif", "read_aiger", "strash", "print_stats", "pdr", "write_aiger_cex"}
	for index, name := range names {
		if got := CmdCommandNameToID(name); int(got) != index {
			t.Fatalf("ID(%q)=%d want %d", name, got, index)
		}
	}
	if CmdCommandNameToID("missing") != CmdCommandUnknown {
		t.Fatal("unknown command acquired stable ID")
	}
	want := []string{"cmd.c", "cmdAlias.c", "cmdApi.c", "cmdAuto.c", "cmdFlag.c", "cmdHist.c", "cmdLoad.c", "cmdPlugin.c", "cmdStarter.c", "cmdUtils.c"}
	if got := CmdBuildInventory(); !reflect.DeepEqual(got, want) {
		t.Fatalf("inventory=%v", got)
	}
	got := CmdBuildInventory()
	got[0] = "changed"
	if CmdBuildInventory()[0] != "cmd.c" {
		t.Fatal("inventory aliases caller storage")
	}
	units := CmdSourceInventory()
	if len(units) != 10 || units[0].Source != "cmd.c" || units[9].Source != "module.make" || !units[9].Complete || len(units[0].Definitions) != 30 {
		t.Fatalf("source inventory=%+v", units)
	}
	units[0].Definitions[0] = "changed"
	if CmdSourceInventory()[0].Definitions[0] != "Cmd_Init" {
		t.Fatal("source inventory aliases caller storage")
	}
}

func TestParseCmdTextQuotesEscapesCommentsAndErrors(t *testing.T) {
	input := " echo  alpha 'two words' \"three  words\" four\\ five; set x 7 # ignored\n echo '#kept' ;\n"
	got, err := ParseCmdText(context.Background(), input, CmdLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"echo", "alpha", "two words", "three  words", "four five"}, {"set", "x", "7"}, {"echo", "#kept"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parse=%q want %q", got, want)
	}
	for _, input := range []string{"echo 'open", "echo trailing\\"} {
		if _, err = ParseCmdText(context.Background(), input, CmdLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("input %q error=%v", input, err)
		}
	}
	if _, err = ParseCmdText(context.Background(), "echo a b", CmdLimits{MaxArguments: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("argument limit error=%v", err)
	}
	if _, err = ParseCmdText(context.Background(), "echo a; echo b", CmdLimits{MaxCommands: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("command limit error=%v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = ParseCmdText(cancelled, "echo a", CmdLimits{}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation error=%v", err)
	}
	if got, err := ParseCmdText(nil, "echo nil-context", CmdLimits{}); err != nil || len(got) != 1 {
		t.Fatalf("nil context parse=%v err=%v", got, err)
	}
}

func TestCmdRegistryOrderReplacementOwnershipAndClose(t *testing.T) {
	registry, err := NewCmdRegistry(CmdLimits{MaxCommands: 2, MaxBytes: 64})
	if err != nil {
		t.Fatal(err)
	}
	handler := func(_ context.Context, _ *CmdExecution, args []string) (CmdResult, error) {
		return CmdResult{Output: strings.Join(args, ",")}, nil
	}
	if err = registry.Add(CmdDefinition{Name: "one", Group: "g", Handler: handler}); err != nil {
		t.Fatal(err)
	}
	if err = registry.Add(CmdDefinition{Name: "two", Group: "g", Handler: handler}); err != nil {
		t.Fatal(err)
	}
	if err = registry.Add(CmdDefinition{Name: "one", Group: "new", Handler: handler, Changes: true}); err != nil {
		t.Fatal(err)
	}
	definitions, err := registry.Definitions()
	if err != nil || len(definitions) != 2 || definitions[0].Name != "one" || definitions[0].Group != "new" || definitions[1].Name != "two" {
		t.Fatalf("definitions=%+v err=%v", definitions, err)
	}
	if err = registry.Add(CmdDefinition{Name: "three", Group: "g", Handler: handler}); !errors.Is(err, ErrLimit) {
		t.Fatalf("capacity error=%v", err)
	}
	if err = registry.Add(CmdDefinition{Name: "bad-name", Group: "g", Handler: handler}); !errors.Is(err, ErrDimension) {
		t.Fatalf("name error=%v", err)
	}
	registry.Close()
	if _, _, err = registry.Definition("one"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed lookup error=%v", err)
	}
}

func TestCmdSessionBuiltinsAliasesHistorySourceAndHelp(t *testing.T) {
	session, err := NewCmdSession(CmdLimits{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := session.Execute(context.Background(), "echo hello 'wide world'; set mode fast; set")
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "hello wide world\nmode=fast\n" {
		t.Fatalf("output=%q", result.Output)
	}
	result, err = session.Execute(context.Background(), "alias e echo; e aliased output")
	if err != nil || result.Output != "aliased output\n" {
		t.Fatalf("alias result=%+v err=%v", result, err)
	}
	if err = session.AddSource("script", "echo sourced; unset mode; echo done"); err != nil {
		t.Fatal(err)
	}
	result, err = session.Execute(context.Background(), "source script")
	if err != nil || result.Output != "sourced\ndone\n" {
		t.Fatalf("source result=%+v err=%v", result, err)
	}
	result, err = session.Execute(context.Background(), "set")
	if err != nil || strings.Contains(result.Output, "mode=") {
		t.Fatalf("unset output=%q err=%v", result.Output, err)
	}
	result, err = session.Execute(context.Background(), "echo recalled")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.Execute(context.Background(), "!!"); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("disabled history substitution error=%v", err)
	}
	result, err = session.Execute(context.Background(), "history")
	if err != nil || !strings.Contains(result.Output, "echo recalled\n") {
		t.Fatalf("history listing result=%+v err=%v", result, err)
	}
	result, err = session.Execute(context.Background(), "help Basic")
	if err != nil || !strings.Contains(result.Output, "Basic: echo\n") || !strings.Contains(result.Output, "Basic: source\n") {
		t.Fatalf("help=%q err=%v", result.Output, err)
	}
	result, err = session.Execute(context.Background(), "empty; quit; echo skipped")
	if err != nil || !result.Changed || !result.Quit || result.Output != "" || !session.QuitRequested() {
		t.Fatalf("quit result=%+v err=%v", result, err)
	}
}

func TestCmdAggregateRegistryBytesAndAtomicMultiDelete(t *testing.T) {
	session, err := NewCmdSession(CmdLimits{MaxBytes: 256, MaxOperations: 128})
	if err != nil {
		t.Fatal(err)
	}
	handler := func(context.Context, *CmdExecution, []string) (CmdResult, error) { return CmdResult{}, nil }
	if err = session.Registry().Add(CmdDefinition{Name: "large", Group: strings.Repeat("g", 120), Handler: handler}); err != nil {
		t.Fatal(err)
	}
	retained, err := session.retainedBytes()
	if err != nil || retained != session.registry.bytes+session.bytes {
		t.Fatalf("retained=%d registry=%d session=%d err=%v", retained, session.registry.bytes, session.bytes, err)
	}
	if err = session.AddSource("s", strings.Repeat("x", 256-retained)); !errors.Is(err, ErrLimit) {
		t.Fatalf("aggregate source limit error=%v retained=%d", err, retained)
	}
	other, err := NewCmdSession(CmdLimits{MaxBytes: 256})
	if err != nil {
		t.Fatal(err)
	}
	available, err := other.retainedBytes()
	if err != nil {
		t.Fatal(err)
	}
	if err = other.AddSource("held", strings.Repeat("x", 256-available-len("held")-2)); err != nil {
		t.Fatal(err)
	}
	if err = other.Registry().Add(CmdDefinition{Name: "later", Group: "group", Handler: handler}); !errors.Is(err, ErrLimit) {
		t.Fatalf("aggregate registry limit error=%v", err)
	}
	for _, command := range []string{"set a 1", "set b 2", "alias aa echo a", "alias bb echo b"} {
		if _, err = session.Execute(context.Background(), command); err != nil {
			t.Fatal(err)
		}
	}
	budget := &cmdBudget{ctx: context.Background(), limit: session.limits, ops: session.limits.MaxOperations - 1}
	execution := &CmdExecution{session: session, budget: budget}
	if _, err = cmdBuiltinUnset(context.Background(), execution, []string{"unset", "a", "b"}); !errors.Is(err, ErrLimit) {
		t.Fatalf("unset budget error=%v", err)
	}
	for _, name := range []string{"a", "b"} {
		if _, found, readErr := session.flags.Read(name); readErr != nil || !found {
			t.Fatalf("flag %q partially deleted found=%v err=%v", name, found, readErr)
		}
	}
	budget = &cmdBudget{ctx: context.Background(), limit: session.limits, ops: session.limits.MaxOperations - 1}
	execution = &CmdExecution{session: session, budget: budget}
	if _, err = cmdBuiltinUnalias(context.Background(), execution, []string{"unalias", "aa", "bb"}); !errors.Is(err, ErrLimit) {
		t.Fatalf("unalias budget error=%v", err)
	}
	for _, name := range []string{"aa", "bb"} {
		if _, found, aliasErr := session.aliases.Alias(name); aliasErr != nil || !found {
			t.Fatalf("alias %q partially deleted found=%v err=%v", name, found, aliasErr)
		}
	}
	budget = &cmdBudget{ctx: context.Background(), limit: session.limits}
	execution = &CmdExecution{session: session, budget: budget}
	if _, err = cmdBuiltinUnset(context.Background(), execution, []string{"unset", "a", "a"}); err != nil {
		t.Fatalf("duplicate unset: %v", err)
	}
	if _, found, readErr := session.flags.Read("a"); readErr != nil || found {
		t.Fatalf("duplicate unset retained flag found=%v err=%v", found, readErr)
	}
	budget = &cmdBudget{ctx: context.Background(), limit: session.limits}
	execution = &CmdExecution{session: session, budget: budget}
	if _, err = cmdBuiltinUnalias(context.Background(), execution, []string{"unalias", "aa", "aa"}); err != nil {
		t.Fatalf("duplicate unalias: %v", err)
	}
	if _, found, aliasErr := session.aliases.Alias("aa"); aliasErr != nil || found {
		t.Fatalf("duplicate unalias retained alias found=%v err=%v", found, aliasErr)
	}
}

func TestCmdCustomHandlerStableArgsAndChanges(t *testing.T) {
	session, err := NewCmdSession(CmdLimits{})
	if err != nil {
		t.Fatal(err)
	}
	var retained []string
	err = session.Registry().Add(CmdDefinition{Name: "capture", Group: "Test", Changes: true, Handler: func(_ context.Context, execution *CmdExecution, args []string) (CmdResult, error) {
		retained = args
		if execution.Session() != session {
			return CmdResult{}, ErrInvariant
		}
		if err := execution.Charge(2, 3); err != nil {
			return CmdResult{}, err
		}
		return CmdResult{Output: strings.Join(args[1:], ":") + "\n"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := session.Execute(context.Background(), "capture a b")
	if err != nil || result.Output != "a:b\n" || !result.Changed {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	retained[0] = "mutated"
	definition, found, err := session.Registry().Definition("capture")
	if err != nil || !found || definition.Name != "capture" {
		t.Fatalf("registry corrupted: %+v %v %v", definition, found, err)
	}
}

func TestCmdLimitsCancellationRecursionAndAtomicState(t *testing.T) {
	session, err := NewCmdSession(CmdLimits{MaxDepth: 3, MaxSources: 2, MaxText: 256, MaxBytes: 256, MaxOperations: 128})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.AddSource("loop", "source loop"); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Execute(context.Background(), "source loop"); !errors.Is(err, ErrLimit) {
		t.Fatalf("recursion error=%v", err)
	}
	if err = session.AddSource("two", "echo two"); err != nil {
		t.Fatal(err)
	}
	if err = session.AddSource("three", "echo three"); !errors.Is(err, ErrLimit) {
		t.Fatalf("source capacity error=%v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = session.Execute(cancelled, "set untouched value"); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel error=%v", err)
	}
	if value, found, readErr := session.flags.Read("untouched"); readErr != nil || found || value != "" {
		t.Fatalf("cancelled mutation value=%q found=%v err=%v", value, found, readErr)
	}
	if _, err = session.Execute(context.Background(), "alias a b; alias b a; a"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("alias loop error=%v", err)
	}
	session.Close()
	if _, err = session.Execute(context.Background(), "echo no"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed error=%v", err)
	}
	if err = session.AddSource("x", "echo x"); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed source error=%v", err)
	}
}

func TestCmdHistoryLimitPreflightsBeforeHandlerMutation(t *testing.T) {
	session, err := NewCmdSession(CmdLimits{MaxHistory: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.Execute(context.Background(), "echo first"); err != nil {
		t.Fatal(err)
	}
	called := false
	if err = session.Registry().Add(CmdDefinition{Name: "mark", Group: "Test", Handler: func(_ context.Context, _ *CmdExecution, _ []string) (CmdResult, error) {
		called = true
		return CmdResult{}, nil
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err = session.Execute(context.Background(), "mark"); !errors.Is(err, ErrLimit) {
		t.Fatalf("history limit error=%v", err)
	}
	if called {
		t.Fatal("handler ran before history capacity preflight")
	}
}

func TestCmdAutoChoices(t *testing.T) {
	choices, err := ParseCmdAutoChoices(context.Background(), "# comment\nb\nA1 A2.5\n c B3\nC4D5\n", CmdLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := []CmdAutoChoice{
		{Candidates: []CmdAutoCandidate{{Parameter: 'b', Binary: true}}},
		{Candidates: []CmdAutoCandidate{{Parameter: 'A', Value: 1}, {Parameter: 'A', Value: 2.5}}},
		{Candidates: []CmdAutoCandidate{{Parameter: 'C', Value: 4}, {Parameter: 'D', Value: 5}}},
	}
	if !reflect.DeepEqual(choices, want) {
		t.Fatalf("choices=%v want=%v", choices, want)
	}
	for _, text := range []string{"B A1", "A", "A-1", "?1"} {
		if _, err = ParseCmdAutoChoices(context.Background(), text, CmdLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("text %q error=%v", text, err)
		}
	}
	if _, err = ParseCmdAutoChoices(nil, "A1", CmdLimits{}); err != nil {
		t.Fatalf("nil context auto choices: %v", err)
	}
	if _, err = ParseCmdAutoChoices(context.Background(), "A1 A2", CmdLimits{MaxArguments: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("candidate limit error=%v", err)
	}
}

func TestCmdDeterministicIndependentModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0x434d44))
	for trial := 0; trial < 5000; trial++ {
		session, err := NewCmdSession(CmdLimits{MaxOperations: 1 << 16, MaxBytes: 1 << 20, MaxText: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		model := make(map[string]string)
		order := make([]string, 0)
		var script strings.Builder
		steps := 1 + rng.Intn(20)
		for step := 0; step < steps; step++ {
			name := fmt.Sprintf("v%d", rng.Intn(8))
			if rng.Intn(4) == 0 {
				script.WriteString("unset ")
				script.WriteString(name)
				script.WriteByte(';')
				delete(model, name)
				for i, existing := range order {
					if existing == name {
						order = append(order[:i], order[i+1:]...)
						break
					}
				}
				continue
			}
			value := fmt.Sprintf("x%d", rng.Intn(32))
			script.WriteString("set ")
			script.WriteString(name)
			script.WriteByte(' ')
			script.WriteString(value)
			script.WriteByte(';')
			if _, found := model[name]; !found {
				order = append(order, name)
			}
			model[name] = value
		}
		script.WriteString("set")
		result, err := session.Execute(context.Background(), script.String())
		if err != nil {
			t.Fatalf("trial %d: %v script=%q", trial, err, script.String())
		}
		var want strings.Builder
		for _, name := range order {
			if value, found := model[name]; found {
				want.WriteString(name)
				want.WriteByte('=')
				want.WriteString(value)
				want.WriteByte('\n')
			}
		}
		if result.Output != want.String() {
			t.Fatalf("trial %d output=%q want=%q", trial, result.Output, want.String())
		}
	}
}

func FuzzCmdParserAndSession(f *testing.F) {
	for _, seed := range []string{"echo hello", "echo 'a b'; set x 1; set", "alias e echo; e yes", "# comment\necho done", "echo a\\ b"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 1024 {
			t.Skip()
		}
		limits := CmdLimits{MaxCommands: 64, MaxArguments: 256, MaxOperations: 1 << 15, MaxBytes: 1 << 16, MaxText: 2048, MaxDepth: 16}
		parsed1, err1 := ParseCmdText(context.Background(), text, limits)
		parsed2, err2 := ParseCmdText(context.Background(), text, limits)
		if (err1 == nil) != (err2 == nil) || !reflect.DeepEqual(parsed1, parsed2) {
			t.Fatalf("nondeterministic parse: %v/%v %v/%v", parsed1, err1, parsed2, err2)
		}
		if err1 != nil {
			return
		}
		session, err := NewCmdSession(limits)
		if err != nil {
			t.Fatal(err)
		}
		_, err = session.Execute(context.Background(), text)
		if err != nil && !errors.Is(err, ErrDimension) && !errors.Is(err, ErrLimit) && !errors.Is(err, ErrInvariant) && !errors.Is(err, ErrUnsupported) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
