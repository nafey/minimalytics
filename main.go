package main

import (
	"database/sql"
	"fmt"
	"log"
	"errors"
	"io"
	"net/http"
	"encoding/json"
	"time"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Message struct {
	Event string
}

type EventRow struct {
	Id int64
	Timestamp int64
	Event string
	Count int64
}


func dailyEvent(sodTimestamp int64, msgEvent string) {
	row := db.QueryRow("select * from days where timestamp = ? and event = ?", sodTimestamp, msgEvent)

	var eventRow EventRow

	err := row.Scan(&eventRow.Id, &eventRow.Timestamp, &eventRow.Event, &eventRow.Count)
	if err != nil {
		// create a row
		db.Exec("insert into days (timestamp, event, count) values (?, ?, ?)", sodTimestamp, msgEvent, 1)
	} else {
		// increment a row
		rowId := eventRow.Id
		nextCount := eventRow.Count + 1
		db.Exec("update days set count = ? where id = ?", nextCount, rowId)
	}
}


func hourEvent(sohTimestamp int64, msgEvent string) {
	row := db.QueryRow("select * from hours where timestamp = ? and event = ?", sohTimestamp, msgEvent)

	var eventRow EventRow

	err := row.Scan(&eventRow.Id, &eventRow.Timestamp, &eventRow.Event, &eventRow.Count)
	if err != nil {
		// create a row
		db.Exec("insert into hours (timestamp, event, count) values (?, ?, ?)", sohTimestamp, msgEvent, 1)
	} else {
		// increment a row
		rowId := eventRow.Id
		nextCount := eventRow.Count + 1
		db.Exec("update hours set count = ? where id = ?", nextCount, rowId)
	}
}


func minuteEvent(somTimestamp int64, msgEvent string) {
	row := db.QueryRow("select * from minutes where timestamp = ? and event = ?", somTimestamp, msgEvent)

	var eventRow EventRow

	err := row.Scan(&eventRow.Id, &eventRow.Timestamp, &eventRow.Event, &eventRow.Count)
	if err != nil {
		// create a row
		db.Exec("insert into minutes (timestamp, event, count) values (?, ?, ?)", somTimestamp, msgEvent, 1)
	} else {
		// increment a row
		rowId := eventRow.Id
		nextCount := eventRow.Count + 1
		db.Exec("update minutes set count = ? where id = ?", nextCount, rowId)
	}
}

func handleEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var t Message
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	msgEvent := t.Event

	log.Println("Event Receieved", msgEvent)

	currentTime := time.Now()

	startOfMinute := currentTime.Truncate(time.Minute)
	somTimestamp := startOfMinute.Unix()

	startOfHour := currentTime.Truncate(time.Hour)
	sohTimestamp := startOfHour.Unix()

	startOfDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location(), )
	sodTimestamp := startOfDay.Unix()

	minuteEvent(somTimestamp, msgEvent)
	hourEvent(sohTimestamp, msgEvent)
	dailyEvent(sodTimestamp, msgEvent)

	io.WriteString(w, "OK")
}

func main() {	
	db, _ = sql.Open("sqlite3", "./events.db")

	mux := http.NewServeMux()
	mux.HandleFunc("/event", handleEvent)

	fmt.Println("Starting server on port 3333")

	err := http.ListenAndServe(":3333", mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
