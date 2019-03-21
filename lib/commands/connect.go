package commands

import "fmt"

// ConnectOverSSH sets up an ssh connection to a hostname
func ConnectOverSSH(hostname string) error {
	fmt.Printf("Connecting to %s\n", hostname)
	return nil
}
