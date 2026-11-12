package extra

/**CFile****************************************************************

  FileName    [extraUtilCfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [CF simulation.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilCfs.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

// Extra_CommandCfs is intentionally empty: the original extraUtilCfs.c command
// has no executable statements. graph is dependency-neutral to avoid a cycle
// with Gia; no simulation behavior is claimed for this source-empty command.
func Extra_CommandCfs(graph any, limit, reps, unseenUse, rareUse, replaceRare, constSim, dagNodes int32, flipProbability float32, verbose int32) {
}
