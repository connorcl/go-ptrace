import os
import subprocess
import signal
import time
import ptrace.debugger

os.chdir("/mnt/hgfs/cmp320/go-ptrace")
program_name = "/home/connor/test-program/test"

TEST_RUNS = 30


results_file = open('timings_python.csv', 'w')
results_file.write("Attach,Single Step,Breakpoint,Registers,Memory,Memory Search,Total\n")

for i in range(TEST_RUNS):
    # start child process
    args = [program_name]
    p = subprocess.Popen(args)
    print(p.pid)
    time.sleep(1)
    debugger = ptrace.debugger.PtraceDebugger()
    
    start = time.time()
    
    # attach to child process
    process = debugger.addProcess(p.pid, False)
    
    # get ip and calculate pointer to set breakpoint at
    ip = process.getInstrPointer()
    breakpoint_addr = ((ip // 0x1000) * 0x1000) + 0x248
    
    init_end_time = (time.time() - start) * 1000
    init_time = init_end_time
    
    #print("0x{:X}".format(breakpoint_addr))
    
    # single step
    for i in range(1000):
        process.singleStep()
        process.waitSignals(signal.SIGTRAP)
        #ip = process.getInstrPointer()
        #print("0x{:X}".format(ip))
        
    single_step_end_time = (time.time() - start) * 1000
    single_step_time = single_step_end_time - init_end_time
        
    # breakpoint
    for i in range(1000):
        process.createBreakpoint(breakpoint_addr)
        process.cont()
        process.waitSignals(signal.SIGTRAP)
        bp = process.findBreakpoint(breakpoint_addr)
        #print("0x{:X}".format(bp.address))
        bp.desinstall(True)
        
    breakpoint_end_time = (time.time() - start) * 1000
    breakpoint_time = breakpoint_end_time - single_step_end_time
    
    # read and write regs
    for i in range(1000):
        regs = process.getregs()
        regs.rax = regs.rbx
        regs.rip = regs.rsp
        process.setregs(regs)
        
    regs_end_time = (time.time() - start) * 1000
    regs_time = regs_end_time - breakpoint_end_time
        
    # read and write memory
    sp = process.getStackPointer()
    for i in range(1000):
        w = process.readWord(sp)
        w += 5
        process.writeWord(sp, w)
        
    mem_end_time = (time.time() - start) * 1000
    mem_time = mem_end_time - regs_end_time
        
    # search memory
    for i in range(100):
        done = False
        for m in process.readMappings():
            if done:
                break
            for a in range(m.start, m.end, 8):
                w = process.readWord(a)
                if w == 6464:
                    done = True
                    break
                
    search_end_time = (time.time() - start) * 1000
    search_time = search_end_time - mem_end_time
    total_time = search_end_time

    process.terminate()

    
    results_file.write("{:d},{:d},{:d},{:d},{:d},{:d},{:d}\n".format(
            int(init_time),
            int(single_step_time),
            int(breakpoint_time),
            int(regs_time),
            int(mem_time),
            int(search_time),
            int(total_time),
            ))

results_file.close()
