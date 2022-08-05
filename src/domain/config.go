package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	toml "github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v2"
)

const (
	Ini  = "ini"
	Hcl  = "hcl"
	Yml  = "yml"
	JSON = "json"
	Yaml = "yaml"
	Toml = "toml"

	// default delimiter
	defaultDelimiter = "."
)

// Config represents a configuration file.
type Config struct {
	filename string
	cache    *AppCfg
}

// New creates a new Config object.
func newConfig(filename string) *Config {
	config := Config{filename, nil}
	err := config.Reload()
	if err != nil {
		log.Println("newConfig error:", err)
		return nil
	}
	go config.watch()
	return &config
}

// Reload clears the config cache.
func (config *Config) Reload() error {
	cache, err := primeCacheFromFile(config.filename)
	config.cache = cache

	if err != nil {
		return err
	}

	return nil
}

func (config *Config) watch() {

	// Catch SIGHUP to automatically reload cache
	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)

	for {
		<-sighup
		log.Printf("Caught SIGHUP, reloading config...")
		config.Reload()
		cfg = config.cache

	}
}

// fix inc/conf/yaml format
func fixFormat(f string) string {
	if f == Yml {
		f = Yaml
	}

	if f == "inc" {
		f = Ini
	}

	// eg nginx config file.
	if f == "conf" {
		f = Hcl
	}

	return f
}

func primeCacheFromFile(file string) (*AppCfg, error) {
	// File exists?
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}

	//get fmt
	posts := strings.Split(file, defaultDelimiter)
	format := fixFormat(posts[len(posts)-1])

	// Read file
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var config AppCfg
	switch format {
	case Yaml:
		if err := yaml.Unmarshal(raw, &config); err != nil {
			return nil, err
		}
	case JSON:
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, err
		}
	case Toml:
		if _, err := toml.Decode(string(raw), &config); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("not supported yet")
	}

	return &config, nil
}
