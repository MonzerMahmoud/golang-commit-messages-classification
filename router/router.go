package router

import (
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"test/model"
	"test/sqldb"
)

var db *sql.DB = sqldb.ConnectDB()

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Routes() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/commit", addCommitMessage).Methods("POST")
	router.HandleFunc("/commits", getAllCommitMessage).Methods("GET")
	router.HandleFunc("/commits/{id}", getCommitMessageById).Methods("GET")
	router.HandleFunc("/commits/{id}", updateCommitMessageById).Methods("PATCH").Queries("label", "{label}")
	router.HandleFunc("/commits/{id}", deleteCommitMessage).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func addCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Add Commit")

	var message model.Message
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Enter Message correctly")
		return
	}

	json.Unmarshal(reqBody, &message)

	stmt, err := db.Prepare("INSERT INTO commits (label, message) VALUES (?, ?)")
	checkErr(err)
	stmt.Exec("not_labeled", message.Message)
	defer stmt.Close()

	fmt.Println("Added commit to DB")

	w.WriteHeader(http.StatusCreated)

	fmt.Fprintf(w, "Message Added Successfully!")
	//json.NewEncoder(w).Encode(commit)
}

func getCommitMessageById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Get Commit")
	// TODO: - Show error message if commit not found

	ourCommit := searchForCommitInDB(mux.Vars(r)["id"])

	json.NewEncoder(w).Encode(ourCommit)
}

// Todo - Get the next un labeled commit

func getAllCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Get All Commits")
	// TODO:- Show error message if no commits found
	rows, err := db.Query("SELECT * FROM commits")

	defer rows.Close()

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	
	commits := make([]model.CommitMessage, 0)
	
	for rows.Next() {
		ourCommit := model.CommitMessage{}
		err = rows.Scan(&ourCommit.ID, &ourCommit.Label, &ourCommit.Message)
		if err != nil {
			log.Fatal(err)
		}
		commits = append(commits, ourCommit)
		fmt.Println(ourCommit)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(commits)
}

func updateCommitMessageById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Update Commit")
	
	// TODO: - Show error message if label query is not found

	ourMessage := searchForCommitInDB(mux.Vars(r)["id"])

	ourMessage.Label = r.FormValue("label")

	stmt, err := db.Prepare("UPDATE commits set label = ? where id = ?")
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(ourMessage.Label, ourMessage.ID)
	checkErr(err)

	affected, err := res.RowsAffected()
	checkErr(err)

	fmt.Println(affected)

	json.NewEncoder(w).Encode(searchForCommitInDB(mux.Vars(r)["id"]))
}

func deleteCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Delete Commit")

	ourCommit := searchForCommitInDB(mux.Vars(r)["id"])

	stmt, err := db.Prepare("DELETE FROM commits where id = ?")
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(ourCommit.ID)
	checkErr(err)

	affected, err := res.RowsAffected()
	checkErr(err)
	fmt.Println(affected)

	fmt.Fprintf(w, "The Commit with ID %v has been deleted successfully", ourCommit.ID)

}

func searchForCommitInDB(id string) model.CommitMessage {
	fmt.Println("Searching For Commit in DB")

	rows, err := db.Query("SELECT id, label, message FROM commits WHERE id = " + id )
	
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	ourCommit := model.CommitMessage{}
	for rows.Next() {
		
		err = rows.Scan(&ourCommit.ID, &ourCommit.Label, &ourCommit.Message)
		if err != nil {
			log.Fatal(err)
		}

	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return ourCommit
}

