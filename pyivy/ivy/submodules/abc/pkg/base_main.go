package abc

// Owned main-frame behavior from src/base/main.

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/glycerine/insdict"
)

var ErrMainClosed = errors.New("main frame is closed")

// MainBuildInventory is the exact default module.make implementation list.
var MainBuildInventory = [...]string{
	"main.c", "mainFrame.c", "mainInit.c", "mainLib.c", "mainReal.c", "libSupport.c", "mainUtils.c",
}

type MainSourceUnit struct {
	Source       string
	TrackedNames int
	Complete     bool
	Remaining    string
}

// MainSourceInventory records every assigned source unit. TrackedNames counts
// the definitions or declarations discovered during the source audit; repeated
// declarations across headers are intentionally counted in their owning file.
var MainSourceInventory = [...]MainSourceUnit{
	{"abcapis.h", 33, false, "opaque MiniAIG, MiniLUT, NDR, mapping and native frame adapters"},
	{"abcapis_old.h", 19, false, "legacy opaque API adapters"},
	{"libSupport.c", 12, false, "dynamic library loading and symbol callbacks"},
	{"main.c", 1, false, "process entry point"},
	{"main.h", 88, false, "full native pointer and global-frame declaration surface"},
	{"mainFrame.c", 87, false, "native ABC/GIA managers, raw libraries and process-global singleton"},
	{"mainInit.c", 5, false, "callback package initialization; deterministic names and order are represented"},
	{"mainInt.h", 111, false, "raw frame layout, streams and callback fields"},
	{"mainLib.c", 2, false, "process-global start and stop; owned MainRuntime is represented"},
	{"mainMC.c", 1, false, "alternate process entry point"},
	{"mainReal.c", 9, false, "external command execution and filesystem I/O; argument semantics are represented"},
	{"mainUtils.c", 6, false, "interactive input, filesystem resource sourcing and process streams"},
	{"module.make", len(MainBuildInventory), true, ""},
}

type MainLimits struct {
	MaxNetworks, MaxLibraries, MaxFlags, MaxSubsystems int
	MaxBytes, MaxText, MaxOperations, MaxDepth         int
	Stop                                               func() bool
}

func normalizeMainLimits(in MainLimits) (MainLimits, error) {
	values := []*int{&in.MaxNetworks, &in.MaxLibraries, &in.MaxFlags, &in.MaxSubsystems, &in.MaxBytes, &in.MaxText, &in.MaxOperations, &in.MaxDepth}
	defaults := [...]int{1024, 1024, 4096, 256, 64 << 20, 4 << 20, 1 << 24, 1024}
	for i, value := range values {
		if *value < 0 {
			return MainLimits{}, ErrDimension
		}
		if *value == 0 {
			*value = defaults[i]
		}
	}
	return in, nil
}

type mainMeter struct {
	limits MainLimits
	work   int
}

func (m *mainMeter) take(amount int) error {
	if amount < 0 || m.work > m.limits.MaxOperations-amount {
		return ErrLimit
	}
	m.work += amount
	if m.limits.Stop != nil && m.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func mainAdd(a, b int) (int, bool) {
	if a < 0 || b < 0 || a > math.MaxInt-b {
		return 0, false
	}
	return a + b, true
}

func mainTextSize(parts ...string) (int, error) {
	total := 0
	for _, part := range parts {
		var ok bool
		if total, ok = mainAdd(total, len(part)); !ok {
			return 0, ErrLimit
		}
	}
	return total, nil
}

type MainNetworkKind uint8

const (
	MainNetworkLogic MainNetworkKind = iota
	MainNetworkStrashed
	MainNetworkGIA
)

type MainNetwork struct {
	ID             uint64
	Name           string
	Kind           MainNetworkKind
	Mapped         bool
	PrimaryOutputs []int8 // -1 means nonconstant; 0 and 1 are constants.
	Payload        []byte
	Step           uint64
}

func (network MainNetwork) clone() MainNetwork {
	result := network
	result.PrimaryOutputs = append([]int8(nil), network.PrimaryOutputs...)
	result.Payload = append([]byte(nil), network.Payload...)
	return result
}

func (network MainNetwork) validate() error {
	if network.ID == 0 || network.Name == "" || strings.IndexByte(network.Name, 0) >= 0 || network.Kind > MainNetworkGIA {
		return ErrInvariant
	}
	for _, value := range network.PrimaryOutputs {
		if value < -1 || value > 1 {
			return ErrInvariant
		}
	}
	return nil
}

func (network MainNetwork) bytes() (int, error) {
	result, ok := mainAdd(len(network.Name), len(network.Payload))
	if !ok {
		return 0, ErrLimit
	}
	result, ok = mainAdd(result, len(network.PrimaryOutputs))
	if !ok {
		return 0, ErrLimit
	}
	return result, nil
}

func mainNetworkSameInput(left, right MainNetwork) bool {
	if left.ID != right.ID || left.Name != right.Name || left.Kind != right.Kind || left.Mapped != right.Mapped || len(left.PrimaryOutputs) != len(right.PrimaryOutputs) || len(left.Payload) != len(right.Payload) {
		return false
	}
	for i := range left.PrimaryOutputs {
		if left.PrimaryOutputs[i] != right.PrimaryOutputs[i] {
			return false
		}
	}
	for i := range left.Payload {
		if left.Payload[i] != right.Payload[i] {
			return false
		}
	}
	return true
}

func mainNetworksBytes(networks []MainNetwork) (int, error) {
	total := 0
	for _, network := range networks {
		size, err := network.bytes()
		if err != nil {
			return 0, err
		}
		var ok bool
		if total, ok = mainAdd(total, size); !ok {
			return 0, ErrLimit
		}
	}
	return total, nil
}

func mainNetworkCopyWork(network MainNetwork) (int, error) {
	work, ok := mainAdd(len(network.PrimaryOutputs), len(network.Payload))
	if !ok {
		return 0, ErrLimit
	}
	return work, nil
}

type MainLibrary struct {
	Name, Kind string
	Data       []byte
}

func (library MainLibrary) clone() MainLibrary {
	result := library
	result.Data = append([]byte(nil), library.Data...)
	return result
}

func (library MainLibrary) validate() error {
	if library.Name == "" || library.Kind == "" || strings.IndexByte(library.Name, 0) >= 0 || strings.IndexByte(library.Kind, 0) >= 0 {
		return ErrInvariant
	}
	return nil
}

func (library MainLibrary) bytes() (int, error) {
	size, err := mainTextSize(library.Name, library.Kind)
	if err != nil {
		return 0, err
	}
	if size > math.MaxInt-len(library.Data) {
		return 0, ErrLimit
	}
	return size + len(library.Data), nil
}

type MainCounterexample struct {
	PIs, Registers, Output, Frame int
	Bits                          []bool
}

func (counterexample MainCounterexample) clone() MainCounterexample {
	result := counterexample
	result.Bits = append([]bool(nil), counterexample.Bits...)
	return result
}

func (counterexample MainCounterexample) validate() error {
	if counterexample.PIs < 0 || counterexample.Registers < 0 || counterexample.Output < 0 || counterexample.Frame < 0 {
		return ErrDimension
	}
	if counterexample.Frame == math.MaxInt || counterexample.PIs > 0 && counterexample.Frame+1 > math.MaxInt/counterexample.PIs {
		return ErrLimit
	}
	minimum, ok := mainAdd(counterexample.Registers, counterexample.PIs*(counterexample.Frame+1))
	if !ok {
		return ErrLimit
	}
	if len(counterexample.Bits) < minimum {
		return ErrInvariant
	}
	return nil
}

type MainVerification struct {
	Status int
	Frames int
	CEX    *MainCounterexample
}

type MainLogEntry struct{ Kind, Text string }
type MainFlag struct{ Name, Value string }

type MainFrameSnapshot struct {
	Version, SpecName, DrivingCell string
	Batch, Bridge, Initialized     bool
	NameMode, Steps                int
	MaxLoad                        float64
	Networks                       []MainNetwork
	Libraries                      []MainLibrary
	Flags                          []MainFlag
	Subsystems                     []string
	Logs                           []MainLogEntry
	Verification                   MainVerification
}

type MainFrame struct {
	mu           sync.RWMutex
	limits       MainLimits
	closed       bool
	initialized  bool
	version      string
	steps        int
	batch        bool
	bridge       bool
	nameMode     int
	specName     string
	drivingCell  string
	maxLoad      float64
	retained     int
	networks     []MainNetwork // current first, followed by backups
	flags        *insdict.Dict[string, string]
	libraries    *insdict.Dict[string, MainLibrary]
	subsystems   []string
	logs         []MainLogEntry
	verification MainVerification
}

func NewMainFrame(version string, limits MainLimits) (*MainFrame, error) {
	normalized, err := normalizeMainLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(version) > normalized.MaxText || strings.IndexByte(version, 0) >= 0 {
		return nil, ErrLimit
	}
	if len(version) > normalized.MaxBytes {
		return nil, ErrLimit
	}
	if normalized.Stop != nil && normalized.Stop() {
		return nil, ErrCancelled
	}
	frame := &MainFrame{
		limits: normalized, version: version, steps: 1,
		flags:        insdict.NewDictSize[string, string](min(normalized.MaxFlags, 1024)),
		libraries:    insdict.NewDictSize[string, MainLibrary](min(normalized.MaxLibraries, 1024)),
		verification: MainVerification{Status: -1, Frames: -1},
	}
	frame.retained = len(version)
	return frame, nil
}

func (frame *MainFrame) ensureOpen() error {
	if frame == nil || frame.closed || frame.flags == nil || frame.libraries == nil {
		return ErrMainClosed
	}
	return nil
}

func (frame *MainFrame) checkRetained(delta int) error {
	if delta < -frame.retained || delta > frame.limits.MaxBytes-frame.retained {
		return ErrLimit
	}
	return nil
}

func (frame *MainFrame) RegisterSubsystem(name string) error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if name == "" || strings.IndexByte(name, 0) >= 0 {
		return ErrInvariant
	}
	if len(frame.subsystems) >= frame.limits.MaxSubsystems || len(name) > frame.limits.MaxText {
		return ErrLimit
	}
	meter := mainMeter{limits: frame.limits}
	work, ok := mainAdd(len(frame.subsystems)+1, len(name))
	if !ok {
		return ErrLimit
	}
	if err := meter.take(work); err != nil {
		return err
	}
	for _, existing := range frame.subsystems {
		if existing == name {
			return ErrInvariant
		}
	}
	if err := frame.checkRetained(len(name)); err != nil {
		return err
	}
	frame.subsystems = append(frame.subsystems, name)
	frame.retained += len(name)
	return nil
}

func (frame *MainFrame) Initialize() error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if frame.initialized {
		return nil
	}
	meter := mainMeter{limits: frame.limits}
	for range frame.subsystems {
		if err := meter.take(1); err != nil {
			return err
		}
	}
	frame.initialized = true
	return nil
}

func (frame *MainFrame) End() error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if !frame.initialized {
		return nil
	}
	meter := mainMeter{limits: frame.limits}
	for range frame.subsystems {
		if err := meter.take(1); err != nil {
			return err
		}
	}
	frame.initialized = false
	return nil
}

func (frame *MainFrame) Close() error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if frame == nil || frame.closed {
		return ErrMainClosed
	}
	frame.closed = true
	frame.version = ""
	frame.specName = ""
	frame.drivingCell = ""
	frame.networks = nil
	frame.subsystems = nil
	frame.logs = nil
	frame.flags = nil
	frame.libraries = nil
	frame.verification = MainVerification{Status: -1, Frames: -1}
	frame.retained = 0
	return nil
}

func (frame *MainFrame) SetFlag(name, value string) error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	return frame.setFlagLocked(name, value)
}

func (frame *MainFrame) setFlagLocked(name, value string) error {
	if name == "" || strings.IndexByte(name, 0) >= 0 || strings.IndexByte(value, 0) >= 0 {
		return ErrInvariant
	}
	if len(name) > frame.limits.MaxText || len(value) > frame.limits.MaxText {
		return ErrLimit
	}
	work, ok := mainAdd(len(name), len(value))
	if !ok {
		return ErrLimit
	}
	meter := mainMeter{limits: frame.limits}
	if err := meter.take(work); err != nil {
		return err
	}
	mode := frame.nameMode
	if name == "namemode" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 || parsed > 9 {
			return ErrDimension
		}
		mode = parsed
	}
	old, found := frame.flags.Get2(name)
	if !found && frame.flags.Len() >= frame.limits.MaxFlags {
		return ErrLimit
	}
	delta := len(value) - len(old)
	if !found {
		delta += len(name)
	}
	if err := frame.checkRetained(delta); err != nil {
		return err
	}
	if frame.limits.Stop != nil && frame.limits.Stop() {
		return ErrCancelled
	}
	frame.flags.Set(name, value)
	frame.retained += delta
	frame.nameMode = mode
	return nil
}

func (frame *MainFrame) DeleteFlag(name string) (bool, error) {
	if frame == nil {
		return false, ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return false, err
	}
	old, found := frame.flags.Get2(name)
	if !found {
		return false, nil
	}
	if frame.limits.Stop != nil && frame.limits.Stop() {
		return false, ErrCancelled
	}
	frame.flags.Del(name)
	frame.retained -= len(name) + len(old)
	if name == "namemode" {
		frame.nameMode = 0
	}
	return true, nil
}

func (frame *MainFrame) Flag(name string) (string, bool, error) {
	if frame == nil {
		return "", false, ErrMainClosed
	}
	frame.mu.RLock()
	defer frame.mu.RUnlock()
	if err := frame.ensureOpen(); err != nil {
		return "", false, err
	}
	value, found := frame.flags.Get2(name)
	return value, found, nil
}

func (frame *MainFrame) FlagEnabled(name string) (bool, error) {
	value, found, err := frame.Flag(name)
	return found && (value == "" || value == "1"), err
}

func (frame *MainFrame) SetNameMode(mode int) (int, error) {
	if frame == nil {
		return 0, ErrMainClosed
	}
	if mode < 0 || mode > 9 {
		return 0, ErrDimension
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return 0, err
	}
	old := frame.nameMode
	if err := frame.setFlagLocked("namemode", strconv.Itoa(mode)); err != nil {
		return old, err
	}
	return old, nil
}

func (frame *MainFrame) SetCurrentNetwork(network MainNetwork, saveSteps int) error {
	if frame == nil {
		return ErrMainClosed
	}
	if saveSteps < 0 {
		return ErrDimension
	}
	if err := network.validate(); err != nil {
		return err
	}
	networkBytes, err := network.bytes()
	if err != nil {
		return err
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if len(network.Name) > frame.limits.MaxText || networkBytes > frame.limits.MaxBytes {
		return ErrLimit
	}
	for i, existing := range frame.networks {
		if existing.ID != network.ID {
			continue
		}
		if i == 0 && mainNetworkSameInput(existing, network) {
			return nil
		}
		return ErrInvariant
	}
	keep := min(saveSteps, len(frame.networks))
	if keep+1 > frame.limits.MaxDepth {
		return ErrLimit
	}
	if keep+1 > frame.limits.MaxNetworks {
		return ErrLimit
	}
	meter := mainMeter{limits: frame.limits}
	copyWork, err := mainNetworkCopyWork(network)
	if err != nil {
		return err
	}
	work, ok := mainAdd(keep+1, copyWork)
	if !ok {
		return ErrLimit
	}
	if err := meter.take(work); err != nil {
		return err
	}
	if frame.steps == math.MaxInt {
		return ErrLimit
	}
	copyNetwork := network.clone()
	copyNetwork.Step = uint64(frame.steps + 1)
	newNetworks := make([]MainNetwork, 1, keep+1)
	newNetworks[0] = copyNetwork
	newNetworks = append(newNetworks, frame.networks[:keep]...)
	oldBytes, err := mainNetworksBytes(frame.networks)
	if err != nil {
		return err
	}
	newBytes, err := mainNetworksBytes(newNetworks)
	if err != nil {
		return err
	}
	if err := frame.checkRetained(newBytes - oldBytes); err != nil {
		return err
	}
	frame.networks = newNetworks
	frame.retained += newBytes - oldBytes
	frame.steps++
	return nil
}

func (frame *MainFrame) SwapCurrentAndBackup() error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if len(frame.networks) < 2 {
		return nil
	}
	if frame.limits.Stop != nil && frame.limits.Stop() {
		return ErrCancelled
	}
	frame.networks[0], frame.networks[1] = frame.networks[1], frame.networks[0]
	frame.networks[0].Step, frame.networks[1].Step = frame.networks[1].Step, frame.networks[0].Step
	return nil
}

func (frame *MainFrame) ReplaceCurrentNetwork(network MainNetwork, preserveBackups bool) error {
	if frame == nil {
		return ErrMainClosed
	}
	if err := network.validate(); err != nil {
		return err
	}
	networkBytes, err := network.bytes()
	if err != nil {
		return err
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if len(network.Name) > frame.limits.MaxText || networkBytes > frame.limits.MaxBytes {
		return ErrLimit
	}
	for i, existing := range frame.networks {
		if i > 0 && existing.ID == network.ID && preserveBackups {
			return ErrInvariant
		}
	}
	start := len(frame.networks)
	if preserveBackups && len(frame.networks) > 0 {
		start = 1
	}
	newNetworks := make([]MainNetwork, 1, len(frame.networks)-start+1)
	newNetworks = append(newNetworks, frame.networks[start:]...)
	if len(newNetworks) > frame.limits.MaxNetworks || len(newNetworks) > frame.limits.MaxDepth {
		return ErrLimit
	}
	copyWork, err := mainNetworkCopyWork(network)
	if err != nil {
		return err
	}
	work, ok := mainAdd(len(newNetworks), copyWork)
	if !ok {
		return ErrLimit
	}
	meter := mainMeter{limits: frame.limits}
	if err := meter.take(work); err != nil {
		return err
	}
	nextSteps := frame.steps
	copyNetwork := network.clone()
	if preserveBackups && len(frame.networks) > 0 {
		copyNetwork.Step = frame.networks[0].Step
		newNetworks[0] = copyNetwork
	} else {
		if nextSteps == math.MaxInt {
			return ErrLimit
		}
		nextSteps++
		copyNetwork.Step = uint64(nextSteps)
		newNetworks[0] = copyNetwork
	}
	oldBytes, err := mainNetworksBytes(frame.networks)
	if err != nil {
		return err
	}
	newBytes, err := mainNetworksBytes(newNetworks)
	if err != nil {
		return err
	}
	if err := frame.checkRetained(newBytes - oldBytes); err != nil {
		return err
	}
	if frame.limits.Stop != nil && frame.limits.Stop() {
		return ErrCancelled
	}
	frame.networks = newNetworks
	frame.retained += newBytes - oldBytes
	frame.steps = nextSteps
	return nil
}

func (frame *MainFrame) DeleteAllNetworks() error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	meter := mainMeter{limits: frame.limits}
	bytes := 0
	for _, n := range frame.networks {
		if err := meter.take(1); err != nil {
			return err
		}
		s, e := n.bytes()
		if e != nil {
			return e
		}
		var ok bool
		if bytes, ok = mainAdd(bytes, s); !ok {
			return ErrLimit
		}
	}
	frame.networks = nil
	frame.retained -= bytes
	return nil
}

func (frame *MainFrame) UnmapAllNetworks() error {
	if frame == nil {
		return ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	m := mainMeter{limits: frame.limits}
	copyNetworks := make([]MainNetwork, len(frame.networks))
	copy(copyNetworks, frame.networks)
	for i := range copyNetworks {
		if err := m.take(1); err != nil {
			return err
		}
		copyNetworks[i].Mapped = false
	}
	frame.networks = copyNetworks
	return nil
}

func (frame *MainFrame) CheckPOConstant(index int) (int, error) {
	if frame == nil {
		return -1, ErrMainClosed
	}
	frame.mu.RLock()
	defer frame.mu.RUnlock()
	if err := frame.ensureOpen(); err != nil {
		return -1, err
	}
	if len(frame.networks) == 0 || frame.networks[0].Kind != MainNetworkStrashed || index < 0 || index >= len(frame.networks[0].PrimaryOutputs) {
		return -1, nil
	}
	return int(frame.networks[0].PrimaryOutputs[index]), nil
}

func (frame *MainFrame) SetLibrary(library MainLibrary) error {
	if frame == nil {
		return ErrMainClosed
	}
	if err := library.validate(); err != nil {
		return err
	}
	size, err := library.bytes()
	if err != nil {
		return err
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if len(library.Name) > frame.limits.MaxText || len(library.Kind) > frame.limits.MaxText || size > frame.limits.MaxBytes {
		return ErrLimit
	}
	old, found := frame.libraries.Get2(library.Name)
	oldSize := 0
	if found {
		oldSize, err = old.bytes()
		if err != nil {
			return err
		}
	} else if frame.libraries.Len() >= frame.limits.MaxLibraries {
		return ErrLimit
	}
	if err := frame.checkRetained(size - oldSize); err != nil {
		return err
	}
	meter := mainMeter{limits: frame.limits}
	if err := meter.take(len(library.Data)); err != nil {
		return err
	}
	copyLibrary := library.clone()
	frame.libraries.Set(copyLibrary.Name, copyLibrary)
	frame.retained += size - oldSize
	return nil
}

func (frame *MainFrame) TakeLibrary(name string) (MainLibrary, bool, error) {
	if frame == nil {
		return MainLibrary{}, false, ErrMainClosed
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return MainLibrary{}, false, err
	}
	library, found := frame.libraries.Get2(name)
	if !found {
		return MainLibrary{}, false, nil
	}
	if frame.limits.Stop != nil && frame.limits.Stop() {
		return MainLibrary{}, false, ErrCancelled
	}
	frame.libraries.Del(name)
	size, _ := library.bytes()
	frame.retained -= size
	return library.clone(), true, nil
}

func (frame *MainFrame) SetVerification(status, frames int, counterexample *MainCounterexample) error {
	if frame == nil {
		return ErrMainClosed
	}
	if frames < -1 {
		return ErrDimension
	}
	bytes := 0
	if counterexample != nil {
		if err := counterexample.validate(); err != nil {
			return err
		}
		bytes = len(counterexample.Bits)
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	oldBytes := 0
	if frame.verification.CEX != nil {
		oldBytes = len(frame.verification.CEX.Bits)
	}
	if err := frame.checkRetained(bytes - oldBytes); err != nil {
		return err
	}
	meter := mainMeter{limits: frame.limits}
	if err := meter.take(bytes); err != nil {
		return err
	}
	var copied *MainCounterexample
	if counterexample != nil {
		value := counterexample.clone()
		copied = &value
	}
	frame.verification = MainVerification{Status: status, Frames: frames, CEX: copied}
	frame.retained += bytes - oldBytes
	return nil
}

func (frame *MainFrame) ClearVerification() error { return frame.SetVerification(-1, -1, nil) }

func (frame *MainFrame) AddLog(kind, text string) error {
	if frame == nil {
		return ErrMainClosed
	}
	if kind == "" || strings.IndexByte(kind, 0) >= 0 || strings.IndexByte(text, 0) >= 0 {
		return ErrInvariant
	}
	size, err := mainTextSize(kind, text)
	if err != nil {
		return err
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if len(kind) > frame.limits.MaxText || len(text) > frame.limits.MaxText || size > frame.limits.MaxText || len(frame.logs) >= frame.limits.MaxOperations {
		return ErrLimit
	}
	if err := frame.checkRetained(size); err != nil {
		return err
	}
	meter := mainMeter{limits: frame.limits}
	if err := meter.take(size); err != nil {
		return err
	}
	frame.logs = append(frame.logs, MainLogEntry{Kind: kind, Text: text})
	frame.retained += size
	return nil
}

func (frame *MainFrame) SetSession(batch, bridge bool, specName, drivingCell string, maxLoad float64) error {
	if frame == nil {
		return ErrMainClosed
	}
	if strings.IndexByte(specName, 0) >= 0 || strings.IndexByte(drivingCell, 0) >= 0 || math.IsNaN(maxLoad) || math.IsInf(maxLoad, 0) || maxLoad < 0 {
		return ErrDimension
	}
	newText, err := mainTextSize(specName, drivingCell)
	if err != nil {
		return err
	}
	frame.mu.Lock()
	defer frame.mu.Unlock()
	if err := frame.ensureOpen(); err != nil {
		return err
	}
	if newText > frame.limits.MaxText {
		return ErrLimit
	}
	oldText := len(frame.specName) + len(frame.drivingCell)
	if err := frame.checkRetained(newText - oldText); err != nil {
		return err
	}
	meter := mainMeter{limits: frame.limits}
	if err := meter.take(newText); err != nil {
		return err
	}
	frame.batch = batch
	frame.bridge = bridge
	frame.specName = specName
	frame.drivingCell = drivingCell
	frame.maxLoad = maxLoad
	frame.retained += newText - oldText
	return nil
}

func (frame *MainFrame) Snapshot() (MainFrameSnapshot, error) {
	if frame == nil {
		return MainFrameSnapshot{}, ErrMainClosed
	}
	frame.mu.RLock()
	defer frame.mu.RUnlock()
	if err := frame.ensureOpen(); err != nil {
		return MainFrameSnapshot{}, err
	}
	total := len(frame.networks)
	var ok bool
	if total, ok = mainAdd(total, frame.flags.Len()); !ok {
		return MainFrameSnapshot{}, ErrLimit
	}
	if total, ok = mainAdd(total, frame.libraries.Len()); !ok {
		return MainFrameSnapshot{}, ErrLimit
	}
	if total, ok = mainAdd(total, len(frame.subsystems)); !ok {
		return MainFrameSnapshot{}, ErrLimit
	}
	if total, ok = mainAdd(total, len(frame.logs)); !ok {
		return MainFrameSnapshot{}, ErrLimit
	}
	copyWork := 0
	for _, network := range frame.networks {
		work, err := mainNetworkCopyWork(network)
		if err != nil {
			return MainFrameSnapshot{}, err
		}
		if copyWork, ok = mainAdd(copyWork, work); !ok {
			return MainFrameSnapshot{}, ErrLimit
		}
	}
	for _, library := range frame.libraries.All() {
		if copyWork, ok = mainAdd(copyWork, len(library.Data)); !ok {
			return MainFrameSnapshot{}, ErrLimit
		}
	}
	if frame.verification.CEX != nil {
		if copyWork, ok = mainAdd(copyWork, len(frame.verification.CEX.Bits)); !ok {
			return MainFrameSnapshot{}, ErrLimit
		}
	}
	if total, ok = mainAdd(total, copyWork); !ok {
		return MainFrameSnapshot{}, ErrLimit
	}
	m := mainMeter{limits: frame.limits}
	if err := m.take(total); err != nil {
		return MainFrameSnapshot{}, err
	}
	result := MainFrameSnapshot{Version: frame.version, SpecName: frame.specName, DrivingCell: frame.drivingCell, Batch: frame.batch, Bridge: frame.bridge, Initialized: frame.initialized, NameMode: frame.nameMode, Steps: frame.steps, MaxLoad: frame.maxLoad, Verification: frame.verification}
	result.Networks = make([]MainNetwork, len(frame.networks))
	for i, n := range frame.networks {
		result.Networks[i] = n.clone()
	}
	for name, value := range frame.flags.All() {
		result.Flags = append(result.Flags, MainFlag{Name: name, Value: value})
	}
	for _, library := range frame.libraries.All() {
		result.Libraries = append(result.Libraries, library.clone())
	}
	for _, subsystem := range frame.subsystems {
		result.Subsystems = append(result.Subsystems, subsystem)
	}
	for _, entry := range frame.logs {
		result.Logs = append(result.Logs, entry)
	}
	if result.Verification.CEX != nil {
		copy := result.Verification.CEX.clone()
		result.Verification.CEX = &copy
	}
	return result, nil
}

type MainRuntime struct {
	mu    sync.Mutex
	frame *MainFrame
}

func (runtime *MainRuntime) Start(version string, limits MainLimits) (*MainFrame, error) {
	if runtime == nil {
		return nil, ErrMainClosed
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.frame != nil {
		return nil, ErrInvariant
	}
	frame, err := NewMainFrame(version, limits)
	if err != nil {
		return nil, err
	}
	if err = frame.Initialize(); err != nil {
		return nil, err
	}
	runtime.frame = frame
	return frame, nil
}
func (runtime *MainRuntime) Frame() *MainFrame {
	if runtime == nil {
		return nil
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.frame
}
func (runtime *MainRuntime) Stop() error {
	if runtime == nil {
		return ErrMainClosed
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.frame == nil {
		return ErrMainClosed
	}
	if err := runtime.frame.End(); err != nil {
		return err
	}
	if err := runtime.frame.Close(); err != nil {
		return err
	}
	runtime.frame = nil
	return nil
}

type MainInvocation struct {
	Commands                                     []string
	Script, Output, Input, InputType, OutputType string
	Quiet, Interactive, SkipSource, Bridge, Help bool
	Debug, SMT, EchoScript                       bool
	MemoryLimitMB, TimeLimitSeconds              int
}

func mainFileType(value string) bool {
	return value == "blif" || value == "bench" || value == "pla" || value == "none"
}

func ParseMainArguments(arguments []string, limits MainLimits) (MainInvocation, error) {
	normalized, err := normalizeMainLimits(limits)
	if err != nil {
		return MainInvocation{}, err
	}
	if len(arguments) > normalized.MaxOperations {
		return MainInvocation{}, ErrLimit
	}
	inputBytes := 0
	for _, argument := range arguments {
		if strings.IndexByte(argument, 0) >= 0 {
			return MainInvocation{}, ErrFormat
		}
		var ok bool
		if inputBytes, ok = mainAdd(inputBytes, len(argument)); !ok || inputBytes > normalized.MaxBytes {
			return MainInvocation{}, ErrLimit
		}
	}
	result := MainInvocation{InputType: "blif_mv", OutputType: "blif_mv"}
	m := mainMeter{limits: normalized}
	take := func(i *int) (string, error) {
		*i++
		if *i >= len(arguments) {
			return "", ErrFormat
		}
		return arguments[*i], nil
	}
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		if err := m.take(1 + len(arg)); err != nil {
			return MainInvocation{}, err
		}
		if len(arg) > normalized.MaxText {
			return MainInvocation{}, ErrLimit
		}
		if arg == "--" {
			if i+2 < len(arguments) {
				return MainInvocation{}, ErrFormat
			}
			if i+1 < len(arguments) {
				result.Input = arguments[i+1]
			}
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			if result.Input != "" {
				return MainInvocation{}, ErrFormat
			}
			result.Input = arg
			continue
		}
		switch arg {
		case "-d":
			result.Debug = !result.Debug
		case "-m", "-l":
			value, e := take(&i)
			if e != nil {
				return MainInvocation{}, e
			}
			number, e := strconv.Atoi(value)
			if e != nil || number < 0 {
				return MainInvocation{}, ErrDimension
			}
			if arg == "-m" {
				result.MemoryLimitMB = number
			} else {
				result.TimeLimitSeconds = number
			}
		case "-c", "-q", "-C", "-Q":
			value, e := take(&i)
			if e != nil {
				return MainInvocation{}, e
			}
			result.Commands = append(result.Commands, value)
			result.Quiet = arg == "-q" || arg == "-Q"
			result.Interactive = arg == "-C" || arg == "-Q"
			result.SMT = false
		case "-S":
			value, e := take(&i)
			if e != nil {
				return MainInvocation{}, e
			}
			result.Commands = append(result.Commands, value)
			result.SMT = true
			result.Quiet = false
			result.Interactive = false
		case "-f", "-F":
			value, e := take(&i)
			if e != nil {
				return MainInvocation{}, e
			}
			result.Script = value
			result.EchoScript = arg == "-F"
			result.SMT = false
			result.Quiet = false
			result.Interactive = false
			prefix := "source "
			if result.EchoScript {
				prefix = "source -x "
			}
			result.Commands = append(result.Commands, prefix+value)
		case "-o":
			value, e := take(&i)
			if e != nil {
				return MainInvocation{}, e
			}
			result.Output = value
		case "-t":
			value, e := take(&i)
			if e != nil {
				return MainInvocation{}, e
			}
			if !mainFileType(value) {
				return MainInvocation{}, ErrUnsupported
			}
			result.InputType = value
			result.SMT = false
			result.Quiet = false
			result.Interactive = false
		case "-T":
			value, e := take(&i)
			if e != nil {
				return MainInvocation{}, e
			}
			if !mainFileType(value) {
				return MainInvocation{}, ErrUnsupported
			}
			result.OutputType = value
			result.SMT = false
			result.Quiet = false
			result.Interactive = false
		case "-h":
			result.Help = true
		case "-s":
			result.SkipSource = true
		case "-x":
			result.InputType = "none"
			result.OutputType = "none"
			result.SMT = false
			result.Quiet = false
			result.Interactive = false
		case "-b":
			result.Bridge = true
		default:
			return MainInvocation{}, ErrUnsupported
		}
	}
	textSize, err := mainTextSize(result.Script, result.Output, result.Input, result.InputType, result.OutputType)
	if err != nil {
		return MainInvocation{}, err
	}
	for _, command := range result.Commands {
		var ok bool
		textSize, ok = mainAdd(textSize, len(command))
		if !ok {
			return MainInvocation{}, ErrLimit
		}
	}
	if textSize > normalized.MaxText {
		return MainInvocation{}, ErrLimit
	}
	return result, nil
}

func MainPrompt(step int) (string, error) {
	if step < 0 {
		return "", ErrDimension
	}
	return fmt.Sprintf("abc %02d> ", step), nil
}
func MainVersion(base, date, clock string) (string, error) {
	if strings.IndexByte(base, 0) >= 0 || strings.IndexByte(date, 0) >= 0 || strings.IndexByte(clock, 0) >= 0 {
		return "", ErrInvariant
	}
	size, err := mainTextSize(base, date, clock)
	if err != nil {
		return "", ErrLimit
	}
	size, ok := mainAdd(size, len(" (compiled  )"))
	if !ok || size > 999 {
		return "", ErrLimit
	}
	return fmt.Sprintf("%s (compiled %s %s)", base, date, clock), nil
}

func MainBriefDate(value string) string {
	var day, month, zone string
	var date, hour, minute, second, year int
	if _, err := fmt.Sscanf(value, "%s %s %d %d:%d:%d %s %d", &day, &month, &date, &hour, &minute, &second, &zone, &year); err != nil {
		return value
	}
	suffix := "AM"
	if hour >= 12 {
		suffix = "PM"
		if hour >= 13 {
			hour -= 12
		}
	} else if hour == 0 {
		hour = 12
	}
	return fmt.Sprintf("%d-%3s-%02d at %d:%02d %s", date, month, year%100, hour, minute, suffix)
}
