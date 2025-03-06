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
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

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

func setup() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	path := filepath.Join(homeDir, ".minimalytics")
	isDir, err := exists(path)

	if err != nil {
		fmt.Println("Error:", err)
	}

	if !isDir {
		os.MkdirAll(path, 0755)
	}

	model.Init()
	model.DeleteEvents()
}

func servermain() {

	setup()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			model.DeleteEvents()
		}
	}()

	http.HandleFunc("/event/", api.Middleware(api.HandleEvent))
	http.HandleFunc("/api/dashboards/", api.Middleware(api.HandleDashboard))
	http.HandleFunc("/api/graphs/", api.Middleware(api.HandleGraphs))

	http.HandleFunc("/api/stat/", api.Middleware(api.HandleStat))
	http.HandleFunc("/api/events/", api.Middleware(api.HandleEventDefsApi))
	http.HandleFunc("/api/test/", api.Middleware(api.HandleTest))

	log.Println("Starting server on port 3333")

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "status",
				Usage: "Shows the status of Minimalytics",
				Action: func(cCtx *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "login",
				Usage: "Generate a code to login to the Minimalytics UI/API",
				Action: func(cCtx *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "server",
				Usage: "Control the Minimalytics server",
				Subcommands: []*cli.Command{
					{
						Name:  "start",
						Usage: "Start the server daemon",
						Action: func(cCtx *cli.Context) error {
							// fmt.Println("new task template: ", cCtx.Args().First())
							return nil
						},
					},
					{
						Name:  "stop",
						Usage: "stop the server daemon",
						Action: func(cCtx *cli.Context) error {
							return nil
						},
					},
				},
			},
			{
				Name:  "ui",
				Usage: "Control the Minimalytics UI",
				Subcommands: []*cli.Command{
					{
						Name:  "enable",
						Usage: "Enable the Minimalytics UI",
						Action: func(cCtx *cli.Context) error {
							return nil
						},
					},
					{
						Name:  "disable",
						Usage: "Disable the Minimalytics UI",
						Action: func(cCtx *cli.Context) error {
							return nil
						},
					},
				},
			},
			{
				Name:  "version",
				Usage: "Print the Minimalytics version",
				Action: func(cCtx *cli.Context) error {
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
