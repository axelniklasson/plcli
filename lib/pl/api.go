package pl

import (
	"fmt"
	"log"
)

// GetSlices queries the PL API and returns all slices matching sliceName
func GetSlices(sliceName string) ([]Slice, error) {
	client := GetClient()
	args := make([]interface{}, 2)
	args[0] = GetClientAuth()
	args[1] = sliceName

	slices := []Slice{}
	err := client.Call("GetSlices", args, &slices)
	if err != nil {
		log.Fatal(err)
	}

	return slices, nil
}

// GetNodeDetails returns details about a given node
func GetNodeDetails(nodeID int) (Node, error) {
	log.Printf("Fetching details about node with ID %d", nodeID)
	client := GetClient()
	args := make([]interface{}, 2)
	args[0] = GetClientAuth()
	args[1] = nodeID

	nodes := []Node{}
	err := client.Call("GetNodes", args, &nodes)
	if err != nil {
		log.Fatal(err)
	}

	if len(nodes) == 0 {
		panic("Could not find a node")
	} else if len(nodes) > 1 {
		panic("Found more than one node, should not happen..")
	}

	return nodes[0], nil
}

// GetNodesDetails returns details about given nodes
func GetNodesDetails(nodeIDs []int) []Node {
	log.Printf("Fetching details about nodes with IDs %v", nodeIDs)
	client := GetClient()
	args := make([]interface{}, 2)
	args[0] = GetClientAuth()
	args[1] = nodeIDs

	nodes := []Node{}
	err := client.Call("GetNodes", args, &nodes)
	if err != nil {
		log.Fatal(err)
	}

	return nodes
}

// GetAllNodes returns all nodes in the system
func GetAllNodes() []Node {
	client := GetClient()
	args := make([]interface{}, 2)
	args[0] = GetClientAuth()

	nodes := []Node{}
	err := client.Call("GetNodes", args, &nodes)
	if err != nil {
		log.Fatal(err)
	}

	return nodes
}

// GetNodeIDsForSlice returns the IDs of all nodes for a given slice
func GetNodeIDsForSlice(sliceName string) []int {
	slices, _ := GetSlices(sliceName)

	if len(slices) > 1 {
		log.Fatal("Found more than one slice, please enter slice name correctly")
	} else if len(slices) == 0 {
		log.Fatalf("Found no slice matching %s, please enter slice name correctly", sliceName)
	}

	return slices[0].NodeIDs
}

// GetNodesForSlice fetches IDs of all attached nodes for the slice and then returns detailed
// info about all of them
func GetNodesForSlice(sliceName string) ([]Node, error) {
	slices, _ := GetSlices(sliceName)

	if len(slices) > 1 {
		log.Fatal("Found more than one slice, please enter slice name correctly")
		return nil, nil
	} else if len(slices) == 0 {
		log.Fatal(fmt.Sprintf("Found no slice matching %s, please enter slice name correctly", sliceName))
	}

	nodeIDs := slices[0].NodeIDs
	detailedNodes := GetNodesDetails(nodeIDs)

	log.Printf("Finished fetching details of all %d nodes", len(detailedNodes))
	return detailedNodes, nil
}

// SetNodeIDsForSlice updates the field nodes of a given slice with the list of node ids
func SetNodeIDsForSlice(sliceName string, nodeIDs []int) error {
	client := GetClient()
	args := make([]interface{}, 3)
	args[0] = GetClientAuth()
	args[1] = sliceName

	// build list of node ids
	nodeIDsArg := struct {
		Nodes []int `xmlrpc:"nodes"`
	}{nodeIDs}
	args[2] = nodeIDsArg

	var res int
	err := client.Call("UpdateSlice", args, &res)
	if err != nil {
		log.Fatal(err)
	}

	if res != 1 {
		log.Fatalf("Something went wrong when updating nodes of slice %s", sliceName)
	}

	log.Printf("Updated nodes of slice %s to be %v", sliceName, nodeIDs)
	return nil
}

// SetNodesForSlice is used to set what nodes should be attached to a given slice
func SetNodesForSlice(sliceName string, nodes []Node) error {
	nodeIDs := []int{}
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.NodeID)
	}

	return SetNodeIDsForSlice(sliceName, nodeIDs)
}
