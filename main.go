package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"minim/api"
	"minim/model"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
	Data    any    `json:"data,omitempty"`
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

func getMinimDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	minimDir := filepath.Join(homeDir, ".minim")
	isDir, err := exists(minimDir)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	if isDir {
		return minimDir, nil
	}

	err = os.MkdirAll(minimDir, 0755)
	if err != nil {
		return "", err
	}

	return minimDir, nil
}

func readPID() (int, error) {
	minimDir, err := getMinimDir()
	if err != nil {
		return -1, err
	}

	pidFile := filepath.Join(minimDir, "minim.pid")
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return -1, err
	}

	pidStr := strings.TrimSpace(string(pidBytes))

	pid, err := strconv.Atoi(pidStr)
	return pid, err
}

func isServerRunning() (bool, error) {
	pid, err := readPID()
	if err != nil {
		return false, err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		if err.Error() == "os: process already finished" {
			return false, nil
		}

		return false, err
	}

	if process == nil {
		return false, nil
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		if err.Error() == "os: process already finished" {
			return false, nil
		}
		return false, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	port := strconv.Itoa(PORT)
	resp, err := client.Get("http://localhost:" + port + "/api/")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var apiResp Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return false, err
	}

	if apiResp.Status != "OK" || apiResp.Message != "Success" {
		return false, err
	}

	return true, nil
}

func stopServer() error {
	pid, err := readPID()
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process with PID %d: %v\n", pid, err)
		return err
	}

	err = process.Kill()
	if err != nil {
		fmt.Printf("Failed to kill process %d: %v\n", pid, err)
		return err
	}

	_, err = process.Wait()

	if err != nil {
		if err.Error() == "wait: no child processes" {
			fmt.Println("Server has been stopped")
			return nil
		}

		fmt.Printf("Error waiting for process to terminate: %v\n", err)
		return err
	}

	fmt.Println("Server has been stopped")
	return nil
}

func startServer() error {
	minimDir, err := getMinimDir()
	if err != nil {
		fmt.Println("Error accessing minim directory")
		return err
	}

	pidFile := filepath.Join(minimDir, "minim.pid")
	logFile := filepath.Join(minimDir, "minim.log")

	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	// Write to file
	err = os.WriteFile(pidFile, []byte(pidStr), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("-------------- Starting Server ---------------")

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

func cmdServerStart() {
	out, err := isServerRunning()

	if err != nil {
		fmt.Println(err)
		return
	}

	if out {
		fmt.Println("Server is already running")
		return
	}

	exepath := os.Args[0]
	cmd := exec.Command(exepath, "execserver")
	cmd.Start()
}

func cmdServerStop() {
	out, err := isServerRunning()

	if err != nil {
		fmt.Println(err)
		return
	}

	if !out {
		fmt.Print("Server is not running")
	}

	stopServer()
}

func cmdExecServer() {
	out := startServer()
	log.Print(out)
}

func cmdStatus() {
	out, err := isServerRunning()
	if err != nil {
		fmt.Print("Error encountered: ", err)
	}

	if out {
		fmt.Print("Server is running")
	} else {
		fmt.Print("Server is not running")
	}
}

func main() {
	mcli.Add("status", cmdStatus, "View the status")

	mcli.AddGroup("server", "Commands for managing Minimalytics server")
	mcli.Add("server start", cmdServerStart, "Start the server")
	mcli.Add("server stop", cmdServerStop, "Stop the server")

	mcli.AddHidden("execserver", cmdExecServer, "")

	mcli.Run()
}
