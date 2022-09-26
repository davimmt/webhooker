package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
)

type Push struct {
	Push struct {
		Changes []struct {
			Old struct {
				Target struct {
					Hash string `json:"hash"`
				} `json:"target"`
			} `json:"old"`
			New struct {
				Name   string `json:"name"`
				Target struct {
					Hash string `json:"hash"`
				} `json:"target"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
}

type PullRequest struct {
	Repository struct {
		MergeCommit struct {
			Hash string `json:"hash"`
		} `json:"merge_commit"`
		Destination struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
			Commit struct {
				Hash string `json:"hash"`
			} `json:"commit"`
		} `json:"destination"`
	} `json:"pullrequest"`
}

func bitbucketPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var payload Push
	err = json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if len(payload.Push.Changes) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Extract commit and send to pipeline
	old_commit_hash := payload.Push.Changes[0].Old.Target.Hash
	new_commit_hash := payload.Push.Changes[0].New.Target.Hash
	branch := payload.Push.Changes[0].New.Name

	if new_commit_hash == "" || old_commit_hash == "" || branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	execPipe(branch, new_commit_hash, old_commit_hash)
}

func bitbucketPullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var payload PullRequest
	err = json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Extract commit and send to pipeline
	old_commit_hash := payload.Repository.Destination.Commit.Hash
	new_commit_hash := payload.Repository.MergeCommit.Hash
	branch := payload.Repository.Destination.Branch.Name

	if new_commit_hash == "" || old_commit_hash == "" || branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	execPipe(branch, new_commit_hash, old_commit_hash)
}

func execPipe(branch, new_commit_hash, old_commit_hash string) {
	go func() {
		out, err := exec.Command("./pipe.sh", branch, new_commit_hash, old_commit_hash).CombinedOutput()
		fmt.Println(string(out))
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
}

func main() {
	http.HandleFunc("/bitbucket-webhook/push", bitbucketPush)
	http.HandleFunc("/bitbucket-webhook/pull-request", bitbucketPullRequest)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
