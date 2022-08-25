package router

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"test/model"
	"test/sqldb"
)

var db *sql.DB = sqldb.ConnectDB()

var jwtKey = []byte("secret_key")

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}



func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Routes() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/home", Home).Methods("GET")
	router.HandleFunc("/refresh", Refresh).Methods("GET")
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

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: Login")

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expectedPassword, ok := users[creds.Username]

	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Minute *5)

	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w,
	&http.Cookie{
		Name: "token",
		Value: tokenString,
		Expires: expirationTime,
	})


}

func Home(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr := cookie.Value

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr := cookie.Value

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	
	expirationTime := time.Now().Add(time.Minute *5)

	claims.ExpiresAt = expirationTime.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w,
		&http.Cookie{
			Name: "refresh_token",
			Value: tokenString,
			Expires: expirationTime,
		})
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

