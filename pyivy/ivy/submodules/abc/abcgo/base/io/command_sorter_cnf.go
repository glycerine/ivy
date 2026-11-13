package io

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	abcmain "github.com/glycerine/abcgo/base/main"
)

var sorterCNFWriter struct {
	sync.RWMutex
	fn func(string, int32, int32)
}

func RegisterSorterCNFWriter(fn func(string, int32, int32)) {
	sorterCNFWriter.Lock()
	sorterCNFWriter.fn = fn
	sorterCNFWriter.Unlock()
}

func IoCommandWriteSortCnf(frame *abcmain.Frame, args []string) int32 {
	variables, queens := int32(16), int32(4)
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, option := range args[i][1:] {
			if option != 'N' && option != 'Q' {
				fmt.Fprintln(frame.Err, "usage: write_sorter_cnf [-N num] [-Q num] <file>")
				return 1
			}
			i++
			if i >= len(args) {
				fmt.Fprintln(frame.Err, "usage: write_sorter_cnf [-N num] [-Q num] <file>")
				return 1
			}
			value, err := strconv.ParseInt(args[i], 10, 32)
			if err != nil || value <= 0 {
				fmt.Fprintln(frame.Err, "usage: write_sorter_cnf [-N num] [-Q num] <file>")
				return 1
			}
			if option == 'N' {
				variables = int32(value)
			} else {
				queens = int32(value)
			}
		}
		i++
	}
	if i+1 != len(args) || queens >= variables {
		fmt.Fprintln(frame.Err, "usage: write_sorter_cnf [-N num] [-Q num] <file> (0 < Q < N)")
		return 1
	}
	sorterCNFWriter.RLock()
	writer := sorterCNFWriter.fn
	sorterCNFWriter.RUnlock()
	if writer == nil {
		fmt.Fprintln(frame.Err, "io: sorter CNF writer is not registered")
		return 1
	}
	writer(args[i], variables, queens)
	return 0
}
