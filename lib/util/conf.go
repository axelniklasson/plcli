package util

import (
	"fmt"
	"log"
	"os"

	"github.com/axelniklasson/plcli/lib"

	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
)

var conf = Conf{}

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
		log.Fatalf("Failed to get users home dir: %v\n", err)
	}

	return fmt.Sprintf("%s/%s", homeDir, lib.ConfFile), nil
}

// ConfFileExists returns bool indicating whether the conf file exists or not
func ConfFileExists() (bool, error) {
	path, err := ConfFilePath()
	if err != nil {
		log.Fatalf("Failed to get conf file path: %v\n", err)
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
		log.Fatalf("Could not write conf file: %v\n", err)
	}

	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalf("Could not load conf file: %v\n", err)
	}

	return cfg
}

func loadConfFromFile() (*ini.File, error) {
	return nil, nil
}

// GetConf returns the current user config
func GetConf() *Conf {
	if (conf != Conf{}) {
		return &conf
	}

	path, _ := ConfFilePath()
	cfg, err := ini.Load(path)

	if err != nil {
		log.Printf("Could not load conf file: %v. Run plcli init.\n", err)
		return &Conf{Slice: ""}
	}

	conf = Conf{
		cfg.Section("auth").Key("pl_username").String(),
		cfg.Section("auth").Key("pl_password").String(),
		cfg.Section("auth").Key("pl_slice").String(),
		cfg.Section("auth").Key("ssh_key_abs_path").String(),
	}

	return &conf
}
