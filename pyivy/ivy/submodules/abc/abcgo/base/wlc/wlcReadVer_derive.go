package wlc

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func wlcInitValue(nameID, width int32) int32 {
	if nameID > 0 {
		return nameID
	}
	return -width
}

func wlcReadWriteType(read int32) int32 {
	if read != 0 {
		return int32(WLC_OBJ_READ)
	}
	return int32(WLC_OBJ_WRITE)
}

func wlcRotationType(right int32) int32 {
	if right != 0 {
		return int32(WLC_OBJ_ROTATE_R)
	}
	return int32(WLC_OBJ_ROTATE_L)
}

func verFallbackName(name string) string {
	if name == "" {
		return `"?"`
	}
	return name
}

// wlcVerChar models C reads of the NUL terminator when a suffix is exhausted.
func wlcVerChar(s string, i int) byte {
	if i < 0 || i >= len(s) {
		return 0
	}
	return s[i]
}

func Wlc_PrsReadDeclaration(p *Wlc_Prs_t, pStart string) int32 {
	var fFound int32
	var Type int32 = int32(WLC_OBJ_NONE)
	var iObj int32
	var pLine string
	var Signed, Beg, End, NameId, fIsPo int32
	if Wlc_PrsStrCmp(pStart, "input") != 0 {
		pStart = pStart[len("input"):]
		Type = int32(WLC_OBJ_PI)
	} else if Wlc_PrsStrCmp(pStart, "output") != 0 {
		pStart = pStart[len("output"):]
		fIsPo = 1
	}
	pStart = Wlc_PrsSkipSpaces(pStart)
	if Wlc_PrsStrCmp(pStart, "wire") != 0 {
		pStart = pStart[len("wire"):]
	} else if Wlc_PrsStrCmp(pStart, "reg") != 0 {
		pStart = pStart[len("reg"):]
	}
	// read 'signed'
	pStart = Wlc_PrsFindWord(pStart, "signed", &Signed)
	// read range
	pLine = pStart
	pStart = Wlc_PrsFindRange(pStart, &End, &Beg)
	if pStart == "" {
		return Wlc_PrsWriteErrorMessage(p, pLine, "Non-standard range.")
	}
	if End != 0 && Beg != 0 {
		if p.nNonZero[0] == 0 {
			p.nNonZero[1] = End
			p.nNonZero[2] = Beg
			p.nNonZero[3] = Wlc_PrsFindLine(p, pStart)
		}
		p.nNonZero[0]++
	}
	if End < 0 || Beg < 0 {
		if p.nNegative[0] == 0 {
			p.nNegative[1] = End
			p.nNegative[2] = Beg
			p.nNegative[3] = Wlc_PrsFindLine(p, pStart)
		}
		p.nNegative[0]++
	}
	if End < Beg {
		if p.nReverse[0] == 0 {
			p.nReverse[1] = End
			p.nReverse[2] = Beg
			p.nReverse[3] = Wlc_PrsFindLine(p, pStart)
		}
		p.nReverse[0]++
	}
	for {
		var pName string
		var XValue, TypeNew int32
		// read name
		pStart = Wlc_PrsFindName(pStart, &pName)
		if pStart == "" {
			return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name in declaration.")
		}
		NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
		if fFound != 0 {
			return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is declared more than once.", pName)
		}
		iObj = Wlc_ObjAlloc(p.pNtk, Type, Signed, End, Beg)
		if fIsPo != 0 {
			Wlc_ObjSetCo(p.pNtk, Wlc_NtkObj(p.pNtk, iObj), 0)
		}
		assert(iObj == NameId, "iObj == NameId")
		// check next definition
		pStart = Wlc_PrsSkipSpaces(pStart)
		if len(pStart) > 0 && wlcVerChar(pStart, 0) == ',' {
			pStart = pStart[1:]
			continue
		}
		// check definition
		TypeNew = Wlc_PrsFindDefinition(p, pStart, p.vFanins, &XValue)
		if TypeNew != 0 {
			var pObj *Wlc_Obj_t = Wlc_NtkObj(p.pNtk, iObj)
			Wlc_ObjUpdateType(p.pNtk, pObj, TypeNew)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			pObj.fXConst = uint32(XValue)
		}
		break
	}
	return 1
}
func Wlc_PrsDerive(p *Wlc_Prs_t) int32 {
	var pObj *Wlc_Obj_t
	var pStart, pName string
	var i int32
	// go through the directives
	// inlining of macro Wlc_PrsForEachLine
	for i = int32(0); i < Vec_IntSize(p.vStarts); i++ {
		pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i))
		if pStart == "" {
			continue
		}
		start := wlcPrsCursorAt(p, int(Vec_IntEntry(p.vStarts, i)))
	startword:
		module, isModule := start.findWord("module")
		if isModule {
			// get module name
			nameEnd, parsedName := module.advance(len("module")).findName()
			pName = parsedName
			spStatic_Wlc_PrsStrtok_last = nameEnd.rest()
			if pName == "" {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read model name.")
			}
			// THIS IS A HACK to skip definitions of modules beginning with "CPL_"
			if strings.HasPrefix(pName, "CPL_") {
				for i++; i < Vec_IntSize(p.vStarts); i++ {
					start = wlcPrsCursorAt(p, int(Vec_IntEntry(p.vStarts, i)))
					if end, found := start.findSubstring("endmodule"); found {
						start = end
						pStart = start.rest()
						break
					}
				}
				continue
			}
			if strings.HasPrefix(pName, "table") {
				// THIS IS A HACK to detect table module descriptions
				var Width1 int32 = -1
				var Width2 int32 = -1
				var v, b, Value, nBits, nInts int32
				var pTable []uint32
				var vValues *Vec_Int_t = Vec_IntAlloc(256)
				// inlining of macro Wlc_PrsForEachLineStart
				for i = i + 1; i < Vec_IntSize((p).vStarts); i++ {
					start = wlcPrsCursorAt(p, int(Vec_IntEntry(p.vStarts, i)))
					pStart = start.rest()
					if endcase, found := start.findWord("endcase"); found {
						pStart = endcase.rest()
						break
					}
					quote := start.findSymbol('\'')
					if !quote.valid() {
						continue
					}
					Width1 = atoi(quote.retreat(1).rest())
					quote = quote.advance(2).findSymbol('\'')
					if !quote.valid() {
						continue
					}
					Width2 = atoi(quote.retreat(1).rest())
					Value = 0
					var parsed [1]word
					util.Abc_TtReadHexNumber(parsed[:], quote.advance(2).rest())
					Value = int32(parsed[0])
					Vec_IntPush(vValues, Value)
				}
				//Vec_IntPrint( vValues );
				nBits = Abc_Base2Log(Vec_IntSize(vValues))
				if Vec_IntSize(vValues) != (1 << nBits) {
					Vec_IntFree(vValues)
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read module \"%s\".", pName)
				}
				assert(Width1 == nBits, "Width1 == nBits")
				// create bitmap
				nInts = util.Abc_BitWordNum(Width2 * Vec_IntSize(vValues))
				tableBytes := mem.Mem_FlexEntryFetch(p.pMemTable, nInts*4)
				pTable = unsafe.Slice((*uint32)(unsafe.Pointer(&tableBytes[0])), int(nInts))
				clear(pTable)
				for v = 0; v < Vec_IntSize(vValues); v++ {
					Value = Vec_IntEntry(vValues, v)
					for b = 0; b < Width2; b++ {
						if (Value>>b)&1 != 0 {
							Abc_InfoSetBit(pTable, v*Width2+b)
						}
					}
				}
				Vec_PtrPush(p.vTables, pTable)
				Vec_IntFree(vValues)
				continue
			}
			if p.pNtk != nil {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Network is already defined.")
			}
			p.pNtk = Wlc_NtkAlloc(pName, Vec_IntSize(p.vStarts))
			p.pNtk.pManName = Abc_NamStart(Vec_IntSize(p.vStarts), 20)
			p.pNtk.pMemTable = p.pMemTable
			p.pMemTable = nil
			p.pNtk.vTables = p.vTables
			p.vTables = nil
			// read the argument definitions
			for pName = Wlc_PrsStrtok("", "(,)"); pName != ""; pName = Wlc_PrsStrtok("", "(,)") {
				pName = Wlc_PrsSkipSpaces(pName)
				if Wlc_PrsStrCmp(pName, "input") == 0 || Wlc_PrsStrCmp(pName, "output") == 0 || Wlc_PrsStrCmp(pName, "wire") == 0 {
					if Wlc_PrsReadDeclaration(p, pName) == 0 {
						return 0
					}
				}
			}

		} else if Wlc_PrsStrCmp(pStart, "endmodule") != 0 {
			var vTemp *Vec_Int_t = Vec_IntStartNatural(Wlc_NtkObjNumMax(p.pNtk))
			Vec_IntAppend(&p.pNtk.vNameIds, vTemp)
			Vec_IntFree(vTemp)
			if p.pNtk.vInits != nil {
				// move FO/FI to be part of CI/CO
				assert((Vec_IntSize(&p.pNtk.vFfs)&1) == 0, "(Vec_IntSize(&p.pNtk.vFfs) & 1) == 0")
				assert(Vec_IntSize(&p.pNtk.vFfs) == 2*Vec_IntSize(p.pNtk.vInits), "Vec_IntSize(&p.pNtk.vFfs) == 2 * Vec_IntSize(p.pNtk.vInits)")
				// inlining of macro Wlc_NtkForEachFf
				for i = 0; i < Vec_IntSize(&p.pNtk.vFfs); i++ {
					pObj = Wlc_NtkFf(p.pNtk, i)
					if i&1 != 0 {
						Wlc_ObjSetCo(p.pNtk, pObj, 1)
					} else {
						Wlc_ObjSetCi(p.pNtk, pObj)
					}
				}
				Vec_IntClear(&p.pNtk.vFfs)
				// convert init values into binary string
				//Vec_IntPrint( p->pNtk->vInits );
				p.pNtk.pInits = Wlc_PrsConvertInitValues(p.pNtk)
				//printf( "%s\n", p->pNtk->pInits );
			}
			if p.pNtk.vArsts != nil && p.pNtk.fAsyncRst == 0 {
				var i, NameIdArst int32
				for i = 0; i < Vec_IntSize(p.pNtk.vArsts); i++ {
					NameIdArst = Vec_IntEntry(p.pNtk.vArsts, i)
					if Wlc_PrsCheckBitConst0(p.pNtk, NameIdArst) != 0 {
						continue
					}
					p.pNtk.fAsyncRst = 1
					fmt.Printf("Detected async reset \"%s\".\n", Abc_NamStr(p.pNtk.pManName, NameIdArst))
					break
				}
			}
			if p.vPoPairs != nil {
				assert(vec.Vec_StrEntryLast(p.vPoPairs) == 0, "Vec_StrEntryLast(p.vPoPairs) == 0")
				Vec_StrPush(p.vPoPairs, 0)
				pairNames := Vec_StrArray(p.vPoPairs)
				for offset := 0; offset < len(pairNames) && pairNames[offset] != 0; {
					end := offset
					for end < len(pairNames) && pairNames[end] != 0 {
						end++
					}
					pName = string(pairNames[offset:end])
					// inlining of macro Wlc_NtkForEachPo
					for i = 0; i < Wlc_NtkPoNum(p.pNtk); i++ {
						pObj = Wlc_NtkPo(p.pNtk, i)
						if Wlc_ObjName(p.pNtk, Wlc_ObjId(p.pNtk, pObj)) == pName {
							Vec_IntPush(&p.pNtk.vPoPairs, i)
							break
						}
					}
					assert(i < Wlc_NtkPoNum(p.pNtk), "i < Wlc_NtkPoNum(p.pNtk)")
					offset = end + 1
				}
				assert(Vec_IntSize(&p.pNtk.vPoPairs)%2 == 0, "Vec_IntSize(&p.pNtk.vPoPairs) % 2 == 0")
				fmt.Printf("Finished parsing %d output pairs to be checked for equivalence.\n", Vec_IntSize(&p.pNtk.vPoPairs)/2)
			}
			break

			// these are read as part of the interface
		} else if (Wlc_PrsStrCmp(pStart, "input") != 0) || (Wlc_PrsStrCmp(pStart, "output") != 0) || (Wlc_PrsStrCmp(pStart, "wire") != 0) || (Wlc_PrsStrCmp(pStart, "reg") != 0) {
			if !(Wlc_PrsReadDeclaration(p, pStart) != 0) {
				return 0
			}

		} else if Wlc_PrsStrCmp(pStart, "assign") != 0 {
			var Type, NameId, fFound, XValue int32
			cStart := wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart))
			cStart = cStart.advance(len("assign"))
			// read name
			cStart, pName = cStart.findName()
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read name after assign.")
			}
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			if fFound == 0 {
				return Wlc_PrsWriteErrorMessage(p, cStart.rest(), "Name %s is not declared.", pName)
			}
			// read definition
			Type = Wlc_PrsFindDefinition(p, cStart.rest(), p.vFanins, &XValue)
			if Type != 0 {
				pObj = Wlc_NtkObj(p.pNtk, NameId)
				Wlc_ObjUpdateType(p.pNtk, pObj, Type)
				Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
				pObj.fXConst = uint32(XValue)

			} else {
				return 0
			}

		} else if Wlc_PrsStrCmp(pStart, "table") != 0 {
			// THIS IS A HACK to detect tables
			var NameId int32
			var fFound int32
			cStart := wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart))
			cIndex := cStart.advance(len("table"))
			indexText := cIndex.rest()
			indexEnd := 0
			for indexEnd < len(indexText) && Wlc_PrsIsDigit(indexText[indexEnd:]) != 0 {
				indexEnd++
			}
			var iTable int32 = atoi(indexText[:indexEnd])
			// find opening
			cStart = cStart.findSymbol('(')
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read table.")
			}
			// read input
			cStart, pName = cStart.advance(1).findName()
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read name after assign.")
			}
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			if fFound == 0 {
				return Wlc_PrsWriteErrorMessage(p, cStart.rest(), "Name %s is not declared.", pName)
			}
			// save inputs
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameId)
			Vec_IntPush(p.vFanins, iTable)
			// find comma
			cStart = cStart.findSymbol(',')
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read table.")
			}
			// read output
			cStart, pName = cStart.advance(1).findName()
			if !cStart.valid() {
				return Wlc_PrsWriteErrorMessage(p, "", "Cannot read name after assign.")
			}
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			if fFound == 0 {
				return Wlc_PrsWriteErrorMessage(p, cStart.rest(), "Name %s is not declared.", pName)
			}
			pObj = Wlc_NtkObj(p.pNtk, NameId)
			Wlc_ObjUpdateType(p.pNtk, pObj, int32(WLC_OBJ_TABLE))
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if Wlc_PrsStrCmp(pStart, "always") != 0 {
			// THIS IS A HACK to detect always statement representing combinational MUX
			var NameId int32
			var NameIdOut int32 = -1
			var nValues int32
			var fDefaultFound bool
			cur := wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart))
			// find control
			var found bool
			cur, found = cur.findWord("case")
			if !found {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read case statement.")
			}
			// read the name
			cur = cur.findSymbol('(')
			if !cur.valid() {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read table.")
			}
			cur = cur.advance(1).findSymbol('(')
			if !cur.valid() {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read table.")
			}
			cur, pName = cur.advance(1).findName()
			if !cur.valid() {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read name after case.")
			}
			fFound := int32(0)
			NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
			found = fFound != 0
			if !found {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Name %s is not declared.", pName)
			}
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameId)
			// read data inputs
			pObj = Wlc_NtkObj(p.pNtk, NameId)
			if pObj == nil {
				return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot find the object in case statement.")
			}
			// remember the number of values
			nValues = (1 << Wlc_ObjRange(pObj))
			for {
				// find opening
				cur = cur.findSymbol(':')
				if !cur.valid() {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot find colon in the case statement.")
				}
				// find output name
				cur, pName = cur.advance(1).findName()
				if !cur.valid() {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read name after case.")
				}
				fFound = 0
				NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				found = fFound != 0
				if !found {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Name %s is not declared.", pName)
				}
				// find equality
				cur = cur.findSymbol('=')
				if !cur.valid() {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot find equality in the case statement.")
				}
				// find input name
				pStart = cur.advance(1).skipSpaces().rest()
				pStart = Wlc_PrsReadName(p, pStart, p.vFanins)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside case statement.")
				}
				cur = cur.withStringSuffix(pStart)
				// consider default
				if fDefaultFound {
					EntryLast := Vec_IntEntryLast(p.vFanins)
					if nValues != Vec_IntSize(p.vFanins)-2 {
						Vec_IntFillExtra(p.vFanins, nValues+1, EntryLast)
					} else {
						Vec_IntPop(p.vFanins)
					}
					// get next line and check its opening character
					i++
					if i >= Vec_IntSize(p.vStarts) {
						return Wlc_PrsWriteErrorMessage(p, "", "Cannot read case statement.")
					}
					pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i))
					cur = wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart)).skipSpaces()

				} else {
					// get next line and check its opening character
					i++
					if i >= Vec_IntSize(p.vStarts) {
						return Wlc_PrsWriteErrorMessage(p, "", "Cannot read case statement.")
					}
					pStart = Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, i))
					cur = wlcPrsCursorAt(p, len(p.pBuffer)-len(pStart)).skipSpaces()
					if cur.byte() != 0 && Wlc_PrsIsDigit(cur.rest()) != 0 {
						continue
					}
					if !cur.hasPrefix("default") {
						fDefaultFound = true
						continue
					}
				}
				// find closing
				cur, found = cur.findWord("endcase")
				if !found {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read case statement.")
				}
				// find closing
				cur, found = cur.findWord("end")
				if !found {
					return Wlc_PrsWriteErrorMessage(p, cur.rest(), "Cannot read case statement.")
				}
				cur = cur.skipSpaces()
				break
			}
			// check range of the control
			if nValues < Vec_IntSize(p.vFanins)-1 { // may occur if default is not there

				//return Wlc_PrsWriteErrorMessage( p, pStart, "The number of values in the case statement is wrong.", pName );
				fmt.Printf("Warning:  The number of values in the case statement is wrong.\n")
				Vec_IntShrink(p.vFanins, nValues+1)

			} else if nValues > Vec_IntSize(p.vFanins)-1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "The number of values in the case statement is wrong.")
			}
			if Wlc_ObjRange(pObj) == 1 {
				//                return Wlc_PrsWriteErrorMessage( p, pStart, "Always-statement with 1-bit control is not bit-blasted correctly.", pName );
				fmt.Printf("Warning:  Case-statement with 1-bit control is treated as a 2:1 MUX (correct for unsigned signals only).\n")
			}
			pStart = cur.rest()
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, int32(WLC_OBJ_MUX))
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			goto startword

		} else if Wlc_PrsStrCmp(pStart, "CPL_FF") != 0 {
			var NameId int32 = -1
			var NameIdIn int32 = -1
			var NameIdOut int32 = -1
			var fFound int32
			var nBits int32 = 1
			var fFlopIn int32
			var fFlopOut int32
			pStart = pStart[len("CPL_FF"):]
			if len(pStart) > 0 && wlcVerChar(pStart, 0) == '#' {
				nBits = atoi(pStart[1:])
			}
			// read names
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == "" {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart[1:])
				if p.pNtk.fAsyncRst == 0 && strings.HasPrefix(pStart, "arst") && (len(pStart) < 5 || pStart[4] != 'v') {
					var NameIdArst int32
					pStart = Wlc_PrsFindSymbol(pStart, '(')
					if pStart == "" {
						return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
					}
					pStart = Wlc_PrsSkipSpaces(pStart[1:])
					if Wlc_PrsIsDigit(pStart) != 0 {
						var Range, Signed, XValue int32
						var vFanins *Vec_Int_t = Vec_IntAlloc(100)
						pStart = Wlc_PrsReadConstant(p, pStart, vFanins, &Range, &Signed, &XValue)
						if pStart != "" && Vec_IntSize(vFanins) == 1 && Vec_IntEntry(vFanins, 0) == 0 {
							Vec_IntFree(vFanins)
							continue
						}
						fmt.Printf("Detected async reset.\n")
						p.pNtk.fAsyncRst = 1
						Vec_IntFree(vFanins)
						continue
					}
					pStart = Wlc_PrsFindName(pStart, &pName)
					if pStart == "" {
						return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
					}
					NameIdArst = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
					if p.pNtk.vArsts == nil {
						p.pNtk.vArsts = Vec_IntAlloc(100)
					}
					Vec_IntPushUnique(p.pNtk.vArsts, NameIdArst)
					continue
				}
				if wlcVerChar(pStart, 0) != 'd' && (wlcVerChar(pStart, 0) != 'q' || (len(pStart) > 1 && wlcVerChar(pStart, 1) == 'b')) && !strings.HasPrefix(pStart, "arstval") {
					continue
				}
				fFlopIn = boolToInt32(wlcVerChar(pStart, 0) == 'd')
				fFlopOut = boolToInt32(wlcVerChar(pStart, 0) == 'q')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart[1:], &pName)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				if fFlopIn != 0 {
					NameIdIn = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopOut != 0 {
					NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				}
				if fFound == 0 {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameIdIn == -1 || NameIdOut == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of flop input or flop output is missing.")
			}
			// create flop output
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, int32(WLC_OBJ_FO))
			Vec_IntPush(&p.pNtk.vFfs, NameIdOut)
			//if ( nBits != Wlc_ObjRange(pObj) )
			//    printf( "Warning!  Flop output \"%s\" has bit-width (%d) that differs from the flop declaration (%d)\n", Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits );
			// create flop input
			pObj = Wlc_NtkObj(p.pNtk, NameIdIn)
			Vec_IntPush(&p.pNtk.vFfs, NameIdIn)
			//if ( nBits != Wlc_ObjRange(pObj) )
			//    printf( "Warning!  Flop input \"%s\" has bit-width (%d) that differs from the flop declaration (%d)\n", Abc_NamStr(p->pNtk->pManName, NameIdIn), Wlc_ObjRange(pObj), nBits );
			// save flop init value
			if NameId == -1 {
				fmt.Printf("Initial value of flop \"%s\" is not specified. Zero is assumed.\n", Abc_NamStr(p.pNtk.pManName, NameIdOut))
			} else {
				pObj = Wlc_NtkObj(p.pNtk, NameId)
				if nBits != Wlc_ObjRange(pObj) {
					fmt.Printf("Warning!  Flop init signal \"%s\" bit-width (%d) is different from the flop declaration (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(pObj), nBits)
				}
			}
			if p.pNtk.vInits == nil {
				p.pNtk.vInits = Vec_IntAlloc(100)
			}
			Vec_IntPush(p.pNtk.vInits, wlcInitValue(NameId, nBits))
			// printf( "Created flop %s with range %d and init value %d (nameId = %d)\n",
			//     Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits, NameId );

		} else if Wlc_PrsStrCmp(pStart, "ABC_DFFRSE") != 0 {
			var NameId [10]int32
			var fFound int32
			var fFlopIn int32
			var fFlopClk int32
			var fFlopRst int32
			var fFlopSet int32
			var fFlopEna int32
			var fFlopAsync int32
			var fFlopSre int32
			var fFlopInit int32
			var fFlopOut int32
			pStart = pStart[len("ABC_DFFRSE"):]
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == "" {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart[1:])
				fFlopIn = boolToInt32(wlcVerChar(pStart, 0) == 'd')
				fFlopClk = boolToInt32(wlcVerChar(pStart, 0) == 'c')
				fFlopRst = boolToInt32(wlcVerChar(pStart, 0) == 'r')
				fFlopSet = boolToInt32(wlcVerChar(pStart, 0) == 's' && len(pStart) > 1 && wlcVerChar(pStart, 1) == 'e')
				fFlopEna = boolToInt32(wlcVerChar(pStart, 0) == 'e')
				fFlopAsync = boolToInt32(wlcVerChar(pStart, 0) == 'a')
				fFlopSre = boolToInt32(wlcVerChar(pStart, 0) == 's' && len(pStart) > 1 && wlcVerChar(pStart, 1) == 'r')
				fFlopInit = boolToInt32(wlcVerChar(pStart, 0) == 'i')
				fFlopOut = boolToInt32(wlcVerChar(pStart, 0) == 'q')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart[1:], &pName)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				if fFlopIn != 0 {
					NameId[0] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopClk != 0 {
					NameId[1] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopRst != 0 {
					NameId[2] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopSet != 0 {
					NameId[3] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopEna != 0 {
					NameId[4] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopAsync != 0 {
					NameId[5] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopSre != 0 {
					NameId[6] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopInit != 0 {
					NameId[7] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopOut != 0 {
					NameId[8] = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					assert(false, "0")
				}
				if fFound == 0 {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameId[0] == -1 || NameId[7] == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of flop input or flop output is missing.")
			}
			// create output
			pObj = Wlc_NtkObj(p.pNtk, NameId[8])
			Wlc_ObjUpdateType(p.pNtk, pObj, int32(WLC_OBJ_FF))
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameId[0])
			Vec_IntPush(p.vFanins, NameId[1])
			Vec_IntPush(p.vFanins, NameId[2])
			Vec_IntPush(p.vFanins, NameId[3])
			Vec_IntPush(p.vFanins, NameId[4])
			Vec_IntPush(p.vFanins, NameId[5])
			Vec_IntPush(p.vFanins, NameId[6])
			Vec_IntPush(p.vFanins, NameId[7])
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if Wlc_PrsStrCmp(pStart, "ABC_DFF") != 0 {
			var NameId int32 = -1
			var NameIdIn int32 = -1
			var NameIdOut int32 = -1
			var fFound int32
			var nBits int32 = 1
			var fFlopIn int32
			var fFlopOut int32
			pStart = pStart[len("ABC_DFF"):]
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == "" {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart[1:])
				fFlopIn = boolToInt32(wlcVerChar(pStart, 0) == 'd')
				fFlopOut = boolToInt32(wlcVerChar(pStart, 0) == 'q')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart[1:], &pName)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				if fFlopIn != 0 {
					NameIdIn = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fFlopOut != 0 {
					NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				}
				if fFound == 0 {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameIdIn == -1 || NameIdOut == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of flop input or flop output is missing.")
			}
			// create flop output
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, int32(WLC_OBJ_FO))
			Vec_IntPush(&p.pNtk.vFfs, NameIdOut)
			nBits = Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdOut))
			// create flop input
			pObj = Wlc_NtkObj(p.pNtk, NameIdIn)
			Vec_IntPush(&p.pNtk.vFfs, NameIdIn)
			// compare bit-width
			if Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdIn)) != nBits {
				fmt.Printf("Warning!  Flop input \"%s\" bit-width (%d) is different from that of flop output (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdIn)), nBits)
			}
			// save flop init value
			if NameId == -1 {
				fmt.Printf("Initial value of flop \"%s\" is not specified. Zero is assumed.\n", Abc_NamStr(p.pNtk.pManName, NameIdOut))
			} else {
				if Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameId)) != nBits {
					fmt.Printf("Warning!  Flop init signal \"%s\" bit-width (%d) is different from that of flop output (%d)\n", Abc_NamStr(p.pNtk.pManName, NameId), Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameId)), nBits)
				}
			}
			if p.pNtk.vInits == nil {
				p.pNtk.vInits = Vec_IntAlloc(100)
			}
			Vec_IntPush(p.pNtk.vInits, wlcInitValue(NameId, Wlc_ObjRange(Wlc_NtkObj(p.pNtk, NameIdOut))))
			// printf( "Created flop %s with range %d and init value %d (nameId = %d)\n",
			//     Abc_NamStr(p->pNtk->pManName, NameIdOut), Wlc_ObjRange(pObj), nBits, NameId );
			p.pNtk.fEasyFfs = 1

		} else if Wlc_PrsStrCmp(pStart, "CPL_MEM_") != 0 {
			var pNameId *int32 = nil
			var NameOutput int32
			var NameMi int32 = -1
			var NameMo int32 = -1
			var NameAddr int32 = -1
			var NameDi int32 = -1
			var NameDo int32 = -1
			var fFound int32
			var fRead int32 = 1
			pStart = pStart[len("CPL_MEM_"):]
			if wlcVerChar(pStart, 0) == 'W' {
				fRead = 0
			}
			// read names
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == "" {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart[1:])
				if strings.HasPrefix(pStart, "mem_data_in") {
					pNameId = &NameMi
				} else if strings.HasPrefix(pStart, "data_in") {
					pNameId = &NameDi
				} else if strings.HasPrefix(pStart, "data_out") {
					if fRead != 0 {
						pNameId = &NameDo
					} else {
						pNameId = &NameMo
					}
				} else if strings.HasPrefix(pStart, "addr_in") {
					pNameId = &NameAddr
				} else {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name of the input/output port.")
				}
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart[1:], &pName)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				*pNameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				if fFound == 0 {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if (fRead != 0) && (NameMi == -1 || NameAddr == -1 || NameDo == -1) {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of read port is missing.")
			}
			if !(fRead != 0) && (NameMi == -1 || NameAddr == -1 || NameDi == -1 || NameMo == -1) {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of write port is missing.")
			}
			// create output
			if fRead != 0 {
				NameOutput = NameDo
			} else {
				NameOutput = NameMo
			}
			pObj = Wlc_NtkObj(p.pNtk, NameOutput)
			Wlc_ObjUpdateType(p.pNtk, pObj, wlcReadWriteType(fRead))
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameMi)
			Vec_IntPush(p.vFanins, NameAddr)
			if !(fRead != 0) {
				Vec_IntPush(p.vFanins, NameDi)
			}
			//printf( "Memory %s ", fRead ? "Read" : "Write" ); printf( "Fanins: " );  Vec_IntPrint( p->vFanins );
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if Wlc_PrsStrCmp(pStart, "ABC_READ") != 0 {
			var pNameId *int32 = nil
			var NameMemIn int32 = -1
			var NameData int32 = -1
			var NameAddr int32 = -1
			var fFound int32
			pStart = pStart[len("ABC_READ"):]
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == "" {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart[1:])
				if strings.HasPrefix(pStart, "mem_in") {
					pNameId = &NameMemIn
				} else if strings.HasPrefix(pStart, "addr") {
					pNameId = &NameAddr
				} else if strings.HasPrefix(pStart, "data") {
					pNameId = &NameData
				} else {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name of the input/output port.")
				}
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart[1:], &pName)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				*pNameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				if fFound == 0 {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameMemIn == -1 || NameAddr == -1 || NameData == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of read port is missing.")
			}
			// create output
			pObj = Wlc_NtkObj(p.pNtk, NameData)
			Wlc_ObjUpdateType(p.pNtk, pObj, int32(WLC_OBJ_READ))
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameMemIn)
			Vec_IntPush(p.vFanins, NameAddr)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			p.pNtk.fMemPorts = 1

		} else if Wlc_PrsStrCmp(pStart, "ABC_WRITE") != 0 {
			var pNameId *int32 = nil
			var NameMemIn int32 = -1
			var NameMemOut int32 = -1
			var NameData int32 = -1
			var NameAddr int32 = -1
			var fFound int32
			pStart = pStart[len("ABC_WRITE"):]
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == "" {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart[1:])
				if strings.HasPrefix(pStart, "mem_in") {
					pNameId = &NameMemIn
				} else if strings.HasPrefix(pStart, "mem_out") {
					pNameId = &NameMemOut
				} else if strings.HasPrefix(pStart, "data") {
					pNameId = &NameData
				} else if strings.HasPrefix(pStart, "addr") {
					pNameId = &NameAddr
				} else {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name of the input/output port.")
				}
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the flop description.")
				}
				pStart = Wlc_PrsFindName(pStart[1:], &pName)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside flop description.")
				}
				*pNameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				if fFound == 0 {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameMemIn == -1 || NameAddr == -1 || NameData == -1 || NameMemOut == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Name of one of signals of write port is missing.")
			}
			// create output
			pObj = Wlc_NtkObj(p.pNtk, NameMemOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, int32(WLC_OBJ_WRITE))
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameMemIn)
			Vec_IntPush(p.vFanins, NameAddr)
			Vec_IntPush(p.vFanins, NameData)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)
			p.pNtk.fMemPorts = 1

		} else if (Wlc_PrsStrCmp(pStart, "CPL_RROT") != 0) || (Wlc_PrsStrCmp(pStart, "CPL_LROT") != 0) {
			// CPL_RROT #(128, 6) I_47479(.o ( E_46713 )  , .d ( E_46718 )  , .s ( E_46712 )  );
			var right_rotation int32 = Wlc_PrsStrCmp(pStart, "CPL_RROT")
			var NameId int32 = -1
			var NameIdOut int32 = -1
			var NameIdInD int32 = -1
			var NameIdInS int32 = -1
			var fFound int32
			var fRotInD int32
			var fRotInS int32
			var fRotOut int32
			pStart = pStart[len("CPL_RROT"):]

			// NOTE: no need to parse the parameter values
			//if ( wlcVerChar(pStart, 0) == '#' )

			// read names
			for {
				pStart = Wlc_PrsFindSymbol(pStart, '.')
				if pStart == "" {
					break
				}
				pStart = Wlc_PrsSkipSpaces(pStart[1:])
				if wlcVerChar(pStart, 0) != 'o' && wlcVerChar(pStart, 0) != 'd' && wlcVerChar(pStart, 0) != 's' {
					continue
				}
				fRotInD = boolToInt32(wlcVerChar(pStart, 0) == 'd')
				fRotInS = boolToInt32(wlcVerChar(pStart, 0) == 's')
				fRotOut = boolToInt32(wlcVerChar(pStart, 0) == 'o')
				pStart = Wlc_PrsFindSymbol(pStart, '(')
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read opening parenthesis in the rotation description.")
				}
				pStart = Wlc_PrsFindName(pStart[1:], &pName)
				if pStart == "" {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read name inside rotation description.")
				}
				if fRotInD != 0 {
					NameIdInD = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fRotInS != 0 {
					NameIdInS = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else if fRotOut != 0 {
					NameIdOut = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				} else {
					NameId = Abc_NamStrFindOrAdd(p.pNtk.pManName, pName, &fFound)
				}
				_ = NameId // The source parser retains this instance name only locally.
				if fFound == 0 {
					return Wlc_PrsWriteErrorMessage(p, pStart, "Name %s is not declared.", pName)
				}
			}
			if NameIdOut == -1 || NameIdInD == -1 || NameIdInS == -1 {
				return Wlc_PrsWriteErrorMessage(p, pStart, "Some fields of CPL_ROT are missing.")
			}
			// create rot output
			pObj = Wlc_NtkObj(p.pNtk, NameIdOut)
			Wlc_ObjUpdateType(p.pNtk, pObj, wlcRotationType(right_rotation))
			Vec_IntClear(p.vFanins)
			Vec_IntPush(p.vFanins, NameIdInD)
			Vec_IntPush(p.vFanins, NameIdInS)
			Wlc_ObjAddFanins(p.pNtk, pObj, p.vFanins)

		} else if len(pStart) > 1 && wlcVerChar(pStart, 0) == '(' && wlcVerChar(pStart, 1) == '*' { // skip comments
			if end := strings.IndexByte(pStart, ')'); end >= 0 {
				pStart = pStart[end+1:]
			} else {
				pStart = ""
			}
			pStart = Wlc_PrsSkipSpaces(pStart)
			goto startword

		} else if wlcVerChar(pStart, 0) != '`' {
			var iLine int32 = Wlc_PrsFindLine(p, pStart)
			pStart = Wlc_PrsFindName(pStart, &pName)
			return Wlc_PrsWriteErrorMessage(p, pStart, "Cannot read line %d beginning with %s.", iLine, verFallbackName(pName))
		}
	}
	if p.nNonZero[0] != 0 {
		fmt.Printf("Warning: Input file contains %d objects with non-zero-based ranges.\n", p.nNonZero[0])
		fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nNonZero[1], p.nNonZero[2], p.nNonZero[3])
	}
	if p.nNegative[0] != 0 {
		fmt.Printf("Warning: Input file contains %d objects with negative ranges.\n", p.nNegative[0])
		fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nNegative[1], p.nNegative[2], p.nNegative[3])
	}
	if p.nReverse[0] != 0 {
		fmt.Printf("Warning: Input file contains %d objects with reversed ranges.\n", p.nReverse[0])
		fmt.Printf("For example, signal with range [%d:%d] is declared in line %d.\n", p.nReverse[1], p.nReverse[2], p.nReverse[3])
	}
	return 1
}
func Wlc_ReadVer(pFileName string, pStr string) *Wlc_Ntk_t {
	var p *Wlc_Prs_t
	var pNtk *Wlc_Ntk_t = nil
	// either file name or string is given
	assert((pFileName == "") != (pStr == ""), "exactly one of file name and source text must be provided")
	// start the parser
	p = Wlc_PrsStart(pFileName, pStr)
	if p == nil {
		return nil
	}
	// detect lines
	if !(Wlc_PrsPrepare(p) != 0) {
		goto finish
	}
	// parse models
	if !(Wlc_PrsDerive(p) != 0) {
		goto finish
	}
	// derive topological order
	if p.pNtk != nil {
		var pObj *Wlc_Obj_t
		var i int32
		// inlining of macro Wlc_NtkForEachObj
		for i = 1; i < Wlc_NtkObjNumMax(p.pNtk); i++ {
			pObj = Wlc_NtkObj(p.pNtk, i)
			if pObj.Type == uint32(WLC_OBJ_FF) {
				Vec_IntPush(&p.pNtk.vFfs2, Wlc_ObjId(p.pNtk, pObj))
			}
		}
		pNtk = Wlc_NtkDupDfs(p.pNtk, 0, 1)
		pNtk.pSpec = Abc_UtilStrsav(pFileName)
	}
finish:
	Wlc_PrsPrintErrorMessage(p)
	Wlc_PrsStop(p)
	return pNtk
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Io_ReadWordTest(pFileName string) {
	var pNew *Gia_Man_t
	var pNtk *Wlc_Ntk_t = Wlc_ReadVer(pFileName, "")
	if pNtk == nil {
		return
	}
	Wlc_WriteVer(pNtk, "test.v", 0, 0)

	pNew = Wlc_NtkBitBlast(pNtk, nil)
	Gia_AigerWrite(pNew, "test.aig", 0, 0, 0)
	Gia_ManStop(pNew)

	Wlc_NtkFree(pNtk)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
