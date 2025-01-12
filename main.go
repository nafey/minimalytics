package main

import (
	"database/sql"
	"fmt"
	"log"
	"errors"
	"io"
	"net/http"
	"encoding/json"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

func openDb() {
	db, err := sql.Open("sqlite3", "./events.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	rows, err := db.Query(`SELECT id, name FROM example`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Printf("User %d: %s\n", id, name)
	}
}

type Message struct {
    Event string
}

func handleEvent(w http.ResponseWriter, r *http.Request) {
	log.Println("Receieved Event")

	decoder := json.NewDecoder(r.Body)

    var t Message
    err := decoder.Decode(&t)
    if err != nil {
        panic(err)
    }

	io.WriteString(w, "OK")
}

func main() {	
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleEvent)

	fmt.Println("Starting server on port 3333")

	err := http.ListenAndServe(":3333", mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
