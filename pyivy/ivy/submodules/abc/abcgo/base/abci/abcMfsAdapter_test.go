package abci

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestMfsParamsDefaultAndTranslation(t *testing.T) {
	var p Mfs_Par_t
	Abc_NtkMfsParsDefault(&p)
	if p.nWinTfoLevs != 2 || p.nFanoutsMax != 30 || p.nDepthMax != 20 || p.nWinMax != 300 || p.nGrowthLevel != 0 || p.nBTLimit != 5000 || p.fResub != 1 {
		t.Fatalf("unexpected ABC defaults: %+v", p)
	}
	p.nGrowthLevel, p.nBTLimit = 4, 123
	p.fRrOnly, p.fArea, p.fMoreEffort, p.fSwapEdge = 1, 1, 1, 1
	p.fOneHotness, p.fDelay, p.fPower, p.fGiaSat, p.fVerbose, p.fVeryVerbose = 1, 1, 1, 1, 1, 1
	got := mfsParamsFromABC(&p)
	if got.WinTfoLevs != 2 || got.FanoutsMax != 30 || got.DepthMax != 20 || got.WinMax != 300 || got.GrowthLevel != 4 || got.BTLimit != 123 {
		t.Fatalf("integer parameter translation: %+v", got)
	}
	if !got.RrOnly || !got.Resub || !got.Area || !got.MoreEffort || !got.SwapEdge || !got.OneHotness || !got.Delay || !got.Power || !got.GiaSat || !got.Verbose || !got.VeryVerbose {
		t.Fatalf("flag parameter translation: %+v", got)
	}
}

func TestMfsABCAdapterPreservesNetworkAndCoherence(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	first := abc.Abc_NtkCreateNode(n)
	for _, input := range []*abc.Abc_Obj_t{a, b, c} {
		abc.Abc_ObjAddFanin(first, input)
	}
	first.Data = abc.Abc_SopCreateAnd(nil, 3, []int32{0, 0, 0})
	second := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(second, first)
	abc.Abc_ObjAddFanin(second, a)
	second.Data = abc.Abc_SopCreateOr(nil, 2, []int32{0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, second)

	// A valid ABC logic network has names for all interface terminals.
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)
	before, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import before optimization: %v", err)
	}
	var pars Mfs_Par_t
	Abc_NtkMfsParsDefault(&pars)
	if !Abc_NtkMfs(n, &pars) {
		t.Fatal("Abc_NtkMfs returned failure")
	}
	if abc.Abc_NtkObj(n, int32(first.ID)) != first || abc.Abc_NtkObj(n, int32(second.ID)) != second || abc.Abc_NtkObj(n, int32(po.ID)) != po {
		t.Fatal("adapter replaced ABC objects instead of updating the current network")
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("ABC network reciprocal fanin/fanout check failed after writeback")
	}
	after, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import after optimization: %v", err)
	}
	for assignment := 0; assignment < 1<<3; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, err := before.network.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := after.network.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("assignment %03b changed outputs: got %v want %v", assignment, got, want)
		}
	}
}

func TestMfsABCImportNormalizesComplementedFanin(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	node := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b.Signal().Not())
	node.Data = abc.Abc_SopCreateAnd(nil, 2, []int32{0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, node)

	imported, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import complemented fanin: %v", err)
	}
	if len(imported.network.Nodes) != 4 || imported.network.Nodes[3].Fanins[1] == 0 {
		t.Fatalf("unexpected regular fanin mapping: %+v", imported.network.Nodes[3])
	}
	truth := imported.network.Nodes[3].Truth[0]
	if truth != 0x22222222 {
		t.Fatalf("AND(a, !b) regular-coordinate truth = %#08x, want %#08x", truth, uint32(0x22222222))
	}
}

func TestMfsABCImportRetainsExcareClauses(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	inputs := make([]*abc.Abc_Obj_t, 3)
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(n)
	}
	node := abc.Abc_NtkCreateNode(n)
	for _, input := range inputs {
		abc.Abc_ObjAddFanin(node, input)
	}
	node.Data = abc.Abc_SopCreateAnd(nil, 3, []int32{0, 0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, node)

	care := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	careInputs := make([]*abc.Abc_Obj_t, 3)
	for i := range careInputs {
		careInputs[i] = abc.Abc_NtkCreatePi(care)
	}
	careNode := abc.Abc_NtkCreateNode(care)
	for _, input := range careInputs {
		abc.Abc_ObjAddFanin(careNode, input)
	}
	careNode.Data = abc.Abc_SopCreateOr(nil, 3, []int32{0, 0, 0})
	carePO := abc.Abc_NtkCreatePo(care)
	abc.Abc_ObjAddFanin(carePO, careNode)
	n.Excare = care

	imported, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import external care clauses: %v", err)
	}
	if imported.network.Care == nil || imported.network.Care.InputCount() != 3 || imported.network.Care.OutputCount() != 1 {
		t.Fatalf("external care graph was not retained: %+v", imported.network.Care)
	}
}

func TestMfsABCImportIgnoresExcareWithMismatchedPiCount(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	input := abc.Abc_NtkCreatePi(n)
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, input)

	care := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	careInput0, careInput1 := abc.Abc_NtkCreatePi(care), abc.Abc_NtkCreatePi(care)
	careNode := abc.Abc_NtkCreateNode(care)
	abc.Abc_ObjAddFanin(careNode, careInput0)
	abc.Abc_ObjAddFanin(careNode, careInput1)
	careNode.Data = abc.Abc_SopCreateAnd(nil, 2, []int32{0, 0})
	carePO := abc.Abc_NtkCreatePo(care)
	abc.Abc_ObjAddFanin(carePO, careNode)
	n.Excare = care

	imported, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("mismatched care PI count should be ignored like C: %v", err)
	}
	if imported.network.Care != nil {
		t.Fatal("mismatched external care network was used")
	}
}

func TestMfsABCImportRetainsRegisterTerminalOrder(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	pi := abc.Abc_NtkCreatePi(n)
	bo := abc.Abc_NtkAddLatch(n, nil, abc.ABC_INIT_ZERO)
	bi := n.ObjectsOfKind(abc.ObjectBI)[0]
	logic := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(logic, pi)
	abc.Abc_ObjAddFanin(logic, bo)
	logic.Data = abc.Abc_SopCreateOr(nil, 2, []int32{0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, logic)
	abc.Abc_ObjAddFanin(bi, logic)
	abc.Abc_NtkOrderCisCos(n)

	imported, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import registered network: %v", err)
	}
	if imported.network.Registers != 1 || len(imported.network.Inputs) != 2 || len(imported.network.Outputs) != 2 {
		t.Fatalf("register partition mismatch: registers=%d CIs=%d COs=%d", imported.network.Registers, len(imported.network.Inputs), len(imported.network.Outputs))
	}
	if imported.objects[imported.network.Inputs[1]] != bo {
		t.Fatal("register output is not the trailing MFS CI")
	}
	if imported.network.Outputs[1].ID() == 0 || imported.objects[imported.network.Outputs[1].ID()] != logic {
		t.Fatal("register input is not the trailing MFS CO driver")
	}
	if abc.Abc_NtkObj(n, int32(pi.ID)) != pi || abc.Abc_NtkObj(n, int32(bo.ID)) != bo || abc.Abc_NtkObj(n, int32(bi.ID)) != bi || abc.Abc_NtkObj(n, int32(po.ID)) != po {
		t.Fatal("register import changed the existing ABC interface objects")
	}
}

func TestMfsABCAdapterWritebackPreservesComplementedFanin(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	node := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b.Signal().Not())
	node.Data = abc.Abc_SopCreateAnd(nil, 2, []int32{0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, node.Signal().Not())

	before, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import before optimization: %v", err)
	}
	var pars Mfs_Par_t
	Abc_NtkMfsParsDefault(&pars)
	if !Abc_NtkMfs(n, &pars) {
		t.Fatal("Abc_NtkMfs returned failure")
	}
	if abc.Abc_ObjFaninC1(node) != 0 {
		t.Fatal("writeback left the original complemented fanin outside the local function")
	}
	if abc.Abc_ObjFanin0(po) != node || abc.Abc_ObjFaninC0(po) == 0 {
		t.Fatal("writeback changed the complemented PO edge or its driver identity")
	}
	after, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import after optimization: %v", err)
	}
	for assignment := 0; assignment < 1<<2; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		want, err := before.network.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := after.network.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("assignment %02b changed outputs: got %v want %v", assignment, got, want)
		}
	}
}

func TestMfsABCAdapterWritebackPreservesExcareAssignments(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	inputs := make([]*abc.Abc_Obj_t, 3)
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(n)
	}
	node := abc.Abc_NtkCreateNode(n)
	for _, input := range inputs {
		abc.Abc_ObjAddFanin(node, input)
	}
	node.Data = abc.Abc_SopCreateAnd(nil, 3, []int32{0, 0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, node)

	care := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	careInputs := make([]*abc.Abc_Obj_t, 3)
	for i := range careInputs {
		careInputs[i] = abc.Abc_NtkCreatePi(care)
	}
	careNode := abc.Abc_NtkCreateNode(care)
	for _, input := range careInputs {
		abc.Abc_ObjAddFanin(careNode, input)
	}
	careNode.Data = abc.Abc_SopCreateOr(nil, 3, []int32{0, 0, 0})
	carePO := abc.Abc_NtkCreatePo(care)
	abc.Abc_ObjAddFanin(carePO, careNode)
	n.Excare = care

	before, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import before optimization: %v", err)
	}
	var pars Mfs_Par_t
	Abc_NtkMfsParsDefault(&pars)
	if !Abc_NtkMfs(n, &pars) {
		t.Fatal("Abc_NtkMfs returned failure")
	}
	after, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import after optimization: %v", err)
	}
	if after.network.Care == nil || after.network.Care.InputCount() != len(inputs) {
		t.Fatal("writeback lost the external care network")
	}
	for assignment := 0; assignment < 1<<3; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		careValues, err := before.network.Care.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if !careValues[0] {
			continue
		}
		want, err := before.network.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		got, err := after.network.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("care assignment %03b changed outputs: got %v want %v", assignment, got, want)
		}
	}
}

func TestMfsABCAdapterWritebackPreservesRegisterTerminalOrder(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	pi := abc.Abc_NtkCreatePi(n)
	bo := abc.Abc_NtkAddLatch(n, nil, abc.ABC_INIT_ZERO)
	bi := n.ObjectsOfKind(abc.ObjectBI)[0]
	logic := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(logic, pi)
	abc.Abc_ObjAddFanin(logic, bo)
	logic.Data = abc.Abc_SopCreateOr(nil, 2, []int32{0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, logic)
	abc.Abc_ObjAddFanin(bi, logic)
	abc.Abc_NtkOrderCisCos(n)

	before, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import before optimization: %v", err)
	}
	var pars Mfs_Par_t
	Abc_NtkMfsParsDefault(&pars)
	if !Abc_NtkMfs(n, &pars) {
		t.Fatal("Abc_NtkMfs returned failure")
	}
	after, err := mfsImportABC(n)
	if err != nil {
		t.Fatalf("import after optimization: %v", err)
	}
	if after.network.Registers != 1 || len(after.network.Inputs) != 2 || len(after.network.Outputs) != 2 {
		t.Fatalf("register partition changed: registers=%d CIs=%d COs=%d", after.network.Registers, len(after.network.Inputs), len(after.network.Outputs))
	}
	if after.objects[after.network.Inputs[1]] != bo || after.network.Outputs[1].ID() == 0 || after.objects[after.network.Outputs[1].ID()] != logic {
		t.Fatal("writeback changed the trailing register terminal mapping")
	}
	for assignment := 0; assignment < 1<<2; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0}
		want, err := before.network.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		got, err := after.network.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("CI assignment %02b changed CO values: got %v want %v", assignment, got, want)
		}
	}
}
