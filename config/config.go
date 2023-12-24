package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Http           string `json:"http,omitempty"`
	StartBlock     uint64 `json:"startBlock,omitempty"`
	DataDir        string `json:"datadir,omitempty"`
	ServerHttpPort int    `json:"serverHttpPort,omitempty"`
}

func InitConfig(path string) *Config {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		fmt.Printf("Read config file error: %v\n", e)
		return nil
	}
	i := Config{
		Http:           "https://api.elastos.io/esc",
		StartBlock:     22291376,
		DataDir:        "./data",
		ServerHttpPort: 20456,
	}

	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	e = json.Unmarshal(file, &i)
	if e != nil {
		fmt.Printf("Unmarshal config file error: %v\n", e)
		return nil
	}

	return &i
}
