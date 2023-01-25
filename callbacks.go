package rip_tracer

import (
	"fmt"
	"syscall"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func CBPrintRegisters(pid int, bp BreakPoint) {
	fmt.Println(Blue, "----------REGS----------", Reset)
	var regs syscall.PtraceRegs
	check(syscall.PtraceGetRegs(pid, &regs))

	fmt.Printf("%srax:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rax, regs.Rax, Reset)
	fmt.Printf("%srbx:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rbx, regs.Rbx, Reset)
	fmt.Printf("%srcx:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rcx, regs.Rcx, Reset)
	fmt.Printf("%srdx:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rdx, regs.Rdx, Reset)
	fmt.Printf("%srdi:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rdi, regs.Rdi, Reset)
	fmt.Printf("%srsi:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rsi, regs.Rsi, Reset)
	fmt.Printf("%srbp:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rbp, regs.Rbp, Reset)
	fmt.Printf("%srsp:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rsp, regs.Rsp, Reset)
	fmt.Printf("%srip:%s 0x%012x (%d)%s\n", Blue, Green, regs.Rip, regs.Rip, Reset)
	/*
		fmt.Printf("%s r8:%s 0x%012x (%d)%s\n", Blue, Green, regs.R8, regs.R8, Reset)
		fmt.Printf("%s r9:%s 0x%012x (%d)%s\n", Blue, Green, regs.R9, regs.R9, Reset)
		fmt.Printf("%sr10:%s 0x%012x (%d)%s\n", Blue, Green, regs.R10, regs.R10, Reset)
		fmt.Printf("%sr11:%s 0x%012x (%d)%s\n", Blue, Green, regs.R11, regs.R11, Reset)
		fmt.Printf("%sr12:%s 0x%012x (%d)%s\n", Blue, Green, regs.R12, regs.R12, Reset)
		fmt.Printf("%sr13:%s 0x%012x (%d)%s\n", Blue, Green, regs.R13, regs.R13, Reset)
		fmt.Printf("%sr14:%s 0x%012x (%d)%s\n", Blue, Green, regs.R14, regs.R14, Reset)
		fmt.Printf("%sr15:%s 0x%012x (%d)%s\n", Blue, Green, regs.R15, regs.R15, Reset)
	*/
}

func CBPrintStack(pid int, bp BreakPoint) {
	fmt.Println(Blue, "----------STACK----------", Reset)
	var regs syscall.PtraceRegs
	check(syscall.PtraceGetRegs(pid, &regs))

	data := make([]byte, 0x30)
	syscall.PtracePeekData(pid, uintptr(regs.Rsp), data)
	Dump(data)
}

func CBFunctionArgs(pid int, bp BreakPoint) {
	var regs syscall.PtraceRegs
	check(syscall.PtraceGetRegs(pid, &regs))
	fmt.Printf("%sThread: %d: arg1: 0x%012x arg2: 0x%012x arg3: 0x%012x %s\n", Green, pid, regs.Rdi, regs.Rsi, regs.Rdx, Reset)
}
