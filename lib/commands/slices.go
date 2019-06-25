package commands

import (
	"fmt"
	"plcli/lib/pl"
)

// GetDetailsForSlice gets all details for a slice through the API and prints it
func GetDetailsForSlice(slice string) error {
	client := pl.GetClient()
	args := make([]interface{}, 2)
	args[0] = pl.GetClientAuth()
	args[1] = slice

	slices := []pl.Slice{}
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
