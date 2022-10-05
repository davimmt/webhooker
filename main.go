package main

import (
	"fmt"
	"net/http"
	"os/exec"
)

func execPipe(commit_info string) {
	go func() {
		fmt.Println(commit_info)
		fmt.Println()
		cmd, err := exec.Command("./pipe.sh", commit_info).CombinedOutput()
		fmt.Println(string(cmd))
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
}

func main() {
	http.HandleFunc("/bitbucket/", bitbucket)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
