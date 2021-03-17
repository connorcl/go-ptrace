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
	maps := process.ReadMappings()
	// for _, m := range maps {
	// 	fmt.Printf("%+v\n", m)
	// }
	DumpMaps(maps)
	word, err := process.ReadWord(maps[0].Start)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%X\n", word)
	}
	s, _ := process.ReadCString(maps[0].Start)
	fmt.Println(s)
	fmt.Println("---------------------------")
	bytes, _ := process.ReadBytes(maps[0].Start, 4)
	fmt.Printf("%v\n", bytes)
	err = process.WriteBytes(maps[0].Start, []byte{0, 0, 0, 0})
	if err != nil {
		fmt.Println(err)
	}
	word, _ = process.ReadWord(maps[0].Start)
	fmt.Printf("%x\n", word)
	// reg, _ := process.GetReg("rax")
	// fmt.Printf("%X\n", reg)
	// process.SetReg("rax", 0x123)
	// reg, _ = process.GetReg("rax")
	// fmt.Printf("%X\n", reg)
	// eip, _ := process.GetInstrPointer()
	// process.DumpRegs()
	// fmt.Printf("%X\n", eip)
	// process.SingleStep()
	// process.Wait()
	// eip, _ = process.GetInstrPointer()
	// fmt.Printf("%X\n", eip)
}
