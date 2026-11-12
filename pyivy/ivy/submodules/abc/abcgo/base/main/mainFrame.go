package abcmain

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/glycerine/abcgo/bool/dec"
)

var globalFrame struct {
	sync.RWMutex
	frame *Frame
}

func NewFrame() *Frame {
	f := &Frame{Version: ABC_VERSION, In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Status: -1, NFrames: -1, steps: 1, commands: map[string]Command{}, aliases: map[string][]string{}, flags: map[string]string{"checkread": ""}, resources: map[string]any{}, strings: map[string]string{}, ints: map[string][]int32{}, lists: map[string][]any{}, rows: map[string][][]int32{}}
	registerBuiltins(f)
	return f
}
func Abc_FrameAllocate() *Frame { return NewFrame() }
func Abc_FrameGetGlobalFrame() *Frame {
	globalFrame.Lock()
	f := globalFrame.frame
	created := f == nil
	if created {
		f = NewFrame()
		globalFrame.frame = f
	}
	globalFrame.Unlock()
	if created {
		Abc_FrameInit(f)
	}
	return f
}
func Abc_FrameReadGlobalFrame() *Frame {
	globalFrame.RLock()
	defer globalFrame.RUnlock()
	return globalFrame.frame
}
func Abc_FrameSetGlobalFrame(f *Frame) {
	globalFrame.Lock()
	globalFrame.frame = f
	globalFrame.Unlock()
}
func Abc_FrameDeallocate(f *Frame) {
	if f == nil {
		return
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(f.Err, err)
	}
}
func resourceNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}
func sameResource(a, b any) bool {
	if resourceNil(a) || resourceNil(b) {
		return resourceNil(a) && resourceNil(b)
	}
	return reflect.TypeOf(a) == reflect.TypeOf(b) && reflect.TypeOf(a).Comparable() && a == b
}
func closeResource(value any) error {
	if resourceNil(value) {
		return nil
	}
	if closer, ok := value.(io.Closer); ok {
		return closer.Close()
	}
	if closer, ok := value.(interface{ CloseFrameResource() error }); ok {
		return closer.CloseFrameResource()
	}
	return nil
}
func (f *Frame) Close() error {
	f.mu.Lock()
	if f.closed || f.closing {
		f.mu.Unlock()
		return nil
	}
	f.closing = true
	f.mu.Unlock()
	Abc_FrameEnd(f)
	f.lifecycle.Lock()
	defer f.lifecycle.Unlock()
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return nil
	}
	f.closed = true
	networks := append([]networkEntry{f.current}, f.backups...)
	f.current = networkEntry{}
	f.backups = nil
	resources := make([]any, 0, len(f.resources))
	for _, value := range f.resources {
		resources = append(resources, value)
	}
	f.resources = map[string]any{}
	manager := f.manDec
	f.manDec = nil
	f.commands = nil
	f.aliases = nil
	f.flags = nil
	f.history = nil
	f.ints = nil
	f.lists = nil
	f.rows = nil
	f.strings = nil
	f.Cex = nil
	f.mu.Unlock()
	globalFrame.Lock()
	if globalFrame.frame == f {
		globalFrame.frame = nil
	}
	globalFrame.Unlock()
	var failures []error
	var closed []any
	release := func(value any, network bool) {
		if resourceNil(value) {
			return
		}
		for _, old := range closed {
			if sameResource(old, value) {
				return
			}
		}
		closed = append(closed, value)
		var err error
		if network && f.Networks != nil {
			err = f.Networks.Close(value)
		} else {
			err = closeResource(value)
		}
		if err != nil {
			failures = append(failures, err)
		}
	}
	for _, network := range networks {
		release(network.value, true)
	}
	for _, resource := range resources {
		release(resource, false)
	}
	if manager != nil {
		dec.Dec_ManStop(manager)
	}
	return errors.Join(failures...)
}
func (f *Frame) RegisterCommand(command Command) error {
	if command.Name == "" || command.Handler == nil {
		return errors.New("abcmain: command requires a name and handler")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return errors.New("abcmain: closed frame")
	}
	f.commands[command.Name] = command
	return nil
}
func (f *Frame) LookupCommand(name string) (Command, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	command, ok := f.commands[name]
	return command, ok
}
func (f *Frame) RemoveCommand(name string) { f.mu.Lock(); delete(f.commands, name); f.mu.Unlock() }
func (f *Frame) CommandList() []Command {
	f.mu.RLock()
	commands := make([]Command, 0, len(f.commands))
	for _, command := range f.commands {
		commands = append(commands, command)
	}
	f.mu.RUnlock()
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name < commands[j].Name })
	return commands
}
func (f *Frame) RegisterAlias(name string, args []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		panic("abcmain: closed frame")
	}
	f.aliases[name] = append([]string(nil), args...)
}
func (f *Frame) LookupAlias(name string) ([]string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	args, ok := f.aliases[name]
	return append([]string(nil), args...), ok
}
func (f *Frame) RemoveAlias(name string) { f.mu.Lock(); delete(f.aliases, name); f.mu.Unlock() }
func (f *Frame) AliasList() map[string][]string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	aliases := make(map[string][]string, len(f.aliases))
	for name, args := range f.aliases {
		aliases[name] = append([]string(nil), args...)
	}
	return aliases
}
func (f *Frame) SetFlag(name, value string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		panic("abcmain: closed frame")
	}
	f.flags[name] = value
}
func (f *Frame) LookupFlag(name string) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	value, ok := f.flags[name]
	return value, ok
}
func (f *Frame) DeleteFlag(name string) { f.mu.Lock(); delete(f.flags, name); f.mu.Unlock() }
func (f *Frame) FlagList() map[string]string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	flags := make(map[string]string, len(f.flags))
	for name, value := range f.flags {
		flags[name] = value
	}
	return flags
}
func (f *Frame) FlagEnabled(name string) bool {
	value, exists := f.LookupFlag(name)
	return exists && (value == "" || value == "1")
}
func (f *Frame) AppendHistory(line string) {
	f.mu.Lock()
	f.history = append(f.history, line)
	f.mu.Unlock()
}
func (f *Frame) SetHistory(lines []string) {
	f.mu.Lock()
	f.history = append([]string(nil), lines...)
	f.mu.Unlock()
}
func (f *Frame) History() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]string(nil), f.history...)
}
func (f *Frame) BeginSource() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.sourceDepth >= 100 {
		return errors.New("abcmain: script nesting limit exceeded")
	}
	f.sourceDepth++
	return nil
}
func (f *Frame) EndSource() {
	f.mu.Lock()
	if f.sourceDepth > 0 {
		f.sourceDepth--
	}
	f.mu.Unlock()
}
func (f *Frame) SourceDepth() int { f.mu.RLock(); defer f.mu.RUnlock(); return f.sourceDepth }
func (f *Frame) EnterAutoexec() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.autoexec {
		return false
	}
	f.autoexec = true
	return true
}
func (f *Frame) LeaveAutoexec()     { f.mu.Lock(); f.autoexec = false; f.mu.Unlock() }
func (f *Frame) Network() any       { f.mu.RLock(); defer f.mu.RUnlock(); return f.current.value }
func (f *Frame) NetworkStep() int32 { f.mu.RLock(); defer f.mu.RUnlock(); return f.current.step }
func (f *Frame) closeNetwork(value any) error {
	if resourceNil(value) {
		return nil
	}
	if f.Networks != nil {
		return f.Networks.Close(value)
	}
	return closeResource(value)
}
func (f *Frame) SetCurrentNetwork(network any) {
	if resourceNil(network) {
		return
	}
	f.mu.Lock()
	if sameResource(f.current.value, network) {
		f.mu.Unlock()
		return
	}
	var backups []networkEntry
	for _, entry := range f.backups {
		if !sameResource(entry.value, network) {
			backups = append(backups, entry)
		}
	}
	if !resourceNil(f.current.value) {
		backups = append([]networkEntry{f.current}, backups...)
	}
	f.steps++
	f.current = networkEntry{network, f.steps}
	limit := 1
	if value, exists := f.flags["savesteps"]; exists {
		limit, _ = strconv.Atoi(value)
		if limit < 0 {
			limit = 0
		}
	}
	var removed []networkEntry
	if len(backups) > limit {
		removed = backups[limit:]
		backups = backups[:limit]
	}
	f.backups = backups
	f.mu.Unlock()
	for _, entry := range removed {
		if err := f.closeNetwork(entry.value); err != nil {
			fmt.Fprintln(f.Err, err)
		}
	}
}
func (f *Frame) SwapCurrentAndBackup() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.backups) == 0 {
		return
	}
	old := f.current
	backup := f.backups[0]
	f.current = networkEntry{backup.value, old.step}
	f.backups[0] = networkEntry{old.value, backup.step}
}
func (f *Frame) ReplaceCurrentNetwork(network any) {
	if resourceNil(network) {
		return
	}
	backupEnabled := f.FlagEnabled("backup")
	f.mu.Lock()
	if sameResource(f.current.value, network) {
		f.mu.Unlock()
		return
	}
	removed := []networkEntry{f.current}
	step := f.current.step
	if !backupEnabled || resourceNil(f.current.value) {
		f.steps++
		step = f.steps
		removed = append(removed, f.backups...)
		f.backups = nil
	}
	f.current = networkEntry{network, step}
	f.mu.Unlock()
	for _, entry := range removed {
		if err := f.closeNetwork(entry.value); err != nil {
			fmt.Fprintln(f.Err, err)
		}
	}
}
func (f *Frame) PrepareNetworkBackup() error {
	network := f.Network()
	if resourceNil(network) || !f.FlagEnabled("backup") {
		return nil
	}
	var copy any
	var err error
	if f.Networks != nil {
		copy, err = f.Networks.Clone(network)
	} else if cloner, ok := network.(interface{ CloneFrameNetwork() any }); ok {
		copy = cloner.CloneFrameNetwork()
	} else {
		return errors.New("abcmain: changing command requires a network cloning adapter for backup")
	}
	if err != nil {
		return err
	}
	if resourceNil(copy) || sameResource(copy, network) {
		return errors.New("abcmain: network cloning adapter returned no independent copy")
	}
	f.SetCurrentNetwork(copy)
	f.SwapCurrentAndBackup()
	return nil
}
func (f *Frame) DeleteNetworks() {
	f.mu.Lock()
	entries := append([]networkEntry{f.current}, f.backups...)
	f.current = networkEntry{}
	f.backups = nil
	f.mu.Unlock()
	for _, entry := range entries {
		if err := f.closeNetwork(entry.value); err != nil {
			fmt.Fprintln(f.Err, err)
		}
	}
	for _, key := range []string{"pGia", "pGia2", "pGiaBest", "pGiaBest2", "pGiaSaved"} {
		f.SetResource(key, nil)
	}
}
func (f *Frame) UnmapNetworks() error {
	if f.Networks == nil {
		return errors.New("abcmain: unmapping requires a network adapter")
	}
	f.mu.RLock()
	entries := append([]networkEntry{f.current}, f.backups...)
	f.mu.RUnlock()
	for _, entry := range entries {
		if !resourceNil(entry.value) {
			if err := f.Networks.Unmap(entry.value); err != nil {
				return err
			}
		}
	}
	return nil
}
func (f *Frame) SetResource(name string, value any) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return
	}
	old := f.resources[name]
	f.resources[name] = value
	f.mu.Unlock()
	if !sameResource(old, value) {
		if err := closeResource(old); err != nil {
			fmt.Fprintln(f.Err, err)
		}
	}
}
func (f *Frame) Resource(name string) any {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.resources[name]
}
func (f *Frame) TakeResource(name string) any {
	f.mu.Lock()
	defer f.mu.Unlock()
	value := f.resources[name]
	delete(f.resources, name)
	return value
}
func (f *Frame) Dispatch(args []string) int32 {
	if len(args) == 0 {
		return 0
	}
	command, exists := f.LookupCommand(args[0])
	if !exists {
		fmt.Fprintf(f.Err, "unknown command: %s\n", args[0])
		return 1
	}
	if command.ChangesNetwork {
		if err := f.PrepareNetworkBackup(); err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
	}
	start := time.Now()
	status := command.Handler(f, append([]string(nil), args...))
	duration := time.Since(start)
	f.mu.Lock()
	f.TimeCommand += duration
	f.TimeTotal += duration
	f.mu.Unlock()
	return status
}
func (f *Frame) Execute(line string) int32 {
	f.mu.RLock()
	closed, dispatcher := f.closed, f.Dispatcher
	f.mu.RUnlock()
	if closed {
		fmt.Fprintln(f.Err, "abcmain: closed frame")
		return 1
	}
	if dispatcher != nil {
		return dispatcher.Execute(f, line)
	}
	return executeDefault(f, line, 0)
}
func Cmd_CommandExecute(f *Frame, line string) int32       { return f.Execute(line) }
func Abc_FrameReadNtk(f *Frame) any                        { return f.Network() }
func Abc_FrameSetCurrentNetwork(f *Frame, network any)     { f.SetCurrentNetwork(network) }
func Abc_FrameSwapCurrentAndBackup(f *Frame)               { f.SwapCurrentAndBackup() }
func Abc_FrameReplaceCurrentNetwork(f *Frame, network any) { f.ReplaceCurrentNetwork(network) }
func Abc_FrameDeleteAllNetworks(f *Frame)                  { f.DeleteNetworks() }
func Abc_FrameUnmapAllNetworks(f *Frame) {
	if err := f.UnmapNetworks(); err != nil {
		fmt.Fprintln(f.Err, err)
	}
}
func Abc_FrameClearVerifStatus(f *Frame) {
	f.mu.Lock()
	f.Status, f.NFrames, f.Cex = -1, -1, nil
	f.mu.Unlock()
}
func Abc_FrameReadOut(f *Frame) io.Writer { return f.Out }
func Abc_FrameReadErr(f *Frame) io.Writer { return f.Err }
func Abc_FrameReadMode(f *Frame) int32 {
	value, _ := f.LookupFlag("namemode")
	mode, _ := strconv.ParseInt(value, 10, 32)
	return int32(mode)
}
func Abc_FrameSetMode(f *Frame, mode int32) int32 {
	old := Abc_FrameReadMode(f)
	f.SetFlag("namemode", strconv.Itoa(int(mode)))
	return old
}
func Abc_FrameShowProgress(f *Frame) int32 {
	if f.FlagEnabled("progressbar") {
		return 1
	}
	return 0
}

// Restart is intentionally an empty extension hook in the original C source.
// Explicit module restart is available through Abc_FrameEnd then Abc_FrameInit.
func Abc_FrameRestart(f *Frame) {}

func (f *Frame) RecallNetwork(step int32) error {
	if step == 0 {
		f.DeleteNetworks()
		return nil
	}
	f.mu.RLock()
	var network any
	if step < 0 {
		index := int(-step - 1)
		if index < len(f.backups) {
			network = f.backups[index].value
		}
	} else {
		if f.current.step == step {
			network = f.current.value
		}
		for _, entry := range f.backups {
			if entry.step == step {
				network = entry.value
				break
			}
		}
	}
	f.mu.RUnlock()
	if resourceNil(network) {
		return fmt.Errorf("abcmain: no network at saved step %d", step)
	}
	var copy any
	var err error
	if f.Networks != nil {
		copy, err = f.Networks.Clone(network)
	} else if cloner, ok := network.(interface{ CloneFrameNetwork() any }); ok {
		copy = cloner.CloneFrameNetwork()
	} else {
		return errors.New("abcmain: recall requires a network cloning adapter")
	}
	if err != nil {
		return err
	}
	if resourceNil(copy) || sameResource(copy, network) {
		return errors.New("abcmain: recalled network is not an independent clone")
	}
	f.SetCurrentNetwork(copy)
	return nil
}
func (f *Frame) CommandTiming(clear bool) (elapsed, total time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	elapsed, total = f.TimeCommand, f.TimeTotal
	if clear {
		f.TimeCommand = 0
	}
	return
}
func (f *Frame) SetDispatcher(dispatcher Dispatcher) {
	f.mu.Lock()
	f.Dispatcher = dispatcher
	f.mu.Unlock()
}
