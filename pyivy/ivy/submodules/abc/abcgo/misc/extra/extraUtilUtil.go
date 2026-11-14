package extra

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

/**CFile****************************************************************

  FileName    [extraUtilUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Old SIS utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilUtil.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

const EXTRA_RLIMIT_DATA_DEFAULT int32 = 67108864

func Extra_GetSoftDataLimit() int32 { return EXTRA_RLIMIT_DATA_DEFAULT }

// Extra_OptionParser is an independent instance of ABC's short-option parser.
// Index counts argv entries including argv[0], and Arg is reset on every Next.
type Extra_OptionParser struct {
	Index int
	Arg   string
	scan  string
}

func (p *Extra_OptionParser) Reset() { *p = Extra_OptionParser{} }
func (p *Extra_OptionParser) Next(argv []string, spec string) int32 {
	p.Arg = ""
	if p.Index < 0 {
		return -1
	}
	if p.scan == "" {
		if p.Index == 0 {
			p.Index = 1
		}
		if p.Index >= len(argv) {
			return -1
		}
		place := argv[p.Index]
		if len(place) < 2 || place[0] != '-' {
			return -1
		}
		p.Index++
		if place == "--" {
			return -1
		}
		p.scan = place[1:]
	}
	ch := p.scan[0]
	p.scan = p.scan[1:]
	position := strings.IndexByte(spec, ch)
	program := ""
	if len(argv) > 0 {
		program = argv[0]
	}
	if position < 0 || ch == ':' {
		fmt.Fprintf(os.Stderr, "%s: unknown option %c\n", program, ch)
		return '?'
	}
	if position+1 < len(spec) && spec[position+1] == ':' {
		if p.scan != "" {
			p.Arg = p.scan
			p.scan = ""
		} else {
			if p.Index >= len(argv) {
				fmt.Fprintf(os.Stderr, "%s: %c requires an argument\n", program, ch)
				return '?'
			}
			p.Arg = argv[p.Index]
			p.Index++
		}
	}
	return int32(ch)
}

// Legacy global access remains for command ports. Independent callers should
// use Extra_OptionParser instances rather than sharing option state.
var globalUtilOptarg string
var globalUtilOptind int32
var legacyOptionParser Extra_OptionParser
var legacyOptionMu sync.Mutex

func Extra_UtilGetoptReset() {
	legacyOptionMu.Lock()
	defer legacyOptionMu.Unlock()
	legacyOptionParser.Reset()
	globalUtilOptarg = ""
	globalUtilOptind = 0
}
func Extra_UtilGetopt(argc int32, argv []string, spec string) int32 {
	legacyOptionMu.Lock()
	defer legacyOptionMu.Unlock()
	if argc < 0 || int64(argc) > int64(len(argv)) {
		return -1
	}
	legacyOptionParser.Index = int(globalUtilOptind)
	result := legacyOptionParser.Next(argv[:argc], spec)
	globalUtilOptarg = legacyOptionParser.Arg
	globalUtilOptind = int32(legacyOptionParser.Index)
	return result
}
func Extra_UtilGetoptArg() string {
	legacyOptionMu.Lock()
	defer legacyOptionMu.Unlock()
	return globalUtilOptarg
}
func Extra_UtilGetoptIndex() int32 {
	legacyOptionMu.Lock()
	defer legacyOptionMu.Unlock()
	return globalUtilOptind
}
func Extra_UtilGetoptSetIndex(index int32) {
	legacyOptionMu.Lock()
	defer legacyOptionMu.Unlock()
	globalUtilOptind = index
	legacyOptionParser.Index = int(index)
}
func Extra_UtilPrintTime(milliseconds int64) string {
	return fmt.Sprintf("%d.%02d sec", milliseconds/1000, milliseconds%1000/10)
}
func Extra_UtilStrsav(s string) string { return s }

// Tilde expansion is deliberately disabled in the original C implementation.
func Extra_UtilTildeExpand(name string) string { return name }
func Extra_UtilCheckFile(name, mode string) int32 {
	if mode == "x" {
		mode = "r"
	}
	if mode == "" {
		return 0
	}
	flags := 0
	switch mode[0] {
	case 'r':
		flags = os.O_RDONLY
	case 'w':
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case 'a':
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	default:
		return 0
	}
	if strings.Contains(mode, "+") {
		flags &^= os.O_WRONLY
		flags |= os.O_RDWR
	}
	if mode[0] == 'w' && strings.Contains(mode, "x") {
		flags |= os.O_EXCL
	}
	file, err := os.OpenFile(name, flags, 0666)
	if err != nil {
		return 0
	}
	if err := file.Close(); err != nil {
		return 0
	}
	return 1
}
func Extra_UtilFileSearch(file, path, mode string) string {
	if path == "" {
		path = "."
	}
	for _, dir := range strings.Split(path, ":") {
		candidate := file
		if dir != "." {
			candidate = filepath.Join(dir, file)
		}
		candidate = Extra_UtilTildeExpand(candidate)
		if Extra_UtilCheckFile(candidate, mode) != 0 {
			return candidate
		}
	}
	return ""
}
func Extra_UtilMMout_Of_Memory(size int64) {
	panic(fmt.Sprintf("out of memory allocating %d bytes", size))
}

var Extra_UtilMMoutOfMemory = Extra_UtilMMout_Of_Memory

// The Go clock policy matches util.Abc_Clock: monotonic elapsed nanoseconds,
// not platform-specific user CPU time. Both functions share the same origin.
func Extra_CpuTime() int64         { return util.Abc_Clock() }
func Extra_CpuTimeDouble() float64 { return float64(util.Abc_Clock()) / float64(util.CLOCKS_PER_SEC) }
func Extra_MemTest()               {} // The corresponding C diagnostic has no active statements.
