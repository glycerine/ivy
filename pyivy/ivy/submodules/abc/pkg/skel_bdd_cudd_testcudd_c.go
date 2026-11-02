package abc

import "os"

// Hand-checked skeleton of src/bdd/cudd/testcudd.c.

// DdManager is mapped to *CuddFoundation (the concrete port of cuddInt.h's
// manager); DdNode is declared literally in skel_bdd_cudd_cudd_h.go.

// main (the only non-static function; renamed because package abc is a library)
// C: int main(int argc, char **argv)
func testcuddMain(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// usage (static)
// C: void usage(char *prog)
func testcuddUsage(prog string) {
	panic("TODO, stubbed")
}

// open_file (static)
// C: FILE * open_file(char *filename, const char *mode)
func testcuddOpenFile(filename string, mode string) *os.File {
	panic("TODO, stubbed")
}

// testIterators (static)
// C: int testIterators( DdManager *dd, DdNode *M, DdNode *C, int pr)
func testIterators(dd *CuddFoundation, M *DdNode, C *DdNode, pr int) int {
	panic("TODO, stubbed")
}

// testXor (static)
// C: int testXor(DdManager *dd, DdNode *f, int pr, int nvars)
func testXor(dd *CuddFoundation, f *DdNode, pr int, nvars int) int {
	panic("TODO, stubbed")
}

// testHamming (static)
// C: int testHamming( DdManager *dd, DdNode *f, int pr)
func testHamming(dd *CuddFoundation, f *DdNode, pr int) int {
	panic("TODO, stubbed")
}

// testWalsh (static)
// C: int testWalsh( DdManager *dd , int N , int cmu , int approach , int pr )
func testWalsh(dd *CuddFoundation, N int, cmu int, approach int, pr int) int {
	panic("TODO, stubbed")
}
