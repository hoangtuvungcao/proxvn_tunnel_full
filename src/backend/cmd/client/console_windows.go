package main

import (
	"log"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func enableVirtualTerminalProcessing() {
	if os.PathSeparator != '\\' {
		return // Not Windows
	}

	var h syscall.Handle
	var err error
	// Stdout
	h, err = syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}

	var mode uint32
	ret, _, _ := procGetConsoleMode.Call(uintptr(h), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return
	}

	// ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	mode |= 0x0004
	ret, _, err = procSetConsoleMode.Call(uintptr(h), uintptr(mode))
	if ret == 0 {
		log.Printf("[client] failed to set console mode: %v", err)
	}
}
