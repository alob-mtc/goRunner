// This sample program demonstrates how to use a channel to
// monitor the amount of time the program is running and terminate
// the program if it runs too long.
package main

import (
	"log"
	"os"
	"time"

	"github.com/alob-mtc/runner"
)

// timeout is the number of second the program has to finish.
const timeout = 3 * time.Second

// main is the entry point for the program.
func main() {
	log.Println("Starting work.")

	// Create a new timer value for this run and the number of worker.
	r := runner.New(timeout, 3)

	// Add the tasks to be run.
	r.Add(createTask("A"), createTask("B"), createTask("C"), createTask("D"), createTask("E"), createTask("F"))

	// Run the tasks and handle the result.
	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeout:
			log.Println("Terminating due to timeout.")
			os.Exit(1)
		case runner.ErrInterrupt:
			log.Println("Terminating due to interrupt.")
			os.Exit(2)
		}
	}

	log.Println("Process ended.")
}

// createTask returns an example task that sleeps for the specified
// number of seconds based on the id.
func createTask(name string) func(int) {
	return func(id int) {
		duration := 1
		log.Printf("Processor - Task #%s....worker #%d\n", name, id)
		time.Sleep(time.Duration(time.Duration(duration)) * time.Second)
		log.Printf("Processor - Task #%s., completed Time - #%d ", name, duration)
	}
}
