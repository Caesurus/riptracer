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

	fmt.Printf("%seax:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Eax), uintptr(regs.Eax), Reset)
	fmt.Printf("%sebx:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Ebx), uintptr(regs.Ebx), Reset)
	fmt.Printf("%secx:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Ecx), uintptr(regs.Ecx), Reset)
	fmt.Printf("%sedx:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Edx), uintptr(regs.Edx), Reset)
	fmt.Printf("%sedi:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Edi), uintptr(regs.Edi), Reset)
	fmt.Printf("%sesi:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Esi), uintptr(regs.Esi), Reset)
	fmt.Printf("%sebp:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Ebp), uintptr(regs.Ebp), Reset)
	fmt.Printf("%sesp:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Esp), uintptr(regs.Esp), Reset)
	fmt.Printf("%seip:%s 0x%08x (%d)%s\n", Blue, Green, uintptr(regs.Eip), uintptr(regs.Eip), Reset)
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
