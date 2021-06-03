package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigFileName = "stunserver.yaml"
)

type ServerConfig struct {
	Server1 string `yaml:"server1"`
	Server2 string `yaml:"server2"`
}

func Load() (*ServerConfig, error) {
	fileData, err := ioutil.ReadFile("stunserver.yaml")
	if err != nil {
		return nil, err
	}

	var serverConfig ServerConfig
	err = yaml.Unmarshal(fileData, &serverConfig)
	if err != nil {
		return nil, err
	}

	fmt.Println(serverConfig.Server2)
	return &serverConfig, nil
}
