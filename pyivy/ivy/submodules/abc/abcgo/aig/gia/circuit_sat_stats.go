package gia

import (
	"fmt"
	"time"
)

type cbsStatisticsRow struct {
	label            string
	calls, conflicts int32
	elapsed          time.Duration
}

// The source CBS engines use the same statistics format and denominators.
func cbsPrintStatistics(graph *Gia_Man_t, budget, frontier, total int32, elapsed time.Duration, rows [3]cbsStatisticsRow) {
	fmt.Printf("CO = %8d  AND = %8d  Conf = %6d  JustMax = %5d  \n", Gia_ManCoNum(graph), Gia_ManAndNum(graph), budget, frontier)
	ratio := func(n, d float64) float64 {
		if d == 0 {
			return 0
		}
		return n / d
	}
	for _, row := range rows {
		fmt.Printf("%s calls %6d  (%6.2f %%)   Ave conf = %8.1f   Time =%9.2f sec (%6.2f %%)\n", row.label, row.calls, 100*ratio(float64(row.calls), float64(total)), ratio(float64(row.conflicts), float64(row.calls)), row.elapsed.Seconds(), 100*ratio(float64(row.elapsed), float64(elapsed)))
	}
	fmt.Printf("Total time =%9.2f sec\n", elapsed.Seconds())
}
