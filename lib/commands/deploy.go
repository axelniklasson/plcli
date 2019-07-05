package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"plcli/lib/pl"
	"strings"

	"gopkg.in/yaml.v2"
)

type plcliYmlFile struct {
	BootstrapCmds []string          `yaml:"bootstrap_cmds"`
	Env           map[string]string `yaml:"env"`
	LaunchCmds    []string          `yaml:"launch_cmds"`
}

// checks that there is a valid .plcli.yml file in the repo at gitURL and parses it
func parseYML(gitURL string) *plcliYmlFile {
	if !strings.HasSuffix(gitURL, ".git") {
		log.Fatal(errors.New("Please provide a valid git url"))
	}

	cmd := fmt.Sprintf("rm -rf ./tmp && git clone %s ./tmp", gitURL)
	_, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}

	// parse file to bytes arr, fail if not existing or other err
	data, err := ioutil.ReadFile("./tmp/.plcli.yml")
	if err != nil {
		if _, err := os.Stat("./tmp/.plcli.yml"); os.IsNotExist(err) {
			log.Fatal(errors.New("No .plcli.yml in repo"))
		}

		log.Fatal(err)
	}

	// parse yaml to conf struct
	conf := plcliYmlFile{}
	err = yaml.Unmarshal([]byte(data), &conf)
	if err != nil {
		log.Fatal(err)
	}

	// remove tmp dir
	os.RemoveAll("./tmp")
	return &conf
}

func bootstrapNodes(sliceName string, nodes []pl.Node, gitURL string, cmds []string) error {
	for _, n := range nodes {
		log.Printf("Bootstrapping %s", n.HostName)
		err := ExecCmdOnNode(sliceName, n.HostName, "kill -9 -1", false)
		if err != nil {
			return err
		}
		err = ExecCmdOnNode(sliceName, n.HostName, "cd && rm -rf app", false)
		if err != nil {
			return err
		}

		err = ExecCmdOnNode(sliceName, n.HostName, fmt.Sprintf("cd && git clone %s app", gitURL), false)
		if err != nil {
			return err
		}

		cmdString := "cd ~/app"
		for _, cmd := range cmds {
			cmdString += fmt.Sprintf(" && %s", cmd)
		}
		err = ExecCmdOnNode(sliceName, n.HostName, cmdString, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func launchNodes(sliceName string, nodes []pl.Node, env map[string]string, cmds []string) error {
	// build string of env variables and commands that makes up start script
	scriptString := ""
	for k, v := range env {
		scriptString += fmt.Sprintf("export %s=%s; ", k, v)
	}
	for _, cmd := range cmds {
		scriptString += fmt.Sprintf("%s; ", cmd)
	}

	for _, n := range nodes {
		log.Printf("Writing start script ")

		// write start script to node
		err := ExecCmdOnNode(sliceName, n.HostName, fmt.Sprintf("cd ~/app && echo '%s' > start.sh && chmod +x start.sh", scriptString), false)
		if err != nil {
			return err
		}
		// launch app using start script in background
		err = ExecCmdOnNode(sliceName, n.HostName, "cd ~/app; nohup sh start.sh > ~/app.log 2>&1 &", false)
		if err != nil {
			return err
		}
	}

	return nil
}

// Deploy performs a PlanetLab deployment of app at gitUrl on nodeCount nodes using slice sliceName
func Deploy(sliceName string, nodeCount int, gitURL string, skipHealthcheck bool) error {
	log.Printf("Initiating deployment of %s to %d nodes using slice %s", gitURL, nodeCount, sliceName)

	conf := parseYML(gitURL)

	var nodes []pl.Node
	var err error
	if !skipHealthcheck {
		nodes = HealthCheck(sliceName)
	} else {
		log.Printf("Skipping healthcheck of nodes")
		nodes, err = pl.GetNodesForSlice(sliceName)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(nodes) < nodeCount {
		log.Fatal(fmt.Errorf("Could not find enough nodes.. Found %d/%d. Run health check to learn more", len(nodes), nodeCount))
	}

	nodes = nodes[0:nodeCount]
	var hostnames string
	for idx, n := range nodes {
		hostnames += n.HostName
		if idx < len(nodes)-2 {
			hostnames += ", "
		} else if idx == len(nodes)-2 {
			hostnames += " and "
		}
	}
	log.Printf("Nodes that will be used for deployment: %s\n", hostnames)

	err = bootstrapNodes(sliceName, nodes, gitURL, conf.BootstrapCmds)
	if err != nil {
		log.Fatal(err)
	}

	launchNodes(sliceName, nodes, conf.Env, conf.LaunchCmds)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Deployment finished!")
	return nil
}
