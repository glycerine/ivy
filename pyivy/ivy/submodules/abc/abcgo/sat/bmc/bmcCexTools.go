package bmc

import (
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Bmc_CexBitMask(bit int32) uint64 {
	if bit < 0 || bit >= 64 {
		panic("bmc: invalid simulation bit")
	}
	return ^(uint64(1) << uint(bit))
}
func Bmc_CexCopySim(v *vec.Vec_Wrd_t, to, from int32) {
	vec.Vec_WrdWriteEntry(v, int(to), vec.Vec_WrdEntry(v, int(from)))
}
func Bmc_CexAndSim(v *vec.Vec_Wrd_t, to, a, b int32) {
	vec.Vec_WrdWriteEntry(v, int(to), vec.Vec_WrdEntry(v, int(a))&vec.Vec_WrdEntry(v, int(b)))
}
func Bmc_CexOrSim(v *vec.Vec_Wrd_t, to, a, b int32) {
	vec.Vec_WrdWriteEntry(v, int(to), vec.Vec_WrdEntry(v, int(a))|vec.Vec_WrdEntry(v, int(b)))
}
func Bmc_CexSim(v *vec.Vec_Wrd_t, obj, bit int32) int32 {
	return int32(vec.Vec_WrdEntry(v, int(obj)) >> uint(bit) & 1)
}
func Bmc_CexBitCount(c *util.Abc_Cex_t, inputs int32) int32 {
	if c == nil {
		return -1
	}
	if c.NPis <= 0 {
		return 0
	}
	var count int32
	for k := c.NRegs; k < c.NBits; k++ {
		if (k-c.NRegs)%c.NPis < inputs && util.Abc_InfoHasBit(c.PData, k) {
			count++
		}
	}
	return count
}
func Bmc_CexPrint(c *util.Abc_Cex_t, realPis, verbose int32) {
	if c == nil {
		fmt.Println("The counterexample is nil.")
		return
	}
	util.Abc_CexPrintStatsInputs(c, realPis)
	if verbose == 0 {
		return
	}
	for f := int32(0); f <= c.IFrame; f++ {
		fmt.Printf("%3d : ", f)
		for i := int32(0); i < c.NPis; i++ {
			if i == realPis {
				fmt.Print(" ")
			}
			if util.Abc_InfoHasBit(c.PData, c.NRegs+f*c.NPis+i) {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
		}
		fmt.Println()
	}
}
func Bmc_CexDumpStats(p *gia.Gia_Man_t, c, care, essential, minimum *util.Abc_Cex_t, elapsed time.Duration) {
	f, err := os.OpenFile("cex/stats.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	inputs := gia.Gia_ManPiNum(p)
	total := (c.IFrame + 1) * inputs
	careCount, ess, minCount := Bmc_CexBitCount(care, inputs), Bmc_CexBitCount(essential, inputs), Bmc_CexBitCount(minimum, inputs)
	pct := func(n int32) float64 {
		if total == 0 {
			return 0
		}
		return 100 * float64(n) / float64(total)
	}
	fmt.Fprintf(f, "%s %d %d %d %d %.2f %.2f %.2f %.2f %.2f\n", gia.Gia_ManName(p), inputs, gia.Gia_ManRegNum(p), c.IFrame+1, total, pct(total-careCount), pct(careCount-ess), pct(ess), pct(minCount), elapsed.Seconds())
}
func Bmc_CexDumpAogStats(p *gia.Gia_Man_t, elapsed time.Duration) {
	f, err := os.OpenFile("cex/aig_stats.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %d %d %d %.2f\n", gia.Gia_ManName(p), gia.Gia_ManPiNum(p), gia.Gia_ManAndNum(p), gia.Gia_ManLevelNum(p), elapsed.Seconds())
}
func Bmc_CexPerformUnrollingTest(p *gia.Gia_Man_t, c *util.Abc_Cex_t) {
	if q := Bmc_CexPerformUnrolling(p, c); q != nil {
		if err := gia.Gia_AigerWrite(q, "unroll.aig", 0, 0, 0); err != nil {
			fmt.Println(err)
		}
	}
}
func Bmc_CexBuildNetworkTest(p *gia.Gia_Man_t, c *util.Abc_Cex_t) {
	if q := Bmc_CexBuildNetwork(p, c); q != nil {
		if err := gia.Gia_AigerWrite(q, "unate.aig", 0, 0, 0); err != nil {
			fmt.Println(err)
		}
	}
}
func Bmc_CexTest(p *gia.Gia_Man_t, c *util.Abc_Cex_t, verbose int32) {
	var implied *util.Abc_Cex_t
	states := Bmc_CexInnerStates(p, c, &implied, verbose)
	care := Bmc_CexCareBits(p, states, implied, nil, 1, verbose)
	if Bmc_CexVerify(p, c, care) == 0 {
		fmt.Println("Counter-example care-set verification has failed.")
	}
	essential := Bmc_CexEssentialBits(p, states, care, verbose)
	minimum := Bmc_CexCareBits(p, states, implied, essential, 0, verbose)
	if Bmc_CexVerify(p, c, minimum) == 0 {
		fmt.Println("Counter-example min-set verification has failed.")
	}
}
