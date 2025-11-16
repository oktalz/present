package execution

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	executedCommands = map[string]string{}
	execMutex        = &sync.Mutex{}
)

// RunOnce runs the given command only once in the given directory.
// If the command has already been executed, it will simply return nil.
// The command will be executed in the given directory, and its output
// will be streamed to the terminal.
func RunOnce(path, command string) error {
	execMutex.Lock()
	defer execMutex.Unlock()

	// Check if the command has already been executed
	if _, ok := executedCommands[path+command]; ok {
		log.Println("command has already been executed: ", path, command)
		return nil
	}

	split := strings.Split(command, " ")

	cmd := exec.Command(split[0], split[1:]...)

	// Set the working directory
	cmd.Dir = path

	// Stream output to terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Mark the command as executed
	executedCommands[path+command] = path
	log.Println("executing", path, command)

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}
