package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// Saig_ManCexFirstFlopPi finds the first abstracted CI originating as a latch.
func Saig_ManCexFirstFlopPi(g, abstraction *aig.Graph) int32 {
	if len(abstraction.InputNumbers) != len(abstraction.Inputs()) {
		panic("abstraction lacks original input numbers")
	}
	for i, input := range abstraction.InputNumbers {
		if input >= saig.Saig_ManPiNum(g) {
			return int32(i)
		}
	}
	return -1
}
