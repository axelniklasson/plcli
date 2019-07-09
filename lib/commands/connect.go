package commands

import (
	"log"
	"os"
	"os/exec"
	"plcli/lib/util"
	"syscall"
)

// ConnectOverSSH sets up an ssh connection to a hostname
func ConnectOverSSH(slice string, hostname string) error {
	log.Printf("Connecting to %s\n", hostname)

	binary, lookErr := exec.LookPath("ssh")
	if lookErr != nil {
		panic(lookErr)
	}

	conf := util.GetConf()
	args := []string{"ssh", "-l", slice, "-i", conf.PrivateKey, hostname}
	env := os.Environ()

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
	return nil
}
