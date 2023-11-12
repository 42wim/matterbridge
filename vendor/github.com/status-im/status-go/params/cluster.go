package params

// Define available fleets.
const (
	FleetUndefined  = ""
	FleetProd       = "eth.prod"
	FleetStaging    = "eth.staging"
	FleetTest       = "eth.test"
	FleetWakuV2Prod = "wakuv2.prod"
	FleetWakuV2Test = "wakuv2.test"
	FleetStatusTest = "status.test"
	FleetStatusProd = "status.prod"
	FleetShardsTest = "shards.test"
)

// Cluster defines a list of Ethereum nodes.
type Cluster struct {
	StaticNodes     []string `json:"staticnodes"`
	BootNodes       []string `json:"bootnodes"`
	MailServers     []string `json:"mailservers"` // list of trusted mail servers
	RendezvousNodes []string `json:"rendezvousnodes"`
}
