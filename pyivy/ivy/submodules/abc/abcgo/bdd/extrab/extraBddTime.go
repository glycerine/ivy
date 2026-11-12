/**CFile****************************************************************

  FileName    [extraBddTime.c]

  PackageName [extra]

  Synopsis    [Procedures to control runtime in BDD operators.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddTime.c,v 1.0 2003/05/21 18:03:50 alanmi Exp $]

***********************************************************************/

package extrab

import (
	"context"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"time"
)

// C used absolute process CPU-clock deadlines, with zero disabling timeout.
// Go uses absolute time.Time deadlines (including their monotonic component).
// An expired deadline returns nil, including for trivial/cached operations.
// Completed nodes remain canonical and can be reused after cancellation.
func timeContext(deadline time.Time) (context.Context, context.CancelFunc) {
	if deadline.IsZero() {
		return context.Background(), func() {}
	}
	return context.WithDeadline(context.Background(), deadline)
}
func Extra_bddAndTime(dd *DdManager, f, g *DdNode, deadline time.Time) *DdNode {
	ctx, cancel := timeContext(deadline)
	defer cancel()
	return cudd.BDDAndContext(ctx, dd, f, g)
}
func Extra_bddAndAbstractTime(dd *DdManager, f, g, cube *DdNode, deadline time.Time) *DdNode {
	ctx, cancel := timeContext(deadline)
	defer cancel()
	return cudd.BDDAndAbstractContext(ctx, dd, f, g, cube)
}
func Extra_TransferPermuteTime(source, destination *DdManager, f *DdNode, permutation []int, deadline time.Time) *DdNode {
	ctx, cancel := timeContext(deadline)
	defer cancel()
	return cudd.BDDTransferContext(ctx, source, destination, f, permutation)
}
