//go:build !windows
// +build !windows

package main

import (
	"log"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/caesurus/riptracer"
)

var gHit = 0

func CBHits(pid int, bp riptracer.BreakPoint) {
	gHit += 1
}

func main() {
	parser := argparse.NewParser("tracer", "General tracer")
	var verbose *bool = parser.Flag("v", "verbose", &argparse.Options{Help: "Verbose Output"})
	var breakPointStr *string = parser.String("b", "breakpoint", &argparse.Options{Required: true, Help: "Breakpoint in hex"})

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

	tracer.EnableVerbose()

	breakPointInt, err := strconv.ParseInt(*breakPointStr, 16, 64)
	if err != nil {
		panic(err)
	}
	tracer.SetBreakpointRelative(uintptr(breakPointInt), riptracer.CBPrintRegisters)
	tracer.SetBreakpointRelative(uintptr(breakPointInt), riptracer.CBPrintStack)
	tracer.SetBreakpointRelative(uintptr(breakPointInt), CBHits)
	tracer.SetFollowForks(true)

	tracer.Start()

	if gHit == 0 {
		os.Exit(1)
	}

}
