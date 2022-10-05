package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
					Hash    string `json:"hash"`
					Message string `json:"message"`
					Author  struct {
						Raw string `json:"raw"`
					} `json:"author"`
				} `json:"target"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
}

type BitbucketPullRequest struct {
	Pullrequest struct {
		Title       string `json:"title"`
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
		Source struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
			Commit struct {
				Hash string `json:"hash"`
			} `json:"commit"`
		} `json:"source"`
	} `json:"pullrequest"`
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
	case "pullrequest:created":
		commit_info = bitbucketPullRequestCreatedOrUpdated(w, b)
	case "pullrequest:updated":
		commit_info = bitbucketPullRequestCreatedOrUpdated(w, b)
	default:
		w.WriteHeader(500)
		w.Write([]byte("Event type not yet implemented"))
		return
	}

	// Error if commit info is empty
	for k := range commit_info {
		if commit_info[k] == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Unsupported payload"))
			return
		}
	}

	COMMIT_MESSAGE_PREFIX_TO_IGNORE := os.Getenv("COMMIT_MESSAGE_PREFIX_TO_IGNORE")
	COMMIT_AUTHOR_TO_IGNORE := os.Getenv("COMMIT_AUTHOR_TO_IGNORE")
	PUSH_TRIGGER_ONLY_IF_BRANCHES := os.Getenv("PUSH_TRIGGER_ONLY_IF_BRANCHES")

	if COMMIT_MESSAGE_PREFIX_TO_IGNORE != "" {
		if strings.HasPrefix(commit_info["message"], COMMIT_MESSAGE_PREFIX_TO_IGNORE) {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Accepted, but ignored"))
			return
		}
	}

	if COMMIT_AUTHOR_TO_IGNORE != "" {
		if commit_info["author"] == COMMIT_AUTHOR_TO_IGNORE {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Accepted, but ignored"))
			return
		}
	}

	if PUSH_TRIGGER_ONLY_IF_BRANCHES != "" {
		if !stringInSlice(strings.Split(PUSH_TRIGGER_ONLY_IF_BRANCHES, ","), commit_info["destination_branch"]) {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Accepted, but ignored"))
			return
		}
	}

	fmt.Println("\nWebhooker: Requested accepted, proceeding...")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	commit_info["event_type"] = required_headers["event_type"]

	commit_info_json, err := json.Marshal(commit_info)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	execPipe(string(commit_info_json))
}

func bitbucketRepoPush(w http.ResponseWriter, b []byte) map[string]string {
	var payload BitbucketRepoPush
	err := json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	if len(payload.Push.Changes) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported payload"))
	}

	return map[string]string{
		"destination_branch": payload.Push.Changes[0].New.Name,
		"new_commit_hash":    payload.Push.Changes[0].New.Target.Hash,
		"old_commit_hash":    payload.Push.Changes[0].Old.Target.Hash,
		"source_branch":      payload.Push.Changes[0].New.Name,
		"message":            payload.Push.Changes[0].New.Target.Message,
		"author":             payload.Push.Changes[0].New.Target.Author.Raw,
	}
}

func bitbucketPullRequestCreatedOrUpdated(w http.ResponseWriter, b []byte) map[string]string {
	var payload BitbucketPullRequest
	err := json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	return map[string]string{
		"destination_branch": payload.Pullrequest.Destination.Branch.Name,
		"new_commit_hash":    payload.Pullrequest.Source.Commit.Hash,
		"old_commit_hash":    payload.Pullrequest.Destination.Commit.Hash,
		"source_branch":      payload.Pullrequest.Source.Branch.Name,
		"message":            payload.Pullrequest.Title,
	}
}
