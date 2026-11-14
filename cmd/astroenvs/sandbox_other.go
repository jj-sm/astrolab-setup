//go:build !linux

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// The sandbox uses Linux namespaces; elsewhere (building and trying the manager on a Mac)
// it is simply not offered

func sandboxSupported() bool { return false }

// sandboxCommand never runs prog unsandboxed: the command fails with runSandbox's message
func sandboxCommand(rw string, ro []string, prog string, args ...string) *exec.Cmd {
	return exec.Command(os.Args[0], sandboxArg)
}

func runSandbox(argv []string) {
	fmt.Fprintln(os.Stderr, "sandbox: only available on Linux")
	os.Exit(1)
}
