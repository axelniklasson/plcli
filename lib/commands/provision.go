package commands

import (
	"log"
	"os"
	"sync"

	"github.com/axelniklasson/plcli/lib/util"
)

// Provision provisions a set of nodes using a provided script
func Provision(scriptPath string, hostnames []string, options *util.Options) error {
	log.Printf("Initiaing provisioning of %d node(s)", len(hostnames))

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		log.Fatalf("Could not find provision script at %s. Got error: %v", scriptPath, err)
	}

	wg := sync.WaitGroup{}

	for idx, n := range hostnames {
		wg.Add(1)
		go func(id int, hostname string) {
			log.Printf("Provision of node %s started by worker %d!", hostname, id)
			defer wg.Done()

			// transfer provision script to node
			err := Transfer(options.Slice, hostname, scriptPath, "~/provision.sh")
			if err != nil {
				log.Printf("Could not transfer provision script to node %s. Error: %v", hostname, err)
				return
			}

			// run provision script on node
			if options.Sudo {
				err = ExecCmdOnNode(options.Slice, hostname, "cd; chmod +x provision.sh; sudo sh provision.sh", true)
			} else {
				err = ExecCmdOnNode(options.Slice, hostname, "cd; chmod +x provision.sh; sh provision.sh", true)
			}
			if err != nil {
				log.Printf("Could not run provision script on node %s. Error: %v", hostname, err)
				return
			}

			// cleanup, remove provision script from node
			err = ExecCmdOnNode(options.Slice, hostname, "cd; rm provision.sh", false)

			log.Printf("Provision of node %s done!", hostname)
		}(idx, n)
	}

	wg.Wait()
	log.Printf("Nodes provisioned!")

	return nil
}
