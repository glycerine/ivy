package abc

// File output from sat/cnf/cnfMan.c:Cnf_DataWriteIntoFile and Gz variant.

import (
	"os"
	"strings"
)

type CNFFileOptions struct {
	DIMACS             CNFDIMACSOptions
	MaxCompressedBytes int // zero: 64 MiB
}

// WriteDIMACSFile validates and serializes before creating/truncating path.
// A case-sensitive .gz suffix selects gzip, as in native ABC. Serialization
// failures leave an existing destination unchanged. Filesystem errors are
// returned; a failed OS write may leave partial output, as with native fopen.
func (f *CNFFormula) WriteDIMACSFile(path string, o CNFFileOptions) error {
	if o.MaxCompressedBytes < 0 {
		return ErrDimension
	}
	var data []byte
	var e error
	if strings.HasSuffix(path, ".gz") {
		data, e = f.DIMACSGZIP(CNFGZIPOptions{DIMACS: o.DIMACS, MaxBytes: o.MaxCompressedBytes})
	} else {
		data, e = f.DIMACS(o.DIMACS)
	}
	if e != nil {
		return e
	}
	if o.DIMACS.Limits.Stop != nil && o.DIMACS.Limits.Stop() {
		return ErrCancelled
	}
	return os.WriteFile(path, data, 0666)
}
func (c *AigMappedCNF) WriteDIMACSFile(path string, o CNFFileOptions) error {
	if c == nil {
		return ErrDimension
	}
	return (&CNFFormula{variables: c.variables, clauses: c.clauses}).WriteDIMACSFile(path, o)
}
func (f *AigCNFFrames) WriteDIMACSFile(path string, o CNFFileOptions) error {
	if f == nil {
		return ErrDimension
	}
	return (&CNFFormula{variables: f.variables, clauses: f.clauses}).WriteDIMACSFile(path, o)
}
