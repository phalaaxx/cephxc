package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

/* HandleLxcList renders list of containers in the system */
func HandleLxcList(w http.ResponseWriter, r *http.Request) {
	// response data structure
	type ResponseData struct {
		Server     string          `json:"server"`
		Containers []ContainerInfo `json:"containers"`
	}
	data := ResponseData{
		Server: GetHostname(),
	}
	// get query container name
	name := r.URL.Query().Get("name")
	// get list of current mounts
	mounts := GetSystemMounts()
	// loop over containers
	for _, c := range lxc.DefinedContainers(lxcpath) {
		// lookup container if name is provided
		if len(name) != 0 && name != c.Name() {
			continue
		}
		// add container data
		container, err := GetContainerInfo(&c, mounts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.Containers = append(
			data.Containers,
			*container,
		)
	}
	// render to json
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/* HandleLxcStart maps ceph rbd device, mounts it and starts lxc container */
func HandleLxcStart(w http.ResponseWriter, r *http.Request) {
	// get query parameters
	query := r.URL.Query()
	name := query.Get("name")
	pool := query.Get("pool")
	fstype := query.Get("fstype")
	// default values
	if len(pool) == 0 {
		pool = "rbd"
	}
	if len(fstype) == 0 {
		fstype = "xfs"
	}
	// map ceph rbd device
	rbd, err := RbdMap(pool, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// get mount options
	source := fmt.Sprintf("/dev/rbd%d", rbd)
	target := fmt.Sprintf("/var/lib/lxc/%s", name)
	// make sure mount target exists
	if _, err := os.Stat(target); err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// create mount target directory
		if err = os.MkdirAll(target, 0755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// mount container root loop
MOUNT:
	for {
		switch err := syscall.Mount(source, target, fstype, syscall.MS_NOATIME, ""); err {
		case syscall.ENODEV:
			// wait for 10 milliseconds and try again
			time.Sleep(time.Millisecond * 10)
		case nil:
			break MOUNT
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// start lxc container
	for _, c := range lxc.DefinedContainers(lxcpath) {
		// locate named container
		if name != c.Name() {
			continue
		}
		// start container
		if c.State() == lxc.STOPPED {
			go c.Start()
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}
	// container was not found in newly mounted rbd
	w.WriteHeader(http.StatusNotFound)
}

// HandleLxcStop shuts a container down, unmounts its root directory and unmaps ceph device
func HandleLxcStop(w http.ResponseWriter, r *http.Request) {
	// get container name
	name := r.URL.Query().Get("name")
	if len(name) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// get system mounts
	mounts := GetSystemMounts()
	// lookup proper container
	for _, container := range lxc.DefinedContainers(lxcpath) {
		// locate named container object
		if name != container.Name() {
			continue
		}
		// parse container information
		c, err := GetContainerInfo(&container, mounts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// handle shutdown action
		switch c.State() {
		case lxc.RUNNING:
			go c.DoShutdown()
			w.WriteHeader(http.StatusAccepted)
		case lxc.STOPPED:
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}

/* HandleLxcMoveTo stops container from the current server and starts it on a remote server */
func HandleLxcMoveTo(w http.ResponseWriter, r *http.Request) {
	// get query parameters
	query := r.URL.Query()
	name := query.Get("name")
	next := query.Get("next")
	port := query.Get("port")
	// default values
	if len(port) == 0 {
		port = "8000"
	}
	// make sure next server is not the current one
	if next == GetHostname() {
		w.WriteHeader(http.StatusOK)
		return
	}
	// get list of current mounts
	mounts := GetSystemMounts()
	// lookup named container
	for _, container := range lxc.DefinedContainers(lxcpath) {
		// filter out unspecified containers
		if name != container.Name() {
			continue
		}
		// get ContainerInfo object
		c, err := GetContainerInfo(&container, mounts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// make sure container is running
		if c.State() == lxc.RUNNING {
			// get remove server port as integer
			portInt, err := strconv.Atoi(port)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// initiate move operation
			go c.MoveTo(next, portInt)
			w.WriteHeader(http.StatusAccepted)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}
