package abci

import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/cut"
)

var nTotal, nGood, nEqual int32

type Cut_Oracle_t = cut.Cut_Oracle_t

type abcRewriteCutNetwork struct{ network *abc.Network }

func (n abcRewriteCutNetwork) CutNodes() []cut.NetworkNode {
	objects := n.network.Objects()
	nodes := make([]cut.NetworkNode, 0, len(objects))
	for _, object := range objects {
		nodes = append(nodes, abcCutNetworkNode(object))
	}
	return nodes
}

func abcCutNetworkNode(object *abc.Object) cut.NetworkNode {
	fanins := object.Fanins()
	node := cut.NetworkNode{
		ID:           int32(object.ID),
		FanoutAdjust: abc.Abc_ObjFanoutNum(object) - int32(len(object.Fanouts())),
		Input:        object.IsCI(),
		Output:       object.IsCO(),
		Constant:     object.Kind == abc.ObjectConst1,
		Value:        object.Kind == abc.ObjectConst1,
		XOR:          abc.Abc_NodeIsExorType(object) != 0,
		MuxControl:   abc.Abc_NodeIsMuxControlType(object) != 0,
		Buffer:       !object.IsCO() && len(fanins) == 1,
	}
	seen := make(map[*abc.Object]bool)
	for choice := object.Choice; choice != nil && !seen[choice]; choice = choice.Choice {
		seen[choice] = true
		node.Choices = append(node.Choices, int32(choice.ID))
	}
	for i, edge := range fanins {
		if i >= len(node.Fanins) {
			break
		}
		node.Fanins[i] = int32(edge.Object.ID)
		node.Compl[i] = edge.Complement
	}
	return node
}

func (n abcRewriteCutNetwork) CutNodeAttributes() []int32 {
	return vec.Vec_IntArray(Abc_NtkGetNodeAttributes(n.network))
}

// Mux fanout adjustments are temporary reference counts in ABC. The graph's
// actual adjacency remains intact while the cut manager sees adjusted counts.
func Abc_NtkCutsSubtractFanunt(pNtk *Abc_Ntk_t) {
	var counter int32
	for _, object := range pNtk.Objects() {
		if abc.Abc_NodeIsMuxType(object) == 0 {
			continue
		}
		var thenEdge, elseEdge abc.Edge
		control := abc.Abc_NodeRecognizeMux(object, &thenEdge, &elseEdge)
		assert(abc.Abc_ObjFanoutNum(control) > 1, "control fanout count > 1")
		abc.Abc_ObjAdjustFanoutCount(control, -1)
		counter++
		if abc.Abc_NodeIsExorType(object) != 0 {
			assert(abc.Abc_ObjFanoutNum(elseEdge.Object) > 1, "data fanout count > 1")
			abc.Abc_ObjAdjustFanoutCount(elseEdge.Object, -1)
			counter++
		}
	}
	fmt.Printf("Substracted %d fanouts\n", counter)
}

func Abc_NtkCutsAddFanunt(pNtk *Abc_Ntk_t) {
	var counter int32
	for _, object := range pNtk.Objects() {
		if abc.Abc_NodeIsMuxType(object) == 0 {
			continue
		}
		var thenEdge, elseEdge abc.Edge
		control := abc.Abc_NodeRecognizeMux(object, &thenEdge, &elseEdge)
		abc.Abc_ObjAdjustFanoutCount(control, 1)
		counter++
		if abc.Abc_NodeIsExorType(object) != 0 {
			abc.Abc_ObjAdjustFanoutCount(elseEdge.Object, 1)
			counter++
		}
	}
	abc.Abc_NtkFinishFanoutCountAdjustments(pNtk)
	fmt.Printf("Added %d fanouts\n", counter)
}

func Abc_NtkCuts(pNtk *Abc_Ntk_t, pParams *Cut_Params_t) *Cut_Man_t {
	assert(pNtk != nil && abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkCuts requires a strashed network")
	assert(pParams != nil, "Abc_NtkCuts requires cut parameters")
	pParams.NIdsMax = abc.Abc_NtkObjNumMax(pNtk)
	nTotal, nGood, nEqual = 0, 0, 0
	started := time.Now()
	if pParams.FAdjust != 0 {
		Abc_NtkCutsSubtractFanunt(pNtk)
		defer Abc_NtkCutsAddFanunt(pNtk)
	}
	network := abcRewriteCutNetwork{network: pNtk}
	p := cut.Cut_ManStartForNetwork(network, pParams)
	if pParams.FGlobal != 0 || pParams.FLocal != 0 {
		cut.Cut_ManSetNodeAttrs(p, cut.Abc_NtkGetNodeAttributes(network))
	}
	if pParams.FDrop != 0 {
		counts := abc.Abc_NtkFanoutCounts(pNtk)
		cut.Cut_ManSetFanoutCounts(p, vec.Vec_IntArray(counts))
		vec.Vec_IntFree(counts)
	}
	objects := abc.Abc_AigDfs(pNtk, 0, 1)
	progress := extra.Extra_ProgressBarStart(os.Stdout, int32(len(objects)))
	progressStopped := false
	defer func() {
		if !progressStopped {
			extra.Extra_ProgressBarStop(progress)
		}
	}()
	for i, object := range objects {
		if abc.Abc_ObjIsCo(object) != 0 {
			if pParams.FDrop != 0 {
				cut.Cut_NodeTryDroppingCuts(p, int32(abc.Abc_ObjFanin0(object).ID))
			}
			continue
		}
		extra.Extra_ProgressBarUpdate(progress, int32(i), "")
		cuts := Abc_NodeGetCuts(p, object, pParams.FDag, pParams.FTree)
		if pParams.FNpnSave != 0 && cuts != nil {
			for current := cuts; current != nil; current = current.Next {
				if cut.Cut_CutReadLeaveNum(current) >= 4 {
					truth := cut.Cut_CutReadTruth(current)
					if len(truth) != 0 {
						Npn_ManSaveOne(&truth[0], cut.Cut_CutReadLeaveNum(current))
					}
				}
			}
		}
		if pParams.FDrop != 0 {
			cut.Cut_NodeTryDroppingCuts(p, int32(abc.Abc_ObjFanin0(object).ID))
			cut.Cut_NodeTryDroppingCuts(p, int32(abc.Abc_ObjFanin1(object).ID))
		}
		if abc.Abc_AigNodeIsChoice(object) != 0 {
			choices := []int32{int32(object.ID)}
			for choice := object.Choice; choice != nil; choice = choice.Choice {
				choices = append(choices, int32(choice.ID))
			}
			cut.Cut_NodeUnionCuts(p, choices)
		}
	}
	extra.Extra_ProgressBarStop(progress)
	progressStopped = true
	cut.Cut_ManPrintStats(p)
	fmt.Printf("TOTAL =%9.2f sec\n", time.Since(started).Seconds())
	if nTotal != 0 {
		fmt.Printf("Total cuts = %d. Good cuts = %d.  Ratio = %5.2f\n", nTotal, nGood, float64(nGood)/float64(nTotal))
	}
	return p
}

func Abc_NtkCutsOracle(pNtk *Abc_Ntk_t, p *Cut_Oracle_t) {
	assert(pNtk != nil && abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkCutsOracle requires a strashed network")
	assert(p != nil, "Abc_NtkCutsOracle requires oracle")
	fDrop := cut.Cut_OracleReadDrop(p)
	if fDrop != 0 {
		counts := abc.Abc_NtkFanoutCounts(pNtk)
		cut.Cut_OracleSetFanoutCounts(p, vec.Vec_IntArray(counts))
		vec.Vec_IntFree(counts)
	}
	for _, object := range pNtk.Objects() {
		if object.IsCI() && abc.Abc_ObjFanoutNum(object) > 0 {
			cut.Cut_OracleNodeSetTriv(p, int32(object.ID))
		}
	}
	for _, object := range abc.Abc_AigDfs(pNtk, 0, 1) {
		if abc.Abc_ObjIsCo(object) != 0 {
			if fDrop != 0 {
				cut.Cut_OracleTryDroppingCuts(p, int32(abc.Abc_ObjFanin0(object).ID))
			}
			continue
		}
		fanins := object.Fanins()
		cut.Cut_OracleComputeCuts(p, int32(object.ID), int32(fanins[0].Object.ID), int32(fanins[1].Object.ID), boolToInt32(fanins[0].Complement), boolToInt32(fanins[1].Complement))
		if fDrop != 0 {
			cut.Cut_OracleTryDroppingCuts(p, int32(fanins[0].Object.ID))
			cut.Cut_OracleTryDroppingCuts(p, int32(fanins[1].Object.ID))
		}
	}
}

func boolToInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func Abc_NtkSeqCuts(pNtk *Abc_Ntk_t, pParams *Cut_Params_t) *Cut_Man_t {
	assert(pNtk != nil && pParams != nil, "Abc_NtkSeqCuts requires network and parameters")
	pParams.NIdsMax = abc.Abc_NtkObjNumMax(pNtk)
	return cut.Abc_NtkSeqCuts(abcRewriteCutNetwork{network: pNtk}, pParams)
}

func Abc_NtkComputeArea(pNtk *Abc_Ntk_t, p *Cut_Man_t) int32 {
	var area int32
	for i := int32(0); i < abc.Abc_NtkCoNum(pNtk); i++ {
		co := abc.Abc_NtkCo(pNtk, i)
		area += cut.Cut_ManMappingArea_rec(p, int32(co.Fanins()[0].Object.ID))
	}
	return area
}

func Abc_NodeGetCutsRecursive(p any, pObj *Abc_Obj_t, fDag int32, fTree int32) *Cut_Cut_t {
	return cut.Abc_NodeGetCutsRecursive(p.(*Cut_Man_t), abcCutNetworkNode(pObj), fDag, fTree)
}

func Abc_NodeGetCuts(p any, pObj *Abc_Obj_t, fDag int32, fTree int32) *Cut_Cut_t {
	return cut.Abc_NodeGetCuts(p.(*Cut_Man_t), abcCutNetworkNode(pObj), fDag, fTree)
}

func Abc_NodeGetCutsSeq(p any, pObj *Abc_Obj_t, fTriv int32) {
	cut.Abc_NodeGetCutsSeq(p.(*Cut_Man_t), abcCutNetworkNode(pObj), fTriv)
}

func Abc_NodeReadCuts(p any, pObj *Abc_Obj_t) *Cut_Cut_t {
	return cut.Abc_NodeReadCuts(p.(*Cut_Man_t), abcCutNetworkNode(pObj))
}

func Abc_NodeFreeCuts(p any, pObj *Abc_Obj_t) {
	cut.Abc_NodeFreeCuts(p.(*Cut_Man_t), abcCutNetworkNode(pObj))
}

func Abc_NtkPrintCuts(p any, pNtk *Abc_Ntk_t, fSeq int32) {
	cut.Abc_NtkPrintCuts(p.(*Cut_Man_t), abcRewriteCutNetwork{network: pNtk}, fSeq)
}

func Abc_NtkPrintCuts_(p any, pNtk *Abc_Ntk_t, fSeq int32) {
	objects := pNtk.Objects()
	if len(objects) == 0 {
		return
	}
	object := abc.Abc_NtkObj(pNtk, int32(2*abc.Abc_NtkObjNum(pNtk)/3))
	fmt.Printf("Node %s:\n", abc.Abc_ObjName(object))
	cut.Cut_CutPrintList(Abc_NodeReadCuts(p, object), fSeq)
}

// Abc_NtkGetNodeAttributes follows the MFFC threshold used by abcCut.c.
func Abc_NtkGetNodeAttributes(pNtk *Abc_Ntk_t) *Vec_Int_t {
	attrs := vec.Vec_IntStart(int(abc.Abc_NtkObjNumMax(pNtk) + 1))
	var nodesTotal, mffcsTotal int32
	for _, object := range pNtk.Objects() {
		if abc.Abc_ObjIsNode(object) != 0 {
			nodesTotal++
		}
		if abc.Abc_ObjIsCo(object) != 0 && abc.Abc_ObjIsNode(abc.Abc_ObjFanin0(object)) != 0 {
			mffcsTotal += abc.Abc_NodeMffcSize(abc.Abc_ObjFanin0(object))
		}
		if abc.Abc_ObjIsNode(object) != 0 && abc.Abc_ObjFanoutNum(object) > 1 && abc.Abc_NodeIsMuxControlType(object) == 0 {
			mffc := abc.Abc_NodeMffcSize(object)
			mffcsTotal += mffc
			if mffc > 2 || abc.Abc_ObjFanoutNum(object) > 8 {
				vec.Vec_IntWriteEntry(attrs, int(object.ID), 1)
			}
		}
	}
	fmt.Printf("Total nodes = %d. Total MFFC nodes = %d.\n", nodesTotal, mffcsTotal)
	return attrs
}

func Abc_NtkSubDagSize_rec(pObj *Abc_Obj_t, vAttrs *Vec_Int_t) int32 {
	if abc.Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return 0
	}
	abc.Abc_NodeSetTravIdCurrent(pObj)
	if vec.Vec_IntEntry(vAttrs, int(pObj.ID)) != 0 {
		return 0
	}
	if abc.Abc_ObjIsCi(pObj) != 0 {
		return 1
	}
	if abc.Abc_ObjFaninNum(pObj) != 2 {
		panic("Abc_NtkSubDagSize_rec requires a CI or two-fanin node")
	}
	return 1 + Abc_NtkSubDagSize_rec(abc.Abc_ObjFanin0(pObj), vAttrs) + Abc_NtkSubDagSize_rec(abc.Abc_ObjFanin1(pObj), vAttrs)
}

func Abc_NtkGetNodeAttributes2(pNtk *Abc_Ntk_t) *Vec_Int_t {
	if abc.Abc_NtkIsDfsOrdered(pNtk) == 0 {
		panic("cut: node attributes require DFS ordered network")
	}
	attrs := vec.Vec_IntStart(int(abc.Abc_NtkObjNumMax(pNtk) + 1))
	for _, object := range pNtk.Objects() {
		if object.ID == 0 || abc.Abc_ObjIsNode(object) == 0 || abc.Abc_ObjFanoutNum(object) <= 1 || abc.Abc_NodeIsMuxControlType(object) != 0 {
			continue
		}
		abc.Abc_NtkIncrementTravId(pNtk)
		if Abc_NtkSubDagSize_rec(object, attrs) > 15 {
			vec.Vec_IntWriteEntry(attrs, object.ID, 1)
		}
	}
	return attrs
}
