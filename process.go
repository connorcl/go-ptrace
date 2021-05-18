package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type GenericError struct {
	Message string
}

func (p GenericError) Error() string {
	return p.Message
}

func MakeGenericError(message string) GenericError {
	return GenericError{
		Message: message,
	}
}

func filterSignal(signal unix.Signal) unix.Signal {
	if signal == syscall.SIGTRAP {
		return 0
	} else {
		return signal
	}
}

const WordSize = 8

type PtraceProcess struct {
	Pid         int
	IsAttached  bool
	IsThread    bool
	Breakpoints map[uintptr]Breakpoint
}

func MakePtraceProcess(pid int, isAttached bool, isThread bool) (PtraceProcess, error) {
	var err error = nil
	// attach to process if necessary and record error
	if !isAttached {
		attachErr := unix.PtraceAttach(pid)
		if attachErr != nil {
			errMsg := fmt.Sprintf("Could not attach to PID %d: %s", pid, attachErr.Error())
			err = MakeGenericError(errMsg)
		} else {
			err = nil
		}
	}
	// create process struct
	process := PtraceProcess{
		Pid:         pid,
		IsAttached:  true,
		IsThread:    isThread,
		Breakpoints: make(map[uintptr]Breakpoint),
	}
	// return process struct and error
	return process, err
}

func (p *PtraceProcess) SingleStep() error {
	return unix.PtraceSingleStep(p.Pid)
}

func (p *PtraceProcess) SignalAndCont(signum unix.Signal) error {
	return unix.PtraceCont(p.Pid, int(filterSignal(signum)))
}

func (p *PtraceProcess) Cont() error {
	return unix.PtraceCont(p.Pid, 0)
}

func (p *PtraceProcess) SignalAndSyscall(signum unix.Signal) error {
	return unix.PtraceSyscall(p.Pid, int(filterSignal(signum)))
}

func (p *PtraceProcess) Syscall() error {
	return unix.PtraceSyscall(p.Pid, 0)
}

func (p *PtraceProcess) Kill(signum unix.Signal) error {
	return unix.Kill(p.Pid, signum)
}

func (p *PtraceProcess) Terminate() error {
	return unix.Kill(p.Pid, syscall.SIGKILL)
}

func (p *PtraceProcess) GetRegs() (unix.PtraceRegs, error) {
	var regs unix.PtraceRegs
	err := unix.PtraceGetRegs(p.Pid, &regs)
	return regs, err
}

func (p *PtraceProcess) GetReg(name string) (uint64, error) {
	name = strings.Title(name)
	regs, err := p.GetRegs()
	if err != nil {
		return 0, err
	}
	v := reflect.ValueOf(&regs).Elem().FieldByName(name)
	if v.IsValid() {
		return v.Uint(), nil
	}
	return 0, nil
}

func (p *PtraceProcess) SetReg(name string, value uint64) error {
	name = strings.Title(name)
	regs, err := p.GetRegs()
	if err != nil {
		return err
	}
	v := reflect.ValueOf(&regs).Elem().FieldByName(name)
	if v.IsValid() {
		v.SetUint(value)
		p.SetRegs(&regs)
		return nil
	}
	return nil
}

func (p *PtraceProcess) SetInstrPointer(ip uint64) error {
	regs, err := p.GetRegs()
	if err != nil {
		return err
	}
	regs.Rip = ip
	return p.SetRegs(&regs)
}

func (p *PtraceProcess) GetInstrPointer() (uint64, error) {
	regs, err := p.GetRegs()
	return regs.Rip, err
}

func (p *PtraceProcess) GetStackPointer() (uint64, error) {
	regs, err := p.GetRegs()
	return regs.Rsp, err
}

func (p *PtraceProcess) GetFramePointer() (uint64, error) {
	regs, err := p.GetRegs()
	return regs.Rbp, err
}

func (p *PtraceProcess) SetRegs(regs *unix.PtraceRegs) error {
	return unix.PtraceSetRegs(p.Pid, regs)
}

func (p *PtraceProcess) Detach() error {
	if !p.IsAttached {
		return nil
	}
	err := unix.PtraceDetach(p.Pid)
	p.IsAttached = false
	return err
}

func (p *PtraceProcess) Wait(blocking bool) (unix.WaitStatus, error) {
	options := 0
	if !blocking {
		options |= unix.WNOHANG
	}
	var wstatus unix.WaitStatus
	var rusage unix.Rusage
	_, err := unix.Wait4(p.Pid, &wstatus, options, &rusage)
	return wstatus, err
}

func (p *PtraceProcess) DumpRegs() {
	regs, _ := p.GetRegs()
	s := reflect.ValueOf(&regs).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%-8s = 0x%012X\n",
			typeOfT.Field(i).Name, f.Interface())
	}
}

func (p *PtraceProcess) ReadWord(address uintptr) (uint64, error) {
	word := make([]byte, 8)
	_, err := unix.PtracePeekData(p.Pid, address, word)
	return binary.LittleEndian.Uint64(word), err
}

func (p *PtraceProcess) WriteWord(address uintptr, value uint64) error {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, value)
	_, err := unix.PtracePokeData(p.Pid, address, data)
	return err
}

func (p *PtraceProcess) ReadBytes(address uintptr, size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := unix.PtracePeekData(p.Pid, address, bytes)
	return bytes, err
}

func (p *PtraceProcess) WriteBytes(address uintptr, bytes []byte) error {
	_, err := unix.PtracePokeData(p.Pid, address, bytes)
	return err
}

func (p *PtraceProcess) ReadMem(address uintptr, size int) ([]byte, error) {
	buf := make([]byte, size)
	localVec := []unix.Iovec{{
		Base: &buf[0],
		Len:  uint64(size),
	}}
	remoteVec := []unix.RemoteIovec{{
		Base: address,
		Len:  size,
	}}
	_, err := unix.ProcessVMReadv(p.Pid, localVec, remoteVec, 0)
	return buf, err
}

func (p *PtraceProcess) WriteMem(address uintptr, data []byte) error {
	size := len(data)
	localVec := []unix.Iovec{{
		Base: &data[0],
		Len:  uint64(size),
	}}
	remoteVec := []unix.RemoteIovec{{
		Base: address,
		Len:  size,
	}}
	_, err := unix.ProcessVMWritev(p.Pid, localVec, remoteVec, 0)
	return err
}

func (p *PtraceProcess) ReadCString(address uintptr) (string, error) {
	var buf []byte
	count := 0
	l := 4
	for {
		word, err := p.ReadBytes(address+uintptr(count*l), l)
		if err != nil {
			return "", err
		}
		count++
		for _, b := range word {
			buf = append(buf, b)
			if b == 0 {
				return string(buf), nil
			}
		}
	}
}

func (p *PtraceProcess) ReadMappings() ([]MemoryMapping, error) {
	return ReadProcessMappings(p.Pid)
}

func (p *PtraceProcess) DumpMaps() {
	maps, _ := p.ReadMappings()
	for _, m := range maps {
		fmt.Println(m)
	}
}

func (p *PtraceProcess) FindBreakpoint(address uintptr) (Breakpoint, bool) {
	b, ok := p.Breakpoints[address]
	return b, ok
}

func (p *PtraceProcess) CreateBreakpoint(address uintptr) error {
	_, ok := p.FindBreakpoint(address)
	if !ok {
		b := CreateBreakpoint(address)
		err := b.Install(p)
		if err != nil {
			return MakeGenericError(err.Error())
		}
		p.Breakpoints[address] = b
		return nil
	} else {
		return MakeGenericError("Breakpoint already exists!")
	}
}

func (p *PtraceProcess) RemoveBreakpoint(address uintptr, setIp bool) error {
	b, ok := p.FindBreakpoint(address)
	if ok {
		err := b.Deinstall(p, setIp)
		if err != nil {
			return MakeGenericError(err.Error())
		}
		delete(p.Breakpoints, b.Address)
		return nil
	} else {
		return MakeGenericError("Breakpoint does not exist!")
	}
}

func (p *PtraceProcess) FindStack() (MemoryMapping, error) {
	maps, err := p.ReadMappings()
	if err != nil {
		return MemoryMapping{}, err
	}
	for _, m := range maps {
		if m.Pathname == "[stack]" {
			return m, nil
		}
	}
	return MemoryMapping{}, MakeGenericError("No stack found!")
}

func (p *PtraceProcess) DumpStack(len int) error {
	// get address of top of stack
	sp, err := p.GetStackPointer()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read stack pointer: %s", err.Error())
		return MakeGenericError(errMsg)
	}
	var addr uintptr
	for i := 0; i < len; i++ {
		addr = uintptr(sp) + uintptr(i*8)
		word, err := p.ReadWord(addr)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read byte at 0x%X: %s\n", addr, err.Error())
			return MakeGenericError(errMsg)
		}
		fmt.Printf("%08X\n", word)
	}
	return nil
}

func (p *PtraceProcess) SetOptions(options int) error {
	return unix.PtraceSetOptions(p.Pid, options)
}
