package configuration

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
)

type Update struct {
	Latest bool `ff:"long: latest, usage: 'update to latest version if go is installed'"`
}

func (u *Update) UpdateToLatest() bool {
	cmd := exec.CommandContext(context.Background(), "go", "install", "github.com/oktalz/present@latest")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("error getting stdout pipe: %v\n", err)
		return false
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("error getting stderr pipe: %v\n", err)
		return false
	}

	if err := cmd.Start(); err != nil {
		log.Printf("error starting command: %v\n", err)
		return false
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		log.Printf("error waiting for command to finish: %v\n", err)
		return false
	}

	return true
}
