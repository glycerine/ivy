package io

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func ioWriteOptions(frame *abcmain.Frame, args []string, name, switches string, number rune) (map[rune]bool, int32, string, bool) {
	flags := map[rune]bool{}
	var value int32
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, opt := range args[i][1:] {
			if opt == number {
				flags[opt] = true
				i++
				if i >= len(args) {
					fmt.Fprintf(frame.Err, "usage: %s [options] <file>\n", name)
					return nil, 0, "", false
				}
				v, err := strconv.ParseInt(args[i], 10, 32)
				if err != nil {
					fmt.Fprintf(frame.Err, "usage: %s [options] <file>\n", name)
					return nil, 0, "", false
				}
				value = int32(v)
			} else if strings.ContainsRune(switches, opt) {
				flags[opt] = !flags[opt]
			} else {
				fmt.Fprintf(frame.Err, "usage: %s [options] <file>\n", name)
				return nil, 0, "", false
			}
		}
		i++
	}
	if i+1 != len(args) {
		fmt.Fprintf(frame.Err, "usage: %s [options] <file>\n", name)
		return nil, 0, "", false
	}
	return flags, value, args[i], true
}

func IoCommandWritePla(frame *abcmain.Frame, args []string) int32 {
	flags, count, filename, ok := ioWriteOptions(frame, args, "write_pla", "m", 'M')
	if !ok {
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	if count != 0 {
		if abc.Abc_NtkIsBddLogic(n) != 0 {
			Io_WriteMoPlaM(n, filename, count)
		} else {
			strashed := abc.Abc_NtkStrash(n, 0, 0, 0)
			if strashed != nil {
				Io_WriteMoPlaM(strashed, filename, count)
				abc.Abc_NtkDelete(strashed)
			}
		}
		return 0
	}
	kind := IO_FILE_PLA
	if flags['m'] {
		kind = IO_FILE_MOPLA
	}
	if err := WriteNetwork(n, filename, kind); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}

func writeVerilogFile(n *abc.Network, filename string, onlyAnds int32) error {
	net := abc.Abc_NtkToNetlist(n)
	if net == nil {
		return fmt.Errorf("io: cannot convert network to netlist")
	}
	defer abc.Abc_NtkDelete(net)
	if abc.Abc_NtkHasAig(net) == 0 && abc.Abc_NtkHasMapping(net) == 0 && abc.Abc_NtkToAig(net) == 0 {
		return fmt.Errorf("io: cannot convert network to AIG functions")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	writeErr := WriteVerilog(f, net, onlyAnds)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func IoCommandWriteVerilog(frame *abcmain.Frame, args []string) int32 {
	flags, size, filename, ok := ioWriteOptions(frame, args, "write_verilog", "a", 'K')
	if !ok || flags['K'] && (size < 2 || size > 6) {
		if ok {
			fmt.Fprintln(frame.Err, "usage: write_verilog [-K 2..6] [-ah] <file>")
		}
		return 1
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	var err error
	if flags['a'] {
		err = writeVerilogFile(n, filename, 1)
	} else if size != 0 {
		f, openErr := os.Create(filename)
		if openErr != nil {
			err = openErr
		} else {
			err = WriteVerilogLut(f, n, size)
			closeErr := f.Close()
			if err == nil {
				err = closeErr
			}
		}
	} else {
		err = WriteNetwork(n, filename, IO_FILE_VERILOG)
	}
	if err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}

func init() {
	RegisterFormat(IO_FILE_VERILOG, nil, func(n *abc.Network, filename string) error { return writeVerilogFile(n, filename, 0) })
}
