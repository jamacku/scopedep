package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

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
	cmd := exec.Command("sleep", "infinity")
	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start worker: %v\n", err)
	}

	log.Printf("worker running as PID: %d\n", cmd.Process.Pid)

	// Wrap a scope unit around child process and make sure add Before dependecy on the main unit
	// We are adding Before dependecy because want deterministic shutdown order
	resultChan := make(chan string)
	scopeProperties := []dbus.Property{dbus.PropBefore(UNIT), dbus.PropSlice("system.slice"), dbus.PropPids(uint32(cmd.Process.Pid))}

	_, err = sd.StartTransientUnit(SCOPE, "replace", scopeProperties, resultChan)
	if err != nil {
		log.Fatalf("failed to create worker scope unit: %v\n", err)
	}

	r := <-resultChan
	if r != "done" {
		log.Fatalf("StartTransientUnit() failed with job result: %v\n", r)
	}

	log.Printf("%s started\n", SCOPE)

	// Setup signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Ignore()
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for {
		sig := <-signalChan

		if sig == syscall.SIGTERM || sig == syscall.SIGINT || sig == syscall.SIGHUP {
			// We want to reuse the channel so it must be empty before we do that
			if len(resultChan) != 0 {
				panic("unexpected message in the channel")
			}

			log.Printf("about to stop %s\n", SCOPE)

			// Kill the worker process and wait for it
			// Note that calling StopUnit() DBus API with "replace" job mode would create deadlock
			err = cmd.Process.Kill()
			if err != nil {
				log.Fatalf("failed to kill worker process: %v\n", err)
			}

			cmd.Wait()

			log.Printf("worker killed\n")

			break
		}
	}

	log.Printf("%s exiting\n", UNIT)
}
