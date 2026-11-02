package abc

// Hand-checked skeleton of src/opt/dar/darPrec.c.

// Dar_ArrayAlloc (returns a 2-D char array)
// C: char ** Dar_ArrayAlloc( int nCols, int nRows, int Size )
func DarArrayAlloc(nCols int, nRows int, Size int) [][]int8 {
	panic("TODO, stubbed")
}

// Dar_Factorial
// C: int Dar_Factorial( int n )
func DarFactorial(n int) int {
	panic("TODO, stubbed")
}

// Dar_Permutations_rec
// C: void Dar_Permutations_rec( char ** pRes, int nFact, int n, char Array[] )
func DarPermutations_rec(pRes [][]int8, nFact int, n int, Array []int8) {
	panic("TODO, stubbed")
}

// Dar_Permutations (returns a 2-D char array)
// C: char ** Dar_Permutations( int n )
func DarPermutations(n int) [][]int8 {
	panic("TODO, stubbed")
}

// Dar_TruthPermute_int
// C: void Dar_TruthPermute_int( int * pMints, int nMints, char * pPerm, int nVars, int * pMintsP )
func DarTruthPermute_int(pMints []int, nMints int, pPerm []int8, nVars int, pMintsP []int) {
	panic("TODO, stubbed")
}

// Dar_TruthPermute
// C: unsigned Dar_TruthPermute( unsigned Truth, char * pPerms, int nVars, int fReverse )
func DarTruthPermute(Truth uint32, pPerms []int8, nVars int, fReverse int) uint32 {
	panic("TODO, stubbed")
}

// Dar_TruthPolarize
// C: unsigned Dar_TruthPolarize( unsigned uTruth, int Polarity, int nVars )
func DarTruthPolarize(uTruth uint32, Polarity int, nVars int) uint32 {
	panic("TODO, stubbed")
}

// Dar_Truth4VarNPN (the four parameters are out parameters; the char ** ones are filled with 2-D char arrays)
// C: void Dar_Truth4VarNPN( unsigned short ** puCanons, char ** puPhases, char ** puPerms, unsigned char ** puMap )
func DarTruth4VarNPN(puCanons *[]uint16, puPhases *[][]int8, puPerms *[][]int8, puMap *[]uint8) {
	panic("TODO, stubbed")
}
