package pl

import "fmt"

// Slice models a PlanetLab slice
type Slice struct {
	Creator           int    `xmlrpc:"creator_person_id"`
	Instantiation     string `xmlrpc:"instantiation"`
	SliceAttributeIDs []int  `xmlrpc:"slice_attribute_ids"`
	Name              string `xmlrpc:"name"`
	SliceID           int    `xmlrpc:"slice_id"`
	Created           int    `xmlrpc:"created"`
	URL               string `xmlrpc:"url"`
	MaxNodes          int    `xmlrpc:"max_nodes"`
	PersonIDs         []int  `xmlrpc:"person_ids"`
	Expires           int    `xmlrpc:"expires"`
	SiteID            int    `xmlrpc:"site_id"`
	PeerSliceID       int    `xmlrpc:"peer_slice_id"`
	NodeIDs           []int  `xmlrpc:"node_ids"`
	PeerID            int    `xmlrpc:"peer_id"`
	Description       string `xmlrpc:"description"`
}

// ToString returns the string representation of a given slice
func (s Slice) ToString() string {
	return fmt.Sprintf("Slice details\n\nCreator: %d\nInstantiation: %s\nSliceAttributeIDS: %v\n"+
		"Name: %s\nSliceID: %d\nCreated: %d\nURL: %s\nMaxNodes: %d\nPersonIDs: %v\nExpires: %d\n"+
		"SiteID: %d\nPeerSliceID: %d\nNodeIDs: %v\nPeerID: %d\nDescription: %s", s.Creator, s.Instantiation,
		s.SliceAttributeIDs, s.Name, s.SliceID, s.Created, s.URL, s.MaxNodes, s.PersonIDs, s.Expires, s.SiteID,
		s.PeerSliceID, s.NodeIDs, s.PeerID, s.Description)
}

// Node models a PlanetLab node
type Node struct {
	LastUpdated       int    `xmlrpc:"last_updated"`
	BootState         string `xmlrpc:"boot_state"`
	SiteID            int    `xmlrpc:"site_id"`
	PcuIDs            []int  `xmlrpc:"pcu_ids"`
	Session           string `xmlrpc:"session"`
	SSHRSAKey         string `xmlrpc:"ssh_rsa_key"`
	LastContact       int    `xmlrpc:"last_contact"`
	PeerNodeID        int    `xmlrpc:"peer_node_id"`
	HostName          string `xmlrpc:"hostname"`
	SliceIDs          []int  `xmlrpc:"slice_ids"`
	Version           string `xmlrpc:"version"`
	PeerID            int    `xmlrpc:"peer_id"`
	NodeID            int    `xmlrpc:"node_id"`
	Key               string `xmlrpc:"key"`
	ConfFileIDs       []int  `xmlrpc:"conf_file_ids"`
	NodegroupIDs      []int  `xmlrpc:"nodegroup_ids"`
	SliceIDsWhitelist []int  `xmlrpc:"slice_ids_whitelist"`
	NodenetworkIDs    []int  `xmlrpc:"nodenetwork_ids"`
	BootNonce         string `xmlrpc:"boot_nonce"`
	DateCreated       int    `xmlrpc:"date_created"`
	Model             string `xmlrpc:"model"`
	Ports             []int  `xmlrpc:"ports"`
}
