package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/axelniklasson/plcli/lib"
	"github.com/axelniklasson/plcli/lib/pl"
	"github.com/axelniklasson/plcli/lib/util"

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
func parseYML(gitURL string, gitBranch string) *plcliYmlFile {
	if !strings.HasSuffix(gitURL, ".git") {
		log.Fatal(errors.New("Please provide a valid git url"))
	}

	cmd := fmt.Sprintf("rm -rf ./tmp && git clone %s ./tmp && cd ./tmp && git checkout %s", gitURL, gitBranch)
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
func bootstrap(node pl.Node, gitURL string, cmds []string, options *util.Options) error {
	log.Printf("Bootstrapping %s", node.HostName)

	cmdsToRun := []string{
		"kill -9 -1",
		fmt.Sprintf("cd && rm -rf logs && rm -rf %s && mkdir logs", options.AppPath),
		fmt.Sprintf("cd && git clone %s %s", gitURL, options.AppPath),
	}

	s := fmt.Sprintf("cd %s && git checkout %s", options.AppPath, options.GitBranch)
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
	err := ExecCmdOnNode(options.Slice, node.HostName, cmdString, false)
	if err != nil {
		return nil
	}

	if options.NodeExporter {
		err := Transfer(options.Slice, node.HostName, fmt.Sprintf("%s/scripts/node_exporter.sh", lib.BasePath), "~/node_exporter.sh")
		if err != nil {
			return err
		}

		log.Println("Launching node_exporter")
		ExecCmdOnNode(options.Slice, node.HostName, "cd ~; pkill node_exporter; chmod +x node_exporter.sh; nohup sh node_exporter.sh > ~/logs/node_exporter.log 2>&1 &", false)
		if err != nil {
			return err
		}
	}

	return nil
}

// launches an application on a given node
func launch(node pl.Node, scriptString string, instanceID int, options *util.Options) error {
	scriptString = fmt.Sprintf("export PLCLI_INSTANCE_ID=%d; ", instanceID) + scriptString

	cmdsToRun := []string{
		fmt.Sprintf("cd %s && echo '%s' > start_instance_%d.sh && chmod +x start_instance_%d.sh", options.AppPath, scriptString, instanceID, instanceID),
	}

	if options.Sudo {
		cmdsToRun = append(cmdsToRun, fmt.Sprintf("cd %s; sudo nohup sh start_instance_%d.sh > ~/logs/instance_%d.log 2>&1 &", options.AppPath, instanceID, instanceID))
	} else {
		cmdsToRun = append(cmdsToRun, fmt.Sprintf("cd %s; nohup sh start_instance_%d.sh > ~/logs/instance_%d.log 2>&1 &", options.AppPath, instanceID, instanceID))
	}

	cmdString := ""
	for idx, c := range cmdsToRun {
		if idx < len(cmdsToRun)-1 {
			cmdString += fmt.Sprintf("%s && ", c)
		} else {
			cmdString += c
		}
	}

	err := ExecCmdOnNode(options.Slice, node.HostName, cmdString, false)
	return err

}

// worker that takes care of bootstrapping a node prior to app launch
func bootstrapWorker(id int, jobs <-chan pl.Node, results chan<- jobResult, gitURL string, cmds []string, options *util.Options) {
	for node := range jobs {
		log.Printf("Worker %d bootstrapping node %s", id, node.HostName)
		bootstrapError := bootstrap(node, gitURL, cmds, options)
		// write result of job back to main thread
		results <- jobResult{node, bootstrapError}
	}
}

// worker that takes care of launching app on a node
func launchWorker(id int, jobs <-chan job, results chan<- jobResult, scriptString string, options *util.Options) {
	for job := range jobs {
		log.Printf("Worker %d launching app instance %d on node %s", id, job.ID, job.Node.HostName)
		launchError := launch(job.Node, scriptString, job.ID, options)
		// write result of job back to main thread
		results <- jobResult{job.Node, launchError}
	}
}

// bootstrap nodes concurrently using workers
// func bootstrapNodes(sliceName string, nodes []pl.Node, gitURL string, gitBranch string, cmds []string) error {
func bootstrapNodes(nodes []pl.Node, gitURL string, cmds []string, options *util.Options) error {
	jobs := make(chan pl.Node, len(nodes))
	results := make(chan jobResult, len(nodes))

	// launch workers
	workerCount := lib.WorkerPoolSize
	if len(nodes) < workerCount {
		workerCount = len(nodes)
	}
	i := 0
	for i < workerCount {
		go bootstrapWorker(i, jobs, results, gitURL, cmds, options)
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
func launchNodes(nodes []pl.Node, env map[string]string, cmds []string, options *util.Options) error {
	instanceCount := len(nodes) * options.Scale
	scriptString := ""
	for k, v := range env {
		scriptString += fmt.Sprintf("export %s=%s; ", k, v)
	}

	if options.EnvVars != "" {
		vars := strings.Split(options.EnvVars, ",")
		for _, v := range vars {
			parts := strings.Split(v, "=")
			if len(parts) != 2 {
				log.Fatal("Badly formatted env string. Should be VAR1=VAL1,VAR2=VAL2,etc")
			}

			scriptString += fmt.Sprintf("export %s=%s; ", parts[0], parts[1])
		}
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
		go launchWorker(i, jobs, results, scriptString, options)
	}

	// create jobs
	jobSlice := []job{}
	for i, n := range nodes {
		for j := i * options.Scale; j < i*options.Scale+options.Scale; j++ {
			jobSlice = append(jobSlice, job{Node: n, ID: i})
		}
	}

	// shuffle jobs
	log.Print("Shuffling jobs")
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

func transferHostFile(nodes []pl.Node, options *util.Options) error {
	buf := bytes.Buffer{}

	for i, n := range nodes {
		for j := i * options.Scale; j < i*options.Scale+options.Scale; j++ {
			ip, err := net.LookupIP(n.HostName)
			if err != nil {
				return err
			}

			ipString := ""
			for idx, x := range ip {
				if idx < len(ip)-1 {
					ipString += fmt.Sprintf("%s.", x)
				} else {
					ipString += fmt.Sprintf("%s", x)
				}
			}

			if i == len(nodes)-1 && j == i*options.Scale+options.Scale-1 {
				buf.WriteString(fmt.Sprintf("%d,%s,%s", j, n.HostName, ipString))
			} else {
				buf.WriteString(fmt.Sprintf("%d,%s,%s\n", j, n.HostName, ipString))
			}
		}
	}

	// transfer hosts file to all nodes
	for _, n := range nodes {
		err := ExecCmdOnNode(options.Slice, n.HostName, fmt.Sprintf("echo '%s' >> %s/hosts.txt", buf.String(), options.AppPath), true)
		if err != nil {
			return err
		}
	}

	if !options.SkipWriteHostsFile {
		f, _ := os.Create("./hosts_deployment.txt")
		defer f.Close()

		f.WriteString(buf.String())
	}

	return nil
}

func removeBlackListed(nodes []pl.Node, blacklistedHostnames []string) []pl.Node {
	okNodes := []pl.Node{}

	for _, node := range nodes {
		ok := true
		for _, blacklisted := range blacklistedHostnames {
			if node.HostName == blacklisted {
				ok = false
				break
			}
		}

		if ok {
			okNodes = append(okNodes, node)
		}
	}

	log.Printf("Removed blacklisted nodes %v from deployment", blacklistedHostnames)

	return okNodes
}

// Deploy performs a PlanetLab deployment of app at gitUrl on nodeCount nodes using slice sliceName
func Deploy(gitURL string, options *util.Options) error {
	start := time.Now()
	log.Printf("Initiating deployment of %d instances of app %s to %d nodes using slice %s ", options.NodeCount*options.Scale, gitURL, options.NodeCount, options.Slice)

	conf := parseYML(gitURL, options.GitBranch)
	var nodes []pl.Node
	var err error

	// possible healthcheck of nodes
	if !options.SkipHealthCheck {
		nodes = HealthCheck(options.Slice, false)
	} else {
		log.Printf("Skipping healthcheck of nodes")
		nodes, err = pl.GetNodesForSlice(options.Slice)
		if err != nil {
			log.Fatal(err)
		}
	}

	if options.BlacklistedHostnames != "" {
		nodes = removeBlackListed(nodes, strings.Split(options.BlacklistedHostnames, ","))
	}

	if len(nodes) < options.NodeCount {
		log.Fatal(fmt.Errorf("Could not find enough nodes.. Found %d/%d. Run health check to learn more", len(nodes), options.NodeCount))
	}

	// shuffle nodes
	if options.ShuffleNodes {
		log.Print("Shuffling nodes")
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })
	}

	// pretty-print nodes that will be used for deployment
	nodes = nodes[0:options.NodeCount]
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
	// err = bootstrapNodes(sliceName, nodes, gitURL, gitBranch, conf.BootstrapCmds)
	err = bootstrapNodes(nodes, gitURL, conf.BootstrapCmds, options)
	if err != nil {
		log.Fatal(err)
	}

	// transfer hosts file to app repo
	err = transferHostFile(nodes, options)
	if err != nil {
		log.Fatal(err)
	}

	// launch app on all nodes
	err = launchNodes(nodes, conf.Env, conf.LaunchCmds, options)
	if err != nil {
		log.Fatal(err)
	}

	if options.PrometheusSDPath != "" {
		writePromSD(nodes, options)
	}

	log.Println("Deployment finished!")
	elapsed := time.Since(start)
	log.Printf("Deployment of %d app instances to %d nodes took %s", options.NodeCount*options.Scale, options.NodeCount, elapsed)
	return nil
}

// writePromSD builds and writes an sd file for prometheus to the given path
func writePromSD(nodes []pl.Node, options *util.Options) {
	// remove file if exists
	os.Remove(options.PrometheusSDPath)

	// create file
	f, err := os.Create(options.PrometheusSDPath)
	if err != nil {
		log.Fatalf("Could not create prom sd file: error: %v", err)
	}
	defer f.Close()

	nodeExporterTargets := ""
	ssurbTargets := ""
	for i, n := range nodes {
		for j := i * options.Scale; j < i*options.Scale+options.Scale; j++ {
			if i == len(nodes)-1 && j == i*options.Scale+options.Scale-1 {
				ssurbTargets += fmt.Sprintf("\"%s:%d\"", n.HostName, 2112+j)
			} else {
				ssurbTargets += fmt.Sprintf("\"%s:%d\",", n.HostName, 2112+j)
			}
		}
		if i == len(nodes)-1 {
			nodeExporterTargets += fmt.Sprintf("\"%s:2100\"", n.HostName)
		} else {
			nodeExporterTargets += fmt.Sprintf("\"%s:2100\",", n.HostName)
		}
	}

	sdString := fmt.Sprintf("[{\"targets\": [%s],\"labels\": { \"env\": \"planetlab\", \"job\": \"self-stabilizing-urb\" }},", ssurbTargets)
	sdString += fmt.Sprintf("{\"targets\": [%s],\"labels\": { \"env\": \"planetlab\", \"job\": \"node_exporter\" }}]", nodeExporterTargets)

	// write to file
	f.WriteString(sdString)

	log.Printf("Wrote sd.json to %s", options.PrometheusSDPath)
}
