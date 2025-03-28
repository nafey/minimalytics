package model

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

var didInit bool = false
var db *sql.DB

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

func InitCreateDb() {

}

func Init() error {
	if didInit {
		return nil
	}

	var err error
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".minim", "data.db")

	exists, err := exists(dbPath)
	if !exists {

	}

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
		err = InitConfig()
		if err != nil {
			return err
		}
	}

	tab, _ = tableExists("graphs")
	if !tab {
		err = InitGraphs()
		if err != nil {
			return err
		}
	}

	tab, _ = tableExists("dashboards")
	if !tab {
		err = InitDashboards()
		if err != nil {
			return err
		}
	}

	tab, _ = tableExists("events")
	if !tab {
		err = InitEvents()
		if err != nil {
			return err
		}
	}

	didInit = true
	return nil
}
