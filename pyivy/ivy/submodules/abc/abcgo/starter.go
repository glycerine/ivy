package abcgo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MAX_COMM_NUM is the original starter's command limit.
const MAX_COMM_NUM = 1000

// Abc_UtilStrsav returns an independently owned immutable Go string value.
func Abc_UtilStrsav(s string) string { return strings.Clone(s) }

// Abc_RunThread executes a command with the platform shell, as system() did.
// The caller supplies the streams; command failure does not terminate siblings.
func Abc_RunThread(command string, out, errOut io.Writer) error {
	var child *exec.Cmd
	if runtime.GOOS == "windows" {
		child = exec.Command("cmd", "/C", command)
	} else {
		child = exec.Command("/bin/sh", "-c", command)
	}
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, out, errOut
	return child.Run()
}

type starterWriter struct {
	mutex  *sync.Mutex
	writer io.Writer
}

func (w starterWriter) Write(p []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.writer.Write(p)
}

// RunStarter executes eligible lines in parallel. The original reserves one CPU
// for the controller; a single-CPU request uses one worker instead of deadlocking.
// Lines beginning with whitespace or # are skipped, just as in src/starter.c.
// Individual command failures are diagnosed and do not change the final status.
func RunStarter(commands io.Reader, filename string, cpus int, out, errOut io.Writer) error {
	return runStarterCommands(commands, filename, cpus, out, errOut, Abc_RunThread)
}

func runStarterCommands(commands io.Reader, filename string, cpus int, out, errOut io.Writer, execute func(string, io.Writer, io.Writer) error) error {
	if cpus <= 0 {
		return fmt.Errorf("number of CPUs must be positive")
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	var outputMutex sync.Mutex
	stdout, stderr := starterWriter{&outputMutex, out}, starterWriter{&outputMutex, errOut}
	workers := cpus - 1
	if workers < 1 {
		workers = 1
	}
	if workers > MAX_COMM_NUM {
		workers = MAX_COMM_NUM
	}
	slots := make(chan struct{}, workers)
	var running sync.WaitGroup
	defer running.Wait()
	reader := bufio.NewReader(commands)
	count := 0
	for {
		line, readErr := reader.ReadString('\n')
		if len(line) > 0 && !strings.ContainsRune("\n\r\t #", rune(line[0])) {
			command := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			slots <- struct{}{}
			fmt.Fprintf(stdout, "Calling:  %s\n", command)
			running.Add(1)
			go func() {
				defer running.Done()
				defer func() { <-slots }()
				if execute(command, stdout, stderr) != nil {
					outputMutex.Lock()
					fmt.Fprintf(errOut, "The following command has returned non-zero exit status:\n\"%s\"\nSorry for the inconvenience.\n", command)
					outputMutex.Unlock()
				}
			}()
			count++
			if count == MAX_COMM_NUM {
				fmt.Fprintf(stdout, "Cannot execute more than %d commands from file \"%s\".\n", count, filename)
				break
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				return readErr
			}
			break
		}
	}
	return nil
}

// StarterMain is the standalone parallel command-file launcher. Args excludes
// the executable name. Child failures retain the original successful exit code.
func StarterMain(args []string, out, errOut io.Writer) int {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	start := time.Now()
	usage := func() int {
		fmt.Fprintf(out, "usage: %s <num> <file>\n       executes command listed in <file> in parallel on <num> CPUs\n\n", filepath.Base(os.Args[0]))
		return 1
	}
	if len(args) != 2 {
		fmt.Fprintln(errOut, "Wrong number of command line arguments.")
		return usage()
	}
	cpus, err := strconv.Atoi(args[0])
	if err != nil || cpus <= 0 {
		fmt.Fprintln(out, "Cannot read an integer represting the number of CPUs.")
		return usage()
	}
	file, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintf(out, "Input file \"%s\" cannot be opened.\n", args[1])
		return usage()
	}
	defer file.Close()
	if err = RunStarter(file, args[1], cpus, out, errOut); err != nil {
		fmt.Fprintf(errOut, "Cannot read commands: %v\n", err)
		return 1
	}
	fmt.Fprintf(out, "Finished processing commands in file \"%s\".  Total time = %7.2f sec\n", args[1], time.Since(start).Seconds())
	return 0
}
