package config

import (
	"encoding/json"
	"os"
)

type PackageConfig struct {
	PackageName string
	Apks        int
	Apk         bool
	AppData     bool
}

type Config struct {
	DeviceInfo string
	IsRoot     bool
	Contacts   bool
	// Message     bool
	// CallRecords bool
	// Wifi        bool
	// Magisk      bool
	Packages []*PackageConfig
}

func (c *Config) Json() ([]byte, error) {
	return json.Marshal(c)
}
func (c *Config) Save(filename string) error {
	b, err := c.Json()
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}
