package huobi

// EndpointsInterface is Huobi exchange API endpoints interface.
type EndpointsInterface interface {
	// PublicEndpoint returns the endpoint that does not requires authentication.
	PublicEndpoint() string
	// AuthenticatedEndpoint returns the endpoint that requires authentication.
	// In simulation mode, authenticated endpoint is the Huobi mock server.
	AuthenticatedEndpoint() string
}

type Endpoints struct {
	baseURL string
}

func (r *Endpoints) PublicEndpoint() string {
	return r.baseURL
}

func (r *Endpoints) AuthenticatedEndpoint() string {
	return r.baseURL
}

// NewEndpoints ...
func NewEndpoints(baseURL string) EndpointsInterface {
	return &Endpoints{baseURL: baseURL}
}
