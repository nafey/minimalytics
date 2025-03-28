package model

import (
	"time"
)

type Config struct {
	Id        int64
	Key       string
	Value     string
	CreatedOn string
}

func InitConfig() error {
	query := `
		CREATE TABLE IF NOT EXISTS config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT,
			value TEXT,
			createdOn TEXT
		);`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	_, err = GetConfig("PORT")
	if err != nil {
		_, err = db.Exec("insert into config (key, value, createdOn) values (?, ?, ?)", "PORT", "3333", formattedTime)

		if err != nil {
			return err
		}
	}

	_, err = GetConfig("UI_ENABLE")
	if err != nil {
		_, err = db.Exec("insert into config (key, value, createdOn) values (?, ?, ?)", "UI_ENABLE", "1", formattedTime)

		if err != nil {
			return err
		}
	}

	return nil
}

func GetConfig(key string) (Config, error) {
	row := db.QueryRow("select * from config where key = ?", key)

	var configItem Config
	err := row.Scan(&configItem.Id, &configItem.Key, &configItem.Value, &configItem.CreatedOn)

	return configItem, err
}

func GetConfigValue(key string) (string, error) {
	configItem, err := GetConfig(key)
	return configItem.Value, err
}

func SetConfig(key string, val string) error {
	_, err := db.Exec("update config set value = ? where key = ?", val, key)

	return err
}
