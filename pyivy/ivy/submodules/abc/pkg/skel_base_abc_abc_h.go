package abc

// Hand-written skeleton of types in src/base/abc/abc.h that have no literal
// Go counterpart (declarations only).

// AbcDes is Abc_Des_t: a design (library of modules).
type AbcDes struct {
	pName    string                            // char * : the name of the library
	pManFunc any                               // void * : functionality manager for the nodes
	vTops    *PointerVector[ABCNetwork]        // the array of top-level modules
	vModules *PointerVector[ABCNetwork]        // the array of modules
	tModules *StableTable[string, *ABCNetwork] // st__table * : the table hashing module names into their networks
	pLibrary *AbcDes                           // the library used to map this design
	pGenlib  *MioLibrary                       // void * : the genlib library used to map this design
}
