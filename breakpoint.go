package main

import "fmt"

type Breakpoint struct {
	Installed bool
	Address uintptr
	SavedBytes uint32
}

func CreateBreakpoint(address uintptr) Breakpoint {
	return Breakpoint{
		Installed: false,
		Address:   address,
		SavedBytes: 0,
	}
}

func (b *Breakpoint) Install(p *PtraceProcess) {
	word, err := p.ReadWord(b.Address)
	if err == nil {
		fmt.Printf("Installing breakpoint at 0x%X\n", b.Address)
		// save old instruction bytes
		b.SavedBytes = word
		// install INT3 instruction(s)
		err = p.WriteWord(b.Address, 0xCCCCCCCC)
		if err != nil {
			fmt.Printf("Error writing memory when installing breakpoint: %s\n", err)
		}
		b.Installed = true
	} else {
		fmt.Printf("Error reading memory when installing breakpoint: %s\n", err)
	}
}

func (b *Breakpoint) Deinstall(p *PtraceProcess, setIp bool) {
	if b.Installed {
		b.Installed = false
		fmt.Printf("Restoring bytes 0x%X to 0x%X\n", b.SavedBytes, b.Address)
		err := p.WriteWord(b.Address, b.SavedBytes)
		if err != nil {
			fmt.Println(err)
		}
		if setIp {
			p.SetInstrPointer(uint64(b.Address))
		}
	}
}
