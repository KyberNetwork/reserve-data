package binance

type Interface interface {
	PublicEndpoint() string
	AuthenticatedEndpoint() string
	SocketPublicEndpoint() string
	SocketAuthenticatedEndpoint() string
}

type RealInterface struct{}

func getOrSetDefaultURL(base_url string) string {
	if len(base_url) > 1 {
		return base_url + ":5100"
	} else {
		return "http://127.0.0.1:5100"
	}

}

func (self *RealInterface) PublicEndpoint() string {
	return "https://api.binance.com"
}

func (self *RealInterface) AuthenticatedEndpoint() string {
	return "https://api.binance.com"
}

func (self *RealInterface) SocketPublicEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func (self *RealInterface) SocketAuthenticatedEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func NewRealInterface() *RealInterface {
	return &RealInterface{}
}

type SimulatedInterface struct {
	base_url string
}

func (self *SimulatedInterface) baseurl() string {
	return getOrSetDefaultURL(self.base_url)
}

func (self *SimulatedInterface) PublicEndpoint() string {
	return self.baseurl()
}

func (self *SimulatedInterface) AuthenticatedEndpoint() string {
	return self.baseurl()
}

func (self *SimulatedInterface) SocketPublicEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func (self *SimulatedInterface) SocketAuthenticatedEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func NewSimulatedInterface() *SimulatedInterface {
	return &SimulatedInterface{}
}

type KovanInterface struct {
	base_url string
}

func (self *KovanInterface) baseurl() string {
	return getOrSetDefaultURL(self.base_url)
}

func (self *KovanInterface) PublicEndpoint() string {
	return "https://api.binance.com"
}

func (self *KovanInterface) AuthenticatedEndpoint() string {
	return self.baseurl()
}

func (self *KovanInterface) SocketPublicEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func (self *KovanInterface) SocketAuthenticatedEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func NewKovanInterface() *KovanInterface {
	return &KovanInterface{}
}

type DevInterface struct{}

func (self *DevInterface) PublicEndpoint() string {
	return "https://api.binance.com"
	// return "http://192.168.25.16:5100"
}

func (self *DevInterface) AuthenticatedEndpoint() string {
	return "https://api.binance.com"
	// return "http://192.168.25.16:5100"
}

func (self *DevInterface) SocketPublicEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func (self *DevInterface) SocketAuthenticatedEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func NewDevInterface() *DevInterface {
	return &DevInterface{}
}

type SocketInterface struct{}

func (self *SocketInterface) PuclicEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func (self *SocketInterface) AuthenticatedEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func (self *SocketInterface) SocketPublicEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func (self *SocketInterface) SocketAuthenticatedEndpoint() string {
	return "wss://stream.binance.com:9443/ws/"
}

func NewSocketInterface() *SocketInterface {
	return &SocketInterface{}
}
