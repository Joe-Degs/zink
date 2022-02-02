package config

import (
	"encoding/json"
	"os"
)

var sampleConfig = `{
	"name": "joe",
	"addr": "localhost:6009",
	"peers": [{
		"name": "kofi",
		"addr": "localhost:7000",
	},{
		"name": "messi",
		"addr": "localhost:30011",
	}]
}`

type Config struct {
	Name  string `json:"name"`
	Addr  string `json:"addr"`
	Id    string `json:"id"`
	Peers []struct {
		Name string `json:"name"`
		Addr string `json:"addr"`
		Id   string `json:"id"`
	} `json:"peers"`
}

func LoadJSON(b []byte) (*Config, error) {
	var config Config
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadFile(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return LoadJSON(file)
}

func LoadSampleConfig() (*Config, error) {
	return LoadJSON([]byte(sampleConfig))
}
