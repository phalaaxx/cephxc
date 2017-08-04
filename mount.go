package main

import (
	"bufio"
	"os"
	"strings"
)

/* GetSystemMounts returns a map of system mount points to block devices */
func GetSystemMounts() map[string]string {
	// open /proc/mounts for reading
	fd, err := os.Open("/proc/mounts")
	if err != nil {
		return nil
	}
	defer fd.Close()
	// parse mounts file and return result
	scanner := bufio.NewScanner(fd)
	result := make(map[string]string)
	for scanner.Scan() {
		items := strings.Split(scanner.Text(), " ")
		result[items[1]] = items[0]
	}
	return result
}
