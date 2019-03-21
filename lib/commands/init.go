package commands

import "fmt"

// Init generates the ~/.plcli file and performs some other init tasks
func Init() error {
	fmt.Println("Initializing plcli")
	return nil
}
