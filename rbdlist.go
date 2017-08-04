package main

import (
	"os"
	"strconv"
)

/* RbdList represents list of ceph rbd devices */
type RbdList []int

/* Exists returns true if n is present in RbdList, false otherwise */
func (r RbdList) Exists(n int) bool {
	for _, i := range r {
		if i == n {
			return true
		}
	}
	return false
}

/* GetNew finds and returns first outersection item in RbdList if it exists, -1 otherwise */
func (r RbdList) GetNew(n RbdList) int {
	for _, i := range n {
		if !r.Exists(i) {
			return i
		}
	}
	return -1
}

/* NewRbdList returns list of currently mapped rbd devices */
func NewRbdList(path string) (result RbdList) {
	// open devices list for reading
	d, err := os.Open(path)
	if err != nil {
		return
	}
	// read list of devices
	names, err := d.Readdirnames(-1)
	if err != nil {
		return
	}
	// parse names
	for _, idStr := range names {
		// parse integer
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}
		result = append(result, id)
	}
	return
}
