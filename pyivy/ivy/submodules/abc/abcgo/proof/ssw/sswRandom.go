package ssw

import "github.com/glycerine/abcgo/aig/aig"

// Ssw_RarManPrepareRandom resets ABC's random stream and advances it by seed
// draws, matching the original rarity simulation initialization.
func Ssw_RarManPrepareRandom(seed int32) {
	aig.Aig_ManRandom(1)
	for i := int32(0); i < seed; i++ {
		aig.Aig_ManRandom(0)
	}
}
