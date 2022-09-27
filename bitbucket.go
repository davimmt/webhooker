package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type BitbucketRepoPush struct {
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

type BitbucketPullRequestFulfilled struct {
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
	} `json:"repository"`
}

func bitbucket(w http.ResponseWriter, r *http.Request) {
	// Get required headers
	required_headers := map[string]string{
		"user_agent": r.Header.Get("User-Agent"),
		"event_type": r.Header.Get("X-Event-Key"),
	}

	// Error if headers missing
	for k := range required_headers {
		if required_headers[k] == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Required headers missing"))
			return
		}
	}

	// Check if User Agent is Bitbucket, just in case
	if required_headers["user_agent"] != "Bitbucket-Webhooks/2.0" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Check event type from header
	var commit_info map[string]string
	switch required_headers["event_type"] {
	case "repo:push":
		commit_info = bitbucketRepoPush(w, b)
	case "pullrequest:fulfilled":
		commit_info = bitbucketPullRequestFulfilled(w, b)
	case "pullrequest:created":
		w.WriteHeader(500)
		w.Write([]byte("Not yet implemented"))
		return
	case "pullrequest:updated":
		w.WriteHeader(500)
		w.Write([]byte("Not yet implemented"))
		return
	default:
		w.WriteHeader(500)
		w.Write([]byte("Not yet implemented"))
		return
	}

	// Error if commit info is empty
	for k := range commit_info {
		if commit_info[k] == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong commit info"))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	execPipe(required_headers["event_type"], commit_info)
}

func bitbucketRepoPush(w http.ResponseWriter, b []byte) map[string]string {
	var payload BitbucketRepoPush
	err := json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	if len(payload.Push.Changes) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("500 - Something bad happened!"))
	}

	return map[string]string{
		"branch":          payload.Push.Changes[0].New.Name,
		"new_commit_hash": payload.Push.Changes[0].New.Target.Hash,
		"old_commit_hash": payload.Push.Changes[0].Old.Target.Hash,
	}
}

func bitbucketPullRequestFulfilled(w http.ResponseWriter, b []byte) map[string]string {
	var payload BitbucketPullRequestFulfilled
	err := json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	return map[string]string{
		"branch":          payload.Repository.Destination.Branch.Name,
		"new_commit_hash": payload.Repository.MergeCommit.Hash,
		"old_commit_hash": payload.Repository.Destination.Commit.Hash,
	}
}
