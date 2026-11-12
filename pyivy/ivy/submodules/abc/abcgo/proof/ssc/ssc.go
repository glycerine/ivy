// Package ssc implements simulation-guided SAT sweeping under constraints.
// Care outputs are violation predicates: every care output must be false.
package ssc

import "github.com/glycerine/abcgo/aig/gia"

type Gia_Man_t = gia.Gia_Man_t
type Ssc_Pars_t struct {
	NWords                  int32
	BTLimit                 int32
	SatVarMax               int32
	CallsRecycle            int32
	Append, Verbose, Verify bool
}

func Ssc_ManSetDefaultParams(p *Ssc_Pars_t) {
	*p = Ssc_Pars_t{NWords: 1, BTLimit: 1000, SatVarMax: 5000, CallsRecycle: 100}
}
func DefaultParams() Ssc_Pars_t { var p Ssc_Pars_t; Ssc_ManSetDefaultParams(&p); return p }
