package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	programName := os.Args[1]
	fmt.Println(programName)
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid
	fmt.Println(pid)
	time.Sleep(time.Second * 1)

	process, _ := makePtraceProcess(pid, false, false)
	process.Wait()

	eip, _ := process.GetInstrPointer()
	fmt.Printf("IP: 0x%X\n", eip)
	process.SingleStep()
	process.Wait()
	eip, _ = process.GetInstrPointer()
	fmt.Printf("IP: 0x%X\n", eip)

	b := GetBacktrace(&process, 6, 25)
	DumpBacktrace(&b)

	//eip, _ := process.GetInstrPointer()
	//fmt.Printf("IP: 0x%X\n", eip)
	//predictedEip := eip  -4;
	//process.CreateBreakpoint(uintptr(predictedEip))
	//process.Cont()
	//process.Wait()
	//eip, _ = process.GetInstrPointer()
	//fmt.Printf("IP: 0x%X\n", eip)
	//process.RemoveBreakpoint(uintptr(predictedEip), true)
	//eip, _ = process.GetInstrPointer()
	//fmt.Printf("IP: 0x%X\n", eip)
	//process.Cont()
	//process.Wait()

	//process.SingleStep()
	//process.Wait()
	//fmt.Printf("Predicted next IP: 0x%X\n", predictedEip)
	//eip, _ = process.GetInstrPointer()
	//fmt.Printf("Next IP: 0x%X\n", eip)

	//process.CreateBreakpoint(0x55C290A40173)
	//process.Cont()
	//process.Wait()
	//eip, _ = process.GetInstrPointer()
	//fmt.Printf("IP: 0x%X\n", eip)

	//maps := process.ReadMappings()
	//// for _, m := range maps {
	//// 	fmt.Printf("%+v\n", m)
	//// }
	//DumpMaps(maps)
	//word, err := process.ReadWord(maps[0].Start)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%X\n", word)
	//}
	//s, _ := process.ReadCString(maps[0].Start)
	//fmt.Println(s)
	//fmt.Println("---------------------------")
	//bytes, _ := process.ReadBytes(maps[0].Start, 4)
	//fmt.Printf("%v\n", bytes)
	//err = process.WriteBytes(maps[0].Start, []byte{0, 0, 0, 0})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//word, _ = process.ReadWord(maps[0].Start)
	//fmt.Printf("%x\n", word)
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
