package abci

import (
	"github.com/glycerine/abcgo/aig/gia"
	ifpkg "github.com/glycerine/abcgo/map/if"
)

func init() {
	gia.RegisterRecordingState(func() bool { return Abc_NtkRecIsRunning3() != 0 })
	ifpkg.RegisterRecordedLibrary(ifpkg.RecordedLibrary{DelayCost: If_CutDelayRecCost3, BuildGia: Abc_RecToGia3})
}
