package commands

import (
	"log"
	"plcli/lib/pl"
	"time"
)

// DiscoverHealthyNodes checks all nodes in the system to find out which are healthy
func DiscoverHealthyNodes(sliceName string, attachToSlice bool) error {
	nodes := pl.GetNodeIDsForSlice(sliceName)
	log.Printf("Current list of nodes attached to slice %s: %v", sliceName, nodes)

	// get all nodes in the system
	allNodes := pl.GetAllNodes()

	// attach all nodes in system to this slice
	err := pl.SetNodesForSlice(sliceName, allNodes)
	if err != nil {
		return err
	}

	// wait for 20 minutes to allow the change in node ids of the slice to propagate throughout the system..
	log.Print("Sleeping for 20 mins to allow for slice update to propagate throughout the system..")
	time.Sleep(time.Minute * 20)

	// perform health check on all attached nodes
	healthyNodes := HealthCheck(sliceName, false)

	// attach all healthy nodes to slice if desired, otherwise restore
	if attachToSlice {
		// healthyNodes is a list of Nodes, hence the call to SetNodeForSlice
		pl.SetNodesForSlice(sliceName, healthyNodes)
	} else {
		// nodes is a list of ids, hence the call to SetNodeIDsForSlice
		pl.SetNodeIDsForSlice(sliceName, nodes)
	}

	err = writeNodesToFile(healthyNodes)
	if err != nil {
		return err
	}

	return nil
}
