package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type MemoryMapping struct {
	Start       uintptr
	End         uintptr
	Permissions string
	Offset      uint64
	MajorDevice uint64
	MinorDevice uint64
	Inode       uint64
	Pathname    string
}

func ReadProcessMappings(pid int) []MemoryMapping {
	maps := make([]MemoryMapping, 0)
	path := fmt.Sprintf("/proc/%d/maps", pid)
	mapsFile, err := os.Open(path)
	catch(err)
	defer mapsFile.Close()
	re := getProcMapsRegexp()
	scanner := bufio.NewScanner(mapsFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := re.FindStringSubmatch(line)
		mapping := parseMemoryMapping(parts)
		maps = append(maps, mapping)
	}
	return maps
}

func DumpMaps(maps []MemoryMapping) {
	for _, m := range maps {
		DumpMap(&m)
	}
}

func DumpMap(m *MemoryMapping) {
	fmt.Printf("0x%016X-0x%016X\n", m.Start, m.End)
	fmt.Printf("\tPermissions: %s\n", m.Permissions)
	fmt.Printf("\tOffset: 0x%X ", m.Offset)
	fmt.Printf("Device: %02X:%02X ", m.MajorDevice, m.MinorDevice)
	fmt.Printf("Inode: %d\n", m.Inode)
	fmt.Printf("\tPath: %s", m.Pathname)
	fmt.Printf("\n")
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}

func getProcMapsRegexp() *regexp.Regexp {
	re, _ := regexp.Compile(`([0-9a-f]+)-([0-9a-f]+) (.{4}) ([0-9a-f]+) ([0-9a-f]{2,3}):([0-9a-f]{2}) ([0-9]+)(?: +(.*))?`)
	return re
}

func parseMemoryMapping(parts []string) MemoryMapping {
	start, _ := strconv.ParseUint(parts[1], 16, 64)
	end, _ := strconv.ParseUint(parts[2], 16, 64)
	permissions := parts[3]
	offset, _ := strconv.ParseUint(parts[4], 16, 64)
	majorDevice, _ := strconv.ParseUint(parts[5], 16, 64)
	minorDevice, _ := strconv.ParseUint(parts[6], 16, 64)
	inode, _ := strconv.ParseUint(parts[7], 10, 64)
	pathName := parts[8]
	mapping := MemoryMapping{
		Start:       uintptr(start),
		End:         uintptr(end),
		Permissions: permissions,
		Offset:      offset,
		MajorDevice: majorDevice,
		MinorDevice: minorDevice,
		Inode:       inode,
		Pathname:    pathName,
	}
	return mapping
}
