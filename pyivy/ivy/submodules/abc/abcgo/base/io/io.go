// Package io implements ABC network input and output commands and formats.
package io

import (
	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// Io_Init registers ABC's input and output commands on the supplied frame.
func Io_Init(frame *abcmain.Frame) {
	commands := []struct {
		name    string
		handler func(*abcmain.Frame, []string) int32
		changes bool
	}{
		{"read", IoCommandRead, true}, {"read_aiger", IoCommandReadAiger, true}, {"read_baf", IoCommandReadBaf, true},
		{"read_bblif", IoCommandReadBblif, true}, {"read_blif", IoCommandReadBlif, true}, {"read_blif_mv", IoCommandReadBlifMv, true},
		{"read_bench", IoCommandReadBench, true}, {"read_dsd", IoCommandReadDsd, true}, {"read_formula", IoCommandReadDsd, true},
		{"read_eqn", IoCommandReadEqn, true}, {"read_fins", IoCommandReadFins, false}, {"read_init", IoCommandReadInit, true},
		{"read_pla", IoCommandReadPla, true}, {"read_plamo", IoCommandReadPlaMo, true}, {"read_truth", IoCommandReadTruth, true},
		{"read_verilog", IoCommandReadVerilog, true}, {"read_status", IoCommandReadStatus, false}, {"&read_gig", IoCommandReadGig, false},
		{"read_json", IoCommandReadJson, false}, {"read_sf", IoCommandReadSF, false},
		{"write", IoCommandWrite, false}, {"write_hie", IoCommandWriteHie, false}, {"write_aiger", IoCommandWriteAiger, false},
		{"write_aiger_cex", IoCommandWriteAigerCex, false}, {"write_baf", IoCommandWriteBaf, false}, {"write_bblif", IoCommandWriteBblif, false},
		{"write_blif", IoCommandWriteBlif, false}, {"write_blif_mv", IoCommandWriteBlifMv, false}, {"write_bench", IoCommandWriteBench, false},
		{"write_book", IoCommandWriteBook, false}, {"write_cellnet", IoCommandWriteCellNet, false}, {"write_cex", IoCommandWriteCex, false},
		{"write_cnf", IoCommandWriteCnf, false}, {"&write_cnf", IoCommandWriteCnf2, false}, {"write_dot", IoCommandWriteDot, false},
		{"write_eqn", IoCommandWriteEqn, false}, {"write_gml", IoCommandWriteGml, false}, {"write_pla", IoCommandWritePla, false},
		{"write_verilog", IoCommandWriteVerilog, false}, {"write_sorter_cnf", IoCommandWriteSortCnf, false}, {"write_truth", IoCommandWriteTruth, false},
		{"&write_truths", IoCommandWriteTruths, false}, {"write_status", IoCommandWriteStatus, false}, {"write_smv", IoCommandWriteSmv, false}, {"write_json", IoCommandWriteJson, false},
	}
	for _, command := range commands {
		if err := frame.RegisterCommand(abcmain.Command{Name: command.name, Group: "I/O", Handler: command.handler, ChangesNetwork: command.changes}); err != nil {
			panic(err)
		}
	}
}

func IoCommandDispatchById(id cmd.Cmd_CommandId_t, frame *abcmain.Frame, args []string, result *int32) int32 {
	switch id {
	case cmd.CMD_COMMAND_READ_BLIF:
		*result = IoCommandReadBlif(frame, args)
	case cmd.CMD_COMMAND_READ_AIGER:
		*result = IoCommandReadAiger(frame, args)
	case cmd.CMD_COMMAND_WRITE_AIGER_CEX:
		*result = IoCommandWriteAigerCex(frame, args)
	default:
		return 0
	}
	return 1
}

func Io_End(*abcmain.Frame) {}
