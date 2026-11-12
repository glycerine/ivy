package bsat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [satStore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Records the trace of SAT solving in the CNF form.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satStore.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

func Sto_ManAlloc() *Sto_Man_t { return &Sto_Man_t{nChunkSize: 1 << 16} }
func Sto_ManMemoryFetch(p *Sto_Man_t, n int32) []byte {
	if n < 0 {
		panic("negative proof storage request")
	}
	if n > p.nChunkSize-p.nChunkUsed || p.pChunkLast == nil {
		size := max(p.nChunkSize, n)
		p.pChunkLast = make([]byte, size)
		p.chunks = append(p.chunks, p.pChunkLast)
		p.nChunkUsed = 0
	}
	start := p.nChunkUsed
	p.nChunkUsed += n
	return p.pChunkLast[start:p.nChunkUsed]
}
func Sto_ManMemoryStop(p *Sto_Man_t) {
	p.chunks = nil
	p.pChunkLast = nil
	p.nChunkUsed = 0
	p.pHead = nil
	p.pTail = nil
	p.pEmpty = nil
	p.nVars = 0
	p.nClauses = 0
	p.nRoots = 0
	p.nClausesA = 0
}
func Sto_ManFree(p *Sto_Man_t) {
	if p != nil {
		*p = Sto_Man_t{}
	}
}

// MemoryReport estimates owned payload from actual Go type sizes, not C layout.
func Sto_ManMemoryReport(p *Sto_Man_t) int32 {
	total := int64(reflect.TypeOf(*p).Size())
	for _, chunk := range p.chunks {
		total += int64(cap(chunk))
	}
	size := int64(reflect.TypeOf(Sto_Cls_t{}).Size())
	for c := p.pHead; c != nil; c = c.pNext {
		total += size + int64(cap(c.pLits))*4
	}
	if total > 1<<31-1 {
		return 1<<31 - 1
	}
	return int32(total)
}
func Sto_ManAddClause(p *Sto_Man_t, data []lit) int32 {
	if len(data) > 0xffffff || len(data) == 0 && p.pEmpty != nil {
		return 0
	}
	literals := append([]lit(nil), data...)
	slices.Sort(literals)
	for i, l := range literals {
		if l < 0 || i > 0 && lit_var(literals[i-1]) == lit_var(l) {
			return 0
		}
	}
	c := &Sto_Cls_t{Id: p.nClauses, nLits: int32(len(literals)), pLits: literals}
	p.nClauses++
	if len(literals) > 0 {
		p.nVars = max(p.nVars, lit_var(literals[len(literals)-1])+1)
	} else {
		p.pEmpty = c
	}
	if p.pHead == nil {
		p.pHead = c
	} else {
		p.pTail.pNext = c
	}
	p.pTail = c
	return 1
}
func Sto_ManMarkRoots(p *Sto_Man_t) {
	p.nRoots = 0
	for c := p.pHead; c != nil; c = c.pNext {
		c.fRoot = 1
		p.nRoots++
	}
}
func Sto_ManMarkClausesA(p *Sto_Man_t) {
	p.nClausesA = 0
	for c := p.pHead; c != nil; c = c.pNext {
		c.fA = 1
		p.nClausesA++
	}
}
func Sto_ManChangeLastClause(p *Sto_Man_t) int32 {
	c := p.pTail
	if c == nil || c.fA != 1 || c.nLits != 1 {
		panic("last A-clause must be a unit")
	}
	c.fA = 0
	p.nClausesA--
	return lit_var(c.pLits[0])
}
func WriteClauseStore(w io.Writer, p *Sto_Man_t) error {
	if _, err := fmt.Fprintf(w, "p %d %d %d %d\n", p.nVars, p.nClauses, p.nRoots, p.nClausesA); err != nil {
		return err
	}
	for c := p.pHead; c != nil; c = c.pNext {
		for _, l := range c.pLits {
			if _, err := fmt.Fprintf(w, " %d", lit_print(l)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, " 0"); err != nil {
			return err
		}
	}
	return nil
}
func Sto_ManDumpClauses(p *Sto_Man_t, name string) {
	file, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = WriteClauseStore(file, p)
	closeErr := file.Close()
	if err != nil {
		fmt.Println(err)
	} else if closeErr != nil {
		fmt.Println(closeErr)
	}
}

// readStoreNumber consumes delimiters without reading past the token. It also
// accepts DIMACS comment lines and EOF immediately after a number.
func readStoreNumber(r io.ByteReader) (int32, error) {
	var token []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			if len(token) > 0 && errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}
		if len(token) == 0 && b == 'c' {
			for b != '\n' {
				b, err = r.ReadByte()
				if err != nil {
					return 0, err
				}
			}
			continue
		}
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			if len(token) == 0 {
				continue
			}
			break
		}
		token = append(token, b)
		if len(token) > 12 {
			return 0, errors.New("proof literal exceeds signed integer range")
		}
	}
	n, err := strconv.ParseInt(string(token), 10, 32)
	return int32(n), err
}

type storeFileBytes struct{ file *os.File }

func (r storeFileBytes) ReadByte() (byte, error) {
	var b [1]byte
	_, err := io.ReadFull(r.file, b[:])
	return b[0], err
}
func Sto_ManLoadNumber(file *os.File, value *int32) int32 {
	n, err := readStoreNumber(storeFileBytes{file})
	if err != nil {
		return 0
	}
	*value = n
	return 1
}
func ReadClauseStore(input io.Reader) (*Sto_Man_t, error) {
	r := bufio.NewReader(input)
	var vars, count, roots, aCount int32
	found := false
	for {
		line, err := r.ReadString('\n')
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "p ") || strings.HasPrefix(trimmed, "p\t") {
			fields := strings.Fields(trimmed)
			if len(fields) != 5 {
				return nil, errors.New("invalid clause-store header")
			}
			values := []*int32{&vars, &count, &roots, &aCount}
			for i, dst := range values {
				n, e := strconv.ParseInt(fields[i+1], 10, 32)
				if e != nil || n < 0 {
					return nil, errors.New("invalid clause-store header count")
				}
				*dst = int32(n)
			}
			found = true
			break
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	if !found || roots > count || aCount > count {
		return nil, errors.New("missing or inconsistent clause-store header")
	}
	p := Sto_ManAlloc()
	var literals []lit
	for {
		n, err := readStoreNumber(r)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if n == 0 {
			if p.nClauses >= count || Sto_ManAddClause(p, literals) == 0 {
				return nil, errors.New("invalid or excess stored clause")
			}
			c := p.pTail
			if c.Id < roots {
				c.fRoot = 1
			}
			if c.Id < aCount {
				c.fA = 1
			}
			literals = nil
			continue
		}
		if n < -0x40000000 || n > 0x40000000 {
			return nil, errors.New("literal exceeds SAT encoding")
		}
		literals = append(literals, lit_read(n))
	}
	if len(literals) != 0 || p.nClauses != count || p.nVars > vars {
		return nil, errors.New("clause data does not match header")
	}
	p.nVars = vars
	p.nRoots = roots
	p.nClausesA = aCount
	return p, nil
}
func Sto_ManLoadClauses(name string) *Sto_Man_t {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()
	p, err := ReadClauseStore(file)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return p
}
