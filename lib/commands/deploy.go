package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"plcli/lib"
	"plcli/lib/pl"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type plcliYmlFile struct {
	BootstrapCmds []string          `yaml:"bootstrap_cmds"`
	Env           map[string]string `yaml:"env"`
	LaunchCmds    []string          `yaml:"launch_cmds"`
}

type jobResult struct {
	Node  pl.Node
	Error error
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

// bootstraps a node prior to application launch
func bootstrap(sliceName string, node pl.Node, gitURL string, cmds []string) error {
	log.Printf("Bootstrapping %s", node.HostName)
	err := ExecCmdOnNode(sliceName, node.HostName, "kill -9 -1", false)
	if err != nil {
		return err
	}
	err = ExecCmdOnNode(sliceName, node.HostName, "cd && rm -rf app", false)
	if err != nil {
		return err
	}

	err = ExecCmdOnNode(sliceName, node.HostName, fmt.Sprintf("cd && git clone %s app", gitURL), false)
	if err != nil {
		return err
	}

	cmdString := "cd ~/app"
	for _, cmd := range cmds {
		cmdString += fmt.Sprintf(" && %s", cmd)
	}
	err = ExecCmdOnNode(sliceName, node.HostName, cmdString, false)
	return err
}

// launches an application on a given node
func launch(sliceName string, node pl.Node, scriptString string) error {
	// write start script to node
	err := ExecCmdOnNode(sliceName, node.HostName, fmt.Sprintf("cd ~/app && echo '%s' > start.sh && chmod +x start.sh", scriptString), false)
	if err != nil {
		return err
	}
	// launch app using start script in background
	err = ExecCmdOnNode(sliceName, node.HostName, "cd ~/app; nohup sh start.sh > ~/app.log 2>&1 &", false)
	return err
}

// worker that takes care of bootstrapping a node prior to app launch
func bootstrapWorker(id int, jobs <-chan pl.Node, results chan<- jobResult, sliceName string, gitURL string, cmds []string) {
	log.Printf("Worker %d launched", id)
	for node := range jobs {
		log.Printf("Worker %d bootstrapping node %s", id, node.HostName)
		bootstrapError := bootstrap(sliceName, node, gitURL, cmds)
		// write result of job back to main thread
		results <- jobResult{node, bootstrapError}
	}
}

// worker that takes care of launching app on a node
func launchWorker(id int, jobs <-chan pl.Node, results chan<- jobResult, sliceName string, scriptString string) {
	log.Printf("Worker %d launched", id)
	for node := range jobs {
		log.Printf("Worker %d launching app on node %s", id, node.HostName)
		launchError := launch(sliceName, node, scriptString)
		// write result of job back to main thread
		results <- jobResult{node, launchError}
	}
}

// bootstrap nodes concurrently using workers
func bootstrapNodes(sliceName string, nodes []pl.Node, gitURL string, cmds []string) error {
	jobs := make(chan pl.Node, len(nodes))
	results := make(chan jobResult, len(nodes))

	// launch workers
	workerCount := lib.WorkerPoolSize
	if len(nodes) < workerCount {
		workerCount = len(nodes)
	}
	i := 0
	for i < workerCount {
		go bootstrapWorker(i, jobs, results, sliceName, gitURL, cmds)
		i++
	}

	// write nodes to jobs channel
	for _, n := range nodes {
		jobs <- n
	}
	close(jobs)

	for j := 0; j < len(nodes); j++ {
		res := <-results
		if res.Error != nil {
			log.Fatalf("Bootstrapping of node %s failed with errror: %v", res.Node.HostName, res.Error)
		} else {
			log.Printf("Bootstrapping of node %s succeeded!", res.Node.HostName)
		}
	}
	log.Print("Bootstrapping of nodes completed")
	return nil
}

// launch nodes concurrently using workers
func launchNodes(sliceName string, nodes []pl.Node, env map[string]string, cmds []string) error {
	scriptString := ""
	for k, v := range env {
		scriptString += fmt.Sprintf("export %s=%s; ", k, v)
	}
	for _, cmd := range cmds {
		scriptString += fmt.Sprintf("%s; ", cmd)
	}

	jobs := make(chan pl.Node, len(nodes))
	results := make(chan jobResult, len(nodes))

	// launch workers
	workerCount := lib.WorkerPoolSize
	if len(nodes) < workerCount {
		workerCount = len(nodes)
	}
	i := 0
	for i < workerCount {
		go launchWorker(i, jobs, results, sliceName, scriptString)
		i++
	}

	// write nodes to jobs channel
	for _, n := range nodes {
		jobs <- n
	}
	close(jobs)

	for j := 0; j < len(nodes); j++ {
		res := <-results
		if res.Error != nil {
			log.Fatalf("App launch on node %s failed with errror: %v", res.Node.HostName, res.Error)
		} else {
			log.Printf("App launch on node %s succeeded!", res.Node.HostName)
		}
	}
	log.Print("App launched on all nodes!")
	return nil

}

// Deploy performs a PlanetLab deployment of app at gitUrl on nodeCount nodes using slice sliceName
func Deploy(sliceName string, nodeCount int, gitURL string, skipHealthcheck bool) error {
	start := time.Now()
	log.Printf("Initiating deployment of %s to %d nodes using slice %s", gitURL, nodeCount, sliceName)

	conf := parseYML(gitURL)
	var nodes []pl.Node
	var err error

	// possible healthcheck of nodes
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

	// shuffle nodes
	log.Print("Shuffling nodes")
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })

	// pretty-print nodes that will be used for deployment
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

	// bootstrap all nodes
	err = bootstrapNodes(sliceName, nodes, gitURL, conf.BootstrapCmds)
	if err != nil {
		log.Fatal(err)
	}

	// launch app on all nodes
	err = launchNodes(sliceName, nodes, conf.Env, conf.LaunchCmds)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Deployment finished!")
	elapsed := time.Since(start)
	log.Printf("Deployment to %d nodes took %s", nodeCount, elapsed)
	return nil
}
