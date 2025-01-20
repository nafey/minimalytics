package model

type Config struct {
	Id int64
	Key string
	Value string
	CreatedOn string
}

func GetConfig(key string) Config {
	row := db.QueryRow("select * from config where key = ?", key)	

	var configItem Config
	err := row.Scan(&configItem.Id, &configItem.Key, &configItem.Value, &configItem.CreatedOn)

	if err != nil {
		panic ("Not found conifg item")
	}

	return configItem
}
