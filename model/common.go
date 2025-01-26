package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func tableExists(tableName string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error checking table existence: %w", err)
	}
	return true, nil

}

func Init() {
	if didInit {
		return
	}

	var err error
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".minimalytics", "data.db")
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Println("Error:", err)
	}

	_, err = db.Query("select 1")
	if err != nil {
		log.Println("Error:", err)
	}

	if err != nil {
		fmt.Println("Error:", err)
		panic("Unable to connect to database")
	}

	tab, _ := tableExists("config")
	if !tab {
		InitConfig()
	}

	tab, _ = tableExists("dashboards")
	if !tab {
		InitDashboards()
	}

	tab, _ = tableExists("graphs")
	if !tab {
		InitGraphs()
	}

	didInit = true
}
