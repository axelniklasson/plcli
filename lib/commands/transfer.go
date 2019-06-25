package commands

import (
	"fmt"
	"os"
	"plcli/lib/util"

	scp "github.com/bramvdbogaerde/go-scp"
)

// Transfer copies a local file to a remote PlanetLab node
func Transfer(slice string, node string, srcBlob string, targetPath string) error {
	clientConfig := util.GetClientConfig(slice)
	client := scp.NewClient(fmt.Sprintf("%s:22", node), clientConfig)

	// Connect to the remote server
	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return err
	}

	f, _ := os.Open(srcBlob)

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer f.Close()

	err = client.CopyFile(f, targetPath, "0644")

	if err != nil {
		return err
	}

	fmt.Printf("Successfully transferred %s to %s:%s\n", srcBlob, node, targetPath)

	return nil
}
