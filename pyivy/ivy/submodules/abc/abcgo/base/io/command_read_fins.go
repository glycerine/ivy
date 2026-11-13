package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
)

var finTypes = map[string]int32{"SA0": -99, "SA1": -98, "NEG": -97, "RDOB_AND": -96, "RDOB_NAND": -95, "RDOB_OR": -94, "RDOB_NOR": -93, "RDOB_XOR": -92, "RDOB_NXOR": -91, "RDOB_NOT": -90, "RDOB_BUFF": -89}

func ReadFaultList(n *abc.Network, input io.Reader, verbose bool, diagnostics io.Writer) ([]int32, error) {
	objects := make(map[string]*abc.Object)
	for _, object := range n.Objects() {
		if !object.IsCI() && object.Kind != abc.ObjectNode {
			continue
		}
		name := abc.Abc_ObjName(object)
		if objects[name] != nil {
			return nil, fmt.Errorf("duplicate fault-list object name %q", name)
		}
		objects[name] = object
	}
	scanner := bufio.NewScanner(input)
	pairs := []int32{-1, -1}
	line := 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			break
		}
		line++
		index, err := strconv.Atoi(fields[0])
		if err != nil || index != line {
			return nil, fmt.Errorf("fault-list line numbers are not consecutive")
		}
		if len(fields) < 3 {
			return nil, fmt.Errorf("fault-list line %d is incomplete", line)
		}
		object := objects[fields[1]]
		if object == nil {
			fmt.Fprintf(diagnostics, "Cannot find object with name %q.\n", fields[1])
			continue
		}
		kind, found := finTypes[fields[2]]
		if abc.Abc_NtkIsMappedLogic(n) != 0 && fields[2] != "SA0" && fields[2] != "SA1" && fields[2] != "NEG" {
			found = false
			lib, ok := n.FunctionManager.(*mio.Mio_Library_t)
			if ok && lib != nil {
				gate := mio.Mio_LibraryReadGateByName(lib, fields[2], "")
				if gate != nil {
					kind, found = mio.Mio_GateReadCell(gate), true
				}
			}
		}
		if !found {
			fmt.Fprintf(diagnostics, "Cannot read type %q of object %q.\n", fields[2], fields[1])
			continue
		}
		pairs = append(pairs, int32(object.ID), kind)
		if verbose {
			fmt.Fprintf(diagnostics, "%-10d%-10s%-10s\n", len(pairs)/2-1, fields[1], fields[2])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return pairs, nil
}

func IoCommandReadFins(frame *abcmain.Frame, args []string) int32 {
	flags, filename, ok := ioReadOptions(frame, args, "read_fins", "v")
	if !ok {
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Err, "Empty network.")
		return 1
	}
	f, err := os.Open(filename)
	if err != nil {
		n.Faults = nil
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	defer f.Close()
	n.Faults, err = ReadFaultList(n, f, flags['v'], frame.Out)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
