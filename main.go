package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/u-root/u-root/pkg/curl"
)

var db *sql.DB


type Message struct {
	Event string
}


type EventRow struct {
	Id int64
	Timestamp int64
	Date string
	Event string
	Count int64
}


type StatRequest struct {
	Event string
}


type DateStatItem struct {
	Date string `json:"date"`
	Count int64 `json:"count"`
}


func dailyEvent(msgEvent string) {
	sodTimestamp := int64(0)

	currentTime := time.Now()
	date := currentTime.Format("2006-01-02")

	row := db.QueryRow("select * from days where date = ? and event = ?", date, msgEvent)

	var eventRow EventRow
	err := row.Scan(&eventRow.Id, &eventRow.Timestamp, &eventRow.Date, &eventRow.Event, &eventRow.Count)
	if err != nil {
		db.Exec("insert into days (timestamp, date, event, count) values (?, ?, ?, ?)", sodTimestamp, date, msgEvent, 1)

	} else {

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

	currentTime := time.Now()

	startOfMinute := currentTime.Truncate(time.Minute)
	somTimestamp := startOfMinute.Unix()

	startOfHour := currentTime.Truncate(time.Hour)
	sohTimestamp := startOfHour.Unix()

	// startOfDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location(), )
	// sodTimestamp := startOfDay.Unix()

	minuteEvent(somTimestamp, msgEvent)
	hourEvent(sohTimestamp, msgEvent)
	dailyEvent(msgEvent)

	io.WriteString(w, "OK")
}


func serveTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

	// Return a 404 if the template doesn't exist
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		// Log the detailed error
		log.Print(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func getDailyStat(event string) [30]DateStatItem{
	currentTime := time.Now()
	toDate := currentTime.Format("2006-01-02")
	fromDate := currentTime.AddDate(0, 0, -30).Format("2006-01-02")
	log.Println(fromDate)

	rows, err := db.Query("select * from days where date between ? and ? and event = ?", fromDate, toDate, event)
	if err != nil {
		panic(err)
	}

	countMap := make(map[string]int64)
	for rows.Next() {
		var eventRow EventRow
		err = rows.Scan(&eventRow.Id, &eventRow.Timestamp, &eventRow.Date, &eventRow.Event, &eventRow.Count)
		if err != nil {
			panic(err)
		}

		countMap[eventRow.Date] = eventRow.Count
	}

	var statsArray [30]DateStatItem
	for i := 0; i < 30; i++ {
		iDate := currentTime.AddDate(0, 0, -1 * i).Format("2006-01-02")
		iCount := int64(0)
		
		realCount, ok := countMap[iDate]
		if ok {
			iCount = realCount
		}

		iStatItem := DateStatItem{
			Date: iDate,
			Count: iCount,
		}

		statsArray[i] = iStatItem
	}

	return statsArray
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if (r.URL.Path == "/api/stat/daily/") {

		var statRequest StatRequest
		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(&statRequest)
		if err != nil {
			panic(err)
		}


		stats := getDailyStat(statRequest.Event)

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		}
	}

}


func main() {	
	db, _ = sql.Open("sqlite3", "./events.db")

	fs := http.FileServer(http.Dir("./static"))

	// mux := http.NewServeMux()
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/event", handleEvent)
	http.HandleFunc("/", serveTemplate)
	http.HandleFunc("/api/", handleAPI)

	log.Println("Starting server on port 3333")

	// err := http.ListenAndServe(":3333", mux)
	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
