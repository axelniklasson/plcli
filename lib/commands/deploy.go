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

type job struct {
	Node pl.Node
	ID   int
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

	cmdsToRun := []string{
		"kill -9 -1",
		"cd && rm -rf * && mkdir logs",
		fmt.Sprintf("cd && git clone %s app", gitURL),
	}

	s := "cd ~/app"
	for _, cmd := range cmds {
		s += fmt.Sprintf(" && %s", cmd)
	}
	cmdsToRun = append(cmdsToRun, s)

	cmdString := ""
	for idx, c := range cmdsToRun {
		if idx < len(cmdsToRun)-1 {
			cmdString += fmt.Sprintf("%s && ", c)
		} else {
			cmdString += c
		}
	}

	// execute all commands chained as one
	err := ExecCmdOnNode(sliceName, node.HostName, cmdString, false)
	return err
}

// launches an application on a given node
func launch(sliceName string, node pl.Node, scriptString string, instanceID int) error {
	scriptString = fmt.Sprintf("export PLCLI_INSTANCE_ID=%d; ", instanceID) + scriptString

	cmdsToRun := []string{
		fmt.Sprintf("cd ~/app && echo '%s' > start_instance_%d.sh && chmod +x start_instance_%d.sh", scriptString, instanceID, instanceID),
		fmt.Sprintf("cd ~/app; nohup sh start_instance_%d.sh > ~/logs/instance_%d.log 2>&1 &", instanceID, instanceID),
	}

	cmdString := ""
	for idx, c := range cmdsToRun {
		if idx < len(cmdsToRun)-1 {
			cmdString += fmt.Sprintf("%s && ", c)
		} else {
			cmdString += c
		}
	}

	err := ExecCmdOnNode(sliceName, node.HostName, cmdString, false)
	return err

}

// worker that takes care of bootstrapping a node prior to app launch
func bootstrapWorker(id int, jobs <-chan pl.Node, results chan<- jobResult, sliceName string, gitURL string, cmds []string) {
	for node := range jobs {
		log.Printf("Worker %d bootstrapping node %s", id, node.HostName)
		bootstrapError := bootstrap(sliceName, node, gitURL, cmds)
		// write result of job back to main thread
		results <- jobResult{node, bootstrapError}
	}
}

// worker that takes care of launching app on a node
func launchWorker(id int, jobs <-chan job, results chan<- jobResult, sliceName string, scriptString string) {
	for job := range jobs {
		log.Printf("Worker %d launching app instance %d on node %s", id, job.ID, job.Node.HostName)
		launchError := launch(sliceName, job.Node, scriptString, job.ID)
		// write result of job back to main thread
		results <- jobResult{job.Node, launchError}
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
func launchNodes(sliceName string, nodes []pl.Node, env map[string]string, cmds []string, scale int) error {
	instanceCount := len(nodes) * scale
	scriptString := ""
	for k, v := range env {
		scriptString += fmt.Sprintf("export %s=%s; ", k, v)
	}
	for _, cmd := range cmds {
		scriptString += fmt.Sprintf("%s; ", cmd)
	}

	jobs := make(chan job, len(nodes))
	results := make(chan jobResult, len(nodes))

	// launch workers
	workerCount := lib.WorkerPoolSize
	if instanceCount < workerCount {
		workerCount = len(nodes)
	}
	for i := 0; i < workerCount; i++ {
		go launchWorker(i, jobs, results, sliceName, scriptString)
	}

	// create jobs
	jobSlice := []job{}
	for _, n := range nodes {
		for i := 0; i < scale; i++ {
			jobSlice = append(jobSlice, job{Node: n, ID: i})
		}
	}

	// shuffle jobs
	log.Print("Shuffling nodes")
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(jobSlice), func(i, j int) { jobSlice[i], jobSlice[j] = jobSlice[j], jobSlice[i] })

	// write shuffled jobs to job channel
	for _, j := range jobSlice {
		jobs <- j
	}
	close(jobs)

	launches := 0
	for j := 0; j < instanceCount; j++ {
		res := <-results
		if res.Error != nil {
			log.Fatalf("Instance launch on node %s failed with errror: %v", res.Node.HostName, res.Error)
		}
		launches++
		log.Printf("%d/%d instances launched! ", launches, instanceCount)
	}
	log.Print("App launched on all nodes!")
	return nil

}

// Deploy performs a PlanetLab deployment of app at gitUrl on nodeCount nodes using slice sliceName
func Deploy(sliceName string, nodeCount int, gitURL string, skipHealthcheck bool, scale int) error {
	start := time.Now()
	log.Printf("Initiating deployment of %d instances of app %s to %d nodes using slice %s ", nodeCount*scale, gitURL, nodeCount, sliceName)

	conf := parseYML(gitURL)
	var nodes []pl.Node
	var err error

	// possible healthcheck of nodes
	if !skipHealthcheck {
		nodes = HealthCheck(sliceName, false)
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
	err = launchNodes(sliceName, nodes, conf.Env, conf.LaunchCmds, scale)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Deployment finished!")
	elapsed := time.Since(start)
	log.Printf("Deployment of %d app instances to %d nodes took %s", nodeCount*scale, nodeCount, elapsed)
	return nil
}
