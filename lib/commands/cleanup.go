package commands

import (
	"log"
	"sync"
)

// Cleanup performs a node cleanup on the supplied hostname(s)
func Cleanup(sliceName string, hostnames []string) error {
	var wg sync.WaitGroup

	for _, hostname := range hostnames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()

			log.Printf("Cleaning up node %s", hostname)
			err := ExecCmdOnNode(sliceName, hostname, "rm -rf ~/app", false)
			if err != nil {
				log.Fatal(err)
			}
			err = ExecCmdOnNode(sliceName, hostname, "kill -9 -1", false)
			if err != nil {
				log.Fatal(err)
			}
		}(hostname)
	}

	wg.Wait()
	log.Print("Cleanup completed")

	return nil
}
