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

type RealInterface struct {
	publicEndpoint        string
	authenticatedEndpoint string
}

func NewRealInterface(publicEndpoint string, authenticatedEndpoint string) *RealInterface {
	return &RealInterface{publicEndpoint: publicEndpoint, authenticatedEndpoint: authenticatedEndpoint}
}

func (r *RealInterface) isBinance() {}

func (r *RealInterface) PublicEndpoint() string {
	return r.publicEndpoint
}

func (r *RealInterface) AuthenticatedEndpoint() string {
	return r.authenticatedEndpoint
}
