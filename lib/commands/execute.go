package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"plcli/lib/util"

	"golang.org/x/crypto/ssh"
)

// ExecCmdOnNode executes a command on a hostname over ssh
func ExecCmdOnNode(slice string, hostname string, cmd string) error {
	sshConfig := util.GetClientConfig(slice)

	if cmd == "" {
		return errors.New("Can't execute empty command")
	}

	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", hostname), sshConfig)
	if err != nil {
		fmt.Printf("Failed to dial: %s\n", err)
		fmt.Println("Try adding the selected key to the ssh agent through ssh-add")
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

	fmt.Printf("Executing \"%s\" on %s\n", cmd, hostname)
	err = session.Run(cmd)
	return err
}
