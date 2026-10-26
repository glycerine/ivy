package abc

// Package abc implements the native Go port of ABC, the logic synthesis and
// formal verification system. The port is in progress; supported operations
// are documented individually and tracked in PORT_STATUS.md.
// Managers and their literals are single-owner mutable values. A literal is
// meaningful only in the manager that created it. Independent managers may
// be used concurrently. AIG constant one is literal 0, unlike AIGER/GIA.
