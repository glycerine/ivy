package abc

// Cover and cube formatting from src/misc/mvc/mvcPrint.c.

import (
	"strconv"
	"strings"
)

// BitText returns the native Mvc_CubePrint bit sequence without its newline.
func (cube *MvcCube) BitText() (string, error) {
	if err := cube.check(); err != nil {
		return "", err
	}
	var text strings.Builder
	text.Grow(cube.cover.bits)
	for bit := 0; bit < cube.cover.bits; bit++ {
		if cube.words[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 {
			text.WriteByte('1')
		} else {
			text.WriteByte('0')
		}
	}
	return text.String(), nil
}

// BinaryText returns the native private binary-cube rendering without its
// newline. An odd final bit is ignored, and a 00 pair terminates iteration as
// it does in Mvc_CubeForEachVarValue.
func (cube *MvcCube) BinaryText() (string, error) {
	if err := cube.check(); err != nil {
		return "", err
	}
	var text strings.Builder
	text.Grow(cube.cover.bits / 2)
	for variable := 0; variable < cube.cover.bits/2; variable++ {
		bit := variable * 2
		value := (cube.words[bit>>5] >> uint(bit&31)) & 3
		if value == 0 {
			break
		}
		switch value {
		case 1:
			text.WriteByte('0')
		case 2:
			text.WriteByte('1')
		default:
			text.WriteByte('-')
		}
	}
	return text.String(), nil
}

func (cover *MvcCover) printText(binary bool) (string, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return "", err
	}
	var text strings.Builder
	text.WriteString("The cover contains ")
	text.WriteString(strconv.Itoa(len(cover.cubes)))
	text.WriteString(" cubes (")
	text.WriteString(strconv.Itoa(cover.bits))
	text.WriteString(" bits and ")
	text.WriteString(strconv.Itoa(cover.words))
	text.WriteString(" words)\n")
	for _, cube := range cover.cubes {
		var line string
		var err error
		if binary {
			line, err = cube.BinaryText()
		} else {
			line, err = cube.BitText()
		}
		if err != nil {
			return "", err
		}
		text.WriteString(line)
		text.WriteByte('\n')
	}
	if cover.literalCounts != nil {
		if len(cover.literalCounts) != cover.bits {
			return "", ErrInvariant
		}
		for _, count := range cover.literalCounts {
			text.WriteByte(' ')
			text.WriteString(strconv.Itoa(count))
		}
		text.WriteByte('\n')
	}
	text.WriteString("End of cover printout\n")
	return text.String(), nil
}

// PrintText returns the exact text emitted by native Mvc_CoverPrint.
func (cover *MvcCover) PrintText() (string, error) {
	return cover.printText(false)
}

// PrintBinaryText returns the exact text emitted by native
// Mvc_CoverPrintBinary.
func (cover *MvcCover) PrintBinaryText() (string, error) {
	return cover.printText(true)
}
