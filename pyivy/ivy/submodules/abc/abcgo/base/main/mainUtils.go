package abcmain

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Abc_UtilsGetVersion(f *Frame) string {
	if f != nil && f.Version != "" {
		return f.Version
	}
	return ABC_VERSION
}
func (f *Frame) ReadLine() (string, error) {
	f.inputMu.Lock()
	defer f.inputMu.Unlock()
	if f.inputReader == nil || !sameResource(f.inputSource, f.In) {
		f.inputSource = f.In
		f.inputReader = bufio.NewReader(f.In)
	}
	line, err := f.inputReader.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}
func Abc_UtilsGetUsersInput(f *Frame) string {
	f.mu.RLock()
	step := f.steps
	f.mu.RUnlock()
	fmt.Fprintf(f.Out, "abc %02d> ", step)
	line, err := f.ReadLine()
	if err != nil && err != io.EOF {
		fmt.Fprintln(f.Err, err)
	}
	return line
}
func Abc_UtilsPrintHello(f *Frame) { fmt.Fprintln(f.Out, f.Version) }
func Abc_UtilsPrintUsage(f *Frame, program string) {
	fmt.Fprintf(f.Err, "usage: %s [-c cmd] [-q cmd] [-C cmd] [-Q cmd] [-f script] [-F script] [-h] [-o file] [-s] [-t type] [-T type] [-x] [-b] [file]\n", program)
	fmt.Fprintln(f.Err, "  -c/-q execute commands and quit; -C/-Q continue interactively\n  -f/-F source a script (with -F echo); -s skip initialization files\n  -t/-T select input/output formats; -x skip read/write; -b bridge mode")
}

// SourceResourceFiles uses the home resource file in preference to a duplicate
// current-directory file, matching ABC's startup policy. No shell is invoked.
func (f *Frame) SourceResourceFiles() int32 {
	home, _ := os.UserHomeDir()
	var chosen string
	if home != "" {
		path := filepath.Join(home, ".abc.rc")
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			chosen = path
		}
	}
	if chosen == "" {
		if info, err := os.Stat(".abc.rc"); err == nil && !info.IsDir() {
			chosen = ".abc.rc"
		}
	}
	if chosen != "" {
		if status := f.Execute("source -s " + strconvQuote(chosen)); status != 0 {
			return status
		}
	}
	return f.Execute("source -s abc.rc")
}
func Abc_UtilsSource(f *Frame) { f.SourceResourceFiles() }
func DateReadFromDateString(date string) string {
	parsed, err := time.Parse("Mon Jan 2 15:04:05 MST 2006", strings.TrimSpace(date))
	if err != nil {
		return date
	}
	hour := parsed.Hour() % 12
	if hour == 0 {
		hour = 12
	}
	period := "AM"
	if parsed.Hour() >= 12 {
		period = "PM"
	}
	return fmt.Sprintf("%d-%s-%02d at %d:%02d %s", parsed.Day(), parsed.Format("Jan"), parsed.Year()%100, hour, parsed.Minute(), period)
}
func TypeCheck(f *Frame, format string) int32 {
	for _, allowed := range []string{"blif_mv", "blif_mvs", "blif", "none", "aiger", "aig", "bench", "pla", "verilog"} {
		if format == allowed {
			return 1
		}
	}
	fmt.Fprintf(f.Err, "Unknown file format: %s\n", format)
	return 0
}
