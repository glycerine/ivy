#include "aig/aig/aig.h"

/* Capture the source graph after its real movement and sequential cleanup.
   The original final DupOrdered assumes increasing topological object IDs;
   frontier rewrites introduce forward buffer references, violating that
   precondition even for the basic two-register AND example. */
static Aig_Man_t *capture_frontier(Aig_Man_t *g) { return g; }
#define Aig_ManDupOrdered capture_frontier
#include "aig/aig/aigRetF.c"
#undef Aig_ManDupOrdered

#define RETIMING_FRONTIER
#include "retiming_driver_compare.c"
