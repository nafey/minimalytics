package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port int  `yaml:"port"`
	UI   bool `yaml:"ui"`
}

func CreateConfig() error {
	minimDir, err := getMinimDir()
	if err != nil {
		return err
	}

	configFilepath := filepath.Join(minimDir, "minim.yaml")
	configExists, err := exists(configFilepath)

	if err != nil {
		return err
	}

	if configExists {
		return nil
	}

	config := Config{
		Port: 3333,
		UI:   true,
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilepath, yamlData, 0644)
	if err != nil {
		return err
	}

	fmt.Println("Config file created at", configFilepath)
	return nil

}

func GetConfigFile() (*os.File, error) {
	var configFile *os.File
	minimDir, err := getMinimDir()
	if err != nil {
		return configFile, err
	}

	configFilepath := filepath.Join(minimDir, "minim.yaml")
	configExists, err := exists(configFilepath)

	if err != nil {
		return configFile, err
	}

	if !configExists {
		CreateConfig()
	}

	configFile, err = os.OpenFile(configFilepath, os.O_EXCL, 0644)

	return configFile, err
}

func GetConfig() (Config, error) {
	var config Config

	configFile, err := GetConfigFile()
	if err != nil {
		return config, err
	}
	defer configFile.Close()

	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)

	return config, err
}

func SetConfig(s Config) error {

	return nil

}
