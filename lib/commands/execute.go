package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"plcli/lib/util"

	"golang.org/x/crypto/ssh"
)

// func getColor(i int) func(string, ...interface{}) {
// 	possibleColors := []func(string, ...interface{}){
// 		color.Blue, color.Red, color.Green, color.Cyan, color.Magenta, color.Yellow,
// 	}

// 	return possibleColors[i%len(possibleColors)]
// }

// ExecCmdOnNode executes a command on a hostname over ssh
func ExecCmdOnNode(slice string, hostname string, cmd string, showOutput bool) error {
	sshConfig := util.GetClientConfig(slice)

	if cmd == "" {
		return errors.New("Can't execute empty command")
	}

	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", hostname), sshConfig)
	if err != nil {
		log.Printf("Failed to dial: %s\n", err)
		log.Println("Try adding the selected key to the ssh agent through ssh-add")
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
	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stderr for session: %v", err)
	}

	if showOutput {
		go func(stdout io.Reader) {
			reader := io.MultiReader(stdout)
			scanner := bufio.NewScanner(reader)
			c := util.GetColorForHostname(hostname)
			for scanner.Scan() {
				msg := scanner.Text()
				c("%s [stdout] ===> %s\n", hostname, msg)
				// fmt.Printf("%s [stdout] ===> %s\n", hostname, msg)
			}
		}(stdout)

		go func(stderr io.Reader) {
			reader := io.MultiReader(stdout)
			scanner := bufio.NewScanner(reader)
			c := util.GetColorForHostname(hostname)
			for scanner.Scan() {
				msg := scanner.Text()
				c("%s [stderr] ===> %s\n", hostname, msg)
				// fmt.Printf("%s [stderr] ===> %s\n", hostname, msg)
			}
		}(stderr)
	}

	log.Printf("Executing \"%s\" on %s\n", cmd, hostname)
	err = session.Run(cmd)
	return err
}
