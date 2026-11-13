package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"time"
)

type cecSweepTiming struct {
	start                                                         int64
	sat, unsat, undec                                             int32
	satTime, unsatTime, undecTime, simTime, refineTime, extraTime int64
}

func (s *cecSweepTiming) classes(g *gia.Gia_Man_t, verbose bool, veryVerbose bool) {
	if !verbose {
		return
	}
	fmt.Printf("S =%5d U =%5d F =%5d ", s.sat, s.unsat, s.undec)
	v := int32(0)
	if veryVerbose {
		v = 1
	}
	gia.Gia_ManEquivPrintClasses(g, v, 0)
}
func (s *cecSweepTiming) print(verbose bool) {
	if !verbose {
		return
	}
	total := (cecCPUClock() - s.start)
	sat := s.satTime + s.unsatTime + s.undecTime
	entries := []struct {
		name string
		ns   int64
	}{
		{"SAT solving", sat}, {"  sat      ", s.satTime}, {"  unsat    ", s.unsatTime}, {"  fail     ", s.undecTime},
		{"Simulation ", s.simTime}, {"Refinement ", s.refineTime}, {"Rollback   ", s.extraTime},
		{"Other      ", total - sat - s.simTime - s.refineTime - s.extraTime}, {"TOTAL      ", total},
	}
	for _, e := range entries {
		percent := float64(0)
		if total > 0 {
			percent = 100 * float64(e.ns) / float64(total)
		}
		fmt.Printf("%s =%9.2f sec (%6.2f %%)\n", e.name, float64(e.ns)/float64(time.Second), percent)
	}
}
