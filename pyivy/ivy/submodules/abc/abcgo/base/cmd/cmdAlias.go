package cmd

import (
	"fmt"
	"strings"
)

func CmdCommandAliasAdd(f *Abc_Frame_t, name string, count int32, args []string) {
	if count < 0 || int(count) > len(args) {
		panic("invalid alias argc")
	}
	f.RegisterAlias(name, args[:count])
}
func CmdCommandAliasPrint(f *Abc_Frame_t, alias *Abc_Alias) {
	fmt.Fprintf(f.Out, "%-15s %s\n", alias.Name, strings.Join(alias.Args, " "))
}
func CmdCommandAliasLookup(f *Abc_Frame_t, name string) string {
	if args, ok := f.LookupAlias(name); ok && len(args) > 0 {
		return args[0]
	}
	return name
}
func CmdCommandAliasFree(alias *Abc_Alias) { alias.Args = nil; alias.Name = "" }
