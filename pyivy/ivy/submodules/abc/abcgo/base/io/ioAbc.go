// Public file kinds and ABC network types shared by native readers and writers.
package io

import "github.com/glycerine/abcgo/base/abc"

type Abc_Ntk_t = abc.Network
type Abc_Obj_t = abc.Object
type Io_FileType_t int32

const (
	IO_FILE_NONE Io_FileType_t = iota
	IO_FILE_AIGER
	IO_FILE_BAF
	IO_FILE_BBLIF
	IO_FILE_BLIF
	IO_FILE_BLIFMV
	IO_FILE_BENCH
	IO_FILE_BOOK
	IO_FILE_CNF
	IO_FILE_DOT
	IO_FILE_EDIF
	IO_FILE_EQN
	IO_FILE_GML
	IO_FILE_JSON
	IO_FILE_LIST
	IO_FILE_PLA
	IO_FILE_MOPLA
	IO_FILE_SMV
	IO_FILE_VERILOG
	IO_FILE_UNKNOWN
)
const IO_WRITE_LINE_LENGTH = 78
