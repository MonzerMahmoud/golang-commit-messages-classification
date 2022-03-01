package model

type CommitMessage struct {
	ID				int    `json:"id"`
	Label          string `json:"label"`
	Message       string `json:"message"`
}

type Message struct {
	Message string `json:"message"`
}