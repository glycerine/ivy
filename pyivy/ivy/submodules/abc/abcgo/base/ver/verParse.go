package ver

/**CFile****************************************************************

  FileName    [verParse.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Performs some Verilog parsing tasks.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 19, 2006.]

  Revision    [$Id: verParse.c,v 1.00 2006/08/19 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Skips the comments of they are present.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseSkipComments(pMan *Ver_Man_t) int32 {
    var p *Ver_Stream_t  = pMan.pReader;
    var Symbol byte;
    // skip spaces
    Ver_StreamSkipChars( p, " \t\n\r" );
    if !Ver_StreamIsOkey(pMan.pReader) {
        return 1; }
    // read the first symbol
    Symbol = Ver_StreamScanChar( p );
    if Symbol != '/' {
        return 1; }
    Ver_StreamPopChar( p );
    // read the second symbol
    Symbol = Ver_StreamScanChar( p );
    if Symbol == '/' { // skip till the end of line
        Ver_StreamSkipToChars( p, "\n" );
        return Ver_ParseSkipComments( pMan );
    }
    if Symbol == '*' { // skip till the next occurrence of */
        Ver_StreamPopChar( p );
        for {
            Ver_StreamSkipToChars( p, "*" );
            Ver_StreamPopChar( p );
        
if !( Ver_StreamScanChar( p ) != '/' ) { break }
} 
        Ver_StreamPopChar( p );
        return Ver_ParseSkipComments( pMan );
    }
    sprintf( pMan.sError, "Cannot parse after symbol \"/\"." );
    Ver_ParsePrintErrorMessage( pMan );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Parses a Verilog name that can be being with a slash.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseGetName(pMan *Ver_Man_t) string {
    var p *Ver_Stream_t  = pMan.pReader;
    var Symbol byte;
    var pWord string;
    pMan.fNameLast = 0;
    if !Ver_StreamIsOkey(p) {
        return nil; }
    if !(Ver_ParseSkipComments( pMan ) != 0) {
        return nil; }
    Symbol = Ver_StreamScanChar( p );
    if Symbol == '\\' {
        pMan.fNameLast = 1;
        Ver_StreamPopChar( p );
        pWord = Ver_StreamGetWord( p, " \r\n" );
        Ver_StreamSkipChars( p, " \r\n" );
        if Ver_StreamScanChar(p) == '[' {
            var This byte;
            var pEnd string = pWord + len(pWord);
            for  (This = Ver_StreamPopChar(p)) != ']'  {
                *pEnd++ = This;
            }
            *pEnd++ = This;
            *pEnd = 0;
        }
    
    } else {
        pWord = Ver_StreamGetWord( p, " \t\n\r(),;" ); }
    if Ver_StreamIsOkey(p) && !(Ver_ParseSkipComments( pMan ) != 0) {
        return nil; }
    return pWord;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

