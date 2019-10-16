# planetlab-cli
CLI for managing slices, deploying applications and various other tasks related to PlanetLab nodes.

## Installation

### Build from source
First, make sure that Go is installed and that Go binaries are available in your PATH:
```
export PATH=$PATH:$GOPATH/bin
```

Add it to your bash_profile or similar if you want this to be persistent.

```
go get -u github.com/axelniklasson/plcli
```

You can now run `plcli` and start using the PlanetLab CLI!

## Usage
```
NAME:
   plcli - CLI for PlanetLab

USAGE:
   plcli [global options] command [command options] [arguments...]

VERSION:
   1.0

AUTHOR:
   Axel Niklasson <axel.niklasson@live.com>

COMMANDS:
     init, i           Init plcli
     connect, c        Connect to a PlanetLab node over ssh
     execute, e        Execute a command on a PlanetLab node
     transfer, t       Transfer a file/directory to a PlanetLab node
     slice-details     Lists details for the current slice
     list-nodes        Lists all nodes attached to the current slice
     health-check      Performs a health check of all nodes attached to the slice and outputs healthy nodes
     discover-healthy  Performs a health check of all nodes in the system and outputs hostnames and ids to an output file
     deploy            Deploys an application on PlanetLab nodes
     provision         Provisions node(s) using a provided script
     cleanup           Performs node cleanup on the given nodes
     help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --slice value       name of slice to use when connecting to PlanetLab (default: "chalmersple_2018_10_29")
   --workers value     number of workers to use (default: 20)
   --nodes-file value  file containing node hostnames and ids of the form "ID,HOSTNAME" on each line
   --sudo              if set, everything will be run as sudo on nodes
   --help, -h          show help
   --version, -v       print the version
```