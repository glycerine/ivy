package abci

import (

	acbcore "github.com/glycerine/abcgo/base/acb"
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
		NWinTfoLevs: p.nWinTfoLevs, NFanoutsMax: p.nFanoutsMax, NDepthMax: p.nDepthMax,
		NWinMax: p.nWinMax, NGrowthLevel: p.nGrowthLevel, NBTLimit: p.nBTLimit,
		FRrOnly: p.fRrOnly, FResub: p.fResub, FArea: p.fArea, FMoreEffort: p.fMoreEffort,
		FSwapEdge: p.fSwapEdge, FOneHotness: p.fOneHotness, FDelay: p.fDelay, FPower: p.fPower,
		FGiaSat: p.fGiaSat, FVerbose: p.fVerbose, FVeryVerbose: p.fVeryVerbose,
	}
}

func init() {
	mfscore.NodeIfNodeResyn = Abc_NodeIfNodeResyn
	mfscore.NtkPowerEstimate = Abc_NtkPowerEstimate
	mfscore.NtkMfsTotalSwitching = Abc_NtkMfsTotalSwitching
	mfscore.NtkStrash = Abc_NtkStrash
}

// Abc_NtkMfs runs the port of src/opt/mfs directly on the logic network.
func Abc_NtkMfs(pNtk *Abc_Ntk_t, pPars *Mfs_Par_t) bool {
	if pNtk == nil || pPars == nil {
		return false
	}
	params := mfsParamsFromABC(pPars)
	return mfscore.Abc_NtkMfs(pNtk, &params) != 0
}
