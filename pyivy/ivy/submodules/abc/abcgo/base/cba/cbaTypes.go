package cba

/**CFile****************************************************************

  FileName    [cbaTypes.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: cbaTypes.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// network objects
type Cba_ObjType_t = int32

const (
	CBA_OBJ_NONE Cba_ObjType_t = 0
	// 00:  unused
	CBA_OBJ_PI Cba_ObjType_t = CBA_OBJ_NONE + 1
	// 01:  input
	CBA_OBJ_PO Cba_ObjType_t = CBA_OBJ_PI + 1
	// 02:  output
	CBA_OBJ_BOX Cba_ObjType_t = CBA_OBJ_PO + 1
	// 03:  box

	CBA_BOX_CF Cba_ObjType_t = CBA_OBJ_BOX + 1
	// 04:
	CBA_BOX_CT Cba_ObjType_t = CBA_BOX_CF + 1
	// 05:
	CBA_BOX_CX Cba_ObjType_t = CBA_BOX_CT + 1
	// 06:
	CBA_BOX_CZ Cba_ObjType_t = CBA_BOX_CX + 1
	// 07:

	CBA_BOX_BUF Cba_ObjType_t = CBA_BOX_CZ + 1
	// 08:
	CBA_BOX_INV Cba_ObjType_t = CBA_BOX_BUF + 1
	// 09:
	CBA_BOX_AND Cba_ObjType_t = CBA_BOX_INV + 1
	// 10:
	CBA_BOX_NAND Cba_ObjType_t = CBA_BOX_AND + 1
	// 11:
	CBA_BOX_OR Cba_ObjType_t = CBA_BOX_NAND + 1
	// 12:
	CBA_BOX_NOR Cba_ObjType_t = CBA_BOX_OR + 1
	// 13:
	CBA_BOX_XOR Cba_ObjType_t = CBA_BOX_NOR + 1
	// 14:
	CBA_BOX_XNOR Cba_ObjType_t = CBA_BOX_XOR + 1
	// 15:
	CBA_BOX_SHARP Cba_ObjType_t = CBA_BOX_XNOR + 1
	// 16:
	CBA_BOX_SHARPL Cba_ObjType_t = CBA_BOX_SHARP + 1
	// 17:
	CBA_BOX_MUX Cba_ObjType_t = CBA_BOX_SHARPL + 1
	// 18:
	CBA_BOX_MAJ Cba_ObjType_t = CBA_BOX_MUX + 1
	// 19:

	CBA_BOX_ABC Cba_ObjType_t = CBA_BOX_MAJ + 1
	// 20:
	CBA_BOX_BA Cba_ObjType_t = CBA_BOX_ABC + 1
	// 21:
	CBA_BOX_BO Cba_ObjType_t = CBA_BOX_BA + 1
	// 22:
	CBA_BOX_BX Cba_ObjType_t = CBA_BOX_BO + 1
	// 23:
	CBA_BOX_BN Cba_ObjType_t = CBA_BOX_BX + 1
	// 24:
	CBA_BOX_BAO Cba_ObjType_t = CBA_BOX_BN + 1
	// 25:
	CBA_BOX_BOA Cba_ObjType_t = CBA_BOX_BAO + 1
	// 26:

	CBA_BOX_RAND Cba_ObjType_t = CBA_BOX_BOA + 1
	// 27:
	CBA_BOX_RNAND Cba_ObjType_t = CBA_BOX_RAND + 1
	// 28:
	CBA_BOX_ROR Cba_ObjType_t = CBA_BOX_RNAND + 1
	// 29:
	CBA_BOX_RNOR Cba_ObjType_t = CBA_BOX_ROR + 1
	// 30:
	CBA_BOX_RXOR Cba_ObjType_t = CBA_BOX_RNOR + 1
	// 31:
	CBA_BOX_RXNOR Cba_ObjType_t = CBA_BOX_RXOR + 1
	// 32:

	CBA_BOX_LNOT Cba_ObjType_t = CBA_BOX_RXNOR + 1
	// 33
	CBA_BOX_LAND Cba_ObjType_t = CBA_BOX_LNOT + 1
	// 34:
	CBA_BOX_LNAND Cba_ObjType_t = CBA_BOX_LAND + 1
	// 35:
	CBA_BOX_LOR Cba_ObjType_t = CBA_BOX_LNAND + 1
	// 36:
	CBA_BOX_LNOR Cba_ObjType_t = CBA_BOX_LOR + 1
	// 37:
	CBA_BOX_LXOR Cba_ObjType_t = CBA_BOX_LNOR + 1
	// 38:
	CBA_BOX_LXNOR Cba_ObjType_t = CBA_BOX_LXOR + 1
	// 39:

	CBA_BOX_NMUX Cba_ObjType_t = CBA_BOX_LXNOR + 1
	// 40:
	CBA_BOX_SEL Cba_ObjType_t = CBA_BOX_NMUX + 1
	// 41:
	CBA_BOX_PSEL Cba_ObjType_t = CBA_BOX_SEL + 1
	// 42:
	CBA_BOX_ENC Cba_ObjType_t = CBA_BOX_PSEL + 1
	// 43:
	CBA_BOX_PENC Cba_ObjType_t = CBA_BOX_ENC + 1
	// 44:
	CBA_BOX_DEC Cba_ObjType_t = CBA_BOX_PENC + 1
	// 45:
	CBA_BOX_EDEC Cba_ObjType_t = CBA_BOX_DEC + 1
	// 46:

	CBA_BOX_ADD Cba_ObjType_t = CBA_BOX_EDEC + 1
	// 47:
	CBA_BOX_SUB Cba_ObjType_t = CBA_BOX_ADD + 1
	// 48:
	CBA_BOX_MUL Cba_ObjType_t = CBA_BOX_SUB + 1
	// 49:
	CBA_BOX_SMUL Cba_ObjType_t = CBA_BOX_MUL + 1
	// 50:
	CBA_BOX_DIV Cba_ObjType_t = CBA_BOX_SMUL + 1
	// 51:
	CBA_BOX_MOD Cba_ObjType_t = CBA_BOX_DIV + 1
	// 52:
	CBA_BOX_REM Cba_ObjType_t = CBA_BOX_MOD + 1
	// 53:
	CBA_BOX_POW Cba_ObjType_t = CBA_BOX_REM + 1
	// 54:
	CBA_BOX_MIN Cba_ObjType_t = CBA_BOX_POW + 1
	// 55:
	CBA_BOX_SQRT Cba_ObjType_t = CBA_BOX_MIN + 1
	// 56:
	CBA_BOX_ABS Cba_ObjType_t = CBA_BOX_SQRT + 1
	// 57:

	CBA_BOX_SLTHAN Cba_ObjType_t = CBA_BOX_ABS + 1
	// 58:
	CBA_BOX_LTHAN Cba_ObjType_t = CBA_BOX_SLTHAN + 1
	// 59:
	CBA_BOX_LETHAN Cba_ObjType_t = CBA_BOX_LTHAN + 1
	// 60:
	CBA_BOX_METHAN Cba_ObjType_t = CBA_BOX_LETHAN + 1
	// 61:
	CBA_BOX_MTHAN Cba_ObjType_t = CBA_BOX_METHAN + 1
	// 62:
	CBA_BOX_EQU Cba_ObjType_t = CBA_BOX_MTHAN + 1
	// 63:
	CBA_BOX_NEQU Cba_ObjType_t = CBA_BOX_EQU + 1
	// 64:

	CBA_BOX_SHIL Cba_ObjType_t = CBA_BOX_NEQU + 1
	// 65:
	CBA_BOX_SHIR Cba_ObjType_t = CBA_BOX_SHIL + 1
	// 66:
	CBA_BOX_SHILA Cba_ObjType_t = CBA_BOX_SHIR + 1
	// 67:
	CBA_BOX_SHIRA Cba_ObjType_t = CBA_BOX_SHILA + 1
	// 68:
	CBA_BOX_ROTL Cba_ObjType_t = CBA_BOX_SHIRA + 1
	// 69:
	CBA_BOX_ROTR Cba_ObjType_t = CBA_BOX_ROTL + 1
	// 70:

	CBA_BOX_NODE Cba_ObjType_t = CBA_BOX_ROTR + 1
	// 71:
	CBA_BOX_LUT Cba_ObjType_t = CBA_BOX_NODE + 1
	// 72:
	CBA_BOX_GATE Cba_ObjType_t = CBA_BOX_LUT + 1
	// 73:
	CBA_BOX_TABLE Cba_ObjType_t = CBA_BOX_GATE + 1
	// 74:

	CBA_BOX_TRI Cba_ObjType_t = CBA_BOX_TABLE + 1
	// 75:
	CBA_BOX_RAM Cba_ObjType_t = CBA_BOX_TRI + 1
	// 76:
	CBA_BOX_RAMR Cba_ObjType_t = CBA_BOX_RAM + 1
	// 77:
	CBA_BOX_RAMW Cba_ObjType_t = CBA_BOX_RAMR + 1
	// 78:
	CBA_BOX_RAMWC Cba_ObjType_t = CBA_BOX_RAMW + 1
	// 79:
	CBA_BOX_RAML Cba_ObjType_t = CBA_BOX_RAMWC + 1
	// 80:
	CBA_BOX_RAMS Cba_ObjType_t = CBA_BOX_RAML + 1
	// 81:
	CBA_BOX_RAMBOX Cba_ObjType_t = CBA_BOX_RAMS + 1
	// 82:

	CBA_BOX_LATCH Cba_ObjType_t = CBA_BOX_RAMBOX + 1
	// 83:
	CBA_BOX_LATCHRS Cba_ObjType_t = CBA_BOX_LATCH + 1
	// 84:
	CBA_BOX_DFF Cba_ObjType_t = CBA_BOX_LATCHRS + 1
	// 85:
	CBA_BOX_DFFCPL Cba_ObjType_t = CBA_BOX_DFF + 1
	// 86:
	CBA_BOX_DFFRS Cba_ObjType_t = CBA_BOX_DFFCPL + 1
	// 87:

	CBA_BOX_SLICE Cba_ObjType_t = CBA_BOX_DFFRS + 1
	// 88:
	CBA_BOX_CONCAT Cba_ObjType_t = CBA_BOX_SLICE + 1
	// 89:

	CBA_BOX_LAST Cba_ObjType_t = CBA_BOX_CONCAT + 1 // 90

)

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                          ITERATORS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
