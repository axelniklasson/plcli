package util

// Conf represents a user config related to PlanetLab and plcli
type Conf struct {
	Slice           string
	DefaultHostname string
	PrivateKey      string
}

// GetConf returns the current user config
func GetConf() *Conf {
	// TODO should be loaded from ~/.plcli
	return &Conf{
		"chalmersple_2018_10_29",
		"cse-yellow.cse.chalmers.se",
		"~/.ssh/id_rsa_chalmers",
	}
}
