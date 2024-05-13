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

// DefaultWakuNodes is a list of "supported" fleets. This list is populated to clients UI settings.
var supportedFleets = map[string][]string{
	FleetWakuV2Prod: {"enrtree://ANEDLO25QVUGJOUTQFRYKWX6P4Z4GKVESBMHML7DZ6YK4LGS5FC5O@prod.wakuv2.nodes.status.im"},
	FleetWakuV2Test: {"enrtree://AO47IDOLBKH72HIZZOXQP6NMRESAN7CHYWIBNXDXWRJRZWLODKII6@test.wakuv2.nodes.status.im"},
	FleetShardsTest: {"enrtree://AMOJVZX4V6EXP7NTJPMAYJYST2QP6AJXYW76IU6VGJS7UVSNDYZG4@boot.test.shards.nodes.status.im"},
}

func DefaultWakuNodes(fleet string) []string {
	return supportedFleets[fleet]
}

func IsFleetSupported(fleet string) bool {
	_, ok := supportedFleets[fleet]
	return ok
}

func GetSupportedFleets() map[string][]string {
	return supportedFleets
}
