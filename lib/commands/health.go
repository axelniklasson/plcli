package commands

import (
	"fmt"
	"log"
	"plcli/lib"
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

	// transfer healthcheck script

	// run healthcheck script

	// check if port open

	// kill healthcheck script and remove from host

	// if all succeeds, return true
	return true
}

// worker used to healthcheck nodes
func worker(id int, jobs <-chan pl.Node, results chan<- JobResult) {
	log.Printf("Worker %d launched", id)
	// collect jobs from channel
	for n := range jobs {
		log.Printf("Worker %d checking node %s", id, n.HostName)
		nodeHealthy := isHealthy(n)
		// write result of job back to main thread
		results <- JobResult{n, nodeHealthy}
	}
}

// HealthCheck checks all nodes attached to a slice to find out which ones are healthy
// healthy nodes are online and able to open a random port between 3000 and 9999
func HealthCheck(sliceName string) []pl.Node {
	// get all nodes attached to slice
	nodes, err := pl.GetNodesForSlice(sliceName)
	if err != nil {
		log.Fatal(err)
	}

	jobs := make(chan pl.Node, len(nodes))
	results := make(chan JobResult, len(nodes))

	// launch workers
	i := 0
	for i < lib.WorkerPoolSize {
		go worker(i, jobs, results)
		i++
	}

	// write nodes to jobs channel
	for _, n := range nodes {
		jobs <- n
	}
	// done with writing to jobs channel, close it
	close(jobs)

	// gather results, store all healthy nodes in healthyNodes slice
	healthyNodes := []pl.Node{}
	faultyNodes := []pl.Node{}
	for j := 0; j < len(nodes); j++ {
		r := <-results
		if r.IsHealthy {
			healthyNodes = append(healthyNodes, r.Node)
		} else {
			faultyNodes = append(faultyNodes, r.Node)
		}
	}

	// pretty-print results from health check
	log.Printf("Found %d healthy and %d faulty nodes!\n", len(healthyNodes), len(faultyNodes))
	prettyPrint("### Healthy nodes ###", healthyNodes)
	prettyPrint("### Faulty nodes ###", faultyNodes)
	fmt.Println("")
	return healthyNodes
}

func prettyPrint(header string, nodes []pl.Node) {
	fmt.Printf("\n%s\n", header)
	for _, n := range nodes {
		fmt.Printf("%s [%d]\n", n.HostName, n.NodeID)
	}
}

// JobResult is used to write results from workers back to main thread
type JobResult struct {
	Node      pl.Node
	IsHealthy bool
}
