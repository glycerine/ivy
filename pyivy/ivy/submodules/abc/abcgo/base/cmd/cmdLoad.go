package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CmdCollectFileNames discovers external ABC commands in the current directory.
// Executable Unix files are accepted in addition to the original .exe binaries.
func CmdCollectFileNames() []string {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "abccmd_") {
			continue
		}
		info, err := entry.Info()
		if err == nil && (strings.HasSuffix(entry.Name(), ".exe") || info.Mode()&0111 != 0) {
			names = append(names, entry.Name())
		}
	}
	return names
}
func externalCommand(name string) (string, error) {
	for _, candidate := range []string{"abccmd_" + name + ".exe", "abccmd_" + name} {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf("cannot find binary abccmd_%s", name)
}
func CmdCommandLoad(f *Abc_Frame_t, argc int32, argv []string) int32 {
	args := commandArgs(argc, argv)
	if len(args) == 0 {
		return 1
	}
	binary, err := externalCommand(args[0])
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	command := exec.Command(binary, args[1:]...)
	command.Stdin = f.In
	command.Stdout = f.Out
	command.Stderr = f.Err
	if err := command.Run(); err != nil {
		fmt.Fprintf(f.Err, "External command %s: %v\n", args[0], err)
		return 1
	}
	return 0
}
func Load_Init(f *Abc_Frame_t) {
	for _, name := range CmdCollectFileNames() {
		name = strings.TrimPrefix(name, "abccmd_")
		if dot := strings.IndexByte(name, '.'); dot >= 0 {
			name = name[:dot]
		}
		if name != "" {
			Cmd_CommandAdd(f, "ZZ", name, CmdCommandLoad, 0)
		}
	}
}

// External commands hold no process or file resources between calls.
func Load_End(f *Abc_Frame_t) {}
func init()                   { extensionInitializers = append(extensionInitializers, Load_Init) }
