package mio

// The historical mioForm.c duplicates Mio_LibraryParseFormulas,
// Mio_GateParseFormula, Mio_GateCollectNames and Mio_SopRegister. It is absent
// from ABC's module.make; the active expression-based versions are ported in
// mioFunc.go. Its obsolete CUDD-dependent alternative is not a second Go API.
