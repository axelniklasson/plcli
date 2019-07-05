package commands

import "log"

// Cleanup performs a node cleanup on the supplied hostname(s)
func Cleanup(sliceName string, hostnames []string) error {
	for _, hostname := range hostnames {
		log.Printf("Cleaning up node %s", hostname)
		ExecCmdOnNode(sliceName, hostname, "rm -rf ~/app", false)
		ExecCmdOnNode(sliceName, hostname, "kill -9 -1", false)
	}

	log.Print("Cleanup completed")
	return nil
}
