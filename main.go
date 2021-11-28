package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/mux"
)

type commitMessage struct {
	ID				string    `json:"Id"`
	Label          string `json:"Label"`
	Message       string `json:"Message"`
}

type allCommitMessages []commitMessage

var commitMessages = allCommitMessages{}

func addCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Add Commit")

	var commit commitMessage
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Enter Message correctly")
		return
	}

	json.Unmarshal(reqBody, &commit)
	commitMessages = append(commitMessages, commit)
	fmt.Println(commit)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(commit)
}

func getCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Get Commit")

	messageID := mux.Vars(r)["id"]

	for _, commit := range commitMessages {
		if commit.ID == messageID {
			json.NewEncoder(w).Encode(commit)
		}
	}
}

func getAllCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Get All Commits")

	json.NewEncoder(w).Encode(commitMessages)
}

func updateCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Update Commit")

	messageID := mux.Vars(r)["id"]
	var updatedCommit commitMessage

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Enter Message correctly")
		return
	}

	json.Unmarshal(reqBody, &updatedCommit)

	for i, commit := range commitMessages {
		if commit.ID == messageID {
			commit.Label = updatedCommit.Label
			commitMessages = append(commitMessages[:i], commitMessages[i:]...)
			commitMessages[i] = commit
			json.NewEncoder(w).Encode(commitMessages)
		}
	}
}

func deleteCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Delete Commit")

	messageID := mux.Vars(r)["id"]

	for i, commit := range commitMessages {
		if commit.ID == messageID {
			commitMessages = append(commitMessages[:i], commitMessages[i+1:]...)
			fmt.Fprintf(w, "The Commit with ID %v has been deleted successfully", messageID)
		}
	}
}
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}


func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/commit", addCommitMessage).Methods("POST")
	router.HandleFunc("/commits", getAllCommitMessage).Methods("GET")
	router.HandleFunc("/commits/{id}", getCommitMessage).Methods("GET")
	router.HandleFunc("/commits/{id}", updateCommitMessage).Methods("PATCH")
	router.HandleFunc("/commits/{id}", deleteCommitMessage).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8092", router))
}