package main

import (
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/* RbdDeviceId returns the id of a rbd block device */
func RbdDeviceId(rbd string) int {
	// parse rbd device id
	if strings.HasPrefix(rbd, "/dev/rbd") {
		if id, err := strconv.Atoi(rbd[8:len(rbd)]); err == nil {
			return id
		}
	}
	return -1
}

// container info data structure
type ContainerInfo struct {
	*lxc.Container
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	IPv4Address *string `json:"ipv4address"`
	IPv6Address *string `json:"ipv6address"`
	Mount       string  `json:"mount"`
	BlockDevice string  `json:"rbd"`
	Pool        string  `json:"pool"`
}

/* RootDir returns container root directory name */
func (c ContainerInfo) RootDir() string {
	return fmt.Sprintf("/var/lib/lxc/%s", c.Name)
}

/* FullShutdown stops running container and umounts root filesystem and block device */
func (c *ContainerInfo) DoShutdown() error {
	if c.State() == lxc.RUNNING {
		// shutdown container
		if err := c.Shutdown(time.Second * 60); err != nil {
			return err
		}
		// umount filesystem loop
	UMOUNT:
		for {
			// attempt to umount filesystem
			switch err := syscall.Unmount(c.RootDir(), 0); err {
			case syscall.EBUSY:
				// sleep for 10 milliseconds
				time.Sleep(time.Millisecond * 10)
			case nil:
				// break out of the umount loop
				break UMOUNT
			default:
				return err
			}
		}
		// unmap rbd device
		if err := RbdUnmap(c.BlockDevice); err != nil {
			return err
		}
	}
	return nil
}

/* MoveTo stops container from current server and starts it on the specified one */
func (c *ContainerInfo) MoveTo(server string, port int) error {
	// stop the container from current server
	if err := c.DoShutdown(); err != nil {
		fmt.Println("DoShutdown:", err)
		return err
	}
	// run command on remote server to start container there
	nexturl := fmt.Sprintf("http://%s:%d/lxc/start?pool=%s&name=%s",
		server,
		port,
		c.Pool,
		c.Name,
	)
	client := http.Client{Timeout: time.Duration(time.Second * 60)}
	resp, err := client.Get(nexturl)
	if err != nil {
		fmt.Println("client.Get:", err)
		return err
	}
	// check if remote server accepted new connection
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Unable to start container on remote server")
	}
	return nil
}

/* GetContainerInfo retrieves information about a container from the system */
func GetContainerInfo(c *lxc.Container, mounts map[string]string) (*ContainerInfo, error) {
	// get ipv4 address
	// TODO: ipv4addr and ipv6addr - parse list of strings
	var ipv4addr *string
	var ipv6addr *string
	// get container addresses if container is running
	if c.State() == lxc.RUNNING {
		// parse list of ipv4 addresses
		if ipv4list, err := c.IPv4Addresses(); err == nil {
			if len(ipv4list) != 0 {
				ipv4addr = &ipv4list[0]
			}
		}
		// parse list of ipv6 addresses
		if ipv6list, err := c.IPv6Addresses(); err == nil {
			if len(ipv6list) != 0 {
				ipv6addr = &ipv6list[0]
			}
		}
	}
	// get mount point
	mount := fmt.Sprintf("/var/lib/lxc/%s", c.Name())
	var device string
	if m, ok := mounts[mount]; ok {
		device = m
	}
	// get pool name
	poolByte, err := ioutil.ReadFile(fmt.Sprintf("/sys/devices/rbd/%d/pool", RbdDeviceId(device)))
	if err != nil {
		return nil, err
	}
	pool := strings.TrimSpace(string(poolByte))
	// prepare result structure
	container := ContainerInfo{c, c.Name(), c.State().String(), ipv4addr, ipv6addr, mount, device, pool}
	// return result
	return &container, nil
}
