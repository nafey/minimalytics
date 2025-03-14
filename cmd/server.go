package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"minim/api"
	"minim/model"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/natefinch/lumberjack"
)

func isServerRunning() (bool, error) {
	pid, err := readPID()
	if err != nil {
		return false, err
	}

	if pid < 1 {
		return false, nil
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

	port, err := model.GetConfigValue("PORT")
	if err != nil {
		return false, err
	}

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

func isUiEnabled() (bool, error) {
	ui_enable_row, err := model.GetConfig("UI_ENABLE")

	if err != nil {
		log.Println("Unable to read UI config")
		return false, err
	}

	ui_enable, err := strconv.Atoi(ui_enable_row.Value)
	if err != nil {
		log.Println("Invalid UI config value")
		return false, err
	}

	if ui_enable != 1 {
		log.Println("UI has been disabled")
		return false, err
	}

	return true, nil
}

func uiMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ui_enabled, err := isUiEnabled()

		if err != nil {
			log.Println(err)
		}

		if !ui_enabled {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ui_enabled, err := isUiEnabled()

		if err != nil {
			log.Println(err)
		}

		if !ui_enabled {
			return
		}

		next(w, r)
	}
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

	err = os.WriteFile(pidFile, []byte(pidStr), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    20, // megabytes
		MaxBackups: 3,
	})

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

	// buildDir := "./static"

	fs := http.FileServer(http.Dir("static"))

	// http.Handle("/", uiMiddleware(fs))
	http.Handle("/ui/", uiMiddleware(fs))

	http.HandleFunc("/api/", (api.Middleware(api.HandleAPIBase)))

	http.HandleFunc("/api/event/", api.Middleware(api.HandleEvent))

	http.HandleFunc("/api/dashboards/", middleware(api.Middleware(api.HandleDashboard)))
	http.HandleFunc("/api/graphs/", middleware(api.Middleware(api.HandleGraphs)))

	http.HandleFunc("/api/stat/", middleware(api.Middleware(api.HandleStat)))
	http.HandleFunc("/api/events/", middleware(api.Middleware(api.HandleEventDefsApi)))
	http.HandleFunc("/api/test/", middleware(api.Middleware(api.HandleTest)))

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	path := filepath.Join(buildDir, r.URL.Path)
	// 	_, err := os.Stat(path)
	// 	if os.IsNotExist(err) {
	// 		http.ServeFile(w, r, filepath.Join(buildDir, "index.html"))
	// 		return
	// 	} else if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	fs.ServeHTTP(w, r)
	// })
	port, err := model.GetConfigValue("PORT")
	if err != nil {
		return err
	}

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
