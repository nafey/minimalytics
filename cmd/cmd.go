package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"minim/model"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

func Init() {
	_, err := getMinimDir()

	if err != nil {
		fmt.Println(err)
	}

	model.Init()
}

func readPID() (int, error) {
	minimDir, err := getMinimDir()
	if err != nil {
		return -1, err
	}

	pidFile := filepath.Join(minimDir, "minim.pid")

	exists, err := exists(pidFile)
	if err != nil {
		return -1, err
	}

	if !exists {
		_, err := os.Create(pidFile)
		if err != nil {
			return -1, err
		}
	}

	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return -1, err
	}

	pidStr := strings.TrimSpace(string(pidBytes))

	if pidStr == "" {
		return -1, nil
	}

	pid, err := strconv.Atoi(pidStr)
	return pid, err
}

func execserver() {
	exepath := os.Args[0]
	cmd := exec.Command(exepath, "execserver")
	cmd.Start()
}

func CmdServerStart() {
	out, err := isServerRunning()

	if err != nil {
		fmt.Println(err)
		return
	}

	if out {
		fmt.Println("Server is already running")
		return
	}

	execserver()
}

func CmdServerStop() {
	out, err := isServerRunning()

	if err != nil {
		fmt.Println(err)
		return
	}

	if !out {
		fmt.Println("Server is not running")
		return
	}

	stopServer()
}

func CmdServerRestart() {
	running, err := isServerRunning()

	if err != nil {
		fmt.Println(err)
		return
	}

	if running {
		CmdServerStop()
	}

	execserver()

	fmt.Println("Restarted the server")
}

func CmdExecServer() {
	err := startServer()
	if err != nil {
		log.Print(err)
	}
}

func CmdStatus() {
	out, err := isServerRunning()
	if err != nil {
		fmt.Println("Error encountered: ", err)
	}

	port, err := model.GetConfigValue("PORT")
	if err != nil {
		fmt.Println(err)
		return
	}

	if out {
		fmt.Println("Server is running on Port:", port)
	} else {
		fmt.Println("Server is not running")
	}

	minimDir, err := getMinimDir()

	if err != nil {
		fmt.Println("Unable to access minim dir at.", minimDir)
	} else {
		fmt.Println("Minimalytics directory location:", minimDir)
	}

}

func CmdUiEnable() {
	err := model.SetConfig("UI_ENABLE", "1")
	if err != nil {
		fmt.Println(err)

	}
}

func CmdUiDisable() {
	err := model.SetConfig("UI_ENABLE", "0")
	if err != nil {
		fmt.Println(err)
	}
}
