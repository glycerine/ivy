package gia

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManComputeSwitching sums switching activity on gate or LUT inputs.
func Gia_ManComputeSwitching(p *Gia_Man_t, frames, prefix, probOne int32) float32 {
	if LfSwitchingProvider == nil {
		panic("gia: switching simulator is not connected")
	}
	probs := LfSwitchingProvider(p, frames, prefix, probOne)
	if probs == nil || vec.Vec_FltSize(probs) < int(p.nObjs) {
		panic("gia: invalid switching probabilities")
	}
	var total float32
	if Gia_ManHasMapping(p) != 0 {
		for id := int32(1); id < p.nObjs; id++ {
			if Gia_ObjIsLut(p, id) != 0 {
				for _, fanin := range Gia_ObjLutFanins(p, id) {
					total += vec.Vec_FltEntry(probs, int(fanin))
				}
			}
		}
	} else {
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				total += vec.Vec_FltEntry(probs, int(Gia_ObjFanin0(obj).id)) + vec.Vec_FltEntry(probs, int(Gia_ObjFanin1(obj).id))
			}
		}
	}
	return total
}

var giaPowerHistory struct {
	sync.Mutex
	count    int32
	previous float32
}

func Gia_ManPrintStats(p *Gia_Man_t, pars *Gps_Par_t) {
	hadLevels := p.vLevels != nil
	if pars != nil && pars.FMiter != 0 {
		Gia_ManPrintStatsMiter(p, 0)
		return
	}
	noColor := pars != nil && pars.FNoColor != 0
	if p.pName != "" {
		if noColor {
			fmt.Printf("%-8s : ", p.pName)
		} else {
			fmt.Printf("\033[1;37m%-8s\033[0m : ", p.pName)
		}
	}
	fmt.Printf("i/o =%7d/%7d", Gia_ManPiNum(p)-Gia_ManBoxCiNum(p)-Gia_ManRegBoxNum(p), Gia_ManPoNum(p)-Gia_ManBoxCoNum(p)-Gia_ManRegBoxNum(p))
	if n := Gia_ManConstrNum(p); n != 0 {
		fmt.Printf("(c=%d)", n)
	}
	if n := Gia_ManRegNum(p); n != 0 {
		fmt.Printf("  ff =%7d", n)
	}
	if n := Gia_ManRegBoxNum(p); n != 0 {
		fmt.Printf("  boxff =%d(%d)", n, Gia_ManClockDomainNum(p))
	}
	kind := "and"
	if p.PMuxes != nil {
		kind = "nod"
	}
	level, average := Gia_ManLevelNum(p), Gia_ManLevelAve(p)
	if noColor {
		fmt.Printf("  %s =%8d  lev =%5d (%.2f)", kind, Gia_ManAndNum(p), level, average)
	} else {
		fmt.Printf("  \033[1;36m%s =%8d\033[0m  \033[1;35mlev =%5d\033[0m \033[1;35m(%.2f)\033[0m", kind, Gia_ManAndNum(p), level, average)
	}
	if !hadLevels {
		p.vLevels = nil
	}
	if pars != nil && pars.FCut != 0 {
		fmt.Printf("  cut = %d(%d)", Gia_ManCrossCut(p, 0), Gia_ManCrossCut(p, 1))
	}
	fmt.Printf("  mem =%5.2f MB", Gia_ManMemory(p)/(1<<20))
	if Gia_ManHasChoices(p) != 0 {
		fmt.Printf("  ch =%5d", Gia_ManChoiceNum(p))
	}
	if p.PManTime != nil {
		fmt.Printf("  box = %d  bb = %d", Gia_ManNonRegBoxNum(p), Gia_ManBlackBoxNum(p))
	}
	if n := Gia_ManBufNum(p); n != 0 {
		fmt.Printf("  buf = %d", n)
	}
	if n := Gia_ManXorNum(p); n != 0 && p.PMuxes == nil {
		fmt.Printf("  xor = %d", n)
	}
	if pars != nil && pars.FMuxXor != 0 {
		fmt.Print("\nXOR/MUX ")
		Gia_ManPrintMuxStats(p)
	}
	if pars != nil && pars.FSwitch != 0 {
		power := Gia_ManComputeSwitching(p, 48, 16, 0)
		fmt.Printf("  power =%8.1f", power)
		giaPowerHistory.Lock()
		n := Gia_ManCiNum(p) + Gia_ManCoNum(p)
		if giaPowerHistory.previous > 0 && giaPowerHistory.count == n {
			fmt.Printf(" %6.2f %%", 100*(giaPowerHistory.previous-power)/giaPowerHistory.previous)
		} else if giaPowerHistory.previous == 0 || giaPowerHistory.count != n {
			giaPowerHistory.previous, giaPowerHistory.count = power, n
		}
		giaPowerHistory.Unlock()
	}
	fmt.Println()
	if p.PReprs != nil && p.PNexts != nil {
		Gia_ManEquivPrintClasses(p, 0, 0)
	}
	dump := ""
	if pars != nil {
		dump = pars.PDumpFile
	}
	if Gia_ManHasMapping(p) != 0 && (pars == nil || pars.FSkipMap == 0) {
		Gia_ManPrintMappingStats(p, dump)
	} else if dump != "" {
		Gia_ManLogAigStats(p, dump)
	}
	if pars != nil && pars.FNpn != 0 && Gia_ManHasMapping(p) != 0 && Gia_ManLutSizeMax(p) <= 4 {
		Gia_ManPrintNpnClasses(p)
	}
	if p.VPacking != nil {
		Gia_ManPrintPackingStats(p)
	}
	if p.VEdge1 != nil {
		Gia_ManPrintEdges(p)
	}
	if pars != nil && pars.FLutProf != 0 && Gia_ManHasMapping(p) != 0 {
		Gia_ManPrintLutStats(p)
	}
	if p.pPlacement != nil {
		Gia_ManPrintPlacement(p)
	}
	Gia_ManPrintFlopClasses(p)
	Gia_ManPrintGateClasses(p)
	Gia_ManPrintObjClasses(p)
	if p.VInitClasses != nil {
		Gia_ManPrintInitClasses(p.VInitClasses)
	}
	Gia_ManCheckIntegrityWithBoxes(p)
	if pars != nil && pars.FTents != 0 {
		Gia_ManPrintTents(p)
	}
	if pars != nil && pars.FSlacks != 0 {
		Gia_ManDfsSlacksPrint(p)
	}
}
