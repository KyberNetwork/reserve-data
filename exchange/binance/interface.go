package binance

// Interface is Binance exchange API endpoints interface.
type Interface interface {
	// isBinance is a safe guard to make sure nothing outside this package can implement this interface.
	isBinance()
	// PublicEndpoint returns the endpoint that does not requires authentication.
	PublicEndpoint() string
	// AuthenticatedEndpoint returns the endpoint that requires authentication.
	// In simulation mode, authenticated endpoint is the Binance mock server.
	AuthenticatedEndpoint() string
}

// RealInterface ...
type RealInterface struct {
	publicEndpoint string
}

// NewRealInterface ...
func NewRealInterface(publicEndpoint string) *RealInterface {
	return &RealInterface{publicEndpoint: publicEndpoint}
}

func (r *RealInterface) isBinance() {}

// PublicEndpoint ...
func (r *RealInterface) PublicEndpoint() string {
	return r.publicEndpoint
}

// AuthenticatedEndpoint ...
func (r *RealInterface) AuthenticatedEndpoint() string {
	return r.publicEndpoint
}
