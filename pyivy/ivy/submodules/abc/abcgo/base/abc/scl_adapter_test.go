package abc

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"reflect"
	"testing"
)

func abcSCLFixture(t *testing.T) (*Network, *scl.SC_Lib, *mio.Mio_Library_t, *Object, *Object, *Object) {
	t.Helper()
	gates, err := mio.ReadLibrary("mapped", "GATE buf1 1 Y=A;\nPIN A NONINV 1 99 1 0 1 0\nGATE buf2 2 Y=A;\nPIN A NONINV 2 99 1 0 1 0\nGATE and2 2 Y=A*B;\nPIN * NONINV 1 99 1 0 1 0\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	library := &scl.SC_Lib{Name: "mapped", UnitTime: 12, UnitCapMultiplier: 1, UnitCapExponent: 15}
	surface := func(value float32) scl.SC_Surface {
		return scl.SC_Surface{Index0: []float32{0, 10}, Index1: []float32{0, 10}, Data: [][]float32{{value, value}, {value, value}}}
	}
	for i, name := range []string{"buf1", "buf2"} {
		timing := &scl.SC_Timing{RelatedPin: "A", Sense: scl.ScTSPos, CellRise: surface(1), CellFall: surface(1), RiseTrans: surface(.1), FallTrans: surface(.1)}
		input := &scl.SC_Pin{Name: "A", Direction: scl.ScDirInput, RiseCap: float32(i + 1), FallCap: float32(i + 1)}
		output := &scl.SC_Pin{Name: "Y", Direction: scl.ScDirOutput, Function: "A", Truth: []uint64{0xaaaaaaaaaaaaaaaa}, Timings: []*scl.SC_Timings{{Name: "A", Timings: []*scl.SC_Timing{timing}}}}
		library.Cells = append(library.Cells, &scl.SC_Cell{Name: name, ID: i, Inputs: 1, Outputs: 1, Area: float32(i + 1), Pins: []*scl.SC_Pin{input, output}})
	}
	scl.Abc_SclHashCells(library)
	scl.Abc_SclLinkCells(library)
	n := NewNetwork("real mapped")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_MAP)
	pi := n.NewObject(ObjectPI)
	consumer, driver := n.NewObject(ObjectNode), n.NewObject(ObjectNode)
	gate := mio.Mio_LibraryReadGateByName(gates, "buf1", "")
	consumer.Data, driver.Data = gate, gate
	Abc_ObjAddFanin(driver, pi)
	Abc_ObjAddFanin(consumer, driver)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, consumer)
	if err := n.ConfigureSCL(library, gates); err != nil {
		t.Fatal(err)
	}
	return n, library, gates, pi, driver, consumer
}
func TestSCLRealNetworkTimingAndAtomicGateCommit(t *testing.T) {
	n, library, gates, _, driver, consumer := abcSCLFixture(t)
	manager, err := scl.NewTimingManager(library, n)
	if err != nil {
		t.Fatal(err)
	}
	if err := scl.Abc_SclTimeNtkRecompute(manager, &manager.InitialArea, &manager.InitialDelay, 1, 0); err != nil {
		t.Fatal(err)
	}
	mapping := n.SclMapping()
	mapping.GateIDs[driver.ID] = 1
	if err := scl.Abc_SclSclGates2MioGates(library, n); err != nil {
		t.Fatal(err)
	}
	if driver.Data != mio.Mio_LibraryReadGateByName(gates, "buf2", "") {
		t.Fatal("staged gate never reached actual object")
	}
	before := consumer.Data
	mapping.MioGates[consumer.ID] = mio.Mio_LibraryReadGateByName(gates, "and2", "")
	if err := n.SclCommitMapping(); err == nil {
		t.Fatal("invalid gate arity accepted")
	}
	if consumer.Data != before {
		t.Fatal("failed commit changed real object")
	}
}
func TestSCLTopologicalCloneRetainsMappingAndNames(t *testing.T) {
	n, library, _, pi, driver, consumer := abcSCLFixture(t)
	driver.SetName("driver")
	consumer.SetName("consumer")
	consumer.Copy = pi
	if err := scl.Abc_SclMioGates2SclGates(library, n); err != nil {
		t.Fatal(err)
	}
	cloned, err := n.SclCloneTopological()
	if err != nil {
		t.Fatal(err)
	}
	clone := cloned.(*Network)
	if consumer.Copy != pi {
		t.Fatal("clone overwrote caller scratch")
	}
	for _, o := range clone.ObjectsOfKind(ObjectNode) {
		for _, edge := range o.fanins {
			if edge.Object.ID >= o.ID {
				t.Fatal("clone retains forward node IDs", o.ID, edge.Object.ID)
			}
		}
	}
	if clone.SclMapping().MioGates[clone.ObjectsOfKind(ObjectNode)[0].ID] != driver.Data {
		t.Fatal("mapping was not remapped")
	}
	if clone.ObjectsOfKind(ObjectNode)[0].Name() != "driver" {
		t.Fatal("name lost")
	}
	if err := clone.Check(); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(clone.Objects(), n.Objects()) {
		t.Fatal("clone shares graph objects")
	}
}
func TestSCLMutationsUseRealAdjacencyAndPhases(t *testing.T) {
	n, library, _, pi, driver, consumer := abcSCLFixture(t)
	if err := scl.Abc_SclMioGates2SclGates(library, n); err != nil {
		t.Fatal(err)
	}
	if err := n.SclSetPhases(true); err != nil {
		t.Fatal(err)
	}
	if err := n.SclToggleFaninPhase(consumer.ID, 0); err != nil {
		t.Fatal(err)
	}
	if err := n.SclPatchFanin(consumer.ID, 0, pi.ID); err != nil {
		t.Fatal(err)
	}
	if !consumer.fanins[0].Complement || consumer.fanins[0].Object != pi {
		t.Fatal("phase or topology lost")
	}
	if err := n.SclSetPhases(false); err == nil {
		t.Fatal("discarded unmaterialized complement")
	}
	if err := n.SclToggleFaninPhase(consumer.ID, 0); err != nil {
		t.Fatal(err)
	}
	if err := n.SclSetPhases(false); err != nil {
		t.Fatal(err)
	}
	if err := n.SclDeleteNode(driver.ID); err != nil {
		t.Fatal(err)
	}
	if n.Object(driver.ID) != nil {
		t.Fatal("deleted driver retained")
	}
	id, err := n.SclCreateNode([]int{pi.ID}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if n.Object(id).Data == nil || n.SclMapping().GateIDs[id] != 1 {
		t.Fatal("new mapping absent")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestSCLDetectsComplementAddedThroughGraphAPI(t *testing.T) {
	n, _, _, _, driver, consumer := abcSCLFixture(t)
	if n.SclHasPhases() {
		t.Fatal("fresh graph has phases")
	}
	if err := n.PatchFanin(consumer, 0, driver.Signal().Not()); err != nil {
		t.Fatal(err)
	}
	if !n.SclHasPhases() || !n.SclFaninPhase(consumer.ID, 0) {
		t.Fatal("generic graph phase mutation was invisible")
	}
	if err := n.SclToggleFaninPhase(consumer.ID, 0); err != nil {
		t.Fatal(err)
	}
}
