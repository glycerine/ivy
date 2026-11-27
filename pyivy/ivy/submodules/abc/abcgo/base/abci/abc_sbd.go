package abci

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/opt/sbd"
)

// abcSbdPerform runs Sbd_NtkPerform (opt/sbd/sbdCore.c) on the GIA itself,
// as Abc_CommandAbc9Mfsd does.
func abcSbdPerform(p *Gia_Man_t, pPars *sbd.Sbd_Par_t) (*Gia_Man_t, error) {
	return sbd.Sbd_NtkPerform(p, pPars), nil
}

func abcGiaMuxNum(p *Gia_Man_t) int32 { return gia.Gia_ManMuxNum(p) }
