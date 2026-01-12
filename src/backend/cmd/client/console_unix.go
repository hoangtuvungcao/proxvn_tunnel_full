//go:build !windows
// +build !windows

package main

func enableVirtualTerminalProcessing() {
	// No-op for Unix/Linux/MacOS as they typically support ANSI by default
}
