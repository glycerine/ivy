package io

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func IoCommandRead(frame *abcmain.Frame, args []string) int32 {
	check, barriers, flatten, mapped := true, false, false, false
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, option := range args[i][1:] {
			switch option {
			case 'm':
				mapped = !mapped
			case 'c':
				check = !check
			case 'b':
				barriers = !barriers
			case 'g':
				flatten = !flatten
			default:
				fmt.Fprintln(frame.Err, "usage: read [-mcbgh] <file>")
				return 1
			}
		}
		i++
	}
	if i+1 != len(args) {
		fmt.Fprintln(frame.Err, "usage: read [-mcbgh] <file>")
		return 1
	}
	filename := strings.NewReplacer(">", "/", "\\", "/").Replace(args[i])
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".genlib", ".lib", ".scl", ".super", ".constr", ".c", ".script", ".dsd":
		commands := map[string]string{".genlib": "read_genlib", ".lib": "read_lib", ".scl": "read_scl", ".super": "read_super", ".constr": "read_constr", ".c": "so", ".script": "so", ".dsd": "dsd_load"}
		frame.Execute(commands[strings.ToLower(filepath.Ext(filename))] + " " + filename)
		return 0
	}
	if mapped && frame.Resource("pLibGen") == nil {
		fmt.Fprintln(frame.Out, "Cannot read mapped design when the library is not given.")
		return 0
	}
	if flatten {
		n, err := ReadNetwork(filename, Io_ReadFileType(filename), check)
		if err != nil {
			fmt.Fprintln(frame.Err, err)
			return 0
		}
		graph := abc.Abc_NtkFlattenHierarchyGia(n, nil, 0)
		abc.Abc_NtkDelete(n)
		if graph == nil {
			fmt.Fprintln(frame.Out, "Abc_CommandBlast(): Bit-blasting has failed.")
			return 0
		}
		abcmain.Abc_FrameUpdateGia(frame, graph)
		return 0
	}
	n, err := ReadLogic(filename, Io_ReadFileType(filename), check, barriers)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	if abc.Abc_NtkPiNum(n) == 0 {
		fmt.Fprintln(frame.Out, "The new network has no primary inputs. It is recommended\nto add a dummy PI to make sure all commands work correctly.")
	}
	abcmain.Abc_FrameReplaceCurrentNetwork(frame, n)
	abcmain.Abc_FrameClearVerifStatus(frame)
	return 0
}

func IoCommandWrite(frame *abcmain.Frame, args []string) int32 {
	if len(args) != 2 || strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: write [-h] <file>")
		return 1
	}
	filename := args[1]
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".genlib", ".lib", ".dsd":
		commands := map[string]string{".genlib": "write_genlib", ".lib": "write_liberty", ".dsd": "dsd_save"}
		frame.Execute(commands[strings.ToLower(filepath.Ext(filename))] + " " + filename)
		return 0
	}
	n, ok := abcmain.Abc_FrameReadNtk(frame).(*abc.Network)
	if !ok || n == nil {
		fmt.Fprintln(frame.Out, "Empty network.")
		return 0
	}
	if err := WriteNetwork(n, filename, Io_ReadFileType(filename)); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
