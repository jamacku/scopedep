package main

import (
	"log"
	"os/exec"

	"github.com/coreos/go-systemd/v22/dbus"
)

const (
	UNIT  string = "scopedep.service"
	SCOPE string = "scopedep-worker.scope"
)

func main() {
	// Create dbus connection to systemd
	sd, err := dbus.New()
	if err != nil {
		log.Fatalf("failed to setup DBus connection: %v\n", err)
	}
	defer sd.Close()

	// Spawn "worker" process
	cmd := exec.Command("sleep", "1")
	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start worker: %v\n", err)
	}

	log.Printf("worker running as PID: %d\n", cmd.Process.Pid)

	// Wrap a scope unit around child process and make sure add Before dependecy on the main unit
	// We are adding Before dependecy because want deterministic shutdown order
	resultChan := make(chan string)
	scopeProperties := []dbus.Property{dbus.PropSlice("system.slice"), dbus.PropPids(uint32(cmd.Process.Pid))}

	go func() {
		cmd.Wait();
	}()

	_, err = sd.StartTransientUnit(SCOPE, "replace", scopeProperties, resultChan)
	if err != nil {
		log.Fatalf("failed to create worker scope unit: %v\n", err)
	}

	r := <-resultChan
	if r != "done" {
		log.Fatalf("StartTransientUnit() failed with job result: %v\n", r)
	}

	log.Printf("%s started\n", SCOPE)
}
