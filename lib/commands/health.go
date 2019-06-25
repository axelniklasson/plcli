package commands

import (
	"fmt"
	"log"
	"plcli/lib/pl"
	"plcli/lib/util"
)

func isHealthy(node pl.Node) bool {
	conf := util.GetConf()

	// ping node and see if it is online
	canPing := util.CanPingHost(node.HostName)
	if canPing == false {
		log.Printf("Could not ping node %s", node.HostName)
		return false
	}

	// try executing a command on node
	err := ExecCmdOnNode(conf.Slice, node.HostName, "ls /", false)
	if err != nil {
		log.Printf("Could not connect/execute command on node %s", node.HostName)
		return false
	}

	// try to ping google.com from node

	// try to open port

	// if all succeeds, return true
	return true
}

// HealthCheck checks all nodes attached to a slice to find out which ones are healthy
// healthy nodes are online and able to open a random port between 3000 and 9999
func HealthCheck(sliceName string) error {
	// get all nodes attached to slice
	nodes, err := pl.GetNodesForSlice(sliceName)
	if err != nil {
		panic(err)
	}

	// perform healthchecks on all nodes
	healthyNodes := []pl.Node{}
	for _, node := range nodes {
		nodeHealthy := isHealthy(node)
		if nodeHealthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	// pretty-print results from health check
	fmt.Printf("Found %d healthy nodes!\n", len(healthyNodes))
	return nil
}
