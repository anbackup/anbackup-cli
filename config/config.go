package config

import (
	"encoding/json"
	"os"
)

type PackageConfig struct {
	PackageName string
	Apk         bool
	AppData     bool
}

type Config struct {
	DeviceInfo  string
	IsRoot      bool
	AddressBook bool
	Message     bool
	CallRecords bool
	Packages    []*PackageConfig
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