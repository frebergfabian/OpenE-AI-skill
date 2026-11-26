//go:build !windows

package main

import "os/exec"

// These tests exercise helpers and simulators. Windows process management needs a Windows host.
func lock(string) (func(), error) {
	panic("Windows process management requires a Windows host")
}

func hidden(*exec.Cmd, bool) {
	panic("Windows process management requires a Windows host")
}
