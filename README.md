# planetlab-cli
CLI for managing slices, deploying applications and various other tasks related to PlanetLab nodes.

## Installation
First, make sure that Go is installed and that Go binaries are available in your PATH:
```
export PATH=$PATH:$GOPATH/bin
```

Add it to your bash_profile or similar if you want this to be persistent.

```
cd $GOPATH/src && git clone https://github.com/axelniklasson/plcli.git && cd plcli
go get -v
go install
```

You can now run `plcli` and start using the PlanetLab CLI!