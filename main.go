package main

import (
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	// "html/template"
	"log"
	"net/http"
	"os"

	// "path/filepath"
	"minimalytics/api"
	"minimalytics/model"
)


type Message struct {
	Event string
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type StatRequest struct {
	Event string `json:"event"`
}

// func serveTemplate(w http.ResponseWriter, r *http.Request) {
// 	lp := filepath.Join("templates", "layout.html")
// 	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

// 	info, err := os.Stat(fp)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			http.NotFound(w, r)
// 			return
// 		}
// 	}

// 	if info.IsDir() {
// 		http.NotFound(w, r)
// 		return
// 	}

// 	tmpl, err := template.ParseFiles(lp, fp)
// 	if err != nil {
// 		log.Print(err.Error())
// 		http.Error(w, http.StatusText(500), 500)
// 		return
// 	}

// 	err = tmpl.ExecuteTemplate(w, "layout", nil)
// 	if err != nil {
// 		log.Print(err.Error())
// 		http.Error(w, http.StatusText(500), 500)
// 	}
// }


func main() {
	model.Init()

	// model.Hello()

	// db, _ = sql.Open("sqlite3", "./events.db")

	// fs := http.FileServer(http.Dir("./static"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/event", api.Middleware(api.HandleEvent))
	// http.HandleFunc("/", serveTemplate)

	http.HandleFunc("/api/dashboards/", api.Middleware(api.HandleDashboard))
	http.HandleFunc("/api/config/", api.Middleware(api.HandleStat))
	http.HandleFunc("/api/stat/", api.Middleware(api.HandleStat))

	log.Println("Starting server on port 3333")

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
