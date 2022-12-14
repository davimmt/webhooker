package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-cmd/cmd"
)

func execPipe(commit_info string) {
	go func() {
		fmt.Println(commit_info)
		fmt.Println()

		// Disable output buffering, enable streaming
		cmdOptions := cmd.Options{
			Buffered:  false,
			Streaming: true,
		}

		// Create Cmd with options
		envCmd := cmd.NewCmdOptions(cmdOptions, "./pipe.sh", commit_info)

		// Print STDOUT and STDERR lines streaming from Cmd
		doneChan := make(chan struct{})
		go func() {
			defer close(doneChan)
			// Done when both channels have been closed
			// https://dave.cheney.net/2013/04/30/curious-channels
			for envCmd.Stdout != nil || envCmd.Stderr != nil {
				select {
				case line, open := <-envCmd.Stdout:
					if !open {
						envCmd.Stdout = nil
						continue
					}
					fmt.Println(line)
				case line, open := <-envCmd.Stderr:
					if !open {
						envCmd.Stderr = nil
						continue
					}
					fmt.Fprintln(os.Stderr, line)
				}
			}
		}()

		// Run and wait for Cmd to return, discard Status
		<-envCmd.Start()

		// Wait for goroutine to print everything
		<-doneChan
	}()
}

func main() {
	http.HandleFunc("/bitbucket/", bitbucket)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
