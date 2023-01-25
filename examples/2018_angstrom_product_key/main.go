//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/akamensky/argparse"
	"github.com/caesurus/rip_tracer"
)

var g_cnt = 0
var g_serial []int32

func CBKeyBreakPoint(pid int) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatalln("Error", err)
	}

	eax_ := int32(regs.Rax & 0xffffffff)
	edx_ := int32(regs.Rdx & 0xffffffff)
	fmt.Printf("*** eax(%6d) - edx(%6d)\n", eax_, edx_)
	serial_char := eax_ - edx_
	fmt.Printf("eax(%6d) - edx(%6d) = key[%d] %d\n", eax_, edx_, g_cnt, serial_char)

	g_cnt = g_cnt + 1

	g_serial[g_cnt] = serial_char
}

func CBPrintSerialKey(pid int) {
	fmt.Printf("\nValid Serial Key = %d-%d-%d-%d-%d-%d\n", g_serial[5], g_serial[1], g_serial[2], g_serial[6], g_serial[3], g_serial[4])
}

func main() {
	parser := argparse.NewParser("tracer", "General tracer")
	var verbose *bool = parser.Flag("v", "verbose", &argparse.Options{Help: "Verbose Output"})

	startCmd := parser.NewCommand("start", "Will start a process")
	var cmd_str *string = startCmd.String("c", "cmd", &argparse.Options{Required: true, Help: "Cmd to execute"})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Print(parser.Usage(err))
		return
	}
	var tracer *rip_tracer.Tracer

	g_serial = make([]int32, 10)

	if startCmd.Happened() {
		log.Println("Started process")
		tracer, err = rip_tracer.NewTracerStartCommand(*cmd_str)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if *verbose {
		tracer.EnableVerbose()
	}
	tracer.SetBreakpoint(uintptr(0x400fb8), CBKeyBreakPoint, true)
	tracer.SetBreakpoint(uintptr(0x40115f), CBPrintSerialKey, true)
	tracer.Start()

}