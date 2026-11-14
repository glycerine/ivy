package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
	"time"
)

// Ga2_ManStart constructs the source GLA manager and its marked cut map.
func Ga2_ManStart(g *gia.Gia_Man_t, pars *Abs_Par_t) *Ga2_Man_t {
	if g == nil || pars == nil {
		panic("abs: missing GLA inputs")
	}
	p := &Ga2_Man_t{pGia: g, pPars: pars, timeStart: time.Now().UnixNano(), fUseNewLine: 1}
	if pars.NTimeOut != 0 {
		p.cpuDeadline = absCPUClock() + int64(pars.NTimeOut)*int64(time.Second)
	}
	p.nMarked = Ga2_ManMarkup(g, 5, pars.FUseSimple)
	p.vCnfs = []*vec.Vec_Int_t{vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)}
	p.vIds = vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(g)))
	p.vProofIds = vec.Vec_IntAlloc(0)
	p.vAbs = vec.Vec_IntAlloc(1000)
	p.vValues = vec.Vec_IntAlloc(1000)
	Ga2_ObjSetId(p, gia.Gia_ManConst0(g), 0)
	vec.Vec_IntPush(p.vValues, 0)
	vec.Vec_IntPush(p.vAbs, 0)
	p.pRnm = Rnm_ManStart(g)
	p.vId2Lit = make([]*vec.Vec_Int_t, 0, 1000)
	p.vLits = vec.Vec_IntAlloc(100)
	p.vIsopMem = vec.Vec_IntAlloc(100)
	p.pSops = cnf.Cnf_ReadMsops()
	p.nTable = util.Abc_PrimeCudd(1 << 18)
	p.pTable = make([]int32, 6*int(p.nTable))
	return p
}
