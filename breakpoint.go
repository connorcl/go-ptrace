package main

import "fmt"

type Breakpoint struct {
	Installed  bool
	Address    uintptr
	SavedBytes uint64
}

func CreateBreakpoint(address uintptr) Breakpoint {
	return Breakpoint{
		Installed:  false,
		Address:    address,
		SavedBytes: 0,
	}
}

func (b *Breakpoint) Install(p *PtraceProcess) error {
	word, err := p.ReadWord(b.Address)
	if err == nil {
		// save old instruction bytes
		b.SavedBytes = word
		// install INT3 instruction(s)
		err = p.WriteWord(b.Address, 0xCCCCCCCCCCCCCCCC)
		if err != nil {
			errMsg := fmt.Sprintf("Error installing breakpoint at address 0x%X: %s\n", b.Address, err.Error())
			return MakeGenericError(errMsg)
		}
		b.Installed = true
		return nil
	} else {
		errMsg := fmt.Sprintf("Error reading bytes at address 0x%X: %s\n", b.Address, err.Error())
		return MakeGenericError(errMsg)
	}
}

func (b *Breakpoint) Deinstall(p *PtraceProcess, setIp bool) error {
	if b.Installed {
		b.Installed = false
		err := p.WriteWord(b.Address, b.SavedBytes)
		if err != nil {
			errMsg := fmt.Sprintf("Error restoring saved bytes to address 0x%X: %s", b.Address, err.Error())
			return MakeGenericError(errMsg)
		}
		if setIp {
			err = p.SetInstrPointer(uint64(b.Address))
			if err != nil {
				errMsg := fmt.Sprintf("Error restoring IP to 0x%X: %s", b.Address, err.Error())
				return MakeGenericError(errMsg)
			}
		}
		return nil
	}
	errMsg := "Breakpoint not installed!"
	return MakeGenericError(errMsg)
}
