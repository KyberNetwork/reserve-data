package httputil

// HTTPPort define custom type for port
type HTTPPort int

const (
	// GatewayPort is the port number of API gateway service
	GatewayPort HTTPPort = iota + 8001
	// V3ServicePort is the port number of API service
	V3ServicePort
	FeedProviderPort
)
