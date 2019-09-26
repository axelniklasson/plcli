package pl

import (
	"log"

	"github.com/axelniklasson/plcli/lib"
	"github.com/axelniklasson/plcli/lib/util"

	"github.com/kolo/xmlrpc"
)

// Auth models the authentication object to use when authanticating against PL API
type Auth struct {
	AuthMethod string
	Username   string
	AuthString string
}

// GetClientAuth returns an Auth struct needed to authenticate against PL API
func GetClientAuth() Auth {
	conf := util.GetConf()
	return Auth{
		"password",
		conf.Username,
		conf.Password,
	}
}

var clientInstance *xmlrpc.Client

// GetClient creates and returns
func GetClient() *xmlrpc.Client {
	if clientInstance == nil {
		clientInstance, err := xmlrpc.NewClient(lib.PLApiURL, nil)
		if err != nil {
			log.Fatal(err)
		}
		return clientInstance
	}
	return clientInstance
}
