package scl

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"io"
	"testing"
)

func upsizeFixture(t *testing.T) (*SC_Lib, *downsizeNetwork) {
	lib, net := downsizeFixture(t)
	for i, c := range lib.Cells {
		delay := float32(3)
		if i == 1 {
			delay = 1
		}
		for _, g := range c.Pins[1].Timings {
			for _, timing := range g.Timings {
				for _, s := range []*SC_Surface{&timing.CellRise, &timing.CellFall, &timing.RiseTrans, &timing.FallTrans} {
					s.Data[0][0] = delay
				}
			}
		}
	}
	net.mapping.GateIDs[1], net.mapping.GateIDs[2] = 0, 0
	return lib, net
}
func TestUpsizeCriticalChain(t *testing.T) {
	lib, net := upsizeFixture(t)
	p, e := Abc_SclManStart(lib, net, 0, 1, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer Abc_SclManFree(p)
	p.Output = io.Discard
	if p.MaxDelay != 6 {
		t.Fatal(p.MaxDelay)
	}
	if e = UpsizeManager(p, &SC_SizePars{Iterations: 10, IterationsNoChange: 5, Window: 1, Ratio: 100, Notches: 4, UseDepartures: true}); e != nil {
		t.Fatal(e)
	}
	if p.MaxDelay != 2 || p.SumArea != 4 || net.mapping.GateIDs[1] != 1 || net.mapping.GateIDs[2] != 1 {
		t.Fatal(p.MaxDelay, p.SumArea, net.mapping.GateIDs)
	}
}
func TestUpsizeWithoutDepartureAndMissingMutable(t *testing.T) {
	lib, net := upsizeFixture(t)
	p, e := Abc_SclManStart(lib, net, 0, 0, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer Abc_SclManFree(p)
	p.Output = io.Discard
	if e = UpsizeManager(p, &SC_SizePars{Iterations: 10, Window: 1, Ratio: 100, Notches: 4}); e != nil {
		t.Fatal(e)
	}
	if p.MaxDelay != 2 {
		t.Fatal(p.MaxDelay)
	}
	if e = UpsizeManager(p, &SC_SizePars{Window: 1, BypassFrequency: 1}); e == nil {
		t.Fatal("bypass silently accepted immutable network")
	}
	if e = UpsizeManager(p, &SC_SizePars{Window: 0}); e == nil {
		t.Fatal("accepted zero window")
	}
}
func TestUpsizeCandidateRestoresTiming(t *testing.T) {
	lib, net := upsizeFixture(t)
	p, e := Abc_SclManStart(lib, net, 0, 1, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer Abc_SclManFree(p)
	outputs := Abc_SclFindCriticalCoWindow(p, 1)
	path, critical := Abc_SclFindCriticalNodeWindow(p, outputs, 1, true)
	if len(path) != 2 {
		t.Fatal(path)
	}
	nodes, evals, e := Abc_SclFindNodesToUpdate(net, 1, critical, -1)
	if e != nil {
		t.Fatal(e)
	}
	before := append([]SC_Pair(nil), p.Times...)
	var gain float32
	cell, e := Abc_SclFindBestCell(p, 1, nodes, evals, 4, 0, &gain)
	if e != nil {
		t.Fatal(e)
	}
	if cell != 1 || gain != 2 || net.mapping.GateIDs[1] != 0 {
		t.Fatal(cell, gain, net.mapping.GateIDs)
	}
	for i, v := range before {
		if p.Times[i] != v {
			t.Fatal("trial leaked timing", i, v, p.Times[i])
		}
	}
	if Abc_SclCountNearCriticalNodes(p) != 2 {
		t.Fatal("near critical count")
	}
}

type mutableUpsizeNetwork struct {
	mapping   NetworkMapping
	fanins    map[int][]int
	deleted   map[int]bool
	failPatch bool
}

func (n *mutableUpsizeNetwork) SclName() string { return "mutable" }
func (n *mutableUpsizeNetwork) SclObjectIDs() []int {
	var ids []int
	for i := 0; i < 5; i++ {
		if !n.deleted[i] {
			ids = append(ids, i)
		}
	}
	return ids
}
func (n *mutableUpsizeNetwork) SclInputs() []int       { return []int{0} }
func (n *mutableUpsizeNetwork) SclOutputs() []int      { return []int{4} }
func (n *mutableUpsizeNetwork) SclFanins(id int) []int { return n.fanins[id] }
func (n *mutableUpsizeNetwork) SclFanouts(id int) []int {
	var out []int
	for _, node := range n.SclObjectIDs() {
		for _, fanin := range n.fanins[node] {
			if fanin == id {
				out = append(out, node)
			}
		}
	}
	return out
}
func (n *mutableUpsizeNetwork) SclIsNode(id int) bool       { return id > 0 && id < 4 && !n.deleted[id] }
func (n *mutableUpsizeNetwork) SclIsBarrier(id int) bool    { return false }
func (n *mutableUpsizeNetwork) SclMapping() *NetworkMapping { return &n.mapping }
func (n *mutableUpsizeNetwork) SclPatchFanin(node, pin, id int) error {
	if n.failPatch {
		return fmt.Errorf("injected patch error")
	}
	n.fanins[node][pin] = id
	return nil
}
func (n *mutableUpsizeNetwork) SclDeleteNode(id int) error {
	if len(n.SclFanouts(id)) != 0 {
		return fmt.Errorf("deleting used node")
	}
	n.deleted[id] = true
	delete(n.fanins, id)
	return nil
}
func mutableUpsizeFixture(t *testing.T) (*SC_Man, *mutableUpsizeNetwork) {
	t.Helper()
	lib, _ := upsizeFixture(t)
	mioLib, e := Abc_SclDeriveGenlibSimple(lib)
	if e != nil {
		t.Fatal(e)
	}
	net := &mutableUpsizeNetwork{mapping: NetworkMapping{Lib: lib, Mio: mioLib, GateIDs: []int32{-1, 0, 0, 0, -1}}, fanins: map[int][]int{1: {0}, 2: {1}, 3: {2}, 4: {3}}, deleted: map[int]bool{}}
	p, e := Abc_SclManStart(lib, net, 0, 1, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	p.Output = io.Discard
	return p, net
}
func TestUpsizeBypassAndUndo(t *testing.T) {
	p, net := mutableUpsizeFixture(t)
	defer Abc_SclManFree(p)
	outputs := Abc_SclFindCriticalCoWindow(p, 1)
	path, critical := Abc_SclFindCriticalNodeWindow(p, outputs, 1, true)
	state := &UpsizeState{Critical: critical}
	count, e := Abc_SclFindBypasses(p, path, state, 100, 4, 1, 0, false)
	if e != nil {
		t.Fatal(e)
	}
	if count != 1 || net.fanins[3][0] != 1 || net.mapping.GateIDs[1] != 1 || len(state.Edits) != 1 {
		t.Fatal(count, net.fanins, net.mapping.GateIDs, state.Edits)
	}
	if e = Abc_SclTimeNtkRecompute(p, &p.SumArea, &p.MaxDelay, 1, 0); e != nil {
		t.Fatal(e)
	}
	if p.MaxDelay != 4 {
		t.Fatal(p.MaxDelay)
	}
	if e = Abc_SclUndoRecentChanges(p, state.Edits); e != nil {
		t.Fatal(e)
	}
	net.mapping.GateIDs[1] = 0
	if e = Abc_SclTimeNtkRecompute(p, nil, &p.MaxDelay, 1, 0); e != nil {
		t.Fatal(e)
	}
	if p.MaxDelay != 9 || net.fanins[3][0] != 2 {
		t.Fatal(p.MaxDelay, net.fanins)
	}
	if e = net.SclPatchFanin(3, 0, 1); e != nil {
		t.Fatal(e)
	}
	if e = Abc_SclUpsizeRemoveDangling(p, net); e != nil {
		t.Fatal(e)
	}
	if !net.deleted[2] {
		t.Fatal("dangling bypass buffer retained")
	}
}
func TestUpsizeBypassFailedPatchRestoresLoad(t *testing.T) {
	p, net := mutableUpsizeFixture(t)
	defer Abc_SclManFree(p)
	path, critical := Abc_SclFindCriticalNodeWindow(p, Abc_SclFindCriticalCoWindow(p, 1), 1, true)
	before := append([]SC_Pair(nil), p.Loads...)
	net.failPatch = true
	if _, e := Abc_SclFindBypasses(p, path, &UpsizeState{Critical: critical}, 100, 4, 1, 0, false); e == nil {
		t.Fatal("ignored patch error")
	}
	for i, v := range before {
		if p.Loads[i] != v {
			t.Fatal("trial leaked loads", i, v, p.Loads[i])
		}
	}
	if net.fanins[3][0] != 2 {
		t.Fatal(net.fanins)
	}
}

func TestUpsizeSingleStageWithoutDeparture(t *testing.T) {
	p, net := mutableUpsizeFixture(t)
	Abc_SclManFree(p)
	net.deleted[2], net.deleted[3] = true, true
	delete(net.fanins, 2)
	delete(net.fanins, 3)
	net.fanins[4] = []int{1}
	p, e := Abc_SclManStart(net.mapping.Lib, net, 0, 0, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer Abc_SclManFree(p)
	p.Output = io.Discard
	if e = UpsizeManager(p, &SC_SizePars{Iterations: 2, Window: 1, Ratio: 100, Notches: 4}); e != nil {
		t.Fatal(e)
	}
	if p.MaxDelay != 1 || net.mapping.GateIDs[1] != 1 {
		t.Fatal("single-stage improvement lost", p.MaxDelay, net.mapping.GateIDs)
	}
}

type phasedUpsizeNetwork struct {
	*mutableUpsizeNetwork
	phase map[[2]int]bool
}

func (n *phasedUpsizeNetwork) SclHasPhases() bool { return true }
func (n *phasedUpsizeNetwork) SclToggleFaninPhase(node, pin int) error {
	key := [2]int{node, pin}
	n.phase[key] = !n.phase[key]
	return nil
}
func TestUpsizeInverterBypassRestoresPhase(t *testing.T) {
	old, base := mutableUpsizeFixture(t)
	Abc_SclManFree(old)
	lib := base.mapping.Lib
	inv := *lib.Cells[0]
	inv.Name = "INV"
	inv.ID = 2
	inv.Pins = append([]*SC_Pin(nil), inv.Pins...)
	pin := *inv.Pins[1]
	pin.Function = "!A"
	pin.Truth = []uint64{0x5555555555555555}
	inv.Pins[1] = &pin
	lib.Cells = append(lib.Cells, &inv)
	Abc_SclHashCells(lib)
	Abc_SclLinkCells(lib)
	base.mapping.GateIDs[2] = 2
	net := &phasedUpsizeNetwork{mutableUpsizeNetwork: base, phase: map[[2]int]bool{}}
	p, e := Abc_SclManStart(lib, net, 0, 1, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer Abc_SclManFree(p)
	p.Output = io.Discard
	path, critical := Abc_SclFindCriticalNodeWindow(p, Abc_SclFindCriticalCoWindow(p, 1), 1, true)
	state := &UpsizeState{Critical: critical}
	count, e := Abc_SclFindBypasses(p, path, state, 100, 4, 1, 0, false)
	if e != nil {
		t.Fatal(e)
	}
	if count != 1 || !net.phase[[2]int{3, 0}] {
		t.Fatal("inverter bypass lost polarity", count, net.phase)
	}
	if e = Abc_SclUndoRecentChanges(p, state.Edits); e != nil {
		t.Fatal(e)
	}
	if net.phase[[2]int{3, 0}] || net.fanins[3][0] != 2 {
		t.Fatal("inverter undo lost polarity", net.phase, net.fanins)
	}
}

func TestUpsizeCommandUsesRealMapping(t *testing.T) {
	lib, net := upsizeFixture(t)
	f := abcmain.NewFrame()
	defer f.Close()
	f.Out = io.Discard
	f.Err = io.Discard
	f.SetCurrentNetwork(net)
	f.SetResource("pLibScl", lib)
	Scl_Init(f)
	command, ok := f.LookupCommand("upsize")
	if !ok || !command.ChangesNetwork {
		t.Fatal(command, ok)
	}
	args := []string{"upsize", "-I", "10", "-W", "1", "-R", "100", "-N", "4"}
	if status := command.Handler(f, args); status != 0 {
		t.Fatal(status)
	}
	if net.mapping.GateIDs[1] != 1 || net.mapping.GateIDs[2] != 1 {
		t.Fatal(net.mapping.GateIDs)
	}
	if status := command.Handler(f, []string{"upsize", "-W", "0"}); status != 1 {
		t.Fatal("accepted invalid window", status)
	}
}
