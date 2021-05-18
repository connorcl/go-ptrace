package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const TestRuns = 30

func main() {
	testPerformance(TestRuns)
}

func testExecutionControlAndRegisterAccess() {
	programName := os.Args[1]
	fmt.Println(programName)
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid
	fmt.Println(pid)
	time.Sleep(time.Second * 1)

	process, _ := MakePtraceProcess(pid, false, false)
	process.Wait(true)

	ip1, _ := process.GetInstrPointer()
	fmt.Printf("IP: 0x%X\n", ip1)
	ip2, _ := process.GetReg("rip")
	if ip1 == ip2 {
		fmt.Println("IPs match!")
	}

	for i := 0; i < 10; i++ {
		process.SingleStep()
		process.Wait(true)
		ip1, _ = process.GetInstrPointer()
		fmt.Printf("IP: 0x%X\n", ip1)
	}

	process.SetReg("Rcx", 0x123)
	rax, _ := process.GetReg("rcx")
	if rax == 0x123 {
		fmt.Println("Register set successfully!")
	}

	process.Cont()
	process.Terminate()
	process.Wait(true)
	process.Detach()
	fmt.Println("Process terminated successfully!")
}

func testMemoryAccess() {
	programName := os.Args[1]
	fmt.Println(programName)
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid
	fmt.Println(pid)
	time.Sleep(time.Second * 1)

	process, _ := MakePtraceProcess(pid, false, false)
	process.Wait(true)

	process.DumpMaps()

	maps, _ := process.ReadMappings()
	m := maps[0]
	w, _ := process.ReadWord(m.Start)
	fmt.Printf("0x%X\n", w)
	w += 5
	process.WriteWord(m.Start, w)

	data, _ := process.ReadBytes(m.Start, int(m.End) - int(m.Start))
	data2, _ := process.ReadMem(m.Start, int(m.End) - int(m.Start))

	if data2[0] == data[0] && data[0] == byte(w) {
		fmt.Println("Memory reads successful!")
	}

	data[0] = 12
	process.WriteBytes(m.Start, data)
	data, _ = process.ReadBytes(m.Start, (int(m.End) - int(m.Start)))
	if data[0] == 12 {
		fmt.Println("Memory write successful!")
	}

	process.Terminate()
	process.Wait(true)
	process.Detach()
	fmt.Println("Process terminated successfully")
}

func testBreakpoints() {
	programName := os.Args[1]
	fmt.Println(programName)
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid
	fmt.Println(pid)
	time.Sleep(time.Second * 1)

	process, _ := MakePtraceProcess(pid, false, false)
	process.Wait(true)

	ip, _ := process.GetInstrPointer()
	// offset depends on test program
	breakpointAddr := ((ip / 0x1000) * 0x1000) + 0x248
	fmt.Printf("IP: 0x%X\n", ip)
	fmt.Printf("Breakpoint Address: 0x%X\n", breakpointAddr)

	err := process.CreateBreakpoint(uintptr(breakpointAddr))
	if err != nil {
		fmt.Println(err.Error())
	}
	process.Cont()
	process.Wait(true)
	ip, _ = process.GetInstrPointer()
	if ip == breakpointAddr + 1 {
		fmt.Println("Breakpoint hit successfully!")
	}
	err = process.RemoveBreakpoint(uintptr(breakpointAddr), true)
	if err != nil {
		fmt.Println(err.Error())
	}
	ip, _ = process.GetInstrPointer()
	if ip == breakpointAddr {
		fmt.Println("Breakpoint removed successfully!")
	}
	process.Cont()
	process.Wait(true)
}

func testStackManagement() {
	programName := os.Args[1]
	fmt.Println(programName)
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid
	fmt.Println(pid)
	time.Sleep(time.Second * 1)

	process, _ := MakePtraceProcess(pid, false, false)
	process.Wait(true)

	stackMapping, err := process.FindStack()
	if err == nil {
		fmt.Printf("Stack found successfully: %s\n", stackMapping.String())
	}
	fmt.Println("Printing stack...")
	process.DumpStack(6)
	fmt.Println("Printing backtrace...")
	backtrace := process.GetBacktrace(6, 25)
	backtrace.Dump()

	process.Terminate()
	process.Wait(true)
	process.Detach()
	fmt.Println("Process terminated successfully")
}

func testPerformance(iterations int) {
	resultsFile, err := os.OpenFile("timings_go.csv", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	writer := bufio.NewWriter(resultsFile)
	writer.WriteString("Attach,Single Step,Breakpoint,Registers,Memory,Memory Search,Total\n")
	defer writer.Flush()

	for i := 0; i < iterations; i++ {
		fmt.Printf("%d\n", i)
		// start child process
		programName := os.Args[1]
		fmt.Println(programName)
		cmd := exec.Command(programName)
		cmd.Start()
		pid := cmd.Process.Pid
		fmt.Println(pid)
		time.Sleep(time.Second * 1)
		start := time.Now()

		// attach to child process
		process, _ := MakePtraceProcess(pid, false, false)
		process.Wait(true)

		// get ip and calculate a pointer to set a breakpoint at
		eip, _ := process.GetInstrPointer()
		breakpointAddr := ((eip / 0x1000) * 0x1000) + 0x228

		initEndTime := int(time.Since(start).Milliseconds())
		initTime := initEndTime

		// single step 1000 times
		for j := 0; j < 1000; j++ {
			process.SingleStep()
			process.Wait(true)
		}

		singleStepEndTime := int(time.Since(start).Milliseconds())
		singleStepTime := singleStepEndTime - initEndTime

		// create bp, run and then remove breakpoint 1000 times
		for j := 0; j < 1000; j++ {
			process.CreateBreakpoint(uintptr(breakpointAddr))
			process.Cont()
			process.Wait(true)
			process.RemoveBreakpoint(uintptr(breakpointAddr), true)
		}

		breakpointEndTime := int(time.Since(start).Milliseconds())
		breakpointTime := breakpointEndTime - singleStepEndTime

		// read and write registers
		for j := 0; j < 1000; j++ {
			regs, _ := process.GetRegs()
			regs.Rax = regs.Rbx
			regs.Rip = regs.Rsp
			process.SetRegs(&regs)
		}

		regEndTime := int(time.Since(start).Milliseconds())
		regTime := regEndTime - breakpointEndTime

		// read and write memory
		sp, _ := process.GetStackPointer()
		for j := 0; j < 1000; j++ {
			w, _ := process.ReadWord(uintptr(sp))
			w += 5
			process.WriteWord(uintptr(sp), w)
		}

		memEndTime := int(time.Since(start).Milliseconds())
		memTime := memEndTime - regEndTime

		// search memory
		for j := 0; j < 100; j++ {
			m, _ := process.ReadMappings()
			done := false
			for _, v := range m {
				for a := v.Start; a < v.End && !done; a += 8 {
					w, _ := process.ReadWord(a)
					if w == 6464 {
						done = true
					}
				}
				if done {
					break
				}
			}
		}

		searchEndTime := int(time.Since(start).Milliseconds())
		searchTime := searchEndTime - memEndTime
		totalTime := searchEndTime

		process.Terminate()
		process.Wait(true)
		process.Detach()

		results := fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d",
			initTime,
			singleStepTime,
			breakpointTime,
			regTime, memTime,
			searchTime,
			totalTime)
		writer.WriteString(results + "\n")
	}
}