package extra

/**CFile****************************************************************

  FileName    [extraUtilGen.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [CF simulation.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilGen.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

// Extra_CommandGen preserves the original source's unconditional NULL result.
// The source contains no generation algorithm. Its graph boundary is the same
// dependency-neutral interface used by the ZDD diagnostic, avoiding a Gia cycle.
func Extra_CommandGen(algo, lutSize, nLuts, nLevels, limit, nBestTries, multi, xor, verbose int32) Abc_ZddGiaGraph {
	return nil
}
