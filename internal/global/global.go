package global

const (
	HeaderContentTypeAppJson = "application/json"

	EndpointHealth          = "/health"
	EndpointInfo            = "/info"
	EndpointRequestObject   = "/transfer"
	EndpointRequestDatabase = "/request/database"
)

var (
	// HashRing    *hashring.HashRingT
	ServerState ServerReadyT

	// ServerInstancesPool = pools.NewServerPool()
	// TransferRequestPool = pools.NewTransferRequestPool()
	// DatabaseRequestPool = pools.NewDatabaseRequestPool()
)
