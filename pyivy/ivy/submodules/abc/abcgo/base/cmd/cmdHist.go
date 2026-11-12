package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Cmd_HistoryAddCommand(f *Abc_Frame_t, line string) {
	if f.BatchMode {
		return
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	if len(line) <= 3 || strings.HasSuffix(line, "?") {
		return
	}
	for _, prefix := range []string{"set", "unset", "time", "quit", "alias", "source abc.rc", "source ..\\abc.rc", "history", "hi "} {
		if strings.HasPrefix(line, prefix) {
			return
		}
	}
	if line == "hi" {
		return
	}
	history := f.History()
	start := len(history) - 10
	if start < 0 {
		start = 0
	}
	for i := start; i < len(history); i++ {
		if history[i] == line {
			history = append(history[:i], history[i+1:]...)
			history = append(history, line)
			f.SetHistory(history)
			return
		}
	}
	f.AppendHistory(line)
	if _, exists := f.LookupFlag("history_file"); exists {
		Cmd_HistoryWrite(f, 1000)
	}
}
func historyPath(f *Abc_Frame_t) string {
	if path, exists := f.LookupFlag("history_file"); exists {
		return path
	}
	return "abc.history"
}
func Cmd_HistoryRead(f *Abc_Frame_t) {
	file, err := os.Open(historyPath(f))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(f.Err, err)
		return
	}
	f.SetHistory(lines)
}
func Cmd_HistoryWrite(f *Abc_Frame_t, limit int32) {
	history := f.History()
	if limit < 0 {
		limit = 0
	}
	if int(limit) < len(history) {
		history = history[len(history)-int(limit):]
	}
	file, err := os.Create(historyPath(f))
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return
	}
	for _, line := range history {
		if _, err := fmt.Fprintln(file, line); err != nil {
			fmt.Fprintln(f.Err, err)
			break
		}
	}
	if err := file.Close(); err != nil {
		fmt.Fprintln(f.Err, err)
	}
}
func Cmd_HistoryPrint(f *Abc_Frame_t, limit int32) {
	history := f.History()
	start := len(history) - int(limit)
	if start < 0 {
		start = 0
	}
	if start > len(history) {
		start = len(history)
	}
	for i := start; i < len(history); i++ {
		fmt.Fprintf(f.Out, "%2d : %s\n", len(history)-i, history[i])
	}
}
