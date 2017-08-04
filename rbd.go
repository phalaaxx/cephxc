package main

import (
	"fmt"
	"os"
	"strings"
)

/* RbdMap instructs kernel rbd module to map a ceph device */
func RbdMap(pool, name string) (int, error) {
	devpath := "/sys/bus/rbd/devices"
	// prepare kernel rbd command
	rbdcmd := fmt.Sprintf("%s name=%s,secret=%s %s %s -",
		strings.Join(config.MonHosts, ","),
		keyring.Name,
		keyring.Secret,
		pool,
		name,
	)
	// get devices list
	mapped := NewRbdList(devpath)
	// map ceph rbd device
	f, err := os.OpenFile("/sys/bus/rbd/add_single_major", os.O_WRONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			// try to open add file
			f, err = os.OpenFile("/sys/bus/rbd/add", os.O_WRONLY, 0644)
		}
		if err != nil {
			return -1, err
		}
	}
	defer f.Close()
	// send command to kernel module
	if _, err := f.WriteString(rbdcmd); err != nil {
		return -1, err
	}
	return mapped.GetNew(NewRbdList(devpath)), nil
}

/* RbdUnmap removes a kernel rbd device map */
func RbdUnmap(rbd string) error {
	if !strings.HasPrefix(rbd, "/dev/rbd") {
		return fmt.Errorf("Bad device name")
	}
	// attempt to openkernel module interface
	f, err := os.OpenFile("/sys/bus/rbd/remove_single_major", os.O_WRONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			// try to open remove
			f, err = os.OpenFile("/sys/bus/rbd/remove", os.O_WRONLY, 0644)
		}
		if err != nil {
			return err
		}
	}
	defer f.Close()
	// instruct kernel module to unmap device
	if _, err := f.WriteString(rbd[8:len(rbd)]); err != nil {
		return err
	}
	return nil
}