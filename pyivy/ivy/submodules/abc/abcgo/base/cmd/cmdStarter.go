package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RunStarter executes a command list using at most cores-1 concurrent children.
// With a nonempty script each list entry is an input filename passed to ABC.
func RunStarter(filename, binary, script string, cores int, out, errOut io.Writer) error {
	if cores < 2 {
		return fmt.Errorf("starter needs at least two cores")
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t\r")
		if line == "" || strings.ContainsRune(" \t#", rune(line[0])) {
			continue
		}
		if script != "" {
			if _, err := os.Stat(line); err != nil {
				return err
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	start := time.Now()
	jobs := make(chan string)
	var workers sync.WaitGroup
	var mu sync.Mutex
	var failures []error
	// Writers are shared by child copy goroutines, including bytes.Buffer in callers.
	lockedOut := &synchronizedWriter{writer: out, mu: &mu}
	lockedErr := &synchronizedWriter{writer: errOut, mu: &mu}
	for i := 0; i < cores-1; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for line := range jobs {
				var command *exec.Cmd
				var log *os.File
				if script != "" {
					command = exec.Command(binary, "-c", line+"; "+script)
					var createErr error
					log, createErr = os.Create(line + ".txt")
					if createErr != nil {
						mu.Lock()
						failures = append(failures, createErr)
						mu.Unlock()
						continue
					}
					command.Stdout = log
				} else {
					if runtime.GOOS == "windows" {
						command = exec.Command("cmd", "/C", line)
					} else {
						command = exec.Command("/bin/sh", "-c", line)
					}
					command.Stdout = lockedOut
				}
				command.Stderr = lockedErr
				fmt.Fprintf(lockedOut, "Calling:  %s\n", line)
				runErr := command.Run()
				if log != nil {
					runErr = errors.Join(runErr, log.Close())
				}
				if runErr != nil {
					mu.Lock()
					failures = append(failures, fmt.Errorf("%s: %w", line, runErr))
					mu.Unlock()
				}
			}
		}()
	}
	for _, line := range lines {
		jobs <- line
	}
	close(jobs)
	workers.Wait()
	fmt.Fprintf(out, "Finished processing commands in file %q. Total wall time = %.2f sec\n", filename, time.Since(start).Seconds())
	return errors.Join(failures...)
}

type synchronizedWriter struct {
	writer io.Writer
	mu     *sync.Mutex
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}
func Cmd_RunStarter(filename, binary, script string, cores int32) {
	if err := RunStarter(filename, binary, script, int(cores), os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
func CmdCommandStarter(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	cores := 3
	script := ""
	var files []string
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-N", "-C":
			option := args[i]
			i++
			if i == len(args) {
				return usage(f, "starter [-N cores] [-C command] file")
			}
			if option == "-N" {
				value, err := strconv.Atoi(args[i])
				if err != nil {
					return 1
				}
				cores = value
			} else {
				script = args[i]
			}
		case "-v":
		case "-h":
			return usage(f, "starter [-N cores] [-C command] file")
		default:
			files = append(files, args[i])
		}
	}
	if len(files) != 1 {
		return usage(f, "starter [-N cores] [-C command] file")
	}
	if err := RunStarter(files[0], f.Binary, script, cores, f.Out, f.Err); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	return 0
}
func init() {
	extensionInitializers = append(extensionInitializers, func(f *Abc_Frame_t) { Cmd_CommandAdd(f, "Various", "starter", CmdCommandStarter, 0) })
}
