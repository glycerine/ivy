package abc

// Hand-checked skeleton of src/misc/zlib/inftrees.c.

// inflate_table (ZLIB_INTERNAL; table is code FAR * FAR * (an in/out cursor into the codes array) and bits is an in/out scalar)
// C: int ZLIB_INTERNAL inflate_table(codetype type, unsigned short FAR *lens, unsigned codes, code FAR * FAR *table, unsigned FAR *bits, unsigned short FAR *work)
func Inflate_table(type_ CodeType, lens []uint16, codes uint32, table *[]Code, bits *uint32, work []uint16) int {
	panic("TODO, stubbed")
}
