package util

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHConnection struct {
	Hostname string
	Session  *ssh.Session
	StdOut   io.Reader
	StdErr   io.Reader
}

var sessionPool = map[string]SSHConnection{}

func GetSession(sliceName string, hostname string) (*SSHConnection, error) {
	sshConfig := GetClientConfig(sliceName)

	if session, ok := sessionPool[hostname]; ok {
		return &session, nil
	}

	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", hostname), sshConfig)
	if err != nil {
		log.Printf("Failed to dial: %s\n", err)
		log.Println("Try adding the selected key to the ssh agent through ssh-add")
		return nil, err
	}

	session, err := connection.NewSession()
	if err != nil {
		session.Close()
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("Unable to setup stderr for session: %v", err)
	}

	conn := SSHConnection{hostname, session, stdout, stderr}
	sessionPool[hostname] = conn
	return &conn, nil

}

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

// GetClientConfig returns the client config to use in SSH connections
func GetClientConfig(user string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}
