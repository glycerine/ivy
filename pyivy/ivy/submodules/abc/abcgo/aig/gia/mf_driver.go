package gia

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"io"
	"os"
	"time"
)

type Cnf_Dat_t = mf.CNF

func Mf_ManGenerateCnfResult(p *Gia_Man_t, lutSize int32, objectIDs, addOr, withMapping bool) (*mf.CNF, error) {
	cnf, err := mf.GenerateCNF(p, int(lutSize), objectIDs, addOr, withMapping)
	if err == nil {
		p.pData = cnf
	}
	return cnf, err
}
func Mf_ManGenerateCnf(p *Gia_Man_t, lutSize, objectIDs, addOr, withMapping, verbose int32) *Cnf_Dat_t {
	cnf, err := Mf_ManGenerateCnfResult(p, lutSize, objectIDs != 0, addOr != 0, withMapping != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	if verbose != 0 {
		fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d.\n", cnf.Variables, len(cnf.Clauses))
	}
	return cnf
}
func giaWriteMappedCNF(writer io.Writer, cnf *mf.CNF) error {
	buffer := bufio.NewWriter(writer)
	if _, err := fmt.Fprintf(buffer, "p cnf %d %d\n", cnf.Variables, len(cnf.Clauses)); err != nil {
		return err
	}
	for _, clause := range cnf.Clauses {
		for _, literal := range clause {
			value := int64(literal>>1) + 1
			if literal&1 != 0 {
				value = -value
			}
			if _, err := fmt.Fprintf(buffer, "%d ", value); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(buffer, "0"); err != nil {
			return err
		}
	}
	return buffer.Flush()
}
func Mf_ManDumpCnf(p *Gia_Man_t, path string, lutSize, objectIDs, addOr, verbose int32) {
	start := time.Now()
	cnf := Mf_ManGenerateCnf(p, lutSize, objectIDs, addOr, 0, verbose)
	if cnf == nil {
		return
	}
	file, err := os.Create(path)
	if err == nil {
		err = giaWriteMappedCNF(file, cnf)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	literals := 0
	for _, clause := range cnf.Clauses {
		literals += len(clause)
	}
	fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. Time = %.2f sec\n", cnf.Variables, len(cnf.Clauses), literals, time.Since(start).Seconds())
	p.pData = nil
}
