package extrab

import "github.com/glycerine/abcgo/bdd/extrab/primes"

// Extra_zddPrimes uses the authoritative, independently importable prime cover
// recursion while the remaining root extrab source families are being ported.
func Extra_zddPrimes(dd *DdManager, f *DdNode) *DdNode { return primes.Extra_zddPrimes(dd, f) }
func extraZddPrimes(dd *DdManager, f *DdNode) *DdNode  { return Extra_zddPrimes(dd, f) }
