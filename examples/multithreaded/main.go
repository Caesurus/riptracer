//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/caesurus/riptracer"
)

var gHit = 0
var gHitHW = 0

func CBHits(pid int, bp riptracer.BreakPoint) {
	gHit += 1
}
func CBHWHits(pid int, bp riptracer.BreakPoint) {
	gHitHW += 1
}

func main() {
	parser := argparse.NewParser("tracer", "General tracer")
	var verbose *bool = parser.Flag("v", "verbose", &argparse.Options{Help: "Verbose Output"})
	var breakPointStr *string = parser.String("b", "breakpoint", &argparse.Options{Required: true, Help: "Breakpoint in hex"})
	var hwbreakPointStr *string = parser.String("w", "hwbreakpoint", &argparse.Options{Required: true, Help: "HWBreakpoint in hex"})

	startCmd := parser.NewCommand("start", "Will start a process")
	var cmd_str *string = startCmd.String("c", "cmd", &argparse.Options{Required: true, Help: "Cmd to execute"})

	attachPid := parser.NewCommand("attach", "Will attach to an existing process")
	var pidOfProcess *int = attachPid.Int("p", "pid", &argparse.Options{Required: true, Help: "Process ID of process to search in"})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Print(parser.Usage(err))
		return
	}
	var tracer *riptracer.Tracer

	if startCmd.Happened() {
		log.Println("Started process")
		tracer, err = riptracer.NewTracerStartCommand(*cmd_str)
		if err != nil {
			log.Fatalln(err)
		}
	} else if attachPid.Happened() {
		log.Println("Connecting to PID:", *pidOfProcess)
		tracer, err = riptracer.NewTracerFromPid(*pidOfProcess)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if *verbose {
		tracer.EnableVerbose()
	}

	breakPointInt, err := strconv.ParseInt(*breakPointStr, 16, 64)
	if err != nil {
		panic(err)
	}

	hwbreakPointInt, err := strconv.ParseInt(*hwbreakPointStr, 16, 64)
	fmt.Println(hwbreakPointInt)
	if err != nil {
		panic(err)
	}
	tracer.SetBreakpointRelative(uintptr(breakPointInt), riptracer.CBPrintRegisters)
	tracer.SetBreakpointRelative(uintptr(breakPointInt), riptracer.CBPrintStack)
	tracer.SetBreakpointRelative(uintptr(breakPointInt), riptracer.CBFunctionArgs)
	tracer.SetBreakpointRelative(uintptr(breakPointInt), CBHits)

	tracer.SetHWBreakpointRelative(uintptr(hwbreakPointInt), CBHWHits)
	tracer.SetHWBreakpointRelative(uintptr(hwbreakPointInt), riptracer.CBFunctionArgs)

	tracer.Start()

	fmt.Printf("Hit SW: %d / HW: %d times\n", gHit, gHitHW)

	if gHit == 0 {
		os.Exit(1)
	}

	if gHitHW == 0 {
		os.Exit(2)
	}

}
