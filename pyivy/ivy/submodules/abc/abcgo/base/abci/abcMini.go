package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/miniaig"
	baseabc "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// Abc_NtkInputMiniAig imports a MiniAIG into the frame as an ABC network.
// The input manager remains owned by the caller, matching the C API.
func Abc_NtkInputMiniAig(frame *Abc_Frame_t, value any) {
	if frame == nil {
		fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n")
		return
	}
	p, ok := value.(*miniaig.Mini_Aig_t)
	if !ok || p == nil {
		fmt.Printf("Input value is not a MiniAIG.\n")
		return
	}
	network := Abc_NtkFromMiniAig(p)
	if network != nil {
		abcmain.Abc_FrameReplaceCurrentNetwork(frame, network)
	}
}

// Abc_NtkOutputMiniAig exports the current strashed network. The returned
// manager is owned by the caller and must be released with Mini_AigStop.
func Abc_NtkOutputMiniAig(frame *Abc_Frame_t) any {
	if frame == nil {
		fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n")
		return nil
	}
	network := Abc_FrameReadNtk(frame)
	if network == nil {
		fmt.Printf("Current network in ABC framework is not defined.\n")
		return nil
	}
	return Abc_NtkToMiniAig(network)
}

func Abc_NtkSetFlopNum(frame *Abc_Frame_t, nFlops int32) {
	if frame == nil {
		fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n")
		return
	}
	network := Abc_FrameReadNtk(frame)
	if network == nil {
		fmt.Printf("Current network in ABC framework is not defined.\n")
		return
	}
	baseabc.Abc_NtkMakeSeq(network, nFlops)
}

// Abc_NtkMiniAigTest exercises conversion and the MiniAIG file roundtrip.
func Abc_NtkMiniAigTest(network *Abc_Ntk_t) {
	if network == nil {
		return
	}
	mini := Abc_NtkToMiniAig(network)
	converted := Abc_NtkFromMiniAig(mini)
	miniaig.Mini_AigStop(mini)
	Abc_NtkPrintStats(converted, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	Abc_NtkDelete(converted)

	mini = Abc_NtkToMiniAig(network)
	miniaig.Mini_AigDump(mini, "miniaig.data")
	miniaig.Mini_AigPrintStats(mini)
	miniaig.Mini_AigStop(mini)

	mini = miniaig.Mini_AigLoad("miniaig.data")
	if mini != nil {
		miniaig.Mini_AigPrintStats(mini)
		miniaig.Mini_AigStop(mini)
	}
}
