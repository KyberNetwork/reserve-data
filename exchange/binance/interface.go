package binance

// Interface is Binance exchange API endpoints interface.
type Interface interface {
	// PublicEndpoint returns the endpoint that does not requires authentication.
	PublicEndpoint() string
	// AuthenticatedEndpoint returns the endpoint that requires authentication.
	// In simulation mode, authenticated endpoint is the Binance mock server.
	AuthenticatedEndpoint() string
}

type RealInterface struct {
	baseURL string
}

func (r *RealInterface) PublicEndpoint() string {
	return r.baseURL
}

func (r *RealInterface) AuthenticatedEndpoint() string {
	return r.baseURL
}

// NewRealInterface ...
func NewRealInterface(baseURL string) Interface {
	return &RealInterface{baseURL: baseURL}
}
