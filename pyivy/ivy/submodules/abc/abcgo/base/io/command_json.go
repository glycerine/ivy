package io

import (
	"fmt"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

func IoCommandReadJson(frame *abcmain.Frame, args []string) int32 {
	if len(args) != 2 || strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: read_json [-h] <file>")
		return 1
	}
	root, err := ReadJSONFile(args[1])
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	objects, names, err := JSONVectors(root)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	rows := make([][]int32, vec.Vec_WecSize(objects))
	for i := range rows {
		rows[i] = append([]int32(nil), vec.Vec_IntArray(vec.Vec_WecEntry(objects, i))...)
	}
	frame.SetResource("pJsonStrs", names)
	frame.SetRows("vJsonObjs", rows)
	return 0
}

func IoCommandWriteJson(frame *abcmain.Frame, args []string) int32 {
	extract := false
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, option := range args[i][1:] {
			if option == 'c' {
				extract = !extract
			} else {
				fmt.Fprintln(frame.Err, "usage: write_json [-ch] <file>")
				return 1
			}
		}
		i++
	}
	names, ok := frame.Resource("pJsonStrs").(*Abc_Nam_t)
	if !ok || names == nil {
		fmt.Fprintln(frame.Out, "No JSON info is available.")
		return 0
	}
	if i+1 != len(args) {
		fmt.Fprintln(frame.Err, "usage: write_json [-ch] <file>")
		return 1
	}
	rows := frame.Rows("vJsonObjs")
	objects := vec.Vec_WecStart(len(rows))
	for i, row := range rows {
		for _, value := range row {
			vec.Vec_WecPush(objects, i, value)
		}
	}
	var err error
	if extract {
		err = Json_Extract(args[i], names, objects)
	} else {
		err = Json_Write(args[i], names, objects)
	}
	if err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
