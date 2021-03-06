# Internals

This page documents internal code details and design decisions.

The resulting Docker image contains the following:

* Base linux distribution - this provides standard Linux libraries (such as "glibc") and utilities (such as "ls" and "grep") required by MQ
* MQ installation (under `/opt/mqm`)
* Three additional programs, to enable running in a containerized environment:
   - `runmqserver` - The main process, which creates and runs a queue manager
   - `runmqdevserver` - The main process for MQ Advanced for Developers
   - `chkmqhealthy` - Checks the health of the queue manager.  This can be used by (say) a Kubernetes liveness probe.
   - `chkmqready` - Checks if the queue manager is ready for work.  This can be used by (say) a Kubernetes readiness probe.

## runmqserver
The `runmqserver` command has the following responsibilities:

* Checks license acceptance
* Sets up `/var/mqm`
    - MQ data directory needs to be set up at container creation time.  This is done using the `crtmqdir` utility, which was introduced in MQ V9.0.3
    - It assumes that a storage volume for data is mounted under `/mnt/mqm`.  It creates a sub-directory for the MQ data, so `/var/mqm` is a symlink which resolves to `/mnt/mqm/data`.  The reason for this is that it's not always possible to change the ownership of an NFS mount point directly (`/var/mqm` needs to be owned by "mqm"), but you can change the ownership of a sub-directory.
* Acts like a daemon
    - Handles UNIX signals, like SIGTERM
    - Works as PID 1, so is responsible for [reaping zombie processes](https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/)
* Creating and starting a queue manager
* Configuring the queue manager, by running any MQSC scripts found under `/etc/mqm`
* Indicates to the `chkmqready` command that configuration is complete, and that normal readiness checking can happen.  This is done by writing a file into `/run/runmqserver`

In addition, for MQ Advanced for Developers only, the web server is started.

## runmqdevserver
The `runmqdevserver` command is added to the MQ Advanced for Developers image only.  It does the following, before invoking `runmqserver`:

1. Sets passwords based on supplied environment variables
2. Generates MQSC files to put in `/etc/mqm`, based on a template, which is updated with values based on supplied environment variables.
3. If requested, it creates TLS key stores under `/run/runmqdevserver`, and configures MQ and the web server to use them

A special version of `runmqserver` is used in the developer image, which performs extra actions like starting the web server.  This is built using the `mqdev` [build constraint](https://golang.org/pkg/go/build/#hdr-Build_Constraints).