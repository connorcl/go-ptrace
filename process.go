package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type PtraceProcess struct {
	Pid        int
	IsAttached bool
	IsThread   bool
	Breakpoints map[uintptr]Breakpoint
}

func makePtraceProcess(pid int, isAttached bool, isThread bool) (PtraceProcess, error) {
	var err error = nil
	// attach to process if necessary and record error
	if !isAttached {
		err = unix.PtraceAttach(pid)
	}
	// create process struct
	process := PtraceProcess{
		Pid:        pid,
		IsAttached: true,
		IsThread:   isThread,
		Breakpoints: make(map[uintptr]Breakpoint),
	}
	// return process struct and error
	return process, err
}

func (p *PtraceProcess) SingleStep() error {
	return unix.PtraceSingleStep(p.Pid)
}

func (p *PtraceProcess) SignalAndCont(signum unix.Signal) error {
	return unix.PtraceCont(p.Pid, int(signum))
}

func (p *PtraceProcess) Cont() error {
	return unix.PtraceCont(p.Pid, 0)
}

func (p *PtraceProcess) SignalAndSyscall(signum unix.Signal) error {
	return unix.PtraceSyscall(p.Pid, int(signum))
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
	err := unix.PtraceDetach(p.Pid)
	p.IsAttached = false
	return err
}

func (p *PtraceProcess) Wait() error {
	var wstatus unix.WaitStatus
	var rusage unix.Rusage
	_, err := unix.Wait4(p.Pid, &wstatus, 0, &rusage)
	return err
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

func (p *PtraceProcess) ReadWord(address uintptr) (uint32, error) {
	word := make([]byte, 4)
	_, err := unix.PtracePeekData(p.Pid, address, word)
	return binary.LittleEndian.Uint32(word), err
}

func (p *PtraceProcess) WriteWord(address uintptr, value uint32) error {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, value)
	_, err := unix.PtracePokeData(p.Pid, address, data)
	return err
}

func (p *PtraceProcess) ReadBytes(address uintptr, size uint) ([]byte, error) {
	buf := make([]byte, size)
	localVec := []unix.Iovec{{
		Base: &buf[0],
		Len:  uint64(size),
	}}
	remoteVec := []unix.RemoteIovec{{
		Base: address,
		Len:  int(size),
	}}
	_, err := unix.ProcessVMReadv(p.Pid, localVec, remoteVec, 0)
	return buf, err
}

func (p *PtraceProcess) WriteBytes(address uintptr, data []byte) error {
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
	var count uint = 0
	var l uint = 4
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

func (p *PtraceProcess) ReadMappings() []MemoryMapping {
	return ReadProcessMappings(p.Pid)
}

func (p *PtraceProcess) FindBreakpoint(address uintptr) (Breakpoint, bool) {
	b, ok := p.Breakpoints[address]
	return b, ok
}

func (p *PtraceProcess) CreateBreakpoint(address uintptr) {
	_, ok := p.FindBreakpoint(address)
	if !ok {
		fmt.Println("Process: Setting breakpoint...")
		b := CreateBreakpoint(address)
		b.Install(p)
		p.Breakpoints[address] = b
	}
}

func (p* PtraceProcess) RemoveBreakpoint(address uintptr, setIp bool) {
	b, ok := p.FindBreakpoint(address)
	if ok {
		b.Deinstall(p, setIp)
		delete(p.Breakpoints, b.Address)
	}
}