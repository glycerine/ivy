package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
)

// ReadBenchInit applies named register initial values to a network.
func ReadBenchInit(n *abc.Network, input io.Reader, diagnostics io.Writer) error {
	outputs := make(map[string]*abc.Object)
	for _, output := range n.ObjectsOfKind(abc.ObjectBO) {
		outputs[abc.Abc_ObjName(output)] = output
	}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		output := outputs[fields[0]]
		if output == nil {
			fmt.Fprintf(diagnostics, "Io_ReadBenchInit(): Cannot find register with output %s.\n", fields[0])
			continue
		}
		if abc.Abc_ObjFaninNum(output) == 0 || abc.Abc_ObjFanin0(output).Kind != abc.ObjectLatch {
			fmt.Fprintf(diagnostics, "Io_ReadBenchInit(): The signal is not a register output %s.\n", fields[0])
			continue
		}
		if len(fields) < 2 {
			fmt.Fprintf(diagnostics, "Io_ReadBenchInit(): Missing initial value for %s.\n", fields[0])
			continue
		}
		latch := abc.Abc_ObjFanin0(output)
		switch fields[1][0] {
		case '0':
			abc.Abc_LatchSetInit0(latch)
		case '1':
			abc.Abc_LatchSetInit1(latch)
		case '2':
			abc.Abc_LatchSetInitDc(latch)
		default:
			fmt.Fprintf(diagnostics, "Io_ReadBenchInit(): The signal %s has unknown initial value (%s).\n", fields[0], fields[1])
		}
	}
	return scanner.Err()
}

func Io_ReadBenchInit(n *abc.Network, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Io_ReadBenchInit(): Failed to open file %q.\n", filename)
		return
	}
	defer f.Close()
	if err := ReadBenchInit(n, f, os.Stdout); err != nil {
		fmt.Println(err)
	}
}

func IoCommandReadInit(frame *abcmain.Frame, args []string) int32 {
	if len(args) > 2 || len(args) == 2 && strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: read_init [-h] <file>")
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Err, "Empty network.")
		return 1
	}
	var filename string
	if len(args) == 2 {
		filename = args[1]
	} else if n.Spec != "" {
		filename = extra.Extra_FileNameGenericAppend(n.Spec, ".init")
	} else {
		fmt.Fprintln(frame.Out, "File name should be given on the command line.")
		return 1
	}
	copy := abc.Abc_NtkDup(n)
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(frame.Out, "Io_ReadBenchInit(): Failed to open file %q.\n", filename)
	} else {
		if err = ReadBenchInit(copy, f, frame.Out); err != nil {
			fmt.Fprintln(frame.Err, err)
		}
		f.Close()
	}
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, copy)
	abcmain.Abc_FrameClearVerifStatus(frame)
	return 0
}
