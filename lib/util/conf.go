package util

import (
	"fmt"
	"os"
	"plcli/lib"

	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
)

// Conf represents a user config related to PlanetLab and plcli
type Conf struct {
	Username   string
	Password   string
	Slice      string
	PrivateKey string
}

// ConfFilePath returns the path for the .plcli file
func ConfFilePath() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		fmt.Printf("Failed to get users home dir: %v\n", err)
		os.Exit(1)
	}

	return fmt.Sprintf("%s/%s", homeDir, lib.ConfFile), nil
}

// ConfFileExists returns bool indicating whether the conf file exists or not
func ConfFileExists() (bool, error) {
	path, err := ConfFilePath()
	if err != nil {
		fmt.Printf("Failed to get conf file path: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
	}
	return true, nil
}

// WriteConfFile writes an empty .plcli file in ~
func WriteConfFile() *ini.File {
	path, _ := ConfFilePath()
	_, err := os.Create(path)
	if err != nil {
		fmt.Printf("Could not write conf file: %v\n", err)
		os.Exit(1)
	}

	cfg, err := ini.Load(path)
	if err != nil {
		fmt.Printf("Could not load conf file: %v\n", err)
		os.Exit(1)
	}

	return cfg
}

func loadConfFromFile() (*ini.File, error) {
	return nil, nil
}

// GetConf returns the current user config
func GetConf() *Conf {
	path, _ := ConfFilePath()
	cfg, err := ini.Load(path)

	if err != nil {
		fmt.Printf("Could not load conf file: %v\n", err)
		os.Exit(1)
	}

	return &Conf{
		cfg.Section("auth").Key("pl_username").String(),
		cfg.Section("auth").Key("pl_password").String(),
		cfg.Section("auth").Key("pl_slice").String(),
		cfg.Section("auth").Key("ssh_key_abs_path").String(),
	}
}
