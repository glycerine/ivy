// Port of verStream.c by Alan Mishchenko (UC Berkeley, 2006).
package ver

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	VER_BUFFER_SIZE = 1048576
	VER_OFFSET_SIZE = 65536
	VER_WORD_SIZE   = 65536
)

type Ver_Stream_t_ struct {
	pFileName            string
	pFile                *os.File
	nFileSize, nFileRead int64
	nLineCounter         int32
	buffer               []byte
	cur, end, stop       int
	chars                [VER_WORD_SIZE + 5]byte
	nChars               int
	fStop                bool
}

func Ver_StreamAlloc(name string) *Ver_Stream_t {
	file, err := os.Open(name)
	if err != nil {
		fmt.Printf("Ver_StreamAlloc(): Cannot open input file %q.\n", name)
		return nil
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil
	}
	p := &Ver_Stream_t_{pFileName: name, pFile: file, nFileSize: info.Size(), nLineCounter: 1, buffer: make([]byte, VER_BUFFER_SIZE+1)}
	p.end, _ = io.ReadFull(file, p.buffer[:min(int64(VER_BUFFER_SIZE), p.nFileSize)])
	p.nFileRead = int64(p.end)
	p.updateStop()
	return p
}
func (p *Ver_Stream_t_) updateStop() {
	if p.nFileRead == p.nFileSize {
		p.stop = p.end
	} else {
		p.stop = VER_BUFFER_SIZE - VER_OFFSET_SIZE
	}
}
func Ver_StreamReload(p *Ver_Stream_t) {
	if p.fStop || p.cur <= p.stop || p.cur >= p.end {
		panic("ver: invalid stream reload")
	}
	remaining := copy(p.buffer, p.buffer[p.cur:p.end])
	p.cur = 0
	want := min(int64(VER_BUFFER_SIZE-remaining), p.nFileSize-p.nFileRead)
	n, _ := io.ReadFull(p.pFile, p.buffer[remaining:remaining+int(want)])
	p.nFileRead += int64(n)
	p.end = remaining + n
	p.updateStop()
}
func Ver_StreamFree(p *Ver_Stream_t) {
	if p != nil && p.pFile != nil {
		p.pFile.Close()
		p.pFile = nil
	}
	if p != nil {
		p.buffer = nil
	}
}
func Ver_StreamGetFileName(p *Ver_Stream_t) string   { return p.pFileName }
func Ver_StreamGetFileSize(p *Ver_Stream_t) int32    { return int32(p.nFileSize) }
func Ver_StreamGetCurPosition(p *Ver_Stream_t) int32 { return int32(p.nFileRead - int64(p.end-p.cur)) }
func Ver_StreamGetLineNumber(p *Ver_Stream_t) int32  { return p.nLineCounter }
func Ver_StreamIsOkey(p *Ver_Stream_t) int32 {
	if p.fStop {
		return 0
	}
	return 1
}
func Ver_StreamScanChar(p *Ver_Stream_t) byte {
	if p.fStop {
		panic("ver: stopped stream")
	}
	return p.buffer[p.cur]
}
func Ver_StreamPopChar(p *Ver_Stream_t) byte {
	if p.fStop {
		panic("ver: stopped stream")
	}
	if p.cur > p.stop {
		Ver_StreamReload(p)
	}
	if p.cur == p.end {
		p.fStop = true
		return 0xff
	}
	c := p.buffer[p.cur]
	p.cur++
	if c == '\n' {
		p.nLineCounter++
	}
	return c
}
func Ver_StreamSkipChars(p *Ver_Stream_t, chars string)   { p.skip(chars, true) }
func Ver_StreamSkipToChars(p *Ver_Stream_t, chars string) { p.skip(chars, false) }
func (p *Ver_Stream_t_) skip(chars string, skipMembers bool) {
	if p.fStop {
		panic("ver: stopped stream")
	}
	if p.cur > p.stop {
		Ver_StreamReload(p)
	}
	for pos := p.cur; pos < p.end; pos++ {
		if strings.ContainsRune(chars, rune(p.buffer[pos])) != skipMembers {
			p.cur = pos
			return
		}
		if p.buffer[pos] == '\n' {
			p.nLineCounter++
		}
	}
	if p.stop == p.end {
		p.fStop = true
		return
	}
	if skipMembers {
		fmt.Printf("Ver_StreamSkipSymbol() failed to parse the file %q.\n", p.pFileName)
	} else {
		fmt.Printf("Ver_StreamSkipToSymbol() failed to parse the file %q.\n", p.pFileName)
	}
}
func Ver_StreamGetWord(p *Ver_Stream_t, stops string) string {
	if p.fStop {
		return ""
	}
	if p.cur > p.stop {
		Ver_StreamReload(p)
	}
	p.nChars = 0
	for pos := p.cur; pos < p.end; pos++ {
		c := p.buffer[pos]
		if strings.ContainsRune(stops, rune(c)) {
			p.cur = pos
			return string(p.chars[:p.nChars])
		}
		p.chars[p.nChars] = c
		p.nChars++
		if p.nChars == VER_WORD_SIZE {
			fmt.Println("Ver_StreamGetWord(): The buffer size is exceeded.")
			return ""
		}
		if c == '\n' {
			p.nLineCounter++
		}
	}
	if p.stop == p.end {
		p.fStop = true
		return string(p.chars[:p.nChars])
	}
	fmt.Printf("Ver_StreamGetWord() failed to parse the file %q.\n", p.pFileName)
	return ""
}
func Ver_StreamMove(p *Ver_Stream_t) {
	remaining := string(p.buffer[p.cur+1 : p.end])
	if strings.HasPrefix(remaining, "z_g_") || strings.HasPrefix(remaining, "co_") {
		for p.cur < p.end && p.buffer[p.cur] != '(' {
			p.cur++
		}
	}
}
