package io

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"strconv"
	"strings"
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
