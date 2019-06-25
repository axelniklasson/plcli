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