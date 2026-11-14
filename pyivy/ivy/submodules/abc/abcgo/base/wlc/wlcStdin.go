package wlc

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [wlcStdin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [stdin processing for STM interface.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcStdin.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converts a bit-string into a number in a given radix.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    [Report results.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Wlc_NtkReport(p *Wlc_Ntk_t, pCex *Abc_Cex_t, pName string, radix int32) {
	if p == nil || pCex == nil {
		return
	}
	nameID := Abc_NamStrFind(p.pManName, pName)
	if nameID <= 0 {
		fmt.Printf("Cannot find %q among nodes of the network.\n", pName)
		return
	}
	var start, bits int32
	found := false
	for i := int32(0); i+2 < Vec_IntSize(&p.vValues); i += 3 {
		if Vec_IntEntry(&p.vValues, i) == nameID {
			start, bits, found = Vec_IntEntry(&p.vValues, i+1), Vec_IntEntry(&p.vValues, i+2), true
			break
		}
	}
	if !found {
		fmt.Printf("Cannot find %q among primary inputs of the network.\n", pName)
		return
	}
	assert(radix == 2 || radix == 10 || radix == 16, "radix is supported")
	value := Wlc_ConvertToRadix(pCex.pData, start, bits, radix)
	prefix := ""
	if radix == 16 {
		prefix = "#x"
	} else if radix == 2 {
		prefix = "#b"
	}
	fmt.Printf("((%s %s%s))\n", pName, prefix, Vec_StrArray(value))
	Vec_StrFree(value)
}

func Wlc_StdinCollectStop(vInput *Vec_Str_t, line string, lineSize int32) int32 {
	data := Vec_StrArray(vInput)
	n := int(lineSize)
	if len(data) < n {
		return 0
	}
	return int32(strings.HasPrefix(string(data[len(data)-n:]), line))
}

func readStdinByte() (byte, error) {
	var one [1]byte
	_, err := os.Stdin.Read(one[:])
	return one[0], err
}

func Wlc_StdinCollectProblem(directive string) *Vec_Str_t {
	v := Vec_StrAlloc(1000)
	for {
		c, err := readStdinByte()
		if err != nil {
			break
		}
		Vec_StrPush(v, c)
		if c == ')' && Wlc_StdinCollectStop(v, directive, int32(len(directive))) != 0 {
			break
		}
	}
	Vec_StrPush(v, 0)
	return v
}

func Wlc_StdinCollectQuery() *Vec_Str_t {
	v := Vec_StrAlloc(1000)
	depth, started := 0, false
	for {
		c, err := readStdinByte()
		if err != nil {
			if !started {
				Vec_StrFreeP(&v)
				return nil
			}
			break
		}
		Vec_StrPush(v, c)
		if c == '(' {
			depth++
			started = true
		} else if c == ')' && started {
			depth--
			if depth == 0 {
				break
			}
		}
	}
	Vec_StrPush(v, 0)
	return v
}

func Wlc_StdinProcessSmt(pAbc *Abc_Frame_t, pCmd string) int32 {
	input := Wlc_StdinCollectProblem("(check-sat)")
	pNtk := Wlc_ReadSmtBuffer("top", string(Vec_StrArray(input)), Vec_StrSize(input), 0, 0)
	Vec_StrFree(input)
	Wlc_SetNtk(pAbc, pNtk)
	if Cmd_CommandExecute(pAbc, pCmd) != 0 {
		Abc_Print(1, "Something did not work out with the command %q.\n", pCmd)
		return 0
	}
	status := Abc_FrameReadProbStatus(pAbc)
	switch status {
	case -1:
		fmt.Println("undecided")
	case 1:
		fmt.Println("unsat")
	case 0:
		fmt.Println("sat")
	default:
		assert(false, "known solver status")
	}
	_ = os.Stdout.Sync()
	for input = Wlc_StdinCollectQuery(); input != nil; input = Wlc_StdinCollectQuery() {
		fields := strings.Fields(strings.Trim(string(Vec_StrArray(input)), "\x00 \t\r\n()"))
		if len(fields) == 0 || fields[0] != "get-value" {
			Abc_Print(1, "ABC is expecting \"get-value\" in a follow-up input of the satisfiable problem.\n")
			Vec_StrFree(input)
			return 0
		}
		if Abc_FrameReadProbStatus(pAbc) != 0 {
			Abc_Print(1, "ABC received a follow-up input for a problem that is not known to be satisfiable.\n")
			Vec_StrFree(input)
			return 0
		}
		cex := Abc_FrameReadCex(pAbc)
		if cex == nil {
			Abc_Print(1, "ABC does not have a counter-example available to process a \"get-value\" request.\n")
			Vec_StrFree(input)
			return 0
		}
		if len(fields) > 1 {
			Wlc_NtkReport(pAbc.pAbcWlc, cex, fields[1], 16)
		}
		Vec_StrFree(input)
		_ = os.Stdout.Sync()
	}
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
