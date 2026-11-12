// Package bbr implements partitioned BDD reachability and witness extraction.
// The algorithms follow ABC's original bdd/bbr sources.
package bbr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
)

type DdManager = cudd.DdManager
type DdNode = cudd.DdNode
type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Saig_ParBbr_t = saig.Saig_ParBbr_t
type Abc_Cex_t = util.Abc_Cex_t

func Aig_ObjGlobalBdd(p *Aig_Man_t, e aig.Edge) *DdNode       { f, _ := p.ObjectData(e).(*DdNode); return f }
func Aig_ObjSetGlobalBdd(p *Aig_Man_t, e aig.Edge, f *DdNode) { p.SetObjectData(e, f) }
func Aig_ObjCleanGlobalBdd(dd *DdManager, p *Aig_Man_t, e aig.Edge) {
	cudd.Cudd_RecursiveDeref(dd, Aig_ObjGlobalBdd(p, e))
	p.SetObjectData(e, nil)
}
func dimensions(p *Aig_Man_t) (pis, regs, pos int, ok bool) {
	if p == nil || p.Check() != nil {
		return
	}
	regs = int(p.RegisterCount())
	pis = p.InputCount() - regs
	pos = p.OutputCount() - regs
	return pis, regs, pos, regs >= 0 && pis >= 0 && pos >= 0
}
