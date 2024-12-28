package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

func main() {
	cmdChan := make(chan string)
	outputChan := make(chan string)

	go func() {
		cmd := exec.Command("alien.py")

		stdin, err := cmd.StdinPipe()
		if err != nil {
			fmt.Printf("Error setting up stdin: %v\n", err)
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Printf("Error setting up stdout: %v\n", err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Printf("Error setting up stderr: %v\n", err)
			return
		}

		if err := cmd.Start(); err != nil {
			fmt.Printf("Error starting bash: %v\n", err)
			return
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				outputChan <- scanner.Text()
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				outputChan <- scanner.Text()
			}
		}()

		for cmdStr := range cmdChan {
			_, err := fmt.Fprintln(stdin, cmdStr)
			if err != nil {
				fmt.Printf("Error writing to stdin: %v\n", err)
				return
			}
		}

		stdin.Close()

		if err := cmd.Wait(); err != nil {
			fmt.Printf("Error waiting for bash to exit: %v\n", err)
		}
	}()

	for output := range outputChan {
		fmt.Println(output)
		matched, err := regexp.Match("AliEn[.*]:/alice/cern/user.*", []byte(output))
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			fmt.Println("MATCHED")
		}
	}

	cmdChan <- "ls -la /"

	for output := range outputChan {
		fmt.Println(output)
	}
}
