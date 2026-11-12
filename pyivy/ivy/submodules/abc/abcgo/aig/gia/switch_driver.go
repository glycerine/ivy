package gia

import (
	"fmt"
	"math"
	"strconv"
	"time"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func Gia_ManSwiComputeSwitching(ones, words int32) float32 {
	total := float32(32 * words)
	return 2 * float32(ones) / total * (total - float32(ones)) / total
}
func Gia_ManSwiComputeProbOne(ones, words int32) float32 { return float32(ones) / float32(32*words) }

// Gia_ManSwiSimulate returns float32 probabilities encoded as int32 bit
// patterns, matching ABC's Vec_Int_t representation.
func Gia_ManSwiSimulate(graph *Gia_Man_t, pars *Gia_ParSwi_t) *Vec_Int_t {
	if pars == nil || pars.nWords <= 0 || pars.nIters <= pars.nPref || pars.nPref < 0 {
		panic("gia: invalid switching simulation frames")
	}
	if pars.fProbOne != 0 && pars.fProbTrans != 0 {
		fmt.Println("Conflict of options: Can either compute probability of 1, or probability of switching by observing transitions.")
	}
	start := time.Now()
	p := Gia_ManSwiCreate(graph, pars)
	defer Gia_ManSwiDelete(p)
	Gia_ManRandom(1)
	Gia_ManSwiSimInfoInit(p)
	for i := int32(0); i < pars.nIters; i++ {
		Gia_ManSwiSimulateRound(p, giaBool(i >= pars.nPref))
		if i == pars.nIters-1 {
			break
		}
		if pars.fProbTrans != 0 {
			Gia_ManSwiSimInfoTransferShift(p, pars.nRandPiFactor)
		} else {
			Gia_ManSwiSimInfoTransfer(p, pars.nRandPiFactor)
		}
	}
	if pars.fVerbose != 0 {
		fmt.Printf("Simulated %d frames with %d words in %s.\n", pars.nIters, pars.nWords, time.Since(start))
	}
	result := Vec_IntStart(int(Gia_ManObjNum(graph)))
	words := pars.nWords * (pars.nIters - pars.nPref)
	for id, count := range p.pData1 {
		var value float32
		if pars.fProbOne != 0 || pars.fProbTrans != 0 {
			value = Gia_ManSwiComputeProbOne(count, words)
		} else {
			value = Gia_ManSwiComputeSwitching(count, words)
		}
		Vec_IntWriteEntry(result, id, int32(math.Float32bits(value)))
	}
	if pars.fProbOne != 0 {
		for i := int32(0); i < Gia_ManCoNum(graph); i++ {
			obj := Gia_ManCo(graph, i)
			value := math.Float32frombits(uint32(Vec_IntEntry(result, int(Gia_ObjFaninId0p(graph, obj)))))
			if Gia_ObjFaninC0(obj) != 0 {
				value = 1 - value
			}
			Vec_IntWriteEntry(result, int(obj.id), int32(math.Float32bits(value)))
		}
	}
	return result
}
func Gia_ManComputeSwitchProbs(graph *Gia_Man_t, frames, prefix, probOne int32) *Vec_Int_t {
	var pars Gia_ParSwi_t
	Gia_ManSetDefaultParamsSwi(&pars)
	pars.nIters, pars.nPref = frames, prefix
	if flag := abcmain.Abc_FrameReadFlag("seqsimframes"); flag != "" {
		n, _ := strconv.Atoi(flag)
		pars.nIters = int32(n)
	}
	if probOne != 0 {
		pars.fProbOne, pars.fProbTrans = 1, 0
	} else {
		pars.fProbOne, pars.fProbTrans = 0, 1
	}
	return Gia_ManSwiSimulate(graph, &pars)
}
