package main

import (
	"log"
	"os"
	"plcli/lib/commands"
	"plcli/lib/util"

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

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "slice",
			Value:       conf.Slice,
			Usage:       "slice to use when connecting to PlanetLab",
			Destination: &slice,
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
			UsageText: "plcli connect [hostname]",
			Action: func(c *cli.Context) error {
				hostname := c.Args().Get(0)
				return commands.ConnectOverSSH(slice, hostname)
			},
		},
		{
			Name:      "execute",
			Aliases:   []string{"e"},
			Usage:     "Execute a command on a PlanetLab node",
			UsageText: "plcli execute [hostname] [command]",
			Action: func(c *cli.Context) error {
				hostname := c.Args().Get(0)
				cmd := c.Args().Get(1)
				return commands.ExecCmdOnNode(slice, hostname, cmd)
			},
		},
		{
			Name: "slice-details",
			Action: func(c *cli.Context) error {
				return commands.GetDetailsForSlice(slice)
			},
		},
		{
			Name: "list-nodes",
			Action: func(c *cli.Context) error {
				return commands.GetNodesForSlice(slice)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
