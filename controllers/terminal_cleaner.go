package controllers

import (
	"os"
	"os/exec"
	"runtime"
)

var (
	TerminalCleaner terminalCleanerInterface = &terminalCleaner{}
)

type terminalCleaner struct {}

type terminalCleanerInterface interface {
	SetCleaner() *map[string]func()
	Clean(*map[string]func())
}

// SetCleaner initializer the cleaner for the current terminal session
func (tc *terminalCleaner) SetCleaner() *map[string]func() {
	clear := make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	return &clear
}

// Clean cleans the the current terminal session
func (tc *terminalCleaner) Clean(clear *map[string]func()) {
	value, ok := (*clear)[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok { //if we defined a clear func for that platform:
		value()  //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}
