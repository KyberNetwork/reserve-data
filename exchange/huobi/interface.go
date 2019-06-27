package huobi

// Interface is Huobi exchange API endpoints interface.
type Interface interface {
	// isHuobi is a safe guard to make sure nothing outside this package can implement this interface.
	isHuobi()
	// PublicEndpoint returns the endpoint that does not requires authentication.
	PublicEndpoint() string
	// AuthenticatedEndpoint returns the endpoint that requires authentication.
	// In simulation mode, authenticated endpoint is the Huobi mock server.
	AuthenticatedEndpoint() string
}

type RealInterface struct {
	publicEndpoint        string
	authenticatedEndpoint string
}

func NewRealInterface(publicEndpoint string, authenticatedEndpoint string) *RealInterface {
	return &RealInterface{publicEndpoint: publicEndpoint, authenticatedEndpoint: authenticatedEndpoint}
}

func (r *RealInterface) isHuobi() {}

func (r *RealInterface) PublicEndpoint() string {
	return r.publicEndpoint
}

func (r *RealInterface) AuthenticatedEndpoint() string {
	return r.authenticatedEndpoint
}
