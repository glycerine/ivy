// Signal declarations from verCore.c.
package ver

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
)

func Ver_ParseSignal(p *Ver_Man_t, n *abc.Network, kind Ver_SignalType_t) int32 {
	msb, lsb := int32(-1), int32(-1)
	for {
		ptr := Ver_ParseGetName(p)
		if ptr == nil {
			return 0
		}
		word := *ptr
		if word == "wire" {
			continue
		}
		if len(word) > 0 && word[0] == '[' && p.fNameLast == 0 {
			if msb != -1 || lsb != -1 {
				panic("ver: duplicate range")
			}
			Ver_ParseSignalPrefix(p, &word, &msb, &lsb)
			if word == "" {
				ptr = Ver_ParseGetName(p)
				if ptr == nil {
					return 0
				}
				word = *ptr
			}
		}
		create := func(name string) {
			if kind == VER_SIG_INPUT || kind == VER_SIG_INOUT {
				Ver_ParseCreatePi(n, name)
			}
			if kind == VER_SIG_OUTPUT || kind == VER_SIG_INOUT {
				Ver_ParseCreatePo(n, name)
			}
			if kind == VER_SIG_WIRE || kind == VER_SIG_REG {
				abc.Abc_NtkFindOrCreateNet(n, name)
			}
		}
		if msb == -1 && lsb == -1 {
			create(word)
		} else {
			if msb < 0 || lsb < 0 {
				panic("ver: negative signal range")
			}
			Ver_ParseInsertsSuffix(p, word, msb, lsb)
			count := msb - lsb
			if count < 0 {
				count = -count
			}
			count++
			bit := lsb
			for i := int32(0); i < count; i++ {
				name := word
				if count > 1 {
					name = fmt.Sprintf("%s[%d]", word, bit)
				}
				create(name)
				if msb > lsb {
					bit++
				} else {
					bit--
				}
			}
		}
		symbol := Ver_StreamPopChar(p.pReader)
		if symbol == ',' {
			continue
		}
		if symbol == ';' {
			return 1
		}
		break
	}
	return verRangeError(p, "Cannot parse signal line (expected , or ;).")
}
