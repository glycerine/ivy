// Signal range table from verCore.c.
package ver

import "github.com/glycerine/abcgo/misc/st"

func Ver_ParseLookupSuffix(p *Ver_Man_t, name string, msb, lsb *int32) int32 {
	*msb, *lsb = -1, -1
	if p.tName2Suffix != nil {
		if value, ok := p.tName2Suffix.Lookup(name); ok {
			packed := value.(uint32)
			*msb = int32((packed >> 8) & 0xff)
			*lsb = int32(packed & 0xff)
		}
	}
	return 1
}
func Ver_ParseInsertsSuffix(p *Ver_Man_t, name string, msb, lsb int32) int32 {
	if p.tName2Suffix == nil {
		p.tName2Suffix = st.New(st.StringCompare, st.StringHash)
	}
	if _, ok := p.tName2Suffix.Lookup(name); ok {
		return 1
	}
	if msb < 0 || msb >= 128 || lsb < 0 || lsb >= 128 {
		panic("ver: signal range out of bounds")
	}
	p.tName2Suffix.Insert(name, uint32(msb<<8|lsb))
	return 1
}
func Ver_ParseRemoveSuffixTable(p *Ver_Man_t) {
	if p.tName2Suffix != nil {
		p.tName2Suffix.Clear()
		p.tName2Suffix = nil
	}
}
