package commands

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/axelniklasson/plcli/lib/pl"
)

// GetNodesForSlice returns the list of nodes attached to the given slice
func GetNodesForSlice(slice string) error {
	slices, err := pl.GetSlices(slice)
	if err != nil {
		log.Fatal(err)
	}

	output := ""
	if len(slices) == 0 {
		output = "No slices found"
	} else {
		output = fmt.Sprintf("IDs of nodes attached to slice %s: %v", slice, slices[0].NodeIDs)
	}

	log.Println(output)
	return nil
}

const nodesFile = "nodes.txt"

func writeNodesToFile(nodes []pl.Node) error {
	f, err := os.Create(nodesFile)
	defer f.Close()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		f.WriteString(fmt.Sprintf("%s,%d\n", n.HostName, n.NodeID))
	}

	log.Printf("Wrote list of nodes to %s!", nodesFile)

	return nil
}

// ParseNodesFile parses a file of hostnames/IDs per line and returns a slice of the found IDs
func ParseNodesFile(path string) ([]int, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("Could not load file at %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	nodeIDs := []int{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return nil, errors.New("malformed nodes file")
		}
		nodeID, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		nodeIDs = append(nodeIDs, nodeID)
	}

	return nodeIDs, nil
}
