package commands

import (
	"fmt"
	"plcli/lib/api"
)

// GetNodesForSlice returns the list of nodes attached to the given slice
func GetNodesForSlice(slice string) error {
	client := api.GetClient()
	args := make([]interface{}, 2)
	args[0] = api.GetClientAuth()
	args[1] = slice

	slices := []api.Slice{}
	err := client.Call("GetSlices", args, &slices)
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
