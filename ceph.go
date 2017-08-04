package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

/* CephConfig is a ceph configuration data structure */
type CephConfig struct {
	FsId        string
	MonHosts    []string
	MonMembers  []string
	AuthCluster string
	AuthService string
	AuthClient  string
}

/* GetCephConfig reads configuration from named path */
func GetCephConfig(path string) (*CephConfig, error) {
	// open configuration file for reading
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// parse read configuration
	config := new(CephConfig)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// get next line
		line := strings.TrimSpace(scanner.Text())
		if strings.Index(line, "=") == -1 {
			continue
		}
		// split to items
		split := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])
		// parse line
		switch strings.ToLower(key) {
		case "fsid":
			config.FsId = strings.ToLower(value)
		case "mon_initial_members":
			config.MonMembers = strings.Split(value, ",")
		case "mon_host":
			config.MonHosts = strings.Split(value, ",")
			for idx := range config.MonHosts {
				if strings.Index(config.MonHosts[idx], ":") == -1 {
					config.MonHosts[idx] = fmt.Sprintf("%s:6789", config.MonHosts[idx])
				}
			}
		case "auth_cluster_required":
			config.AuthCluster = value
		case "auth_service_required":
			config.AuthService = value
		case "auth_client_required":
			config.AuthClient = value
		}
	}
	// check for scan errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}

/* CephKeyring contains contains client keyring information for ceph cluster */
type CephKeyring struct {
	Name   string
	Secret string
}

/* ReadSecret reads ceph configuration client secret */
func ReadCephKeyring(path string) (*CephKeyring, error) {
	// open client keyring configuration file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// read keyring file
	keyring := new(CephKeyring)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// parse line
		line := strings.TrimSpace(scanner.Text())
		// check for keyring name
		if strings.HasPrefix(strings.ToLower(line), "[client.") {
			keyring.Name = strings.TrimSpace(line[8 : len(line)-1])
			continue
		}
		// look for key=value pairs
		if strings.Index(line, "=") == -1 {
			continue
		}
		// split to items
		split := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])
		// look for secret
		if strings.ToLower(key) == "key" {
			keyring.Secret = value
		}
	}
	// check scanner for error
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	// return result
	return keyring, nil
}
