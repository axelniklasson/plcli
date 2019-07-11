package commands

import (
	"fmt"
	"log"
	"os"
	"plcli/lib/util"

	scp "github.com/bramvdbogaerde/go-scp"
)

// Transfer copies a local file to a remote PlanetLab node
func Transfer(slice string, hostname string, srcBlob string, targetPath string) error {
	clientConfig := util.GetClientConfig(slice)
	client := scp.NewClient(fmt.Sprintf("%s:22", hostname), clientConfig)

	// Connect to the remote server
	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return err
	}

	f, err := os.Open(srcBlob)
	if err != nil {
		return err
	}

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer f.Close()

	err = client.CopyFile(f, targetPath, "0644")

	if err != nil {
		return err
	}

	log.Printf("Successfully transferred %s to %s:%s\n", srcBlob, hostname, targetPath)

	return nil
}
