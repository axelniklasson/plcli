package util

import (
	"fmt"
	"log"
	"net"
	"time"
)

// CanPingHost is a helper that tries to ping a given hostname
func CanPingHost(hostName string) bool {
	log.Printf("Pinging node %s\n", hostName)
	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", hostName), timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	log.Printf("Node %s responded!", hostName)
	return true
}

// PortOpen tries to connect to hostname:port over TCP and returns whether it succeeds or not
func PortOpen(hostName string, port int) bool {
	log.Printf("Checking if port %d on node %s is open", port, hostName)
	_, err := net.Dial("tcp", fmt.Sprintf("http://%s:%d", hostName, port))
	return err != nil
}
