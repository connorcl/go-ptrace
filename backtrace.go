package main

import (
	"fmt"
	"math"
)

type BacktraceFrame struct {
	Ip        uintptr
	Name      string
	Arguments []uint64
}

func CreateBacktraceFrame(p *PtraceProcess, ip uintptr, fp uintptr, nArgs int) BacktraceFrame {
	frame := BacktraceFrame{
		Ip:        ip,
		Name:      "???",
		Arguments: make([]uint64, nArgs),
	}
	address := fp + 8
	for i := 0; i < nArgs; i++ {
		address += 8
		word, err := p.ReadWord(address)
		if err == nil {
			frame.Arguments[i] = word
		}
	}
	return frame
}

type Backtrace struct {
	Frames    []BacktraceFrame
	Truncated bool
}

func (p *PtraceProcess) GetBacktrace(maxArgs int, maxDepth int) Backtrace {
	backtrace := Backtrace{
		Frames:    make([]BacktraceFrame, 0),
		Truncated: false,
	}

	// get instruction and frame pointer
	ip, _ := p.GetInstrPointer()
	fp, _ := p.GetFramePointer()
	depth := 0
	for {
		if maxDepth <= depth {
			backtrace.Truncated = true
			break
		}

		nextFp, err := p.ReadWord(uintptr(fp))
		if err != nil {
			//fmt.Println("Next FP err!")
			break
		}

		// Guess the number of function arguments
		var nArgs int
		//fmt.Printf("Next FP: 0x%X\n", nextFp)
		//fmt.Printf("FP: 0x%X\n", fp)
		nArgs = int(((nextFp - fp) / 8) - 2)
		//fmt.Printf("Args guess: %d\n", nArgs)
		if nArgs > maxArgs {
			nArgs = maxArgs
		}

		frame := CreateBacktraceFrame(p, uintptr(ip), uintptr(fp), nArgs)
		backtrace.Frames = append(backtrace.Frames, frame)

		nextIp, _ := p.ReadWord(uintptr(fp + 8))
		//fmt.Printf("Next IP: 0x%X\n", nextIp)
		//fmt.Printf("IP: 0x%X\n", ip)
		ip = nextIp
		if ip == math.MaxUint64 {
			//fmt.Println("Max int")
			break
		}
		fp = nextFp
		depth += 1
		//fmt.Println("-----------------------")
	}
	return backtrace
}

func (f BacktraceFrame) String() string {
	return fmt.Sprintf("IP: 0x%X, Args: %v", f.Ip, f.Arguments)
}

func (b *Backtrace) Dump() {
	for _, f := range b.Frames {
		fmt.Println(f)
	}
}
