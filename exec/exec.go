package exec

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/oktalz/present/types"
)

func Cmd(tc types.TerminalCommand) []byte {
	cmd := exec.Command(tc.App, tc.Cmd...) //nolint:gosec
	cmd.Dir = tc.Dir
	output, err := cmd.Output()
	if err != nil {
		return []byte(err.Error())
	}
	return output
}

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func CmdStream(tc types.TerminalCommand) {
	fmt.Println("======== executing", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, tc.App, tc.Cmd...) //nolint:gosec
	if DirectoryExists(tc.Dir) {
		cmd.Dir = tc.Dir
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return
		}
		cmd.Dir = path.Join(dir, tc.Dir)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	if err := cmd.Start(); err != nil {
		return
	}

	go func() {
		scannerOut := bufio.NewScanner(stdout)
		for scannerOut.Scan() {
			fmt.Println(scannerOut.Text())
		}
	}()
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			fmt.Println(scannerErr.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Println(err.Error())
		fmt.Println("======== finished ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
		return
	}
	fmt.Println("======== finished ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
}

func CmdStreamWS(tc types.TerminalCommand, ch chan string, timeout time.Duration, raw bool) {
	go cmdStreamWS(tc, ch, timeout, raw)
}

func cmdStreamWS(tc types.TerminalCommand, ch chan string, timeout time.Duration, raw bool) { //nolint:funlen
	fmt.Println("======== executing", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		cancel()
		close(ch)
		fmt.Println("======== finished ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
	}()
	cmd := exec.CommandContext(ctx, tc.App, tc.Cmd...) //nolint:gosec
	if DirectoryExists(tc.Dir) {
		cmd.Dir = tc.Dir
	} else {
		dir, err := os.Getwd()
		if err != nil {
			ch <- err.Error()
			return
		}
		cmd.Dir = path.Join(dir, tc.Dir)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ch <- err.Error()
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		ch <- err.Error()
		return
	}

	if err := cmd.Start(); err != nil {
		ch <- err.Error()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		scannerOut := bufio.NewScanner(stdout)
		// scannerOut.Split(bufio.ScanLines)
		for scannerOut.Scan() {
			txt := scannerOut.Text()
			if raw {
				ch <- txt
			} else {
				ch <- strings.ReplaceAll(txt, " ", "&nbsp;")
			}
			// fmt.Println(scannerOut.Text())
			log.Println(txt)
		}
		wg.Done()
	}()
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			ch <- scannerErr.Text()
			log.Println(scannerErr.Text())
		}
		wg.Done()
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		ch <- err.Error()
		fmt.Println(err.Error())
		fmt.Println("======== error ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
		return
	}
}
