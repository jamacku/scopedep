scopedep
========

Example service that demonstrates how to wrap a workload
(e.g. container payload) in a scope unit as well as set up
dependencies between main service and workload scope.

Application runs dummy workload (sleep) in a child process. Child is
then wrapped in systemd's scope unit. Newly created scope unit has
*Before* dependency on the main service. Hence on shutdown the main
service receives stop signal (SIGTERM) before the worker process is
killed. After signal is received, service will stop the worker process
and exit.

Building
--------
```shell
make build
```

Installation
------------
```shell
make install
```

Removal
-------
```shell
make uninstall
```

Dependencies
------------
* https://github.com/coreos/go-systemd
