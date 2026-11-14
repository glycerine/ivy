package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
	baseabc "github.com/glycerine/abcgo/base/abc"
	acbcore "github.com/glycerine/abcgo/base/acb"
	"github.com/glycerine/abcgo/misc/mem"
	mfscore "github.com/glycerine/abcgo/opt/mfs"
	sfmcore "github.com/glycerine/abcgo/opt/sfm"
)

// Mfs_Par_t keeps the C command's field names at the ABC command boundary.
// The optimizer itself uses the typed Go parameters in opt/mfs.
type Mfs_Par_t struct {
	nWinTfoLevs, nFanoutsMax, nDepthMax, nWinMax, nGrowthLevel int32
	nBTLimit                                                   int32
	fRrOnly, fResub, fArea, fMoreEffort, fSwapEdge             int32
	fOneHotness, fDelay, fPower, fGiaSat, fVerbose             int32
	fVeryVerbose                                               int32
}

func Sfm_ParSetDefault(p *Sfm_Par_t)  { sfmcore.Sfm_ParSetDefault(p) }
func Sfm_ParSetDefault3(p *Sfm_Par_t) { sfmcore.Sfm_ParSetDefault3(p) }

// Acb_Par_t preserves the legacy command fields; base/acb owns the private
// optimizer representation and receives these values through its setter.
type Acb_Par_t struct {
	nLutSize, nTfoLevMax, nTfiLevMax, nFanoutMax, nWinNodeMax int32
	nGrowthLevel, nBTLimit, nNodesMax, iNodeOne               int32
	fArea, fUseAshen, fMoreEffort, fVerbose, fVeryVerbose     int32
}

func Acb_ParSetDefault(p *Acb_Par_t) {
	*p = Acb_Par_t{nLutSize: 4, nTfoLevMax: 2, nTfiLevMax: 3, nFanoutMax: 20, nWinNodeMax: 100, fArea: 1}
}

func Abc_NtkOptMfse(n *Abc_Ntk_t, p *Acb_Par_t) *Abc_Ntk_t {
	if n == nil || p == nil {
		return nil
	}
	var native acbcore.Acb_Par_t
	acbcore.Acb_ParSetDefault(&native)
	acbcore.Acb_ParSetCommand(&native, p.nLutSize, p.nTfoLevMax, p.nTfiLevMax, p.nFanoutMax,
		p.nWinNodeMax, p.nGrowthLevel, p.nBTLimit, p.nNodesMax, p.iNodeOne,
		p.fArea, p.fUseAshen, p.fMoreEffort, p.fVerbose, p.fVeryVerbose)
	return acbcore.Abc_NtkOptMfse(n, &native)
}

func Abc_NtkMfsParsDefault(p *Mfs_Par_t) {
	*p = Mfs_Par_t{nWinTfoLevs: 2, nFanoutsMax: 30, nDepthMax: 20, nWinMax: 300, nBTLimit: 5000, fResub: 1}
}

func mfsParamsFromABC(p *Mfs_Par_t) mfscore.Mfs_Par_t {
	return mfscore.Mfs_Par_t{
		WinTfoLevs: int(p.nWinTfoLevs), FanoutsMax: int(p.nFanoutsMax), DepthMax: int(p.nDepthMax),
		WinMax: int(p.nWinMax), GrowthLevel: int(p.nGrowthLevel), BTLimit: int64(p.nBTLimit),
		RrOnly: p.fRrOnly != 0, Resub: p.fResub != 0, Area: p.fArea != 0,
		MoreEffort: p.fMoreEffort != 0, SwapEdge: p.fSwapEdge != 0, OneHotness: p.fOneHotness != 0,
		Delay: p.fDelay != 0, Power: p.fPower != 0, GiaSat: p.fGiaSat != 0,
		Verbose: p.fVerbose != 0, VeryVerbose: p.fVeryVerbose != 0,
	}
}

type mfsABCMap struct {
	network  *mfscore.Network
	objects  []*baseabc.Object // index is MFS node ID; zero is ABC's constant one
	nodeIDs  []int
	original []mfsLocal // indexed by MFS node ID
}

type mfsLocal struct {
	fanins []int
	truth  []uint32
}

// mfsTruth imports a node's Hop function in its ordered local-variable basis.
// ABC_NtkToAig has already incorporated SOP output phase; complemented root
// edges are included by Hop_ManConvertAigToTruth.
func mfsTruth(manager *hop.Hop_Man_t, object *baseabc.Object) ([]uint32, error) {
	root, ok := object.Data.(*hop.Hop_Obj_t)
	if !ok || root == nil {
		return nil, fmt.Errorf("MFS: node %d has no Hop local function", object.ID)
	}
	return hop.Hop_ManConvertAigToTruth(manager, root, int32(len(object.Fanins())), nil, 0), nil
}

func mfsImportABC(pNtk *Abc_Ntk_t) (*mfsABCMap, error) {
	if baseabc.Abc_NtkToAig(pNtk) == 0 {
		return nil, fmt.Errorf("MFS: converting network functions to AIGs failed")
	}
	manager, ok := pNtk.FunctionManager.(*hop.Hop_Man_t)
	if !ok || manager == nil {
		return nil, fmt.Errorf("MFS: AIG network has no Hop manager")
	}
	if baseabc.Abc_NtkLatchNum(pNtk) != 0 && baseabc.Abc_NtkLatchNum(pNtk) > baseabc.Abc_NtkCiNum(pNtk) {
		return nil, fmt.Errorf("MFS: invalid register partition")
	}

	n := mfscore.NewNetwork()
	mapping := &mfsABCMap{network: n, objects: []*baseabc.Object{pNtk.Object(0)}, original: []mfsLocal{{}}}
	ids := map[*baseabc.Object]int{pNtk.Object(0): 0}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		object := baseabc.Abc_NtkCi(pNtk, i)
		if object == nil {
			return nil, fmt.Errorf("MFS: missing CI %d", i)
		}
		id := n.AddInput()
		mapping.objects = append(mapping.objects, object)
		mapping.original = append(mapping.original, mfsLocal{})
		ids[object] = id
	}
	order, err := pNtk.Topological(pNtk.Objects())
	if err != nil {
		return nil, fmt.Errorf("MFS: network is not topological: %w", err)
	}
	for _, object := range order {
		if object.Kind != baseabc.ObjectNode {
			continue
		}
		if len(object.Fanins()) > mfscore.MFS_FANIN_MAX {
			return nil, fmt.Errorf("MFS: node %d has more than %d fanins", object.ID, mfscore.MFS_FANIN_MAX)
		}
		truth, err := mfsTruth(manager, object)
		if err != nil {
			return nil, err
		}
		fanins := make([]int, len(object.Fanins()))
		for i, edge := range object.Fanins() {
			id, found := ids[edge.Object]
			if !found {
				return nil, fmt.Errorf("MFS: node %d has unmapped fanin %d", object.ID, edge.Object.ID)
			}
			fanins[i] = id
			// Current ABC logic-local functions are in regular fanin coordinates.
			// A complemented network edge is represented in the local truth basis.
			if edge.Complement {
				truth = mfsNegateInput(truth, len(fanins), i)
			}
		}
		id := n.AddNode(fanins, truth)
		mapping.objects = append(mapping.objects, object)
		ids[object] = id
		mapping.nodeIDs = append(mapping.nodeIDs, id)
		mapping.original = append(mapping.original, mfsLocal{fanins: append([]int(nil), fanins...), truth: append([]uint32(nil), truth...)})
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		co := baseabc.Abc_NtkCo(pNtk, i)
		if co == nil || len(co.Fanins()) != 1 {
			return nil, fmt.Errorf("MFS: CO %d has invalid fanin count", i)
		}
		edge := co.Fanins()[0]
		id, found := ids[edge.Object]
		if !found {
			return nil, fmt.Errorf("MFS: CO %d has an unmapped driver", i)
		}
		literal := mfscore.Literal(id << 1)
		if edge.Complement {
			literal = literal.Not()
		}
		n.AddOutput(literal)
	}
	n.Registers = int(baseabc.Abc_NtkLatchNum(pNtk))
	if pNtk.Excare != nil {
		carePIs := baseabc.Abc_NtkPiNum(pNtk.Excare)
		if carePIs != baseabc.Abc_NtkCiNum(pNtk) {
			fmt.Printf("The PI count of careset (%d) and logic network (%d) differ. Careset is not used.\n", carePIs, baseabc.Abc_NtkCiNum(pNtk))
		} else {
			careNtk := Abc_NtkStrash(pNtk.Excare, 0, 0, 0)
			if careNtk == nil {
				return nil, fmt.Errorf("MFS: converting external care network to AIG failed")
			}
			care := baseabc.Abc_NtkToDar(careNtk, 0, 0)
			baseabc.Abc_NtkDelete(careNtk)
			if care == nil {
				return nil, fmt.Errorf("MFS: converting external care network to AIG failed")
			}
			if care.InputCount() != len(n.Inputs) {
				return nil, fmt.Errorf("MFS: external care CI count %d differs from network CI count %d", care.InputCount(), len(n.Inputs))
			}
			n.Care = care
		}
	}
	if err := n.Check(); err != nil {
		return nil, fmt.Errorf("MFS: invalid imported network: %w", err)
	}
	return mapping, nil
}

// mfsNegateInput reorders the local truth table so variable i denotes the
// regular ABC object rather than a complemented fanin edge.
func mfsNegateInput(truth []uint32, vars, variable int) []uint32 {
	out := make([]uint32, len(truth))
	for assignment := 0; assignment < 1<<uint(vars); assignment++ {
		from := assignment ^ (1 << uint(variable))
		if truth[from>>5]&(1<<uint(from&31)) != 0 {
			out[assignment>>5] |= 1 << uint(assignment&31)
		}
	}
	return out
}

func mfsWriteABC(pNtk *Abc_Ntk_t, imported *mfsABCMap) error {
	if len(imported.objects) != len(imported.network.Nodes) || len(imported.original) != len(imported.network.Nodes) {
		return fmt.Errorf("MFS: object map is inconsistent")
	}
	manager := hop.Hop_ManStart()
	// Keep old local functions for dead, now-dangling nodes; source ABC's update
	// leaves those objects allocated until the usual network sweep.
	for _, id := range imported.nodeIDs {
		object := imported.objects[id]
		local := imported.original[id]
		node := imported.network.Nodes[id]
		if !node.Deleted {
			local = mfsLocal{fanins: node.Fanins, truth: node.Truth}
		}
		if object == nil || object.Kind != baseabc.ObjectNode {
			return fmt.Errorf("MFS: object map node %d is invalid", id)
		}
		newFanins := make([]baseabc.Edge, len(local.fanins))
		for j, id := range local.fanins {
			if id < 0 || id >= len(imported.objects) || imported.objects[id] == nil {
				return fmt.Errorf("MFS: node %d has invalid replacement fanin %d", object.ID, id)
			}
			newFanins[j] = imported.objects[id].Signal()
		}
		text := baseabc.Abc_SopCreateFromTruth((*mem.Mem_Flex_t)(nil), int32(len(newFanins)), local.truth)
		root := baseabc.Abc_ConvertSopToAig(manager, text)
		if root == nil {
			return fmt.Errorf("MFS: rebuilding node %d failed", object.ID)
		}
		baseabc.Abc_ObjRemoveFanins(object)
		for _, edge := range newFanins {
			baseabc.Abc_ObjAddFanin(object, edge)
		}
		object.Data = root
	}
	pNtk.FunctionManager = manager
	pNtk.Function = uint8(baseabc.ABC_FUNC_AIG)
	return nil
}

// Abc_NtkMfs is the source-command adapter: the optimizer runs on an explicit
// MFS network, and its result is committed back into the same ABC network.
func Abc_NtkMfs(pNtk *Abc_Ntk_t, pPars *Mfs_Par_t) bool {
	if pNtk == nil || pPars == nil {
		return false
	}
	imported, err := mfsImportABC(pNtk)
	if err != nil {
		fmt.Println(err)
		return false
	}
	params := mfsParamsFromABC(pPars)
	if !mfscore.Abc_NtkMfs(imported.network, &params) {
		return false
	}
	if err := mfsWriteABC(pNtk, imported); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
