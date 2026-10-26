package abc

// Compressed text output from sat/cnf/cnfMan.c:Cnf_DataWriteIntoFileGz.

import "compress/gzip"

type CNFGZIPOptions struct {
	DIMACS   CNFDIMACSOptions
	MaxBytes int // compressed byte budget; zero: 64 MiB
}
type cnfGZIPBuffer struct {
	bytes []byte
	limit int
	stop  func() bool
}

func (b *cnfGZIPBuffer) Write(p []byte) (int, error) {
	if b.stop != nil && b.stop() {
		return 0, ErrCancelled
	}
	if len(p) > b.limit-len(b.bytes) {
		return 0, ErrLimit
	}
	b.bytes = append(b.bytes, p...)
	return len(p), nil
}

// DIMACSGZIP returns a deterministic gzip stream containing native-format text.
// Text and compressed budgets are separate. Compression bytes may differ from
// native zlib; decompressed text is identical. No partial bytes escape on error.
func (f *CNFFormula) DIMACSGZIP(o CNFGZIPOptions) ([]byte, error) {
	if o.MaxBytes < 0 {
		return nil, ErrDimension
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 64 << 20
	}
	plain, e := f.DIMACS(o.DIMACS)
	if e != nil {
		return nil, e
	}
	buffer := cnfGZIPBuffer{limit: o.MaxBytes, stop: o.DIMACS.Limits.Stop}
	writer := gzip.NewWriter(&buffer)
	for offset := 0; offset < len(plain); {
		if buffer.stop != nil && buffer.stop() {
			writer.Close()
			return nil, ErrCancelled
		}
		end := min(len(plain), offset+(32<<10))
		if _, e = writer.Write(plain[offset:end]); e != nil {
			writer.Close()
			return nil, e
		}
		offset = end
	}
	if e = writer.Close(); e != nil {
		return nil, e
	}
	if buffer.stop != nil && buffer.stop() {
		return nil, ErrCancelled
	}
	return buffer.bytes, nil
}
func (c *AigMappedCNF) DIMACSGZIP(o CNFGZIPOptions) ([]byte, error) {
	if c == nil {
		return nil, ErrDimension
	}
	return (&CNFFormula{variables: c.variables, clauses: c.clauses}).DIMACSGZIP(o)
}
func (f *AigCNFFrames) DIMACSGZIP(o CNFGZIPOptions) ([]byte, error) {
	if f == nil {
		return nil, ErrDimension
	}
	return (&CNFFormula{variables: f.variables, clauses: f.clauses}).DIMACSGZIP(o)
}
