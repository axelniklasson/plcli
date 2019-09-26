package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/axelniklasson/plcli/lib/util"
)

func getStringFromUser(msg string) (string, error) {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		return scanner.Text(), nil
	}

	if scanner.Err() != nil {
		return "", scanner.Err()
	}
	return "", nil
}

// Init generates the ~/.plcli file and performs some other init tasks
func Init() error {
	if exists, _ := util.ConfFileExists(); exists == true {
		log.Printf("~/.plcli file already exists, aborting..\n")
		os.Exit(0)
	}

	log.Println("Initializing plcli..")

	// conf file does not exist, first write empty file
	cfg := util.WriteConfFile()

	plUsername, _ := getStringFromUser("What is your PlanetLab username? ")
	plPassword, _ := getStringFromUser("What is your PlanetLab password? ")
	plSlice, _ := getStringFromUser("What PlanetLab slice is your default one when connecting? ")
	sshKeyAbsPath, _ := getStringFromUser("What is the absolute path to your ssh key used when connecting to PlanetLab? ")

	cfg.NewSection("auth")
	cfg.Section("auth").NewKey("pl_username", plUsername)
	cfg.Section("auth").NewKey("pl_password", plPassword)
	cfg.Section("auth").NewKey("pl_slice", plSlice)
	cfg.Section("auth").NewKey("ssh_key_abs_path", sshKeyAbsPath)

	path, _ := util.ConfFilePath()
	err := cfg.SaveTo(path)

	log.Println("Saving .plcli file")

	if err != nil {
		log.Printf("Could not save .plcli file: %v\n", err)
		os.Exit(1)
	}

	return nil
}
