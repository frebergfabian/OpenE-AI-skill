//go:build windows

package main

import (
	"crypto/sha256"
	"fmt"
	"os/exec"
	"runtime"
	"syscall"
	"unsafe"
)

var kernel = syscall.NewLazyDLL("kernel32.dll")
var createMutex = kernel.NewProc("CreateMutexW")
var releaseMutex = kernel.NewProc("ReleaseMutex")
var waitObject = kernel.NewProc("WaitForSingleObject")

func lock(name string) (func(), error) {
	runtime.LockOSThread()
	hash := sha256.Sum256([]byte(name))
	p, _ := syscall.UTF16PtrFromString(fmt.Sprintf("Local\\OpenE-Sandbox-%x", hash))
	h, _, err := createMutex.Call(0, 0, uintptr(unsafe.Pointer(p)))
	if h == 0 {
		runtime.UnlockOSThread()
		return nil, err
	}
	r, _, _ := waitObject.Call(h, 0)
	if r != 0 && r != 0x80 {
		syscall.CloseHandle(syscall.Handle(h))
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("instance operation already in progress")
	}
	return func() { releaseMutex.Call(h); syscall.CloseHandle(syscall.Handle(h)); runtime.UnlockOSThread() }, nil
}
func hidden(cmd *exec.Cmd, detached bool) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if detached {
		cmd.SysProcAttr.CreationFlags |= 0x00000200
	}
}
