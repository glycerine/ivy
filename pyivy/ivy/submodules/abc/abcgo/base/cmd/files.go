package cmd

import (
	"compress/bzip2"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Directory helpers use explicit paths, never process-wide chdir.
func directoryOption(args []string) (string, error) {
	directory := "."
	for i := 1; i < len(args); i++ {
		if args[i] != "-D" || i+1 == len(args) {
			return "", fmt.Errorf("expected -D directory")
		}
		i++
		directory = args[i]
	}
	return directory, nil
}
func CmdCommandScanDir(f *Abc_Frame_t, argc int32, argv []string) int32 {
	directory, err := directoryOption(commandArgs(argc, argv))
	if err != nil {
		return usage(f, "scandir [-D directory]")
	}
	files, err := filepath.Glob(filepath.Join(directory, "*.txt"))
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	for _, file := range files {
		name := filepath.Base(file)
		underscore, dot := strings.LastIndexByte(name, '_'), strings.IndexByte(name, '.')
		if underscore < 0 || dot <= underscore {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(f.Err, err)
			continue
		}
		for _, token := range []string{"Property UNDECIDED.  Time =", "Property proved.  Time =", "Time ="} {
			if index := strings.Index(string(data), token); index >= 0 {
				var seconds float32
				fmt.Sscan(string(data[index+len(token):]), &seconds)
				fmt.Fprintf(f.Out, "%s %.2f\n", name[underscore+1:dot], seconds)
				break
			}
		}
	}
	return 0
}
func CmfFindNumber(name string) int32 {
	dot := strings.IndexByte(name, '.')
	if dot < 0 {
		return -1
	}
	start := dot
	for start > 0 && name[start-1] >= '0' && name[start-1] <= '9' {
		start--
	}
	if start == dot {
		return -2
	}
	number, err := strconv.ParseInt(name[start:dot], 10, 32)
	if err != nil {
		return -2
	}
	return int32(number)
}
func UnzipBzip2File(name string) error {
	input, err := os.Open(name)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(name + ".v")
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, bzip2.NewReader(input))
	return errors.Join(copyErr, output.Close())
}
func CnfDupFileUnzip(name string) {
	if err := UnzipBzip2File(name); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
func CmdCommandRenameFiles(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	directory, extension, prefix := ".", "", ""
	base := 0
	for i := 1; i < len(args); i++ {
		option := args[i]
		if i+1 == len(args) {
			return usage(f, "renamefiles -E extension [-D directory] [-N prefix] [-B base]")
		}
		i++
		switch option {
		case "-D":
			directory = args[i]
		case "-E":
			extension = args[i]
		case "-N":
			prefix = args[i]
		case "-B":
			value, err := strconv.Atoi(args[i])
			if err != nil || value < 0 {
				return 1
			}
			base = value
		default:
			return usage(f, "renamefiles -E extension [-D directory] [-N prefix] [-B base]")
		}
	}
	if extension == "" {
		return usage(f, "renamefiles -E extension [-D directory] [-N prefix] [-B base]")
	}
	files, err := filepath.Glob(filepath.Join(directory, "*."+extension))
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	type rename struct {
		old, new string
		number   int
	}
	moves := make([]rename, 0, len(files))
	maximum := 0
	for _, file := range files {
		number := int(CmfFindNumber(filepath.Base(file)))
		if number < 0 {
			fmt.Fprintf(f.Err, "File has no number: %s\n", file)
			return 1
		}
		if base > int(^uint(0)>>1)-number {
			fmt.Fprintln(f.Err, "Filename number overflow")
			return 1
		}
		number += base
		if number > maximum {
			maximum = number
		}
		moves = append(moves, rename{old: file, number: number})
	}
	width := len(strconv.Itoa(maximum))
	sort.SliceStable(moves, func(i, j int) bool { return moves[i].number < moves[j].number })
	destinations := make(map[string]bool)
	for i := range moves {
		move := &moves[i]
		move.new = filepath.Join(directory, fmt.Sprintf("%s%0*d.%s", prefix, width, move.number, extension))
		if destinations[move.new] {
			fmt.Fprintln(f.Err, "Duplicate rename destination", move.new)
			return 1
		}
		destinations[move.new] = true
		if move.new != move.old {
			if _, err := os.Lstat(move.new); err == nil {
				fmt.Fprintln(f.Err, "Rename destination exists", move.new)
				return 1
			} else if !os.IsNotExist(err) {
				fmt.Fprintln(f.Err, err)
				return 1
			}
		}
	}
	for _, move := range moves {
		if move.old == move.new {
			continue
		}
		if err := os.Rename(move.old, move.new); err != nil {
			fmt.Fprintln(f.Err, err)
			return 1
		}
		fmt.Fprintf(f.Out, "%s -> %s\n", filepath.Base(move.old), filepath.Base(move.new))
	}
	return 0
}
func CmdCommandLs(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	long, onlyMV := false, false
	for _, arg := range args[1:] {
		switch arg {
		case "-l":
			long = true
		case "-b":
			onlyMV = true
		default:
			return usage(f, "ls [-l] [-b]")
		}
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	count := 0
	columns := 4
	if long {
		columns = 2
		fmt.Fprintln(f.Out, " File              Date           Size | File              Date           Size")
	}
	for _, entry := range entries {
		if onlyMV && !strings.HasSuffix(entry.Name(), ".mv") {
			continue
		}
		if long {
			info, err := entry.Info()
			if err != nil {
				fmt.Fprintln(f.Err, err)
				continue
			}
			fmt.Fprintf(f.Out, " %-17s %s%7d", entry.Name(), info.ModTime().Format("Jan _2 15:04"), info.Size())
		} else {
			fmt.Fprintf(f.Out, " %-18s", entry.Name())
		}
		count++
		if count%columns == 0 {
			fmt.Fprintln(f.Out)
		} else if long {
			fmt.Fprint(f.Out, " |")
		} else {
			fmt.Fprint(f.Out, " ")
		}
	}
	if count%columns != 0 {
		fmt.Fprintln(f.Out)
	}
	if count == 0 {
		fmt.Fprintln(f.Out, "No files in the current directory.")
	}
	return 0
}
func CmdCommandScrGen(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	filename, directory, commands, writeDirectory := "test.s", "", "ps", ""
	current := false
	for i := 1; i < len(args); i++ {
		option := args[i]
		if option == "-c" {
			current = !current
			continue
		}
		if i+1 == len(args) {
			return usage(f, "scrgen [-F file] [-D directory] [-C commands] [-W output-directory] [-c]")
		}
		i++
		switch option {
		case "-F":
			filename = args[i]
		case "-D":
			directory = args[i]
		case "-C":
			commands = args[i]
		case "-W":
			writeDirectory = args[i]
		default:
			return usage(f, "scrgen [-F file] [-D directory] [-C commands] [-W output-directory] [-c]")
		}
	}
	if directory == "" {
		directory = "."
		current = true
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	target := filename
	if !current {
		target = filepath.Join(directory, filename)
	}
	file, err := os.Create(target)
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	fmt.Fprintf(file, "# Script file produced by ABC on %s\n", time.Now().Format(time.RFC1123))
	var writeErr error
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasSuffix(name, ".") || strings.HasSuffix(name, ".s") || strings.HasSuffix(name, ".txt") {
			continue
		}
		input := filepath.ToSlash(filepath.Join(directory, name))
		line := "r " + shellQuote(input) + " ; " + commands
		if writeDirectory != "" {
			line += " ; w " + shellQuote(filepath.ToSlash(filepath.Join(writeDirectory, name)))
		}
		if _, err := fmt.Fprintln(file, line); err != nil {
			writeErr = err
			break
		}
	}
	if err := errors.Join(writeErr, file.Close()); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	fmt.Fprintf(f.Out, "Script file %q was saved.\n", target)
	return 0
}
func init() {
	if runtime.GOOS == "windows" {
		extensionInitializers = append(extensionInitializers, func(f *Abc_Frame_t) {
			Cmd_CommandAdd(f, "Basic", "scandir", CmdCommandScanDir, 0)
			Cmd_CommandAdd(f, "Basic", "renamefiles", CmdCommandRenameFiles, 0)
			Cmd_CommandAdd(f, "Basic", "ls", CmdCommandLs, 0)
			Cmd_CommandAdd(f, "Basic", "scrgen", CmdCommandScrGen, 0)
		})
	}
}
