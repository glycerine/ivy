package abc

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var networkReader struct {
	sync.RWMutex
	read func(string, int32) *Network
}

// RegisterNetworkReader connects the IO package without creating an import cycle.
func RegisterNetworkReader(read func(string, int32) *Network) {
	networkReader.Lock()
	networkReader.read = read
	networkReader.Unlock()
}

// Abc_NtkPrepareTwoNtks reads or duplicates two networks and strashes them.
func Abc_NtkPrepareTwoNtks(errout io.Writer, current *Network, argv []string, argc int32, first, second **Network, deleteFirst, deleteSecond *int32, check int32) int32 {
	*deleteFirst, *deleteSecond = 0, 0
	if argc < 0 || argc > 2 || int(argc) > len(argv) {
		fmt.Fprintln(errout, "Wrong number of arguments.")
		return 0
	}
	if argc < 2 && current == nil {
		fmt.Fprintln(errout, "Empty current network.")
		return 0
	}
	if argc == 0 {
		if current.Spec == "" {
			fmt.Fprintln(errout, "The external spec is not given.")
			return 0
		}
		f, err := os.Open(current.Spec)
		if err != nil {
			fmt.Fprintf(errout, "Cannot open the external spec file %q.\n", current.Spec)
			return 0
		}
		f.Close()
	}
	networkReader.RLock()
	read := networkReader.read
	networkReader.RUnlock()
	if read == nil {
		panic("abc: network reader is not registered")
	}
	var a, b *Network
	if argc < 2 {
		a = Abc_NtkDup(current)
	} else {
		a = read(argv[0], check)
	}
	if a == nil {
		return 0
	}
	name := currentSpec(current, argv, argc)
	b = read(name, check)
	if b == nil {
		Abc_NtkDelete(a)
		return 0
	}
	if Abc_NtkIsStrash(a) == 0 {
		old := a
		a = Abc_NtkStrash(old, 0, 1, 0)
		Abc_NtkDelete(old)
	}
	if a == nil {
		Abc_NtkDelete(b)
		return 0
	}
	if Abc_NtkIsStrash(b) == 0 {
		old := b
		b = Abc_NtkStrash(old, 0, 1, 0)
		Abc_NtkDelete(old)
	}
	if b == nil {
		Abc_NtkDelete(a)
		return 0
	}
	*first, *second = a, b
	*deleteFirst, *deleteSecond = 1, 1
	return 1
}

func currentSpec(current *Network, argv []string, argc int32) string {
	if argc == 0 {
		return current.Spec
	}
	return argv[argc-1]
}
