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
	"runtime"
	"strconv"
	"time"

	"github.com/jxskiss/mcli"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
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

// func setup() {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	path := filepath.Join(homeDir, ".minim")
// 	isDir, err := exists(path)

// 	if err != nil {
// 		fmt.Println("Error:", err)
// 	}

// 	if !isDir {
// 		os.MkdirAll(path, 0755)
// 	}

// 	cntxt := &daemon.Context{
// 		PidFileName: "minim.pid",
// 		PidFilePerm: 0644,
// 		LogFileName: "minim.log",
// 		LogFilePerm: 0640,
// 		WorkDir:     homeDir,
// 		Umask:       027,
// 		Args:        []string{"minim server"},
// 	}

// 	d, err := cntxt.Reborn()
// 	if err != nil {
// 		log.Fatal("Unable to run: ", err)
// 	}
// 	if d != nil {
// 		return
// 	}
// 	defer cntxt.Release()

// 	log.Print("- - - - - - - - - - - - - - -")
// 	log.Print("daemon started")

// 	model.Init()
// 	model.DeleteEvents()
// }

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

func climain() {
	// startServer()

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
							log.Print(">>>>>>>>>>>>>>>>>> 1")
							log.Print(runtime.Version())
							err := startServer()
							if err == nil {
								print("Err is nil")
							} else {
								print(err)
							}
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

	if len(os.Args) < 2 {
		return
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runCmd1() {
	exepath := os.Args[0]
	cmd := exec.Command(exepath, "server")
	cmd.Start()
}

func runCmd2() {
}

func runCmd3() {
	startServer()
}

func main() {
	// fmt.Println(len(os.Args), os.Args)
	mcli.Add("start", runCmd1, "An awesome command cmd1")
	mcli.Add("dummy", runCmd2, "")
	mcli.Add("server", runCmd3, "")
	mcli.Run()

	// exepath := (os.Args[0])

	// cmd := exec.Command(exepath, "start")
	// cmd.Start()

	// if len(os.Args) > 1 && os.Args[1] == "start" {
	// 	startServer()
	// }

}
