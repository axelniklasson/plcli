package commands

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

// func readConfFile(filename string) error {

// }

// func writeConfFile(filename string) error {
// 	err := ioutil.WriteFile(filename, nil, 0644)
// 	return err
// }

func getConfFile() (*ini.File, error) {
	return nil, nil

	// homeDir, err := homedir.Dir()
	// if err != nil {
	// 	fmt.Printf("Failed to get users home dir: %v\n", err)
	// 	os.Exit(1)
	// }
	// path := fmt.Sprintf("%s/%s", homeDir, lib.ConfFile)
	// _, err = ini.Load(path)

	// if err != nil {
	// 	fmt.Printf("Fail to read file: %v\n", err)
	// 	writeConfFile(path)
	// 	readConfFile(path)
	// 	os.Exit(1)
	// }
}

// Init generates the ~/.plcli file and performs some other init tasks
func Init() error {
	fmt.Println("Initializing plcli")

	_, err := getConfFile()
	if err != nil {
		fmt.Printf("Could not get plcli conf file: %v\n", err)
		os.Exit(1)
	}

	// write values to conf

	return nil
}
