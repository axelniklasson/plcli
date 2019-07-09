package main

import (
	"log"
	"os"
	"plcli/lib/commands"
	"plcli/lib/pl"
	"plcli/lib/util"
	"strings"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "plcli"
	app.Usage = "CLI for PlanetLab"
	app.Version = "0.1"
	app.Author = "Axel Niklasson <axel.niklasson@live.com>"

	conf := util.GetConf()

	var slice string
	var nodeCount int
	var skipHealthcheck bool
	var output string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "slice",
			Value:       conf.Slice,
			Usage:       "name of slice to use when connecting to PlanetLab",
			Destination: &slice,
		},
		cli.IntFlag{
			Name:        "node-count",
			Usage:       "number of nodes to deploy to",
			Destination: &nodeCount,
		},
		cli.BoolFlag{
			Name:        "skip-healthcheck",
			Usage:       "skip health check when deploying",
			Destination: &skipHealthcheck,
		},
		cli.StringFlag{
			Name:        "output",
			Usage:       "file to write output to (if applicable)",
			Destination: &output,
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
				return commands.ConnectOverSSH(slice, node)
			},
		},
		{
			Name:      "execute",
			Aliases:   []string{"e"},
			Usage:     "Execute a command on a PlanetLab node",
			UsageText: "plcli execute [node] [command]",
			Action: func(c *cli.Context) error {
				node := c.Args().Get(0)
				cmd := c.Args().Get(1)
				return commands.ExecCmdOnNode(slice, node, cmd, true)
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
				return commands.Transfer(slice, node, src, target)
			},
		},
		{
			Name:      "slice-details",
			Usage:     "Lists details for the current slice",
			UsageText: "plcli slice-details",
			Action: func(c *cli.Context) error {
				return commands.GetDetailsForSlice(slice)
			},
		},
		{
			Name:      "list-nodes",
			Usage:     "Lists all nodes attached to the current slice",
			UsageText: "plcli list-nodes",
			Action: func(c *cli.Context) error {
				return commands.GetNodesForSlice(slice)
			},
		},
		{
			Name:      "health-check",
			Usage:     "Performs a health check of all nodes attached to the slice and outputs healthy nodes",
			UsageText: "plcli health-check",
			Action: func(c *cli.Context) error {
				commands.HealthCheck(slice)
				return nil
			},
		},
		{
			Name:      "deploy",
			Usage:     "Deploys an application on PlanetLab nodes",
			UsageText: "plcli deploy GIT_URL",
			Action: func(c *cli.Context) error {
				gitURL := c.Args().Get(0)
				return commands.Deploy(slice, nodeCount, gitURL, skipHealthcheck)
			},
		},
		{
			Name:      "provision",
			Usage:     "Provisions node(s) using a provided script",
			UsageText: "plcli provision PATH_TO_SCRIPT HOSTNAME|HOSTNAME1,HOSTNAME1",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 2 {
					log.Fatal("Run as provision PATH_TO_SCRIPT HOSTNAME|HOSTNAME1,HOSTNAME1")
				}
				provisionScriptPath := c.Args().Get(0)
				hostnamesString := c.Args().Get(1)
				var hostnames []string

				if len(hostnamesString) == 0 {
					log.Fatal("No hostnames found. Run as provision PATH_TO_SCRIPT HOSTNAME|HOSTNAME1,HOSTNAME2..")
				} else if hostnamesString == "all" {
					log.Printf("Finding all nodes attached to slice %s", slice)
					nodes, _ := pl.GetNodesForSlice(slice)
					for _, n := range nodes {
						hostnames = append(hostnames, n.HostName)
					}
				} else {
					hostnames = strings.Split(hostnamesString, ",")
				}

				return commands.Provision(slice, provisionScriptPath, hostnames)
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
					nodes, _ := pl.GetNodesForSlice(slice)
					for _, n := range nodes {
						hostnames = append(hostnames, n.HostName)
					}
				} else {
					hostnames = strings.Split(hostnamesString, ",")
				}
				return commands.Cleanup(slice, hostnames)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
