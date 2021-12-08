package main

import (
	"log"
	"os"
	"strings"

	"github.com/axelniklasson/plcli/lib"
	"github.com/axelniklasson/plcli/lib/commands"
	"github.com/axelniklasson/plcli/lib/pl"
	"github.com/axelniklasson/plcli/lib/util"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "plcli"
	app.Usage = "CLI for PlanetLab"
	app.Version = "1.1"
	app.Authors = []cli.Author{{Name: "Axel Niklasson", Email: "axel.r.niklasson@gmail.com"}}

	conf := util.GetConf()
	options := &util.Options{}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "slice",
			Value:       conf.Slice,
			Usage:       "name of slice to use when connecting to PlanetLab",
			Destination: &options.Slice,
		},
		&cli.IntFlag{
			Name:        "workers",
			Value:       lib.WorkerPoolSize,
			Usage:       "number of workers to use",
			Destination: &lib.WorkerPoolSize,
		},
		&cli.StringFlag{
			Name:        "nodes-file",
			Usage:       "file containing node hostnames and ids of the form \"ID,HOSTNAME\" on each line",
			Destination: &options.NodesFile,
		},
		&cli.BoolFlag{
			Name:        "sudo",
			Usage:       "if set, everything will be run as sudo on nodes",
			Destination: &options.Sudo,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "init",
			Aliases:   []string{"i"},
			Usage:     "Init plcli",
			UsageText: "plcli init",
			Action: func(c *cli.Context) error {
				return commands.Init()
			},
		},
		{
			Name:      "connect",
			Aliases:   []string{"c"},
			Usage:     "Connect to a PlanetLab node over ssh",
			UsageText: "plcli connect [node]",
			Action: func(c *cli.Context) error {
				node := c.Args().Get(0)
				return commands.ConnectOverSSH(options.Slice, node)
			},
		},
		{
			Name:      "execute",
			Aliases:   []string{"e"},
			Usage:     "Execute a command on a PlanetLab node",
			UsageText: "plcli execute [command] [HOSTNAME|all|HOSTNAME1,HOSTNAME2..]",
			Action: func(c *cli.Context) error {
				cmd := c.Args().Get(0)
				hostnamesString := c.Args().Get(1)
				var hostnames []string

				if len(hostnamesString) == 0 {
					log.Fatal("No hostnames found. Run as execute [command] [HOSTNAME|all|HOSTNAME1,HOSTNAME2..]")
				} else if hostnamesString == "all" {
					log.Printf("Finding all nodes attached to slice %s", options.Slice)
					nodes, _ := pl.GetNodesForSlice(options.Slice)
					for _, n := range nodes {
						hostnames = append(hostnames, n.HostName)
					}
				} else {
					hostnames = strings.Split(hostnamesString, ",")
				}

				return commands.ExecCmdOnNodes(hostnames, cmd, options)
			},
		},
		{
			Name:      "transfer",
			Aliases:   []string{"t"},
			Usage:     "Transfer a file/directory to a PlanetLab node",
			UsageText: "plcli transfer [node] [path_to_source_file] [path_to_target]",
			Action: func(c *cli.Context) error {
				node := c.Args().Get(0)
				src := c.Args().Get(1)
				target := c.Args().Get(2)
				return commands.Transfer(options.Slice, node, src, target)
			},
		},
		{
			Name:      "slice-details",
			Usage:     "Lists details for the current slice",
			UsageText: "plcli slice-details",
			Action: func(c *cli.Context) error {
				return commands.GetDetailsForSlice(options.Slice)
			},
		},
		{
			Name:      "list-nodes",
			Usage:     "Lists all nodes attached to the current slice",
			UsageText: "plcli list-nodes",
			Action: func(c *cli.Context) error {
				return commands.GetNodesForSlice(options.Slice)
			},
		},
		{
			Name:      "health-check",
			Usage:     "Performs a health check of all nodes attached to the slice and outputs healthy nodes",
			UsageText: "plcli [--remove-faulty] health-check",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "remove-faulty",
					Usage:       "remove faulty nodes from slice during healthcheck",
					Destination: &options.RemoveFaulty,
				},
				&cli.BoolFlag{
					Name:        "attach-to-slice",
					Usage:       "attach all healthy nodes to slice",
					Destination: &options.AttachToSlice,
				},
			},
			Action: func(c *cli.Context) error {
				commands.HealthCheck(options.Slice, options.RemoveFaulty)
				return nil
			},
		},
		{
			Name:      "discover-healthy",
			Usage:     "Performs a health check of all nodes in the system and outputs hostnames and ids to an output file",
			UsageText: "plcli [--attach-to-slice] discover-healthy",
			Action: func(c *cli.Context) error {
				return commands.DiscoverHealthyNodes(options.Slice, options.AttachToSlice)
			},
		},
		{
			Name:      "deploy",
			Usage:     "Deploys an application on PlanetLab nodes",
			UsageText: "plcli deploy GIT_URL",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:        "node-count",
					Usage:       "number of nodes to deploy to",
					Destination: &options.NodeCount,
				},
				&cli.BoolFlag{
					Name:        "skip-healthcheck",
					Usage:       "skip health check when deploying",
					Destination: &options.SkipHealthCheck,
				},
				&cli.IntFlag{
					Name:        "scale",
					Value:       1,
					Usage:       "number of instances of app to launch on each node",
					Destination: &options.Scale,
				},
				&cli.StringFlag{
					Name:        "git-branch",
					Value:       "master",
					Usage:       "what branch to use in deployment",
					Destination: &options.GitBranch,
				},
				&cli.StringFlag{
					Name:        "app-path",
					Value:       "~/app",
					Usage:       "where the app should be stored on a node during deployment",
					Destination: &options.AppPath,
				},
				&cli.StringFlag{
					Name:        "prometheus-sd-path",
					Usage:       "if present, plcli will generate sd.json for prometheus and write to supplied path",
					Destination: &options.PrometheusSDPath,
				},
				&cli.BoolFlag{
					Name:        "node-exporter",
					Usage:       "if set, node-exporter will be installed and launched on port 2100",
					Destination: &options.NodeExporter,
				},
				&cli.BoolFlag{
					Name:        "shuffle-nodes",
					Usage:       "if set, nodes form PL api will be shuffled prior to deployment",
					Destination: &options.ShuffleNodes,
				},
				&cli.BoolFlag{
					Name:        "skip-write-hosts-file",
					Usage:       "if set, no file called hosts_deployment.txt will be written to current directory",
					Destination: &options.SkipWriteHostsFile,
				},
				&cli.StringFlag{
					Name:        "blacklist",
					Usage:       "HOST1,HOST2,... string of hostnames to blacklist in the deployment",
					Destination: &options.BlacklistedHostnames,
				},
				&cli.StringFlag{
					Name:        "env",
					Usage:       "VAR1=VAL1,VAR2=VAL2,... string of env vars to use in deployment",
					Destination: &options.EnvVars,
				},
			},
			Action: func(c *cli.Context) error {
				gitURL := c.Args().Get(0)
				return commands.Deploy(gitURL, options)
			},
		},
		{
			Name:      "provision",
			Usage:     "Provisions node(s) using a provided script",
			UsageText: "plcli provision PATH_TO_SCRIPT HOSTNAME|all|HOSTNAME1,HOSTNAME1",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 2 {
					log.Fatal("Run as provision PATH_TO_SCRIPT HOSTNAME|HOSTNAME1,HOSTNAME1")
				}
				provisionScriptPath := c.Args().Get(0)
				hostnamesString := c.Args().Get(1)
				var hostnames []string

				if len(hostnamesString) == 0 {
					log.Fatal("No hostnames found. Run as provision PATH_TO_SCRIPT HOSTNAME|all|HOSTNAME1,HOSTNAME2..")
				} else if hostnamesString == "all" {
					log.Printf("Finding all nodes attached to slice %s", options.Slice)
					nodes, _ := pl.GetNodesForSlice(options.Slice)
					for _, n := range nodes {
						hostnames = append(hostnames, n.HostName)
					}
				} else {
					hostnames = strings.Split(hostnamesString, ",")
				}

				return commands.Provision(provisionScriptPath, hostnames, options)
			},
		},
		{
			Name:      "cleanup",
			Usage:     "Performs node cleanup on the given nodes",
			UsageText: "plcli cleanup HOSTNAME|HOSTNAME1,HOSTNAME2..",
			Action: func(c *cli.Context) error {
				hostnamesString := c.Args().Get(0)
				var hostnames []string

				if len(hostnamesString) == 0 {
					log.Fatal("No hostnames found. Run as cleanup all|HOSTNAME|HOSTNAME1,HOSTNAME2..")
				} else if hostnamesString == "all" {
					nodes, _ := pl.GetNodesForSlice(options.Slice)
					for _, n := range nodes {
						hostnames = append(hostnames, n.HostName)
					}
				} else {
					hostnames = strings.Split(hostnamesString, ",")
				}
				return commands.Cleanup(options.Slice, hostnames)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
