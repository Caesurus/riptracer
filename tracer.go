package riptracer

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/procfs"
	"golang.org/x/sys/unix"
)

type CallBackFunction func(int, BreakPoint) // CallBack Function Pointer

type BreakPoint struct {
	Address      uintptr
	OriginalCode *[]byte
	Hits         int
	Callbacks    []CallBackFunction
}

type Tracer struct {
	Process          *os.Process
	ProcFS           procfs.FS
	ws               syscall.WaitStatus
	breakpoints      map[uintptr]*BreakPoint
	threads          map[int]bool
	verbose          bool
	exeCompareLength int
	baseAddress      uintptr
	ptraceOptions    int
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// How many bytes we want to use to compare mem to executable
const DEFAULTEXECMPLENGTH = 32

var shutdownFlag = false

func readBytesFromFile(filePath string, length int, offset int64) []byte {
	f, err := os.Open(filePath)
	check(err)

	_, err = f.Seek(offset, 0)
	check(err)

	data := make([]byte, length)
	numBytes, err := f.Read(data)
	check(err)
	if length != numBytes {
		log.Fatalln("Couldn't read expected number of bytes from exe file")
	}

	return data
}

func _attachToPid(pid int) (int, error) {
	err := syscall.PtraceAttach(pid)
	if err == syscall.EPERM {
		_, err := syscall.PtraceGetEventMsg(pid)
		if err != nil {
			log.Fatalln("Permissions Error attaching to PID, please run as root, err message: ", err)
			return 0, err
		}
	} else if err != nil {
		log.Fatalln("Error attaching to pid:", err)
		return 0, err
	}

	return pid, nil
}

func NewTracerStartCommand(cmd_str string) (*Tracer, error) {
	runtime.LockOSThread()
	threads := make(map[int]bool)

	cmds := strings.Split(cmd_str, " ")

	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}
	check(cmd.Start())

	cmd.Wait() //Ignore the error, we hit our starting breakpoint trap

	check(syscall.PtraceSetOptions(cmd.Process.Pid, syscall.PTRACE_O_TRACECLONE))

	log.Printf("CMD PID: %s : %v\n", cmd_str, cmd.Process.Pid)
	syscall.PtraceSingleStep(cmd.Process.Pid)

	var ws syscall.WaitStatus
	wpid, err := syscall.Wait4(cmd.Process.Pid, &ws, syscall.WALL, nil)

	// Add this pid to known threads. We need to continue this pid once breakpoints are set.
	threads[wpid] = true

	procFS, err := procfs.NewFS("/proc")
	if err != nil {
		log.Fatalln("Couldn't access proc fs", err)
		return nil, err
	}

	return &Tracer{
		Process:          cmd.Process,
		ProcFS:           procFS,
		breakpoints:      make(map[uintptr]*BreakPoint),
		threads:          threads,
		exeCompareLength: DEFAULTEXECMPLENGTH,
		baseAddress:      0,
		ptraceOptions:    syscall.PTRACE_O_TRACECLONE,
	}, nil

}

func NewTracerFromPid(pid int) (*Tracer, error) {
	var ws syscall.WaitStatus
	runtime.LockOSThread()

	procFS, err := procfs.NewFS("/proc")
	if err != nil {
		log.Fatalln("Couldn't access proc fs", err)
		return nil, err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalf("Failed to find process: %s\n", err)
	}

	all_pids, err := procFS.AllThreads(pid)
	check(err)

	tracer := Tracer{
		Process:          proc,
		ProcFS:           procFS,
		breakpoints:      make(map[uintptr]*BreakPoint),
		threads:          make(map[int]bool),
		exeCompareLength: DEFAULTEXECMPLENGTH,
		baseAddress:      0,
		ptraceOptions:    syscall.PTRACE_O_TRACECLONE,
	}

	for i := range all_pids {
		p := all_pids[i].PID

		_attachToPid(p)

		for {
			_, err := syscall.Wait4(p, &ws, syscall.WALL, nil)
			if ws.Stopped() {
				check(err)
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
		tracer.threads[p] = true
	}
	return &tracer, nil
}

func (t *Tracer) EnableVerbose() {
	t.verbose = true
}

func (t *Tracer) SetExeComparisonLength(length int) {
	t.exeCompareLength = length
}
func (t *Tracer) SetFollowForks(enable bool) {

	if enable {
		t.ptraceOptions = t.ptraceOptions | syscall.PTRACE_EVENT_FORK | syscall.PTRACE_EVENT_VFORK
	} else {
		t.ptraceOptions = t.ptraceOptions & ^(syscall.PTRACE_EVENT_FORK | syscall.PTRACE_EVENT_VFORK)
	}

	if t.verbose {
		log.Printf("SetFollowForks: %t, 0x%x", enable, t.ptraceOptions)
	}

}
func (t *Tracer) Start() {
	var ws syscall.WaitStatus
	var regs syscall.PtraceRegs

	sig_chan := make(chan os.Signal, 1)
	signal.Notify(sig_chan, syscall.SIGTERM)
	signal.Notify(sig_chan, syscall.SIGINT)

	go func() {
		for {
			sig := <-sig_chan
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				log.Println("Got SIGTERM/SIGINT SIGNAL")
				shutdownFlag = true
				//Send our main signal handler a USR2 signal, this will cause a blocking wait to return
				syscall.Kill(t.Process.Pid, syscall.SIGUSR2)
				// Give 5 seconds to shut down gracefully
				time.Sleep(5 * time.Second)
				log.Println("No exit detected yet, calling Exit")
				os.Exit(1)
			}
		}
	}()

	// At this point breakpoints should be configured. Let's continue all threads...
	t.continueAllThreads()

	for {
		var rusage syscall.Rusage
		// Wait for any trap from any thread
		wpid, err := syscall.Wait4(-1, &ws, syscall.WALL, &rusage)
		if t.verbose {
			log.Printf("PPID:%d / PID:%d wait4 returned... 0x%x, %v, %v, %v\n", t.Process.Pid, wpid, ws, ws.StopSignal(), ws.TrapCause(), err)
			log.Printf("-> signal: 0x%x\n", (ws>>8)&0xFF)
		}

		if err != nil {
			log.Fatalln("ERROR: ", err)
		}

		if shutdownFlag {
			log.Printf("%sDisable all breakpoints... %s", Red, Reset)
			for b := range t.breakpoints {
				breakPoint := t.breakpoints[b]
				replaceCode(wpid, breakPoint.Address, *breakPoint.OriginalCode)
			}
			log.Printf("%sDetaching from Process...%s", Red, Reset)
			syscall.PtraceDetach(wpid)
			os.Exit(0)
		}

		t.threads[wpid] = true

		if ws.Exited() == true {
			delete(t.threads, wpid)
			if t.verbose {
				log.Printf("Child pid %v finished.\n", wpid)
			}
			if len(t.threads) == 0 {
				break
			}
			continue
		}
		if ws.Signaled() == true {
			log.Printf("Error: Other pid signalled %v %v", wpid, ws)
			delete(t.threads, wpid)
			continue
		}
		err = syscall.PtraceGetRegs(wpid, &regs)
		if err != nil {
			log.Printf("Error (ptrace): %v", err)
			continue
		}

		switch uint32(ws) >> 8 {

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_SECCOMP << 8):
			if err != nil {
				log.Printf("Error (ptrace): %v", err)
				continue
			}
			log.Printf("SECCOMP REGS: %v\n", regs)
		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_EXIT << 8):
			if t.verbose {
				log.Printf("Ptrace exit event detected pid %v ", wpid)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_CLONE << 8):
			if t.verbose {
				log.Printf("Ptrace clone event detected pid %v ", wpid)
			}
			newPid := t.getEventMsg(wpid)
			t.threads[int(newPid)] = true
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_FORK << 8):
			if t.verbose {
				log.Printf("PTrace fork event detected pid %v ", wpid)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_VFORK << 8):
			if t.verbose {
				log.Printf("Ptrace vfork event detected pid %v ", wpid)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_VFORK_DONE << 8):
			if t.verbose {
				log.Printf("Ptrace vfork done event detected pid %v ", wpid)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_EXEC << 8):
			if t.verbose {
				log.Printf("Ptrace exec event detected pid %v ", wpid)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_STOP << 8):
			if t.verbose {
				log.Printf("Ptrace stop event detected pid %v ", wpid)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP):
			if t.verbose {
				log.Printf("SIGTRAP detected in pid %v ", wpid)
			}
			breakPoint, ok := t.breakpoints[uintptr(regs.Rip)-1]
			if ok {
				breakPoint.Hits += 1
				if t.verbose {
					msgId := t.getEventMsg(wpid)
					log.Printf("PID: %d (msg:%d) Hit Breakpoint at 0x%x (%d times)", wpid, msgId, breakPoint.Address, breakPoint.Hits)
				}

				replaceCode(wpid, breakPoint.Address, *breakPoint.OriginalCode)
				regs.Rip = uint64(breakPoint.Address)
				check(syscall.PtraceSetRegs(wpid, &regs))

				// Call the callback print handlers
				for idx := range breakPoint.Callbacks {
					cb := breakPoint.Callbacks[idx]
					cb(wpid, *breakPoint)
				}

				// we need to step forward once before setting the breakpoint again
				check(syscall.PtraceSingleStep(wpid))
				wpid, err = syscall.Wait4(wpid, &ws, syscall.WALL, nil)
				check(err)
				// set the breakpoint back again
				replaceCode(wpid, breakPoint.Address, []byte{0xCC})
			} else {
				log.Printf("Got SIGTRAP without known Breakpoint at 0x%x\n", regs.Rip)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGCHLD):
			if t.verbose {
				log.Printf("SIGCHLD detected pid %v ", wpid)
			}
			check(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGSTOP):
			if t.verbose {
				msg := t.getEventMsg(wpid)
				log.Printf("SIGSTOP detected pid %v, msg: %v ", wpid, msg)
			}
			check(syscall.PtraceCont(wpid, 0))
		case uint32(unix.SIGINT):
			if wpid == t.Process.Pid {
				log.Printf("SIGINT on PID %d, Start detaching and exit", wpid)
				for p := range t.threads {
					err = syscall.PtraceDetach(p)
					log.Printf("PID %d Detach returned: %v", p, err)
				}
				os.Exit(0)
			} else {
				log.Printf("SIGINT on child PID %d", wpid)
				check(syscall.PtraceCont(wpid, 0))
			}

		default:
			y := ws.StopSignal()
			log.Printf("Child stopped for unknown reasons pid %v status %v signal %d", wpid, ws, y)
			check(syscall.PtraceCont(wpid, int(ws.StopSignal())))
		}

	}
}

func (t *Tracer) continueAllThreads() {
	for p := range t.threads {
		if t.verbose {
			log.Printf("Setting configuration on pid: %d", p)
		}
		check(syscall.PtraceSetOptions(p, t.ptraceOptions))
		check(syscall.PtraceCont(p, 0))
	}
}

func (*Tracer) getEventMsg(wpid int) uint {
	msgID, err := syscall.PtraceGetEventMsg(wpid)
	check(err)
	return msgID
}

func (t *Tracer) GetBaseAddress() (uintptr, error) {
	if t.baseAddress > 0 {
		return t.baseAddress, nil
	} else {
		p, err := t.ProcFS.Proc(t.Process.Pid)

		if err != nil {
			log.Fatalln(err)
		}
		procMaps, err := p.ProcMaps()
		exePath := fmt.Sprintf("/proc/%d/exe", t.Process.Pid)
		exeMemPath := fmt.Sprintf("/proc/%d/mem", t.Process.Pid)

		exeData := readBytesFromFile(exePath, t.exeCompareLength, 0)

		for i := range procMaps {
			// Only check if we're at a base address
			if 0 == procMaps[i].Offset {
				if int(procMaps[i].EndAddr-procMaps[i].StartAddr) > t.exeCompareLength {
					memData := readBytesFromFile(exeMemPath, t.exeCompareLength, int64(procMaps[i].StartAddr))

					if 0 == bytes.Compare(exeData, memData) {
						if t.verbose {
							log.Printf("start:%x offset:%x, pathname: %s", procMaps[i].StartAddr, procMaps[i].Offset, procMaps[i].Pathname)
						}
						t.baseAddress = procMaps[i].StartAddr
						return t.baseAddress, nil
					}
				}
			}
		}
	}

	log.Fatalln("Unable to find the base address of the process")
	return 0, nil
}

func (t *Tracer) GetMemMaps() ([]*procfs.ProcMap, error) {
	p, err := t.ProcFS.Proc(t.Process.Pid)

	if err != nil {
		log.Fatalln(err)
	}
	procMaps, err := p.ProcMaps()

	return procMaps, err
}

func replaceCode(pid int, breakpoint uintptr, code []byte) []byte {
	original := make([]byte, len(code))
	_, err := syscall.PtracePeekData(pid, breakpoint, original)
	check(err)
	//log.Printf("peek: cnt: %d, err: %v, original %v", cnt, err, original)

	_, err = syscall.PtracePokeData(pid, breakpoint, code)
	//log.Printf("poke: cnt: %d, err: %v", cnt, err)
	check(err)

	return original
}

func (t *Tracer) ConvertOffsetToAddress(breakAddress uintptr) uintptr {
	baseAddress, err := t.GetBaseAddress()
	if err != nil {
		log.Fatalln(err)
	}

	bp := baseAddress + breakAddress
	return bp
}

func (t *Tracer) setBreakpoint(breakAddress uintptr, cb CallBackFunction) {
	bp := breakAddress

	breakpoint, ok := t.breakpoints[bp]

	if ok {
		log.Printf("Breakpoint at 0x%x already set, adding cb...", bp)
		breakpoint.Callbacks = append(breakpoint.Callbacks, cb)
	} else {
		log.Printf("Setting Breakpoint at 0x%x", bp)
		org := replaceCode(t.Process.Pid, bp, []byte{0xCC})

		callBacks := make([]CallBackFunction, 0)
		callBacks = append(callBacks, cb)

		t.breakpoints[bp] = &BreakPoint{Address: bp, OriginalCode: &org, Hits: 0, Callbacks: callBacks}
	}

	return
}

func (t *Tracer) SetBreakpointRelative(breakAddress uintptr, cb CallBackFunction) {
	bp := t.ConvertOffsetToAddress(breakAddress)
	t.setBreakpoint(bp, cb)
}

func (t *Tracer) SetBreakpointAbsolute(breakAddress uintptr, cb CallBackFunction) {
	t.setBreakpoint(breakAddress, cb)
}
