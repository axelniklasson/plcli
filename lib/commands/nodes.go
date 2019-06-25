package commands

import (
	"fmt"
	"plcli/lib/pl"
)

// GetNodesForSlice returns the list of nodes attached to the given slice
func GetNodesForSlice(slice string) error {
	slices, err := pl.GetSlices(slice)
	if err != nil {
		panic(err)
	}

	output := ""
	if len(slices) == 0 {
		output = "No slices found"
	} else {
		output = fmt.Sprintf("IDs of nodes attached to slice %s: %v", slice, slices[0].NodeIDs)
	}

	fmt.Println(output)
	return nil
}
