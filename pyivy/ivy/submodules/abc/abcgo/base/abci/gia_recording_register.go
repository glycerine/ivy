package abci

import "github.com/glycerine/abcgo/aig/gia"

func init() {
	gia.RegisterRecordingState(func() bool { return Abc_NtkRecIsRunning3() != 0 })
}
