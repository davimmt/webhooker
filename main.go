package main

import (
	"fmt"
	"net/http"
	"os/exec"
)

func execPipe(event_type string, commit_info map[string]string) {
	go func() {
		out, err := exec.Command(
			"./pipe.sh",
			event_type,
			commit_info["destination_branch"],
			commit_info["new_commit_hash"],
			commit_info["old_commit_hash"],
			commit_info["source_branch"],
		).CombinedOutput()
		fmt.Println(string(out))
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
