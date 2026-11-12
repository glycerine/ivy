package miniaig

/**CFile****************************************************************

  FileName    [acbTypes.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: acbTypes.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

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
 type Acb_ObjType_t int32
const ( 
    ABC_OPER_NONE Acb_ObjType_t = 0
     // 00  unused
    ABC_OPER_PI Acb_ObjType_t = ABC_OPER_NONE + 1
           // 01  input
    ABC_OPER_PO Acb_ObjType_t = ABC_OPER_PI + 1
           // 02  output
    ABC_OPER_CI Acb_ObjType_t = ABC_OPER_PO + 1
           // 03  combinational input
    ABC_OPER_CO Acb_ObjType_t = ABC_OPER_CI + 1
           // 04  combinational output
    ABC_OPER_FON Acb_ObjType_t = ABC_OPER_CO + 1
          // 05  output placeholder
    ABC_OPER_BOX Acb_ObjType_t = ABC_OPER_FON + 1
          // 06  box

    ABC_OPER_CONST_F Acb_ObjType_t = ABC_OPER_BOX + 1
      // 07
    ABC_OPER_CONST_T Acb_ObjType_t = ABC_OPER_CONST_F + 1
      // 08
    ABC_OPER_CONST_X Acb_ObjType_t = ABC_OPER_CONST_T + 1
      // 09
    ABC_OPER_CONST_Z Acb_ObjType_t = ABC_OPER_CONST_X + 1
      // 10

    ABC_OPER_BIT_BUF Acb_ObjType_t = ABC_OPER_CONST_Z + 1
      // 11
    ABC_OPER_BIT_INV Acb_ObjType_t = ABC_OPER_BIT_BUF + 1
      // 12
    ABC_OPER_BIT_AND Acb_ObjType_t = ABC_OPER_BIT_INV + 1
      // 13
    ABC_OPER_BIT_NAND Acb_ObjType_t = ABC_OPER_BIT_AND + 1
     // 14
    ABC_OPER_BIT_OR Acb_ObjType_t = ABC_OPER_BIT_NAND + 1
       // 15
    ABC_OPER_BIT_NOR Acb_ObjType_t = ABC_OPER_BIT_OR + 1
      // 16
    ABC_OPER_BIT_XOR Acb_ObjType_t = ABC_OPER_BIT_NOR + 1
      // 17
    ABC_OPER_BIT_NXOR Acb_ObjType_t = ABC_OPER_BIT_XOR + 1
     // 18
    ABC_OPER_BIT_SHARP Acb_ObjType_t = ABC_OPER_BIT_NXOR + 1
    // 19
    ABC_OPER_BIT_SHARPL Acb_ObjType_t = ABC_OPER_BIT_SHARP + 1
   // 20
    ABC_OPER_BIT_MUX Acb_ObjType_t = ABC_OPER_BIT_SHARPL + 1
      // 21  fanins are: {Ctrl, Data1, Data0}
    ABC_OPER_BIT_MAJ Acb_ObjType_t = ABC_OPER_BIT_MUX + 1
      // 22

    ABC_OPER_ABC Acb_ObjType_t = ABC_OPER_BIT_MAJ + 1
          // 23
    ABC_OPER_BA Acb_ObjType_t = ABC_OPER_ABC + 1
           // 24
    ABC_OPER_BO Acb_ObjType_t = ABC_OPER_BA + 1
           // 25
    ABC_OPER_BX Acb_ObjType_t = ABC_OPER_BO + 1
           // 26
    ABC_OPER_BN Acb_ObjType_t = ABC_OPER_BX + 1
           // 27
    ABC_OPER_BAO Acb_ObjType_t = ABC_OPER_BN + 1
          // 28
    ABC_OPER_BOA Acb_ObjType_t = ABC_OPER_BAO + 1
          // 29

    ABC_OPER_RED_AND Acb_ObjType_t = ABC_OPER_BOA + 1
      // 30
    ABC_OPER_RED_NAND Acb_ObjType_t = ABC_OPER_RED_AND + 1
     // 31
    ABC_OPER_RED_OR Acb_ObjType_t = ABC_OPER_RED_NAND + 1
       // 32
    ABC_OPER_RED_NOR Acb_ObjType_t = ABC_OPER_RED_OR + 1
      // 33
    ABC_OPER_RED_XOR Acb_ObjType_t = ABC_OPER_RED_NOR + 1
      // 34
    ABC_OPER_RED_NXOR Acb_ObjType_t = ABC_OPER_RED_XOR + 1
     // 35

    ABC_OPER_LOGIC_NOT Acb_ObjType_t = ABC_OPER_RED_NXOR + 1
    // 36
    ABC_OPER_LOGIC_AND Acb_ObjType_t = ABC_OPER_LOGIC_NOT + 1
    // 37
    ABC_OPER_LOGIC_NAND Acb_ObjType_t = ABC_OPER_LOGIC_AND + 1
   // 38
    ABC_OPER_LOGIC_OR Acb_ObjType_t = ABC_OPER_LOGIC_NAND + 1
     // 39
    ABC_OPER_LOGIC_NOR Acb_ObjType_t = ABC_OPER_LOGIC_OR + 1
    // 40
    ABC_OPER_LOGIC_XOR Acb_ObjType_t = ABC_OPER_LOGIC_NOR + 1
    // 41
    ABC_OPER_LOGIC_XNOR Acb_ObjType_t = ABC_OPER_LOGIC_XOR + 1
   // 42

    ABC_OPER_SEL_NMUX Acb_ObjType_t = ABC_OPER_LOGIC_XNOR + 1
     // 43  fanins are: {Ctrl, Data0, Data1, Data2, ...}
    ABC_OPER_SEL_SEL Acb_ObjType_t = ABC_OPER_SEL_NMUX + 1
      // 44
    ABC_OPER_SEL_PSEL Acb_ObjType_t = ABC_OPER_SEL_SEL + 1
     // 45
    ABC_OPER_SEL_ENC Acb_ObjType_t = ABC_OPER_SEL_PSEL + 1
      // 46
    ABC_OPER_SEL_PENC Acb_ObjType_t = ABC_OPER_SEL_ENC + 1
     // 47
    ABC_OPER_SEL_DEC Acb_ObjType_t = ABC_OPER_SEL_PENC + 1
      // 48
    ABC_OPER_SEL_EDEC Acb_ObjType_t = ABC_OPER_SEL_DEC + 1
     // 49

    ABC_OPER_ARI_ADD Acb_ObjType_t = ABC_OPER_SEL_EDEC + 1
      // 50
    ABC_OPER_ARI_SUB Acb_ObjType_t = ABC_OPER_ARI_ADD + 1
      // 51
    ABC_OPER_ARI_MUL Acb_ObjType_t = ABC_OPER_ARI_SUB + 1
      // 52
    ABC_OPER_ARI_SMUL Acb_ObjType_t = ABC_OPER_ARI_MUL + 1
     // 53
    ABC_OPER_ARI_DIV Acb_ObjType_t = ABC_OPER_ARI_SMUL + 1
      // 54
    ABC_OPER_ARI_MOD Acb_ObjType_t = ABC_OPER_ARI_DIV + 1
      // 55
    ABC_OPER_ARI_REM Acb_ObjType_t = ABC_OPER_ARI_MOD + 1
      // 56
    ABC_OPER_ARI_POW Acb_ObjType_t = ABC_OPER_ARI_REM + 1
      // 57
    ABC_OPER_ARI_MIN Acb_ObjType_t = ABC_OPER_ARI_POW + 1
      // 58
    ABC_OPER_ARI_SQRT Acb_ObjType_t = ABC_OPER_ARI_MIN + 1
     // 59
    ABC_OPER_ARI_ABS Acb_ObjType_t = ABC_OPER_ARI_SQRT + 1
      // 60

    ABC_OPER_COMP_SLESS Acb_ObjType_t = ABC_OPER_ARI_ABS + 1
   // 61
    ABC_OPER_COMP_LESS Acb_ObjType_t = ABC_OPER_COMP_SLESS + 1
    // 62
    ABC_OPER_COMP_LESSEQU Acb_ObjType_t = ABC_OPER_COMP_LESS + 1
 // 63
    ABC_OPER_COMP_MOREEQU Acb_ObjType_t = ABC_OPER_COMP_LESSEQU + 1
 // 64
    ABC_OPER_COMP_MORE Acb_ObjType_t = ABC_OPER_COMP_MOREEQU + 1
    // 65
    ABC_OPER_COMP_EQU Acb_ObjType_t = ABC_OPER_COMP_MORE + 1
     // 66
    ABC_OPER_COMP_NOTEQU Acb_ObjType_t = ABC_OPER_COMP_EQU + 1
  // 67

    ABC_OPER_SHIFT_L Acb_ObjType_t = ABC_OPER_COMP_NOTEQU + 1
      // 68
    ABC_OPER_SHIFT_R Acb_ObjType_t = ABC_OPER_SHIFT_L + 1
      // 69
    ABC_OPER_SHIFT_LA Acb_ObjType_t = ABC_OPER_SHIFT_R + 1
     // 70
    ABC_OPER_SHIFT_RA Acb_ObjType_t = ABC_OPER_SHIFT_LA + 1
     // 71
    ABC_OPER_SHIFT_ROTL Acb_ObjType_t = ABC_OPER_SHIFT_RA + 1
   // 72
    ABC_OPER_SHIFT_ROTR Acb_ObjType_t = ABC_OPER_SHIFT_ROTL + 1
   // 73

    ABC_OPER_NODE Acb_ObjType_t = ABC_OPER_SHIFT_ROTR + 1
         // 74
    ABC_OPER_LUT Acb_ObjType_t = ABC_OPER_NODE + 1
          // 75
    ABC_OPER_GATE Acb_ObjType_t = ABC_OPER_LUT + 1
         // 76
    ABC_OPER_TABLE Acb_ObjType_t = ABC_OPER_GATE + 1
        // 77

    ABC_OPER_TRI Acb_ObjType_t = ABC_OPER_TABLE + 1
          // 78
    ABC_OPER_RAM Acb_ObjType_t = ABC_OPER_TRI + 1
          // 79
    ABC_OPER_RAMR Acb_ObjType_t = ABC_OPER_RAM + 1
         // 80
    ABC_OPER_RAMW Acb_ObjType_t = ABC_OPER_RAMR + 1
         // 81
    ABC_OPER_RAMWC Acb_ObjType_t = ABC_OPER_RAMW + 1
        // 82
    ABC_OPER_RAML Acb_ObjType_t = ABC_OPER_RAMWC + 1
         // 83
    ABC_OPER_RAMS Acb_ObjType_t = ABC_OPER_RAML + 1
         // 84
    ABC_OPER_RAMBOX Acb_ObjType_t = ABC_OPER_RAMS + 1
       // 85

    ABC_OPER_LATCH Acb_ObjType_t = ABC_OPER_RAMBOX + 1
        // 86
    ABC_OPER_LATCHRS Acb_ObjType_t = ABC_OPER_LATCH + 1
      // 87
    ABC_OPER_DFF Acb_ObjType_t = ABC_OPER_LATCHRS + 1
          // 88
    ABC_OPER_DFFRSE Acb_ObjType_t = ABC_OPER_DFF + 1
       // 89
    ABC_OPER_DFFLAST Acb_ObjType_t = ABC_OPER_DFFRSE + 1
      // 90

    ABC_OPER_SLICE Acb_ObjType_t = ABC_OPER_DFFLAST + 1
        // 91
    ABC_OPER_CONCAT Acb_ObjType_t = ABC_OPER_SLICE + 1
       // 92
    ABC_OPER_ZEROPAD Acb_ObjType_t = ABC_OPER_CONCAT + 1
      // 93
    ABC_OPER_SIGNEXT Acb_ObjType_t = ABC_OPER_ZEROPAD + 1
      // 94

    ABC_OPER_LOGIC_IMPL Acb_ObjType_t = ABC_OPER_SIGNEXT + 1
   // 95
    ABC_OPER_ARI_SQUARE Acb_ObjType_t = ABC_OPER_LOGIC_IMPL + 1
   // 96
    ABC_OPER_CONST Acb_ObjType_t = ABC_OPER_ARI_SQUARE + 1
        // 97

    ABC_OPER_ARI_ADDSUB Acb_ObjType_t = ABC_OPER_CONST + 1
   // 98

    ABC_OPER_LAST Acb_ObjType_t = ABC_OPER_ARI_ADDSUB + 1          // 99

)

  


// printing operator types
func Abc_OperName(Type int32) string {
    if Type == ABC_OPER_NONE {   return nil; }
    if Type == ABC_OPER_PI {   return "pi"; }     
    if Type == ABC_OPER_PO {   return "po"; }     
    if Type == ABC_OPER_CI {   return "ci"; }     
    if Type == ABC_OPER_CO {   return "co"; }     
    if Type == ABC_OPER_FON {   return "fon"; }     
    if Type == ABC_OPER_BOX {   return "box"; }    
    
    if Type == ABC_OPER_BIT_BUF {   return "buf"; }    
    if Type == ABC_OPER_BIT_INV {   return "~"; }      
    if Type == ABC_OPER_BIT_MUX {   return "mux"; }    
    if Type == ABC_OPER_BIT_MAJ {   return "maj"; }    
    if Type == ABC_OPER_BIT_AND {   return "&"; }      
    if Type == ABC_OPER_BIT_OR {   return "|"; }      
    if Type == ABC_OPER_BIT_XOR {   return "^"; }      
    if Type == ABC_OPER_BIT_NAND {   return "~&"; }     
    if Type == ABC_OPER_BIT_NOR {   return "~|"; }     
    if Type == ABC_OPER_BIT_NXOR {   return "~^"; }     

    if Type == ABC_OPER_RED_AND {   return "&"; }      
    if Type == ABC_OPER_RED_OR {   return "|"; }      
    if Type == ABC_OPER_RED_XOR {   return "^"; }      
    if Type == ABC_OPER_RED_NAND {   return "~&"; }     
    if Type == ABC_OPER_RED_NOR {   return "~|"; }     
    if Type == ABC_OPER_RED_NXOR {   return "~^"; }     

    if Type == ABC_OPER_LOGIC_NOT {   return "!"; }      
    if Type == ABC_OPER_LOGIC_IMPL {   return "=>"; }     
    if Type == ABC_OPER_LOGIC_AND {   return "&&"; }     
    if Type == ABC_OPER_LOGIC_OR {   return "||"; }     
    if Type == ABC_OPER_LOGIC_XOR {   return "^^"; }   

    if Type == ABC_OPER_ARI_ADD {   return "+"; }      
    if Type == ABC_OPER_ARI_SUB {   return "-"; }      
    if Type == ABC_OPER_ARI_MUL {   return "*"; }      
    if Type == ABC_OPER_ARI_SMUL {   return "*"; }      
    if Type == ABC_OPER_ARI_DIV {   return "/"; }      
    if Type == ABC_OPER_ARI_REM {   return "%"; }      
    if Type == ABC_OPER_ARI_MOD {   return "mod"; }    
    if Type == ABC_OPER_ARI_POW {   return "**"; }     
    if Type == ABC_OPER_ARI_MIN {   return "-"; }      
    if Type == ABC_OPER_ARI_SQRT {   return "sqrt"; }   
    if Type == ABC_OPER_ARI_SQUARE {   return "squar"; }  
    
    if Type == ABC_OPER_COMP_EQU {   return "=="; }     
    if Type == ABC_OPER_COMP_NOTEQU {   return "!="; }     
    if Type == ABC_OPER_COMP_LESS {   return "<"; }      
    if Type == ABC_OPER_COMP_MORE {   return ">"; }      
    if Type == ABC_OPER_COMP_LESSEQU {   return "<="; }     
    if Type == ABC_OPER_COMP_MOREEQU {   return ">="; }     

    if Type == ABC_OPER_SHIFT_L {   return "<<"; }     
    if Type == ABC_OPER_SHIFT_R {   return ">>"; }     
    if Type == ABC_OPER_SHIFT_LA {   return "<<<"; }    
    if Type == ABC_OPER_SHIFT_RA {   return ">>>"; }    
    if Type == ABC_OPER_SHIFT_ROTL {   return "rotL"; }   
    if Type == ABC_OPER_SHIFT_ROTR {   return "rotR"; }   

    if Type == ABC_OPER_DFFRSE {   return "DFFRSE"; }    

    if Type == ABC_OPER_SLICE {   return "[:]"; }    
    if Type == ABC_OPER_CONCAT {   return "{}"; }     
    if Type == ABC_OPER_ZEROPAD {   return "zPad"; }   
    if Type == ABC_OPER_SIGNEXT {   return "sExt"; }   

    if Type == ABC_OPER_CONST {   return "const"; }  
    if Type == ABC_OPER_TABLE {   return "table"; }  
    if Type == ABC_OPER_LUT {   return "lut"; }  
    if Type == ABC_OPER_LAST {   return nil; }     
    assert(false, "0");
    return nil;
}

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

