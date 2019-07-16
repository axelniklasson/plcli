package commands

import (
	"fmt"
	"log"
	"plcli/lib"
	"plcli/lib/pl"
	"plcli/lib/util"
	"time"
)

type funcArgs struct {
	SliceName string
	Node      pl.Node
}

type healthCheckResult struct {
	Node      pl.Node
	IsHealthy bool
}

func isHealthy(i interface{}) (interface{}, error) {
	// ping node and see if it is online
	args := i.(funcArgs)
	sliceName := args.SliceName
	node := args.Node

	log.Printf("Performing health check for node %s", node.HostName)

	err := util.PingHost(node.HostName)
	if err != nil {
		log.Printf("Could not ping node %s", node.HostName)
		return healthCheckResult{node, false}, nil
	}

	// try executing a command on node
	err = ExecCmdOnNode(sliceName, node.HostName, "ls /", false)
	if err != nil {
		log.Printf("Could not connect/execute command on node %s", node.HostName)
		return healthCheckResult{node, false}, err
	}

	// transfer healthcheck script
	err = Transfer(sliceName, node.HostName, fmt.Sprintf("%s/scripts/healthcheck.sh", lib.BasePath), "~/healthcheck.sh")
	if err != nil {
		log.Printf("Could not transfer healthcheck script to node %s", node.HostName)
		return healthCheckResult{node, false}, err
	}

	// run healthcheck script
	err = ExecCmdOnNode(sliceName, node.HostName, "cd ~; nohup sh healthcheck.sh > /dev/null 2>&1 &", false)
	if err != nil {
		log.Printf("Something went wrong with running healthcheck script on node %s", node.HostName)
		return healthCheckResult{node, false}, err
	}

	// sleep 3s to wait for healthcheck script to start
	time.Sleep(time.Second * 3)

	// check if port 9876 is opened by healthcheck script. maximum 10 tries.
	tries := 1
	portOpen := false
	for {
		tries = tries + 1

		err = util.CheckPortOpen(node.HostName, 9876)
		if err == nil {
			portOpen = true
			break
		} else {
			// wait for 2s before retry
			time.Sleep(time.Second * 2)
		}

		if tries == 10 {
			break
		}
	}

	if !portOpen {
		log.Printf("Could not open port 9876 on node %s", node.HostName)
		return healthCheckResult{node, false}, err
	}

	// kill healthcheck script and remove from host
	err = ExecCmdOnNode(sliceName, node.HostName, "kill -9 -1; rm ~/healthcheck.sh", false)
	if err != nil {
		return healthCheckResult{node, false}, err
	}

	// if all succeeds, return true
	return healthCheckResult{node, true}, nil
}

// HealthCheck checks all nodes attached to a slice to find out which ones are healthy
// healthy nodes are online and able to open a random port between 3000 and 9999
func HealthCheck(sliceName string, removeFaulty bool) []pl.Node {
	// get all nodes attached to slice
	nodes, err := pl.GetNodesForSlice(sliceName)
	if err != nil {
		log.Fatal(err)
	}

	// setup channels to write jobs and get back jobresults
	jobs := make(chan util.Job, len(nodes))
	results := make(chan util.JobResult, len(nodes))

	// construct jobs and write over channel
	for _, n := range nodes {
		args := funcArgs{sliceName, n}
		workerFunc := func(i interface{}) (interface{}, error) {
			args = i.(funcArgs)
			return isHealthy(args)
		}
		jobs <- util.Job{Func: workerFunc, Args: args}
	}
	close(jobs)

	// launch workers
	workerCount := lib.WorkerPoolSize
	if len(nodes) < workerCount {
		workerCount = len(nodes)
	}
	for i := 0; i < workerCount; i++ {
		go util.Worker(i, jobs, results)
	}

	// gather results, store all healthy nodes in healthyNodes slice
	healthyNodes := []pl.Node{}
	faultyNodes := []pl.Node{}
	for j := 0; j < len(nodes); j++ {
		r := <-results
		jobResult := r.Result.(healthCheckResult)
		if jobResult.IsHealthy {
			healthyNodes = append(healthyNodes, jobResult.Node)
		} else {
			faultyNodes = append(faultyNodes, jobResult.Node)
		}

		log.Printf("Job %d/%d finished!", len(healthyNodes)+len(faultyNodes), len(nodes))
	}

	// pretty-print results from health check
	log.Printf("Found %d healthy and %d faulty nodes!\n", len(healthyNodes), len(faultyNodes))
	prettyPrint("### Healthy nodes ###", healthyNodes)
	prettyPrint("### Faulty nodes ###", faultyNodes)
	fmt.Println("")

	if removeFaulty {
		if len(faultyNodes) == 0 {
			log.Print("No faulty nodes to remove!")
		} else {
			pl.SetNodesForSlice(sliceName, healthyNodes)
		}
	}

	return healthyNodes
}

func prettyPrint(header string, nodes []pl.Node) {
	fmt.Printf("\n%s\n", header)
	if len(nodes) == 0 {
		fmt.Println("No nodes to print!")
	}
	for _, n := range nodes {
		fmt.Printf("%s [%d]\n", n.HostName, n.NodeID)
	}
}

// JobResult is used to write results from workers back to main thread
type JobResult struct {
	Node      pl.Node
	IsHealthy bool
}
