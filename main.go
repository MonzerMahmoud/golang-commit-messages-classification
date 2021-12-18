package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"strconv"
	"test/sqldb"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type commitMessage struct {
	ID				int    `json:"id"`
	Label          string `json:"label"`
	Message       string `json:"message"`
}

type message struct {
	Message string `json:"message"`
}

type allCommitMessages []commitMessage

var commitMessages = allCommitMessages{}

var db *sql.DB = sqldb.ConnectDB()

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func addCommitMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Add Commit")

	var message message
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

	//messageID := mux.Vars(r)["id"]
	//messageID,err := strconv.Atoi(mux.Vars(r)["id"])
	//checkErr(err)

	// rows, err := db.Query("SELECT id, label, message FROM commits WHERE id = " + mux.Vars(r)["id"] )

	// defer rows.Close()

	// err = rows.Err()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	
	//ourCommit := commitMessage{}
	// for rows.Next() {
		
	// 	err = rows.Scan(&ourCommit.ID, &ourCommit.Label, &ourCommit.Message)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(ourCommit)
	// }

	// err = rows.Err()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	ourCommit := searchForCommitInDB(mux.Vars(r)["id"])

	json.NewEncoder(w).Encode(ourCommit)
	// for _, commit := range commitMessages {
	// 	if commit.ID == messageID {
	// 		json.NewEncoder(w).Encode(commit)
	// 	}
	// }
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

	
	commits := make([]commitMessage, 0)
	
	for rows.Next() {
		ourCommit := commitMessage{}
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
	// //messageID := mux.Vars(r)["id"]
	// messageID,err := strconv.Atoi(mux.Vars(r)["id"])
	// checkErr(err)
	// var updatedCommit commitMessage

	// reqBody, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	fmt.Fprintf(w, "Enter Message correctly")
	// 	return
	// }

	// json.Unmarshal(reqBody, &updatedCommit)

	// for i, commit := range commitMessages {
	// 	if commit.ID == messageID {
	// 		commit.Label = updatedCommit.Label
	// 		commitMessages = append(commitMessages[:i], commitMessages[i:]...)
	// 		commitMessages[i] = commit
	// 		json.NewEncoder(w).Encode(commitMessages)
	// 	}
	// }
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




	// //messageID := mux.Vars(r)["id"]
	// messageID,err := strconv.Atoi(mux.Vars(r)["id"])
	// checkErr(err)

	// for i, commit := range commitMessages {
	// 	if commit.ID == messageID {
	// 		commitMessages = append(commitMessages[:i], commitMessages[i+1:]...)
	// 		fmt.Fprintf(w, "The Commit with ID %v has been deleted successfully", messageID)
	// 	}
	// }
}


func searchForCommitInDB(id string) commitMessage {
	fmt.Println("Searching For Commit in DB")

	rows, err := db.Query("SELECT id, label, message FROM commits WHERE id = " + id )
	
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	ourCommit := commitMessage{}
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




func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}


func main() {
	//connectToDB()
	routes()
}

// func connectToDB() {
// 	db, err := sql.Open("sqlite3", "./commitMessages.db")

// 	checkErr(err)

// 	defer db.Close()

// 	fmt.Println("Connected to DB")

// 	myDB = db

// }

func routes() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/commit", addCommitMessage).Methods("POST")
	router.HandleFunc("/commits", getAllCommitMessage).Methods("GET")
	router.HandleFunc("/commits/{id}", getCommitMessageById).Methods("GET")
	router.HandleFunc("/commits/{id}", updateCommitMessageById).Methods("PATCH").Queries("label", "{label}")
	router.HandleFunc("/commits/{id}", deleteCommitMessage).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8094", router))
}