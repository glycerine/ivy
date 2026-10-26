package abc

// CRC lookup data from src/misc/bzlib/crctable.c.

var bzip2CRC32Table = makeBZip2CRC32Table()

func makeBZip2CRC32Table() [256]uint32 {
	var table [256]uint32
	for index := range table {
		value := uint32(index) << 24
		for bit := 0; bit < 8; bit++ {
			if value&0x80000000 != 0 {
				value = value<<1 ^ 0x04c11db7
			} else {
				value <<= 1
			}
		}
		table[index] = value
	}
	return table
}

// BZip2CRC32Table returns an owned copy of the complete native lookup table.
func BZip2CRC32Table() [256]uint32 { return bzip2CRC32Table }
