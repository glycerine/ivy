package io

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

var darCNFWriter struct {
	sync.RWMutex
	fn func(*abc.Network, string, int32, int32, int32) *abc.Network
}

func RegisterDarCNFWriter(fn func(*abc.Network, string, int32, int32, int32) *abc.Network) {
	darCNFWriter.Lock()
	darCNFWriter.fn = fn
	darCNFWriter.Unlock()
}

func IoCommandWriteCnf(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "write_cnf", "nfpcv")
	if !ok {
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	allPrimes := flags['p']
	if abc.Abc_NtkIsStrash(n) != 0 && allPrimes {
		allPrimes = false
		fmt.Fprintln(frame.Out, "Warning: Selected option to write all primes has no effect when deriving CNF from AIG.")
	}
	if flags['f'] || !flags['n'] {
		darCNFWriter.RLock()
		writer := darCNFWriter.fn
		darCNFWriter.RUnlock()
		if writer == nil {
			fmt.Fprintln(frame.Err, "io: DAR CNF writer is not registered")
			return 1
		}
		fast, changePol, verbose := int32(0), int32(1), int32(0)
		if flags['f'] {
			fast = 1
		}
		if flags['c'] {
			changePol = 0
		}
		if flags['v'] {
			verbose = 1
		}
		writer(n, filename, fast, changePol, verbose)
	} else if allPrimes {
		Io_WriteCnf(n, filename, 1)
	} else if err := WriteNetwork(n, filename, IO_FILE_CNF); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
