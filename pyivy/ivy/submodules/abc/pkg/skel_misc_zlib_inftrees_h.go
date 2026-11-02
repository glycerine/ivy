package abc

// Hand-written skeleton of the types in src/misc/zlib/inftrees.h.

// Code is `code`: structure for decoding tables.
type Code struct {
	op   uint8  // unsigned char : operation, extra bits, table bits
	bits uint8  // unsigned char : bits in this part of the code
	val  uint16 // unsigned short : offset in table or code value
}

// CodeType is the codetype enum: type of code to build for inflate_table().
type CodeType int

const (
	CODES CodeType = iota
	LENS
	DISTS
)
