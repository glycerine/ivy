// Always and initial statements from verCore.c.
package ver

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"strings"
)

func Ver_ParseAlways(p *Ver_Man_t, n *abc.Network) int32 {
	word := Ver_ParseGetName(p)
	if word == nil {
		return 0
	}
	if strings.HasPrefix(*word, "@") {
		Ver_StreamSkipToChars(p.pReader, ")")
		Ver_StreamPopChar(p.pReader)
		word = Ver_ParseGetName(p)
		if word == nil {
			return 0
		}
	}
	one := *word != "begin"
	for {
		if !one {
			word = Ver_ParseGetName(p)
			if word == nil {
				return 0
			}
			if *word == "end" {
				break
			}
		}
		output := Ver_ParseFindNet(n, *word)
		if output == nil {
			return verRangeError(p, fmt.Sprintf("Cannot read the always statement for %s (output wire is not defined).", *word))
		}
		sym := Ver_StreamPopChar(p.pReader)
		if sym != '<' && sym != '=' {
			return verRangeError(p, fmt.Sprintf("Cannot read the assign statement for %s (expected <= or =).", *word))
		}
		if sym == '<' {
			Ver_StreamPopChar(p.pReader)
		}
		if Ver_ParseSkipComments(p) == 0 {
			return 0
		}
		inputName := Ver_ParseGetName(p)
		if inputName == nil {
			return 0
		}
		var input *abc.Object
		if strings.HasPrefix(*inputName, "~") {
			input = Ver_ParseCreateInv(n, Ver_ParseFindNet(n, (*inputName)[1:]))
		} else {
			input = Ver_ParseFindNet(n, *inputName)
		}
		if input == nil {
			return verRangeError(p, fmt.Sprintf("Cannot read the always statement for %s (input wire is not defined).", *inputName))
		}
		Ver_ParseCreateLatch(n, input, output)
		if Ver_StreamPopChar(p.pReader) != ';' {
			panic("ver: always statement missing semicolon")
		}
		if one {
			break
		}
	}
	return 1
}

func Ver_ParseInitial(p *Ver_Man_t, n *abc.Network) int32 {
	word := Ver_ParseGetName(p)
	if word == nil {
		return 0
	}
	one := *word != "begin"
	for {
		if !one {
			word = Ver_ParseGetName(p)
			if word == nil {
				return 0
			}
			if *word == "end" {
				break
			}
		}
		net := Ver_ParseFindNet(n, *word)
		if net == nil {
			return verRangeError(p, fmt.Sprintf("Cannot read the initial statement for %s (output wire is not defined).", *word))
		}
		sym := Ver_StreamPopChar(p.pReader)
		if sym != '<' && sym != '=' {
			return verRangeError(p, fmt.Sprintf("Cannot read the assign statement for %s (expected <= or =).", *word))
		}
		if sym == '<' {
			Ver_StreamPopChar(p.pReader)
		}
		if Ver_ParseSkipComments(p) == 0 {
			return 0
		}
		equation := Ver_StreamGetWord(p.pReader, ";")
		if equation == nil {
			return 0
		}
		if abc.Abc_ObjFaninNum(net) == 0 {
			return verRangeError(p, "Cannot find the latch to assign the initial value.")
		}
		latch := abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(net))
		if abc.Abc_ObjIsLatch(latch) == 0 {
			panic("ver: initial target is not a latch")
		}
		switch *equation {
		case "0", "1'b0":
			abc.Abc_LatchSetInit0(latch)
		case "1", "1'b1":
			abc.Abc_LatchSetInit1(latch)
		default:
			return verRangeError(p, fmt.Sprintf("Incorrect initial value of the latch %s.", abc.Abc_ObjName(net)))
		}
		if Ver_StreamPopChar(p.pReader) != ';' {
			panic("ver: initial statement missing semicolon")
		}
		if one {
			break
		}
	}
	return 1
}
