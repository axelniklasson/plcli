package commands

import (
	"fmt"
	"io"
	"os"
	"plcli/lib/util"

	"golang.org/x/crypto/ssh"
)

// ExecCmdOnNode executes a command on a hostname over ssh
func ExecCmdOnNode(hostname string, cmd string) error {
	fmt.Printf("Executing \"%s\" on %s\n", cmd, hostname)
	conf := util.GetConf()
	sshConfig := util.GetClientConfig(conf.Slice)

	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", hostname), sshConfig)
	if err != nil {
		fmt.Printf("Failed to dial: %s\n", err)
		return err
	}

	session, err := connection.NewSession()
	if err != nil {
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stdout for session: %v", err)
	}

	go io.Copy(os.Stdout, stdout)

	err = session.Run(cmd)
	return err
}
