package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"minim/api"
	"minim/model"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jxskiss/mcli"
	_ "github.com/mattn/go-sqlite3"
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

const PORT = 3333

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

//		err = tmpl.ExecuteTemplate(w, "layout", nil)
//		if err != nil {
//			log.Print(err.Error())
//			http.Error(w, http.StatusText(500), 500)
//		}
//	}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func startServer() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	minimDir := filepath.Join(homeDir, ".minim")
	isDir, err := exists(minimDir)

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	if !isDir {
		os.MkdirAll(minimDir, 0755)
	}

	log.Print(">>>>>>>>>>>>>>>>>> 2")

	pidFile := filepath.Join(minimDir, "minim.pid")
	logFile := filepath.Join(minimDir, "minim.log")

	log.Print(pidFile)
	log.Print(logFile)

	log.Print(">>>>>>>>>>>>>>>>>> 6")

	model.Init()
	model.DeleteEvents()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			model.DeleteEvents()
		}
	}()

	http.HandleFunc("/event/", api.Middleware(api.HandleEvent))

	http.HandleFunc("/api/", api.Middleware(api.HandleAPIBase))
	http.HandleFunc("/api/dashboards/", api.Middleware(api.HandleDashboard))
	http.HandleFunc("/api/graphs/", api.Middleware(api.HandleGraphs))

	http.HandleFunc("/api/stat/", api.Middleware(api.HandleStat))
	http.HandleFunc("/api/events/", api.Middleware(api.HandleEventDefsApi))
	http.HandleFunc("/api/test/", api.Middleware(api.HandleTest))

	port := strconv.Itoa(PORT)
	log.Println("Starting server on port " + port)

	err = http.ListenAndServe(":"+port, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
		return err

	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
		return err
	}

	return nil
}

func runCmd1() {
	exepath := os.Args[0]
	cmd := exec.Command(exepath, "server")
	cmd.Start()
}

func runCmd2() {
	log.Print(">>>>>>>> dummy")
}

func runCmd3() {
	startServer()
}

func main() {

	mcli.Add("status", runCmd1, "View the status")

	mcli.AddGroup("server", "Commands for managing Minimalytics server")
	mcli.Add("server start", runCmd1, "Start the server")
	mcli.Add("server stop", runCmd1, "Stop the server")

	// mcli.Add("dummy", runCmd2, "")

	mcli.AddHidden("execserver", runCmd3, "")
	mcli.Run()

}
