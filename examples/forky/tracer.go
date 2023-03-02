package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/akamensky/argparse"
	"github.com/caesurus/riptracer"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadMem(pid int, addr uintptr, length int) []byte {
	data := make([]byte, length)
	_, err := syscall.PtracePeekData(pid, addr, data)
	if err != nil {
		fmt.Printf("Error reading memory at 0x%012x\n", addr)
	}
	return data
}

func Read32bitValue(pid int, addr uintptr) uint32 {
	data := ReadMem(pid, addr, 4)
	ret := binary.LittleEndian.Uint32(data)
	return ret
}

func CBFuncCalls(pid int, bp riptracer.BreakPoint) {
	var regs syscall.PtraceRegs
	check(syscall.PtraceGetRegs(pid, &regs))
	fmt.Printf("pid:%d -> doNothing called with arg: %d\n", pid, int32(regs.Rax))
}

func main() {
	parser := argparse.NewParser("tracer", "RE stuff...")
	var verbose *bool = parser.Flag("v", "verbose", &argparse.Options{Help: "Verbose Output"})

	startCmd := parser.NewCommand("start", "Will start a process")
	var cmd_str *string = startCmd.String("c", "cmd", &argparse.Options{Required: true, Help: "Cmd to execute"})

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
	}

	if *verbose {
		tracer.EnableVerbose()
	}

	tracer.SetFollowForks(true)

	tracer.SetBreakpointRelative(0x560, CBFuncCalls)

	tracer.Start()
}
