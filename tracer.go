package rip_tracer

import (
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

type BreakPoint struct {
	Address      uintptr
	OriginalCode *[]byte
	Hits         int
	Callbacks    []func(int)
}

type Tracer struct {
	Process     *os.Process
	ProcFS      procfs.FS
	ws          syscall.WaitStatus
	breakpoints map[uintptr]*BreakPoint
	threads     map[int]bool
	verbose     bool
}

func must(err error) {
	if err != nil {
		panic(err)
	}
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
	must(cmd.Start())

	cmd.Wait() //Ignore the error, we hit our starting breakpoint trap

	must(syscall.PtraceSetOptions(cmd.Process.Pid, syscall.PTRACE_O_TRACECLONE))

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
		Process:     cmd.Process,
		ProcFS:      procFS,
		breakpoints: make(map[uintptr]*BreakPoint),
		threads:     threads,
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
	must(err)

	tracer := Tracer{
		Process:     proc,
		ProcFS:      procFS,
		breakpoints: make(map[uintptr]*BreakPoint),
		threads:     make(map[int]bool),
	}

	for i := range all_pids {
		p := all_pids[i].PID

		_attachToPid(p)

		for {
			_, err := syscall.Wait4(p, &ws, syscall.WALL, nil)
			if ws.Stopped() {
				must(err)
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
				//Send our main signal handler a USR2 signal
				syscall.Kill(t.Process.Pid, syscall.SIGUSR2)
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

		if ws.StopSignal() == syscall.SIGUSR2 {

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
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_CLONE << 8):
			if t.verbose {
				log.Printf("Ptrace clone event detected pid %v ", wpid)
			}
			newPid := t.getEventMsg(wpid)
			t.threads[int(newPid)] = true
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_FORK << 8):
			if t.verbose {
				log.Printf("PTrace fork event detected pid %v ", wpid)
			}
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_VFORK << 8):
			if t.verbose {
				log.Printf("Ptrace vfork event detected pid %v ", wpid)
			}
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_VFORK_DONE << 8):
			if t.verbose {
				log.Printf("Ptrace vfork done event detected pid %v ", wpid)
			}
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_EXEC << 8):
			if t.verbose {
				log.Printf("Ptrace exec event detected pid %v ", wpid)
			}
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP) | (unix.PTRACE_EVENT_STOP << 8):
			if t.verbose {
				log.Printf("Ptrace stop event detected pid %v ", wpid)
			}
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGTRAP):
			if t.verbose {
				log.Printf("SIGTRAP detected in pid %v ", wpid)
			}
			breakPoint, ok := t.breakpoints[uintptr(regs.Rip)-1]
			if ok {
				breakPoint.Hits += 1
				msgId := t.getEventMsg(wpid)
				log.Printf("PID: %d (msg:%d) Hit Breakpoint at 0x%x (%d times)", wpid, msgId, breakPoint.Address, breakPoint.Hits)

				replaceCode(wpid, breakPoint.Address, *breakPoint.OriginalCode)
				regs.Rip = uint64(breakPoint.Address)
				must(syscall.PtraceSetRegs(wpid, &regs))

				// Call the callback print handlers
				for idx := range breakPoint.Callbacks {
					cb := breakPoint.Callbacks[idx]
					cb(wpid)
				}

				// we need to step forward once before setting the breakpoint again
				must(syscall.PtraceSingleStep(wpid))
				wpid, err = syscall.Wait4(wpid, &ws, syscall.WALL, nil)
				must(err)
				// set the breakpoint back again
				replaceCode(wpid, breakPoint.Address, []byte{0xCC})
			} else {
				log.Printf("Got SIGTRAP without known Breakpoint at 0x%x\n", regs.Rip)
			}
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGCHLD):
			if t.verbose {
				log.Printf("SIGCHLD detected pid %v ", wpid)
			}
			must(syscall.PtraceCont(wpid, 0))

		case uint32(unix.SIGSTOP):
			if t.verbose {
				msg := t.getEventMsg(wpid)
				log.Printf("SIGSTOP detected pid %v, msg: %v ", wpid, msg)
			}
			must(syscall.PtraceCont(wpid, 0))
		case uint32(unix.SIGINT):
			log.Println("SIGINT, start detaching and exit")
			for p := range t.threads {
				err = syscall.PtraceDetach(p)
				log.Printf("PID %d Detach returned: %v", p, err)
			}
			os.Exit(0)
		default:
			y := ws.StopSignal()
			log.Printf("Child stopped for unknown reasons pid %v status %v signal %d", wpid, ws, y)
			must(syscall.PtraceCont(wpid, int(ws.StopSignal())))
		}

	}
}

func (t *Tracer) continueAllThreads() {
	for p := range t.threads {
		if t.verbose {
			log.Printf("Setting configuration on pid: %d", p)
		}
		must(syscall.PtraceSetOptions(p, syscall.PTRACE_O_TRACECLONE))
		must(syscall.PtraceCont(p, 0))
	}
}

func (*Tracer) getEventMsg(wpid int) uint {
	msgID, err := syscall.PtraceGetEventMsg(wpid)
	must(err)
	return msgID
}

func (t *Tracer) GetBaseAddress() (uintptr, error) {
	p, err := t.ProcFS.Proc(t.Process.Pid)

	if err != nil {
		log.Fatalln(err)
	}
	procMaps, err := p.ProcMaps()
	cmdline, err := p.CmdLine()
	for i := range procMaps {
		if 0 == procMaps[i].Offset && strings.Contains(procMaps[i].Pathname, cmdline[0]) {
			log.Printf("start:%x offset:%x, pathname: %s", procMaps[i].StartAddr, procMaps[i].Offset, procMaps[i].Pathname)
			return procMaps[i].StartAddr, nil
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
	must(err)
	//log.Printf("peek: cnt: %d, err: %v, original %v", cnt, err, original)

	_, err = syscall.PtracePokeData(pid, breakpoint, code)
	//log.Printf("poke: cnt: %d, err: %v", cnt, err)
	must(err)

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

func (t *Tracer) SetBreakpoint(breakAddress uintptr, cb func(int), absolute bool) {
	bp := breakAddress
	if false == absolute {
		bp = t.ConvertOffsetToAddress(breakAddress)
	}

	breakpoint, ok := t.breakpoints[bp]

	if ok {
		log.Printf("Breakpoint at 0x%x already set, adding cb...", bp)
		breakpoint.Callbacks = append(breakpoint.Callbacks, cb)
	} else {
		log.Printf("Setting Breakpoint at 0x%x", bp)
		org := replaceCode(t.Process.Pid, bp, []byte{0xCC})

		callBacks := make([]func(int), 0)
		callBacks = append(callBacks, cb)

		t.breakpoints[bp] = &BreakPoint{Address: bp, OriginalCode: &org, Hits: 0, Callbacks: callBacks}
	}

	return
}
