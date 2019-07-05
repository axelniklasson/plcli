package main

import (
	"log"
	"os"
	"plcli/lib/commands"
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
			Usage:     "Performs a health check of all nodes attached to the slice and outputs IDs of healthy nodes",
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
			Name:      "cleanup",
			Usage:     "Performs node cleanup on the given nodes",
			UsageText: "plcli cleanup HOSTNAME|HOSTNAME1,HOSTNAME2..",
			Action: func(c *cli.Context) error {
				hostnames := strings.Split(c.Args().Get(0), ",")
				return commands.Cleanup(slice, hostnames)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
