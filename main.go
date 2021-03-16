package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	programName := os.Args[1]
	fmt.Println(programName)
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid
	fmt.Println(pid)
	process, _ := makePtraceProcess(pid, false, false)
	process.Wait()
	// reg, _ := process.GetReg("rax")
	// fmt.Printf("%X\n", reg)
	// process.SetReg("rax", 0x123)
	// reg, _ = process.GetReg("rax")
	// fmt.Printf("%X\n", reg)
	eip, _ := process.GetInstrPointer()
	process.DumpRegs()
	fmt.Printf("%X\n", eip)
	process.SingleStep()
	process.Wait()
	eip, _ = process.GetInstrPointer()
	fmt.Printf("%X\n", eip)
}
