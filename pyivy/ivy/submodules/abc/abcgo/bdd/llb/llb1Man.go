package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func Llb_ManPrepareVarMap(manager *Llb_Man_t) {
	assert(manager.vNs2Glo == nil && manager.vCs2Glo == nil && manager.vGlo2Cs == nil && manager.vGlo2Ns == nil, "variable maps are not initialized")
	variables := vec.Vec_IntSize(manager.vVar2Obj)
	registers := int(aig.Aig_ManRegNum(manager.pAig))
	manager.vNs2Glo = vec.Vec_IntStartFull(variables)
	manager.vCs2Glo = vec.Vec_IntStartFull(variables)
	manager.vGlo2Cs = vec.Vec_IntStartFull(registers)
	manager.vGlo2Ns = vec.Vec_IntStartFull(registers)
	inputStart := len(manager.pAig.Inputs()) - registers
	outputStart := len(manager.pAig.Outputs()) - registers
	for i := 0; i < registers; i++ {
		li := aig.Aig_ManCo(manager.pAig, int32(outputStart+i))
		lo := aig.Aig_ManCi(manager.pAig, int32(inputStart+i))
		varLi := vec.Vec_IntEntry(manager.vObj2Var, int(aig.Aig_ObjId(li)))
		varLo := vec.Vec_IntEntry(manager.vObj2Var, int(aig.Aig_ObjId(lo)))
		assert(varLi >= 0 && int(varLi) < variables, "latch input has an indexed variable")
		assert(varLo >= 0 && int(varLo) < variables, "latch output has an indexed variable")
		vec.Vec_IntWriteEntry(manager.vNs2Glo, int(varLi), int32(i))
		vec.Vec_IntWriteEntry(manager.vCs2Glo, int(varLo), int32(i))
		vec.Vec_IntWriteEntry(manager.vGlo2Cs, i, varLo)
		vec.Vec_IntWriteEntry(manager.vGlo2Ns, i, varLi)
	}
	primaryInputs := int32(inputStart)
	for i := int32(0); i < primaryInputs; i++ {
		pi := aig.Aig_ManCi(manager.pAig, i)
		variable := vec.Vec_IntEntry(manager.vObj2Var, int(aig.Aig_ObjId(pi)))
		assert(variable >= 0 && int(variable) < variables, "primary input has an indexed variable")
		global := int32(registers) + i
		vec.Vec_IntWriteEntry(manager.vCs2Glo, int(variable), global)
		vec.Vec_IntWriteEntry(manager.vNs2Glo, int(variable), global)
	}
}

func Llb_ManPrepareVarLimits(manager *Llb_Man_t) {
	assert(manager.vVarBegs == nil && manager.vVarEnds == nil, "variable limits are not initialized")
	objCount := int(aig.Aig_ManObjNumMax(manager.pAig))
	manager.vVarEnds = vec.Vec_IntStart(objCount)
	manager.vVarBegs = vec.Vec_IntStart(objCount)
	vec.Vec_IntFill(manager.vVarBegs, objCount, manager.pMatrix.nCols)
	for groupID := int32(0); groupID < manager.pMatrix.nCols; groupID++ {
		group := manager.pMatrix.pColGrps[groupID]
		if group == nil {
			continue
		}
		for _, boundary := range []*Vec_Ptr_t{group.vIns, group.vOuts} {
			for i := 0; i < vec.Vec_PtrSize(boundary); i++ {
				edge := llbGroupEdge(boundary, i)
				id := int(aig.Aig_ObjId(edge))
				if vec.Vec_IntEntry(manager.vVarBegs, id) > groupID {
					vec.Vec_IntWriteEntry(manager.vVarBegs, id, groupID)
				}
				if vec.Vec_IntEntry(manager.vVarEnds, id) < groupID {
					vec.Vec_IntWriteEntry(manager.vVarEnds, id, groupID)
				}
			}
		}
	}
}

func Llb_ManStop(manager *Llb_Man_t) {
	if manager == nil {
		return
	}
	if manager.pMatrix != nil {
		Llb_MtrFree(manager.pMatrix)
		manager.pMatrix = nil
	}
	for i := 0; i < vec.Vec_PtrSize(manager.vGroups); i++ {
		if group, ok := vec.Vec_PtrEntry(manager.vGroups, i).(*Llb_Grp_t); ok {
			Llb_ManGroupStop(group)
		}
	}
	vec.Vec_PtrFree(manager.vGroups)
	manager.vGroups = nil
	for _, dd := range []*core.DdManager{manager.dd, manager.ddG, manager.ddR} {
		core.Cudd_Quit(dd)
	}
	manager.dd, manager.ddG, manager.ddR = nil, nil, nil
	if manager.pAig != nil {
		aig.Aig_ManStop(manager.pAig)
		manager.pAig = nil
	}
	for _, v := range []*Vec_Int_t{
		manager.vVar2Obj, manager.vObj2Var, manager.vVarBegs, manager.vVarEnds,
		manager.vNs2Glo, manager.vCs2Glo, manager.vGlo2Cs, manager.vGlo2Ns,
	} {
		vec.Vec_IntFree(v)
	}
	manager.vVar2Obj, manager.vObj2Var, manager.vVarBegs, manager.vVarEnds = nil, nil, nil, nil
	manager.vNs2Glo, manager.vCs2Glo, manager.vGlo2Cs, manager.vGlo2Ns = nil, nil, nil, nil
	vec.Vec_PtrFree(manager.vRings)
	manager.vRings = nil
}

func Llb_ManStart(globalAig, workingAig *Aig_Man_t, params *Gia_ParLlb_t) *Llb_Man_t {
	if globalAig == nil || workingAig == nil || params == nil {
		return nil
	}
	aig.Aig_ManCleanMarkA(workingAig)
	manager := &Llb_Man_t_{pAigGlo: globalAig, pPars: params, pAig: workingAig}
	manager.vVar2Obj = Llb_ManMarkPivotNodes(manager.pAig, params.fUsePivots)
	manager.vObj2Var = vec.Vec_IntInvert(manager.vVar2Obj, -1)
	manager.vRings = vec.Vec_PtrAlloc(100)
	Llb_ManPrepareVarMap(manager)
	Llb_ManPrepareGroups(manager)
	aig.Aig_ManCleanMarkA(workingAig)
	manager.pMatrix = Llb_MtrCreate(manager)
	manager.pMatrix.pMan = manager
	return manager
}
