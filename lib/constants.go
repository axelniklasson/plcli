package lib

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// BasePath represents the absolute path to the lib/ folder
	BasePath = filepath.Dir(b)
)

// ConfFile is the plcli conf file residing in users home dir
const ConfFile = ".plcli"

// SSHPort is the port to use when connecting over ssh
const SSHPort = 22

// PLApiURL is the URL to the PlanetLab API
const PLApiURL = "https://www.planet-lab.eu/PLCAPI/"

// PLApiConcurrentWorkers controls the number of workers talking to the PL API allowed to run concurrently
const PLApiConcurrentWorkers = 20

// WorkerPoolSize controls the number of workers allowed to run concurrently
// var because it can be overridden using --workers flag
var WorkerPoolSize = 20
