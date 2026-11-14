// Verilog whitespace, comment and identifier parsing from verParse.c
// by Alan Mishchenko (UC Berkeley, 2006).
package ver

func Ver_ParseSkipComments(pMan *Ver_Man_t) int32 {
	p := pMan.pReader
	Ver_StreamSkipChars(p, " \t\n\r")
	if Ver_StreamIsOkey(p) == 0 {
		return 1
	}
	symbol := Ver_StreamScanChar(p)
	if symbol != '/' {
		return 1
	}
	Ver_StreamPopChar(p)
	symbol = Ver_StreamScanChar(p)
	if symbol == '/' {
		Ver_StreamSkipToChars(p, "\n")
		return Ver_ParseSkipComments(pMan)
	}
	if symbol == '*' {
		Ver_StreamPopChar(p)
		for {
			Ver_StreamSkipToChars(p, "*")
			Ver_StreamPopChar(p)
			if Ver_StreamScanChar(p) == '/' {
				break
			}
		}
		Ver_StreamPopChar(p)
		return Ver_ParseSkipComments(pMan)
	}
	copy(pMan.sError[:], "Cannot parse after symbol \"/\".\x00")
	Ver_ParsePrintErrorMessage(pMan)
	return 0
}

func Ver_ParseGetName(pMan *Ver_Man_t) *string {
	p := pMan.pReader
	pMan.fNameLast = 0
	if Ver_StreamIsOkey(p) == 0 || Ver_ParseSkipComments(pMan) == 0 {
		return nil
	}
	var word *string
	if Ver_StreamScanChar(p) == '\\' {
		pMan.fNameLast = 1
		Ver_StreamPopChar(p)
		word = Ver_StreamGetWord(p, " \r\n")
		Ver_StreamSkipChars(p, " \r\n")
		if Ver_StreamScanChar(p) == '[' {
			for {
				c := Ver_StreamPopChar(p)
				*word += string(c)
				if c == ']' {
					break
				}
			}
		}
	} else {
		word = Ver_StreamGetWord(p, " \t\n\r(),;")
	}
	if Ver_StreamIsOkey(p) != 0 && Ver_ParseSkipComments(pMan) == 0 {
		return nil
	}
	return word
}
