// Continuous assignments from verCore.c.
package ver

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

func Ver_ParseAssign(p *Ver_Man_t, n *abc.Network) int32 {
	if Ver_ParseConvertNetwork(p, n, p.fMapped) == 0 {
		return 0
	}
	for {
		token := Ver_ParseGetName(p)
		if token == nil {
			return 0
		}
		word := *token
		if word == "#1" {
			continue
		}
		var msb, lsb int32
		Ver_ParseLookupSuffix(p, word, &msb, &lsb)
		limit := msb - lsb
		if limit < 0 {
			limit = -limit
		}
		limit++
		if msb >= 0 && lsb >= 0 && limit > 1 {
			output := word
			if output == "1'h0" {
				output = "1'b0"
			} else if output == "1'h1" {
				output = "1'b1"
			}
			if Ver_StreamPopChar(p.pReader) != '=' {
				return verRangeError(p, fmt.Sprintf("Cannot read the assign statement for %s (expected equality sign).", word))
			}
			token = Ver_ParseGetName(p)
			if token == nil {
				return 0
			}
			word = *token
			if len(word) == 0 || word[0] < '0' || word[0] > '9' {
				return verRangeError(p, fmt.Sprintf("Currently can only assign vector-signal \"%s\" to be a constant.", output))
			}
			if Ver_ParseConstant(p, word) == 0 {
				return 0
			}
			if vec.Vec_PtrSize(p.vNames) != int(limit) {
				return verRangeError(p, fmt.Sprintf("The constant size (%d) is different from the signal\"%s\" size (%d).", vec.Vec_PtrSize(p.vNames), output, limit))
			}
			bit := lsb
			for i := int32(0); i < limit; i++ {
				constant := "1'b0"
				if vec.Vec_PtrEntry(p.vNames, int(limit-1-i)) != nil {
					constant = "1'b1"
				}
				input := Ver_ParseFindNet(n, constant)
				if input == nil {
					panic("ver: constant net missing")
				}
				node := abc.Abc_NtkCreateNodeBuf(n, input)
				target := Ver_ParseFindNet(n, fmt.Sprintf("%s[%d]", output, bit))
				if target == nil {
					return verRangeError(p, fmt.Sprintf("Cannot read the assign statement for %s (output wire is not defined).", word))
				}
				abc.Abc_ObjAddFanin(target, node)
				if msb > lsb {
					bit++
				} else {
					bit--
				}
			}
			Ver_ParseSkipComments(p)
		} else {
			reduction := strings.HasPrefix(word, "{") && p.fNameLast == 0
			if reduction {
				if len(word) < 2 {
					panic("ver: invalid reduction output")
				}
				word = word[1 : len(word)-1]
				if strings.HasPrefix(word, "\\") {
					panic("ver: escaped reduction output")
				}
			}
			output := Ver_ParseFindNet(n, word)
			if output == nil {
				return verRangeError(p, fmt.Sprintf("Cannot read the assign statement for %s (output wire is not defined).", word))
			}
			if Ver_StreamPopChar(p.pReader) != '=' {
				return verRangeError(p, fmt.Sprintf("Cannot read the assign statement for %s (expected equality sign).", word))
			}
			if Ver_ParseSkipComments(p) == 0 {
				return 0
			}
			stops := ",;"
			if reduction {
				stops = ";"
			}
			equation := Ver_StreamGetWord(p.pReader, stops)
			if equation == nil {
				return verRangeError(p, fmt.Sprintf("Cannot read the equation for %s.", abc.Abc_ObjName(output)))
			}
			vec.Vec_PtrClear(p.vNames)
			var function any
			if p.fMapped != 0 {
				lib, _ := abcmain.Abc_FrameReadLibGen().(*mio.Mio_Library_t)
				if *equation == "1'b0" {
					function = mio.Mio_LibraryReadConst0(lib)
				} else if *equation == "1'b1" {
					function = mio.Mio_LibraryReadConst1(lib)
				} else {
					name := *equation
					if strings.HasPrefix(name, "\\") {
						name = name[1 : len(name)-1]
					}
					if Ver_ParseFindNet(n, name) == nil {
						return verRangeError(p, "Cannot read Verilog with non-trivial assignments in the mapped netlist.")
					}
					vec.Vec_PtrPush(p.vNames, len(name))
					vec.Vec_PtrPush(p.vNames, name)
					function = mio.Mio_LibraryReadBuf(lib)
					if function.(*mio.Mio_Gate_t) == nil {
						return verRangeError(p, fmt.Sprintf("Reading assign statement for node %s has failed because the genlib library has no buffer.", abc.Abc_ObjName(output)))
					}
				}
			} else {
				manager := n.FunctionManager.(*hop.Hop_Man_t)
				var h *hop.Hop_Obj_t
				switch *equation {
				case "0", "1'b0", "1'bx":
					h = hop.Hop_ManConst0(manager)
				case "1", "1'b1":
					h = hop.Hop_ManConst1(manager)
				default:
					if reduction {
						h = Ver_FormulaReduction(*equation, manager, p.vNames, p.sError[:])
					} else {
						h = Ver_FormulaParser(*equation, manager, p.vNames, p.vStackFn, p.vStackOp, p.sError[:])
					}
				}
				if h == nil {
					Ver_ParsePrintErrorMessage(p)
					return 0
				}
				function = h
			}
			node := abc.Abc_NtkCreateNode(n)
			node.Data = function
			abc.Abc_ObjAddFanin(output, node)
			for i := 0; i < vec.Vec_PtrSize(p.vNames)/2; i++ {
				name := vec.Vec_PtrEntry(p.vNames, 2*i+1).(string)
				lookup := name
				if name == "1'h0" {
					lookup = "1'b0"
				} else if name == "1'h1" {
					lookup = "1'b1"
				}
				input := Ver_ParseFindNet(n, lookup)
				if input == nil {
					return verRangeError(p, fmt.Sprintf("Cannot read the assign statement for %s (input wire %s is not defined).", word, name))
				}
				abc.Abc_ObjAddFanin(node, input)
			}
		}
		symbol := Ver_StreamPopChar(p.pReader)
		if symbol == ',' {
			continue
		}
		if symbol == ';' {
			return 1
		}
	}
}
