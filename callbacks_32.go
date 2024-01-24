//go:build 386
// +build 386

package riptracer

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func CBPrintRegisters(pid int, bp BreakPoint) {
	fmt.Println(Blue, "----------REGS----------", Reset)
	var regs unix.PtraceRegs
	check(unix.PtraceGetRegs(pid, &regs))

	fmt.Printf("%seax:%s 0x%012x (%d)%s\n", Blue, Green, regs.Eax, regs.Eax, Reset)
	fmt.Printf("%sebx:%s 0x%012x (%d)%s\n", Blue, Green, regs.Ebx, regs.Ebx, Reset)
	fmt.Printf("%secx:%s 0x%012x (%d)%s\n", Blue, Green, regs.Ecx, regs.Ecx, Reset)
	fmt.Printf("%sedx:%s 0x%012x (%d)%s\n", Blue, Green, regs.Edx, regs.Edx, Reset)
	fmt.Printf("%sedi:%s 0x%012x (%d)%s\n", Blue, Green, regs.Edi, regs.Edi, Reset)
	fmt.Printf("%sesi:%s 0x%012x (%d)%s\n", Blue, Green, regs.Esi, regs.Esi, Reset)
	fmt.Printf("%sebp:%s 0x%012x (%d)%s\n", Blue, Green, regs.Ebp, regs.Ebp, Reset)
	fmt.Printf("%sesp:%s 0x%012x (%d)%s\n", Blue, Green, regs.Esp, regs.Esp, Reset)
	fmt.Printf("%seip:%s 0x%012x (%d)%s\n", Blue, Green, regs.Eip, regs.Eip, Reset)
}

func CBPrintStack(pid int, bp BreakPoint) {
	fmt.Println(Blue, "----------STACK----------", Reset)
	var regs unix.PtraceRegs
	check(unix.PtraceGetRegs(pid, &regs))

	data := make([]byte, 0x30)
	unix.PtracePeekData(pid, uintptr(regs.Esp), data)
	Dump(data)
}

func CBFunctionArgs(pid int, bp BreakPoint) {
	//TODO
}
