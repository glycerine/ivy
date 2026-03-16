// Package ivylaunch provides process launching for compiled Ivy test/build.
// This is a port of Python's ivy_launch.py (221 lines).
//
// It reads .dsc (descriptor) files that describe multi-process Ivy
// deployments (with network endpoints) and launches them.
package ivylaunch

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Descriptor describes a multi-process Ivy deployment.
type Descriptor struct {
	Processes  []Process `json:"processes"`
	TestParams []string  `json:"test_params,omitempty"`
}

// Process describes a single process in the deployment.
type Process struct {
	Name    string  `json:"name"`
	Binary  string  `json:"binary"`
	Indices []Index `json:"indices"`
	Params  []Param `json:"params"`
}

// Index describes a parameterized dimension of a process.
type Index struct {
	Name  string `json:"name"`
	Range []int  `json:"range,omitempty"`
}

// Param describes a process parameter.
type Param struct {
	Name    string      `json:"name"`
	Type    interface{} `json:"type"` // string or object with "name"
	Default interface{} `json:"default,omitempty"`
}

// Config holds launch configuration.
type Config struct {
	DSCFile string
	Params  map[string]string
	Runs    int
	Seed    int
}

// NextUnusedPort is the next available port for UDP/TCP endpoints.
var NextUnusedPort = 49123

// GetUnusedPort returns the next unused port number.
func GetUnusedPort() int {
	NextUnusedPort++
	return NextUnusedPort
}

// LoadDescriptor reads a .dsc file.
func LoadDescriptor(filename string) (*Descriptor, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", filename, err)
	}
	var desc Descriptor
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("error in %s: %w", filename, err)
	}
	return &desc, nil
}

// Launch launches the processes described in the descriptor.
func Launch(cfg *Config) error {
	desc, err := LoadDescriptor(cfg.DSCFile)
	if err != nil {
		return err
	}

	var cmds []*exec.Cmd
	for _, proc := range desc.Processes {
		binary := proc.Binary
		if !strings.Contains(binary, "/") {
			binary = "./" + binary
		}
		args := []string{}
		for _, param := range proc.Params {
			if val, ok := cfg.Params[param.Name]; ok {
				if param.Default != nil {
					args = append(args, fmt.Sprintf("%s=%s", param.Name, val))
				} else {
					args = append(args, val)
				}
			}
		}
		if cfg.Runs > 1 {
			args = append(args, fmt.Sprintf("seed=%d", cfg.Seed))
		}

		cmd := exec.Command(binary, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start %s: %w", binary, err)
		}
		cmds = append(cmds, cmd)
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all processes
	var firstErr error
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Main is the entry point for ivy_launch.
func Main() error {
	params := make(map[string]string)
	args := os.Args[1:]
	for len(args) > 0 && strings.Contains(args[0], "=") {
		parts := strings.SplitN(args[0], "=", 2)
		params[parts[0]] = parts[1]
		args = args[1:]
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: ivy_launch [option=value...] <file>.dsc")
	}
	dscFile := args[0]
	if !strings.HasSuffix(dscFile, ".dsc") {
		if strings.HasSuffix(dscFile, ".ivy") {
			dscFile = dscFile[:len(dscFile)-4]
		}
		dscFile += ".dsc"
	}

	runs := 1
	if r, ok := params["runs"]; ok {
		fmt.Sscanf(r, "%d", &runs)
		delete(params, "runs")
	}

	for run := 0; run < runs; run++ {
		cfg := &Config{
			DSCFile: dscFile,
			Params:  params,
			Runs:    runs,
			Seed:    run,
		}
		if err := Launch(cfg); err != nil {
			return err
		}
	}
	return nil
}
