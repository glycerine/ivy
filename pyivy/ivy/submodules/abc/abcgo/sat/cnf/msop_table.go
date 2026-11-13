package cnf

import "sync"

var msopOnce sync.Once
var msopCovers [][]byte

// Cnf_ReadMsops returns ABC's immutable minimum SOP covers indexed by the
// 16-bit truth table. Each cube is encoded in four base-3 digits.
func Cnf_ReadMsops() [][]byte {
	msopOnce.Do(func() {
		var alphabet [256]int16
		for i := range alphabet {
			alphabet[i] = -1
		}
		for i, ch := range s_Data3 {
			alphabet[ch] = int16(i)
		}
		msopCovers = make([][]byte, 1<<16)
		index := 1
		var cube []byte
	rows:
		for _, row := range s_Data4 {
			for i := 0; i < len(row); i++ {
				ch := row[i]
				if ch == ' ' {
					if index >= len(msopCovers) {
						break rows
					}
					msopCovers[index] = cube
					index++
					cube = nil
					continue
				}
				value := alphabet[ch]
				if value < 0 {
					panic("cnf: invalid MSOP encoding")
				}
				cube = append(cube, byte(value))
			}
		}
		if index != len(msopCovers) || len(cube) != 0 {
			panic("cnf: incomplete MSOP table")
		}
	})
	return msopCovers
}
