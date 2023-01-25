//go:build !windows
// +build !windows

package main

import (
	"log"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/caesurus/rip_tracer"
)

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
	var tracer *rip_tracer.Tracer

	if startCmd.Happened() {
		log.Println("Started process")
		tracer, err = rip_tracer.NewTracerStartCommand(*cmd_str)
		if err != nil {
			log.Fatalln(err)
		}
	} else if attachPid.Happened() {
		log.Println("Connecting to PID:", *pidOfProcess)
		tracer, err = rip_tracer.NewTracerFromPid(*pidOfProcess)
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
	tracer.SetBreakpointRelative(uintptr(breakPointInt), rip_tracer.CBPrintRegisters)
	tracer.SetBreakpointRelative(uintptr(breakPointInt), rip_tracer.CBPrintStack)
	tracer.SetBreakpointRelative(uintptr(breakPointInt), rip_tracer.CBFunctionArgs)
	tracer.Start()
}
