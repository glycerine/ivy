package io

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

var aigerCanonicalizer struct {
	sync.RWMutex
	fn func(*aig.Graph, int32) *aig.Graph
}

func RegisterAigerCanonicalizer(fn func(*aig.Graph, int32) *aig.Graph) {
	aigerCanonicalizer.Lock()
	aigerCanonicalizer.fn = fn
	aigerCanonicalizer.Unlock()
}

func IoCommandWriteAiger(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "write_aiger", "scuv")
	if !ok {
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	if abc.Abc_NtkIsStrash(n) == 0 {
		fmt.Fprintln(frame.Out, "Writing this format is only possible for structurally hashed AIGs.")
		return 1
	}
	options := ioa.WriteOptions{Symbols: flags['s'], Compact: flags['c']}
	if flags['u'] {
		aigerCanonicalizer.RLock()
		canonicalize := aigerCanonicalizer.fn
		aigerCanonicalizer.RUnlock()
		if canonicalize == nil {
			fmt.Fprintln(frame.Err, "io: AIGER canonicalizer is not registered")
			return 1
		}
		graph := abc.Abc_NtkToDar(n, 0, 1)
		if graph == nil {
			fmt.Fprintln(frame.Err, "io: cannot convert network to AIG")
			return 1
		}
		verbose := int32(0)
		if flags['v'] {
			verbose = 1
		}
		canonical := canonicalize(graph, verbose)
		if canonical == nil {
			fmt.Fprintln(frame.Err, "io: AIG canonicalization failed")
			return 1
		}
		model, err := canonical.AigerModel()
		if err != nil {
			fmt.Fprintln(frame.Err, err)
			return 1
		}
		n, err = NetworkFromAigerModel(model)
		if err != nil {
			fmt.Fprintln(frame.Err, err)
			return 1
		}
		defer abc.Abc_NtkDelete(n)
		options.Symbols, options.Compact = false, false
	}
	if err := WriteAigerWithOptions(n, filename, options); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
