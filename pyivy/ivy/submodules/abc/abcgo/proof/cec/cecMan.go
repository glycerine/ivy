package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
)

// Cec_ManSat_t owns incremental cone proofs and the most recent validated model.
// The source-shaped entry points retain their status convention: 1 proved,
// 0 disproved, -1 interrupted or invalid (Error describes the latter).
type Cec_ManSat_t struct {
	pAig   *Gia_Man_t
	pPars  *Cec_ParSat_t
	proof  *core.Manager
	result core.Result
	values []bool
	vCex   *Vec_Int_t
	Error  error
}

func Cec_ManSatCreate(graph *Gia_Man_t, params *Cec_ParSat_t) *Cec_ManSat_t {
	if params == nil {
		params = new(Cec_ParSat_t)
		Cec_ManSatSetDefaultParams(params)
	}
	p := &Cec_ManSat_t{pAig: graph, pPars: params}
	p.restart()
	return p
}
func (p *Cec_ManSat_t) restart() {
	p.values = nil
	p.vCex = vec.Vec_IntAlloc(16)
	o := core.Options{ConflictLimit: int64(p.pPars.nBTLimit), MaxSATVariables: int(p.pPars.nSatVarMax), RecycleQueries: int(p.pPars.nCallsRecycle), PolarFlip: p.pPars.fPolarFlip != 0, TreatRegistersAsInputs: true, Deadline: p.pPars.deadline}
	p.proof, p.Error = core.NewManager(p.pAig, o)
}
func Cec_ManSatPrintStats(p *Cec_ManSat_t) {
	if p.proof == nil {
		fmt.Printf("CEC SAT: %v\n", p.Error)
		return
	}
	s := p.proof.Statistics()
	fmt.Printf("CEC SAT: %d queries, %d SAT calls, %d conflicts, %d recycles\n", s.Queries, s.SATCalls, s.Conflicts, s.Recycles)
}
func Cec_ManSatStop(p *Cec_ManSat_t) {
	if p != nil {
		p.proof = nil
		p.values = nil
	}
}

func Cec_ManFraStart(graph *Gia_Man_t, params *Cec_ParFra_t) *Cec_ManFra_t {
	if params == nil {
		params = new(Cec_ParFra_t)
		Cec_ManFraSetDefaultParams(params)
	}
	return &Cec_ManFra_t{pAig: graph, pPars: params, vXorNodes: vec.Vec_IntAlloc(1000)}
}
func Cec_ManFraStop(p *Cec_ManFra_t) {
	if p != nil {
		p.vXorNodes = nil
	}
}
