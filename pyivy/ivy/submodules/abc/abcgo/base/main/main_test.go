package abcmain

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func isolatedFrame(t *testing.T) *Frame {
	t.Helper()
	Abc_Stop()
	t.Setenv("ABC_LIB_PATH", "")
	f := NewFrame()
	f.Out = new(bytes.Buffer)
	f.Err = new(bytes.Buffer)
	f.In = strings.NewReader("")
	Abc_FrameSetGlobalFrame(f)
	t.Cleanup(Abc_Stop)
	return f
}
func isolateInitializers(t *testing.T) {
	t.Helper()
	initializerRegistry.Lock()
	saved := initializerRegistry.list
	initializerRegistry.list = nil
	initializerRegistry.Unlock()
	t.Cleanup(func() {
		Abc_Stop()
		initializerRegistry.Lock()
		initializerRegistry.list = saved
		initializerRegistry.Unlock()
	})
}
func TestLifecycleAndReentrantGlobalLookup(t *testing.T) {
	Abc_Stop()
	isolateInitializers(t)
	var events []string
	first := &Abc_FrameInitializer_t{Init: func(f *Frame) {
		if Abc_FrameGetGlobalFrame() != f {
			t.Fatal("wrong global during initialization")
		}
		events = append(events, "first init")
	}, Destroy: func(*Frame) { events = append(events, "first end") }}
	second := &Abc_FrameInitializer_t{Init: func(*Frame) { events = append(events, "second init") }, Destroy: func(*Frame) { events = append(events, "second end") }}
	Abc_FrameAddInitializer(first)
	Abc_FrameAddInitializer(first)
	Abc_FrameAddInitializer(second)
	Abc_Start()
	f := Abc_FrameReadGlobalFrame()
	if f == nil || f.Status != -1 || f.NFrames != -1 || !f.FlagEnabled("checkread") {
		t.Fatal("default state")
	}
	Abc_Start()
	Abc_FrameInit(f)
	Abc_Stop()
	Abc_Stop()
	if Abc_FrameReadGlobalFrame() != nil {
		t.Fatal("global retained after stop")
	}
	want := []string{"first init", "second init", "second end", "first end"}
	if !reflect.DeepEqual(events, want) {
		t.Fatal(events)
	}
}
func TestCommandsAliasesFlagsAndDispatcher(t *testing.T) {
	f := isolatedFrame(t)
	f.RegisterCommand(Command{Name: "record", Group: "Test", Handler: func(f *Frame, args []string) int32 { f.SetFlag("recorded", strings.Join(args[1:], "|")); return 0 }})
	if status := f.Execute(`set empty; set enabled 1; alias collect record; collect "one two" 'three;four'; echo "hello;world" # ignored`); status != 0 {
		t.Fatal(status, f.Err)
	}
	if value, _ := f.LookupFlag("recorded"); value != "one two|three;four" {
		t.Fatal(value)
	}
	if !f.FlagEnabled("empty") || !f.FlagEnabled("enabled") || f.FlagEnabled("missing") {
		t.Fatal("flag presence")
	}
	aliases := f.AliasList()
	aliases["collect"][0] = "bad"
	f.FlagList()["enabled"] = "0"
	if args, _ := f.LookupAlias("collect"); args[0] != "record" {
		t.Fatal("aliased snapshot")
	}
	if status := f.Execute("missing-command"); status != 1 {
		t.Fatal(status)
	}
	if status := f.Execute("alias loop loop; loop"); status != 1 {
		t.Fatal("alias cycle", status)
	}
	if status := f.Execute(`echo "unterminated`); status != 1 {
		t.Fatal(status)
	}
	f.SetDispatcher(testDispatcher{})
	if status := f.Execute("dispatch payload"); status != 7 {
		t.Fatal("dispatcher bypassed", status)
	}
}

type testDispatcher struct{}

func (testDispatcher) Execute(f *Frame, line string) int32 { f.SetFlag("dispatched", line); return 7 }

type testNetwork struct {
	value  int
	closed int
}

func (n *testNetwork) CloneFrameNetwork() any { return &testNetwork{value: n.value} }
func (n *testNetwork) Close() error           { n.closed++; return nil }
func TestNetworkBackupUndoRecallAndOwnership(t *testing.T) {
	f := isolatedFrame(t)
	f.SetFlag("savesteps", "2")
	a, b, c := &testNetwork{value: 1}, &testNetwork{value: 2}, &testNetwork{value: 3}
	f.SetCurrentNetwork(a)
	f.SetCurrentNetwork(b)
	f.SetCurrentNetwork(c)
	if f.BackupCount() != 2 {
		t.Fatal(f.BackupCount())
	}
	f.SwapCurrentAndBackup()
	if f.Network() != b {
		t.Fatal("swap failed")
	}
	if err := f.RecallNetwork(-1); err != nil {
		t.Fatal(err)
	}
	recalled := f.Network().(*testNetwork)
	if recalled == c || recalled.value != 3 {
		t.Fatal("recall must clone", recalled)
	}
	if err := f.RecallNetwork(9999); err == nil {
		t.Fatal("missing recall accepted")
	}
	f.DeleteNetworks()
	if recalled.closed != 1 || a.closed != 1 || b.closed != 1 || c.closed != 1 {
		t.Fatal("network ownership", a.closed, b.closed, c.closed, recalled.closed)
	}
}
func TestChangingCommandBacksUpBeforeMutation(t *testing.T) {
	f := isolatedFrame(t)
	original := &testNetwork{value: 1}
	f.SetCurrentNetwork(original)
	f.SetFlag("backup", "")
	f.RegisterCommand(Command{Name: "mutate", ChangesNetwork: true, Handler: func(f *Frame, args []string) int32 { f.Network().(*testNetwork).value = 2; return 0 }})
	if f.Execute("mutate") != 0 {
		t.Fatal(f.Err)
	}
	if original.value != 2 || f.BackupCount() != 1 {
		t.Fatal("mutation/backup")
	}
	f.SwapCurrentAndBackup()
	if f.Network().(*testNetwork).value != 1 {
		t.Fatal("backup was mutated")
	}
	elapsed, total := f.CommandTiming(true)
	if elapsed < 0 || total < elapsed {
		t.Fatal(elapsed, total)
	}
	elapsed, _ = f.CommandTiming(false)
	if elapsed != 0 {
		t.Fatal("timing not reset")
	}
}
func TestResourceTransferAndFrameSpecificState(t *testing.T) {
	f := isolatedFrame(t)
	resource := &testNetwork{}
	f.SetResource("pSave1", resource)
	if Abc_FrameReadSave1() != resource || Abc_FrameReadSave1() != nil {
		t.Fatal("take semantics")
	}
	other := NewFrame()
	defer other.Close()
	f.Status = 0
	other.Status = 1
	other.Cex = &Counterexample{PIs: 5, Registers: 2, Output: 3, Frame: 7}
	if Abc_FrameReadProbStatus(other) != 1 || Abc_FrameReadCexPiNum(other) != 5 || Abc_FrameReadCexFrame(other) != 7 {
		t.Fatal("getter used global frame")
	}
	f.Close()
	if resource.closed != 0 {
		t.Fatal("transferred resource closed")
	}
	resource.Close()
}
func TestSourceScriptAndCLI(t *testing.T) {
	f := isolatedFrame(t)
	path := filepath.Join(t.TempDir(), "commands with spaces.abc")
	if err := os.WriteFile(path, []byte("set sourced 1\necho script-output\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if status := f.Execute("source " + strconvQuote(path)); status != 0 {
		t.Fatal(status, f.Err)
	}
	if !f.FlagEnabled("sourced") || f.SourceDepth() != 0 {
		t.Fatal("source state")
	}
	status := Abc_RealMain([]string{"abc", "-s", "-q", "echo cli-output;set cli 1"})
	if status != 0 {
		t.Fatal(status, f.Err)
	}
	if !strings.Contains(f.Out.(*bytes.Buffer).String(), "cli-output") || Abc_FrameReadGlobalFrame() != nil {
		t.Fatal("CLI lifecycle/output")
	}
}
func TestInteractiveBufferedInputPreserved(t *testing.T) {
	f := isolatedFrame(t)
	f.In = strings.NewReader("echo first\necho second\nquit\n")
	if status := Abc_RealMain([]string{"abc", "-s"}); status != 0 {
		t.Fatal(status)
	}
	out := f.Out.(*bytes.Buffer).String()
	if !strings.Contains(out, "first") || !strings.Contains(out, "second") {
		t.Fatal(out)
	}
}
func TestMissingAdaptersReportErrors(t *testing.T) {
	f := isolatedFrame(t)
	if err := Abc_NtkInputMiniAig(f, new(int)); err == nil {
		t.Fatal("unsupported bridge silently accepted")
	}
	if _, err := Abc_FrameGiaOutputMiniLut(f); err == nil {
		t.Fatal("unsupported export accepted")
	}
	if err := Abc_NtkSetCiArrivalTime(f, 0, 1, 2); err == nil {
		t.Fatal("timing adapter missing")
	}
	if status, err := RunModelChecking(f, "file.aig", ModelCheckOptions{}); status != -1 || err == nil {
		t.Fatal(status, err)
	}
	f.SetCurrentNetwork(42)
	f.SetFlag("backup", "")
	called := false
	f.RegisterCommand(Command{Name: "change", ChangesNetwork: true, Handler: func(*Frame, []string) int32 { called = true; return 0 }})
	if status := f.Execute("change"); status != 1 || called {
		t.Fatal("missing backup adapter ignored")
	}
}
func TestDateFormatting(t *testing.T) {
	if got := DateReadFromDateString("Sun Sep 16 01:03:52 EDT 2001"); got != "16-Sep-01 at 1:03 AM" {
		t.Fatal(got)
	}
	if got := DateReadFromDateString("not a date"); got != "not a date" {
		t.Fatal(got)
	}
}

type testModelChecker struct {
	calls  []int
	result int32
	closed bool
}

func (c *testModelChecker) ReadAiger(string) (any, error) { return c, nil }
func (c *testModelChecker) RegisterCount(any) int         { return 1 }
func (c *testModelChecker) BMC(_ any, frames, size, limit int) (int32, int32, error) {
	c.calls = append(c.calls, frames)
	if frames == 200 {
		return c.result, 12, nil
	}
	return -1, 0, nil
}
func (c *testModelChecker) SEC(any, time.Duration, int) (int32, error) { return -1, nil }
func (c *testModelChecker) WriteCounterexample(w io.Writer, _ any) error {
	_, err := io.WriteString(w, "witness\n")
	return err
}
func (c *testModelChecker) Close(any) error { c.closed = true; return nil }
func TestModelCheckWorkflow(t *testing.T) {
	f := isolatedFrame(t)
	checker := &testModelChecker{result: 0}
	f.ModelChecker = checker
	status, err := RunModelChecking(f, "example.aig", ModelCheckOptions{})
	if err != nil || status != 0 || !checker.closed || !reflect.DeepEqual(checker.calls, []int{50, 200}) {
		t.Fatal(status, err, checker)
	}
	if !strings.Contains(f.Out.(*bytes.Buffer).String(), "1\nwitness") {
		t.Fatal(f.Out)
	}
}
func TestModuleLibraryLifecycle(t *testing.T) {
	f := isolatedFrame(t)
	name := t.Name()
	var events []string
	if err := RegisterLibrary(&Library{Name: name, Init: func(*Frame) { events = append(events, "init") }, End: func(*Frame) { events = append(events, "end") }}); err != nil {
		t.Fatal(err)
	}
	if err := f.LoadLibraries([]string{name}); err != nil {
		t.Fatal(err)
	}
	f.UnloadLibraries()
	if !reflect.DeepEqual(events, []string{"init", "end"}) {
		t.Fatal(events)
	}
	if err := f.LoadLibraries([]string{"missing-extension"}); err == nil {
		t.Fatal("missing loader accepted")
	}
}

func TestReentrantLifecycleCallbacks(t *testing.T) {
	Abc_Stop()
	isolateInitializers(t)
	calls := 0
	Abc_FrameAddInitializer(&Abc_FrameInitializer_t{Init: func(f *Frame) { calls++; Abc_FrameInit(f) }, Destroy: func(f *Frame) { calls++; Abc_FrameEnd(f); f.Close() }})
	Abc_Start()
	Abc_Stop()
	if calls != 2 {
		t.Fatal(calls)
	}
}
func TestAliasLimitPerCommand(t *testing.T) {
	f := isolatedFrame(t)
	f.RegisterAlias("say", []string{"echo", "ok"})
	line := strings.Repeat("say;", 40)
	if status := f.Execute(line); status != 0 {
		t.Fatal(status, f.Err)
	}
	if strings.Count(f.Out.(*bytes.Buffer).String(), "ok\n") != 40 {
		t.Fatal(f.Out)
	}
}

func TestClosedFrameRejectsStateMutation(t *testing.T) {
	f := isolatedFrame(t)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	f.SetStringValue("x", "value")
	f.SetIntValues("x", []int32{1})
	f.SetListValues("x", []any{1})
	f.SetRows("x", [][]int32{{1}})
	f.AppendStore(1)
	f.SetResource("x", 1)
	if f.StringValue("x") != "" || len(f.IntValues("x")) != 0 || len(f.ListValues("x")) != 0 || len(f.Rows("x")) != 0 || f.Resource("x") != nil {
		t.Fatal("closed frame retained state")
	}
}
