package main

import (
	"log"
	"test/router"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	router.Routes()
	// Todo - Implement test and production environments
}


