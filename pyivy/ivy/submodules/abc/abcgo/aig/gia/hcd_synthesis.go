package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

type Hcd_Pars_t struct {
	nWords, nBTLimit, nSatVarMax, fSynthesis, fPolarFlip, fSimulateTfo, fPower, fUseGia, fUseCSat, fVerbose int32
	timeSynth                                                                                               time.Duration
	nNodesAhead, nCallsRecycle                                                                              int32
}

func Hcd_ManSetDefaultParams(p *Hcd_Pars_t) {
	*p = Hcd_Pars_t{nWords: 8, nBTLimit: 1000, nSatVarMax: 5000, fSynthesis: 1, fPolarFlip: 1, fSimulateTfo: 1, nNodesAhead: 1000, nCallsRecycle: 100}
}
func hcdRewrite(g *aig.Graph, p *dar.RewriteParams) *aig.Graph {
	if err := dar.Dar_ManRewrite(g, p); err != nil {
		panic(fmt.Sprintf("hcd: rewrite: %v", err))
	}
	return aig.Aig_ManDupDfs(g)
}
func hcdRefactor(g *aig.Graph, p *dar.RefactorParams) *aig.Graph {
	if err := dar.Dar_ManRefactor(g, p); err != nil {
		panic(fmt.Sprintf("hcd: refactor: %v", err))
	}
	return aig.Aig_ManDupDfs(g)
}
func Hcd_Compress(g *aig.Graph, balance, update, power, verbose int32) *aig.Graph {
	r := dar.DefaultRewriteParams()
	r.UpdateLevel = update != 0
	r.Power = power != 0
	f := dar.DefaultRefactorParams()
	f.UpdateLevel = update != 0
	if verbose != 0 {
		aig.Aig_ManPrintStats(g)
	}
	g = hcdRewrite(g, &r)
	g = hcdRefactor(g, &f)
	if balance != 0 {
		g = dar.Dar_ManBalance(g, update != 0)
	}
	r.UseZeros = true
	g = hcdRewrite(g, &r)
	if verbose != 0 {
		aig.Aig_ManPrintStats(g)
	}
	return g
}
func Hcd_Compress2(g *aig.Graph, balance, update, fanout, power, verbose int32) *aig.Graph {
	r := dar.DefaultRewriteParams()
	r.UpdateLevel = update != 0
	r.Fanout = fanout != 0
	r.Power = power != 0
	f := dar.DefaultRefactorParams()
	f.UpdateLevel = update != 0
	if verbose != 0 {
		aig.Aig_ManPrintStats(g)
	}
	g = hcdRewrite(g, &r)
	g = hcdRefactor(g, &f)
	g = dar.Dar_ManBalance(g, update != 0)
	g = hcdRewrite(g, &r)
	r.UseZeros = true
	f.UseZeros = true
	g = hcdRewrite(g, &r)
	if balance != 0 {
		g = dar.Dar_ManBalance(g, update != 0)
	}
	g = hcdRefactor(g, &f)
	g = hcdRewrite(g, &r)
	if balance != 0 {
		g = dar.Dar_ManBalance(g, update != 0)
	}
	if verbose != 0 {
		aig.Aig_ManPrintStats(g)
	}
	return g
}
func Hcd_ChoiceSynthesis(g *aig.Graph, balance, update, power, verbose int32) *Vec_Ptr_t {
	result := Vec_PtrAlloc(3)
	Vec_PtrPush(result, Gia_ManFromAig(g))
	g = Hcd_Compress(g, balance, update, power, verbose)
	Vec_PtrPush(result, Gia_ManFromAig(g))
	g = Hcd_Compress2(g, balance, update, 1, power, verbose)
	Vec_PtrPush(result, Gia_ManFromAig(g))
	return result
}
func Hcd_ComputeChoices(g *aig.Graph, limit, synthesis, mini, verbose int32) *aig.Graph {
	start := time.Now()
	snapshots := int32(1)
	var miter *Gia_Man_t
	if synthesis != 0 {
		networks := Hcd_ChoiceSynthesis(aig.Aig_ManDupDfs(g), 1, 1, 0, 0)
		snapshots = int32(Vec_PtrSize(networks))
		miter = Hcd_ManChoiceMiter(networks)
	} else {
		miter = Gia_ManFromAig(g)
	}
	Gia_ComputeEquivalences(miter, limit, mini, verbose)
	choices := Hcd_ManEquivToChoices(miter, snapshots)
	Gia_ManSetRegNum(choices, aig.Aig_ManRegNum(g))
	if verbose != 0 {
		fmt.Printf("Choicing time: %s\n", time.Since(start))
	}
	return Gia_ManToAig(choices, 1)
}
func Hcd_ComputeChoicesTest(g *Gia_Man_t, limit, synthesis, mini, verbose int32) {
	Hcd_ComputeChoices(Gia_ManToAig(g, 0), limit, synthesis, mini, verbose)
}
