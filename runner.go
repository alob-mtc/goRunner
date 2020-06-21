// Package runner manages the running and lifetime of a process.
package runner

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Runner runs a set of tasks within a given timeout and can be
// shut down on an operating system interrupt.
type Runner struct {
	// interrupt channel reports a signal from the
	// operating system.
	interrupt chan os.Signal

	// complete channel reports that processing is done.
	complete chan error

	// complete channel reports that processing is done.
	completeMain chan error

	// timeout reports that time has run out.
	timeout <-chan time.Time

	// tasks holds a set of functions that are executed
	// synchronously in index order.
	tasks []func(int)

	//mutex
	m sync.Mutex

	// terminate controlles the termination of workers
	terminate bool
}

// ErrTimeout is returned when a value is received on the timeout channel.
var ErrTimeout = errors.New("received timeout")

// ErrInterrupt is returned when an event from the OS is received.
var ErrInterrupt = errors.New("received interrupt")

// New returns a new ready-to-use Runner.
func New(d time.Duration) *Runner {
	return &Runner{
		interrupt:    make(chan os.Signal, 1),
		complete:     make(chan error),
		timeout:      time.After(d),
		completeMain: make(chan error),
	}
}

// Add attaches tasks to the Runner. A task is a function that
// takes an int ID.
func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

// Start runs all tasks and monitors channel events.
func (r *Runner) Start() error {
	// We want to receive all interrupt based signals.
	signal.Notify(r.interrupt, os.Interrupt)

	// Run the different tasks on a different goroutine.
	r.run()
	// spin up the master GOR
	go func() {
		// check if all the task as been precessed
		completedTask := 0
		for range r.complete {
			completedTask++
			if completedTask == len(r.tasks) {
				close(r.complete)
				r.completeMain <- nil
				return
			}
		}
	}()
	select {
	// Signaled when processing is done.
	case err := <-r.completeMain:
		// id the err is ErrInterrupt => tell all the running workers to terminate
		return err

	// Signaled when we run out of time.
	case <-r.timeout:
		return ErrTimeout
	}
}

// run executes each registered task.
func (r *Runner) run() error {
	for id := 0; id < 3; id++ {
		// spin up the worker GORs to Execute the registered task.
		go func(i int) {
			//get the task
			task, ok := r.getTask()
			for ok {
				// Check for an interrupt signal from the OS.
				if r.gotInterrupt() {
					r.completeMain <- ErrInterrupt
					return
				}
				// run the task
				task(i)
				task, ok = r.getTask()

			}
			r.complete <- nil
		}(id)
	}

	return nil
}

// gotInterrupt verifies if the interrupt signal has been issued.
func (r *Runner) gotInterrupt() bool {
	select {
	// Signaled when an interrupt event is sent.
	case <-r.interrupt:
		r.terminate = true
		// Stop receiving any further signals.
		signal.Stop(r.interrupt)
		return true

		// Continue running as normal.
	default:
		// check if ternimate
		if r.terminate {
			return true
		}
		return false
	}
}

// getTask
func (r *Runner) getTask() (task func(int), found bool) {
	// secure this operation with lock
	r.m.Lock()
	defer r.m.Unlock()
	//TODO: fetch the task form the task queue
	for i, value := range r.tasks {
		if value == nil {
			continue
		} else {
			task = value
			found = true
			// set the index to nil
			r.tasks[i] = nil
			break
		}
	}
	return
}
