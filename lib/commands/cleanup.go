package commands

import (
	"log"
	"sync"
)

// Cleanup performs a node cleanup on the supplied hostname(s)
func Cleanup(sliceName string, hostnames []string) error {
	var wg sync.WaitGroup
	cmds := []string{
		"rm -rf ~/app",
		"rm ~/healthcheck.sh",
		"kill -9 -1",
	}

	for _, hostname := range hostnames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			log.Printf("Cleaning up node %s", hostname)
			for _, c := range cmds {
				err := ExecCmdOnNode(sliceName, hostname, c, false)
				if err != nil {
					log.Printf("Got error when executing command %s on %s: %v", c, hostname, err)
				}
			}

		}(hostname)
	}

	wg.Wait()
	log.Print("Cleanup completed")

	return nil
}
