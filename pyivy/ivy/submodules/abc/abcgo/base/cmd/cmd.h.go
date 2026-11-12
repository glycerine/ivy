// Package cmd implements ABC command parsing, aliases and interactive commands.
// The registry, flags, history and network lifecycle belong to base/main.Frame.
package cmd

import abcmain "github.com/glycerine/abcgo/base/main"

type Abc_Frame_t = abcmain.Frame
type Abc_Command = abcmain.Command
type Cmd_CommandFuncType func(*Abc_Frame_t, int32, []string) int32
type Cmd_CommandId_t int32
type Cmd_CommandId_t_ = Cmd_CommandId_t

const (
	CMD_COMMAND_UNKNOWN Cmd_CommandId_t = iota
	CMD_COMMAND_ALIAS
	CMD_COMMAND_SOURCE
	CMD_COMMAND_SET
	CMD_COMMAND_READ_BLIF
	CMD_COMMAND_READ_AIGER
	CMD_COMMAND_STRASH
	CMD_COMMAND_PRINT_STATS
	CMD_COMMAND_PDR
	CMD_COMMAND_WRITE_AIGER_CEX
)

type Abc_Alias struct {
	Name string
	Args []string
}
