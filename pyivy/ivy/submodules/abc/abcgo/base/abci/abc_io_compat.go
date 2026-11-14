package abci

import (
	baseabc "github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
)

// Keep the translated ABC reader entry points available in abci while using
// the native format dispatch and logic conversion implemented by base/io.
type Io_FileType_t = abcio.Io_FileType_t

const (
	IO_FILE_NONE    = abcio.IO_FILE_NONE
	IO_FILE_AIGER   = abcio.IO_FILE_AIGER
	IO_FILE_BAF     = abcio.IO_FILE_BAF
	IO_FILE_BBLIF   = abcio.IO_FILE_BBLIF
	IO_FILE_BLIF    = abcio.IO_FILE_BLIF
	IO_FILE_BLIFMV  = abcio.IO_FILE_BLIFMV
	IO_FILE_BENCH   = abcio.IO_FILE_BENCH
	IO_FILE_BOOK    = abcio.IO_FILE_BOOK
	IO_FILE_CNF     = abcio.IO_FILE_CNF
	IO_FILE_DOT     = abcio.IO_FILE_DOT
	IO_FILE_EDIF    = abcio.IO_FILE_EDIF
	IO_FILE_EQN     = abcio.IO_FILE_EQN
	IO_FILE_GML     = abcio.IO_FILE_GML
	IO_FILE_JSON    = abcio.IO_FILE_JSON
	IO_FILE_LIST    = abcio.IO_FILE_LIST
	IO_FILE_PLA     = abcio.IO_FILE_PLA
	IO_FILE_MOPLA   = abcio.IO_FILE_MOPLA
	IO_FILE_SMV     = abcio.IO_FILE_SMV
	IO_FILE_VERILOG = abcio.IO_FILE_VERILOG
	IO_FILE_UNKNOWN = abcio.IO_FILE_UNKNOWN
)

func Io_ReadFileType(filename string) Io_FileType_t {
	return abcio.Io_ReadFileType(filename)
}

func Io_Read(filename string, kind Io_FileType_t, check, barBuffers int32) *baseabc.Network {
	return abcio.Io_Read(filename, kind, check, barBuffers)
}
