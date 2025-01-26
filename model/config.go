package model

import (
	"log"
)

type Config struct {
	Id        int64
	Key       string
	Value     string
	CreatedOn string
}

func InitConfig() {
	query := `
		CREATE TABLE IF NOT EXISTS config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT,
			value TEXT,
			createdOn TEXT
		);`
	_, err := db.Exec(query)
	if err != nil {
		log.Println("failed to create table: %w", err)
		return
	}
	return
}

func GetConfig(key string) Config {

	row := db.QueryRow("select * from config where key = ?", key)

	var configItem Config
	err := row.Scan(&configItem.Id, &configItem.Key, &configItem.Value, &configItem.CreatedOn)

	if err != nil {
		panic("Not found conifg item")
	}

	return configItem
}
