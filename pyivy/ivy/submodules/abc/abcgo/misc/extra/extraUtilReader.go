package extra

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Extra_FileReader_t = Extra_FileReader_t_

/**CFile****************************************************************

  FileName    [extraUtilReader.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [File reading utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilReader.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// The streaming reader owns token strings, so callers may retain them across
// subsequent reads without depending on a reused C input buffer.
type Extra_FileReader_t_ struct {
	name           string
	file           *os.File
	input          *bufio.Reader
	size, position int64
	line           int32
	kinds          [256]byte
	tokens         []string
	lines          []int32
	stopped        bool
	err            error
}

const (
	EXTRA_CHAR_COMMENT byte = iota
	EXTRA_CHAR_NORMAL
	EXTRA_CHAR_STOP
	EXTRA_CHAR_CLEAN
)

func Extra_FileReaderAlloc(name, comments, stops, clean string) *Extra_FileReader_t_ {
	file, err := os.Open(name)
	if err != nil {
		fmt.Printf("Extra_FileReaderAlloc(): Cannot open input file %q.\n", name)
		return nil
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil
	}
	p := &Extra_FileReader_t_{name: name, file: file, input: bufio.NewReader(file), size: info.Size(), line: 1}
	for i := range p.kinds {
		p.kinds[i] = EXTRA_CHAR_NORMAL
	}
	for _, group := range []struct {
		text string
		kind byte
	}{{comments, EXTRA_CHAR_COMMENT}, {stops, EXTRA_CHAR_STOP}, {clean, EXTRA_CHAR_CLEAN}} {
		for i := 0; i < len(group.text); i++ {
			p.kinds[group.text[i]] = group.kind
		}
	}
	return p
}
func Extra_FileReaderFree(p *Extra_FileReader_t_) {
	if p != nil {
		if p.file != nil {
			if err := p.file.Close(); err != nil && p.err == nil {
				p.err = err
			}
		}
		p.file = nil
		p.input = nil
		p.tokens = nil
		p.lines = nil
		p.stopped = true
	}
}
func Extra_FileReaderGetFileName(p *Extra_FileReader_t_) string               { return p.name }
func Extra_FileReaderGetFileSize(p *Extra_FileReader_t_) int64                { return p.size }
func Extra_FileReaderGetCurPosition(p *Extra_FileReader_t_) int64             { return p.position }
func Extra_FileReaderGetLineNumber(p *Extra_FileReader_t_, token int32) int32 { return p.lines[token] }
func Extra_FileReaderErr(p *Extra_FileReader_t_) error                        { return p.err }
func Extra_FileReaderGetTokens(p *Extra_FileReader_t_) []string {
	for {
		tokens := Extra_FileReaderGetTokens_int(p)
		if tokens == nil || len(tokens) > 0 {
			return tokens
		}
	}
}
func Extra_FileReaderGetTokens_int(p *Extra_FileReader_t_) []string {
	if p.stopped {
		return nil
	}
	p.tokens = make([]string, 0)
	p.lines = make([]int32, 0)
	var token strings.Builder
	var tokenLine int32
	started := false
	finish := func() {
		if started {
			p.tokens = append(p.tokens, token.String())
			p.lines = append(p.lines, tokenLine)
			token.Reset()
			started = false
		}
	}
	for {
		ch, err := p.input.ReadByte()
		if err != nil {
			finish()
			p.stopped = true
			if err != io.EOF {
				p.err = err
			}
			return p.tokens
		}
		p.position++
		if ch == '\n' {
			p.line++
		}
		kind := p.kinds[ch]
		if kind == EXTRA_CHAR_COMMENT {
			isComment := ch != '/'
			if !isComment {
				next, _ := p.input.Peek(1)
				isComment = len(next) > 0 && next[0] == '/'
			}
			if isComment {
				finish()
				for {
					next, err := p.input.Peek(1)
					if err != nil || next[0] == '\n' {
						break
					}
					p.input.ReadByte()
					p.position++
				}
				continue
			}
			kind = EXTRA_CHAR_NORMAL
		}
		switch kind {
		case EXTRA_CHAR_STOP:
			finish()
			return p.tokens
		case EXTRA_CHAR_CLEAN:
			finish()
		case EXTRA_CHAR_NORMAL:
			if !started {
				started = true
				tokenLine = p.line
			}
			token.WriteByte(ch)
		}
	}
}

// Reload is retained for source compatibility; Peek asks bufio to refill the
// stream without consuming the next token byte.
func Extra_FileReaderReload(p *Extra_FileReader_t_) {
	if !p.stopped {
		_, err := p.input.Peek(1)
		if err != nil && err != io.EOF {
			p.err = err
		}
	}
}
