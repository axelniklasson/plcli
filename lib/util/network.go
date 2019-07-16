package util

import (
	"fmt"
	"log"
	"net"
	"time"
)

// PingHost is a helper that tries to ping a given hostname
func PingHost(hostName string) error {
	log.Printf("Pinging node %s\n", hostName)
	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", hostName), timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return err
	}
	log.Printf("Node %s responded!", hostName)
	return nil
}

// CheckPortOpen tries to connect to hostname:port over TCP and returns whether it succeeds or not
func CheckPortOpen(hostName string, port int) error {
	log.Printf("Checking if port %d on node %s is open", port, hostName)
	_, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostName, port))
	return err
}
