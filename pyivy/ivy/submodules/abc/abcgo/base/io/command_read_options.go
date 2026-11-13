package io

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"strings"
)

func ioReadOptions(frame *abcmain.Frame, args []string, name, allowed string) (map[rune]bool, string, bool) {
	flags := make(map[rune]bool)
	i := 1
	for i < len(args) && strings.HasPrefix(args[i], "-") && args[i] != "-" {
		for _, option := range args[i][1:] {
			if option == 'h' || !strings.ContainsRune(allowed, option) {
				fmt.Fprintf(frame.Err, "usage: %s [-%sh] <file>\n", name, allowed)
				return nil, "", false
			}
			flags[option] = !flags[option]
		}
		i++
	}
	if i+1 != len(args) {
		fmt.Fprintf(frame.Err, "usage: %s [-%sh] <file>\n", name, allowed)
		return nil, "", false
	}
	return flags, args[i], true
}
