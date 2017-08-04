package main

import (
	"flag"
	"gopkg.in/lxc/go-lxc.v2"
	"log"
	"net/http"
)

/* global variables */
var (
	lxcpath string
	listen  string
	config  *CephConfig
	keyring *CephKeyring
)

/* initialize configuration */
func init() {
	flag.StringVar(&lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&listen, "listen", ":8000", "Bind to the specified address")
	flag.Parse()
	// parse ceph configuration
	var err error
	config, err = GetCephConfig("/etc/ceph/ceph.conf")
	if err != nil {
		log.Fatal(err)
	}
	// parse ceph keyring
	keyring, err = ReadCephKeyring("/etc/ceph/ceph.client.admin.keyring")
	if err != nil {
		log.Fatal(err)
	}
}

/* main program */
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/lxc/start", HandleLxcStart)
	mux.HandleFunc("/lxc/stop", HandleLxcStop)
	mux.HandleFunc("/lxc/moveto", HandleLxcMoveTo)
	mux.HandleFunc("/lxc", HandleLxcList)
	http.ListenAndServe(listen, mux)
}
