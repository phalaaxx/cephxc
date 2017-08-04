CephXC
==

CephXC is an API interface to LXC containers installed on Ceph RBD devices each. It can be used to automatically map RBD device, mount it in the proper location and start LXC container (all in one step) and takes care with the same in reverse order when stopping containers from a LXC host environment.

Requirements
--

There are several assumptions which for configuration options which are not configurable yet, which need to be met:

* LXC host environment should load rbd.ko module automatically
* Containers will run from /var/lib/lxc/<name> directory
* Each container is installed and properly configured in their own RBD device
* All host servers have similar configurations and containers can run in each without modifications
* All host servers can communicate with each other through HTTP requests on a random port (8000 by default)
* Ceph cluster is properly configured and host environments can map/unmap RBD devices from the storage
* CephXC is configured to automatically start after host environment boots up
* RBD names are the same as container names

API endpoints:
--

Currently CephXC API consists of 4 endpoints for controlling containers:

* /lxc - lists all containers currently visible to the host environment
  * parameters:
    * optional: name - specifies the name of the container for which to dump current status/information
* /lxc/start - maps rbd device, mounts it in /var/lib/lxc/<name> and starts lxc container
  * parameters:
    * optional: pool - name of the ceph rbd device pool, default is "rbd"
	* optional: fstype - filesystem type as required by mount syscall, default is "xfs"
	* required: name - the name of container and rbd device to map and start
* /lxc/stop - stops lxc container, umounts it from /var/lib/lxc/<name> and unmaps rbd device
  * parameters:
    * required: name - the name of the container to stop
* /lxc/moveto - stops lxc container from current host environment and start it on another
  * parameters:
    * required: name - the name of the container which is to be migrated to another server
	* required: next - name of the server on which to start the container
	* optional: port - numeric port number of CephXC running on next server, default is 8000

Usage
--

Let's assume host servers follow a naming convention, for example: host1.example.com, host2.example.com, hostN.example.com.
Also let's assume containers are named like lxc1, lxc2, lxcN and their rbd devices are in a pool called containers.
Examples below will be demonstrated with curl.

Start lxc1 container on server host1.example.com (assume container's rbd is formatted with ext3 filesystem):

	curl 'http://host1.example.com:8000/lxc/start?pool=containers&name=lxc1&fstype=ext3'

Stop lxc2 container running on host2.example.com:

	curl 'http://host2.example.com:8000/lxc/stop?name=lxc2'

Stop lxc3 container running on host1.example.com and immediately start it on host2.example.com:

	curl 'http://host1.example.com:8000/lxc/moveto?name=lxc3&next=host2.example.com


Limitations
--

* Currently all API endpoints are accessible with GET requests only, this will change in future
* It is necessary to know precisely on which host server a container is running in order to stop it or move it elswhere, currently there's no way to automatically find it
