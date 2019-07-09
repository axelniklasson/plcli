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

	detailedNodes := make([]Node, len(nodeIDs))
	// fetch details about each node and return slice of all nodes, populated with information
	for idx, id := range nodeIDs {
		node, err := GetNodeDetails(id)
		if err != nil {
			log.Fatal(err)
		}

		detailedNodes[idx] = node
	}

	return detailedNodes, nil
}
