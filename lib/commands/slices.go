package commands

import (
	"fmt"
	"plcli/lib/api"
)

// GetDetailsForSlice gets all details for a slice through the API and prints it
func GetDetailsForSlice(slice string) error {
	client := api.GetClient()
	args := make([]interface{}, 2)
	args[0] = api.GetClientAuth()
	args[1] = slice

	slices := []api.Slice{}
	err := client.Call("GetSlices", args, &slices)
	if err != nil {
		panic(err)
	}

	if len(slices) == 0 {
		fmt.Printf("No slice with name %s found\n", slice)
		return nil
	}

	fmt.Println(slices[0].ToString())

	return nil
}
