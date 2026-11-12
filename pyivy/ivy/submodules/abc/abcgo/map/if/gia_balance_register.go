package if_

import "github.com/glycerine/abcgo/aig/gia"

func init() {
	gia.RegisterBalanceEngines(gia.BalanceEngines{
		SOP: Gia_ManPerformSopBalance,
		DSD: Gia_ManPerformDsdBalance,
	})
}
