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

func ReadProcessMappings(pid int) ([]MemoryMapping, error) {
	maps := make([]MemoryMapping, 0)
	path := fmt.Sprintf("/proc/%d/maps", pid)
	mapsFile, err := os.Open(path)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to open memory maps file: %s", err.Error())
		return maps, MakeGenericError(errMsg)
	}
	defer mapsFile.Close()
	re := getProcMapsRegexp()
	scanner := bufio.NewScanner(mapsFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := re.FindStringSubmatch(line)
		mapping := parseMemoryMapping(parts)
		maps = append(maps, mapping)
	}
	return maps, nil
}

func (m MemoryMapping) String() string {
	return fmt.Sprintf("0x%016X-0x%016X %s %s", m.Start, m.End, m.Permissions, m.Pathname)
	//result := fmt.Sprintf("0x%016X-0x%016X\n", m.Start, m.End)
	//result += fmt.Sprintf("\tPermissions: %s\n", m.Permissions)
	//result += fmt.Sprintf("\tOffset: 0x%X ", m.Offset)
	//result += fmt.Sprintf("Device: %02X:%02X ", m.MajorDevice, m.MinorDevice)
	//result += fmt.Sprintf("Inode: %d\n", m.Inode)
	//result += fmt.Sprintf("\tPath: %s", m.Pathname)
	//result += fmt.Sprintf("\n")
	//return result
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
